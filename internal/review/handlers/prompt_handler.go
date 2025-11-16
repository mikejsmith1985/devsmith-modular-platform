package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// PromptTemplateService defines the interface for prompt template business logic
type PromptTemplateService interface {
	GetEffectivePrompt(ctx context.Context, userID int, mode, userLevel, outputMode string) (*review_models.PromptTemplate, error)
	SaveCustomPrompt(ctx context.Context, userID int, mode, userLevel, outputMode, promptText string) (*review_models.PromptTemplate, error)
	FactoryReset(ctx context.Context, userID int, mode, userLevel, outputMode string) error
	GetExecutionHistory(ctx context.Context, userID int, limit int) ([]*review_models.PromptExecution, error)
	RateExecution(ctx context.Context, userID int, executionID int64, rating int) error
}

// PromptHandler handles HTTP requests for prompt management
type PromptHandler struct {
	service PromptTemplateService
}

// NewPromptHandler creates a new PromptHandler
func NewPromptHandler(service PromptTemplateService) *PromptHandler {
	return &PromptHandler{
		service: service,
	}
}

// GetPrompt returns the effective prompt for the given mode/level/output
// GET /api/review/prompts?mode={mode}&user_level={level}&output_mode={output}
func (h *PromptHandler) GetPrompt(c *gin.Context) {
	// Extract user_id from context (set by auth middleware)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userID := userIDVal.(int)

	// Extract query parameters
	mode := c.Query("mode")
	userLevel := c.Query("user_level")
	outputMode := c.Query("output_mode")

	// Validate required parameters
	if mode == "" || userLevel == "" || outputMode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters: mode, user_level, output_mode"})
		return
	}

	// Get effective prompt from service
	prompt, err := h.service.GetEffectivePrompt(c.Request.Context(), userID, mode, userLevel, outputMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prompt"})
		return
	}

	// Return prompt with metadata (flat structure)
	c.JSON(http.StatusOK, gin.H{
		"id":          prompt.ID,
		"mode":        prompt.Mode,
		"user_level":  prompt.UserLevel,
		"output_mode": prompt.OutputMode,
		"prompt_text": prompt.PromptText,
		"variables":   prompt.Variables,
		"is_custom":   prompt.IsCustom(),
		"can_reset":   prompt.CanBeDeleted(),
		"is_default":  prompt.IsDefault,
		"version":     prompt.Version,
		"created_at":  prompt.CreatedAt,
		"updated_at":  prompt.UpdatedAt,
	})
}

// SavePrompt creates or updates a custom prompt
// PUT /api/review/prompts
func (h *PromptHandler) SavePrompt(c *gin.Context) {
	// Extract user_id from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userID := userIDVal.(int)

	// Parse request body
	var req struct {
		Mode       string `json:"mode" binding:"required"`
		UserLevel  string `json:"user_level" binding:"required"`
		OutputMode string `json:"output_mode" binding:"required"`
		PromptText string `json:"prompt_text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Save custom prompt
	prompt, err := h.service.SaveCustomPrompt(c.Request.Context(), userID, req.Mode, req.UserLevel, req.OutputMode, req.PromptText)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return saved prompt
	c.JSON(http.StatusOK, prompt)
}

// ResetPrompt deletes a user's custom prompt (factory reset)
// DELETE /api/review/prompts?mode={mode}&user_level={level}&output_mode={output}
func (h *PromptHandler) ResetPrompt(c *gin.Context) {
	// Extract user_id from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userID := userIDVal.(int)

	// Extract query parameters
	mode := c.Query("mode")
	userLevel := c.Query("user_level")
	outputMode := c.Query("output_mode")

	// Validate required parameters
	if mode == "" || userLevel == "" || outputMode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters: mode, user_level, output_mode"})
		return
	}

	// Factory reset
	err := h.service.FactoryReset(c.Request.Context(), userID, mode, userLevel, outputMode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No custom prompt found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Prompt reset to factory default",
	})
}

// GetHistory returns the user's prompt execution history
// GET /api/review/prompts/history?limit=50
func (h *PromptHandler) GetHistory(c *gin.Context) {
	// Extract user_id from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userID := userIDVal.(int)

	// Extract limit parameter (default 50)
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get execution history
	executions, err := h.service.GetExecutionHistory(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve history"})
		return
	}

	// Return executions array directly
	c.JSON(http.StatusOK, executions)
}

// RateExecution updates the user rating for a prompt execution
// POST /api/review/prompts/:execution_id/rate
func (h *PromptHandler) RateExecution(c *gin.Context) {
	// Extract user_id from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userID := userIDVal.(int)

	// Extract execution_id from URL
	executionIDStr := c.Param("execution_id")
	executionID, err := strconv.ParseInt(executionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution_id"})
		return
	}

	// Parse request body
	var req struct {
		Rating int `json:"rating" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate rating range
	if req.Rating < 1 || req.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rating must be between 1 and 5"})
		return
	}

	// Update rating
	err = h.service.RateExecution(c.Request.Context(), userID, executionID, req.Rating)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rating"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rating updated successfully",
	})
}
