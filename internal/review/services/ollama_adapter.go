package review_services

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai/providers"
	reviewcontext "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/context"
)

const (
	defaultOllamaModel = "mistral:7b-instruct" // Fallback if context empty
)

// OllamaClientAdapter implements OllamaClientInterface by wrapping providers.OllamaClient
// This adapter bridges the gap between the complex ai.Request/Response interface
// and the simpler string-based interface used by review services.
type OllamaClientAdapter struct {
	client *providers.OllamaClient
}

// NewOllamaClientAdapter creates a new adapter wrapping an OllamaClient
func NewOllamaClientAdapter(client *providers.OllamaClient) *OllamaClientAdapter {
	return &OllamaClientAdapter{
		client: client,
	}
}

// Generate implements OllamaClientInterface.Generate
// Converts simple string prompt to ai.Request, calls wrapped client, returns string response
func (a *OllamaClientAdapter) Generate(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	if a.client == nil {
		return "", fmt.Errorf("ollama client is not initialized")
	}

	// Try to get model from context
	model, ok := ctx.Value(reviewcontext.ModelContextKey).(string)
	if !ok || model == "" {
		// Defensive fallback: use environment variable or default
		model = os.Getenv("OLLAMA_MODEL")
		if model == "" {
			model = defaultOllamaModel
		}
		// Log warning but continue
		log.Printf("Warning: model not in context, using fallback: %s", model)
	}

	// Construct ai.Request from simple prompt
	req := &ai.Request{
		Model:       model, // Use resolved model (from context, env, or default)
		Prompt:      prompt,
		Temperature: 0.7,  // Default temperature for code analysis
		MaxTokens:   2048, // Reasonable limit for analysis
	}

	// Call wrapped client
	resp, err := a.client.Generate(ctx, req)
	if err != nil {
		return "", fmt.Errorf("ollama generation failed: %w", err)
	}

	if resp == nil {
		return "", fmt.Errorf("ollama returned nil response")
	}

	return resp.Content, nil
}
