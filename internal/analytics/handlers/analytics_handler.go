package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/services"
	"github.com/sirupsen/logrus"
)

type AnalyticsHandler struct {
	aggregatorService *services.AggregatorService
	trendService      *services.TrendService
	anomalyService    *services.AnomalyService
	topIssuesService  *services.TopIssuesService
	exportService     *services.ExportService
	logger            *logrus.Logger
}

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

func (h *AnalyticsHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/analytics")
	{
		api.POST("/aggregate", h.RunAggregation)
		api.GET("/trends", h.GetTrends)
		api.GET("/anomalies", h.GetAnomalies)
		api.GET("/top-issues", h.GetTopIssues)
		api.POST("/export", h.ExportData)
	}
}

func (h *AnalyticsHandler) RunAggregation(c *gin.Context) {
	if err := h.aggregatorService.RunHourlyAggregation(c.Request.Context()); err != nil {
		h.logger.WithError(err).Error("Failed to run aggregation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run aggregation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Aggregation completed successfully"})
}

func (h *AnalyticsHandler) GetTrends(c *gin.Context) {
	// Implementation for fetching trends
}

func (h *AnalyticsHandler) GetAnomalies(c *gin.Context) {
	// Implementation for fetching anomalies
}

func (h *AnalyticsHandler) GetTopIssues(c *gin.Context) {
	// Implementation for fetching top issues
}

func (h *AnalyticsHandler) ExportData(c *gin.Context) {
	// Implementation for exporting data
}
