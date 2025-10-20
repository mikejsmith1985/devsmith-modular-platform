package services

import (
	"context"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

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
	FindByReviewAndMode(ctx context.Context, reviewID int64, mode string) (*models.AnalysisResult, error)
	Create(ctx context.Context, result *models.AnalysisResult) error
}
