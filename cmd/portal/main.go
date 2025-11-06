// Package main starts the portal service for the DevSmith platform.
// The portal service provides an interface for users to access
// various features of the DevSmith platform, including authentication,
// dashboard access, and more. It serves as the entry point for users
// and handles routing, middleware, and template rendering.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib" // Import pgx PostgreSQL driver for DB connection
	handlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/common/debug"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/config"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/instrumentation"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/middleware"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/session"
)

func main() {
	// Get port from environment or default to 3001
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	// Create Gin router
	router := gin.Default()

	// Initialize instrumentation logger for this service (use validated config)
	logsServiceURL, logsEnabled, err := config.LoadLogsConfigWithFallbackFor("portal")
	if err != nil {
		log.Fatalf("Failed to load logging configuration: %v", err)
	}
	if !logsEnabled {
		log.Printf("Instrumentation/logging disabled: continuing startup without external logs")
		logsServiceURL = "" // instrumentation will treat empty URL as disabled
	}
	instrLogger := instrumentation.NewServiceInstrumentationLogger("portal", logsServiceURL)

	// Middleware for logging requests (skip health checks to reduce noise)
	router.Use(func(c *gin.Context) {
		if c.Request.URL.Path != "/health" {
			log.Printf("Incoming request: %s %s", c.Request.Method, c.Request.URL.Path)
			// Log to instrumentation service asynchronously
			//nolint:errcheck,gosec // Logger always returns nil, safe to ignore
			instrLogger.LogEvent(c.Request.Context(), "request_received", map[string]interface{}{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
			})
		}
		c.Next()
	})

	// Health check endpoint (required for Docker health checks)
	router.GET("/health", func(c *gin.Context) {
		//nolint:errcheck,gosec // Logger always returns nil, safe to ignore
		instrLogger.LogEvent(c.Request.Context(), "health_check", map[string]interface{}{
			"status": "healthy",
		})
		c.JSON(http.StatusOK, gin.H{
			"service": "portal",
			"status":  "healthy",
		})
	})

	// Root endpoint: render login page
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	dbConn, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return
	}

	// Configure connection pool to prevent exhaustion
	dbConn.SetMaxOpenConns(10)               // Max 10 connections per service
	dbConn.SetMaxIdleConns(5)                // Keep 5 idle
	dbConn.SetConnMaxLifetime(3600000000000) // 1 hour
	dbConn.SetConnMaxIdleTime(600000000000)  // 10 minutes

	// Ping the database to verify connection
	if err := dbConn.Ping(); err != nil {
		log.Printf("Failed to ping database: %v", err)
		if closeErr := dbConn.Close(); closeErr != nil {
			log.Printf("Error closing DB connection: %v", closeErr)
		}
		return
	}

	// Initialize Redis session store
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379" // Default for local development
	}
	sessionStore, err := session.NewRedisStore(redisURL, 7*24*time.Hour) // 7 day session TTL
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
		if closeErr := dbConn.Close(); closeErr != nil {
			log.Printf("Error closing DB connection: %v", closeErr)
		}
		return
	}
	defer func() {
		if closeErr := sessionStore.Close(); closeErr != nil {
			log.Printf("Error closing Redis connection: %v", closeErr)
		}
	}()
	log.Printf("Redis session store initialized at %s", redisURL)

	// Register authentication routes (pass session store)
	handlers.RegisterAuthRoutesWithSession(router, dbConn, sessionStore)

	// Register debug routes (development only)
	debug.RegisterDebugRoutes(router, "portal")

	// Register dashboard routes (use RedisSessionAuth for SSO)
	authenticated := router.Group("/")
	authenticated.Use(middleware.RedisSessionAuthMiddleware(sessionStore))
	authenticated.GET("/dashboard", handlers.DashboardHandler)
	authenticated.GET("/dashboard/logs", handlers.LogsDashboardHandler)
	authenticated.GET("/api/v1/dashboard/user", handlers.GetUserInfoHandler)

	// Load templates (path works in both local dev and Docker)
	templatePath := "apps/portal/templates/*.html"
	if _, err := os.Stat("./templates"); err == nil {
		// Running in Docker container
		templatePath = "./templates/*.html"
	}
	router.LoadHTMLGlob(templatePath)
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	// Serve static files (path works in both local dev and Docker)
	staticPath := "apps/portal/static"
	if _, err := os.Stat("./static"); err == nil {
		// Running in Docker container
		staticPath = "./static"
	}
	router.Static("/static", staticPath)

	// Custom 404 handler
	router.NoRoute(func(c *gin.Context) {
		log.Printf("404 Not Found: %s", c.Request.URL.Path)
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
	})

	// Validate required OAuth environment variables
	if err := validateOAuthEnvironment(); err != nil {
		log.Printf("FATAL: %v", err)
		if closeErr := dbConn.Close(); closeErr != nil {
			log.Printf("Error closing DB connection: %v", closeErr)
		}
		os.Exit(1)
	}
	log.Printf("OAuth configured: redirect_uri=%s", os.Getenv("REDIRECT_URI"))

	// Replace fmt.Printf with log.Printf for better logging consistency
	log.Printf("Portal service starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		// Replace os.Exit with proper error handling
		log.Printf("Failed to start server: %v", err)
		if closeErr := dbConn.Close(); closeErr != nil {
			log.Printf("Error closing DB connection: %v", closeErr)
		}
		os.Exit(1) // Ensure the application exits with a non-zero status
	}
}

func validateOAuthEnvironment() error {
	required := []string{"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET", "REDIRECT_URI"}
	for _, key := range required {
		if os.Getenv(key) == "" {
			return fmt.Errorf("%s environment variable not set", key)
		}
	}
	return nil
}
