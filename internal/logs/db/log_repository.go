// Package logs_db provides database access and query/filter types for log entries.
package logs_db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// LogEntry represents a log entry in the database.
type LogEntry struct {
	ID        int64
	CreatedAt time.Time
	Metadata  map[string]interface{}
	Tags      []string // Auto-generated and manual tags
	Message   string
	Service   string
	Level     string
}

// QueryFilters represents filtering options for log queries.
type QueryFilters struct {
	From       time.Time         // Filter logs created at or after this time
	To         time.Time         // Filter logs created at or before this time
	MetaEquals map[string]string // Filter logs where metadata keys equal given values
	Service    string            // Filter logs by service name
	Level      string            // Filter logs by level (e.g., "error", "info")
	Search     string            // Full-text search on message field (ILIKE)
}

// PageOptions holds pagination parameters for query results.
type PageOptions struct {
	Limit  int // Number of results to return (must be > 0)
	Offset int // Number of results to skip (must be >= 0)
}

// LogRepository handles CRUD operations for log entries.
type LogRepository struct {
	db *sql.DB
}

// NewLogRepository creates a new LogRepository.
func NewLogRepository(db *sql.DB) *LogRepository {
	return &LogRepository{db: db}
}

// Save inserts a new log entry and returns its ID.
func (r *LogRepository) Save(ctx context.Context, entry *LogEntry) (int64, error) {
	if entry == nil {
		return 0, fmt.Errorf("entry cannot be nil")
	}

	if entry.Message == "" {
		return 0, fmt.Errorf("message is required")
	}

	if entry.Level == "" {
		return 0, fmt.Errorf("level is required")
	}

	if entry.Service == "" {
		return 0, fmt.Errorf("service is required")
	}

	if entry.CreatedAt.IsZero() {
		return 0, fmt.Errorf("created_at is required")
	}

	// If no database connection, return mock ID for testing
	if r.db == nil {
		// Still check context deadline even for mock responses
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}
		return 1, nil
	}

	// Check context is not cancelled
	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// Marshal metadata to JSON
	metadataJSON := "{}"
	if len(entry.Metadata) > 0 {
		b, err := json.Marshal(entry.Metadata)
		if err != nil {
			return 0, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(b)
	}

	// Insert and return ID
	query := `INSERT INTO logs.entries (service, level, message, metadata, created_at)
	         VALUES ($1, $2, $3, $4::jsonb, $5)
	         RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query, entry.Service, entry.Level, entry.Message, metadataJSON, entry.CreatedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert log entry: %w", err)
	}

	return id, nil
}

// buildWhereClause builds parameterized WHERE clause fragments from filters.
// nolint:gocritic // return values are self-explanatory (WHERE fragments, args, count)
func buildWhereClause(filters *QueryFilters) ([]string, []interface{}, int) {
	var fragments []string
	var args []interface{}
	argNum := 1

	if filters == nil {
		return fragments, args, argNum
	}

	if filters.Service != "" && filters.Service != "all" {
		fragments = append(fragments, fmt.Sprintf("service = $%d", argNum))
		args = append(args, filters.Service)
		argNum++
	}

	if filters.Level != "" && filters.Level != "all" {
		fragments = append(fragments, fmt.Sprintf("level = $%d", argNum))
		args = append(args, filters.Level)
		argNum++
	}

	if !filters.From.IsZero() {
		fragments = append(fragments, fmt.Sprintf("created_at >= $%d", argNum))
		args = append(args, filters.From)
		argNum++
	}

	if !filters.To.IsZero() {
		fragments = append(fragments, fmt.Sprintf("created_at <= $%d", argNum))
		args = append(args, filters.To)
		argNum++
	}

	if filters.Search != "" {
		fragments = append(fragments, fmt.Sprintf("message ILIKE $%d", argNum))
		args = append(args, "%"+filters.Search+"%")
		argNum++
	}

	if len(filters.MetaEquals) > 0 {
		for k, v := range filters.MetaEquals {
			fragments = append(fragments, fmt.Sprintf("metadata @> jsonb_build_object($%d::text, $%d::text)::jsonb", argNum, argNum+1))
			args = append(args, k, v)
			argNum += 2
		}
	}

	return fragments, args, argNum
}

// Query retrieves log entries matching specified filters with pagination support.
// nolint:gocognit,gocyclo // complexity is necessary for comprehensive query building and filtering
func (r *LogRepository) Query(ctx context.Context, filters *QueryFilters, page PageOptions) ([]*LogEntry, error) {
	// Validate pagination
	if page.Limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than 0")
	}
	if page.Offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative")
	}

	// Check context
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// If no database connection, return empty slice for testing
	if r.db == nil {
		return []*LogEntry{}, nil
	}

	// Build WHERE clause
	whereFragments, args, argNum := buildWhereClause(filters)
	args = append(args, page.Limit, page.Offset)

	// Build query - select actual columns (no tags column exists)
	query := "SELECT id, service, level, message, metadata, created_at FROM logs.entries"
	if len(whereFragments) > 0 {
		query += " WHERE " + strings.Join(whereFragments, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query log entries: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("failed to close rows: %w", closeErr)
			}
		}
	}()

	// Scan results
	var entries []*LogEntry
	for rows.Next() {
		var id int64
		var service, level, message string
		var metadataJSON sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&id, &service, &level, &message, &metadataJSON, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}

		entry := &LogEntry{
			ID:        id,
			Service:   service,
			Level:     level,
			Message:   message,
			Tags:      []string{}, // No tags column in schema
			CreatedAt: createdAt,
			Metadata:  make(map[string]interface{}),
		}

		// Parse metadata JSON if it exists
		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &entry.Metadata); err != nil {
				// Log the error but continue with empty metadata
				log.Printf("Failed to unmarshal metadata for log entry %d: %v", entry.ID, err)
				entry.Metadata = make(map[string]interface{})
			}
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	if entries == nil {
		entries = []*LogEntry{}
	}

	return entries, nil
}

// GetByID retrieves a single log entry by ID.
func (r *LogRepository) GetByID(ctx context.Context, id int64) (*LogEntry, error) {
	// Validate ID
	if id <= 0 {
		return nil, fmt.Errorf("id must be greater than 0")
	}

	// Check context
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// If no database connection, return nil for testing
	if r.db == nil {
		return nil, nil
	}

	// Query single entry
	query := "SELECT id, service, level, message, metadata, created_at FROM logs.entries WHERE id = $1"

	var id64 int64
	var service, level, message string
	var metadataJSON sql.NullString
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(&id64, &service, &level, &message, &metadataJSON, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("log entry not found")
		}
		return nil, fmt.Errorf("failed to query log entry: %w", err)
	}

	entry := &LogEntry{
		ID:        id64,
		Service:   service,
		Level:     level,
		Message:   message,
		CreatedAt: createdAt,
		Metadata:  make(map[string]interface{}),
	}

	// Parse metadata JSON if it exists
	if metadataJSON.Valid && metadataJSON.String != "" {
		if err := json.Unmarshal([]byte(metadataJSON.String), &entry.Metadata); err != nil {
			log.Printf("Warning: Failed to unmarshal metadata JSON for log entry %d: %v", entry.ID, err)
			// Continue with empty metadata map
		}
	}

	return entry, nil
}

// getCountsByLevel aggregates log count grouped by level.
func (r *LogRepository) getCountsByLevel(ctx context.Context) (map[string]int64, error) {
	return r.aggregateCount(ctx, "level", "failed to query by level", "failed to scan level stats", "rows iteration error (by_level)")
}

// getCountsByService aggregates log count grouped by service.
func (r *LogRepository) getCountsByService(ctx context.Context) (map[string]int64, error) {
	return r.aggregateCount(ctx, "service", "failed to query by service", "failed to scan service stats", "rows iteration error (by_service)")
}

// aggregateCount performs aggregation on a specific column.
// nolint:gocritic,gosec // return values are self-explanatory; SQL column names are controlled
func (r *LogRepository) aggregateCount(ctx context.Context, column, queryErr, scanErr, iterErr string) (map[string]int64, error) {
	if r.db == nil {
		return map[string]int64{}, nil
	}

	query := fmt.Sprintf("SELECT %s, COUNT(*) FROM logs.entries GROUP BY %s", column, column)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf(queryErr+": %w", err)
	}
	defer func() {
		//nolint:errcheck // Best effort to close rows
		rows.Close()
		if err == nil {
			err = nil
		}
	}()

	result := make(map[string]int64)
	for rows.Next() {
		var key string
		var count int64
		if err := rows.Scan(&key, &count); err != nil {
			return nil, fmt.Errorf(scanErr+": %w", err)
		}
		result[key] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(iterErr+": %w", err)
	}

	return result, nil
}

// FindAllServices returns all unique service names in the logs.
func (r *LogRepository) FindAllServices(ctx context.Context) ([]string, error) {
	if r.db == nil {
		return []string{}, nil
	}

	query := "SELECT DISTINCT service FROM logs.entries ORDER BY service"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find services: %w", err)
	}
	//nolint:errcheck // Best effort to close rows
	defer rows.Close()

	var services []string
	for rows.Next() {
		var service string
		if err := rows.Scan(&service); err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, service)
	}

	return services, nil
}

// CountByServiceAndLevel counts log entries matching service, level, and time range.
func (r *LogRepository) CountByServiceAndLevel(ctx context.Context, service, level string, start, end time.Time) (int64, error) {
	if r.db == nil {
		return 0, nil
	}

	query := "SELECT COUNT(*) FROM logs.entries WHERE service = $1 AND level = $2 AND created_at >= $3 AND created_at <= $4"
	var count int64
	err := r.db.QueryRowContext(ctx, query, service, level, start, end).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to count logs: %w", err)
	}

	return count, nil
}

// FindTopMessages finds the most frequent log messages matching criteria.
// nolint:dupl // Similar query pattern is acceptable for database operations
func (r *LogRepository) FindTopMessages(ctx context.Context, service, level string, start, end time.Time, limit int) ([]logs_models.LogMessage, error) {
	if r.db == nil {
		return []logs_models.LogMessage{}, nil
	}

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	query := `SELECT message, COUNT(*) as count, MAX(created_at) as last_seen 
	         FROM logs.entries 
	         WHERE service = $1 AND level = $2 AND created_at >= $3 AND created_at <= $4
	         GROUP BY message 
	         ORDER BY count DESC 
	         LIMIT $5`

	rows, err := r.db.QueryContext(ctx, query, service, level, start, end, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find top messages: %w", err)
	}
	//nolint:errcheck // Best effort to close rows
	defer rows.Close()

	var messages []logs_models.LogMessage
	for rows.Next() {
		var message string
		var count int64
		var lastSeen time.Time
		if err := rows.Scan(&message, &count, &lastSeen); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, logs_models.LogMessage{
			Message:  message,
			Count:    int(count),
			LastSeen: lastSeen,
			Service:  service,
			Level:    level,
		})
	}

	return messages, nil
}

// GetStats returns aggregated statistics about log entries.
func (r *LogRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// If no database connection, return mock stats for testing
	if r.db == nil {
		return map[string]interface{}{
			"total":      0,
			"by_level":   map[string]int64{},
			"by_service": map[string]int64{},
		}, nil
	}

	stats := make(map[string]interface{})

	// Get total count
	var totalCount int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM logs.entries").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats["total"] = totalCount

	// Get counts by level
	byLevel, err := r.getCountsByLevel(ctx)
	if err != nil {
		return nil, err
	}
	stats["by_level"] = byLevel

	// Get counts by service
	byService, err := r.getCountsByService(ctx)
	if err != nil {
		return nil, err
	}
	stats["by_service"] = byService

	return stats, nil
}

// DeleteOld removes log entries older than the given timestamp.
func (r *LogRepository) DeleteOld(ctx context.Context, ts time.Time) (int64, error) {
	// Validate timestamp
	if ts.IsZero() {
		return 0, fmt.Errorf("timestamp cannot be zero")
	}

	// Check context
	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// If no database connection, return 0 for testing
	if r.db == nil {
		return 0, nil
	}

	// Delete entries older than timestamp
	result, err := r.db.ExecContext(ctx, "DELETE FROM logs.entries WHERE created_at < $1", ts)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old log entries: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// validateBulkEntries validates all entries before insertion.
func validateBulkEntries(entries []*LogEntry) error {
	if entries == nil {
		return fmt.Errorf("entries cannot be nil")
	}
	if len(entries) == 0 {
		return fmt.Errorf("entries cannot be empty")
	}

	for i, entry := range entries {
		if entry == nil {
			return fmt.Errorf("entry at index %d is nil", i)
		}
		if entry.Message == "" {
			return fmt.Errorf("entry at index %d: message is required", i)
		}
		if entry.Level == "" {
			return fmt.Errorf("entry at index %d: level is required", i)
		}
		if entry.Service == "" {
			return fmt.Errorf("entry at index %d: service is required", i)
		}
		if entry.CreatedAt.IsZero() {
			return fmt.Errorf("entry at index %d: created_at is required", i)
		}
	}
	return nil
}

// BulkInsert inserts multiple log entries at once.
// nolint:gocognit,gocyclo // complexity is necessary for transaction handling and error management
func (r *LogRepository) BulkInsert(ctx context.Context, entries []*LogEntry) (int64, error) {
	// Validate input
	if err := validateBulkEntries(entries); err != nil {
		return 0, err
	}

	// Check context
	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// If no database connection, return count for testing
	if r.db == nil {
		return int64(len(entries)), nil
	}

	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			// Log but don't override previous errors
			if err == nil {
				err = fmt.Errorf("failed to rollback transaction: %w", rollbackErr)
			}
		}
	}()

	// Prepare statement
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO logs.entries (service, level, message, metadata, created_at)
		VALUES ($1, $2, $3, $4::jsonb, $5)
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			// Log but don't override previous errors
			if err == nil {
				err = fmt.Errorf("failed to close statement: %w", closeErr)
			}
		}
	}()

	// Insert each entry
	var insertedCount int64
	for _, entry := range entries {
		metadataJSON := "{}"
		if len(entry.Metadata) > 0 {
			b, err := json.Marshal(entry.Metadata)
			if err != nil {
				return 0, fmt.Errorf("failed to marshal metadata: %w", err)
			}
			metadataJSON = string(b)
		}

		_, err := stmt.ExecContext(ctx, entry.Service, entry.Level, entry.Message, metadataJSON, entry.CreatedAt)
		if err != nil {
			return 0, fmt.Errorf("failed to insert log entry: %w", err)
		}
		insertedCount++
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return insertedCount, nil
}

// GetLogStatsByLevel returns the count of logs grouped by level for the React frontend StatCards.
func (r *LogRepository) GetLogStatsByLevel(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT 
			LOWER(level) as level,
			COUNT(*) as count
		FROM logs.entries
		GROUP BY LOWER(level)
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query log stats: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Log error but don't return it to avoid masking the primary error
			fmt.Printf("Error closing rows: %v\n", closeErr)
		}
	}()

	stats := map[string]int{
		"debug":    0,
		"info":     0,
		"warning":  0,
		"error":    0,
		"critical": 0,
	}

	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		stats[level] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return stats, nil
}

// GetAllTags retrieves all unique tags from the database.
func (r *LogRepository) GetAllTags(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	if r.db == nil {
		return []string{}, nil
	}

	// Tags column doesn't exist in schema - return empty array
	return []string{}, nil
}

// updateTagsHelper is a helper to reduce code duplication for tag operations
func (r *LogRepository) updateTagsHelper(ctx context.Context, logID int64, tag, operation string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	if r.db == nil {
		return nil // No-op for testing
	}

	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}

	var query string
	if operation == "add" {
		query = `UPDATE logs.entries 
			 SET tags = array_append(tags, $1)
			 WHERE id = $2 AND NOT ($1 = ANY(tags))`
	} else {
		query = `UPDATE logs.entries 
			 SET tags = array_remove(tags, $1)
			 WHERE id = $2 AND $1 = ANY(tags)`
	}

	result, err := r.db.ExecContext(ctx, query, tag, logID)
	if err != nil {
		return fmt.Errorf("failed to %s tag: %w", operation, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		if operation == "add" {
			return fmt.Errorf("log entry not found or tag already exists")
		}
		return fmt.Errorf("log entry not found or tag does not exist")
	}

	return nil
}

// AddTag adds a manual tag to a log entry.
func (r *LogRepository) AddTag(ctx context.Context, logID int64, tag string) error {
	return r.updateTagsHelper(ctx, logID, tag, "add")
}

// RemoveTag removes a tag from a log entry.
func (r *LogRepository) RemoveTag(ctx context.Context, logID int64, tag string) error {
	return r.updateTagsHelper(ctx, logID, tag, "remove")
}
