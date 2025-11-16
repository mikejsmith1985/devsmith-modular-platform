package review_services

import (
	"context"
	"fmt"
	"strings"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai/providers"
	reviewcontext "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/context"
)

// UnifiedAIClient implements OllamaClientInterface by routing to the appropriate
// AI provider (Ollama, Anthropic, OpenAI) based on the user's AI Factory configuration.
// This client queries the Portal service's AI Factory API to get the user's configured model.
type UnifiedAIClient struct {
	portalClient *PortalClient
}

// NewUnifiedAIClient creates a new unified AI client that fetches configs from Portal's AI Factory
func NewUnifiedAIClient(portalURL string) *UnifiedAIClient {
	return &UnifiedAIClient{
		portalClient: NewPortalClient(portalURL),
	}
}

// Generate implements OllamaClientInterface.Generate
// Routes the request to the appropriate AI provider based on user's AI Factory configuration
func (c *UnifiedAIClient) Generate(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	// Get session token from context
	// The token is set by RedisSessionAuthMiddleware as "session_token" in Gin context
	// Handlers should pass it through to context using reviewcontext.SessionTokenKey
	sessionToken, ok := ctx.Value(reviewcontext.SessionTokenKey).(string)
	if !ok || sessionToken == "" {
		return "", fmt.Errorf("no session token in context - user must be authenticated. Please ensure RedisSessionAuthMiddleware is active and session token is passed to context")
	}

	// Get user's AI configuration from Portal's AI Factory
	config, err := c.portalClient.GetEffectiveConfigForApp(ctx, sessionToken, "review")
	if err != nil {
		return "", fmt.Errorf("failed to get AI configuration from Portal: %w. Please configure an AI model in AI Factory (/llm-config)", err)
	}

	// Allow model override from context (for advanced users selecting different models)
	model := config.ModelName
	if contextModel, ok := ctx.Value(reviewcontext.ModelContextKey).(string); ok && contextModel != "" {
		model = contextModel
	}

	// Instantiate the appropriate provider based on configuration
	provider, err := c.createProvider(config, model)
	if err != nil {
		return "", fmt.Errorf("failed to create AI provider: %w", err)
	}

	// Construct ai.Request
	req := &ai.Request{
		Model:       model,
		Prompt:      prompt,
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
	}

	// Call the provider
	resp, err := provider.Generate(ctx, req)
	if err != nil {
		return "", fmt.Errorf("%s generation failed: %w", config.Provider, err)
	}

	if resp == nil {
		return "", fmt.Errorf("%s returned nil response", config.Provider)
	}

	return resp.Content, nil
}

// createProvider instantiates the correct AI provider based on LLM configuration
func (c *UnifiedAIClient) createProvider(config *LLMConfig, model string) (ai.Provider, error) {
	providerLower := strings.ToLower(strings.TrimSpace(config.Provider))

	switch providerLower {
	case "ollama":
		// Ollama requires endpoint but no API key
		endpoint := config.APIEndpoint
		if endpoint == "" {
			return nil, fmt.Errorf("Ollama endpoint not configured in AI Factory")
		}
		return providers.NewOllamaClient(endpoint, model), nil

	case "anthropic":
		// Anthropic (Claude) requires API key
		if config.APIKey == "" {
			return nil, fmt.Errorf("Anthropic API key not configured in AI Factory")
		}
		return providers.NewAnthropicClient(config.APIKey, model), nil

	case "openai":
		// OpenAI (GPT) requires API key
		if config.APIKey == "" {
			return nil, fmt.Errorf("OpenAI API key not configured in AI Factory")
		}
		return providers.NewOpenAIClient(config.APIKey, model), nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s. Supported providers: ollama, anthropic, openai", config.Provider)
	}
}

// GetModelCapabilities returns the capabilities of a model (local vs API features)
func (c *UnifiedAIClient) GetModelCapabilities(provider string) ModelCapabilities {
	providerLower := strings.ToLower(strings.TrimSpace(provider))

	// Local models (Ollama) have limited capabilities
	if providerLower == "ollama" {
		return ModelCapabilities{
			Provider:                  "ollama",
			IsLocal:                   true,
			SupportsAnalogies:         false,
			SupportsAdvancedReasoning: false,
			SupportsCodeExecution:     false,
			MaxContextWindow:          8192, // Conservative estimate
			Description:               "Local model - Fast, private, but limited reasoning",
		}
	}

	// API models have full capabilities
	return ModelCapabilities{
		Provider:                  providerLower,
		IsLocal:                   false,
		SupportsAnalogies:         true,
		SupportsAdvancedReasoning: true,
		SupportsCodeExecution:     false,  // Future feature
		MaxContextWindow:          200000, // Claude 3.5 Sonnet supports 200k
		Description:               "Cloud API model - Advanced reasoning and analogies",
	}
}

// ModelCapabilities describes what a model can do
type ModelCapabilities struct {
	Provider                  string `json:"provider"`
	IsLocal                   bool   `json:"is_local"`
	SupportsAnalogies         bool   `json:"supports_analogies"`
	SupportsAdvancedReasoning bool   `json:"supports_advanced_reasoning"`
	SupportsCodeExecution     bool   `json:"supports_code_execution"`
	MaxContextWindow          int    `json:"max_context_window"`
	Description               string `json:"description"`
}
