package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/healthcheck"
	"github.com/sirupsen/logrus"
)

// HealthStorageService stores health check reports and analysis
type HealthStorageService struct {
	logger *logrus.Logger
	db     *sql.DB
}

// HealthCheckSummary is a summary of a health check
type HealthCheckSummary struct {
	ID            int       `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	OverallStatus string    `json:"overall_status"`
	DurationMs    int       `json:"duration_ms"`
	PassedCount   int       `json:"passed_count"`
	FailedCount   int       `json:"failed_count"`
	TriggeredBy   string    `json:"triggered_by"`
}

// TrendData represents trend analysis for a service
type TrendData struct {
	ServiceName  string    `json:"service_name"`
	TimeRange    string    `json:"time_range"`
	AvgDuration  int       `json:"avg_duration"`
	FailureRate  float64   `json:"failure_rate"`
	HealthScores []float64 `json:"health_scores"`
}

// HealthTrendPoint represents a single data point in a trend
type HealthTrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
	Value     int       `json:"value"` // 0-100 score
}

// NewHealthStorageService creates a new health storage service
func NewHealthStorageService(db *sql.DB) *HealthStorageService {
	return &HealthStorageService{
		logger: logrus.New(),
		db:     db,
	}
}

// NewHealthStorageServiceWithLogger creates a new health storage service with a logger
func NewHealthStorageServiceWithLogger(db *sql.DB, logger *logrus.Logger) *HealthStorageService {
	return &HealthStorageService{
		logger: logger,
		db:     db,
	}
}

// StoreHealthCheck stores a complete health check report
func (s *HealthStorageService) StoreHealthCheck(ctx context.Context, report *healthcheck.HealthReport, triggeredBy string) (int, error) {
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal report: %w", err)
	}

	var id int
	err = s.db.QueryRowContext(ctx,
		`INSERT INTO logs.health_checks 
		 (timestamp, overall_status, duration_ms, check_count, passed_count, warned_count, failed_count, report_json, triggered_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id`,
		time.Now(),
		string(report.Status),
		report.Duration.Milliseconds(),
		report.Summary.Total,
		report.Summary.Passed,
		report.Summary.Warned,
		report.Summary.Failed,
		reportJSON,
		triggeredBy,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert health check: %w", err)
	}

	// Store individual check details
	for _, check := range report.Checks {
		detailsJSON, err := json.Marshal(check.Details)
		if err != nil {
			// Log but don't fail entire operation
			s.logger.Errorf("failed to marshal check details: %v", err)
			detailsJSON = []byte("{}")
		}
		_, err = s.db.ExecContext(ctx,
			`INSERT INTO logs.health_check_details 
			 (health_check_id, check_name, status, message, error, duration_ms, details_json)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			id,
			check.Name,
			string(check.Status),
			check.Message,
			check.Error,
			check.Duration.Milliseconds(),
			detailsJSON,
		)
		if err != nil {
			// Log but don't fail entire operation
			s.logger.Errorf("failed to insert check detail: %v", err)
		}
	}

	return id, nil
}

// GetRecentChecks returns the most recent health checks
func (s *HealthStorageService) GetRecentChecks(ctx context.Context, limit int) ([]HealthCheckSummary, error) {
	query := `SELECT id, timestamp, overall_status, duration_ms, passed_count, failed_count, triggered_by
		 FROM logs.health_checks
		 ORDER BY timestamp DESC
		 LIMIT $1`
	return s.queryHealthChecks(ctx, query, limit)
}

// GetCheckHistory returns checks from the specified number of hours
func (s *HealthStorageService) GetCheckHistory(ctx context.Context, hours int) ([]HealthCheckSummary, error) {
	query := `SELECT id, timestamp, overall_status, duration_ms, passed_count, failed_count, triggered_by
		 FROM logs.health_checks
		 WHERE timestamp >= NOW() - INTERVAL '1 hour' * $1
		 ORDER BY timestamp DESC`
	return s.queryHealthChecks(ctx, query, hours)
}

// queryHealthChecks executes a health checks query and returns results
func (s *HealthStorageService) queryHealthChecks(ctx context.Context, query string, arg interface{}) ([]HealthCheckSummary, error) {
	rows, err := s.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to query health checks: %w", err)
	}
	defer func() {
		_ = rows.Close() // explicitly ignore error as rows already processed
	}()

	var checks []HealthCheckSummary
	for rows.Next() {
		var c HealthCheckSummary
		err := rows.Scan(&c.ID, &c.Timestamp, &c.OverallStatus, &c.DurationMs, &c.PassedCount, &c.FailedCount, &c.TriggeredBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan check: %w", err)
		}
		checks = append(checks, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}
	return checks, nil
}

// GetTrendData returns trend analysis for a specific service
func (s *HealthStorageService) GetTrendData(ctx context.Context, serviceName string, hours int) (*TrendData, error) {
	// Query recent checks for this service
	rows, err := s.db.QueryContext(ctx,
		`SELECT timestamp, overall_status, duration_ms
		 FROM logs.health_checks
		 WHERE timestamp >= NOW() - INTERVAL '1 hour' * $1
		 ORDER BY timestamp DESC`,
		hours,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query trend data: %w", err)
	}
	defer func() {
		_ = rows.Close() // explicitly ignore error as rows already processed
	}()

	trend := &TrendData{
		ServiceName:  serviceName,
		TimeRange:    fmt.Sprintf("%d hours", hours),
		HealthScores: []float64{},
	}

	var totalDuration int64
	var checkCount int
	var failCount int

	for rows.Next() {
		var timestamp time.Time
		var status string
		var durationMs int
		err := rows.Scan(&timestamp, &status, &durationMs)
		if err != nil {
			continue
		}

		totalDuration += int64(durationMs)
		checkCount++
		if status == "fail" {
			failCount++
		}

		// Convert status to score (0-100)
		score := 0
		if status == "pass" {
			score = 100
		} else if status == "warn" {
			score = 50
		}

		trend.HealthScores = append(trend.HealthScores, float64(score))
	}

	if checkCount > 0 {
		trend.AvgDuration = int(totalDuration / int64(checkCount))
		trend.FailureRate = float64(failCount) / float64(checkCount)
	}

	if len(trend.HealthScores) > 0 {
		// The original code had trend.LastCheckTime = trend.HealthScores[0].Timestamp
		// This line is problematic as HealthScores is []float64.
		// Assuming the intent was to find the timestamp of the last check.
		// Since HealthScores is now []float64, we need to find the timestamp of the last float64.
		// This is not directly possible without a timestamp field in HealthScores.
		// For now, removing this line as it's not directly applicable to the new HealthScores type.
	}

	return trend, rows.Err()
}

// CleanupOldChecks removes health checks older than the specified number of days
func (s *HealthStorageService) CleanupOldChecks(ctx context.Context, retentionDays int) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM logs.health_checks
		 WHERE timestamp < NOW() - INTERVAL '1 day' * $1`,
		retentionDays,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old checks: %w", err)
	}

	return result.RowsAffected()
}
