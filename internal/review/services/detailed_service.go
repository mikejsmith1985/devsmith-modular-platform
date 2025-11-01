// Package review_services contains business logic for review service reading modes, including Detailed Mode.
package review_services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// DetailedService provides line-by-line code analysis for Detailed Mode.
// It identifies code complexity, side effects, and data flow between elements.
type DetailedService struct {
	ollamaClient OllamaClientInterface
	analysisRepo AnalysisRepositoryInterface
	logger       logger.Interface
}

// NewDetailedService creates a new DetailedService with the given Ollama client and analysis repository.
func NewDetailedService(ollama OllamaClientInterface, repo AnalysisRepositoryInterface, logger logger.Interface) *DetailedService {
	return &DetailedService{
		ollamaClient: ollama,
		analysisRepo: repo,
		logger:       logger,
	}
}

// DetailedLine represents a single line of code and its analysis in Detailed Mode.
// It includes the line number, code snippet, explanation, complexity, and side effects.
type DetailedLine struct {
	Code              string   `json:"code"`
	Explanation       string   `json:"explanation"`
	Complexity        string   `json:"complexity"`
	SideEffects       []string `json:"side_effects"`
	VariablesModified []string `json:"variables_modified"`
	LineNum           int      `json:"line_num"`
}

// DataFlow describes the flow of data between code elements in Detailed Mode.
// It includes the source, destination, and a description of the data flow.
type DataFlow struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Description string `json:"description"`
}

// DetailedAnalysisOutput is the result of a Detailed Mode analysis.
// It includes line-by-line explanations, data flow, and a summary.
type DetailedAnalysisOutput struct {
	Summary  string         `json:"summary"`
	Lines    []DetailedLine `json:"lines"`
	DataFlow []DataFlow     `json:"data_flow"`
}

// AnalyzeDetailed performs a line-by-line analysis of code in Detailed Mode.
// It generates a detailed report and stores the result in the analysis repository.
func (s *DetailedService) AnalyzeDetailed(ctx context.Context, reviewID int64, code string, filename string) (*DetailedAnalysisOutput, error) {
	correlationID := ctx.Value(logger.CorrelationIDKey)
	s.logger.Info("AnalyzeDetailed called", "correlation_id", correlationID, "review_id", reviewID, "filename", filename, "code_length", len(code))
	
	if code == "" {
		s.logger.Error("DetailedService: code empty", "correlation_id", correlationID, "review_id", reviewID)
		return nil, errors.New("code cannot be empty")
	}
	
	// Build prompt using template
	prompt := BuildDetailedPrompt(code, filename)
	
	start := time.Now()
	resp, err := s.ollamaClient.Generate(ctx, prompt)
	duration := time.Since(start)
	if err != nil {
		s.logger.Error("DetailedService: AI call failed", "correlation_id", correlationID, "review_id", reviewID, "error", err, "duration_ms", duration.Milliseconds())
		// Return fallback on error
		return s.getFallbackDetailedOutput(), nil
	}
	s.logger.Info("DetailedService: AI call succeeded", "correlation_id", correlationID, "review_id", reviewID, "duration_ms", duration.Milliseconds())
	
	var output DetailedAnalysisOutput
	if err := json.Unmarshal([]byte(resp), &output); err != nil {
		s.logger.Error("DetailedService: failed to unmarshal output", "correlation_id", correlationID, "review_id", reviewID, "error", err)
		return s.getFallbackDetailedOutput(), nil
	}
	
	metadataJSON, marshalErr := json.Marshal(output)
	if marshalErr != nil {
		s.logger.Error("DetailedService: failed to marshal output", "correlation_id", correlationID, "review_id", reviewID, "error", marshalErr)
		return nil, fmt.Errorf("failed to marshal detailed analysis output: %w", marshalErr)
	}
	
	result := &review_models.AnalysisResult{
		ReviewID:  reviewID,
		Mode:      review_models.DetailedMode,
		Prompt:    prompt,
		RawOutput: resp,
		Summary:   output.Summary,
		Metadata:  string(metadataJSON),
		ModelUsed: "mistral:7b-instruct",
	}
	if err := s.analysisRepo.Create(ctx, result); err != nil {
		s.logger.Error("DetailedService: failed to save analysis result", "correlation_id", correlationID, "review_id", reviewID, "error", err)
		return nil, fmt.Errorf("failed to create analysis result: %w", err)
	}
	s.logger.Info("DetailedService: analysis completed and saved", "correlation_id", correlationID, "review_id", reviewID)
	return &output, nil
}

// getFallbackDetailedOutput returns safe fallback data when Ollama fails
func (s *DetailedService) getFallbackDetailedOutput() *DetailedAnalysisOutput {
	return &DetailedAnalysisOutput{
		Summary:  "Analysis unavailable - detailed line-by-line analysis not currently available",
		Lines:    []DetailedLine{},
		DataFlow: []DataFlow{},
	}
}
