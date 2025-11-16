// Package testutils provides testing utilities and mocks for the platform.
package testutils

import (
	"context"
	"errors"
)

// MockOllamaClient provides a mock implementation of OllamaClientInterface for testing.
// It allows tests to control responses and inject errors without calling real Ollama service.
type MockOllamaClient struct {
	GenerateResponse string
	GenerateError    string
}

// Generate returns the mocked response or error.
// This allows tests to control Ollama behavior without a running service.
func (m *MockOllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	if m.GenerateError != "" {
		return "", errors.New(m.GenerateError)
	}
	return m.GenerateResponse, nil
}
