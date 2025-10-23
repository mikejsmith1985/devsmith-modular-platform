package handlers

import (
	"database/sql"
	"net/http"

	"log"

	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers authentication-related routes
func RegisterAuthRoutes(router *gin.Engine, dbConn *sql.DB) {
	log.Println("[DEBUG] Registering authentication routes")

	// Login route redirects to GitHub OAuth
	router.GET("/auth/login", func(c *gin.Context) {
		log.Println("[DEBUG] Login route registered")
		redirectURL := "https://github.com/login/oauth/authorize?client_id=YOUR_CLIENT_ID"
		log.Printf("[DEBUG] Redirecting to GitHub OAuth: %s", redirectURL)
		c.Redirect(http.StatusFound, redirectURL)
	})

	// Dashboard route (placeholder)
	router.GET("/auth/github/dashboard", func(c *gin.Context) {
		log.Println("[DEBUG] Dashboard route registered")
		userID := c.Query("user_id")
		if userID == "" {
			log.Println("[ERROR] Missing user_id in dashboard request")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_id"})
			return
		}
		log.Printf("[DEBUG] Dashboard accessed by user_id: %s", userID)
		c.JSON(http.StatusOK, gin.H{"message": "Dashboard route", "user_id": userID})
	})
}
