package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// ScanService provides Scan Mode analysis for code review sessions.
type ScanService struct {
	ollamaClient OllamaClientInterface
	analysisRepo AnalysisRepositoryInterface
}

// NewScanService creates a new ScanService with the given dependencies.
func NewScanService(ollamaClient OllamaClientInterface, analysisRepo AnalysisRepositoryInterface) *ScanService {
	return &ScanService{ollamaClient, analysisRepo}
}

// AnalyzeScan performs Scan Mode analysis for the given review session and query.
func (s *ScanService) AnalyzeScan(ctx context.Context, reviewID int64, query string, repoOwner, repoName string) (*models.ScanModeOutput, error) {
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}
	prompt := fmt.Sprintf(`Find code related to: "%s" in repository %s/%s.\n\nReturn JSON:\n{\n  "matches": [\n    {"file": "path/to/file.go", "line": 42, "code_snippet": "...", "relevance": 0.95, "context": "Why this matches"}\n  ],\n  "summary": "Found X matches in Y files"\n}`,
		query, repoOwner, repoName)

	rawOutput, err := s.ollamaClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}
	var output models.ScanModeOutput
	if err := json.Unmarshal([]byte(rawOutput), &output); err != nil {
		return nil, err
	}

	metadataJSON, _ := json.Marshal(output)
	result := &models.AnalysisResult{
		ReviewID:  reviewID,
		Mode:      models.ScanMode,
		Prompt:    prompt,
		RawOutput: rawOutput,
		Summary:   output.Summary,
		Metadata:  string(metadataJSON),
		ModelUsed: "qwen2.5-coder:32b",
	}
	s.analysisRepo.Create(ctx, result)

	return &output, nil
}
