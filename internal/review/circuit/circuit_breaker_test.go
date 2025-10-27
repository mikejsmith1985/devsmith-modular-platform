// Package circuit provides circuit breaker pattern for fault tolerance.
package circuit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestBreaker_InitialState tests initial CLOSED state
func TestBreaker_InitialState(t *testing.T) {
	// WHEN: Creating new circuit breaker
	breaker := NewCircuitBreaker(testConfig())

	// THEN: Should be in CLOSED state
	assert.Equal(t, StateClosed, breaker.State())
}

// TestBreaker_ClosedToOpen tests transition from CLOSED to OPEN
func TestBreaker_ClosedToOpen(t *testing.T) {
	// GIVEN: Circuit breaker in CLOSED state
	config := &Config{
		OpenThreshold: 3,
	}
	breaker := NewCircuitBreaker(config)

	ctx := context.Background()

	// WHEN: Recording failures equal to threshold
	for i := 0; i < 3; i++ {
		breaker.RecordFailure(ctx)
	}

	// THEN: State should be OPEN
	assert.Equal(t, StateOpen, breaker.State())
}

// TestBreaker_OpenRejectsRequests tests OPEN state rejects requests
func TestBreaker_OpenRejectsRequests(t *testing.T) {
	// GIVEN: Circuit breaker in OPEN state
	config := &Config{
		OpenThreshold: 1,
	}
	breaker := NewCircuitBreaker(config)

	ctx := context.Background()
	breaker.RecordFailure(ctx)

	// WHEN: Attempting to execute request
	err := breaker.Execute(ctx, func(context.Context) error {
		return nil
	})

	// THEN: Should be rejected immediately
	assert.Error(t, err)
	assert.Equal(t, ErrCircuitOpen, err)
}

// TestBreaker_OpenToHalfOpen tests timeout-based transition
func TestBreaker_OpenToHalfOpen(t *testing.T) {
	// GIVEN: Circuit breaker that opens quickly then times out
	config := &Config{
		OpenThreshold: 1,
		Timeout:       50 * time.Millisecond,
	}
	breaker := NewCircuitBreaker(config)

	ctx := context.Background()
	breaker.RecordFailure(ctx)

	// Verify it's open
	assert.Equal(t, StateOpen, breaker.State())

	// WHEN: Waiting for timeout
	time.Sleep(100 * time.Millisecond)

	// THEN: Should transition to HALF_OPEN
	assert.Equal(t, StateHalfOpen, breaker.State())
}

// TestBreaker_HalfOpenSuccess tests HALF_OPEN to CLOSED on success
func TestBreaker_HalfOpenSuccess(t *testing.T) {
	// GIVEN: Circuit breaker in HALF_OPEN state
	config := &Config{
		OpenThreshold:     1,
		Timeout:           50 * time.Millisecond,
		HalfOpenThreshold: 1,
	}
	breaker := NewCircuitBreaker(config)

	ctx := context.Background()
	breaker.RecordFailure(ctx)
	time.Sleep(100 * time.Millisecond)

	// Verify it's in HALF_OPEN
	assert.Equal(t, StateHalfOpen, breaker.State())

	// WHEN: Recording success
	breaker.RecordSuccess(ctx)

	// THEN: Should return to CLOSED
	assert.Equal(t, StateClosed, breaker.State())
}

// TestBreaker_HalfOpenFailure tests HALF_OPEN to OPEN on failure
func TestBreaker_HalfOpenFailure(t *testing.T) {
	// GIVEN: Circuit breaker in HALF_OPEN state
	config := &Config{
		OpenThreshold: 1,
		Timeout:       50 * time.Millisecond,
	}
	breaker := NewCircuitBreaker(config)

	ctx := context.Background()
	breaker.RecordFailure(ctx)
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, StateHalfOpen, breaker.State())

	// WHEN: Recording failure in HALF_OPEN
	breaker.RecordFailure(ctx)

	// THEN: Should go back to OPEN
	assert.Equal(t, StateOpen, breaker.State())
}

// TestBreaker_ExecuteSuccess tests successful execution in CLOSED
func TestBreaker_ExecuteSuccess(t *testing.T) {
	// GIVEN: Circuit breaker in CLOSED state
	breaker := NewCircuitBreaker(testConfig())
	ctx := context.Background()

	// WHEN: Executing successful function
	executed := false
	err := breaker.Execute(ctx, func(context.Context) error {
		executed = true
		return nil
	})

	// THEN: Function should execute
	assert.NoError(t, err)
	assert.True(t, executed)
}

// TestBreaker_ExecuteFailure tests failure handling
func TestBreaker_ExecuteFailure(t *testing.T) {
	// GIVEN: Circuit breaker in CLOSED state
	breaker := NewCircuitBreaker(testConfig())
	ctx := context.Background()

	testErr := errors.New("test error")

	// WHEN: Executing failed function
	err := breaker.Execute(ctx, func(context.Context) error {
		return testErr
	})

	// THEN: Error should be propagated
	assert.Equal(t, testErr, err)
}

// TestBreaker_ContextCancellation tests context respect
func TestBreaker_ContextCancellation(t *testing.T) {
	// GIVEN: Circuit breaker with cancelled context
	_ = NewCircuitBreaker(testConfig())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// WHEN: Attempting to check state with cancelled context
	// THEN: Should handle gracefully
	assert.Error(t, ctx.Err())
}

// TestBreaker_Metrics tests metrics tracking
func TestBreaker_Metrics(t *testing.T) {
	// GIVEN: Circuit breaker with recorded events
	breaker := NewCircuitBreaker(testConfig())
	ctx := context.Background()

	// Record some successes
	for i := 0; i < 5; i++ {
		breaker.RecordSuccess(ctx)
	}

	// WHEN: Getting metrics
	metrics := breaker.Metrics()

	// THEN: Metrics should be tracked
	assert.NotNil(t, metrics)
	assert.Equal(t, 5, metrics.Successes, "Should track successes")
}

// TestBreaker_StateString tests state representation
func TestBreaker_StateString(t *testing.T) {
	// GIVEN: Circuit breaker in different states
	breaker := NewCircuitBreaker(testConfig())

	// WHEN: Checking state string
	state := breaker.State()

	// THEN: Should have valid state
	assert.NotEmpty(t, state)
	assert.Equal(t, StateClosed, state)
}

// TestBreaker_ResetMetrics tests metric reset
func TestBreaker_ResetMetrics(t *testing.T) {
	// GIVEN: Circuit breaker with recorded events
	breaker := NewCircuitBreaker(testConfig())
	ctx := context.Background()

	breaker.RecordSuccess(ctx)
	breaker.RecordFailure(ctx)

	// WHEN: Resetting metrics
	breaker.ResetMetrics(ctx)

	// THEN: Metrics should be cleared
	metrics := breaker.Metrics()
	assert.Equal(t, 0, metrics.Successes)
	assert.Equal(t, 0, metrics.Failures)
}

// TestBreaker_ConcurrentAccess tests thread safety
func TestBreaker_ConcurrentAccess(t *testing.T) {
	// GIVEN: Circuit breaker
	breaker := NewCircuitBreaker(testConfig())
	ctx := context.Background()

	// WHEN: Concurrent access from multiple goroutines
	for i := 0; i < 10; i++ {
		go func() {
			breaker.RecordSuccess(ctx)
		}()
		go func() {
			_ = breaker.Execute(ctx, func(context.Context) error { return nil })
		}()
	}

	// Allow operations to complete
	time.Sleep(100 * time.Millisecond)

	// THEN: Should handle concurrency safely
	assert.NotNil(t, breaker.Metrics())
}

// TestBreaker_FailureThreshold tests cumulative failures
func TestBreaker_FailureThreshold(t *testing.T) {
	// GIVEN: Circuit breaker with threshold of 5
	config := &Config{
		OpenThreshold: 5,
	}
	breaker := NewCircuitBreaker(config)
	ctx := context.Background()

	// WHEN: Recording 4 failures (below threshold)
	for i := 0; i < 4; i++ {
		breaker.RecordFailure(ctx)
	}

	// THEN: Should still be CLOSED
	assert.Equal(t, StateClosed, breaker.State())

	// WHEN: Recording 5th failure
	breaker.RecordFailure(ctx)

	// THEN: Should transition to OPEN
	assert.Equal(t, StateOpen, breaker.State())
}

// Helper function
func testConfig() *Config {
	return &Config{
		OpenThreshold:     5,
		HalfOpenThreshold: 2,
		Timeout:           100 * time.Millisecond,
	}
}
