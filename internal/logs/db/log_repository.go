// Package db provides database access and query/filter types for log entries.
package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// LogEntry represents a log entry in the database.
type LogEntry struct {
	Metadata  map[string]interface{}
	Message   string
	CreatedAt time.Time
	Service   string
	Level     string
	ID        int64
}

// QueryFilters holds filter criteria for log queries.
type QueryFilters struct {
	MetaEquals map[string]string
	From       time.Time
	To         time.Time
	Search     string
	Service    string
	Level      string
	Sort       string // Added for sorting
}

// PageOptions holds pagination parameters.
type PageOptions struct {
	Limit  int
	Offset int
}

// LogRepository handles CRUD operations for log entries.
type LogRepository struct {
	db *sql.DB
}

// NewLogRepository creates a new LogRepository.
func NewLogRepository(db *sql.DB) *LogRepository {
	return &LogRepository{db: db}
}

// Insert adds a new log entry and returns its ID.
func (r *LogRepository) Insert(ctx context.Context, e *LogEntry) (int64, error) {
	if e == nil {
		return 0, fmt.Errorf("db: log entry cannot be nil")
	}

	metadataJSON := "{}"
	if len(e.Metadata) > 0 {
		// Marshal metadata map to JSON
		b, err := json.Marshal(e.Metadata)
		if err != nil {
			return 0, fmt.Errorf("db: failed to marshal metadata: %w", err)
		}
		metadataJSON = string(b)
	}

	query := `INSERT INTO logs.entries (service, level, message, metadata) 
	         VALUES ($1, $2, $3, $4::jsonb) RETURNING id, created_at`

	var id int64
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, query, e.Service, e.Level, e.Message, metadataJSON).Scan(&id, &createdAt)
	if err != nil {
		return 0, fmt.Errorf("db: failed to insert log entry: %w", err)
	}

	return id, nil
}

// buildWhereClause builds WHERE clause fragments and args for query.
func buildWhereClause(filters *QueryFilters) (fragments []string, args []interface{}, nextArgNum int) {
	fragments = []string{}
	args = []interface{}{}
	nextArgNum = 1

	if filters.Service != "" {
		fragments = append(fragments, fmt.Sprintf("service = $%d", nextArgNum))
		args = append(args, filters.Service)
		nextArgNum++
	}

	if filters.Level != "" {
		fragments = append(fragments, fmt.Sprintf("level = $%d", nextArgNum))
		args = append(args, filters.Level)
		nextArgNum++
	}

	if !filters.From.IsZero() {
		fragments = append(fragments, fmt.Sprintf("created_at >= $%d", nextArgNum))
		args = append(args, filters.From)
		nextArgNum++
	}

	if !filters.To.IsZero() {
		fragments = append(fragments, fmt.Sprintf("created_at <= $%d", nextArgNum))
		args = append(args, filters.To)
		nextArgNum++
	}

	if filters.Search != "" {
		fragments = append(fragments, fmt.Sprintf("message ILIKE $%d", nextArgNum))
		args = append(args, "%"+filters.Search+"%")
		nextArgNum++
	}

	if len(filters.MetaEquals) > 0 {
		for k, v := range filters.MetaEquals {
			fragments = append(fragments, fmt.Sprintf("metadata @> jsonb_build_object($%d::text, $%d::text)::jsonb", nextArgNum, nextArgNum+1))
			args = append(args, k, v)
			nextArgNum += 2
		}
	}

	return
}

// Query retrieves log entries matching the filters with pagination.
func (r *LogRepository) Query(ctx context.Context, filters *QueryFilters, page PageOptions) ([]*LogEntry, error) {
	if page.Limit <= 0 {
		return nil, fmt.Errorf("db: limit must be positive")
	}
	if page.Offset < 0 {
		return nil, fmt.Errorf("db: offset cannot be negative")
	}

	whereFragments, args, argNum := buildWhereClause(filters)
	args = append(args, page.Limit, page.Offset)

	query := "SELECT id, service, level, message, metadata, created_at FROM logs.entries"
	if len(whereFragments) > 0 {
		query += " WHERE " + strings.Join(whereFragments, " AND ")
	}

	// Determine sort order (defaults to DESC if not specified)
	sortOrder := "DESC"
	if filters.Sort == "asc" {
		sortOrder = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY created_at %s, id %s LIMIT $%d OFFSET $%d", sortOrder, sortOrder, argNum, argNum+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db: failed to query log entries: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("db: failed to close rows: %w", closeErr)
		}
	}()

	var entries []*LogEntry
	for rows.Next() {
		var id int64
		var service, level, message string
		var metadataJSON sql.NullString
		var createdAt time.Time

		err := rows.Scan(&id, &service, &level, &message, &metadataJSON, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("db: failed to scan log entry: %w", err)
		}

		entry := &LogEntry{
			ID:        id,
			Service:   service,
			Level:     level,
			Message:   message,
			CreatedAt: createdAt,
			Metadata:  make(map[string]interface{}),
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db: rows iteration error: %w", err)
	}

	return entries, nil
}

// DeleteByID removes a log entry by ID.
func (r *LogRepository) DeleteByID(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM logs.entries WHERE id = $1", id)
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

// DeleteBefore removes all log entries created before the specified timestamp.
func (r *LogRepository) DeleteBefore(ctx context.Context, ts time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM logs.entries WHERE created_at < $1", ts)
	if err != nil {
		return 0, fmt.Errorf("db: failed to delete old log entries: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("db: failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// Stats returns aggregated statistics about log entries.
func (r *LogRepository) Stats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get total count
	var total int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM logs.entries").Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("db: failed to count log entries: %w", err)
	}
	stats["total"] = total

	// Get counts by level
	byLevel := make(map[string]interface{})
	levelRows, err := r.db.QueryContext(ctx, "SELECT level, COUNT(*) as count FROM logs.entries GROUP BY level ORDER BY level")
	if err != nil {
		return nil, fmt.Errorf("db: failed to query level stats: %w", err)
	}
	defer func() {
		if closeErr := levelRows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("db: failed to close level rows: %w", closeErr)
		}
	}()

	for levelRows.Next() {
		var level string
		var count int64
		if err := levelRows.Scan(&level, &count); err != nil {
			return nil, fmt.Errorf("db: failed to scan level stat: %w", err)
		}
		byLevel[level] = float64(count)
	}
	if err := levelRows.Err(); err != nil {
		return nil, fmt.Errorf("db: level rows iteration error: %w", err)
	}
	stats["by_level"] = byLevel

	// Get counts by service
	byService := make(map[string]interface{})
	serviceRows, err := r.db.QueryContext(ctx, "SELECT service, COUNT(*) as count FROM logs.entries GROUP BY service ORDER BY service")
	if err != nil {
		return nil, fmt.Errorf("db: failed to query service stats: %w", err)
	}
	defer func() {
		if closeErr := serviceRows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("db: failed to close service rows: %w", closeErr)
		}
	}()

	for serviceRows.Next() {
		var service string
		var count int64
		if err := serviceRows.Scan(&service, &count); err != nil {
			return nil, fmt.Errorf("db: failed to scan service stat: %w", err)
		}
		byService[service] = float64(count)
	}
	if err := serviceRows.Err(); err != nil {
		return nil, fmt.Errorf("db: service rows iteration error: %w", err)
	}
	stats["by_service"] = byService

	return stats, nil
}
