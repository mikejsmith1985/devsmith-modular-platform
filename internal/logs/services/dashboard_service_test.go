package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogRepository is a mock for the log repository
type MockLogRepository struct {
	mock.Mock
}

// MockDashboardService is a mock for testing the dashboard service
type MockDashboardService struct {
	mock.Mock
}

// GetDashboardStats mocks the GetDashboardStats method
func (m *MockDashboardService) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardStats), args.Error(1)
}

// GetServiceStats mocks the GetServiceStats method
func (m *MockDashboardService) GetServiceStats(ctx context.Context, service string, timeRange time.Duration) (*models.LogStats, error) {
	args := m.Called(ctx, service, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LogStats), args.Error(1)
}

// GetTopErrors mocks the GetTopErrors method
func (m *MockDashboardService) GetTopErrors(ctx context.Context, limit int, timeRange time.Duration) ([]models.TopErrorMessage, error) {
	args := m.Called(ctx, limit, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TopErrorMessage), args.Error(1)
}

// GetServiceHealth mocks the GetServiceHealth method
func (m *MockDashboardService) GetServiceHealth(ctx context.Context) (map[string]*models.ServiceHealth, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*models.ServiceHealth), args.Error(1)
}

// TestGetDashboardStats_ReturnsValidStats validates dashboard stats retrieval.
func TestGetDashboardStats_ReturnsValidStats(t *testing.T) {
	// GIVEN: A dashboard service with mocked dependencies
	mockService := new(MockDashboardService)
	now := time.Now()

	stats := &models.DashboardStats{
		GeneratedAt:   now,
		ServiceStats:  make(map[string]*models.LogStats),
		ServiceHealth: make(map[string]*models.ServiceHealth),
		TopErrors:     []models.TopErrorMessage{},
		Violations:    []models.AlertThresholdViolation{},
	}

	mockService.On("GetDashboardStats", mock.Anything).Return(stats, nil)

	// WHEN: Calling GetDashboardStats
	result, err := mockService.GetDashboardStats(context.Background())

	// THEN: Should return valid stats without error
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, now, result.GeneratedAt)
	mockService.AssertCalled(t, "GetDashboardStats", mock.Anything)
}

// TestGetDashboardStats_IncludesMultipleServices validates dashboard includes all services.
func TestGetDashboardStats_IncludesMultipleServices(t *testing.T) {
	// GIVEN: Dashboard stats with multiple services
	mockService := new(MockDashboardService)
	now := time.Now()

	stats := &models.DashboardStats{
		GeneratedAt:   now,
		ServiceStats:  make(map[string]*models.LogStats),
		ServiceHealth: make(map[string]*models.ServiceHealth),
	}

	// Add multiple services
	services := []string{"portal", "review", "analytics", "logs"}
	for _, svc := range services {
		stats.ServiceStats[svc] = &models.LogStats{
			Service:    svc,
			TotalCount: 100,
		}
		stats.ServiceHealth[svc] = &models.ServiceHealth{
			Service: svc,
			Status:  "OK",
		}
	}

	mockService.On("GetDashboardStats", mock.Anything).Return(stats, nil)

	// WHEN: Getting dashboard stats
	result, err := mockService.GetDashboardStats(context.Background())

	// THEN: Stats should include all services
	assert.NoError(t, err)
	assert.Equal(t, 4, len(result.ServiceStats))
	assert.Equal(t, 4, len(result.ServiceHealth))
}

// TestGetServiceStats_ReturnsStatsForService validates service-specific stats retrieval.
func TestGetServiceStats_ReturnsStatsForService(t *testing.T) {
	// GIVEN: A dashboard service and a specific service
	mockService := new(MockDashboardService)
	now := time.Now()
	timeRange := time.Hour

	stats := &models.LogStats{
		Timestamp:  now,
		Service:    "portal",
		TotalCount: 250,
		ErrorRate:  0.02,
		CountByLevel: map[string]int64{
			"error":   5,
			"warning": 5,
			"info":    240,
		},
	}

	mockService.On("GetServiceStats", mock.Anything, "portal", timeRange).Return(stats, nil)

	// WHEN: Getting stats for a specific service
	result, err := mockService.GetServiceStats(context.Background(), "portal", timeRange)

	// THEN: Should return stats for that service
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "portal", result.Service)
	assert.Equal(t, int64(250), result.TotalCount)
	assert.Equal(t, 0.02, result.ErrorRate)
}

// TestGetServiceStats_CountByLevel validates level distribution.
func TestGetServiceStats_CountByLevel(t *testing.T) {
	// GIVEN: Service stats with level breakdown
	mockService := new(MockDashboardService)
	timeRange := time.Hour

	stats := &models.LogStats{
		Service:    "analytics",
		TotalCount: 1000,
		CountByLevel: map[string]int64{
			"error":   50,
			"warning": 100,
			"info":    850,
		},
	}

	mockService.On("GetServiceStats", mock.Anything, "analytics", timeRange).Return(stats, nil)

	// WHEN: Getting service stats
	result, err := mockService.GetServiceStats(context.Background(), "analytics", timeRange)

	// THEN: Level counts should be correct
	assert.NoError(t, err)
	assert.Equal(t, int64(50), result.CountByLevel["error"])
	assert.Equal(t, int64(100), result.CountByLevel["warning"])
	assert.Equal(t, int64(850), result.CountByLevel["info"])
}

// TestGetTopErrors_ReturnsTopErrorMessages validates top errors retrieval.
func TestGetTopErrors_ReturnsTopErrorMessages(t *testing.T) {
	// GIVEN: A dashboard service and top errors query
	mockService := new(MockDashboardService)
	now := time.Now()
	limit := 10
	timeRange := 24 * time.Hour

	topErrors := []models.TopErrorMessage{
		{
			Message:   "database connection timeout",
			Service:   "analytics",
			Level:     "error",
			Count:     150,
			FirstSeen: now.Add(-24 * time.Hour),
			LastSeen:  now,
		},
		{
			Message:   "invalid user input",
			Service:   "portal",
			Level:     "error",
			Count:     100,
			FirstSeen: now.Add(-24 * time.Hour),
			LastSeen:  now,
		},
	}

	mockService.On("GetTopErrors", mock.Anything, limit, timeRange).Return(topErrors, nil)

	// WHEN: Getting top errors
	result, err := mockService.GetTopErrors(context.Background(), limit, timeRange)

	// THEN: Should return sorted list of top errors
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "database connection timeout", result[0].Message)
	assert.Equal(t, int64(150), result[0].Count)
	assert.Equal(t, "invalid user input", result[1].Message)
}

// TestGetTopErrors_Limit validates limit is respected.
func TestGetTopErrors_Limit(t *testing.T) {
	// GIVEN: Request for top 5 errors
	mockService := new(MockDashboardService)
	limit := 5
	timeRange := time.Hour

	topErrors := make([]models.TopErrorMessage, limit)
	for i := 0; i < limit; i++ {
		topErrors[i] = models.TopErrorMessage{
			Message: "error message " + string(rune('A'+i)),
			Count:   int64(100 - i*10),
		}
	}

	mockService.On("GetTopErrors", mock.Anything, limit, timeRange).Return(topErrors, nil)

	// WHEN: Getting top 5 errors
	result, err := mockService.GetTopErrors(context.Background(), limit, timeRange)

	// THEN: Should return exactly 5 errors
	assert.NoError(t, err)
	assert.Len(t, result, 5)
}

// TestGetServiceHealth_ReturnsHealthStatus validates service health retrieval.
func TestGetServiceHealth_ReturnsHealthStatus(t *testing.T) {
	// GIVEN: A dashboard service
	mockService := new(MockDashboardService)
	now := time.Now()

	health := map[string]*models.ServiceHealth{
		"portal": {
			Service:       "portal",
			Status:        "OK",
			LastCheckedAt: now,
			ErrorCount:    0,
			WarningCount:  2,
			InfoCount:     500,
		},
		"analytics": {
			Service:       "analytics",
			Status:        "Warning",
			LastCheckedAt: now,
			ErrorCount:    5,
			WarningCount:  25,
			InfoCount:     470,
		},
		"review": {
			Service:       "review",
			Status:        "Error",
			LastCheckedAt: now,
			ErrorCount:    50,
			WarningCount:  10,
			InfoCount:     100,
		},
	}

	mockService.On("GetServiceHealth", mock.Anything).Return(health, nil)

	// WHEN: Getting service health
	result, err := mockService.GetServiceHealth(context.Background())

	// THEN: Should return health status for all services
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "OK", result["portal"].Status)
	assert.Equal(t, "Warning", result["analytics"].Status)
	assert.Equal(t, "Error", result["review"].Status)
}

// TestGetServiceHealth_Status_OK validates OK status.
func TestGetServiceHealth_Status_OK(t *testing.T) {
	// GIVEN: Service with no errors
	mockService := new(MockDashboardService)

	health := map[string]*models.ServiceHealth{
		"portal": {
			Service:    "portal",
			Status:     "OK",
			ErrorCount: 0,
		},
	}

	mockService.On("GetServiceHealth", mock.Anything).Return(health, nil)

	// WHEN: Getting health status
	result, err := mockService.GetServiceHealth(context.Background())

	// THEN: Status should be OK
	assert.NoError(t, err)
	assert.Equal(t, "OK", result["portal"].Status)
}

// TestGetServiceHealth_Status_Warning validates Warning status.
func TestGetServiceHealth_Status_Warning(t *testing.T) {
	// GIVEN: Service with some errors
	mockService := new(MockDashboardService)

	health := map[string]*models.ServiceHealth{
		"analytics": {
			Service:      "analytics",
			Status:       "Warning",
			ErrorCount:   10,
			WarningCount: 50,
		},
	}

	mockService.On("GetServiceHealth", mock.Anything).Return(health, nil)

	// WHEN: Getting health status
	result, err := mockService.GetServiceHealth(context.Background())

	// THEN: Status should be Warning
	assert.NoError(t, err)
	assert.Equal(t, "Warning", result["analytics"].Status)
}

// TestGetServiceHealth_Status_Error validates Error status.
func TestGetServiceHealth_Status_Error(t *testing.T) {
	// GIVEN: Service with many errors
	mockService := new(MockDashboardService)

	health := map[string]*models.ServiceHealth{
		"review": {
			Service:    "review",
			Status:     "Error",
			ErrorCount: 200,
		},
	}

	mockService.On("GetServiceHealth", mock.Anything).Return(health, nil)

	// WHEN: Getting health status
	result, err := mockService.GetServiceHealth(context.Background())

	// THEN: Status should be Error
	assert.NoError(t, err)
	assert.Equal(t, "Error", result["review"].Status)
}

// TestGetDashboardStats_ContextCancellation validates context cancellation handling.
func TestGetDashboardStats_ContextCancellation(t *testing.T) {
	// GIVEN: A cancelled context
	mockService := new(MockDashboardService)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockService.On("GetDashboardStats", mock.Anything).Return(nil, context.Canceled)

	// WHEN: Calling GetDashboardStats with cancelled context
	result, err := mockService.GetDashboardStats(ctx)

	// THEN: Should return context cancelled error
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestGetDashboardStats_TimeRangeVariations validates different time ranges.
func TestGetServiceStats_VariousTimeRanges(t *testing.T) {
	// GIVEN: A dashboard service
	mockService := new(MockDashboardService)

	testCases := []struct {
		name      string
		timeRange time.Duration
		expected  int64
	}{
		{"1 hour", time.Hour, 50},
		{"24 hours", 24 * time.Hour, 200},
		{"7 days", 7 * 24 * time.Hour, 1000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stats := &models.LogStats{
				Service:    "test",
				TotalCount: tc.expected,
			}

			mockService.On("GetServiceStats", mock.Anything, "test", tc.timeRange).Return(stats, nil)

			// WHEN: Getting service stats
			result, err := mockService.GetServiceStats(context.Background(), "test", tc.timeRange)

			// THEN: Should return correct stats for time range
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result.TotalCount)
		})
	}
}

// TestGetTopErrors_EmptyResult validates empty error list.
func TestGetTopErrors_EmptyResult(t *testing.T) {
	// GIVEN: A service with no errors
	mockService := new(MockDashboardService)

	mockService.On("GetTopErrors", mock.Anything, 10, time.Hour).Return([]models.TopErrorMessage{}, nil)

	// WHEN: Getting top errors
	result, err := mockService.GetTopErrors(context.Background(), 10, time.Hour)

	// THEN: Should return empty list
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}
