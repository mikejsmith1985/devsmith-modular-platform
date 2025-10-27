// Package circuit provides circuit breaker pattern for fault tolerance.
package circuit

import (
	"context"
	"errors"
	"sync"
	"time"
)

// circuitBreaker implements the circuit breaker pattern
type circuitBreaker struct {
	config            *CircuitBreakerConfig
	metrics           *Metrics
	lastFailureTime   time.Time
	lastStateChangeAt time.Time
	mu                sync.RWMutex
	state             CircuitBreakerState
	failureCount      int
	successCount      int
}

// NewCircuitBreaker creates a new circuit breaker with given config
func NewCircuitBreaker(config *CircuitBreakerConfig) CircuitBreaker {
	if config == nil {
		config = &CircuitBreakerConfig{
			OpenThreshold:     5,
			HalfOpenThreshold: 2,
			Timeout:           30 * time.Second,
			MetricsWindow:     1 * time.Minute,
		}
	}

	// Apply defaults for zero values
	if config.OpenThreshold == 0 {
		config.OpenThreshold = 5
	}
	if config.HalfOpenThreshold == 0 {
		config.HalfOpenThreshold = 2
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MetricsWindow == 0 {
		config.MetricsWindow = 1 * time.Minute
	}

	return &circuitBreaker{
		state:             StateClosed,
		config:            config,
		lastStateChangeAt: time.Now(),
		metrics:           &Metrics{},
	}
}

// checkTimeout checks if we should transition from OPEN to HALF_OPEN
func (cb *circuitBreaker) checkTimeout() {
	if cb.state == StateOpen && time.Since(cb.lastStateChangeAt) >= cb.config.Timeout {
		cb.state = StateHalfOpen
		cb.successCount = 0
		cb.failureCount = 0
		cb.lastStateChangeAt = time.Now()
	}
}

// Execute executes a function with circuit breaker protection
func (cb *circuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	cb.mu.Lock()
	cb.checkTimeout()
	currentState := cb.state

	// Reject if still OPEN
	if currentState == StateOpen {
		cb.mu.Unlock()
		return ErrCircuitOpen
	}

	cb.mu.Unlock()

	// Execute the function
	err := fn(ctx)

	// Record result
	if err != nil {
		cb.RecordFailure(ctx)
	} else {
		cb.RecordSuccess(ctx)
	}

	return err
}

// State returns the current state
func (cb *circuitBreaker) State() CircuitBreakerState {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.checkTimeout()
	return cb.state
}

// RecordSuccess records a successful request
func (cb *circuitBreaker) RecordSuccess(ctx context.Context) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.metrics.Successes++

	switch cb.state {
	case StateClosed:
		// Stay in CLOSED
	case StateHalfOpen:
		// Transition to CLOSED after success
		cb.successCount++
		if cb.successCount >= cb.config.HalfOpenThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
			cb.lastStateChangeAt = time.Now()
		}
	}
}

// RecordFailure records a failed request
func (cb *circuitBreaker) RecordFailure(ctx context.Context) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.metrics.Failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		// Transition to OPEN if threshold reached
		cb.failureCount++
		if cb.failureCount >= cb.config.OpenThreshold {
			cb.state = StateOpen
			cb.lastStateChangeAt = time.Now()
		}
	case StateHalfOpen:
		// Return to OPEN on any failure
		cb.state = StateOpen
		cb.lastStateChangeAt = time.Now()
		cb.failureCount = 0
		cb.successCount = 0
	}
}

// Metrics returns current metrics
func (cb *circuitBreaker) Metrics() *Metrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// Return a copy
	return &Metrics{
		Failures:  cb.metrics.Failures,
		Successes: cb.metrics.Successes,
	}
}

// ResetMetrics resets all metrics
func (cb *circuitBreaker) ResetMetrics(ctx context.Context) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.metrics = &Metrics{}
	cb.failureCount = 0
	cb.successCount = 0
}

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState string

// Circuit breaker states
const (
	StateClosed   CircuitBreakerState = "CLOSED"
	StateOpen     CircuitBreakerState = "OPEN"
	StateHalfOpen CircuitBreakerState = "HALF_OPEN"
)

// CircuitBreakerConfig configuration for circuit breaker
type CircuitBreakerConfig struct {
	OpenThreshold     int
	HalfOpenThreshold int
	Timeout           time.Duration
	MetricsWindow     time.Duration
}

// Metrics tracks circuit breaker metrics
type Metrics struct {
	Failures  int
	Successes int
}

// CircuitBreaker defines the circuit breaker interface
type CircuitBreaker interface {
	Execute(ctx context.Context, fn func(context.Context) error) error
	State() CircuitBreakerState
	RecordSuccess(ctx context.Context)
	RecordFailure(ctx context.Context)
	Metrics() *Metrics
	ResetMetrics(ctx context.Context)
}

// ErrCircuitOpen is returned when circuit is open
var ErrCircuitOpen = errors.New("circuit breaker is open")
