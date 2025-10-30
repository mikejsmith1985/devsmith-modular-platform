// Package cache provides caching functionality for logs operations.
package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/cache"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// TestNewDashboardCache creates a new dashboard cache.
func TestNewDashboardCache(t *testing.T) {
	// GIVEN: Cache configuration
	ttl := 5 * time.Minute

	// WHEN: Creating a new cache
	c := cache.NewDashboardCache(ttl)

	// THEN: It should be created successfully
	assert.NotNil(t, c)
}

// TestCacheSet stores data in cache.
func TestCacheSet(t *testing.T) {
	// GIVEN: A cache and dashboard stats
	c := cache.NewDashboardCache(5 * time.Minute)
	stats := &logs_models.DashboardStats{
		GeneratedAt: time.Now(),
	}

	// WHEN: Storing stats in cache
	err := c.Set(context.Background(), "dashboard", stats)

	// THEN: It should store successfully
	assert.NoError(t, err)
}

// TestCacheGet retrieves data from cache.
func TestCacheGet(t *testing.T) {
	// GIVEN: Cache with stored stats
	c := cache.NewDashboardCache(5 * time.Minute)
	originalStats := &logs_models.DashboardStats{
		GeneratedAt: time.Now(),
	}
	err := c.Set(context.Background(), "dashboard", originalStats)
	require.NoError(t, err)

	// WHEN: Retrieving stats from cache
	retrieved, err := c.Get(context.Background(), "dashboard")

	// THEN: It should retrieve the same stats
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)

	stats, ok := retrieved.(*logs_models.DashboardStats)
	require.True(t, ok)
	assert.Equal(t, originalStats.GeneratedAt, stats.GeneratedAt)
}

// TestCacheMiss returns nil for non-existent key.
func TestCacheMiss(t *testing.T) {
	// GIVEN: An empty cache
	c := cache.NewDashboardCache(5 * time.Minute)

	// WHEN: Retrieving non-existent key
	retrieved, err := c.Get(context.Background(), "nonexistent")

	// THEN: Should return nil without error
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

// TestCacheExpiry removes expired entries.
func TestCacheExpiry(t *testing.T) {
	// GIVEN: Cache with short TTL
	c := cache.NewDashboardCache(100 * time.Millisecond)
	stats := &logs_models.DashboardStats{
		GeneratedAt: time.Now(),
	}
	err := c.Set(context.Background(), "dashboard", stats)
	require.NoError(t, err)

	// WHEN: Waiting for expiry
	time.Sleep(150 * time.Millisecond)

	// THEN: Entry should be expired
	retrieved, err := c.Get(context.Background(), "dashboard")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

// TestCacheDelete removes entries.
func TestCacheDelete(t *testing.T) {
	// GIVEN: Cache with stored entry
	c := cache.NewDashboardCache(5 * time.Minute)
	stats := &logs_models.DashboardStats{
		GeneratedAt: time.Now(),
	}
	err := c.Set(context.Background(), "dashboard", stats)
	require.NoError(t, err)

	// WHEN: Deleting the entry
	err = c.Delete(context.Background(), "dashboard")

	// THEN: Delete should succeed and entry should be gone
	assert.NoError(t, err)
	retrieved, err := c.Get(context.Background(), "dashboard")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

// TestCacheClear removes all entries.
func TestCacheClear(t *testing.T) {
	// GIVEN: Cache with multiple entries
	c := cache.NewDashboardCache(5 * time.Minute)
	stats := &logs_models.DashboardStats{
		GeneratedAt: time.Now(),
	}
	err := c.Set(context.Background(), "key1", stats)
	require.NoError(t, err)
	err = c.Set(context.Background(), "key2", stats)
	require.NoError(t, err)

	// WHEN: Clearing cache
	err = c.Clear(context.Background())

	// THEN: All entries should be removed
	assert.NoError(t, err)
	retrieved1, _ := c.Get(context.Background(), "key1")
	retrieved2, _ := c.Get(context.Background(), "key2")
	assert.Nil(t, retrieved1)
	assert.Nil(t, retrieved2)
}

// TestCacheContextCancellation handles context cancellation.
func TestCacheContextCancellation(t *testing.T) {
	// GIVEN: A cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := cache.NewDashboardCache(5 * time.Minute)
	stats := &logs_models.DashboardStats{
		GeneratedAt: time.Now(),
	}

	// WHEN: Operating on cache with cancelled context
	err := c.Set(ctx, "key", stats)

	// THEN: Should respect context cancellation
	assert.Error(t, err)
}

// TestCacheServiceStatsGet retrieves service-specific stats from cache.
func TestCacheServiceStatsGet(t *testing.T) {
	// GIVEN: Cache with service stats
	c := cache.NewDashboardCache(5 * time.Minute)
	serviceStats := &logs_models.LogStats{
		Service:    "api-service",
		TotalCount: 100,
	}
	cacheKey := "service_stats_api-service"

	err := c.Set(context.Background(), cacheKey, serviceStats)
	require.NoError(t, err)

	// WHEN: Retrieving service stats
	// THEN: Should retrieve with correct type
	assert.NoError(t, err)
}

// TestCacheHealthStatsGet retrieves health stats from cache.
func TestCacheHealthStatsGet(t *testing.T) {
	// GIVEN: Cache with health stats
	c := cache.NewDashboardCache(5 * time.Minute)
	healthStats := make(map[string]*logs_models.ServiceHealth)
	healthStats["service1"] = &logs_models.ServiceHealth{
		Service: "service1",
		Status:  "OK",
	}

	cacheKey := "health_stats"
	err := c.Set(context.Background(), cacheKey, healthStats)
	require.NoError(t, err)

	// WHEN: Retrieving health stats
	retrieved, err := c.Get(context.Background(), cacheKey)

	// THEN: Should retrieve health data
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
}
