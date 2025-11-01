// Package analytics_handlers provides HTTP handlers for the analytics UI service.
package analytics_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	templates "github.com/mikejsmith1985/devsmith-modular-platform/apps/analytics/templates"
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

// ContentHandler returns dashboard content sections (HTMX)
func (h *UIHandler) ContentHandler(c *gin.Context) {
	timeRange := c.DefaultQuery("time_range", "24h")
	h.logger.WithField("time_range", timeRange).Debug("Loading dashboard content")

	// For now, return the dashboard grid sections as HTML
	// In production, this would fetch real data from the analytics service
	html := `
	<div class="trends-section card">
		<h2>Log Trends</h2>
		<div class="chart-container">
			<canvas id="trends-chart"></canvas>
		</div>
		<div class="loading">Trends chart loading...</div>
	</div>
	<div class="anomalies-section card">
		<h2>Detected Anomalies</h2>
		<div id="anomalies-container">
			<div class="alert alert-info">No anomalies detected in the last ` + timeRange + `</div>
		</div>
	</div>
	<div class="issues-section card">
		<h2>Top Issues</h2>
		<div class="issues-filters">
			<select
				id="issues-level"
				name="level"
				hx-get="/api/analytics/issues"
				hx-target="#issues-container"
				hx-swap="innerHTML"
				hx-trigger="change"
				hx-include="[name='time_range']">
				<option value="all">All Levels</option>
				<option value="error">Errors Only</option>
				<option value="warn">Warnings Only</option>
			</select>
		</div>
		<div id="issues-container">
			<div class="alert alert-info">No issues in the last ` + timeRange + `</div>
		</div>
	</div>
	`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// IssuesHandler returns filtered top issues (HTMX)
func (h *UIHandler) IssuesHandler(c *gin.Context) {
	level := c.DefaultQuery("level", "all")
	timeRange := c.DefaultQuery("time_range", "24h")
	h.logger.WithFields(map[string]interface{}{
		"level":      level,
		"time_range": timeRange,
	}).Debug("Loading issues")

	// For now, return placeholder HTML
	html := `
	<div class="issues-list">
		<div class="alert alert-info">
			Showing ` + level + ` issues from the last ` + timeRange + `
		</div>
	</div>
	`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// RegisterUIRoutes registers the UI routes for the analytics handler.
func RegisterUIRoutes(router *gin.Engine, logger *logrus.Logger) {
	uiHandler := NewUIHandler(logger)

	// Dashboard UI route
	router.GET("/", uiHandler.DashboardHandler)
	router.GET("/dashboard", uiHandler.DashboardHandler)

	// Health check route
	router.GET("/health", uiHandler.HealthHandler)

	// HTMX API endpoints (Phase 12.5)
	router.GET("/api/analytics/content", uiHandler.ContentHandler) // Dashboard content with time range
	router.GET("/api/analytics/issues", uiHandler.IssuesHandler)   // Filtered issues
	// Note: /api/analytics/export is already registered in internal/analytics/handlers/analytics_handler.go
}
