package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
	"github.com/stretchr/testify/assert"
)

// TestCorrelationMiddleware_GeneratesID tests that middleware generates correlation ID when missing
func TestCorrelationMiddleware_GeneratesID(t *testing.T) {
	contextService := services.NewContextService(nil)
	middleware := CorrelationMiddleware(contextService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", http.NoBody)

	middleware(c)

	ctx := GetCorrelationContext(c)
	assert.NotNil(t, ctx, "Context should exist")
	assert.NotEmpty(t, ctx.CorrelationID, "Should generate correlation ID")
	assert.Len(t, ctx.CorrelationID, 32, "Correlation ID should be 32 characters")
}

// TestCorrelationMiddleware_PropagatesID tests that middleware preserves incoming correlation ID
func TestCorrelationMiddleware_PropagatesID(t *testing.T) {
	contextService := services.NewContextService(nil)
	middleware := CorrelationMiddleware(contextService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Correlation-ID", "incoming-123")
	c.Request = req

	middleware(c)

	ctx := GetCorrelationContext(c)
	assert.Equal(t, "incoming-123", ctx.CorrelationID, "Should preserve incoming correlation ID")
	assert.Equal(t, "incoming-123", w.Header().Get("X-Correlation-ID"), "Should set response header")
}

// TestCorrelationMiddleware_ExtractsTraceParent tests W3C traceparent format
func TestCorrelationMiddleware_ExtractsTraceParent(t *testing.T) {
	contextService := services.NewContextService(nil)
	middleware := CorrelationMiddleware(contextService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	// W3C format: version-traceid-spanid-traceflags
	req.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	c.Request = req

	middleware(c)

	ctx := GetCorrelationContext(c)
	assert.NotEmpty(t, ctx.TraceID, "Should extract trace ID")
	assert.Equal(t, "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01", ctx.TraceID)
	assert.Equal(t, "b7ad6b7169203331", ctx.SpanID, "Should extract span ID")
}

// TestCorrelationMiddleware_ExtractsCustomTraceID tests custom X-Trace-ID header
func TestCorrelationMiddleware_ExtractsCustomTraceID(t *testing.T) {
	contextService := services.NewContextService(nil)
	middleware := CorrelationMiddleware(contextService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Trace-ID", "custom-trace-123")
	c.Request = req

	middleware(c)

	ctx := GetCorrelationContext(c)
	assert.Equal(t, "custom-trace-123", ctx.TraceID, "Should extract custom trace ID")
}

// TestCorrelationMiddleware_ExtractsRequestID tests request ID extraction
func TestCorrelationMiddleware_ExtractsRequestID(t *testing.T) {
	contextService := services.NewContextService(nil)
	middleware := CorrelationMiddleware(contextService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/api/users", http.NoBody)
	req.Header.Set("X-Request-ID", "req-456")
	c.Request = req

	middleware(c)

	ctx := GetCorrelationContext(c)
	assert.Equal(t, "req-456", ctx.RequestID, "Should extract request ID")
}

// TestCorrelationMiddleware_CapturesHTTPContext tests HTTP context capture
func TestCorrelationMiddleware_CapturesHTTPContext(t *testing.T) {
	contextService := services.NewContextService(nil)
	middleware := CorrelationMiddleware(contextService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/users", http.NoBody)
	c.Request.Header.Set("X-Forwarded-For", "192.168.1.1")

	middleware(c)

	ctx := GetCorrelationContext(c)
	assert.Equal(t, "POST", ctx.Method, "Should capture HTTP method")
	assert.Equal(t, "/api/users", ctx.Path, "Should capture request path")
	assert.NotEmpty(t, ctx.RemoteAddr, "Should capture remote address")
}

// TestCorrelationMiddleware_AddsResponseHeaders tests response header setting
func TestCorrelationMiddleware_AddsResponseHeaders(t *testing.T) {
	contextService := services.NewContextService(nil)
	middleware := CorrelationMiddleware(contextService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Correlation-ID", "test-123")
	req.Header.Set("X-Trace-ID", "trace-789")
	c.Request = req

	middleware(c)

	assert.Equal(t, "test-123", w.Header().Get("X-Correlation-ID"), "Should set correlation ID in response")
	assert.Equal(t, "trace-789", w.Header().Get("X-Trace-ID"), "Should set trace ID in response")
}

// TestExtractSpanID_ValidFormat tests span ID extraction
func TestExtractSpanID_ValidFormat(t *testing.T) {
	spanID := extractSpanID("00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	assert.Equal(t, "b7ad6b7169203331", spanID, "Should extract span ID from traceparent")
}

// TestExtractSpanID_InvalidFormat tests span ID extraction with invalid format
func TestExtractSpanID_InvalidFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty", "", ""},
		{"SinglePart", "part1", ""},
		{"TwoParts", "part1-part2", ""},
		{"ThreeParts", "part1-part2-part3", "part3"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractSpanID(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestGetCorrelationContext_Exists tests context retrieval when set
func TestGetCorrelationContext_Exists(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	expected := &models.CorrelationContext{
		CorrelationID: "test-123",
		TraceID:       "trace-456",
	}
	c.Set("correlation_context", expected)

	result := GetCorrelationContext(c)
	assert.Equal(t, expected, result, "Should return stored context")
}

// TestGetCorrelationContext_Missing tests context retrieval fallback
func TestGetCorrelationContext_Missing(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Correlation-ID", "fallback-123")
	c.Request = req

	result := GetCorrelationContext(c)
	assert.NotNil(t, result, "Should create fallback context")
	assert.Equal(t, "fallback-123", result.CorrelationID, "Should use header value")
}

// TestCorrelationMiddleware_UserContextExtraction tests user ID extraction
func TestCorrelationMiddleware_UserContextExtraction(t *testing.T) {
	contextService := services.NewContextService(nil)
	middleware := CorrelationMiddleware(contextService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", http.NoBody)

	// Set user context before middleware
	userID := 123
	c.Set("user_id", userID)
	c.Set("session_id", "sess-abc123")

	middleware(c)

	ctx := GetCorrelationContext(c)
	assert.NotNil(t, ctx.UserID, "Should extract user ID")
	assert.Equal(t, userID, *ctx.UserID, "User ID should match")
	assert.Equal(t, "sess-abc123", ctx.SessionID, "Session ID should match")
}

// TestCorrelationMiddleware_EnrichesContext tests context enrichment
func TestCorrelationMiddleware_EnrichesContext(t *testing.T) {
	contextService := services.NewContextService(nil)
	middleware := CorrelationMiddleware(contextService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", http.NoBody)

	middleware(c)

	ctx := GetCorrelationContext(c)
	assert.NotEmpty(t, ctx.Hostname, "Should enrich with hostname")
	assert.NotEmpty(t, ctx.Environment, "Should enrich with environment")
	assert.NotEmpty(t, ctx.Version, "Should enrich with version")
	assert.False(t, ctx.CreatedAt.IsZero(), "Should set created timestamp")
}

// TestCorrelationMiddleware_CompleteFlow tests complete flow with all headers
func TestCorrelationMiddleware_CompleteFlow(t *testing.T) {
	contextService := services.NewContextService(nil)
	middleware := CorrelationMiddleware(contextService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/api/data", http.NoBody)
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("traceparent", "00-trace-span-flags")
	req.Header.Set("X-Request-ID", "req-456")
	c.Request = req
	c.Set("user_id", 789)

	middleware(c)

	ctx := GetCorrelationContext(c)
	assert.NotNil(t, ctx, "Context should exist")
	assert.Equal(t, "corr-123", ctx.CorrelationID)
	assert.Equal(t, "req-456", ctx.RequestID)
	assert.Equal(t, "POST", ctx.Method)
	assert.Equal(t, "/api/data", ctx.Path)
	assert.Equal(t, 789, *ctx.UserID)
	assert.NotEmpty(t, ctx.Hostname)
	assert.NotEmpty(t, ctx.Environment)

	// Verify response headers
	assert.Equal(t, "corr-123", w.Header().Get("X-Correlation-ID"))
}
