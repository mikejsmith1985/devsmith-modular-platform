package ai

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCostMonitor_NewCostMonitor_CreatesValidMonitor verifies constructor
func TestCostMonitor_NewCostMonitor_CreatesValidMonitor(t *testing.T) {
	monitor := NewCostMonitor()

	assert.NotNil(t, monitor, "Monitor should be created")
	assert.NotNil(t, monitor.userCosts)
	assert.NotNil(t, monitor.appCosts)
	assert.NotNil(t, monitor.budgets)
}

// TestCostMonitor_RecordUsage_TracksUserCost verifies usage recording
func TestCostMonitor_RecordUsage_TracksUserCost(t *testing.T) {
	monitor := NewCostMonitor()

	req := &AIRequest{Prompt: "test"}
	resp := &AIResponse{
		InputTokens:  1000,
		OutputTokens: 500,
		CostUSD:      0.05,
		ResponseTime: 2 * time.Second,
	}

	err := monitor.RecordUsage(context.Background(), 111, "review", req, resp)
	assert.NoError(t, err)

	// Verify cost was recorded
	userCost := monitor.GetUserTotalCost(111)
	assert.Equal(t, 0.05, userCost)
}

// TestCostMonitor_RecordUsage_AccumulateCosts verifies accumulation
func TestCostMonitor_RecordUsage_AccumulateCosts(t *testing.T) {
	monitor := NewCostMonitor()

	// Record multiple usages
	for i := 0; i < 5; i++ {
		resp := &AIResponse{CostUSD: 0.10}
		monitor.RecordUsage(context.Background(), 222, "review", &AIRequest{}, resp)
	}

	userCost := monitor.GetUserTotalCost(222)
	assert.Equal(t, 0.50, userCost)
}

// TestCostMonitor_RecordUsage_AppIsolation verifies per-app tracking
func TestCostMonitor_RecordUsage_AppIsolation(t *testing.T) {
	monitor := NewCostMonitor()

	monitor.RecordUsage(context.Background(), 333, "review", &AIRequest{}, &AIResponse{CostUSD: 0.10})
	monitor.RecordUsage(context.Background(), 333, "logs", &AIRequest{}, &AIResponse{CostUSD: 0.05})

	reviewCost := monitor.GetAppCostForUser(333, "review")
	logsCost := monitor.GetAppCostForUser(333, "logs")

	assert.Equal(t, 0.10, reviewCost)
	assert.Equal(t, 0.05, logsCost)
}

// TestCostMonitor_GetUserTotalCost_CorrectSum verifies total calculation
func TestCostMonitor_GetUserTotalCost_CorrectSum(t *testing.T) {
	monitor := NewCostMonitor()

	monitor.RecordUsage(context.Background(), 444, "review", &AIRequest{}, &AIResponse{CostUSD: 0.15})
	monitor.RecordUsage(context.Background(), 444, "review", &AIRequest{}, &AIResponse{CostUSD: 0.10})
	monitor.RecordUsage(context.Background(), 444, "logs", &AIRequest{}, &AIResponse{CostUSD: 0.05})

	totalCost := monitor.GetUserTotalCost(444)
	assert.Equal(t, 0.30, totalCost)
}

// TestCostMonitor_SetBudget_StoresBudgetLimit verifies budget setting
func TestCostMonitor_SetBudget_StoresBudgetLimit(t *testing.T) {
	monitor := NewCostMonitor()

	err := monitor.SetUserBudget(context.Background(), 555, 10.0)
	assert.NoError(t, err)

	budget := monitor.GetUserBudget(555)
	assert.Equal(t, 10.0, budget)
}

// TestCostMonitor_IsWithinBudget_TrueWhenUnderLimit verifies budget check
func TestCostMonitor_IsWithinBudget_TrueWhenUnderLimit(t *testing.T) {
	monitor := NewCostMonitor()

	monitor.SetUserBudget(context.Background(), 666, 1.0)
	monitor.RecordUsage(context.Background(), 666, "review", &AIRequest{}, &AIResponse{CostUSD: 0.50})

	withinBudget := monitor.IsWithinBudget(666)
	assert.True(t, withinBudget)
}

// TestCostMonitor_IsWithinBudget_FalseWhenOverLimit verifies budget exceeded
func TestCostMonitor_IsWithinBudget_FalseWhenOverLimit(t *testing.T) {
	monitor := NewCostMonitor()

	monitor.SetUserBudget(context.Background(), 777, 0.50)
	monitor.RecordUsage(context.Background(), 777, "review", &AIRequest{}, &AIResponse{CostUSD: 0.75})

	withinBudget := monitor.IsWithinBudget(777)
	assert.False(t, withinBudget)
}

// TestCostMonitor_IsWithinBudget_TrueWhenNoBudget verifies no budget = unlimited
func TestCostMonitor_IsWithinBudget_TrueWhenNoBudget(t *testing.T) {
	monitor := NewCostMonitor()

	// Don't set a budget
	monitor.RecordUsage(context.Background(), 888, "review", &AIRequest{}, &AIResponse{CostUSD: 100.0})

	withinBudget := monitor.IsWithinBudget(888)
	assert.True(t, withinBudget, "Should be unlimited without budget")
}

// TestCostMonitor_GetRemainingBudget_CalculatesCorrectly verifies remaining calc
func TestCostMonitor_GetRemainingBudget_CalculatesCorrectly(t *testing.T) {
	monitor := NewCostMonitor()

	monitor.SetUserBudget(context.Background(), 999, 5.0)
	monitor.RecordUsage(context.Background(), 999, "review", &AIRequest{}, &AIResponse{CostUSD: 2.0})

	remaining := monitor.GetRemainingBudget(999)
	assert.Equal(t, 3.0, remaining)
}

// TestCostMonitor_GetPercentageUsed_CalculatesCorrectly verifies percentage
func TestCostMonitor_GetPercentageUsed_CalculatesCorrectly(t *testing.T) {
	monitor := NewCostMonitor()

	monitor.SetUserBudget(context.Background(), 1001, 10.0)
	monitor.RecordUsage(context.Background(), 1001, "review", &AIRequest{}, &AIResponse{CostUSD: 5.0})

	percentage := monitor.GetPercentageUsed(1001)
	assert.Equal(t, 50.0, percentage)
}

// TestCostMonitor_RecordUsage_TracksDuration verifies response time
func TestCostMonitor_RecordUsage_TracksDuration(t *testing.T) {
	monitor := NewCostMonitor()

	resp := &AIResponse{
		CostUSD:      0.01,
		ResponseTime: 5 * time.Second,
	}

	monitor.RecordUsage(context.Background(), 1002, "review", &AIRequest{}, resp)

	usage := monitor.GetUserUsageStats(1002)
	assert.NotNil(t, usage)
	assert.Equal(t, int64(1), usage.RequestCount)
	assert.Equal(t, int64(5000), usage.TotalResponseTimeMs)
}

// TestCostMonitor_GetAppTotalCost_SumAcrossAllUsers verifies app-wide cost
func TestCostMonitor_GetAppTotalCost_SumAcrossAllUsers(t *testing.T) {
	monitor := NewCostMonitor()

	monitor.RecordUsage(context.Background(), 1003, "review", &AIRequest{}, &AIResponse{CostUSD: 0.10})
	monitor.RecordUsage(context.Background(), 1004, "review", &AIRequest{}, &AIResponse{CostUSD: 0.15})
	monitor.RecordUsage(context.Background(), 1005, "review", &AIRequest{}, &AIResponse{CostUSD: 0.05})

	appCost := monitor.GetAppTotalCost("review")
	assert.Equal(t, 0.30, appCost)
}

// TestCostMonitor_GetAverageCostPerRequest verifies average cost
func TestCostMonitor_GetAverageCostPerRequest_Calculates(t *testing.T) {
	monitor := NewCostMonitor()

	monitor.RecordUsage(context.Background(), 1006, "review", &AIRequest{}, &AIResponse{CostUSD: 0.10})
	monitor.RecordUsage(context.Background(), 1006, "review", &AIRequest{}, &AIResponse{CostUSD: 0.20})
	monitor.RecordUsage(context.Background(), 1006, "review", &AIRequest{}, &AIResponse{CostUSD: 0.30})

	stats := monitor.GetAppStats("review")
	assert.NotNil(t, stats)
	assert.InDelta(t, 0.20, stats.AverageCostPerRequest, 0.0001)
}

// TestCostMonitor_GetMostExpensiveUser verifies ranking
func TestCostMonitor_GetMostExpensiveUser_IdentifiesHighestUser(t *testing.T) {
	monitor := NewCostMonitor()

	monitor.RecordUsage(context.Background(), 2001, "review", &AIRequest{}, &AIResponse{CostUSD: 0.50})
	monitor.RecordUsage(context.Background(), 2002, "review", &AIRequest{}, &AIResponse{CostUSD: 1.50})
	monitor.RecordUsage(context.Background(), 2003, "review", &AIRequest{}, &AIResponse{CostUSD: 0.75})

	topUser := monitor.GetTopUsers(1)
	assert.Equal(t, 1, len(topUser))
	assert.Equal(t, int64(2002), topUser[0].UserID)
	assert.Equal(t, 1.50, topUser[0].TotalCost)
}

// TestCostMonitor_GetTopUsers_ReturnsMultiple verifies top N
func TestCostMonitor_GetTopUsers_ReturnsMultiple(t *testing.T) {
	monitor := NewCostMonitor()

	users := []struct {
		id   int64
		cost float64
	}{
		{3001, 3.0},
		{3002, 2.0},
		{3003, 1.0},
		{3004, 0.5},
	}

	for _, u := range users {
		monitor.RecordUsage(context.Background(), u.id, "review", &AIRequest{}, &AIResponse{CostUSD: u.cost})
	}

	topThree := monitor.GetTopUsers(3)
	assert.Equal(t, 3, len(topThree))
	assert.Equal(t, int64(3001), topThree[0].UserID)
	assert.Equal(t, int64(3002), topThree[1].UserID)
	assert.Equal(t, int64(3003), topThree[2].UserID)
}

// TestCostMonitor_GetCostTrend_TracksCostOverTime verifies trend analysis
func TestCostMonitor_GetCostTrend_TracksCostOverTime(t *testing.T) {
	monitor := NewCostMonitor()

	// Simulate usage over time (in real impl would track timestamps)
	for i := 0; i < 5; i++ {
		monitor.RecordUsage(context.Background(), 4001, "review", &AIRequest{}, &AIResponse{CostUSD: 0.10})
	}

	trend := monitor.GetUserCostTrend(4001)
	assert.NotNil(t, trend)
	assert.Equal(t, 0.50, trend.TotalCost)
	assert.Equal(t, int64(5), trend.RequestCount)
}

// TestCostMonitor_AlertThreshold_TriggersWhenExceeded verifies alerts
func TestCostMonitor_AlertThreshold_TriggersWhenExceeded(t *testing.T) {
	monitor := NewCostMonitor()

	monitor.SetAlertThreshold(context.Background(), 5001, 0.50)

	// First usage: under threshold
	monitor.RecordUsage(context.Background(), 5001, "review", &AIRequest{}, &AIResponse{CostUSD: 0.30})
	alert1 := monitor.HasPendingAlert(5001)
	assert.False(t, alert1, "No alert when under threshold")

	// Second usage: exceeds threshold
	monitor.RecordUsage(context.Background(), 5001, "review", &AIRequest{}, &AIResponse{CostUSD: 0.30})
	alert2 := monitor.HasPendingAlert(5001)
	assert.True(t, alert2, "Alert when exceeds threshold")
}

// TestCostMonitor_ClearAlert_RemovesAlert verifies alert clearing
func TestCostMonitor_ClearAlert_RemovesAlert(t *testing.T) {
	monitor := NewCostMonitor()

	monitor.SetAlertThreshold(context.Background(), 6001, 0.50)
	monitor.RecordUsage(context.Background(), 6001, "review", &AIRequest{}, &AIResponse{CostUSD: 0.60})

	alert1 := monitor.HasPendingAlert(6001)
	assert.True(t, alert1)

	monitor.ClearAlert(context.Background(), 6001)

	alert2 := monitor.HasPendingAlert(6001)
	assert.False(t, alert2)
}

// TestCostMonitor_Concurrency_ThreadSafe verifies concurrent access
func TestCostMonitor_Concurrency_ThreadSafe(t *testing.T) {
	monitor := NewCostMonitor()

	done := make(chan bool, 20)
	for i := 0; i < 20; i++ {
		go func(userID int64) {
			for j := 0; j < 50; j++ {
				monitor.RecordUsage(context.Background(), userID, "review", &AIRequest{}, &AIResponse{CostUSD: 0.01})
			}
			done <- true
		}(int64(7000 + i))
	}

	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify all costs were recorded correctly
	totalCost := 0.0
	for i := 0; i < 20; i++ {
		userID := int64(7000 + i)
		cost := monitor.GetUserTotalCost(userID)
		totalCost += cost
		assert.InDelta(t, 0.50, cost, 0.0001, "Each user should have 50 requests * 0.01 = 0.50")
	}

	assert.InDelta(t, 10.0, totalCost, 0.001, "Total across all 20 users should be 10.0")
}
