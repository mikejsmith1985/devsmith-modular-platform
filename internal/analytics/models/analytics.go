package models

import "time"

// MetricType represents the type of aggregated metric
type MetricType string

const (
	ErrorFrequency  MetricType = "error_frequency"
	ServiceActivity MetricType = "service_activity"
)

// Aggregation represents a pre-computed metric for a time bucket. Fields are ordered to optimize memory alignment.
type Aggregation struct {
	MetricType MetricType `json:"metric_type" db:"metric_type"`
	Service    string     `json:"service" db:"service"`
	Value      float64    `json:"value" db:"value"`
	TimeBucket time.Time  `json:"time_bucket" db:"time_bucket"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	ID         int64      `json:"id" db:"id"`
}

// Trend represents a detected pattern over time
type Trend struct {
	ID         int64      `json:"id" db:"id"`
	MetricType MetricType `json:"metric_type" db:"metric_type"`
	Service    string     `json:"service" db:"service"`
	Direction  string     `json:"direction" db:"direction"`
	Confidence float64    `json:"confidence" db:"confidence"`
	StartTime  time.Time  `json:"start_time" db:"start_time"`
	EndTime    time.Time  `json:"end_time" db:"end_time"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// Anomaly represents an unusual spike or dip in the data.
// It includes details about the metric type, service, and the severity of the anomaly.
type Anomaly struct {
	ID         int64      `json:"id" db:"id"`
	MetricType MetricType `json:"metric_type" db:"metric_type"`
	Service    string     `json:"service" db:"service"`
	Severity   string     `json:"severity" db:"severity"`
	TimeBucket time.Time  `json:"time_bucket" db:"time_bucket"`
	Value      float64    `json:"value" db:"value"`
	ZScore     float64    `json:"z_score" db:"z_score"`
	DetectedAt time.Time  `json:"detected_at" db:"detected_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// LogEntry represents a log from logs.entries (READ-ONLY model)
type LogEntry struct {
	ID        int64     `json:"id" db:"id"`
	Service   string    `json:"service" db:"service"`
	Level     string    `json:"level" db:"level"`
	Message   string    `json:"message" db:"message"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TrendResponse is the API response for trend analysis
type TrendResponse struct {
	MetricType MetricType    `json:"metric_type"`
	Service    string        `json:"service"`
	Trend      *TrendSummary `json:"trend,omitempty"`
}

// AggregationDataPoint represents a single point in time-series data
type AggregationDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// TrendSummary provides high-level trend insights
type TrendSummary struct {
	Direction  string  `json:"direction"` // "increasing", "decreasing", "stable"
	Confidence float64 `json:"confidence"`
	Summary    string  `json:"summary"` // Human-readable description
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
	Service  string    `json:"service"`
	Level    string    `json:"level"`
	Message  string    `json:"message"`
	Count    int       `json:"count"`
	LastSeen time.Time `json:"last_seen"`
	Value    float64   `json:"value"`
}

// AnomalyResponse returns detected anomalies
type AnomalyResponse struct {
	TimeRange TimeRange `json:"time_range"`
	Anomalies []Anomaly `json:"anomalies"`
}

// Update TrendAnalysis struct to include required fields
// TrendAnalysis provides insights into trends over time
type TrendAnalysis struct {
	MetricType       MetricType `json:"metric_type"`
	Service          string     `json:"service"`
	Start            time.Time  `json:"start"`
	End              time.Time  `json:"end"`
	Change           float64    `json:"change"`
	PercentageChange float64    `json:"percentage_change"`
	Direction        string     `json:"direction"` // "increasing", "decreasing", "stable"
}

// Replace MinTime and MaxTime constants with variables
var (
	MinTime = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	MaxTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
)
