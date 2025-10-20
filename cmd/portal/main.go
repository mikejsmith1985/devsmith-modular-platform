package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/cmd/portal/handlers"
	_ "github.com/jackc/pgx/v4/stdlib" // Fix: Import pgx PostgreSQL driver for DB connection
)

func main() {
	// Create Gin router
	router := gin.Default()

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
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// Register authentication routes
	// Import handlers package
	// ...existing code...
	// This import is implied: "github.com/mikejsmith1985/devsmith-modular-platform/cmd/portal/handlers"
	handlers.RegisterAuthRoutes(router, dbConn)

	// Load Templ templates
	router.LoadHTMLGlob("cmd/portal/templates/*.templ")
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.templ", nil)
	})

	// Get port from environment or default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Portal service starting on port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
