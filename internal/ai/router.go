package ai

import (
	"context"
	"fmt"
	"sync"
)

// userPreference stores a user's model selection for an app
type userPreference struct {
	ProviderModel string // Format: "provider:model"
	Persist       bool   // Whether to persist to database
}

// DefaultRouter implements the Router interface for intelligent AI provider selection
type DefaultRouter struct {
	providers       map[string]Provider       // Key: "provider:model"
	userPreferences map[string]userPreference // Key: "userID:app"
	mu              sync.RWMutex              // Protect concurrent access
}

// NewDefaultRouter creates a new default router
func NewDefaultRouter() *DefaultRouter {
	return &DefaultRouter{
		providers:       make(map[string]Provider),
		userPreferences: make(map[string]userPreference),
	}
}

// RegisterProvider registers an AI provider with the router
func (r *DefaultRouter) RegisterProvider(providerName, model string, provider Provider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	if providerName == "" || model == "" {
		return fmt.Errorf("provider name and model cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", providerName, model)
	r.providers[key] = provider

	return nil
}

// Route selects the best provider for a user in a given app context
func (r *DefaultRouter) Route(ctx context.Context, appName string, userID int64) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check if user has a preference for this app
	prefKey := r.getUserAppKey(userID, appName)
	if pref, exists := r.userPreferences[prefKey]; exists {
		// User has a stored preference
		if provider, providerExists := r.providers[pref.ProviderModel]; providerExists {
			return provider, nil
		}
		// Preference refers to non-existent provider, fall through to default
	}

	// No preference set, use intelligent default
	return r.getDefaultProvider()
}

// GetAvailableModels returns all models available to a user in an app
func (r *DefaultRouter) GetAvailableModels(ctx context.Context, appName string, userID int64) ([]*ModelInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []*ModelInfo

	for _, provider := range r.providers {
		models = append(models, provider.GetModelInfo())
	}

	return models, nil
}

// SetUserPreference updates user's model selection for an app
func (r *DefaultRouter) SetUserPreference(ctx context.Context, userID int64, appName, provider, model string, persist bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verify provider exists
	key := fmt.Sprintf("%s:%s", provider, model)
	if _, exists := r.providers[key]; !exists {
		return fmt.Errorf("provider %s:%s is not registered", provider, model)
	}

	// Store preference
	prefKey := r.getUserAppKey(userID, appName)
	r.userPreferences[prefKey] = userPreference{
		ProviderModel: key,
		Persist:       persist,
	}

	// TODO: If persist is true, save to database
	// For now, this is stored in-memory

	return nil
}

// LogUsage records API usage for cost tracking
func (r *DefaultRouter) LogUsage(ctx context.Context, userID int64, appName string, req *Request, resp *Response) error {
	// TODO: Implement usage logging to database
	// This will be used by the cost monitoring service
	return nil
}

// getDefaultProvider returns the default provider (usually the cheapest)
func (r *DefaultRouter) getDefaultProvider() (Provider, error) {
	// Strategy: Prefer free providers (Ollama), then cheapest
	var bestProvider Provider
	bestCost := 1000000.0

	for _, provider := range r.providers {
		info := provider.GetModelInfo()
		cost := info.CostPer1kInputTokens

		// Prefer free/cheaper providers
		if cost < bestCost {
			bestCost = cost
			bestProvider = provider
		}
	}

	if bestProvider == nil {
		return nil, fmt.Errorf("no providers registered")
	}

	return bestProvider, nil
}

// getUserAppKey generates a unique key for user+app combination
func (r *DefaultRouter) getUserAppKey(userID int64, appName string) string {
	return fmt.Sprintf("%d:%s", userID, appName)
}
