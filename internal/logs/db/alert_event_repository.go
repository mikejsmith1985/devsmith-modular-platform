// Package logs_db provides database access and repository implementations for logs.
package logs_db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// AlertEventRepository handles database operations for alert events.
type AlertEventRepository struct {
	db *sql.DB
}

// NewAlertEventRepository creates a new AlertEventRepository.
func NewAlertEventRepository(db *sql.DB) *AlertEventRepository {
	return &AlertEventRepository{db: db}
}

// Create inserts a new alert event.
func (r *AlertEventRepository) Create(ctx context.Context, event *logs_models.AlertEvent) error {
	if event == nil {
		return errors.New("alert event cannot be nil")
	}

	query := `
		INSERT INTO logs.alert_events 
		(config_id, error_count, threshold_value, error_type, alert_sent, triggered_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	now := time.Now()
	event.TriggeredAt = now

	err := r.db.QueryRowContext(ctx, query,
		event.ConfigID,
		event.ErrorCount,
		event.ThresholdValue,
		event.ErrorType,
		event.AlertSent,
		event.TriggeredAt,
	).Scan(&event.ID)

	if err != nil {
		return fmt.Errorf("failed to create alert event: %w", err)
	}

	return nil
}

// GetByID retrieves an alert event by ID.
func (r *AlertEventRepository) GetByID(ctx context.Context, id int64) (*logs_models.AlertEvent, error) {
	query := `
		SELECT id, config_id, error_count, threshold_value, error_type, alert_sent, triggered_at
		FROM logs.alert_events
		WHERE id = $1
	`

	event := &logs_models.AlertEvent{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.ConfigID,
		&event.ErrorCount,
		&event.ThresholdValue,
		&event.ErrorType,
		&event.AlertSent,
		&event.TriggeredAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("alert event not found")
		}
		return nil, fmt.Errorf("failed to get alert event: %w", err)
	}

	return event, nil
}

// GetByConfigID retrieves all alert events for a specific config.
//
//nolint:dupl // Acceptable duplication: similar query pattern but different domain models
func (r *AlertEventRepository) GetByConfigID(ctx context.Context, configID int64) ([]logs_models.AlertEvent, error) {
	query := `
		SELECT id, config_id, error_count, threshold_value, error_type, alert_sent, triggered_at
		FROM logs.alert_events
		WHERE config_id = $1
		ORDER BY triggered_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert events: %w", err)
	}
	defer rows.Close() //nolint:errcheck // ignore close error in defer block

	var events []logs_models.AlertEvent
	for rows.Next() {
		event := logs_models.AlertEvent{}
		scanErr := rows.Scan(
			&event.ID,
			&event.ConfigID,
			&event.ErrorCount,
			&event.ThresholdValue,
			&event.ErrorType,
			&event.AlertSent,
			&event.TriggeredAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan alert event: %w", scanErr)
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return events, nil
}

// Delete removes an alert event by ID.
func (r *AlertEventRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM logs.alert_events WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert event: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("alert event with ID %d not found", id)
	}

	return nil
}
