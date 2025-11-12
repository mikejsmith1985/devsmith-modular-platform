// Gin Middleware for DevSmith Logging
//
// Automatically logs HTTP requests/responses to DevSmith platform.
//
// Installation:
// 1. Copy logger.go into your project
// 2. Copy this file (gin_middleware.go) into your project
// 3. Add to your Gin router
//
// Usage:
//   import (
//       "github.com/gin-gonic/gin"
//       "os"
//   )
//
//   func main() {
//       logger := NewLogger(
//           os.Getenv("DEVSMITH_API_KEY"),
//           os.Getenv("DEVSMITH_API_URL"),
//           "my-app",
//           "gin-api",
//       )
//       defer logger.Close()
//
//       router := gin.Default()
//       router.Use(DevSmithMiddleware(logger, DevSmithMiddlewareConfig{}))
//
//       router.GET("/", func(c *gin.Context) {
//           c.String(200, "Hello World")
//       })
//
//       router.Run(":8080")
//   }

package main

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// DevSmithMiddlewareConfig configures the DevSmith logging middleware
type DevSmithMiddlewareConfig struct {
	// LogBody enables logging of request/response bodies (default: false)
	LogBody bool

	// SkipPaths is a list of paths to skip logging (e.g., ["/health", "/metrics"])
	SkipPaths []string

	// RedactHeaders is a list of headers to redact (e.g., ["Authorization", "Cookie"])
	RedactHeaders []string

	// MaxBodySize limits the size of logged bodies in bytes (default: 1024)
	MaxBodySize int
}

// responseWriter is a wrapper around gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// DevSmithMiddleware creates a Gin middleware for automatic request/response logging
//
// Args:
//
//	logger: DevSmithLogger instance
//	config: Middleware configuration (optional)
//
// Returns:
//
//	gin.HandlerFunc
func DevSmithMiddleware(logger *DevSmithLogger, config DevSmithMiddlewareConfig) gin.HandlerFunc {
	// Set defaults
	if config.SkipPaths == nil {
		config.SkipPaths = []string{"/health", "/metrics"}
	}
	if config.RedactHeaders == nil {
		config.RedactHeaders = []string{"Authorization", "Cookie", "X-Api-Key"}
	}
	if config.MaxBodySize == 0 {
		config.MaxBodySize = 1024 // 1KB
	}

	// Build skip path map for O(1) lookup
	skipPathMap := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPathMap[path] = true
	}

	// Build redact header map for O(1) lookup
	redactHeaderMap := make(map[string]bool)
	for _, header := range config.RedactHeaders {
		redactHeaderMap[header] = true
	}

	return func(c *gin.Context) {
		// Skip logging for configured paths
		if skipPathMap[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Record start time
		startTime := time.Now()

		// Capture request body (if logging enabled)
		var requestBody []byte
		if config.LogBody && c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Restore body for handler to read
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Wrap response writer to capture response body
		var responseBody *bytes.Buffer
		if config.LogBody {
			responseBody = &bytes.Buffer{}
			writer := &responseWriter{
				ResponseWriter: c.Writer,
				body:           responseBody,
			}
			c.Writer = writer
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Determine log level based on status code
		level := "INFO"
		if c.Writer.Status() >= 500 {
			level = "ERROR"
		} else if c.Writer.Status() >= 400 {
			level = "WARN"
		}

		// Build context
		context := map[string]interface{}{
			// Request info
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.Query(),
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),

			// Response info
			"status_code": c.Writer.Status(),
			"duration":    duration.String(),

			// Headers (redacted)
			"request_headers":  redactHeaders(c.Request.Header, redactHeaderMap),
			"response_headers": redactHeaders(c.Writer.Header(), redactHeaderMap),
		}

		// Optionally include bodies
		if config.LogBody {
			if len(requestBody) > 0 {
				if len(requestBody) > config.MaxBodySize {
					context["request_body"] = string(requestBody[:config.MaxBodySize]) + "..."
				} else {
					context["request_body"] = string(requestBody)
				}
			}
			if responseBody != nil && responseBody.Len() > 0 {
				if responseBody.Len() > config.MaxBodySize {
					context["response_body"] = responseBody.String()[:config.MaxBodySize] + "..."
				} else {
					context["response_body"] = responseBody.String()
				}
			}
		}

		// Add error if exists
		if len(c.Errors) > 0 {
			context["errors"] = c.Errors.String()
		}

		// Log message
		message := c.Request.Method + " " + c.Request.URL.Path + " " +
			string(rune(c.Writer.Status())) + " " + duration.String()

		// Log with appropriate level
		switch level {
		case "ERROR":
			logger.Error(message, context)
		case "WARN":
			logger.Warn(message, context)
		default:
			logger.Info(message, context)
		}
	}
}

// redactHeaders creates a copy of headers with sensitive values redacted
func redactHeaders(headers map[string][]string, redactMap map[string]bool) map[string]string {
	redacted := make(map[string]string)
	for key, values := range headers {
		if redactMap[key] {
			redacted[key] = "[REDACTED]"
		} else if len(values) > 0 {
			redacted[key] = values[0] // Use first value
		}
	}
	return redacted
}

// DevSmithErrorHandler creates a middleware to log panic recovery
//
// Usage:
//
//	router.Use(DevSmithErrorHandler(logger))
func DevSmithErrorHandler(logger *DevSmithLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered", map[string]interface{}{
					"method": c.Request.Method,
					"path":   c.Request.URL.Path,
					"error":  err,
					"ip":     c.ClientIP(),
				})

				// Return 500 error
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}

// Example usage
/*
package main

import (
	"os"
	"github.com/gin-gonic/gin"
)

func main() {
	// Create logger
	logger := NewLogger(
		os.Getenv("DEVSMITH_API_KEY"),
		os.Getenv("DEVSMITH_API_URL"),
		"test-project",
		"gin-api",
	)
	defer logger.Close()

	// Create router
	router := gin.Default()

	// Add DevSmith middleware
	router.Use(DevSmithMiddleware(logger, DevSmithMiddlewareConfig{
		LogBody:       false, // Set to true to log request/response bodies
		SkipPaths:     []string{"/health", "/metrics"},
		RedactHeaders: []string{"Authorization", "Cookie", "X-Api-Key"},
		MaxBodySize:   1024, // 1KB
	}))

	// Add panic recovery middleware
	router.Use(DevSmithErrorHandler(logger))

	// Routes
	router.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World")
	})

	router.GET("/api/users", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"users": []gin.H{
				{"id": 1, "name": "Alice"},
				{"id": 2, "name": "Bob"},
			},
		})
	})

	router.GET("/api/error", func(c *gin.Context) {
		// This will trigger error logging
		panic("Intentional panic for testing")
	})

	// Start server
	router.Run(":8080")
}
*/
