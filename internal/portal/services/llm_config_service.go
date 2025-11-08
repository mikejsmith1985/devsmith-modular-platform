package portal_services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	portal_repositories "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/repositories"
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

	// Create config struct
	config := &portal_repositories.LLMConfig{
		ID:        configID,
		UserID:    userID,
		Provider:  provider,
		ModelName: model,
		IsDefault: isDefault,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Set endpoint if provided
	if endpoint != "" {
		config.APIEndpoint = sql.NullString{String: endpoint, Valid: true}
	}

	// Encrypt API key if not Ollama and key provided
	if provider != "ollama" && apiKey != "" {
		encrypted, err := s.encryption.EncryptAPIKey(apiKey, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API key: %w", err)
		}
		config.APIKeyEncrypted = sql.NullString{String: encrypted, Valid: true}
	} else {
		// Ollama or no API key - set to NULL
		config.APIKeyEncrypted = sql.NullString{Valid: false}
	}

	// Save to repository
	if err := s.repo.Create(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
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
	// Fetch existing config
	existing, err := s.repo.FindByID(ctx, configID)
	if err != nil {
		return fmt.Errorf("failed to find config: %w", err)
	}

	// Validate ownership
	if existing.UserID != userID {
		return fmt.Errorf("permission denied: config does not belong to user")
	}

	// Apply updates
	if provider, ok := updates["provider"]; ok {
		existing.Provider = provider.(string)
	}
	if model, ok := updates["model_name"]; ok {
		existing.ModelName = model.(string)
	}
	if endpoint, ok := updates["endpoint"]; ok {
		if endpointStr, ok := endpoint.(string); ok && endpointStr != "" {
			existing.APIEndpoint = sql.NullString{String: endpointStr, Valid: true}
		}
	}
	if isDefault, ok := updates["is_default"]; ok {
		existing.IsDefault = isDefault.(bool)
	}

	// Handle API key update with re-encryption
	if apiKey, ok := updates["api_key"]; ok {
		apiKeyStr := apiKey.(string)
		if existing.Provider != "ollama" && apiKeyStr != "" {
			encrypted, err := s.encryption.EncryptAPIKey(apiKeyStr, userID)
			if err != nil {
				return fmt.Errorf("failed to encrypt API key: %w", err)
			}
			existing.APIKeyEncrypted = sql.NullString{String: encrypted, Valid: true}
		} else {
			existing.APIKeyEncrypted = sql.NullString{Valid: false}
		}
	}

	// Update timestamp
	existing.UpdatedAt = time.Now()

	// Save to repository
	if err := s.repo.Update(ctx, existing); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
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
	config, err := s.repo.FindByID(ctx, configID)
	if err != nil {
		return fmt.Errorf("failed to find config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("config not found")
	}

	// Validate ownership
	if config.UserID != userID {
		return fmt.Errorf("permission denied: config does not belong to user")
	}

	// Delete from repository
	if err := s.repo.Delete(ctx, configID); err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
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
	config, err := s.repo.FindByID(ctx, configID)
	if err != nil {
		return fmt.Errorf("failed to find config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("config not found")
	}

	// Validate ownership
	if config.UserID != userID {
		return fmt.Errorf("permission denied: config does not belong to user")
	}

	// Set as default via repository
	if err := s.repo.SetDefault(ctx, userID, configID); err != nil {
		return fmt.Errorf("failed to set default config: %w", err)
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
	appPref, err := s.repo.GetAppPreference(ctx, userID, appName)
	if err == nil && appPref != nil {
		config, err := s.repo.FindByID(ctx, appPref.LLMConfigID)
		if err == nil && config != nil {
			return config, nil
		}
	}

	// Priority 2: Check user's default configuration
	defaultConfig, err := s.repo.FindDefaultByUser(ctx, userID)
	if err == nil && defaultConfig != nil {
		return defaultConfig, nil
	}

	// Priority 3: Return system default (Ollama)
	systemDefault := &portal_repositories.LLMConfig{
		ID:              "system-default-ollama",
		UserID:          0, // System config
		Provider:        "ollama",
		ModelName:       "deepseek-coder:6.7b",
		APIEndpoint:     sql.NullString{String: "http://localhost:11434", Valid: true},
		APIKeyEncrypted: sql.NullString{Valid: false}, // No API key for Ollama
		IsDefault:       true,
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
	config, err := s.repo.FindByID(ctx, configID)
	if err != nil {
		return fmt.Errorf("failed to find config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("config not found")
	}

	// Validate ownership
	if config.UserID != userID {
		return fmt.Errorf("permission denied: config does not belong to user")
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
		return nil, fmt.Errorf("failed to list user configs: %w", err)
	}
	return configs, nil
}
