package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

func TestInMemoryCache_Set_Get_Success(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Stop() // Ensure cleanup goroutine stops
	defer cache.Stop() // Ensure cleanup goroutine stops
	ctx := context.Background()

	// GIVEN: A cache and analysis result
	result := &review_models.AnalysisResult{
		ReviewID: 1,
		Mode:     "skim",
		Summary:  "Test summary",
		Metadata: `{"key":"value"}`,
	}

	// WHEN: Setting a value
	err := cache.Set(ctx, 1, "skim", result, 1*time.Hour)

	// THEN: No error should occur
	assert.NoError(t, err)

	// AND: Value should be retrievable
	retrieved, err := cache.Get(ctx, 1, "skim")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, result.Summary, retrieved.Summary)
}

func TestInMemoryCache_Get_Miss(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Stop() // Ensure cleanup goroutine stops
	ctx := context.Background()

	// WHEN: Getting non-existent key
	retrieved, err := cache.Get(ctx, 999, "nonexistent")

	// THEN: Should return nil without error (miss)
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestInMemoryCache_Expiration(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Stop() // Ensure cleanup goroutine stops
	ctx := context.Background()

	// GIVEN: A cache entry with short TTL
	result := &review_models.AnalysisResult{
		ReviewID: 1,
		Mode:     "scan",
		Summary:  "Quick expiry test",
	}

	err := cache.Set(ctx, 1, "scan", result, 100*time.Millisecond)
	require.NoError(t, err)

	// WHEN: Immediately retrieving
	retrieved, err := cache.Get(ctx, 1, "scan")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)

	// AND: Waiting for expiration
	time.Sleep(150 * time.Millisecond)

	// THEN: Entry should have expired
	retrieved, err = cache.Get(ctx, 1, "scan")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestInMemoryCache_Delete(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Stop() // Ensure cleanup goroutine stops
	ctx := context.Background()

	// GIVEN: A cached entry
	result := &review_models.AnalysisResult{
		ReviewID: 1,
		Mode:     "detailed",
		Summary:  "Delete test",
	}
	err := cache.Set(ctx, 1, "detailed", result, 1*time.Hour)
	require.NoError(t, err)

	// WHEN: Deleting the entry
	err = cache.Delete(ctx, 1, "detailed")

	// THEN: No error should occur
	assert.NoError(t, err)

	// AND: Entry should no longer exist
	retrieved, err := cache.Get(ctx, 1, "detailed")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestInMemoryCache_Clear(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Stop() // Ensure cleanup goroutine stops
	ctx := context.Background()

	// GIVEN: Multiple cached entries
	for i := int64(1); i <= 5; i++ {
		result := &review_models.AnalysisResult{
			ReviewID: i,
			Mode:     "skim",
			Summary:  "Test",
		}
		err := cache.Set(ctx, i, "skim", result, 1*time.Hour)
		require.NoError(t, err)
	}

	// WHEN: Clearing all entries
	err := cache.Clear(ctx)

	// THEN: No error should occur
	assert.NoError(t, err)

	// AND: All entries should be gone
	for i := int64(1); i <= 5; i++ {
		retrieved, err := cache.Get(ctx, i, "skim")
		assert.NoError(t, err)
		assert.Nil(t, retrieved)
	}
}

func TestInMemoryCache_Stats_HitRate(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Stop() // Ensure cleanup goroutine stops
	ctx := context.Background()

	// GIVEN: Cache with some data
	result := &review_models.AnalysisResult{
		ReviewID: 1,
		Mode:     "skim",
		Summary:  "Stats test",
	}
	err := cache.Set(ctx, 1, "skim", result, 1*time.Hour)
	require.NoError(t, err)

	// WHEN: Generating cache hits and misses
	cache.Get(ctx, 1, "skim")   // Hit
	cache.Get(ctx, 1, "skim")   // Hit
	cache.Get(ctx, 2, "scan")   // Miss
	cache.Get(ctx, 3, "detail") // Miss

	// THEN: Stats should reflect correct hit rate
	stats := cache.Stats(ctx)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(2), stats.Hits)
	assert.Equal(t, int64(2), stats.Misses)
	assert.Equal(t, int64(4), stats.TotalRequests)
	assert.Equal(t, 50.0, stats.HitRate)
}

func TestInMemoryCache_Stats_Size(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Stop() // Ensure cleanup goroutine stops
	ctx := context.Background()

	// GIVEN: Empty cache
	stats := cache.Stats(ctx)
	assert.Equal(t, 0, stats.CurrentSize)

	// WHEN: Adding entries
	for i := int64(1); i <= 10; i++ {
		result := &review_models.AnalysisResult{
			ReviewID: i,
			Mode:     "skim",
		}
		err := cache.Set(ctx, i, "skim", result, 1*time.Hour)
		require.NoError(t, err)
	}

	// THEN: Size should be 10
	stats = cache.Stats(ctx)
	assert.Equal(t, 10, stats.CurrentSize)
}

func TestInMemoryCache_ContextCancellation(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Stop() // Ensure cleanup goroutine stops
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// WHEN: Trying to Get with cancelled context
	result := &review_models.AnalysisResult{ReviewID: 1, Mode: "skim"}

	// THEN: Should return context cancelled error
	_, err := cache.Get(ctx, 1, "skim")
	assert.Error(t, err)

	// AND: Set should also fail
	err = cache.Set(ctx, 1, "skim", result, 1*time.Hour)
	assert.Error(t, err)

	// AND: Delete should also fail
	err = cache.Delete(ctx, 1, "skim")
	assert.Error(t, err)
}

func TestInMemoryCache_NilResultRejected(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Stop() // Ensure cleanup goroutine stops
	ctx := context.Background()

	// WHEN: Trying to set nil result
	err := cache.Set(ctx, 1, "skim", nil, 1*time.Hour)

	// THEN: Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot store nil result")
}

func TestInMemoryCache_MultipleModes_SameReview(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Stop() // Ensure cleanup goroutine stops
	ctx := context.Background()

	// GIVEN: Same review ID with different modes
	modes := []string{"skim", "scan", "detailed", "critical", "preview"}

	for _, mode := range modes {
		result := &review_models.AnalysisResult{
			ReviewID: 1,
			Mode:     mode,
			Summary:  "Mode: " + mode,
		}
		err := cache.Set(ctx, 1, mode, result, 1*time.Hour)
		require.NoError(t, err)
	}

	// WHEN: Retrieving all modes
	for _, mode := range modes {
		retrieved, err := cache.Get(ctx, 1, mode)

		// THEN: Each should be distinct and correct
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, mode, retrieved.Mode)
		assert.Contains(t, retrieved.Summary, mode)
	}
}
