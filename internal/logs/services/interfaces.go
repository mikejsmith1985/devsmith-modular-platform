// Package services provides service interfaces for logs operations.
package services

import (
	"context"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// DashboardServiceInterface defines the contract for dashboard operations.
type DashboardServiceInterface interface {
	// GetDashboardStats returns aggregated statistics for the dashboard
	GetDashboardStats(ctx context.Context) (*models.DashboardStats, error)

	// GetServiceStats returns statistics for a specific service
	GetServiceStats(ctx context.Context, service string, timeRange time.Duration) (*models.LogStats, error)

	// GetTopErrors returns the top error messages within a time range
	GetTopErrors(ctx context.Context, limit int, timeRange time.Duration) ([]models.TopErrorMessage, error)

	// GetServiceHealth returns health status for all services
	GetServiceHealth(ctx context.Context) (map[string]*models.ServiceHealth, error)
}

// AlertServiceInterface defines the contract for alert operations.
type AlertServiceInterface interface {
	// CreateAlertConfig creates a new alert configuration
	CreateAlertConfig(ctx context.Context, config *models.AlertConfig) error

	// UpdateAlertConfig updates an existing alert configuration
	UpdateAlertConfig(ctx context.Context, config *models.AlertConfig) error

	// GetAlertConfig retrieves alert configuration for a service
	GetAlertConfig(ctx context.Context, service string) (*models.AlertConfig, error)

	// CheckThresholds checks if current log counts exceed alert thresholds
	CheckThresholds(ctx context.Context) ([]models.AlertThresholdViolation, error)

	// SendAlert sends an alert via email or webhook
	SendAlert(ctx context.Context, violation *models.AlertThresholdViolation) error
}

// LogAggregationServiceInterface defines the contract for log aggregation operations.
type LogAggregationServiceInterface interface {
	// AggregateLogsHourly performs hourly aggregation of logs
	AggregateLogsHourly(ctx context.Context) error

	// AggregateLogsDaily performs daily aggregation of logs
	AggregateLogsDaily(ctx context.Context) error

	// GetErrorRate calculates error rate for a service within a time window
	GetErrorRate(ctx context.Context, service string, start, end time.Time) (float64, error)

	// CountLogsByServiceAndLevel counts logs by service and level
	CountLogsByServiceAndLevel(ctx context.Context, service, level string, start, end time.Time) (int64, error)
}

// WebSocketRealtimeServiceInterface defines the contract for real-time WebSocket updates.
type WebSocketRealtimeServiceInterface interface {
	// RegisterConnection registers a new WebSocket connection
	RegisterConnection(ctx context.Context, connectionID string) error

	// UnregisterConnection removes a WebSocket connection
	UnregisterConnection(ctx context.Context, connectionID string) error

	// BroadcastStats broadcasts current statistics to all connected clients
	BroadcastStats(ctx context.Context, stats *models.DashboardStats) error

	// BroadcastAlert broadcasts an alert to all connected clients
	BroadcastAlert(ctx context.Context, violation *models.AlertThresholdViolation) error

	// GetConnectionCount returns the number of active connections
	GetConnectionCount(ctx context.Context) (int, error)
}
