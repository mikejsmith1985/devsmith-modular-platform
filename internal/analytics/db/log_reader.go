package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
)

// LogReader provides READ-ONLY access to logs.entries
type LogReader struct {
	db *pgxpool.Pool
}

// NewLogReader creates a new instance of LogReader.
func NewLogReader(db *pgxpool.Pool) *LogReader {
	return &LogReader{db: db}
}

// CountByServiceAndLevel counts log entries by service and level within a time range
func (r *LogReader) CountByServiceAndLevel(ctx context.Context, service, level string, start, end time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM logs.entries
		WHERE service = $1 AND level = $2 AND created_at BETWEEN $3 AND $4
	`
	var count int
	err := r.db.QueryRow(ctx, query, service, level, start, end).Scan(&count)
	return count, err
}

// FindTopMessages finds most frequent log messages within a time range
func (r *LogReader) FindTopMessages(ctx context.Context, service, level string, start, end time.Time, limit int) ([]models.IssueItem, error) {
	query := `
		SELECT message, COUNT(*) AS count, MAX(created_at) AS last_seen
		FROM logs.entries
		WHERE service = $1 AND level = $2 AND created_at BETWEEN $3 AND $4
		GROUP BY message
		ORDER BY count DESC
		LIMIT $5
	`
	rows, err := r.db.Query(ctx, query, service, level, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []models.IssueItem
	for rows.Next() {
		issue := models.IssueItem{}
		if err := rows.Scan(&issue.Message, &issue.Count, &issue.LastSeen); err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, nil
}

// FindAllServices returns list of all services that have logged
func (r *LogReader) FindAllServices(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT service FROM logs.entries ORDER BY service`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var service string
		if err := rows.Scan(&service); err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	return services, nil
}
