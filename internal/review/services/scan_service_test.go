package services

import (
	"context"
	"testing"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test 1: Successful scan with matches
func TestScanService_AnalyzeScan_Success(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	mockLogger := &testutils.MockLogger{}
	service := NewScanService(mockOllama, mockRepo, mockLogger)

	aiResponse := `{
		"matches": [
			{
				"file": "auth.go",
				"line": 10,
				"code_snippet": "func Login()",
				"relevance": 0.9,
				"context": "Main authentication entry point"
			}
		],
		"summary": "Found 1 match"
	}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	output, err := service.AnalyzeScan(context.Background(), 1, "authentication")

	assert.NoError(t, err)
	assert.Len(t, output.Matches, 1)
	assert.Equal(t, "auth.go", output.Matches[0].FilePath)
	assert.Equal(t, 0.9, output.Matches[0].Relevance)
}

// Test 2: Empty query returns error
func TestScanService_AnalyzeScan_EmptyQuery(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	mockLogger := &testutils.MockLogger{}
	service := NewScanService(mockOllama, mockRepo, mockLogger)

	_, err := service.AnalyzeScan(context.Background(), 1, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query cannot be empty")
}

// Test 3: No matches found
func TestScanService_AnalyzeScan_NoMatches(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	mockLogger := &testutils.MockLogger{}
	service := NewScanService(mockOllama, mockRepo, mockLogger)

	aiResponse := `{"matches": [], "summary": "No matches found"}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	output, err := service.AnalyzeScan(context.Background(), 1, "nonexistent")

	assert.NoError(t, err)
	assert.Empty(t, output.Matches)
}
