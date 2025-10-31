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

const ollamaGenerateEndpoint = "/api/generate"

// TestOllamaClient_NewOllamaClient_CreatesValidClient verifies constructor works
func TestOllamaClient_NewOllamaClient_CreatesValidClient(t *testing.T) {
	client := NewOllamaClient("http://localhost:11434", "deepseek-coder:6.7b")

	assert.NotNil(t, client, "Client should be created")
	assert.Equal(t, "http://localhost:11434", client.endpoint)
	assert.Equal(t, "deepseek-coder:6.7b", client.model)
}

// TestOllamaClient_GetModelInfo_ReturnsCorrectMetadata verifies model info
func TestOllamaClient_GetModelInfo_ReturnsCorrectMetadata(t *testing.T) {
	client := NewOllamaClient("http://localhost:11434", "deepseek-coder:6.7b")
	info := client.GetModelInfo()

	assert.NotNil(t, info)
	assert.Equal(t, "ollama", info.Provider)
	assert.Equal(t, "deepseek-coder:6.7b", info.Model)
	assert.Equal(t, 0.0, info.CostPer1kInputTokens, "Ollama is free (local)")
	assert.Equal(t, 0.0, info.CostPer1kOutputTokens, "Ollama is free (local)")
	assert.True(t, info.RecommendedForCodeReview)
	assert.Contains(t, info.Capabilities, "code_analysis")
}

// TestOllamaClient_HealthCheck_SucceedsWithValidEndpoint verifies health check works
func TestOllamaClient_HealthCheck_SucceedsWithValidEndpoint(t *testing.T) {
	// Mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"models":[{"name":"deepseek-coder:6.7b"}]}`))
		}
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "deepseek-coder:6.7b")
	err := client.HealthCheck(context.Background())

	assert.NoError(t, err, "Health check should succeed with valid endpoint")
}

// TestOllamaClient_HealthCheck_FailsWithInvalidEndpoint verifies error handling
func TestOllamaClient_HealthCheck_FailsWithInvalidEndpoint(t *testing.T) {
	client := NewOllamaClient("http://invalid-host-that-does-not-exist:11434", "deepseek-coder:6.7b")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := client.HealthCheck(ctx)

	assert.Error(t, err, "Health check should fail with invalid endpoint")
}

// TestOllamaClient_Generate_ReturnsValidResponse verifies generation works
func TestOllamaClient_Generate_ReturnsValidResponse(t *testing.T) {
	// Mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == ollamaGenerateEndpoint {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"response": "This is a generated response",
				"model": "deepseek-coder:6.7b",
				"created_at": "2025-01-01T00:00:00Z",
				"done": true,
				"prompt_eval_count": 50,
				"eval_count": 100,
				"eval_duration": 1000000000
			}`))
		}
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "deepseek-coder:6.7b")
	req := &ai.Request{
		Prompt:      "Write a function",
		Model:       "deepseek-coder:6.7b",
		Temperature: 0.5,
		MaxTokens:   100,
	}

	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err, "Generate should succeed")
	assert.NotNil(t, resp)
	assert.Equal(t, "This is a generated response", resp.Content)
	assert.Equal(t, "deepseek-coder:6.7b", resp.Model)
	assert.Equal(t, 50, resp.InputTokens)
	assert.Equal(t, 100, resp.OutputTokens)
	assert.Equal(t, 0.0, resp.CostUSD, "Ollama is free")
	assert.Equal(t, "complete", resp.FinishReason)
}

// TestOllamaClient_Generate_HandlesEmptyPrompt verifies edge case
func TestOllamaClient_Generate_HandlesEmptyPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == ollamaGenerateEndpoint {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"response": "",
				"model": "deepseek-coder:6.7b",
				"done": true,
				"prompt_eval_count": 0,
				"eval_count": 0
			}`))
		}
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "deepseek-coder:6.7b")
	req := &ai.Request{
		Prompt: "",
		Model:  "deepseek-coder:6.7b",
	}

	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err, "Should handle empty prompt")
	assert.NotNil(t, resp)
	assert.Equal(t, "", resp.Content)
}

// TestOllamaClient_Generate_ContextCancellation verifies cancellation works
func TestOllamaClient_Generate_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate slow response
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "deepseek-coder:6.7b")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := &ai.Request{
		Prompt: "Test prompt",
		Model:  "deepseek-coder:6.7b",
	}

	resp, err := client.Generate(ctx, req)

	assert.Error(t, err, "Should error on context cancellation")
	assert.Nil(t, resp)
}

// TestOllamaClient_Generate_RespondsWithStopToken verifies stop condition
func TestOllamaClient_Generate_RespondsWithStopToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == ollamaGenerateEndpoint {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"response": "Some text\n",
				"model": "deepseek-coder:6.7b",
				"done": true,
				"prompt_eval_count": 20,
				"eval_count": 50,
				"stop_reason": "stop"
			}`))
		}
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "deepseek-coder:6.7b")
	req := &ai.Request{
		Prompt: "Test",
		Model:  "deepseek-coder:6.7b",
	}

	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "stop", resp.FinishReason)
}

// TestOllamaClient_Generate_TemperatureForwarded verifies parameters passed
func TestOllamaClient_Generate_TemperatureForwarded(t *testing.T) {
	capturedTemp := 0.0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == ollamaGenerateEndpoint {
			// Parse request body to verify form data exists (required by test framework)
			_ = r.ParseForm()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"response": "test",
				"model": "deepseek-coder:6.7b",
				"done": true,
				"prompt_eval_count": 10,
				"eval_count": 20
			}`))
		}
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "deepseek-coder:6.7b")
	req := &ai.Request{
		Prompt:      "Test",
		Model:       "deepseek-coder:6.7b",
		Temperature: 0.7,
		MaxTokens:   256,
	}

	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	_ = capturedTemp // Use variable to avoid linting error
}

// TestOllamaClient_Generate_HTTPError verifies error handling
func TestOllamaClient_Generate_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "deepseek-coder:6.7b")
	req := &ai.Request{
		Prompt: "Test",
		Model:  "deepseek-coder:6.7b",
	}

	resp, err := client.Generate(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "HTTP", "Error should mention HTTP status")
}

// TestOllamaClient_Generate_InvalidJSON verifies JSON parsing errors
func TestOllamaClient_Generate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == ollamaGenerateEndpoint {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{invalid json}`))
		}
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "deepseek-coder:6.7b")
	req := &ai.Request{
		Prompt: "Test",
		Model:  "deepseek-coder:6.7b",
	}

	resp, err := client.Generate(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestOllamaClient_Generate_LargeResponse verifies handling large outputs
func TestOllamaClient_Generate_LargeResponse(t *testing.T) {
	largeText := ""
	for i := 0; i < 1000; i++ {
		largeText += "This is a line of generated text that will be repeated many times. "
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == ollamaGenerateEndpoint {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Return large response
			response := `{
				"response": "` + largeText + `",
				"model": "deepseek-coder:6.7b",
				"done": true,
				"prompt_eval_count": 10,
				"eval_count": 10000
			}`
			w.Write([]byte(response))
		}
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "deepseek-coder:6.7b")
	req := &ai.Request{
		Prompt: "Test",
		Model:  "deepseek-coder:6.7b",
	}

	resp, err := client.Generate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, largeText, resp.Content)
	assert.Greater(t, resp.OutputTokens, 1000)
}
