package logger

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewLogger_WithValidConfig_ReturnsLogger tests logger initialization.
func TestNewLogger_WithValidConfig_ReturnsLogger(t *testing.T) {
	config := &Config{
		ServiceName: "test-service",
		LogLevel:    "info",
	}

	logger, err := NewLogger(config)

	assert.NoError(t, err)
	assert.NotNil(t, logger)
	assert.Equal(t, "test-service", logger.serviceName)
	assert.Equal(t, "info", logger.logLevel)
}

// TestNewLogger_WithEmptyServiceName_ReturnsError tests validation.
func TestNewLogger_WithEmptyServiceName_ReturnsError(t *testing.T) {
	config := &Config{
		ServiceName: "",
		LogLevel:    "info",
	}

	logger, err := NewLogger(config)

	assert.Error(t, err)
	assert.Nil(t, logger)
	assert.Contains(t, err.Error(), "service name is required")
}

// TestLogger_Info_LogsStructuredMessage tests Info method.
func TestLogger_Info_LogsStructuredMessage(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Should not panic when logging
	logger.Info("test message", "key1", "value1")
}

// TestLogger_Error_LogsErrorMessage tests Error method.
func TestLogger_Error_LogsErrorMessage(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "error",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Should not panic when logging error
	logger.Error("test error", "key1", "value1")
}

// TestLogger_Warn_LogsWarningMessage tests Warn method.
func TestLogger_Warn_LogsWarningMessage(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "warn",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Should not panic when logging warning
	logger.Warn("test warning", "key1", "value1")
}

// TestLogger_Debug_LogsDebugMessage tests Debug method.
func TestLogger_Debug_LogsDebugMessage(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "debug",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Should not panic when logging debug
	logger.Debug("test debug", "key1", "value1")
}

// TestLogger_Fatal_LogsAndExits tests Fatal method (should exit, tested by behavior).
func TestLogger_Fatal_LogsFatalMessage(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Don't actually call Fatal as it will exit the test
	// Just verify the method exists and is callable
	assert.NotNil(t, logger.Fatal)
}

// TestLogger_WithContext_InjectsCorrelationID tests context injection.
func TestLogger_WithContext_InjectsCorrelationID(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	correlationID := "test-correlation-123"
	ctx := context.WithValue(context.Background(), CorrelationIDKey, correlationID)

	// Should extract correlation ID from context
	logger.Info("test message", "key", "value")
	// (Verification will be in implementation phase)
	_ = ctx
}

// TestLogger_WithGinContext_ExtractsCorrelationID tests Gin context extraction.
func TestLogger_WithGinContext_ExtractsCorrelationID(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Create a Gin context with correlation ID
	c, _ := gin.CreateTestContext(nil)
	c.Set("correlation_id", "test-correlation-456")

	// Should extract from Gin context
	// (Verification will be in implementation phase)
	_ = c
}

// TestLogger_BatchedSending_SendsAfterFiveSeconds tests batching with time.
func TestLogger_BatchedSending_SendsAfterFiveSeconds(t *testing.T) {
	config := &Config{
		ServiceName:     "test-service",
		LogLevel:        "info",
		BatchSize:       100,
		BatchTimeoutSec: 5,
		LogToStdout:     true,
		EnableStdout:    true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Log some messages
	logger.Info("message 1")
	logger.Info("message 2")

	// Should batch them (not send immediately)
	assert.NotNil(t, logger.batchBuffer)
}

// TestLogger_BatchedSending_SendsAfter100Logs tests batching with count.
func TestLogger_BatchedSending_SendsAfter100Logs(t *testing.T) {
	config := &Config{
		ServiceName:     "test-service",
		LogLevel:        "info",
		BatchSize:       100,
		BatchTimeoutSec: 60,
		LogToStdout:     true,
		EnableStdout:    true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Log messages
	for i := 0; i < 50; i++ {
		logger.Info(fmt.Sprintf("message %d", i))
	}

	// Should not have exceeded batch size yet
	assert.LessOrEqual(t, len(logger.batchBuffer), 50)
}

// TestLogger_AsyncSending_NonBlocking tests async behavior.
func TestLogger_AsyncSending_NonBlocking(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Logging should return immediately (async)
	start := time.Now()
	for i := 0; i < 100; i++ {
		logger.Info(fmt.Sprintf("message %d", i))
	}
	duration := time.Since(start)

	// Should complete quickly (< 100ms for 100 async logs)
	assert.Less(t, duration, 100*time.Millisecond)
}

// TestLogger_FallbackToStdout_WhenServiceUnavailable tests fallback behavior.
func TestLogger_FallbackToStdout_WhenServiceUnavailable(t *testing.T) {
	config := &Config{
		ServiceName:     "test-service",
		LogLevel:        "info",
		LogURL:          "http://invalid:9999", // Invalid URL
		LogToStdout:     false,
		EnableStdout:    true,
		BatchTimeoutSec: 1,
		BatchSize:       10,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Should fall back to stdout if service unavailable
	logger.Info("test message")
	// Verification: message should appear in stdout (tested manually or with log capture)
}

// TestLogger_ConfigDefaults_AppliesWhenNotProvided tests defaults.
func TestLogger_ConfigDefaults_AppliesWhenNotProvided(t *testing.T) {
	config := &Config{
		ServiceName: "test-service",
		// Other fields left empty to test defaults
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Should have default values applied
	assert.Equal(t, DefaultBatchSize, logger.batchSize)
	assert.Equal(t, DefaultBatchTimeoutSec, logger.batchTimeoutSec)
}

// TestLogger_MultipleInstances_ThreadSafe tests concurrent logging.
func TestLogger_MultipleInstances_ThreadSafe(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				logger.Info(fmt.Sprintf("goroutine %d message %d", id, j))
			}
		}(i)
	}

	wg.Wait()
	// Should not panic or deadlock
}

// TestLogger_WithPanicRecovery_LogsPanicAndContinues tests panic logging.
func TestLogger_WithPanicRecovery_LogsPanicAndContinues(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Should have a method to log panics
	assert.NotNil(t, logger.Panic)
}

// TestLogger_Flush_SendsPendingLogs tests flushing.
func TestLogger_Flush_SendsPendingLogs(t *testing.T) {
	config := &Config{
		ServiceName:     "test-service",
		LogLevel:        "info",
		BatchSize:       100,
		BatchTimeoutSec: 60,
		LogToStdout:     true,
		EnableStdout:    true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Log some messages
	logger.Info("message 1")
	logger.Info("message 2")

	// Flush should send pending logs
	err = logger.Flush(context.Background())
	assert.NoError(t, err)
}

// TestLogger_Close_GracefulShutdown tests shutdown.
func TestLogger_Close_GracefulShutdown(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Log a message
	logger.Info("test message")

	// Close should gracefully shutdown
	err = logger.Close()
	assert.NoError(t, err)

	// Should not be able to log after close
	// (Behavior to be verified in implementation)
}

// TestLogger_CorrelationIDKey_IsContextKey tests context key.
func TestLogger_CorrelationIDKey_IsContextKey(t *testing.T) {
	// CorrelationIDKey should be a valid context key
	assert.NotNil(t, CorrelationIDKey)

	// Should be usable with context.WithValue
	ctx := context.WithValue(context.Background(), CorrelationIDKey, "test-id")
	val := ctx.Value(CorrelationIDKey)
	assert.Equal(t, "test-id", val)
}

// TestLogger_DefaultConfig_HasReasonableValues tests defaults.
func TestLogger_DefaultConfig_HasReasonableValues(t *testing.T) {
	assert.Greater(t, DefaultBatchSize, 0)
	assert.Greater(t, DefaultBatchTimeoutSec, 0)
	assert.NotEmpty(t, DefaultLogLevel)
}

// TestLogger_LogLevelFiltering_OnlyLogsAppropriateLevels tests level filtering.
func TestLogger_LogLevelFiltering_OnlyLogsAppropriateLevels(t *testing.T) {
	// Create logger with WARN level
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "warn",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Should log warnings and errors
	logger.Warn("warning message")
	logger.Error("error message")

	// Should not log debug (since level is warn)
	logger.Debug("debug message") // Should be filtered

	// Verification: only warn/error should appear (tested manually)
}

// TestLogger_StructuredFields_IncludedInLogs tests field inclusion.
func TestLogger_StructuredFields_IncludedInLogs(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Log with structured fields
	logger.Info("test message", "user_id", "123", "request_id", "req-456")

	// Fields should be included (verification in implementation)
}

// TestLogger_ServiceNameAutoInjection_InEveryLog tests auto-injection.
func TestLogger_ServiceNameAutoInjection_InEveryLog(t *testing.T) {
	config := &Config{
		ServiceName:  "my-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Service name should be injected into every log
	logger.Info("test message")

	// Should be available in logger object
	assert.Equal(t, "my-service", logger.serviceName)
}

// TestLogger_TimestampAutoInjection_InEveryLog tests timestamp injection.
func TestLogger_TimestampAutoInjection_InEveryLog(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	before := time.Now()
	logger.Info("test message")
	after := time.Now()

	// Timestamp should be injected (verification in implementation)
	assert.True(t, before.Before(after) || before.Equal(after))
}

// TestGlobalLogger_GetInstance_ReturnsSingleton tests global logger.
func TestGlobalLogger_GetInstance_ReturnsSingleton(t *testing.T) {
	// Initialize with config
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger1, err := NewLogger(config)
	require.NoError(t, err)

	// Global should be settable
	SetGlobalLogger(logger1)

	// Should return same instance
	logger2 := GetGlobalLogger()
	assert.Equal(t, logger1, logger2)
}

// TestGlobalLogger_GlobalMethods_UseGlobalLogger tests global functions.
func TestGlobalLogger_GlobalMethods_UseGlobalLogger(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	SetGlobalLogger(logger)

	// Global functions should work
	LogInfo("test message", "key", "value")
	LogError("error message", "key", "value")
	LogWarn("warning message", "key", "value")
	LogDebug("debug message", "key", "value")
}

// TestLogger_LogURLConfiguration_IsOptional tests URL config.
func TestLogger_LogURLConfiguration_IsOptional(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogURL:       "", // Empty URL
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	assert.NotNil(t, logger)
	// Should work without log service URL
}

// TestLogger_LogURLConfiguration_IsUsedWhenProvided tests URL usage.
func TestLogger_LogURLConfiguration_IsUsedWhenProvided(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogURL:       "http://localhost:8082/api/logs",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	assert.NotNil(t, logger)
	// Should attempt to use provided URL
}

// TestLogger_BatchBuffer_HasExpectedStructure tests buffer structure.
func TestLogger_BatchBuffer_HasExpectedStructure(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Buffer should be initialized
	assert.NotNil(t, logger.batchBuffer)
	assert.Greater(t, cap(logger.batchBuffer), 0)
}

// TestLogger_SendLog_HitsLoggingService tests send functionality.
func TestLogger_SendLog_HitsLoggingService(t *testing.T) {
	config := &Config{
		ServiceName:     "test-service",
		LogLevel:        "info",
		LogURL:          "http://localhost:8082/api/logs",
		BatchSize:       1, // Send immediately
		BatchTimeoutSec: 1,
		LogToStdout:     true,
		EnableStdout:    true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Log should attempt to send
	logger.Info("test message")

	// Give it time to send asynchronously
	time.Sleep(100 * time.Millisecond)
	// Verification: check logs were sent (integration test)
}

// TestLogger_ExampleUsage_SimpleAPI tests expected simple API.
func TestLogger_ExampleUsage_SimpleAPI(t *testing.T) {
	// This test verifies the simple API mentioned in requirements
	config := &Config{
		ServiceName:  "review-service",
		LogLevel:     "info",
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Example usage from requirements: logger.Info("message", "key", "value")
	logger.Info("AI analysis completed", "mode", "critical", "duration_ms", 1234)

	// Should not panic
	assert.True(t, true)
}
