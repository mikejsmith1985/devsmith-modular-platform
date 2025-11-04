// Package internal_review_middleware provides middleware components for the Review service
package internal_review_middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/security"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// JWTAuthMiddleware validates JWT tokens and adds user claims to the context
// This middleware enforces authentication for all protected Review service endpoints
func JWTAuthMiddleware(log logger.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for token in cookies first (primary method)
		tokenString, err := c.Cookie("devsmith_token")
		if err != nil || tokenString == "" {
			// Fallback to Authorization header (for API clients)
			tokenString = c.GetHeader("Authorization")
			if tokenString == "" {
				log.Warn("Authentication required - no token provided", "path", c.Request.URL.Path)
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Authentication required",
					"message": "Please log in to access this resource",
				})
				c.Abort()
				return
			}
		}

		// Validate token using shared security package
		claims, err := security.ValidateJWT(tokenString)
		if err != nil {
			log.Warn("Invalid JWT token", "error", err.Error(), "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authentication token",
				"message": "Your session may have expired. Please log in again.",
			})
			c.Abort()
			return
		}

		// Add claims to context for downstream handlers
		c.Set("user", claims)
		c.Set("user_id", claims.GithubID) // Convenient access to user_id
		c.Set("username", claims.Username)

		log.Info("User authenticated",
			"user_id", claims.GithubID,
			"username", claims.Username,
			"path", c.Request.URL.Path)

		c.Next()
	}
}

// OptionalAuthMiddleware validates JWT if present, but doesn't require it
// Useful for endpoints that work both authenticated and unauthenticated
func OptionalAuthMiddleware(log logger.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for token
		tokenString, err := c.Cookie("devsmith_token")
		if err != nil || tokenString == "" {
			tokenString = c.GetHeader("Authorization")
		}

		// If no token, continue without authentication
		if tokenString == "" {
			log.Debug("No authentication token provided - continuing as unauthenticated")
			c.Next()
			return
		}

		// If token exists, validate it
		claims, err := security.ValidateJWT(tokenString)
		if err != nil {
			log.Warn("Invalid JWT token in optional auth", "error", err.Error())
			// Continue anyway (don't block request)
			c.Next()
			return
		}

		// Add claims to context if valid
		c.Set("user", claims)
		c.Set("user_id", claims.GithubID)
		c.Set("username", claims.Username)

		log.Debug("User authenticated (optional auth)", "user_id", claims.GithubID)
		c.Next()
	}
}
