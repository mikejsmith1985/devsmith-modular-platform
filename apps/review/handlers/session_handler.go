// Package review_handlers provides HTTP handlers for the review service.
package review_handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// SessionHandlers provides HTTP handlers for session operations.
type SessionHandlers struct {
	sessionService *review_services.SessionService
	logger         logger.Interface
}

// NewSessionHandlers creates a new SessionHandlers instance.
func NewSessionHandlers(sessionService *review_services.SessionService, logger logger.Interface) *SessionHandlers {
	return &SessionHandlers{
		sessionService: sessionService,
		logger:         logger,
	}
}

// RegisterRoutes registers all session-related routes.
func (h *SessionHandlers) RegisterRoutes(router *gin.Engine) {
	sessions := router.Group("/api/review/sessions")
	{
		sessions.POST("", h.CreateSession)
		sessions.GET("", h.ListSessions)
		sessions.GET("/:id", h.GetSession)
		sessions.PUT("/:id", h.UpdateSession)
		sessions.DELETE("/:id", h.DeleteSession)
		sessions.POST("/:id/modes/:mode", h.UpdateSessionMode)
		sessions.POST("/:id/complete", h.CompleteSession)
		sessions.POST("/:id/archive", h.ArchiveSession)
		sessions.GET("/:id/history", h.GetSessionHistory)
		sessions.POST("/:id/notes", h.AddSessionNote)
	}
}

// CreateSession creates a new code review session.
// POST /api/review/sessions
func (h *SessionHandlers) CreateSession(c *gin.Context) {
	var req review_services.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	if req.CodeSource == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CodeSource is required"})
		return
	}

	// Get user ID from context (would come from auth middleware)
	userID := int64(1) // TODO: Get from authenticated user

	req.UserID = userID
	session, err := h.sessionService.CreateSession(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// GetSession retrieves a session by ID.
// GET /api/review/sessions/:id
func (h *SessionHandlers) GetSession(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	session, err := h.sessionService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to get session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve session"})
		return
	}

	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// ListSessions retrieves sessions with filtering and pagination.
// GET /api/review/sessions?status=active&limit=10&offset=0
func (h *SessionHandlers) ListSessions(c *gin.Context) {
	// Parse query parameters
	userID := int64(1) // TODO: Get from authenticated user
	status := c.Query("status")
	language := c.Query("language")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "DESC")

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	filter := &review_models.SessionFilter{
		UserID:    userID,
		Status:    status,
		Language:  language,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Limit:     limit,
		Offset:    offset,
	}

	summaries, err := h.sessionService.ListSessions(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list sessions", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": summaries,
		"count":    len(summaries),
		"limit":    limit,
		"offset":   offset,
	})
}

// UpdateSession updates an existing session.
// PUT /api/review/sessions/:id
func (h *SessionHandlers) UpdateSession(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		CurrentMode string `json:"current_mode"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Retrieve session
	session, err := h.sessionService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to get session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve session"})
		return
	}

	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Update fields
	if req.Title != "" {
		session.Title = req.Title
	}
	if req.Description != "" {
		session.Description = req.Description
	}
	if req.Status != "" {
		session.Status = req.Status
	}
	if req.CurrentMode != "" {
		session.CurrentMode = req.CurrentMode
	}
	session.UpdatedAt = time.Now()

	// TODO: Call repository to persist changes

	c.JSON(http.StatusOK, session)
}

// DeleteSession removes a session.
// DELETE /api/review/sessions/:id
func (h *SessionHandlers) DeleteSession(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	err = h.sessionService.DeleteSession(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to delete session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// UpdateSessionMode updates the state for a specific mode within a session.
// POST /api/review/sessions/:id/modes/:mode
func (h *SessionHandlers) UpdateSessionMode(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	mode := c.Param("mode")
	if mode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mode is required"})
		return
	}

	var req review_services.ModeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.Mode = mode
	err = h.sessionService.UpdateSessionMode(c.Request.Context(), sessionID, &req)
	if err != nil {
		h.logger.Error("Failed to update session mode", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session mode"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mode updated successfully"})
}

// CompleteSession marks a session as completed and calculates statistics.
// POST /api/review/sessions/:id/complete
func (h *SessionHandlers) CompleteSession(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	stats, err := h.sessionService.CompleteSession(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to complete session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete session"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ArchiveSession moves a session to archived status.
// POST /api/review/sessions/:id/archive
func (h *SessionHandlers) ArchiveSession(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	err = h.sessionService.ArchiveSession(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to archive session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session archived successfully"})
}

// GetSessionHistory retrieves the audit trail for a session.
// GET /api/review/sessions/:id/history
func (h *SessionHandlers) GetSessionHistory(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	history, err := h.sessionService.GetSessionHistory(c.Request.Context(), sessionID, limit)
	if err != nil {
		h.logger.Error("Failed to get session history", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"count":   len(history),
	})
}

// AddSessionNote adds user notes to a session or specific mode.
// POST /api/review/sessions/:id/notes
func (h *SessionHandlers) AddSessionNote(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var req struct {
		Mode string `json:"mode" binding:"required"`
		Note string `json:"note" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mode and note are required"})
		return
	}

	err = h.sessionService.AddSessionNote(c.Request.Context(), sessionID, req.Mode, req.Note)
	if err != nil {
		h.logger.Error("Failed to add session note", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note added successfully"})
}
