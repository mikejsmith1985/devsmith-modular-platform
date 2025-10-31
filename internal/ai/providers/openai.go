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

// OpenAIClient implements the AIProvider interface for OpenAI models
type OpenAIClient struct {
	apiKey     string
	model      string
	apiBaseURL string
	httpClient *http.Client
}

// openaiRequest represents the JSON request sent to OpenAI API
type openaiRequest struct {
	Model       string  `json:"model"`
	Messages    []map[string]string `json:"messages"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// openaiResponse represents the JSON response from OpenAI API
type openaiResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// openaiModelPricing contains cost information for different GPT models
type openaiModelPricing struct {
	inputCostPer1k  float64
	outputCostPer1k float64
}

var gptModels = map[string]openaiModelPricing{
	"gpt-4-turbo": {
		inputCostPer1k:  0.01,   // $10.00 per 1M input tokens
		outputCostPer1k: 0.03,   // $30.00 per 1M output tokens
	},
	"gpt-4o": {
		inputCostPer1k:  0.005,  // $5.00 per 1M input tokens
		outputCostPer1k: 0.015,  // $15.00 per 1M output tokens
	},
	"gpt-4-32k": {
		inputCostPer1k:  0.06,   // $60.00 per 1M input tokens
		outputCostPer1k: 0.12,   // $120.00 per 1M output tokens
	},
}

// NewOpenAIClient creates a new OpenAI AI client
func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		apiKey:     apiKey,
		model:      model,
		apiBaseURL: "https://api.openai.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Generate sends a prompt to OpenAI and returns the response
func (c *OpenAIClient) Generate(ctx context.Context, req *ai.AIRequest) (*ai.AIResponse, error) {
	// Prepare OpenAI request
	openaiReq := openaiRequest{
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
		openaiReq.MaxTokens = req.MaxTokens
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with context
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/v1/chat/completions", c.apiBaseURL), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set required headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// Execute request
	startTime := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to OpenAI: %w", err)
	}
	defer httpResp.Body.Close()

	// Check HTTP status
	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("HTTP %d from OpenAI: %s", httpResp.StatusCode, string(bodyBytes))
	}

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var openaiResp openaiResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Extract first choice (we only use first choice)
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI response has no choices")
	}

	choice := openaiResp.Choices[0]

	// Calculate cost
	cost := c.calculateCost(openaiResp.Usage.PromptTokens, openaiResp.Usage.CompletionTokens)

	return &ai.AIResponse{
		Content:      choice.Message.Content,
		InputTokens:  openaiResp.Usage.PromptTokens,
		OutputTokens: openaiResp.Usage.CompletionTokens,
		ResponseTime: time.Since(startTime),
		CostUSD:      cost,
		Model:        openaiResp.Model,
		FinishReason: choice.FinishReason,
	}, nil
}

// HealthCheck verifies that the API key is valid and can reach OpenAI
func (c *OpenAIClient) HealthCheck(ctx context.Context) error {
	// Create a minimal test request
	req := &ai.AIRequest{
		Prompt:    "test",
		MaxTokens: 5,
	}

	_, err := c.Generate(ctx, req)
	if err != nil {
		// Check if it's an auth error
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "Unauthorized") {
			return fmt.Errorf("OpenAI authentication failed: invalid API key")
		}
		return fmt.Errorf("OpenAI health check failed: %w", err)
	}

	return nil
}

// GetModelInfo returns metadata about this OpenAI model
func (c *OpenAIClient) GetModelInfo() *ai.ModelInfo {
	pricing, exists := gptModels[c.model]
	if !exists {
		// Default pricing if model not found
		pricing = openaiModelPricing{
			inputCostPer1k:  0.01,
			outputCostPer1k: 0.03,
		}
	}

	return &ai.ModelInfo{
		Provider:                 "openai",
		Model:                    c.model,
		DisplayName:              fmt.Sprintf("OpenAI - %s", c.model),
		MaxTokens:                128000, // GPT-4 Turbo has 128K context
		CostPer1kInputTokens:     pricing.inputCostPer1k,
		CostPer1kOutputTokens:    pricing.outputCostPer1k,
		Capabilities:             []string{"code_analysis", "code_review", "explanation", "documentation"},
		SupportsStreaming:        true,
		DefaultTemperature:       0.7,
		RecommendedForCodeReview: true,
	}
}

// calculateCost calculates the estimated API cost based on tokens used
func (c *OpenAIClient) calculateCost(inputTokens, outputTokens int) float64 {
	pricing, exists := gptModels[c.model]
	if !exists {
		// Default pricing
		pricing = openaiModelPricing{
			inputCostPer1k:  0.01,
			outputCostPer1k: 0.03,
		}
	}

	// Calculate cost: (tokens / 1000) * cost_per_1k_tokens
	inputCost := float64(inputTokens) / 1000.0 * pricing.inputCostPer1k
	outputCost := float64(outputTokens) / 1000.0 * pricing.outputCostPer1k

	return inputCost + outputCost
}
