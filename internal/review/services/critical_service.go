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
func (s *CriticalService) AnalyzeCritical(ctx context.Context, reviewID int64, code string) (*review_models.CriticalModeOutput, error) {
	correlationID := ctx.Value(logger.CorrelationIDKey)
	s.logger.Info("AnalyzeCritical called", "correlation_id", correlationID, "review_id", reviewID, "code_length", len(code))

	// Build prompt using template
	prompt := BuildCriticalPrompt(code)

	// Call Ollama for real analysis
	start := time.Now()
	rawOutput, err := s.ollamaClient.Generate(ctx, prompt)
	duration := time.Since(start)
	
	if err != nil {
		s.logger.Error("Critical analysis AI call failed", "correlation_id", correlationID, "review_id", reviewID, "error", err, "duration_ms", duration.Milliseconds())
		// Fallback to mock response on error
		return s.getFallbackCriticalOutput()
	}
	s.logger.Info("Critical analysis AI call succeeded", "correlation_id", correlationID, "review_id", reviewID, "duration_ms", duration.Milliseconds(), "output_length", len(rawOutput))

	// Parse JSON response
	var output review_models.CriticalModeOutput
	if unmarshalErr := json.Unmarshal([]byte(rawOutput), &output); unmarshalErr != nil {
		s.logger.Error("Failed to unmarshal critical analysis output", "correlation_id", correlationID, "review_id", reviewID, "error", unmarshalErr)
		// Fallback on JSON parsing error
		return s.getFallbackCriticalOutput()
	}

	// Validate output structure
	if output.Summary == "" {
		s.logger.Warn("Critical analysis returned empty summary", "correlation_id", correlationID, "review_id", reviewID)
		output.Summary = "Analysis completed but summary was empty"
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
		ModelUsed: "mistral:7b-instruct",
	}
	
	if err := s.analysisRepo.Create(ctx, result); err != nil {
		s.logger.Error("Failed to save critical analysis result", "correlation_id", correlationID, "review_id", reviewID, "error", err)
		return nil, fmt.Errorf("failed to save analysis result: %w", err)
	}
	s.logger.Info("Critical analysis completed and saved", "correlation_id", correlationID, "review_id", reviewID)
	return &output, nil
}

// getFallbackCriticalOutput returns a safe fallback response when Ollama fails
func (s *CriticalService) getFallbackCriticalOutput() (*review_models.CriticalModeOutput, error) {
	return &review_models.CriticalModeOutput{
		Issues: []review_models.CodeIssue{
			{
				Severity:      "info",
				Category:      "quality",
				Description:   "Unable to perform AI analysis at this time. Please try again later.",
				FixSuggestion: "Ensure Ollama service is running on localhost:11434",
			},
		},
		OverallGrade: "N/A",
		Summary:      "Analysis unavailable - Ollama service error",
	}, nil
}
