// Package testutils provides testing utilities and mocks for the platform.
package testutils

import (
	"context"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// MockAnalysisRepository provides a mock implementation of AnalysisRepositoryInterface for testing.
// It captures saved results and allows tests to verify persistence behavior.
type MockAnalysisRepository struct {
	SavedResult *review_models.AnalysisResult
	FindError   error
	CreateError error
}

// Create stores the analysis result and returns any configured error.
// This allows tests to verify that results are being persisted correctly.
func (m *MockAnalysisRepository) Create(ctx context.Context, result *review_models.AnalysisResult) error {
	m.SavedResult = result
	return m.CreateError
}

// FindByReviewAndMode retrieves an analysis result and returns any configured error.
// This allows tests to verify retrieval of stored results.
func (m *MockAnalysisRepository) FindByReviewAndMode(ctx context.Context, reviewID int64, mode string) (*review_models.AnalysisResult, error) {
	if m.FindError != nil {
		return nil, m.FindError
	}
	if m.SavedResult != nil && m.SavedResult.ReviewID == reviewID && m.SavedResult.Mode == mode {
		return m.SavedResult, nil
	}
	return nil, nil
}
