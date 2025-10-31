package review_handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	templates "github.com/mikejsmith1985/devsmith-modular-platform/apps/review/templates"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logging"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// UIHandler provides HTTP handlers for the Review UI with logging.
type UIHandler struct {
	logger    logger.Interface
	logClient *logging.Client
}

// NewUIHandler creates a new UIHandler with the given logger and optional logging client.
func NewUIHandler(logger logger.Interface, client *logging.Client) *UIHandler {
	return &UIHandler{logger: logger, logClient: client}
}

// HomeHandler serves the main Review UI (mode selector + repo input)
func (h *UIHandler) HomeHandler(c *gin.Context) {
	correlationID := c.Request.Context().Value("correlation_id")
	h.logger.Info("HomeHandler called", "correlation_id", correlationID, "path", c.Request.URL.Path)
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.Home().Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.Error("Failed to render Home template", "error", err)
		c.String(http.StatusInternalServerError, "Failed to render page")
	}
}

// AnalysisResultHandler displays analysis results
func (h *UIHandler) AnalysisResultHandler(c *gin.Context) {
	correlationID := c.Request.Context().Value("correlation_id")
	mode := c.Query("mode")
	repo := c.Query("repo")
	branch := c.Query("branch")
	analysisMarkdown := c.Query("analysis")

	h.logger.Info("AnalysisResultHandler called", "correlation_id", correlationID, "mode", mode, "repo", repo, "branch", branch)

	data := map[string]interface{}{
		"AnalysisID":   generateAnalysisID(),
		"Mode":         mode,
		"Repository":   repo,
		"Branch":       branch,
		"AnalysisHTML": analysisMarkdown,
		"CreatedAt":    time.Now().Format("2006-01-02 15:04:05"),
	}

	c.HTML(http.StatusOK, "analysis.html", data)
}

// CreateSessionHandler handles session creation from the UI form
func (h *UIHandler) CreateSessionHandler(c *gin.Context) {
	correlationID := c.Request.Context().Value("correlation_id")
	h.logger.Info("CreateSessionHandler called", "correlation_id", correlationID)

	// Parse form data
	var req struct {
		PastedCode string `form:"pasted_code" json:"pasted_code"`
		GitHubURL  string `form:"github_url" json:"github_url"`
		Title      string `form:"title" json:"title"`
	}

	if err := c.ShouldBind(&req); err != nil {
		h.logger.Error("Failed to bind session creation request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Validate that at least one input method is provided
	if req.PastedCode == "" && req.GitHubURL == "" {
		h.logger.Warn("Session creation request missing code or URL", "correlation_id", correlationID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either pasted_code or github_url is required"})
		return
	}

	// Generate session ID
	sessionID := uuid.New().String()

	h.logger.Info("Session created successfully", "correlation_id", correlationID, "session_id", sessionID)

	// Send a lightweight log event to the Logs service if a client is configured.
	if h.logClient != nil {
		// Best-effort: do not block the request path
		go func(ctx context.Context, sid string) {
			if err := h.logClient.Post(ctx, map[string]interface{}{
				"service": "review",
				"event":   "session_created",
				"session": sid,
				"time":    time.Now().UTC().Format(time.RFC3339),
			}); err != nil {
				log.Printf("warning: failed to post session_created event: %v", err)
			}
		}(c.Request.Context(), sessionID)
	}
	// Return session info (for now, just the ID)
	c.JSON(http.StatusCreated, gin.H{
		"session_id": sessionID,
		"message":    "Session created successfully",
	})
}

// SessionProgressSSE streams progress updates for a given session via SSE.
// This is a lightweight simulator for UI integration and demos. In production
// this should be driven by the actual analysis pipeline (publish progress to
// a channel/store that this handler reads from).
func (h *UIHandler) SessionProgressSSE(c *gin.Context) {
	sessionID := c.Param("id")
	correlationID := c.Request.Context().Value("correlation_id")
	h.logger.Info("SessionProgressSSE connected", "session_id", sessionID, "correlation_id", correlationID)

	// Set headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	// Flush helper
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		h.logger.Error("SSE unsupported by writer")
		c.Status(http.StatusInternalServerError)
		return
	}

	// Send initial progress event and begin streaming
	h.streamSessionProgress(c, flusher, sessionID)
}

// streamSessionProgress handles the main SSE streaming loop for session progress.
// Extracted to reduce cognitive complexity of SessionProgressSSE.
func (h *UIHandler) streamSessionProgress(c *gin.Context, flusher http.Flusher, sessionID string) {
	percent := 0
	ticker := time.NewTicker(400 * time.Millisecond)
	defer ticker.Stop()

	// Send initial event
	if !h.writeSSEEvent(c, flusher, 0, "Queued") {
		return
	}

	// Loop and send updates until complete or client disconnect
	for percent < 100 {
		select {
		case <-c.Request.Context().Done():
			h.logger.Info("SSE client disconnected", "session_id", sessionID)
			return
		case <-ticker.C:
			percent = updateProgressPercent(percent)
			if percent > 100 {
				percent = 100
			}

			if !h.writeSSEEvent(c, flusher, percent, "Processing") {
				return
			}

			if percent >= 100 {
				h.writeFinalSSEEvent(c, flusher)
				return
			}
		}
	}
}

// writeSSEEvent writes a progress event to the SSE stream.
func (h *UIHandler) writeSSEEvent(c *gin.Context, flusher http.Flusher, percent int, message string) bool {
	msg := fmt.Sprintf("event: progress\n data: {\"percent\": %d, \"message\": %q}\n\n", percent, message)
	if _, err := c.Writer.WriteString(msg); err != nil {
		h.logger.Error("failed to write SSE event", "error", err)
		return false
	}
	flusher.Flush()
	return true
}

// writeFinalSSEEvent writes the completion event to the SSE stream.
func (h *UIHandler) writeFinalSSEEvent(c *gin.Context, flusher http.Flusher) {
	if _, err := c.Writer.WriteString("event: progress\n"); err != nil {
		h.logger.Error("failed to write SSE final header", "error", err)
		return
	}
	if _, err := c.Writer.WriteString("data: {\"percent\": 100, \"message\": \"Complete\"}\n\n"); err != nil {
		h.logger.Error("failed to write SSE final data", "error", err)
		return
	}
	flusher.Flush()
}

// updateProgressPercent calculates the next progress percentage based on current value.
func updateProgressPercent(current int) int {
	switch {
	case current < 30:
		return current + 5
	case current < 70:
		return current + 8
	default:
		return current + 10
	}
}

// generateAnalysisID creates a unique ID for analysis sessions (backwards compat).
func generateAnalysisID() string {
	return GenerateAnalysisID()
}

// GenerateAnalysisID creates a unique ID for analysis sessions.
func GenerateAnalysisID() string {
	return uuid.New().String()
}
