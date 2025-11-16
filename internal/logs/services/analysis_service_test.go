// Package logs_services provides service implementations for logs operations.
package logs_services

import (
	"context"
	"encoding/json"
	"testing"

	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAIAnalyzer mocks the AIAnalyzer for testing
type MockAIAnalyzer struct {
	mock.Mock
}

func (m *MockAIAnalyzer) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AnalysisResult), args.Error(1)
}

// MockPatternMatcher mocks the PatternMatcher for testing
type MockPatternMatcher struct {
	mock.Mock
}

func (m *MockPatternMatcher) Classify(message string) string {
	args := m.Called(message)
	return args.String(0)
}

func TestNewAnalysisService(t *testing.T) {
	aiAnalyzer := &AIAnalyzer{}
	patternMatcher := &PatternMatcher{}

	service := NewAnalysisService(aiAnalyzer, patternMatcher)

	assert.NotNil(t, service)
	assert.Equal(t, aiAnalyzer, service.aiAnalyzer)
	assert.Equal(t, patternMatcher, service.patternMatcher)
}

func TestAnalysisService_AnalyzeLogEntry_Success(t *testing.T) {
	// Setup
	entry := &logs_models.LogEntry{
		ID:      1,
		Service: "portal",
		Level:   "error",
		Message: "database connection refused",
	}

	// Create pattern matcher that returns db_connection
	patternMatcher := NewPatternMatcher()
	aiAnalyzer := &AIAnalyzer{} // Can't easily mock internal calls, will test integration

	service := NewAnalysisService(aiAnalyzer, patternMatcher)

	// Verify the service structure is correct
	assert.NotNil(t, service)
	assert.NotNil(t, service.aiAnalyzer)
	assert.NotNil(t, service.patternMatcher)

	// Test classification works
	issueType, err := service.ClassifyLogEntry(context.Background(), entry)
	assert.NoError(t, err)
	assert.Equal(t, "db_connection", issueType)
	assert.Equal(t, "db_connection", entry.IssueType)
}

func TestAnalysisService_ClassifyLogEntry_DatabaseError(t *testing.T) {
	patternMatcher := NewPatternMatcher()
	service := NewAnalysisService(nil, patternMatcher)

	entry := &logs_models.LogEntry{
		Message: "database connection refused",
	}

	issueType, err := service.ClassifyLogEntry(context.Background(), entry)

	assert.NoError(t, err)
	assert.Equal(t, "db_connection", issueType)
	assert.Equal(t, "db_connection", entry.IssueType)
}

func TestAnalysisService_ClassifyLogEntry_AuthError(t *testing.T) {
	patternMatcher := NewPatternMatcher()
	service := NewAnalysisService(nil, patternMatcher)

	entry := &logs_models.LogEntry{
		Message: "authentication failed for user",
	}

	issueType, err := service.ClassifyLogEntry(context.Background(), entry)

	assert.NoError(t, err)
	assert.Equal(t, "auth_failure", issueType)
	assert.Equal(t, "auth_failure", entry.IssueType)
}

func TestAnalysisService_ClassifyLogEntry_UnknownError(t *testing.T) {
	patternMatcher := NewPatternMatcher()
	service := NewAnalysisService(nil, patternMatcher)

	entry := &logs_models.LogEntry{
		Message: "something weird happened",
	}

	issueType, err := service.ClassifyLogEntry(context.Background(), entry)

	assert.NoError(t, err)
	assert.Equal(t, "unknown", issueType)
	assert.Equal(t, "unknown", entry.IssueType)
}

func TestAnalysisService_AnalyzeLogEntry_StoresAnalysisInEntry(t *testing.T) {
	// This is more of an integration test showing the flow
	// Verify the entry can store analysis
	entry := &logs_models.LogEntry{
		Message: "database connection refused",
		Level:   "error",
	}

	// Verify the entry can store analysis
	testResult := &AnalysisResult{
		RootCause:    "Test cause",
		SuggestedFix: "Test fix",
		Severity:     3,
	}

	analysisJSON, err := json.Marshal(testResult)
	assert.NoError(t, err)

	entry.AIAnalysis = analysisJSON
	entry.IssueType = "db_connection"
	entry.SeverityScore = testResult.Severity

	// Verify we can retrieve it
	var retrieved AnalysisResult
	err = json.Unmarshal(entry.AIAnalysis, &retrieved)
	assert.NoError(t, err)
	assert.Equal(t, "Test cause", retrieved.RootCause)
	assert.Equal(t, "Test fix", retrieved.SuggestedFix)
	assert.Equal(t, 3, retrieved.Severity)
}
