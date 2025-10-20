# Issue #007: [COPILOT] Review Service - Detailed Mode

**Labels:** `copilot`, `review`, `reading-mode`
**Created:** 2025-10-19
**Issue:** #7
**Estimated Time:** 90-120 minutes
**Depends On:** Issue #006

---

# 🚨 STEP 0: CREATE FEATURE BRANCH FIRST 🚨

```bash
git checkout development && git pull origin development
git checkout -b feature/007-copilot-review-detailed-mode
git branch --show-current
```

---

## Task Description

Implement Detailed Mode - line-by-line explanations. Most complex mode. AI explains every line, shows data flow, traces variables.

---

## Success Criteria
- [ ] User selects file or function
- [ ] AI provides line-by-line explanations
- [ ] Shows data flow and variable tracing
- [ ] Identifies side effects and dependencies
- [ ] 70%+ test coverage

---

## Implementation

### Phase 1: Detailed Service

**File:** `internal/review/services/detailed_service.go`
```go
package services

type DetailedService struct {
	ollamaClient *OllamaClient
	analysisRepo AnalysisRepositoryInterface
}

func NewDetailedService(ollamaClient *OllamaClient, analysisRepo AnalysisRepositoryInterface) *DetailedService {
	return &DetailedService{ollamaClient, analysisRepo}
}

func (s *DetailedService) AnalyzeDetailed(ctx context.Context, reviewID int64, filePath string, repoOwner, repoName string) (*models.DetailedModeOutput, error) {
	prompt := fmt.Sprintf(`Analyze file %s in repository %s/%s in Detailed Mode.

Provide line-by-line explanations.

Return JSON:
{
  "lines": [
    {"line_num": 1, "code": "package main", "explanation": "Declares package", "complexity": "low"},
    {"line_num": 5, "code": "func Login()", "explanation": "Authenticates user", "complexity": "medium", "side_effects": ["Database query"], "variables_modified": ["userSession"]}
  ],
  "data_flow": [{"from": "input", "to": "database", "description": "User credentials validated"}],
  "summary": "This file handles authentication with 3 main functions"
}`, filePath, repoOwner, repoName)

	rawOutput, _ := s.ollamaClient.Generate(ctx, prompt)
	var output models.DetailedModeOutput
	json.Unmarshal([]byte(rawOutput), &output)

	metadataJSON, _ := json.Marshal(output)
	result := &models.AnalysisResult{
		ReviewID:  reviewID,
		Mode:      models.DetailedMode,
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

**Commit:** `git add internal/review/services/detailed* && git commit -m "feat(review): add Detailed Mode service"`

---

### Phase 2: Models

**Append to:** `internal/review/models/review.go`
```go
type DetailedModeOutput struct {
	Lines    []LineExplanation `json:"lines"`
	DataFlow []DataFlowStep    `json:"data_flow"`
	Summary  string            `json:"summary"`
}

type LineExplanation struct {
	LineNum           int      `json:"line_num"`
	Code              string   `json:"code"`
	Explanation       string   `json:"explanation"`
	Complexity        string   `json:"complexity"`
	SideEffects       []string `json:"side_effects,omitempty"`
	VariablesModified []string `json:"variables_modified,omitempty"`
}

type DataFlowStep struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Description string `json:"description"`
}
```

**Commit:** `git add internal/review/models/ && git commit -m "feat(review): add Detailed Mode models"`

---

### Phase 3: Handler & Route

**File:** `cmd/review/handlers/review_handler.go`
```go
func (h *ReviewHandler) GetDetailedAnalysis(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	filePath := c.Query("file")  // GET /api/reviews/:id/detailed?file=auth.go

	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file parameter required"})
		return
	}

	review, _ := h.reviewService.GetReview(c.Request.Context(), id)
	output, _ := h.detailedService.AnalyzeDetailed(c.Request.Context(), review.ID, filePath, "owner", "repo")

	c.JSON(http.StatusOK, output)
}
```

**Add route:** `api.GET("/reviews/:id/detailed", reviewHandler.GetDetailedAnalysis)`

**Commit:** `git add cmd/review/ && git commit -m "feat(review): add Detailed Mode endpoint"`

---

### Phase 4: Tests

**File:** `internal/review/services/detailed_service_test.go`
```go
func TestDetailedService_AnalyzeDetailed_Success(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewDetailedService(mockOllama, mockRepo)

	aiResponse := `{"lines": [{"line_num": 1, "code": "package main", "explanation": "Package declaration", "complexity": "low"}], "data_flow": [], "summary": "Simple file"}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	output, err := service.AnalyzeDetailed(context.Background(), 1, "main.go", "owner", "repo")

	assert.NoError(t, err)
	assert.Len(t, output.Lines, 1)
}
```

**Commit:** `git add internal/review/services/detailed* && git commit -m "test(review): add Detailed Mode tests"`

---

### Phase 5: Push

```bash
git push -u origin feature/007-copilot-review-detailed-mode
```

---

## References
- `ARCHITECTURE.md` lines 814-870 (Detailed Mode spec)

**Time:** 90-120 minutes (most complex mode)
