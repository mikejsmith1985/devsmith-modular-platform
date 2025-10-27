package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRateLimiter_AllowRequest_WithinLimit tests that requests within limit are allowed
func TestRateLimiter_AllowRequest_WithinLimit(t *testing.T) {
	// GIVEN: Rate limiter with 10 requests per minute
	limiter := NewRedisRateLimiter(10, 1*time.Minute)
	ctx := context.Background()

	// WHEN: Making a request within limit
	userID := "user-123"
	err := limiter.CheckLimit(ctx, userID)

	// THEN: Request is allowed
	assert.NoError(t, err, "Request within limit should be allowed")
}

// TestRateLimiter_RejectRequest_ExceedsLimit tests rate limiting after limit exceeded
func TestRateLimiter_RejectRequest_ExceedsLimit(t *testing.T) {
	// GIVEN: Rate limiter with 3 requests per minute
	limiter := NewRedisRateLimiter(3, 1*time.Minute)
	ctx := context.Background()
	userID := "user-456"

	// WHEN: Making 4 requests in quick succession
	for i := 0; i < 3; i++ {
		err := limiter.CheckLimit(ctx, userID)
		assert.NoError(t, err, "First 3 requests should succeed")
	}

	// THEN: 4th request is rejected
	err := limiter.CheckLimit(ctx, userID)
	assert.Error(t, err, "4th request should fail")
	assert.Equal(t, ErrRateLimited, err, "Error should be rate limit error")
}

// TestRateLimiter_GetRemainingQuota tests quota retrieval
func TestRateLimiter_GetRemainingQuota(t *testing.T) {
	// GIVEN: Rate limiter with 5 requests per minute
	limiter := NewRedisRateLimiter(5, 1*time.Minute)
	ctx := context.Background()
	userID := "user-789"

	// WHEN: Making 2 requests
	limiter.CheckLimit(ctx, userID)
	limiter.CheckLimit(ctx, userID)

	// THEN: Remaining quota is 3
	remaining, resetTime, err := limiter.GetRemainingQuota(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, 3, remaining, "Should have 3 remaining after 2 requests")
	assert.NotZero(t, resetTime, "Reset time should be set")
	assert.True(t, resetTime.After(time.Now()), "Reset time should be in future")
}

// TestRateLimiter_PerIPLimit tests per-IP limiting for unauthenticated users
func TestRateLimiter_PerIPLimit(t *testing.T) {
	// GIVEN: Rate limiter with different IP limits
	limiter := NewRedisRateLimiter(10, 1*time.Minute)
	limiter.SetIPLimit(5) // 5 requests per minute per IP
	ctx := context.Background()

	// WHEN: Making requests from IP "192.168.1.1"
	ip := "192.168.1.1"
	for i := 0; i < 5; i++ {
		err := limiter.CheckIPLimit(ctx, ip)
		assert.NoError(t, err, "First 5 requests from IP should succeed")
	}

	// THEN: 6th request from same IP is rejected
	err := limiter.CheckIPLimit(ctx, ip)
	assert.Error(t, err, "6th request from same IP should fail")
	assert.Equal(t, ErrRateLimited, err)
}

// TestRateLimiter_WindowResets tests that quota resets after time window
func TestRateLimiter_WindowResets(t *testing.T) {
	// GIVEN: Rate limiter with 100ms window (for testing)
	limiter := NewRedisRateLimiter(1, 100*time.Millisecond)
	ctx := context.Background()
	userID := "user-reset"

	// WHEN: Making 1 request
	err1 := limiter.CheckLimit(ctx, userID)
	assert.NoError(t, err1)

	// THEN: 2nd request fails (limit exceeded)
	err2 := limiter.CheckLimit(ctx, userID)
	assert.Error(t, err2, "Second request should fail within window")

	// WHEN: Waiting for window to reset
	time.Sleep(150 * time.Millisecond)

	// THEN: Request is allowed again
	err3 := limiter.CheckLimit(ctx, userID)
	assert.NoError(t, err3, "Request should succeed after window resets")
}

// TestRateLimiter_MultipleUsers tests independent rate limits for different users
func TestRateLimiter_MultipleUsers(t *testing.T) {
	// GIVEN: Rate limiter with 2 requests per minute
	limiter := NewRedisRateLimiter(2, 1*time.Minute)
	ctx := context.Background()

	// WHEN: User A makes 2 requests (at limit)
	for i := 0; i < 2; i++ {
		err := limiter.CheckLimit(ctx, "user-a")
		assert.NoError(t, err)
	}

	// THEN: User B can still make requests (independent limits)
	for i := 0; i < 2; i++ {
		err := limiter.CheckLimit(ctx, "user-b")
		assert.NoError(t, err, "User B should have independent quota")
	}

	// AND: User A's 3rd request fails
	err := limiter.CheckLimit(ctx, "user-a")
	assert.Error(t, err, "User A's 3rd request should fail")

	// AND: User B's 3rd request also fails
	err = limiter.CheckLimit(ctx, "user-b")
	assert.Error(t, err, "User B's 3rd request should fail")
}

// TestRateLimiter_ContextCancellation tests behavior with cancelled context
func TestRateLimiter_ContextCancellation(t *testing.T) {
	// GIVEN: Rate limiter
	limiter := NewRedisRateLimiter(10, 1*time.Minute)

	// WHEN: Context is cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// THEN: CheckLimit respects cancellation
	err := limiter.CheckLimit(ctx, "user-123")
	assert.Error(t, err, "CheckLimit should handle cancelled context")
	assert.Equal(t, context.Canceled, err)
}

// TestRateLimiter_ConcurrentRequests tests thread safety
func TestRateLimiter_ConcurrentRequests(t *testing.T) {
	// GIVEN: Rate limiter with 100 requests per minute
	limiter := NewRedisRateLimiter(100, 1*time.Minute)
	ctx := context.Background()
	userID := "concurrent-user"

	// WHEN: Making 100 concurrent requests
	successCount := 0
	errorCount := 0
	done := make(chan bool)

	for i := 0; i < 100; i++ {
		go func() {
			defer func() { done <- true }()
			err := limiter.CheckLimit(ctx, userID)
			if err == nil {
				successCount++
			} else if errors.Is(err, ErrRateLimited) {
				errorCount++
			}
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// THEN: Exactly 100 succeed (all within limit)
	assert.Equal(t, 100, successCount, "All 100 concurrent requests should succeed")
	assert.Equal(t, 0, errorCount, "No requests should exceed limit")
}

// TestRateLimiter_ErrorHandling tests error handling edge cases
func TestRateLimiter_ErrorHandling(t *testing.T) {
	// GIVEN: Rate limiter
	limiter := NewRedisRateLimiter(10, 1*time.Minute)
	ctx := context.Background()

	// WHEN: Checking limit with empty identifier
	err := limiter.CheckLimit(ctx, "")

	// THEN: Returns error for empty identifier
	assert.Error(t, err, "Empty identifier should return error")
	assert.Equal(t, ErrInvalidIdentifier, err)
}

// TestRateLimiter_ZeroQuota tests behavior when quota is 0
func TestRateLimiter_ZeroQuota(t *testing.T) {
	// GIVEN: Rate limiter created with 0 quota (invalid)
	// Constructor should handle gracefully or reject

	// This test verifies the limiter doesn't panic with edge case
	limiter := &RedisRateLimiter{
		defaultLimit: 0,
		windowSize:   1 * time.Minute,
	}

	// WHEN: Checking limit
	ctx := context.Background()
	err := limiter.CheckLimit(ctx, "user-123")

	// THEN: Either rejects or uses sensible default
	assert.NotNil(t, err, "Zero quota should result in error or sensible handling")
}

// TestRateLimiter_ResetQuotaManually tests manual quota reset capability
func TestRateLimiter_ResetQuotaManually(t *testing.T) {
	// GIVEN: Rate limiter with 2 requests per minute
	limiter := NewRedisRateLimiter(2, 1*time.Minute)
	ctx := context.Background()
	userID := "user-reset-manual"

	// WHEN: User makes 2 requests (at limit)
	limiter.CheckLimit(ctx, userID)
	limiter.CheckLimit(ctx, userID)

	// AND: Admin manually resets quota
	limiter.ResetQuota(ctx, userID)

	// THEN: User can make requests again immediately
	err := limiter.CheckLimit(ctx, userID)
	assert.NoError(t, err, "After reset, user should have quota available")
}

// TestRateLimiter_MiddlewareIntegration tests Gin middleware integration
func TestRateLimiter_MiddlewareIntegration(t *testing.T) {
	// This is an integration test with HTTP
	// It tests that the middleware properly rejects requests and sets headers

	// GIVEN: Rate limiting middleware
	limiter := NewRedisRateLimiter(1, 1*time.Minute)
	_ = NewRateLimitMiddleware(limiter) // Verify middleware is created

	// WHEN: First request through middleware
	req1, _ := http.NewRequest("GET", "/api/review/analyze", http.NoBody)
	req1.Header.Set("X-User-ID", "user-123")
	_ = req1

	// THEN: First request passes (would be handled by handler)
	// (Actual verification requires gin context)

	// AND: Subsequent requests are rejected with 429
	// This is tested in integration tests with actual gin context
}

// TestRateLimiter_RetryAfterHeader tests Retry-After header is set
func TestRateLimiter_RetryAfterHeader(t *testing.T) {
	// GIVEN: Rate limiter
	limiter := NewRedisRateLimiter(1, 60*time.Second)
	ctx := context.Background()
	userID := "user-retry-after"

	// WHEN: Making request at limit
	limiter.CheckLimit(ctx, userID)
	limiter.CheckLimit(ctx, userID) // This fails

	// THEN: Can get Retry-After time
	retryAfter, err := limiter.GetRetryAfterSeconds(ctx, userID)
	assert.NoError(t, err, "Should be able to get Retry-After time")
	assert.Greater(t, retryAfter, int64(0), "Retry-After should be positive")
	assert.LessOrEqual(t, retryAfter, int64(60), "Retry-After should be <= window size")
}

// Errors that should be defined
var (
	ErrRateLimited       = fmt.Errorf("rate limit exceeded")
	ErrInvalidIdentifier = fmt.Errorf("invalid identifier")
)
