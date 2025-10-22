// Package services provides utility functions for analytics testing.
package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/stretchr/testify/mock"
)

// MockAggregationRepository is a mock implementation of the AggregationRepository interface.
type MockAggregationRepository struct {
	mock.Mock
}

// FindByRange retrieves aggregations within the specified time range.
func (m *MockAggregationRepository) FindByRange(ctx context.Context, metricType models.MetricType, service string, start, end time.Time) ([]*models.Aggregation, error) {
	args := m.Called(ctx, metricType, service, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	aggregations, ok := args.Get(0).([]*models.Aggregation)
	if !ok {
		return nil, fmt.Errorf("unexpected type for aggregations: %T", args.Get(0))
	}
	return aggregations, args.Error(1)
}

// FindTopIssues retrieves the top issues based on the specified criteria.
func (m *MockAggregationRepository) FindTopIssues(ctx context.Context, metricType models.MetricType, service string, start, end time.Time, limit int) ([]*models.Aggregation, error) {
	args := m.Called(ctx, metricType, service, start, end, limit)
	// Correct the type assertion and handle errors properly.
	result, ok := args.Get(0).([]*models.Aggregation)
	if !ok {
		return nil, fmt.Errorf("type assertion to []*models.Aggregation failed")
	}
	if err := args.Error(1); err != nil {
		log.Printf("Error in MockAggregationRepository.FindTopIssues: %v", err)
		return nil, err
	}
	return result, nil
}
