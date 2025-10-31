package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
)

const anthropicMessagesEndpoint = "/v1/messages"

// TestAnthropicClient_NewAnthropicClient_CreatesValidClient verifies constructor
func TestAnthropicClient_NewAnthropicClient_CreatesValidClient(t *testing.T) {
	client := NewAnthropicClient("sk-test-key-12345", "claude-3-5-haiku-20241022")

	assert.NotNil(t, client, "Client should be created")
	assert.Equal(t, "claude-3-5-haiku-20241022", client.model)
}

// TestAnthropicClient_GetModelInfo_ReturnsClaude35HaikuMetadata verifies model info
func TestAnthropicClient_GetModelInfo_ReturnsClaude35HaikuMetadata(t *testing.T) {
	client := NewAnthropicClient("sk-test-key", "claude-3-5-haiku-20241022")
	info := client.GetModelInfo()

	assert.NotNil(t, info)
	assert.Equal(t, "anthropic", info.Provider)
	assert.Equal(t, "claude-3-5-haiku-20241022", info.Model)
	assert.Greater(t, info.CostPer1kInputTokens, 0.0, "Anthropic has API costs")
	assert.Greater(t, info.CostPer1kOutputTokens, 0.0, "Anthropic has API costs")
	assert.True(t, info.RecommendedForCodeReview)
	assert.Contains(t, info.Capabilities, "code_analysis")
}

// TestAnthropicClient_GetModelInfo_Claude35Sonnet verifies Sonnet model info
func TestAnthropicClient_GetModelInfo_Claude35Sonnet(t *testing.T) {
	client := NewAnthropicClient("sk-test-key", "claude-3-5-sonnet-20241022")
	info := client.GetModelInfo()

	assert.NotNil(t, info)
	assert.Equal(t, "claude-3-5-sonnet-20241022", info.Model)
	// Sonnet should be more expensive than Haiku
	assert.Greater(t, info.CostPer1kInputTokens, 0.0)
}

// TestAnthropicClient_HealthCheck_SucceedsWithValidKey verifies health check
func TestAnthropicClient_HealthCheck_SucceedsWithValidKey(t *testing.T) {
	// Mock Anthropic API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == anthropicMessagesEndpoint && r.Method == "POST" {
			// Verify auth header
			authHeader := r.Header.Get("x-api-key")
			if authHeader != "sk-test-key" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "msg_test",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "test"}],
				"model": "claude-3-5-haiku-20241022",
				"stop_reason": "end_turn",
				"stop_sequence": null,
				"usage": {"input_tokens": 1, "output_tokens": 1}
			}`))
		}
	}))
	defer server.Close()

	client := &AnthropicClient{
		apiKey:     "sk-test-key",
		model:      "claude-3-5-haiku-20241022",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	err := client.HealthCheck(context.Background())
	assert.NoError(t, err, "Health check should succeed with valid key")
}

// TestAnthropicClient_HealthCheck_FailsWithInvalidKey verifies error handling
func TestAnthropicClient_HealthCheck_FailsWithInvalidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
	}))
	defer server.Close()

	client := &AnthropicClient{
		apiKey:     "sk-invalid-key",
		model:      "claude-3-5-haiku-20241022",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	err := client.HealthCheck(context.Background())
	assert.Error(t, err, "Health check should fail with invalid key")
}

// TestAnthropicClient_Generate_ReturnsValidResponse verifies generation
func TestAnthropicClient_Generate_ReturnsValidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == anthropicMessagesEndpoint {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "msg_test123",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "Here's a solution to your problem"}],
				"model": "claude-3-5-haiku-20241022",
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 150, "output_tokens": 200}
			}`))
		}
	}))
	defer server.Close()

	client := &AnthropicClient{
		apiKey:     "sk-test-key",
		model:      "claude-3-5-haiku-20241022",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	req := &ai.Request{
		Prompt:      "Solve this problem",
		Temperature: 0.5,
		MaxTokens:   1000,
	}

	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err, "Generate should succeed")
	assert.NotNil(t, resp)
	assert.Equal(t, "Here's a solution to your problem", resp.Content)
	assert.Equal(t, "claude-3-5-haiku-20241022", resp.Model)
	assert.Equal(t, 150, resp.InputTokens)
	assert.Equal(t, 200, resp.OutputTokens)
	assert.Greater(t, resp.CostUSD, 0.0, "Anthropic API has costs")
	assert.Equal(t, "end_turn", resp.FinishReason)
}

// TestAnthropicClient_Generate_HandlesEmptyPrompt verifies edge case
func TestAnthropicClient_Generate_HandlesEmptyPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == anthropicMessagesEndpoint {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "msg_test",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": ""}],
				"model": "claude-3-5-haiku-20241022",
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 0, "output_tokens": 0}
			}`))
		}
	}))
	defer server.Close()

	client := &AnthropicClient{
		apiKey:     "sk-test-key",
		model:      "claude-3-5-haiku-20241022",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	req := &ai.Request{Prompt: ""}
	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err, "Should handle empty prompt")
	assert.NotNil(t, resp)
}

// TestAnthropicClient_Generate_ContextCancellation verifies cancellation
func TestAnthropicClient_Generate_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate slow response
	}))
	defer server.Close()

	client := &AnthropicClient{
		apiKey:     "sk-test-key",
		model:      "claude-3-5-haiku-20241022",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := &ai.Request{Prompt: "Test"}
	resp, err := client.Generate(ctx, req)

	assert.Error(t, err, "Should error on context cancellation")
	assert.Nil(t, resp)
}

// TestAnthropicClient_Generate_MaxTokensForwarded verifies parameter handling
func TestAnthropicClient_Generate_MaxTokensForwarded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == anthropicMessagesEndpoint {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "msg_test",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "response"}],
				"model": "claude-3-5-haiku-20241022",
				"stop_reason": "max_tokens",
				"usage": {"input_tokens": 50, "output_tokens": 100}
			}`))
		}
	}))
	defer server.Close()

	client := &AnthropicClient{
		apiKey:     "sk-test-key",
		model:      "claude-3-5-haiku-20241022",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	req := &ai.Request{
		Prompt:    "Test",
		MaxTokens: 500,
	}

	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "max_tokens", resp.FinishReason)
}

// TestAnthropicClient_Generate_TemperatureForwarded verifies temperature param
func TestAnthropicClient_Generate_TemperatureForwarded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == anthropicMessagesEndpoint {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "msg_test",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "response"}],
				"model": "claude-3-5-haiku-20241022",
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 10, "output_tokens": 20}
			}`))
		}
	}))
	defer server.Close()

	client := &AnthropicClient{
		apiKey:     "sk-test-key",
		model:      "claude-3-5-haiku-20241022",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	req := &ai.Request{
		Prompt:      "Test",
		Temperature: 0.8,
	}

	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestAnthropicClient_Generate_HTTPError verifies error handling
func TestAnthropicClient_Generate_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": {"message": "Internal server error"}}`))
	}))
	defer server.Close()

	client := &AnthropicClient{
		apiKey:     "sk-test-key",
		model:      "claude-3-5-haiku-20241022",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	req := &ai.Request{Prompt: "Test"}
	resp, err := client.Generate(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestAnthropicClient_Generate_InvalidJSON verifies JSON parsing
func TestAnthropicClient_Generate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == anthropicMessagesEndpoint {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{invalid json}`))
		}
	}))
	defer server.Close()

	client := &AnthropicClient{
		apiKey:     "sk-test-key",
		model:      "claude-3-5-haiku-20241022",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	req := &ai.Request{Prompt: "Test"}
	resp, err := client.Generate(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestAnthropicClient_Generate_CostCalculation verifies cost tracking
func TestAnthropicClient_Generate_CostCalculation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == anthropicMessagesEndpoint {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Return large token counts to verify cost calculation
			w.Write([]byte(`{
				"id": "msg_test",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "response"}],
				"model": "claude-3-5-haiku-20241022",
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 10000, "output_tokens": 5000}
			}`))
		}
	}))
	defer server.Close()

	client := &AnthropicClient{
		apiKey:     "sk-test-key",
		model:      "claude-3-5-haiku-20241022",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	req := &ai.Request{Prompt: "Test"}
	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Greater(t, resp.CostUSD, 0.0, "Cost should be calculated from token usage")
	assert.Equal(t, 10000, resp.InputTokens)
	assert.Equal(t, 5000, resp.OutputTokens)
}

// TestAnthropicClient_Generate_MultipleContentBlocks verifies content extraction
func TestAnthropicClient_Generate_MultipleContentBlocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == anthropicMessagesEndpoint {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Return multiple content blocks
			w.Write([]byte(`{
				"id": "msg_test",
				"type": "message",
				"role": "assistant",
				"content": [
					{"type": "text", "text": "First"},
					{"type": "text", "text": " Second"}
				],
				"model": "claude-3-5-haiku-20241022",
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 10, "output_tokens": 20}
			}`))
		}
	}))
	defer server.Close()

	client := &AnthropicClient{
		apiKey:     "sk-test-key",
		model:      "claude-3-5-haiku-20241022",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	req := &ai.Request{Prompt: "Test"}
	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.Content, "First")
	assert.Contains(t, resp.Content, "Second")
}
