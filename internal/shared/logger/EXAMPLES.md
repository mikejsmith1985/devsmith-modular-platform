# Structured Logging SDK - Usage Examples

This document provides examples of how to use the structured logging SDK across all services in the DevSmith platform.

## Quick Start

### Basic Usage

```go
package main

import (
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

func main() {
	// Create logger with configuration
	config := &logger.Config{
		ServiceName:     "my-service",
		LogLevel:        "info",          // debug, info, warn, error, fatal
		LogURL:          "http://logs-service:8082/api/logs",
		BatchSize:       100,             // Send when 100 logs accumulated
		BatchTimeoutSec: 5,               // Or after 5 seconds
		LogToStdout:     true,            // Also log to stdout
	}

	log, err := logger.NewLogger(config)
	if err != nil {
		panic(err)
	}
	defer log.Close()

	// Simple logging
	log.Info("Server started successfully")
	log.Warn("Running in development mode")
	log.Error("Failed to connect to database")
}
```

## Structured Fields

Log events with key-value pairs for better searchability and analysis:

```go
// Basic field logging
log.Info("User login attempt", "user_id", "123", "ip_address", "192.168.1.1")

// Multiple fields of different types
log.Info("API request completed",
	"method", "POST",
	"endpoint", "/api/users",
	"status_code", 200,
	"duration_ms", 1234,
	"success", true,
	"records_processed", 500,
)

// Nested complex data
log.Error("Payment processing failed",
	"transaction_id", "txn-456",
	"amount", 99.99,
	"retry_count", 3,
	"error_code", "PAYMENT_GATEWAY_TIMEOUT",
)
```

## Context Integration

Extract request-scoped values from Go context and inject them into logs:

```go
import (
	"context"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// In your request handler
func handleUserUpdate(w http.ResponseWriter, r *http.Request) {
	// Get context from request
	ctx := r.Context()

	// Add context values
	ctx = context.WithValue(ctx, logger.CorrelationIDKey, "req-abc123")
	ctx = context.WithValue(ctx, logger.UserIDKey, "user-789")
	ctx = context.WithValue(ctx, logger.RequestIDKey, "request-xyz")

	// Create logger with context - all logs will include correlation ID and user ID
	contextLogger := log.WithContext(ctx)
	contextLogger.Info("User update started")
	// Logs will automatically include: correlation_id=req-abc123, user_id=user-789, request_id=request-xyz
}
```

## Method Chaining

Chain `WithContext()` and `WithFields()` for flexible logging:

```go
// Chain context and fields
ctx := context.WithValue(context.Background(), logger.CorrelationIDKey, "req-123")
log.WithContext(ctx).
	WithFields("user_id", "user-456", "action", "create").
	Info("User created successfully", "email", "user@example.com")

// Chaining order doesn't matter
log.WithFields("service_version", "2.1.0").
	WithContext(ctx).
	Warn("Deprecated API endpoint used", "endpoint", "/api/v1/users")
```

## Log Levels

Use appropriate log levels for different severity:

```go
// DEBUG - Detailed diagnostic information
log.Debug("Connection pool stats", "active", 42, "idle", 8, "pending", 2)

// INFO - General informational messages
log.Info("Service initialized", "version", "1.0.0", "environment", "production")

// WARN - Warning messages (potential issues)
log.Warn("High latency detected", "duration_ms", 5000, "threshold_ms", 1000)

// ERROR - Error messages (operation failed)
log.Error("Database query failed", "query", "SELECT ...", "retries", 3)

// FATAL - Critical error, application exit
log.Fatal("Configuration file not found", "expected_path", "/etc/config.yaml")
// Program exits after this log

// PANIC - Critical error, panic
log.Panic("Null pointer dereference detected", "function", "processPayment")
// Panic is triggered after this log
```

## Global Logger

Use global logger functions as shortcuts:

```go
import "github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"

// Initialize global logger
logger.SetGlobalLogger(config)

// Use global functions directly
logger.Info("Global logging example")
logger.Error("Error occurred", "code", "ERR_001")
logger.WithContext(ctx).Warn("Context-aware global logging")
```

## Service-Specific Examples

### Portal Service

```go
// cmd/portal/main.go
package main

import (
	"context"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

func init() {
	config := &logger.Config{
		ServiceName:     "portal",
		LogLevel:        os.Getenv("LOG_LEVEL"),
		LogURL:          os.Getenv("LOGS_SERVICE_URL"),
		BatchSize:       100,
		BatchTimeoutSec: 5,
		LogToStdout:     true,
	}

	log, err := logger.NewLogger(config)
	if err != nil {
		panic(err)
	}
	logger.SetGlobalLogger(log)
}

// In authentication handler
func authenticateUser(c *gin.Context) {
	ctx := context.WithValue(c.Request.Context(), logger.CorrelationIDKey, c.GetString("request_id"))
	
	logger.WithContext(ctx).Info("Authentication attempt", 
		"email", c.PostForm("email"),
		"method", "oauth",
	)

	// ... authentication logic ...

	logger.WithContext(ctx).Info("User authenticated successfully",
		"user_id", user.ID,
		"session_duration_sec", 3600,
	)
}
```

### Review Service

```go
// cmd/review/main.go
package main

import (
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

func init() {
	config := &logger.Config{
		ServiceName:     "review",
		LogLevel:        "info",
		LogURL:          "http://logs-service:8082/api/logs",
		BatchSize:       50,
		BatchTimeoutSec: 3,
		LogToStdout:     true,
	}

	log, err := logger.NewLogger(config)
	if err != nil {
		panic(err)
	}
	logger.SetGlobalLogger(log)
}

// In code review handler
func submitCodeReview(c *gin.Context) {
	logger.Info("Code review submission started",
		"pull_request_id", pr.ID,
		"reviewer_count", len(pr.Reviewers),
		"files_changed", len(pr.Files),
	)

	// Process review...

	logger.Info("Code review completed",
		"pull_request_id", pr.ID,
		"status", "approved",
		"comments", 5,
		"suggestions", 2,
	)
}
```

### Analytics Service

```go
// cmd/analytics/main.go
package main

import (
	"context"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

func init() {
	config := &logger.Config{
		ServiceName:     "analytics",
		LogLevel:        "debug",
		LogURL:          "http://logs-service:8082/api/logs",
		BatchSize:       200,           // Larger batch for high volume
		BatchTimeoutSec: 2,
		LogToStdout:     true,
	}

	log, err := logger.NewLogger(config)
	if err != nil {
		panic(err)
	}
	logger.SetGlobalLogger(log)
}

// In analytics processor
func processAnalytics(ctx context.Context, events []Event) {
	logger.WithContext(ctx).Info("Analytics processing started",
		"event_count", len(events),
		"batch_id", ctx.Value("batch_id"),
	)

	for i, event := range events {
		if err := processEvent(event); err != nil {
			logger.WithContext(ctx).Error("Failed to process event",
				"index", i,
				"event_id", event.ID,
				"error", err.Error(),
			)
			continue
		}
	}

	logger.WithContext(ctx).Info("Analytics batch completed",
		"processed_count", len(events),
		"duration_ms", time.Since(startTime).Milliseconds(),
	)
}
```

### Logs Service

```go
// cmd/logs/main.go
package main

import (
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

func init() {
	config := &logger.Config{
		ServiceName:     "logs",
		LogLevel:        "info",
		LogURL:          "", // Logs service doesn't send to itself
		BatchSize:       100,
		BatchTimeoutSec: 5,
		LogToStdout:     true,
	}

	log, err := logger.NewLogger(config)
	if err != nil {
		panic(err)
	}
	logger.SetGlobalLogger(log)
}

// In log ingestion handler
func ingestLogs(c *gin.Context) {
	var req LogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid log request",
			"error", err.Error(),
			"content_type", c.ContentType(),
		)
		return
	}

	logger.Info("Logs received",
		"log_count", len(req.Logs),
		"services", len(groupByService(req.Logs)),
	)

	// Store logs...
}
```

## Performance Best Practices

### 1. Batch Configuration
```go
// For high-volume services (Analytics, Review)
config.BatchSize = 200        // Larger batches
config.BatchTimeoutSec = 2    // Shorter timeout

// For normal services
config.BatchSize = 100        // Default
config.BatchTimeoutSec = 5    // Default

// For low-volume services (Logs ingestion)
config.BatchSize = 50         // Smaller batches
config.BatchTimeoutSec = 3    // Shorter timeout
```

### 2. Field Cardinality
```go
// ✅ GOOD - Low cardinality fields
log.Info("Request processed",
	"method", "POST",        // Low cardinality (GET, POST, etc)
	"status", 200,           // Low cardinality (200, 404, 500, etc)
	"endpoint", "/api/users", // Moderate cardinality
)

// ❌ AVOID - High cardinality fields
log.Info("Request processed",
	"user_id", userID,       // High cardinality (millions of values)
	"session_id", sessionID, // High cardinality (unique per session)
	"timestamp", now,        // Created automatically - don't include
)
```

### 3. Graceful Shutdown
```go
import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log, _ := logger.NewLogger(config)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Shutdown signal received")
		_ = log.Close() // Flush all pending logs before exit
		os.Exit(0)
	}()

	// ... rest of application ...
}
```

### 4. Error Handling
```go
// Always handle Flush errors
if err := log.Flush(ctx); err != nil {
	log.Error("Failed to flush logs", "error", err.Error())
}

// Always handle Close errors
if err := log.Close(); err != nil {
	log.Error("Failed to close logger", "error", err.Error())
}
```

## Integration with Gin Middleware

```go
// Middleware to inject correlation ID from request headers
func LoggingMiddleware(log logger.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract or generate correlation ID
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = generateID()
		}

		// Inject into context
		ctx := context.WithValue(c.Request.Context(), logger.CorrelationIDKey, correlationID)
		ctx = context.WithValue(ctx, logger.UserIDKey, c.GetString("user_id"))

		// Create context logger
		c.Set("logger", log.WithContext(ctx))

		// Log request
		log.WithContext(ctx).Info("Request started",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)

		c.Next()

		// Log response
		log.WithContext(ctx).Info("Request completed",
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}

// Use in main
func main() {
	router := gin.Default()
	router.Use(LoggingMiddleware(log))
	// ... rest of setup ...
}
```

## Testing with Logger

```go
// In tests, create logger with disabled service URL
testLogger, _ := logger.NewLogger(&logger.Config{
	ServiceName:     "test-service",
	LogLevel:        "debug",
	LogURL:          "",  // No external service in tests
	BatchSize:       1,
	BatchTimeoutSec: 1,
	LogToStdout:     false, // Suppress output in tests
})

// Use in tests
testLogger.Info("Test message", "test_id", "123")
```

## Migration from logrus

If migrating from logrus:

```go
// OLD (logrus)
logrus.WithFields(logrus.Fields{
	"user_id": 123,
	"action": "login",
}).Info("User logged in")

// NEW (structured logger)
log.WithFields("user_id", 123, "action", "login").
	Info("User logged in")
```

For more information, see the integration tests in `logger_integration_test.go` and unit tests in `logger_test.go`.
