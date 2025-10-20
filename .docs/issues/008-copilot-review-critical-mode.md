# Issue #008: [COPILOT] Review Service - Critical Mode

**Labels:** `copilot`, `review`, `reading-mode`, `security`
**Created:** 2025-10-19
**Issue:** #8
**Estimated Time:** 90-120 minutes
**Depends On:** Issue #007

---

# 🚨 STEP 0: CREATE FEATURE BRANCH FIRST 🚨

```bash
git checkout development && git pull origin development
git checkout -b feature/008-copilot-review-critical-mode
git branch --show-current
```

---

## Task Description

Implement Critical Mode - evaluative review. AI identifies bugs, security vulnerabilities, performance issues, anti-patterns. Most valuable for HITL supervision.

**This is the platform's key differentiator** - teaches developers to supervise AI output critically.

---

## Success Criteria
- [ ] AI identifies security vulnerabilities (SQL injection, XSS, etc.)
- [ ] Detects bugs and logic errors
- [ ] Flags performance issues (N+1 queries, inefficient algorithms)
- [ ] Identifies anti-patterns and code smells
- [ ] Provides severity ratings (critical/high/medium/low)
- [ ] Suggests fixes
- [ ] 70%+ test coverage

---

## ⚠️ CRITICAL: Test-Driven Development (TDD) Required

**YOU MUST WRITE TESTS FIRST, THEN IMPLEMENTATION.**

Follow the Red-Green-Refactor cycle from DevsmithTDD.md:
1. **RED**: Write failing test
2. **GREEN**: Write minimal code to pass
3. **REFACTOR**: Improve code quality

### TDD Workflow for This Issue

**Step 1: RED PHASE (Write Failing Tests) - DO THIS FIRST!**

Create `internal/review/services/critical_service_test.go` BEFORE writing `critical_service.go`:

```go
package services

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test 1: Finds security vulnerabilities
func TestCriticalService_AnalyzeCritical_FindsSecurityIssues(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewCriticalService(mockOllama, mockRepo)

	aiResponse := `{
		"issues": [
			{
				"severity": "critical",
				"category": "security",
				"file": "auth.go",
				"line": 10,
				"code_snippet": "db.Query(userInput)",
				"description": "SQL injection vulnerability",
				"impact": "Attacker can access entire database",
				"fix_suggestion": "Use parameterized queries: db.Query(sql, userInput)"
			}
		],
		"summary": "Found 1 critical security issue",
		"overall_grade": "D"
	}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	output, err := service.AnalyzeCritical(context.Background(), 1, "owner", "repo")

	assert.NoError(t, err)
	assert.Len(t, output.Issues, 1)
	assert.Equal(t, "critical", output.Issues[0].Severity)
	assert.Equal(t, "security", output.Issues[0].Category)
	assert.Contains(t, output.Issues[0].Description, "SQL injection")
	assert.Equal(t, "D", output.OverallGrade)
}

// Test 2: Finds multiple issue types
func TestCriticalService_AnalyzeCritical_MultipleIssueTypes(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewCriticalService(mockOllama, mockRepo)

	aiResponse := `{
		"issues": [
			{
				"severity": "critical",
				"category": "security",
				"file": "auth.go",
				"line": 10,
				"code_snippet": "eval(userInput)",
				"description": "Code injection vulnerability",
				"impact": "Remote code execution",
				"fix_suggestion": "Never use eval with user input"
			},
			{
				"severity": "high",
				"category": "performance",
				"file": "users.go",
				"line": 25,
				"code_snippet": "for user in users: db.query()",
				"description": "N+1 query problem",
				"impact": "Database overload with many users",
				"fix_suggestion": "Use JOIN or batch query"
			},
			{
				"severity": "medium",
				"category": "maintainability",
				"file": "utils.go",
				"line": 100,
				"code_snippet": "if ... elif ... elif ... (50 lines)",
				"description": "Excessive cyclomatic complexity",
				"impact": "Hard to test and maintain",
				"fix_suggestion": "Refactor to switch statement or strategy pattern"
			}
		],
		"summary": "Found 1 critical, 1 high, 1 medium issue",
		"overall_grade": "C"
	}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	output, err := service.AnalyzeCritical(context.Background(), 1, "owner", "repo")

	assert.NoError(t, err)
	assert.Len(t, output.Issues, 3)

	// Verify severity levels
	severities := []string{output.Issues[0].Severity, output.Issues[1].Severity, output.Issues[2].Severity}
	assert.Contains(t, severities, "critical")
	assert.Contains(t, severities, "high")
	assert.Contains(t, severities, "medium")

	// Verify categories
	categories := []string{output.Issues[0].Category, output.Issues[1].Category, output.Issues[2].Category}
	assert.Contains(t, categories, "security")
	assert.Contains(t, categories, "performance")
	assert.Contains(t, categories, "maintainability")
}

// Test 3: Clean code (no issues)
func TestCriticalService_AnalyzeCritical_CleanCode(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewCriticalService(mockOllama, mockRepo)

	aiResponse := `{
		"issues": [],
		"summary": "No issues found - excellent code quality",
		"overall_grade": "A"
	}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	output, err := service.AnalyzeCritical(context.Background(), 1, "owner", "repo")

	assert.NoError(t, err)
	assert.Empty(t, output.Issues)
	assert.Equal(t, "A", output.OverallGrade)
}

// Test 4: Handles AI parsing errors
func TestCriticalService_AnalyzeCritical_InvalidJSON(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewCriticalService(mockOllama, mockRepo)

	aiResponse := `Invalid JSON response`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)

	output, err := service.AnalyzeCritical(context.Background(), 1, "owner", "repo")

	// Should handle gracefully, not crash
	assert.NoError(t, err)
	assert.NotNil(t, output)
	// Fallback response expected
}
```

**Run tests (they should FAIL):**
```bash
go test ./internal/review/services/...
# Expected: FAIL - NewCriticalService undefined
```

**Commit the failing tests:**
```bash
git add internal/review/services/critical_service_test.go
git commit -m "test(review): add failing tests for Critical Mode (RED phase)"
```

**Step 2: GREEN PHASE (Make Tests Pass)**

Now implement `critical_service.go` to make tests pass. See Phase 1 below.

**Step 3: Verify Build (CRITICAL)**

Before committing implementation:
```bash
# Build must succeed
go build -o /dev/null ./cmd/review

# If build fails, fix errors before committing
```

**Step 4: Commit Implementation**
```bash
git add internal/review/services/critical_service.go
git commit -m "feat(review): implement Critical Mode service (GREEN phase)"
```

**Reference:** DevsmithTDD.md lines 15-36 (Red-Green-Refactor cycle)

**Note:** This is the platform's centerpiece mode - most important for HITL training. Tests must be comprehensive.

---

## Implementation

**IMPORTANT: Follow TDD workflow above. Write tests FIRST (already shown), then implement.**

### Phase 1: Critical Service

**File:** `internal/review/services/critical_service.go`
```go
package services

type CriticalService struct {
	ollamaClient *OllamaClient
	analysisRepo AnalysisRepositoryInterface
}

func NewCriticalService(ollamaClient *OllamaClient, analysisRepo AnalysisRepositoryInterface) *CriticalService {
	return &CriticalService{ollamaClient, analysisRepo}
}

func (s *CriticalService) AnalyzeCritical(ctx context.Context, reviewID int64, repoOwner, repoName string) (*models.CriticalModeOutput, error) {
	prompt := fmt.Sprintf(`Review repository %s/%s in Critical Mode.

Identify:
1. Security vulnerabilities (SQL injection, XSS, secrets in code, etc.)
2. Bugs and logic errors
3. Performance issues (N+1 queries, inefficient algorithms)
4. Anti-patterns and code smells
5. Missing error handling
6. Concurrency issues

Return JSON:
{
  "issues": [
    {
      "severity": "critical|high|medium|low",
      "category": "security|bug|performance|maintainability",
      "file": "path/to/file.go",
      "line": 42,
      "code_snippet": "...",
      "description": "SQL injection vulnerability",
      "impact": "Attacker can access database",
      "fix_suggestion": "Use parameterized queries"
    }
  ],
  "summary": "Found 5 critical, 10 high, 15 medium issues",
  "overall_grade": "C"
}`, repoOwner, repoName)

	rawOutput, _ := s.ollamaClient.Generate(ctx, prompt)
	var output models.CriticalModeOutput
	json.Unmarshal([]byte(rawOutput), &output)

	metadataJSON, _ := json.Marshal(output)
	result := &models.AnalysisResult{
		ReviewID:  reviewID,
		Mode:      models.CriticalMode,
		Prompt:    prompt,
		RawOutput: rawOutput,
		Summary:   output.Summary,
		Metadata:  string(metadataJSON),
		ModelUsed: "qwen2.5-coder:32b",
	}
	s.analysisRepo.Create(ctx, result)

	return &output, nil
}
```

**Commit:** `git add internal/review/services/critical* && git commit -m "feat(review): add Critical Mode service"`

---

### Phase 2: Models

**Append to:** `internal/review/models/review.go`
```go
type CriticalModeOutput struct {
	Issues       []CodeIssue `json:"issues"`
	Summary      string      `json:"summary"`
	OverallGrade string      `json:"overall_grade"`
}

type CodeIssue struct {
	Severity      string `json:"severity"`       // critical, high, medium, low
	Category      string `json:"category"`       // security, bug, performance, maintainability
	File          string `json:"file"`
	Line          int    `json:"line"`
	CodeSnippet   string `json:"code_snippet"`
	Description   string `json:"description"`
	Impact        string `json:"impact"`
	FixSuggestion string `json:"fix_suggestion"`
}
```

**Commit:** `git add internal/review/models/ && git commit -m "feat(review): add Critical Mode models"`

---

### Phase 3: Handler & Route

**File:** `cmd/review/handlers/review_handler.go`
```go
func (h *ReviewHandler) GetCriticalAnalysis(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	review, _ := h.reviewService.GetReview(c.Request.Context(), id)
	output, _ := h.criticalService.AnalyzeCritical(c.Request.Context(), review.ID, "owner", "repo")

	c.JSON(http.StatusOK, output)
}
```

**Add route:** `api.GET("/reviews/:id/critical", reviewHandler.GetCriticalAnalysis)`

**Commit:** `git add cmd/review/ && git commit -m "feat(review): add Critical Mode endpoint"`

---

### Phase 4: Tests

**File:** `internal/review/services/critical_service_test.go`
```go
func TestCriticalService_AnalyzeCritical_FindsVulnerabilities(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewCriticalService(mockOllama, mockRepo)

	aiResponse := `{
		"issues": [
			{"severity": "critical", "category": "security", "file": "auth.go", "line": 10,
			 "code_snippet": "db.Query(userInput)", "description": "SQL injection",
			 "impact": "Database compromise", "fix_suggestion": "Use parameterized queries"}
		],
		"summary": "Found 1 critical issue",
		"overall_grade": "D"
	}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	output, err := service.AnalyzeCritical(context.Background(), 1, "owner", "repo")

	assert.NoError(t, err)
	assert.Len(t, output.Issues, 1)
	assert.Equal(t, "critical", output.Issues[0].Severity)
	assert.Equal(t, "security", output.Issues[0].Category)
}
```

**Commit:** `git add internal/review/services/critical* && git commit -m "test(review): add Critical Mode tests"`

---

### Phase 5: Push

```bash
git push -u origin feature/008-copilot-review-critical-mode
```

---

## References
- `ARCHITECTURE.md` lines 872-966 (Critical Mode spec)

**Time:** 90-120 minutes
**Note:** This mode is the platform centerpiece - most important for HITL training
