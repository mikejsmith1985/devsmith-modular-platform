package logs_services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// AIInsightsService handles AI-powered log analysis
type AIInsightsService struct {
	aiClient AIProvider
	logRepo  LogRepository
	repo     AIInsightsRepository
}

// AIProvider interface for AI model integration
type AIProvider interface {
	Generate(ctx context.Context, request *AIRequest) (*AIResponse, error)
}

// AIRequest represents a request to the AI model
type AIRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// AIResponse represents AI model response
type AIResponse struct {
	Content string `json:"content"`
}

// LogRepository interface for fetching logs
type LogRepository interface {
	GetByID(ctx context.Context, id int64) (*logs_models.LogEntry, error)
}

// AIInsightsRepository interface for database operations
type AIInsightsRepository interface {
	Upsert(ctx context.Context, insight *logs_models.AIInsight) (*logs_models.AIInsight, error)
	GetByLogID(ctx context.Context, logID int64) (*logs_models.AIInsight, error)
}

// NewAIInsightsService creates a new AI insights service
func NewAIInsightsService(aiClient AIProvider, logRepo LogRepository, insightsRepo AIInsightsRepository) *AIInsightsService {
	return &AIInsightsService{
		aiClient: aiClient,
		logRepo:  logRepo,
		repo:     insightsRepo,
	}
}

// GenerateInsights generates AI insights for a log entry
func (s *AIInsightsService) GenerateInsights(ctx context.Context, logID int64, model string) (*logs_models.AIInsight, error) {
	// 1. Fetch log entry
	log, err := s.logRepo.GetByID(ctx, logID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch log: %w", err)
	}

	// 2. Build prompt
	prompt := s.buildAnalysisPrompt(log)

	// 3. Call AI
	response, err := s.aiClient.Generate(ctx, &AIRequest{
		Model:  model,
		Prompt: prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// 4. Parse response
	insight, err := s.parseAIResponse(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	insight.LogID = logID
	insight.ModelUsed = model
	insight.GeneratedAt = time.Now()

	// 5. Save to database (upsert - overwrites existing)
	savedInsight, err := s.repo.Upsert(ctx, insight)
	if err != nil {
		return nil, fmt.Errorf("failed to save insight: %w", err)
	}

	return savedInsight, nil
}

// GetInsights retrieves existing AI insights for a log
func (s *AIInsightsService) GetInsights(ctx context.Context, logID int64) (*logs_models.AIInsight, error) {
	return s.repo.GetByLogID(ctx, logID)
}

// buildAnalysisPrompt constructs the AI prompt for log analysis
func (s *AIInsightsService) buildAnalysisPrompt(log *logs_models.LogEntry) string {
	metadataJSON := "{}"
	if log.Metadata != nil {
		if bytes, err := json.Marshal(log.Metadata); err == nil {
			metadataJSON = string(bytes)
		}
	}

	return fmt.Sprintf(`Analyze this log entry and provide insights:

Level: %s
Service: %s
Message: %s
Timestamp: %s
Metadata: %s

Please provide:
1. Analysis: What does this log indicate? (2-3 sentences)
2. Root Cause: What likely caused this? (1-2 sentences, leave empty if not applicable)
3. Suggestions: How to fix or prevent this? (3-5 actionable items)

Format your response as JSON:
{
  "analysis": "Brief analysis of what this log indicates",
  "root_cause": "Brief explanation of the root cause (or empty string if not applicable)",
  "suggestions": ["Actionable suggestion 1", "Actionable suggestion 2", "Actionable suggestion 3"]
}

Respond ONLY with valid JSON, no additional text.`,
		log.Level,
		log.Service,
		log.Message,
		log.CreatedAt.Format(time.RFC3339),
		metadataJSON,
	)
}

// parseAIResponse parses the AI response into an AIInsight struct
// Handles both pure JSON and JSON wrapped in markdown code blocks
func (s *AIInsightsService) parseAIResponse(content string) (*logs_models.AIInsight, error) {
	// Try to find JSON in the response (AI might wrap it in markdown)
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonContent := content[jsonStart : jsonEnd+1]

	var parsed struct {
		Analysis    string   `json:"analysis"`
		RootCause   string   `json:"root_cause"`
		Suggestions []string `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(jsonContent), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	insight := &logs_models.AIInsight{
		Analysis:    parsed.Analysis,
		RootCause:   parsed.RootCause,
		Suggestions: parsed.Suggestions,
	}

	// Ensure Suggestions is not nil
	if insight.Suggestions == nil {
		insight.Suggestions = []string{}
	}

	return insight, nil
}
