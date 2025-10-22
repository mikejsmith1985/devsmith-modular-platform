package db

import (
	"context"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
)

type LogReaderInterface interface {
	FindTopMessages(ctx context.Context, service, level string, start, end time.Time, limit int) ([]models.IssueItem, error)
	FindAllServices(ctx context.Context) ([]string, error)
}
