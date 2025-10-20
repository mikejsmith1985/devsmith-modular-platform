// Package handlers contains HTTP handlers for the portal service.
package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/services"

	// "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/interfaces"
	"database/sql"

	"github.com/rs/zerolog"
)

// RegisterAuthRoutes registers authentication-related routes for the portal service.
func RegisterAuthRoutes(r *gin.Engine, dbConn *sql.DB) {
	logger := zerolog.New(os.Stdout)
	userRepo := db.NewUserRepository(dbConn)
	githubClient := services.NewGitHubClient(os.Getenv("GITHUB_CLIENT_ID"), os.Getenv("GITHUB_CLIENT_SECRET"))
	authService := services.NewAuthService(userRepo, githubClient, os.Getenv("JWT_SECRET"), &logger)

	r.GET("/auth/github/login", func(c *gin.Context) {
		clientID := os.Getenv("GITHUB_CLIENT_ID")
		redirectURI := os.Getenv("GITHUB_REDIRECT_URI")
		url := "https://github.com/login/oauth/authorize?client_id=" + clientID + "&redirect_uri=" + redirectURI + "&scope=read:user user:email"
		c.Redirect(http.StatusFound, url)
	})

	r.GET("/auth/github/callback", func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
			return
		}
		user, token, err := authService.AuthenticateWithGitHub(c.Request.Context(), code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		// Set JWT as cookie
		c.SetCookie("devsmith_token", token, 86400, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
	})

	r.POST("/auth/logout", func(c *gin.Context) {
		token, err := c.Cookie("devsmith_token")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No token"})
			return
		}
		err = authService.RevokeSession(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.SetCookie("devsmith_token", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
	})
}
