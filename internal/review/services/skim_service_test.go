package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOllamaClient struct{ mock.Mock }

func (m *MockOllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

type MockAnalysisRepository struct{ mock.Mock }

func (m *MockAnalysisRepository) FindByReviewAndMode(ctx context.Context, reviewID int64, mode string) (*models.AnalysisResult, error) {
	args := m.Called(ctx, reviewID, mode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AnalysisResult), args.Error(1)
}
func (m *MockAnalysisRepository) Create(ctx context.Context, result *models.AnalysisResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

func TestSkimService_AnalyzeSkim_Success(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewSkimService(mockOllama, mockRepo)

	mockRepo.On("FindByReviewAndMode", mock.Anything, int64(1), models.SkimMode).
		Return(nil, fmt.Errorf("not found"))

	aiResponse := `{"functions": [], "interfaces": [], "data_models": [], "workflows": [], "summary": "Test"}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	output, err := service.AnalyzeSkim(context.Background(), 1, "owner", "repo")

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "Test", output.Summary)
}
