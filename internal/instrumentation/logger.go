// Package instrumentation provides logging infrastructure for services.
package instrumentation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// ServiceInstrumentationLogger handles async logging for services.
type ServiceInstrumentationLogger struct {
	httpClient     *http.Client
	serviceName    string
	logsServiceURL string
}

// NewServiceInstrumentationLogger creates a new service instrumentation logger.
func NewServiceInstrumentationLogger(serviceName, logsServiceURL string) *ServiceInstrumentationLogger {
	return &ServiceInstrumentationLogger{
		serviceName:    serviceName,
		logsServiceURL: logsServiceURL,
		httpClient: &http.Client{
			Timeout: 2 * time.Second, // Fast timeout to avoid blocking
		},
	}
}

// LogEvent logs a generic event asynchronously.
func (l *ServiceInstrumentationLogger) LogEvent(ctx context.Context, eventType string, metadata map[string]interface{}) error {
	logEntry := l.buildLogEntry("info", eventType, metadata, ctx)
	l.sendAsync(logEntry)
	return nil // Always return nil - never block on logging
}

// LogValidationFailure logs a validation failure.
func (l *ServiceInstrumentationLogger) LogValidationFailure(ctx context.Context, errorType, message string, metadata map[string]interface{}) error {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["error_type"] = errorType
	logEntry := l.buildLogEntry("warning", "validation_failure", metadata, ctx)
	logEntry["message"] = message
	l.sendAsync(logEntry)
	return nil
}

// LogSecurityViolation logs a security violation.
func (l *ServiceInstrumentationLogger) LogSecurityViolation(ctx context.Context, errorType, message string, metadata map[string]interface{}) error {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["error_type"] = errorType
	logEntry := l.buildLogEntry("error", "security_violation", metadata, ctx)
	logEntry["message"] = message
	l.sendAsync(logEntry)
	return nil
}

// LogError logs an error event.
func (l *ServiceInstrumentationLogger) LogError(ctx context.Context, errorType, message string, metadata map[string]interface{}) error {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["error_type"] = errorType
	logEntry := l.buildLogEntry("error", "service_error", metadata, ctx)
	logEntry["message"] = message
	l.sendAsync(logEntry)
	return nil
}

// HasCircularDependencyPrevention returns true if circular dependency prevention is enabled.
func (l *ServiceInstrumentationLogger) HasCircularDependencyPrevention() bool {
	// Logs service should have its own circular dependency prevention
	return l.serviceName == "logs"
}

// buildLogEntry constructs a log entry with context information.
func (l *ServiceInstrumentationLogger) buildLogEntry(level, eventType string, metadata map[string]interface{}, ctx context.Context) map[string]interface{} {
	logEntry := map[string]interface{}{
		"service":  l.serviceName,
		"level":    level,
		"message":  eventType, // Use eventType as message for logs service compatibility
		"metadata": metadata,
	}

	// Extract request_id from context if available
	if requestID := l.extractRequestID(ctx); requestID != "" {
		logEntry["request_id"] = requestID
	}

	return logEntry
}

// extractRequestID tries to extract request ID from context.
func (l *ServiceInstrumentationLogger) extractRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Try common request ID context keys
	if rid := ctx.Value("request_id"); rid != nil {
		if ridStr, ok := rid.(string); ok {
			return ridStr
		}
	}
	if rid := ctx.Value("X-Request-ID"); rid != nil {
		if ridStr, ok := rid.(string); ok {
			return ridStr
		}
	}
	if rid := ctx.Value("request-id"); rid != nil {
		if ridStr, ok := rid.(string); ok {
			return ridStr
		}
	}

	return ""
}

// sendAsync sends the log asynchronously without blocking.
func (l *ServiceInstrumentationLogger) sendAsync(logEntry map[string]interface{}) {
	// Circular dependency prevention for logs service
	if l.serviceName == "logs" && logEntry["event_type"] == "log_entry_ingested" {
		// Don't re-log the logs service's own log ingestion events
		return
	}

	go func() {
		defer func() {
			// Silently recover from any panic in async logging
			//nolint:errcheck // Intentionally ignoring recover errors in async logging
			_ = recover()
		}()

		jsonData, err := json.Marshal(logEntry)
		if err != nil {
			return // Can't marshal, give up silently
		}

		// DEBUG: Log to stderr so we can see if this is being called
		fmt.Fprintf(os.Stderr, "[DEBUG] Sending log to %s from %s\n", l.logsServiceURL, l.serviceName)

		// Create a context with timeout for the HTTP request
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "POST", l.logsServiceURL+"/api/logs", bytes.NewReader(jsonData))
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to create request: %v\n", err)
			return // Can't create request, give up silently
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := l.httpClient.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] HTTP request failed: %v\n", err)
			return // Network error, fail silently (don't block)
		}

		if resp != nil {
			// Best effort to close response body
			//nolint:errcheck // Intentionally ignoring close errors in async logging
			_ = resp.Body.Close()
		}

		fmt.Fprintf(os.Stderr, "[DEBUG] Log sent successfully\n")
	}()
}
