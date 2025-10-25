// Package testutils provides mock implementations for testing analytics services.
package testutils

import (
	"context"
	"log"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/stretchr/testify/mock"
)

// MockAggregationRepository is a mock implementation of AggregationRepositoryInterface.
// It is used for testing purposes.
type MockAggregationRepository struct {
	mock.Mock
}

// FindByRange retrieves mock aggregations within a specified time range.
// It simulates the behavior of the actual repository method.
func (m *MockAggregationRepository) FindByRange(ctx context.Context, metricType models.MetricType, service string, start, end time.Time) ([]*models.Aggregation, error) {
	args := m.Called(ctx, metricType, service, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).([]*models.Aggregation)
	if !ok {
		log.Printf("Unexpected type for FindByRange result")
		return nil, args.Error(1)
	}
	return result, args.Error(1)
}

// FindTopIssues retrieves the top issues based on the specified criteria.
// It simulates the behavior of the actual repository method.
func (m *MockAggregationRepository) FindTopIssues(ctx context.Context, metricType models.MetricType, service string, start, end time.Time, limit int) ([]*models.Aggregation, error) {
	args := m.Called(ctx, metricType, service, start, end, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).([]*models.Aggregation)
	if !ok {
		log.Printf("Unexpected type for FindTopIssues result")
		return nil, args.Error(1)
	}
	return result, args.Error(1)
}

// SaveAggregation saves a mock aggregation.
// It simulates the behavior of the actual repository method.
func (m *MockAggregationRepository) SaveAggregation(ctx context.Context, aggregation *models.Aggregation) error {
	args := m.Called(ctx, aggregation)
	return args.Error(0)
}

// Upsert creates or updates a mock aggregation.
// It simulates the behavior of the actual repository method.
func (m *MockAggregationRepository) Upsert(ctx context.Context, aggregation *models.Aggregation) error {
	args := m.Called(ctx, aggregation)
	return args.Error(0)
}

// FindAllServices returns a list of all services with mock aggregations.
// It simulates the behavior of the actual repository method.
func (m *MockAggregationRepository) FindAllServices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).([]string)
	if !ok {
		log.Printf("Unexpected type for FindAllServices result")
		return nil, args.Error(1)
	}
	return result, args.Error(1)
}
