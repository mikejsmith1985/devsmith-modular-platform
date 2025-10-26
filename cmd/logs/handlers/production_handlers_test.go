// Package handlers contains HTTP request handlers for the logs service.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockValidationAggregation implements aggregation service interface for testing
type MockValidationAggregation struct {
	mock.Mock
}

func (m *MockValidationAggregation) GetTopErrors(ctx context.Context, service string, limit, days int) ([]models.ValidationError, error) {
	if m.Called(ctx, service, limit, days).Get(0) != nil {
		return m.Called(ctx, service, limit, days).Get(0).([]models.ValidationError), m.Called(ctx, service, limit, days).Error(1)
	}
	return []models.ValidationError{}, nil
}

func (m *MockValidationAggregation) GetErrorTrends(ctx context.Context, service string, days int, interval string) ([]models.ErrorTrend, error) {
	if m.Called(ctx, service, days, interval).Get(0) != nil {
		return m.Called(ctx, service, days, interval).Get(0).([]models.ErrorTrend), m.Called(ctx, service, days, interval).Error(1)
	}
	return []models.ErrorTrend{}, nil
}

// MockAlertThresholdService implements alert service interface for testing
type MockAlertThresholdService struct {
	mock.Mock
}

func (m *MockAlertThresholdService) Create(ctx context.Context, config *models.AlertConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockAlertThresholdService) GetByID(ctx context.Context, service string) (*models.AlertConfig, error) {
	args := m.Called(ctx, service)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AlertConfig), args.Error(1)
}

func (m *MockAlertThresholdService) Update(ctx context.Context, config *models.AlertConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockAlertThresholdService) GetAll(ctx context.Context) ([]models.AlertConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return []models.AlertConfig{}, args.Error(1)
	}
	return args.Get(0).([]models.AlertConfig), args.Error(1)
}

// TestGetDashboardStats_Valid tests retrieving dashboard statistics
func TestGetDashboardStats_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAgg := &MockValidationAggregation{}

	router.GET("/api/logs/dashboard/stats", GetDashboardStats(mockAgg))

	req := httptest.NewRequest("GET", "/api/logs/dashboard/stats?service=review&time_range=last_hour", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotNil(t, resp["total_errors"])
}

// TestGetDashboardStats_InvalidTimeRange tests with invalid time range
func TestGetDashboardStats_InvalidTimeRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAgg := &MockValidationAggregation{}
	router.GET("/api/logs/dashboard/stats", GetDashboardStats(mockAgg))

	req := httptest.NewRequest("GET", "/api/logs/dashboard/stats?time_range=invalid", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetTopErrors_Valid tests retrieving top validation errors
func TestGetTopErrors_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAgg := &MockValidationAggregation{}
	mockAgg.On("GetTopErrors", mock.MatchedBy(func(ctx context.Context) bool { return true }), "review", 10, 7).
		Return([]models.ValidationError{
			{
				ErrorType:        "validation_error",
				Message:          "code exceeds maximum size",
				Count:            245,
				LastOccurrence:   time.Now(),
				AffectedServices: []string{"review"},
			},
		}, nil)

	router.GET("/api/logs/validations/top-errors", GetTopErrors(mockAgg))

	req := httptest.NewRequest("GET", "/api/logs/validations/top-errors?limit=10&days=7", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotNil(t, resp["errors"])
}

// TestGetErrorTrends_Valid tests retrieving error trends
func TestGetErrorTrends_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAgg := &MockValidationAggregation{}
	mockAgg.On("GetErrorTrends", mock.MatchedBy(func(ctx context.Context) bool { return true }), "review", 7, "hourly").
		Return([]models.ErrorTrend{
			{
				Timestamp:        time.Now(),
				ErrorCount:       42,
				ErrorRatePercent: 0.5,
				ByType: map[string]int64{
					"validation_error": 42,
				},
			},
		}, nil)

	router.GET("/api/logs/validations/trends", GetErrorTrends(mockAgg))

	req := httptest.NewRequest("GET", "/api/logs/validations/trends?days=7&interval=hourly", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotNil(t, resp["trend"])
}

// TestExportLogs_JSON tests exporting logs as JSON
func TestExportLogs_JSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock exporter would be provided here
	router.GET("/api/logs/export", ExportLogs())

	req := httptest.NewRequest("GET", "/api/logs/export?format=json&service=review", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

// TestExportLogs_CSV tests exporting logs as CSV
func TestExportLogs_CSV(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/api/logs/export", ExportLogs())

	req := httptest.NewRequest("GET", "/api/logs/export?format=csv&service=review", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
}

// TestExportLogs_InvalidFormat tests export with invalid format
func TestExportLogs_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/api/logs/export", ExportLogs())

	req := httptest.NewRequest("GET", "/api/logs/export?format=invalid", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCreateAlertConfig_Valid tests creating alert configuration
func TestCreateAlertConfig_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := &MockAlertThresholdService{}
	mockService.On("Create", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.MatchedBy(func(config *models.AlertConfig) bool {
		config.ID = 1
		return true
	})).Return(nil)

	router.POST("/api/logs/alert-config", CreateAlertConfig(mockService))

	body := map[string]interface{}{
		"service":                   "review",
		"error_threshold_per_min":   10,
		"warning_threshold_per_min": 5,
		"alert_email":               "admin@example.com",
		"enabled":                   true,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/logs/alert-config", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

// TestGetAlertConfig_Valid tests retrieving alert configuration
func TestGetAlertConfig_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := &MockAlertThresholdService{}
	mockService.On("GetByID", mock.MatchedBy(func(ctx context.Context) bool { return true }), "review").
		Return(&models.AlertConfig{
			ID:                     1,
			Service:                "review",
			ErrorThresholdPerMin:   10,
			WarningThresholdPerMin: 5,
			AlertEmail:             "admin@example.com",
			Enabled:                true,
		}, nil)

	router.GET("/api/logs/alert-config/:service", GetAlertConfig(mockService))

	req := httptest.NewRequest("GET", "/api/logs/alert-config/review", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "review", resp["service"])
}

// TestUpdateAlertConfig_Valid tests updating alert configuration
func TestUpdateAlertConfig_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := &MockAlertThresholdService{}
	mockService.On("Update", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.MatchedBy(func(config *models.AlertConfig) bool { return true })).Return(nil)

	router.PUT("/api/logs/alert-config/:service", UpdateAlertConfig(mockService))

	body := map[string]interface{}{
		"error_threshold_per_min":   15,
		"warning_threshold_per_min": 7,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/api/logs/alert-config/review", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGetAlertEvents tests retrieving triggered alert events
func TestGetAlertEvents_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/api/logs/alert-events", GetAlertEvents())

	req := httptest.NewRequest("GET", "/api/logs/alert-events?service=review&limit=20", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotNil(t, resp["events"])
}
