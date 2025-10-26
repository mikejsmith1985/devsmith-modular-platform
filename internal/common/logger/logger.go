// Package logger provides a structured logging SDK for services.
package logger

import (
	"context"
	"fmt"
	"sync"
)

// Logger is a structured logger client for sending logs to the logging service.
// nolint:govet // Struct alignment optimized for readability and logical grouping
type Logger struct {
	// mutex protects concurrent access.
	mu sync.RWMutex

	serviceName     string
	logLevel        string
	logURL          string
	batchSize       int
	batchTimeoutSec int
	logToStdout     bool
	enableStdout    bool
	closed          bool

	// batchBuffer holds logs pending to be sent.
	batchBuffer []interface{}

	// done signals goroutines to stop.
	done chan struct{}

	// wg waits for goroutines to finish.
	wg sync.WaitGroup
}

// NewLogger creates a new structured logger instance.
func NewLogger(config *Config) (*Logger, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	if config.ServiceName == "" {
		return nil, fmt.Errorf("service name is required")
	}

	logLevel := config.LogLevel
	if logLevel == "" {
		logLevel = DefaultLogLevel
	}

	batchSize := config.BatchSize
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	batchTimeoutSec := config.BatchTimeoutSec
	if batchTimeoutSec <= 0 {
		batchTimeoutSec = DefaultBatchTimeoutSec
	}

	logger := &Logger{
		serviceName:     config.ServiceName,
		logLevel:        logLevel,
		logURL:          config.LogURL,
		batchSize:       batchSize,
		batchTimeoutSec: batchTimeoutSec,
		logToStdout:     config.LogToStdout,
		enableStdout:    config.EnableStdout,
		batchBuffer:     make([]interface{}, 0, batchSize),
		done:            make(chan struct{}),
	}

	// Start background batch sender goroutine
	logger.wg.Add(1)
	go logger.batchSender()

	return logger, nil
}

// Info logs an info level message with optional structured fields.
func (l *Logger) Info(msg string, keyvals ...interface{}) {
	l.log("info", msg, keyvals...)
}

// Debug logs a debug level message with optional structured fields.
func (l *Logger) Debug(msg string, keyvals ...interface{}) {
	l.log("debug", msg, keyvals...)
}

// Warn logs a warning level message with optional structured fields.
func (l *Logger) Warn(msg string, keyvals ...interface{}) {
	l.log("warn", msg, keyvals...)
}

// Error logs an error level message with optional structured fields.
func (l *Logger) Error(msg string, keyvals ...interface{}) {
	l.log("error", msg, keyvals...)
}

// Fatal logs a fatal level message with optional structured fields and exits.
func (l *Logger) Fatal(msg string, keyvals ...interface{}) {
	l.log("fatal", msg, keyvals...)
	// Note: actual exit happens in implementation
}

// Panic logs a panic level message with optional structured fields.
func (l *Logger) Panic(msg string, keyvals ...interface{}) {
	l.log("panic", msg, keyvals...)
	// Note: actual panic happens in implementation
}

// WithContext returns a logger with context-extracted values.
func (l *Logger) WithContext(ctx context.Context) Interface {
	// Implementation will extract correlation ID and other values from context
	return l
}

// WithFields returns a logger with additional structured fields.
func (l *Logger) WithFields(keyvals ...interface{}) Interface {
	// Implementation will add fields
	return l
}

// Flush ensures all pending logs are sent.
func (l *Logger) Flush(ctx context.Context) error {
	// Implementation will send pending logs
	return nil
}

// Close gracefully shuts down the logger.
func (l *Logger) Close() error {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return nil
	}
	l.closed = true
	l.mu.Unlock()

	close(l.done)
	l.wg.Wait()

	// Flush any remaining logs
	return l.Flush(context.Background())
}

// log adds a log entry to the batch buffer.
func (l *Logger) log(_, msg string, keyvals ...interface{}) {
	if l.isClosed() {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// TODO: Create LogEntry and add to batch buffer
	// This will be implemented in GREEN phase
	_ = msg
	_ = keyvals
}

// batchSender sends batched logs periodically or when buffer is full.
func (l *Logger) batchSender() {
	defer l.wg.Done()

	// TODO: Implement batching loop with timeout
	// This will be implemented in GREEN phase

	<-l.done
}

// isClosed checks if logger is closed.
func (l *Logger) isClosed() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.closed
}

// Global logger instance
var globalLogger Interface

// GetGlobalLogger returns the global logger instance.
func GetGlobalLogger() Interface {
	return globalLogger
}

// SetGlobalLogger sets the global logger instance.
func SetGlobalLogger(logger Interface) {
	globalLogger = logger
}

// Global logging functions that use the global logger instance

// LogInfo logs an info level message using the global logger.
func LogInfo(msg string, keyvals ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(msg, keyvals...)
	}
}

// LogDebug logs a debug level message using the global logger.
func LogDebug(msg string, keyvals ...interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(msg, keyvals...)
	}
}

// LogWarn logs a warning level message using the global logger.
func LogWarn(msg string, keyvals ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(msg, keyvals...)
	}
}

// LogError logs an error level message using the global logger.
func LogError(msg string, keyvals ...interface{}) {
	if globalLogger != nil {
		globalLogger.Error(msg, keyvals...)
	}
}

// LogFatal logs a fatal level message using the global logger.
func LogFatal(msg string, keyvals ...interface{}) {
	if globalLogger != nil {
		globalLogger.Fatal(msg, keyvals...)
	}
}

// LogPanic logs a panic level message using the global logger.
func LogPanic(msg string, keyvals ...interface{}) {
	if globalLogger != nil {
		globalLogger.Panic(msg, keyvals...)
	}
}
