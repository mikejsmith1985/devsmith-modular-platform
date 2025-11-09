package portal_services

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewClientFactory_CreatesValidFactory tests factory construction
func TestNewClientFactory_CreatesValidFactory(t *testing.T) {
	factory := NewClientFactory(nil, nil) // Pass nil dependencies for now

	require.NotNil(t, factory, "Factory should be created")
}

// TestClientFactory_GetClientForApp_ReturnsClientForUserPreference tests preference lookup
func TestClientFactory_GetClientForApp_ReturnsClientForUserPreference(t *testing.T) {
	// Mock dependencies
	mockConfigService := &MockLLMConfigService{
		configs: map[string]*LLMConfig{
			"user1-review": {
				UserID:      1,
				AppName:     "review",
				Provider:    "deepseek",
				Model:       "deepseek-chat",
				APIKey:      "encrypted-key-123",
				APIEndpoint: "",
				IsDefault:   true,
			},
		},
	}
	mockEncryptionService := &MockEncryptionService{
		decryptedKeys: map[string]string{
			"encrypted-key-123": "sk-deepseek-actual-key",
		},
	}

	factory := NewClientFactory(mockConfigService, mockEncryptionService)

	// Act
	client, err := factory.GetClientForApp(context.Background(), 1, "review")

	// Assert
	require.NoError(t, err, "Should successfully get client")
	require.NotNil(t, client, "Client should be returned")

	modelInfo := client.GetModelInfo()
	assert.Equal(t, "deepseek", modelInfo.Provider)
	assert.Equal(t, "deepseek-chat", modelInfo.Model)
}

// TestClientFactory_GetClientForApp_FallsBackToOllama tests fallback when no preference
func TestClientFactory_GetClientForApp_FallsBackToOllama(t *testing.T) {
	// Mock with no user preferences
	mockConfigService := &MockLLMConfigService{
		configs: map[string]*LLMConfig{},
	}
	mockEncryptionService := &MockEncryptionService{}

	factory := NewClientFactory(mockConfigService, mockEncryptionService)

	// Act
	client, err := factory.GetClientForApp(context.Background(), 999, "review")

	// Assert
	require.NoError(t, err, "Should fall back to Ollama")
	require.NotNil(t, client, "Ollama client should be returned")

	modelInfo := client.GetModelInfo()
	assert.Equal(t, "ollama", modelInfo.Provider)
}

// TestClientFactory_GetClientForApp_DecryptsAPIKey tests conditional decryption
func TestClientFactory_GetClientForApp_DecryptsAPIKey(t *testing.T) {
	// Mock with API-based provider (requires decryption)
	mockConfigService := &MockLLMConfigService{
		configs: map[string]*LLMConfig{
			"user1-review": {
				UserID:      1,
				AppName:     "review",
				Provider:    "anthropic",
				Model:       "claude-3-5-haiku-20241022",
				APIKey:      "encrypted-anthropic-key",
				APIEndpoint: "",
				IsDefault:   true,
			},
		},
	}

	decryptionCalled := false
	mockEncryptionService := &MockEncryptionService{
		decryptedKeys: map[string]string{
			"encrypted-anthropic-key": "sk-ant-actual-key",
		},
		onDecrypt: func(encrypted string, userID int) {
			decryptionCalled = true
			assert.Equal(t, "encrypted-anthropic-key", encrypted)
			assert.Equal(t, 1, userID)
		},
	}

	factory := NewClientFactory(mockConfigService, mockEncryptionService)

	// Act
	client, err := factory.GetClientForApp(context.Background(), 1, "review")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.True(t, decryptionCalled, "DecryptAPIKey should be called for API providers")
}

// TestClientFactory_GetClientForApp_SkipsDecryptionForOllama tests Ollama doesn't decrypt
func TestClientFactory_GetClientForApp_SkipsDecryptionForOllama(t *testing.T) {
	// Mock with Ollama (should NOT decrypt)
	mockConfigService := &MockLLMConfigService{
		configs: map[string]*LLMConfig{
			"user1-review": {
				UserID:      1,
				AppName:     "review",
				Provider:    "ollama",
				Model:       "deepseek-coder:6.7b",
				APIKey:      "", // NULL for Ollama
				APIEndpoint: "http://localhost:11434",
				IsDefault:   true,
			},
		},
	}

	decryptionCalled := false
	mockEncryptionService := &MockEncryptionService{
		onDecrypt: func(encrypted string, userID int) {
			decryptionCalled = true
		},
	}

	factory := NewClientFactory(mockConfigService, mockEncryptionService)

	// Act
	client, err := factory.GetClientForApp(context.Background(), 1, "review")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.False(t, decryptionCalled, "DecryptAPIKey should NOT be called for Ollama")

	modelInfo := client.GetModelInfo()
	assert.Equal(t, "ollama", modelInfo.Provider)
}

// TestClientFactory_GetClientForApp_CachesClients tests client caching
func TestClientFactory_GetClientForApp_CachesClients(t *testing.T) {
	mockConfigService := &MockLLMConfigService{
		configs: map[string]*LLMConfig{
			"user1-review": {
				UserID:   1,
				AppName:  "review",
				Provider: "ollama",
				Model:    "deepseek-coder:6.7b",
			},
		},
	}
	mockEncryptionService := &MockEncryptionService{}

	factory := NewClientFactory(mockConfigService, mockEncryptionService)

	// Act - get client twice
	client1, err1 := factory.GetClientForApp(context.Background(), 1, "review")
	client2, err2 := factory.GetClientForApp(context.Background(), 1, "review")

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotNil(t, client1)
	require.NotNil(t, client2)

	// Should return same instance (pointer equality)
	assert.Same(t, client1, client2, "Factory should cache and reuse clients")
}

// TestClientFactory_GetClientForApp_ReturnsErrorOnDecryptionFailure tests error handling
func TestClientFactory_GetClientForApp_ReturnsErrorOnDecryptionFailure(t *testing.T) {
	mockConfigService := &MockLLMConfigService{
		configs: map[string]*LLMConfig{
			"user1-review": {
				UserID:   1,
				AppName:  "review",
				Provider: "anthropic",
				Model:    "claude-3-5-haiku-20241022",
				APIKey:   "corrupted-encrypted-key",
			},
		},
	}

	// Mock encryption service that fails decryption
	mockEncryptionService := &MockEncryptionService{
		decryptError: "decryption failed: invalid ciphertext",
	}

	factory := NewClientFactory(mockConfigService, mockEncryptionService)

	// Act
	client, err := factory.GetClientForApp(context.Background(), 1, "review")

	// Assert
	require.Error(t, err, "Should return error on decryption failure")
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to create Anthropic client")
	assert.Contains(t, err.Error(), "user 1")
	assert.Contains(t, err.Error(), "app review")
}

// Mock implementations for testing

type MockLLMConfigService struct {
	configs map[string]*LLMConfig
}

func (m *MockLLMConfigService) GetUserConfigForApp(ctx context.Context, userID int, appName string) (*LLMConfig, error) {
	key := "user" + string(rune(userID+'0')) + "-" + appName
	config, exists := m.configs[key]
	if !exists {
		return nil, nil // No preference found
	}
	return config, nil
}

type MockEncryptionService struct {
	decryptedKeys map[string]string
	decryptError  string
	onDecrypt     func(encrypted string, userID int)
}

func (m *MockEncryptionService) EncryptAPIKey(apiKey string, userID int) (string, error) {
	// For ai_factory tests, we don't need encryption, just return the key as-is
	return "encrypted-" + apiKey, nil
}

func (m *MockEncryptionService) DecryptAPIKey(encrypted string, userID int) (string, error) {
	if m.onDecrypt != nil {
		m.onDecrypt(encrypted, userID)
	}

	if m.decryptError != "" {
		return "", fmt.Errorf("encryption error: %s", m.decryptError)
	}

	decrypted, exists := m.decryptedKeys[encrypted]
	if !exists {
		return "", fmt.Errorf("encryption error: key not found")
	}

	return decrypted, nil
}
