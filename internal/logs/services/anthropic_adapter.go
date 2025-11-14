package logs_services

import (
	"context"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai/providers"
)

// AnthropicAdapter adapts the AnthropicClient to the AIProvider interface
type AnthropicAdapter struct {
	client *providers.AnthropicClient
}

// NewAnthropicAdapter creates a new adapter for AnthropicClient
func NewAnthropicAdapter(client *providers.AnthropicClient) AIProvider {
	return &AnthropicAdapter{client: client}
}

// Generate adapts the AnthropicClient Generate method to AIProvider interface
func (a *AnthropicAdapter) Generate(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	// Convert AIRequest to ai.Request
	aiReq := &ai.Request{
		Model:  request.Model,
		Prompt: request.Prompt,
	}

	// Call AnthropicClient
	aiResp, err := a.client.Generate(ctx, aiReq)
	if err != nil {
		return nil, err
	}

	// Convert ai.Response to AIResponse
	return &AIResponse{
		Content: aiResp.Content,
	}, nil
}
