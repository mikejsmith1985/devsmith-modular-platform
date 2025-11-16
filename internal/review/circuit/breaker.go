// Package circuit provides circuit breaker patterns for resilient external dependency calls.
package circuit

import (
	"context"
	"time"

	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
	"github.com/sony/gobreaker"
)

// OllamaCircuitBreaker wraps an Ollama client with circuit breaker protection.
// Prevents cascading failures when Ollama service is unhealthy.
type OllamaCircuitBreaker struct {
	breaker *gobreaker.CircuitBreaker
	client  review_services.OllamaClientInterface
	logger  *logger.Logger
}

// NewOllamaCircuitBreaker creates a circuit breaker wrapper for Ollama client.
// Configuration:
// - MaxRequests: 3 (max concurrent half-open requests)
// - Interval: 60s (window for counting failures)
// - Timeout: 60s (half-open→open timeout)
// - ReadyToTrip: 5 consecutive failures triggers open state
func NewOllamaCircuitBreaker(client review_services.OllamaClientInterface, logger *logger.Logger) *OllamaCircuitBreaker {
	settings := gobreaker.Settings{
		Name:        "ollama",
		MaxRequests: 3,                // Allow 3 requests in half-open state
		Interval:    60 * time.Second, // Reset failure count every minute
		Timeout:     60 * time.Second, // Stay open for 60s before attempting half-open
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Open circuit after 5 consecutive failures
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warn("Circuit breaker state change", "name", name, "from", from.String(), "to", to.String())
		},
	}

	breaker := gobreaker.NewCircuitBreaker(settings)

	return &OllamaCircuitBreaker{
		breaker: breaker,
		client:  client,
		logger:  logger,
	}
}

// Generate wraps the Ollama Generate call with circuit breaker protection.
// Returns error if circuit is open (fail-fast instead of waiting for timeout).
func (cb *OllamaCircuitBreaker) Generate(ctx context.Context, prompt string) (string, error) {
	// Execute through circuit breaker
	result, err := cb.breaker.Execute(func() (interface{}, error) {
		cb.logger.Debug("Circuit breaker: calling Ollama", "state", cb.breaker.State().String())
		return cb.client.Generate(ctx, prompt)
	})

	if err != nil {
		// Log circuit breaker specific errors
		if err == gobreaker.ErrOpenState {
			cb.logger.Error("Circuit breaker is open - Ollama calls blocked", "state", cb.breaker.State().String())
		} else if err == gobreaker.ErrTooManyRequests {
			cb.logger.Warn("Circuit breaker throttling - too many concurrent requests", "state", cb.breaker.State().String())
		}
		return "", err
	}

	// Type assert result back to string
	return result.(string), nil
}

// State returns the current state of the circuit breaker.
// States: Closed (normal), Open (failing), HalfOpen (testing recovery).
func (cb *OllamaCircuitBreaker) State() gobreaker.State {
	return cb.breaker.State()
}

// Counts returns current failure/success counts.
func (cb *OllamaCircuitBreaker) Counts() gobreaker.Counts {
	return cb.breaker.Counts()
}
