// Package service provides business logic for log operations.
package service

import (
	"context"
)

// Repository defines the interface for log persistence.
type Repository interface {
	Insert(ctx context.Context, entry interface{}) (int64, error)
	Query(ctx context.Context, filters interface{}, page interface{}) ([]interface{}, error)
	GetByID(ctx context.Context, id int64) (interface{}, error)
	DeleteByID(ctx context.Context, id int64) error
	DeleteBefore(ctx context.Context, ts interface{}) (int64, error)
	Stats(ctx context.Context) (map[string]interface{}, error)
}

// LogService provides business logic for logs.
type LogService struct {
	repo Repository
}

// NewLogService creates a new LogService.
func NewLogService(repo Repository) *LogService {
	return &LogService{repo: repo}
}

// Insert adds a new log entry.
func (s *LogService) Insert(ctx context.Context, entry map[string]interface{}) (int64, error) {
	return s.repo.Insert(ctx, entry)
}

// Query retrieves logs matching filters.
func (s *LogService) Query(ctx context.Context, filters map[string]interface{}, page map[string]int) ([]interface{}, error) {
	return s.repo.Query(ctx, filters, page)
}

// GetByID retrieves a single log entry.
func (s *LogService) GetByID(ctx context.Context, id int64) (interface{}, error) {
	return s.repo.GetByID(ctx, id)
}

// Stats returns aggregated log statistics.
func (s *LogService) Stats(ctx context.Context) (map[string]interface{}, error) {
	// Call repository to get stats
	return s.repo.Stats(ctx)
}

// DeleteByID removes a log entry.
func (s *LogService) DeleteByID(ctx context.Context, id int64) error {
	return s.repo.DeleteByID(ctx, id)
}

// Delete removes logs matching criteria.
func (s *LogService) Delete(ctx context.Context, filters map[string]interface{}) (int64, error) {
	return s.repo.DeleteBefore(ctx, filters["before"])
}
