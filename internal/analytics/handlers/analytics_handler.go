// Package handlers provides HTTP handlers for the analytics service.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/services"
	"github.com/sirupsen/logrus"
)

// AnalyticsHandler handles HTTP requests for the analytics service.
// It provides methods for running aggregations, fetching trends, anomalies, top issues, and exporting data.
type AnalyticsHandler struct {
	aggregatorService *services.AggregatorService
	trendService      *services.TrendService
	anomalyService    *services.AnomalyService
	topIssuesService  *services.TopIssuesService
	exportService     *services.ExportService
	logger            *logrus.Logger
}

// NewAnalyticsHandler creates a new instance of AnalyticsHandler.
func NewAnalyticsHandler(aggregatorService *services.AggregatorService, trendService *services.TrendService, anomalyService *services.AnomalyService, topIssuesService *services.TopIssuesService, exportService *services.ExportService, logger *logrus.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		aggregatorService: aggregatorService,
		trendService:      trendService,
		anomalyService:    anomalyService,
		topIssuesService:  topIssuesService,
		exportService:     exportService,
		logger:            logger,
	}
}

// RegisterRoutes registers the HTTP routes for the analytics handler.
func (h *AnalyticsHandler) RegisterRoutes(router *gin.Engine) {
	router.Group("/api/analytics").POST("/aggregate", h.RunAggregation)
	router.Group("/api/analytics").GET("/trends", h.GetTrends)
	router.Group("/api/analytics").GET("/anomalies", h.GetAnomalies)
	router.Group("/api/analytics").GET("/top-issues", h.GetTopIssues)
	router.Group("/api/analytics").POST("/export", h.ExportData)
}

// RunAggregation triggers the hourly aggregation process.
// It responds with a success message or an error if the aggregation fails.
func (h *AnalyticsHandler) RunAggregation(c *gin.Context) {
	if err := h.aggregatorService.RunHourlyAggregation(c.Request.Context()); err != nil {
		h.logger.WithError(err).Error("Failed to run aggregation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run aggregation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Aggregation completed successfully"})
}

// GetTrends retrieves trend data for the analytics service.
// It responds with the trend data or an error if the operation fails.
func (h *AnalyticsHandler) GetTrends(c *gin.Context) {
	// Implementation for fetching trends
}

// GetAnomalies retrieves anomaly data for the analytics service.
// It responds with the anomaly data or an error if the operation fails.
func (h *AnalyticsHandler) GetAnomalies(c *gin.Context) {
	// Implementation for fetching anomalies
}

// GetTopIssues retrieves the top issues for the analytics service.
// It responds with the top issues data or an error if the operation fails.
func (h *AnalyticsHandler) GetTopIssues(c *gin.Context) {
	// Implementation for fetching top issues
}

// ExportData exports analytics data to a specified format.
// It responds with a success message or an error if the export fails.
func (h *AnalyticsHandler) ExportData(c *gin.Context) {
	// Implementation for exporting data
}
