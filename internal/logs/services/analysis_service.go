// Package logs_services provides service implementations for logs operations.
package logs_services

import (
	"context"
	"encoding/json"
	"fmt"

	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// AnalysisServiceImpl implements the analysis service interface
type AnalysisServiceImpl struct {
	aiAnalyzer     *AIAnalyzer
	patternMatcher *PatternMatcher
}

// NewAnalysisService creates a new analysis service
func NewAnalysisService(aiAnalyzer *AIAnalyzer, patternMatcher *PatternMatcher) *AnalysisServiceImpl {
	return &AnalysisServiceImpl{
		aiAnalyzer:     aiAnalyzer,
		patternMatcher: patternMatcher,
	}
}

// AnalyzeLogEntry analyzes a log entry using AI
func (s *AnalysisServiceImpl) AnalyzeLogEntry(ctx context.Context, entry *logs_models.LogEntry) (*AnalysisResult, error) {
	// First classify the log to provide context to AI
	issueType := s.patternMatcher.Classify(entry.Message)

	// Build analysis request
	req := AnalysisRequest{
		LogEntries: []logs_models.LogEntry{*entry},
		Context:    entry.Level, // "error", "warning", "info"
	}

	// Perform AI analysis
	result, err := s.aiAnalyzer.Analyze(ctx, req)
	if err != nil {
		return nil, err
	}

	// Store analysis in log entry (for future caching/persistence)
	analysisJSON, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("Warning: failed to marshal analysis result: %v\n", err)
		analysisJSON = []byte("{}")
	}
	entry.AIAnalysis = analysisJSON
	entry.IssueType = issueType
	entry.SeverityScore = result.Severity

	return result, nil
}

// ClassifyLogEntry classifies a log entry into a known issue type
func (s *AnalysisServiceImpl) ClassifyLogEntry(ctx context.Context, entry *logs_models.LogEntry) (string, error) {
	issueType := s.patternMatcher.Classify(entry.Message)

	// Update entry with classification
	entry.IssueType = issueType

	return issueType, nil
}
