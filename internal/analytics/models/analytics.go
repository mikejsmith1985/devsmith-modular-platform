// Package models contains data structures for analytics.
package models

import "time"

// MetricType represents the type of aggregated metric
type MetricType string

const (
	// ErrorFrequency represents the frequency of errors.
	ErrorFrequency MetricType = "error_frequency"
	// ServiceActivity represents a metric type for service activity.
	ServiceActivity MetricType = "service_activity"
)

// Aggregation represents a pre-computed metric for a time bucket. Fields are ordered to optimize memory alignment.
type Aggregation struct {
	TimeBucket time.Time  `json:"time_bucket" db:"time_bucket"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	MetricType MetricType `json:"metric_type" db:"metric_type"`
	Service    string     `json:"service" db:"service"`
	Value      float64    `json:"value" db:"value"`
	ID         int64      `json:"id" db:"id"`
}

// Trend represents a detected pattern over time
type Trend struct {
	StartTime  time.Time  `json:"start_time" db:"start_time"`
	EndTime    time.Time  `json:"end_time" db:"end_time"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	MetricType MetricType `json:"metric_type" db:"metric_type"`
	Service    string     `json:"service" db:"service"`
	Direction  string     `json:"direction" db:"direction"`
	ID         int64      `json:"id" db:"id"`
	Confidence float64    `json:"confidence" db:"confidence"`
}

// Anomaly represents an unusual spike or dip in the data.
// It includes details about the metric type, service, and the severity of the anomaly.
type Anomaly struct {
	TimeBucket time.Time  `json:"time_bucket" db:"time_bucket"`
	DetectedAt time.Time  `json:"detected_at" db:"detected_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	MetricType MetricType `json:"metric_type" db:"metric_type"`
	Service    string     `json:"service" db:"service"`
	Severity   string     `json:"severity" db:"severity"`
	ID         int64      `json:"id" db:"id"`
	Value      float64    `json:"value" db:"value"`
	ZScore     float64    `json:"z_score" db:"z_score"`
}

// LogEntry represents a log from logs.entries (READ-ONLY model)
type LogEntry struct {
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Service   string    `json:"service" db:"service"`
	Level     string    `json:"level" db:"level"`
	Message   string    `json:"message" db:"message"`
	ID        int64     `json:"id" db:"id"`
}

// TrendResponse is the API response for trend analysis
type TrendResponse struct {
	Trend      *TrendSummary `json:"trend,omitempty"`
	MetricType MetricType    `json:"metric_type"`
	Service    string        `json:"service"`
}

// AggregationDataPoint represents a single point in time-series data
type AggregationDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// TrendSummary provides high-level trend insights
type TrendSummary struct {
	Direction  string  `json:"direction"` // "increasing", "decreasing", "stable"
	Summary    string  `json:"summary"`   // Human-readable description
	Confidence float64 `json:"confidence"`
}

// TimeRange represents a time window for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// TopIssuesResponse returns most frequent errors/warnings
type TopIssuesResponse struct {
	TimeRange TimeRange   `json:"time_range"`
	Issues    []IssueItem `json:"issues"`
}

// IssueItem represents a frequent issue
// Added Value field to align with test expectations
type IssueItem struct {
	LastSeen time.Time
	Service  string
	Level    string
	Message  string
	Count    int
	Value    float64
}

// AnomalyResponse returns detected anomalies
type AnomalyResponse struct {
	TimeRange TimeRange `json:"time_range"`
	Anomalies []Anomaly `json:"anomalies"`
}

// TrendAnalysis represents the analysis of trends.
// Update TrendAnalysis struct to include required fields
type TrendAnalysis struct {
	Start            time.Time
	End              time.Time
	MetricType       MetricType
	Service          string
	Direction        string
	Change           float64
	PercentageChange float64
}

// Replace MinTime and MaxTime constants with variables
var (
	MinTime = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	MaxTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
)
