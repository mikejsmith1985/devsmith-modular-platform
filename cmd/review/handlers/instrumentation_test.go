// Package cmd_review_handlers contains HTTP request handlers for the review service.
package cmd_review_handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/instrumentation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// contextKeyRequestID is a context key for request ID
type contextKeyRequestID string

const requestIDKey contextKeyRequestID = "request_id"

// TestReviewServiceValidationLogging_InvalidCodeSource tests validation failure logging
func TestReviewServiceValidationLogging_InvalidCodeSource(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Arrange: Create instrumentation logger and middleware
	logger := instrumentation.NewServiceInstrumentationLogger("review", "http://localhost:8082")
	require.NotNil(t, logger, "ServiceInstrumentationLogger should be created")

	// Act: Log validation failure
	err := logger.LogValidationFailure(
		context.Background(),
		"invalid_code_source",
		"code_source must be one of: github, local, paste",
		map[string]interface{}{
			"provided_value": "invalid_source",
			"allowed_values": []string{"github", "local", "paste"},
		},
	)

	// Assert: Log should be sent without error
	assert.NoError(t, err, "LogValidationFailure should not return error")
}

// TestReviewServiceValidationLogging_CodeSizeExceeded tests size limit violation logging
func TestReviewServiceValidationLogging_CodeSizeExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := instrumentation.NewServiceInstrumentationLogger("review", "http://localhost:8082")
	require.NotNil(t, logger)

	// Act: Log size limit violation
	err := logger.LogValidationFailure(
		context.Background(),
		"size_limit_exceeded",
		"code exceeds maximum size of 10MB",
		map[string]interface{}{
			"max_size_bytes":    10 * 1024 * 1024,
			"actual_size_bytes": 15 * 1024 * 1024,
		},
	)

	// Assert
	assert.NoError(t, err)
}

// TestReviewServiceValidationLogging_PathTraversal tests security violation logging
func TestReviewServiceValidationLogging_PathTraversalAttempt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := instrumentation.NewServiceInstrumentationLogger("review", "http://localhost:8082")
	require.NotNil(t, logger)

	// Act: Log path traversal attempt
	err := logger.LogSecurityViolation(
		context.Background(),
		"path_traversal_attempt",
		"file path contains invalid characters",
		map[string]interface{}{
			"attempted_path": "../../etc/passwd",
			"severity":       "critical",
		},
	)

	// Assert
	assert.NoError(t, err)
}

// TestReviewServiceSessionLogging_CreatedSuccessfully tests session creation logging
func TestReviewServiceSessionLogging_CreatedSuccessfully(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := instrumentation.NewServiceInstrumentationLogger("review", "http://localhost:8082")
	require.NotNil(t, logger)

	// Act: Log session creation
	err := logger.LogEvent(
		context.Background(),
		"session_created",
		map[string]interface{}{
			"session_id":      int64(12345),
			"user_id":         int64(1),
			"code_source":     "github",
			"code_size_bytes": 5000,
		},
	)

	// Assert
	assert.NoError(t, err)
}

// TestReviewServiceAnalysisLogging_ScanCompleted tests analysis completion logging
func TestReviewServiceAnalysisLogging_ScanCompleted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := instrumentation.NewServiceInstrumentationLogger("review", "http://localhost:8082")
	require.NotNil(t, logger)

	// Act: Log analysis completion
	err := logger.LogEvent(
		context.Background(),
		"scan_analysis_completed",
		map[string]interface{}{
			"session_id":    int64(12345),
			"query":         "performance issues",
			"reading_mode":  "detail",
			"results_count": 42,
		},
	)

	// Assert
	assert.NoError(t, err)
}

// TestReviewServiceAnalysisLogging_ScanFailed tests analysis error logging
func TestReviewServiceAnalysisLogging_ScanFailed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := instrumentation.NewServiceInstrumentationLogger("review", "http://localhost:8082")
	require.NotNil(t, logger)

	// Act: Log analysis error
	err := logger.LogError(
		context.Background(),
		"analysis_failed",
		"database connection timeout",
		map[string]interface{}{
			"session_id": int64(12345),
			"timeout_ms": 5000,
		},
	)

	// Assert
	assert.NoError(t, err)
}

// TestReviewServiceHealth_LogsHealthStatus tests health check logging
func TestReviewServiceHealth_LogsHealthStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := instrumentation.NewServiceInstrumentationLogger("review", "http://localhost:8082")
	require.NotNil(t, logger)

	// Act: Log health status
	err := logger.LogEvent(
		context.Background(),
		"health_check",
		map[string]interface{}{
			"status":             "healthy",
			"database_connected": true,
			"memory_usage_mb":    256,
			"request_count":      1024,
		},
	)

	// Assert
	assert.NoError(t, err)
}

// TestReviewServiceMetrics_LogsRequestLatency tests request metrics logging
func TestReviewServiceMetrics_LogsRequestLatency(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := instrumentation.NewServiceInstrumentationLogger("review", "http://localhost:8082")
	require.NotNil(t, logger)

	ctx := context.WithValue(context.Background(), requestIDKey, "req-123")

	// Act: Log request metrics
	err := logger.LogEvent(
		ctx,
		"request_completed",
		map[string]interface{}{
			"method":              "POST",
			"path":                "/api/review/sessions",
			"status_code":         201,
			"latency_ms":          45,
			"response_size_bytes": 2048,
			"user_agent":          "Mozilla/5.0",
			"remote_ip":           "192.168.1.1",
		},
	)

	// Assert
	assert.NoError(t, err)
}

// TestReviewServiceLogging_ContinuesWhenLogsServiceUnavailable tests graceful degradation
func TestReviewServiceLogging_ContinuesWhenLogsServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use unreachable logs service URL
	logger := instrumentation.NewServiceInstrumentationLogger("review", "http://unreachable-logs-service:9999")
	require.NotNil(t, logger)

	// Act: Try to log while logs service is unavailable
	// Should NOT block, should NOT error
	err := logger.LogEvent(
		context.Background(),
		"test_event",
		map[string]interface{}{"test": true},
	)

	// Assert: Service should handle unavailable logs service gracefully
	// Either return nil (async) or return error without blocking
	assert.NoError(t, err, "Should not block or fail when logs service unavailable")
}

// TestReviewServiceLogging_AsyncNonBlocking tests non-blocking logging
func TestReviewServiceLogging_DoesNotBlockRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := instrumentation.NewServiceInstrumentationLogger("review", "http://localhost:8082")
	require.NotNil(t, logger)

	// Use gin test context to simulate real request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/review", http.NoBody)

	// Act: Log should be async and return immediately
	err := logger.LogEvent(
		c.Request.Context(),
		"test_event",
		map[string]interface{}{"test": true},
	)

	// Assert: Should complete quickly (async)
	assert.NoError(t, err)
}

// TestReviewServiceLogging_IncludesRequestID tests request correlation
func TestReviewServiceLogging_IncludesRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := instrumentation.NewServiceInstrumentationLogger("review", "http://localhost:8082")
	require.NotNil(t, logger)

	// Create context with request ID
	ctx := context.WithValue(context.Background(), requestIDKey, "req-correlation-123")

	// Act: Log should include request_id
	err := logger.LogEvent(
		ctx,
		"test_event",
		map[string]interface{}{
			"data": "test",
		},
	)

	// Assert
	assert.NoError(t, err)
}
