// Package main is the entry point for the Portal service.
package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	portal_handlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/handlers"
)

func main() {
	// Initialize Gin router
	r := gin.Default()

	// Dashboard routes (middleware temporarily disabled for testing)
	r.GET("/dashboard", portal_handlers.DashboardHandler)
	r.GET("/api/v1/dashboard/user", portal_handlers.GetUserInfoHandler)

	// Start the server
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("[ERROR] Failed to start server: %v\n", err)
	}
}
