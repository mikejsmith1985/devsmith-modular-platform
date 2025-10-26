// Package logger provides a structured logging SDK for services in the DevSmith platform.
//
// This package implements an asynchronous, batched structured logging client that sends
// logs to a centralized Logs service. It supports automatic context injection, multiple
// log levels, and falls back to stdout if the service is unavailable.
//
// Core Features:
//   - Async structured logging with automatic batching (by count or timeout)
//   - All log levels: DEBUG, INFO, WARN, ERROR, FATAL, PANIC
//   - Automatic injection of service name and timestamp
//   - Structured key-value fields (any type supported)
//   - Context-aware logging (correlation ID, user ID, request ID extraction)
//   - Non-blocking async HTTP sends to Logs service
//   - Thread-safe concurrent logging
//   - Graceful flush and shutdown
//   - Method chaining support (WithContext, WithFields)
//   - Global logger singleton pattern
//   - Automatic fallback to stdout on service failure
//
// Quick Start:
//
//	config := &logger.Config{
//	    ServiceName:     "my-service",
//	    LogLevel:        "info",
//	    LogURL:          "http://logs-service:8082/api/logs",
//	    BatchSize:       100,
//	    BatchTimeoutSec: 5,
//	    LogToStdout:     true,
//	}
//
//	log, err := logger.NewLogger(config)
//	if err != nil {
//	    panic(err)
//	}
//	defer log.Close()
//
//	// Simple logging
//	log.Info("Operation completed", "duration_ms", 1234, "success", true)
//
//	// With context
//	ctx := context.WithValue(context.Background(), logger.CorrelationIDKey, "req-123")
//	log.WithContext(ctx).Error("Request failed", "error_code", "TIMEOUT")
//
// Context Keys:
//   - CorrelationIDKey: Unique identifier for tracing requests across services
//   - UserIDKey: Identifier of the user making the request
//   - RequestIDKey: Unique identifier for the individual request
//
// Performance Considerations:
//   - Logs are batched and sent asynchronously in a background goroutine
//   - Batches are sent when either: (1) batch size is reached, or (2) timeout expires
//   - Logger.Close() flushes all pending logs before returning
//   - All operations are non-blocking and thread-safe
//
// See EXAMPLES.md for detailed usage examples and best practices.
package logger

import (
	"context"
)

// Interface defines the contract for logging operations.
// All methods are safe for concurrent use and non-blocking.
type Interface interface {
	// Info logs an info level message with optional structured fields.
	// Fields should be provided as alternating key-value pairs.
	// Example: logger.Info("user created", "user_id", 123, "email", "user@example.com")
	Info(msg string, keyvals ...interface{})

	// Debug logs a debug level message with optional structured fields.
	// Messages at DEBUG level are only logged if LogLevel is set to "debug".
	Debug(msg string, keyvals ...interface{})

	// Warn logs a warning level message with optional structured fields.
	// Use for potentially problematic situations that don't prevent operation.
	Warn(msg string, keyvals ...interface{})

	// Error logs an error level message with optional structured fields.
	// Use for error conditions that prevented an operation but service continues.
	Error(msg string, keyvals ...interface{})

	// Fatal logs a fatal level message with optional structured fields and exits.
	// This will call os.Exit(1) after logging. Use only for unrecoverable errors.
	Fatal(msg string, keyvals ...interface{})

	// Panic logs a panic level message with optional structured fields.
	// This will call panic() after logging. Use for critical invariant violations.
	Panic(msg string, keyvals ...interface{})

	// WithContext returns a new logger instance that extracts values from the context.
	// Automatically extracts CorrelationIDKey, UserIDKey, and RequestIDKey from context.
	// All logs from this logger will include the extracted context values.
	// Example:
	//   ctx := context.WithValue(context.Background(), logger.CorrelationIDKey, "req-123")
	//   logger.WithContext(ctx).Info("message") // Will include correlation_id in metadata
	WithContext(ctx context.Context) Interface

	// WithFields returns a new logger instance with additional structured fields.
	// Fields should be provided as alternating key-value pairs.
	// Can be chained with WithContext for combined context and field logging.
	// Example:
	//   logger.WithFields("user_id", 123, "action", "create").
	//       Info("user created")
	WithFields(keyvals ...interface{}) Interface

	// Flush ensures all pending logs are sent to the service synchronously.
	// Blocks until all buffered logs are sent or context is cancelled.
	// Returns an error if the flush operation fails.
	// Should be called before graceful shutdown to ensure no logs are lost.
	Flush(ctx context.Context) error

	// Close gracefully shuts down the logger, flushing all pending logs.
	// Should be called using defer to ensure proper cleanup.
	// Returns an error if shutdown fails.
	// After Close(), the logger should not be used.
	Close() error
}

// contextKeyType is used for context keys to avoid collisions with other packages.
// Context keys in Go should be custom types, not strings, to prevent collisions.
type contextKeyType string

const (
	// CorrelationIDKey is the key for storing correlation ID in context.
	// Used to trace requests across multiple services.
	// Example: ctx = context.WithValue(ctx, logger.CorrelationIDKey, "req-abc123")
	CorrelationIDKey contextKeyType = "correlation_id"

	// UserIDKey is the key for storing user ID in context.
	// Used to identify which user made the request.
	// Example: ctx = context.WithValue(ctx, logger.UserIDKey, "user-456")
	UserIDKey contextKeyType = "user_id"

	// RequestIDKey is the key for storing request ID in context.
	// Used to uniquely identify individual requests within a service.
	// Example: ctx = context.WithValue(ctx, logger.RequestIDKey, "request-xyz")
	RequestIDKey contextKeyType = "request_id"
)
