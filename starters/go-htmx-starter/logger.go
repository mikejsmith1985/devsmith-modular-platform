package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
)

// LogLevel represents logging severity
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

// Logger provides structured JSON logging with correlation IDs
type Logger struct {
	serviceName string
	minLevel    LogLevel
}

// NewLogger creates a new logger instance
func NewLogger(serviceName, level string) *Logger {
	logLevel := INFO
	switch level {
	case "DEBUG":
		logLevel = DEBUG
	case "INFO":
		logLevel = INFO
	case "WARN":
		logLevel = WARN
	case "ERROR":
		logLevel = ERROR
	}

	return &Logger{
		serviceName: serviceName,
		minLevel:    logLevel,
	}
}

// Debug logs a debug-level message
func (l *Logger) Debug(ctx context.Context, event string, metadata map[string]interface{}) {
	l.log(ctx, DEBUG, event, metadata)
}

// Info logs an info-level message
func (l *Logger) Info(ctx context.Context, event string, metadata map[string]interface{}) {
	l.log(ctx, INFO, event, metadata)
}

// Warn logs a warning-level message
func (l *Logger) Warn(ctx context.Context, event string, metadata map[string]interface{}) {
	l.log(ctx, WARN, event, metadata)
}

// Error logs an error-level message
func (l *Logger) Error(ctx context.Context, event string, metadata map[string]interface{}) {
	l.log(ctx, ERROR, event, metadata)
}

// log writes a structured JSON log entry
func (l *Logger) log(ctx context.Context, level LogLevel, event string, metadata map[string]interface{}) {
	// Check if we should log this level
	if !l.shouldLog(level) {
		return
	}

	// Extract correlation ID from context
	correlationID := getCorrelationID(ctx)

	// Build log entry
	entry := map[string]interface{}{
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"service":        l.serviceName,
		"level":          level,
		"event":          event,
		"correlation_id": correlationID,
	}

	// Merge metadata
	if metadata != nil {
		for k, v := range metadata {
			entry[k] = v
		}
	}

	// Output as JSON
	output, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	// Write to stdout (captured by Docker logs)
	os.Stdout.Write(output)
	os.Stdout.Write([]byte("\n"))
}

// shouldLog determines if a message should be logged based on level
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
	}

	return levels[level] >= levels[l.minLevel]
}

// Context key for correlation ID
type contextKey string

const correlationIDKey contextKey = "correlation_id"

// getCorrelationID extracts correlation ID from context
func getCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}

	return ""
}

// generateCorrelationID creates a new correlation ID
func generateCorrelationID() string {
	return uuid.New().String()
}
