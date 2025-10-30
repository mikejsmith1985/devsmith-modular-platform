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

// AlertConfigRepository handles database operations for alert configurations.
type AlertConfigRepository struct {
	db *sql.DB
}

// NewAlertConfigRepository creates a new AlertConfigRepository.
func NewAlertConfigRepository(db *sql.DB) *AlertConfigRepository {
	return &AlertConfigRepository{db: db}
}

// Create inserts a new alert configuration.
func (r *AlertConfigRepository) Create(ctx context.Context, config *logs_models.AlertConfig) error {
	if config == nil {
		return errors.New("alert config cannot be nil")
	}

	query := `
		INSERT INTO logs.alert_configs 
		(service, error_threshold_per_min, warning_threshold_per_min, alert_email, alert_webhook_url, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		config.Service,
		config.ErrorThresholdPerMin,
		config.WarningThresholdPerMin,
		config.AlertEmail,
		config.AlertWebhookURL,
		config.Enabled,
		config.CreatedAt,
		config.UpdatedAt,
	).Scan(&config.ID)

	if err != nil {
		return fmt.Errorf("failed to create alert config: %w", err)
	}

	return nil
}

// GetByService retrieves an alert configuration by service name.
func (r *AlertConfigRepository) GetByService(ctx context.Context, service string) (*logs_models.AlertConfig, error) {
	query := `
		SELECT id, service, error_threshold_per_min, warning_threshold_per_min, 
		       alert_email, alert_webhook_url, enabled, created_at, updated_at
		FROM logs.alert_configs
		WHERE service = $1
	`

	config := &logs_models.AlertConfig{}
	err := r.db.QueryRowContext(ctx, query, service).Scan(
		&config.ID,
		&config.Service,
		&config.ErrorThresholdPerMin,
		&config.WarningThresholdPerMin,
		&config.AlertEmail,
		&config.AlertWebhookURL,
		&config.Enabled,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("alert config not found for service %s", service)
		}
		return nil, fmt.Errorf("failed to get alert config: %w", err)
	}

	return config, nil
}

// Update modifies an existing alert configuration.
func (r *AlertConfigRepository) Update(ctx context.Context, config *logs_models.AlertConfig) error {
	if config == nil || config.ID == 0 {
		return errors.New("alert config must have valid ID")
	}

	query := `
		UPDATE logs.alert_configs
		SET error_threshold_per_min = $1,
		    warning_threshold_per_min = $2,
		    alert_email = $3,
		    alert_webhook_url = $4,
		    enabled = $5,
		    updated_at = $6
		WHERE id = $7
	`

	now := time.Now()
	config.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		config.ErrorThresholdPerMin,
		config.WarningThresholdPerMin,
		config.AlertEmail,
		config.AlertWebhookURL,
		config.Enabled,
		config.UpdatedAt,
		config.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update alert config: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("alert config with ID %d not found", config.ID)
	}

	return nil
}

// GetAll retrieves all alert configurations.
func (r *AlertConfigRepository) GetAll(ctx context.Context) ([]logs_models.AlertConfig, error) {
	query := `
		SELECT id, service, error_threshold_per_min, warning_threshold_per_min,
		       alert_email, alert_webhook_url, enabled, created_at, updated_at
		FROM logs.alert_configs
		ORDER BY service
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert configs: %w", err)
	}
	defer rows.Close() //nolint:errcheck // close error ignored in defer

	var configs []logs_models.AlertConfig
	for rows.Next() {
		config := logs_models.AlertConfig{}
		scanErr := rows.Scan(
			&config.ID,
			&config.Service,
			&config.ErrorThresholdPerMin,
			&config.WarningThresholdPerMin,
			&config.AlertEmail,
			&config.AlertWebhookURL,
			&config.Enabled,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan alert config: %w", scanErr)
		}
		configs = append(configs, config)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return configs, nil
}

// Delete removes an alert configuration by ID.
func (r *AlertConfigRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM logs.alert_configs WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert config: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("alert config with ID %d not found", id)
	}

	return nil
}
