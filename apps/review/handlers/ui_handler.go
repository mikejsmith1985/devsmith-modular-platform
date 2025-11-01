package review_handlers

import (
	"fmt"
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

// CreateSessionHandler handles POST /api/review/sessions (HTMX form submission)
func (h *UIHandler) CreateSessionHandler(c *gin.Context) {
	var req struct {
		PastedCode string `form:"pasted_code" json:"pasted_code"`
		GitHubURL  string `form:"github_url" json:"github_url"`
		File       string `form:"file" json:"file"`
	}

	// Parse form data
	if err := c.ShouldBind(&req); err != nil {
		h.logger.Error("Failed to parse form", "error", err)
		c.String(http.StatusBadRequest, `<div class="alert alert-error"><p>Invalid form submission</p></div>`)
		return
	}

	// Validate at least one input
	if req.PastedCode == "" && req.GitHubURL == "" && req.File == "" {
		c.String(http.StatusBadRequest, `<div class="alert alert-error"><p>Please provide code, GitHub URL, or upload a file</p></div>`)
		return
	}

	// Generate session ID
	sessionID := uuid.New().String()
	h.logger.Info("Session created", "session_id", sessionID, "source", "form")

	// Return HTML with SSE progress indicator
	progressHTML := fmt.Sprintf(`
<div class="mt-8 space-y-4">
	<div class="alert alert-info">
		<p>Session %s created. Starting analysis...</p>
	</div>
	<div id="progress-stream" hx-sse="connect:/api/review/sessions/%s/progress" class="mt-4">
		<div class="flex items-center gap-2 p-4 bg-blue-50 dark:bg-blue-900 rounded-lg">
			<span class="loading loading-spinner loading-sm text-blue-600 dark:text-blue-400"></span>
			<span class="text-sm text-blue-900 dark:text-blue-100">Analyzing your code...</span>
		</div>
	</div>
</div>
`, sessionID, sessionID)

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, progressHTML)
}

// HandlePreviewMode handles POST /api/review/modes/preview (HTMX)
func (h *UIHandler) HandlePreviewMode(c *gin.Context) {
	var req struct {
		Code string `form:"pasted_code" json:"code"`
	}

	// Try JSON binding first, then form
	if err := c.ShouldBindJSON(&req); err != nil {
		if err := c.ShouldBind(&req); err != nil {
			c.String(http.StatusBadRequest, "Code required")
			return
		}
	}

	if req.Code == "" {
		c.String(http.StatusBadRequest, "Code required")
		return
	}

	// Return preview result component (for now, simple HTML)
	html := `
	<section class="card">
		<h3 class="text-xl font-bold mb-4">👁️ Preview Mode Results</h3>
		<div class="space-y-4">
			<div>
				<h4 class="font-semibold text-gray-700 dark:text-gray-300">File Tree</h4>
				<ul class="list-disc list-inside text-sm text-gray-600 dark:text-gray-400">
					<li>main.go</li>
					<li>utils.go</li>
					<li>handlers/</li>
					<li>services/</li>
				</ul>
			</div>
			<div>
				<h4 class="font-semibold text-gray-700 dark:text-gray-300">Bounded Contexts</h4>
				<ul class="list-disc list-inside text-sm text-gray-600 dark:text-gray-400">
					<li>Core Logic</li>
					<li>API Layer</li>
					<li>Data Access</li>
				</ul>
			</div>
			<div>
				<h4 class="font-semibold text-gray-700 dark:text-gray-300">Tech Stack</h4>
				<div class="flex gap-2 flex-wrap">
					<span class="px-3 py-1 bg-indigo-100 dark:bg-indigo-900 text-indigo-800 dark:text-indigo-200 rounded-full text-sm font-medium">Go</span>
					<span class="px-3 py-1 bg-indigo-100 dark:bg-indigo-900 text-indigo-800 dark:text-indigo-200 rounded-full text-sm font-medium">PostgreSQL</span>
					<span class="px-3 py-1 bg-indigo-100 dark:bg-indigo-900 text-indigo-800 dark:text-indigo-200 rounded-full text-sm font-medium">Gin</span>
				</div>
			</div>
			<div>
				<h4 class="font-semibold text-gray-700 dark:text-gray-300">Architecture Pattern</h4>
				<p class="text-sm text-gray-600 dark:text-gray-400">Layered Architecture</p>
			</div>
			<div>
				<h4 class="font-semibold text-gray-700 dark:text-gray-300">Summary</h4>
				<p class="text-sm text-gray-600 dark:text-gray-400">Backend service with clean layering</p>
			</div>
		</div>
	</section>
	`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// HandleSkimMode handles POST /api/review/modes/skim (HTMX)
func (h *UIHandler) HandleSkimMode(c *gin.Context) {
	var req struct {
		Code string `form:"pasted_code" json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		if err := c.ShouldBind(&req); err != nil {
			c.String(http.StatusBadRequest, "Code required")
			return
		}
	}

	if req.Code == "" {
		c.String(http.StatusBadRequest, "Code required")
		return
	}

	// Placeholder response
	c.String(http.StatusOK, "<p>Skim mode analysis in progress...</p>")
}

// HandleScanMode handles POST /api/review/modes/scan (HTMX)
func (h *UIHandler) HandleScanMode(c *gin.Context) {
	var req struct {
		Code string `form:"pasted_code" json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		if err := c.ShouldBind(&req); err != nil {
			c.String(http.StatusBadRequest, "Code required")
			return
		}
	}

	if req.Code == "" {
		c.String(http.StatusBadRequest, "Code required")
		return
	}

	// Placeholder response
	c.String(http.StatusOK, "<p>Scan mode analysis in progress...</p>")
}

// HandleDetailedMode handles POST /api/review/modes/detailed (HTMX)
func (h *UIHandler) HandleDetailedMode(c *gin.Context) {
	var req struct {
		Code string `form:"pasted_code" json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		if err := c.ShouldBind(&req); err != nil {
			c.String(http.StatusBadRequest, "Code required")
			return
		}
	}

	if req.Code == "" {
		c.String(http.StatusBadRequest, "Code required")
		return
	}

	// Placeholder response
	c.String(http.StatusOK, "<p>Detailed mode analysis in progress...</p>")
}

// HandleCriticalMode handles POST /api/review/modes/critical (HTMX)
func (h *UIHandler) HandleCriticalMode(c *gin.Context) {
	var req struct {
		Code string `form:"pasted_code" json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		if err := c.ShouldBind(&req); err != nil {
			c.String(http.StatusBadRequest, "Code required")
			return
		}
	}

	if req.Code == "" {
		c.String(http.StatusBadRequest, "Code required")
		return
	}

	// Placeholder response
	c.String(http.StatusOK, "<p>Critical mode analysis in progress...</p>")
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
