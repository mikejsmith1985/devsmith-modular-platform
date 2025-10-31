// Package cache provides caching for review analysis results
package cache

import (
	"context"
	"time"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// Interface defines the contract for caching analysis results.
type Interface interface {
	// Get retrieves a cached analysis result by review ID and mode
	Get(ctx context.Context, reviewID int64, mode string) (*review_models.AnalysisResult, error)

	// Set stores an analysis result in the cache
	Set(ctx context.Context, reviewID int64, mode string, result *review_models.AnalysisResult, ttl time.Duration) error

	// Delete removes a cached analysis result
	Delete(ctx context.Context, reviewID int64, mode string) error

	// Clear removes all cached entries
	Clear(ctx context.Context) error

	// Stats returns cache performance statistics
	Stats(ctx context.Context) *Stats
}

// Stats holds cache performance metrics
type Stats struct {
	Hits          int64
	Misses        int64
	Evictions     int64
	CurrentSize   int
	TotalRequests int64
	HitRate       float64
}
