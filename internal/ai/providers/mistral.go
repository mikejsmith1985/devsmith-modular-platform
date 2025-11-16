// Package providers contains AI provider implementations for different services.
//
// Architecture Note: Mistral is an API-based provider (unlike local Ollama).
// API keys are encrypted at rest using AES-256-GCM encryption (see internal/portal/services/encryption_service.go).
// The factory/service layer decrypts API keys before passing them to this client.
// Local providers like Ollama do not require API keys or encryption.
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
)

// MistralClient implements the Provider interface for Mistral models.
// This client requires an API key for authentication, unlike local models (Ollama).
// API keys should be encrypted when stored in the database using AES-256-GCM.
type MistralClient struct {
	httpClient *http.Client
	apiKey     string
	model      string
	apiBaseURL string
}

// mistralRequest represents the JSON request sent to Mistral API
type mistralRequest struct {
	Model       string                   `json:"model"`
	Messages    []map[string]interface{} `json:"messages"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Temperature float64                  `json:"temperature,omitempty"`
}

// mistralResponse represents the JSON response from Mistral API
type mistralResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// modelPricing contains cost information for different Mistral models
type mistralPricing struct {
	inputCostPer1k  float64
	outputCostPer1k float64
}

var mistralModels = map[string]mistralPricing{
	"mistral-large-latest": {
		inputCostPer1k:  0.002, // $2.00 per 1M input tokens
		outputCostPer1k: 0.006, // $6.00 per 1M output tokens
	},
	"mistral-small-latest": {
		inputCostPer1k:  0.0002, // $0.20 per 1M input tokens
		outputCostPer1k: 0.0006, // $0.60 per 1M output tokens
	},
	"codestral-latest": {
		inputCostPer1k:  0.0002, // $0.20 per 1M input tokens
		outputCostPer1k: 0.0006, // $0.60 per 1M output tokens
	},
}

// NewMistralClient creates a new Mistral AI client.
// Note: apiKey should be decrypted before passing to this constructor.
// API keys are stored encrypted in the database and decrypted by the
// factory/service layer before creating the client.
func NewMistralClient(apiKey, model string) *MistralClient {
	return &MistralClient{
		apiKey:     apiKey,
		model:      model,
		apiBaseURL: "https://api.mistral.ai",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Generate sends a prompt to Mistral and returns the response
func (c *MistralClient) Generate(ctx context.Context, req *ai.Request) (*ai.Response, error) {
	// Prepare Mistral request (OpenAI-compatible format)
	mistralReq := mistralRequest{
		Model: c.model,
		Messages: []map[string]interface{}{
			{
				"role":    "user",
				"content": req.Prompt,
			},
		},
		Temperature: req.Temperature,
	}

	// Set MaxTokens if provided
	if req.MaxTokens > 0 {
		mistralReq.MaxTokens = req.MaxTokens
	} else {
		// Default to a reasonable max if not specified
		mistralReq.MaxTokens = 4096
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(mistralReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with context
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/v1/chat/completions", c.apiBaseURL), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set required headers (OpenAI-compatible)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// Execute request
	startTime := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Mistral: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close() //nolint:errcheck // error after response processed
	}()

	// Check HTTP status
	if httpResp.StatusCode != http.StatusOK {
		readErr := error(nil)
		bodyBytes, readErr := io.ReadAll(httpResp.Body)
		if readErr != nil {
			bodyBytes = []byte("(unable to read error body)")
		}
		return nil, fmt.Errorf("HTTP %d from Mistral: %s", httpResp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var mistralResp mistralResponse
	if err := json.Unmarshal(bodyBytes, &mistralResp); err != nil {
		return nil, fmt.Errorf("failed to parse Mistral response: %w", err)
	}

	// Extract content from first choice
	content := ""
	finishReason := ""
	if len(mistralResp.Choices) > 0 {
		content = mistralResp.Choices[0].Message.Content
		finishReason = mistralResp.Choices[0].FinishReason
	}

	// Calculate cost
	cost := c.calculateCost(mistralResp.Usage.PromptTokens, mistralResp.Usage.CompletionTokens)

	return &ai.Response{
		Content:      content,
		InputTokens:  mistralResp.Usage.PromptTokens,
		OutputTokens: mistralResp.Usage.CompletionTokens,
		ResponseTime: time.Since(startTime),
		CostUSD:      cost,
		Model:        mistralResp.Model,
		FinishReason: finishReason,
	}, nil
}

// HealthCheck verifies that the API key is valid and can reach Mistral
func (c *MistralClient) HealthCheck(ctx context.Context) error {
	// Create a minimal test request
	req := &ai.Request{
		Prompt:    "test",
		MaxTokens: 10,
	}

	_, err := c.Generate(ctx, req)
	if err != nil {
		// Check if it's an auth error
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "Unauthorized") {
			return fmt.Errorf("Mistral authentication failed: invalid API key")
		}
		return fmt.Errorf("Mistral health check failed: %w", err)
	}

	return nil
}

// GetModelInfo returns metadata about this Mistral model
func (c *MistralClient) GetModelInfo() *ai.ModelInfo {
	pricing, exists := mistralModels[c.model]
	if !exists {
		// Default pricing if model not found (use large model pricing)
		pricing = mistralPricing{
			inputCostPer1k:  0.002,
			outputCostPer1k: 0.006,
		}
	}

	return &ai.ModelInfo{
		Provider:                 "mistral",
		Model:                    c.model,
		DisplayName:              fmt.Sprintf("Mistral - %s", c.model),
		MaxTokens:                32000, // Mistral models support 32k context
		CostPer1kInputTokens:     pricing.inputCostPer1k,
		CostPer1kOutputTokens:    pricing.outputCostPer1k,
		Capabilities:             []string{"code_analysis", "code_review", "code_generation", "explanation"},
		SupportsStreaming:        true,
		DefaultTemperature:       0.7,
		RecommendedForCodeReview: true,
	}
}

// calculateCost calculates the estimated API cost based on tokens used
func (c *MistralClient) calculateCost(inputTokens, outputTokens int) float64 {
	pricing, exists := mistralModels[c.model]
	if !exists {
		// Default pricing
		pricing = mistralPricing{
			inputCostPer1k:  0.002,
			outputCostPer1k: 0.006,
		}
	}

	// Calculate cost: (tokens / 1000) * cost_per_1k_tokens
	inputCost := float64(inputTokens) / 1000.0 * pricing.inputCostPer1k
	outputCost := float64(outputTokens) / 1000.0 * pricing.outputCostPer1k

	return inputCost + outputCost
}
