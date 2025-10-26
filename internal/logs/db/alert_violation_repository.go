// Package db provides database access and repository implementations for logs.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// AlertViolationRepository handles database operations for alert threshold violations.
type AlertViolationRepository struct {
	db *sql.DB
}

// NewAlertViolationRepository creates a new AlertViolationRepository.
func NewAlertViolationRepository(db *sql.DB) *AlertViolationRepository {
	return &AlertViolationRepository{db: db}
}

// Create inserts a new alert violation.
func (r *AlertViolationRepository) Create(ctx context.Context, violation *models.AlertThresholdViolation) error {
	if violation == nil {
		return errors.New("violation cannot be nil")
	}

	query := `
		INSERT INTO logs.alert_violations
		(service, level, current_count, threshold_value, timestamp, alert_sent_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		violation.Service,
		violation.Level,
		violation.CurrentCount,
		violation.ThresholdValue,
		violation.Timestamp,
		violation.AlertSentAt,
	).Scan(&violation.ID)

	if err != nil {
		return fmt.Errorf("failed to create violation: %w", err)
	}

	return nil
}

// UpdateAlertSent marks a violation as having an alert sent.
func (r *AlertViolationRepository) UpdateAlertSent(ctx context.Context, id int64) error {
	query := `
		UPDATE logs.alert_violations
		SET alert_sent_at = $1
		WHERE id = $2
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to update violation: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("violation with ID %d not found", id)
	}

	return nil
}

// GetByService retrieves violations for a service within a time range.
func (r *AlertViolationRepository) GetByService(ctx context.Context, service string, start, end time.Time) ([]models.AlertThresholdViolation, error) {
	query := `
		SELECT id, service, level, current_count, threshold_value, timestamp, alert_sent_at
		FROM logs.alert_violations
		WHERE service = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, service, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query violations: %w", err)
	}
	defer closeRows(rows)

	var violations []models.AlertThresholdViolation
	for rows.Next() {
		v := models.AlertThresholdViolation{}
		scanErr := rows.Scan(
			&v.ID,
			&v.Service,
			&v.Level,
			&v.CurrentCount,
			&v.ThresholdValue,
			&v.Timestamp,
			&v.AlertSentAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan violation: %w", scanErr)
		}
		violations = append(violations, v)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return violations, nil
}

// GetUnsent retrieves violations where alert has not been sent.
func (r *AlertViolationRepository) GetUnsent(ctx context.Context) ([]models.AlertThresholdViolation, error) {
	query := `
		SELECT id, service, level, current_count, threshold_value, timestamp, alert_sent_at
		FROM logs.alert_violations
		WHERE alert_sent_at IS NULL
		ORDER BY timestamp DESC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unsent violations: %w", err)
	}
	defer closeRows(rows)

	var violations []models.AlertThresholdViolation
	for rows.Next() {
		v := models.AlertThresholdViolation{}
		scanErr := rows.Scan(
			&v.ID,
			&v.Service,
			&v.Level,
			&v.CurrentCount,
			&v.ThresholdValue,
			&v.Timestamp,
			&v.AlertSentAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan violation: %w", scanErr)
		}
		violations = append(violations, v)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return violations, nil
}

// GetRecent retrieves recent alert violations ordered by timestamp.
//
//nolint:dupl // Acceptable duplication: similar query pattern but different domain models
func (r *AlertViolationRepository) GetRecent(ctx context.Context, limit int) ([]models.AlertThresholdViolation, error) {
	query := `
		SELECT id, service, level, current_count, threshold_value, timestamp, alert_sent_at
		FROM logs.alert_violations
		ORDER BY timestamp DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent violations: %w", err)
	}
	defer closeRows(rows)

	var violations []models.AlertThresholdViolation
	for rows.Next() {
		v := models.AlertThresholdViolation{}
		scanErr := rows.Scan(
			&v.ID,
			&v.Service,
			&v.Level,
			&v.CurrentCount,
			&v.ThresholdValue,
			&v.Timestamp,
			&v.AlertSentAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan violation: %w", scanErr)
		}
		violations = append(violations, v)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return violations, nil
}
