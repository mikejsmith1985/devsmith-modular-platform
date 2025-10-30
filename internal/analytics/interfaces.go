// Package analytics provides interfaces and services for analytics operations.
package analytics

import (
	"context"
	"time"

	analytics_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
)

// AggregationRepositoryInterface defines the contract for aggregation storage operations
type AggregationRepositoryInterface interface {
	// Upsert creates or updates an aggregation for a time bucket
	Upsert(ctx context.Context, agg *analytics_models.Aggregation) error

	// FindByRange retrieves aggregations within a time range for a specific metric and service
	FindByRange(ctx context.Context, metricType analytics_models.MetricType, service string, start, end time.Time) ([]*analytics_models.Aggregation, error)

	// FindAllServices returns list of all services that have aggregations
	FindAllServices(ctx context.Context) ([]string, error)
}

// LogReaderInterface defines the contract for read-only log access
type LogReaderInterface interface {
	// CountByServiceAndLevel counts log entries by service and level within a time range
	CountByServiceAndLevel(ctx context.Context, service, level string, start, end time.Time) (int, error)

	// FindTopMessages finds most frequent log messages within a time range
	FindTopMessages(ctx context.Context, service, level string, start, end time.Time, limit int) ([]analytics_models.IssueItem, error)

	// FindAllServices returns list of all services that have logged
	FindAllServices(ctx context.Context) ([]string, error)
}

// AggregatorServiceInterface defines the contract for aggregation operations
type AggregatorServiceInterface interface {
	// RunHourlyAggregation performs hourly aggregation for all services
	RunHourlyAggregation(ctx context.Context) error
}

// TrendServiceInterface defines the contract for trend analysis operations.
// It includes methods for detecting trends and analyzing data over time.
type TrendServiceInterface interface {
	// AnalyzeTrend performs trend analysis for a specific metric and service
	AnalyzeTrend(ctx context.Context, metricType analytics_models.MetricType, service string, start, end time.Time) (*analytics_models.TrendAnalysis, error)
}

// AnomalyServiceInterface defines the contract for anomaly detection
type AnomalyServiceInterface interface {
	// DetectAnomalies identifies anomalies in aggregated metrics
	DetectAnomalies(ctx context.Context, metricType analytics_models.MetricType, service string, start, end time.Time) ([]analytics_models.Anomaly, error)
}

// TopIssuesServiceInterface defines the contract for top issues analysis
type TopIssuesServiceInterface interface {
	// GetTopIssues returns the most frequent errors/warnings within a time range
	GetTopIssues(ctx context.Context, service, level string, start, end time.Time, limit int) ([]analytics_models.IssueItem, error)
}

// ExportServiceInterface defines the contract for data export operations
type ExportServiceInterface interface {
	// ExportData exports aggregations in the specified format (csv, json)
	ExportData(ctx context.Context, metricType analytics_models.MetricType, service string, start, end time.Time, format string) ([]byte, error)
}
