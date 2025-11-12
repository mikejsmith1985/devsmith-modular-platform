// Package middleware provides HTTP middleware for the logs service.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	logs_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
)

// SimpleAPITokenAuth validates X-API-Key header using plain token lookup.
// This is simpler and faster than bcrypt hashing (O(1) vs O(n) complexity).
//
// Authentication Flow:
// 1. Extract X-API-Key from header
// 2. Query logs.projects with indexed lookup (fast!)
// 3. Validate project is active
// 4. Set project in context for handler use
//
// Example external usage (Node.js):
//
//	fetch('https://devsmith.example.com/api/logs/batch', {
//	  method: 'POST',
//	  headers: {
//	    'X-API-Key': 'dsk_abc123...',
//	    'Content-Type': 'application/json'
//	  },
//	  body: JSON.stringify({
//	    project_slug: 'my-nodejs-app',
//	    logs: [...]
//	  })
//	});
//
// Security: API tokens are NOT passwords. They're transmitted over HTTPS,
// stored plain-text in database (indexed), and rotated frequently.
// This is standard practice for API authentication (GitHub, Stripe, etc.).
func SimpleAPITokenAuth(projectRepo *logs_db.ProjectRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract API key from header
		token := c.GetHeader("X-API-Key")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Missing X-API-Key header",
				"message": "API key is required. Get your key from DevSmith Portal.",
			})
			c.Abort()
			return
		}

		// Single database query with indexed lookup (O(1) performance)
		project, err := projectRepo.FindByAPIToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid API key",
				"message": "API key not found or inactive. Check your key in DevSmith Portal.",
			})
			c.Abort()
			return
		}

		// Verify project is active
		if !project.IsActive {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Project disabled",
				"message": "This project has been deactivated. Contact support or reactivate in Portal.",
			})
			c.Abort()
			return
		}

		// Set project in context for handler
		c.Set("project", project)
		c.Next()
	}
}
