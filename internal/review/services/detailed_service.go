// Package services contains business logic for review service reading modes, including Detailed Mode.
package services

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// DetailedService provides line-by-line code analysis for Detailed Mode.
type DetailedService struct {
	ollamaClient OllamaClientInterface
	analysisRepo AnalysisRepositoryInterface
}

// NewDetailedService creates a new DetailedService with the given Ollama client and analysis repository.
func NewDetailedService(ollama OllamaClientInterface, repo AnalysisRepositoryInterface) *DetailedService {
	return &DetailedService{
		ollamaClient: ollama,
		analysisRepo: repo,
	}
}

// DetailedLine represents a single line of code and its analysis in Detailed Mode.
type DetailedLine struct {
	LineNum           int      `json:"line_num"`
	Code              string   `json:"code"`
	Explanation       string   `json:"explanation"`
	Complexity        string   `json:"complexity"`
	SideEffects       []string `json:"side_effects"`
	VariablesModified []string `json:"variables_modified"`
}

// DataFlow describes the flow of data between code elements in Detailed Mode.
type DataFlow struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Description string `json:"description"`
}

// DetailedAnalysisOutput is the result of a Detailed Mode analysis, including line explanations and data flow.
type DetailedAnalysisOutput struct {
	Lines    []DetailedLine `json:"lines"`
	DataFlow []DataFlow     `json:"data_flow"`
	Summary  string         `json:"summary"`
}

// AnalyzeDetailed performs a line-by-line analysis of the specified file in Detailed Mode.
func (s *DetailedService) AnalyzeDetailed(ctx context.Context, sessionID int, filePath, _, _ string) (*DetailedAnalysisOutput, error) {
	if filePath == "" {
		return nil, errors.New("file path cannot be empty")
	}
	// Construct prompt (simplified for now)
	prompt := "Analyze file in detailed mode: " + filePath
	resp, err := s.ollamaClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}
	var output DetailedAnalysisOutput
	err = json.Unmarshal([]byte(resp), &output)
	if err != nil {
		return nil, err
	}
	// Store result in repository
	result := &models.AnalysisResult{
		ReviewID:  int64(sessionID),
		Mode:      "detailed",
		Prompt:    prompt,
		RawOutput: resp,
		Summary:   output.Summary,
		Metadata:  "",
		ModelUsed: "ollama",
	}
	_ = s.analysisRepo.Create(ctx, result)
	return &output, nil
}
