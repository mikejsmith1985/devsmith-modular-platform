package internal_logs_handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	logs_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

// AIInsightsHandler handles AI insights API requests
type AIInsightsHandler struct {
	service *logs_services.AIInsightsService
	logger  *logrus.Logger
	logRepo *logs_db.LogEntryRepository
}

// NewAIInsightsHandler creates a new AI insights handler
func NewAIInsightsHandler(service *logs_services.AIInsightsService, logger *logrus.Logger, logRepo *logs_db.LogEntryRepository) *AIInsightsHandler {
	return &AIInsightsHandler{
		service: service,
		logger:  logger,
		logRepo: logRepo,
	}
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
		// Log the AI Insights failure to the logs system
		h.logger.WithFields(logrus.Fields{
			"log_id": logID,
			"model":  req.Model,
			"error":  err.Error(),
		}).Error("AI Insights generation failed")

		// Also create a log entry in the database so it appears in the UI
		if h.logRepo != nil {
			// Prepare metadata as JSON
			metadata := map[string]interface{}{
				"log_id":     logID,
				"model":      req.Model,
				"error":      err.Error(),
				"error_type": "ai_generation_failure",
				"failed_at":  time.Now().Format(time.RFC3339),
			}
			metadataJSON, _ := json.Marshal(metadata)

			logEntry := &logs_models.LogEntry{
				Level:    "ERROR",
				Message:  fmt.Sprintf("AI Insights generation failed for log %d with model %s: %v", logID, req.Model, err),
				Service:  "ai-insights",
				Metadata: metadataJSON,
				UserID:   0, // System-generated log
			}

			// Best-effort logging (don't fail the request if logging fails)
			if _, logErr := h.logRepo.Create(c.Request.Context(), logEntry); logErr != nil {
				h.logger.WithError(logErr).Warn("Failed to log AI Insights error to database")
			}
		}

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
