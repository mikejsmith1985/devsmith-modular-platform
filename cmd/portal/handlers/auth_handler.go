// Package cmd_portal_handlers contains HTTP handlers for the portal service.
package cmd_portal_handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	portal_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/db"
	portal_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/services"

	// "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/interfaces"
	"database/sql"

	"github.com/rs/zerolog"
)

// RegisterAuthRoutes registers authentication-related routes for the portal service.
func RegisterAuthRoutes(r *gin.Engine, dbConn *sql.DB) {
	logger := zerolog.New(os.Stdout)
	userRepo := portal_db.NewUserRepository(dbConn)
	githubClient := portal_services.NewGitHubClient(os.Getenv("GITHUB_CLIENT_ID"), os.Getenv("GITHUB_CLIENT_SECRET"))
	authService := portal_services.NewAuthService(userRepo, githubClient, os.Getenv("JWT_SECRET"), &logger, nil, nil)

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
		_, token, err := authService.AuthenticateWithGitHub(c.Request.Context(), code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Set cookie
		c.SetCookie(
			"devsmith_token",
			token,
			3600*24*7, // 7 days
			"/",
			"",
			false, // secure - set to true in production
			true,  // httpOnly
		)

		// Redirect to React frontend with token in URL for React to store in localStorage
		// React will extract token from URL and redirect to dashboard
		redirectURL := "http://localhost:3000/auth/callback?token=" + token
		c.Redirect(http.StatusFound, redirectURL)
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
