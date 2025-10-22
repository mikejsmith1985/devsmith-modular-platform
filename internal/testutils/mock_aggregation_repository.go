package testutils

import (
	"context"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
)

func (m *MockAggregationRepository) FindAllServices(ctx context.Context) ([]string, error) {
	// Mock implementation
	return nil, nil
}

func (m *MockAggregationRepository) FindByRange(ctx context.Context, metricType models.MetricType, serviceName string, startTime time.Time, endTime time.Time) ([]*models.Aggregation, error) {
	args := m.Called(ctx, metricType, serviceName, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Aggregation), args.Error(1)
}

func (m *MockAggregationRepository) Upsert(ctx context.Context, agg *models.Aggregation) error {
	// Mock implementation
	return nil
}
