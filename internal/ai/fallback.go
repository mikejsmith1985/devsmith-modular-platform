package ai

import (
	"context"
	"fmt"
	"sync"
)

// FallbackChain manages a chain of AI providers with automatic failover
type FallbackChain struct {
	providers      []AIProvider
	maxRetries     int
	failures       map[string]int64 // Track failures per provider
	mu             sync.RWMutex
}

// NewFallbackChain creates a new fallback chain
func NewFallbackChain() *FallbackChain {
	return &FallbackChain{
		providers:  make([]AIProvider, 0),
		maxRetries: 1, // Default: try each provider once
		failures:   make(map[string]int64),
	}
}

// AddProvider adds a provider to the fallback chain
func (fc *FallbackChain) AddProvider(provider AIProvider) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.providers = append(fc.providers, provider)
}

// Generate tries each provider in sequence until one succeeds
func (fc *FallbackChain) Generate(ctx context.Context, req *AIRequest) (*AIResponse, error) {
	fc.mu.RLock()
	providers := fc.providers
	fc.mu.RUnlock()

	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers registered in fallback chain")
	}

	var lastErr error

	// Try each provider in order
	for _, provider := range providers {
		resp, err := provider.Generate(ctx, req)
		if err == nil {
			return resp, nil
		}

		// Record failure
		providerName := provider.GetModelInfo().Provider
		fc.RecordFailure(ctx, providerName)
		lastErr = err
	}

	// All providers failed
	return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
}

// GetSuccessfulProvider returns the first provider that succeeds
func (fc *FallbackChain) GetSuccessfulProvider(ctx context.Context) (AIProvider, error) {
	fc.mu.RLock()
	providers := fc.providers
	fc.mu.RUnlock()

	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers in chain")
	}

	// Try a simple request to each provider
	for _, provider := range providers {
		err := provider.HealthCheck(ctx)
		if err == nil {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("no successful providers found")
}

// GetHealthyProvider returns the first provider that passes health check
func (fc *FallbackChain) GetHealthyProvider(ctx context.Context) (AIProvider, error) {
	fc.mu.RLock()
	providers := fc.providers
	fc.mu.RUnlock()

	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers in chain")
	}

	// Try health check on each provider
	for _, provider := range providers {
		err := provider.HealthCheck(ctx)
		if err == nil {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("all providers are unhealthy")
}

// SetMaxRetries sets the maximum retries per provider
func (fc *FallbackChain) SetMaxRetries(retries int) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.maxRetries = retries
}

// RecordFailure records a failure for a provider
func (fc *FallbackChain) RecordFailure(ctx context.Context, providerName string) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.failures[providerName]++
}

// GetFailureCount returns the failure count for a provider
func (fc *FallbackChain) GetFailureCount(providerName string) int64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	return fc.failures[providerName]
}

// ResetFailures resets all failure counters
func (fc *FallbackChain) ResetFailures(ctx context.Context) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.failures = make(map[string]int64)
}
