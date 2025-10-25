# Issue #005: [COPILOT] Review Service - Skim Mode

**Labels:** `copilot`, `review`, `reading-mode`
**Assignee:** Mike (with Copilot assistance)
**Created:** 2025-10-19
**Issue:** #5
**Estimated Time:** 60-90 minutes
**Depends On:** Issue #004 (Preview Mode)

---

# 🚨 STEP 0: CREATE FEATURE BRANCH FIRST 🚨

```bash
git checkout development && git pull origin development
git checkout -b feature/005-copilot-review-skim-mode
git branch --show-current  # Verify: feature/005-copilot-review-skim-mode
```

**✅ Only proceed after creating branch**

---

## Task Description

Implement Skim Mode - the second reading mode. Shows function signatures, interfaces, and data models WITHOUT implementation details. Teaches "what the code does" without "how it does it."

**Builds on:** Preview Mode (#004) - same database, same service structure
**What's new:** Different AI prompt, different parsing logic, similar UI patterns

---

## Success Criteria
- [ ] Skim Mode analyzes code via Ollama (function signatures only)
- [ ] AI extracts interfaces, function signatures, data models
- [ ] Results stored in analysis_results table (mode='skim')
- [ ] UI displays signatures with collapsible sections
- [ ] Can switch between Preview and Skim modes
- [ ] 70%+ test coverage

---

## Implementation (Reuse #004 Patterns)

### Phase 1: Add Skim Service

**File:** `internal/review/services/skim_service.go`
```go
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

type SkimService struct {
	ollamaClient *OllamaClient
	analysisRepo AnalysisRepositoryInterface
}

func NewSkimService(ollamaClient *OllamaClient, analysisRepo AnalysisRepositoryInterface) *SkimService {
	return &SkimService{
		ollamaClient: ollamaClient,
		analysisRepo: analysisRepo,
	}
}

func (s *SkimService) AnalyzeSkim(ctx context.Context, reviewID int64, repoOwner, repoName string) (*models.SkimModeOutput, error) {
	// Check cache
	existing, _ := s.analysisRepo.FindByReviewAndMode(ctx, reviewID, models.SkimMode)
	if existing != nil {
		var output models.SkimModeOutput
		json.Unmarshal([]byte(existing.Metadata), &output)
		return &output, nil
	}

	// Generate prompt
	prompt := s.buildSkimPrompt(repoOwner, repoName)

	// Call AI
	rawOutput, err := s.ollamaClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse response
	output, _ := s.parseSkimOutput(rawOutput)

	// Store in DB
	metadataJSON, _ := json.Marshal(output)
	result := &models.AnalysisResult{
		ReviewID:   reviewID,
		Mode:       models.SkimMode,
		Prompt:     prompt,
		RawOutput:  rawOutput,
		Summary:    output.Summary,
		Metadata:   string(metadataJSON),
		ModelUsed:  "qwen2.5-coder:32b",
	}
	s.analysisRepo.Create(ctx, result)

	return output, nil
}

func (s *SkimService) buildSkimPrompt(owner, repo string) string {
	return fmt.Sprintf(`Analyze repository %s/%s in Skim Mode.

Goal: Extract function signatures, interfaces, and data models WITHOUT implementation details.

Provide JSON:
{
  "functions": [{"name": "FunctionName", "signature": "func(arg Type) ReturnType", "description": "What it does"}],
  "interfaces": [{"name": "InterfaceName", "methods": ["Method1", "Method2"], "purpose": "Why it exists"}],
  "data_models": [{"name": "StructName", "fields": ["field1", "field2"], "purpose": "What it represents"}],
  "workflows": [{"name": "User Login Flow", "steps": ["Step1", "Step2"]}],
  "summary": "2-3 sentences"
}

Repository: https://github.com/%s/%s`, owner, repo, owner, repo)
}

func (s *SkimService) parseSkimOutput(raw string) (*models.SkimModeOutput, error) {
	var output models.SkimModeOutput
	json.Unmarshal([]byte(raw), &output)
	return &output, nil
}
```

**Commit:** `git add internal/review/services/skim* && git commit -m "feat(review): add Skim Mode service"`

---

### Phase 2: Add Skim Models

**File:** `internal/review/models/review.go` (append to existing)
```go
// SkimModeOutput represents Skim Mode analysis
type SkimModeOutput struct {
	Functions   []FunctionSignature `json:"functions"`
	Interfaces  []InterfaceInfo     `json:"interfaces"`
	DataModels  []DataModelInfo     `json:"data_models"`
	Workflows   []WorkflowInfo      `json:"workflows"`
	Summary     string              `json:"summary"`
}

type FunctionSignature struct {
	Name        string `json:"name"`
	Signature   string `json:"signature"`
	Description string `json:"description"`
}

type InterfaceInfo struct {
	Name    string   `json:"name"`
	Methods []string `json:"methods"`
	Purpose string   `json:"purpose"`
}

type DataModelInfo struct {
	Name    string   `json:"name"`
	Fields  []string `json:"fields"`
	Purpose string   `json:"purpose"`
}

type WorkflowInfo struct {
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
}
```

**Commit:** `git add internal/review/models/ && git commit -m "feat(review): add Skim Mode data models"`

---

### Phase 3: Add Handler Endpoint

**File:** `cmd/review/handlers/review_handler.go` (add method)
```go
// GetSkimAnalysis handles GET /api/reviews/:id/skim
func (h *ReviewHandler) GetSkimAnalysis(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	review, err := h.reviewService.GetReview(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	output, err := h.skimService.AnalyzeSkim(c.Request.Context(), review.ID, "owner", "repo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}
```

**File:** `cmd/review/main.go` (add route)
```go
// In main(), add:
skimService := services.NewSkimService(ollamaClient, analysisRepo)
reviewHandler := handlers.NewReviewHandler(reviewService, previewService, skimService)

// Add route:
api.GET("/reviews/:id/skim", reviewHandler.GetSkimAnalysis)
```

**Commit:** `git add cmd/review/ && git commit -m "feat(review): add Skim Mode API endpoint"`

---

### Phase 4: Update ReviewService

**File:** `internal/review/services/review_service.go` (add field and method)
```go
type ReviewService struct {
	// ... existing fields
	skimService *SkimService
}

// Update constructor
func NewReviewService(..., skimService *SkimService) *ReviewService {
	return &ReviewService{
		// ... existing
		skimService: skimService,
	}
}
```

**Commit:** `git add internal/review/services/review_service.go && git commit -m "feat(review): wire Skim Mode into ReviewService"`

---

### Phase 5: Tests

**File:** `internal/review/services/skim_service_test.go`
```go
package services

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSkimService_AnalyzeSkim_Success(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewSkimService(mockOllama, mockRepo)

	mockRepo.On("FindByReviewAndMode", mock.Anything, int64(1), models.SkimMode).
		Return(nil, fmt.Errorf("not found"))

	aiResponse := `{"functions": [], "interfaces": [], "data_models": [], "workflows": [], "summary": "Test"}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	output, err := service.AnalyzeSkim(context.Background(), 1, "owner", "repo")

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "Test", output.Summary)
}
```

**Commit:** `git add internal/review/services/skim_service_test.go && git commit -m "test(review): add Skim Mode service tests"`

---

### Phase 6: Push & Test

```bash
git push -u origin feature/005-copilot-review-skim-mode
make test
```

**PR auto-created by GitHub Actions**

---

## References
- Issue #004 (Preview Mode) - follow same patterns
- `ARCHITECTURE.md` lines 746-776 (Skim Mode spec)

**Estimated Time:** 60-90 minutes
**Pattern:** Copy Preview Mode, change prompt and models
