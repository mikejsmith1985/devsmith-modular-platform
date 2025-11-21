// Package logs_handlers provides HTTP handlers for the logs UI service.
package logs_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
	"github.com/sirupsen/logrus"
)

// UIHandler handles UI requests for the logs service
type UIHandler struct {
	logger        *logrus.Logger
	policyService *services.HealthPolicyService
}

// NewUIHandler creates a new UI handler
func NewUIHandler(logger *logrus.Logger, policyService *services.HealthPolicyService) *UIHandler {
	return &UIHandler{
		logger:        logger,
		policyService: policyService,
	}
}

// HealthHandler serves health check in JSON format.
func (h *UIHandler) HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "logs",
		"status":  "healthy",
	})
}

// RegisterUIRoutes registers the UI routes for the logs handler.
// Note: Dashboard UI and health check dashboard now handled by React frontend.
// REST API health check available at /api/logs/healthcheck (see cmd/logs/handlers/healthcheck_handler.go)
func RegisterUIRoutes(router *gin.Engine, uiHandler *UIHandler) {
	// Health check route (simple JSON)
	router.GET("/health", uiHandler.HealthHandler)
}
