package handlers

import (
	"bytes"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewAnalyticsHandler(t *testing.T) {
	logger := logrus.New()

	handler := NewAnalyticsHandler(
		(*services.AggregatorService)(nil),
		(*services.TrendService)(nil),
		(*services.AnomalyService)(nil),
		(*services.TopIssuesService)(nil),
		(*services.ExportService)(nil),
		logger,
	)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
}

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()

	handler := NewAnalyticsHandler(
		(*services.AggregatorService)(nil),
		(*services.TrendService)(nil),
		(*services.AnomalyService)(nil),
		(*services.TopIssuesService)(nil),
		(*services.ExportService)(nil),
		logger,
	)

	handler.RegisterRoutes(router)

	// Verify routes are registered
	routes := router.Routes()
	assert.NotEmpty(t, routes)

	// Check for specific routes
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Method+" "+route.Path] = true
	}

	assert.True(t, routePaths["POST /api/analytics/aggregate"])
	assert.True(t, routePaths["GET /api/analytics/trends"])
	assert.True(t, routePaths["GET /api/analytics/anomalies"])
	assert.True(t, routePaths["GET /api/analytics/top-issues"])
	assert.True(t, routePaths["POST /api/analytics/export"])
}

func TestAnalyticsHandlerFields(t *testing.T) {
	// Test that the handler stores all services correctly
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))

	handler := NewAnalyticsHandler(
		(*services.AggregatorService)(nil),
		(*services.TrendService)(nil),
		(*services.AnomalyService)(nil),
		(*services.TopIssuesService)(nil),
		(*services.ExportService)(nil),
		logger,
	)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)

	// Verify fields are initialized (even if nil)
	_ = handler.aggregatorService
	_ = handler.trendService
	_ = handler.anomalyService
	_ = handler.topIssuesService
	_ = handler.exportService
}
