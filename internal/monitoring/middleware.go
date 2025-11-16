package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// APIMetrics represents a single API call measurement
type APIMetrics struct {
	ID           int64     `json:"id" db:"id"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	Method       string    `json:"method" db:"method"`
	Endpoint     string    `json:"endpoint" db:"endpoint"`
	StatusCode   int       `json:"status_code" db:"status_code"`
	ResponseTime int64     `json:"response_time_ms" db:"response_time_ms"` // milliseconds
	PayloadSize  int64     `json:"payload_size_bytes" db:"payload_size_bytes"`
	UserID       string    `json:"user_id" db:"user_id"`
	ErrorType    string    `json:"error_type" db:"error_type"`
	ErrorMessage string    `json:"error_message" db:"error_message"`
	ServiceName  string    `json:"service_name" db:"service_name"`
}

// MetricsCollector handles collecting and storing API metrics
type MetricsCollector interface {
	RecordAPICall(ctx context.Context, metrics APIMetrics) error
	GetErrorRate(ctx context.Context, window time.Duration) (float64, error)
	GetResponseTimes(ctx context.Context, window time.Duration) ([]float64, error)
}

// MetricsMiddleware creates a Gin middleware that collects API metrics
func MetricsMiddleware(collector MetricsCollector, serviceName string) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Extract user ID from context if available
		userID := ""
		if uid, exists := param.Keys["user_id"]; exists {
			if uidStr, ok := uid.(string); ok {
				userID = uidStr
			}
		}

		// Determine error type and message for non-2xx responses
		errorType := ""
		errorMessage := ""
		if param.StatusCode >= 400 {
			switch {
			case param.StatusCode >= 500:
				errorType = "server_error"
				errorMessage = "Internal server error"
			case param.StatusCode == 400:
				errorType = "client_error"
				errorMessage = "Bad request - likely payload validation failure"
			case param.StatusCode == 401:
				errorType = "auth_error"
				errorMessage = "Authentication required"
			case param.StatusCode == 404:
				errorType = "not_found"
				errorMessage = "Resource not found"
			default:
				errorType = "client_error"
				errorMessage = "Client request error"
			}
		}

		// Create metrics record
		metrics := APIMetrics{
			Timestamp:    param.TimeStamp,
			Method:       param.Method,
			Endpoint:     param.Path,
			StatusCode:   param.StatusCode,
			ResponseTime: param.Latency.Milliseconds(),
			PayloadSize:  int64(param.BodySize),
			UserID:       userID,
			ErrorType:    errorType,
			ErrorMessage: errorMessage,
			ServiceName:  serviceName,
		}

		// Store metrics asynchronously to avoid blocking requests
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := collector.RecordAPICall(ctx, metrics); err != nil {
				log.Printf("Failed to record API metrics: %v", err)
			}
		}()

		// Return formatted log message
		return fmt.Sprintf("%s - [%s] \"%s %s %s\" %d %s %s\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.ErrorMessage,
		)
	})
}

// PayloadValidationMetrics tracks specific validation failures
type PayloadValidationMetrics struct {
	Endpoint        string    `json:"endpoint"`
	InvalidFields   []string  `json:"invalid_fields"`
	ExtraFields     []string  `json:"extra_fields"` // Like session_id
	MissingFields   []string  `json:"missing_fields"`
	OriginalPayload string    `json:"original_payload"`
	Timestamp       time.Time `json:"timestamp"`
}

// RecordPayloadValidationFailure logs detailed validation failure information
func RecordPayloadValidationFailure(c *gin.Context, invalidFields, extraFields, missingFields []string) {
	// Extract original payload for analysis
	var originalPayload string
	if c.Request.Body != nil {
		// Note: Body can only be read once, so this should be called after validation
		// In practice, we'd need to capture this in the validation middleware itself
		originalPayload = "Body already consumed - implement body capture in validation middleware"
	}

	validationMetrics := PayloadValidationMetrics{
		Endpoint:        c.FullPath(),
		InvalidFields:   invalidFields,
		ExtraFields:     extraFields,
		MissingFields:   missingFields,
		OriginalPayload: originalPayload,
		Timestamp:       time.Now(),
	}

	// Log structured validation failure data
	validationJSON, _ := json.MarshalIndent(validationMetrics, "", "  ")
	log.Printf("PAYLOAD_VALIDATION_FAILURE: %s", string(validationJSON))

	// Store in context for middleware to capture
	c.Set("validation_failure", validationMetrics)
}

// AlertThresholds defines when to trigger alerts
type AlertThresholds struct {
	APIErrorRate    float64 // errors per minute
	ResponseTimeP95 int64   // milliseconds
	ServiceDown     int     // consecutive failures
}

// DefaultAlertThresholds returns sensible default alert thresholds
func DefaultAlertThresholds() AlertThresholds {
	return AlertThresholds{
		APIErrorRate:    5.0,  // 5 errors per minute triggers alert
		ResponseTimeP95: 2000, // 2 second P95 response time triggers alert
		ServiceDown:     2,    // 2 consecutive health check failures
	}
}

// MonitoringConfig holds all monitoring configuration
type MonitoringConfig struct {
	ServiceName    string
	Thresholds     AlertThresholds
	EnableRealTime bool
	RetentionDays  int
}

// DefaultMonitoringConfig returns default monitoring configuration
func DefaultMonitoringConfig(serviceName string) MonitoringConfig {
	return MonitoringConfig{
		ServiceName:    serviceName,
		Thresholds:     DefaultAlertThresholds(),
		EnableRealTime: true,
		RetentionDays:  30,
	}
}
