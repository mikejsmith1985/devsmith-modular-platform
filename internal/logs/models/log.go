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
