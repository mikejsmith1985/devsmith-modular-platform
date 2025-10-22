package services_test

import (
	"context"
	"testing"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/testutils"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExportService_ExportData(t *testing.T) {
	mockRepo := new(testutils.MockAggregationRepository)
	logger, _ := test.NewNullLogger()

	service := services.NewExportService(mockRepo, logger)

	mockRepo.On("FindByRange", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*models.Aggregation{
		{MetricType: "error_rate", Service: "service1", Value: 10},
		{MetricType: "error_rate", Service: "service2", Value: 20},
	}, nil)

	err := service.ExportData(context.Background(), "error_rate", "service1", "output.csv")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
