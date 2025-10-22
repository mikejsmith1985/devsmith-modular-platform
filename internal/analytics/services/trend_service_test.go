package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/testutils"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTrendService_AnalyzeTrends(t *testing.T) {
	logger, _ := test.NewNullLogger()
	mockAggRepo := &testutils.MockAggregationRepository{}

	service := services.NewTrendService(mockAggRepo, logger)

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	mockAggRepo.On("FindByRange", mock.Anything, models.MetricType("log_count"), "service1", start, end).Return([]*models.Aggregation{
		{
			MetricType: "log_count",
			Service:    "service1",
			Value:      10.0,
			TimeBucket: time.Now().Add(-30 * time.Minute),
			CreatedAt:  time.Now(),
		},
		{
			MetricType: "log_count",
			Service:    "service1",
			Value:      20.0,
			TimeBucket: time.Now().Add(-20 * time.Minute),
			CreatedAt:  time.Now(),
		},
		{
			MetricType: "log_count",
			Service:    "service1",
			Value:      30.0,
			TimeBucket: time.Now().Add(-10 * time.Minute),
			CreatedAt:  time.Now(),
		},
	}, nil)

	trend, err := service.AnalyzeTrends(context.Background(), models.MetricType("log_count"), "service1", start, end)

	assert.NoError(t, err)
	assert.NotNil(t, trend)
	assert.Equal(t, "increasing", trend.Direction)
	assert.Equal(t, 20.0, trend.Change)
	mockAggRepo.AssertExpectations(t)
}

func TestTrendService_GetTrends(t *testing.T) {
	mockRepo := new(testutils.MockAggregationRepository)
	logger, _ := test.NewNullLogger()

	service := services.NewTrendService(mockRepo, logger)

	startTime := time.Date(2025, 10, 20, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 10, 21, 0, 0, 0, 0, time.UTC)

	mockRepo.On("FindByMetric", mock.Anything, models.MetricType("error_rate"), "service", startTime, endTime).Return([]models.Aggregation{
		{MetricType: "error_rate", Service: "service1", Value: 5},
		{MetricType: "error_rate", Service: "service2", Value: 10},
	}, nil)

	trends, err := service.GetTrends(context.Background(), models.MetricType("error_rate"), "service", startTime, endTime)

	assert.NoError(t, err)
	assert.NotNil(t, trends.Trend)
	assert.Equal(t, "stable", trends.Trend.Direction)
	assert.Equal(t, 0.95, trends.Trend.Confidence)
	assert.Contains(t, trends.Trend.Summary, "Trend analysis completed successfully.")

	mockRepo.AssertExpectations(t)
}
