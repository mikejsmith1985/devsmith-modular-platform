package portal_repositories

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateLLMConfig_Success tests successful creation of LLM config
func TestCreateLLMConfig_Success(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil) // Will fail - no implementation yet

	config := &LLMConfig{
		UserID:          1,
		Provider:        "anthropic",
		ModelName:       "claude-3-5-sonnet-20241022",
		APIKeyEncrypted: sql.NullString{String: "encrypted_key_abc123", Valid: true},
		APIEndpoint:     sql.NullString{String: "https://api.anthropic.com", Valid: true},
		IsDefault:       true,
		MaxTokens:       4096,
		Temperature:     0.7,
	}

	// Act
	err := repo.Create(context.Background(), config)

	// Assert
	require.NoError(t, err, "Should create LLM config successfully")
	assert.NotEmpty(t, config.ID, "Should generate UUID for config")
	assert.NotZero(t, config.CreatedAt, "Should set created_at timestamp")
	assert.NotZero(t, config.UpdatedAt, "Should set updated_at timestamp")
}

// TestCreateLLMConfig_DuplicateProviderModel tests unique constraint
func TestCreateLLMConfig_DuplicateProviderModel(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	config1 := &LLMConfig{
		UserID:    1,
		Provider:  "openai",
		ModelName: "gpt-4-turbo-preview",
	}

	// Act - Create first config
	err1 := repo.Create(context.Background(), config1)
	require.NoError(t, err1)

	// Act - Try to create duplicate
	config2 := &LLMConfig{
		UserID:    1,
		Provider:  "openai",
		ModelName: "gpt-4-turbo-preview", // Same provider + model for same user
	}
	err2 := repo.Create(context.Background(), config2)

	// Assert
	require.Error(t, err2, "Should reject duplicate user+provider+model")
	assert.Contains(t, err2.Error(), "already exists")
}

// TestCreateLLMConfig_NullAPIKey tests Ollama without API key
func TestCreateLLMConfig_NullAPIKey(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	config := &LLMConfig{
		UserID:          1,
		Provider:        "ollama",
		ModelName:       "deepseek-coder:6.7b",
		APIKeyEncrypted: sql.NullString{Valid: false}, // NULL for Ollama
		APIEndpoint:     sql.NullString{String: "http://localhost:11434", Valid: true},
		IsDefault:       false,
	}

	// Act
	err := repo.Create(context.Background(), config)

	// Assert
	require.NoError(t, err, "Should create Ollama config without API key")
	assert.NotEmpty(t, config.ID)
	assert.False(t, config.APIKeyEncrypted.Valid, "API key should remain NULL")
}

// TestFindByID_Success tests retrieving config by ID
func TestFindByID_Success(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	created := &LLMConfig{
		UserID:    1,
		Provider:  "deepseek",
		ModelName: "deepseek-coder",
	}
	repo.Create(context.Background(), created)

	// Act
	found, err := repo.FindByID(context.Background(), created.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.Provider, found.Provider)
	assert.Equal(t, created.ModelName, found.ModelName)
}

// TestFindByID_NotFound tests missing config
func TestFindByID_NotFound(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	// Act
	found, err := repo.FindByID(context.Background(), "nonexistent-id-xyz")

	// Assert
	require.NoError(t, err, "FindByID should not error on missing record")
	assert.Nil(t, found, "Should return nil for missing config")
}

// TestFindByUser_Success tests listing all configs for a user
func TestFindByUser_Success(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	// Create multiple configs for user 1
	repo.Create(context.Background(), &LLMConfig{UserID: 1, Provider: "anthropic", ModelName: "claude"})
	repo.Create(context.Background(), &LLMConfig{UserID: 1, Provider: "openai", ModelName: "gpt-4"})
	repo.Create(context.Background(), &LLMConfig{UserID: 1, Provider: "ollama", ModelName: "deepseek"})

	// Create config for different user (should not appear)
	repo.Create(context.Background(), &LLMConfig{UserID: 2, Provider: "openai", ModelName: "gpt-3"})

	// Act
	configs, err := repo.FindByUser(context.Background(), 1)

	// Assert
	require.NoError(t, err)
	assert.Len(t, configs, 3, "Should return all 3 configs for user 1")
	for _, cfg := range configs {
		assert.Equal(t, 1, cfg.UserID, "All configs should belong to user 1")
	}
}

// TestFindDefaultByUser_Success tests retrieving default config
func TestFindDefaultByUser_Success(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	// Create non-default config
	repo.Create(context.Background(), &LLMConfig{
		UserID:    1,
		Provider:  "openai",
		ModelName: "gpt-3.5",
		IsDefault: false,
	})

	// Create default config
	defaultConfig := &LLMConfig{
		UserID:    1,
		Provider:  "anthropic",
		ModelName: "claude-3",
		IsDefault: true,
	}
	repo.Create(context.Background(), defaultConfig)

	// Act
	found, err := repo.FindDefaultByUser(context.Background(), 1)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, defaultConfig.ID, found.ID)
	assert.True(t, found.IsDefault)
}

// TestFindDefaultByUser_NoDefault tests when user has no default set
func TestFindDefaultByUser_NoDefault(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	// Create config without default
	repo.Create(context.Background(), &LLMConfig{
		UserID:    1,
		Provider:  "ollama",
		ModelName: "deepseek",
		IsDefault: false,
	})

	// Act
	found, err := repo.FindDefaultByUser(context.Background(), 1)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, found, "Should return nil when no default config exists")
}

// TestUpdate_Success tests updating existing config
func TestUpdate_Success(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	config := &LLMConfig{
		UserID:      1,
		Provider:    "openai",
		ModelName:   "gpt-4",
		MaxTokens:   4096,
		Temperature: 0.7,
	}
	repo.Create(context.Background(), config)

	// Act - Update fields
	config.MaxTokens = 8192
	config.Temperature = 0.5
	err := repo.Update(context.Background(), config)

	// Assert
	require.NoError(t, err)

	// Verify update persisted
	updated, _ := repo.FindByID(context.Background(), config.ID)
	assert.Equal(t, 8192, updated.MaxTokens)
	assert.Equal(t, 0.5, updated.Temperature)
	assert.True(t, updated.UpdatedAt.After(updated.CreatedAt), "Should update updated_at timestamp")
}

// TestDelete_Success tests deleting config
func TestDelete_Success(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	config := &LLMConfig{
		UserID:    1,
		Provider:  "openai",
		ModelName: "gpt-4",
	}
	repo.Create(context.Background(), config)

	// Act
	err := repo.Delete(context.Background(), config.ID)

	// Assert
	require.NoError(t, err)

	// Verify deletion
	found, _ := repo.FindByID(context.Background(), config.ID)
	assert.Nil(t, found, "Config should be deleted")
}

// TestSetDefault_OnlyOneDefault tests ensuring single default per user
func TestSetDefault_OnlyOneDefault(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	config1 := &LLMConfig{UserID: 1, Provider: "openai", ModelName: "gpt-4", IsDefault: true}
	config2 := &LLMConfig{UserID: 1, Provider: "anthropic", ModelName: "claude", IsDefault: false}
	repo.Create(context.Background(), config1)
	repo.Create(context.Background(), config2)

	// Act - Set config2 as default
	err := repo.SetDefault(context.Background(), 1, config2.ID)

	// Assert
	require.NoError(t, err)

	// Verify only config2 is default
	updated1, _ := repo.FindByID(context.Background(), config1.ID)
	updated2, _ := repo.FindByID(context.Background(), config2.ID)
	assert.False(t, updated1.IsDefault, "Previous default should be cleared")
	assert.True(t, updated2.IsDefault, "New config should be default")
}

// TestGetAppPreference_Success tests retrieving app-specific preference
func TestGetAppPreference_Success(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	// Create config
	config := &LLMConfig{UserID: 1, Provider: "anthropic", ModelName: "claude"}
	repo.Create(context.Background(), config)

	// Set app preference
	repo.SetAppPreference(context.Background(), 1, "review", config.ID)

	// Act
	pref, err := repo.GetAppPreference(context.Background(), 1, "review")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, pref)
	assert.Equal(t, 1, pref.UserID)
	assert.Equal(t, "review", pref.AppName)
	assert.Equal(t, config.ID, pref.LLMConfigID)
}

// TestGetAppPreference_NotSet tests when no preference exists
func TestGetAppPreference_NotSet(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	// Act
	pref, err := repo.GetAppPreference(context.Background(), 1, "logs")

	// Assert
	require.NoError(t, err)
	assert.Nil(t, pref, "Should return nil when no preference set")
}

// TestSetAppPreference_UpdatesExisting tests updating preference
func TestSetAppPreference_UpdatesExisting(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	config1 := &LLMConfig{UserID: 1, Provider: "openai", ModelName: "gpt-4"}
	config2 := &LLMConfig{UserID: 1, Provider: "anthropic", ModelName: "claude"}
	repo.Create(context.Background(), config1)
	repo.Create(context.Background(), config2)

	// Set initial preference
	repo.SetAppPreference(context.Background(), 1, "review", config1.ID)

	// Act - Update to different config
	err := repo.SetAppPreference(context.Background(), 1, "review", config2.ID)

	// Assert
	require.NoError(t, err)

	// Verify update
	pref, _ := repo.GetAppPreference(context.Background(), 1, "review")
	assert.Equal(t, config2.ID, pref.LLMConfigID)
}

// TestClearAppPreference_Success tests removing preference
func TestClearAppPreference_Success(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	config := &LLMConfig{UserID: 1, Provider: "openai", ModelName: "gpt-4"}
	repo.Create(context.Background(), config)
	repo.SetAppPreference(context.Background(), 1, "review", config.ID)

	// Act
	err := repo.ClearAppPreference(context.Background(), 1, "review")

	// Assert
	require.NoError(t, err)

	// Verify cleared
	pref, _ := repo.GetAppPreference(context.Background(), 1, "review")
	assert.Nil(t, pref, "Preference should be cleared")
}

// TestGetAllAppPreferences_Success tests listing all user preferences
func TestGetAllAppPreferences_Success(t *testing.T) {
	// Arrange
	repo := NewLLMConfigRepository(nil)

	config1 := &LLMConfig{UserID: 1, Provider: "openai", ModelName: "gpt-4"}
	config2 := &LLMConfig{UserID: 1, Provider: "anthropic", ModelName: "claude"}
	repo.Create(context.Background(), config1)
	repo.Create(context.Background(), config2)

	repo.SetAppPreference(context.Background(), 1, "review", config1.ID)
	repo.SetAppPreference(context.Background(), 1, "logs", config2.ID)

	// Create preference for different user (should not appear)
	config3 := &LLMConfig{UserID: 2, Provider: "ollama", ModelName: "deepseek"}
	repo.Create(context.Background(), config3)
	repo.SetAppPreference(context.Background(), 2, "review", config3.ID)

	// Act
	prefs, err := repo.GetAllAppPreferences(context.Background(), 1)

	// Assert
	require.NoError(t, err)
	assert.Len(t, prefs, 2, "Should return 2 preferences for user 1")

	// Verify correct apps
	appNames := []string{prefs[0].AppName, prefs[1].AppName}
	assert.Contains(t, appNames, "review")
	assert.Contains(t, appNames, "logs")
}
