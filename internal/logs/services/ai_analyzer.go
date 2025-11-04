// Package logs_services provides service implementations for logs operations.
package logs_services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// AIAnalyzer analyzes log entries using AI to provide root cause analysis and fixes
type AIAnalyzer struct {
	aiProvider ai.Provider
	cache      *AnalysisCache
}

// AnalysisCache caches analysis results to avoid redundant AI calls
type AnalysisCache struct {
	mu      sync.RWMutex
	results map[string]*AnalysisResult
}

// AnalysisRequest represents a request for AI analysis of log entries
type AnalysisRequest struct {
	LogEntries []logs_models.LogEntry `json:"log_entries"`
	Context    string                 `json:"context"` // "error", "warning", "info"
}

// AnalysisResult represents the AI analysis result
type AnalysisResult struct {
	RootCause    string   `json:"root_cause"`
	SuggestedFix string   `json:"suggested_fix"`
	Severity     int      `json:"severity"`       // 1-5
	RelatedLogs  []string `json:"related_logs"`   // correlation_ids
	FixSteps     []string `json:"fix_steps"`      // Step-by-step instructions
}

// NewAIAnalyzer creates a new AI analyzer service
func NewAIAnalyzer(provider ai.Provider) *AIAnalyzer {
	return &AIAnalyzer{
		aiProvider: provider,
		cache: &AnalysisCache{
			results: make(map[string]*AnalysisResult),
		},
	}
}

// Analyze performs AI analysis on the given log entries
func (a *AIAnalyzer) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error) {
	// Generate cache key based on log messages
	cacheKey := a.generateCacheKey(req)
	
	// Check cache first
	if cachedResult := a.cache.Get(cacheKey); cachedResult != nil {
		return cachedResult, nil
	}
	
	// Build prompt for AI
	prompt := a.buildPrompt(req)
	
	// Call AI provider
	aiReq := &ai.Request{
		Prompt:      prompt,
		Model:       "qwen2.5-coder:7b-instruct-q4_K_M", // Default model
		Temperature: 0.3,                                  // Low temperature for consistent analysis
		MaxTokens:   2000,
	}
	
	resp, err := a.aiProvider.Generate(ctx, aiReq)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}
	
	// Parse AI response
	result, err := a.parseResponse(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}
	
	// Cache the result
	a.cache.Set(cacheKey, result)
	
	return result, nil
}

// generateCacheKey creates a cache key from the log entries
func (a *AIAnalyzer) generateCacheKey(req AnalysisRequest) string {
	var messages []string
	for _, entry := range req.LogEntries {
		messages = append(messages, entry.Message)
	}
	
	combined := strings.Join(messages, "|")
	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash)
}

// buildPrompt constructs the AI prompt from the analysis request
func (a *AIAnalyzer) buildPrompt(req AnalysisRequest) string {
	var sb strings.Builder
	
	sb.WriteString("You are a systems diagnostics expert analyzing application logs.\n\n")
	sb.WriteString(fmt.Sprintf("Context: %s\n\n", req.Context))
	sb.WriteString("Log Entries:\n")
	
	for _, entry := range req.LogEntries {
		sb.WriteString(fmt.Sprintf("[%s] %s - %s\n", entry.Level, entry.Service, entry.Message))
		if len(entry.Metadata) > 0 {
			sb.WriteString(fmt.Sprintf("Metadata: %s\n", string(entry.Metadata)))
		}
	}
	
	sb.WriteString("\nTasks:\n")
	sb.WriteString("1. Identify the root cause (be specific - which component/function/line is failing?)\n")
	sb.WriteString("2. Suggest a fix (concrete code change or configuration adjustment)\n")
	sb.WriteString("3. Rate severity (1=info, 2=minor, 3=moderate, 4=serious, 5=critical)\n")
	sb.WriteString("4. List related log correlation IDs if this is part of a larger issue\n")
	sb.WriteString("5. Provide step-by-step fix instructions\n\n")
	sb.WriteString("Respond in JSON format:\n")
	sb.WriteString(`{
    "root_cause": "...",
    "suggested_fix": "...",
    "severity": 3,
    "related_logs": ["correlation-id-1", "correlation-id-2"],
    "fix_steps": ["Step 1", "Step 2"]
}`)
	
	return sb.String()
}

// parseResponse parses the AI response into an AnalysisResult
func (a *AIAnalyzer) parseResponse(content string) (*AnalysisResult, error) {
	// Try to extract JSON from the content (in case AI adds explanation before/after)
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")
	
	if startIdx == -1 || endIdx == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}
	
	jsonContent := content[startIdx : endIdx+1]
	
	var result AnalysisResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}
	
	return &result, nil
}

// Get retrieves a cached analysis result
func (c *AnalysisCache) Get(key string) *AnalysisResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.results[key]
}

// Set stores an analysis result in the cache
func (c *AnalysisCache) Set(key string, result *AnalysisResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results[key] = result
}

// Clear clears the cache
func (c *AnalysisCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results = make(map[string]*AnalysisResult)
}
