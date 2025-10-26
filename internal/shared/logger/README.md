# Structured Logging SDK

The structured logging SDK provides a centralized, async logging client for all services in the DevSmith platform. It automatically batches logs and sends them to the Logs service with support for context injection, multiple log levels, and graceful fallback.

## Table of Contents

- [Quick Start](#quick-start)
- [Core Features](#core-features)
- [Log Levels](#log-levels)
- [Structured Fields](#structured-fields)
- [Context Integration](#context-integration)
- [Configuration](#configuration)
- [Performance](#performance)
- [Architecture](#architecture)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [API Reference](#api-reference)

## Quick Start

```go
package main

import (
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

func main() {
	// Create logger
	config := &logger.Config{
		ServiceName:     "my-service",
		LogLevel:        "info",
		LogURL:          "http://logs-service:8082/api/logs",
		BatchSize:       100,
		BatchTimeoutSec: 5,
		LogToStdout:     true,
	}

	log, err := logger.NewLogger(config)
	if err != nil {
		panic(err)
	}
	defer log.Close()

	// Log a message
	log.Info("Service started", "port", 8080, "environment", "production")
}
```

## Core Features

### Asynchronous Batching
- Logs are accumulated in memory and sent to the service in batches
- Batches are sent when either:
  1. Batch size threshold is reached (e.g., 100 logs), OR
  2. Timeout expires (e.g., 5 seconds)
- Whichever happens first
- Non-blocking: logging calls return immediately

### All Log Levels
```go
logger.Debug("diagnostic info")     // Only if log level is "debug"
logger.Info("operation completed")   // Default threshold
logger.Warn("potential issue")       // Needs attention
logger.Error("operation failed")     // Error occurred
logger.Fatal("critical error")       // Exits process
logger.Panic("invariant violated")   // Panics
```

### Structured Fields
```go
// All types supported: string, int, float, bool, etc.
logger.Info("operation completed",
	"duration_ms", 1234,
	"success", true,
	"retry_count", 3,
	"error_code", "TIMEOUT",
)
```

### Context Injection
```go
// Extract values from context automatically
ctx := context.WithValue(context.Background(), logger.CorrelationIDKey, "req-123")
ctx = context.WithValue(ctx, logger.UserIDKey, "user-456")

logger.WithContext(ctx).Info("user action")
// Logs will include: correlation_id, user_id
```

### Graceful Fallback
- If the Logs service is unavailable, logs automatically fall back to stdout
- No logs are lost
- Service continues operating normally

## Log Levels

Log levels in order of severity:

| Level | Usage | When to Use |
|-------|-------|-----------|
| DEBUG | Detailed diagnostic info | Development, detailed tracing |
| INFO | General information | Normal operation milestones |
| WARN | Potential issues | Degraded operation, retries |
| ERROR | Error occurred | Operation failed, service continues |
| FATAL | Critical error, exit | Unrecoverable error, must exit |
| PANIC | Critical error, panic | Invariant violated, program error |

### Log Level Filtering

Only logs at or above the configured level are sent:

```
LogLevel="debug"  → sends: DEBUG, INFO, WARN, ERROR, FATAL, PANIC
LogLevel="info"   → sends: INFO, WARN, ERROR, FATAL, PANIC
LogLevel="warn"   → sends: WARN, ERROR, FATAL, PANIC
LogLevel="error"  → sends: ERROR, FATAL, PANIC
LogLevel="fatal"  → sends: FATAL, PANIC
```

## Structured Fields

### Supported Types

```go
logger.Info("example",
	"string", "value",
	"int", 42,
	"float", 3.14,
	"bool", true,
	"interface", someObject,
)
```

### Key-Value Pairs

Fields must be provided as alternating key-value pairs:

```go
// ✅ CORRECT
logger.Info("message", "key1", "value1", "key2", "value2")

// ❌ WRONG - Missing value for last key
logger.Info("message", "key1", "value1", "key2")
```

### Field Naming

- Use snake_case for field names
- Keep names short but descriptive
- Avoid high-cardinality fields (like user_id for every log)

```go
// ✅ GOOD
logger.Info("request processed",
	"method", "POST",           // Low cardinality
	"status", 200,              // Low cardinality
	"duration_ms", 1234,        // Numeric
)

// ❌ AVOID
logger.Info("request processed",
	"user_id", userID,          // High cardinality
	"timestamp", now,           // Already created automatically
)
```

## Context Integration

### Extracting Values from Context

The logger automatically extracts:
- `CorrelationIDKey` - Trace requests across services
- `UserIDKey` - Identify the user
- `RequestIDKey` - Track individual requests

```go
ctx := context.WithValue(context.Background(), logger.CorrelationIDKey, "req-abc123")
ctx = context.WithValue(ctx, logger.UserIDKey, "user-789")

logger.WithContext(ctx).Info("user action")
// Logs will automatically include: correlation_id, user_id
```

### Method Chaining

Combine context and fields:

```go
ctx := context.WithValue(context.Background(), logger.CorrelationIDKey, "req-123")

logger.
	WithContext(ctx).
	WithFields("action", "create", "resource", "user").
	Info("user created")
```

### Middleware Integration

Extract correlation ID from request headers:

```go
func LoggingMiddleware(log logger.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract or generate correlation ID
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Inject into context
		ctx := context.WithValue(c.Request.Context(), logger.CorrelationIDKey, correlationID)
		ctx = context.WithValue(ctx, logger.UserIDKey, c.GetString("user_id"))

		// Use throughout request
		c.Set("logger", log.WithContext(ctx))

		c.Next()
	}
}

// Use middleware
router := gin.Default()
router.Use(LoggingMiddleware(log))
```

## Configuration

### Required Settings

```go
config := &logger.Config{
	ServiceName: "my-service",  // REQUIRED - name of your service
}
```

### Optional Settings

| Setting | Default | Recommendation |
|---------|---------|-----------------|
| LogLevel | "info" | Change based on needs |
| LogURL | "" | Leave empty for stdout only |
| BatchSize | 100 | 50-200 depending on volume |
| BatchTimeoutSec | 5 | 2-5 depending on latency needs |
| LogToStdout | false | true for development |
| EnableStdout | false | true for safety |

### Service-Specific Configurations

**High-Volume Services** (Analytics, Review):
```go
config := &logger.Config{
	ServiceName:     "analytics",
	BatchSize:       200,           // Larger batches
	BatchTimeoutSec: 2,             // Shorter timeout
	LogLevel:        "info",
}
```

**Normal Services** (Portal, Logs):
```go
config := &logger.Config{
	ServiceName:     "portal",
	BatchSize:       100,           // Default
	BatchTimeoutSec: 5,             // Default
	LogLevel:        "info",
}
```

**Low-Volume Services**:
```go
config := &logger.Config{
	ServiceName:     "worker",
	BatchSize:       50,            // Smaller batches
	BatchTimeoutSec: 10,            // Longer timeout
	LogLevel:        "info",
}
```

## Performance

### Batching Strategy

The logger uses two thresholds to determine when to send:

```
┌─────────────────────────────────────────────────────┐
│ Log Buffer                                          │
│ ┌──────────────────────────────────────────────┐   │
│ │ Log 1: msg="event"                           │   │
│ │ Log 2: msg="event"                           │   │
│ │ Log 3: msg="event"                           │   │
│ │ ...                                          │   │
│ │ Log 99: msg="event"                          │   │
│ │ Log 100: msg="event"  ← Batch Size Reached! │   │
│ └──────────────────────────────────────────────┘   │
│                           ↓                        │
│                       SEND BATCH                   │
└─────────────────────────────────────────────────────┘

If timeout expires before batch size reached:
- Timer fires after 5 seconds
- All buffered logs sent immediately
```

### Thread Safety

All operations are thread-safe for concurrent logging:

```go
// Safe from multiple goroutines
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
	wg.Add(1)
	go func(id int) {
		logger.Info("concurrent log", "goroutine", id)
		wg.Done()
	}(i)
}
wg.Wait()
```

### Non-Blocking Operations

Logging calls return immediately:

```go
// This returns immediately, sends happen in background
logger.Info("message", "field", "value")
```

### Graceful Shutdown

Always call Close() to ensure pending logs are sent:

```go
log, _ := logger.NewLogger(config)

// Handle graceful shutdown
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

go func() {
	<-sigChan
	_ = log.Close()  // Flush and close
	os.Exit(0)
}()
```

## Architecture

### Internal Components

```
Application
    ↓
    logger.Info("msg")
    ↓
    ┌─────────────────────────────────────────┐
    │ Logger                                  │
    │ ┌──────────────────────────────────┐   │
    │ │ Batch Buffer (sync.RWMutex)      │   │
    │ │ - Thread-safe                    │   │
    │ │ - Auto-grow as needed            │   │
    │ └──────────────────────────────────┘   │
    │                ↓                       │
    │ ┌──────────────────────────────────┐   │
    │ │ Batch Sender (goroutine)         │   │
    │ │ - time.Ticker for timeout        │   │
    │ │ - Sends on size OR timeout       │   │
    │ └──────────────────────────────────┘   │
    └─────────────────────────────────────────┘
            ↓
        HTTP POST
            ↓
    Logs Service (/api/logs)
            ↓
        Database
```

### Design Patterns

1. **Wrapper Pattern**: `loggerWithFields` decorates the base logger without copying mutex
2. **Background Goroutine**: Async batching in separate goroutine
3. **Sink Pattern**: Fallback to stdout on service failure
4. **Singleton Pattern**: Global logger for convenience

## Best Practices

### 1. Always Defer Close()

```go
log, _ := logger.NewLogger(config)
defer log.Close()  // Ensures pending logs are sent
```

### 2. Use Appropriate Log Levels

```go
// ✅ GOOD - Clear intent
log.Debug("connection established")        // Very detailed
log.Info("user authenticated")              // Milestone
log.Warn("retry attempt 3 of 5")            // Potential issue
log.Error("database query failed")          // Operation failed
log.Fatal("config file not found")          // Must exit

// ❌ POOR - Unclear intent
log.Info("user authenticated")
log.Info("something went wrong")  // Should be Error or Warn
log.Info("debug: status = " + status)  // Should be Debug
```

### 3. Provide Context for Debugging

```go
// ❌ POOR - Not enough info
log.Error("operation failed")

// ✅ GOOD - Rich context
log.Error("database query failed",
	"query", "SELECT * FROM users",
	"timeout_ms", 5000,
	"retry_count", 3,
	"error_code", "QUERY_TIMEOUT",
)
```

### 4. Use Correlation IDs

```go
// Extract from request headers
ctx := context.WithValue(
	r.Context(),
	logger.CorrelationIDKey,
	r.Header.Get("X-Correlation-ID"),
)

logger.WithContext(ctx).Info("processing request")
```

### 5. Batch Configuration Tuning

```go
// Monitor actual batch sizes and adjust accordingly
// High-volume service sending too frequently?
// → Increase BatchSize, decrease BatchTimeoutSec

// Low-volume service with stale logs?
// → Decrease BatchTimeoutSec to increase freshness
```

### 6. Avoid High-Cardinality Fields

```go
// ❌ POOR - Each log has different user_id (millions of values)
for _, user := range users {
	log.Info("processing", "user_id", user.ID)  // Bad!
}

// ✅ GOOD - Logs by type/status (few values)
log.Info("batch processed",
	"total_users", len(users),
	"status", "completed",
	"duration_ms", elapsed,
)
```

## Troubleshooting

### Logs Not Appearing

**Check 1: Log Level**
```go
// If LogLevel="error", info messages won't be sent
config.LogLevel = "info"  // or "debug"
```

**Check 2: Service URL**
```go
// Verify the Logs service is running
config.LogURL = "http://logs-service:8082/api/logs"
```

**Check 3: Fallback to Stdout**
```go
// Enable stdout to see logs even if service fails
config.LogToStdout = true
config.EnableStdout = true
```

### High Memory Usage

**Issue**: Buffer growing too large
```go
// Reduce batch timeout or increase batch size
config.BatchSize = 50         // Send more frequently
config.BatchTimeoutSec = 2    // Don't wait as long
```

### Network Timeouts

**Issue**: Logs taking too long to send
```go
// Reduce batch size for smaller payloads
config.BatchSize = 50
// Or check network connectivity to Logs service
```

### Logs Lost During Shutdown

**Issue**: Logs not flushed before exit
```go
// Always call Close() in defer
defer log.Close()

// Or explicitly flush before shutdown
log.Flush(context.Background())
```

## API Reference

### NewLogger(config *Config) (Interface, error)

Creates a new logger with the given configuration.

```go
log, err := logger.NewLogger(&logger.Config{
	ServiceName: "my-service",
})
if err != nil {
	return err
}
```

### Info(msg string, keyvals ...interface{})

Logs an info level message.

```go
log.Info("user created", "user_id", 123, "email", "user@example.com")
```

### Debug(msg string, keyvals ...interface{})

Logs a debug level message (only if LogLevel="debug").

```go
log.Debug("database query", "query", "SELECT * FROM users", "rows", 1000)
```

### Warn(msg string, keyvals ...interface{})

Logs a warning level message.

```go
log.Warn("high latency", "duration_ms", 5000, "threshold_ms", 1000)
```

### Error(msg string, keyvals ...interface{})

Logs an error level message.

```go
log.Error("request failed", "status", 500, "error", "Internal Server Error")
```

### Fatal(msg string, keyvals ...interface{})

Logs a fatal message and exits the process.

```go
log.Fatal("configuration invalid", "file", "config.yaml")
// Process exits here
```

### Panic(msg string, keyvals ...interface{})

Logs a panic message and panics.

```go
log.Panic("critical error", "invariant", "connection != nil")
// Program panics here
```

### WithContext(ctx context.Context) Interface

Returns a logger with context-extracted values.

```go
ctx := context.WithValue(context.Background(), logger.CorrelationIDKey, "req-123")
log.WithContext(ctx).Info("message")
```

### WithFields(keyvals ...interface{}) Interface

Returns a logger with additional fields.

```go
log.WithFields("user_id", 123).Info("message")
```

### Flush(ctx context.Context) error

Sends all pending logs synchronously.

```go
if err := log.Flush(context.Background()); err != nil {
	return err
}
```

### Close() error

Gracefully shuts down the logger, flushing all pending logs.

```go
if err := log.Close(); err != nil {
	return err
}
```

---

For detailed usage examples, see [EXAMPLES.md](EXAMPLES.md).
