package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// Entry holds cached analysis data with expiration time
type Entry struct {
	Data      *review_models.AnalysisResult
	ExpiresAt time.Time
	CreatedAt time.Time
}

// InMemoryCache provides in-memory caching for analysis results
type InMemoryCache struct {
	store       map[string]*Entry
	mu          sync.RWMutex
	statsMu     sync.RWMutex
	stats       Stats
	stopCleanup chan struct{} // Channel to signal cleanup goroutine to stop
}

// NewInMemoryCache creates a new in-memory cache for analysis results
func NewInMemoryCache() *InMemoryCache {
	cache := &InMemoryCache{
		store:       make(map[string]*Entry),
		stopCleanup: make(chan struct{}),
	}
	// Start cleanup goroutine to evict expired entries
	go cache.cleanupExpired()
	return cache
}

// cacheKey generates a consistent cache key from review ID and mode
func cacheKey(reviewID int64, mode string) string {
	return fmt.Sprintf("analysis:%d:%s", reviewID, mode)
}

// Get retrieves a cached analysis result
func (c *InMemoryCache) Get(ctx context.Context, reviewID int64, mode string) (*review_models.AnalysisResult, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	key := cacheKey(reviewID, mode)
	entry, exists := c.store[key]
	if !exists {
		c.recordMiss()
		return nil, nil
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		c.recordMiss()
		return nil, nil
	}

	c.recordHit()
	return entry.Data, nil
}

// Set stores an analysis result in the cache
func (c *InMemoryCache) Set(ctx context.Context, reviewID int64, mode string, result *review_models.AnalysisResult, ttl time.Duration) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	if result == nil {
		return fmt.Errorf("cache: cannot store nil result")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(reviewID, mode)
	entry := &Entry{
		Data:      result,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}
	c.store[key] = entry
	return nil
}

// Delete removes a cached analysis result
func (c *InMemoryCache) Delete(ctx context.Context, reviewID int64, mode string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(reviewID, mode)
	delete(c.store, key)
	return nil
}

// Clear removes all cached entries
func (c *InMemoryCache) Clear(ctx context.Context) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.store = make(map[string]*Entry)
	return nil
}

// Stats returns cache performance statistics
func (c *InMemoryCache) Stats(ctx context.Context) *Stats {
	if ctx.Err() != nil {
		return nil
	}

	c.mu.RLock()
	currentSize := len(c.store)
	c.mu.RUnlock()

	c.statsMu.RLock()
	defer c.statsMu.RUnlock()

	stats := c.stats
	stats.CurrentSize = currentSize

	// Calculate hit rate
	if stats.TotalRequests > 0 {
		stats.HitRate = float64(stats.Hits) / float64(stats.TotalRequests) * 100
	}

	return &stats
}

// recordHit increments cache hit counter
func (c *InMemoryCache) recordHit() {
	c.statsMu.Lock()
	defer c.statsMu.Unlock()

	c.stats.Hits++
	c.stats.TotalRequests++
}

// recordMiss increments cache miss counter
func (c *InMemoryCache) recordMiss() {
	c.statsMu.Lock()
	defer c.statsMu.Unlock()

	c.stats.Misses++
	c.stats.TotalRequests++
}

// recordEviction increments eviction counter
func (c *InMemoryCache) recordEviction() {
	c.statsMu.Lock()
	defer c.statsMu.Unlock()

	c.stats.Evictions++
}

// cleanupExpired periodically removes expired entries
func (c *InMemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			evicted := 0

			for key, entry := range c.store {
				if now.After(entry.ExpiresAt) {
					delete(c.store, key)
					evicted++
				}
			}
			c.mu.Unlock()

			if evicted > 0 {
				for i := 0; i < evicted; i++ {
					c.recordEviction()
				}
			}
		case <-c.stopCleanup:
			// Stop signal received - exit goroutine
			return
		}
	}
}

// Stop gracefully stops the cache cleanup goroutine
func (c *InMemoryCache) Stop() {
	close(c.stopCleanup)
}
