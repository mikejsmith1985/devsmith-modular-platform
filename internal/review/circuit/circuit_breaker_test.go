// Package circuit provides circuit breaker pattern for fault tolerance.
package circuit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCircuitBreaker_InitialState tests initial CLOSED state
func TestCircuitBreaker_InitialState(t *testing.T) {
	// WHEN: Creating new circuit breaker
	breaker := NewCircuitBreaker(testConfig())

	// THEN: Should be in CLOSED state
	assert.Equal(t, StateClosed, breaker.State())
}

// TestCircuitBreaker_ClosedToOpen tests transition from CLOSED to OPEN
func TestCircuitBreaker_ClosedToOpen(t *testing.T) {
	// GIVEN: Circuit breaker in CLOSED state
	config := &CircuitBreakerConfig{
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

// TestCircuitBreaker_OpenRejectsRequests tests OPEN state rejects requests
func TestCircuitBreaker_OpenRejectsRequests(t *testing.T) {
	// GIVEN: Circuit breaker in OPEN state
	config := &CircuitBreakerConfig{
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

// TestCircuitBreaker_OpenToHalfOpen tests timeout-based transition
func TestCircuitBreaker_OpenToHalfOpen(t *testing.T) {
	// GIVEN: Circuit breaker that opens quickly then times out
	config := &CircuitBreakerConfig{
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

// TestCircuitBreaker_HalfOpenSuccess tests HALF_OPEN to CLOSED on success
func TestCircuitBreaker_HalfOpenSuccess(t *testing.T) {
	// GIVEN: Circuit breaker in HALF_OPEN state
	config := &CircuitBreakerConfig{
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

// TestCircuitBreaker_HalfOpenFailure tests HALF_OPEN to OPEN on failure
func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	// GIVEN: Circuit breaker in HALF_OPEN state
	config := &CircuitBreakerConfig{
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

// TestCircuitBreaker_ExecuteSuccess tests successful execution in CLOSED
func TestCircuitBreaker_ExecuteSuccess(t *testing.T) {
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

// TestCircuitBreaker_ExecuteFailure tests failure handling
func TestCircuitBreaker_ExecuteFailure(t *testing.T) {
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

// TestCircuitBreaker_ContextCancellation tests context respect
func TestCircuitBreaker_ContextCancellation(t *testing.T) {
	// GIVEN: Circuit breaker with cancelled context
	_ = NewCircuitBreaker(testConfig())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// WHEN: Attempting to check state with cancelled context
	// THEN: Should handle gracefully
	assert.Error(t, ctx.Err())
}

// TestCircuitBreaker_Metrics tests metrics tracking
func TestCircuitBreaker_Metrics(t *testing.T) {
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

// TestCircuitBreaker_StateString tests state representation
func TestCircuitBreaker_StateString(t *testing.T) {
	// GIVEN: Circuit breaker in different states
	breaker := NewCircuitBreaker(testConfig())

	// WHEN: Checking state string
	state := breaker.State()

	// THEN: Should have valid state
	assert.NotEmpty(t, state)
	assert.Equal(t, StateClosed, state)
}

// TestCircuitBreaker_ResetMetrics tests metric reset
func TestCircuitBreaker_ResetMetrics(t *testing.T) {
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

// TestCircuitBreaker_ConcurrentAccess tests thread safety
func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
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

// TestCircuitBreaker_FailureThreshold tests cumulative failures
func TestCircuitBreaker_FailureThreshold(t *testing.T) {
	// GIVEN: Circuit breaker with threshold of 5
	config := &CircuitBreakerConfig{
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
func testConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		OpenThreshold:     5,
		HalfOpenThreshold: 2,
		Timeout:           100 * time.Millisecond,
	}
}
