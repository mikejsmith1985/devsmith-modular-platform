// Package db provides database access for log entries.
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

// ContextRepository handles database operations for correlation context
type ContextRepository struct {
	db *sql.DB
}

// NewContextRepository creates a new ContextRepository
func NewContextRepository(db *sql.DB) *ContextRepository {
	return &ContextRepository{db: db}
}

// GetCorrelatedLogs retrieves all logs for a correlation ID with pagination
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

// GetCorrelationCount returns the count of logs for a correlation ID
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

// GetRecentCorrelations returns active correlation IDs from the last N minutes
func (r *ContextRepository) GetRecentCorrelations(
	ctx context.Context,
	minutes, limit int,
) ([]string, error) {
	if r.db == nil {
		return []string{}, nil
	}

	if limit <= 0 {
		limit = 100
	}
	if minutes <= 0 {
		minutes = 5
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

// parsePostgresArray extracts values from PostgreSQL array format {val1,val2,...}
func parsePostgresArray(arrayStr string) []string {
	var result []string
	if arrayStr == "{}" || !strings.HasPrefix(arrayStr, "{") || !strings.HasSuffix(arrayStr, "}") {
		return result
	}

	inner := arrayStr[1 : len(arrayStr)-1]
	if inner == "" {
		return result
	}

	for _, item := range strings.Split(inner, ",") {
		if item != "" && item != "null" {
			result = append(result, strings.Trim(item, "\""))
		}
	}

	return result
}

// GetContextMetadata retrieves metadata aggregated from all logs for a correlation
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

	// Get distinct services involved in this correlation
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
