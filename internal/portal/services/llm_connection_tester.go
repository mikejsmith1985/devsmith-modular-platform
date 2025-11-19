package portal_services

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai/providers"
)

// LLMConnectionTester tests connections to various LLM providers
type LLMConnectionTester struct{}

// NewLLMConnectionTester creates a new connection tester
func NewLLMConnectionTester() *LLMConnectionTester {
	return &LLMConnectionTester{}
}

// TestConnectionRequest contains the parameters for testing a connection
type TestConnectionRequest struct {
	Provider string
	Model    string
	APIKey   string
	Endpoint string
}

// TestConnectionResponse contains the result of a connection test
type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// TestConnection tests the connection to an LLM provider
func (t *LLMConnectionTester) TestConnection(ctx context.Context, req TestConnectionRequest) TestConnectionResponse {
	// Create context with timeout
	testCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Create provider based on type
	var provider ai.Provider
	var err error

	switch req.Provider {
	case "ollama":
		// Provide default Ollama endpoint if not specified
		endpoint := req.Endpoint
		if endpoint == "" {
			// Use Docker-compatible endpoint that resolves to host machine
			endpoint = os.Getenv("OLLAMA_ENDPOINT")
			if endpoint == "" {
				endpoint = "http://host.docker.internal:11434"
			}
		}
		provider = providers.NewOllamaClient(endpoint, req.Model)

	case "anthropic":
		if req.APIKey == "" {
			return TestConnectionResponse{
				Success: false,
				Message: "Connection failed",
				Details: "Anthropic API key is required",
			}
		}
		provider = providers.NewAnthropicClient(req.APIKey, req.Model)

	case "openai":
		if req.APIKey == "" {
			return TestConnectionResponse{
				Success: false,
				Message: "Connection failed",
				Details: "OpenAI API key is required",
			}
		}
		provider = providers.NewOpenAIClient(req.APIKey, req.Model)

	case "deepseek":
		if req.APIKey == "" {
			return TestConnectionResponse{
				Success: false,
				Message: "Connection failed",
				Details: "DeepSeek API key is required",
			}
		}
		provider = providers.NewDeepSeekClient(req.APIKey, req.Model)

	case "mistral":
		if req.APIKey == "" {
			return TestConnectionResponse{
				Success: false,
				Message: "Connection failed",
				Details: "Mistral API key is required",
			}
		}
		provider = providers.NewMistralClient(req.APIKey, req.Model)

	default:
		return TestConnectionResponse{
			Success: false,
			Message: "Connection failed",
			Details: fmt.Sprintf("Unsupported provider: %s", req.Provider),
		}
	}

	// Test the connection using a minimal prompt
	testReq := &ai.Request{
		Model:       req.Model,
		Prompt:      "Hello", // Minimal test prompt
		MaxTokens:   10,      // Minimal response
		Temperature: 0.1,
	}

	_, err = provider.Generate(testCtx, testReq)
	if err != nil {
		// Provide more helpful error messages based on error type
		details := fmt.Sprintf("Failed to connect to %s: %v", req.Provider, err)

		// Add helpful suggestions based on provider
		if req.Provider == "ollama" {
			endpoint := req.Endpoint
			if endpoint == "" {
				endpoint = os.Getenv("OLLAMA_ENDPOINT")
				if endpoint == "" {
					endpoint = "http://host.docker.internal:11434"
				}
			}
			details += fmt.Sprintf("\n\nTroubleshooting:\n• Ensure Ollama is running at %s\n• Try running: curl %s/api/generate -d '{\"model\":\"%s\",\"prompt\":\"test\"}'", endpoint, endpoint, req.Model)
		}

		return TestConnectionResponse{
			Success: false,
			Message: "Connection failed",
			Details: details,
		}
	}

	// Success!
	return TestConnectionResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully connected to %s with model %s", req.Provider, req.Model),
	}
}
