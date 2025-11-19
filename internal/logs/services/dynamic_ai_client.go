package logs_services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai/providers"
)

// DynamicAIClient implements AIProvider by dynamically fetching LLM configuration
// from Portal's AI Factory at request time (not startup time).
// This allows the Logs service to work even if AI Factory is configured AFTER startup.
type DynamicAIClient struct {
	portalURL string
}

// NewDynamicAIClient creates a new dynamic AI client
func NewDynamicAIClient(portalURL string) *DynamicAIClient {
	return &DynamicAIClient{
		portalURL: portalURL,
	}
}

// Generate implements AIProvider interface
func (c *DynamicAIClient) Generate(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	// Fetch LLM configuration from Portal's AI Factory
	config, err := c.fetchLLMConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch LLM configuration: %w. Please configure an AI model in AI Factory (/llm-config)", err)
	}

	// Use model from request if provided, otherwise use config default
	model := request.Model
	if model == "" {
		model = config.ModelName
	}

	// Create provider based on configuration
	provider, err := c.createProvider(config, model)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	// Convert AIRequest to ai.Request
	aiReq := &ai.Request{
		Model:       model,
		Prompt:      request.Prompt,
		Temperature: 0.7,  // Default for log analysis
		MaxTokens:   2048, // Default for log insights
	}

	// Call provider
	aiResp, err := provider.Generate(ctx, aiReq)
	if err != nil {
		return nil, fmt.Errorf("%s generation failed: %w", config.Provider, err)
	}

	if aiResp == nil {
		return nil, fmt.Errorf("%s returned nil response", config.Provider)
	}

	// Convert ai.Response to AIResponse
	return &AIResponse{
		Content: aiResp.Content,
	}, nil
}

// LLMConfig represents Portal's AI Factory configuration
type LLMConfig struct {
	ID          string  `json:"id"`
	UserID      int     `json:"user_id"`
	Provider    string  `json:"provider"`
	ModelName   string  `json:"model_name"`
	APIEndpoint string  `json:"api_endpoint"`
	APIKey      string  `json:"api_key"`
	IsDefault   bool    `json:"is_default"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// fetchLLMConfig queries Portal's AI Factory for default LLM configuration
func (c *DynamicAIClient) fetchLLMConfig() (*LLMConfig, error) {
	// Query Portal's AI Factory API for logs app preference or default config
	url := fmt.Sprintf("%s/api/portal/app-llm-preferences", c.portalURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query Portal API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Portal API returned %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var preferences map[string]LLMConfig
	if err := json.NewDecoder(resp.Body).Decode(&preferences); err != nil {
		return nil, fmt.Errorf("failed to parse Portal response: %w", err)
	}

	// Get logs app preference
	config, ok := preferences["logs"]
	if !ok {
		return nil, fmt.Errorf("no LLM configuration found for logs app. Please configure an AI model in AI Factory")
	}

	return &config, nil
}

// createProvider instantiates the correct AI provider based on LLM configuration
func (c *DynamicAIClient) createProvider(config *LLMConfig, model string) (ai.Provider, error) {
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

	case "deepseek":
		// DeepSeek requires API key
		if config.APIKey == "" {
			return nil, fmt.Errorf("DeepSeek API key not configured in AI Factory")
		}
		// DeepSeek uses OpenAI-compatible API
		return providers.NewOpenAIClient(config.APIKey, model), nil

	case "mistral":
		// Mistral requires API key
		if config.APIKey == "" {
			return nil, fmt.Errorf("Mistral API key not configured in AI Factory")
		}
		// Mistral uses OpenAI-compatible API
		return providers.NewOpenAIClient(config.APIKey, model), nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s. Supported providers: ollama, anthropic, openai, deepseek, mistral", config.Provider)
	}
}
