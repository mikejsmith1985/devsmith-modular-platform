package logger

const (
	// DefaultBatchSize is the default number of logs to batch before sending.
	DefaultBatchSize = 100

	// DefaultBatchTimeoutSec is the default timeout in seconds for batching.
	DefaultBatchTimeoutSec = 5

	// DefaultLogLevel is the default log level.
	DefaultLogLevel = "info"
)

// Config represents the configuration for the logger.
type Config struct {
	// ServiceName is the name of the service using the logger (required).
	ServiceName string

	// LogLevel is the logging level (debug, info, warn, error, fatal).
	LogLevel string

	// LogURL is the URL of the logging service (optional, logs to stdout if not provided).
	LogURL string

	// BatchSize is the number of logs to batch before sending.
	BatchSize int

	// BatchTimeoutSec is the timeout in seconds for batching.
	BatchTimeoutSec int

	// LogToStdout indicates whether to also log to stdout.
	LogToStdout bool

	// EnableStdout indicates whether stdout is enabled as fallback.
	EnableStdout bool
}
