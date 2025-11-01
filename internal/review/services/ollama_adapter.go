package review_services

import (
	"context"
	"fmt"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai/providers"
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

	// Construct ai.Request from simple prompt
	req := &ai.Request{
		Model:       "",             // Will use client's default model
		Prompt:      prompt,
		Temperature: 0.7,            // Default temperature for code analysis
		MaxTokens:   2048,           // Reasonable limit for analysis
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
