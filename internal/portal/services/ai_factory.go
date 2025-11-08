// Package portal_services provides business logic for the DevSmith Portal.
//
// AI Client Factory Architecture:
//
// The ClientFactory is responsible for creating and managing AI provider clients
// based on user preferences. It implements several key patterns:
//
//  1. Conditional Decryption: API keys are ONLY decrypted for cloud providers
//     (Anthropic, OpenAI, DeepSeek, Mistral). Local providers like Ollama
//     do NOT use encryption, avoiding unnecessary overhead.
//
//  2. Performance Caching: Clients are cached per user+app+provider combination
//     to avoid recreating HTTP clients and re-establishing connections. Cache
//     uses sync.RWMutex for thread-safe concurrent access.
//
//  3. Fallback Strategy: When no user preference exists, factory falls back
//     to local Ollama with deepseek-coder:6.7b model. This ensures the platform
//     always has a working LLM even for new users.
//
//  4. Error Context: All errors include user_id and app_name context to aid
//     debugging in multi-user scenarios.
//
// Thread Safety: The factory is safe for concurrent use. Cache reads use
// RLock for high performance, cache writes use Lock for safety.
//
// Example Usage:
//
//	factory := NewClientFactory(configService, encryptionService)
//	client, err := factory.GetClientForApp(ctx, userID, "review")
//	if err != nil {
//		return fmt.Errorf("failed to get AI client: %w", err)
//	}
//	response, err := client.Generate(ctx, request)
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

// GetClientForApp returns an AI client for the specified user and app.
// It resolves the user's LLM preference, creates the appropriate provider client,
// and caches it for future requests. Falls back to Ollama if no preference exists.
//
// Logic flow:
//  1. Validate inputs (userID > 0, appName non-empty)
//  2. Check cache for existing client (read lock)
//  3. Look up user's LLM preference for this app
//  4. If API-based provider (not Ollama), decrypt API key
//  5. Create appropriate client (Anthropic, OpenAI, DeepSeek, Mistral, or Ollama)
//  6. Cache and return the client (write lock)
//  7. Fall back to Ollama if no preference or error
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - userID: User identifier (must be > 0)
//   - appName: Application name (must be non-empty, e.g., "review", "logs")
//
// Returns:
//   - ai.Provider: Ready-to-use AI client
//   - error: Validation error, config lookup error, decryption error, or provider creation error
//
// Thread Safety: Safe for concurrent calls. Uses read lock for cache hits,
// write lock for cache misses.
func (f *ClientFactory) GetClientForApp(ctx context.Context, userID int, appName string) (ai.Provider, error) {
	// Validate inputs
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user_id: %d (must be positive)", userID)
	}
	if appName == "" {
		return nil, fmt.Errorf("invalid app_name: empty string (must be non-empty)")
	}

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
			return nil, fmt.Errorf("failed to create Anthropic client for user %d, app %s: %w", userID, appName, decryptErr)
		}
		client = providers.NewAnthropicClient(apiKey, config.Model)

	case "openai":
		// API-based - decrypt API key
		apiKey, decryptErr := f.decryptAPIKey(config.APIKey, userID)
		if decryptErr != nil {
			return nil, fmt.Errorf("failed to create OpenAI client for user %d, app %s: %w", userID, appName, decryptErr)
		}
		client = providers.NewOpenAIClient(apiKey, config.Model)

	case "deepseek":
		// API-based - decrypt API key
		apiKey, decryptErr := f.decryptAPIKey(config.APIKey, userID)
		if decryptErr != nil {
			return nil, fmt.Errorf("failed to create DeepSeek client for user %d, app %s: %w", userID, appName, decryptErr)
		}
		client = providers.NewDeepSeekClient(apiKey, config.Model)

	case "mistral":
		// API-based - decrypt API key
		apiKey, decryptErr := f.decryptAPIKey(config.APIKey, userID)
		if decryptErr != nil {
			return nil, fmt.Errorf("failed to create Mistral client for user %d, app %s: %w", userID, appName, decryptErr)
		}
		client = providers.NewMistralClient(apiKey, config.Model)

	default:
		// Unknown provider - fall back to Ollama with warning
		// (Could log warning here once logging is integrated)
		return f.createDefaultOllamaClient(), nil
	}

	// Cache client (write lock)
	f.cacheMutex.Lock()
	f.clientCache[cacheKey] = client
	f.cacheMutex.Unlock()

	return client, nil
}

// decryptAPIKey decrypts an encrypted API key for the specified user.
// Returns an error if encryption service is nil or decryption fails.
//
// This method SHOULD NOT be called for local providers like Ollama,
// as they don't use encrypted API keys (performance optimization).
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

// createDefaultOllamaClient creates a default Ollama client as fallback.
// Uses localhost:11434 endpoint and deepseek-coder:6.7b model.
//
// This is called when:
// - User has no LLM preference configured
// - Config service returns an error
// - Unknown provider is specified
//
// Ensures the platform always has a working LLM, even for new users.
func (f *ClientFactory) createDefaultOllamaClient() ai.Provider {
	return providers.NewOllamaClient("http://localhost:11434", "deepseek-coder:6.7b")
}

// ClearCache removes all cached clients from the factory.
//
// Use Cases:
//   - Testing: Clear state between test cases
//   - Config Changes: When LLM configurations are updated globally
//   - Memory Management: Periodic cache clearing to free resources
//
// Thread Safety: Acquires write lock during operation.
func (f *ClientFactory) ClearCache() {
	f.cacheMutex.Lock()
	defer f.cacheMutex.Unlock()
	f.clientCache = make(map[string]ai.Provider)
}

// ClearCacheForUser removes all cached clients for a specific user.
//
// Use Cases:
//   - User Updates Preferences: When user changes LLM provider or API key
//   - User Logout: Clean up user-specific state
//   - API Key Rotation: Force re-creation with new credentials
//
// Parameters:
//   - userID: User identifier (must match userID used in GetClientForApp)
//
// Implementation: Scans cache for keys matching "user{userID}-" prefix
// and removes all matches.
//
// Thread Safety: Acquires write lock during operation.
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
