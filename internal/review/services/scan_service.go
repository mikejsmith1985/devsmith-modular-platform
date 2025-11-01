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

// buildScanPrompt constructs the prompt for Scan Mode analysis.
func buildScanPrompt(query string) string {
	return fmt.Sprintf(`Find code related to: %q.\n\nReturn JSON:\n{\n  "matches": [\n    {"file": "path/to/file.go", "line": 42, "code_snippet": "...", "relevance": 0.95, "context": "Why this matches"}\n  ],\n  "summary": "Found X matches in Y files"\n}`,
		query)
}

// ScanService provides Scan Mode analysis for code review sessions.
// It integrates with Ollama for AI-powered code search and stores results in the analysis repository.
// All operations are logged with structured context for observability.
type ScanService struct {
	ollamaClient OllamaClientInterface
	analysisRepo AnalysisRepositoryInterface
	logger       logger.Interface
}

// NewScanService creates a new ScanService with the given dependencies and logger.
// ollamaClient: AI client for code search
// analysisRepo: Repository for persisting analysis results
// logger: Structured logger for observability
func NewScanService(ollamaClient OllamaClientInterface, analysisRepo AnalysisRepositoryInterface, logger logger.Interface) *ScanService {
	return &ScanService{ollamaClient: ollamaClient, analysisRepo: analysisRepo, logger: logger}
}

// AnalyzeScan performs Scan Mode analysis for the given review session and query.
// Returns a ScanModeOutput with matches and summary, or an error if analysis fails.
// Logs all major steps and errors with correlation ID for traceability.
func (s *ScanService) AnalyzeScan(ctx context.Context, reviewID int64, code string, query string) (*review_models.ScanModeOutput, error) {
	correlationID := ctx.Value(logger.CorrelationIDKey)
	s.logger.Info("AnalyzeScan called", "correlation_id", correlationID, "review_id", reviewID, "query", query, "code_length", len(code))

	if query == "" {
		s.logger.Warn("AnalyzeScan: empty query", "correlation_id", correlationID, "review_id", reviewID)
		return nil, errors.New("query cannot be empty")
	}
	
	// Build prompt using template
	prompt := BuildScanPrompt(code, query)

	start := time.Now()
	rawOutput, aiErr := s.ollamaClient.Generate(ctx, prompt)
	durationMs := time.Since(start).Milliseconds()
	if aiErr != nil {
		s.logger.Error("AI call failed", "correlation_id", correlationID, "review_id", reviewID, "duration_ms", durationMs, "error", aiErr)
		// Return fallback on error
		return s.getFallbackScanOutput(query), nil
	}
	s.logger.Info("AI call succeeded", "correlation_id", correlationID, "review_id", reviewID, "duration_ms", durationMs)

	var output review_models.ScanModeOutput
	unmarshalErr := json.Unmarshal([]byte(rawOutput), &output)
	if unmarshalErr != nil {
		s.logger.Error("Failed to unmarshal scan analysis output", "correlation_id", correlationID, "review_id", reviewID, "error", unmarshalErr)
		return s.getFallbackScanOutput(query), nil
	}

	metadataJSON, marshalErr := json.Marshal(output)
	if marshalErr != nil {
		s.logger.Error("Failed to marshal scan analysis output", "correlation_id", correlationID, "review_id", reviewID, "error", marshalErr)
		return nil, fmt.Errorf("scan analysis marshal error: %w", marshalErr)
	}

	result := &review_models.AnalysisResult{
		ReviewID:  reviewID,
		Mode:      review_models.ScanMode,
		Prompt:    prompt,
		RawOutput: rawOutput,
		Summary:   output.Summary,
		Metadata:  string(metadataJSON),
		ModelUsed: "mistral:7b-instruct",
	}
	saveErr := s.analysisRepo.Create(ctx, result)
	if saveErr != nil {
		s.logger.Error("Failed to save scan analysis result", "correlation_id", correlationID, "review_id", reviewID, "error", saveErr)
		return nil, fmt.Errorf("scan analysis save error: %w", saveErr)
	}

	s.logger.Info("AnalyzeScan completed", "correlation_id", correlationID, "review_id", reviewID, "summary", output.Summary)
	return &output, nil
}

// getFallbackScanOutput returns safe fallback data when Ollama fails
func (s *ScanService) getFallbackScanOutput(query string) *review_models.ScanModeOutput {
	return &review_models.ScanModeOutput{
		Matches: []review_models.CodeMatch{},
		Summary: fmt.Sprintf("Analysis unavailable for query: %s", query),
	}
}
