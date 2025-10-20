package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)


type SkimService struct {
	ollamaClient OllamaClientInterface
	analysisRepo AnalysisRepositoryInterface
}

func NewSkimService(ollamaClient OllamaClientInterface, analysisRepo AnalysisRepositoryInterface) *SkimService {
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
		ReviewID:  reviewID,
		Mode:      models.SkimMode,
		Prompt:    prompt,
		RawOutput: rawOutput,
		Summary:   output.Summary,
		Metadata:  string(metadataJSON),
		ModelUsed: "qwen2.5-coder:32b",
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
