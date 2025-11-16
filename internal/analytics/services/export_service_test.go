package analytics_services_test

import (
	"context"
	"os"
	"testing"

	analytics_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	analytics_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/testutils"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExportService_ExportData(t *testing.T) {
	mockRepo := new(testutils.MockAggregationRepository)
	logger, _ := test.NewNullLogger()

	service := analytics_services.NewExportService(mockRepo, logger)

	mockRepo.On("FindByRange", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*analytics_models.Aggregation{
		{MetricType: "error_rate", Service: "service1", Value: 10},
		{MetricType: "error_rate", Service: "service2", Value: 20},
	}, nil)

	// Use the required safe directory path (hardcoded in export_service.go isValidFilePath)
	// In CI, this test may fail if the directory cannot be created - that's a known limitation
	dir := "/safe/export/directory"
	
	// Try to create the directory, skip test if it fails (CI restriction)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Skipf("Cannot create export directory %s (expected in CI): %v", dir, err)
		return
	}
	
	// Clean up after test
	defer os.RemoveAll(dir)
	
	err := service.ExportData(context.Background(), "error_rate", "service1", dir+"/output.csv")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
