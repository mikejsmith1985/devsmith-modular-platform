package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/testutils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAnomalyService_DetectAnomalies(t *testing.T) {
	mockRepo := new(testutils.MockAggregationRepository)
	logger := logrus.New()

	service := services.NewAnomalyService(mockRepo, logger)

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	mockRepo.On("FindByRange", mock.Anything, models.MetricType("error_frequency"), "service1", start, end).Return([]*models.Aggregation{
		{Value: 100, TimeBucket: time.Now().Add(-1 * time.Hour)},
		{Value: 105, TimeBucket: time.Now().Add(-45 * time.Minute)},
		{Value: 400, TimeBucket: time.Now().Add(-30 * time.Minute)},
		{Value: 410, TimeBucket: time.Now().Add(-15 * time.Minute)},
		{Value: 1000, TimeBucket: time.Now().Add(-15 * time.Minute)}, // Extreme anomaly with reduced variance
	}, nil)

	anomalies, err := service.DetectAnomalies(context.Background(), models.MetricType("error_frequency"), "service1", start, end)

	assert.NoError(t, err)
	assert.Greater(t, len(anomalies), 0, "Anomalies should not be empty")
	if len(anomalies) > 0 {
		assert.Equal(t, 1000.0, anomalies[0].Value, "The highest anomaly value should be detected")
		assert.Greater(t, anomalies[0].ZScore, 1.5, "Z-score should exceed threshold")
	}
	mockRepo.AssertExpectations(t)
}
