package circuit

import (
	"context"
	"errors"
	"testing"
	"time"
)

// BenchmarkCircuitBreaker_Execute_Success measures overhead of circuit breaker when successful
func BenchmarkCircuitBreaker_Execute_Success(b *testing.B) {
	config := &Config{
		OpenThreshold:     5,
		HalfOpenThreshold: 2,
		Timeout:           60 * time.Second,
		MetricsWindow:     1 * time.Minute,
	}
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	successFunc := func(ctx context.Context) error {
		return nil // Always succeed
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Execute(ctx, successFunc)
	}
}

// BenchmarkCircuitBreaker_Execute_Failure measures circuit breaker with failures
func BenchmarkCircuitBreaker_Execute_Failure(b *testing.B) {
	ctx := context.Background()
	errTest := errors.New("test error")

	failFunc := func(ctx context.Context) error {
		return errTest
	}

	config := &Config{
		OpenThreshold:     5,
		HalfOpenThreshold: 2,
		Timeout:           60 * time.Second,
		MetricsWindow:     1 * time.Minute,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create new circuit breaker for each iteration to avoid OPEN state
		cb := NewCircuitBreaker(config)
		_ = cb.Execute(ctx, failFunc)
	}
}

// BenchmarkCircuitBreaker_StateCheck measures cost of state checking
func BenchmarkCircuitBreaker_StateCheck(b *testing.B) {
	config := &Config{
		OpenThreshold:     5,
		HalfOpenThreshold: 2,
		Timeout:           60 * time.Second,
		MetricsWindow:     1 * time.Minute,
	}
	cb := NewCircuitBreaker(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.State()
	}
}

// BenchmarkCircuitBreaker_Execute_Open measures fail-fast when circuit open
func BenchmarkCircuitBreaker_Execute_Open(b *testing.B) {
	config := &Config{
		OpenThreshold:     5,
		HalfOpenThreshold: 2,
		Timeout:           60 * time.Second,
		MetricsWindow:     1 * time.Minute,
	}
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	// Open the circuit
	errTest := errors.New("test error")
	for i := 0; i < 6; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error { return errTest })
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error { return nil })
	}
}

// BenchmarkCircuitBreaker_Concurrent measures performance under concurrent load
func BenchmarkCircuitBreaker_Concurrent(b *testing.B) {
	config := &Config{
		OpenThreshold:     5,
		HalfOpenThreshold: 2,
		Timeout:           60 * time.Second,
		MetricsWindow:     1 * time.Minute,
	}
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	successFunc := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cb.Execute(ctx, successFunc)
		}
	})
}
