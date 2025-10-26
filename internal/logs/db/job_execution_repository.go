// Package db provides database access and repository implementations for logs.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// JobExecution represents a background job execution record.
type JobExecution struct { //nolint:govet // struct alignment optimized for readability
	ID           int64
	StartedAt    time.Time
	CreatedAt    time.Time
	JobType      string
	Status       string // pending, running, success, failed
	ErrorMessage *string
	CompletedAt  *time.Time
}

// JobExecutionRepository handles database operations for job execution history.
type JobExecutionRepository struct {
	db *sql.DB
}

// NewJobExecutionRepository creates a new JobExecutionRepository.
func NewJobExecutionRepository(db *sql.DB) *JobExecutionRepository {
	return &JobExecutionRepository{db: db}
}

// Create inserts a new job execution record.
func (r *JobExecutionRepository) Create(ctx context.Context, jobType string) (int64, error) {
	if jobType == "" {
		return 0, errors.New("job type cannot be empty")
	}

	query := `
		INSERT INTO logs.job_executions
		(job_type, started_at, status, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	now := time.Now()
	var id int64

	err := r.db.QueryRowContext(ctx, query, jobType, now, "running", now).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create job execution: %w", err)
	}

	return id, nil
}

// MarkSuccess marks a job execution as successfully completed.
func (r *JobExecutionRepository) MarkSuccess(ctx context.Context, id int64) error {
	query := `
		UPDATE logs.job_executions
		SET status = $1, completed_at = $2
		WHERE id = $3
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, "success", now, id)
	if err != nil {
		return fmt.Errorf("failed to update job execution: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("job execution with ID %d not found", id)
	}

	return nil
}

// MarkFailure marks a job execution as failed with an error message.
func (r *JobExecutionRepository) MarkFailure(ctx context.Context, id int64, errMsg string) error {
	query := `
		UPDATE logs.job_executions
		SET status = $1, completed_at = $2, error_message = $3
		WHERE id = $4
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, "failed", now, errMsg, id)
	if err != nil {
		return fmt.Errorf("failed to update job execution: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("job execution with ID %d not found", id)
	}

	return nil
}

// GetByJobType retrieves execution history for a specific job type.
func (r *JobExecutionRepository) GetByJobType(ctx context.Context, jobType string, limit int) ([]JobExecution, error) {
	query := `
		SELECT id, job_type, started_at, completed_at, status, error_message, created_at
		FROM logs.job_executions
		WHERE job_type = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, jobType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query job executions: %w", err)
	}
	defer rows.Close() //nolint:errcheck // close error ignored in defer

	var executions []JobExecution
	for rows.Next() {
		e := JobExecution{}
		scanErr := rows.Scan(
			&e.ID,
			&e.JobType,
			&e.StartedAt,
			&e.CompletedAt,
			&e.Status,
			&e.ErrorMessage,
			&e.CreatedAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan job execution: %w", scanErr)
		}
		executions = append(executions, e)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return executions, nil
}

// GetRecent retrieves the most recent job executions across all types.
func (r *JobExecutionRepository) GetRecent(ctx context.Context, limit int) ([]JobExecution, error) {
	query := `
		SELECT id, job_type, started_at, completed_at, status, error_message, created_at
		FROM logs.job_executions
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent job executions: %w", err)
	}
	defer rows.Close() //nolint:errcheck // close error ignored in defer

	var executions []JobExecution
	for rows.Next() {
		e := JobExecution{}
		scanErr := rows.Scan(
			&e.ID,
			&e.JobType,
			&e.StartedAt,
			&e.CompletedAt,
			&e.Status,
			&e.ErrorMessage,
			&e.CreatedAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan job execution: %w", scanErr)
		}
		executions = append(executions, e)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return executions, nil
}

// GetFailures retrieves failed job executions within a time range.
func (r *JobExecutionRepository) GetFailures(ctx context.Context, start, end time.Time) ([]JobExecution, error) {
	query := `
		SELECT id, job_type, started_at, completed_at, status, error_message, created_at
		FROM logs.job_executions
		WHERE status = $1 AND created_at BETWEEN $2 AND $3
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, "failed", start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query job failures: %w", err)
	}
	defer rows.Close() //nolint:errcheck // close error ignored in defer

	var executions []JobExecution
	for rows.Next() {
		e := JobExecution{}
		scanErr := rows.Scan(
			&e.ID,
			&e.JobType,
			&e.StartedAt,
			&e.CompletedAt,
			&e.Status,
			&e.ErrorMessage,
			&e.CreatedAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan job execution: %w", scanErr)
		}
		executions = append(executions, e)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return executions, nil
}

// DeleteOlder removes job executions older than the specified date.
func (r *JobExecutionRepository) DeleteOlder(ctx context.Context, before time.Time) (int64, error) {
	query := "DELETE FROM logs.job_executions WHERE created_at < $1"

	result, err := r.db.ExecContext(ctx, query, before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old job executions: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to check rows affected: %w", err)
	}

	return rows, nil
}
