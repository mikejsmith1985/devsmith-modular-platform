package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/handlers"
)

func main() {
	// Initialize Gin router
	r := gin.Default()

	// Dashboard routes (middleware temporarily disabled for testing)
	r.GET("/dashboard", handlers.DashboardHandler)
	r.GET("/api/v1/dashboard/user", handlers.GetUserInfoHandler)

	// Start the server
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("[ERROR] Failed to start server: %v\n", err)
	}
}
