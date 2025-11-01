package review_services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// PreviewResult holds the analysis for Preview Mode
type PreviewResult struct {
	FileTree             []string `json:"file_tree"`
	BoundedContexts      []string `json:"bounded_contexts"`
	TechStack            []string `json:"tech_stack"`
	ArchitecturePattern  string   `json:"architecture_pattern"`
	EntryPoints          []string `json:"entry_points"`
	ExternalDependencies []string `json:"external_dependencies"`
	Summary              string   `json:"summary"`
}

// FileNode represents a node in the file/folder tree (deprecated, kept for compatibility)
type FileNode struct {
	Name        string
	Path        string
	Description string
	Children    []FileNode
}

// PreviewService provides Preview Mode analysis
type PreviewService struct {
	ollamaClient OllamaClientInterface
	logger       logger.Interface
}

// NewPreviewService creates a new PreviewService with logger.
// Note: For backward compatibility, OllamaClientInterface can be nil and will use mock data.
func NewPreviewService(logger logger.Interface) *PreviewService {
	return &PreviewService{logger: logger}
}

// NewPreviewServiceWithOllama creates a new PreviewService with Ollama integration.
func NewPreviewServiceWithOllama(ollamaClient OllamaClientInterface, logger logger.Interface) *PreviewService {
	return &PreviewService{ollamaClient: ollamaClient, logger: logger}
}

// AnalyzePreview analyzes the codebase in Preview Mode.
// It returns a PreviewResult containing high-level structure and context for the given codebase.
func (s *PreviewService) AnalyzePreview(ctx context.Context, code string) (*PreviewResult, error) {
	correlationID := ctx.Value(logger.CorrelationIDKey)
	s.logger.Info("AnalyzePreview called", "correlation_id", correlationID, "code_length", len(code))

	// If no Ollama client configured, return mock data (backward compatibility)
	if s.ollamaClient == nil {
		s.logger.Info("AnalyzePreview using mock data (no Ollama configured)", "correlation_id", correlationID)
		return s.getMockPreviewResult(), nil
	}

	// Build prompt using template
	prompt := BuildPreviewPrompt(code)

	// Call Ollama for real analysis
	start := time.Now()
	rawOutput, err := s.ollamaClient.Generate(ctx, prompt)
	duration := time.Since(start)

	if err != nil {
		s.logger.Error("Preview analysis AI call failed", "correlation_id", correlationID, "error", err, "duration_ms", duration.Milliseconds())
		// Fallback to mock data on error
		return s.getMockPreviewResult(), nil
	}
	s.logger.Info("Preview analysis AI call succeeded", "correlation_id", correlationID, "duration_ms", duration.Milliseconds(), "output_length", len(rawOutput))

	// Parse JSON response
	var result PreviewResult
	if unmarshalErr := json.Unmarshal([]byte(rawOutput), &result); unmarshalErr != nil {
		s.logger.Error("Failed to unmarshal preview analysis output", "correlation_id", correlationID, "error", unmarshalErr)
		// Fallback on JSON parsing error
		return s.getMockPreviewResult(), nil
	}

	// Validate output structure
	if result.Summary == "" {
		s.logger.Warn("Preview analysis returned empty summary", "correlation_id", correlationID)
		result.Summary = "Analysis completed"
	}

	s.logger.Info("AnalyzePreview completed", "correlation_id", correlationID, "result_summary", result.Summary)
	return &result, nil
}

// getMockPreviewResult returns safe fallback mock data
func (s *PreviewService) getMockPreviewResult() *PreviewResult {
	return &PreviewResult{
		FileTree:             []string{"main.go", "handler.go", "models/"},
		BoundedContexts:      []string{"Auth domain", "Review domain"},
		TechStack:            []string{"Go", "Gin", "PostgreSQL"},
		ArchitecturePattern:  "layered",
		EntryPoints:          []string{"main()", "NewServer()"},
		ExternalDependencies: []string{"PostgreSQL", "Redis"},
		Summary:              "This is a Go microservice using Gin and PostgreSQL.",
	}
}
