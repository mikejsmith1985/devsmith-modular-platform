package analytics_services_test

import (
	"bytes"
	"context"
	"testing"

	analytics_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	analytics_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/testutils"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	service1      = "service1"
	service2      = "service2"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
)

func TestAggregatorService_RunHourlyAggregation(t *testing.T) {
	logger, _ := test.NewNullLogger()
	mockAggRepo := new(testutils.MockAggregationRepository)
	mockLogReader := new(testutils.MockLogReader)

	service := analytics_services.NewAggregatorService(mockAggRepo, mockLogReader, logger)

	// Capture logs programmatically
	var logBuffer bytes.Buffer
	logger.SetOutput(&logBuffer)

	logger.SetLevel(logrus.DebugLevel)
	logger.Debug("Logger configured for debug output")

	logger.Infof("Test started: Running RunHourlyAggregation")

	// Mock FindAllServices to return services with levels
	mockLogReader.On("FindAllServices", mock.Anything).Return([]string{"service1", "service2"}, nil).Run(func(args mock.Arguments) {
		logger.Debug("FindAllServices mock invoked")
	})

	// Expand CountByServiceAndLevel mock to cover all log levels for each service
	levels := []string{"info", "warn", "error"}
	for _, service := range []string{"service1", "service2"} {
		for _, level := range levels {
			mockLogReader.On("CountByServiceAndLevel", mock.Anything, service, level, mock.Anything, mock.Anything).
				Return(10, nil).Once()
		}
	}

	// Refine Upsert mock setup to ensure it matches Aggregation objects
	mockAggRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(agg *analytics_models.Aggregation) bool {
		logger.Debugf("Upsert called with aggregation: %+v", agg)
		return (agg.Service == "service1" || agg.Service == "service2") &&
			(agg.Value == 10)
	})).Return(nil).Times(6)

	logger.Debug("Calling RunHourlyAggregation with refined mocks")

	// Execute the actual method under test
	err := service.RunHourlyAggregation(context.Background())

	// Verify no errors
	assert.NoError(t, err, "RunHourlyAggregation should complete without errors")

	logger.Debug("RunHourlyAggregation completed")

	// Print captured logs at the end of the test
	defer func() {
		t.Log("Captured Logs:")
		t.Log(logBuffer.String())
	}()

	mockLogReader.AssertExpectations(t)
	mockAggRepo.AssertExpectations(t)

	// Ensure mock setups are properly scoped within the test function
	mockAggRepo.On("CountByServiceAndLevel", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(
		func(args mock.Arguments) {
			t.Logf("CountByServiceAndLevel called with args: %v", args)
		},
	).Return(10, nil).Twice()

	mockAggRepo.On("Upsert", mock.Anything, mock.Anything).Run(
		func(args mock.Arguments) {
			t.Logf("Upsert called with args: %v", args)
		},
	).Return(nil).Once()

	// Add detailed logs to capture arguments passed to CountByServiceAndLevel
	mockLogReader.On("CountByServiceAndLevel", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(
		func(args mock.Arguments) {
			logger.Debugf("CountByServiceAndLevel called with args: service=%v, level=%v, start=%v, end=%v", args.Get(1), args.Get(2), args.Get(3), args.Get(4))
		},
	).Return(10, nil).Maybe()

	// Add detailed logs to capture arguments passed to Upsert
	mockAggRepo.On("Upsert", mock.Anything, mock.Anything).Run(
		func(args mock.Arguments) {
			logger.Debugf("Upsert called with aggregation: %+v", args.Get(1))
		},
	).Return(nil).Maybe()
}

func TestAggregatorService_AnalyzeAggregations(t *testing.T) {
	mockRepo := new(testutils.MockAggregationRepository)
	logReader := new(testutils.MockLogReader) // Updated to use the mock implementation
	logger, _ := test.NewNullLogger()

	service := analytics_services.NewAggregatorService(mockRepo, logReader, logger)

	// Add mock setup for FindAllServices
	logReader.On("FindAllServices", mock.Anything).Return([]string{"service1", "service2"}, nil)

	// Add mock setup for CountByServiceAndLevel
	logReader.On("CountByServiceAndLevel", mock.Anything, "service1", "info", mock.Anything, mock.Anything).Return(10, nil)
	logReader.On("CountByServiceAndLevel", mock.Anything, "service1", "warn", mock.Anything, mock.Anything).Return(10, nil)
	logReader.On("CountByServiceAndLevel", mock.Anything, "service1", "error", mock.Anything, mock.Anything).Return(10, nil)
	logReader.On("CountByServiceAndLevel", mock.Anything, "service2", "info", mock.Anything, mock.Anything).Return(10, nil)
	logReader.On("CountByServiceAndLevel", mock.Anything, "service2", "warn", mock.Anything, mock.Anything).Return(10, nil)
	logReader.On("CountByServiceAndLevel", mock.Anything, "service2", "error", mock.Anything, mock.Anything).Return(10, nil)

	// Mock setup for Upsert method
	mockRepo.On("Upsert", mock.Anything, mock.AnythingOfType("*analytics_models.Aggregation")).Return(nil).Maybe()

	// Define test cases and assertions here
	t.Log("Invoking RunHourlyAggregation")
	err := service.RunHourlyAggregation(context.Background())
	assert.NoError(t, err, "RunHourlyAggregation should not return an error")

	// Validate that FindAllServices and CountByServiceAndLevel were called
	logReader.AssertCalled(t, "FindAllServices", mock.Anything)
	logReader.AssertCalled(t, "CountByServiceAndLevel", mock.Anything, "service1", "info", mock.Anything, mock.Anything)
	logReader.AssertCalled(t, "CountByServiceAndLevel", mock.Anything, "service1", "warn", mock.Anything, mock.Anything)
	logReader.AssertCalled(t, "CountByServiceAndLevel", mock.Anything, "service1", "error", mock.Anything, mock.Anything)
	logReader.AssertCalled(t, "CountByServiceAndLevel", mock.Anything, "service2", "info", mock.Anything, mock.Anything)
	logReader.AssertCalled(t, "CountByServiceAndLevel", mock.Anything, "service2", "warn", mock.Anything, mock.Anything)
	logReader.AssertCalled(t, "CountByServiceAndLevel", mock.Anything, "service2", "error", mock.Anything, mock.Anything)
}

func TestAggregatorService_FindAllServices_IsCalled(t *testing.T) {
	logger, _ := test.NewNullLogger()
	mockAggRepo := new(testutils.MockAggregationRepository)
	mockLogReader := new(testutils.MockLogReader)

	service := analytics_services.NewAggregatorService(mockAggRepo, mockLogReader, logger)

	mockLogReader.On("FindAllServices", mock.Anything).Return([]string{service1}, nil)
	mockLogReader.On("CountByServiceAndLevel", mock.Anything, service1, logLevelInfo, mock.Anything, mock.Anything).Return(10, nil)
	mockLogReader.On("CountByServiceAndLevel", mock.Anything, service1, logLevelWarn, mock.Anything, mock.Anything).Return(5, nil)
	mockLogReader.On("CountByServiceAndLevel", mock.Anything, service1, logLevelError, mock.Anything, mock.Anything).Return(2, nil)

	mockLogReader.On("CountByServiceAndLevel", mock.Anything, service2, logLevelInfo, mock.Anything, mock.Anything).Return(8, nil)
	mockLogReader.On("CountByServiceAndLevel", mock.Anything, service2, logLevelWarn, mock.Anything, mock.Anything).Return(4, nil)
	mockLogReader.On("CountByServiceAndLevel", mock.Anything, service2, logLevelError, mock.Anything, mock.Anything).Return(1, nil)

	// Add mock setup for the Upsert method
	mockAggRepo.On("Upsert", mock.Anything, mock.Anything).Return(nil).Times(6)

	err := service.RunHourlyAggregation(context.Background())

	assert.NoError(t, err)
	mockLogReader.AssertCalled(t, "FindAllServices", mock.Anything)
}
