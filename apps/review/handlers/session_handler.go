package review_handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	review_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// SessionHandler handles HTTP requests for review session management
type SessionHandler struct {
	repo   *review_db.ReviewRepository
	logger logger.Interface
}

// NewSessionHandler creates a new SessionHandler
func NewSessionHandler(repo *review_db.ReviewRepository, logger logger.Interface) *SessionHandler {
	return &SessionHandler{
		repo:   repo,
		logger: logger,
	}
}

// ListSessions returns paginated list of user's review sessions
// GET /api/review/sessions?limit=10&offset=0
func (h *SessionHandler) ListSessions(c *gin.Context) {
	// Extract user ID from context (set by auth middleware)
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}
	userIDInt, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// Parse pagination parameters
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Fetch sessions from database
	sessions, total, err := h.repo.ListByUserID(c.Request.Context(), userIDInt, limit, offset)
	if err != nil {
		h.logger.Error("failed to list sessions", "error", err, "user_id", userIDInt)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch sessions"})
		return
	}

	// Return paginated response
	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetSession returns a specific review session by ID
// GET /api/review/sessions/:id
func (h *SessionHandler) GetSession(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	// Fetch session from database
	session, err := h.repo.GetByID(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("failed to get session", "error", err, "session_id", sessionID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch session"})
		return
	}

	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// DeleteSession removes a review session
// DELETE /api/review/sessions/:id
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	// Delete session from database
	err = h.repo.DeleteByID(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("failed to delete session", "error", err, "session_id", sessionID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session deleted successfully"})
}
