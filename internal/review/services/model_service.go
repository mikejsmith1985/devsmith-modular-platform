package review_services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// ModelInfo describes an available Ollama model
type ModelInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ModelService queries Ollama for available models
type ModelService struct {
	logger logger.Interface
}

// NewModelService creates a ModelService instance
func NewModelService(logger logger.Interface) *ModelService {
	return &ModelService{logger: logger}
}

// ListAvailableModels queries `ollama list` and returns available models
func (s *ModelService) ListAvailableModels(ctx context.Context) ([]ModelInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ollama", "list")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		s.logger.Error("Failed to run 'ollama list'", "error", err.Error(), "stderr", stderr.String())
		return s.fallbackModels(), fmt.Errorf("ollama list failed: %w", err)
	}

	output := stdout.String()
	lines := strings.Split(output, "\n")

	var models []ModelInfo
	for i, line := range lines {
		// Skip header line (first line typically contains column headers)
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// Parse line format: "NAME    ID    SIZE    MODIFIED"
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}

		modelName := fields[0]
		description := s.inferDescription(modelName)
		models = append(models, ModelInfo{
			Name:        modelName,
			Description: description,
		})
	}

	if len(models) == 0 {
		s.logger.Warn("No models detected from 'ollama list', using fallback list")
		return s.fallbackModels(), nil
	}

	s.logger.Info("Detected available models", "count", len(models))
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

// fallbackModels returns a hardcoded list when `ollama list` fails
func (s *ModelService) fallbackModels() []ModelInfo {
	return []ModelInfo{
		{Name: "mistral:7b-instruct", Description: "Fast, General (Recommended)"},
		{Name: "codellama:13b", Description: "Better for code"},
		{Name: "deepseek-coder:6.7b", Description: "Code specialist"},
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
