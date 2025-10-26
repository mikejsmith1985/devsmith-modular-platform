package models_test

//nolint:govet // Test file: struct literals need fields for assertions
import (
	"testing"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/assert"
)

// TestLogStatsStructure validates LogStats model structure and fields.
func TestLogStatsStructure(t *testing.T) {
	// GIVEN: A LogStats instance
	now := time.Now()
	stats := &models.LogStats{
		Timestamp:  now,
		Service:    "test-service",
		TotalCount: 100,
		ErrorRate:  0.05,
		ID:         1,
		CountByLevel: map[string]int64{
			"error":   5,
			"warning": 10,
			"info":    85,
		},
	}

	// WHEN: Accessing fields
	// THEN: Fields are properly stored
	assert.Equal(t, now, stats.Timestamp)
	assert.Equal(t, "test-service", stats.Service)
	assert.Equal(t, int64(100), stats.TotalCount)
	assert.Equal(t, 0.05, stats.ErrorRate)
	assert.Equal(t, int64(1), stats.ID)
	assert.Equal(t, int64(5), stats.CountByLevel["error"])
	assert.Equal(t, int64(10), stats.CountByLevel["warning"])
	assert.Equal(t, int64(85), stats.CountByLevel["info"])
}

// TestAlertConfigStructure validates AlertConfig model structure and fields.
func TestAlertConfigStructure(t *testing.T) {
	// GIVEN: An AlertConfig instance
	now := time.Now()
	config := &models.AlertConfig{
		Service:                "test-service",
		ErrorThresholdPerMin:   100,
		WarningThresholdPerMin: 50,
		AlertEmail:             "admin@example.com",
		AlertWebhookURL:        "https://example.com/webhook",
		Enabled:                true,
		ID:                     1,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	// WHEN: Accessing fields
	// THEN: Fields are properly stored
	assert.Equal(t, "test-service", config.Service)
	assert.Equal(t, 100, config.ErrorThresholdPerMin)
	assert.Equal(t, 50, config.WarningThresholdPerMin)
	assert.Equal(t, "admin@example.com", config.AlertEmail)
	assert.Equal(t, "https://example.com/webhook", config.AlertWebhookURL)
	assert.True(t, config.Enabled)
	assert.Equal(t, int64(1), config.ID)
}

// TestServiceHealthStructure validates ServiceHealth model structure and fields.
func TestServiceHealthStructure(t *testing.T) {
	// GIVEN: A ServiceHealth instance
	now := time.Now()
	health := &models.ServiceHealth{
		Service:       "test-service",
		Status:        "OK",
		LastCheckedAt: now,
		ErrorCount:    0,
		WarningCount:  2,
		InfoCount:     150,
		ID:            1,
	}

	// WHEN: Accessing fields
	// THEN: Fields are properly stored
	assert.Equal(t, "test-service", health.Service)
	assert.Equal(t, "OK", health.Status)
	assert.Equal(t, now, health.LastCheckedAt)
	assert.Equal(t, int64(0), health.ErrorCount)
	assert.Equal(t, int64(2), health.WarningCount)
	assert.Equal(t, int64(150), health.InfoCount)
}

// TestServiceHealthWarningStatus validates service health warning status.
func TestServiceHealthWarningStatus(t *testing.T) {
	// GIVEN: A ServiceHealth with warnings
	health := &models.ServiceHealth{
		Service:      "test-service",
		Status:       "Warning",
		ErrorCount:   0,
		WarningCount: 10,
	}

	// WHEN: Checking status
	// THEN: Status is Warning
	assert.Equal(t, "Warning", health.Status)
}

// TestServiceHealthErrorStatus validates service health error status.
func TestServiceHealthErrorStatus(t *testing.T) {
	// GIVEN: A ServiceHealth with errors
	health := &models.ServiceHealth{
		Service:    "test-service",
		Status:     "Error",
		ErrorCount: 50,
	}

	// WHEN: Checking status
	// THEN: Status is Error
	assert.Equal(t, "Error", health.Status)
}

// TestTopErrorMessageStructure validates TopErrorMessage model structure.
func TestTopErrorMessageStructure(t *testing.T) {
	// GIVEN: A TopErrorMessage instance
	now := time.Now()
	msg := &models.TopErrorMessage{
		Message:   "database connection failed",
		Service:   "test-service",
		Level:     "error",
		Count:     25,
		FirstSeen: now.Add(-1 * time.Hour),
		LastSeen:  now,
	}

	// WHEN: Accessing fields
	// THEN: Fields are properly stored
	assert.Equal(t, "database connection failed", msg.Message)
	assert.Equal(t, "test-service", msg.Service)
	assert.Equal(t, "error", msg.Level)
	assert.Equal(t, int64(25), msg.Count)
	assert.Equal(t, now.Add(-1*time.Hour), msg.FirstSeen)
	assert.Equal(t, now, msg.LastSeen)
}

// TestAlertThresholdViolationStructure validates AlertThresholdViolation model.
func TestAlertThresholdViolationStructure(t *testing.T) {
	// GIVEN: An AlertThresholdViolation instance
	now := time.Now()
	violation := &models.AlertThresholdViolation{
		Service:        "test-service",
		Level:          "error",
		CurrentCount:   150,
		ThresholdValue: 100,
		Timestamp:      now,
		ID:             1,
	}

	// WHEN: Accessing fields
	// THEN: Fields are properly stored
	assert.Equal(t, "test-service", violation.Service)
	assert.Equal(t, "error", violation.Level)
	assert.Equal(t, int64(150), violation.CurrentCount)
	assert.Equal(t, 100, violation.ThresholdValue)
	assert.Equal(t, now, violation.Timestamp)
	assert.Nil(t, violation.AlertSentAt)
}

// TestAlertThresholdViolationWithAlert validates AlertThresholdViolation with sent alert.
func TestAlertThresholdViolationWithAlert(t *testing.T) {
	// GIVEN: An AlertThresholdViolation with alert sent
	now := time.Now()
	violation := &models.AlertThresholdViolation{
		Service:        "test-service",
		Level:          "error",
		CurrentCount:   150,
		ThresholdValue: 100,
		Timestamp:      now,
		AlertSentAt:    &now,
		ID:             1,
	}

	// WHEN: Checking if alert was sent
	// THEN: AlertSentAt is not nil
	assert.NotNil(t, violation.AlertSentAt)
	assert.Equal(t, now, *violation.AlertSentAt)
}

// TestDashboardStatsStructure validates DashboardStats model structure.
func TestDashboardStatsStructure(t *testing.T) { //nolint:govet
	// GIVEN: A DashboardStats instance
	now := time.Now()
	stats := &models.DashboardStats{
		GeneratedAt:      now,
		ServiceStats:     make(map[string]*models.LogStats),
		ServiceHealth:    make(map[string]*models.ServiceHealth),
		TopErrors:        []models.TopErrorMessage{},
		Violations:       []models.AlertThresholdViolation{},
		TimestampOne:     now.Add(-1 * time.Hour),
		TimestampOneDay:  now.Add(-24 * time.Hour),
		TimestampOneWeek: now.Add(-168 * time.Hour),
	}

	// WHEN: Adding service stats
	stats.ServiceStats["service1"] = &models.LogStats{
		Service:    "service1",
		TotalCount: 100,
	}

	// THEN: DashboardStats properly stores aggregated data
	assert.Equal(t, now, stats.GeneratedAt)
	assert.NotNil(t, stats.ServiceStats)
	assert.NotNil(t, stats.ServiceHealth)
	assert.Equal(t, 1, len(stats.ServiceStats))
	assert.Equal(t, "service1", stats.ServiceStats["service1"].Service)
}

// TestDashboardStatsMultipleServices validates DashboardStats with multiple services.
func TestDashboardStatsMultipleServices(t *testing.T) { //nolint:govet
	// GIVEN: A DashboardStats with multiple services
	now := time.Now()
	stats := &models.DashboardStats{
		GeneratedAt:   now,
		ServiceStats:  make(map[string]*models.LogStats),
		ServiceHealth: make(map[string]*models.ServiceHealth),
	}

	// WHEN: Adding multiple services
	for i := 1; i <= 3; i++ {
		service := "service" + string(rune('0'+i))
		stats.ServiceStats[service] = &models.LogStats{
			Service:    service,
			TotalCount: int64(100 * i),
		}
		stats.ServiceHealth[service] = &models.ServiceHealth{
			Service: service,
			Status:  "OK",
		}
	}

	// THEN: All services are stored
	assert.Equal(t, 3, len(stats.ServiceStats))
	assert.Equal(t, 3, len(stats.ServiceHealth))
}

// TestEmptyLogStats validates LogStats with zero values.
func TestEmptyLogStats(t *testing.T) {
	// GIVEN: An empty LogStats
	stats := &models.LogStats{}

	// THEN: Fields have zero values
	assert.Empty(t, stats.Service)
	assert.Equal(t, int64(0), stats.TotalCount)
	assert.Equal(t, float64(0), stats.ErrorRate)
	assert.Nil(t, stats.CountByLevel)
}
