// Package review_services provides business logic services for the Review Service
package review_services

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AnalyticsMetric represents a single analytics data point
type AnalyticsMetric struct {
	TimestampRecorded time.Time     `json:"timestamp"`
	ReadingMode       string        `json:"reading_mode"`
	ModelUsed         string        `json:"model_used"`
	Duration          time.Duration `json:"duration_ms"`
	CostInCents       int           `json:"cost_cents"`
	SessionID         int64         `json:"session_id"`
	UserID            int64         `json:"user_id"`
	Success           bool          `json:"success"`
}

// ModeUsageStats tracks usage statistics per reading mode
type ModeUsageStats struct {
	LastUpdated     time.Time
	Mode            string
	TotalCalls      int64
	SuccessfulCalls int64
	FailedCalls     int64
	AvgDurationMs   float64
	TotalCostCents  int64
}

// AnalyticsService tracks review service metrics for cost and performance
type AnalyticsService struct {
	modeStats      map[string]*ModeUsageStats
	metrics        []AnalyticsMetric
	totalCostCents int64
	mu             sync.RWMutex
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{
		metrics:   make([]AnalyticsMetric, 0),
		modeStats: make(map[string]*ModeUsageStats),
	}
}

// RecordMetric records a new analytics metric
func (a *AnalyticsService) RecordMetric(ctx context.Context, metric *AnalyticsMetric) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	if metric == nil {
		return fmt.Errorf("analytics: cannot record nil metric")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Add metric to history
	metric.TimestampRecorded = time.Now()
	a.metrics = append(a.metrics, *metric)

	// Update mode statistics
	a.updateModeStats(metric)

	// Update total cost
	if metric.Success {
		a.totalCostCents += int64(metric.CostInCents)
	}

	return nil
}

// updateModeStats updates statistics for a specific reading mode
func (a *AnalyticsService) updateModeStats(metric *AnalyticsMetric) {
	stats, exists := a.modeStats[metric.ReadingMode]
	if !exists {
		stats = &ModeUsageStats{
			Mode: metric.ReadingMode,
		}
		a.modeStats[metric.ReadingMode] = stats
	}

	stats.TotalCalls++
	if metric.Success {
		stats.SuccessfulCalls++
		stats.TotalCostCents += int64(metric.CostInCents)
	} else {
		stats.FailedCalls++
	}

	// Update average duration
	avgMs := float64(metric.Duration.Milliseconds())
	stats.AvgDurationMs = (stats.AvgDurationMs*(float64(stats.TotalCalls)-1) + avgMs) / float64(stats.TotalCalls)
	stats.LastUpdated = time.Now()
}

// GetModeStats returns statistics for a specific reading mode
func (a *AnalyticsService) GetModeStats(ctx context.Context, mode string) (*ModeUsageStats, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	stats, exists := a.modeStats[mode]
	if !exists {
		return nil, fmt.Errorf("analytics: no stats for mode %s", mode)
	}

	return stats, nil
}

// GetAllModeStats returns statistics for all reading modes
func (a *AnalyticsService) GetAllModeStats(ctx context.Context) (map[string]*ModeUsageStats, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	// Create a copy to avoid external modification
	result := make(map[string]*ModeUsageStats)
	for mode, stats := range a.modeStats {
		result[mode] = stats
	}

	return result, nil
}

// GetTotalCost returns total cost in cents
func (a *AnalyticsService) GetTotalCost(ctx context.Context) (int64, error) {
	if ctx.Err() != nil {
		return 0, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.totalCostCents, nil
}

// GetMetricsCount returns the number of recorded metrics
func (a *AnalyticsService) GetMetricsCount(ctx context.Context) (int, error) {
	if ctx.Err() != nil {
		return 0, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	return len(a.metrics), nil
}

// GetSuccessRate returns the percentage of successful calls
func (a *AnalyticsService) GetSuccessRate(ctx context.Context) (float64, error) {
	if ctx.Err() != nil {
		return 0, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	if len(a.metrics) == 0 {
		return 0, nil
	}

	successful := 0
	for _, m := range a.metrics {
		if m.Success {
			successful++
		}
	}

	return float64(successful) / float64(len(a.metrics)) * 100, nil
}

// Reset clears all analytics data
func (a *AnalyticsService) Reset(ctx context.Context) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.metrics = make([]AnalyticsMetric, 0)
	a.modeStats = make(map[string]*ModeUsageStats)
	a.totalCostCents = 0

	return nil
}
