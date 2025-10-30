// Package cmd_logs_middleware provides HTTP middleware for the logs service.
// This package handles correlation context extraction and propagation for distributed tracing.
package cmd_logs_middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

// HTTP header constants for correlation context
const (
	// HeaderCorrelationID is the header for correlation ID
	HeaderCorrelationID = "X-Correlation-ID"

	// HeaderTraceID is the header for OpenTelemetry trace ID
	HeaderTraceID = "X-Trace-ID"

	// HeaderTraceparent is the W3C Trace Context standard header
	HeaderTraceparent = "traceparent"

	// ContextKey is the key for storing correlation context in Gin request
	ContextKey = "correlation_context"

	// TraceparentMinParts is the minimum number of parts needed for span ID extraction (version-traceid-spanid)
	TraceparentMinParts = 3

	// TraceparentSpanIDIndex is the index of span ID in traceparent format (version-traceid-spanid-[traceflags])
	TraceparentSpanIDIndex = 2
)

// CorrelationMiddleware creates HTTP middleware that manages correlation context for distributed tracing.
//
// The middleware performs the following operations:
// 1. Extracts or generates a correlation ID for the request
// 2. Parses OpenTelemetry W3C traceparent headers (version-traceid-spanid-traceflags)
// 3. Captures HTTP context (method, path, remote address)
// 4. Extracts user context if authenticated (user_id, session_id)
// 5. Enriches context with automatic metadata (hostname, environment, version)
// 6. Stores context in request for downstream handlers
// 7. Sets response headers for trace propagation
//
// Example:
//
//	contextService := logs_services.NewContextService(repo)
//	router.Use(CorrelationMiddleware(contextService))
func CorrelationMiddleware(contextService *logs_services.ContextService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract or generate correlation ID
		correlationID := c.GetHeader(HeaderCorrelationID)
		if correlationID == "" {
			correlationID = contextService.GenerateCorrelationID()
		}

		// Extract trace ID (OpenTelemetry)
		traceID := c.GetHeader(HeaderTraceparent) // W3C Trace Context
		if traceID == "" {
			traceID = c.GetHeader(HeaderTraceID) // Custom header fallback
		}

		// Extract OpenTelemetry span ID from W3C traceparent format
		spanID := extractSpanID(traceID)

		// Build correlation context from request
		ctx := &logs_models.CorrelationContext{
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

		// Enrich with automatic metadata (hostname, environment, version, timestamps)
		ctx = contextService.EnrichContext(ctx)

		// Store in request context for use by downstream handlers
		c.Set(ContextKey, ctx)

		// Add response headers for trace propagation to other services
		c.Header(HeaderCorrelationID, correlationID)
		if traceID != "" {
			c.Header(HeaderTraceID, traceID)
		}

		c.Next()
	}
}

// extractSpanID extracts the span ID from W3C Trace Context format.
//
// W3C Trace Context (traceparent) format: version-traceid-spanid-traceflags
// Example: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
//
// Returns empty string if format is invalid or doesn't contain a span ID.
//
// Specification: https://www.w3.org/TR/trace-context/
func extractSpanID(traceparent string) string {
	if traceparent == "" {
		return ""
	}

	parts := strings.Split(traceparent, "-")
	if len(parts) >= TraceparentMinParts {
		return parts[TraceparentSpanIDIndex]
	}
	return ""
}

// GetCorrelationContext retrieves the correlation context from a Gin request.
//
// Returns the correlation context if present in the request,
// otherwise returns a minimal context with just the correlation ID from headers.
//
// This function is safe to call even if the middleware hasn't been applied.
//
// Example:
//
//	func MyHandler(c *gin.Context) {
//	    ctx := GetCorrelationContext(c)
//	    log.Printf("Correlation ID: %s", ctx.CorrelationID)
//	}
func GetCorrelationContext(c *gin.Context) *logs_models.CorrelationContext {
	if ctx, exists := c.Get(ContextKey); exists {
		if correlationCtx, ok := ctx.(*logs_models.CorrelationContext); ok {
			return correlationCtx
		}
	}

	// Fallback: create minimal context from headers
	return &logs_models.CorrelationContext{
		CorrelationID: c.GetHeader(HeaderCorrelationID),
	}
}
