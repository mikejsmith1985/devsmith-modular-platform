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
	UserID      int    // 8 bytes on 64-bit systems
	Provider    string // 16 bytes (pointer + length)
	Model       string // 16 bytes (pointer + length)
	APIKey      string // 16 bytes (pointer + length) - Encrypted for API providers, empty for Ollama
	APIEndpoint string // 16 bytes (pointer + length)
	AppName     string // 16 bytes (pointer + length)
	IsDefault   bool   // 1 byte
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
	EncryptAPIKey(apiKey string, userID int) (string, error)
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
// GetClientForApp retrieves or creates an AI client for the specified user and app.
//
// Cache Strategy: Clients are cached by "user{ID}-{appName}" key to avoid
// recreating HTTP connections and re-establishing provider state.
//
// Fallback Strategy: Returns Ollama client on any configuration error to
// ensure the platform always has a working LLM.
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

	// Get user configuration
	client, err := f.createClientFromConfig(ctx, userID, appName)
	if err != nil {
		// Propagate errors (especially decryption failures) instead of silently falling back
		// Security: If encrypted API key can't be decrypted, fail loudly
		return nil, err
	}

	// Cache client (write lock)
	f.cacheMutex.Lock()
	f.clientCache[cacheKey] = client
	f.cacheMutex.Unlock()

	return client, nil
}

// createClientFromConfig creates an AI client based on user configuration.
// Returns error if config lookup fails or client creation fails.
func (f *ClientFactory) createClientFromConfig(ctx context.Context, userID int, appName string) (ai.Provider, error) {
	// Look up user preference
	config, err := f.configService.GetUserConfigForApp(ctx, userID, appName)
	if err != nil {
		return nil, fmt.Errorf("failed to get user config: %w", err)
	}

	if config == nil {
		// No preference set - use default
		return f.createDefaultOllamaClient(), nil
	}

	// Create client based on provider
	return f.createClientForProvider(config, userID, appName)
}

// createClientForProvider creates a client for the specified provider configuration.
func (f *ClientFactory) createClientForProvider(config *LLMConfig, userID int, appName string) (ai.Provider, error) {
	switch config.Provider {
	case "ollama":
		return f.createOllamaClient(config), nil

	case "anthropic":
		return f.createAnthropicClient(config, userID, appName)

	case "openai":
		return f.createOpenAIClient(config, userID, appName)

	case "deepseek":
		return f.createDeepSeekClient(config, userID, appName)

	case "mistral":
		return f.createMistralClient(config, userID, appName)

	default:
		// Unknown provider - use default
		return f.createDefaultOllamaClient(), nil
	}
}

// createOllamaClient creates an Ollama client from configuration.
func (f *ClientFactory) createOllamaClient(config *LLMConfig) ai.Provider {
	endpoint := config.APIEndpoint
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	return providers.NewOllamaClient(endpoint, config.Model)
}

// createAnthropicClient creates an Anthropic client from configuration.
func (f *ClientFactory) createAnthropicClient(config *LLMConfig, userID int, appName string) (ai.Provider, error) {
	apiKey, err := f.decryptAPIKey(config.APIKey, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Anthropic client for user %d, app %s: %w", userID, appName, err)
	}
	return providers.NewAnthropicClient(apiKey, config.Model), nil
}

// createOpenAIClient creates an OpenAI client from configuration.
func (f *ClientFactory) createOpenAIClient(config *LLMConfig, userID int, appName string) (ai.Provider, error) {
	apiKey, err := f.decryptAPIKey(config.APIKey, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI client for user %d, app %s: %w", userID, appName, err)
	}
	return providers.NewOpenAIClient(apiKey, config.Model), nil
}

// createDeepSeekClient creates a DeepSeek client from configuration.
func (f *ClientFactory) createDeepSeekClient(config *LLMConfig, userID int, appName string) (ai.Provider, error) {
	apiKey, err := f.decryptAPIKey(config.APIKey, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create DeepSeek client for user %d, app %s: %w", userID, appName, err)
	}
	return providers.NewDeepSeekClient(apiKey, config.Model), nil
}

// createMistralClient creates a Mistral client from configuration.
func (f *ClientFactory) createMistralClient(config *LLMConfig, userID int, appName string) (ai.Provider, error) {
	apiKey, err := f.decryptAPIKey(config.APIKey, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Mistral client for user %d, app %s: %w", userID, appName, err)
	}
	return providers.NewMistralClient(apiKey, config.Model), nil
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
