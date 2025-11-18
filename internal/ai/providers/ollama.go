// Package providers contains AI provider implementations for different services.
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
)

// OllamaClient implements the Provider interface for local Ollama models
type OllamaClient struct {
	httpClient *http.Client
	endpoint   string
	model      string
}

// ollamaRequest represents the JSON request sent to Ollama API
type ollamaRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Stream      bool    `json:"stream"`
	Format      string  `json:"format,omitempty"` // Set to "json" to force JSON-only output
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

// ollamaResponse represents the JSON response from Ollama API
type ollamaResponse struct {
	Response           string `json:"response"`
	Model              string `json:"model"`
	StopReason         string `json:"stop_reason,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count"`
	EvalCount          int    `json:"eval_count"`
	EvalDuration       int64  `json:"eval_duration"`
	PromptEvalDuration int64  `json:"prompt_eval_duration"`
	Done               bool   `json:"done"`
}

// ollamaTagsResponse represents the response from /api/tags endpoint
type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// NewOllamaClient creates a new Ollama AI client
func NewOllamaClient(endpoint, model string) *OllamaClient {
	return &OllamaClient{
		endpoint: endpoint,
		model:    model,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Ollama can be slow
		},
	}
}

// Generate sends a prompt to Ollama and returns the response
func (c *OllamaClient) Generate(ctx context.Context, req *ai.Request) (*ai.Response, error) {
	// Prepare Ollama request
	ollamaReq := ollamaRequest{
		Model:       req.Model,
		Prompt:      req.Prompt,
		Stream:      false,
		Format:      "json", // CRITICAL: Force JSON-only output mode
		Temperature: req.Temperature,
	}

	// Set MaxTokens if provided
	if req.MaxTokens > 0 {
		ollamaReq.NumPredict = req.MaxTokens
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with context
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/generate", c.endpoint), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	startTime := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close() //nolint:errcheck // error after response processed
	}()

	// Check HTTP status
	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, readErr := io.ReadAll(httpResp.Body)
		if readErr != nil {
			bodyBytes = []byte("(unable to read error body)")
		}
		return nil, fmt.Errorf("HTTP %d from Ollama: %s", httpResp.StatusCode, string(bodyBytes))
	}

	bodyBytes, readErr := io.ReadAll(httpResp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body: %w", readErr)
	}

	// Parse JSON response
	var resp ollamaResponse
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	// Convert to Response
	finishReason := "complete"
	if resp.StopReason != "" {
		finishReason = resp.StopReason
	}

	return &ai.Response{
		Content:      resp.Response,
		InputTokens:  resp.PromptEvalCount,
		OutputTokens: resp.EvalCount,
		ResponseTime: time.Since(startTime),
		CostUSD:      0.0, // Ollama is local, no cost
		Model:        resp.Model,
		FinishReason: finishReason,
	}, nil
}

// HealthCheck verifies that Ollama is reachable and the model is available
func (c *OllamaClient) HealthCheck(ctx context.Context) error {
	// Create HTTP request to /api/tags to list available models
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/tags", c.endpoint), http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Execute request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("Ollama is unreachable: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close() //nolint:errcheck // error after response processed
	}()

	// Check status
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama health check returned HTTP %d", httpResp.StatusCode)
	}

	// Parse response to verify model is available
	var tagsResp ollamaTagsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&tagsResp); err != nil {
		return fmt.Errorf("failed to parse Ollama tags response: %w", err)
	}

	// Check if our model is in the list
	for _, m := range tagsResp.Models {
		if m.Name == c.model {
			return nil // Model found
		}
	}

	return fmt.Errorf("model %s not found in Ollama", c.model)
}

// GetModelInfo returns metadata about this Ollama model
func (c *OllamaClient) GetModelInfo() *ai.ModelInfo {
	return &ai.ModelInfo{
		Provider:                 "ollama",
		Model:                    c.model,
		DisplayName:              fmt.Sprintf("Ollama - %s", c.model),
		MaxTokens:                8192,
		CostPer1kInputTokens:     0.0, // Local - free
		CostPer1kOutputTokens:    0.0, // Local - free
		Capabilities:             []string{"code_analysis", "code_review", "explanation"},
		SupportsStreaming:        true,
		DefaultTemperature:       0.7,
		RecommendedForCodeReview: true,
	}
}
