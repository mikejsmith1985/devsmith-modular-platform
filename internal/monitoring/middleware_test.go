package monitoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBasicMonitoring tests that the monitoring package can be imported and compiled
func TestBasicMonitoring(t *testing.T) {
	// Test that alert thresholds work
	thresholds := DefaultAlertThresholds()
	assert.Equal(t, 5.0, thresholds.APIErrorRate)
	assert.Equal(t, int64(2000), thresholds.ResponseTimeP95)
	assert.Equal(t, 2, thresholds.ServiceDown)
}

// TestMonitoringConfig tests monitoring configuration
func TestMonitoringConfig(t *testing.T) {
	serviceName := "review-service"
	config := DefaultMonitoringConfig(serviceName)

	assert.Equal(t, serviceName, config.ServiceName)
	assert.True(t, config.EnableRealTime)
	assert.Equal(t, 30, config.RetentionDays)
	assert.Equal(t, 5.0, config.Thresholds.APIErrorRate)
}
