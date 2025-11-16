package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SQLMetricsCollector implements MetricsCollector using database/sql
type SQLMetricsCollector struct {
	db *sql.DB
}

// NewSQLMetricsCollector creates a new SQL-based metrics collector
func NewSQLMetricsCollector(db *sql.DB) *SQLMetricsCollector {
	return &SQLMetricsCollector{
		db: db,
	}
}

// RecordAPICall stores an API call metric in the database
func (c *SQLMetricsCollector) RecordAPICall(ctx context.Context, metrics APIMetrics) error {
	query := `
		INSERT INTO monitoring.api_metrics (
			timestamp, method, endpoint, status_code, response_time_ms,
			payload_size_bytes, user_id, error_type, error_message, service_name
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)`

	_, err := c.db.ExecContext(ctx, query,
		metrics.Timestamp,
		metrics.Method,
		metrics.Endpoint,
		metrics.StatusCode,
		metrics.ResponseTime,
		metrics.PayloadSize,
		metrics.UserID,
		metrics.ErrorType,
		metrics.ErrorMessage,
		metrics.ServiceName,
	)

	if err != nil {
		return fmt.Errorf("failed to record API call: %w", err)
	}

	return nil
}

// GetErrorRate calculates the error rate (4xx/5xx responses per minute) for the given time window
func (c *SQLMetricsCollector) GetErrorRate(ctx context.Context, window time.Duration) (float64, error) {
	since := time.Now().Add(-window)

	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status_code >= 400) as error_count,
			COUNT(*) as total_count
		FROM monitoring.api_metrics 
		WHERE timestamp >= $1`

	var errorCount, totalCount int64
	err := c.db.QueryRowContext(ctx, query, since).Scan(&errorCount, &totalCount)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate error rate: %w", err)
	}

	if totalCount == 0 {
		return 0, nil
	}

	// Convert to errors per minute
	windowMinutes := window.Minutes()
	errorRate := (float64(errorCount) / windowMinutes)

	return errorRate, nil
}

// GetResponseTimes retrieves response times for the given time window
func (c *SQLMetricsCollector) GetResponseTimes(ctx context.Context, window time.Duration) ([]float64, error) {
	since := time.Now().Add(-window)

	query := `
		SELECT response_time_ms 
		FROM monitoring.api_metrics 
		WHERE timestamp >= $1 
		AND status_code < 400
		ORDER BY timestamp DESC
		LIMIT 1000`

	rows, err := c.db.QueryContext(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get response times: %w", err)
	}
	defer rows.Close()

	var responseTimes []float64
	for rows.Next() {
		var responseTime int64
		if err := rows.Scan(&responseTime); err != nil {
			return nil, fmt.Errorf("failed to scan response time: %w", err)
		}
		responseTimes = append(responseTimes, float64(responseTime))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating response times: %w", err)
	}

	return responseTimes, nil
}
