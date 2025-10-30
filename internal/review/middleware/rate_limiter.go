// Package internal_review_middleware provides HTTP middleware for the review service.
package internal_review_middleware

import (
	"context"
	"errors"
	"sync"
	"time"
)

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	CheckLimit(ctx context.Context, identifier string) error
	GetRemainingQuota(ctx context.Context, identifier string) (int, time.Time, error)
}

// tokenBucket represents a token bucket for rate limiting
//
//nolint:govet // field alignment optimized for readability
type tokenBucket struct {
	lastReset time.Time
	tokens    float64
	maxTokens float64
	mu        sync.Mutex
}

// RedisRateLimiter implements token bucket rate limiting using in-memory storage
// (Redis integration deferred to refactor phase)
//
//nolint:govet // field alignment optimized for readability
type RedisRateLimiter struct {
	buckets      map[string]*tokenBucket
	ipBuckets    map[string]*tokenBucket
	defaultLimit int
	ipLimit      int
	windowSize   time.Duration
	mu           sync.RWMutex
}

// NewRedisRateLimiter creates a new rate limiter with specified limit and window
func NewRedisRateLimiter(limit int, window time.Duration) *RedisRateLimiter {
	if limit <= 0 {
		limit = 10 // Default to 10 requests per minute
	}
	if window <= 0 {
		window = 1 * time.Minute
	}

	return &RedisRateLimiter{
		defaultLimit: limit,
		windowSize:   window,
		ipLimit:      limit,
		buckets:      make(map[string]*tokenBucket),
		ipBuckets:    make(map[string]*tokenBucket),
	}
}

// checkBucketLimit is a helper function that implements the token bucket algorithm
func (r *RedisRateLimiter) checkBucketLimit(bucket *tokenBucket, limit int) error {
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(bucket.lastReset)
	if elapsed >= r.windowSize {
		// Window expired, reset all tokens
		bucket.tokens = float64(limit)
		bucket.lastReset = now
	} else {
		// Add tokens based on time elapsed
		refillRate := float64(limit) / r.windowSize.Seconds()
		tokensToAdd := refillRate * elapsed.Seconds()
		bucket.tokens = minFloat(bucket.tokens+tokensToAdd, float64(limit))
		bucket.lastReset = now
	}

	// Check if we have tokens available
	if bucket.tokens < 1.0 {
		return ErrRateLimited
	}

	// Consume one token
	bucket.tokens -= 1.0
	return nil
}

// getOrCreateBucket retrieves or creates a token bucket from the appropriate map
func (r *RedisRateLimiter) getOrCreateBucket(buckets map[string]*tokenBucket, identifier string, limit int) *tokenBucket {
	bucket, exists := buckets[identifier]
	if !exists {
		bucket = &tokenBucket{
			tokens:    float64(limit),
			maxTokens: float64(limit),
			lastReset: time.Now(),
		}
		buckets[identifier] = bucket
	}
	return bucket
}

// CheckLimit checks if identifier has reached rate limit using token bucket algorithm
func (r *RedisRateLimiter) CheckLimit(ctx context.Context, identifier string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if identifier == "" {
		return ErrInvalidIdentifier
	}

	r.mu.Lock()
	bucket := r.getOrCreateBucket(r.buckets, identifier, r.defaultLimit)
	r.mu.Unlock()

	return r.checkBucketLimit(bucket, r.defaultLimit)
}

// GetRemainingQuota returns remaining requests and reset time
func (r *RedisRateLimiter) GetRemainingQuota(ctx context.Context, identifier string) (int, time.Time, error) {
	if ctx.Err() != nil {
		return 0, time.Time{}, ctx.Err()
	}

	if identifier == "" {
		return 0, time.Time{}, ErrInvalidIdentifier
	}

	r.mu.RLock()
	bucket, exists := r.buckets[identifier]
	r.mu.RUnlock()

	if !exists {
		return r.defaultLimit, time.Now().Add(r.windowSize), nil
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastReset)

	var resetTime time.Time
	if elapsed >= r.windowSize {
		resetTime = now
	} else {
		resetTime = bucket.lastReset.Add(r.windowSize)
	}

	remaining := int(bucket.tokens)
	return remaining, resetTime, nil
}

// SetIPLimit sets the rate limit for IP addresses
func (r *RedisRateLimiter) SetIPLimit(limit int) {
	r.ipLimit = limit
}

// CheckIPLimit checks if IP has reached rate limit
func (r *RedisRateLimiter) CheckIPLimit(ctx context.Context, ip string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if ip == "" {
		return ErrInvalidIdentifier
	}

	r.mu.Lock()
	bucket := r.getOrCreateBucket(r.ipBuckets, ip, r.ipLimit)
	r.mu.Unlock()

	return r.checkBucketLimit(bucket, r.ipLimit)
}

// ResetQuota manually resets user's quota
func (r *RedisRateLimiter) ResetQuota(ctx context.Context, identifier string) {
	if ctx.Err() != nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if bucket, exists := r.buckets[identifier]; exists {
		bucket.mu.Lock()
		defer bucket.mu.Unlock()
		bucket.tokens = float64(r.defaultLimit)
		bucket.lastReset = time.Now()
	}
}

// GetRetryAfterSeconds returns seconds to wait before retry
func (r *RedisRateLimiter) GetRetryAfterSeconds(ctx context.Context, identifier string) (int64, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	r.mu.RLock()
	bucket, exists := r.buckets[identifier]
	r.mu.RUnlock()

	if !exists {
		return 0, nil
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastReset)

	if elapsed >= r.windowSize {
		return 0, nil
	}

	// Return remaining time in this window
	remaining := r.windowSize - elapsed
	return int64(remaining.Seconds()), nil
}

// NewRateLimitMiddleware creates Gin middleware for rate limiting
func NewRateLimitMiddleware(limiter RateLimiter) interface{} {
	// Placeholder for Gin middleware wrapper
	// Full implementation in handlers package
	return nil
}

// Helper function
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Errors
var (
	ErrRateLimited       = errors.New("rate limit exceeded")
	ErrInvalidIdentifier = errors.New("invalid identifier")
)
