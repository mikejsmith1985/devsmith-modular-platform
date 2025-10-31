package ai

import (
	"context"
	"time"
)

// AIProvider is the universal interface all AI providers must implement
type AIProvider interface {
	// Generate sends a prompt and returns the AI response
	Generate(ctx context.Context, req *AIRequest) (*AIResponse, error)

	// HealthCheck verifies the provider is reachable
	HealthCheck(ctx context.Context) error

	// GetModelInfo returns model capabilities and metadata
	GetModelInfo() *ModelInfo
}

// AIRequest represents a request to any AI provider
type AIRequest struct {
	Metadata    map[string]interface{} // Provider-specific options
	Prompt      string                 // The prompt to send to AI
	Model       string                 // Model identifier
	Temperature float64                // 0.0-1.0, controls randomness
	MaxTokens   int                    // Response length limit
}

// AIResponse represents the response from any AI provider
type AIResponse struct {
	Content      string        // The generated text
	Model        string        // Model that fulfilled request
	FinishReason string        // 'complete', 'length', 'error'
	InputTokens  int           // Number of input tokens used
	OutputTokens int           // Number of output tokens used
	ResponseTime time.Duration // Time taken to generate response
	CostUSD      float64       // Estimated cost (0 for local providers like Ollama)
}

// ModelInfo describes model capabilities
type ModelInfo struct {
	Provider                 string   // 'ollama', 'anthropic', 'openai'
	Model                    string   // Model identifier
	DisplayName              string   // Human-readable name
	Capabilities             []string // ['code_analysis', 'critical_review', etc.]
	MaxTokens                int      // Max context window
	CostPer1kInputTokens     float64  // Cost per 1k input tokens ($)
	CostPer1kOutputTokens    float64  // Cost per 1k output tokens ($)
	DefaultTemperature       float64  // Recommended temperature for this model
	SupportsStreaming        bool     // Whether model supports streaming responses
	RecommendedForCodeReview bool     // Is this model good for code review tasks
}

// Router determines which provider to use for a request
type Router interface {
	// Route selects the appropriate provider based on app, user preferences, and config
	Route(ctx context.Context, appName string, userID int64) (AIProvider, error)

	// GetAvailableModels returns all models available to a user in an app
	GetAvailableModels(ctx context.Context, appName string, userID int64) ([]*ModelInfo, error)

	// SetUserPreference updates user's model selection for an app
	SetUserPreference(ctx context.Context, userID int64, appName string, provider string, model string, persist bool) error

	// LogUsage records API usage for cost tracking
	LogUsage(ctx context.Context, userID int64, appName string, req *AIRequest, resp *AIResponse) error
}
