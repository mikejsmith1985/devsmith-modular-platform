package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

type CriticalService struct {
	ollamaClient OllamaClientInterface
	analysisRepo AnalysisRepositoryInterface
}

func NewCriticalService(ollamaClient OllamaClientInterface, analysisRepo AnalysisRepositoryInterface) *CriticalService {
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

// OllamaClient represents the AI client used for generating analysis.
type OllamaClient struct{}

// Generate simulates AI generation for the given prompt.
func (o *OllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	// Simulated implementation
	return "", nil
}
