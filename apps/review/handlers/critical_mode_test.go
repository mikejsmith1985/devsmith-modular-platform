// Package review_handlers contains HTTP handlers for the review app.
package review_handlers

import (
	"context"
	"encoding/json"
	"testing"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestCriticalMode_IdentifiesSecurityIssues tests that Critical Mode finds security vulnerabilities
// RED phase: Tests define expected behavior - Ollama integration needed
func TestCriticalMode_IdentifiesSecurityIssues(t *testing.T) {
	// GIVEN: Code with SQL injection vulnerability
	mockLogger := &testutils.MockLogger{}
	mockOllamaClient := &testutils.MockOllamaClient{}
	analysisRepo := &testutils.MockAnalysisRepository{}

	service := review_services.NewCriticalService(mockOllamaClient, analysisRepo, mockLogger)

	// Mock Ollama response with SQL injection issue
	mockOllamaClient.GenerateResponse = `{
		"issues": [
			{
				"severity": "critical",
				"category": "security",
				"file": "main.go",
				"line": 42,
				"code_snippet": "db.Query(query + userInput)",
				"description": "SQL injection vulnerability - user input not parameterized",
				"impact": "Attacker can execute arbitrary SQL",
				"fix_suggestion": "Use parameterized query: db.Query(query, userInput)"
			}
		],
		"overall_grade": "F",
		"summary": "Found 1 critical security issue"
	}`

	// WHEN: Analyzing code
	result, err := service.AnalyzeCritical(context.Background(), 1, "package main\nfunc query(userInput string) {}")

	// THEN: Should identify the security issue
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Issues, "Must find security issues")
	assert.Equal(t, "critical", result.Issues[0].Severity)
	assert.Equal(t, "security", result.Issues[0].Category)
}

// TestCriticalMode_IdentifiesPerformanceIssues tests detection of performance problems
func TestCriticalMode_IdentifiesPerformanceIssues(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	mockOllamaClient := &testutils.MockOllamaClient{}
	analysisRepo := &testutils.MockAnalysisRepository{}

	service := review_services.NewCriticalService(mockOllamaClient, analysisRepo, mockLogger)

	mockOllamaClient.GenerateResponse = `{
		"issues": [
			{
				"severity": "high",
				"category": "performance",
				"file": "main.go",
				"line": 15,
				"code_snippet": "for _, user := range users { user.Details = db.GetDetails(user.ID) }",
				"description": "N+1 query problem - database called in loop",
				"impact": "Slow performance on large datasets",
				"fix_suggestion": "Fetch all details in single batch query"
			}
		],
		"overall_grade": "C",
		"summary": "Found 1 high severity performance issue"
	}`

	result, _ := service.AnalyzeCritical(context.Background(), 2, "test code")

	assert.NotEmpty(t, result.Issues)
	assert.Equal(t, "performance", result.Issues[0].Category)
	assert.Equal(t, "high", result.Issues[0].Severity)
}

// TestCriticalMode_DetectsMissingErrorHandling tests error handling detection
func TestCriticalMode_DetectsMissingErrorHandling(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	mockOllamaClient := &testutils.MockOllamaClient{}
	analysisRepo := &testutils.MockAnalysisRepository{}

	service := review_services.NewCriticalService(mockOllamaClient, analysisRepo, mockLogger)

	mockOllamaClient.GenerateResponse = `{
		"issues": [
			{
				"severity": "high",
				"category": "reliability",
				"file": "main.go",
				"line": 28,
				"code_snippet": "db.Close()",
				"description": "Error not checked after database close",
				"impact": "Silent failures could occur",
				"fix_suggestion": "Check error: if err := db.Close(); err != nil { log.Error(err) }"
			}
		],
		"overall_grade": "D",
		"summary": "Found 1 high severity reliability issue"
	}`

	result, _ := service.AnalyzeCritical(context.Background(), 3, "test code")

	assert.NotEmpty(t, result.Issues)
	assert.Equal(t, "reliability", result.Issues[0].Category)
}

// TestCriticalMode_CalculatesQualityScore tests that overall grade reflects code quality
func TestCriticalMode_CalculatesQualityScore(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	mockOllamaClient := &testutils.MockOllamaClient{}
	analysisRepo := &testutils.MockAnalysisRepository{}

	service := review_services.NewCriticalService(mockOllamaClient, analysisRepo, mockLogger)

	mockOllamaClient.GenerateResponse = `{
		"issues": [],
		"overall_grade": "A",
		"summary": "No issues found - excellent code quality"
	}`

	result, _ := service.AnalyzeCritical(context.Background(), 4, "test code")

	assert.Equal(t, "A", result.OverallGrade)
	assert.Empty(t, result.Issues)
}

// TestCriticalMode_HandleOllamaUnavailability tests graceful degradation when Ollama fails
func TestCriticalMode_HandleOllamaUnavailability(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	mockOllamaClient := &testutils.MockOllamaClient{}
	analysisRepo := &testutils.MockAnalysisRepository{}

	service := review_services.NewCriticalService(mockOllamaClient, analysisRepo, mockLogger)

	// Simulate Ollama failure
	mockOllamaClient.GenerateError = "Ollama service unavailable"

	result, err := service.AnalyzeCritical(context.Background(), 5, "test code")

	// Should return fallback response, not error
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "N/A", result.OverallGrade)
	assert.Contains(t, result.Summary, "unavailable")
}

// TestCriticalMode_ParsesComplexJSON tests handling of complex multi-issue responses
func TestCriticalMode_ParsesComplexJSON(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	mockOllamaClient := &testutils.MockOllamaClient{}
	analysisRepo := &testutils.MockAnalysisRepository{}

	service := review_services.NewCriticalService(mockOllamaClient, analysisRepo, mockLogger)

	mockOllamaClient.GenerateResponse = `{
		"issues": [
			{
				"severity": "critical",
				"category": "security",
				"file": "auth.go",
				"line": 10,
				"code_snippet": "credentials in code",
				"description": "Hardcoded credentials",
				"impact": "Credentials exposed in repository",
				"fix_suggestion": "Use environment variables"
			},
			{
				"severity": "high",
				"category": "performance",
				"file": "db.go",
				"line": 45,
				"code_snippet": "N+1 query",
				"description": "N+1 database queries",
				"impact": "Slow performance",
				"fix_suggestion": "Use batch query"
			},
			{
				"severity": "medium",
				"category": "maintainability",
				"file": "utils.go",
				"line": 88,
				"code_snippet": "complex logic",
				"description": "Complex function hard to understand",
				"impact": "Maintenance difficulty",
				"fix_suggestion": "Split into smaller functions"
			}
		],
		"overall_grade": "D",
		"summary": "Found 3 issues: 1 critical, 1 high, 1 medium"
	}`

	result, err := service.AnalyzeCritical(context.Background(), 6, "test code")

	assert.NoError(t, err)
	assert.Equal(t, 3, len(result.Issues))
	assert.Equal(t, "critical", result.Issues[0].Severity)
	assert.Equal(t, "high", result.Issues[1].Severity)
	assert.Equal(t, "medium", result.Issues[2].Severity)
}

// TestCriticalMode_StoreResults tests that analysis results are persisted
func TestCriticalMode_StoreResults(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	mockOllamaClient := &testutils.MockOllamaClient{}
	analysisRepo := &testutils.MockAnalysisRepository{}

	service := review_services.NewCriticalService(mockOllamaClient, analysisRepo, mockLogger)

	mockOllamaClient.GenerateResponse = `{
		"issues": [{"severity": "info", "category": "test", "description": "Test issue"}],
		"overall_grade": "B",
		"summary": "Test summary"
	}`

	_, err := service.AnalyzeCritical(context.Background(), 7, "test code")

	assert.NoError(t, err)
	// Check that Create was called on the repository
	assert.NotNil(t, analysisRepo.SavedResult)
	assert.Equal(t, review_models.CriticalMode, analysisRepo.SavedResult.Mode)
}

// Helper to validate JSON structure can be parsed into CriticalModeOutput
func validateCriticalModeJSON(t *testing.T, jsonStr string) *review_models.CriticalModeOutput {
	var output review_models.CriticalModeOutput
	err := json.Unmarshal([]byte(jsonStr), &output)
	assert.NoError(t, err, "Should parse as valid JSON")
	return &output
}
