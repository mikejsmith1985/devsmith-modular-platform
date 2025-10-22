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

func (m *MockAggregationRepository) FindByRange(ctx context.Context, metricType models.MetricType, service string, start, end time.Time) ([]*models.Aggregation, error) {
	args := m.Called(ctx, metricType, service, start, end)
	return args.Get(0).([]*models.Aggregation), args.Error(1)
}

func (m *MockAggregationRepository) FindTopIssues(ctx context.Context, metricType models.MetricType, service string, start, end time.Time, limit int) ([]*models.Aggregation, error) {
	args := m.Called(ctx, metricType, service, start, end, limit)
	return args.Get(0).([]*models.Aggregation), args.Error(1)
}

func (m *MockAggregationRepository) SaveAggregation(ctx context.Context, aggregation *models.Aggregation) error {
	args := m.Called(ctx, aggregation)
	return args.Error(0)
}

func (m *MockAggregationRepository) Upsert(ctx context.Context, aggregation *models.Aggregation) error {
	args := m.Called(ctx, aggregation)
	return args.Error(0)
}

func (m *MockAggregationRepository) FindAllServices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}
