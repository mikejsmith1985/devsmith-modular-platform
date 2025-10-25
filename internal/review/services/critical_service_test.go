package services

import (
	"context"
	"testing"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestCriticalService_AnalyzeCritical_FindsSecurityIssues verifies that the CriticalService can identify security vulnerabilities in the codebase.
// Test 1: Finds security vulnerabilities
func TestCriticalService_AnalyzeCritical_FindsSecurityIssues(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewCriticalService(mockOllama, mockRepo)

	aiResponse := `{
		"issues": [
			{
				"severity": "critical",
				"category": "security",
				"file": "auth.go",
				"line": 10,
				"code_snippet": "db.Query(userInput)",
				"description": "SQL injection vulnerability",
				"impact": "Attacker can access entire database",
				"fix_suggestion": "Use parameterized queries: db.Query(sql, userInput)"
			}
		],
		"summary": "Found 1 critical security issue",
		"overall_grade": "D"
	}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(result *models.AnalysisResult) bool {
		return result.ReviewID == 1 && result.Mode == models.CriticalMode
	})).Return(nil)

	output, err := service.AnalyzeCritical(context.Background(), 1, "owner", "repo")

	assert.NoError(t, err)
	assert.Len(t, output.Issues, 1)
	assert.Equal(t, "critical", output.Issues[0].Severity)
	assert.Equal(t, "security", output.Issues[0].Category)
	assert.Contains(t, output.Issues[0].Description, "SQL injection")
	assert.Equal(t, "D", output.OverallGrade)
}

// TestCriticalService_AnalyzeCritical_MultipleIssueTypes ensures that the CriticalService can identify and categorize multiple types of issues, such as security and performance.
// Test 2: Finds multiple issue types
func TestCriticalService_AnalyzeCritical_MultipleIssueTypes(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewCriticalService(mockOllama, mockRepo)

	aiResponse := `{
		"issues": [
			{
				"severity": "critical",
				"category": "security",
				"file": "auth.go",
				"line": 10,
				"code_snippet": "eval(userInput)",
				"description": "Code injection vulnerability",
				"impact": "Remote code execution",
				"fix_suggestion": "Never use eval with user input"
			},
			{
				"severity": "high",
				"category": "performance",
				"file": "users.go",
				"line": 25,
				"code_snippet": "for user in users: db.query()",
				"description": "N+1 query problem",
				"impact": "Database overload with many users",
				"fix_suggestion": "Use JOIN or batch query"
			},
			{
				"severity": "medium",
				"category": "maintainability",
				"file": "utils.go",
				"line": 100,
				"code_snippet": "if ... elif ... elif ... (50 lines)",
				"description": "Excessive cyclomatic complexity",
				"impact": "Hard to test and maintain",
				"fix_suggestion": "Refactor to switch statement or strategy pattern"
			}
		],
		"summary": "Found 1 critical, 1 high, 1 medium issue",
		"overall_grade": "C"
	}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(result *models.AnalysisResult) bool {
		return result.ReviewID == 1 && result.Mode == models.CriticalMode
	})).Return(nil)

	output, err := service.AnalyzeCritical(context.Background(), 1, "owner", "repo")

	assert.NoError(t, err)
	assert.Len(t, output.Issues, 3)

	// Verify severity levels
	severities := []string{output.Issues[0].Severity, output.Issues[1].Severity, output.Issues[2].Severity}
	assert.Contains(t, severities, "critical")
	assert.Contains(t, severities, "high")
	assert.Contains(t, severities, "medium")

	// Verify categories
	categories := []string{output.Issues[0].Category, output.Issues[1].Category, output.Issues[2].Category}
	assert.Contains(t, categories, "security")
	assert.Contains(t, categories, "performance")
	assert.Contains(t, categories, "maintainability")
}

// Test 3: Clean code (no issues)
func TestCriticalService_AnalyzeCritical_CleanCode(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewCriticalService(mockOllama, mockRepo)

	aiResponse := `{
		"issues": [],
		"summary": "No issues found - excellent code quality",
		"overall_grade": "A"
	}`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(result *models.AnalysisResult) bool {
		return result.ReviewID == 1 && result.Mode == models.CriticalMode
	})).Return(nil)

	output, err := service.AnalyzeCritical(context.Background(), 1, "owner", "repo")

	assert.NoError(t, err)
	assert.Empty(t, output.Issues)
	assert.Equal(t, "A", output.OverallGrade)
}

// Test 4: Handles AI parsing errors
// TestCriticalService_AnalyzeCritical_InvalidJSON ensures that the CriticalService gracefully handles invalid JSON responses from the AI.
func TestCriticalService_AnalyzeCritical_InvalidJSON(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewCriticalService(mockOllama, mockRepo)

	aiResponse := `Invalid JSON response`
	mockOllama.On("Generate", mock.Anything, mock.Anything).Return(aiResponse, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(result *models.AnalysisResult) bool {
		return result.ReviewID == 1 && result.Mode == models.CriticalMode
	})).Return(nil)

	output, err := service.AnalyzeCritical(context.Background(), 1, "owner", "repo")

	// Adjusted test to expect an error for invalid JSON
	assert.Error(t, err, "Expected error for invalid JSON response")
	assert.Nil(t, output, "Output should be nil for invalid JSON")
}
