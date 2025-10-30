// Package analytics_services provides utility functions for analytics testing.
package analytics_services

import (
	"context"
	"fmt"
	"log"
	"time"

	analytics_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/stretchr/testify/mock"
)

// MockAggregationRepository is a mock implementation of the AggregationRepository interface.
type MockAggregationRepository struct {
	mock.Mock
}

// FindByRange retrieves aggregations within the specified time range.
func (m *MockAggregationRepository) FindByRange(ctx context.Context, metricType analytics_models.MetricType, service string, start, end time.Time) ([]*analytics_models.Aggregation, error) {
	args := m.Called(ctx, metricType, service, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	aggregations, ok := args.Get(0).([]*analytics_models.Aggregation)
	if !ok {
		return nil, fmt.Errorf("unexpected type for aggregations: %T", args.Get(0))
	}
	return aggregations, args.Error(1)
}

// FindTopIssues retrieves the top issues based on the specified criteria.
func (m *MockAggregationRepository) FindTopIssues(ctx context.Context, metricType analytics_models.MetricType, service string, start, end time.Time, limit int) ([]*analytics_models.Aggregation, error) {
	args := m.Called(ctx, metricType, service, start, end, limit)
	// Correct the type assertion and handle errors properly.
	result, ok := args.Get(0).([]*analytics_models.Aggregation)
	if !ok {
		return nil, fmt.Errorf("type assertion to []*analytics_models.Aggregation failed")
	}
	if err := args.Error(1); err != nil {
		log.Printf("Error in MockAggregationRepository.FindTopIssues: %v", err)
		return nil, err
	}
	return result, nil
}
