// Package main starts the logs service for the DevSmith platform.
package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	apphandlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/logs/handlers"
	resthandlers "github.com/mikejsmith1985/devsmith-modular-platform/cmd/logs/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/common/debug"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/instrumentation"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
	"github.com/sirupsen/logrus"
)

var dbConn *sql.DB

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	// Initialize instrumentation logger for this service
	// Note: Logs service has circular dependency prevention built in
	logsServiceURL := os.Getenv("LOGS_SERVICE_URL")
	if logsServiceURL == "" {
		logsServiceURL = "http://localhost:8082" // Default for local development
	}
	instrLogger := instrumentation.NewServiceInstrumentationLogger("logs", logsServiceURL)

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Initialize database
	dbURL := os.Getenv("DATABASE_URL")
	var err error
	dbConn, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Verify connection
	if err := dbConn.Ping(); err != nil {
		if closeErr := dbConn.Close(); closeErr != nil {
			log.Printf("[ERROR] Failed to close database: %v", closeErr)
		}
		log.Fatal("Failed to ping database:", err)
	}

	// OAuth2 configuration (for GitHub)
	required := []string{"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET", "REDIRECT_URI"}
	for _, key := range required {
		if os.Getenv(key) == "" {
			log.Printf("FATAL: %s environment variable not set", key)
			return
		}
	}
	log.Printf("OAuth configured: redirect_uri=%s", os.Getenv("REDIRECT_URI"))

	// Initialize Gin router
	router := gin.Default()

	// Middleware for logging requests (skip health checks)
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

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		//nolint:errcheck,gosec // Logger always returns nil, safe to ignore
		instrLogger.LogEvent(c.Request.Context(), "health_check", map[string]interface{}{
			"status": "healthy",
		})
		c.JSON(http.StatusOK, gin.H{
			"service": "logs",
			"status":  "healthy",
		})
	})

	// Serve static files for logs dashboard
	router.Static("/static", "./apps/logs/static")

	// Register UI routes for dashboard
	apphandlers.RegisterUIRoutes(router, logger)

	// Register debug routes (development only)
	debug.RegisterDebugRoutes(router, "logs")

	// Initialize database repositories for REST API
	logRepo := db.NewLogRepository(dbConn)
	restSvc := services.NewRestLogService(logRepo, logger)

	// Register REST API routes
	router.POST("/api/logs", func(c *gin.Context) {
		resthandlers.PostLogs(restSvc)(c)
	})
	router.GET("/api/logs", func(c *gin.Context) {
		resthandlers.GetLogs(restSvc)(c)
	})
	router.GET("/api/logs/:id", func(c *gin.Context) {
		resthandlers.GetLogByID(restSvc)(c)
	})
	router.GET("/api/logs/stats", func(c *gin.Context) {
		resthandlers.GetStats(restSvc)(c)
	})
	router.DELETE("/api/logs", func(c *gin.Context) {
		resthandlers.DeleteLogs(restSvc)(c)
	})

	// Also register /api/v1/logs routes (for consistency and direct access)
	router.POST("/api/v1/logs", func(c *gin.Context) {
		resthandlers.PostLogs(restSvc)(c)
	})
	router.GET("/api/v1/logs", func(c *gin.Context) {
		resthandlers.GetLogs(restSvc)(c)
	})
	router.GET("/api/v1/logs/:id", func(c *gin.Context) {
		resthandlers.GetLogByID(restSvc)(c)
	})
	router.GET("/api/v1/logs/stats", func(c *gin.Context) {
		resthandlers.GetStats(restSvc)(c)
	})
	router.DELETE("/api/v1/logs", func(c *gin.Context) {
		resthandlers.DeleteLogs(restSvc)(c)
	})

	// Initialize WebSocket hub
	hub := services.NewWebSocketHub()
	go hub.Run()

	// Register WebSocket routes
	services.RegisterWebSocketRoutes(router, hub)

	log.Printf("Starting logs service on port %s", port)

	// Create an HTTP server with timeouts
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		if closeErr := dbConn.Close(); closeErr != nil {
			log.Printf("[ERROR] Failed to close database: %v", closeErr)
		}
		log.Fatalf("[ERROR] Failed to start server: %v", err)
	}
}
