// Package models defines the data structures used in the logs service.
package models

import "time"

// LogEntry represents a log entry in the system.
// Fields are ordered to optimize memory alignment.
//
//nolint:govet // fieldalignment: organized by type for readability
type LogEntry struct {
	Context   *CorrelationContext `json:"context,omitempty"`
	Metadata  []byte              `json:"metadata"`
	Tags      []string            `json:"tags"`
	ID        int64               `json:"id"`
	UserID    int64               `json:"user_id"`
	CreatedAt time.Time           `json:"created_at"`
	Timestamp time.Time           `json:"timestamp"`
	Service   string              `json:"service"`
	Level     string              `json:"level"`
	Message   string              `json:"message"`
}

// LogStats represents aggregated statistics for logs in a time window.
// Used for dashboard display of real-time counts.
type LogStats struct { //nolint:govet // Struct alignment optimized for readability
	Timestamp    time.Time        `json:"timestamp" db:"timestamp"`
	Service      string           `json:"service" db:"service"`
	CountByLevel map[string]int64 `json:"count_by_level"`
	TotalCount   int64            `json:"total_count" db:"total_count"`
	ErrorRate    float64          `json:"error_rate" db:"error_rate"`
	ID           int64            `json:"id" db:"id"`
}

// AlertConfig represents alert threshold configuration.
type AlertConfig struct {
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
	Service                string    `json:"service" db:"service"`
	AlertEmail             string    `json:"alert_email" db:"alert_email"`
	AlertWebhookURL        string    `json:"alert_webhook_url" db:"alert_webhook_url"`
	ErrorThresholdPerMin   int       `json:"error_threshold_per_min" db:"error_threshold_per_min"`
	WarningThresholdPerMin int       `json:"warning_threshold_per_min" db:"warning_threshold_per_min"`
	ID                     int64     `json:"id" db:"id"`
	Enabled                bool      `json:"enabled" db:"enabled"`
}

// ServiceHealth represents the health status of a service.
type ServiceHealth struct {
	LastCheckedAt time.Time `json:"last_checked_at" db:"last_checked_at"`
	Service       string    `json:"service" db:"service"`
	Status        string    `json:"status" db:"status"` // OK, Warning, Error
	ErrorCount    int64     `json:"error_count" db:"error_count"`
	WarningCount  int64     `json:"warning_count" db:"warning_count"`
	InfoCount     int64     `json:"info_count" db:"info_count"`
	ID            int64     `json:"id" db:"id"`
}

// TopErrorMessage represents a frequently occurring error message.
type TopErrorMessage struct {
	FirstSeen time.Time `json:"first_seen" db:"first_seen"`
	LastSeen  time.Time `json:"last_seen" db:"last_seen"`
	Message   string    `json:"message" db:"message"`
	Service   string    `json:"service" db:"service"`
	Level     string    `json:"level" db:"level"`
	Count     int64     `json:"count" db:"count"`
}

// AlertThresholdViolation represents a violation of alert thresholds.
type AlertThresholdViolation struct {
	Timestamp      time.Time  `json:"timestamp" db:"timestamp"`
	AlertSentAt    *time.Time `json:"alert_sent_at" db:"alert_sent_at"`
	Service        string     `json:"service" db:"service"`
	Level          string     `json:"level" db:"level"`
	CurrentCount   int64      `json:"current_count" db:"current_count"`
	ThresholdValue int        `json:"threshold_value" db:"threshold_value"`
	ID             int64      `json:"id" db:"id"`
}

// DashboardStats represents complete dashboard data.
type DashboardStats struct {
	GeneratedAt      time.Time                 `json:"generated_at"`
	TimestampOne     time.Time                 `json:"timestamp_1h"`
	TimestampOneDay  time.Time                 `json:"timestamp_1d"`
	TimestampOneWeek time.Time                 `json:"timestamp_1w"`
	ServiceStats     map[string]*LogStats      `json:"service_stats"`
	ServiceHealth    map[string]*ServiceHealth `json:"service_health"`
	TopErrors        []TopErrorMessage         `json:"top_errors"`
	Violations       []AlertThresholdViolation `json:"violations"`
}

// ValidationError represents a validation error with aggregated metadata.
type ValidationError struct {
	LastOccurrence   time.Time `json:"last_occurrence" db:"last_occurrence"`
	ErrorType        string    `json:"error_type" db:"error_type"`
	Message          string    `json:"message" db:"message"`
	AffectedServices []string  `json:"affected_services"`
	Count            int64     `json:"count" db:"count"`
}

// ErrorTrend represents error counts over a time period.
type ErrorTrend struct {
	Timestamp       time.Time          `json:"timestamp" db:"timestamp"`
	ErrorCount      int64              `json:"error_count" db:"error_count"`
	ErrorRatePercent float64           `json:"error_rate_percent" db:"error_rate_percent"`
	ByType          map[string]int64   `json:"by_type"`
}

// AlertEvent represents a recorded alert event when thresholds are triggered.
type AlertEvent struct {
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	ConfigID         int64     `json:"config_id" db:"config_id"`
	ID               int64     `json:"id" db:"id"`
	ErrorCount       int       `json:"error_count" db:"error_count"`
	ThresholdValue   int       `json:"threshold_value" db:"threshold_value"`
	ErrorType        string    `json:"error_type" db:"error_type"`
	AlertSent        bool      `json:"alert_sent" db:"alert_sent"`
}

// LogExportOptions contains parameters for exporting logs.
type LogExportOptions struct {
	Format    string    `json:"format"`    // json or csv
	Service   string    `json:"service"`   // optional filter
	ErrorType string    `json:"error_type"` // optional filter
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}
