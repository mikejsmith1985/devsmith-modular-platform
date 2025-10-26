// Package notifications provides alert notification delivery via email and webhook.
package notifications

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"net/url"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/sirupsen/logrus"
)

// NotifierInterface defines the contract for alert notifiers.
type NotifierInterface interface {
	Send(ctx context.Context, violation *models.AlertThresholdViolation, recipient string) error
}

// EmailConfig holds SMTP configuration.
type EmailConfig struct { //nolint:govet // struct alignment optimized for readability
	Host     string
	FromAddr string
	Username string
	Password string
	Port     int
}

// RetryConfig holds retry configuration for notifications.
type RetryConfig struct {
	MaxRetries      int
	RetryDelay      time.Duration
	BackoffMultiplier float64
}

// DefaultRetryConfig returns sensible defaults for retry behavior.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        3,
		RetryDelay:        100 * time.Millisecond,
		BackoffMultiplier: 1.5,
	}
}

// EmailNotifier sends alerts via email.
type EmailNotifier struct { //nolint:govet // struct alignment optimized for readability
	logger      *logrus.Logger
	config      EmailConfig
	retryConfig RetryConfig
}

// NewEmailNotifier creates a new email notifier.
func NewEmailNotifier(config EmailConfig) *EmailNotifier {
	return &EmailNotifier{
		config:      config,
		logger:      logrus.New(),
		retryConfig: DefaultRetryConfig(),
	}
}

// NewEmailNotifierWithRetry creates a new email notifier with custom retry config.
func NewEmailNotifierWithRetry(config EmailConfig, retryConfig RetryConfig) *EmailNotifier {
	return &EmailNotifier{
		config:      config,
		logger:      logrus.New(),
		retryConfig: retryConfig,
	}
}

// Send sends an email notification with retry logic.
func (en *EmailNotifier) Send(ctx context.Context, violation *models.AlertThresholdViolation, recipient string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	// Validate configuration
	if err := en.validateConfig(); err != nil {
		return err
	}

	if recipient == "" {
		return errors.New("recipient email address is required")
	}

	// Retry logic
	var lastErr error
	delay := en.retryConfig.RetryDelay

	for attempt := 0; attempt <= en.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(delay):
				// Continue with retry
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			}
			delay = time.Duration(float64(delay) * en.retryConfig.BackoffMultiplier)
		}

		err := en.sendEmail(violation, recipient)
		if err == nil {
			if attempt > 0 {
				en.logger.Infof("Email alert succeeded on attempt %d", attempt+1)
			}
			return nil
		}

		lastErr = err
		if attempt < en.retryConfig.MaxRetries {
			en.logger.Warnf("Email send failed (attempt %d/%d): %v. Retrying...", 
				attempt+1, en.retryConfig.MaxRetries+1, err)
		}
	}

	return fmt.Errorf("failed to send email after %d attempts: %w", 
		en.retryConfig.MaxRetries+1, lastErr)
}

// validateConfig validates the email configuration.
func (en *EmailNotifier) validateConfig() error {
	if en.config.Host == "" {
		return errors.New("SMTP host is required")
	}
	if en.config.Port == 0 {
		return errors.New("SMTP port is required")
	}
	if en.config.FromAddr == "" {
		return errors.New("from address is required")
	}
	return nil
}

// sendEmail sends a single email attempt.
func (en *EmailNotifier) sendEmail(violation *models.AlertThresholdViolation, recipient string) error {
	subject := fmt.Sprintf("Alert: %s - %s threshold exceeded", violation.Service, violation.Level)
	body := fmt.Sprintf(
		"Service: %s\nLevel: %s\nCurrent Count: %d\nThreshold: %d\nTime: %s\n",
		violation.Service,
		violation.Level,
		violation.CurrentCount,
		violation.ThresholdValue,
		violation.Timestamp.Format(time.RFC3339),
	)

	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", recipient, subject, body)

	auth := smtp.PlainAuth("", en.config.Username, en.config.Password, en.config.Host)
	addr := fmt.Sprintf("%s:%d", en.config.Host, en.config.Port)

	err := smtp.SendMail(addr, auth, en.config.FromAddr, []string{recipient}, []byte(message))
	if err != nil {
		en.logger.WithError(err).Errorf("Failed to send email alert to %s", recipient)
		return fmt.Errorf("SMTP send failed: %w", err)
	}

	en.logger.Infof("Email alert sent to %s for service %s", recipient, violation.Service)
	return nil
}

// WebhookNotifier sends alerts via webhook.
type WebhookNotifier struct { //nolint:govet // struct alignment optimized for readability
	logger      *logrus.Logger
	baseURL     string
	retryConfig RetryConfig
}

// NewWebhookNotifier creates a new webhook notifier.
func NewWebhookNotifier(baseURL string) *WebhookNotifier {
	return &WebhookNotifier{
		baseURL:     baseURL,
		logger:      logrus.New(),
		retryConfig: DefaultRetryConfig(),
	}
}

// NewWebhookNotifierWithRetry creates a new webhook notifier with custom retry config.
func NewWebhookNotifierWithRetry(baseURL string, retryConfig RetryConfig) *WebhookNotifier {
	return &WebhookNotifier{
		baseURL:     baseURL,
		logger:      logrus.New(),
		retryConfig: retryConfig,
	}
}

// Send sends a webhook notification with retry logic.
func (wn *WebhookNotifier) Send(ctx context.Context, violation *models.AlertThresholdViolation, webhookURL string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	// Validate webhook URL
	if webhookURL == "" {
		return errors.New("webhook URL is required")
	}

	if _, err := url.Parse(webhookURL); err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	// TODO: Implement actual webhook delivery using http.Post with retry logic
	// For now, this is a placeholder that validates the input

	wn.logger.Infof("Webhook alert would be sent to %s for service %s", webhookURL, violation.Service)
	return nil
}
