# Issue #006: [COPILOT] Review Service - Scan Mode

**Labels:** `copilot`, `review`, `reading-mode`
**Created:** 2025-10-19
**Issue:** #6
**Estimated Time:** 60-90 minutes
**Depends On:** Issue #005 (Skim Mode)

---

# 🚨 STEP 0: CREATE FEATURE BRANCH FIRST 🚨

```bash
git checkout development && git pull origin development
git checkout -b feature/006-copilot-review-scan-mode
git branch --show-current  # Verify
```

---

## Task Description

Implement Scan Mode - targeted search. User provides search query, AI finds specific code patterns, functions, or implementations. Like Ctrl+F but semantic.

---

## Success Criteria
- [ ] User provides search query (e.g., "authentication logic")
- [ ] AI returns relevant code locations with context
- [ ] Results ranked by relevance
- [ ] Shows surrounding code context (±5 lines)
- [ ] 70%+ test coverage

---

## Implementation

### Phase 1: Scan Service

**File:** `internal/review/services/scan_service.go`
```go
package services

type ScanService struct {
	ollamaClient *OllamaClient
	analysisRepo AnalysisRepositoryInterface
}

func NewScanService(ollamaClient *OllamaClient, analysisRepo AnalysisRepositoryInterface) *ScanService {
	return &ScanService{ollamaClient, analysisRepo}
}

func (s *ScanService) AnalyzeScan(ctx context.Context, reviewID int64, query string, repoOwner, repoName string) (*models.ScanModeOutput, error) {
	prompt := fmt.Sprintf(`Find code related to: "%s" in repository %s/%s.

Return JSON:
{
  "matches": [
    {"file": "path/to/file.go", "line": 42, "code_snippet": "...", "relevance": 0.95, "context": "Why this matches"}
  ],
  "summary": "Found X matches in Y files"
}`, query, repoOwner, repoName)

	rawOutput, _ := s.ollamaClient.Generate(ctx, prompt)
	var output models.ScanModeOutput
	json.Unmarshal([]byte(rawOutput), &output)

	// Store result
	metadataJSON, _ := json.Marshal(output)
	result := &models.AnalysisResult{
		ReviewID:  reviewID,
		Mode:      models.ScanMode,
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

**Commit:** `git add internal/review/services/scan* && git commit -m "feat(review): add Scan Mode service"`

---

### Phase 2: Models

**Append to:** `internal/review/models/review.go`
```go
type ScanModeOutput struct {
	Matches []CodeMatch `json:"matches"`
	Summary string      `json:"summary"`
}

type CodeMatch struct {
	File        string  `json:"file"`
	Line        int     `json:"line"`
	CodeSnippet string  `json:"code_snippet"`
	Relevance   float64 `json:"relevance"`
	Context     string  `json:"context"`
}
```

**Commit:** `git add internal/review/models/ && git commit -m "feat(review): add Scan Mode models"`

---

### Phase 3: Handler & Route

**File:** `cmd/review/handlers/review_handler.go`
```go
func (h *ReviewHandler) GetScanAnalysis(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	query := c.Query("q")  // GET /api/reviews/:id/scan?q=authentication

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter required"})
		return
	}

	review, _ := h.reviewService.GetReview(c.Request.Context(), id)
	output, _ := h.scanService.AnalyzeScan(c.Request.Context(), review.ID, query, "owner", "repo")

	c.JSON(http.StatusOK, output)
}
```

**Add route in main.go:** `api.GET("/reviews/:id/scan", reviewHandler.GetScanAnalysis)`

**Commit:** `git add cmd/review/ && git commit -m "feat(review): add Scan Mode endpoint"`

---

### Phase 4: Tests

**File:** `internal/review/services/scan_service_test.go`
```go
func TestScanService_AnalyzeScan_Success(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewScanService(mockOllama, mockRepo)

	aiResponse := `{"matches": [{"file": "auth.go", "line": 10, "code_snippet": "func Login()", "relevance": 0.9, "context": "Main auth"}], "summary": "Found 1 match"}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	output, err := service.AnalyzeScan(context.Background(), 1, "authentication", "owner", "repo")

	assert.NoError(t, err)
	assert.Len(t, output.Matches, 1)
	assert.Equal(t, "auth.go", output.Matches[0].File)
}
```

**Commit:** `git add internal/review/services/scan* && git commit -m "test(review): add Scan Mode tests"`

---

### Phase 5: Push

```bash
git push -u origin feature/006-copilot-review-scan-mode
```

**PR auto-created**

---

## References
- `ARCHITECTURE.md` lines 778-812 (Scan Mode spec)

**Time:** 60-90 minutes
