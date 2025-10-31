// Package internal_logs_handlers provides HTTP handlers for logs operations.
package internal_logs_handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
	"github.com/sirupsen/logrus"
)

// DashboardHandler handles dashboard-related HTTP endpoints.
type DashboardHandler struct {
	dashboardService logs_services.DashboardServiceInterface
	logger           *logrus.Logger
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(dashboardService logs_services.DashboardServiceInterface, logger *logrus.Logger) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
		logger:           logger,
	}
}

// Response wraps API responses.
type Response struct { //nolint:govet // struct alignment optimized for readability
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Success bool        `json:"success"`
}

// GetDashboardStats returns aggregated statistics for the dashboard.
// @Summary Get Dashboard Statistics
// @Description Retrieve aggregated statistics including service stats, health, and top errors
// @Produce json
// @Success 200 {object} Response
// @Router /api/logs/dashboard [get]
func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.dashboardService.GetDashboardStats(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get dashboard stats")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to retrieve dashboard statistics",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    stats,
	})
}

// GetServiceStats returns statistics for a specific service.
// @Summary Get Service Statistics
// @Description Retrieve statistics for a specific service within a time range
// @Param service query string true "Service name"
// @Param timeRange query string false "Time range (1h, 1d, 1w, default: 1h)"
// @Produce json
// @Success 200 {object} Response
// @Router /api/logs/dashboard/service [get]
func (h *DashboardHandler) GetServiceStats(c *gin.Context) {
	ctx := c.Request.Context()

	service := c.Query("service")
	if service == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "service parameter is required",
		})
		return
	}

	timeRangeStr := c.DefaultQuery("timeRange", "1h")
	timeRange, err := parseDuration(timeRangeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "invalid timeRange format",
		})
		return
	}

	stats, err := h.dashboardService.GetServiceStats(ctx, service, timeRange)
	if err != nil {
		h.logger.WithError(err).Warnf("Failed to get stats for service %s", service)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to retrieve service statistics",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    stats,
	})
}

// GetTopErrors returns the top error messages.
// @Summary Get Top Errors
// @Description Retrieve top error messages across all services
// @Param limit query int false "Maximum number of errors to return (default: 10)"
// @Param timeRange query string false "Time range (1h, 1d, 1w, default: 1h)"
// @Produce json
// @Success 200 {object} Response
// @Router /api/logs/dashboard/top-errors [get]
func (h *DashboardHandler) GetTopErrors(c *gin.Context) {
	ctx := c.Request.Context()

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	timeRangeStr := c.DefaultQuery("timeRange", "1h")
	timeRange, err := parseDuration(timeRangeStr)
	if err != nil {
		timeRange = time.Hour
	}

	errors, err := h.dashboardService.GetTopErrors(ctx, limit, timeRange)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get top errors")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to retrieve top errors",
		})
		return
	}

	if errors == nil {
		errors = []logs_models.TopErrorMessage{}
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    errors,
	})
}

// GetServiceHealth returns health status for all logs_services.
// @Summary Get Service Health
// @Description Retrieve health status for all services
// @Produce json
// @Success 200 {object} Response
// @Router /api/logs/dashboard/health [get]
func (h *DashboardHandler) GetServiceHealth(c *gin.Context) {
	ctx := c.Request.Context()

	health, err := h.dashboardService.GetServiceHealth(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get service health")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to retrieve service health",
		})
		return
	}

	if health == nil {
		health = make(map[string]*logs_models.ServiceHealth)
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    health,
	})
}

// parseDuration parses duration strings like "1h", "1d", "1w".
func parseDuration(s string) (time.Duration, error) {
	switch s {
	case "1h":
		return time.Hour, nil
	case "1d":
		return 24 * time.Hour, nil
	case "1w":
		return 7 * 24 * time.Hour, nil
	default:
		return 0, errors.New("invalid duration format")
	}
}
