package review_handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// DebugHandler handles debug/testing endpoints
type DebugHandler struct {
	tracer trace.Tracer
}

// NewDebugHandler creates a new debug handler
func NewDebugHandler() *DebugHandler {
	return &DebugHandler{
		tracer: otel.Tracer("devsmith-review"),
	}
}

// HandleTraceTest creates a test span for Jaeger validation
func (h *DebugHandler) HandleTraceTest(c *gin.Context) {
	ctx := c.Request.Context()

	// Start span
	ctx, span := h.tracer.Start(ctx, "debug.trace.test")
	defer span.End()

	// Add attributes
	span.SetAttributes(
		attribute.String("debug.mode", "trace_validation"),
		attribute.String("debug.endpoint", "/debug/trace"),
		attribute.Int64("debug.timestamp", time.Now().Unix()),
		attribute.String("http.method", c.Request.Method),
		attribute.String("http.path", c.Request.URL.Path),
	)

	// Simulate some work
	time.Sleep(50 * time.Millisecond)

	// Create nested span
	_, childSpan := h.tracer.Start(ctx, "debug.trace.test.nested")
	childSpan.SetAttributes(
		attribute.String("debug.nested", "true"),
		attribute.String("debug.operation", "test_nested_span"),
	)
	time.Sleep(25 * time.Millisecond)
	childSpan.SetStatus(codes.Ok, "Nested span completed")
	childSpan.End()

	// Mark main span as successful
	span.SetStatus(codes.Ok, "Trace test completed successfully")

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"message":      "Test span created successfully",
		"trace_id":     span.SpanContext().TraceID().String(),
		"span_id":      span.SpanContext().SpanID().String(),
		"instructions": "Check Jaeger UI at http://localhost:16686 for service 'devsmith-review'",
		"jaeger_query": "http://localhost:16686/search?service=devsmith-review",
	})
}

// RegisterDebugRoutes registers debug endpoints (TODO: remove in production or guard with env flag)
func RegisterDebugRoutes(router *gin.Engine) {
	handler := NewDebugHandler()

	debug := router.Group("/debug")
	{
		debug.GET("/trace", handler.HandleTraceTest)
		debug.POST("/trace", handler.HandleTraceTest)
	}
}
