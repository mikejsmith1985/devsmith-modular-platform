// Package providers contains AI provider implementations for different services.
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

// AnthropicClient implements the Provider interface for Anthropic models
type AnthropicClient struct {
	httpClient *http.Client
	apiKey     string
	model      string
	apiBaseURL string
}

// anthropicRequest represents the JSON request sent to Anthropic API
type anthropicRequest struct {
	Model       string              `json:"model"`
	Messages    []map[string]string `json:"messages"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature float64             `json:"temperature,omitempty"`
}

// anthropicResponse represents the JSON response from Anthropic API
type anthropicResponse struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Role       string `json:"role"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Content    []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// modelPricing contains cost information for different Claude models
type modelPricing struct {
	inputCostPer1k  float64
	outputCostPer1k float64
}

var claudeModels = map[string]modelPricing{
	"claude-3-5-haiku-20241022": {
		inputCostPer1k:  0.00080, // $0.80 per 1M input tokens
		outputCostPer1k: 0.00240, // $2.40 per 1M output tokens
	},
	"claude-3-5-sonnet-20241022": {
		inputCostPer1k:  0.003, // $3.00 per 1M input tokens
		outputCostPer1k: 0.015, // $15.00 per 1M output tokens
	},
	"claude-3-opus-20250219": {
		inputCostPer1k:  0.015, // $15.00 per 1M input tokens
		outputCostPer1k: 0.075, // $75.00 per 1M output tokens
	},
}

// NewAnthropicClient creates a new Anthropic AI client
func NewAnthropicClient(apiKey, model string) *AnthropicClient {
	return &AnthropicClient{
		apiKey:     apiKey,
		model:      model,
		apiBaseURL: "https://api.anthropic.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Generate sends a prompt to Anthropic and returns the response
func (c *AnthropicClient) Generate(ctx context.Context, req *ai.Request) (*ai.Response, error) {
	// Prepare Anthropic request
	anthropicReq := anthropicRequest{
		Model: c.model,
		Messages: []map[string]string{
			{
				"role":    "user",
				"content": req.Prompt,
			},
		},
		Temperature: req.Temperature,
	}

	// Set MaxTokens if provided
	if req.MaxTokens > 0 {
		anthropicReq.MaxTokens = req.MaxTokens
	} else {
		// Default to a reasonable max if not specified
		anthropicReq.MaxTokens = 4096
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with context
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/v1/messages", c.apiBaseURL), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set required headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Execute request
	startTime := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Anthropic: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close() // error safe to ignore
	}()

	// Check HTTP status
	if httpResp.StatusCode != http.StatusOK {
		readErr := error(nil)
		bodyBytes, readErr := io.ReadAll(httpResp.Body)
		if readErr != nil {
			bodyBytes = []byte("(unable to read error body)")
		}
		return nil, fmt.Errorf("HTTP %d from Anthropic: %s", httpResp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var anthropicResp anthropicResponse
	if err := json.Unmarshal(bodyBytes, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

	// Extract text content from all content blocks
	var textContent strings.Builder
	for _, block := range anthropicResp.Content {
		if block.Type == "text" {
			textContent.WriteString(block.Text)
		}
	}

	// Calculate cost
	cost := c.calculateCost(anthropicResp.Usage.InputTokens, anthropicResp.Usage.OutputTokens)

	return &ai.Response{
		Content:      textContent.String(),
		InputTokens:  anthropicResp.Usage.InputTokens,
		OutputTokens: anthropicResp.Usage.OutputTokens,
		ResponseTime: time.Since(startTime),
		CostUSD:      cost,
		Model:        anthropicResp.Model,
		FinishReason: anthropicResp.StopReason,
	}, nil
}

// HealthCheck verifies that the API key is valid and can reach Anthropic
func (c *AnthropicClient) HealthCheck(ctx context.Context) error {
	// Create a minimal test request
	req := &ai.Request{
		Prompt:    "test",
		MaxTokens: 10,
	}

	_, err := c.Generate(ctx, req)
	if err != nil {
		// Check if it's an auth error
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "Unauthorized") {
			return fmt.Errorf("Anthropic authentication failed: invalid API key")
		}
		return fmt.Errorf("Anthropic health check failed: %w", err)
	}

	return nil
}

// GetModelInfo returns metadata about this Anthropic model
func (c *AnthropicClient) GetModelInfo() *ai.ModelInfo {
	pricing, exists := claudeModels[c.model]
	if !exists {
		// Default pricing if model not found
		pricing = modelPricing{
			inputCostPer1k:  0.003,
			outputCostPer1k: 0.015,
		}
	}

	return &ai.ModelInfo{
		Provider:                 "anthropic",
		Model:                    c.model,
		DisplayName:              fmt.Sprintf("Anthropic - %s", c.model),
		MaxTokens:                100000, // Claude models have very large context windows
		CostPer1kInputTokens:     pricing.inputCostPer1k,
		CostPer1kOutputTokens:    pricing.outputCostPer1k,
		Capabilities:             []string{"code_analysis", "code_review", "explanation", "documentation"},
		SupportsStreaming:        true,
		DefaultTemperature:       0.7,
		RecommendedForCodeReview: true,
	}
}

// calculateCost calculates the estimated API cost based on tokens used
func (c *AnthropicClient) calculateCost(inputTokens, outputTokens int) float64 {
	pricing, exists := claudeModels[c.model]
	if !exists {
		// Default pricing
		pricing = modelPricing{
			inputCostPer1k:  0.003,
			outputCostPer1k: 0.015,
		}
	}

	// Calculate cost: (tokens / 1000) * cost_per_1k_tokens
	inputCost := float64(inputTokens) / 1000.0 * pricing.inputCostPer1k
	outputCost := float64(outputTokens) / 1000.0 * pricing.outputCostPer1k

	return inputCost + outputCost
}
