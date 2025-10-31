// Package internal_logs_handlers provides HTTP handlers for logs operations.
package internal_logs_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
	"github.com/sirupsen/logrus"
)

// AlertHandler handles alert-related HTTP endpoints.
type AlertHandler struct {
	alertService logs_services.AlertServiceInterface
	logger       *logrus.Logger
}

// NewAlertHandler creates a new AlertHandler.
func NewAlertHandler(alertService logs_services.AlertServiceInterface, logger *logrus.Logger) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
		logger:       logger,
	}
}

// CreateAlertConfigRequest is the request body for creating an alert config.
type CreateAlertConfigRequest struct { //nolint:govet // struct alignment optimized for readability
	Service                string `json:"service" binding:"required"`
	AlertEmail             string `json:"alert_email"`
	AlertWebhookURL        string `json:"alert_webhook_url"`
	ErrorThresholdPerMin   int    `json:"error_threshold_per_min" binding:"required,min=1"`
	WarningThresholdPerMin int    `json:"warning_threshold_per_min" binding:"required,min=1"`
	Enabled                bool   `json:"enabled"`
}

// UpdateAlertConfigRequest is the request body for updating an alert config.
type UpdateAlertConfigRequest struct { //nolint:govet // struct alignment optimized for readability
	AlertEmail             string `json:"alert_email"`
	AlertWebhookURL        string `json:"alert_webhook_url"`
	ErrorThresholdPerMin   int    `json:"error_threshold_per_min"`
	WarningThresholdPerMin int    `json:"warning_threshold_per_min"`
	Enabled                bool   `json:"enabled"`
}

// CreateAlertConfig creates a new alert configuration.
// @Summary Create Alert Configuration
// @Description Create a new alert configuration for a service
// @Accept json
// @Produce json
// @Param body body CreateAlertConfigRequest true "Alert config details"
// @Success 201 {object} Response
// @Router /api/logs/alerts/config [post]
func (h *AlertHandler) CreateAlertConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateAlertConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	config := &logs_models.AlertConfig{
		Service:                req.Service,
		ErrorThresholdPerMin:   req.ErrorThresholdPerMin,
		WarningThresholdPerMin: req.WarningThresholdPerMin,
		AlertEmail:             req.AlertEmail,
		AlertWebhookURL:        req.AlertWebhookURL,
		Enabled:                req.Enabled,
	}

	if err := h.alertService.CreateAlertConfig(ctx, config); err != nil {
		h.logger.WithError(err).Errorf("Failed to create alert config for service %s", req.Service)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to create alert configuration",
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    config,
		Message: "Alert configuration created successfully",
	})
}

// GetAlertConfig retrieves alert configuration for a service.
// @Summary Get Alert Configuration
// @Description Retrieve alert configuration for a specific service
// @Param service query string true "Service name"
// @Produce json
// @Success 200 {object} Response
// @Router /api/logs/alerts/config [get]
func (h *AlertHandler) GetAlertConfig(c *gin.Context) {
	ctx := c.Request.Context()

	service := c.Query("service")
	if service == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "service parameter is required",
		})
		return
	}

	config, err := h.alertService.GetAlertConfig(ctx, service)
	if err != nil {
		h.logger.WithError(err).Warnf("Failed to get alert config for service %s", service)
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Error:   "Alert configuration not found",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    config,
	})
}

// UpdateAlertConfig updates an existing alert configuration.
// @Summary Update Alert Configuration
// @Description Update alert configuration for a service
// @Accept json
// @Produce json
// @Param service query string true "Service name"
// @Param body body UpdateAlertConfigRequest true "Updated alert config details"
// @Success 200 {object} Response
// @Router /api/logs/alerts/config [put]
func (h *AlertHandler) UpdateAlertConfig(c *gin.Context) {
	ctx := c.Request.Context()

	service := c.Query("service")
	if service == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "service parameter is required",
		})
		return
	}

	// Get existing config
	existing, err := h.alertService.GetAlertConfig(ctx, service)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Error:   "Alert configuration not found",
		})
		return
	}

	var req UpdateAlertConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Update fields
	if req.ErrorThresholdPerMin > 0 {
		existing.ErrorThresholdPerMin = req.ErrorThresholdPerMin
	}
	if req.WarningThresholdPerMin > 0 {
		existing.WarningThresholdPerMin = req.WarningThresholdPerMin
	}
	if req.AlertEmail != "" {
		existing.AlertEmail = req.AlertEmail
	}
	if req.AlertWebhookURL != "" {
		existing.AlertWebhookURL = req.AlertWebhookURL
	}
	existing.Enabled = req.Enabled

	if err := h.alertService.UpdateAlertConfig(ctx, existing); err != nil {
		h.logger.WithError(err).Errorf("Failed to update alert config for service %s", service)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to update alert configuration",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    existing,
		Message: "Alert configuration updated successfully",
	})
}

// CheckThresholds checks if current log counts exceed alert thresholds.
// @Summary Check Alert Thresholds
// @Description Check if current log counts exceed configured thresholds
// @Produce json
// @Success 200 {object} Response
// @Router /api/logs/alerts/check [post]
func (h *AlertHandler) CheckThresholds(c *gin.Context) {
	ctx := c.Request.Context()

	violations, err := h.alertService.CheckThresholds(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to check alert thresholds")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to check thresholds",
		})
		return
	}

	if violations == nil {
		violations = []logs_models.AlertThresholdViolation{}
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    violations,
	})
}
