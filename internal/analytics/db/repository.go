// Package db provides interfaces for database operations in the analytics service.
package db

import (
	"context"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
)

// LogReaderInterface defines methods for reading log data from the database.
// It includes methods for retrieving top messages and all services.
type LogReaderInterface interface {
	FindTopMessages(ctx context.Context, service, level string, start, end time.Time, limit int) ([]models.IssueItem, error)
	FindAllServices(ctx context.Context) ([]string, error)
}
