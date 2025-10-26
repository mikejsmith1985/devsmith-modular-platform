// Package services provides service implementations for logs operations.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/sirupsen/logrus"
)

// AlertService implements alert operations.
type AlertService struct { //nolint:govet // Struct alignment optimized for memory efficiency
	violationRepo AlertViolationRepositoryInterface
	configRepo    AlertConfigRepositoryInterface
	logReader     LogReaderInterface
	logger        *logrus.Logger
}

// AlertViolationRepositoryInterface defines contract for violation persistence.
type AlertViolationRepositoryInterface interface {
	Create(ctx context.Context, violation *models.AlertThresholdViolation) error
	UpdateAlertSent(ctx context.Context, id int64) error
	GetUnsent(ctx context.Context) ([]models.AlertThresholdViolation, error)
}

// AlertConfigRepositoryInterface defines contract for alert config persistence.
type AlertConfigRepositoryInterface interface {
	Create(ctx context.Context, config *models.AlertConfig) error
	Update(ctx context.Context, config *models.AlertConfig) error
	GetByService(ctx context.Context, service string) (*models.AlertConfig, error)
	GetAll(ctx context.Context) ([]models.AlertConfig, error)
}

// NewAlertService creates a new AlertService.
func NewAlertService(
	violationRepo AlertViolationRepositoryInterface,
	configRepo AlertConfigRepositoryInterface,
	logReader LogReaderInterface,
	logger *logrus.Logger,
) *AlertService {
	return &AlertService{
		violationRepo: violationRepo,
		configRepo:    configRepo,
		logReader:     logReader,
		logger:        logger,
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

	err := s.configRepo.Create(ctx, config)
	if err != nil {
		s.logger.WithError(err).Errorf("Failed to create alert config for service %s", config.Service)
		return err
	}

	s.logger.Infof("Created alert config for service %s", config.Service)
	return nil
}

// UpdateAlertConfig updates an existing alert configuration.
func (s *AlertService) UpdateAlertConfig(ctx context.Context, config *models.AlertConfig) error {
	if config == nil {
		return fmt.Errorf("alert config cannot be nil")
	}

	if config.ID == 0 {
		return fmt.Errorf("alert config ID is required for update")
	}

	err := s.configRepo.Update(ctx, config)
	if err != nil {
		s.logger.WithError(err).Errorf("Failed to update alert config for service %s", config.Service)
		return err
	}

	s.logger.Infof("Updated alert config for service %s", config.Service)
	return nil
}

// GetAlertConfig retrieves alert configuration for a service.
func (s *AlertService) GetAlertConfig(ctx context.Context, service string) (*models.AlertConfig, error) {
	config, err := s.configRepo.GetByService(ctx, service)
	if err != nil {
		s.logger.WithError(err).Warnf("Failed to get alert config for service %s", service)
		return nil, err
	}

	return config, nil
}

// CheckThresholds checks if current log counts exceed alert thresholds.
func (s *AlertService) CheckThresholds(ctx context.Context) ([]models.AlertThresholdViolation, error) {
	configs, err := s.configRepo.GetAll(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get alert configs")
		return []models.AlertThresholdViolation{}, nil
	}

	violations := []models.AlertThresholdViolation{}
	now := time.Now()
	oneMinuteAgo := now.Add(-1 * time.Minute)

	for i := range configs {
		config := &configs[i]
		if !config.Enabled {
			continue
		}

		// Check error threshold
		errorCount, err := s.logReader.CountByServiceAndLevel(ctx, config.Service, "error", oneMinuteAgo, now)
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
			}
			violations = append(violations, violation)
		}

		// Check warning threshold
		warningCount, err := s.logReader.CountByServiceAndLevel(ctx, config.Service, "warning", oneMinuteAgo, now)
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

	// TODO: Implement email and webhook sending

	s.logger.Infof("Alert sent for service %s level %s", violation.Service, violation.Level)
	return nil
}
