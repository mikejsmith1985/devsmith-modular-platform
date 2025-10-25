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
func (r *LogRepository) Insert(ctx context.Context, entry interface{}) (int64, error) {
	// Convert interface{} to map[string]interface{}
	entryMap, ok := entry.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("db: entry must be map[string]interface{}")
	}

	service, ok := entryMap["service"].(string)
	if !ok || service == "" {
		return 0, fmt.Errorf("db: service is required")
	}

	level, ok := entryMap["level"].(string)
	if !ok || level == "" {
		return 0, fmt.Errorf("db: level is required")
	}

	message, ok := entryMap["message"].(string)
	if !ok || message == "" {
		return 0, fmt.Errorf("db: message is required")
	}

	// Extract metadata if provided
	metadataJSON := "{}"
	if metadata, ok := entryMap["metadata"].(map[string]interface{}); ok && len(metadata) > 0 {
		b, err := json.Marshal(metadata)
		if err != nil {
			return 0, fmt.Errorf("db: failed to marshal metadata: %w", err)
		}
		metadataJSON = string(b)
	}

	query := `INSERT INTO logs.entries (service, level, message, metadata) 
	         VALUES ($1, $2, $3, $4::jsonb) RETURNING id, created_at`

	var id int64
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, query, service, level, message, metadataJSON).Scan(&id, &createdAt)
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
func (r *LogRepository) Query(ctx context.Context, filters interface{}, page interface{}) ([]interface{}, error) {
	// Convert interface{} to QueryFilters
	var queryFilters *QueryFilters
	if filters != nil {
		if filtersMap, ok := filters.(map[string]interface{}); ok {
			queryFilters = &QueryFilters{
				Service: getStringValue(filtersMap, "service"),
				Level:   getStringValue(filtersMap, "level"),
				Search:  getStringValue(filtersMap, "search"),
				From:    getTimeValue(filtersMap, "from"),
				To:      getTimeValue(filtersMap, "to"),
				Sort:    getStringValue(filtersMap, "sort"),
			}
			// Handle metadata filters
			if meta, ok := filtersMap["metadata"].(map[string]string); ok {
				queryFilters.MetaEquals = meta
			}
		} else {
			queryFilters = &QueryFilters{}
		}
	} else {
		queryFilters = &QueryFilters{}
	}

	// Convert interface{} to PageOptions
	pageOpts := PageOptions{Limit: 100, Offset: 0}
	if page != nil {
		if pageMap, ok := page.(map[string]int); ok {
			if limit, ok := pageMap["limit"]; ok && limit > 0 {
				pageOpts.Limit = limit
			}
			if offset, ok := pageMap["offset"]; ok && offset >= 0 {
				pageOpts.Offset = offset
			}
		}
	}

	// Call internal query with strongly-typed parameters
	entries, err := r.queryInternal(ctx, queryFilters, pageOpts)
	if err != nil {
		return nil, err
	}

	// Convert []*LogEntry to []interface{}
	result := make([]interface{}, len(entries))
	for i, entry := range entries {
		result[i] = map[string]interface{}{
			"id":         entry.ID,
			"service":    entry.Service,
			"level":      entry.Level,
			"message":    entry.Message,
			"created_at": entry.CreatedAt,
			"metadata":   entry.Metadata,
		}
	}

	return result, nil
}

// queryInternal is the internal implementation that works with strongly-typed parameters.
func (r *LogRepository) queryInternal(ctx context.Context, filters *QueryFilters, page PageOptions) ([]*LogEntry, error) {
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

// Helper functions for type conversion
func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getTimeValue(m map[string]interface{}, key string) time.Time {
	if v, ok := m[key].(time.Time); ok {
		return v
	}
	if v, ok := m[key].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t
		}
	}
	return time.Time{}
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

// GetByID retrieves a single log entry by ID.
func (r *LogRepository) GetByID(ctx context.Context, id int64) (interface{}, error) {
	query := "SELECT id, service, level, message, metadata, created_at FROM logs.entries WHERE id = $1"

	var entryID int64
	var service, level, message string
	var metadataJSON sql.NullString
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(&entryID, &service, &level, &message, &metadataJSON, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("db: log entry not found")
		}
		return nil, fmt.Errorf("db: failed to query log entry: %w", err)
	}

	entry := map[string]interface{}{
		"id":         entryID,
		"service":    service,
		"level":      level,
		"message":    message,
		"created_at": createdAt,
		"metadata":   make(map[string]interface{}),
	}

	if metadataJSON.Valid {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON.String), &metadata); err == nil {
			entry["metadata"] = metadata
		}
	}

	return entry, nil
}

// DeleteBefore removes all log entries created before the specified timestamp.
func (r *LogRepository) DeleteBefore(ctx context.Context, ts interface{}) (int64, error) {
	if ts == nil {
		return 0, fmt.Errorf("db: timestamp cannot be nil")
	}

	// Convert interface{} to time.Time
	var deleteTime time.Time
	switch v := ts.(type) {
	case time.Time:
		deleteTime = v
	case string:
		// Try to parse string timestamp
		parsedTime, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return 0, fmt.Errorf("db: invalid timestamp format: %w", err)
		}
		deleteTime = parsedTime
	default:
		return 0, fmt.Errorf("db: timestamp must be time.Time or string, got %T", ts)
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM logs.entries WHERE created_at < $1", deleteTime)
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
