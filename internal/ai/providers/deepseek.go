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

// DeepSeekClient implements the Provider interface for DeepSeek models
type DeepSeekClient struct {
	httpClient *http.Client
	apiKey     string
	model      string
	apiBaseURL string
}

// deepseekRequest represents the JSON request sent to DeepSeek API
type deepseekRequest struct {
	Model       string                   `json:"model"`
	Messages    []map[string]interface{} `json:"messages"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Temperature float64                  `json:"temperature,omitempty"`
}

// deepseekResponse represents the JSON response from DeepSeek API
type deepseekResponse struct {
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

// modelPricing contains cost information for different DeepSeek models
type deepseekPricing struct {
	inputCostPer1k  float64
	outputCostPer1k float64
}

var deepseekModels = map[string]deepseekPricing{
	"deepseek-chat": {
		inputCostPer1k:  0.00014, // $0.14 per 1M input tokens (cache miss)
		outputCostPer1k: 0.00028, // $0.28 per 1M output tokens
	},
	"deepseek-coder": {
		inputCostPer1k:  0.00014, // $0.14 per 1M input tokens (cache miss)
		outputCostPer1k: 0.00028, // $0.28 per 1M output tokens
	},
}

// NewDeepSeekClient creates a new DeepSeek AI client
func NewDeepSeekClient(apiKey, model string) *DeepSeekClient {
	return &DeepSeekClient{
		apiKey:     apiKey,
		model:      model,
		apiBaseURL: "https://api.deepseek.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Generate sends a prompt to DeepSeek and returns the response
func (c *DeepSeekClient) Generate(ctx context.Context, req *ai.Request) (*ai.Response, error) {
	// Prepare DeepSeek request (OpenAI-compatible format)
	deepseekReq := deepseekRequest{
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
		deepseekReq.MaxTokens = req.MaxTokens
	} else {
		// Default to a reasonable max if not specified
		deepseekReq.MaxTokens = 4096
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(deepseekReq)
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
		return nil, fmt.Errorf("failed to send request to DeepSeek: %w", err)
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
		return nil, fmt.Errorf("HTTP %d from DeepSeek: %s", httpResp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var deepseekResp deepseekResponse
	if err := json.Unmarshal(bodyBytes, &deepseekResp); err != nil {
		return nil, fmt.Errorf("failed to parse DeepSeek response: %w", err)
	}

	// Extract content from first choice
	content := ""
	finishReason := ""
	if len(deepseekResp.Choices) > 0 {
		content = deepseekResp.Choices[0].Message.Content
		finishReason = deepseekResp.Choices[0].FinishReason
	}

	// Calculate cost
	cost := c.calculateCost(deepseekResp.Usage.PromptTokens, deepseekResp.Usage.CompletionTokens)

	return &ai.Response{
		Content:      content,
		InputTokens:  deepseekResp.Usage.PromptTokens,
		OutputTokens: deepseekResp.Usage.CompletionTokens,
		ResponseTime: time.Since(startTime),
		CostUSD:      cost,
		Model:        deepseekResp.Model,
		FinishReason: finishReason,
	}, nil
}

// HealthCheck verifies that the API key is valid and can reach DeepSeek
func (c *DeepSeekClient) HealthCheck(ctx context.Context) error {
	// Create a minimal test request
	req := &ai.Request{
		Prompt:    "test",
		MaxTokens: 10,
	}

	_, err := c.Generate(ctx, req)
	if err != nil {
		// Check if it's an auth error
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "Unauthorized") {
			return fmt.Errorf("DeepSeek authentication failed: invalid API key")
		}
		return fmt.Errorf("DeepSeek health check failed: %w", err)
	}

	return nil
}

// GetModelInfo returns metadata about this DeepSeek model
func (c *DeepSeekClient) GetModelInfo() *ai.ModelInfo {
	pricing, exists := deepseekModels[c.model]
	if !exists {
		// Default pricing if model not found
		pricing = deepseekPricing{
			inputCostPer1k:  0.00014,
			outputCostPer1k: 0.00028,
		}
	}

	return &ai.ModelInfo{
		Provider:                 "deepseek",
		Model:                    c.model,
		DisplayName:              fmt.Sprintf("DeepSeek - %s", c.model),
		MaxTokens:                32000, // DeepSeek models support 32k context
		CostPer1kInputTokens:     pricing.inputCostPer1k,
		CostPer1kOutputTokens:    pricing.outputCostPer1k,
		Capabilities:             []string{"code_analysis", "code_review", "code_generation", "explanation"},
		SupportsStreaming:        true,
		DefaultTemperature:       0.7,
		RecommendedForCodeReview: true,
	}
}

// calculateCost calculates the estimated API cost based on tokens used
func (c *DeepSeekClient) calculateCost(inputTokens, outputTokens int) float64 {
	pricing, exists := deepseekModels[c.model]
	if !exists {
		// Default pricing
		pricing = deepseekPricing{
			inputCostPer1k:  0.00014,
			outputCostPer1k: 0.00028,
		}
	}

	// Calculate cost: (tokens / 1000) * cost_per_1k_tokens
	inputCost := float64(inputTokens) / 1000.0 * pricing.inputCostPer1k
	outputCost := float64(outputTokens) / 1000.0 * pricing.outputCostPer1k

	return inputCost + outputCost
}
