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

func TestTopIssuesService_GetTopIssues(t *testing.T) {
	mockRepo := &testutils.MockLogReader{}
	logger, _ := test.NewNullLogger()

	// Add log to verify service initialization
	logger.Debug("Initializing TopIssuesService")
	service := services.NewTopIssuesService(mockRepo, logger)

	// Simplify mock setup to use mock.Anything for all arguments
	// Add log to verify mock setup
	logger.Debug("Setting up mock for FindTopMessages")
	mockRepo.On("FindTopMessages", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]models.IssueItem{
			{Service: "service1", Level: "error", Message: "Error 1", Count: int(15), Value: 0.0, LastSeen: time.Now()},
			{Service: "service2", Level: "error", Message: "Error 2", Count: int(10), Value: 0.0, LastSeen: time.Now()},
		}, nil).Run(func(args mock.Arguments) {
		// Add log to mock method to confirm invocation
		logger.Debugf("Mock FindTopMessages invoked with args: %v", args)
	})

	// Add mock setup for FindAllServices
	logger.Debug("Setting up mock for FindAllServices")
	mockRepo.On("FindAllServices", mock.Anything).
		Return([]string{"service1", "service2"}, nil).Run(func(args mock.Arguments) {
		logger.Debugf("Mock FindAllServices invoked with args: %v", args)
	})

	// Add debugging log to confirm mock invocation
	logger.Debug("Starting TestTopIssuesService_GetTopIssues")

	startTime := time.Date(2025, 10, 20, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 10, 21, 0, 0, 0, 0, time.UTC)

	// Add log for input parameters
	logger.Infof("Calling GetTopIssues with metricType: %s, service: %s, startTime: %v, endTime: %v, limit: %d", "error_rate", "error", startTime, endTime, 5)

	// Add debug logs to verify input parameters
	logger.Debugf("Test: Calling GetTopIssues with start=%v, end=%v, limit=%d", startTime, endTime, 5)

	// Call the method
	topIssues, err := service.GetTopIssues(context.Background(), "error_rate", "error", startTime, endTime, 5)

	// Log the output for debugging
	if err != nil {
		logger.Errorf("Test: GetTopIssues returned error: %v", err)
	} else {
		logger.Infof("Test: GetTopIssues returned: %v", topIssues)
	}

	// Add assertion to verify mock invocation
	mockRepo.AssertCalled(t, "FindTopMessages", mock.Anything, "error_rate", "error", mock.Anything, mock.Anything, 5)

	// Existing assertions
	assert.NoError(t, err)
	assert.Len(t, topIssues, 2)
	assert.Equal(t, "service1", topIssues[0].Service)
	assert.Equal(t, 15, topIssues[0].Count)
	assert.Equal(t, "service2", topIssues[1].Service)
	assert.Equal(t, 10, topIssues[1].Count)

	mockRepo.AssertExpectations(t)
}
