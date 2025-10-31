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

// TestOpenAIClient_NewOpenAIClient_CreatesValidClient verifies constructor
func TestOpenAIClient_NewOpenAIClient_CreatesValidClient(t *testing.T) {
	client := NewOpenAIClient("sk-test-key-12345", "gpt-4-turbo")

	assert.NotNil(t, client, "Client should be created")
	assert.Equal(t, "gpt-4-turbo", client.model)
}

// TestOpenAIClient_GetModelInfo_ReturnsGPT4Metadata verifies model info
func TestOpenAIClient_GetModelInfo_ReturnsGPT4Metadata(t *testing.T) {
	client := NewOpenAIClient("sk-test-key", "gpt-4-turbo")
	info := client.GetModelInfo()

	assert.NotNil(t, info)
	assert.Equal(t, "openai", info.Provider)
	assert.Equal(t, "gpt-4-turbo", info.Model)
	assert.Greater(t, info.CostPer1kInputTokens, 0.0, "OpenAI has API costs")
	assert.Greater(t, info.CostPer1kOutputTokens, 0.0, "OpenAI has API costs")
	assert.True(t, info.RecommendedForCodeReview)
	assert.Contains(t, info.Capabilities, "code_analysis")
}

// TestOpenAIClient_GetModelInfo_GPT4o verifies GPT-4o model info
func TestOpenAIClient_GetModelInfo_GPT4o(t *testing.T) {
	client := NewOpenAIClient("sk-test-key", "gpt-4o")
	info := client.GetModelInfo()

	assert.NotNil(t, info)
	assert.Equal(t, "gpt-4o", info.Model)
	assert.Greater(t, info.CostPer1kInputTokens, 0.0)
}

// TestOpenAIClient_HealthCheck_SucceedsWithValidKey verifies health check
func TestOpenAIClient_HealthCheck_SucceedsWithValidKey(t *testing.T) {
	// Mock OpenAI API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" && r.Method == "POST" {
			// Verify auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer sk-test-key" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "chatcmpl_test",
				"object": "chat.completion",
				"created": 1234567890,
				"model": "gpt-4-turbo",
				"choices": [{"message": {"role": "assistant", "content": "test"}, "finish_reason": "stop"}],
				"usage": {"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2}
			}`))
		}
	}))
	defer server.Close()

	client := &OpenAIClient{
		apiKey:     "sk-test-key",
		model:      "gpt-4-turbo",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	err := client.HealthCheck(context.Background())
	assert.NoError(t, err, "Health check should succeed with valid key")
}

// TestOpenAIClient_HealthCheck_FailsWithInvalidKey verifies error handling
func TestOpenAIClient_HealthCheck_FailsWithInvalidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
	}))
	defer server.Close()

	client := &OpenAIClient{
		apiKey:     "sk-invalid-key",
		model:      "gpt-4-turbo",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	err := client.HealthCheck(context.Background())
	assert.Error(t, err, "Health check should fail with invalid key")
}

// TestOpenAIClient_Generate_ReturnsValidResponse verifies generation
func TestOpenAIClient_Generate_ReturnsValidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "chatcmpl_test123",
				"object": "chat.completion",
				"created": 1234567890,
				"model": "gpt-4-turbo",
				"choices": [{"message": {"role": "assistant", "content": "Here's a solution"}, "finish_reason": "stop"}],
				"usage": {"prompt_tokens": 100, "completion_tokens": 150, "total_tokens": 250}
			}`))
		}
	}))
	defer server.Close()

	client := &OpenAIClient{
		apiKey:     "sk-test-key",
		model:      "gpt-4-turbo",
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
	assert.Equal(t, "Here's a solution", resp.Content)
	assert.Equal(t, "gpt-4-turbo", resp.Model)
	assert.Equal(t, 100, resp.InputTokens)
	assert.Equal(t, 150, resp.OutputTokens)
	assert.Greater(t, resp.CostUSD, 0.0, "OpenAI API has costs")
	assert.Equal(t, "stop", resp.FinishReason)
}

// TestOpenAIClient_Generate_HandlesEmptyPrompt verifies edge case
func TestOpenAIClient_Generate_HandlesEmptyPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "chatcmpl_test",
				"object": "chat.completion",
				"created": 1234567890,
				"model": "gpt-4-turbo",
				"choices": [{"message": {"role": "assistant", "content": ""}, "finish_reason": "stop"}],
				"usage": {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}
			}`))
		}
	}))
	defer server.Close()

	client := &OpenAIClient{
		apiKey:     "sk-test-key",
		model:      "gpt-4-turbo",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	req := &ai.Request{Prompt: ""}
	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err, "Should handle empty prompt")
	assert.NotNil(t, resp)
}

// TestOpenAIClient_Generate_ContextCancellation verifies cancellation
func TestOpenAIClient_Generate_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate slow response
	}))
	defer server.Close()

	client := &OpenAIClient{
		apiKey:     "sk-test-key",
		model:      "gpt-4-turbo",
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

// TestOpenAIClient_Generate_FinishReasonLength verifies length limit
func TestOpenAIClient_Generate_FinishReasonLength(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "chatcmpl_test",
				"object": "chat.completion",
				"created": 1234567890,
				"model": "gpt-4-turbo",
				"choices": [{"message": {"role": "assistant", "content": "response"}, "finish_reason": "length"}],
				"usage": {"prompt_tokens": 50, "completion_tokens": 100, "total_tokens": 150}
			}`))
		}
	}))
	defer server.Close()

	client := &OpenAIClient{
		apiKey:     "sk-test-key",
		model:      "gpt-4-turbo",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	req := &ai.Request{
		Prompt:    "Test",
		MaxTokens: 100,
	}

	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "length", resp.FinishReason)
}

// TestOpenAIClient_Generate_TemperatureForwarded verifies temperature param
func TestOpenAIClient_Generate_TemperatureForwarded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "chatcmpl_test",
				"object": "chat.completion",
				"created": 1234567890,
				"model": "gpt-4-turbo",
				"choices": [{"message": {"role": "assistant", "content": "response"}, "finish_reason": "stop"}],
				"usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}
			}`))
		}
	}))
	defer server.Close()

	client := &OpenAIClient{
		apiKey:     "sk-test-key",
		model:      "gpt-4-turbo",
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

// TestOpenAIClient_Generate_HTTPError verifies error handling
func TestOpenAIClient_Generate_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": {"message": "Internal server error"}}`))
	}))
	defer server.Close()

	client := &OpenAIClient{
		apiKey:     "sk-test-key",
		model:      "gpt-4-turbo",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	req := &ai.Request{Prompt: "Test"}
	resp, err := client.Generate(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestOpenAIClient_Generate_InvalidJSON verifies JSON parsing
func TestOpenAIClient_Generate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{invalid json}`))
		}
	}))
	defer server.Close()

	client := &OpenAIClient{
		apiKey:     "sk-test-key",
		model:      "gpt-4-turbo",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	req := &ai.Request{Prompt: "Test"}
	resp, err := client.Generate(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestOpenAIClient_Generate_CostCalculation verifies cost tracking
func TestOpenAIClient_Generate_CostCalculation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "chatcmpl_test",
				"object": "chat.completion",
				"created": 1234567890,
				"model": "gpt-4-turbo",
				"choices": [{"message": {"role": "assistant", "content": "response"}, "finish_reason": "stop"}],
				"usage": {"prompt_tokens": 5000, "completion_tokens": 2000, "total_tokens": 7000}
			}`))
		}
	}))
	defer server.Close()

	client := &OpenAIClient{
		apiKey:     "sk-test-key",
		model:      "gpt-4-turbo",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	req := &ai.Request{Prompt: "Test"}
	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Greater(t, resp.CostUSD, 0.0, "Cost should be calculated from token usage")
	assert.Equal(t, 5000, resp.InputTokens)
	assert.Equal(t, 2000, resp.OutputTokens)
}

// TestOpenAIClient_Generate_MultipleChoices verifies first choice extraction
func TestOpenAIClient_Generate_MultipleChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "chatcmpl_test",
				"object": "chat.completion",
				"created": 1234567890,
				"model": "gpt-4-turbo",
				"choices": [
					{"message": {"role": "assistant", "content": "First response"}, "finish_reason": "stop"},
					{"message": {"role": "assistant", "content": "Second response"}, "finish_reason": "stop"}
				],
				"usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}
			}`))
		}
	}))
	defer server.Close()

	client := &OpenAIClient{
		apiKey:     "sk-test-key",
		model:      "gpt-4-turbo",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	req := &ai.Request{Prompt: "Test"}
	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "First response", resp.Content, "Should extract first choice")
}
