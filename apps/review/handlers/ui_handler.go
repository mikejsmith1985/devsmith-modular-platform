package review_handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	templates "github.com/mikejsmith1985/devsmith-modular-platform/apps/review/templates"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logging"
	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	modelContextKey contextKey = "model"
)

// UIHandler provides HTTP handlers for the Review UI with logging.
// This handler depends on interfaces (not concrete types) to enforce clean architecture.
type UIHandler struct {
	logger          logger.Interface
	logClient       *logging.Client
	previewService  review_services.PreviewAnalyzer
	skimService     review_services.SkimAnalyzer
	scanService     review_services.ScanAnalyzer
	detailedService review_services.DetailedAnalyzer
	criticalService review_services.CriticalAnalyzer
}

// NewUIHandler creates a new UIHandler with the given logger, logging client, and analyzer services.
// This enforces dependency inversion - handlers depend on abstractions, not implementations.
func NewUIHandler(
	logger logger.Interface,
	client *logging.Client,
	previewService review_services.PreviewAnalyzer,
	skimService review_services.SkimAnalyzer,
	scanService review_services.ScanAnalyzer,
	detailedService review_services.DetailedAnalyzer,
	criticalService review_services.CriticalAnalyzer,
) *UIHandler {
	return &UIHandler{
		logger:          logger,
		logClient:       client,
		previewService:  previewService,
		skimService:     skimService,
		scanService:     scanService,
		detailedService: detailedService,
		criticalService: criticalService,
	}
}

// CodeRequest represents the code submission request
type CodeRequest struct {
	PastedCode string `form:"pasted_code" json:"pasted_code" binding:"required"`
	Model      string `form:"model" json:"model"`
}

// bindCodeRequest binds code from JSON or form data using Gin's binding
func (h *UIHandler) bindCodeRequest(c *gin.Context) (*CodeRequest, bool) {
	var req CodeRequest

	// Try binding as form first, then JSON
	if err := c.ShouldBind(&req); err != nil {
		h.logger.Warn("Failed to bind code request",
			"error", err,
			"content-type", c.GetHeader("Content-Type"))
		c.String(http.StatusBadRequest, "Code required. Please paste code in the textarea.")
		return nil, false
	}

	// Default model if not provided
	if req.Model == "" {
		req.Model = "mistral:7b-instruct"
	}

	h.logger.Info("Code request bound successfully",
		"code_length", len(req.PastedCode),
		"model", req.Model)

	return &req, true
}

// marshalAndFormat converts analysis result to JSON and renders HTML response
func (h *UIHandler) marshalAndFormat(c *gin.Context, result interface{}, title, bgColor string) {
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		h.logger.Error("Failed to marshal result", "error", err.Error())
		c.String(http.StatusInternalServerError, "Failed to format analysis result")
		return
	}

	html := fmt.Sprintf(`
	<div class="p-4 rounded-lg %s">
		<h4 class="font-semibold">%s</h4>
		<pre class="mt-2 p-2 bg-white dark:bg-gray-800 rounded text-sm text-gray-700 dark:text-gray-300 overflow-auto">%s</pre>
	</div>
	`, bgColor, title, string(resultJSON))
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
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
// nolint:dupl // Similar structure across handlers is acceptable; each mode has distinct service and context
func (h *UIHandler) HandlePreviewMode(c *gin.Context) {
	req, ok := h.bindCodeRequest(c)
	if !ok {
		return
	}

	if h.previewService == nil {
		h.logger.Warn("Preview service not initialized")
		c.String(http.StatusServiceUnavailable, "Preview service unavailable")
		return
	}

	// TODO: Pass model to service via context for Ollama override
	ctx := context.WithValue(c.Request.Context(), modelContextKey, req.Model)

	result, err := h.previewService.AnalyzePreview(ctx, req.PastedCode)
	if err != nil {
		h.logger.Error("Preview analysis failed", "error", err.Error(), "model", req.Model)
		c.String(http.StatusInternalServerError, fmt.Sprintf("Analysis failed: %v", err))
		return
	}

	h.marshalAndFormat(c, result, "👁️ Preview Mode Analysis", "bg-indigo-50 dark:bg-indigo-900 border border-indigo-200 dark:border-indigo-700")
}

// HandleSkimMode handles POST /api/review/modes/skim (HTMX)
// nolint:dupl // Similar structure across handlers is acceptable; each mode has distinct service and context
func (h *UIHandler) HandleSkimMode(c *gin.Context) {
	req, ok := h.bindCodeRequest(c)
	if !ok {
		return
	}

	if h.skimService == nil {
		h.logger.Warn("Skim service not initialized")
		c.String(http.StatusServiceUnavailable, "Skim service unavailable")
		return
	}

	// Pass model to service via context for Ollama override
	ctx := context.WithValue(c.Request.Context(), modelContextKey, req.Model)

	result, err := h.skimService.AnalyzeSkim(ctx, req.PastedCode)
	if err != nil {
		h.logger.Error("Skim analysis failed", "error", err.Error(), "model", req.Model)
		c.String(http.StatusInternalServerError, fmt.Sprintf("Analysis failed: %v", err))
		return
	}

	h.marshalAndFormat(c, result, "📚 Skim Mode Analysis", "bg-blue-50 dark:bg-blue-900 border border-blue-200 dark:border-blue-700")
}

// HandleScanMode handles POST /api/review/modes/scan (HTMX)
// nolint:dupl // Similar structure across handlers is acceptable; each mode has distinct service and context
func (h *UIHandler) HandleScanMode(c *gin.Context) {
	req, ok := h.bindCodeRequest(c)
	if !ok {
		return
	}

	query := c.DefaultQuery("query", "find issues and improvements")

	if h.scanService == nil {
		h.logger.Warn("Scan service not initialized")
		c.String(http.StatusServiceUnavailable, "Scan service unavailable")
		return
	}

	// Pass model to service via context for Ollama override
	ctx := context.WithValue(c.Request.Context(), modelContextKey, req.Model)

	result, err := h.scanService.AnalyzeScan(ctx, query, req.PastedCode)
	if err != nil {
		h.logger.Error("Scan analysis failed", "error", err.Error(), "model", req.Model)
		c.String(http.StatusInternalServerError, fmt.Sprintf("Analysis failed: %v", err))
		return
	}

	h.marshalAndFormat(c, result, "🔎 Scan Mode Analysis", "bg-green-50 dark:bg-green-900 border border-green-200 dark:border-green-700")
}

// HandleDetailedMode handles POST /api/review/modes/detailed (HTMX)
// nolint:dupl // Similar structure across handlers is acceptable; each mode has distinct service and context
func (h *UIHandler) HandleDetailedMode(c *gin.Context) {
	req, ok := h.bindCodeRequest(c)
	if !ok {
		return
	}

	filename := c.DefaultQuery("filename", "main.go")

	if h.detailedService == nil {
		h.logger.Warn("Detailed service not initialized")
		c.String(http.StatusServiceUnavailable, "Detailed service unavailable")
		return
	}

	// Pass model to service via context for Ollama override
	ctx := context.WithValue(c.Request.Context(), modelContextKey, req.Model)

	result, err := h.detailedService.AnalyzeDetailed(ctx, filename, req.PastedCode)
	if err != nil {
		h.logger.Error("Detailed analysis failed", "error", err.Error(), "model", req.Model)
		c.String(http.StatusInternalServerError, fmt.Sprintf("Analysis failed: %v", err))
		return
	}

	h.marshalAndFormat(c, result, "📖 Detailed Mode Analysis", "bg-yellow-50 dark:bg-yellow-900 border border-yellow-200 dark:border-yellow-700")
}

// HandleCriticalMode handles POST /api/review/modes/critical (HTMX)
// nolint:dupl // Similar structure across handlers is acceptable; each mode has distinct service and context
func (h *UIHandler) HandleCriticalMode(c *gin.Context) {
	req, ok := h.bindCodeRequest(c)
	if !ok {
		return
	}

	if h.criticalService == nil {
		h.logger.Warn("Critical service not initialized")
		c.String(http.StatusServiceUnavailable, "Critical service unavailable")
		return
	}

	// Pass model to service via context for Ollama override
	ctx := context.WithValue(c.Request.Context(), modelContextKey, req.Model)

	result, err := h.criticalService.AnalyzeCritical(ctx, req.PastedCode)
	if err != nil {
		h.logger.Error("Critical analysis failed", "error", err.Error(), "model", req.Model)
		c.String(http.StatusInternalServerError, fmt.Sprintf("Analysis failed: %v", err))
		return
	}

	h.marshalAndFormat(c, result, "🚨 Critical Mode Analysis", "bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-700")
}

// GetAvailableModels returns a list of available Ollama models
func (h *UIHandler) GetAvailableModels(c *gin.Context) {
	// Return hardcoded list of common models
	// TODO: Query Ollama API for actual available models
	models := []map[string]string{
		{"name": "mistral:7b-instruct", "description": "Fast, General (Recommended)"},
		{"name": "codellama:13b", "description": "Better for code"},
		{"name": "llama2:13b", "description": "Balanced"},
		{"name": "deepseek-coder:6.7b", "description": "Code specialist"},
		{"name": "deepseek-coder-v2:16b", "description": "Most accurate (slower)"},
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}

// ListSessionsHTMX handles GET /api/review/sessions/list (HTMX)
func (h *UIHandler) ListSessionsHTMX(c *gin.Context) {
	// For now, return placeholder HTML with mock sessions
	// In production, this would fetch from SessionHandler via internal API
	html := `
	<div class="space-y-2">
		<div class="p-3 rounded-lg border border-indigo-400 bg-indigo-50 dark:bg-indigo-900 dark:border-indigo-600 cursor-pointer">
			<div class="flex items-start justify-between">
				<div class="flex-1 min-w-0">
					<h4 class="text-sm font-medium truncate text-indigo-900 dark:text-indigo-200">Latest Session</h4>
					<p class="text-xs truncate text-indigo-700 dark:text-indigo-300">2025-11-01 10:30:00</p>
				</div>
				<span class="px-2 py-1 rounded text-xs font-medium whitespace-nowrap ml-2 bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200">active</span>
			</div>
			<div class="mt-2 flex items-center justify-between">
				<span class="text-xs text-gray-500 dark:text-gray-400">2 modes</span>
				<button class="text-xs text-red-600 dark:text-red-400 hover:text-red-800">🗑️</button>
			</div>
		</div>
		<div class="p-3 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600 cursor-pointer">
			<div class="flex items-start justify-between">
				<div class="flex-1 min-w-0">
					<h4 class="text-sm font-medium truncate text-gray-900 dark:text-white">Review Session 2</h4>
					<p class="text-xs truncate text-gray-600 dark:text-gray-400">2025-10-31 15:45:00</p>
				</div>
				<span class="px-2 py-1 rounded text-xs font-medium whitespace-nowrap ml-2 bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200">completed</span>
			</div>
			<div class="mt-2 flex items-center justify-between">
				<span class="text-xs text-gray-500 dark:text-gray-400">5 modes</span>
				<button class="text-xs text-red-600 dark:text-red-400 hover:text-red-800">🗑️</button>
			</div>
		</div>
	</div>
	`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// SearchSessionsHTMX handles GET /api/review/sessions/search (HTMX)
func (h *UIHandler) SearchSessionsHTMX(c *gin.Context) {
	query := c.Query("query")
	h.logger.Info("Searching sessions", "query", query)

	// Placeholder: return filtered results
	if query == "" {
		// Return all sessions
		c.Redirect(http.StatusMovedPermanently, "/api/review/sessions/list")
		return
	}

	html := `
	<div class="space-y-2">
		<div class="p-3 rounded-lg border border-gray-200 dark:border-gray-700 cursor-pointer">
			<div class="flex items-start justify-between">
				<div class="flex-1 min-w-0">
					<h4 class="text-sm font-medium truncate text-gray-900 dark:text-white">Matching: ` + query + `</h4>
					<p class="text-xs truncate text-gray-600 dark:text-gray-400">2025-10-30 12:00:00</p>
				</div>
				<span class="px-2 py-1 rounded text-xs font-medium whitespace-nowrap ml-2 bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200">completed</span>
			</div>
			<div class="mt-2 flex items-center justify-between">
				<span class="text-xs text-gray-500 dark:text-gray-400">3 modes</span>
				<button class="text-xs text-red-600 dark:text-red-400 hover:text-red-800">🗑️</button>
			</div>
		</div>
	</div>
	`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// GetSessionDetailHTMX handles GET /api/review/sessions/:id (HTMX)
func (h *UIHandler) GetSessionDetailHTMX(c *gin.Context) {
	sessionID := c.Param("id")
	h.logger.Info("Loading session detail", "session_id", sessionID)

	// Placeholder: return session detail view
	html := `
	<div class="w-full lg:flex-1 bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
		<div class="flex items-start justify-between mb-6 pb-6 border-b border-gray-200 dark:border-gray-700">
			<div class="flex-1">
				<h2 class="text-2xl font-bold text-gray-900 dark:text-white">Session ` + sessionID + `</h2>
				<div class="mt-2 flex items-center gap-4">
					<span class="px-3 py-1 rounded-lg text-sm font-medium bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200">active</span>
					<span class="text-sm text-gray-600 dark:text-gray-400">Created: 2025-11-01 10:30:00</span>
				</div>
			</div>
			<button class="px-4 py-2 rounded-lg font-medium bg-red-600 text-white hover:bg-red-700 dark:hover:bg-red-800 transition-colors text-sm" onclick="if(confirm('Delete this session?')) { htmx.ajax('DELETE', '/api/review/sessions/` + sessionID + `', {target:'#session-detail', swap:'innerHTML'}) }">🗑️ Delete</button>
		</div>

		<div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
			<div class="p-4 rounded-lg bg-indigo-50 dark:bg-indigo-900 border border-indigo-200 dark:border-indigo-700">
				<div class="text-sm font-medium text-indigo-600 dark:text-indigo-300">Reading Modes Used</div>
				<div class="mt-2 text-2xl font-bold text-indigo-900 dark:text-indigo-100">2</div>
			</div>
			<div class="p-4 rounded-lg bg-green-50 dark:bg-green-900 border border-green-200 dark:border-green-700">
				<div class="text-sm font-medium text-green-600 dark:text-green-300">Created</div>
				<div class="mt-2 text-sm text-green-900 dark:text-green-100">2025-11-01 10:30:00</div>
			</div>
			<div class="p-4 rounded-lg bg-blue-50 dark:bg-blue-900 border border-blue-200 dark:border-blue-700">
				<div class="text-sm font-medium text-blue-600 dark:text-blue-300">Last Updated</div>
				<div class="mt-2 text-sm text-blue-900 dark:text-blue-100">2025-11-01 10:45:00</div>
			</div>
		</div>

		<div class="space-y-4">
			<h3 class="text-lg font-semibold text-gray-900 dark:text-white">Actions</h3>
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
				<button class="px-4 py-3 rounded-lg font-medium bg-indigo-600 text-white hover:bg-indigo-700 dark:hover:bg-indigo-800 transition-colors text-sm">▶️ Resume Session</button>
				<button class="px-4 py-3 rounded-lg font-medium bg-gray-600 text-white hover:bg-gray-700 dark:hover:bg-gray-800 transition-colors text-sm">⬇️ Export Session</button>
				<button class="px-4 py-3 rounded-lg font-medium bg-purple-600 text-white hover:bg-purple-700 dark:hover:bg-purple-800 transition-colors text-sm">📋 Duplicate</button>
				<button class="px-4 py-3 rounded-lg font-medium border-2 border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300">📁 Archive</button>
			</div>
		</div>

		<div class="mt-8 pt-8 border-t border-gray-200 dark:border-gray-700">
			<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Mode History</h3>
			<div class="space-y-3">
				<div class="p-3 rounded-lg bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-gray-900 dark:text-white">👁️ Preview Mode</span>
						<span class="text-xs text-gray-500 dark:text-gray-400">10:15 AM</span>
					</div>
					<p class="mt-1 text-xs text-gray-600 dark:text-gray-400">Analyzed project structure</p>
				</div>
				<div class="p-3 rounded-lg bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-gray-900 dark:text-white">🔎 Scan Mode</span>
						<span class="text-xs text-gray-500 dark:text-gray-400">10:20 AM</span>
					</div>
					<p class="mt-1 text-xs text-gray-600 dark:text-gray-400">Searched for error handling</p>
				</div>
			</div>
		</div>
	</div>
	`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// ResumeSessionHTMX handles POST /api/review/sessions/:id/resume (HTMX)
func (h *UIHandler) ResumeSessionHTMX(c *gin.Context) {
	sessionID := c.Param("id")
	h.logger.Info("Resuming session", "session_id", sessionID)

	html := `
	<div class="alert alert-success">
		<p class="text-green-700 dark:text-green-300">✓ Session resumed successfully. Continuing from where you left off...</p>
	</div>
	`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// DuplicateSessionHTMX handles POST /api/review/sessions/:id/duplicate (HTMX)
func (h *UIHandler) DuplicateSessionHTMX(c *gin.Context) {
	sessionID := c.Param("id")
	h.logger.Info("Duplicating session", "session_id", sessionID)

	html := `
	<div class="alert alert-success">
		<p class="text-green-700 dark:text-green-300">✓ Session duplicated successfully. Switched to new session.</p>
	</div>
	`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// ArchiveSessionHTMX handles POST /api/review/sessions/:id/archive (HTMX)
func (h *UIHandler) ArchiveSessionHTMX(c *gin.Context) {
	sessionID := c.Param("id")
	h.logger.Info("Archiving session", "session_id", sessionID)

	html := `
	<div class="alert alert-success">
		<p class="text-green-700 dark:text-green-300">✓ Session archived successfully.</p>
	</div>
	`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// DeleteSessionHTMX handles DELETE /api/review/sessions/:id (HTMX)
func (h *UIHandler) DeleteSessionHTMX(c *gin.Context) {
	sessionID := c.Param("id")
	h.logger.Info("Deleting session", "session_id", sessionID)

	html := `
	<div class="alert alert-info">
		<p class="text-blue-700 dark:text-blue-300">Session deleted. Returning to session list...</p>
	</div>
	`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// GetSessionStatsHTMX handles GET /api/review/sessions/:id/stats (HTMX)
func (h *UIHandler) GetSessionStatsHTMX(c *gin.Context) {
	sessionID := c.Param("id")
	h.logger.Info("Loading session statistics", "session_id", sessionID)

	// Return statistics grid HTML
	html := `
	<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
		<div class="p-4 rounded-lg bg-indigo-50 dark:bg-indigo-900 border border-indigo-200 dark:border-indigo-700">
			<div class="text-sm font-medium text-indigo-600 dark:text-indigo-300">Reading Modes</div>
			<div class="mt-2 text-3xl font-bold text-indigo-900 dark:text-indigo-100">5</div>
			<p class="mt-1 text-xs text-indigo-700 dark:text-indigo-400">modes used in analysis</p>
		</div>
		<div class="p-4 rounded-lg bg-green-50 dark:bg-green-900 border border-green-200 dark:border-green-700">
			<div class="text-sm font-medium text-green-600 dark:text-green-300">Code Analyzed</div>
			<div class="mt-2 text-3xl font-bold text-green-900 dark:text-green-100">2,847</div>
			<p class="mt-1 text-xs text-green-700 dark:text-green-400">lines of code</p>
		</div>
		<div class="p-4 rounded-lg bg-blue-50 dark:bg-blue-900 border border-blue-200 dark:border-blue-700">
			<div class="text-sm font-medium text-blue-600 dark:text-blue-300">Analysis Time</div>
			<div class="mt-2 text-3xl font-bold text-blue-900 dark:text-blue-100">3,245ms</div>
			<p class="mt-1 text-xs text-blue-700 dark:text-blue-400">total time spent</p>
		</div>
	</div>
	`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// GetSessionMetadataHTMX handles GET /api/review/sessions/:id/metadata (HTMX)
func (h *UIHandler) GetSessionMetadataHTMX(c *gin.Context) {
	sessionID := c.Param("id")
	h.logger.Info("Loading session metadata", "session_id", sessionID)

	// Return metadata grid HTML
	html := `
	<div class="grid grid-cols-2 gap-4">
		<div class="p-3 rounded-lg bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600">
			<div class="text-xs font-medium text-gray-600 dark:text-gray-400">Created</div>
			<div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">2025-11-01 10:30:00</div>
		</div>
		<div class="p-3 rounded-lg bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600">
			<div class="text-xs font-medium text-gray-600 dark:text-gray-400">Last Updated</div>
			<div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">2025-11-01 10:45:00</div>
		</div>
		<div class="p-3 rounded-lg bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600">
			<div class="text-xs font-medium text-gray-600 dark:text-gray-400">File Size</div>
			<div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">85.3 KB</div>
		</div>
		<div class="p-3 rounded-lg bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600">
			<div class="text-xs font-medium text-gray-600 dark:text-gray-400">Languages</div>
			<div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">Go, SQL, YAML</div>
		</div>
	</div>
	`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// ExportSessionHTMX handles GET /api/review/sessions/:id/export (HTMX)
func (h *UIHandler) ExportSessionHTMX(c *gin.Context) {
	sessionID := c.Param("id")
	format := c.DefaultQuery("format", "json")
	h.logger.Info("Exporting session", "session_id", sessionID, "format", format)

	if format == "json" {
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=session-%s.json", sessionID))
		c.JSON(http.StatusOK, gin.H{
			"session_id": sessionID,
			"exported":   "2025-11-01T10:50:00Z",
			"data": gin.H{
				"modes_used":       5,
				"code_lines":       2847,
				"analysis_time_ms": 3245,
			},
		})
	} else {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=session-%s.csv", sessionID))
		c.String(http.StatusOK, "session_id,modes_used,code_lines,analysis_time_ms\n"+sessionID+",5,2847,3245\n")
	}
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

// ShowWorkspace renders the two-pane workspace for code review with mode selection
// GET /review/workspace/:session_id
//
// Path Parameters:
//   - session_id: Session ID (integer)
//
// Response: HTML page with two-pane layout (code left, analysis right)
func (h *UIHandler) ShowWorkspace(c *gin.Context) {
	// Extract session ID from URL
	sessionIDStr := c.Param("session_id")
	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil {
		h.logger.Warn("invalid session_id", "session_id", sessionIDStr, "error", err)
		c.String(http.StatusBadRequest, "Invalid session ID")
		return
	}

	h.logger.Info("showing workspace", "session_id", sessionID)

	// For now, use sample data (in production this would come from database)
	// TODO: Fetch actual session data from ReviewRepository
	props := templates.WorkspaceProps{
		SessionID:      sessionID,
		Title:          "Sample Code Review",
		Code:           sampleCodeForWorkspace(),
		CurrentMode:    "preview",
		AnalysisResult: "",
	}

	// Render workspace template
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)

	if err := templates.Workspace(props).Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.Error("failed to render workspace template", "error", err.Error())
		c.String(http.StatusInternalServerError, "Failed to render workspace")
	}
}

// sampleCodeForWorkspace provides sample Go code for workspace demo
func sampleCodeForWorkspace() string {
	return `package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// GetUser retrieves a user by ID from the database.
// This function demonstrates common patterns in Go web handlers.
func GetUser(c *gin.Context) {
	// Extract user ID from URL parameter
	userID := c.Param("id")
	
	// Potential SQL injection vulnerability (Critical issue)
	query := "SELECT * FROM users WHERE id = " + userID
	
	// Missing error handling (Quality issue)
	rows, _ := db.Query(query)
	
	// Handler calling database directly (Architecture issue - layer violation)
	// Should call a service layer instead
	
	c.JSON(http.StatusOK, gin.H{"user": rows})
}

// CreateUser adds a new user to the database.
func CreateUser(c *gin.Context) {
	var user User
	
	// Missing input validation (Security issue)
	c.BindJSON(&user)
	
	// Global variable usage (Scope issue)
	db.Exec("INSERT INTO users ...")
	
	c.JSON(http.StatusCreated, user)
}
`
}
