// Package logs_services provides service implementations for logs operations.
package logs_services

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// LogAggregationService implements log aggregation operations.
type LogAggregationService struct {
	logReader LogReaderInterface
	logger    *logrus.Logger
}

// NewLogAggregationService creates a new LogAggregationService.
func NewLogAggregationService(logReader LogReaderInterface, logger *logrus.Logger) *LogAggregationService {
	return &LogAggregationService{
		logReader: logReader,
		logger:    logger,
	}
}

// AggregateLogsHourly performs hourly aggregation of logs.
func (s *LogAggregationService) AggregateLogsHourly(ctx context.Context) error {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	return s.aggregateLogs(ctx, oneHourAgo, now, "hourly")
}

// AggregateLogsDaily performs daily aggregation of logs.
func (s *LogAggregationService) AggregateLogsDaily(ctx context.Context) error {
	now := time.Now()
	oneDayAgo := now.Add(-24 * time.Hour)
	return s.aggregateLogs(ctx, oneDayAgo, now, "daily")
}

// aggregateLogs performs log aggregation for a given time range.
func (s *LogAggregationService) aggregateLogs(ctx context.Context, start, end time.Time, period string) error {
	s.logger.Infof("Starting %s aggregation from %v to %v", period, start, end)

	// Get all services
	services, err := s.logReader.FindAllServices(ctx)
	if err != nil {
		s.logger.WithError(err).Errorf("Failed to get services for %s aggregation", period)
		return fmt.Errorf("failed to get services: %w", err)
	}

	for _, service := range services {
		// Count by level
		for _, level := range []string{"error", "warning", "info", "debug"} {
			count, err := s.logReader.CountByServiceAndLevel(ctx, service, level, start, end)
			if err != nil {
				s.logger.WithError(err).Warnf("Failed to count %s logs for service %s", level, service)
				continue
			}

			if count > 0 {
				s.logger.Debugf("Service %s level %s: %d logs (%s)", service, level, count, period)
			}
		}
	}

	s.logger.Infof("Completed %s aggregation", period)
	return nil
}

// GetErrorRate calculates error rate for a service within a time window.
func (s *LogAggregationService) GetErrorRate(ctx context.Context, service string, start, end time.Time) (float64, error) {
	errorCount, err := s.logReader.CountByServiceAndLevel(ctx, service, "error", start, end)
	if err != nil {
		s.logger.WithError(err).Warnf("Failed to count errors for service %s", service)
		return 0, fmt.Errorf("failed to count errors: %w", err)
	}

	totalCount, err := s.logReader.CountByServiceAndLevel(ctx, service, "all", start, end)
	if err != nil {
		s.logger.WithError(err).Warnf("Failed to count total logs for service %s", service)
		return 0, fmt.Errorf("failed to count total: %w", err)
	}

	if totalCount == 0 {
		return 0, nil
	}

	rate := float64(errorCount) / float64(totalCount)
	return rate, nil
}

// CountLogsByServiceAndLevel counts logs by service and level.
func (s *LogAggregationService) CountLogsByServiceAndLevel(ctx context.Context, service, level string, start, end time.Time) (int64, error) {
	count, err := s.logReader.CountByServiceAndLevel(ctx, service, level, start, end)
	if err != nil {
		s.logger.WithError(err).Warnf("Failed to count logs for service %s level %s", service, level)
		return 0, fmt.Errorf("failed to count logs: %w", err)
	}

	return count, nil
}
