// Package db provides database access for log entries.
package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// LogEntryRepository handles CRUD operations for log entries.
type LogEntryRepository struct {
	db *sql.DB
}

// NewLogEntryRepository creates a new LogEntryRepository with the given database connection.
func NewLogEntryRepository(db *sql.DB) *LogEntryRepository {
	return &LogEntryRepository{db: db}
}

// queryLogEntries executes a query and returns scanned log entries.
func (r *LogEntryRepository) queryLogEntries(ctx context.Context, query string, args ...interface{}) ([]models.LogEntry, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("db: failed to close rows: %w", closeErr)
		}
	}()

	var entries []models.LogEntry
	for rows.Next() {
		var entry models.LogEntry
		scanErr := rows.Scan(&entry.ID, &entry.UserID, &entry.Service, &entry.Level, &entry.Message, &entry.Metadata, &entry.CreatedAt)
		if scanErr != nil {
			return nil, fmt.Errorf("db: failed to scan log entry: %w", scanErr)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db: rows iteration error: %w", err)
	}

	return entries, nil
}

// Create inserts a new log entry and returns the created entry with ID.
func (r *LogEntryRepository) Create(ctx context.Context, entry *models.LogEntry) (*models.LogEntry, error) {
	metadataBytes := entry.Metadata
	if metadataBytes == nil {
		metadataBytes = []byte("{}")
	}

	query := `INSERT INTO logs.log_entries (user_id, service, level, message, metadata) 
	         VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query,
		entry.UserID,
		entry.Service,
		entry.Level,
		entry.Message,
		metadataBytes,
	).Scan(&entry.ID, &entry.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("db: failed to create log entry: %w", err)
	}

	return entry, nil
}

// GetByID retrieves a log entry by its ID.
func (r *LogEntryRepository) GetByID(ctx context.Context, id int64) (*models.LogEntry, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, service, level, message, metadata, created_at FROM logs.log_entries WHERE id = $1`,
		id,
	)

	var entry models.LogEntry
	err := row.Scan(&entry.ID, &entry.UserID, &entry.Service, &entry.Level, &entry.Message, &entry.Metadata, &entry.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("db: failed to get log entry by id: %w", err)
	}

	return &entry, nil
}

// GetByService retrieves log entries filtered by service name.
func (r *LogEntryRepository) GetByService(ctx context.Context, service string, limit, offset int) ([]models.LogEntry, error) {
	query := `SELECT id, user_id, service, level, message, metadata, created_at FROM logs.log_entries 
	         WHERE service = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	entries, err := r.queryLogEntries(ctx, query, service, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("db: failed to query log entries by service: %w", err)
	}
	return entries, nil
}

// GetByLevel retrieves log entries filtered by level.
func (r *LogEntryRepository) GetByLevel(ctx context.Context, level string, limit, offset int) ([]models.LogEntry, error) {
	query := `SELECT id, user_id, service, level, message, metadata, created_at FROM logs.log_entries 
	         WHERE level = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	entries, err := r.queryLogEntries(ctx, query, level, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("db: failed to query log entries by level: %w", err)
	}
	return entries, nil
}

// GetByUser retrieves log entries for a specific user.
func (r *LogEntryRepository) GetByUser(ctx context.Context, userID int64, limit, offset int) ([]models.LogEntry, error) {
	query := `SELECT id, user_id, service, level, message, metadata, created_at FROM logs.log_entries 
	         WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	entries, err := r.queryLogEntries(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("db: failed to query log entries by user: %w", err)
	}
	return entries, nil
}

// GetRecent retrieves the most recent log entries.
func (r *LogEntryRepository) GetRecent(ctx context.Context, limit int) ([]models.LogEntry, error) {
	query := `SELECT id, user_id, service, level, message, metadata, created_at FROM logs.log_entries 
	         ORDER BY created_at DESC LIMIT $1`
	entries, err := r.queryLogEntries(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("db: failed to query recent log entries: %w", err)
	}
	return entries, nil
}

// GetStats returns statistics on log entries by level and service.
func (r *LogEntryRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	levelCounts := make(map[string]int)
	rows, err := r.db.QueryContext(ctx,
		`SELECT level, COUNT(*) as count FROM logs.log_entries GROUP BY level`,
	)
	if err != nil {
		return nil, fmt.Errorf("db: failed to query level stats: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("db: failed to close rows: %w", closeErr)
		}
	}()

	for rows.Next() {
		var level string
		var count int
		scanErr := rows.Scan(&level, &count)
		if scanErr != nil {
			return nil, fmt.Errorf("db: failed to scan stats: %w", scanErr)
		}
		levelCounts[level] = count
	}
	rowErr := rows.Err()
	if rowErr != nil {
		return nil, fmt.Errorf("db: rows iteration error: %w", rowErr)
	}

	stats["by_level"] = levelCounts

	serviceCounts := make(map[string]int)
	rows, err = r.db.QueryContext(ctx,
		`SELECT service, COUNT(*) as count FROM logs.log_entries GROUP BY service`,
	)
	if err != nil {
		return nil, fmt.Errorf("db: failed to query service stats: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("db: failed to close rows: %w", closeErr)
		}
	}()

	for rows.Next() {
		var service string
		var count int
		scanErr := rows.Scan(&service, &count)
		if scanErr != nil {
			return nil, fmt.Errorf("db: failed to scan stats: %w", scanErr)
		}
		serviceCounts[service] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db: rows iteration error: %w", err)
	}

	stats["by_service"] = serviceCounts

	return stats, nil
}

// Count returns the total number of log entries.
func (r *LogEntryRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM logs.log_entries`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("db: failed to count log entries: %w", err)
	}
	return count, nil
}

// Delete removes a log entry by ID.
func (r *LogEntryRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM logs.log_entries WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("db: failed to delete log entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("db: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("db: log entry not found")
	}

	return nil
}

// DeleteByService removes all log entries for a service.
func (r *LogEntryRepository) DeleteByService(ctx context.Context, service string) (int64, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM logs.log_entries WHERE service = $1`, service)
	if err != nil {
		return 0, fmt.Errorf("db: failed to delete log entries: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("db: failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// DeleteOlderThan removes log entries older than the specified days.
func (r *LogEntryRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM logs.log_entries WHERE created_at < NOW() - INTERVAL '1 day' * $1`,
		days,
	)
	if err != nil {
		return 0, fmt.Errorf("db: failed to delete old log entries: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("db: failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// GetMetadataValue extracts a specific value from the metadata JSONB.
func (r *LogEntryRepository) GetMetadataValue(metadata []byte, key string) (interface{}, error) {
	if len(metadata) == 0 {
		return nil, nil
	}

	var data map[string]interface{}
	err := json.Unmarshal(metadata, &data)
	if err != nil {
		return nil, fmt.Errorf("db: failed to unmarshal metadata: %w", err)
	}

	return data[key], nil
}
