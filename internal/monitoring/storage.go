package monitoring

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQLMetricsCollector implements MetricsCollector using PostgreSQL
type PostgreSQLMetricsCollector struct {
	db *pgxpool.Pool
}

// NewPostgreSQLMetricsCollector creates a new PostgreSQL-based metrics collector
func NewPostgreSQLMetricsCollector(db *pgxpool.Pool) *PostgreSQLMetricsCollector {
	return &PostgreSQLMetricsCollector{
		db: db,
	}
}

// RecordAPICall stores an API call metric in the database
func (c *PostgreSQLMetricsCollector) RecordAPICall(ctx context.Context, metrics APIMetrics) error {
	query := `
		INSERT INTO monitoring.api_metrics (
			timestamp, method, endpoint, status_code, response_time_ms,
			payload_size_bytes, user_id, error_type, error_message, service_name
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)`

	_, err := c.db.Exec(ctx, query,
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
		log.Printf("Failed to insert API metrics: %v", err)
		return fmt.Errorf("failed to record API call: %w", err)
	}

	return nil
}

// GetErrorRate calculates the error rate (4xx/5xx responses per minute) for the given time window
func (c *PostgreSQLMetricsCollector) GetErrorRate(ctx context.Context, window time.Duration) (float64, error) {
	since := time.Now().Add(-window)

	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status_code >= 400) as error_count,
			COUNT(*) as total_count
		FROM monitoring.api_metrics 
		WHERE timestamp >= $1`

	var errorCount, totalCount int64
	err := c.db.QueryRow(ctx, query, since).Scan(&errorCount, &totalCount)
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
func (c *PostgreSQLMetricsCollector) GetResponseTimes(ctx context.Context, window time.Duration) ([]float64, error) {
	since := time.Now().Add(-window)

	query := `
		SELECT response_time_ms 
		FROM monitoring.api_metrics 
		WHERE timestamp >= $1 
		AND status_code < 400
		ORDER BY timestamp DESC
		LIMIT 1000`

	rows, err := c.db.Query(ctx, query, since)
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

	return responseTimes, nil
}

// GetEndpointMetrics returns aggregated metrics for all endpoints
func (c *PostgreSQLMetricsCollector) GetEndpointMetrics(ctx context.Context, window time.Duration) ([]EndpointMetrics, error) {
	since := time.Now().Add(-window)

	query := `
		SELECT 
			endpoint,
			service_name,
			COUNT(*) as request_count,
			AVG(response_time_ms) as avg_response_time,
			PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY response_time_ms) as p95_response_time,
			COUNT(*) FILTER (WHERE status_code >= 400) as error_count,
			MAX(timestamp) as last_request
		FROM monitoring.api_metrics 
		WHERE timestamp >= $1
		GROUP BY endpoint, service_name
		ORDER BY request_count DESC`

	rows, err := c.db.Query(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint metrics: %w", err)
	}
	defer rows.Close()

	var metrics []EndpointMetrics
	for rows.Next() {
		var m EndpointMetrics
		var lastRequest *time.Time
		if err := rows.Scan(&m.Endpoint, &m.ServiceName, &m.RequestCount, &m.AvgResponseTime, &m.P95ResponseTime, &m.ErrorCount, &lastRequest); err != nil {
			return nil, fmt.Errorf("failed to scan endpoint metrics: %w", err)
		}
		if lastRequest != nil {
			m.LastRequest = *lastRequest
		}

		// Calculate error rate as percentage
		if m.RequestCount > 0 {
			m.ErrorRate = float64(m.ErrorCount) / float64(m.RequestCount) * 100
		}

		metrics = append(metrics, m)
	}

	return metrics, nil
}

// EndpointMetrics represents aggregated metrics for a single endpoint
type EndpointMetrics struct {
	Endpoint        string    `json:"endpoint"`
	ServiceName     string    `json:"service_name"`
	RequestCount    int64     `json:"request_count"`
	ErrorCount      int64     `json:"error_count"`
	ErrorRate       float64   `json:"error_rate_percent"`
	AvgResponseTime float64   `json:"avg_response_time_ms"`
	P95ResponseTime float64   `json:"p95_response_time_ms"`
	LastRequest     time.Time `json:"last_request"`
}

// InitializeSchema creates the necessary database schema for monitoring
func (c *PostgreSQLMetricsCollector) InitializeSchema(ctx context.Context) error {
	// Create monitoring schema
	schemaQuery := `CREATE SCHEMA IF NOT EXISTS monitoring`
	if _, err := c.db.Exec(ctx, schemaQuery); err != nil {
		return fmt.Errorf("failed to create monitoring schema: %w", err)
	}

	// Create api_metrics table
	tableQuery := `
		CREATE TABLE IF NOT EXISTS monitoring.api_metrics (
			id BIGSERIAL PRIMARY KEY,
			timestamp TIMESTAMPTZ NOT NULL,
			method VARCHAR(10) NOT NULL,
			endpoint VARCHAR(255) NOT NULL,
			status_code INT NOT NULL,
			response_time_ms BIGINT NOT NULL,
			payload_size_bytes BIGINT,
			user_id VARCHAR(50),
			error_type VARCHAR(100),
			error_message TEXT,
			service_name VARCHAR(50) NOT NULL
		)`
	if _, err := c.db.Exec(ctx, tableQuery); err != nil {
		return fmt.Errorf("failed to create api_metrics table: %w", err)
	}

	// Create indexes for performance
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_api_metrics_timestamp ON monitoring.api_metrics(timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_api_metrics_endpoint ON monitoring.api_metrics(endpoint, timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_api_metrics_status ON monitoring.api_metrics(status_code, timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_api_metrics_service ON monitoring.api_metrics(service_name, timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_api_metrics_user ON monitoring.api_metrics(user_id, timestamp DESC)`,
	}

	for _, indexQuery := range indexes {
		if _, err := c.db.Exec(ctx, indexQuery); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Create alerts table
	alertsTableQuery := `
		CREATE TABLE IF NOT EXISTS monitoring.alerts (
			id SERIAL PRIMARY KEY,
			alert_type VARCHAR(50) NOT NULL,
			severity VARCHAR(20) NOT NULL,
			message TEXT NOT NULL,
			threshold_value DECIMAL,
			actual_value DECIMAL,
			service_name VARCHAR(50),
			endpoint VARCHAR(255),
			triggered_at TIMESTAMPTZ DEFAULT NOW(),
			acknowledged_at TIMESTAMPTZ,
			resolved_at TIMESTAMPTZ,
			metadata JSONB
		)`
	if _, err := c.db.Exec(ctx, alertsTableQuery); err != nil {
		return fmt.Errorf("failed to create alerts table: %w", err)
	}

	// Create index for alerts
	alertIndexQuery := `CREATE INDEX IF NOT EXISTS idx_alerts_triggered ON monitoring.alerts(triggered_at DESC)`
	if _, err := c.db.Exec(ctx, alertIndexQuery); err != nil {
		return fmt.Errorf("failed to create alerts index: %w", err)
	}

	log.Println("Successfully initialized monitoring database schema")
	return nil
}

// CleanupOldMetrics removes metrics older than the specified retention period
func (c *PostgreSQLMetricsCollector) CleanupOldMetrics(ctx context.Context, retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	query := `DELETE FROM monitoring.api_metrics WHERE timestamp < $1`
	result, err := c.db.Exec(ctx, query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to cleanup old metrics: %w", err)
	}

	rowsDeleted := result.RowsAffected()
	log.Printf("Cleaned up %d old metric records older than %d days", rowsDeleted, retentionDays)

	return nil
}
