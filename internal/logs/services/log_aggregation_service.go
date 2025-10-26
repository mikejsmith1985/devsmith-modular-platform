// Package services provides service implementations for logs operations.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// LogAggregationService performs log aggregation jobs.
type LogAggregationService struct {
	reader LogReaderInterface
	logger *logrus.Logger
}

// NewLogAggregationService creates a new LogAggregationService.
func NewLogAggregationService(reader LogReaderInterface, logger *logrus.Logger) *LogAggregationService {
	return &LogAggregationService{
		reader: reader,
		logger: logger,
	}
}

// AggregateLogsHourly performs hourly aggregation of logs.
func (s *LogAggregationService) AggregateLogsHourly(ctx context.Context) error {
	now := time.Now()
	hourAgo := now.Add(-1 * time.Hour)

	// Get all services
	services, err := s.reader.FindAllServices(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get services for hourly aggregation")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"services": len(services),
		"hour":     now.Hour(),
	}).Info("Starting hourly aggregation")

	// Aggregate for each service and level
	for _, service := range services {
		levels := []string{"error", "warning", "info", "debug"}
		for _, level := range levels {
			count, err := s.reader.CountByServiceAndLevel(ctx, service, level, hourAgo, now)
			if err != nil {
				s.logger.WithError(err).Warnf("Failed to count logs for %s/%s", service, level)
				continue
			}

			s.logger.WithFields(logrus.Fields{
				"service": service,
				"level":   level,
				"count":   count,
			}).Debug("Aggregated hourly logs")
		}
	}

	s.logger.Info("Hourly aggregation completed")
	return nil
}

// AggregateLogsDaily performs daily aggregation of logs.
func (s *LogAggregationService) AggregateLogsDaily(ctx context.Context) error {
	now := time.Now()
	dayAgo := now.Add(-24 * time.Hour)

	// Get all services
	services, err := s.reader.FindAllServices(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get services for daily aggregation")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"services": len(services),
		"date":     now.Format("2006-01-02"),
	}).Info("Starting daily aggregation")

	// Aggregate for each service and level
	for _, service := range services {
		levels := []string{"error", "warning", "info", "debug"}
		for _, level := range levels {
			count, err := s.reader.CountByServiceAndLevel(ctx, service, level, dayAgo, now)
			if err != nil {
				s.logger.WithError(err).Warnf("Failed to count logs for %s/%s", service, level)
				continue
			}

			s.logger.WithFields(logrus.Fields{
				"service": service,
				"level":   level,
				"count":   count,
			}).Debug("Aggregated daily logs")
		}
	}

	s.logger.Info("Daily aggregation completed")
	return nil
}

// GetErrorRate calculates error rate for a service within a time window.
func (s *LogAggregationService) GetErrorRate(ctx context.Context, service string, start, end time.Time) (float64, error) {
	if start.After(end) {
		return 0, fmt.Errorf("start time must be before end time")
	}

	// Count errors
	errorCount, err := s.reader.CountByServiceAndLevel(ctx, service, "error", start, end)
	if err != nil {
		s.logger.WithError(err).Warnf("Failed to count errors for %s", service)
		return 0, err
	}

	// Count total logs
	levels := []string{"error", "warning", "info", "debug"}
	totalCount := int64(0)

	for _, level := range levels {
		count, err := s.reader.CountByServiceAndLevel(ctx, service, level, start, end)
		if err != nil {
			s.logger.WithError(err).Warnf("Failed to count %s logs for %s", level, service)
			continue
		}
		totalCount += count
	}

	if totalCount == 0 {
		return 0, nil
	}

	rate := float64(errorCount) / float64(totalCount)
	s.logger.WithFields(logrus.Fields{
		"service":     service,
		"error_rate":  rate,
		"error_count": errorCount,
		"total_count": totalCount,
	}).Debug("Calculated error rate")

	return rate, nil
}

// CountLogsByServiceAndLevel counts logs by service and level.
func (s *LogAggregationService) CountLogsByServiceAndLevel(ctx context.Context, service, level string, start, end time.Time) (int64, error) {
	if start.After(end) {
		return 0, fmt.Errorf("start time must be before end time")
	}

	count, err := s.reader.CountByServiceAndLevel(ctx, service, level, start, end)
	if err != nil {
		s.logger.WithError(err).Warnf("Failed to count logs for %s/%s", service, level)
		return 0, err
	}

	return count, nil
}
