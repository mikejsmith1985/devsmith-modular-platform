# Issue #013: [COPILOT] Review Service - UI Integration

**Type:** Feature (Copilot Implementation)
**Service:** Review
**Depends On:** Issues #004-#008 (All 5 reading modes backends), Issue #012 (Portal Dashboard)
**Estimated Duration:** 60-90 minutes

---

## Summary

Create the Review Service web UI that integrates all 5 reading modes (Preview, Skim, Scan, Detailed, Critical) into a single cohesive interface. This UI allows developers to paste repository URLs, select reading modes, and view AI-generated analysis.

**User Story:**
> As a developer, I want to access the Review Service from the Portal dashboard, paste a repository URL, choose a reading mode, and see AI-generated insights about the codebase.

---

## Bounded Context

**Review Service Context:**
- **Responsibility:** Code review analysis with 5 reading modes
- **Does NOT:** Handle authentication (Portal does that), log aggregation (Logs does that), or analytics (Analytics does that)
- **Boundaries:** Review only processes code and generates insights

**Why This Matters:**
- UI lives in Review service (not Portal)
- Authentication handled by Portal (JWT passed to Review)
- Review service is accessed via `http://localhost:8081` from Portal dashboard

---

## Success Criteria

### Must Have (MVP)
- [ ] Landing page at `/` shows mode selector and repository input
- [ ] 5 reading modes selectable: Preview, Skim, Scan, Detailed, Critical
- [ ] Repository URL input with validation (GitHub URLs only)
- [ ] Branch/commit input (defaults to `main`)
- [ ] Submit button triggers analysis
- [ ] Loading state during analysis (with progress indicator)
- [ ] Results page displays AI-generated analysis
- [ ] Each mode shows different UI layout based on mode type
- [ ] Markdown rendering for AI responses
- [ ] Copy analysis to clipboard functionality
- [ ] Back to mode selector button
- [ ] Responsive design (desktop and tablet)

### Nice to Have (Post-MVP)
- Save analysis history
- Export analysis as PDF
- Compare multiple analyses
- Real-time streaming of analysis progress

---

## Database Schema

**Uses existing schemas from Issues #004-#008.**

No new tables needed. UI consumes existing API endpoints.

---

## API Endpoints (Existing - Created in Issues #004-#008)

### POST `/api/v1/review/preview`
From Issue #004 (Preview Mode)

### POST `/api/v1/review/skim`
From Issue #005 (Skim Mode)

### POST `/api/v1/review/scan`
From Issue #006 (Scan Mode)

### POST `/api/v1/review/detailed`
From Issue #007 (Detailed Mode)

### POST `/api/v1/review/critical`
From Issue #008 (Critical Mode)

**All endpoints accept:**
```json
{
  "repository_url": "https://github.com/user/repo",
  "branch": "main",
  "commit_sha": "abc123"
}
```

**All endpoints return:**
```json
{
  "analysis_id": "uuid",
  "mode": "preview|skim|scan|detailed|critical",
  "analysis": "AI-generated markdown content",
  "repository": "user/repo",
  "branch": "main",
  "created_at": "2025-10-20T10:30:00Z"
}
```

---

## File Structure

```
apps/review/
├── templates/
│   ├── layout.templ              # NEW - Base layout
│   ├── home.templ                # NEW - Mode selector & repo input
│   ├── analysis.templ            # NEW - Analysis results display
│   └── components/
│       ├── mode_selector.templ   # NEW - 5 reading modes selector
│       ├── repo_input.templ      # NEW - Repository URL input form
│       └── loading.templ         # NEW - Loading state component
├── handlers/
│   ├── ui_handler.go             # NEW - UI route handlers
│   └── [existing mode handlers]  # From Issues #004-#008
└── static/
    ├── css/
    │   ├── review.css            # NEW - Main styles
    │   └── modes.css             # NEW - Mode-specific styles
    └── js/
        ├── review.js             # NEW - Main UI logic
        └── analysis.js           # NEW - Analysis rendering logic

cmd/review/
└── main.go                       # UPDATE - Add UI routes
```

---

## Implementation Details

### 1. Home Page Template (Mode Selector)

**File:** `apps/review/templates/home.templ`

```go
package templates

templ Home() {
	@Layout("DevSmith Review - Code Analysis") {
		<div class="review-container">
			<header class="review-header">
				<h1>🔍 DevSmith Review</h1>
				<p class="subtitle">AI-powered code analysis with 5 reading modes</p>
			</header>

			<main class="review-main">
				@ModeSelector()
				@RepoInput()
			</main>
		</div>
	}
}

templ ModeSelector() {
	<div class="mode-selector">
		<h2>Choose Your Reading Mode</h2>
		<div class="modes-grid">
			@ModeCard(ModeInfo{
				ID: "preview",
				Name: "Preview",
				Icon: "👁️",
				Description: "Quick 2-minute overview of project structure and purpose",
				Duration: "2-3 min",
				Cognitive: "Low",
			})

			@ModeCard(ModeInfo{
				ID: "skim",
				Name: "Skim",
				Icon: "⚡",
				Description: "Surface-level scan of architecture and key components",
				Duration: "5-7 min",
				Cognitive: "Low-Medium",
			})

			@ModeCard(ModeInfo{
				ID: "scan",
				Name: "Scan",
				Icon: "🔎",
				Description: "Targeted search for specific patterns or issues",
				Duration: "3-5 min",
				Cognitive: "Medium",
			})

			@ModeCard(ModeInfo{
				ID: "detailed",
				Name: "Detailed",
				Icon: "📖",
				Description: "Deep dive into implementation details and logic",
				Duration: "10-15 min",
				Cognitive: "High",
			})

			@ModeCard(ModeInfo{
				ID: "critical",
				Name: "Critical",
				Icon: "🔬",
				Description: "Comprehensive review focusing on correctness and quality",
				Duration: "15-20 min",
				Cognitive: "Very High",
			})
		</div>
	</div>
}

templ ModeCard(mode ModeInfo) {
	<div class="mode-card" data-mode={mode.ID}>
		<div class="mode-icon">{mode.Icon}</div>
		<h3>{mode.Name}</h3>
		<p class="mode-description">{mode.Description}</p>
		<div class="mode-meta">
			<span class="mode-duration">⏱️ {mode.Duration}</span>
			<span class={"cognitive-load " + mode.Cognitive}>🧠 {mode.Cognitive}</span>
		</div>
		<button class="btn-select-mode" data-mode={mode.ID}>Select {mode.Name}</button>
	</div>
}

templ RepoInput() {
	<div id="repo-input-section" class="repo-input-section hidden">
		<h2>Repository Details</h2>
		<form id="review-form">
			<div class="form-group">
				<label for="repository-url">Repository URL</label>
				<input
					type="url"
					id="repository-url"
					name="repository_url"
					placeholder="https://github.com/user/repo"
					required
				/>
				<span class="form-help">Only GitHub repositories are supported in MVP</span>
			</div>

			<div class="form-row">
				<div class="form-group">
					<label for="branch">Branch</label>
					<input
						type="text"
						id="branch"
						name="branch"
						placeholder="main"
						value="main"
					/>
				</div>

				<div class="form-group">
					<label for="commit-sha">Commit SHA (optional)</label>
					<input
						type="text"
						id="commit-sha"
						name="commit_sha"
						placeholder="abc123..."
					/>
				</div>
			</div>

			<input type="hidden" id="selected-mode" name="mode" />

			<div class="form-actions">
				<button type="button" id="back-btn" class="btn-secondary">← Back to Modes</button>
				<button type="submit" id="analyze-btn" class="btn-primary">Start Analysis</button>
			</div>
		</form>
	</div>
}

type ModeInfo struct {
	ID          string
	Name        string
	Icon        string
	Description string
	Duration    string
	Cognitive   string
}
```

---

### 2. Analysis Results Template

**File:** `apps/review/templates/analysis.templ`

```go
package templates

templ Analysis(result AnalysisResult) {
	@Layout("Analysis Results - " + result.Mode) {
		<div class="analysis-container">
			<header class="analysis-header">
				<div class="analysis-meta">
					<span class="mode-badge">{result.Mode}</span>
					<h1>{result.Repository}</h1>
					<span class="branch-info">Branch: {result.Branch}</span>
				</div>
				<div class="analysis-actions">
					<button id="copy-btn" class="btn-icon" title="Copy to clipboard">📋</button>
					<a href="/" class="btn-secondary">← New Analysis</a>
				</div>
			</header>

			<main class="analysis-main">
				<div id="analysis-content" class="markdown-content">
					{templ.Raw(result.AnalysisHTML)}
				</div>
			</main>

			<footer class="analysis-footer">
				<span class="timestamp">Generated: {result.CreatedAt}</span>
				<span class="analysis-id">ID: {result.AnalysisID}</span>
			</footer>
		</div>
	}
}

type AnalysisResult struct {
	AnalysisID   string
	Mode         string
	Repository   string
	Branch       string
	AnalysisHTML string
	CreatedAt    string
}
```

---

### 3. Loading Component

**File:** `apps/review/templates/components/loading.templ`

```go
package components

templ Loading(mode string) {
	<div id="loading-overlay" class="loading-overlay">
		<div class="loading-content">
			<div class="spinner"></div>
			<h2>Analyzing Repository...</h2>
			<p>Running <strong>{mode}</strong> mode analysis</p>
			<p class="loading-message">This may take 30 seconds to 2 minutes depending on repository size.</p>
		</div>
	</div>
}
```

---

### 4. UI Handler

**File:** `apps/review/handlers/ui_handler.go`

```go
package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"devsmith/apps/review/templates"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// HomeHandler serves the main Review UI (mode selector + repo input)
func HomeHandler(c *gin.Context) {
	component := templates.Home()
	component.Render(c.Request.Context(), c.Writer)
}

// AnalysisResultHandler displays analysis results
func AnalysisResultHandler(c *gin.Context) {
	// Get analysis result from query params or session
	// In MVP, we'll receive this from the POST redirect

	mode := c.Query("mode")
	repo := c.Query("repo")
	branch := c.Query("branch")
	analysisMarkdown := c.Query("analysis") // From API response

	// Convert Markdown to HTML
	htmlContent := markdownToHTML(analysisMarkdown)

	result := templates.AnalysisResult{
		AnalysisID:   generateAnalysisID(),
		Mode:         mode,
		Repository:   repo,
		Branch:       branch,
		AnalysisHTML: htmlContent,
		CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
	}

	component := templates.Analysis(result)
	component.Render(c.Request.Context(), c.Writer)
}

func markdownToHTML(md string) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	htmlBytes := markdown.ToHTML([]byte(md), parser, renderer)
	return string(htmlBytes)
}
```

---

### 5. Main JavaScript

**File:** `apps/review/static/js/review.js`

```javascript
let selectedMode = null;

// Mode selection
document.querySelectorAll('.btn-select-mode').forEach(btn => {
  btn.addEventListener('click', (e) => {
    selectedMode = e.target.dataset.mode;
    document.getElementById('selected-mode').value = selectedMode;

    // Hide mode selector, show repo input
    document.querySelector('.mode-selector').classList.add('hidden');
    document.getElementById('repo-input-section').classList.remove('hidden');

    // Update form title with selected mode
    document.querySelector('#repo-input-section h2').textContent =
      `Repository Details (${capitalizeMode(selectedMode)} Mode)`;
  });
});

// Back button
document.getElementById('back-btn').addEventListener('click', () => {
  document.querySelector('.mode-selector').classList.remove('hidden');
  document.getElementById('repo-input-section').classList.add('hidden');
  selectedMode = null;
});

// Form submission
document.getElementById('review-form').addEventListener('submit', async (e) => {
  e.preventDefault();

  const formData = {
    repository_url: document.getElementById('repository-url').value,
    branch: document.getElementById('branch').value || 'main',
    commit_sha: document.getElementById('commit-sha').value || null,
  };

  // Validate GitHub URL
  if (!formData.repository_url.startsWith('https://github.com/')) {
    alert('Please enter a valid GitHub repository URL');
    return;
  }

  // Show loading overlay
  showLoading(selectedMode);

  try {
    // Call appropriate API endpoint based on selected mode
    const response = await fetch(`/api/v1/review/${selectedMode}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(formData),
    });

    if (!response.ok) {
      throw new Error(`Analysis failed: ${response.statusText}`);
    }

    const result = await response.json();

    // Redirect to analysis results page with data
    const params = new URLSearchParams({
      mode: result.mode,
      repo: result.repository,
      branch: result.branch,
      analysis: result.analysis,
    });

    window.location.href = `/analysis?${params.toString()}`;

  } catch (error) {
    hideLoading();
    alert(`Error: ${error.message}`);
  }
});

function showLoading(mode) {
  const overlay = document.createElement('div');
  overlay.id = 'loading-overlay';
  overlay.className = 'loading-overlay';
  overlay.innerHTML = `
    <div class="loading-content">
      <div class="spinner"></div>
      <h2>Analyzing Repository...</h2>
      <p>Running <strong>${capitalizeMode(mode)}</strong> mode analysis</p>
      <p class="loading-message">This may take 30 seconds to 2 minutes.</p>
    </div>
  `;
  document.body.appendChild(overlay);
}

function hideLoading() {
  const overlay = document.getElementById('loading-overlay');
  if (overlay) overlay.remove();
}

function capitalizeMode(mode) {
  return mode.charAt(0).toUpperCase() + mode.slice(1);
}
```

---

### 6. Copy to Clipboard Functionality

**File:** `apps/review/static/js/analysis.js`

```javascript
// Copy analysis to clipboard
document.getElementById('copy-btn')?.addEventListener('click', async () => {
  const content = document.getElementById('analysis-content').innerText;

  try {
    await navigator.clipboard.writeText(content);

    // Visual feedback
    const btn = document.getElementById('copy-btn');
    const originalText = btn.textContent;
    btn.textContent = '✅';
    btn.classList.add('copied');

    setTimeout(() => {
      btn.textContent = originalText;
      btn.classList.remove('copied');
    }, 2000);

  } catch (err) {
    alert('Failed to copy to clipboard');
  }
});
```

---

### 7. CSS Styles

**File:** `apps/review/static/css/review.css`

```css
:root {
  --primary-color: #0366d6;
  --secondary-color: #6c757d;
  --success-color: #28a745;
  --danger-color: #dc3545;
  --bg-light: #f6f8fa;
  --border-color: #e1e4e8;
  --text-primary: #24292e;
  --text-secondary: #586069;
}

.review-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

.review-header {
  text-align: center;
  margin-bottom: 3rem;
}

.review-header h1 {
  font-size: 2.5rem;
  margin-bottom: 0.5rem;
  color: var(--text-primary);
}

.subtitle {
  color: var(--text-secondary);
  font-size: 1.125rem;
}

/* Mode Selector */
.mode-selector {
  margin-bottom: 3rem;
}

.mode-selector h2 {
  text-align: center;
  margin-bottom: 2rem;
  color: var(--text-primary);
}

.modes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1.5rem;
}

.mode-card {
  background: white;
  border: 2px solid var(--border-color);
  border-radius: 8px;
  padding: 1.5rem;
  text-align: center;
  transition: all 0.2s;
  cursor: pointer;
}

.mode-card:hover {
  border-color: var(--primary-color);
  transform: translateY(-4px);
  box-shadow: 0 8px 16px rgba(0,0,0,0.1);
}

.mode-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.mode-card h3 {
  margin: 0 0 0.5rem 0;
  color: var(--text-primary);
}

.mode-description {
  color: var(--text-secondary);
  font-size: 0.875rem;
  margin-bottom: 1rem;
  min-height: 3rem;
}

.mode-meta {
  display: flex;
  justify-content: space-around;
  margin-bottom: 1rem;
  font-size: 0.75rem;
}

.cognitive-load {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-weight: 600;
}

.cognitive-load.Low { background: #d4edda; color: #155724; }
.cognitive-load.Low-Medium { background: #d1ecf1; color: #0c5460; }
.cognitive-load.Medium { background: #fff3cd; color: #856404; }
.cognitive-load.High { background: #f8d7da; color: #721c24; }
.cognitive-load.Very\\ High { background: #dc3545; color: white; }

.btn-select-mode {
  width: 100%;
  padding: 0.75rem;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-select-mode:hover {
  background: #0256c7;
}

/* Repository Input */
.repo-input-section {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 2rem;
  max-width: 600px;
  margin: 0 auto;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: var(--text-primary);
}

.form-group input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 1rem;
}

.form-help {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
}

.btn-primary {
  flex: 1;
  padding: 0.75rem 1.5rem;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}

.btn-secondary {
  padding: 0.75rem 1.5rem;
  background: var(--secondary-color);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  text-decoration: none;
}

/* Loading Overlay */
.loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.loading-content {
  text-align: center;
  color: white;
}

.spinner {
  width: 64px;
  height: 64px;
  border: 4px solid rgba(255,255,255,0.2);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Analysis Results */
.analysis-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 2rem;
}

.analysis-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 2px solid var(--border-color);
}

.mode-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  background: var(--primary-color);
  color: white;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  margin-bottom: 0.5rem;
}

.markdown-content {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  line-height: 1.6;
}

.markdown-content h1 { margin-top: 0; }
.markdown-content pre { background: #f6f8fa; padding: 1rem; border-radius: 6px; overflow-x: auto; }
.markdown-content code { background: #f6f8fa; padding: 0.2rem 0.4rem; border-radius: 3px; }

.hidden { display: none !important; }

@media (max-width: 768px) {
  .modes-grid { grid-template-columns: 1fr; }
  .form-row { grid-template-columns: 1fr; }
  .analysis-header { flex-direction: column; gap: 1rem; }
}
```

---

## TDD Workflow

### TDD Workflow for This Issue

**Step 1: RED PHASE (Write Failing Tests) - DO THIS FIRST!**

Create test files BEFORE implementation:

```go
// apps/review/handlers/ui_handler_test.go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestReviewPageHandler_RendersForm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", ReviewPageHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "DevSmith Code Review")
	assert.Contains(t, w.Body.String(), "repository-url")
	assert.Contains(t, w.Body.String(), "reading-mode")
	assert.Contains(t, w.Body.String(), "analyze-btn")
}

func TestReviewPageHandler_ShowsAllFiveReadingModes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", ReviewPageHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should show all 5 reading modes
	assert.Contains(t, w.Body.String(), "preview")
	assert.Contains(t, w.Body.String(), "skim")
	assert.Contains(t, w.Body.String(), "scan")
	assert.Contains(t, w.Body.String(), "detailed")
	assert.Contains(t, w.Body.String(), "critical")
}

func TestAnalysisResultsHandler_DisplaysMarkdownOutput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/analysis/:id", AnalysisResultsHandler)

	req := httptest.NewRequest(http.MethodGet, "/analysis/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "analysis-output")
	assert.Contains(t, w.Body.String(), "markdown-content")
}

func TestStartAnalysisHandler_ValidatesInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/review/start", StartAnalysisHandler)

	// Test with missing repository URL
	req := httptest.NewRequest(http.MethodPost, "/api/v1/review/start", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStartAnalysisHandler_CreatesAnalysis(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock service layer
	mockService := &MockReviewService{
		CreateAnalysisFn: func(ctx context.Context, req AnalysisRequest) (*Analysis, error) {
			return &Analysis{
				ID:          "test-123",
				RepoURL:     req.RepoURL,
				ReadingMode: req.ReadingMode,
				Status:      "pending",
			}, nil
		},
	}

	handler := NewReviewHandler(mockService)
	router.POST("/api/v1/review/start", handler.StartAnalysis)

	body := `{"repository_url": "https://github.com/user/repo", "reading_mode": "skim"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/review/start", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test-123")
}

// apps/review/templates/review_page_test.go (Templ testing)
package templates

import (
	"context"
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestReviewPageTemplate_RendersFormElements(t *testing.T) {
	var buf strings.Builder
	err := ReviewPage().Render(context.Background(), &buf)

	assert.NoError(t, err)
	html := buf.String()
	assert.Contains(t, html, "repository-url")
	assert.Contains(t, html, "reading-mode")
	assert.Contains(t, html, "analyze-btn")
	assert.Contains(t, html, "form")
}

func TestReadingModeSelector_ShowsAllModes(t *testing.T) {
	var buf strings.Builder
	err := ReadingModeSelector().Render(context.Background(), &buf)

	assert.NoError(t, err)
	html := buf.String()

	modes := []string{"preview", "skim", "scan", "detailed", "critical"}
	for _, mode := range modes {
		assert.Contains(t, html, mode)
	}
}

func TestReadingModeSelector_ShowsDescriptions(t *testing.T) {
	var buf strings.Builder
	err := ReadingModeSelector().Render(context.Background(), &buf)

	assert.NoError(t, err)
	html := buf.String()

	// Each mode should have a description
	assert.Contains(t, html, "Quick overview")       // Preview
	assert.Contains(t, html, "Fast scan")            // Skim
	assert.Contains(t, html, "Pattern detection")    // Scan
	assert.Contains(t, html, "Comprehensive")        // Detailed
	assert.Contains(t, html, "Production-ready")     // Critical
}

func TestAnalysisResultsTemplate_RendersMarkdown(t *testing.T) {
	analysis := &Analysis{
		ID:          "123",
		RepoURL:     "https://github.com/user/repo",
		ReadingMode: "skim",
		Output:      "# Analysis Results\n\n## Overview\n\nThis is markdown.",
		Status:      "completed",
	}

	var buf strings.Builder
	err := AnalysisResults(analysis).Render(context.Background(), &buf)

	assert.NoError(t, err)
	html := buf.String()
	assert.Contains(t, html, "Analysis Results")
	assert.Contains(t, html, "markdown-content")
	assert.Contains(t, html, analysis.ID)
}
```

**Run tests (should FAIL):**
```bash
go test ./apps/review/handlers/...
# Expected: FAIL - ReviewPageHandler undefined

go test ./apps/review/templates/...
# Expected: FAIL - ReviewPage template undefined
```

**Commit failing tests:**
```bash
git add apps/review/handlers/ui_handler_test.go
git add apps/review/templates/review_page_test.go
git commit -m "test(review): add failing tests for Review UI integration (RED phase)"
```

**Step 2: GREEN PHASE - Implement to Pass Tests**

Now implement the templates, handlers, and JavaScript. See Implementation section above.

**After implementation, run tests:**
```bash
go test ./apps/review/...
# Expected: PASS
```

**Step 3: Verify Build**
```bash
templ generate apps/review/templates/*.templ
go build -o /dev/null ./cmd/review
```

**Step 4: Manual Testing**

Follow the manual testing checklist below.

**Step 5: Commit Implementation**
```bash
git add apps/review/
git commit -m "feat(review): implement Review UI with 5 reading modes and analysis display (GREEN phase)"
```

**Step 6: REFACTOR PHASE (Optional)**

If needed, refactor for:
- Better form validation (client-side and server-side)
- Improved markdown rendering (syntax highlighting, code blocks)
- Real-time analysis progress updates (WebSocket or polling)
- Better error handling and user feedback
- Accessibility improvements (ARIA labels, keyboard navigation)

**Commit refactors:**
```bash
git add apps/review/
git commit -m "refactor(review): improve form validation and markdown rendering"
```

**Reference:** DevsmithTDD.md lines 15-36, 38-86 (RED-GREEN-REFACTOR)

**Key TDD Principles for Review UI:**
1. **Test form rendering** (all input fields present)
2. **Test reading mode selector** (all 5 modes shown with descriptions)
3. **Test form submission** (validation, API call, response handling)
4. **Test analysis display** (markdown rendered, metadata shown)
5. **Test error states** (invalid input, failed analysis)
6. **Test navigation** (start analysis → results page → back to form)

**Coverage Target:** 70%+ for Go handlers, 60%+ for Templ templates, 60%+ for JavaScript

**Special Testing Considerations:**
- Mock Review service layer in handler tests
- Test Templ components by rendering to string buffer
- Test markdown rendering with various markdown content
- Test form validation with edge cases (empty URL, invalid mode)
- Test HTMX behavior (form submission, dynamic content loading)

**Integration with Backend:**
- Review UI calls existing API endpoints from Issues #004-#008
- Each reading mode uses different prompt strategy
- Analysis results are stored in `review.analyses` table
- Markdown output is rendered using Go markdown library

---

## Testing Requirements

### Unit Tests

**File:** `apps/review/handlers/ui_handler_test.go`

```go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHomeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", HomeHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "DevSmith Review")
	assert.Contains(t, w.Body.String(), "Preview")
	assert.Contains(t, w.Body.String(), "Critical")
}

func TestMarkdownToHTML(t *testing.T) {
	markdown := "# Heading\n\nParagraph with **bold** text."
	html := markdownToHTML(markdown)

	assert.Contains(t, html, "<h1>Heading</h1>")
	assert.Contains(t, html, "<strong>bold</strong>")
}
```

### Manual Testing Checklist

- [ ] Navigate to `http://localhost:8081/`
- [ ] Verify all 5 mode cards display correctly
- [ ] Click "Select Preview" - verify repo input form appears
- [ ] Click "Back to Modes" - verify mode selector reappears
- [ ] Select "Critical Mode"
- [ ] Enter repository URL: `https://github.com/gin-gonic/gin`
- [ ] Leave branch as "main"
- [ ] Click "Start Analysis"
- [ ] Verify loading overlay appears
- [ ] Wait for analysis to complete
- [ ] Verify analysis results page loads with markdown rendered
- [ ] Click "Copy to clipboard" - verify ✅ feedback
- [ ] Click "New Analysis" - verify returns to home
- [ ] Test all 5 modes with different repositories
- [ ] Test responsive design on mobile viewport

---

## Configuration

**File:** `.env` (Review service)

```bash
# Existing from Issues #004-#008
DATABASE_URL=postgresql://review_user:review_pass@localhost:5432/devsmith_review
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=deepseek-coder-v2:16b

# GitHub API (for repository fetching)
GITHUB_TOKEN=your_personal_access_token
```

---

## Acceptance Criteria

Before marking this issue complete, verify:

- [x] Home page loads at `http://localhost:8081/`
- [x] All 5 reading modes display with correct icons and descriptions
- [x] Mode selection shows repository input form
- [x] Form validates GitHub URLs
- [x] All 5 modes can be triggered via UI
- [x] Loading overlay shows during analysis
- [x] Analysis results render as markdown HTML
- [x] Copy to clipboard works
- [x] Back button returns to mode selector
- [x] Responsive design works on desktop and tablet
- [x] No console errors
- [x] Unit tests pass (70%+ coverage)
- [x] Manual testing checklist complete

---

## Branch Naming

```bash
feature/013-review-ui-integration
```

---

## Notes

- This UI integrates all 5 backend modes (Issues #004-#008)
- Each mode uses its existing API endpoint - no backend changes needed
- Markdown rendering uses `github.com/gomarkdown/markdown` library
- Loading times vary by mode: Preview (30s), Skim (1m), Critical (2m)
- For MVP, analysis results are not persisted (future enhancement)
- Service URLs are configured in Portal dashboard (Issue #012)

---

**Created:** 2025-10-20
**For:** Copilot Implementation
**Estimated Time:** 60-90 minutes
