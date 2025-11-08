package portal_services

import (
	"context"
	"fmt"
	"sync"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai/providers"
)

// LLMConfig represents a user's LLM configuration
type LLMConfig struct {
	UserID      int
	AppName     string
	Provider    string
	Model       string
	APIKey      string // Encrypted for API providers, empty for Ollama
	APIEndpoint string
	IsDefault   bool
}

// ClientFactory creates and caches AI provider clients based on user preferences
type ClientFactory struct {
	configService     LLMConfigServiceInterface
	encryptionService EncryptionServiceInterface
	clientCache       map[string]ai.Provider
	cacheMutex        sync.RWMutex
}

// LLMConfigServiceInterface defines methods for retrieving user LLM configurations
type LLMConfigServiceInterface interface {
	GetUserConfigForApp(ctx context.Context, userID int, appName string) (*LLMConfig, error)
}

// EncryptionServiceInterface defines methods for API key encryption/decryption
type EncryptionServiceInterface interface {
	DecryptAPIKey(encrypted string, userID int) (string, error)
}

// NewClientFactory creates a new AI client factory
func NewClientFactory(configService LLMConfigServiceInterface, encryptionService EncryptionServiceInterface) *ClientFactory {
	return &ClientFactory{
		configService:     configService,
		encryptionService: encryptionService,
		clientCache:       make(map[string]ai.Provider),
	}
}

// GetClientForApp returns an AI client for the specified user and app
// It follows this logic:
// 1. Check cache for existing client
// 2. Look up user's LLM preference for this app
// 3. If API-based provider (not Ollama), decrypt API key
// 4. Create appropriate client (Anthropic, OpenAI, DeepSeek, Mistral, or Ollama)
// 5. Cache and return the client
// 6. Fall back to Ollama if no preference or error
func (f *ClientFactory) GetClientForApp(ctx context.Context, userID int, appName string) (ai.Provider, error) {
	// Generate cache key
	cacheKey := fmt.Sprintf("user%d-%s", userID, appName)

	// Check cache first (read lock)
	f.cacheMutex.RLock()
	if cachedClient, exists := f.clientCache[cacheKey]; exists {
		f.cacheMutex.RUnlock()
		return cachedClient, nil
	}
	f.cacheMutex.RUnlock()

	// Look up user preference
	config, err := f.configService.GetUserConfigForApp(ctx, userID, appName)
	if err != nil {
		// Error retrieving config - fall back to Ollama
		return f.createDefaultOllamaClient(), nil
	}

	if config == nil {
		// No preference set - fall back to Ollama
		return f.createDefaultOllamaClient(), nil
	}

	// Create client based on provider
	var client ai.Provider

	switch config.Provider {
	case "ollama":
		// Local Ollama - no API key needed
		endpoint := config.APIEndpoint
		if endpoint == "" {
			endpoint = "http://localhost:11434"
		}
		client = providers.NewOllamaClient(endpoint, config.Model)

	case "anthropic":
		// API-based - decrypt API key
		apiKey, decryptErr := f.decryptAPIKey(config.APIKey, userID)
		if decryptErr != nil {
			return nil, fmt.Errorf("failed to decrypt API key for Anthropic: %w", decryptErr)
		}
		client = providers.NewAnthropicClient(apiKey, config.Model)

	case "openai":
		// API-based - decrypt API key
		apiKey, decryptErr := f.decryptAPIKey(config.APIKey, userID)
		if decryptErr != nil {
			return nil, fmt.Errorf("failed to decrypt API key for OpenAI: %w", decryptErr)
		}
		client = providers.NewOpenAIClient(apiKey, config.Model)

	case "deepseek":
		// API-based - decrypt API key
		apiKey, decryptErr := f.decryptAPIKey(config.APIKey, userID)
		if decryptErr != nil {
			return nil, fmt.Errorf("failed to decrypt API key for DeepSeek: %w", decryptErr)
		}
		client = providers.NewDeepSeekClient(apiKey, config.Model)

	case "mistral":
		// API-based - decrypt API key
		apiKey, decryptErr := f.decryptAPIKey(config.APIKey, userID)
		if decryptErr != nil {
			return nil, fmt.Errorf("failed to decrypt API key for Mistral: %w", decryptErr)
		}
		client = providers.NewMistralClient(apiKey, config.Model)

	default:
		// Unknown provider - fall back to Ollama
		return f.createDefaultOllamaClient(), nil
	}

	// Cache client (write lock)
	f.cacheMutex.Lock()
	f.clientCache[cacheKey] = client
	f.cacheMutex.Unlock()

	return client, nil
}

// decryptAPIKey is a helper that handles decryption and error wrapping
func (f *ClientFactory) decryptAPIKey(encrypted string, userID int) (string, error) {
	if f.encryptionService == nil {
		return "", fmt.Errorf("encryption service not configured")
	}

	decrypted, err := f.encryptionService.DecryptAPIKey(encrypted, userID)
	if err != nil {
		return "", err
	}

	return decrypted, nil
}

// createDefaultOllamaClient creates a default Ollama client as fallback
func (f *ClientFactory) createDefaultOllamaClient() ai.Provider {
	return providers.NewOllamaClient("http://localhost:11434", "deepseek-coder:6.7b")
}

// ClearCache removes all cached clients (useful for testing or when configs change)
func (f *ClientFactory) ClearCache() {
	f.cacheMutex.Lock()
	defer f.cacheMutex.Unlock()
	f.clientCache = make(map[string]ai.Provider)
}

// ClearCacheForUser removes cached clients for a specific user
func (f *ClientFactory) ClearCacheForUser(userID int) {
	f.cacheMutex.Lock()
	defer f.cacheMutex.Unlock()

	// Find and delete all entries for this user
	for key := range f.clientCache {
		userPrefix := fmt.Sprintf("user%d-", userID)
		if len(key) >= len(userPrefix) && key[:len(userPrefix)] == userPrefix {
			delete(f.clientCache, key)
		}
	}
}
