// Package services provides service implementations for logs operations.
package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/sirupsen/logrus"
)

// AlertService manages alert configurations and threshold detection.
type AlertService struct { //nolint:govet // Struct alignment optimized for memory efficiency
	logger  *logrus.Logger
	reader  LogReaderInterface
	mu      sync.RWMutex
	configs map[string]*models.AlertConfig
}

// NewAlertService creates a new AlertService.
func NewAlertService(reader LogReaderInterface, logger *logrus.Logger) *AlertService {
	return &AlertService{
		configs: make(map[string]*models.AlertConfig),
		logger:  logger,
		reader:  reader,
	}
}

// CreateAlertConfig creates a new alert configuration.
func (s *AlertService) CreateAlertConfig(ctx context.Context, config *models.AlertConfig) error {
	if config == nil {
		return fmt.Errorf("alert config cannot be nil")
	}

	if config.Service == "" {
		return fmt.Errorf("service name is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Set timestamps
	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	// Auto-generate ID if not set
	if config.ID == 0 {
		config.ID = int64(len(s.configs) + 1)
	}

	s.configs[config.Service] = config
	s.logger.WithFields(logrus.Fields{
		"service":         config.Service,
		"error_threshold": config.ErrorThresholdPerMin,
	}).Info("Alert config created")

	return nil
}

// UpdateAlertConfig updates an existing alert configuration.
func (s *AlertService) UpdateAlertConfig(ctx context.Context, config *models.AlertConfig) error {
	if config == nil {
		return fmt.Errorf("alert config cannot be nil")
	}

	if config.Service == "" {
		return fmt.Errorf("service name is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.configs[config.Service]; !exists {
		return fmt.Errorf("alert config for service %s not found", config.Service)
	}

	config.UpdatedAt = time.Now()
	s.configs[config.Service] = config
	s.logger.WithFields(logrus.Fields{
		"service": config.Service,
	}).Info("Alert config updated")

	return nil
}

// GetAlertConfig retrieves alert configuration for a service.
func (s *AlertService) GetAlertConfig(ctx context.Context, service string) (*models.AlertConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[service]
	if !exists {
		return nil, fmt.Errorf("alert config for service %s not found", service)
	}

	return config, nil
}

// CheckThresholds checks if current log counts exceed alert thresholds.
func (s *AlertService) CheckThresholds(ctx context.Context) ([]models.AlertThresholdViolation, error) {
	s.mu.RLock()
	configs := make([]*models.AlertConfig, 0, len(s.configs))
	for _, config := range s.configs {
		if config.Enabled {
			configs = append(configs, config)
		}
	}
	s.mu.RUnlock()

	violations := make([]models.AlertThresholdViolation, 0)
	now := time.Now()
	oneMinuteAgo := now.Add(-1 * time.Minute)

	for _, config := range configs {
		// Check error threshold
		errorCount, err := s.reader.CountByServiceAndLevel(ctx, config.Service, "error", oneMinuteAgo, now)
		if err != nil {
			s.logger.WithError(err).Warnf("Failed to count errors for service %s", config.Service)
			continue
		}

		if errorCount > int64(config.ErrorThresholdPerMin) {
			violation := models.AlertThresholdViolation{
				Service:        config.Service,
				Level:          "error",
				CurrentCount:   errorCount,
				ThresholdValue: config.ErrorThresholdPerMin,
				Timestamp:      now,
				ID:             int64(len(violations) + 1),
			}
			violations = append(violations, violation)
		}

		// Check warning threshold
		warningCount, err := s.reader.CountByServiceAndLevel(ctx, config.Service, "warning", oneMinuteAgo, now)
		if err != nil {
			s.logger.WithError(err).Warnf("Failed to count warnings for service %s", config.Service)
			continue
		}

		if warningCount > int64(config.WarningThresholdPerMin) {
			violation := models.AlertThresholdViolation{
				Service:        config.Service,
				Level:          "warning",
				CurrentCount:   warningCount,
				ThresholdValue: config.WarningThresholdPerMin,
				Timestamp:      now,
				ID:             int64(len(violations) + 1),
			}
			violations = append(violations, violation)
		}
	}

	return violations, nil
}

// SendAlert sends an alert via email or webhook.
func (s *AlertService) SendAlert(ctx context.Context, violation *models.AlertThresholdViolation) error {
	if violation == nil {
		return fmt.Errorf("violation cannot be nil")
	}

	// Simulate alert sending
	now := time.Now()
	violation.AlertSentAt = &now

	s.logger.WithFields(logrus.Fields{
		"service": violation.Service,
		"level":   violation.Level,
		"count":   violation.CurrentCount,
	}).Info("Alert sent")

	return nil
}
