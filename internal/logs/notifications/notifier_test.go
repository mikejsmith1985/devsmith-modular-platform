// Package notifications provides alert notification delivery.
package notifications_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/notifications"
)

// TestEmailNotifierInterface validates the notifier interface exists.
func TestEmailNotifierInterface(t *testing.T) {
	// GIVEN: A notifier interface
	// WHEN: Creating an email notifier
	// THEN: It should implement the NotifierInterface
	var notifier notifications.NotifierInterface
	require.NotNil(t, notifier) // Will fail until interface is created
}

// TestNewEmailNotifier creates a new email notifier.
func TestNewEmailNotifier(t *testing.T) {
	// GIVEN: SMTP configuration
	config := notifications.EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "sender@example.com",
		Password: "password",
		FromAddr: "noreply@example.com",
	}

	// WHEN: Creating an email notifier
	notifier := notifications.NewEmailNotifier(config)

	// THEN: It should be created successfully
	assert.NotNil(t, notifier)
}

// TestEmailNotifierSend sends an email notification.
func TestEmailNotifierSend(t *testing.T) {
	// GIVEN: An email notifier and alert violation
	config := notifications.EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "sender@example.com",
		Password: "password",
		FromAddr: "noreply@example.com",
	}
	notifier := notifications.NewEmailNotifier(config)

	violation := &models.AlertThresholdViolation{
		Service:        "api-service",
		Level:          "error",
		CurrentCount:   150,
		ThresholdValue: 100,
		Timestamp:      time.Now(),
		ID:             1,
	}

	// WHEN: Sending a notification
	err := notifier.Send(context.Background(), violation, "test@example.com")

	// THEN: It should not error (in test mode)
	assert.NoError(t, err)
}

// TestEmailNotifierValidation validates email configuration.
func TestEmailNotifierValidation(t *testing.T) {
	// GIVEN: Invalid SMTP configuration
	config := notifications.EmailConfig{
		Host:     "",
		Port:     0,
		Username: "",
		Password: "",
		FromAddr: "",
	}

	// WHEN: Creating an email notifier with invalid config
	notifier := notifications.NewEmailNotifier(config)

	// THEN: Send should fail with validation error
	violation := &models.AlertThresholdViolation{
		Service: "test-service",
		Level:   "error",
	}

	err := notifier.Send(context.Background(), violation, "test@example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

// TestWebhookNotifier creates a webhook notifier.
func TestNewWebhookNotifier(t *testing.T) {
	// GIVEN: Webhook URL
	baseURL := "https://example.com/webhooks"

	// WHEN: Creating a webhook notifier
	notifier := notifications.NewWebhookNotifier(baseURL)

	// THEN: It should be created successfully
	assert.NotNil(t, notifier)
}

// TestWebhookNotifierSend sends a webhook notification.
func TestWebhookNotifierSend(t *testing.T) {
	// GIVEN: A webhook notifier and alert violation
	notifier := notifications.NewWebhookNotifier("https://example.com/webhooks/alerts")

	violation := &models.AlertThresholdViolation{
		Service:        "api-service",
		Level:          "error",
		CurrentCount:   150,
		ThresholdValue: 100,
		Timestamp:      time.Now(),
		ID:             1,
	}

	// WHEN: Sending a webhook notification
	err := notifier.Send(context.Background(), violation, "https://example.com/webhook")

	// THEN: It should not error (in test mode)
	assert.NoError(t, err)
}

// TestWebhookNotifierValidation validates webhook URL.
func TestWebhookNotifierValidation(t *testing.T) {
	// GIVEN: Webhook notifier with invalid URL
	notifier := notifications.NewWebhookNotifier("")

	// WHEN: Sending to invalid URL
	violation := &models.AlertThresholdViolation{
		Service: "test-service",
		Level:   "error",
	}

	err := notifier.Send(context.Background(), violation, "invalid-url")
	assert.Error(t, err)
}

// TestNotifierMultipleViolations sends multiple violations.
func TestNotifierMultipleViolations(t *testing.T) {
	// GIVEN: Multiple alert violations
	notifier := notifications.NewEmailNotifier(notifications.EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "sender@example.com",
		Password: "password",
		FromAddr: "noreply@example.com",
	})

	violations := []*models.AlertThresholdViolation{
		{
			Service:        "service1",
			Level:          "error",
			CurrentCount:   150,
			ThresholdValue: 100,
			Timestamp:      time.Now(),
		},
		{
			Service:        "service2",
			Level:          "warning",
			CurrentCount:   75,
			ThresholdValue: 50,
			Timestamp:      time.Now(),
		},
	}

	// WHEN: Sending notifications for multiple violations
	for _, v := range violations {
		err := notifier.Send(context.Background(), v, "alerts@example.com")
		// THEN: Each should send successfully
		assert.NoError(t, err)
	}
}

// TestNotifierContextCancellation handles context cancellation.
func TestNotifierContextCancellation(t *testing.T) {
	// GIVEN: A cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	notifier := notifications.NewEmailNotifier(notifications.EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "sender@example.com",
		Password: "password",
		FromAddr: "noreply@example.com",
	})

	violation := &models.AlertThresholdViolation{
		Service: "test-service",
		Level:   "error",
	}

	// WHEN: Sending with cancelled context
	err := notifier.Send(ctx, violation, "test@example.com")

	// THEN: Should respect context cancellation
	assert.Error(t, err)
}
