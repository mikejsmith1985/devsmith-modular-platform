package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewAnalyticsHandler tests handler initialization
func TestNewAnalyticsHandler(t *testing.T) {
	logger := logrus.New()
	agg := &services.AggregatorService{}
	trend := &services.TrendService{}
	anom := &services.AnomalyService{}
	top := &services.TopIssuesService{}
	exp := &services.ExportService{}

	handler := NewAnalyticsHandler(agg, trend, anom, top, exp, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, agg, handler.aggregatorService)
	assert.Equal(t, trend, handler.trendService)
	assert.Equal(t, anom, handler.anomalyService)
	assert.Equal(t, top, handler.topIssuesService)
	assert.Equal(t, exp, handler.exportService)
	assert.Equal(t, logger, handler.logger)
}

// TestRegisterRoutes tests that routes are registered correctly
func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	agg := &services.AggregatorService{}
	trend := &services.TrendService{}
	anom := &services.AnomalyService{}
	top := &services.TopIssuesService{}
	exp := &services.ExportService{}

	handler := NewAnalyticsHandler(agg, trend, anom, top, exp, logger)
	router := gin.New()

	// Should not panic
	require.NotPanics(t, func() {
		handler.RegisterRoutes(router)
	})
}

// TestGetTrends tests the trends endpoint
func TestGetTrends(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	handler := &AnalyticsHandler{
		trendService: &services.TrendService{},
		logger:       logger,
	}

	router := gin.New()
	router.GET("/api/analytics/trends", handler.GetTrends)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/trends?time_range=24h", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should either return 200 or empty response (implementation is incomplete)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNoContent)
}

// TestGetAnomalies tests the anomalies endpoint
func TestGetAnomalies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	handler := &AnalyticsHandler{
		anomalyService: &services.AnomalyService{},
		logger:         logger,
	}

	router := gin.New()
	router.GET("/api/analytics/anomalies", handler.GetAnomalies)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/anomalies?time_range=24h", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should either return 200 or empty response (implementation is incomplete)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNoContent)
}

// TestGetTopIssues tests the top issues endpoint
func TestGetTopIssues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	handler := &AnalyticsHandler{
		topIssuesService: &services.TopIssuesService{},
		logger:           logger,
	}

	router := gin.New()
	router.GET("/api/analytics/top-issues", handler.GetTopIssues)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/top-issues?time_range=24h", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should either return 200 or empty response (implementation is incomplete)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNoContent)
}

// TestExportData tests the export endpoint
func TestExportData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	handler := &AnalyticsHandler{
		exportService: &services.ExportService{},
		logger:        logger,
	}

	router := gin.New()
	router.POST("/api/analytics/export", handler.ExportData)

	req := httptest.NewRequest(http.MethodPost, "/api/analytics/export?format=csv", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should either return 200 or empty response (implementation is incomplete)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNoContent)
}

// TestHandlerFieldAccess tests that handler fields are properly set
func TestHandlerFieldAccess(t *testing.T) {
	logger := logrus.New()
	agg := &services.AggregatorService{}
	trend := &services.TrendService{}
	anom := &services.AnomalyService{}
	top := &services.TopIssuesService{}
	exp := &services.ExportService{}

	handler := NewAnalyticsHandler(agg, trend, anom, top, exp, logger)

	// Verify all fields are accessible
	assert.NotNil(t, handler.aggregatorService)
	assert.NotNil(t, handler.trendService)
	assert.NotNil(t, handler.anomalyService)
	assert.NotNil(t, handler.topIssuesService)
	assert.NotNil(t, handler.exportService)
	assert.NotNil(t, handler.logger)
}

// TestHandlerWithNilLogger tests handler with nil logger
func TestHandlerWithNilLogger(t *testing.T) {
	agg := &services.AggregatorService{}
	trend := &services.TrendService{}
	anom := &services.AnomalyService{}
	top := &services.TopIssuesService{}
	exp := &services.ExportService{}

	handler := NewAnalyticsHandler(agg, trend, anom, top, exp, nil)

	assert.NotNil(t, handler)
	assert.Nil(t, handler.logger)
}

// TestGetTrends_WithParameters tests trends with parameters
func TestGetTrends_WithParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	handler := &AnalyticsHandler{
		trendService: &services.TrendService{},
		logger:       logger,
	}

	router := gin.New()
	router.GET("/api/analytics/trends", handler.GetTrends)

	tests := []string{"24h", "7d", "30d"}
	for _, tr := range tests {
		req := httptest.NewRequest(http.MethodGet, "/api/analytics/trends?time_range="+tr, http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNoContent)
	}
}

// TestGetAnomalies_WithParameters tests anomalies with severity parameter
func TestGetAnomalies_WithParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	handler := &AnalyticsHandler{
		anomalyService: &services.AnomalyService{},
		logger:         logger,
	}

	router := gin.New()
	router.GET("/api/analytics/anomalies", handler.GetAnomalies)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/anomalies?time_range=24h&severity=high", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNoContent)
}

// TestGetTopIssues_WithParameters tests top issues with level and limit
func TestGetTopIssues_WithParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	handler := &AnalyticsHandler{
		topIssuesService: &services.TopIssuesService{},
		logger:           logger,
	}

	router := gin.New()
	router.GET("/api/analytics/top-issues", handler.GetTopIssues)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/top-issues?time_range=24h&level=error&limit=10", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNoContent)
}

// TestExportData_WithFormat tests export with different formats
func TestExportData_WithFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	handler := &AnalyticsHandler{
		exportService: &services.ExportService{},
		logger:        logger,
	}

	router := gin.New()
	router.POST("/api/analytics/export", handler.ExportData)

	formats := []string{"csv", "json"}
	for _, fmt := range formats {
		req := httptest.NewRequest(http.MethodPost, "/api/analytics/export?format="+fmt, http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNoContent)
	}
}
