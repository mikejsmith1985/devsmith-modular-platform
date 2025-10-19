package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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

	// Get port from environment or default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Portal service starting on port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
