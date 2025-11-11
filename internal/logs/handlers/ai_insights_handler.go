package internal_logs_handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

// AIInsightsHandler handles AI insights API requests
type AIInsightsHandler struct {
	service *logs_services.AIInsightsService
}

// NewAIInsightsHandler creates a new AI insights handler
func NewAIInsightsHandler(service *logs_services.AIInsightsService) *AIInsightsHandler {
	return &AIInsightsHandler{service: service}
}

// GenerateInsights handles POST /api/logs/:id/insights
// Generates or regenerates AI insights for a log entry
func (h *AIInsightsHandler) GenerateInsights(c *gin.Context) {
	// Parse log ID from URL
	logIDStr := c.Param("id")
	logID, err := strconv.ParseInt(logIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}

	// Parse request body
	var req struct {
		Model string `json:"model" binding:"required"`
	}
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body - model parameter is required",
			"details": bindErr.Error(),
		})
		return
	}

	// Additional validation: check if model is empty string
	if req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Model parameter cannot be empty",
		})
		return
	}

	// Generate insights
	insight, err := h.service.GenerateInsights(c.Request.Context(), logID, req.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, insight)
}

// GetInsights handles GET /api/logs/:id/insights
// Retrieves cached AI insights for a log entry
func (h *AIInsightsHandler) GetInsights(c *gin.Context) {
	// Parse log ID from URL
	logIDStr := c.Param("id")
	logID, err := strconv.ParseInt(logIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}

	// Get insights from database
	insight, err := h.service.GetInsights(c.Request.Context(), logID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if insight == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No insights found for this log"})
		return
	}

	c.JSON(http.StatusOK, insight)
}
