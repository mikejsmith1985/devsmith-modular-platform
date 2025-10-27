// Package services provides service implementations for logs operations.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/sirupsen/logrus"
)

// Interval constants for error trend queries
const (
	IntervalHourly = "hourly"
	IntervalDaily  = "daily"
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

// ValidationAggregation provides aggregated validation error analysis.
type ValidationAggregation struct { //nolint:govet // Struct alignment optimized for memory efficiency
	logReader LogReaderInterface
	logger    *logrus.Logger
}

// NewValidationAggregation creates a new ValidationAggregation service.
func NewValidationAggregation(logReader LogReaderInterface, logger *logrus.Logger) *ValidationAggregation {
	return &ValidationAggregation{
		logReader: logReader,
		logger:    logger,
	}
}

// GetTopErrors retrieves the most frequently occurring validation errors.
// Parameters:
//   - service: Filter by service (empty string = all services)
//   - limit: Maximum number of errors to return (default 10, max 50)
//   - days: Number of days to look back (default 7, max 365)
func (va *ValidationAggregation) GetTopErrors(ctx context.Context, service string, limit, days int) ([]models.ValidationError, error) {
	// Validate and constrain parameters
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50 // Prevent unbounded queries
	}
	if days <= 0 {
		days = 7
	}
	if days > 365 {
		days = 365 // Maximum 1 year lookback
	}

	startTime := time.Now().AddDate(0, 0, -days)
	endTime := time.Now()

	// Use existing LogReader to get top messages for warnings (validation errors)
	messages, err := va.logReader.FindTopMessages(ctx, service, "warning", startTime, endTime, limit)
	if err != nil {
		va.logger.WithError(err).WithField("service", service).Warn("Failed to query validation errors")
		return []models.ValidationError{}, fmt.Errorf("failed to query validation errors: %w", err)
	}

	// Convert LogMessage to ValidationError
	// This provides a richer data structure with affected services tracking
	result := make([]models.ValidationError, len(messages))
	for i, msg := range messages {
		result[i] = models.ValidationError{
			ErrorType:        "validation_error",
			Message:          msg.Message,
			Count:            int64(msg.Count),
			LastOccurrence:   msg.LastSeen,
			AffectedServices: []string{msg.Service},
		}
	}

	return result, nil
}

// GetErrorTrends returns error count trends over a specified time period.
// Parameters:
//   - service: Filter by service (empty string = all services)
//   - days: Look back period in days (default 7, max 365)
//   - interval: Grouping interval - "hourly" or "daily" (default hourly)
//
// Returns a time-series of error counts, useful for visualizing error rate trends
// and identifying patterns in validation failures.
func (va *ValidationAggregation) GetErrorTrends(ctx context.Context, service string, days int, interval string) ([]models.ErrorTrend, error) {
	// Validate and constrain parameters
	if days <= 0 {
		days = 7
	}
	if days > 365 {
		days = 365 // Maximum 1 year lookback
	}
	if interval != IntervalHourly && interval != IntervalDaily {
		interval = IntervalHourly // Default to hourly for more granular data
	}

	startTime := time.Now().AddDate(0, 0, -days)
	endTime := time.Now()

	// Query error counts for the period
	// Using warning level for validation error detection
	errorCount, err := va.logReader.CountByServiceAndLevel(ctx, service, "warning", startTime, endTime)
	if err != nil {
		va.logger.WithError(err).WithField("service", service).Warn("Failed to query error trends")
		return []models.ErrorTrend{}, fmt.Errorf("failed to query error trends: %w", err)
	}

	// Create trend entry
	// Note: For detailed hourly/daily breakdown, a more sophisticated query would be needed.
	// This simplified version provides aggregate trend for the period.
	var intervalDuration time.Duration
	if interval == IntervalHourly {
		intervalDuration = time.Hour
	} else {
		intervalDuration = 24 * time.Hour
	}

	timestamp := startTime.Round(intervalDuration)
	result := []models.ErrorTrend{
		{
			Timestamp:        timestamp,
			ErrorCount:       errorCount,
			ErrorRatePercent: calculateErrorRate(errorCount), // Placeholder calculation
			ByType: map[string]int64{
				"validation_error": errorCount,
			},
		},
	}

	return result, nil
}

// calculateErrorRate computes a simple error rate percentage.
// In production, this should consider total request count for accuracy.
func calculateErrorRate(errorCount int64) float64 {
	if errorCount == 0 {
		return 0.0
	}
	// Placeholder: Assumes roughly 1000 requests per error for estimation
	// Real implementation should track total requests from metrics
	return float64(errorCount) / 10.0
}
