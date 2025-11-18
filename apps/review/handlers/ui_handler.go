package review_handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	templates "github.com/mikejsmith1985/devsmith-modular-platform/apps/review/templates"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logging"
	reviewcontext "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/context"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
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
	modelService    *review_services.ModelService
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
	modelService *review_services.ModelService,
) *UIHandler {
	return &UIHandler{
		logger:          logger,
		logClient:       client,
		previewService:  previewService,
		skimService:     skimService,
		scanService:     scanService,
		detailedService: detailedService,
		criticalService: criticalService,
		modelService:    modelService,
	}
}

// CodeRequest represents the code submission request
type CodeRequest struct {
	PastedCode string `form:"pasted_code" json:"pasted_code" binding:"required"`
	Model      string `form:"model" json:"model"`
	UserMode   string `form:"user_mode" json:"user_mode"`     // beginner, novice, intermediate, expert
	OutputMode string `form:"output_mode" json:"output_mode"` // quick, full
}

// bindCodeRequest binds code from JSON or form data using Gin's binding
func (h *UIHandler) bindCodeRequest(c *gin.Context) (*CodeRequest, bool) {
	var req CodeRequest

	// Try binding as form first, then JSON
	if err := c.ShouldBind(&req); err != nil {
		// If binding failed, attempt to accept a file upload fallback
		// Some clients may POST a file part (e.g. curl -F '@-') instead of a plain form field.
		// Try to read an uploaded file named 'pasted_code' and use its contents.
		if fileHeader, ferr := c.FormFile("pasted_code"); ferr == nil {
			fh, openErr := fileHeader.Open()
			if openErr == nil {
				defer func() {
					if err := fh.Close(); err != nil {
						h.logger.Error("Failed to close file handle", "error", err.Error())
					}
				}()
				if data, readErr := io.ReadAll(fh); readErr == nil {
					req.PastedCode = string(data)
					// try to bind model separately (optional)
					if m := c.PostForm("model"); m != "" {
						req.Model = m
					}
					// try to bind user_mode (optional)
					if um := c.PostForm("user_mode"); um != "" {
						req.UserMode = um
					}
					// try to bind output_mode (optional)
					if om := c.PostForm("output_mode"); om != "" {
						req.OutputMode = om
					}
					h.logger.Info("Code request bound from uploaded file",
						"code_length", len(req.PastedCode),
						"filename", fileHeader.Filename,
						"content-type", c.GetHeader("Content-Type"),
						"user_mode", req.UserMode,
						"output_mode", req.OutputMode)

					// Default model if not provided
					if req.Model == "" {
						req.Model = "mistral:7b-instruct"
					}

					// Default user_mode if not provided
					if req.UserMode == "" {
						req.UserMode = "intermediate"
					}

					// Default output_mode if not provided
					if req.OutputMode == "" {
						req.OutputMode = "quick"
					}

					return &req, true
				}
			}
		}

		h.logger.Warn("Failed to bind code request",
			"error", err.Error(),
			"content-type", c.GetHeader("Content-Type"))
		c.String(http.StatusBadRequest, "Code required. Please paste code in the textarea.")
		return nil, false
	}

	// Default model if not provided
	if req.Model == "" {
		req.Model = "mistral:7b-instruct"
	}

	// Default user_mode if not provided (defaults to intermediate)
	if req.UserMode == "" {
		req.UserMode = "intermediate"
	}

	// Default output_mode if not provided (defaults to quick)
	if req.OutputMode == "" {
		req.OutputMode = "quick"
	}

	h.logger.Info("Code request bound successfully",
		"code_length", len(req.PastedCode),
		"model", req.Model,
		"user_mode", req.UserMode,
		"output_mode", req.OutputMode)

	return &req, true
}

// looksLikeCode performs a lightweight heuristic check to determine whether the
// provided text looks like source code. This prevents modes that expect source
// code (Skim/Detailed) from hallucinating when given natural language input.
func looksLikeCode(s string) bool {
	if s == "" {
		return false
	}
	// heuristics: common code tokens across languages
	checks := []string{"package ", "func ", "class ", "import ", "def ", "struct ", "interface ", "=>", "->", "{", "}"}
	score := 0
	for _, token := range checks {
		if strings.Contains(s, token) {
			score++
		}
	}
	// if multiple tokens present, very likely code
	return score >= 1
}

// marshalAndFormat converts analysis result to user-friendly HTML


func (h *UIHandler) renderPreviewHTML(w http.ResponseWriter, result *review_models.PreviewModeOutput) {
	html := `<div class="space-y-6 p-6 bg-indigo-50 dark:bg-indigo-900 rounded-lg border border-indigo-200 dark:border-indigo-700">
		<div class="flex items-center gap-3 border-b border-indigo-200 dark:border-gray-700 pb-4">
			<span class="text-3xl">👁️</span>
			<div><h3 class="text-xl font-bold text-indigo-900 dark:text-indigo-50">Quick Preview</h3>
			<p class="text-sm text-indigo-700 dark:text-indigo-200">High-level structure and overview</p></div>
		</div>`

	if result.Summary != "" {
		html += fmt.Sprintf(`<div class="prose prose-sm dark:prose-invert max-w-none">
			<h4 class="text-lg font-semibold text-indigo-900 dark:text-indigo-50 mb-2">📋 Summary</h4>
			<p class="text-gray-700 dark:text-indigo-100 leading-relaxed">%s</p>
		</div>`, result.Summary)
	}

	if len(result.BoundedContexts) > 0 {
		html += `<div><h4 class="text-lg font-semibold text-indigo-900 dark:text-gray-100 mb-3">🎯 Key Areas</h4><div class="grid gap-2">`
		for _, ctx := range result.BoundedContexts {
			html += fmt.Sprintf(`<div class="p-3 bg-white dark:bg-indigo-800 rounded-lg border border-indigo-100 dark:border-indigo-700">
				<span class="font-medium text-indigo-700 dark:text-indigo-50">%s</span></div>`, ctx)
		}
		html += `</div></div>`
	}

	if len(result.TechStack) > 0 {
		html += `<div><h4 class="text-lg font-semibold text-indigo-900 dark:text-gray-100 mb-3">🔧 Technologies Used</h4><div class="flex flex-wrap gap-2">`
		for _, tech := range result.TechStack {
			html += fmt.Sprintf(`<span class="px-3 py-1 bg-indigo-100 dark:bg-indigo-800 text-indigo-800 dark:text-indigo-100 rounded-full text-sm font-medium">%s</span>`, tech)
		}
		html += `</div></div>`
	}

	html += `</div>`
	if _, err := fmt.Fprint(w, html); err != nil {
		h.logger.Error("Failed to write preview response", "error", err.Error())
	}
}

func (h *UIHandler) renderSkimHTML(w http.ResponseWriter, result *review_models.SkimModeOutput) {
	html := `<div class="space-y-6 p-6 bg-blue-50 dark:bg-slate-800 rounded-lg border border-blue-200 dark:border-slate-700">
		<div class="flex items-center gap-3 border-b border-blue-200 dark:border-slate-700 pb-4">
			<span class="text-3xl">📚</span>
			<div><h3 class="text-xl font-bold text-blue-900 dark:text-slate-50">Skim Analysis</h3>
			<p class="text-sm text-blue-700 dark:text-slate-200">Key components and abstractions</p></div>
		</div>`

	// If the service returned a summary (used when input wasn't code), show it prominently
	if result.Summary != "" {
		html += fmt.Sprintf(`<div class="p-3 bg-white dark:bg-slate-800 rounded-lg border border-blue-100 dark:border-slate-700">
			<h4 class="text-sm font-semibold text-slate-900 dark:text-slate-50 mb-2">ℹ️ Note</h4>
			<p class="text-sm text-gray-700 dark:text-slate-200">%s</p>
		</div>`, result.Summary)
	}

	if len(result.Functions) > 0 {
		html += `<div><h4 class="text-lg font-semibold text-blue-900 dark:text-gray-100 mb-3">⚡ Functions & Methods</h4><div class="space-y-3">`
		for _, fn := range result.Functions {
			html += fmt.Sprintf(`<div class="p-4 bg-white dark:bg-blue-800 rounded-lg border border-blue-100 dark:border-blue-700">
				<div class="font-mono text-sm text-blue-700 dark:text-blue-100 font-semibold mb-2">%s</div>
				<div class="font-mono text-xs text-gray-600 dark:text-blue-200 mb-2 pl-4">%s</div>
				<p class="text-sm text-gray-700 dark:text-blue-100 leading-relaxed">%s</p>
			</div>`, fn.Name, fn.Signature, fn.Description)
		}
		html += `</div></div>`
	}

	if len(result.Interfaces) > 0 {
		html += `<div><h4 class="text-lg font-semibold text-blue-900 dark:text-gray-100 mb-3">🔌 Interfaces</h4><div class="space-y-3">`
		for _, iface := range result.Interfaces {
			html += fmt.Sprintf(`<div class="p-4 bg-white dark:bg-blue-800 rounded-lg border border-blue-100 dark:border-blue-700">
				<div class="font-mono text-sm text-blue-700 dark:text-blue-100 font-semibold mb-2">%s</div>
				<p class="text-sm text-gray-700 dark:text-blue-100 leading-relaxed">%s</p>
			</div>`, iface.Name, iface.Description)
		}
		html += `</div></div>`
	}

	html += `</div>`
	if _, err := fmt.Fprint(w, html); err != nil {
		h.logger.Error("Failed to write skim response", "error", err.Error())
	}
}

func (h *UIHandler) renderScanHTML(w http.ResponseWriter, result *review_models.ScanModeOutput) {
	html := fmt.Sprintf(`<div class="space-y-6 p-6 bg-green-50 dark:bg-green-900 rounded-lg border border-green-200 dark:border-green-700">
		<div class="flex items-center gap-3 border-b border-green-200 dark:border-green-700 pb-4">
			<span class="text-3xl">🔎</span>
			<div><h3 class="text-xl font-bold text-green-900 dark:text-green-50">Search Results</h3>
			<p class="text-sm text-green-700 dark:text-green-200">Found %d matches</p></div>
		</div>`, len(result.Matches))

	if len(result.Matches) > 0 {
		html += `<div class="space-y-4">`
		for i, match := range result.Matches {
			html += fmt.Sprintf(`<div class="p-4 bg-white dark:bg-green-800 rounded-lg border border-green-100 dark:border-green-700">
				<div class="flex items-center justify-between mb-3">
					<span class="text-sm font-semibold text-green-700 dark:text-green-300">Match %d</span>
					<span class="text-xs font-mono text-gray-600 dark:text-gray-400">%s</span>
				</div>
				<div class="p-3 bg-gray-50 dark:bg-gray-900 rounded border border-gray-200 dark:border-gray-700 font-mono text-xs text-gray-700 dark:text-gray-300 whitespace-pre-wrap overflow-x-auto">%s</div>
			</div>`, i+1, match.FilePath, match.CodeSnippet)
		}
		html += `</div>`
	} else {
		html += `<div class="text-center py-8"><p class="text-gray-600 dark:text-gray-400">No matches found.</p></div>`
	}

	html += `</div>`
	if _, err := fmt.Fprint(w, html); err != nil {
		h.logger.Error("Failed to write scan response", "error", err.Error())
	}
}

func (h *UIHandler) renderDetailedHTML(w http.ResponseWriter, result *review_models.DetailedModeOutput) {
	html := `<div class="space-y-6 p-6 bg-yellow-50 dark:bg-yellow-900 rounded-lg border border-yellow-200 dark:border-yellow-700">
		<div class="flex items-center gap-3 border-b border-yellow-200 dark:border-yellow-700 pb-4">
			<span class="text-3xl">📖</span>
			<div><h3 class="text-xl font-bold text-yellow-900 dark:text-yellow-50">Detailed Analysis</h3>
			<p class="text-sm text-yellow-700 dark:text-yellow-200">Line-by-line explanation</p></div>
		</div>`

	if result.AlgorithmSummary != "" {
		html += fmt.Sprintf(`<div class="p-4 bg-yellow-100 dark:bg-yellow-800 rounded-lg border border-yellow-200 dark:border-yellow-700">
			<h4 class="text-sm font-semibold text-yellow-900 dark:text-yellow-50 mb-2">🧠 Algorithm Overview</h4>
			<p class="text-sm text-yellow-800 dark:text-yellow-100 leading-relaxed">%s</p>
		</div>`, result.AlgorithmSummary)
	}

	if len(result.LineExplanations) > 0 {
		html += `<div><h4 class="text-lg font-semibold text-yellow-900 dark:text-yellow-100 mb-3">📝 Line-by-Line Walkthrough</h4><div class="space-y-3">`
		for _, line := range result.LineExplanations {
			html += fmt.Sprintf(`<div class="p-4 bg-white dark:bg-yellow-900/10 rounded-lg border border-yellow-100 dark:border-yellow-700">
				<div class="flex items-start gap-3 mb-2">
					<span class="flex-shrink-0 w-12 text-right font-mono text-xs text-yellow-600 dark:text-yellow-300 font-semibold">L%d</span>
					<div class="flex-1">
						<div class="font-mono text-sm text-gray-700 dark:text-yellow-100 mb-2 p-2 bg-gray-50 dark:bg-gray-900 rounded">%s</div>
						<p class="text-sm text-gray-700 dark:text-yellow-100 leading-relaxed">%s</p>
					</div>
				</div>
			</div>`, line.LineNumber, line.Code, line.Explanation)
		}
		html += `</div></div>`
	}

	html += `</div>`
	if _, err := fmt.Fprint(w, html); err != nil {
		h.logger.Error("Failed to write detailed response", "error", err.Error())
	}
}

func (h *UIHandler) renderCriticalHTML(w http.ResponseWriter, result *review_models.CriticalModeOutput) {
	html := fmt.Sprintf(`<div class="space-y-6 p-6 bg-red-50 dark:bg-red-900 rounded-lg border border-red-200 dark:border-red-700">
		<div class="flex items-center gap-3 border-b border-red-200 dark:border-red-700 pb-4">
			<span class="text-3xl">🚨</span>
			<div><h3 class="text-xl font-bold text-red-900 dark:text-red-50">Critical Review</h3>
			<p class="text-sm text-red-700 dark:text-red-200">Found %d issues</p></div>
		</div>`, len(result.Issues))

	if result.OverallGrade != "" {
		html += fmt.Sprintf(`<div class="text-center p-4 bg-white dark:bg-gray-800 rounded-lg">
			<div class="text-3xl font-bold text-red-600 dark:text-red-400">%s</div>
			<div class="text-sm text-gray-600 dark:text-gray-400">Overall Grade</div>
		</div>`, result.OverallGrade)
	}

	if len(result.Issues) > 0 {
		// Group by severity
		critical, high, medium, low := []review_models.CodeIssue{}, []review_models.CodeIssue{}, []review_models.CodeIssue{}, []review_models.CodeIssue{}
		for _, issue := range result.Issues {
			switch issue.Severity {
			case "critical":
				critical = append(critical, issue)
			case "high":
				high = append(high, issue)
			case "medium":
				medium = append(medium, issue)
			default:
				low = append(low, issue)
			}
		}

		// Critical
		if len(critical) > 0 {
			html += fmt.Sprintf(`<div><h4 class="text-lg font-semibold text-red-900 dark:text-red-50 mb-3">🔴 Critical Issues (%d)</h4><div class="space-y-3">`, len(critical))
			for _, issue := range critical {
				html += fmt.Sprintf(`<div class="p-4 bg-white dark:bg-red-900/5 rounded-lg border border-red-200 dark:border-red-700">
					<div class="text-sm font-semibold text-red-700 mb-2">%s <span class="text-xs text-gray-600 dark:text-gray-300">%s:%d</span></div>
					<p class="text-sm text-gray-700 dark:text-gray-300 mb-2">%s</p>
					<div class="p-3 bg-green-50 dark:bg-green-900/20 rounded border border-green-200">
						<div class="text-xs font-semibold text-green-700 mb-1">💡 Fix:</div>
						<p class="text-sm text-green-800 dark:text-green-200">%s</p>
					</div>
				</div>`, issue.Category, issue.File, issue.Line, issue.Description, issue.FixSuggestion)
			}
			html += `</div></div>`
		}

		// High
		if len(high) > 0 {
			html += fmt.Sprintf(`<div><h4 class="text-lg font-semibold text-orange-900 mb-3">🟠 High Priority (%d)</h4><div class="space-y-3">`, len(high))
			for _, issue := range high {
				html += fmt.Sprintf(`<div class="p-4 bg-white dark:bg-gray-800 rounded-lg border border-orange-200">
					<div class="text-sm font-semibold text-orange-700 mb-2">%s</div>
					<p class="text-sm text-gray-700 dark:text-gray-300">%s</p>
				</div>`, issue.Category, issue.Description)
			}
			html += `</div></div>`
		}
	} else {
		html += `<div class="text-center py-8">
			<span class="text-5xl mb-4 block">✅</span>
			<p class="text-lg font-semibold text-green-700">No critical issues found!</p>
		</div>`
	}

	html += `</div>`
	if _, err := fmt.Fprint(w, html); err != nil {
		h.logger.Error("Failed to write critical response", "error", err.Error())
	}
}

// renderError classifies the error and renders appropriate HTMX-compatible error template
func (h *UIHandler) renderError(c *gin.Context, err error, fallbackMessage string) {
	h.logger.Error("Request error", "error", err.Error(), "path", c.Request.URL.Path)

	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusInternalServerError)

	// Classify error and render appropriate template
	errMsg := err.Error()
	if strings.Contains(errMsg, "circuit breaker is open") || strings.Contains(errMsg, "ErrOpenState") {
		if renderErr := templates.CircuitOpen().Render(c.Request.Context(), c.Writer); renderErr != nil {
			h.logger.Error("Failed to render circuit open template", "error", renderErr.Error())
		}
	} else if strings.Contains(errMsg, "context deadline exceeded") || strings.Contains(errMsg, "timeout") {
		if renderErr := templates.AITimeout().Render(c.Request.Context(), c.Writer); renderErr != nil {
			h.logger.Error("Failed to render AI timeout template", "error", renderErr.Error())
		}
	} else if strings.Contains(errMsg, "ollama") && strings.Contains(errMsg, "unavailable") {
		if renderErr := templates.AIServiceUnavailable().Render(c.Request.Context(), c.Writer); renderErr != nil {
			h.logger.Error("Failed to render AI service unavailable template", "error", renderErr.Error())
		}
	} else if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "no such host") {
		if renderErr := templates.AIServiceUnavailable().Render(c.Request.Context(), c.Writer); renderErr != nil {
			h.logger.Error("Failed to render AI service unavailable template", "error", renderErr.Error())
		}
	} else if strings.Contains(errMsg, "ERR_AI_RESPONSE_INVALID") || strings.Contains(strings.ToLower(errMsg), "invalid response") {
		// AI returned malformed JSON or couldn't be repaired. Show a helpful message
		// including any excerpt available in the error string to aid troubleshooting.
		excerpt := errMsg
		// attempt to extract a raw response excerpt marker if present
		if idx := strings.Index(errMsg, "Raw response excerpt:"); idx != -1 {
			excerpt = errMsg[idx:]
		} else if idx := strings.Index(errMsg, "Excerpt:"); idx != -1 {
			excerpt = errMsg[idx:]
		}
		if len(excerpt) > 1200 {
			excerpt = excerpt[:1200] + "..."
		}
		html := fmt.Sprintf(`<div class="p-6 rounded-lg bg-yellow-50 dark:bg-yellow-900 border border-yellow-200 dark:border-yellow-700">
			<h3 class="text-lg font-semibold text-yellow-900 dark:text-yellow-50">Analysis could not be parsed</h3>
			<p class="mt-2 text-sm text-gray-700 dark:text-yellow-100">The AI returned an unexpected response format and automatic repair failed. We've captured a short excerpt below to help debugging.</p>
			<pre class="mt-3 p-3 bg-white dark:bg-gray-800 rounded text-sm text-gray-700 dark:text-gray-200 overflow-auto">%s</pre>
			<p class="mt-3 text-sm text-gray-600 dark:text-gray-300">We saved the full AI output for 14 days for troubleshooting. Try again or choose a different model.</p>
		</div>`, templateEscape(excerpt))
		c.String(http.StatusOK, html)
	} else {
		// Generic error
		message := fallbackMessage
		if message == "" {
			message = fmt.Sprintf("Analysis failed: %v", err)
		}
		if renderErr := templates.ErrorDisplay("error", "Analysis Failed", message, true, "/api/review/retry").Render(c.Request.Context(), c.Writer); renderErr != nil {
			h.logger.Error("Failed to render error display template", "error", renderErr.Error())
		}
	}
}

// templateEscape performs a minimal HTML escape for safe insertion into templates
func templateEscape(s string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;")
	return replacer.Replace(s)
}

// HomeHandler serves the main Review UI - creates new authenticated session
func (h *UIHandler) HomeHandler(c *gin.Context) {
	correlationID := c.Request.Context().Value("correlation_id")
	h.logger.Info("HomeHandler called", "correlation_id", correlationID, "path", c.Request.URL.Path)

	// Extract authenticated user from Redis session context
	userID, exists := c.Get("user_id")
	username, _ := c.Get("github_username")

	if !exists {
		// User not authenticated - redirect to portal login
		h.logger.Info("User not authenticated, redirecting to portal login")
		c.Redirect(http.StatusFound, "/auth/github/login")
		return
	}

	// Generate a new session ID (timestamp-based for uniqueness)
	sessionID := time.Now().Unix()

	h.logger.Info("Creating new session for user",
		"user_id", userID,
		"username", username,
		"session_id", sessionID)

	// Redirect to workspace with new session
	c.Redirect(http.StatusFound, fmt.Sprintf("/review/workspace/%d", sessionID))
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
	// Extract authenticated user from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context - authentication middleware may not be configured")
		c.String(http.StatusUnauthorized, `<div class="alert alert-error"><p>Authentication required</p></div>`)
		return
	}

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
	h.logger.Info("Session created",
		"session_id", sessionID,
		"user_id", userID,
		"source", "form")

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
		c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusServiceUnavailable)
		if renderErr := templates.AIServiceUnavailable().Render(c.Request.Context(), c.Writer); renderErr != nil {
			h.logger.Error("Failed to render AI service unavailable template", "error", renderErr.Error())
		}
		return
	}

	// Extract session token from Gin context (set by RedisSessionAuthMiddleware)
	sessionToken, exists := c.Get("session_token")
	if !exists {
		h.logger.Warn("Session token not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session token missing"})
		return
	}
	sessionTokenStr, ok := sessionToken.(string)
	if !ok {
		h.logger.Warn("Session token type assertion failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid session token"})
		return
	}

	// DEBUG: Log session token extraction
	h.logger.Info("DEBUG HandlePreviewMode session token",
		"has_token", sessionToken != nil,
		"token_length", len(sessionTokenStr),
		"token_empty", sessionTokenStr == "")

	// Create context with 90-second timeout for LLM generation (overrides Gin's default ~12s timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Pass both model and session token to service via context
	ctx = context.WithValue(ctx, reviewcontext.ModelContextKey, req.Model)
	ctx = context.WithValue(ctx, reviewcontext.SessionTokenKey, sessionTokenStr)

	// DEBUG: Verify context values
	h.logger.Info("DEBUG HandlePreviewMode context values",
		"model", req.Model,
		"session_token_length", len(sessionTokenStr))

	result, err := h.previewService.AnalyzePreview(ctx, req.PastedCode, req.UserMode, req.OutputMode)
	if err != nil {
		h.logger.Error("Preview analysis failed", "error", err.Error(), "model", req.Model, "user_mode", req.UserMode, "output_mode", req.OutputMode)
		h.renderError(c, err, "Preview analysis failed")
		return
	}

	// Return JSON response for frontend React app
	c.JSON(http.StatusOK, result)
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
		c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusServiceUnavailable)
		if renderErr := templates.AIServiceUnavailable().Render(c.Request.Context(), c.Writer); renderErr != nil {
			h.logger.Error("Failed to render AI service unavailable template", "error", renderErr.Error())
		}
		return
	}

	// Extract session token from Gin context (set by RedisSessionAuthMiddleware)
	sessionToken, exists := c.Get("session_token")
	if !exists {
		h.logger.Warn("Session token not found in context for skim mode")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session token missing"})
		return
	}

	sessionTokenStr, ok := sessionToken.(string)
	if !ok {
		h.logger.Warn("Session token type assertion failed for skim mode")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid session token"})
		return
	}

	// Create context with 90-second timeout for LLM generation (overrides Gin's default ~12s timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Pass both model and session token to service via context
	ctx = context.WithValue(ctx, reviewcontext.ModelContextKey, req.Model)
	ctx = context.WithValue(ctx, reviewcontext.SessionTokenKey, sessionTokenStr)

	// If the pasted input doesn't look like source code, avoid calling Skim mode
	// which expects actual source files (functions, interfaces, data models).
	if !looksLikeCode(req.PastedCode) {
		// Return a friendly message instead of hallucinated abstractions.
		summary := "The content you pasted looks like natural language text, not source code.\n" +
			"Skim mode extracts functions, interfaces and data models from source files. " +
			"If you want to search or summarize prose, use Scan mode or paste source code."
		out := &review_models.SkimModeOutput{Summary: summary}
		c.JSON(http.StatusOK, out)
		return
	}

	result, err := h.skimService.AnalyzeSkim(ctx, req.PastedCode, req.UserMode, req.OutputMode)
	if err != nil {
		h.logger.Error("Skim analysis failed", "error", err.Error(), "model", req.Model, "user_mode", req.UserMode, "output_mode", req.OutputMode)
		h.renderError(c, err, "Skim analysis failed")
		return
	}

	// Return JSON response for frontend React app
	c.JSON(http.StatusOK, result)
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
		c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusServiceUnavailable)
		if renderErr := templates.AIServiceUnavailable().Render(c.Request.Context(), c.Writer); renderErr != nil {
			h.logger.Error("Failed to render AIServiceUnavailable template in scan", "error", renderErr.Error())
		}
		return
	}

	// Extract session token from Gin context (set by RedisSessionAuthMiddleware)
	sessionToken, exists := c.Get("session_token")
	if !exists {
		h.logger.Warn("Session token not found in context for scan mode")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session token missing"})
		return
	}
	sessionTokenStr, ok := sessionToken.(string)
	if !ok {
		h.logger.Warn("Session token type assertion failed in scan mode")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid session token"})
		return
	}

	// Create context with 90-second timeout for LLM generation (overrides Gin's default ~12s timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	// Pass both model and session token to service via context
	ctx = context.WithValue(ctx, reviewcontext.ModelContextKey, req.Model)
	ctx = context.WithValue(ctx, reviewcontext.SessionTokenKey, sessionTokenStr)
	if !looksLikeCode(req.PastedCode) {
		// If user provided a query, run a local text search (case-insensitive).
		if strings.TrimSpace(query) != "" && query != "find issues and improvements" {
			lines := strings.Split(req.PastedCode, "\n")
			matches := make([]review_models.CodeMatch, 0)
			qLower := strings.ToLower(query)
			for i, line := range lines {
				if strings.Contains(strings.ToLower(line), qLower) {
					ctxBefore := ""
					if i-2 >= 0 {
						ctxBefore = lines[i-2] + "\n"
					}
					if i-1 >= 0 {
						ctxBefore += lines[i-1]
					}
					ctxAfter := ""
					if i+1 < len(lines) {
						ctxAfter = "\n" + lines[i+1]
					}
					match := review_models.CodeMatch{
						FilePath:    "pasted_input",
						CodeSnippet: strings.TrimSpace(line),
						Context:     strings.TrimSpace(ctxBefore + "\n" + line + ctxAfter),
						Relevance:   1.0,
						Line:        i + 1,
					}
					matches = append(matches, match)
				}
			}
			out := &review_models.ScanModeOutput{Summary: "Local text search results for pasted prose", Matches: matches}
			c.JSON(http.StatusOK, out)
			return
		}

		// No meaningful query supplied - return a note guiding the user
		summary := "The content you pasted looks like natural language text, not source code.\n" +
			"Scan mode can search code for patterns (SQL, auth, queries). For prose, provide a search query or paste source code."
		out := &review_models.ScanModeOutput{Summary: summary, Matches: nil}
		c.JSON(http.StatusOK, out)
		return
	}

	result, err := h.scanService.AnalyzeScan(ctx, query, req.PastedCode, req.UserMode, req.OutputMode)
	if err != nil {
		h.logger.Error("Scan analysis failed", "error", err.Error(), "model", req.Model, "user_mode", req.UserMode, "output_mode", req.OutputMode)
		h.renderError(c, err, "Scan analysis failed")
		return
	}

	// Return JSON response for frontend React app
	c.JSON(http.StatusOK, result)
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
		c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusServiceUnavailable)
		if renderErr := templates.AIServiceUnavailable().Render(c.Request.Context(), c.Writer); renderErr != nil {
			h.logger.Error("Failed to render AIServiceUnavailable template in detailed", "error", renderErr.Error())
		}
		return
	}

	// Extract session token from Gin context (set by RedisSessionAuthMiddleware)
	sessionToken, exists := c.Get("session_token")
	if !exists {
		h.logger.Warn("Session token not found in context for detailed mode")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session token missing"})
		return
	}

	sessionTokenStr, ok := sessionToken.(string)
	if !ok {
		h.logger.Warn("Session token type assertion failed for detailed mode")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid session token"})
		return
	}

	// Create context with 90-second timeout for LLM generation (overrides Gin's default ~12s timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	// Pass both model and session token to service via context
	ctx = context.WithValue(ctx, reviewcontext.ModelContextKey, req.Model)
	ctx = context.WithValue(ctx, reviewcontext.SessionTokenKey, sessionTokenStr)

	// If the content doesn't look like code, avoid running Detailed Mode (it expects source)
	if !looksLikeCode(req.PastedCode) {
		html := `<div class="space-y-6 p-6 bg-yellow-50 dark:bg-yellow-900 rounded-lg border border-yellow-200 dark:border-yellow-700">
			<div class="flex items-center gap-3 border-b border-yellow-200 dark:border-yellow-700 pb-4">
				<span class="text-3xl">📖</span>
				<div>
					<h3 class="text-xl font-bold text-yellow-900 dark:text-yellow-50">Detailed Analysis</h3>
					<p class="text-sm text-yellow-700 dark:text-yellow-200">Line-by-line explanation</p>
				</div>
			</div>
			<div class="p-4 bg-white dark:bg-yellow-800 rounded-lg border border-yellow-100 dark:border-yellow-700">
				<h4 class="text-sm font-semibold text-yellow-900 dark:text-yellow-50 mb-2">ℹ️ Note</h4>
				<p class="text-sm text-gray-700 dark:text-yellow-200">The content you pasted looks like natural language text, not source code.</p>
				<p class="text-sm text-gray-700 dark:text-yellow-200 mt-2">Detailed mode performs line-by-line code analysis and requires source code. If you want to summarize prose or search for phrases, use <strong>Scan mode</strong> or paste source code.</p>
			</div>
		</div>`
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, html)
		return
	}

	result, err := h.detailedService.AnalyzeDetailed(ctx, req.PastedCode, filename, req.UserMode, req.OutputMode)
	if err != nil {
		h.logger.Error("Detailed analysis failed", "error", err.Error(), "model", req.Model, "user_mode", req.UserMode, "output_mode", req.OutputMode)
		h.renderError(c, err, "Detailed analysis failed")
		return
	}

	c.JSON(http.StatusOK, result)
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
		c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusServiceUnavailable)
		if renderErr := templates.AIServiceUnavailable().Render(c.Request.Context(), c.Writer); renderErr != nil {
			h.logger.Error("Failed to render AIServiceUnavailable template in critical", "error", renderErr.Error())
		}
		return
	}

	// Extract session token from Gin context (set by RedisSessionAuthMiddleware)
	sessionToken, exists := c.Get("session_token")
	if !exists {
		h.logger.Warn("Session token not found in context for critical mode")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session token missing"})
		return
	}

	sessionTokenStr, ok := sessionToken.(string)
	if !ok {
		h.logger.Warn("Session token type assertion failed for critical mode")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid session token"})
		return
	}

	// Create context with 90-second timeout for LLM generation (overrides Gin's default ~12s timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	// Pass both model and session token to service via context
	ctx = context.WithValue(ctx, reviewcontext.ModelContextKey, req.Model)
	ctx = context.WithValue(ctx, reviewcontext.SessionTokenKey, sessionTokenStr)

	// If pasted content doesn't look like source code, avoid running full Critical
	// analysis which focuses on architecture/layering and code quality.
	if !looksLikeCode(req.PastedCode) {
		summary := "The content you pasted appears to be natural language text rather than source code.\n" +
			"Critical mode evaluates code quality, architecture and security. For prose, please use Scan mode or paste source code."
		out := &review_models.CriticalModeOutput{Summary: summary, Issues: nil}
		c.JSON(http.StatusOK, out)
		return
	}

	result, err := h.criticalService.AnalyzeCritical(ctx, req.PastedCode)
	if err != nil {
		h.logger.Error("Critical analysis failed", "error", err.Error(), "model", req.Model)
		h.renderError(c, err, "Critical analysis failed")
		return
	}

	// Normalize overall grade deterministically based on issues to reduce LLM variance
	deterministic := determineGradeFromIssues(result.Issues)
	if deterministic != "" && deterministic != result.OverallGrade {
		// preserve original grade in the summary for traceability
		orig := result.OverallGrade
		if orig == "" {
			result.Summary = fmt.Sprintf("Grade: %s. %s", deterministic, result.Summary)
		} else {
			result.Summary = fmt.Sprintf("Original grade: %s. Normalized grade: %s. %s", orig, deterministic, result.Summary)
		}
		result.OverallGrade = deterministic
	}

	c.JSON(http.StatusOK, result)
}

// determineGradeFromIssues applies a simple deterministic rubric to compute an overall grade
// based on the counts of issues by severity. This prevents non-deterministic LLM grades.
func determineGradeFromIssues(issues []review_models.CodeIssue) string {
	if len(issues) == 0 {
		return "A"
	}
	var crit, high, med int
	for _, it := range issues {
		switch strings.ToLower(it.Severity) {
		case "critical":
			crit++
		case "high":
			high++
		case "medium":
			med++
		}
	}
	// Rubric (conservative): any critical -> F, 2+ high -> D, 1 high or 3+ med -> C, 1-2 med -> B, else A
	switch {
	case crit > 0:
		return "F"
	case high >= 2:
		return "D"
	case high == 1 || med >= 3:
		return "C"
	case med >= 1:
		return "B"
	default:
		return "A"
	}
}

// GetAvailableModels returns a list of available Ollama models (queries dynamically)
func (h *UIHandler) GetAvailableModels(c *gin.Context) {
	ctx := c.Request.Context()

	// Use model service to query Ollama for actual available models
	modelsJSON, err := h.modelService.ListAvailableModelsJSON(ctx)
	if err != nil {
		h.logger.Error("Failed to retrieve available models", "error", err.Error())
		// Return fallback: only Mistral 7B (guaranteed to be available)
		fallbackModels := []map[string]string{
			{"name": "mistral:7b-instruct", "description": "Fast, General (Recommended)"},
		}
		c.JSON(http.StatusOK, gin.H{"models": fallbackModels})
		return
	}

	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, string(modelsJSON))
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

	// Extract authenticated user info from Redis session context
	userID, _ := c.Get("user_id")
	username, _ := c.Get("github_username")

	var sessionID int
	var props templates.WorkspaceProps

	// Handle special "demo" session for quick access (legacy)
	if sessionIDStr == "demo" {
		h.logger.Info("showing demo workspace (legacy)", "user_id", userID)
		props = templates.WorkspaceProps{
			SessionID:      0,
			Title:          fmt.Sprintf("DevSmith Review - Workspace (User: %v)", username),
			Code:           sampleCodeForWorkspace(),
			CurrentMode:    "preview",
			AnalysisResult: "",
		}
	} else {
		// Parse numeric session ID
		var err error
		sessionID, err = strconv.Atoi(sessionIDStr)
		if err != nil {
			h.logger.Warn("invalid session_id", "session_id", sessionIDStr, "error", err)
			c.String(http.StatusBadRequest, "Invalid session ID")
			return
		}

		h.logger.Info("showing workspace", "session_id", sessionID, "user_id", userID, "username", username)

		// For now, use sample data with user context
		// TODO: Fetch actual session data from ReviewRepository by session_id and user_id
		props = templates.WorkspaceProps{
			SessionID:      sessionID,
			Title:          fmt.Sprintf("Code Review Session #%d (User: %v)", sessionID, username),
			Code:           sampleCodeForWorkspace(),
			CurrentMode:    "preview",
			AnalysisResult: "",
		}
	}

	// Render workspace template
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)

	if err := templates.Workspace(props).Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.Error("failed to render workspace template", "error", err.Error())
		h.renderError(c, err, "Failed to render workspace")
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
