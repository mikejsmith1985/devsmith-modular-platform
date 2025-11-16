package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/security"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/session"
)

// RedisSessionAuthMiddleware validates JWT and retrieves session from Redis
func RedisSessionAuthMiddleware(sessionStore *session.RedisStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get JWT from cookie
		tokenString, err := c.Cookie("devsmith_token")
		if err != nil {
			// No cookie found - redirect to login
			if isHTMLRequest(c) {
				c.Redirect(http.StatusFound, "/auth/github/login")
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			}
			c.Abort()
			return
		}

		// Parse JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return security.GetJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			// Invalid token - redirect to login
			if isHTMLRequest(c) {
				c.Redirect(http.StatusFound, "/auth/github/login")
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			}
			c.Abort()
			return
		}

		// Extract session_id from claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			if isHTMLRequest(c) {
				c.Redirect(http.StatusFound, "/auth/github/login")
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			}
			c.Abort()
			return
		}

		sessionID, ok := claims["session_id"].(string)
		if !ok || sessionID == "" {
			if isHTMLRequest(c) {
				c.Redirect(http.StatusFound, "/auth/github/login")
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing session_id"})
			}
			c.Abort()
			return
		}

		// Retrieve session from Redis
		sess, err := sessionStore.Get(c.Request.Context(), sessionID)
		if err != nil {
			if isHTMLRequest(c) {
				c.Redirect(http.StatusFound, "/auth/github/login")
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Session retrieval failed"})
			}
			c.Abort()
			return
		}

		if sess == nil {
			// Session not found (expired or deleted) - redirect to login
			if isHTMLRequest(c) {
				c.Redirect(http.StatusFound, "/auth/github/login")
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
			}
			c.Abort()
			return
		}

		// Store session data in context for handlers to use
		c.Set("user_id", sess.UserID)
		c.Set("github_username", sess.GitHubUsername)
		c.Set("github_token", sess.GitHubToken)
		c.Set("session_id", sessionID)
		c.Set("session_token", tokenString) // Store JWT for Portal AI Factory API calls

		// Store full session for handlers that need metadata
		c.Set("session", sess)

		// Refresh session TTL on each request
		if err := sessionStore.RefreshTTL(c.Request.Context(), sessionID); err != nil {
			// Log but don't fail - session will eventually expire
		}

		c.Next()
	}
}

// isHTMLRequest checks if the request expects HTML response
func isHTMLRequest(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, "text/html") || accept == ""
}
