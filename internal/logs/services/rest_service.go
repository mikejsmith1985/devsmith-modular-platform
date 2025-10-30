// Package logs_services provides service implementations for logs operations.
package logs_services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	logs_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	"github.com/sirupsen/logrus"
)

// RestLogService implements REST API operations for logs.
type RestLogService struct {
	repo   *logs_db.LogRepository
	logger *logrus.Logger
}

// NewRestLogService creates a new RestLogService.
func NewRestLogService(repo *logs_db.LogRepository, logger *logrus.Logger) *RestLogService {
	return &RestLogService{
		repo:   repo,
		logger: logger,
	}
}

// Insert creates a new log entry.
func (s *RestLogService) Insert(ctx context.Context, entry map[string]interface{}) (int64, error) {
	if s.repo == nil {
		return 0, errors.New("repository not configured")
	}

	logEntry := &logs_db.LogEntry{
		Service:   extractString(entry, "service"),
		Level:     extractString(entry, "level"),
		Message:   extractString(entry, "message"),
		Metadata:  extractMetadata(entry, "metadata"),
		CreatedAt: time.Now(),
	}

	id, err := s.repo.Save(ctx, logEntry)
	if err != nil {
		return 0, fmt.Errorf("insert failed: %w", err)
	}

	return id, nil
}

// Query retrieves logs with optional filters and pagination.
func (s *RestLogService) Query(
	ctx context.Context,
	filters map[string]interface{},
	page map[string]int,
) ([]interface{}, error) {
	if s.repo == nil {
		return nil, errors.New("repository not configured")
	}

	limit := 100
	if l, ok := page["limit"]; ok && l > 0 && l <= 1000 {
		limit = l
	}
	offset := 0
	if o, ok := page["offset"]; ok && o >= 0 {
		offset = o
	}

	queryFilters := &logs_db.QueryFilters{
		Service: extractString(filters, "service"),
		Level:   extractString(filters, "level"),
		Search:  extractString(filters, "search"),
		From:    parseTime(extractString(filters, "from")),
		To:      parseTime(extractString(filters, "to")),
	}

	pageOpts := logs_db.PageOptions{
		Limit:  limit,
		Offset: offset,
	}

	entries, err := s.repo.Query(ctx, queryFilters, pageOpts)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	result := make([]interface{}, len(entries))
	for i, entry := range entries {
		result[i] = mapLogEntryToInterface(entry)
	}

	return result, nil
}

// GetByID retrieves a single log entry by ID.
func (s *RestLogService) GetByID(ctx context.Context, id int64) (interface{}, error) {
	if s.repo == nil {
		return nil, errors.New("repository not configured")
	}

	entry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get by id failed: %w", err)
	}

	return mapLogEntryToInterface(entry), nil
}

// Stats returns aggregated log statistics.
func (s *RestLogService) Stats(ctx context.Context) (map[string]interface{}, error) {
	if s.repo == nil {
		return nil, errors.New("repository not configured")
	}

	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get stats failed: %w", err)
	}

	// Convert stats map to interface{} map
	return stats, nil
}

// DeleteByID deletes a log entry by ID.
func (s *RestLogService) DeleteByID(ctx context.Context, id int64) error {
	if s.repo == nil {
		return errors.New("repository not configured")
	}

	// LogRepository doesn't support single delete yet
	return errors.New("delete by ID not supported")
}

// Delete deletes logs matching filters.
func (s *RestLogService) Delete(ctx context.Context, filters map[string]interface{}) (int64, error) {
	if s.repo == nil {
		return 0, errors.New("repository not configured")
	}

	// LogRepository only supports DeleteOld with a timestamp
	// For now, return not implemented
	return 0, errors.New("delete by filters not supported")
}

// Helper functions

func extractString(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func extractMetadata(data map[string]interface{}, key string) map[string]interface{} {
	if v, ok := data[key]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return make(map[string]interface{})
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}

	// Try to parse as RFC3339
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}

	// Try to parse as Unix timestamp
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(n, 0)
	}

	return time.Time{}
}

func mapLogEntryToInterface(entry *logs_db.LogEntry) map[string]interface{} {
	return map[string]interface{}{
		"id":         entry.ID,
		"service":    entry.Service,
		"level":      entry.Level,
		"message":    entry.Message,
		"metadata":   entry.Metadata,
		"created_at": entry.CreatedAt,
	}
}
