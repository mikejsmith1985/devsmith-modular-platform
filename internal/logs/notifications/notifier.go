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

// EmailNotifier sends alerts via email.
type EmailNotifier struct { //nolint:govet // struct alignment optimized for readability
	logger *logrus.Logger
	config EmailConfig
}

// NewEmailNotifier creates a new email notifier.
func NewEmailNotifier(config EmailConfig) *EmailNotifier {
	return &EmailNotifier{
		config: config,
		logger: logrus.New(),
	}
}

// Send sends an email notification.
func (en *EmailNotifier) Send(ctx context.Context, violation *models.AlertThresholdViolation, recipient string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	// Validate configuration
	if en.config.Host == "" || en.config.Port == 0 || en.config.FromAddr == "" {
		return errors.New("invalid email configuration")
	}

	if recipient == "" {
		return errors.New("recipient email address is required")
	}

	// Create email message
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

	// Send email via SMTP
	auth := smtp.PlainAuth("", en.config.Username, en.config.Password, en.config.Host)
	addr := fmt.Sprintf("%s:%d", en.config.Host, en.config.Port)

	err := smtp.SendMail(addr, auth, en.config.FromAddr, []string{recipient}, []byte(message))
	if err != nil {
		en.logger.WithError(err).Errorf("Failed to send email alert to %s", recipient)
		return fmt.Errorf("failed to send email: %w", err)
	}

	en.logger.Infof("Email alert sent to %s for service %s", recipient, violation.Service)
	return nil
}

// WebhookNotifier sends alerts via webhook.
type WebhookNotifier struct { //nolint:govet // struct alignment optimized for readability
	logger  *logrus.Logger
	baseURL string
}

// NewWebhookNotifier creates a new webhook notifier.
func NewWebhookNotifier(baseURL string) *WebhookNotifier {
	return &WebhookNotifier{
		baseURL: baseURL,
		logger:  logrus.New(),
	}
}

// Send sends a webhook notification.
func (wn *WebhookNotifier) Send(ctx context.Context, violation *models.AlertThresholdViolation, webhookURL string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	// Validate webhook URL
	if webhookURL == "" {
		return errors.New("webhook URL is required")
	}

	_, err := url.Parse(webhookURL)
	if err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	// TODO: Implement actual webhook delivery using http.Post
	// For now, this is a placeholder that validates the input

	wn.logger.Infof("Webhook alert would be sent to %s for service %s", webhookURL, violation.Service)
	return nil
}
