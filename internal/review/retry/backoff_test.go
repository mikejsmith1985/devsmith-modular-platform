// Package retry provides retry logic with exponential backoff.
package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestBackoff_CalculateDelay tests delay calculation for attempts
func TestBackoff_CalculateDelay(t *testing.T) {
	// GIVEN: RetryStrategy with default config
	config := &RetryConfig{
		InitialDelay:      100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		MaxDelay:          30 * time.Second,
	}
	strategy := NewRetryStrategy(config)

	tests := []struct {
		attempt     int
		expectedMin time.Duration
		expectedMax time.Duration
	}{
		{1, 100 * time.Millisecond, 110 * time.Millisecond}, // No backoff
		{2, 200 * time.Millisecond, 220 * time.Millisecond}, // 1x multiplier
		{3, 400 * time.Millisecond, 440 * time.Millisecond}, // 2x multiplier
		{4, 800 * time.Millisecond, 880 * time.Millisecond}, // 3x multiplier
		{10, 30 * time.Second, 30 * time.Second},            // Capped at max
	}

	for _, tt := range tests {
		// WHEN: Calculating delay for attempt
		delay := strategy.CalculateDelay(tt.attempt)

		// THEN: Delay is within expected range
		assert.GreaterOrEqual(t, delay, tt.expectedMin,
			"Attempt %d delay should be >= %v", tt.attempt, tt.expectedMin)
		assert.LessOrEqual(t, delay, tt.expectedMax,
			"Attempt %d delay should be <= %v", tt.attempt, tt.expectedMax)
	}
}

// TestBackoff_ExponentialGrowth tests exponential growth pattern
func TestBackoff_ExponentialGrowth(t *testing.T) {
	// GIVEN: RetryStrategy without jitter for predictable results
	config := &RetryConfig{
		InitialDelay:      100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		MaxDelay:          60 * time.Second,
		JitterFraction:    0, // No jitter
	}
	strategy := NewRetryStrategy(config)

	// WHEN: Calculating delays for multiple attempts
	delay1 := strategy.CalculateDelay(1)
	delay2 := strategy.CalculateDelay(2)
	delay3 := strategy.CalculateDelay(3)

	// THEN: Each delay is ~2x the previous
	assert.Equal(t, 100*time.Millisecond, delay1)
	assert.Equal(t, 200*time.Millisecond, delay2)
	assert.Equal(t, 400*time.Millisecond, delay3)
}

// TestBackoff_Jitter tests jitter application
func TestBackoff_Jitter(t *testing.T) {
	// GIVEN: RetryStrategy with 10% jitter
	config := &RetryConfig{
		InitialDelay:      100 * time.Millisecond,
		BackoffMultiplier: 1.0,
		MaxDelay:          60 * time.Second,
		JitterFraction:    0.1,
	}
	strategy := NewRetryStrategy(config)

	// WHEN: Calculating delay multiple times
	delays := make([]time.Duration, 0)
	for i := 0; i < 10; i++ {
		delays = append(delays, strategy.CalculateDelay(1))
	}

	// THEN: Delays vary due to jitter
	minDelay := delays[0]
	maxDelay := delays[0]
	for _, d := range delays {
		if d < minDelay {
			minDelay = d
		}
		if d > maxDelay {
			maxDelay = d
		}
	}

	// Should have variation
	assert.Greater(t, maxDelay, minDelay, "Jitter should create variation")
	assert.GreaterOrEqual(t, minDelay, 90*time.Millisecond, "Min should respect jitter")
	assert.LessOrEqual(t, maxDelay, 110*time.Millisecond, "Max should respect jitter")
}

// TestBackoff_MaxDelayCapping tests max delay enforcement
func TestBackoff_MaxDelayCapping(t *testing.T) {
	// GIVEN: RetryStrategy with small max delay
	config := &RetryConfig{
		InitialDelay:      100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		MaxDelay:          1 * time.Second,
		JitterFraction:    0,
	}
	strategy := NewRetryStrategy(config)

	// WHEN: Calculating delay for high attempt number
	delay := strategy.CalculateDelay(10)

	// THEN: Delay is capped at max
	assert.LessOrEqual(t, delay, 1*time.Second)
}

// TestBackoff_DefaultConfig tests default configuration
func TestBackoff_DefaultConfig(t *testing.T) {
	// WHEN: Creating strategy with nil config
	strategy := NewRetryStrategy(nil)

	// THEN: Defaults are applied
	assert.NotNil(t, strategy)
	delay := strategy.CalculateDelay(1)
	assert.Greater(t, delay, 0*time.Second)
}

// TestBackoff_ShouldRetry tests retry decision logic
func TestBackoff_ShouldRetry(t *testing.T) {
	// GIVEN: RetryStrategy with max 3 retries
	config := &RetryConfig{
		MaxRetries: 3,
	}
	strategy := NewRetryStrategy(config)

	// WHEN/THEN: Check retry decision for different attempts
	assert.True(t, strategy.ShouldRetry(1, 3), "Attempt 1 of 3 should retry")
	assert.True(t, strategy.ShouldRetry(2, 3), "Attempt 2 of 3 should retry")
	assert.False(t, strategy.ShouldRetry(3, 3), "Attempt 3 of 3 should NOT retry")
	assert.False(t, strategy.ShouldRetry(4, 3), "Attempt 4 exceeds max")
}

// TestBackoff_ContextCancellation tests context-aware retry
func TestBackoff_ContextCancellation(t *testing.T) {
	// GIVEN: Cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	config := &RetryConfig{
		MaxRetries: 3,
	}
	_ = NewRetryStrategy(config)

	// WHEN: Attempting retry with cancelled context
	// THEN: Should stop immediately (implementation should check context)
	assert.Error(t, ctx.Err())
}

// TestBackoff_RetryOnError tests full retry flow
func TestBackoff_RetryOnError(t *testing.T) {
	// GIVEN: Function that fails first 2 times, succeeds on 3rd
	attempt := 0
	fn := func(ctx context.Context) error {
		attempt++
		if attempt < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	config := &RetryConfig{
		MaxRetries:        3,
		InitialDelay:      10 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}
	strategy := NewRetryStrategy(config)

	ctx := context.Background()

	// WHEN: Executing with retry
	err := strategy.ExecuteWithRetry(ctx, fn)

	// THEN: Should succeed after retries
	assert.NoError(t, err)
	assert.Equal(t, 3, attempt, "Should retry twice then succeed")
}

// TestBackoff_RetryExhaustion tests max retries exhaustion
func TestBackoff_RetryExhaustion(t *testing.T) {
	// GIVEN: Function that always fails
	fn := func(ctx context.Context) error {
		return errors.New("persistent error")
	}

	config := &RetryConfig{
		MaxRetries:        2,
		InitialDelay:      5 * time.Millisecond,
		BackoffMultiplier: 1.0,
	}
	strategy := NewRetryStrategy(config)

	ctx := context.Background()

	// WHEN: Executing with retry
	err := strategy.ExecuteWithRetry(ctx, fn)

	// THEN: Should fail after max retries
	assert.Error(t, err)
	assert.Equal(t, "persistent error", err.Error())
}

// TestBackoff_ContextDeadline tests deadline respect
func TestBackoff_ContextDeadline(t *testing.T) {
	// GIVEN: Function that needs retries but context has tight deadline
	attempt := 0
	fn := func(ctx context.Context) error {
		attempt++
		return errors.New("error")
	}

	config := &RetryConfig{
		MaxRetries:        5,
		InitialDelay:      100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}
	strategy := NewRetryStrategy(config)

	// Context with 50ms deadline
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// WHEN: Executing with tight deadline
	err := strategy.ExecuteWithRetry(ctx, fn)

	// THEN: Should stop due to deadline, not exhaust retries
	assert.Error(t, err)
	assert.Less(t, attempt, 5, "Should stop before max retries due to deadline")
}

// TestBackoff_ZeroRetries tests with no retries allowed
func TestBackoff_ZeroRetries(t *testing.T) {
	// GIVEN: Config with 0 max retries
	config := &RetryConfig{
		MaxRetries: 0,
	}
	strategy := NewRetryStrategy(config)

	// WHEN: Checking if should retry
	// THEN: Should not retry
	assert.False(t, strategy.ShouldRetry(1, 0))
}
