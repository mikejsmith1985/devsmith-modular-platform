package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/healthcheck"
)

// HealthStorageService stores and retrieves health check results
type HealthStorageService struct {
	db *sql.DB
}

// NewHealthStorageService creates a new health storage service
func NewHealthStorageService(db *sql.DB) *HealthStorageService {
	return &HealthStorageService{db: db}
}

// HealthCheckSummary represents a stored health check summary
type HealthCheckSummary struct {
	ID            int       `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	OverallStatus string    `json:"overall_status"`
	DurationMS    int       `json:"duration_ms"`
	CheckCount    int       `json:"check_count"`
	PassedCount   int       `json:"passed_count"`
	WarnedCount   int       `json:"warned_count"`
	FailedCount   int       `json:"failed_count"`
	TriggeredBy   string    `json:"triggered_by"`
}

// TrendData represents performance trend data
type TrendData struct {
	Service    string      `json:"service"`
	TimePeriod string      `json:"time_period"`
	DataPoints []DataPoint `json:"data_points"`
	Average    float64     `json:"average"`
	Peak       int64       `json:"peak"`
}

// DataPoint represents a single trend data point
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     int64     `json:"value"`
	Status    string    `json:"status"`
}

// StoreHealthCheck stores a health check result in the database
func (s *HealthStorageService) StoreHealthCheck(ctx context.Context, report healthcheck.HealthReport, triggeredBy string) (int, error) {
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal report: %w", err)
	}

	var healthCheckID int
	query := `
		INSERT INTO logs.health_checks 
		(timestamp, overall_status, duration_ms, check_count, passed_count, warned_count, failed_count, report_json, triggered_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err = s.db.QueryRowContext(ctx,
		query,
		report.Timestamp,
		report.Status,
		report.Duration.Milliseconds(),
		report.Summary.Total,
		report.Summary.Passed,
		report.Summary.Warned,
		report.Summary.Failed,
		reportJSON,
		triggeredBy,
	).Scan(&healthCheckID)

	if err != nil {
		return 0, fmt.Errorf("failed to insert health check: %w", err)
	}

	// Store individual check details
	for _, check := range report.Checks {
		detailsJSON, _ := json.Marshal(check.Details)
		detailQuery := `
			INSERT INTO logs.health_check_details
			(health_check_id, check_name, status, message, error, duration_ms, details_json)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err := s.db.ExecContext(ctx,
			detailQuery,
			healthCheckID,
			check.Name,
			check.Status,
			check.Message,
			check.Error,
			check.Duration.Milliseconds(),
			detailsJSON,
		)
		if err != nil {
			// Log but don't fail if detail insert fails
			fmt.Printf("Warning: failed to insert check detail %s: %v\n", check.Name, err)
		}
	}

	return healthCheckID, nil
}

// GetRecentChecks retrieves the most recent health checks
func (s *HealthStorageService) GetRecentChecks(ctx context.Context, limit int) ([]HealthCheckSummary, error) {
	query := `
		SELECT id, timestamp, overall_status, duration_ms, check_count, passed_count, warned_count, failed_count, triggered_by
		FROM logs.health_checks
		ORDER BY timestamp DESC
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query health checks: %w", err)
	}
	defer rows.Close()

	var checks []HealthCheckSummary
	for rows.Next() {
		var check HealthCheckSummary
		err := rows.Scan(
			&check.ID,
			&check.Timestamp,
			&check.OverallStatus,
			&check.DurationMS,
			&check.CheckCount,
			&check.PassedCount,
			&check.WarnedCount,
			&check.FailedCount,
			&check.TriggeredBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan health check: %w", err)
		}
		checks = append(checks, check)
	}

	return checks, rows.Err()
}

// GetCheckHistory retrieves health checks from the last N hours
func (s *HealthStorageService) GetCheckHistory(ctx context.Context, hours int) ([]HealthCheckSummary, error) {
	query := `
		SELECT id, timestamp, overall_status, duration_ms, check_count, passed_count, warned_count, failed_count, triggered_by
		FROM logs.health_checks
		WHERE timestamp > NOW() - INTERVAL '1 hour' * $1
		ORDER BY timestamp DESC
	`

	rows, err := s.db.QueryContext(ctx, query, hours)
	if err != nil {
		return nil, fmt.Errorf("failed to query health check history: %w", err)
	}
	defer rows.Close()

	var checks []HealthCheckSummary
	for rows.Next() {
		var check HealthCheckSummary
		err := rows.Scan(
			&check.ID,
			&check.Timestamp,
			&check.OverallStatus,
			&check.DurationMS,
			&check.CheckCount,
			&check.PassedCount,
			&check.WarnedCount,
			&check.FailedCount,
			&check.TriggeredBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan health check: %w", err)
		}
		checks = append(checks, check)
	}

	return checks, rows.Err()
}

// GetTrendData retrieves performance trend data for a specific check
func (s *HealthStorageService) GetTrendData(ctx context.Context, checkName string, hours int) (TrendData, error) {
	trend := TrendData{
		Service:    checkName,
		TimePeriod: fmt.Sprintf("Last %d hours", hours),
		DataPoints: []DataPoint{},
	}

	query := `
		SELECT hcd.created_at, 
		       CAST(COALESCE(hcd.duration_ms, 0) AS BIGINT) as duration_ms,
		       hcd.status
		FROM logs.health_check_details hcd
		WHERE hcd.check_name = $1
		  AND hcd.created_at > NOW() - INTERVAL '1 hour' * $2
		ORDER BY hcd.created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, checkName, hours)
	if err != nil {
		return trend, fmt.Errorf("failed to query trend data: %w", err)
	}
	defer rows.Close()

	var total int64
	var count int64

	for rows.Next() {
		var timestamp time.Time
		var duration int64
		var status string

		err := rows.Scan(&timestamp, &duration, &status)
		if err != nil {
			return trend, fmt.Errorf("failed to scan trend data: %w", err)
		}

		trend.DataPoints = append(trend.DataPoints, DataPoint{
			Timestamp: timestamp,
			Value:     duration,
			Status:    status,
		})

		total += duration
		count++
	}

	if count > 0 {
		trend.Average = float64(total) / float64(count)
	}

	// Find peak
	for _, dp := range trend.DataPoints {
		if dp.Value > trend.Peak {
			trend.Peak = dp.Value
		}
	}

	return trend, rows.Err()
}

// GetHealthCheckByID retrieves a specific health check with all details
func (s *HealthStorageService) GetHealthCheckByID(ctx context.Context, id int) (*healthcheck.HealthReport, error) {
	query := `SELECT report_json FROM logs.health_checks WHERE id = $1`

	var reportJSON []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(&reportJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to query health check: %w", err)
	}

	var report healthcheck.HealthReport
	err = json.Unmarshal(reportJSON, &report)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal health check: %w", err)
	}

	return &report, nil
}

// CleanupOldChecks removes health checks older than the retention period
func (s *HealthStorageService) CleanupOldChecks(ctx context.Context, retentionDays int) (int64, error) {
	query := `
		DELETE FROM logs.health_checks
		WHERE created_at < NOW() - INTERVAL '1 day' * $1
	`

	result, err := s.db.ExecContext(ctx, query, retentionDays)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old checks: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
