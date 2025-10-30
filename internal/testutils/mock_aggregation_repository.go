// Package testutils provides mock implementations for testing purposes.
package testutils

import (
	"context"
	"log"
	"time"

	analytics_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/stretchr/testify/mock"
)

// MockAggregationRepository is a mock implementation for testing aggregation repository.
type MockAggregationRepository struct {
	mock.Mock
}

// FindAllServices retrieves all services for testing purposes.
func (m *MockAggregationRepository) FindAllServices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if result, ok := args.Get(0).([]string); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

// FindTopIssues retrieves top issues for testing purposes.
func (m *MockAggregationRepository) FindTopIssues(ctx context.Context, metricType, service string, start, end time.Time, limit int) ([]*analytics_models.Aggregation, error) {
	args := m.Called(ctx, metricType, service, start, end, limit)
	if result, ok := args.Get(0).([]*analytics_models.Aggregation); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

// FindByRange retrieves aggregations within a specified range for testing purposes.
func (m *MockAggregationRepository) FindByRange(ctx context.Context, metricType analytics_models.MetricType, service string, start, end time.Time) ([]*analytics_models.Aggregation, error) {
	args := m.Called(ctx, metricType, service, start, end)
	if result, ok := args.Get(0).([]*analytics_models.Aggregation); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

// Upsert inserts or updates an aggregation for testing purposes.
func (m *MockAggregationRepository) Upsert(ctx context.Context, aggregation *analytics_models.Aggregation) error {
	args := m.Called(ctx, aggregation)
	return args.Error(0)
}

// SetupMockFindByRange sets up the mock for FindByRange method.
func (m *MockAggregationRepository) SetupMockFindByRange() {
	m.On("FindByRange", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*analytics_models.Aggregation{}, nil)
}

// FindByMetric retrieves aggregations by metric for testing purposes.
func (m *MockAggregationRepository) FindByMetric(ctx context.Context, metricType analytics_models.MetricType, service string, start, end time.Time) ([]*analytics_models.Aggregation, error) {
	args := m.Called(ctx, metricType, service, start, end)
	log.Printf("FindByMetric called with: metricType=%v, service=%v, start=%v, end=%v", metricType, service, start, end)
	if result, ok := args.Get(0).([]*analytics_models.Aggregation); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}
