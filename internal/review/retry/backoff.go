// Package retry provides retry logic with exponential backoff.
package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// retryStrategy implements exponential backoff retry logic
type retryStrategy struct {
	config *Config
}

// NewRetryStrategy creates a new retry strategy with given config
func NewRetryStrategy(config *Config) Strategy {
	if config == nil {
		config = &Config{
			MaxRetries:        3,
			InitialDelay:      100 * time.Millisecond,
			BackoffMultiplier: 2.0,
			MaxDelay:          30 * time.Second,
			JitterFraction:    0.1,
		}
	}

	// Apply defaults for zero values
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.InitialDelay == 0 {
		config.InitialDelay = 100 * time.Millisecond
	}
	if config.BackoffMultiplier == 0 {
		config.BackoffMultiplier = 2.0
	}
	if config.MaxDelay == 0 {
		config.MaxDelay = 30 * time.Second
	}

	return &retryStrategy{config: config}
}

// CalculateDelay calculates the delay for a given attempt number using exponential backoff.
// Formula: delay = initialDelay × (multiplier ^ (attempt - 1)), capped at maxDelay, with optional jitter.
func (rs *retryStrategy) CalculateDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return rs.config.InitialDelay
	}

	// Calculate exponential delay: initialDelay * (multiplier ^ (attempt - 1))
	exponent := float64(attempt - 1)
	delay := float64(rs.config.InitialDelay) * math.Pow(rs.config.BackoffMultiplier, exponent)

	// Cap at max delay
	if delay > float64(rs.config.MaxDelay) {
		delay = float64(rs.config.MaxDelay)
	}

	// Apply jitter
	if rs.config.JitterFraction > 0 {
		jitterAmount := delay * rs.config.JitterFraction
		jitter := (rand.Float64() * 2 * jitterAmount) - jitterAmount
		delay += jitter
	}

	return time.Duration(delay)
}

// ShouldRetry determines if we should retry based on attempt count
func (rs *retryStrategy) ShouldRetry(attempt, maxRetries int) bool {
	return attempt < maxRetries
}

// ExecuteWithRetry executes a function with retry logic
func (rs *retryStrategy) ExecuteWithRetry(ctx context.Context, fn func(context.Context) error) error {
	var lastErr error

	for attempt := 1; attempt <= rs.config.MaxRetries; attempt++ {
		// Check context before executing
		if ctx.Err() != nil {
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}

		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry if this was the last attempt
		if attempt >= rs.config.MaxRetries {
			break
		}

		// Calculate delay and sleep
		delay := rs.CalculateDelay(attempt)

		// Create a context-aware sleep
		select {
		case <-time.After(delay):
			// Delay completed, continue to next attempt
		case <-ctx.Done():
			// Context cancelled during sleep
			return fmt.Errorf("retry cancelled during backoff: %w", ctx.Err())
		}
	}

	return lastErr
}

// Config configuration for retry strategy
type Config struct {
	MaxRetries        int
	InitialDelay      time.Duration
	BackoffMultiplier float64
	MaxDelay          time.Duration
	JitterFraction    float64
}

// Strategy defines the retry strategy interface
type Strategy interface {
	CalculateDelay(attempt int) time.Duration
	ShouldRetry(attempt, maxRetries int) bool
	ExecuteWithRetry(ctx context.Context, fn func(context.Context) error) error
}
