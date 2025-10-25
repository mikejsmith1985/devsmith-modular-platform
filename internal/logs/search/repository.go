package search

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const (
	fieldService = "service"
	fieldLevel   = "level"
	fieldMessage = "message"
	whereTrue    = "true"
)

// Repository handles persistence of logs in PostgreSQL with full-text search.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new log repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SearchLogs executes a search query and returns matching logs.
func (r *Repository) SearchLogs(ctx context.Context, query *ParsedQuery) ([]*LogEntry, error) {
	if !query.IsValid {
		return nil, fmt.Errorf("invalid query: %s", query.ErrorMsg)
	}

	if query.RootNode == nil {
		return []*LogEntry{}, nil
	}

	whereClause, args := r.buildWhereClause(query.RootNode)

	// nolint:gosec // query is built from validated AST, not user input
	sql := fmt.Sprintf(`
		SELECT id, service, level, message, created_at
		FROM logs
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT 1000
	`, whereClause)

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer func() {
		// nolint:errcheck // Close errors in defer are not actionable
		_ = rows.Close()
	}()

	var results []*LogEntry
	for rows.Next() {
		var entry LogEntry
		if err := rows.Scan(&entry.ID, &entry.Service, &entry.Level, &entry.Message, &entry.Timestamp); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		results = append(results, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}

	return results, nil
}

// SaveLog inserts a log entry into the database.
func (r *Repository) SaveLog(ctx context.Context, entry *LogEntry) error {
	sql := `
		INSERT INTO logs (service, level, message, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	return r.db.QueryRowContext(ctx, sql,
		entry.Service,
		entry.Level,
		entry.Message,
		"{}",
		time.Now(),
	).Scan(&entry.ID)
}

// DeleteLog removes a log entry by ID.
func (r *Repository) DeleteLog(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM logs WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected check failed: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("log not found: id %d", id)
	}

	return nil
}

// DeleteBefore removes all log entries created before the given timestamp.
func (r *Repository) DeleteBefore(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM logs WHERE created_at < $1", before)
	if err != nil {
		return 0, fmt.Errorf("delete failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected check failed: %w", err)
	}

	return rowsAffected, nil
}

// buildWhereClause constructs SQL WHERE clause from query AST.
func (r *Repository) buildWhereClause(node *QueryNode) (clause string, args []interface{}) {
	if node == nil {
		return whereTrue, []interface{}{}
	}

	switch node.Type {
	case "FIELD":
		return r.buildFieldClause(node)
	case "REGEX":
		return r.buildRegexClause(node)
	case "AND":
		left, leftArgs := r.buildWhereClause(node.Left)
		right, rightArgs := r.buildWhereClause(node.Right)
		allArgs := make([]interface{}, 0, len(leftArgs)+len(rightArgs))
		allArgs = append(allArgs, leftArgs...)
		allArgs = append(allArgs, rightArgs...)
		return fmt.Sprintf("(%s AND %s)", left, right), allArgs
	case "OR":
		left, leftArgs := r.buildWhereClause(node.Left)
		right, rightArgs := r.buildWhereClause(node.Right)
		allArgs := make([]interface{}, 0, len(leftArgs)+len(rightArgs))
		allArgs = append(allArgs, leftArgs...)
		allArgs = append(allArgs, rightArgs...)
		return fmt.Sprintf("(%s OR %s)", left, right), allArgs
	case "NOT":
		inner, innerArgs := r.buildWhereClause(node.Left)
		if node.Left == nil {
			inner, innerArgs = r.buildWhereClause(node.Right)
		}
		return fmt.Sprintf("(NOT %s)", inner), innerArgs
	default:
		return whereTrue, []interface{}{}
	}
}

// buildFieldClause constructs WHERE clause for field:value queries.
func (r *Repository) buildFieldClause(node *QueryNode) (clause string, args []interface{}) {
	field := node.Field
	value := node.Value

	var dbColumn string
	switch field {
	case fieldService:
		dbColumn = fieldService
	case fieldLevel:
		dbColumn = fieldLevel
	case fieldMessage:
		dbColumn = fieldMessage
	default:
		return whereTrue, []interface{}{}
	}

	// Use full-text search if it's a message field
	if field == fieldMessage && value != "" {
		return "search_vector @@ plainto_tsquery('english', $1)", []interface{}{value}
	}

	// Exact match for other fields
	return fmt.Sprintf("%s = $1", dbColumn), []interface{}{value}
}

// buildRegexClause constructs WHERE clause for regex patterns.
func (r *Repository) buildRegexClause(node *QueryNode) (clause string, args []interface{}) {
	field := node.Field
	pattern := node.Value

	var dbColumn string
	switch field {
	case fieldService:
		dbColumn = fieldService
	case fieldLevel:
		dbColumn = fieldLevel
	case fieldMessage:
		dbColumn = fieldMessage
	default:
		return whereTrue, []interface{}{}
	}

	// Use SIMILAR TO for regex (PostgreSQL)
	return fmt.Sprintf("%s SIMILAR TO $1", dbColumn), []interface{}{pattern}
}

// Query performs a complex search with filters and pagination.
func (r *Repository) Query(ctx context.Context, filters, page interface{}) ([]*LogEntry, error) {
	// This will be implemented in next phase for saved searches
	return []*LogEntry{}, nil
}
