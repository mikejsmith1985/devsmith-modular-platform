package portal_services

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	portal_repositories "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/repositories"
)

// MockLLMConfigRepository is a mock implementation of the repository interface
type MockLLMConfigRepository struct {
	mock.Mock
}

func (m *MockLLMConfigRepository) Create(ctx context.Context, config *portal_repositories.LLMConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockLLMConfigRepository) FindByID(ctx context.Context, id string) (*portal_repositories.LLMConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*portal_repositories.LLMConfig), args.Error(1)
}

func (m *MockLLMConfigRepository) FindByUser(ctx context.Context, userID int) ([]*portal_repositories.LLMConfig, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*portal_repositories.LLMConfig), args.Error(1)
}

func (m *MockLLMConfigRepository) Update(ctx context.Context, config *portal_repositories.LLMConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockLLMConfigRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockLLMConfigRepository) FindDefaultByUser(ctx context.Context, userID int) (*portal_repositories.LLMConfig, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*portal_repositories.LLMConfig), args.Error(1)
}

func (m *MockLLMConfigRepository) SetDefault(ctx context.Context, userID int, configID string) error {
	args := m.Called(ctx, userID, configID)
	return args.Error(0)
}

func (m *MockLLMConfigRepository) GetAppPreference(ctx context.Context, userID int, appName string) (*portal_repositories.AppLLMPreference, error) {
	args := m.Called(ctx, userID, appName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*portal_repositories.AppLLMPreference), args.Error(1)
}

func (m *MockLLMConfigRepository) SetAppPreference(ctx context.Context, userID int, appName string, configID string) error {
	args := m.Called(ctx, userID, appName, configID)
	return args.Error(0)
}

func (m *MockLLMConfigRepository) ClearAppPreference(ctx context.Context, userID int, appName string) error {
	args := m.Called(ctx, userID, appName)
	return args.Error(0)
}

func (m *MockLLMConfigRepository) GetAllAppPreferences(ctx context.Context, userID int) ([]*portal_repositories.AppLLMPreference, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*portal_repositories.AppLLMPreference), args.Error(1)
}

// MockEncryptionServiceForService mocks the encryption service for service layer tests
type MockEncryptionServiceForService struct {
	mock.Mock
}

func (m *MockEncryptionServiceForService) EncryptAPIKey(plaintext string, userID int) (string, error) {
	args := m.Called(plaintext, userID)
	return args.String(0), args.Error(1)
}

func (m *MockEncryptionServiceForService) DecryptAPIKey(encrypted string, userID int) (string, error) {
	args := m.Called(encrypted, userID)
	return args.String(0), args.Error(1)
}

// TestCreateConfig_Success verifies successful config creation with API key encryption
func TestCreateConfig_Success(t *testing.T) {
	mockRepo := new(MockLLMConfigRepository)
	mockEncryption := new(MockEncryptionServiceForService)
	service := NewLLMConfigService(mockRepo, mockEncryption)

	ctx := context.Background()
	userID := 123
	provider := "anthropic"
	model := "claude-3-5-sonnet-20241022"
	apiKey := "test-api-key-123"

	// Mock encryption service
	mockEncryption.On("EncryptAPIKey", apiKey, 123).Return("encrypted-key-123", nil)

	// Mock repository - capture the config being created
	mockRepo.On("Create", ctx, mock.MatchedBy(func(cfg *portal_repositories.LLMConfig) bool {
		return cfg.UserID == userID &&
			cfg.Provider == provider &&
			cfg.ModelName == model &&
			cfg.APIKeyEncrypted.Valid &&
			cfg.APIKeyEncrypted.String == "encrypted-key-123"
	})).Return(nil)

	config, err := service.CreateConfig(ctx, userID, provider, model, apiKey, false, "")

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, userID, config.UserID)
	assert.Equal(t, provider, config.Provider)
	assert.Equal(t, model, config.ModelName)
	assert.True(t, config.APIKeyEncrypted.Valid)
	assert.Equal(t, "encrypted-key-123", config.APIKeyEncrypted.String)
	mockEncryption.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// TestCreateConfig_OllamaNoEncryption verifies Ollama configs skip encryption
func TestCreateConfig_OllamaNoEncryption(t *testing.T) {
	mockRepo := new(MockLLMConfigRepository)
	mockEncryption := new(MockEncryptionServiceForService)
	service := NewLLMConfigService(mockRepo, mockEncryption)

	ctx := context.Background()
	userID := 123
	provider := "ollama"
	model := "deepseek-coder:6.7b"
	endpoint := "http://localhost:11434"

	// Mock repository - verify NULL API key
	mockRepo.On("Create", ctx, mock.MatchedBy(func(cfg *portal_repositories.LLMConfig) bool {
		return cfg.UserID == userID &&
			cfg.Provider == provider &&
			cfg.ModelName == model &&
			!cfg.APIKeyEncrypted.Valid && // NULL for Ollama
			cfg.APIEndpoint.Valid &&
			cfg.APIEndpoint.String == endpoint
	})).Return(nil)

	config, err := service.CreateConfig(ctx, userID, provider, model, "", false, endpoint)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "ollama", config.Provider)
	assert.False(t, config.APIKeyEncrypted.Valid)      // NULL = no encryption
	mockEncryption.AssertNotCalled(t, "EncryptAPIKey") // Encryption not called for Ollama
	mockRepo.AssertExpectations(t)
}

// TestCreateConfig_EncryptionFails verifies error handling when encryption fails
func TestCreateConfig_EncryptionFails(t *testing.T) {
	mockRepo := new(MockLLMConfigRepository)
	mockEncryption := new(MockEncryptionServiceForService)
	service := NewLLMConfigService(mockRepo, mockEncryption)

	ctx := context.Background()
	mockEncryption.On("EncryptAPIKey", "test-key", 123).Return("", fmt.Errorf("encryption error"))

	config, err := service.CreateConfig(ctx, 123, "anthropic", "claude", "test-key", false, "")

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to encrypt API key")
	mockRepo.AssertNotCalled(t, "Create") // Repository not called if encryption fails
}

// TestCreateConfig_RepositoryFails verifies error handling when repository fails
func TestCreateConfig_RepositoryFails(t *testing.T) {
	mockRepo := new(MockLLMConfigRepository)
	mockEncryption := new(MockEncryptionServiceForService)
	service := NewLLMConfigService(mockRepo, mockEncryption)

	ctx := context.Background()
	mockEncryption.On("EncryptAPIKey", "test-key", 123).Return("encrypted", nil)
	mockRepo.On("Create", ctx, mock.Anything).Return(fmt.Errorf("repository error"))

	config, err := service.CreateConfig(ctx, 123, "anthropic", "claude", "test-key", false, "")

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to create config")
}

// TestUpdateConfig_ReencryptsAPIKey verifies API key updates require re-encryption
func TestUpdateConfig_ReencryptsAPIKey(t *testing.T) {
	mockRepo := new(MockLLMConfigRepository)
	mockEncryption := new(MockEncryptionServiceForService)
	service := NewLLMConfigService(mockRepo, mockEncryption)

	ctx := context.Background()
	configID := "test-config-id"
	userID := 123
	newAPIKey := "new-api-key-456"

	// Existing config
	existingConfig := &portal_repositories.LLMConfig{
		ID:              configID,
		UserID:          userID,
		Provider:        "anthropic",
		ModelName:       "claude-3",
		APIKeyEncrypted: sql.NullString{String: "old-encrypted", Valid: true},
	}

	mockRepo.On("FindByID", ctx, configID).Return(existingConfig, nil)
	mockEncryption.On("EncryptAPIKey", newAPIKey, 123).Return("new-encrypted-key", nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(cfg *portal_repositories.LLMConfig) bool {
		return cfg.ID == configID &&
			cfg.APIKeyEncrypted.String == "new-encrypted-key"
	})).Return(nil)

	err := service.UpdateConfig(ctx, userID, configID, map[string]interface{}{
		"api_key": newAPIKey,
	})

	assert.NoError(t, err)
	mockEncryption.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// TestUpdateConfig_OwnershipValidation verifies users can only update their own configs
func TestUpdateConfig_OwnershipValidation(t *testing.T) {
	mockRepo := new(MockLLMConfigRepository)
	mockEncryption := new(MockEncryptionServiceForService)
	service := NewLLMConfigService(mockRepo, mockEncryption)

	ctx := context.Background()
	configID := "test-config-id"
	actualOwnerID := 123
	requestingUserID := 456 // Different user!

	existingConfig := &portal_repositories.LLMConfig{
		ID:       configID,
		UserID:   actualOwnerID,
		Provider: "anthropic",
	}

	mockRepo.On("FindByID", ctx, configID).Return(existingConfig, nil)

	err := service.UpdateConfig(ctx, requestingUserID, configID, map[string]interface{}{
		"model_name": "new-model",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
	mockRepo.AssertNotCalled(t, "Update") // Update not called for wrong user
}

// TestDeleteConfig_Success verifies successful deletion with ownership check
func TestDeleteConfig_Success(t *testing.T) {
	mockRepo := new(MockLLMConfigRepository)
	mockEncryption := new(MockEncryptionServiceForService)
	service := NewLLMConfigService(mockRepo, mockEncryption)

	ctx := context.Background()
	configID := "test-config-id"
	userID := 123

	existingConfig := &portal_repositories.LLMConfig{
		ID:     configID,
		UserID: userID,
	}

	mockRepo.On("FindByID", ctx, configID).Return(existingConfig, nil)
	mockRepo.On("Delete", ctx, configID).Return(nil)

	err := service.DeleteConfig(ctx, userID, configID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestSetDefaultConfig_Success verifies setting default config
func TestSetDefaultConfig_Success(t *testing.T) {
	mockRepo := new(MockLLMConfigRepository)
	mockEncryption := new(MockEncryptionServiceForService)
	service := NewLLMConfigService(mockRepo, mockEncryption)

	ctx := context.Background()
	configID := "test-config-id"
	userID := 123

	existingConfig := &portal_repositories.LLMConfig{
		ID:     configID,
		UserID: userID,
	}

	mockRepo.On("FindByID", ctx, configID).Return(existingConfig, nil)
	mockRepo.On("SetDefault", ctx, userID, configID).Return(nil)

	err := service.SetDefaultConfig(ctx, userID, configID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestGetEffectiveConfig_AppPreference verifies app-specific preference takes priority
func TestGetEffectiveConfig_AppPreference(t *testing.T) {
	mockRepo := new(MockLLMConfigRepository)
	mockEncryption := new(MockEncryptionServiceForService)
	service := NewLLMConfigService(mockRepo, mockEncryption)

	ctx := context.Background()
	userID := 123
	appName := "review"
	configID := "app-specific-config"

	appPreference := &portal_repositories.AppLLMPreference{
		UserID:      userID,
		AppName:     appName,
		LLMConfigID: configID,
	}

	appConfig := &portal_repositories.LLMConfig{
		ID:        configID,
		UserID:    userID,
		Provider:  "anthropic",
		ModelName: "claude-3-5-sonnet",
	}

	mockRepo.On("GetAppPreference", ctx, userID, appName).Return(appPreference, nil)
	mockRepo.On("FindByID", ctx, configID).Return(appConfig, nil)

	config, err := service.GetEffectiveConfig(ctx, userID, appName)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, configID, config.ID)
	assert.Equal(t, "anthropic", config.Provider)
	mockRepo.AssertNotCalled(t, "FindDefaultByUser") // Shouldn't check default if app pref exists
}

// TestGetEffectiveConfig_UserDefault verifies fallback to user default when no app preference
func TestGetEffectiveConfig_UserDefault(t *testing.T) {
	mockRepo := new(MockLLMConfigRepository)
	mockEncryption := new(MockEncryptionServiceForService)
	service := NewLLMConfigService(mockRepo, mockEncryption)

	ctx := context.Background()
	userID := 123
	appName := "logs"

	defaultConfig := &portal_repositories.LLMConfig{
		ID:        "default-config",
		UserID:    userID,
		Provider:  "openai",
		ModelName: "gpt-4",
		IsDefault: true,
	}

	mockRepo.On("GetAppPreference", ctx, userID, appName).Return(nil, nil) // No app preference
	mockRepo.On("FindDefaultByUser", ctx, userID).Return(defaultConfig, nil)

	config, err := service.GetEffectiveConfig(ctx, userID, appName)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "default-config", config.ID)
	assert.Equal(t, "openai", config.Provider)
	assert.True(t, config.IsDefault)
}

// TestGetEffectiveConfig_SystemDefault verifies final fallback to Ollama system default
func TestGetEffectiveConfig_SystemDefault(t *testing.T) {
	mockRepo := new(MockLLMConfigRepository)
	mockEncryption := new(MockEncryptionServiceForService)
	service := NewLLMConfigService(mockRepo, mockEncryption)

	ctx := context.Background()
	userID := 123
	appName := "analytics"

	mockRepo.On("GetAppPreference", ctx, userID, appName).Return(nil, nil) // No app preference
	mockRepo.On("FindDefaultByUser", ctx, userID).Return(nil, nil)         // No user default

	config, err := service.GetEffectiveConfig(ctx, userID, appName)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "ollama", config.Provider)
	assert.Equal(t, "deepseek-coder:6.7b", config.ModelName)
	assert.False(t, config.APIKeyEncrypted.Valid) // NULL for Ollama
	// Verify system default endpoint
	if config.APIEndpoint.Valid {
		assert.Equal(t, "http://localhost:11434", config.APIEndpoint.String)
	}
}

// TestSetAppPreference_ValidatesConfig verifies config existence before setting preference
func TestSetAppPreference_ValidatesConfig(t *testing.T) {
	mockRepo := new(MockLLMConfigRepository)
	mockEncryption := new(MockEncryptionServiceForService)
	service := NewLLMConfigService(mockRepo, mockEncryption)

	ctx := context.Background()
	userID := 123
	appName := "review"
	configID := "nonexistent-config"

	mockRepo.On("FindByID", ctx, configID).Return(nil, nil) // Config doesn't exist

	err := service.SetAppPreference(ctx, userID, appName, configID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config not found")
	mockRepo.AssertNotCalled(t, "SetAppPreference") // Preference not set for invalid config
}

// TestListUserConfigs_Success verifies listing all user configs
func TestListUserConfigs_Success(t *testing.T) {
	mockRepo := new(MockLLMConfigRepository)
	mockEncryption := new(MockEncryptionServiceForService)
	service := NewLLMConfigService(mockRepo, mockEncryption)

	ctx := context.Background()
	userID := 123

	expectedConfigs := []*portal_repositories.LLMConfig{
		{
			ID:        "config-1",
			UserID:    userID,
			Provider:  "anthropic",
			ModelName: "claude-3-5-sonnet",
			IsDefault: true,
			CreatedAt: time.Now(),
		},
		{
			ID:        "config-2",
			UserID:    userID,
			Provider:  "openai",
			ModelName: "gpt-4",
			IsDefault: false,
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
	}

	mockRepo.On("FindByUser", ctx, userID).Return(expectedConfigs, nil)

	configs, err := service.ListUserConfigs(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, configs, 2)
	assert.Equal(t, "anthropic", configs[0].Provider)
	assert.Equal(t, "openai", configs[1].Provider)
	mockRepo.AssertExpectations(t)
}
