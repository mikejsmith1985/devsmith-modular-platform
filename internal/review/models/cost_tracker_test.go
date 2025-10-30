// Package review_models provides domain models for the review service.
package review_models

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCostTracker_RecordUsage tests recording API usage
func TestCostTracker_RecordUsage(t *testing.T) {
	// GIVEN: Cost tracker
	tracker := NewCostTracker()
	ctx := context.Background()

	// WHEN: Recording usage
	usage := &APIUsage{
		UserID:       123,
		RequestID:    "req-1",
		APIProvider:  "claude",
		InputTokens:  100,
		OutputTokens: 50,
		TotalCost:    0.0015,
	}
	err := tracker.RecordUsage(ctx, usage)

	// THEN: Usage recorded successfully
	assert.NoError(t, err)
}

// TestCostTracker_GetUserCost tests retrieving user's total cost
func TestCostTracker_GetUserCost(t *testing.T) {
	// GIVEN: Tracker with recorded usage
	tracker := NewCostTracker()
	ctx := context.Background()

	usage1 := &APIUsage{UserID: 123, RequestID: "req-1", TotalCost: 0.01}
	usage2 := &APIUsage{UserID: 123, RequestID: "req-2", TotalCost: 0.02}

	tracker.RecordUsage(ctx, usage1)
	tracker.RecordUsage(ctx, usage2)

	// WHEN: Getting user's total cost
	cost, err := tracker.GetUserCost(ctx, 123)

	// THEN: Total cost is sum of all usage
	assert.NoError(t, err)
	assert.Equal(t, 0.03, cost)
}

// TestCostTracker_CheckQuota tests quota checking
func TestCostTracker_CheckQuota(t *testing.T) {
	// GIVEN: Tracker with user quota
	tracker := NewCostTracker()
	ctx := context.Background()

	// Set quota
	tracker.SetUserQuota(ctx, 123, 1.0)

	// WHEN: User under quota
	allowed, err := tracker.CheckQuota(ctx, 123, 0.5)

	// THEN: Should be allowed
	assert.NoError(t, err)
	assert.True(t, allowed)
}

// TestCostTracker_ExceededQuota tests over-quota rejection
func TestCostTracker_ExceededQuota(t *testing.T) {
	// GIVEN: Tracker with limited quota
	tracker := NewCostTracker()
	ctx := context.Background()

	tracker.SetUserQuota(ctx, 123, 1.0)

	// Record usage to approach quota
	usage := &APIUsage{UserID: 123, RequestID: "req-1", TotalCost: 0.9}
	tracker.RecordUsage(ctx, usage)

	// WHEN: Attempting to use more than remaining quota
	allowed, err := tracker.CheckQuota(ctx, 123, 0.2)

	// THEN: Should be rejected
	assert.NoError(t, err)
	assert.False(t, allowed)
}

// TestCostTracker_GetRemainingQuota tests remaining quota calculation
func TestCostTracker_GetRemainingQuota(t *testing.T) {
	// GIVEN: Tracker with quota and usage
	tracker := NewCostTracker()
	ctx := context.Background()

	tracker.SetUserQuota(ctx, 123, 1.0)

	usage := &APIUsage{UserID: 123, RequestID: "req-1", TotalCost: 0.3}
	tracker.RecordUsage(ctx, usage)

	// WHEN: Getting remaining quota
	remaining, err := tracker.GetRemainingQuota(ctx, 123)

	// THEN: Should be total - used
	assert.NoError(t, err)
	assert.Equal(t, 0.7, remaining)
}

// TestCostTracker_NoQuotaSet tests behavior without quota
func TestCostTracker_NoQuotaSet(t *testing.T) {
	// GIVEN: Tracker without quota set
	tracker := NewCostTracker()
	ctx := context.Background()

	// WHEN: Checking quota
	allowed, err := tracker.CheckQuota(ctx, 999, 100.0)

	// THEN: Should be allowed (no quota limit)
	assert.NoError(t, err)
	assert.True(t, allowed)
}

// TestCostTracker_ResetQuota tests quota reset
func TestCostTracker_ResetQuota(t *testing.T) {
	// GIVEN: Tracker with usage
	tracker := NewCostTracker()
	ctx := context.Background()

	tracker.SetUserQuota(ctx, 123, 1.0)
	usage := &APIUsage{UserID: 123, RequestID: "req-1", TotalCost: 0.5}
	tracker.RecordUsage(ctx, usage)

	// WHEN: Resetting quota
	err := tracker.ResetQuota(ctx, 123)

	// THEN: Quota reset
	assert.NoError(t, err)

	// And usage should be cleared
	cost, _ := tracker.GetUserCost(ctx, 123)
	assert.Equal(t, 0.0, cost)
}

// TestCostTracker_GetUsageHistory tests retrieving usage records
func TestCostTracker_GetUsageHistory(t *testing.T) {
	// GIVEN: Tracker with multiple usage records
	tracker := NewCostTracker()
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		usage := &APIUsage{
			UserID:    123,
			RequestID: "req-" + string(rune(48+i)),
			TotalCost: 0.01 * float64(i),
		}
		tracker.RecordUsage(ctx, usage)
	}

	// WHEN: Getting usage history
	history, err := tracker.GetUsageHistory(ctx, 123)

	// THEN: Should retrieve all records
	assert.NoError(t, err)
	assert.Equal(t, 3, len(history))
}

// TestCostTracker_CalculateCost tests cost calculation
func TestCostTracker_CalculateCost(t *testing.T) {
	// GIVEN: Usage with token counts
	tracker := NewCostTracker()

	// WHEN: Calculate cost for tokens
	// Claude pricing: $0.003/1K input, $0.015/1K output
	cost := tracker.CalculateCost("claude", 1000, 1000)

	// THEN: Cost should be calculated correctly
	expected := 0.003 + 0.015
	assert.InDelta(t, expected, cost, 0.0001)
}

// TestCostTracker_ContextCancellation tests context handling
func TestCostTracker_ContextCancellation(t *testing.T) {
	// GIVEN: Cancelled context
	tracker := NewCostTracker()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	usage := &APIUsage{UserID: 123, RequestID: "req-1", TotalCost: 0.01}

	// WHEN: Recording usage with cancelled context
	err := tracker.RecordUsage(ctx, usage)

	// THEN: Should respect cancellation
	assert.Error(t, err)
}

// TestCostTracker_InvalidUsage tests validation
func TestCostTracker_InvalidUsage(t *testing.T) {
	// GIVEN: Tracker
	tracker := NewCostTracker()
	ctx := context.Background()

	// WHEN: Recording invalid usage
	err := tracker.RecordUsage(ctx, nil)

	// THEN: Should return error
	assert.Error(t, err)
}

// TestCostTracker_MultipleUsers tests isolation
func TestCostTracker_MultipleUsers(t *testing.T) {
	// GIVEN: Tracker with multiple users
	tracker := NewCostTracker()
	ctx := context.Background()

	tracker.SetUserQuota(ctx, 123, 1.0)
	tracker.SetUserQuota(ctx, 456, 2.0)

	usage1 := &APIUsage{UserID: 123, RequestID: "req-1", TotalCost: 0.5}
	usage2 := &APIUsage{UserID: 456, RequestID: "req-2", TotalCost: 1.0}

	tracker.RecordUsage(ctx, usage1)
	tracker.RecordUsage(ctx, usage2)

	// WHEN: Getting costs
	cost1, _ := tracker.GetUserCost(ctx, 123)
	cost2, _ := tracker.GetUserCost(ctx, 456)

	// THEN: Should be separate
	assert.Equal(t, 0.5, cost1)
	assert.Equal(t, 1.0, cost2)
}

// TestCostTracker_ConcurrentRecording tests thread safety
func TestCostTracker_ConcurrentRecording(t *testing.T) {
	// GIVEN: Tracker
	tracker := NewCostTracker()
	ctx := context.Background()

	// WHEN: Concurrent recording
	for i := 0; i < 10; i++ {
		go func(id int) {
			usage := &APIUsage{
				UserID:    123,
				RequestID: "req-" + string(rune(48+id)),
				TotalCost: 0.01,
			}
			tracker.RecordUsage(ctx, usage)
		}(i)
	}

	// Allow operations to complete
	time.Sleep(100 * time.Millisecond)

	// THEN: All recorded safely
	cost, _ := tracker.GetUserCost(ctx, 123)
	assert.Greater(t, cost, 0.0)
}

// TestCostTracker_PricePerProvider tests provider-specific pricing
func TestCostTracker_PricePerProvider(t *testing.T) {
	// GIVEN: Tracker
	tracker := NewCostTracker()

	// WHEN: Calculating costs for different providers
	costClaude := tracker.CalculateCost("claude", 1000, 1000)
	costOpenAI := tracker.CalculateCost("openai", 1000, 1000)

	// THEN: Should have different costs based on provider
	assert.NotEqual(t, costClaude, costOpenAI)
}
