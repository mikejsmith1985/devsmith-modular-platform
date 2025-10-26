package services

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/assert"
)

// TestContextService_GenerateCorrelationID_Unique tests that correlation IDs are unique
func TestContextService_GenerateCorrelationID_Unique(t *testing.T) {
	service := NewContextService(nil)

	id1 := service.GenerateCorrelationID()
	id2 := service.GenerateCorrelationID()

	assert.NotEmpty(t, id1, "Correlation ID should not be empty")
	assert.NotEmpty(t, id2, "Correlation ID should not be empty")
	assert.NotEqual(t, id1, id2, "Correlation IDs should be unique")
	assert.Len(t, id1, 32, "Correlation ID should be 32 characters (hex encoded 16 bytes)")
	assert.Len(t, id2, 32, "Correlation ID should be 32 characters (hex encoded 16 bytes)")
}

// TestContextService_GenerateCorrelationID_Format tests that correlation IDs are valid hex
func TestContextService_GenerateCorrelationID_Format(t *testing.T) {
	service := NewContextService(nil)

	for i := 0; i < 10; i++ {
		id := service.GenerateCorrelationID()

		// Should be valid hex string
		assert.Regexp(t, `^[a-f0-9]{32}$`, id, "Correlation ID should be valid hex")
	}
}

// TestContextService_EnrichContext_GeneratesID tests that enrichment generates missing ID
func TestContextService_EnrichContext_GeneratesID(t *testing.T) {
	service := NewContextService(nil)
	ctx := &models.CorrelationContext{}

	enriched := service.EnrichContext(ctx)

	assert.NotNil(t, enriched, "Enriched context should not be nil")
	assert.NotEmpty(t, enriched.CorrelationID, "Should generate correlation ID")
	assert.Len(t, enriched.CorrelationID, 32, "Correlation ID should be 32 characters")
}

// TestContextService_EnrichContext_PreservesExistingID tests that enrichment preserves existing ID
func TestContextService_EnrichContext_PreservesExistingID(t *testing.T) {
	service := NewContextService(nil)
	existingID := "test-correlation-id-123"
	ctx := &models.CorrelationContext{
		CorrelationID: existingID,
	}

	enriched := service.EnrichContext(ctx)

	assert.Equal(t, existingID, enriched.CorrelationID, "Should preserve existing correlation ID")
}

// TestContextService_EnrichContext_AddsHostname tests that enrichment adds hostname
func TestContextService_EnrichContext_AddsHostname(t *testing.T) {
	service := NewContextService(nil)
	ctx := &models.CorrelationContext{}

	enriched := service.EnrichContext(ctx)

	assert.NotEmpty(t, enriched.Hostname, "Should add hostname")

	// Verify it's the actual system hostname
	expectedHost, _ := os.Hostname()
	assert.Equal(t, expectedHost, enriched.Hostname, "Hostname should be system hostname")
}

// TestContextService_EnrichContext_AddsEnvironment tests that enrichment adds environment
func TestContextService_EnrichContext_AddsEnvironment(t *testing.T) {
	// Setup
	originalEnv := os.Getenv("ENVIRONMENT")
	defer os.Setenv("ENVIRONMENT", originalEnv)

	t.Run("WithEnvironmentVar", func(t *testing.T) {
		os.Setenv("ENVIRONMENT", "production")
		service := NewContextService(nil)
		ctx := &models.CorrelationContext{}

		enriched := service.EnrichContext(ctx)

		assert.Equal(t, "production", enriched.Environment, "Should use ENVIRONMENT var")
	})

	t.Run("WithoutEnvironmentVar", func(t *testing.T) {
		os.Unsetenv("ENVIRONMENT")
		service := NewContextService(nil)
		ctx := &models.CorrelationContext{}

		enriched := service.EnrichContext(ctx)

		assert.Equal(t, "development", enriched.Environment, "Should default to development")
	})
}

// TestContextService_EnrichContext_AddsVersion tests that enrichment adds version
func TestContextService_EnrichContext_AddsVersion(t *testing.T) {
	originalVersion := os.Getenv("SERVICE_VERSION")
	defer os.Setenv("SERVICE_VERSION", originalVersion)

	t.Run("WithVersionVar", func(t *testing.T) {
		os.Setenv("SERVICE_VERSION", "1.2.3")
		service := NewContextService(nil)
		ctx := &models.CorrelationContext{}

		enriched := service.EnrichContext(ctx)

		assert.Equal(t, "1.2.3", enriched.Version, "Should use SERVICE_VERSION var")
	})

	t.Run("WithoutVersionVar", func(t *testing.T) {
		os.Unsetenv("SERVICE_VERSION")
		service := NewContextService(nil)
		ctx := &models.CorrelationContext{}

		enriched := service.EnrichContext(ctx)

		assert.Equal(t, "dev", enriched.Version, "Should default to dev")
	})
}

// TestContextService_EnrichContext_SetsTimestamps tests that enrichment sets timestamps
func TestContextService_EnrichContext_SetsTimestamps(t *testing.T) {
	service := NewContextService(nil)
	ctx := &models.CorrelationContext{}

	beforeEnrich := time.Now()
	enriched := service.EnrichContext(ctx)
	afterEnrich := time.Now()

	assert.False(t, enriched.CreatedAt.IsZero(), "Should set CreatedAt")
	assert.False(t, enriched.UpdatedAt.IsZero(), "Should set UpdatedAt")
	assert.True(t, enriched.CreatedAt.After(beforeEnrich.Add(-time.Second)), "CreatedAt should be approximately now")
	assert.True(t, enriched.UpdatedAt.Before(afterEnrich.Add(time.Second)), "UpdatedAt should be approximately now")
}

// TestContextService_EnrichContext_PreservesExistingTimestamps tests that enrichment preserves existing timestamps
func TestContextService_EnrichContext_PreservesExistingTimestamps(t *testing.T) {
	service := NewContextService(nil)
	oldTime := time.Now().Add(-24 * time.Hour)
	ctx := &models.CorrelationContext{
		CreatedAt: oldTime,
	}

	enriched := service.EnrichContext(ctx)

	assert.Equal(t, oldTime, enriched.CreatedAt, "Should preserve existing CreatedAt")
	assert.NotEqual(t, oldTime, enriched.UpdatedAt, "Should update UpdatedAt")
}

// TestContextService_EnrichContext_NilContext tests that enrichment handles nil context
func TestContextService_EnrichContext_NilContext(t *testing.T) {
	service := NewContextService(nil)

	enriched := service.EnrichContext(nil)

	assert.NotNil(t, enriched, "Should create context from nil")
	assert.NotEmpty(t, enriched.CorrelationID, "Should generate ID")
	assert.NotEmpty(t, enriched.Hostname, "Should add hostname")
}

// TestContextService_GetCorrelatedLogs_Valid tests retrieval of correlated logs
func TestContextService_GetCorrelatedLogs_Valid(t *testing.T) {
	// Mock repository
	mockRepo := NewMockContextRepository()
	mockRepo.On("GetCorrelatedLogs", context.Background(), "test-123", 50, 0).
		Return([]models.LogEntry{
			{ID: 1, Message: "Log 1", Level: "info"},
			{ID: 2, Message: "Log 2", Level: "error"},
		}, nil)

	service := NewContextService(mockRepo)

	logs, err := service.GetCorrelatedLogs(context.Background(), "test-123", 50, 0)

	assert.NoError(t, err, "Should not return error")
	assert.Len(t, logs, 2, "Should return 2 logs")
	assert.Equal(t, "Log 1", logs[0].Message)
	assert.Equal(t, "Log 2", logs[1].Message)
}

// TestContextService_GetCorrelatedLogs_ValidatesLimit tests limit validation
func TestContextService_GetCorrelatedLogs_ValidatesLimit(t *testing.T) {
	mockRepo := NewMockContextRepository()
	mockRepo.On("GetCorrelatedLogs", context.Background(), "test-123", 1000, 0).
		Return([]models.LogEntry{}, nil)

	service := NewContextService(mockRepo)

	// Limit too high (> 1000) should be capped at 1000
	_, _ = service.GetCorrelatedLogs(context.Background(), "test-123", 5000, 0)

	// Verify repository was called with capped limit (1000 max)
	mockRepo.AssertCalled(t, "GetCorrelatedLogs", context.Background(), "test-123", 1000, 0)
}

// TestContextService_GetCorrelatedLogs_DefaultLimit tests default limit
func TestContextService_GetCorrelatedLogs_DefaultLimit(t *testing.T) {
	mockRepo := NewMockContextRepository()
	mockRepo.On("GetCorrelatedLogs", context.Background(), "test-123", 50, 0).
		Return([]models.LogEntry{}, nil)

	service := NewContextService(mockRepo)

	// Default limit (0)
	_, _ = service.GetCorrelatedLogs(context.Background(), "test-123", 0, 0)

	// Verify repository was called with default limit
	mockRepo.AssertCalled(t, "GetCorrelatedLogs", context.Background(), "test-123", 50, 0)
}

// TestContextService_GetCorrelationMetadata_Valid tests metadata retrieval
func TestContextService_GetCorrelationMetadata_Valid(t *testing.T) {
	mockRepo := NewMockContextRepository()
	mockRepo.On("GetCorrelationCount", context.Background(), "test-123").
		Return(5, nil)
	mockRepo.On("GetContextMetadata", context.Background(), "test-123").
		Return(map[string]interface{}{
			"service":  "portal",
			"trace_id": "trace-abc123",
		}, nil)

	service := NewContextService(mockRepo)

	metadata, err := service.GetCorrelationMetadata(context.Background(), "test-123")

	assert.NoError(t, err, "Should not return error")
	assert.Equal(t, 5, metadata["total_logs"], "Should include log count")
	assert.Equal(t, "test-123", metadata["correlation_id"], "Should include correlation ID")
	assert.Equal(t, "portal", metadata["service"], "Should include service")
}

// TestContextService_GetTraceTimeline_Valid tests timeline retrieval
func TestContextService_GetTraceTimeline_Valid(t *testing.T) {
	now := time.Now()
	mockRepo := NewMockContextRepository()
	mockRepo.On("GetCorrelatedLogs", context.Background(), "test-123", 1000, 0).
		Return([]models.LogEntry{
			{
				ID:        1,
				Message:   "Step 1",
				Level:     "info",
				Timestamp: now,
				Context: &models.CorrelationContext{
					TraceID: "trace-123",
					SpanID:  "span-1",
				},
			},
			{
				ID:        2,
				Message:   "Step 2",
				Level:     "info",
				Timestamp: now.Add(time.Second),
				Context: &models.CorrelationContext{
					TraceID: "trace-123",
					SpanID:  "span-2",
				},
			},
		}, nil)

	service := NewContextService(mockRepo)

	timeline, err := service.GetTraceTimeline(context.Background(), "test-123")

	assert.NoError(t, err, "Should not return error")
	assert.Len(t, timeline, 2, "Should return 2 events")
	assert.Equal(t, "Step 1", timeline[0]["message"])
	assert.Equal(t, "trace-123", timeline[0]["trace_id"])
	assert.Equal(t, "span-1", timeline[0]["span_id"])
}

// TestContextService_EnrichContext_AllFields tests complete enrichment
func TestContextService_EnrichContext_AllFields(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("SERVICE_VERSION", "1.0.0")
	defer os.Unsetenv("ENVIRONMENT")
	defer os.Unsetenv("SERVICE_VERSION")

	service := NewContextService(nil)
	ctx := &models.CorrelationContext{
		RequestID: "req-123",
		Method:    "POST",
		Path:      "/api/test",
	}

	enriched := service.EnrichContext(ctx)

	assert.NotEmpty(t, enriched.CorrelationID, "CorrelationID should be set")
	assert.NotEmpty(t, enriched.Hostname, "Hostname should be set")
	assert.Equal(t, "test", enriched.Environment, "Environment should be set")
	assert.Equal(t, "1.0.0", enriched.Version, "Version should be set")
	assert.Equal(t, "req-123", enriched.RequestID, "RequestID should be preserved")
	assert.Equal(t, "POST", enriched.Method, "Method should be preserved")
	assert.Equal(t, "/api/test", enriched.Path, "Path should be preserved")
	assert.False(t, enriched.CreatedAt.IsZero(), "CreatedAt should be set")
}
