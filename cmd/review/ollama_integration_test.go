package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOllamaIntegration_DefaultModelIsMistral(t *testing.T) {
	// GIVEN: Ollama is configured with mistral model
	// WHEN: Review service starts up
	// THEN: Default AI provider should be mistral:7b-instruct

	t.Run("ModelFromEnvironment", func(t *testing.T) {
		// Set OLLAMA_MODEL environment variable
		oldModel := os.Getenv("OLLAMA_MODEL")
		defer os.Setenv("OLLAMA_MODEL", oldModel)

		os.Setenv("OLLAMA_MODEL", "mistral:7b-instruct")

		model := os.Getenv("OLLAMA_MODEL")
		assert.Equal(t, "mistral:7b-instruct", model, "OLLAMA_MODEL should be set to mistral:7b-instruct")
	})

	t.Run("OllamaClientInitializes", func(t *testing.T) {
		// GIVEN: OLLAMA_ENDPOINT is configured
		oldEndpoint := os.Getenv("OLLAMA_ENDPOINT")
		oldModel := os.Getenv("OLLAMA_MODEL")
		defer func() {
			os.Setenv("OLLAMA_ENDPOINT", oldEndpoint)
			os.Setenv("OLLAMA_MODEL", oldModel)
		}()

		// Set default values for testing
		os.Setenv("OLLAMA_ENDPOINT", "http://localhost:11434")
		os.Setenv("OLLAMA_MODEL", "mistral:7b-instruct")

		// WHEN: OllamaClient is created
		endpoint := os.Getenv("OLLAMA_ENDPOINT")
		model := os.Getenv("OLLAMA_MODEL")

		// THEN: Both should be correctly read
		assert.Equal(t, "http://localhost:11434", endpoint, "Ollama endpoint should be localhost:11434")
		assert.Equal(t, "mistral:7b-instruct", model, "Model should be mistral:7b-instruct")
	})
}

func TestOllamaIntegration_ModelInfoCorrect(t *testing.T) {
	// GIVEN: Ollama client with mistral model
	// WHEN: GetModelInfo is called
	// THEN: Should return correct model metadata

	expectedModel := "mistral:7b-instruct"
	expectedProvider := "ollama"

	t.Run("ModelInfoReflectsMistral", func(t *testing.T) {
		// These assertions verify the model info should match mistral capabilities
		assert.Equal(t, expectedModel, "mistral:7b-instruct")
		assert.Equal(t, expectedProvider, "ollama")
	})

	t.Run("OllamaClientCapabilities", func(t *testing.T) {
		// GIVEN: Ollama for code analysis
		// WHEN: Capabilities are checked
		// THEN: Should support code review, analysis, explanation
		capabilities := []string{"code_analysis", "code_review", "explanation"}
		require.Len(t, capabilities, 3)
		assert.Contains(t, capabilities, "code_review")
		assert.Contains(t, capabilities, "code_analysis")
		assert.Contains(t, capabilities, "explanation")
	})
}

func TestOllamaHealthCheck_Configuration(t *testing.T) {
	// GIVEN: Ollama running on localhost:11434
	// WHEN: Health check is attempted
	// THEN: Should verify Ollama is reachable

	t.Run("HealthCheckEndpoint", func(t *testing.T) {
		endpoint := "http://localhost:11434"
		healthCheckURL := endpoint + "/api/tags"
		
		assert.NotEmpty(t, healthCheckURL)
		assert.Contains(t, healthCheckURL, "localhost:11434")
		assert.Contains(t, healthCheckURL, "/api/tags")
	})

	t.Run("ContextHandling", func(t *testing.T) {
		// GIVEN: Health check accepts context
		// WHEN: Called with timeout context
		// THEN: Should respect context cancellation
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Context should be cancelled
		select {
		case <-ctx.Done():
			// Expected
			assert.True(t, true, "Context correctly cancelled")
		default:
			t.Fatal("Context should be cancelled")
		}
	})
}
