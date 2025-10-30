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

	mock.AssertExpectations(t)
}

func TestMockLogReader(t *testing.T) {
	mock := &MockLogReader{}
	ctx := context.Background()

	// Test FindAllServices
	mock.On("FindAllServices", ctx).Return([]string{"service1"}, nil)
	services, err := mock.FindAllServices(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"service1"}, services)

	// Test CountByServiceAndLevel
	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	mock.On("CountByServiceAndLevel", ctx, "test-service", "error", start, end).Return(5, nil)
	count, err := mock.CountByServiceAndLevel(ctx, "test-service", "error", start, end)
	assert.NoError(t, err)
	assert.Equal(t, 5, count)

	mock.AssertExpectations(t)
}
