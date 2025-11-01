package review_services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
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

// AnalyzeSkim performs Skim Mode analysis for the given review session and code.
func (s *SkimService) AnalyzeSkim(ctx context.Context, reviewID int64, code string) (*review_models.SkimModeOutput, error) {
	correlationID := ctx.Value(logger.CorrelationIDKey)
	s.logger.Info("AnalyzeSkim called", "correlation_id", correlationID, "review_id", reviewID, "code_length", len(code))

	// Build prompt using template
	prompt := BuildSkimPrompt(code)

	start := time.Now()
	rawOutput, err := s.ollamaClient.Generate(ctx, prompt)
	duration := time.Since(start)
	if err != nil {
		s.logger.Error("SkimService: AI call failed", "correlation_id", correlationID, "review_id", reviewID, "error", err, "duration_ms", duration.Milliseconds())
		// Return fallback on error
		return s.getFallbackSkimOutput(), nil
	}
	s.logger.Info("SkimService: AI call succeeded", "correlation_id", correlationID, "review_id", reviewID, "duration_ms", duration.Milliseconds())

	output, err := s.parseSkimOutput(rawOutput)
	if err != nil {
		s.logger.Error("SkimService: failed to parse AI output", "correlation_id", correlationID, "review_id", reviewID, "error", err)
		return s.getFallbackSkimOutput(), nil
	}

	metadataJSON, err := json.Marshal(output)
	if err != nil {
		s.logger.Error("SkimService: failed to marshal output", "correlation_id", correlationID, "review_id", reviewID, "error", err)
		return nil, fmt.Errorf("failed to marshal skim analysis output: %w", err)
	}
	result := &review_models.AnalysisResult{
		ReviewID:  reviewID,
		Mode:      review_models.SkimMode,
		Prompt:    prompt,
		RawOutput: rawOutput,
		Summary:   output.Summary,
		Metadata:  string(metadataJSON),
		ModelUsed: "mistral:7b-instruct",
	}
	if err := s.analysisRepo.Create(ctx, result); err != nil {
		s.logger.Error("SkimService: failed to save analysis result", "correlation_id", correlationID, "review_id", reviewID, "error", err)
		return nil, fmt.Errorf("failed to save skim analysis result: %w", err)
	}
	s.logger.Info("SkimService: analysis completed and saved", "correlation_id", correlationID, "review_id", reviewID)
	return output, nil
}

// getFallbackSkimOutput returns safe fallback data when Ollama fails
func (s *SkimService) getFallbackSkimOutput() *review_models.SkimModeOutput {
	return &review_models.SkimModeOutput{
		Functions:  []review_models.FunctionSignature{},
		Interfaces: []review_models.InterfaceInfo{},
		DataModels: []review_models.DataModelInfo{},
		Workflows:  []review_models.WorkflowInfo{},
		Summary:    "Analysis unavailable - using mock data",
	}
}

func (s *SkimService) buildSkimPrompt(owner, repo string) string {
	// Deprecated - use BuildSkimPrompt instead
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
func (s *SkimService) parseSkimOutput(raw string) (*review_models.SkimModeOutput, error) {
	var output review_models.SkimModeOutput
	if err := json.Unmarshal([]byte(raw), &output); err != nil {
		return nil, fmt.Errorf("failed to parse skim output: %w", err)
	}
	return &output, nil
}
