// Package middleware contains HTTP middleware for the logs service.
package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

// CorrelationMiddleware adds correlation context to requests (GREEN phase - full implementation)
func CorrelationMiddleware(contextService *services.ContextService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract or generate correlation ID
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = contextService.GenerateCorrelationID()
		}

		// Extract trace ID (OpenTelemetry)
		traceID := c.GetHeader("traceparent") // W3C Trace Context
		if traceID == "" {
			traceID = c.GetHeader("X-Trace-ID") // Custom header
		}

		// Extract OpenTelemetry span ID
		spanID := extractSpanID(traceID)

		// Build correlation context
		ctx := &models.CorrelationContext{
			CorrelationID: correlationID,
			TraceID:       traceID,
			SpanID:        spanID,
			RequestID:     c.GetHeader("X-Request-ID"),
			Method:        c.Request.Method,
			Path:          c.Request.URL.Path,
			RemoteAddr:    c.ClientIP(),
			Service:       os.Getenv("SERVICE_NAME"),
		}

		// Extract user context if authenticated
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(int); ok {
				ctx.UserID = &uid
			}
		}

		if sessionID, exists := c.Get("session_id"); exists {
			if sid, ok := sessionID.(string); ok {
				ctx.SessionID = sid
			}
		}

		// Enrich with automatic metadata
		ctx = contextService.EnrichContext(ctx)

		// Store in request context
		c.Set("correlation_context", ctx)

		// Add response headers for tracing
		c.Header("X-Correlation-ID", correlationID)
		if traceID != "" {
			c.Header("X-Trace-ID", traceID)
		}

		c.Next()
	}
}

// extractSpanID extracts span ID from W3C traceparent format (GREEN phase)
// Format: traceparent: version-traceid-spanid-traceflags
func extractSpanID(traceparent string) string {
	if traceparent == "" {
		return ""
	}

	parts := strings.Split(traceparent, "-")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

// GetCorrelationContext retrieves correlation context from request (GREEN phase)
func GetCorrelationContext(c *gin.Context) *models.CorrelationContext {
	if ctx, exists := c.Get("correlation_context"); exists {
		if correlationCtx, ok := ctx.(*models.CorrelationContext); ok {
			return correlationCtx
		}
	}

	// Fallback: create minimal context
	return &models.CorrelationContext{
		CorrelationID: c.GetHeader("X-Correlation-ID"),
	}
}
