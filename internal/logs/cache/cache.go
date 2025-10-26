// Package cache provides caching functionality for logs operations.
package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// Entry holds cached data with expiration time.
type Entry struct {
	Data      interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Stats holds cache performance metrics.
type Stats struct {
	Hits          int64
	Misses        int64
	Evictions     int64
	CurrentSize   int
	TotalRequests int64
}

// HitRate returns the cache hit rate as a percentage.
func (cs *Stats) HitRate() float64 {
	if cs.TotalRequests == 0 {
		return 0
	}
	return float64(cs.Hits) / float64(cs.TotalRequests) * 100
}

// DashboardCache provides in-memory caching for dashboard stats.
type DashboardCache struct { //nolint:govet // struct alignment optimized for readability
	mu      sync.RWMutex
	ttl     time.Duration
	store   map[string]*Entry
	stats   Stats
	statsmu sync.RWMutex
}

// NewDashboardCache creates a new dashboard cache.
func NewDashboardCache(ttl time.Duration) *DashboardCache {
	cache := &DashboardCache{
		ttl:   ttl,
		store: make(map[string]*Entry),
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

	entry := &Entry{
		Data:      data,
		ExpiresAt: time.Now().Add(dc.ttl),
		CreatedAt: time.Now(),
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
		dc.recordMiss()
		return nil, nil
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		dc.recordMiss()
		return nil, nil
	}

	dc.recordHit()
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

	dc.store = make(map[string]*Entry)
	return nil
}

// GetStats returns current cache statistics.
func (dc *DashboardCache) GetStats() Stats {
	dc.statsmu.RLock()
	defer dc.statsmu.RUnlock()

	dc.mu.RLock()
	size := len(dc.store)
	dc.mu.RUnlock()

	stats := dc.stats
	stats.CurrentSize = size
	return stats
}

// recordHit increments cache hit counter.
func (dc *DashboardCache) recordHit() {
	dc.statsmu.Lock()
	defer dc.statsmu.Unlock()

	dc.stats.Hits++
	dc.stats.TotalRequests++
}

// recordMiss increments cache miss counter.
func (dc *DashboardCache) recordMiss() {
	dc.statsmu.Lock()
	defer dc.statsmu.Unlock()

	dc.stats.Misses++
	dc.stats.TotalRequests++
}

// recordEviction increments eviction counter.
func (dc *DashboardCache) recordEviction() {
	dc.statsmu.Lock()
	defer dc.statsmu.Unlock()

	dc.stats.Evictions++
}

// cleanupExpired periodically removes expired entries.
func (dc *DashboardCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		dc.mu.Lock()
		now := time.Now()
		evicted := 0

		for key, entry := range dc.store {
			if now.After(entry.ExpiresAt) {
				delete(dc.store, key)
				evicted++
			}
		}

		dc.mu.Unlock()

		if evicted > 0 {
			for i := 0; i < evicted; i++ {
				dc.recordEviction()
			}
		}
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
