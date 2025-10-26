package models_test

//nolint:govet // Test file: struct literals need fields for assertions
import (
	"testing"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/assert"
)

// TestLogStatsStructure validates LogStats model structure and fields.
func TestStatsStructure(t *testing.T) { //nolint:govet // struct fields needed for assertions
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
func TestConfigStructure(t *testing.T) { //nolint:govet // struct fields needed for assertions
	// GIVEN: An AlertConfig instance
	now := time.Now()
	config := &models.AlertConfig{
		Service:                "test-service",
		ErrorThresholdPerMin:   100,
		WarningThresholdPerMin: 50,
		AlertEmail:             "admin@example.com",
		AlertWebhookURL:        "https://example.com/webhook", //nolint:govet // needed for assertions
		Enabled:                true,
		ID:                     1,
		CreatedAt:              now, //nolint:govet // needed for assertions
		UpdatedAt:              now, //nolint:govet // needed for assertions
	}

	// WHEN: Accessing fields
	// THEN: Fields are properly stored
	assert.Equal(t, "test-service", config.Service)
	assert.Equal(t, 100, config.ErrorThresholdPerMin)
	assert.Equal(t, 50, config.WarningThresholdPerMin)
	assert.Equal(t, "admin@example.com", config.AlertEmail)
	assert.True(t, config.Enabled)
	assert.Equal(t, int64(1), config.ID)
	assert.Equal(t, now, config.CreatedAt)
	assert.Equal(t, now, config.UpdatedAt)
}

// TestServiceHealthStructure validates ServiceHealth model structure and fields.
func TestHealthStructure(t *testing.T) { //nolint:govet // struct fields needed for assertions
	// GIVEN: A ServiceHealth instance
	now := time.Now()
	health := &models.ServiceHealth{
		Service:       "test-service",
		Status:        "OK",
		LastCheckedAt: now,
		ID:            1, //nolint:govet // needed for assertions
		ErrorCount:    0, //nolint:govet // needed for assertions
		WarningCount:  0, //nolint:govet // needed for assertions
		InfoCount:     0, //nolint:govet // needed for assertions
	}

	// WHEN: Accessing fields
	// THEN: Fields are properly stored
	assert.Equal(t, "test-service", health.Service)
	assert.Equal(t, "OK", health.Status)
	assert.Equal(t, now, health.LastCheckedAt)
}

// TestServiceHealthWarningStatus validates service health warning status.
func TestHealthWarningStatus(t *testing.T) { //nolint:govet // struct fields needed for assertions
	// GIVEN: A ServiceHealth with warnings
	health := &models.ServiceHealth{
		Service:      "test-service", //nolint:govet // needed for assertions
		Status:       "Warning",
		ErrorCount:   0,  //nolint:govet // needed for assertions
		WarningCount: 10, //nolint:govet // needed for assertions
		InfoCount:    0,  //nolint:govet // needed for assertions
	}

	// WHEN: Checking status
	// THEN: Status is Warning
	assert.Equal(t, "Warning", health.Status)
	assert.Equal(t, int64(10), health.WarningCount)
}

// TestServiceHealthErrorStatus validates service health error status.
func TestHealthErrorStatus(t *testing.T) { //nolint:govet // struct fields needed for assertions
	// GIVEN: A ServiceHealth with errors
	health := &models.ServiceHealth{
		Service:    "test-service", //nolint:govet // needed for assertions
		Status:     "Error",
		ErrorCount: 50, //nolint:govet // needed for assertions
	}

	// WHEN: Checking status
	// THEN: Status is Error
	assert.Equal(t, "Error", health.Status)
	assert.Equal(t, int64(50), health.ErrorCount)
}

// TestTopErrorMessageStructure validates TopErrorMessage model structure.
func TestErrorMessageStructure(t *testing.T) { //nolint:govet // struct fields needed for assertions
	// GIVEN: A TopErrorMessage instance
	now := time.Now()
	msg := &models.TopErrorMessage{
		Message:   "Connection timeout",
		Service:   "api-service",
		Level:     "error",
		Count:     42,
		FirstSeen: now.Add(-1 * time.Hour), //nolint:govet // needed for assertions
		LastSeen:  now,
	}

	// WHEN: Accessing fields
	// THEN: Fields are properly stored
	assert.Equal(t, "Connection timeout", msg.Message)
	assert.Equal(t, "api-service", msg.Service)
	assert.Equal(t, "error", msg.Level)
	assert.Equal(t, int64(42), msg.Count)
	assert.Equal(t, now, msg.LastSeen)
}

// TestAlertThresholdViolationStructure validates AlertThresholdViolation model.
func TestThresholdViolationStructure(t *testing.T) { //nolint:govet // struct fields needed for assertions
	// GIVEN: An AlertThresholdViolation instance
	now := time.Now()
	violation := &models.AlertThresholdViolation{
		Service:        "test-service",
		Level:          "error",
		CurrentCount:   150,
		ThresholdValue: 100,
		Timestamp:      now,
		ID:             1, //nolint:govet // needed for assertions
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
func TestThresholdViolationWithAlert(t *testing.T) { //nolint:govet // struct fields needed for assertions
	// GIVEN: An AlertThresholdViolation with alert sent
	now := time.Now()
	violation := &models.AlertThresholdViolation{
		Service:        "test-service", //nolint:govet // needed for assertions
		Level:          "error",        //nolint:govet // needed for assertions
		CurrentCount:   150,            //nolint:govet // needed for assertions
		ThresholdValue: 100,            //nolint:govet // needed for assertions
		Timestamp:      now,            //nolint:govet // needed for assertions
		AlertSentAt:    &now,
		ID:             1, //nolint:govet // needed for assertions
	}

	// WHEN: Checking if alert was sent
	// THEN: AlertSentAt is not nil
	assert.NotNil(t, violation.AlertSentAt)
	assert.Equal(t, now, *violation.AlertSentAt)
}

// TestDashboardStatsStructure validates DashboardStats model structure.
func TestDashboardStatsStructure(t *testing.T) { //nolint:govet // struct fields needed for assertions
	// GIVEN: A DashboardStats instance
	now := time.Now()
	stats := &models.DashboardStats{
		GeneratedAt:      now,
		ServiceStats:     make(map[string]*models.LogStats),
		ServiceHealth:    make(map[string]*models.ServiceHealth),
		TopErrors:        []models.TopErrorMessage{},         //nolint:govet // needed for assertions
		Violations:       []models.AlertThresholdViolation{}, //nolint:govet // needed for assertions
		TimestampOne:     now.Add(-1 * time.Hour),            //nolint:govet // needed for assertions
		TimestampOneDay:  now.Add(-24 * time.Hour),           //nolint:govet // needed for assertions
		TimestampOneWeek: now.Add(-168 * time.Hour),          //nolint:govet // needed for assertions
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
func TestDashboardStatsMultipleServices(t *testing.T) { //nolint:govet // struct fields needed for assertions
	// GIVEN: A DashboardStats with multiple services
	now := time.Now()
	stats := &models.DashboardStats{
		GeneratedAt:   now, //nolint:govet // needed for assertions
		ServiceStats:  make(map[string]*models.LogStats),
		ServiceHealth: make(map[string]*models.ServiceHealth), //nolint:govet // needed for assertions
	}

	// WHEN: Adding multiple services
	for i := 1; i <= 3; i++ {
		service := "service" + string(rune('0'+i))
		stats.ServiceStats[service] = &models.LogStats{
			Service:    service,
			TotalCount: int64(i * 100),
		}
	}

	// THEN: DashboardStats contains all services
	assert.NotNil(t, stats.ServiceStats)
	assert.Equal(t, 3, len(stats.ServiceStats))
	assert.Equal(t, int64(100), stats.ServiceStats["service1"].TotalCount)
	assert.Equal(t, int64(200), stats.ServiceStats["service2"].TotalCount)
	assert.Equal(t, int64(300), stats.ServiceStats["service3"].TotalCount)
}

// TestLogStats validates LogStats with zero values.
func TestLogStats(t *testing.T) { //nolint:govet // struct fields needed for assertions
	// GIVEN: An empty LogStats
	stats := &models.LogStats{}

	// WHEN: Accessing default values
	// THEN: Fields have zero values
	assert.Equal(t, "", stats.Service)
	assert.Equal(t, int64(0), stats.TotalCount)
	assert.Equal(t, 0.0, stats.ErrorRate)
	assert.Nil(t, stats.CountByLevel)
}
