// Package db provides database access for log entries.
// This package implements repository pattern for correlation context operations.
package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// Database constants for correlation queries
const (
	// DefaultRecentCorrelationWindow is the default time window for recent correlations in minutes
	DefaultRecentCorrelationWindow = 5

	// DefaultRecentCorrelationLimit is the default limit for recent correlations
	DefaultRecentCorrelationLimit = 100

	// PostgresArrayEmpty is the empty PostgreSQL array representation
	PostgresArrayEmpty = "{}"

	// PostgresArrayNullValue is the PostgreSQL null value in array context
	PostgresArrayNullValue = "null"
)

// ContextRepository handles database operations for correlation context.
// It provides methods to query and aggregate logs by correlation ID,
// supporting both direct correlation_id column and JSONB context->>'correlation_id' lookups.
type ContextRepository struct {
	db *sql.DB
}

// NewContextRepository creates a new ContextRepository with the given database connection.
func NewContextRepository(db *sql.DB) *ContextRepository {
	return &ContextRepository{db: db}
}

// GetCorrelatedLogs retrieves all logs associated with a correlation ID.
//
// Supports dual-path lookup:
// - Direct: correlation_id = $1
// - JSONB: context->>'correlation_id' = $2
//
// This enables querying both legacy and new-style correlation tracking.
//
// Parameters:
// - ctx: Request context for cancellation
// - correlationID: Required - the correlation ID to query
// - limit: Max results (capped at 1000, default 50)
// - offset: Results to skip
//
// Returns logs ordered by timestamp DESC, then id DESC for chronological viewing.
func (r *ContextRepository) GetCorrelatedLogs(
	ctx context.Context,
	correlationID string,
	limit, offset int,
) ([]models.LogEntry, error) {
	if correlationID == "" {
		return nil, errors.New("correlation_id required")
	}

	if r.db == nil {
		return []models.LogEntry{}, nil
	}

	// Validate pagination
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, timestamp, level, message, service, context, created_at
		FROM logs.log_entries
		WHERE correlation_id = $1 OR (context IS NOT NULL AND context->>'correlation_id' = $2)
		ORDER BY timestamp DESC, id DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, correlationID, correlationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query correlated logs: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close rows: %w", closeErr)
		}
	}()

	var logs []models.LogEntry
	for rows.Next() {
		var log models.LogEntry
		var contextJSON sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.Level,
			&log.Message,
			&log.Service,
			&contextJSON,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}

		// Parse context if present
		if contextJSON.Valid {
			ctx := &models.CorrelationContext{}
			if err := json.Unmarshal([]byte(contextJSON.String), ctx); err != nil {
				return nil, fmt.Errorf("failed to unmarshal context: %w", err)
			}
			log.Context = ctx
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return logs, nil
}

// GetCorrelationCount returns the count of logs associated with a correlation ID.
// Uses dual-path lookup (correlation_id column or context->>'correlation_id').
func (r *ContextRepository) GetCorrelationCount(
	ctx context.Context,
	correlationID string,
) (int, error) {
	if correlationID == "" {
		return 0, errors.New("correlation_id required")
	}

	if r.db == nil {
		return 0, nil
	}

	query := `
		SELECT COUNT(*)
		FROM logs.log_entries
		WHERE correlation_id = $1 OR (context IS NOT NULL AND context->>'correlation_id' = $2)
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, correlationID, correlationID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count correlated logs: %w", err)
	}

	return count, nil
}

// GetRecentCorrelations returns active correlation IDs from the last N minutes.
// Useful for discovering active traces in the system.
//
// Parameters:
// - ctx: Request context for cancellation
// - minutes: Time window in minutes (0 uses default)
// - limit: Max correlations to return (0 uses default)
//
// Returns correlation IDs ordered by most recent first.
func (r *ContextRepository) GetRecentCorrelations(
	ctx context.Context,
	minutes, limit int,
) ([]string, error) {
	if r.db == nil {
		return []string{}, nil
	}

	if limit <= 0 {
		limit = DefaultRecentCorrelationLimit
	}
	if minutes <= 0 {
		minutes = DefaultRecentCorrelationWindow
	}

	query := `
		SELECT DISTINCT correlation_id
		FROM logs.log_entries
		WHERE created_at > NOW() - INTERVAL '1 minute' * $1
		  AND correlation_id IS NOT NULL
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, minutes, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent correlations: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close rows: %w", closeErr)
		}
	}()

	var correlationIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan correlation id: %w", err)
		}
		correlationIDs = append(correlationIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return correlationIDs, nil
}

// parsePostgresArray extracts string values from PostgreSQL array format.
// Converts PostgreSQL array format {val1,val2,...} to []string.
//
// Example:
// - "{}" → []string{}
// - "{foo,bar,baz}" → []string{"foo", "bar", "baz"}
// - '{foo,"bar baz",null}' → []string{"foo", "bar baz"}
//
// Returns empty slice for invalid format.
func parsePostgresArray(arrayStr string) []string {
	var result []string
	if arrayStr == PostgresArrayEmpty || !strings.HasPrefix(arrayStr, "{") || !strings.HasSuffix(arrayStr, "}") {
		return result
	}

	inner := arrayStr[1 : len(arrayStr)-1]
	if inner == "" {
		return result
	}

	for _, item := range strings.Split(inner, ",") {
		if item != "" && item != PostgresArrayNullValue {
			result = append(result, strings.Trim(item, "\""))
		}
	}

	return result
}

// GetContextMetadata retrieves aggregated metadata for a correlation.
// Queries all logs for a correlation and returns:
// - total_logs: Count of logs
// - correlation_id: The queried correlation ID
// - services: List of unique services involved
// - trace_ids: List of unique OpenTelemetry trace IDs
//
// This provides a summary view suitable for distributed tracing dashboards.
func (r *ContextRepository) GetContextMetadata(
	ctx context.Context,
	correlationID string,
) (map[string]interface{}, error) {
	if correlationID == "" {
		return nil, errors.New("correlation_id required")
	}

	if r.db == nil {
		return make(map[string]interface{}), nil
	}

	// Aggregate data from all logs in this correlation
	query := `
		SELECT ARRAY_AGG(DISTINCT context->>'service') as services,
		       ARRAY_AGG(DISTINCT context->>'trace_id') as trace_ids,
		       COUNT(*) as total_logs
		FROM logs.log_entries
		WHERE correlation_id = $1 OR (context IS NOT NULL AND context->>'correlation_id' = $2)
	`

	var servicesJSON, traceIDsJSON sql.NullString
	var totalLogs int

	err := r.db.QueryRowContext(ctx, query, correlationID, correlationID).
		Scan(&servicesJSON, &traceIDsJSON, &totalLogs)
	if err != nil {
		return nil, fmt.Errorf("failed to query context metadata: %w", err)
	}

	metadata := make(map[string]interface{})
	metadata["total_logs"] = totalLogs
	metadata["correlation_id"] = correlationID

	// Parse and add services if present
	if servicesJSON.Valid {
		services := parsePostgresArray(servicesJSON.String)
		if len(services) > 0 {
			metadata["services"] = services
		}
	}

	// Parse and add trace IDs if present
	if traceIDsJSON.Valid {
		traceIDs := parsePostgresArray(traceIDsJSON.String)
		if len(traceIDs) > 0 {
			metadata["trace_ids"] = traceIDs
		}
	}

	return metadata, nil
}
