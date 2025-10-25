// Package handlers provides HTTP handlers for the analytics UI service.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/apps/analytics/templates"
	"github.com/sirupsen/logrus"
)

// UIHandler handles UI-related HTTP requests for the analytics service.
type UIHandler struct {
	logger *logrus.Logger
}

// NewUIHandler creates a new instance of UIHandler.
func NewUIHandler(logger *logrus.Logger) *UIHandler {
	return &UIHandler{
		logger: logger,
	}
}

// DashboardHandler serves the main Analytics dashboard.
func (h *UIHandler) DashboardHandler(c *gin.Context) {
	component := templates.Dashboard()
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.WithError(err).Error("Failed to render dashboard template")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render dashboard"})
	}
}

// HealthHandler serves health check in plain text/JSON format.
func (h *UIHandler) HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "analytics",
		"status":  "healthy",
	})
}

// RegisterUIRoutes registers the UI routes for the analytics handler.
func RegisterUIRoutes(router *gin.Engine, logger *logrus.Logger) {
	uiHandler := NewUIHandler(logger)

	// Dashboard UI route
	router.GET("/", uiHandler.DashboardHandler)
	router.GET("/dashboard", uiHandler.DashboardHandler)

	// Health check route
	router.GET("/health", uiHandler.HealthHandler)
}
