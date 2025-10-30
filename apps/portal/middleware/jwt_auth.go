// Package portal_middleware provides reusable middleware components for the DevSmith platform.
package portal_middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	portal_handlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/handlers"
)

// JWTAuthMiddleware validates JWT tokens and adds user claims to the context
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for token in cookies first
		tokenString, err := c.Cookie("devsmith_token")
		if err != nil || tokenString == "" {
			// Fallback to Authorization header
			tokenString = c.GetHeader("Authorization")
			if tokenString == "" {
				c.Redirect(http.StatusFound, "/login")
				return
			}
		}

		// Parse token with UserClaims structure
		token, err := jwt.ParseWithClaims(tokenString, &portal_handlers.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte("your-secret-key"), nil // Must match the key in auth_handler.go
		})

		log.Printf("[DEBUG] Authorization header: %s", c.GetHeader("Authorization"))
		log.Printf("[DEBUG] Token validation result: %v", err)

		// Log the token string for debugging
		log.Printf("[DEBUG] Token string: %s", tokenString)

		if err != nil {
			log.Printf("[DEBUG] Token validation error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		log.Printf("[DEBUG] Token valid: %v", token.Valid)
		log.Printf("[DEBUG] Token claims: %v", token.Claims)

		// Add claims to context
		if claims, ok := token.Claims.(*portal_handlers.UserClaims); ok && token.Valid {
			c.Set("user", claims)
			log.Printf("[DEBUG] User claims extracted: %v", claims)

			// Log claims after parsing
			log.Printf("[DEBUG] Parsed claims: %+v", claims)
		} else {
			log.Printf("[DEBUG] Token parsing started")
			log.Printf("[DEBUG] Token: %v", token)
			log.Printf("[DEBUG] Claims type assertion result: %v", ok)

			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}
