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

// AnalyzeDetailed performs a line-by-line analysis of the specified file in Detailed Mode.
// It generates a detailed report and stores the result in the analysis repository.
func (s *DetailedService) AnalyzeDetailed(ctx context.Context, sessionID int, filePath string) (*DetailedAnalysisOutput, error) {
	correlationID := ctx.Value(logger.CorrelationIDKey)
	s.logger.Info("AnalyzeDetailed called", "correlation_id", correlationID, "session_id", sessionID, "file_path", filePath)
	if filePath == "" {
		s.logger.Error("DetailedService: file path empty", "correlation_id", correlationID, "session_id", sessionID)
		return nil, errors.New("file path cannot be empty")
	}
	prompt := "Analyze file in detailed mode: " + filePath
	start := time.Now()
	resp, err := s.ollamaClient.Generate(ctx, prompt)
	duration := time.Since(start)
	if err != nil {
		s.logger.Error("DetailedService: AI call failed", "correlation_id", correlationID, "session_id", sessionID, "error", err, "duration_ms", duration.Milliseconds())
		return nil, err
	}
	s.logger.Info("DetailedService: AI call succeeded", "correlation_id", correlationID, "session_id", sessionID, "duration_ms", duration.Milliseconds())
	var output DetailedAnalysisOutput
	if err := json.Unmarshal([]byte(resp), &output); err != nil {
		s.logger.Error("DetailedService: failed to unmarshal output", "correlation_id", correlationID, "session_id", sessionID, "error", err)
		return nil, fmt.Errorf("failed to unmarshal detailed analysis output: %w", err)
	}
	result := &review_models.AnalysisResult{
		ReviewID:  int64(sessionID),
		Mode:      "detailed",
		Prompt:    prompt,
		RawOutput: resp,
		Summary:   output.Summary,
		Metadata:  "",
		ModelUsed: "ollama",
	}
	if err := s.analysisRepo.Create(ctx, result); err != nil {
		s.logger.Error("DetailedService: failed to save analysis result", "correlation_id", correlationID, "session_id", sessionID, "error", err)
		return nil, fmt.Errorf("failed to create analysis result: %w", err)
	}
	s.logger.Info("DetailedService: analysis completed and saved", "correlation_id", correlationID, "session_id", sessionID)
	return &output, nil
}
