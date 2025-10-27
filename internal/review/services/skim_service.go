package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// SkimService provides Skim Mode analysis for code review sessions.
type SkimService struct {
	ollamaClient OllamaClientInterface
	analysisRepo AnalysisRepositoryInterface
	logger       logger.Interface
}

// NewSkimService creates a new SkimService with the given dependencies.
func NewSkimService(ollamaClient OllamaClientInterface, analysisRepo AnalysisRepositoryInterface, logger logger.Interface) *SkimService {
	return &SkimService{
		ollamaClient: ollamaClient,
		analysisRepo: analysisRepo,
		logger:       logger,
	}
}

// AnalyzeSkim performs Skim Mode analysis for the given review session and repository.
func (s *SkimService) AnalyzeSkim(ctx context.Context, reviewID int64, repoOwner, repoName string) (*models.SkimModeOutput, error) {
	correlationID := ctx.Value(logger.CorrelationIDKey)
	s.logger.Info("AnalyzeSkim called", "correlation_id", correlationID, "review_id", reviewID, "repo_owner", repoOwner, "repo_name", repoName)

	existing, err := s.analysisRepo.FindByReviewAndMode(ctx, reviewID, models.SkimMode)
	if err != nil && err.Error() != "not found" {
		s.logger.Error("SkimService: cache lookup failed", "correlation_id", correlationID, "review_id", reviewID, "error", err)
		return nil, fmt.Errorf("cache lookup failed: %w", err)
	}
	if existing != nil {
		var output models.SkimModeOutput
		if unmarshalErr := json.Unmarshal([]byte(existing.Metadata), &output); unmarshalErr != nil {
			s.logger.Error("SkimService: failed to unmarshal existing metadata", "correlation_id", correlationID, "review_id", reviewID, "error", unmarshalErr)
			return nil, fmt.Errorf("failed to unmarshal existing metadata: %w", unmarshalErr)
		}
		s.logger.Info("SkimService: cache hit", "correlation_id", correlationID, "review_id", reviewID)
		return &output, nil
	}

	prompt := s.buildSkimPrompt(repoOwner, repoName)
	start := time.Now()
	rawOutput, err := s.ollamaClient.Generate(ctx, prompt)
	duration := time.Since(start)
	if err != nil {
		s.logger.Error("SkimService: AI call failed", "correlation_id", correlationID, "review_id", reviewID, "error", err, "duration_ms", duration.Milliseconds())
		return nil, err
	}
	s.logger.Info("SkimService: AI call succeeded", "correlation_id", correlationID, "review_id", reviewID, "duration_ms", duration.Milliseconds())

	output, err := s.parseSkimOutput(rawOutput)
	if err != nil {
		s.logger.Error("SkimService: failed to parse AI output", "correlation_id", correlationID, "review_id", reviewID, "error", err)
		return nil, err
	}

	metadataJSON, err := json.Marshal(output)
	if err != nil {
		s.logger.Error("SkimService: failed to marshal output", "correlation_id", correlationID, "review_id", reviewID, "error", err)
		return nil, fmt.Errorf("failed to marshal skim analysis output: %w", err)
	}
	result := &models.AnalysisResult{
		ReviewID:  reviewID,
		Mode:      models.SkimMode,
		Prompt:    prompt,
		RawOutput: rawOutput,
		Summary:   output.Summary,
		Metadata:  string(metadataJSON),
		ModelUsed: "qwen2.5-coder:32b",
	}
	if err := s.analysisRepo.Create(ctx, result); err != nil {
		s.logger.Error("SkimService: failed to save analysis result", "correlation_id", correlationID, "review_id", reviewID, "error", err)
		return nil, fmt.Errorf("failed to save skim analysis result: %w", err)
	}
	s.logger.Info("SkimService: analysis completed and saved", "correlation_id", correlationID, "review_id", reviewID)
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

// Fix parseSkimOutput to handle errors properly
func (s *SkimService) parseSkimOutput(raw string) (*models.SkimModeOutput, error) {
	var output models.SkimModeOutput
	if err := json.Unmarshal([]byte(raw), &output); err != nil {
		return nil, fmt.Errorf("failed to parse skim output: %w", err)
	}
	return &output, nil
}
