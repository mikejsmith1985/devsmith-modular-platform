// Package models provides domain models for the review service.
package models

import (
	"context"
	"errors"
	"sync"
	"time"
)

// costTracker implements cost tracking and quota management
type costTracker struct {
	userCosts    map[int64]float64
	userQuotas   map[int64]float64
	usageHistory map[int64][]*APIUsage
	mu           sync.RWMutex
}

// NewCostTracker creates a new cost tracker instance
func NewCostTracker() CostTracker {
	return &costTracker{
		userCosts:    make(map[int64]float64),
		userQuotas:   make(map[int64]float64),
		usageHistory: make(map[int64][]*APIUsage),
	}
}

// RecordUsage records an API usage event
func (ct *costTracker) RecordUsage(ctx context.Context, usage *APIUsage) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if usage == nil {
		return errors.New("usage cannot be nil")
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	// Add to history
	ct.usageHistory[usage.UserID] = append(ct.usageHistory[usage.UserID], usage)

	// Update total cost
	ct.userCosts[usage.UserID] += usage.TotalCost

	return nil
}

// GetUserCost returns the total cost for a user
func (ct *costTracker) GetUserCost(ctx context.Context, userID int64) (float64, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	ct.mu.RLock()
	defer ct.mu.RUnlock()

	return ct.userCosts[userID], nil
}

// GetRemainingQuota returns the remaining quota for a user
func (ct *costTracker) GetRemainingQuota(ctx context.Context, userID int64) (float64, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	ct.mu.RLock()
	defer ct.mu.RUnlock()

	quota, hasQuota := ct.userQuotas[userID]
	if !hasQuota {
		return 0, nil
	}

	used := ct.userCosts[userID]
	remaining := quota - used

	if remaining < 0 {
		return 0, nil
	}

	return remaining, nil
}

// CheckQuota checks if a user can afford a cost
func (ct *costTracker) CheckQuota(ctx context.Context, userID int64, cost float64) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	ct.mu.RLock()
	defer ct.mu.RUnlock()

	quota, hasQuota := ct.userQuotas[userID]
	if !hasQuota {
		// No quota limit
		return true, nil
	}

	used := ct.userCosts[userID]
	remaining := quota - used

	return remaining >= cost, nil
}

// SetUserQuota sets the quota for a user
func (ct *costTracker) SetUserQuota(ctx context.Context, userID int64, quota float64) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.userQuotas[userID] = quota
}

// ResetQuota resets a user's usage and quota
func (ct *costTracker) ResetQuota(ctx context.Context, userID int64) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.userCosts[userID] = 0
	ct.usageHistory[userID] = make([]*APIUsage, 0)

	return nil
}

// GetUsageHistory returns all usage records for a user
func (ct *costTracker) GetUsageHistory(ctx context.Context, userID int64) ([]*APIUsage, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	ct.mu.RLock()
	defer ct.mu.RUnlock()

	if history, exists := ct.usageHistory[userID]; exists {
		// Return a copy
		result := make([]*APIUsage, len(history))
		copy(result, history)
		return result, nil
	}

	return make([]*APIUsage, 0), nil
}

// CalculateCost calculates cost based on tokens and provider pricing
func (ct *costTracker) CalculateCost(provider string, inputTokens, outputTokens int) float64 {
	var inputRate, outputRate float64

	// Provider-specific pricing per 1K tokens
	switch provider {
	case "claude":
		inputRate = 0.003  // $0.003 per 1K input tokens
		outputRate = 0.015 // $0.015 per 1K output tokens
	case "openai":
		inputRate = 0.0005  // $0.0005 per 1K input tokens (GPT-3.5)
		outputRate = 0.0015 // $0.0015 per 1K output tokens
	case "ollama":
		inputRate = 0.0 // Free local model
		outputRate = 0.0
	default:
		inputRate = 0.001
		outputRate = 0.001
	}

	inputCost := (float64(inputTokens) / 1000.0) * inputRate
	outputCost := (float64(outputTokens) / 1000.0) * outputRate

	return inputCost + outputCost
}

// APIUsage represents an API usage event
//
//nolint:govet // test struct, field alignment not critical
type APIUsage struct {
	CreatedAt    time.Time
	CompletedAt  time.Time
	UserID       int64
	TotalCost    float64
	InputTokens  int
	OutputTokens int
	RequestID    string
	APIProvider  string
	Status       string
	ErrorMessage string
}

// CostTracker defines the cost tracking interface
type CostTracker interface {
	RecordUsage(ctx context.Context, usage *APIUsage) error
	GetUserCost(ctx context.Context, userID int64) (float64, error)
	GetRemainingQuota(ctx context.Context, userID int64) (float64, error)
	CheckQuota(ctx context.Context, userID int64, cost float64) (bool, error)
	SetUserQuota(ctx context.Context, userID int64, quota float64)
	ResetQuota(ctx context.Context, userID int64) error
	GetUsageHistory(ctx context.Context, userID int64) ([]*APIUsage, error)
	CalculateCost(provider string, inputTokens, outputTokens int) float64
}
