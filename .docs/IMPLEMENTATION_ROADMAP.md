# DevSmith Platform: Implementation Roadmap

**Version:** 1.0  
**Created:** 2025-11-04  
**Status:** In Progress  
**Current Phase:** Phase 0 (Foundation - Model Service Refactor)

---

## Overview

This document tracks the 5-phase implementation plan for DevSmith Platform's major features. Each phase is designed to be completed in a separate conversation with clear acceptance criteria and no inter-phase dependencies requiring context sharing.

### Feature Requests Summary
1. GitHub integration (repo/folder/multi-file scanning in Review app)
2. PR review workflow (files in left pane, AI explanations in right pane)
3. Text-to-Speech with Google Voice (highlight sync)
4. Test output delivery (new app or Review integration)
5. Logs application with AI-driven diagnostics (actionable error resolution)

### Strategic Rationale
**Priority is based on:**
- **Technical Dependencies:** Logs app provides debugging for GitHub integration; GitHub integration enables PR review
- **Business Value:** AI diagnostics and PR review deliver immediate ROI
- **Cognitive Load Management:** Each phase builds on stable foundation before adding complexity
- **Team Efficiency:** Phases can be parallelized if team > 1 developer

---

## Phase 0: Foundation - Model Service Refactor ✅

**Duration:** Completed 2025-11-04  
**Branch:** `development`  
**Goal:** Migrate model discovery from CLI (non-functional in Docker) to Ollama HTTP API

### Acceptance Criteria
- [x] ModelService uses HTTP API (`GET /api/tags`) instead of `exec.Command("ollama", "list")`
- [x] Container accesses host Ollama via `http://host.docker.internal:11434`
- [x] API endpoint `/api/review/models` returns all installed models (not just fallback)
- [x] Frontend JavaScript dynamically populates model dropdown from API response
- [x] HTML fallback shows Mistral if API unreachable
- [x] All code compiles without errors
- [x] Pre-push hook passes (format, imports, build, lint, vet)

### Technical Changes
**Files Modified:**
- `internal/review/services/model_service.go` - Refactored ListAvailableModels() to use HTTP API
- `cmd/review/main.go` - Pass `ollamaEndpoint` to ModelService constructor
- `apps/review/handlers/ui_handler.go` - Updated fallback to single model (Mistral)

**Testing Evidence:**
```bash
# Before: Only Mistral (fallback) returned
curl http://localhost:3000/api/review/models
# {"models":[{"name":"mistral:7b-instruct","description":"Fast, General (Recommended)"}]}

# After: All 3 installed models returned
curl http://localhost:3000/api/review/models
# {"models":[
#   {"name":"qwen2.5-coder:7b-instruct-q4_K_M","description":"Qwen coder model"},
#   {"name":"qwen2.5-coder:7b-instruct-q5_K_M","description":"Qwen coder model"},
#   {"name":"mistral:7b-instruct","description":"Fast, General (Recommended)"}
# ]}
```

### Status
**COMPLETED** - Ready for PR and merge to `development`

---

## Phase 1: Logs Application - AI-Driven Diagnostics ✅

**Duration:** Completed  
**Branch:** `development`  
**Goal:** Build observability layer with AI-powered error analysis and actionable suggestions

**Status:** COMPLETED 2025-11-04 (#103, #104)

### Business Value
- **Developer Productivity:** AI suggests fixes for errors/warnings, reducing debugging time by 50%+
- **Platform Reliability:** Real-time issue detection before users encounter problems
- **Foundation for Later Phases:** Debugging infrastructure for GitHub API integration (Phase 2)

### Acceptance Criteria
- [x] AI Analysis Service implemented (`internal/logs/services/ai_analyzer.go`)
  - [x] Endpoint: `POST /api/logs/analyze` accepts log entries + context
  - [x] Groups logs by `correlation_id` for request tracing
  - [x] Sends to Ollama with prompt: "Analyze these logs. Identify root cause, suggest fix, rate severity"
  - [x] Returns JSON: `{root_cause, suggested_fix, severity, related_logs}`
  - [x] Caches analysis to avoid re-analyzing identical patterns
- [x] Pattern Recognition Service (`internal/logs/services/pattern_matcher.go`)
  - [x] Detects recurring error patterns (e.g., "connection refused", "auth failure")
  - [x] Auto-tags logs with `issue_type` enum
  - [x] Triggers AI analysis on first occurrence, reuses for duplicates
- [x] UI Enhancements (`apps/logs/templates/dashboard.templ`)
  - [x] Issue Cards display:
    - Root cause summary
    - AI-suggested fix (code snippet)
    - "Apply Fix" button (copies to clipboard)
    - Severity indicator (1-5 scale)
  - [x] Filter by: service, severity, issue_type, time range
  - [x] Real-time updates via WebSocket
- [ ] Real-time Alerting
  - [ ] On new ERROR/WARN: trigger AI analysis immediately
  - [ ] Push notification to dashboard: "New issue detected + AI suggestion"
  - [ ] WebSocket message format: `{type: "new_issue", issue: {...}}`
- [ ] Database Schema
  - [ ] `logs.entries` table additions:
    - `issue_type VARCHAR(50)` - categorized error type
    - `ai_analysis JSONB` - cached AI response
    - `severity_score INT` - 1-5 rating
  - [ ] Index: `idx_logs_issue_type` on `(issue_type, created_at DESC)`
- [ ] Testing
  - [ ] Unit tests: AI prompt generation, pattern matching logic
  - [ ] Integration tests: Full flow (log ingestion → AI analysis → UI display)
  - [ ] Load test: 1000 logs/second ingestion without analysis queue backup

### Technical Specifications

#### 1. AI Analyzer Service Structure
```go
// internal/logs/services/ai_analyzer.go
type AIAnalyzer struct {
    ollamaClient *ai.OllamaClient
    cache        *AnalysisCache  // Redis or in-memory map
    logger       *Logger
}

type AnalysisRequest struct {
    LogEntries    []models.LogEntry `json:"log_entries"`
    Context       string            `json:"context"` // "error", "warning", "info"
}

type AnalysisResult struct {
    RootCause     string   `json:"root_cause"`
    SuggestedFix  string   `json:"suggested_fix"`
    Severity      int      `json:"severity"`       // 1-5
    RelatedLogs   []string `json:"related_logs"`   // correlation_ids
    FixSteps      []string `json:"fix_steps"`      // Step-by-step instructions
}

func (a *AIAnalyzer) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error)
```

#### 2. Pattern Matcher Logic
```go
// internal/logs/services/pattern_matcher.go
var ErrorPatterns = map[string]*regexp.Regexp{
    "db_connection":    regexp.MustCompile(`(connection refused|database.*timeout|pg.*connect)`),
    "auth_failure":     regexp.MustCompile(`(unauthorized|authentication.*failed|invalid.*token)`),
    "null_pointer":     regexp.MustCompile(`(nil pointer|null reference|undefined.*nil)`),
    "rate_limit":       regexp.MustCompile(`(rate limit|too many requests|429)`),
    "network_timeout":  regexp.MustCompile(`(timeout|i/o timeout|context deadline)`),
}

func (p *PatternMatcher) Classify(logMsg string) string {
    for issueType, pattern := range ErrorPatterns {
        if pattern.MatchString(logMsg) {
            return issueType
        }
    }
    return "unknown"
}
```

#### 3. Ollama Prompt Template
```go
const LogAnalysisPrompt = `You are a systems diagnostics expert analyzing application logs.

Context: {{.Context}}
Log Entries:
{{range .LogEntries}}
[{{.Level}}] {{.Service}} - {{.Message}}
{{if .Metadata}}Metadata: {{.Metadata}}{{end}}
{{end}}

Tasks:
1. Identify the root cause (be specific - which component/function/line is failing?)
2. Suggest a fix (concrete code change or configuration adjustment)
3. Rate severity (1=info, 2=minor, 3=moderate, 4=serious, 5=critical)
4. List related log correlation IDs if this is part of a larger issue

Respond in JSON format:
{
    "root_cause": "...",
    "suggested_fix": "...",
    "severity": 3,
    "related_logs": ["correlation-id-1", "correlation-id-2"],
    "fix_steps": ["Step 1", "Step 2"]
}
`
```

#### 4. UI Component Structure
```templ
// apps/logs/templates/components/issue_card.templ
package templates

templ IssueCard(issue models.Issue) {
    <div class="issue-card severity-{issue.Severity}" data-issue-id={issue.ID}>
        <div class="issue-header">
            <span class="issue-type">{issue.IssueType}</span>
            <span class="severity-badge">Severity: {issue.Severity}/5</span>
        </div>
        
        <div class="issue-body">
            <h3>Root Cause</h3>
            <p class="root-cause">{issue.AIAnalysis.RootCause}</p>
            
            <h4>Suggested Fix</h4>
            <pre class="suggested-fix"><code>{issue.AIAnalysis.SuggestedFix}</code></pre>
            
            <div class="fix-steps">
                <h4>Steps to Fix</h4>
                <ol>
                    for _, step := range issue.AIAnalysis.FixSteps {
                        <li>{step}</li>
                    }
                </ol>
            </div>
        </div>
        
        <div class="issue-actions">
            <button class="btn-copy" 
                    hx-post="/logs/copy-fix"
                    hx-vals={`{"issue_id": "` + issue.ID + `"}`}>
                📋 Copy Fix
            </button>
            <button class="btn-dismiss"
                    hx-post="/logs/dismiss-issue"
                    hx-vals={`{"issue_id": "` + issue.ID + `"}`}
                    hx-target="closest .issue-card"
                    hx-swap="outerHTML">
                ✓ Dismiss
            </button>
        </div>
    </div>
}
```

#### 5. Real-time WebSocket Integration
```go
// apps/logs/handlers/ws_handler.go
func (h *WSHandler) HandleLogStream(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()

    // Subscribe to Redis pub/sub for new logs
    subscriber := h.redis.Subscribe("logs:new")
    defer subscriber.Close()

    for msg := range subscriber.Channel() {
        var logEntry models.LogEntry
        json.Unmarshal([]byte(msg.Payload), &logEntry)

        // Trigger AI analysis if ERROR or WARN
        if logEntry.Level == "error" || logEntry.Level == "warn" {
            analysis, err := h.aiAnalyzer.Analyze(context.Background(), AnalysisRequest{
                LogEntries: []models.LogEntry{logEntry},
                Context:    logEntry.Level,
            })
            if err == nil {
                logEntry.AIAnalysis = analysis
            }
        }

        // Send to client
        conn.WriteJSON(logEntry)
    }
}
```

### Dependencies
- Ollama running on host (`http://host.docker.internal:11434`)
- PostgreSQL with `logs` schema
- Redis for WebSocket pub/sub (optional - can use in-memory)
- Existing Logs service structure (`apps/logs/`, `internal/logs/`)

### Success Metrics
- AI analysis accuracy: >80% of suggestions actionable
- Response time: <3 seconds for AI analysis per log entry
- Cache hit rate: >60% for recurring errors
- User feedback: "Fix applied" rate >50%

### Risks & Mitigations
| Risk | Impact | Mitigation |
|------|--------|------------|
| Ollama slow response | High latency | Queue analysis requests, show "analyzing..." spinner |
| AI hallucinations | Wrong suggestions | Add disclaimer: "AI suggestion - verify before applying" |
| Analysis queue backup | Logs lost | Implement backpressure: pause ingestion if queue >1000 |
| Cache invalidation | Stale suggestions | TTL cache entries (24 hours), version by codebase hash |

### Phase 1 Handoff Checklist
When starting Phase 1 conversation:
- [ ] Review this section completely
- [ ] Confirm Ollama models installed (qwen2.5-coder or mistral)
- [ ] Verify Logs service currently functional (basic ingestion working)
- [ ] Check PostgreSQL schema exists (`logs.*` tables)
- [ ] Ensure test database available for integration tests

---

## Phase 2: GitHub Integration (Repo/Folder/Multi-File Scanning) 📋

**Duration:** 3-4 weeks  
**Branch:** TBD  
**Goal:** Enable Review app to analyze entire GitHub repositories, folders, and multiple files

### Business Value
- **Code Review at Scale:** Analyze entire repos instead of single files
- **Onboarding Acceleration:** New developers understand codebase structure quickly
- **Architectural Insights:** Cross-file dependency mapping, bounded context validation

### Acceptance Criteria
- [ ] GitHub API Client (`internal/review/github/client.go`)
  - [ ] Method: `GetRepoTree(owner, repo, branch)` returns hierarchical file structure
  - [ ] Method: `GetFileContent(owner, repo, path, branch)` returns file content (decoded)
  - [ ] Method: `GetPullRequest(owner, repo, prNum)` returns PR metadata
  - [ ] Method: `GetPRFiles(owner, repo, prNum)` returns changed files + diffs
  - [ ] Authentication: Uses `GITHUB_TOKEN` from env or user-provided token
  - [ ] Rate limiting: Respects GitHub API limits (5000/hour authenticated)
- [ ] Session Management Enhancement
  - [ ] New session type: `SessionTypeGitHub` with fields:
    - `github_url`, `owner`, `repo`, `branch`, `file_tree` (JSONB cached)
  - [ ] Database table: `reviews.github_sessions`
  - [ ] Endpoint: `POST /review/sessions/github` creates GitHub session
  - [ ] Endpoint: `GET /review/sessions/:id/tree` returns cached or fresh tree
- [ ] UI - File Tree Viewer (LEFT PANE)
  - [ ] Recursive tree component (`apps/review/templates/components/file_tree.templ`)
  - [ ] Click folder → expand/collapse
  - [ ] Click file → open in new tab (loads content via htmx)
  - [ ] File icons by extension (.go, .js, .md, .yaml, etc.)
  - [ ] Breadcrumb navigation (e.g., `repo/src/handlers/auth.go`)
- [ ] UI - Multi-Tab System
  - [ ] Tab bar above code pane with `+ New Tab` button
  - [ ] Each tab: unique `tab_id` (UUID), filename label, close button
  - [ ] Active tab highlighted
  - [ ] Click tab → switches active pane
  - [ ] Close tab → `hx-confirm="Discard changes?"` if unsaved analysis exists
  - [ ] Tab state persisted in session storage (survive page refresh)
- [ ] Multi-File Analysis
  - [ ] Endpoint: `POST /review/analyze-multiple`
  - [ ] Input: `{session_id, file_ids[], mode}`
  - [ ] Process:
    1. Fetch all file contents
    2. Concatenate with separators: `=== FILE: {path} ===`
    3. Send to Ollama with cross-file context prompt
    4. Return unified analysis
  - [ ] Output: Dependencies between files, shared abstractions, architecture patterns
- [ ] Testing
  - [ ] Unit tests: GitHub API client methods (mocked responses)
  - [ ] Integration test: Full flow (paste GitHub URL → tree loads → file opens → analyze)
  - [ ] UI test (Playwright): Navigate tree, open 3 files in tabs, run multi-file analysis

### Technical Specifications

#### GitHub Client Implementation
```go
// internal/review/github/client.go
type GitHubClient struct {
    token      string
    baseURL    string // "https://api.github.com"
    httpClient *http.Client
    logger     *Logger
    rateLimit  *RateLimiter // Tracks remaining API calls
}

type RepoTree struct {
    SHA      string      `json:"sha"`
    URL      string      `json:"url"`
    Tree     []TreeNode  `json:"tree"`
    Truncated bool       `json:"truncated"`
}

type TreeNode struct {
    Path string `json:"path"`
    Mode string `json:"mode"`
    Type string `json:"type"` // "blob" or "tree"
    SHA  string `json:"sha"`
    Size int    `json:"size"`
    URL  string `json:"url"`
}

func (c *GitHubClient) GetRepoTree(ctx context.Context, owner, repo, branch string) (*RepoTree, error) {
    url := fmt.Sprintf("%s/repos/%s/%s/git/trees/%s?recursive=1", c.baseURL, owner, repo, branch)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    req.Header.Set("Authorization", "Bearer "+c.token)
    req.Header.Set("Accept", "application/vnd.github.v3+json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var tree RepoTree
    json.NewDecoder(resp.Body).Decode(&tree)
    return &tree, nil
}
```

#### Session Schema Additions
```sql
-- db/migrations/XXX_add_github_sessions.sql
CREATE TABLE reviews.github_sessions (
    id SERIAL PRIMARY KEY,
    session_id INT NOT NULL REFERENCES reviews.sessions(id) ON DELETE CASCADE,
    github_url VARCHAR(500) NOT NULL,
    owner VARCHAR(100) NOT NULL,
    repo VARCHAR(100) NOT NULL,
    branch VARCHAR(100) NOT NULL DEFAULT 'main',
    pr_number INT,
    file_tree JSONB, -- Cached tree structure
    last_synced TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE reviews.open_files (
    id SERIAL PRIMARY KEY,
    session_id INT NOT NULL REFERENCES reviews.sessions(id) ON DELETE CASCADE,
    file_path VARCHAR(500) NOT NULL,
    content TEXT,
    language VARCHAR(50),
    tab_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_github_sessions_session_id ON reviews.github_sessions(session_id);
CREATE INDEX idx_open_files_session_id ON reviews.open_files(session_id);
```

#### Multi-File Analysis Prompt
```go
const MultiFilePrompt = `You are analyzing multiple files from the repository {{.Owner}}/{{.Repo}}.
Reading Mode: {{.Mode}}

Files:
{{range .Files}}
=== FILE: {{.Path}} ===
{{.Content}}

{{end}}

Cross-File Analysis Tasks:
1. Identify shared abstractions (interfaces, types used across files)
2. Map dependencies (which files import/call which)
3. Detect architectural patterns (layering, bounded contexts)
4. Flag inconsistencies (different error handling styles, naming conventions)
5. Suggest refactoring opportunities (code duplication, missing abstractions)

Respond in JSON:
{
    "shared_abstractions": [...],
    "dependency_graph": {"nodes": [...], "edges": [...]},
    "architecture_pattern": "...",
    "inconsistencies": [...],
    "refactoring_suggestions": [...]
}
`
```

### Dependencies
- Phase 1 (Logs app) for debugging GitHub API issues
- GitHub Personal Access Token (user provides via UI settings)
- Existing Review app infrastructure

### Success Metrics
- Repo tree loads in <5 seconds for repos with <10,000 files
- Multi-file analysis completes in <30 seconds for 10 files
- Tree navigation feels instant (<100ms per click)
- Cache hit rate: >70% for repeated tree requests

---

## Phase 3: PR Review Workflow 📋

**Duration:** 2-3 weeks  
**Branch:** TBD  
**Goal:** Enable PR review with diff display (left pane) and AI change analysis (right pane)

### Business Value
- **Code Review Automation:** AI pre-reviews PRs, highlights risks before human review
- **Review Consistency:** Every PR analyzed with same rigor (Critical Mode)
- **Learning Tool:** Junior developers learn from AI explanations of changes

### Acceptance Criteria
- [ ] PR Session Creation
  - [ ] Endpoint: `POST /review/sessions/pr` accepts GitHub PR URL
  - [ ] Parses PR number, fetches metadata (title, description, author)
  - [ ] Fetches all changed files with base/head versions
  - [ ] Stores both versions + computes unified diff
  - [ ] Creates session with `type="pull_request"`
- [ ] UI - PR Overview (LEFT PANE)
  - [ ] PR header: number, title, author, branch info
  - [ ] Changed files list with stats (+additions, -deletions)
  - [ ] File status badges (added, modified, deleted, renamed)
  - [ ] Click file → loads diff in viewer below
- [ ] UI - Diff Viewer (LEFT PANE)
  - [ ] Unified diff display with line numbers
  - [ ] Color coding: green (+), red (-), gray (context)
  - [ ] Syntax highlighting based on file extension
  - [ ] Click line → opens AI analysis for that specific change
- [ ] AI PR Analysis (RIGHT PANE)
  - [ ] Button: "Analyze This Change" per file
  - [ ] Endpoint: `POST /review/pr/analyze`
  - [ ] Input: `{session_id, file_path, mode="critical"}`
  - [ ] Output: `{summary, issues, suggestions}`
  - [ ] Categories:
    - **What Changed:** Intent analysis ("This change adds OAuth support")
    - **Regression Risk:** Bugs introduced, missing tests
    - **Architecture:** Bounded context violations, layer mixing
    - **Quality:** Error handling, naming, documentation
- [ ] Comment Integration (Optional)
  - [ ] Button: "Post to GitHub" (requires OAuth scope: `repo`)
  - [ ] Converts AI analysis to GitHub PR comment format
  - [ ] Posts via GitHub API: `POST /repos/{owner}/{repo}/pulls/{number}/comments`
- [ ] Testing
  - [ ] Integration test: Paste PR URL → files load → diff displays → AI analyzes
  - [ ] UI test: Navigate changed files, analyze 3 files, verify output format

### Technical Specifications

#### PR Handler Implementation
```go
// apps/review/handlers/pr_handler.go
func (h *PRHandler) CreatePRSession(c *gin.Context) {
    var req struct {
        GitHubURL string `json:"github_url"` // https://github.com/owner/repo/pull/42
    }
    c.BindJSON(&req)

    // Parse PR number
    prNum, err := parsePRNumber(req.GitHubURL)
    owner, repo := parseRepoFromURL(req.GitHubURL)

    // Fetch PR via GitHub API
    pr, err := h.githubClient.GetPullRequest(c, owner, repo, prNum)
    files, err := h.githubClient.GetPRFiles(c, owner, repo, prNum)

    // Create session
    session := &models.Session{
        UserID: getUserID(c),
        Type:   "pull_request",
        Title:  fmt.Sprintf("PR #%d: %s", prNum, pr.Title),
    }
    h.sessionService.Create(c, session)

    // Store PR metadata
    h.prService.CreatePRSession(c, session.ID, owner, repo, prNum, files)

    c.JSON(200, gin.H{"session_id": session.ID})
}
```

#### PR Analysis Prompt
```go
const PRAnalysisPrompt = `You are reviewing a Pull Request.

PR Title: {{.PRTitle}}
File: {{.FilePath}}

BASE VERSION (before changes):
{{.BaseContent}}

HEAD VERSION (after changes):
{{.HeadContent}}

UNIFIED DIFF:
{{.Diff}}

Analyze in Critical Mode:
1. What changed and why? (intent analysis - be specific)
2. Are there bugs introduced? (regression risk - cite line numbers)
3. Does this violate architecture patterns? (bounded context, layering)
4. Are tests needed? (coverage gap - suggest test cases)
5. Suggestions for improvement? (concrete code examples)

Respond in JSON:
{
    "summary": "Brief what/why of changes",
    "issues": [
        {"severity": "critical|important|minor", "title": "...", "description": "...", "line": 42, "suggested_fix": "..."}
    ],
    "suggestions": ["Suggestion 1", "Suggestion 2"]
}
`
```

### Dependencies
- Phase 2 (GitHub integration) for API client and auth
- Existing Review app with Critical Mode analysis

### Success Metrics
- PR loads in <10 seconds (including all file diffs)
- AI analysis completes in <30 seconds per file
- Issue detection accuracy: >70% (verified by human reviewers)

---

## Phase 4: Test Output Integration 📋

**Duration:** 3-4 weeks  
**Branch:** TBD  
**Goal:** Ingest test results (JUnit XML, Go JSON, pytest) and provide AI-driven failure analysis

### Acceptance Criteria
- [ ] Test Ingestion Service (`internal/review/services/test_service.go`)
  - [ ] Supports formats: JUnit XML, Go test JSON, pytest JSON, plain text
  - [ ] Endpoint: `POST /review/sessions/test-upload`
  - [ ] Parses test output into structured data: `{name, status, duration, error_msg, stacktrace, file, line}`
  - [ ] Stores in `reviews.test_results` table
  - [ ] Creates session with `type="test_results"`
- [ ] UI - Test Results View (LEFT PANE)
  - [ ] Summary bar: `X passed, Y failed, Z skipped, total duration`
  - [ ] Test list with status icons (✅ pass, ❌ fail, ⏭️ skip)
  - [ ] Click test → expands details (error message, stacktrace)
  - [ ] Filter: show only failures, search by name
- [ ] AI Test Failure Analysis (RIGHT PANE)
  - [ ] Button: "Analyze Failure" per failed test
  - [ ] Endpoint: `POST /review/test/analyze`
  - [ ] Input: `{test_id}`
  - [ ] Output: `{root_cause, fix_steps, suggested_code, related_tests}`
  - [ ] Displays:
    - Root cause explanation
    - Step-by-step fix instructions
    - Code snippet (before/after)
    - Related tests that might fail
- [ ] CI/CD Webhook Integration (Optional)
  - [ ] Endpoint: `POST /api/review/ci-webhook`
  - [ ] Accepts webhooks from GitHub Actions, GitLab CI, Jenkins
  - [ ] Auto-creates test session + triggers AI analysis
  - [ ] Notifies user: "New test failures + AI analysis ready"
- [ ] Testing
  - [ ] Unit tests: Parsers for all supported formats
  - [ ] Integration test: Upload JUnit XML → results display → analyze failure
  - [ ] UI test: Navigate test list, filter failures, analyze 2 failures

### Technical Specifications

#### Test Parser Implementation
```go
// internal/review/services/test_parser.go
type TestParser interface {
    Parse(data []byte) ([]TestResult, error)
}

type JUnitParser struct{}
func (p *JUnitParser) Parse(data []byte) ([]TestResult, error) {
    var suite JUnitTestSuite
    xml.Unmarshal(data, &suite)
    
    results := []TestResult{}
    for _, testcase := range suite.TestCases {
        result := TestResult{
            Name:     testcase.Name,
            Status:   "pass",
            Duration: testcase.Time,
        }
        if testcase.Failure != nil {
            result.Status = "fail"
            result.ErrorMsg = testcase.Failure.Message
            result.Stacktrace = testcase.Failure.Text
        }
        results = append(results, result)
    }
    return results, nil
}

type GoTestJSONParser struct{}
func (p *GoTestJSONParser) Parse(data []byte) ([]TestResult, error) {
    // Parse JSONL format from `go test -json`
}
```

#### Test Analysis Prompt
```go
const TestAnalysisPrompt = `Test Failure Analysis:

Test: {{.TestName}}
File: {{.File}}:{{.Line}}

Error Message:
{{.ErrorMsg}}

Stack Trace:
{{.Stacktrace}}

Related Code (if available):
{{.CodeContext}}

Analyze:
1. What caused this failure? (root cause - be specific)
2. Is this a code bug or test bug? (categorize)
3. How to fix it? (step-by-step instructions)
4. Related tests that might fail? (impact analysis)

Respond in JSON:
{
    "root_cause": "...",
    "category": "code_bug|test_bug|environment|flaky",
    "fix_steps": ["Step 1", "Step 2"],
    "suggested_code": "...",
    "related_tests": ["test_name_1", "test_name_2"]
}
`
```

### Dependencies
- Phase 1 (Logs app) for error analysis patterns
- Existing Review app infrastructure

### Success Metrics
- Parser accuracy: >95% for supported formats
- AI analysis completes in <10 seconds per test
- Fix suggestion accuracy: >60% (verified by developers)

---

## Phase 5: Text-to-Speech with Highlight Sync 📋

**Duration:** 1-2 weeks  
**Branch:** TBD  
**Goal:** Add Google Cloud TTS with synchronized text highlighting in Review app

### Acceptance Criteria
- [ ] Google Cloud TTS Integration (`internal/review/tts/google_client.go`)
  - [ ] Method: `GenerateSpeech(text, voice string) ([]byte, error)`
  - [ ] Calls: `POST https://texttospeech.googleapis.com/v1/text:synthesize`
  - [ ] Returns: MP3 audio bytes
  - [ ] Requires: `GOOGLE_CLOUD_API_KEY` env var
- [ ] TTS Endpoint
  - [ ] Endpoint: `POST /review/tts`
  - [ ] Input: `{text, voice}`
  - [ ] Output: `audio/mpeg` stream
  - [ ] Streams audio directly to browser
- [ ] UI - Voice Selector
  - [ ] Dropdown with options:
    - US Female (en-US-Wavenet-F)
    - US Male (en-US-Standard-D)
    - UK Male (en-GB-Standard-A)
    - AU Female (en-AU-Standard-C)
  - [ ] Selection persisted in localStorage
- [ ] UI - Reading Mode
  - [ ] Button: "🔊 Read Aloud"
  - [ ] Sends visible code/analysis text to TTS endpoint
  - [ ] Plays audio via `<audio>` tag with autoplay
  - [ ] Pause on tab switch or "Stop" button
- [ ] Text Highlight Sync
  - [ ] Backend: Split text into words with timestamps
  - [ ] Frontend: Track audio playback position
  - [ ] Highlight current word/sentence with `.tts-highlight.active` class
  - [ ] Scroll viewport to keep highlighted text visible
- [ ] Testing
  - [ ] Unit test: TTS client generates valid MP3
  - [ ] UI test: Click "Read Aloud" → audio plays → text highlights → stop works

### Technical Specifications

#### Google TTS Client
```go
// internal/review/tts/google_client.go
type GoogleTTSClient struct {
    apiKey     string
    httpClient *http.Client
}

type TTSRequest struct {
    Input struct {
        Text string `json:"text"`
    } `json:"input"`
    Voice struct {
        LanguageCode string `json:"languageCode"` // "en-US"
        Name         string `json:"name"`         // "en-US-Wavenet-F"
    } `json:"voice"`
    AudioConfig struct {
        AudioEncoding string `json:"audioEncoding"` // "MP3"
    } `json:"audioConfig"`
}

func (c *GoogleTTSClient) GenerateSpeech(ctx context.Context, text, voice string) ([]byte, error) {
    url := "https://texttospeech.googleapis.com/v1/text:synthesize?key=" + c.apiKey
    
    reqBody := TTSRequest{}
    reqBody.Input.Text = text
    reqBody.Voice.LanguageCode = "en-US"
    reqBody.Voice.Name = voice
    reqBody.AudioConfig.AudioEncoding = "MP3"
    
    jsonData, _ := json.Marshal(reqBody)
    req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result struct {
        AudioContent string `json:"audioContent"` // Base64-encoded MP3
    }
    json.NewDecoder(resp.Body).Decode(&result)
    
    audioBytes, _ := base64.StdEncoding.DecodeString(result.AudioContent)
    return audioBytes, nil
}
```

#### Highlight Sync (Alpine.js)
```html
<!-- apps/review/templates/workspace.templ -->
<div x-data="ttsSync()">
    <audio id="audio-player" 
           @timeupdate="updateHighlight($event.target.currentTime)"></audio>
    
    <div id="code-pane">
        <template x-for="(word, idx) in words" :key="idx">
            <span :class="{'tts-highlight': true, 'active': currentWordIndex === idx}"
                  x-text="word"></span>
        </template>
    </div>
</div>

<script>
function ttsSync() {
    return {
        words: [],
        timestamps: [], // [0, 0.5, 1.2, 1.8, ...]
        currentWordIndex: 0,
        
        updateHighlight(currentTime) {
            // Find word index for current playback time
            for (let i = 0; i < this.timestamps.length; i++) {
                if (currentTime >= this.timestamps[i] && 
                    currentTime < this.timestamps[i + 1]) {
                    this.currentWordIndex = i;
                    break;
                }
            }
        }
    }
}
</script>
```

### Dependencies
- Google Cloud API key (user provides)
- Existing Review app infrastructure

### Success Metrics
- TTS latency: <2 seconds from button click to audio start
- Highlight accuracy: <100ms sync lag between audio and text
- Voice quality: User rating >4/5

---

## Cross-Phase Dependencies

```
Phase 1 (Logs AI)
    ↓ (provides debugging infrastructure)
Phase 2 (GitHub Integration)
    ↓ (provides repo/file fetching)
Phase 3 (PR Review)
    ↓ (provides test data source)
Phase 4 (Test Integration)

Phase 5 (TTS) ← Independent, no dependencies
```

**Parallel Execution:**
- If team has 2 developers:
  - Dev 1: Phase 1 → Phase 2 → Phase 3
  - Dev 2: Phase 4 (after Phase 1 complete) + Phase 5 (anytime)

---

## Quality Gates (All Phases)

**Pre-Push Hook Enforces:**
1. ✅ `gofmt` formatting on all modified `.go` files
2. ✅ `goimports` import resolution
3. ✅ `go build` succeeds for modified packages
4. ✅ `golangci-lint` passes (only modified files)
5. ✅ `go vet` clean (modified packages)
6. ⚠️ E2E smoke tests (optional - warns if fails, doesn't block)

**Pre-Merge Checklist:**
- [ ] All acceptance criteria met (checkboxes above)
- [ ] Unit tests written + passing (coverage >70%)
- [ ] Integration tests passing
- [ ] UI tests (Playwright) passing for new features
- [ ] Manual testing checklist completed
- [ ] Documentation updated (if API changes)
- [ ] PR description includes:
  - Summary of changes
  - Testing evidence (screenshots, curl output, test results)
  - Acceptance criteria checklist

**Merge Process:**
1. Create PR from feature branch → `development`
2. Pre-push hook validates quality
3. GitHub Actions runs full CI (all tests, builds, Docker validation)
4. Human review (Mike or Claude)
5. Approval → Squash merge to `development`
6. Delete feature branch
7. Create new branch for next phase

---

## Progress Tracking

### Phase Completion Status

| Phase | Status | Branch | PR | Merged | Started | Completed |
|-------|--------|--------|----|----|---------|-----------|
| Phase 0: Model Service | ✅ DONE | `development` | - | ✅ | 2025-11-04 | 2025-11-04 |
| Phase 1: Logs AI | ✅ DONE | `development` | #103, #104 | ✅ | 2025-11-04 | 2025-11-04 |
| Phase 2: GitHub | � NEXT | TBD | - | - | - | - |
| Phase 3: PR Review | 📋 PLANNED | - | - | - | - | - |
| Phase 4: Tests | 📋 PLANNED | - | - | - | - | - |
| Phase 5: TTS | 📋 PLANNED | - | - | - | - | - |

### Current Focus
**Security Hardening Complete**
- Status: Removed all hardcoded JWT secrets
- Commit: 65859f8 (security: remove all hardcoded JWT secrets)
- Next Action: Push to origin and prepare for Phase 2

**Phase 2 - GitHub Integration**
- Status: Ready to begin
- Next Action: Create feature branch and review acceptance criteria

---

## Conversation Handoff Template

**For starting each phase conversation, provide this context:**

```markdown
# Phase [N] Implementation - [Phase Name]

**Context:** I'm working on Phase [N] of the DevSmith Platform Implementation Roadmap.

**Reference Document:** `.docs/IMPLEMENTATION_ROADMAP.md` - Section "Phase [N]"

**Acceptance Criteria:** See roadmap document (checklist with [ ] items)

**Current Status:**
- Previous Phase Status: [Completed/Merged]
- Current Branch: [branch-name]
- Services Running: [docker-compose ps output]
- Ollama Models Available: [ollama list output]

**Ready to Start:** Please review Phase [N] acceptance criteria and provide:
1. Detailed implementation plan (file-by-file)
2. Database migration scripts (if needed)
3. Test cases to write first (TDD approach)
4. Manual testing checklist

Let's begin with [first task from acceptance criteria].
```

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-11-04 | GitHub Copilot | Initial roadmap creation with 5 phases + Phase 0 completion |

---

**End of Roadmap Document**

This is a living document. Update acceptance criteria checkboxes as work progresses. Add new sections as needed. Keep Phase status table current.
