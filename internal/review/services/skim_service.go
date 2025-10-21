package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// SkimService provides Skim Mode analysis for code review sessions.
type SkimService struct {
	ollamaClient OllamaClientInterface
	analysisRepo AnalysisRepositoryInterface
}

// NewSkimService creates a new SkimService with the given dependencies.
func NewSkimService(ollamaClient OllamaClientInterface, analysisRepo AnalysisRepositoryInterface) *SkimService {
	return &SkimService{
		ollamaClient: ollamaClient,
		analysisRepo: analysisRepo,
	}
}

// AnalyzeSkim performs Skim Mode analysis for the given review session and repository.
func (s *SkimService) AnalyzeSkim(ctx context.Context, reviewID int64, repoOwner, repoName string) (*models.SkimModeOutput, error) {
	// Check cache
	existing, err := s.analysisRepo.FindByReviewAndMode(ctx, reviewID, models.SkimMode)
	// Debugging: Log cache lookup result
	// fmt.Printf("Cache lookup result: %+v, error: %v\n", existing, err)
	if err != nil && err.Error() != "not found" {
		return nil, fmt.Errorf("cache lookup failed: %w", err)
	}
	if existing != nil {
		var output models.SkimModeOutput
		// Properly handle errors for Unmarshal
		if unmarshalErr := json.Unmarshal([]byte(existing.Metadata), &output); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal existing metadata: %w", unmarshalErr)
		}
		return &output, nil
	}

	// Generate prompt
	prompt := s.buildSkimPrompt(repoOwner, repoName)

	// Call AI
	rawOutput, err := s.ollamaClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Debugging: Log raw output from AI
	// fmt.Printf("Raw AI output: %s\n", rawOutput)

	// Parse response
	output, err := s.parseSkimOutput(rawOutput)
	if err != nil {
		return nil, err
	}

	// Debugging: Log parsed output
	// fmt.Printf("Parsed SkimModeOutput: %+v\n", output)

	// Store in DB
	metadataJSON, err := json.Marshal(output)
	if err != nil {
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
	// Ensure the result is saved and handle errors
	if err := s.analysisRepo.Create(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to save skim analysis result: %w", err)
	}

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
