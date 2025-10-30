package review_services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// CriticalService provides methods for analyzing repositories in Critical Mode.
// It identifies issues such as security vulnerabilities, bugs, performance problems, and code smells.
type CriticalService struct {
	ollamaClient OllamaClientInterface
	analysisRepo AnalysisRepositoryInterface
	logger       logger.Interface
}

// NewCriticalService creates a new instance of CriticalService with the provided dependencies.
func NewCriticalService(ollamaClient OllamaClientInterface, analysisRepo AnalysisRepositoryInterface, logger logger.Interface) *CriticalService {
	return &CriticalService{ollamaClient: ollamaClient, analysisRepo: analysisRepo, logger: logger}
}

// AnalyzeCritical performs a detailed analysis of a repository in Critical Mode.
// It generates a report identifying various issues and returns the analysis output.
func (s *CriticalService) AnalyzeCritical(ctx context.Context, reviewID int64, repoOwner, repoName string) (*review_models.CriticalModeOutput, error) {
	correlationID := ctx.Value(logger.CorrelationIDKey)
	s.logger.Info("AnalyzeCritical called", "correlation_id", correlationID, "review_id", reviewID, "repo_owner", repoOwner, "repo_name", repoName)

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

	start := time.Now()
	rawOutput, err := s.ollamaClient.Generate(ctx, prompt)
	duration := time.Since(start)
	if err != nil {
		s.logger.Error("Critical analysis AI call failed", "correlation_id", correlationID, "review_id", reviewID, "error", err, "duration_ms", duration.Milliseconds())
		return nil, fmt.Errorf("failed to generate critical analysis: %w", err)
	}
	s.logger.Info("Critical analysis AI call succeeded", "correlation_id", correlationID, "review_id", reviewID, "duration_ms", duration.Milliseconds())

	var output review_models.CriticalModeOutput
	if unmarshalErr := json.Unmarshal([]byte(rawOutput), &output); unmarshalErr != nil {
		s.logger.Error("Failed to unmarshal critical analysis output", "correlation_id", correlationID, "review_id", reviewID, "error", unmarshalErr)
		return nil, fmt.Errorf("failed to unmarshal critical analysis output: %w", unmarshalErr)
	}

	metadataJSON, err := json.Marshal(output)
	if err != nil {
		s.logger.Error("Failed to marshal critical analysis output", "correlation_id", correlationID, "review_id", reviewID, "error", err)
		return nil, fmt.Errorf("failed to marshal critical analysis output: %w", err)
	}
	result := &review_models.AnalysisResult{
		ReviewID:  reviewID,
		Mode:      review_models.CriticalMode,
		Prompt:    prompt,
		RawOutput: rawOutput,
		Summary:   output.Summary,
		Metadata:  string(metadataJSON),
		ModelUsed: "qwen2.5-coder:32b",
	}
	if err := s.analysisRepo.Create(ctx, result); err != nil {
		s.logger.Error("Failed to save critical analysis result", "correlation_id", correlationID, "review_id", reviewID, "error", err)
		return nil, fmt.Errorf("failed to save analysis result: %w", err)
	}
	s.logger.Info("Critical analysis completed and saved", "correlation_id", correlationID, "review_id", reviewID)
	return &output, nil
}

// OllamaClient represents the AI client used for generating analysis.
// It provides methods to interact with the AI model.
type OllamaClient struct{}

// Generate simulates AI generation for the given prompt.
// It returns the generated output or an error if the operation fails.
func (o *OllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	// Simulated implementation
	return "", nil
}
