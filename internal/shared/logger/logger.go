// Package logger provides a structured logging SDK for services.
package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
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
	batchBuffer []*LogEntry

	// done signals goroutines to stop.
	done chan struct{}

	// wg waits for goroutines to finish.
	wg sync.WaitGroup

	// httpClient for sending logs to service.
	httpClient *http.Client
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
		batchBuffer:     make([]*LogEntry, 0, batchSize),
		done:            make(chan struct{}),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
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
	os.Exit(1)
}

// Panic logs a panic level message with optional structured fields.
func (l *Logger) Panic(msg string, keyvals ...interface{}) {
	l.log("panic", msg, keyvals...)
	panic(msg)
}

// WithContext returns a logger with context-extracted values.
func (l *Logger) WithContext(ctx context.Context) Interface {
	// Create a new logger instance that shares the same components but has additional context fields
	contextFields := make(map[string]interface{})

	// Extract known fields from context
	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		contextFields["correlation_id"] = correlationID
	}
	if userID := ctx.Value(UserIDKey); userID != nil {
		contextFields["user_id"] = userID
	}
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		contextFields["request_id"] = requestID
	}

	// Create a wrapper logger with context fields
	return &loggerWithFields{
		logger:        l,
		contextFields: contextFields,
	}
}

// WithFields returns a logger with additional structured fields.
func (l *Logger) WithFields(keyvals ...interface{}) Interface {
	contextFields := make(map[string]interface{})

	// Add new fields (key-value pairs)
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key := fmt.Sprintf("%v", keyvals[i])
			contextFields[key] = keyvals[i+1]
		}
	}

	// Create a wrapper logger with fields
	return &loggerWithFields{
		logger:        l,
		contextFields: contextFields,
	}
}

// Flush ensures all pending logs are sent.
func (l *Logger) Flush(ctx context.Context) error {
	l.mu.Lock()
	if len(l.batchBuffer) > 0 {
		logs := make([]*LogEntry, len(l.batchBuffer))
		copy(logs, l.batchBuffer)
		l.batchBuffer = l.batchBuffer[:0]
		l.mu.Unlock()

		// Send the batch
		_ = l.sendBatch(ctx, logs) //nolint:errcheck // Send errors are logged elsewhere
		return nil
	}
	l.mu.Unlock()
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
	_ = l.Flush(context.Background()) //nolint:errcheck // Flush errors are not critical on shutdown
	return nil
}

// log adds a log entry to the batch buffer.
func (l *Logger) log(level, msg string, keyvals ...interface{}) {
	if l.isClosed() {
		return
	}

	// Check if log level should be logged
	if !l.shouldLog(level) {
		return
	}

	// Build metadata from keyvals
	metadata := make(map[string]interface{})
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key := fmt.Sprintf("%v", keyvals[i])
			metadata[key] = keyvals[i+1]
		}
	}

	// The contextFields map is now managed by the wrapper, so we don't need to add them here.

	// Create log entry
	entry := &LogEntry{
		CreatedAt: time.Now().UTC(),
		Service:   l.serviceName,
		Level:     level,
		Message:   msg,
		Metadata:  metadata,
		Tags:      []string{level, l.serviceName},
	}

	l.mu.Lock()
	l.batchBuffer = append(l.batchBuffer, entry)
	shouldSend := len(l.batchBuffer) >= l.batchSize
	l.mu.Unlock()

	// If batch is full, send immediately
	if shouldSend {
		l.flushAsync()
	}
}

// flushAsync triggers an async flush without waiting for completion.
func (l *Logger) flushAsync() {
	go func() {
		_ = l.Flush(context.Background()) //nolint:errcheck // Async flush errors are non-critical
	}()
}

// batchSender sends batched logs periodically or when buffer is full.
func (l *Logger) batchSender() {
	defer l.wg.Done()

	ticker := time.NewTicker(time.Duration(l.batchTimeoutSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = l.Flush(context.Background()) //nolint:errcheck // Periodic flush errors are non-critical
		case <-l.done:
			return
		}
	}
}

// sendBatch sends a batch of logs to the logging service or stdout.
func (l *Logger) sendBatch(ctx context.Context, logs []*LogEntry) error {
	if len(logs) == 0 {
		return nil
	}

	// Log to stdout if enabled
	if l.logToStdout || l.enableStdout {
		for _, entry := range logs {
			l.logToStdoutEntry(entry)
		}
	}

	// Send to logging service if URL configured
	if l.logURL == "" {
		return nil
	}

	request := &LogRequest{Logs: logs}
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", l.logURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := l.httpClient.Do(req)
	if err != nil {
		// Fallback to stdout on error
		if l.enableStdout {
			for _, entry := range logs {
				l.logToStdoutEntry(entry)
			}
		}
		return fmt.Errorf("failed to send logs: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Closing response body errors are non-critical

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // ReadAll errors are non-critical for logging
		return fmt.Errorf("logging service returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// logToStdoutEntry logs a single entry to stdout.
func (l *Logger) logToStdoutEntry(entry *LogEntry) {
	prefix := fmt.Sprintf("[%s] %s", entry.Level, entry.Service)
	_, _ = fmt.Fprintf(os.Stdout, "%s: %s\n", prefix, entry.Message) //nolint:errcheck // Stdout write errors are non-critical

	// Include metadata if present
	if len(entry.Metadata) > 0 {
		metaJSON, _ := json.Marshal(entry.Metadata)                         //nolint:errcheck // Marshal errors are non-critical for logging
		_, _ = fmt.Fprintf(os.Stdout, "  metadata: %s\n", string(metaJSON)) //nolint:errcheck // Stdout write errors are non-critical
	}
}

// shouldLog checks if a log level should be logged based on configured level.
func (l *Logger) shouldLog(level string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
		"fatal": 4,
		"panic": 5,
	}

	configLevel, ok := levels[l.logLevel]
	if !ok {
		configLevel = 1 // default to info
	}

	logLevel, ok := levels[level]
	if !ok {
		return false
	}

	return logLevel >= configLevel
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

// loggerWithFields wraps a logger and adds context fields to all logs.
type loggerWithFields struct {
	logger        *Logger
	contextFields map[string]interface{}
}

// Info logs an info level message with additional context fields.
func (lf *loggerWithFields) Info(msg string, keyvals ...interface{}) {
	lf.logWithFields("info", msg, keyvals...)
}

// Debug logs a debug level message with additional context fields.
func (lf *loggerWithFields) Debug(msg string, keyvals ...interface{}) {
	lf.logWithFields("debug", msg, keyvals...)
}

// Warn logs a warning level message with additional context fields.
func (lf *loggerWithFields) Warn(msg string, keyvals ...interface{}) {
	lf.logWithFields("warn", msg, keyvals...)
}

// Error logs an error level message with additional context fields.
func (lf *loggerWithFields) Error(msg string, keyvals ...interface{}) {
	lf.logWithFields("error", msg, keyvals...)
}

// Fatal logs a fatal level message with additional context fields and exits.
func (lf *loggerWithFields) Fatal(msg string, keyvals ...interface{}) {
	lf.logWithFields("fatal", msg, keyvals...)
	os.Exit(1)
}

// Panic logs a panic level message with additional context fields.
func (lf *loggerWithFields) Panic(msg string, keyvals ...interface{}) {
	lf.logWithFields("panic", msg, keyvals...)
	panic(msg)
}

// WithContext returns a logger with additional context values.
func (lf *loggerWithFields) WithContext(ctx context.Context) Interface {
	newFields := make(map[string]interface{})
	for k, v := range lf.contextFields {
		newFields[k] = v
	}

	// Extract additional fields from context
	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		newFields["correlation_id"] = correlationID
	}
	if userID := ctx.Value(UserIDKey); userID != nil {
		newFields["user_id"] = userID
	}
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		newFields["request_id"] = requestID
	}

	return &loggerWithFields{
		logger:        lf.logger,
		contextFields: newFields,
	}
}

// WithFields returns a logger with additional fields.
func (lf *loggerWithFields) WithFields(keyvals ...interface{}) Interface {
	newFields := make(map[string]interface{})
	for k, v := range lf.contextFields {
		newFields[k] = v
	}

	// Add new fields (key-value pairs)
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key := fmt.Sprintf("%v", keyvals[i])
			newFields[key] = keyvals[i+1]
		}
	}

	return &loggerWithFields{
		logger:        lf.logger,
		contextFields: newFields,
	}
}

// Flush ensures all pending logs are sent.
func (lf *loggerWithFields) Flush(ctx context.Context) error {
	return lf.logger.Flush(ctx)
}

// Close gracefully shuts down the logger.
func (lf *loggerWithFields) Close() error {
	return lf.logger.Close()
}

// logWithFields logs a message with context fields.
func (lf *loggerWithFields) logWithFields(level, msg string, keyvals ...interface{}) {
	// Merge context fields with provided keyvals
	allKeyvals := make([]interface{}, 0, len(keyvals)+len(lf.contextFields)*2)

	// Add keyvals first
	allKeyvals = append(allKeyvals, keyvals...)

	// Add context fields
	for k, v := range lf.contextFields {
		allKeyvals = append(allKeyvals, k, v)
	}

	lf.logger.log(level, msg, allKeyvals...)
}
