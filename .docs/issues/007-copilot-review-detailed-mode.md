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

## ⚠️ CRITICAL: Test-Driven Development (TDD) Required

**YOU MUST WRITE TESTS FIRST, THEN IMPLEMENTATION.**

Follow the Red-Green-Refactor cycle from DevsmithTDD.md:
1. **RED**: Write failing test
2. **GREEN**: Write minimal code to pass
3. **REFACTOR**: Improve code quality

### TDD Workflow for This Issue

**Step 1: RED PHASE (Write Failing Tests) - DO THIS FIRST!**

Create `internal/review/services/detailed_service_test.go` BEFORE writing `detailed_service.go`:

```go
package services

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test 1: Successful detailed analysis
func TestDetailedService_AnalyzeDetailed_Success(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewDetailedService(mockOllama, mockRepo)

	aiResponse := `{
		"lines": [
			{
				"line_num": 1,
				"code": "package main",
				"explanation": "Package declaration",
				"complexity": "low"
			},
			{
				"line_num": 5,
				"code": "func Login()",
				"explanation": "Authentication handler",
				"complexity": "medium",
				"side_effects": ["Database query", "Session creation"],
				"variables_modified": ["userSession"]
			}
		],
		"data_flow": [
			{"from": "input", "to": "database", "description": "Credentials validated"}
		],
		"summary": "Authentication file with 2 functions"
	}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	output, err := service.AnalyzeDetailed(context.Background(), 1, "auth.go", "owner", "repo")

	assert.NoError(t, err)
	assert.Len(t, output.Lines, 2)
	assert.Equal(t, "package main", output.Lines[0].Code)
	assert.Contains(t, output.Lines[1].SideEffects, "Database query")
	assert.Len(t, output.DataFlow, 1)
}

// Test 2: Empty file path returns error
func TestDetailedService_AnalyzeDetailed_EmptyFilePath(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewDetailedService(mockOllama, mockRepo)

	_, err := service.AnalyzeDetailed(context.Background(), 1, "", "owner", "repo")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file path cannot be empty")
}

// Test 3: Complex file with side effects
func TestDetailedService_AnalyzeDetailed_WithSideEffects(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewDetailedService(mockOllama, mockRepo)

	aiResponse := `{
		"lines": [
			{
				"line_num": 10,
				"code": "db.Exec(sql)",
				"explanation": "Executes database query",
				"complexity": "high",
				"side_effects": ["Database write", "Triggers audit log"],
				"variables_modified": ["recordCount", "lastModified"]
			}
		],
		"data_flow": [
			{"from": "userInput", "to": "database", "description": "Unvalidated input risk"}
		],
		"summary": "Database operations with side effects"
	}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	output, err := service.AnalyzeDetailed(context.Background(), 1, "db.go", "owner", "repo")

	assert.NoError(t, err)
	assert.Equal(t, "high", output.Lines[0].Complexity)
	assert.Len(t, output.Lines[0].SideEffects, 2)
	assert.Len(t, output.Lines[0].VariablesModified, 2)
}
```

**Run tests (they should FAIL):**
```bash
go test ./internal/review/services/...
# Expected: FAIL - NewDetailedService undefined
```

**Commit the failing tests:**
```bash
git add internal/review/services/detailed_service_test.go
git commit -m "test(review): add failing tests for Detailed Mode (RED phase)"
```

**Step 2: GREEN PHASE (Make Tests Pass)**

Now implement `detailed_service.go` to make tests pass. See Phase 1 below.

**Step 3: Verify Build (CRITICAL)**

Before committing implementation:
```bash
# Build must succeed
go build -o /dev/null ./cmd/review

# If build fails, fix errors before committing
```

**Step 4: Commit Implementation**
```bash
git add internal/review/services/detailed_service.go
git commit -m "feat(review): implement Detailed Mode service (GREEN phase)"
```

**Reference:** DevsmithTDD.md lines 15-36 (Red-Green-Refactor cycle)

---

## Implementation

**IMPORTANT: Follow TDD workflow above. Write tests FIRST (already shown), then implement.**

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
