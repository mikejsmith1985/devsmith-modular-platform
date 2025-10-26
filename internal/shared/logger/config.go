package logger

const (
	// DefaultBatchSize is the default number of logs to batch before sending.
	// Recommended for most services. Higher values reduce network traffic,
	// lower values improve latency. Adjust based on your service's logging volume.
	DefaultBatchSize = 100

	// DefaultBatchTimeoutSec is the default timeout in seconds for batching.
	// Ensures logs are sent even if batch size is not reached.
	// Recommended for most services. Higher values improve efficiency,
	// lower values improve log freshness.
	DefaultBatchTimeoutSec = 5

	// DefaultLogLevel is the default log level.
	DefaultLogLevel = "info"
)

// Config represents the configuration for the logger.
// All fields except LogLevel and LogToStdout have sensible defaults.
type Config struct {
	// ServiceName is the name of the service using the logger (required).
	// This will be automatically injected into every log entry.
	// Example: "portal", "review", "analytics", "logs"
	ServiceName string

	// LogLevel is the logging level (debug, info, warn, error, fatal).
	// Only messages at or above this level will be logged.
	// Case-insensitive. Defaults to "info" if not provided.
	// Levels in order: debug < info < warn < error < fatal
	LogLevel string

	// LogURL is the URL of the logging service (optional, logs to stdout if not provided).
	// Should be the endpoint where logs are POSTed. For example:
	// "http://logs-service:8082/api/logs"
	// If empty or unreachable, logs will fall back to stdout.
	LogURL string

	// BatchSize is the number of logs to batch before sending.
	// Triggers a send when the buffer reaches this size.
	// Defaults to DefaultBatchSize (100) if not provided or <= 0.
	// Recommended: 100 for normal services, 50 for low-volume, 200+ for high-volume.
	BatchSize int

	// BatchTimeoutSec is the timeout in seconds for batching.
	// Triggers a send if this timeout is reached, even if batch size not reached.
	// Defaults to DefaultBatchTimeoutSec (5) if not provided or <= 0.
	// Recommended: 5 for normal services, 3 for low-volume, 2 for high-volume.
	BatchTimeoutSec int

	// LogToStdout indicates whether to also log to stdout.
	// If true, logs will be printed to stdout in addition to being sent to the service.
	// Useful for development or debugging.
	LogToStdout bool

	// EnableStdout indicates whether stdout is enabled as fallback.
	// If true and the service is unavailable, logs will fall back to stdout.
	// Should typically be true to avoid losing logs on service failure.
	EnableStdout bool
}
