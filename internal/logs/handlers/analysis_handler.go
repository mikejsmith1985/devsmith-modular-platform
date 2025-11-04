// Package internal_logs_handlers provides HTTP handlers for logs operations.
package internal_logs_handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
	"github.com/sirupsen/logrus"
)

// AnalysisService defines the interface for log analysis operations
type AnalysisService interface {
	AnalyzeLogEntry(ctx context.Context, entry *logs_models.LogEntry) (*logs_services.AnalysisResult, error)
	ClassifyLogEntry(ctx context.Context, entry *logs_models.LogEntry) (string, error)
}

// AnalysisHandler handles AI-powered log analysis requests
type AnalysisHandler struct {
	service AnalysisService
	logger  *logrus.Logger
}

// AnalyzeLogRequest represents a request to analyze a log entry
type AnalyzeLogRequest struct {
	LogEntry logs_models.LogEntry `json:"log_entry"`
}

// ClassifyLogRequest represents a request to classify a log entry
type ClassifyLogRequest struct {
	LogEntry logs_models.LogEntry `json:"log_entry"`
}

// NewAnalysisHandler creates a new analysis handler
func NewAnalysisHandler(service AnalysisService, logger *logrus.Logger) *AnalysisHandler {
	return &AnalysisHandler{
		service: service,
		logger:  logger,
	}
}

// AnalyzeLog handles POST /api/logs/analyze
// Performs AI-powered root cause analysis on a log entry
func (h *AnalysisHandler) AnalyzeLog(c *gin.Context) {
	var req AnalyzeLogRequest

	// Parse request
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid analyze request")
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Call service
	result, err := h.service.AnalyzeLogEntry(c.Request.Context(), &req.LogEntry)
	if err != nil {
		h.logger.WithError(err).Error("Failed to analyze log entry")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Analysis failed",
		})
		return
	}

	// Return result
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"root_cause":    result.RootCause,
			"suggested_fix": result.SuggestedFix,
			"severity":      result.Severity,
			"related_logs":  result.RelatedLogs,
			"fix_steps":     result.FixSteps,
		},
	})
}

// ClassifyLog handles POST /api/logs/classify
// Classifies a log entry into a known issue type using pattern matching
func (h *AnalysisHandler) ClassifyLog(c *gin.Context) {
	var req ClassifyLogRequest

	// Parse request
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid classify request")
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Call service
	issueType, err := h.service.ClassifyLogEntry(c.Request.Context(), &req.LogEntry)
	if err != nil {
		h.logger.WithError(err).Error("Failed to classify log entry")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Classification failed",
		})
		return
	}

	// Return result
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"issue_type": issueType,
		},
	})
}
