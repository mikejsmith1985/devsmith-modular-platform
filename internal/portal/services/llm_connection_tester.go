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
	// Create context with timeout - increased to 90s for large Ollama models
	// Large models (16b+) can take 60-90 seconds to load on first use
	testCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
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
		// Check if this was a timeout error
		isTimeout := false
		if err == context.DeadlineExceeded || testCtx.Err() == context.DeadlineExceeded {
			isTimeout = true
		}

		// For Ollama timeouts on large models, provide specific guidance
		if req.Provider == "ollama" && isTimeout {
			// Large models may need pre-warming - provide clear instructions
			endpoint := req.Endpoint
			if endpoint == "" {
				endpoint = os.Getenv("OLLAMA_ENDPOINT")
				if endpoint == "" {
					endpoint = "http://host.docker.internal:11434"
				}
			}

			return TestConnectionResponse{
				Success: false,
				Message: "Connection timed out",
				Details: fmt.Sprintf("Connection to Ollama timed out after 90 seconds.\n\nThis usually happens when testing large models (16b+) that need to be loaded into memory for the first time.\n\nTo fix this:\n1. Pre-load the model: ollama run %s 'test'\n2. Or test with a smaller model first (e.g., deepseek-coder:6.7b)\n3. Wait a few minutes for the model to fully load, then try again\n\nTroubleshooting:\n• Check Ollama is responding: curl %s/api/version\n• Verify model exists: ollama list | grep %s\n• Check Ollama logs for loading progress", req.Model, endpoint, req.Model),
			}
		}

		// Provide more helpful error messages based on error type
		details := fmt.Sprintf("Failed to connect to %s: %v", req.Provider, err)

		// Add helpful suggestions based on provider and error type
		if req.Provider == "ollama" && !isTimeout {
			// Non-timeout errors get standard troubleshooting
			endpoint := req.Endpoint
			if endpoint == "" {
				endpoint = os.Getenv("OLLAMA_ENDPOINT")
				if endpoint == "" {
					endpoint = "http://host.docker.internal:11434"
				}
			}
			details += fmt.Sprintf("\n\nTroubleshooting:\n• Ensure Ollama is running at %s\n• Verify model '%s' exists: ollama list | grep %s\n• Test connection: curl %s/api/generate -d '{\"model\":\"%s\",\"prompt\":\"test\"}'", endpoint, req.Model, req.Model, endpoint, req.Model)
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
