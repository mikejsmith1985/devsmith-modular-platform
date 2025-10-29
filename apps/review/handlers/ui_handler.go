package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mikejsmith1985/devsmith-modular-platform/apps/review/templates"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// UIHandler provides HTTP handlers for the Review UI with logging.
type UIHandler struct {
	logger logger.Interface
}

// NewUIHandler creates a new UIHandler with the given logger.
func NewUIHandler(logger logger.Interface) *UIHandler {
	return &UIHandler{logger: logger}
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

// Generate unique analysis ID
func generateAnalysisID() string {
	return uuid.New().String()
}

// GenerateAnalysisID is the exported version for testing
func GenerateAnalysisID() string {
	return generateAnalysisID()
}
