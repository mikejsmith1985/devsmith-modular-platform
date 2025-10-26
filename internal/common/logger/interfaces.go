package logger

import (
	"context"
)

// Interface defines the contract for logging operations.
type Interface interface {
	// Info logs an info level message with optional structured fields.
	Info(msg string, keyvals ...interface{})

	// Debug logs a debug level message with optional structured fields.
	Debug(msg string, keyvals ...interface{})

	// Warn logs a warning level message with optional structured fields.
	Warn(msg string, keyvals ...interface{})

	// Error logs an error level message with optional structured fields.
	Error(msg string, keyvals ...interface{})

	// Fatal logs a fatal level message with optional structured fields and exits.
	Fatal(msg string, keyvals ...interface{})

	// Panic logs a panic level message with optional structured fields.
	Panic(msg string, keyvals ...interface{})

	// WithContext returns a logger with context-extracted values.
	WithContext(ctx context.Context) Interface

	// WithFields returns a logger with additional structured fields.
	WithFields(keyvals ...interface{}) Interface

	// Flush ensures all pending logs are sent.
	Flush(ctx context.Context) error

	// Close gracefully shuts down the logger.
	Close() error
}

// contextKeyType is used for context keys to avoid collisions.
type contextKeyType string

const (
	// CorrelationIDKey is the key for storing correlation ID in context.
	CorrelationIDKey contextKeyType = "correlation_id"

	// UserIDKey is the key for storing user ID in context.
	UserIDKey contextKeyType = "user_id"

	// RequestIDKey is the key for storing request ID in context.
	RequestIDKey contextKeyType = "request_id"
)
