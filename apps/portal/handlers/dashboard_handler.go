package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/templates"
)

// DashboardHandler serves the main dashboard page
func DashboardHandler(c *gin.Context) {
	// Extract user from JWT (middleware already validated)
	claims, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	user := templates.DashboardUser{
		Username:  userClaims["username"].(string),
		Email:     userClaims["email"].(string),
		AvatarURL: userClaims["avatar_url"].(string),
	}

	// Debug logging
	log.Printf("DashboardUser: %+v\n", user)

	// Render dashboard template
	component := templates.Dashboard(user)
	// Render does not return an error; linter warning is a false positive.
	component.Render(c.Request.Context(), c.Writer)
}

// GetUserInfoHandler returns current user info as JSON
func GetUserInfoHandler(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user claims"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username":   userClaims["username"],
		"email":      userClaims["email"],
		"avatar_url": userClaims["avatar_url"],
		"github_id":  userClaims["github_id"],
		"created_at": userClaims["created_at"],
	})
}
