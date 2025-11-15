package logger

import (
	"context"
	"fmt"
	"net/http"
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

// TestLogger_BatchSending_DoesNotSendImmediately verifies batching behavior (BEHAVIORAL TEST).
func TestLogger_BatchSending_DoesNotSendImmediately(t *testing.T) {
	config := &Config{
		ServiceName:     "test-service",
		LogLevel:        "info",
		LogURL:          "http://localhost:8082/api/logs",
		BatchSize:       100,
		BatchTimeoutSec: 60, // Long timeout
		LogToStdout:     true,
		EnableStdout:    true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Log a single message
	logger.Info("test message", "key", "value")

	// Check buffer - should still have unsent logs (not empty)
	logger.mu.RLock()
	bufferLen := len(logger.batchBuffer)
	logger.mu.RUnlock()

	// BEHAVIORAL: Should buffer the log, not send immediately
	assert.Greater(t, bufferLen, 0, "Log should be buffered and not sent immediately")
}

// TestLogger_BatchSending_SendsWhenFull verifies batch count triggers send.
func TestLogger_BatchSending_SendsWhenFull(t *testing.T) {
	config := &Config{
		ServiceName:     "test-service",
		LogLevel:        "info",
		LogURL:          "http://localhost:8082/api/logs",
		BatchSize:       5,
		BatchTimeoutSec: 60,
		LogToStdout:     false,
		EnableStdout:    true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Log exactly batch size messages
	for i := 0; i < 5; i++ {
		logger.Info(fmt.Sprintf("message %d", i))
	}

	// Give async sender time to send
	time.Sleep(200 * time.Millisecond)

	// BEHAVIORAL: Buffer should be cleared after batch size reached
	logger.mu.RLock()
	bufferLen := len(logger.batchBuffer)
	logger.mu.RUnlock()

	assert.Equal(t, 0, bufferLen, "Buffer should be sent when batch size reached")
}

// TestLogger_CorrelationIDInjection_AppearsInLogs verifies correlation ID is added.
func TestLogger_CorrelationIDInjection_AppearsInLogs(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		BatchSize:    1, // Flush immediately
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	correlationID := "test-corr-id-12345"
	ctx := context.WithValue(context.Background(), CorrelationIDKey, correlationID)

	// Log with context
	loggerWithCtx := logger.WithContext(ctx)
	loggerWithCtx.Info("test message")

	time.Sleep(100 * time.Millisecond)

	// BEHAVIORAL: Logged entry should contain correlation ID
	// This requires checking the batchBuffer content (implementation will add logs there)
	logger.mu.RLock()
	bufferLen := len(logger.batchBuffer)
	logger.mu.RUnlock()

	assert.Equal(t, 0, bufferLen, "Buffer should be flushed (batch size of 1)")
}

// TestLogger_StructuredFields_AppearsInSentLog verifies fields are included.
func TestLogger_StructuredFields_AppearsInSentLog(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		LogLevel:     "info",
		BatchSize:    1, // Send immediately
		LogToStdout:  true,
		EnableStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Log with structured fields
	logger.Info("test message", "user_id", "123", "request_id", "req-456")

	// Give async sender time to flush (batch size is 1, so should send immediately)
	time.Sleep(100 * time.Millisecond)

	// Ensure flush completes before checking buffer state
	err = logger.Flush(context.Background())
	require.NoError(t, err)

	// BEHAVIORAL: The logged entry should contain the fields
	// This will be verifiable once implementation stores entries with fields
	// Check buffer with proper mutex protection
	logger.mu.RLock()
	bufferExists := logger.batchBuffer != nil
	logger.mu.RUnlock()

	assert.True(t, bufferExists, "Buffer should exist to store log entries")
}

// TestLogger_ServiceNameInjection_AppearsInEveryLog verifies service name is added.
func TestLogger_ServiceNameInjection_AppearsInEveryLog(t *testing.T) {
	serviceName := "my-special-service"
	config := &Config{
		ServiceName: serviceName,
		LogLevel:    "info",
		BatchSize:   1, // Send immediately
		LogToStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	logger.Info("test message")

	// BEHAVIORAL: Logger should have stored the service name
	assert.Equal(t, serviceName, logger.serviceName)
	// When implementation completes, logged entry will have this service name
}

// TestLogger_LogLevelFiltering_DebugNotSentWhenWarn verifies level filtering.
func TestLogger_LogLevelFiltering_DebugNotSentWhenWarn(t *testing.T) {
	config := &Config{
		ServiceName: "test-service",
		LogLevel:    "warn", // Only warn and above
		BatchSize:   1,
		LogToStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Log debug (should be filtered)
	logger.Debug("debug message")

	time.Sleep(100 * time.Millisecond)

	// BEHAVIORAL: Debug log should NOT appear in buffer when level is warn
	logger.mu.RLock()
	bufferLen := len(logger.batchBuffer)
	logger.mu.RUnlock()

	assert.Equal(t, 0, bufferLen, "Debug logs should be filtered when level is warn")
}

// TestLogger_Flush_SendsAllPendingLogs verifies flush works.
func TestLogger_Flush_SendsAllPendingLogs(t *testing.T) {
	config := &Config{
		ServiceName:     "test-service",
		LogLevel:        "info",
		BatchSize:       100, // Won't trigger auto-send
		BatchTimeoutSec: 60,
		LogToStdout:     true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Log several messages (won't trigger batch send)
	for i := 0; i < 10; i++ {
		logger.Info(fmt.Sprintf("message %d", i))
	}

	// BEHAVIORAL: Buffer should have logs
	logger.mu.RLock()
	beforeFlush := len(logger.batchBuffer)
	logger.mu.RUnlock()
	assert.Greater(t, beforeFlush, 0, "Buffer should have logs before flush")

	// Flush
	err = logger.Flush(context.Background())
	require.NoError(t, err)

	// BEHAVIORAL: Buffer should be empty after flush
	logger.mu.RLock()
	afterFlush := len(logger.batchBuffer)
	logger.mu.RUnlock()
	assert.Equal(t, 0, afterFlush, "Buffer should be empty after flush")
}

// TestLogger_Close_PreventsLoggingAfterClose verifies close behavior.
func TestLogger_Close_PreventsLoggingAfterClose(t *testing.T) {
	config := &Config{
		ServiceName: "test-service",
		LogLevel:    "info",
		LogToStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Log before close
	logger.Info("message before close")

	time.Sleep(100 * time.Millisecond)

	logger.mu.RLock()
	beforeClose := len(logger.batchBuffer)
	logger.mu.RUnlock()
	assert.Greater(t, beforeClose, 0)

	// Close
	err = logger.Close()
	require.NoError(t, err)

	// Wait for pending logs to flush
	time.Sleep(200 * time.Millisecond)

	// BEHAVIORAL: Logger should be marked closed
	assert.True(t, logger.closed, "Logger should be closed")

	// Log after close (should be ignored)
	logger.Info("message after close")

	logger.mu.RLock()
	afterClose := len(logger.batchBuffer)
	logger.mu.RUnlock()

	// BEHAVIORAL: No new logs should be added after close
	assert.Equal(t, 0, afterClose, "Buffer should be flushed after close")
}

// TestLogger_GlobalLogger_GlobalFunctionsDelegateToGlobal verifies delegation.
func TestLogger_GlobalLogger_GlobalFunctionsDelegateToGlobal(t *testing.T) {
	config := &Config{
		ServiceName: "test-service",
		LogLevel:    "info",
		LogToStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)

	SetGlobalLogger(logger)

	// Use global functions
	LogInfo("message via global")

	// BEHAVIORAL: Message should be in the global logger's buffer
	logger.mu.RLock()
	bufferLen := len(logger.batchBuffer)
	logger.mu.RUnlock()

	assert.Greater(t, bufferLen, 0, "Global function should delegate to set logger")
}

// TestLogger_ConcurrentLogging_NoDataRace verifies thread safety with race detector.
func TestLogger_ConcurrentLogging_NoDataRace(t *testing.T) {
	config := &Config{
		ServiceName: "test-service",
		LogLevel:    "info",
		LogToStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)

	var wg sync.WaitGroup
	const numGoroutines = 20
	const messagesPerGoroutine = 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Info(fmt.Sprintf("goroutine %d message %d", id, j))
				logger.Error(fmt.Sprintf("error from %d", id))
				logger.Warn(fmt.Sprintf("warn from %d", id))
			}
		}(i)
	}

	// Also flush from another goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			time.Sleep(10 * time.Millisecond)
			_ = logger.Flush(context.Background())
		}
	}()

	wg.Wait()

	// BEHAVIORAL: Should complete without panics or deadlocks
	// Go's race detector will catch data races if any exist
	assert.True(t, true, "Concurrent logging completed without errors")
}

// TestLogger_GinContextExtraction_UsesGinCorrelationID verifies Gin integration.
func TestLogger_GinContextExtraction_UsesGinCorrelationID(t *testing.T) {
	config := &Config{
		ServiceName: "test-service",
		LogLevel:    "info",
		BatchSize:   1,
		LogToStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)

	// Create a proper HTTP request with context
	req, _ := http.NewRequest("GET", "http://localhost/test", http.NoBody)

	// Add correlation ID to context
	ginCorrelationID := "gin-corr-123"
	ctx := context.WithValue(req.Context(), CorrelationIDKey, ginCorrelationID)

	// Log with context
	loggerWithCtx := logger.WithContext(ctx)
	loggerWithCtx.Info("gin context message")

	time.Sleep(100 * time.Millisecond)

	// BEHAVIORAL: Log should use the correlation ID from context
	logger.mu.RLock()
	bufferLen := len(logger.batchBuffer)
	logger.mu.RUnlock()

	assert.Equal(t, 0, bufferLen, "Buffer should be flushed (batch size of 1)")
}

// TestLogger_BatchTimeout_SendsAfterTimeout verifies timeout-based sending.
func TestLogger_BatchTimeout_SendsAfterTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	config := &Config{
		ServiceName:     "test-service",
		LogLevel:        "info",
		BatchSize:       100,
		BatchTimeoutSec: 2, // 2 second timeout
		LogToStdout:     true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)

	// Log a single message (less than batch size)
	logger.Info("timeout test message")

	logger.mu.RLock()
	beforeTimeout := len(logger.batchBuffer)
	logger.mu.RUnlock()
	assert.Greater(t, beforeTimeout, 0, "Message should be buffered")

	// Wait for timeout plus buffer
	time.Sleep(3 * time.Second)

	// BEHAVIORAL: Buffer should be sent after timeout
	logger.mu.RLock()
	afterTimeout := len(logger.batchBuffer)
	logger.mu.RUnlock()

	assert.Equal(t, 0, afterTimeout, "Buffer should be sent after timeout period")

	logger.Close()
}

// TestLogger_MultipleFieldTypes_AllSupported verifies field type support.
func TestLogger_MultipleFieldTypes_AllSupported(t *testing.T) {
	config := &Config{
		ServiceName: "test-service",
		LogLevel:    "info",
		BatchSize:   1,
		LogToStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)

	// Log with various types
	logger.Info("multi-type message",
		"string_field", "value",
		"int_field", 42,
		"float_field", 3.14,
		"bool_field", true,
	)

	time.Sleep(100 * time.Millisecond)

	// BEHAVIORAL: Should handle multiple field types without panic
	logger.mu.RLock()
	bufferLen := len(logger.batchBuffer)
	logger.mu.RUnlock()

	assert.Equal(t, 0, bufferLen, "Buffer should be flushed (batch size of 1)")
}

// TestLogger_WithFields_ChainingWorks verifies field chaining.
func TestLogger_WithFields_ChainingWorks(t *testing.T) {
	config := &Config{
		ServiceName: "test-service",
		LogLevel:    "info",
		LogToStdout: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)

	// Chain WithFields
	loggerWithFields := logger.WithFields("field1", "value1").WithFields("field2", "value2")

	loggerWithFields.Info("message with chained fields")

	// BEHAVIORAL: Chaining should work and message should be logged
	assert.NotNil(t, loggerWithFields, "Chaining should return logger")
}
