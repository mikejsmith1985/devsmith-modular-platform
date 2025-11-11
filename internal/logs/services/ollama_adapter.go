package logs_services

import (
	"context"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai/providers"
)

// OllamaAdapter adapts the OllamaClient to the AIProvider interface
type OllamaAdapter struct {
	client *providers.OllamaClient
}

// NewOllamaAdapter creates a new adapter for OllamaClient
func NewOllamaAdapter(client *providers.OllamaClient) AIProvider {
	return &OllamaAdapter{client: client}
}

// Generate adapts the OllamaClient Generate method to AIProvider interface
func (a *OllamaAdapter) Generate(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	// Convert AIRequest to ai.Request
	aiReq := &ai.Request{
		Model:  request.Model,
		Prompt: request.Prompt,
	}

	// Call OllamaClient
	aiResp, err := a.client.Generate(ctx, aiReq)
	if err != nil {
		return nil, err
	}

	// Convert ai.Response to AIResponse
	return &AIResponse{
		Content: aiResp.Content,
	}, nil
}
