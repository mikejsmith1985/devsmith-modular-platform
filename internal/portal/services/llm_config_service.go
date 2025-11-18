package portal_services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	portal_repositories "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/repositories"
)

// ListOllamaModels returns the list of installed Ollama models by calling the Ollama HTTP API
func (s *LLMConfigService) ListOllamaModels(ctx context.Context) ([]string, error) {
	// Use context for timeout (10s)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get Ollama endpoint from environment or use default
	ollamaEndpoint := os.Getenv("OLLAMA_ENDPOINT")
	if ollamaEndpoint == "" {
		ollamaEndpoint = "http://host.docker.internal:11434"
	}

	// Call Ollama HTTP API to list models
	req, err := http.NewRequestWithContext(ctx, "GET", ollamaEndpoint+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	models := make([]string, len(result.Models))
	for i, model := range result.Models {
		models[i] = model.Name
	}
	return models, nil
}

// Error messages for configuration validation
const (
	errConfigNotFound       = "config not found"
	errPermissionDenied     = "permission denied: config does not belong to user"
	errFailedToFindConfig   = "failed to find config"
	errFailedToEncrypt      = "failed to encrypt API key"
	errFailedToSaveConfig   = "failed to save config"
	errFailedToUpdateConfig = "failed to update config"
	errFailedToDeleteConfig = "failed to delete config"
	errFailedToSetDefault   = "failed to set default config"
	errFailedToSetPref      = "failed to set app preference"
	errFailedToListConfigs  = "failed to list configs"
)

// LLMConfigService provides business logic for managing LLM configurations
type LLMConfigService struct {
	repo       portal_repositories.LLMConfigRepository
	encryption EncryptionServiceInterface
}

// NewLLMConfigService creates a new LLM configuration service
func NewLLMConfigService(
	repo portal_repositories.LLMConfigRepository,
	encryption EncryptionServiceInterface,
) *LLMConfigService {
	return &LLMConfigService{
		repo:       repo,
		encryption: encryption,
	}
}

// validateConfigOwnership checks if a config exists and belongs to the specified user
// Returns nil if validation passes, error otherwise
func (s *LLMConfigService) validateConfigOwnership(
	ctx context.Context,
	configID string,
	userID int,
) (*portal_repositories.LLMConfig, error) {
	config, err := s.repo.FindByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errFailedToFindConfig, err)
	}
	if config == nil {
		return nil, fmt.Errorf("%s", errConfigNotFound)
	}
	if config.UserID != userID {
		return nil, fmt.Errorf("%s", errPermissionDenied)
	}
	return config, nil
}

// CreateConfig creates a new LLM configuration for a user
// For Ollama, apiKey should be empty string and encryption is skipped
// For other providers, apiKey is encrypted before storage
func (s *LLMConfigService) CreateConfig(
	ctx context.Context,
	userID int,
	provider string,
	model string,
	apiKey string,
	isDefault bool,
	endpoint string,
) (*portal_repositories.LLMConfig, error) {
	// Generate UUID for config
	configID := uuid.New().String()
	now := time.Now()

	// Create config struct with defaults
	config := &portal_repositories.LLMConfig{
		ID:          configID,
		UserID:      userID,
		Provider:    provider,
		ModelName:   model,
		IsDefault:   isDefault,
		MaxTokens:   4096, // Default max tokens
		Temperature: 0.7,  // Default temperature
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Set endpoint if provided
	if endpoint != "" {
		config.APIEndpoint = sql.NullString{String: endpoint, Valid: true}
	}

	// Encrypt API key if not Ollama and key provided
	if provider != "ollama" && apiKey != "" {
		encrypted, err := s.encryption.EncryptAPIKey(apiKey, userID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errFailedToEncrypt, err)
		}
		config.APIKeyEncrypted = sql.NullString{String: encrypted, Valid: true}
	} else {
		// Ollama or no API key - set to NULL
		config.APIKeyEncrypted = sql.NullString{Valid: false}
	}

	// Save to repository
	if err := s.repo.Create(ctx, config); err != nil {
		return nil, fmt.Errorf("%s: %w", errFailedToSaveConfig, err)
	}

	return config, nil
}

// UpdateConfig updates an existing LLM configuration
// Updates are provided as a map for flexibility
// If "api_key" is in updates and provider != "ollama", the key is re-encrypted
func (s *LLMConfigService) UpdateConfig(
	ctx context.Context,
	userID int,
	configID string,
	updates map[string]interface{},
) error {
	// Validate ownership
	existing, err := s.validateConfigOwnership(ctx, configID, userID)
	if err != nil {
		return err
	}

	// Apply updates using helper methods to reduce complexity
	if err := s.applyProviderUpdates(existing, updates); err != nil {
		return err
	}

	if err := s.applyDefaultUpdates(ctx, existing, updates, userID, configID); err != nil {
		return err
	}

	if err := s.applyAPIKeyUpdates(existing, updates, userID); err != nil {
		return err
	}

	// Update timestamp
	existing.UpdatedAt = time.Now()

	// Save to repository
	if err := s.repo.Update(ctx, existing); err != nil {
		return fmt.Errorf("%s: %w", errFailedToUpdateConfig, err)
	}

	return nil
}

// DeleteConfig removes a configuration from the system
// Validates that the config belongs to the requesting user
func (s *LLMConfigService) DeleteConfig(
	ctx context.Context,
	userID int,
	configID string,
) error {
	// Validate ownership before deletion
	if _, err := s.validateConfigOwnership(ctx, configID, userID); err != nil {
		return err
	}

	// Delete from repository
	if err := s.repo.Delete(ctx, configID); err != nil {
		return fmt.Errorf("%s: %w", errFailedToDeleteConfig, err)
	}

	return nil
}

// SetDefaultConfig marks a configuration as the user's default
// Validates that the config belongs to the requesting user
func (s *LLMConfigService) SetDefaultConfig(
	ctx context.Context,
	userID int,
	configID string,
) error {
	// Validate ownership before setting default
	// Validate ownership before setting default
	if _, err := s.validateConfigOwnership(ctx, configID, userID); err != nil {
		return err
	}

	// Set as default via repository
	if err := s.repo.SetDefault(ctx, userID, configID); err != nil {
		return fmt.Errorf("%s: %w", errFailedToSetDefault, err)
	}

	return nil
}

// GetEffectiveConfig returns the effective LLM configuration for a user and app
// Priority order:
// 1. App-specific preference (if set)
// 2. User's default configuration (if set)
// 3. System default (Ollama with deepseek-coder:6.7b)
func (s *LLMConfigService) GetEffectiveConfig(
	ctx context.Context,
	userID int,
	appName string,
) (*portal_repositories.LLMConfig, error) {
	// Priority 1: Check app-specific preference
	appPref, appErr := s.repo.GetAppPreference(ctx, userID, appName)
	if appErr == nil && appPref != nil {
		config, configErr := s.repo.FindByID(ctx, appPref.LLMConfigID)
		if configErr == nil && config != nil {
			return config, nil
		}
	}

	// Priority 2: Check user's default configuration
	defaultConfig, err := s.repo.FindDefaultByUser(ctx, userID)
	if err == nil && defaultConfig != nil {
		return defaultConfig, nil
	}

	// Priority 3: Return system default (Ollama with deepseek-coder:6.7b)
	// Use OLLAMA_ENDPOINT environment variable for Docker compatibility
	ollamaEndpoint := os.Getenv("OLLAMA_ENDPOINT")
	if ollamaEndpoint == "" {
		ollamaEndpoint = "http://host.docker.internal:11434" // Default for Docker
	}

	systemDefault := &portal_repositories.LLMConfig{
		ID:              "system-default",
		UserID:          userID,
		Provider:        "ollama",
		ModelName:       "deepseek-coder:6.7b",
		APIEndpoint:     sql.NullString{String: ollamaEndpoint, Valid: true},
		APIKeyEncrypted: sql.NullString{Valid: false}, // NULL for Ollama
		IsDefault:       false,
		MaxTokens:       8192,
		Temperature:     0.7,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	return systemDefault, nil
}

// SetAppPreference sets the preferred LLM configuration for a specific app
// Validates that the config belongs to the requesting user
func (s *LLMConfigService) SetAppPreference(
	ctx context.Context,
	userID int,
	appName string,
	configID string,
) error {
	// Validate config exists and belongs to user
	if _, err := s.validateConfigOwnership(ctx, configID, userID); err != nil {
		return err
	}

	// Set app preference via repository
	if err := s.repo.SetAppPreference(ctx, userID, appName, configID); err != nil {
		return fmt.Errorf("failed to set app preference: %w", err)
	}

	return nil
}

// ListUserConfigs returns all LLM configurations for a user
func (s *LLMConfigService) ListUserConfigs(
	ctx context.Context,
	userID int,
) ([]*portal_repositories.LLMConfig, error) {
	configs, err := s.repo.FindByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errFailedToListConfigs, err)
	}
	return configs, nil
}

// GetConfigByID returns a single LLM configuration by ID (includes ownership check)
// NOTE: API key is decrypted for use in connection testing
func (s *LLMConfigService) GetConfigByID(
	ctx context.Context,
	userID int,
	configID string,
) (*portal_repositories.LLMConfig, error) {
	config, err := s.repo.FindByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errFailedToFindConfig, err)
	}

	// Verify ownership
	if config.UserID != userID {
		return nil, fmt.Errorf(errPermissionDenied)
	}

	// Decrypt API key if present (needed for connection testing)
	if config.APIKeyEncrypted.Valid && config.APIKeyEncrypted.String != "" {
		decryptedKey, err := s.encryption.DecryptAPIKey(config.APIKeyEncrypted.String, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt API key: %w", err)
		}
		config.APIKeyEncrypted.String = decryptedKey // Replace encrypted with decrypted
	}

	return config, nil
}

// applyProviderUpdates applies provider and model updates to the config
func (s *LLMConfigService) applyProviderUpdates(existing *portal_repositories.LLMConfig, updates map[string]interface{}) error {
	if provider, ok := updates["provider"]; ok {
		if providerStr, ok := provider.(string); ok {
			existing.Provider = providerStr
		}
	}
	if model, ok := updates["model_name"]; ok {
		if modelStr, ok := model.(string); ok {
			existing.ModelName = modelStr
		}
	}
	if endpoint, ok := updates["endpoint"]; ok {
		if endpointStr, ok := endpoint.(string); ok && endpointStr != "" {
			existing.APIEndpoint = sql.NullString{String: endpointStr, Valid: true}
		}
	}
	return nil
}

// applyDefaultUpdates handles is_default flag updates with atomic transaction
// nolint:nestif // complexity is due to necessary type assertion and default handling logic
func (s *LLMConfigService) applyDefaultUpdates(ctx context.Context, existing *portal_repositories.LLMConfig, updates map[string]interface{}, userID int, configID string) error {
	if isDefault, ok := updates["is_default"]; ok {
		if isDefaultBool, ok := isDefault.(bool); ok {
			// If setting this config as default, use the SetDefault method which
			// handles unsetting other configs atomically in a transaction
			if isDefaultBool {
				if err := s.repo.SetDefault(ctx, userID, configID); err != nil {
					return fmt.Errorf("failed to set as default: %w", err)
				}
				// SetDefault already updated the database, so we just update the in-memory object
				existing.IsDefault = true
			} else {
				// If explicitly setting is_default to false, just update this config
				existing.IsDefault = false
			}
		}
	}
	return nil
}

// applyAPIKeyUpdates handles API key encryption and updates
// nolint:nestif // complexity is due to necessary type assertion and encryption logic
func (s *LLMConfigService) applyAPIKeyUpdates(existing *portal_repositories.LLMConfig, updates map[string]interface{}, userID int) error {
	if apiKey, ok := updates["api_key"]; ok {
		if apiKeyStr, ok := apiKey.(string); ok {
			if existing.Provider != "ollama" && apiKeyStr != "" {
				encrypted, err := s.encryption.EncryptAPIKey(apiKeyStr, userID)
				if err != nil {
					return fmt.Errorf("%s: %w", errFailedToEncrypt, err)
				}
				existing.APIKeyEncrypted = sql.NullString{String: encrypted, Valid: true}
			} else {
				existing.APIKeyEncrypted = sql.NullString{Valid: false}
			}
		}
	}
	return nil
}
