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

## Implementation

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
