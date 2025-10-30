package testutils

import (
	"context"
	"testing"
	"time"

	analytics_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/stretchr/testify/assert"
)

func TestMockAggregationRepository(t *testing.T) {
	mock := &MockAggregationRepository{}
	ctx := context.Background()

	// Test FindAllServices
	mock.On("FindAllServices", ctx).Return([]string{"service1", "service2"}, nil)
	services, err := mock.FindAllServices(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"service1", "service2"}, services)

	// Test Upsert
	agg := &analytics_models.Aggregation{
		MetricType: analytics_models.ErrorFrequency,
		Service:    "test",
		Value:      100,
		TimeBucket: time.Now(),
	}
	mock.On("Upsert", ctx, agg).Return(nil)
	err = mock.Upsert(ctx, agg)
	assert.NoError(t, err)

	// Test SaveAggregation
	mock.On("SaveAggregation", ctx, agg).Return(nil)
	err = mock.SaveAggregation(ctx, agg)
	assert.NoError(t, err)

	// Test FindByRange
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	mock.On("FindByRange", ctx, analytics_models.ErrorFrequency, "test-service", start, end).Return([]*analytics_models.Aggregation{agg}, nil)
	results, err := mock.FindByRange(ctx, analytics_models.ErrorFrequency, "test-service", start, end)
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	// Test FindTopIssues
	mock.On("FindTopIssues", ctx, analytics_models.ErrorFrequency, "test-service", start, end, 10).Return([]*analytics_models.Aggregation{agg}, nil)
	topIssues, err := mock.FindTopIssues(ctx, analytics_models.ErrorFrequency, "test-service", start, end, 10)
	assert.NoError(t, err)
	assert.Len(t, topIssues, 1)

	mock.AssertExpectations(t)
}
