// Package ai provides AI provider abstraction, routing, and cost monitoring.
package ai

import (
	"context"
	"fmt"
	"sync"
)

// FallbackChain implements a sequential provider failover mechanism.
type FallbackChain struct {
	failureCounts map[string]int64
	providers     []Provider
	mu            sync.RWMutex
}

// NewFallbackChain creates a new fallback chain with the given providers.
func NewFallbackChain(providers ...Provider) *FallbackChain {
	fc := &FallbackChain{
		providers:     make([]Provider, 0),
		failureCounts: make(map[string]int64),
	}
	if len(providers) > 0 {
		fc.providers = providers
	}
	return fc
}

// AddProvider adds a provider to the fallback chain.
func (fc *FallbackChain) AddProvider(provider Provider) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.providers = append(fc.providers, provider)
}

// Generate tries each provider in sequence until one succeeds.
func (fc *FallbackChain) Generate(ctx context.Context, req *Request) (*Response, error) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	if len(fc.providers) == 0 {
		return nil, fmt.Errorf("no providers available in fallback chain")
	}

	var lastErr error
	for i, provider := range fc.providers {
		resp, err := provider.Generate(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		// Log provider failure for debugging
		fmt.Printf("Provider %d failed: %v, trying next...\n", i, err)
	}

	return nil, fmt.Errorf("all providers failed: %w", lastErr)
}

// GetSuccessfulProvider returns the first provider that succeeds
func (fc *FallbackChain) GetSuccessfulProvider(ctx context.Context) (Provider, error) {
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
func (fc *FallbackChain) GetHealthyProvider(ctx context.Context) (Provider, error) {
	fc.mu.RLock()
	providers := fc.providers
	fc.mu.RUnlock()

	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers in chain")
	}

	for _, provider := range providers {
		err := provider.HealthCheck(ctx)
		if err == nil {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("all providers unhealthy")
}

// SetMaxRetries sets the maximum retries per provider
func (fc *FallbackChain) SetMaxRetries(retries int) {
	// This is a no-op stub. Retry logic moved to individual providers.
}

// RecordFailure records a failure for a provider
func (fc *FallbackChain) RecordFailure(ctx context.Context, providerName string) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.failureCounts[providerName]++
}

// GetFailureCount returns the failure count for a provider
func (fc *FallbackChain) GetFailureCount(providerName string) int64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.failureCounts[providerName]
}

// ResetFailures resets failure counts for a provider
func (fc *FallbackChain) ResetFailures(providerName string) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.failureCounts[providerName] = 0
}
