package review_services

import (
	"context"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// OllamaClientStub is a stub implementation of OllamaClientInterface for local dev/testing.
type OllamaClientStub struct{}

// Generate returns a stubbed AI output for testing.
func (o *OllamaClientStub) Generate(ctx context.Context, prompt string) (string, error) {
	return `{"functions":[],"interfaces":[],"data_models":[],"workflows":[],"summary":"Stubbed AI output"}`, nil
}

// OllamaClientInterface defines the AI client contract
// Accepts context and prompt, returns raw AI output or error
// Enables swapping AI providers and mocking for tests
// Used by ScanService, SkimService, etc.
type OllamaClientInterface interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// AnalysisRepositoryInterface defines the analysis repo contract
// Used for storing and retrieving analysis results
// Enables mocking and swapping DB implementations
type AnalysisRepositoryInterface interface {
	FindByReviewAndMode(ctx context.Context, reviewID int64, mode string) (*review_models.AnalysisResult, error)
	Create(ctx context.Context, result *review_models.AnalysisResult) error
}
