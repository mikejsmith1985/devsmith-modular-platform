// Package cache provides caching functionality for logs operations.
package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// CacheEntry holds cached data with expiration time.
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// DashboardCache provides in-memory caching for dashboard stats.
type DashboardCache struct { //nolint:govet // struct alignment optimized for readability
	ttl   time.Duration
	mu    sync.RWMutex
	store map[string]*CacheEntry
}

// NewDashboardCache creates a new dashboard cache.
func NewDashboardCache(ttl time.Duration) *DashboardCache {
	cache := &DashboardCache{
		ttl:   ttl,
		store: make(map[string]*CacheEntry),
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Set stores data in the cache.
func (dc *DashboardCache) Set(ctx context.Context, key string, data interface{}) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	if key == "" {
		return fmt.Errorf("cache key cannot be empty")
	}

	dc.mu.Lock()
	defer dc.mu.Unlock()

	entry := &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(dc.ttl),
	}

	dc.store[key] = entry
	return nil
}

// Get retrieves data from the cache.
func (dc *DashboardCache) Get(ctx context.Context, key string) (interface{}, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	dc.mu.RLock()
	defer dc.mu.RUnlock()

	entry, exists := dc.store[key]
	if !exists {
		return nil, nil
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, nil
	}

	return entry.Data, nil
}

// Delete removes an entry from the cache.
func (dc *DashboardCache) Delete(ctx context.Context, key string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	dc.mu.Lock()
	defer dc.mu.Unlock()

	delete(dc.store, key)
	return nil
}

// Clear removes all entries from the cache.
func (dc *DashboardCache) Clear(ctx context.Context) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.store = make(map[string]*CacheEntry)
	return nil
}

// cleanupExpired periodically removes expired entries.
func (dc *DashboardCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		dc.mu.Lock()
		now := time.Now()

		for key, entry := range dc.store {
			if now.After(entry.ExpiresAt) {
				delete(dc.store, key)
			}
		}

		dc.mu.Unlock()
	}
}

// GetDashboardStats retrieves cached dashboard stats.
func (dc *DashboardCache) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
	data, err := dc.Get(ctx, "dashboard_stats")
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	stats, ok := data.(*models.DashboardStats)
	if !ok {
		return nil, fmt.Errorf("cached data is not DashboardStats")
	}

	return stats, nil
}

// SetDashboardStats stores dashboard stats in the cache.
func (dc *DashboardCache) SetDashboardStats(ctx context.Context, stats *models.DashboardStats) error {
	return dc.Set(ctx, "dashboard_stats", stats)
}

// GetServiceStats retrieves cached service stats.
func (dc *DashboardCache) GetServiceStats(ctx context.Context, service string) (*models.LogStats, error) {
	key := fmt.Sprintf("service_stats_%s", service)
	data, err := dc.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	stats, ok := data.(*models.LogStats)
	if !ok {
		return nil, fmt.Errorf("cached data is not LogStats")
	}

	return stats, nil
}

// SetServiceStats stores service stats in the cache.
func (dc *DashboardCache) SetServiceStats(ctx context.Context, service string, stats *models.LogStats) error {
	key := fmt.Sprintf("service_stats_%s", service)
	return dc.Set(ctx, key, stats)
}

// GetHealthStats retrieves cached health stats.
func (dc *DashboardCache) GetHealthStats(ctx context.Context) (map[string]*models.ServiceHealth, error) {
	data, err := dc.Get(ctx, "health_stats")
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	health, ok := data.(map[string]*models.ServiceHealth)
	if !ok {
		return nil, fmt.Errorf("cached data is not health stats")
	}

	return health, nil
}

// SetHealthStats stores health stats in the cache.
func (dc *DashboardCache) SetHealthStats(ctx context.Context, health map[string]*models.ServiceHealth) error {
	return dc.Set(ctx, "health_stats", health)
}
