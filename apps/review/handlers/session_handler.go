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
//
// Query Parameters:
//
//   - limit: Number of sessions to return (default: 10)
//
//   - offset: Number of sessions to skip (default: 0)
//
//     Response: {
//     "sessions": [...],
//     "pagination": {"total": 50, "limit": 10, "offset": 0}
//     }
func (h *SessionHandler) ListSessions(c *gin.Context) {
	// Extract user ID from context (set by auth middleware)
	userID, ok := c.Get("user_id")
	if !ok {
		h.logger.Warn("user_id not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: user_id not found"})
		return
	}
	userIDInt, ok := userID.(int64)
	if !ok {
		h.logger.Warn("user_id type assertion failed", "type", userID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id type"})
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

	h.logger.Info("listed sessions", "user_id", userIDInt, "count", len(sessions), "total", total)

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
//
// Path Parameters:
//
//   - id: Session ID (integer)
//
//     Response: {
//     "id": 1,
//     "user_id": 100,
//     "title": "...",
//     "code_source": "paste",
//     "created_at": "...",
//     "last_accessed": "..."
//     }
func (h *SessionHandler) GetSession(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid session id", "id", c.Param("id"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id format"})
		return
	}

	// Fetch session from database
	session, err := h.repo.GetByID(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("failed to get session", "error", err, "session_id", sessionID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if session == nil {
		h.logger.Info("session not found", "session_id", sessionID)
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// DeleteSession removes a review session
// DELETE /api/review/sessions/:id
//
// Path Parameters:
//   - id: Session ID (integer)
//
// Response: {"message": "session deleted successfully"}
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid session id", "id", c.Param("id"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id format"})
		return
	}

	// Delete session from database
	err = h.repo.DeleteByID(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("failed to delete session", "error", err, "session_id", sessionID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete session"})
		return
	}

	h.logger.Info("session deleted", "session_id", sessionID)
	c.JSON(http.StatusOK, gin.H{"message": "session deleted successfully"})
}
