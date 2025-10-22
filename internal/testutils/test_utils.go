package testutils

import (
	"context"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/stretchr/testify/mock"
)

type MockAggregationRepository struct {
	mock.Mock
}

func (m *MockAggregationRepository) FindTopIssues(ctx context.Context, metricType string, service string, start, end time.Time, limit int) ([]*models.Aggregation, error) {
	args := m.Called(ctx, metricType, service, start, end, limit)
	return args.Get(0).([]*models.Aggregation), args.Error(1)
}
