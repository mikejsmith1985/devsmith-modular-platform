// Package ai provides AI provider abstraction, routing, and cost monitoring.
package ai

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// UserCostRecord stores cost data for a single user
type UserCostRecord struct {
	UserID              int64
	TotalCost           float64
	RequestCount        int64
	TotalResponseTimeMs int64
	LastUsedAt          time.Time
}

// AppStats stores aggregate statistics for an application
type AppStats struct {
	AppName               string
	TotalCost             float64
	RequestCount          int64
	AverageCostPerRequest float64
	AverageResponseTimeMs int64
	UniqueUsers           int
}

// CostMonitor tracks AI usage and costs for users and applications
type CostMonitor struct {
	// userCosts: key is "userID:app", value is cost
	userCosts map[string]float64

	// appCosts: key is app name, value is total cost
	appCosts map[string]float64

	// userStats: key is userID, value is detailed stats
	userStats map[int64]*UserCostRecord

	// appStats: key is app name, value is request count
	appRequestCounts map[string]int64

	// budgets: key is userID, value is budget limit
	budgets map[int64]float64

	// alerts: key is userID, value is alert threshold
	alerts map[int64]float64

	// alertTriggered: key is userID, value is true if alert triggered
	alertTriggered map[int64]bool

	mu sync.RWMutex
}

// NewCostMonitor creates a new cost monitor
func NewCostMonitor() *CostMonitor {
	return &CostMonitor{
		userCosts:        make(map[string]float64),
		appCosts:         make(map[string]float64),
		userStats:        make(map[int64]*UserCostRecord),
		appRequestCounts: make(map[string]int64),
		budgets:          make(map[int64]float64),
		alerts:           make(map[int64]float64),
		alertTriggered:   make(map[int64]bool),
	}
}

// RecordUsage records an AI API usage for cost tracking
func (m *CostMonitor) RecordUsage(ctx context.Context, userID int64, appName string, req *AIRequest, resp *AIResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record user + app specific cost
	key := fmt.Sprintf("%d:%s", userID, appName)
	m.userCosts[key] += resp.CostUSD

	// Record app-wide cost
	m.appCosts[appName] += resp.CostUSD

	// Update user stats
	if _, exists := m.userStats[userID]; !exists {
		m.userStats[userID] = &UserCostRecord{UserID: userID}
	}
	stats := m.userStats[userID]
	stats.TotalCost += resp.CostUSD
	stats.RequestCount++
	stats.TotalResponseTimeMs += resp.ResponseTime.Milliseconds()
	stats.LastUsedAt = time.Now()

	// Update app request count
	m.appRequestCounts[appName]++

	// Check alert threshold
	if threshold, hasAlert := m.alerts[userID]; hasAlert {
		if stats.TotalCost > threshold {
			m.alertTriggered[userID] = true
		}
	}

	return nil
}

// GetUserTotalCost returns total cost for a user across all apps
func (m *CostMonitor) GetUserTotalCost(userID int64) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if stats, exists := m.userStats[userID]; exists {
		return stats.TotalCost
	}
	return 0.0
}

// GetAppCostForUser returns cost for a specific user in a specific app
func (m *CostMonitor) GetAppCostForUser(userID int64, appName string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%d:%s", userID, appName)
	return m.userCosts[key]
}

// GetAppTotalCost returns total cost for an app across all users
func (m *CostMonitor) GetAppTotalCost(appName string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.appCosts[appName]
}

// GetUserUsageStats returns detailed usage statistics for a user
func (m *CostMonitor) GetUserUsageStats(userID int64) *UserCostRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if stats, exists := m.userStats[userID]; exists {
		statsCopy := *stats
		return &statsCopy
	}
	return nil
}

// GetAppStats returns aggregate statistics for an app
func (m *CostMonitor) GetAppStats(appName string) *AppStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalCost := m.appCosts[appName]
	requestCount := m.appRequestCounts[appName]

	avgCost := 0.0
	if requestCount > 0 {
		avgCost = totalCost / float64(requestCount)
	}

	// Count unique users for this app
	uniqueUsers := 0
	for key := range m.userCosts {
		// Parse key format "userID:appName"
		var u int64
		var a string
		if _, err := fmt.Sscanf(key, "%d:%s", &u, &a); err != nil {
			continue // Skip malformed key
		}
		if a == appName {
			uniqueUsers++
		}
	}

	return &AppStats{
		AppName:               appName,
		TotalCost:             totalCost,
		RequestCount:          requestCount,
		AverageCostPerRequest: avgCost,
		UniqueUsers:           uniqueUsers,
	}
}

// SetUserBudget sets a budget limit for a user
func (m *CostMonitor) SetUserBudget(ctx context.Context, userID int64, budgetUSD float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.budgets[userID] = budgetUSD
	return nil
}

// GetUserBudget returns the budget limit for a user
func (m *CostMonitor) GetUserBudget(userID int64) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.budgets[userID]
}

// IsWithinBudget checks if user is within their budget
func (m *CostMonitor) IsWithinBudget(userID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	budget, hasBudget := m.budgets[userID]
	if !hasBudget {
		return true // No budget = unlimited
	}

	var totalCost float64
	if stats, exists := m.userStats[userID]; exists {
		totalCost = stats.TotalCost
	}

	return totalCost <= budget
}

// GetRemainingBudget returns remaining budget for a user
func (m *CostMonitor) GetRemainingBudget(userID int64) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	budget, hasBudget := m.budgets[userID]
	if !hasBudget {
		return 0.0 // Unlimited
	}

	var totalCost float64
	if stats, exists := m.userStats[userID]; exists {
		totalCost = stats.TotalCost
	}

	remaining := budget - totalCost
	if remaining < 0 {
		return 0.0
	}
	return remaining
}

// GetPercentageUsed returns the percentage of budget used
func (m *CostMonitor) GetPercentageUsed(userID int64) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	budget, hasBudget := m.budgets[userID]
	if !hasBudget {
		return 0.0
	}

	var totalCost float64
	if stats, exists := m.userStats[userID]; exists {
		totalCost = stats.TotalCost
	}

	if budget == 0 {
		return 0.0
	}

	return (totalCost / budget) * 100.0
}

// GetTopUsers returns the top N users by total cost
func (m *CostMonitor) GetTopUsers(limit int) []UserCostRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect all user records
	var users []UserCostRecord
	for _, record := range m.userStats {
		users = append(users, *record)
	}

	// Sort by total cost descending
	sort.Slice(users, func(i, j int) bool {
		return users[i].TotalCost > users[j].TotalCost
	})

	// Return top N
	if limit > len(users) {
		limit = len(users)
	}
	return users[:limit]
}

// GetUserCostTrend returns trend data for a user
func (m *CostMonitor) GetUserCostTrend(userID int64) *UserCostRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if stats, exists := m.userStats[userID]; exists {
		statsCopy2 := *stats
		return &statsCopy2
	}
	return nil
}

// SetAlertThreshold sets a cost alert threshold for a user
func (m *CostMonitor) SetAlertThreshold(ctx context.Context, userID int64, thresholdUSD float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.alerts[userID] = thresholdUSD
	return nil
}

// HasPendingAlert checks if there's a pending alert for a user
func (m *CostMonitor) HasPendingAlert(userID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.alertTriggered[userID]
}

// ClearAlert clears a pending alert for a user
func (m *CostMonitor) ClearAlert(ctx context.Context, userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.alertTriggered[userID] = false
	return nil
}
