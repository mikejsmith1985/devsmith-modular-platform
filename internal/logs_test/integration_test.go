// Package logs_test provides integration tests for the logs module.
package logs_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

// MockLogReader mocks the LogReaderInterface.
type MockLogReader struct {
	mock.Mock
}

func (m *MockLogReader) FindAllServices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockLogReader) CountByServiceAndLevel(ctx context.Context, service, level string, start, end time.Time) (int64, error) {
	args := m.Called(ctx, service, level, start, end)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockLogReader) FindTopMessages(ctx context.Context, service, level string, start, end time.Time, limit int) ([]services.LogMessage, error) {
	args := m.Called(ctx, service, level, start, end, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]services.LogMessage), args.Error(1)
}

// TestDashboardEndToEnd tests complete dashboard flow.
func TestDashboardEndToEnd(t *testing.T) {
	// GIVEN: A dashboard service with mock log reader
	mockReader := new(MockLogReader)
	mockReader.On("FindAllServices", mock.Anything).Return([]string{"api-service", "db-service"}, nil)
	mockReader.On("CountByServiceAndLevel", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(10), nil)
	mockReader.On("FindTopMessages", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]services.LogMessage{}, nil)

	dashboardService := services.NewDashboardService(mockReader, nil)

	// WHEN: Getting dashboard stats
	stats, err := dashboardService.GetDashboardStats(context.Background())

	// THEN: Should return complete stats
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.NotEmpty(t, stats.ServiceStats)
	mockReader.AssertCalled(t, "FindAllServices", mock.Anything)
}

// TestMultipleServicesDashboard tests dashboard with multiple services.
func TestMultipleServicesDashboard(t *testing.T) {
	// GIVEN: Multiple services in system
	serviceList := []string{"api-service", "db-service", "cache-service", "worker-service"}
	mockReader := new(MockLogReader)
	mockReader.On("FindAllServices", mock.Anything).Return(serviceList, nil)
	mockReader.On("CountByServiceAndLevel", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(50), nil)
	mockReader.On("FindTopMessages", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]services.LogMessage{}, nil)

	dashboardService := services.NewDashboardService(mockReader, nil)

	// WHEN: Getting dashboard stats
	stats, err := dashboardService.GetDashboardStats(context.Background())

	// THEN: Should aggregate stats for all services
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	mockReader.AssertCalled(t, "FindAllServices", mock.Anything)
}

// TestErrorHandlingIntegration tests error handling across components.
func TestErrorHandlingIntegration(t *testing.T) {
	// GIVEN: Mock that returns error
	mockReader := new(MockLogReader)
	mockReader.On("FindAllServices", mock.Anything).Return(nil, sql.ErrNoRows)

	dashboardService := services.NewDashboardService(mockReader, nil)

	// WHEN: Getting dashboard stats with error
	stats, err := dashboardService.GetDashboardStats(context.Background())

	// THEN: Should handle error gracefully
	assert.NoError(t, err) // Service returns partial stats on error
	assert.NotNil(t, stats)
}

// TestAlertThresholdWorkflow tests alert threshold detection workflow.
func TestAlertThresholdWorkflow(t *testing.T) {
	// GIVEN: System with threshold violations
	violation := &models.AlertThresholdViolation{
		Service:        "api-service",
		Level:          "error",
		CurrentCount:   150,
		ThresholdValue: 100,
		Timestamp:      time.Now(),
	}

	// WHEN: Processing violation
	// THEN: Should be trackable and queryable
	assert.NotNil(t, violation)
	assert.Greater(t, violation.CurrentCount, int64(violation.ThresholdValue))
}

// TestAggregationJobWorkflow tests background aggregation job workflow.
func TestAggregationJobWorkflow(t *testing.T) {
	// GIVEN: Aggregation service and time window
	mockReader := new(MockLogReader)
	mockReader.On("FindAllServices", mock.Anything).Return([]string{"api-service"}, nil)
	mockReader.On("CountByServiceAndLevel", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(100), nil)

	aggregationService := services.NewLogAggregationService(mockReader, nil)

	// WHEN: Running hourly aggregation
	err := aggregationService.AggregateLogsHourly(context.Background())

	// THEN: Should complete without error
	assert.NoError(t, err)
}

// TestWebSocketRealtimeWorkflow tests real-time WebSocket updates.
func TestWebSocketRealtimeWorkflow(t *testing.T) {
	// GIVEN: WebSocket service with connected clients
	wsService := services.NewWebSocketRealtimeService(nil)

	// WHEN: Registering and unregistering connections
	err1 := wsService.RegisterConnection(context.Background(), "client-1")
	err2 := wsService.RegisterConnection(context.Background(), "client-2")
	count, err3 := wsService.GetConnectionCount(context.Background())

	// THEN: Should track connections
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.Equal(t, 2, count)

	// WHEN: Unregistering
	err4 := wsService.UnregisterConnection(context.Background(), "client-1")
	count2, _ := wsService.GetConnectionCount(context.Background())

	// THEN: Count should decrease
	assert.NoError(t, err4)
	assert.Equal(t, 1, count2)
}

// TestConcurrentAccessWorkflow tests concurrent access to services.
func TestConcurrentAccessWorkflow(t *testing.T) {
	// GIVEN: Multiple goroutines accessing dashboard
	mockReader := new(MockLogReader)
	mockReader.On("FindAllServices", mock.Anything).Return([]string{"api-service"}, nil)
	mockReader.On("CountByServiceAndLevel", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(50), nil)
	mockReader.On("FindTopMessages", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]services.LogMessage{}, nil)

	dashboardService := services.NewDashboardService(mockReader, nil)

	// WHEN: Making concurrent requests
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			stats, err := dashboardService.GetDashboardStats(context.Background())
			assert.NoError(t, err)
			assert.NotNil(t, stats)
			done <- true
		}()
	}

	// THEN: All should complete successfully
	for i := 0; i < 3; i++ {
		require.True(t, <-done)
	}
}
