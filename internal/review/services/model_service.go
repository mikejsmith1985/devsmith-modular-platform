package review_services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// ModelInfo describes an available Ollama model
type ModelInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// OllamaModel represents a model from Ollama API response
type OllamaModel struct {
	Name       string `json:"name"`
	ModifiedAt string `json:"modified_at"`
	Size       int64  `json:"size"`
}

// OllamaTagsResponse is the response from GET /api/tags
type OllamaTagsResponse struct {
	Models []OllamaModel `json:"models"`
}

// ModelService queries Ollama for available models
type ModelService struct {
	logger         logger.Interface
	ollamaEndpoint string
}

// NewModelService creates a ModelService instance
func NewModelService(logger logger.Interface, ollamaEndpoint string) *ModelService {
	if ollamaEndpoint == "" {
		ollamaEndpoint = "http://localhost:11434"
	}
	return &ModelService{
		logger:         logger,
		ollamaEndpoint: ollamaEndpoint,
	}
}

// ListAvailableModels queries Ollama HTTP API and returns available models
func (s *ModelService) ListAvailableModels(ctx context.Context) ([]ModelInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Call Ollama API: GET /api/tags
	url := s.ollamaEndpoint + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.logger.Error("Failed to create request for Ollama API", "error", err.Error())
		return s.fallbackModels(), fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Failed to call Ollama API", "url", url, "error", err.Error())
		return s.fallbackModels(), fmt.Errorf("ollama API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Ollama API returned non-200 status", "status", resp.StatusCode)
		return s.fallbackModels(), fmt.Errorf("ollama API returned status %d", resp.StatusCode)
	}

	var tagsResp OllamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		s.logger.Error("Failed to decode Ollama API response", "error", err.Error())
		return s.fallbackModels(), fmt.Errorf("failed to decode response: %w", err)
	}

	var models []ModelInfo
	for _, model := range tagsResp.Models {
		description := s.inferDescription(model.Name)
		models = append(models, ModelInfo{
			Name:        model.Name,
			Description: description,
		})
	}

	if len(models) == 0 {
		s.logger.Warn("No models detected from Ollama API, using fallback list")
		return s.fallbackModels(), nil
	}

	s.logger.Info("Detected available models from Ollama", "count", len(models))
	return models, nil
}

// inferDescription provides user-friendly descriptions based on model name
func (s *ModelService) inferDescription(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "mistral") && strings.Contains(lower, "7b"):
		return "Fast, General (Recommended)"
	case strings.Contains(lower, "codellama"):
		return "Better for code"
	case strings.Contains(lower, "deepseek-coder-v2") && strings.Contains(lower, "16b"):
		return "Most accurate (slower)"
	case strings.Contains(lower, "deepseek-coder") && strings.Contains(lower, "6.7b"):
		return "Code specialist"
	case strings.Contains(lower, "qwen") && strings.Contains(lower, "coder"):
		return "Qwen coder model"
	case strings.Contains(lower, "llama"):
		return "Balanced general model"
	default:
		return "Available model"
	}
}

// fallbackModels returns a hardcoded list when Ollama API fails
// Only Mistral 7B is guaranteed to be available
func (s *ModelService) fallbackModels() []ModelInfo {
	return []ModelInfo{
		{Name: "mistral:7b-instruct", Description: "Fast, General (Recommended)"},
	}
}

// ListAvailableModelsJSON returns models as JSON (for API handler)
func (s *ModelService) ListAvailableModelsJSON(ctx context.Context) ([]byte, error) {
	models, err := s.ListAvailableModels(ctx)
	if err != nil {
		s.logger.Warn("Using fallback models due to error", "error", err.Error())
	}
	return json.Marshal(map[string]interface{}{"models": models})
}
