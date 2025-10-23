// Package main starts the portal service for the DevSmith platform.
// The portal service provides an interface for users to access
// various features of the DevSmith platform, including authentication,
// dashboard access, and more. It serves as the entry point for users
// and handles routing, middleware, and template rendering.
package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib" // Import pgx PostgreSQL driver for DB connection
	"github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/middleware"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/common/debug"
)

func main() {
	// Get port from environment or default to 3001
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	// Create Gin router
	router := gin.Default()

	// Middleware for logging requests (skip health checks to reduce noise)
	router.Use(func(c *gin.Context) {
		if c.Request.URL.Path != "/health" {
			log.Printf("Incoming request: %s %s", c.Request.Method, c.Request.URL.Path)
		}
		c.Next()
	})

	// Health check endpoint (required for Docker health checks)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "portal",
			"status":  "healthy",
		})
	})

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "DevSmith Portal",
			"version": "0.1.0",
			"message": "Portal service is running",
		})
	})

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	dbConn, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			log.Printf("Error closing DB connection: %v", err)
		}
	}()

	// Register authentication routes
	// Import handlers package
	// ...existing code...
	// This import is implied: "github.com/mikejsmith1985/devsmith-modular-platform/cmd/portal/handlers"
	handlers.RegisterAuthRoutes(router, dbConn)

	// Register debug routes (development only)
	debug.RegisterDebugRoutes(router, "portal")

	// Register dashboard routes
	authenticated := router.Group("/")
	authenticated.Use(middleware.JWTAuthMiddleware())
	{
		authenticated.GET("/dashboard", handlers.DashboardHandler)
		authenticated.GET("/api/v1/dashboard/user", handlers.GetUserInfoHandler)
	}

	// Load templates (path works in both local dev and Docker)
	templatePath := "apps/portal/templates/*.templ"
	if _, err := os.Stat("./templates"); err == nil {
		// Running in Docker container
		templatePath = "./templates/*.templ"
	}
	router.LoadHTMLGlob(templatePath)
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.templ", nil)
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

	// Replace fmt.Printf with log.Printf for better logging consistency
	log.Printf("Portal service starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		// Replace os.Exit with proper error handling
		log.Printf("Failed to start server: %v", err)
		os.Exit(1) // Ensure the application exits with a non-zero status
	}
}
