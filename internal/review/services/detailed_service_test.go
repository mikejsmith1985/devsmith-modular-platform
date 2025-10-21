package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test 1: Successful detailed analysis
func TestDetailedService_AnalyzeDetailed_Success(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewDetailedService(mockOllama, mockRepo)

	aiResponse := `{
		"lines": [
			{
				"line_num": 1,
				"code": "package main",
				"explanation": "Package declaration",
				"complexity": "low"
			},
			{
				"line_num": 5,
				"code": "func Login()",
				"explanation": "Authentication handler",
				"complexity": "medium",
				"side_effects": ["Database query", "Session creation"],
				"variables_modified": ["userSession"]
			}
		],
		"data_flow": [
			{"from": "input", "to": "database", "description": "Credentials validated"}
		],
		"summary": "Authentication file with 2 functions"
	}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	output, err := service.AnalyzeDetailed(context.Background(), 1, "auth.go")

	assert.NoError(t, err)
	assert.Len(t, output.Lines, 2)
	assert.Equal(t, "package main", output.Lines[0].Code)
	assert.Contains(t, output.Lines[1].SideEffects, "Database query")
	assert.Len(t, output.DataFlow, 1)
}

// Test 2: Empty file path returns error
func TestDetailedService_AnalyzeDetailed_EmptyFilePath(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewDetailedService(mockOllama, mockRepo)

	_, err := service.AnalyzeDetailed(context.Background(), 1, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file path cannot be empty")
}

// Test 3: Complex file with side effects
func TestDetailedService_AnalyzeDetailed_WithSideEffects(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewDetailedService(mockOllama, mockRepo)

	aiResponse := `{
		"lines": [
			{
				"line_num": 10,
				"code": "db.Exec(sql)",
				"explanation": "Executes database query",
				"complexity": "high",
				"side_effects": ["Database write", "Triggers audit log"],
				"variables_modified": ["recordCount", "lastModified"]
			}
		],
		"data_flow": [
			{"from": "userInput", "to": "database", "description": "Unvalidated input risk"}
		],
		"summary": "Database operations with side effects"
	}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	output, err := service.AnalyzeDetailed(context.Background(), 1, "db.go")

	assert.NoError(t, err)
	assert.Equal(t, "high", output.Lines[0].Complexity)
	assert.Len(t, output.Lines[0].SideEffects, 2)
	assert.Len(t, output.Lines[0].VariablesModified, 2)
}
