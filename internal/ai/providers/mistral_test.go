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

const mistralChatEndpoint = "/v1/chat/completions"

// TestMistralClient_NewMistralClient_CreatesValidClient verifies constructor
func TestMistralClient_NewMistralClient_CreatesValidClient(t *testing.T) {
	client := NewMistralClient("sk-test-key-12345", "mistral-large-latest")

	assert.NotNil(t, client, "Client should be created")
	assert.Equal(t, "mistral-large-latest", client.model)
}

// TestMistralClient_GetModelInfo_ReturnsMistralMetadata verifies model info
func TestMistralClient_GetModelInfo_ReturnsMistralMetadata(t *testing.T) {
	client := NewMistralClient("sk-test-key", "mistral-large-latest")
	info := client.GetModelInfo()

	assert.NotNil(t, info)
	assert.Equal(t, "mistral", info.Provider)
	assert.Equal(t, "mistral-large-latest", info.Model)
	assert.Greater(t, info.CostPer1kInputTokens, 0.0, "Mistral has API costs")
	assert.Greater(t, info.CostPer1kOutputTokens, 0.0, "Mistral has API costs")
	assert.True(t, info.RecommendedForCodeReview)
	assert.Contains(t, info.Capabilities, "code_analysis")
}

// TestMistralClient_HealthCheck_SucceedsWithValidKey verifies health check
func TestMistralClient_HealthCheck_SucceedsWithValidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mistralChatEndpoint && r.Method == "POST" {
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
				"model": "mistral-large-latest",
				"choices": [{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "test"
					},
					"finish_reason": "stop"
				}],
				"usage": {
					"prompt_tokens": 1,
					"completion_tokens": 1,
					"total_tokens": 2
				}
			}`))
		}
	}))
	defer server.Close()

	client := &MistralClient{
		apiKey:     "sk-test-key",
		model:      "mistral-large-latest",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	err := client.HealthCheck(context.Background())
	assert.NoError(t, err, "Health check should succeed with valid key")
}

// TestMistralClient_HealthCheck_FailsWithInvalidKey verifies error handling
func TestMistralClient_HealthCheck_FailsWithInvalidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
	}))
	defer server.Close()

	client := &MistralClient{
		apiKey:     "sk-invalid-key",
		model:      "mistral-large-latest",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	err := client.HealthCheck(context.Background())
	assert.Error(t, err, "Health check should fail with invalid key")
}

// TestMistralClient_Generate_ReturnsValidResponse verifies generation
func TestMistralClient_Generate_ReturnsValidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mistralChatEndpoint {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "chatcmpl_test123",
				"object": "chat.completion",
				"model": "mistral-large-latest",
				"choices": [{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "Here's a solution to your problem"
					},
					"finish_reason": "stop"
				}],
				"usage": {
					"prompt_tokens": 150,
					"completion_tokens": 200,
					"total_tokens": 350
				}
			}`))
		}
	}))
	defer server.Close()

	client := &MistralClient{
		apiKey:     "sk-test-key",
		model:      "mistral-large-latest",
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
	assert.Equal(t, "mistral-large-latest", resp.Model)
	assert.Equal(t, 150, resp.InputTokens)
	assert.Equal(t, 200, resp.OutputTokens)
	assert.Greater(t, resp.CostUSD, 0.0, "Mistral API has costs")
	assert.Equal(t, "stop", resp.FinishReason)
}

// TestMistralClient_Generate_APIError verifies error handling
func TestMistralClient_Generate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": {"message": "Internal server error"}}`))
	}))
	defer server.Close()

	client := &MistralClient{
		apiKey:     "sk-test-key",
		model:      "mistral-large-latest",
		apiBaseURL: server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	req := &ai.Request{
		Prompt:      "Test prompt",
		Temperature: 0.7,
		MaxTokens:   100,
	}

	_, err := client.Generate(context.Background(), req)
	assert.Error(t, err, "Generate should fail with API error")
}
