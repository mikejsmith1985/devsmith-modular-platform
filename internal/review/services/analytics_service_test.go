package review_services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAnalyticsService_RecordMetric_Success(t *testing.T) {
	// GIVEN: Analytics service and metric
	service := NewAnalyticsService()
	metric := &AnalyticsMetric{
		ReadingMode: "skim",
		Duration:    100 * time.Millisecond,
		Success:     true,
		CostInCents: 25,
		ModelUsed:   "qwen2.5",
		SessionID:   123,
		UserID:      1,
	}

	// WHEN: Recording metric
	err := service.RecordMetric(context.Background(), metric)

	// THEN: Should succeed
	assert.NoError(t, err)

	// AND: Count should increase
	count, _ := service.GetMetricsCount(context.Background())
	assert.Equal(t, 1, count)
}

func TestAnalyticsService_RecordMetric_Nil(t *testing.T) {
	// GIVEN: Analytics service
	service := NewAnalyticsService()

	// WHEN: Recording nil metric
	err := service.RecordMetric(context.Background(), nil)

	// THEN: Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot record nil metric")
}

func TestAnalyticsService_RecordMetric_ContextCancelled(t *testing.T) {
	// GIVEN: Cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	service := NewAnalyticsService()
	metric := &AnalyticsMetric{ReadingMode: "skim"}

	// WHEN: Recording with cancelled context
	err := service.RecordMetric(ctx, metric)

	// THEN: Should return error
	assert.Error(t, err)
}

func TestAnalyticsService_GetModeStats_Success(t *testing.T) {
	// GIVEN: Service with recorded metrics
	service := NewAnalyticsService()
	_ = service.RecordMetric(context.Background(), &AnalyticsMetric{
		ReadingMode: "skim",
		Duration:    100 * time.Millisecond,
		Success:     true,
		CostInCents: 25,
	})
	_ = service.RecordMetric(context.Background(), &AnalyticsMetric{
		ReadingMode: "skim",
		Duration:    150 * time.Millisecond,
		Success:     true,
		CostInCents: 30,
	})

	// WHEN: Getting stats for mode
	stats, err := service.GetModeStats(context.Background(), "skim")

	// THEN: Should succeed
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "skim", stats.Mode)
	assert.Equal(t, int64(2), stats.TotalCalls)
	assert.Equal(t, int64(2), stats.SuccessfulCalls)
	assert.Equal(t, int64(55), stats.TotalCostCents)
}

func TestAnalyticsService_GetModeStats_NotFound(t *testing.T) {
	// GIVEN: Empty service
	service := NewAnalyticsService()

	// WHEN: Getting stats for non-existent mode
	stats, err := service.GetModeStats(context.Background(), "nonexistent")

	// THEN: Should return error
	assert.Error(t, err)
	assert.Nil(t, stats)
}

func TestAnalyticsService_GetAllModeStats(t *testing.T) {
	// GIVEN: Service with multiple modes
	service := NewAnalyticsService()
	_ = service.RecordMetric(context.Background(), &AnalyticsMetric{
		ReadingMode: "skim",
		Success:     true,
		CostInCents: 25,
	})
	_ = service.RecordMetric(context.Background(), &AnalyticsMetric{
		ReadingMode: "scan",
		Success:     true,
		CostInCents: 35,
	})
	_ = service.RecordMetric(context.Background(), &AnalyticsMetric{
		ReadingMode: "critical",
		Success:     true,
		CostInCents: 50,
	})

	// WHEN: Getting all stats
	allStats, err := service.GetAllModeStats(context.Background())

	// THEN: Should succeed
	assert.NoError(t, err)
	assert.Len(t, allStats, 3)
	assert.NotNil(t, allStats["skim"])
	assert.NotNil(t, allStats["scan"])
	assert.NotNil(t, allStats["critical"])
}

func TestAnalyticsService_GetTotalCost(t *testing.T) {
	// GIVEN: Service with mixed success/failure metrics
	service := NewAnalyticsService()
	_ = service.RecordMetric(context.Background(), &AnalyticsMetric{
		ReadingMode: "skim",
		Success:     true,
		CostInCents: 25,
	})
	_ = service.RecordMetric(context.Background(), &AnalyticsMetric{
		ReadingMode: "skim",
		Success:     false, // Failed - no cost
		CostInCents: 10,
	})
	_ = service.RecordMetric(context.Background(), &AnalyticsMetric{
		ReadingMode: "scan",
		Success:     true,
		CostInCents: 35,
	})

	// WHEN: Getting total cost
	totalCost, err := service.GetTotalCost(context.Background())

	// THEN: Should only count successful calls
	assert.NoError(t, err)
	assert.Equal(t, int64(60), totalCost) // 25 + 35, not including failed 10
}

func TestAnalyticsService_GetSuccessRate(t *testing.T) {
	// GIVEN: Service with known successes/failures
	service := NewAnalyticsService()
	for i := 0; i < 7; i++ {
		_ = service.RecordMetric(context.Background(), &AnalyticsMetric{
			ReadingMode: "skim",
			Success:     true,
		})
	}
	for i := 0; i < 3; i++ {
		_ = service.RecordMetric(context.Background(), &AnalyticsMetric{
			ReadingMode: "skim",
			Success:     false,
		})
	}

	// WHEN: Getting success rate
	rate, err := service.GetSuccessRate(context.Background())

	// THEN: Should be 70%
	assert.NoError(t, err)
	assert.InDelta(t, 70.0, rate, 0.1)
}

func TestAnalyticsService_GetSuccessRate_EmptyMetrics(t *testing.T) {
	// GIVEN: Empty service
	service := NewAnalyticsService()

	// WHEN: Getting success rate with no metrics
	rate, err := service.GetSuccessRate(context.Background())

	// THEN: Should return 0%
	assert.NoError(t, err)
	assert.Equal(t, 0.0, rate)
}

func TestAnalyticsService_Reset(t *testing.T) {
	// GIVEN: Service with metrics
	service := NewAnalyticsService()
	_ = service.RecordMetric(context.Background(), &AnalyticsMetric{
		ReadingMode: "skim",
		Success:     true,
		CostInCents: 25,
	})

	// Verify data exists
	count, _ := service.GetMetricsCount(context.Background())
	assert.Equal(t, 1, count)

	// WHEN: Resetting
	err := service.Reset(context.Background())

	// THEN: Should succeed
	assert.NoError(t, err)

	// AND: All data should be cleared
	count, _ = service.GetMetricsCount(context.Background())
	assert.Equal(t, 0, count)

	cost, _ := service.GetTotalCost(context.Background())
	assert.Equal(t, int64(0), cost)

	allStats, _ := service.GetAllModeStats(context.Background())
	assert.Empty(t, allStats)
}

func TestAnalyticsService_MultipleModesStats(t *testing.T) {
	// GIVEN: Service with mixed mode calls
	service := NewAnalyticsService()

	modes := []struct {
		name  string
		count int
		cost  int
	}{
		{"skim", 10, 250},
		{"scan", 5, 175},
		{"detailed", 3, 300},
		{"critical", 2, 500},
	}

	for _, m := range modes {
		for i := 0; i < m.count; i++ {
			_ = service.RecordMetric(context.Background(), &AnalyticsMetric{
				ReadingMode: m.name,
				Success:     true,
				CostInCents: m.cost / m.count,
			})
		}
	}

	// WHEN: Getting all stats
	allStats, _ := service.GetAllModeStats(context.Background())

	// THEN: Should have correct counts for each mode
	for _, m := range modes {
		assert.Equal(t, int64(m.count), allStats[m.name].TotalCalls)
		assert.Equal(t, int64(m.count), allStats[m.name].SuccessfulCalls)
	}
}

func TestAnalyticsService_AverageDuration(t *testing.T) {
	// GIVEN: Service with known durations
	service := NewAnalyticsService()
	durations := []time.Duration{100, 200, 300, 400, 500}

	for _, d := range durations {
		_ = service.RecordMetric(context.Background(), &AnalyticsMetric{
			ReadingMode: "skim",
			Duration:    d * time.Millisecond,
			Success:     true,
			CostInCents: 25,
		})
	}

	// WHEN: Getting mode stats
	stats, _ := service.GetModeStats(context.Background(), "skim")

	// THEN: Average should be 300ms
	assert.InDelta(t, 300.0, stats.AvgDurationMs, 0.1)
}
