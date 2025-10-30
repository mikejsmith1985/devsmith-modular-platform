// Package analytics_services provides the implementation of analytics services.
package analytics_services

import (
	"context"
	"log"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics"
	analytics_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/sirupsen/logrus"
)

// AggregatorService provides methods to aggregate data.
type AggregatorService struct {
	aggregationRepo analytics.AggregationRepositoryInterface
	logReader       analytics.LogReaderInterface
	logger          *logrus.Logger
}

// NewAggregatorService creates a new instance of AggregatorService.
func NewAggregatorService(aggregationRepo analytics.AggregationRepositoryInterface, logReader analytics.LogReaderInterface, logger *logrus.Logger) *AggregatorService {
	return &AggregatorService{
		aggregationRepo: aggregationRepo,
		logReader:       logReader,
		logger:          logger,
	}
}

// RunHourlyAggregation performs hourly aggregation for all services
func (s *AggregatorService) RunHourlyAggregation(ctx context.Context) error {
	s.logger.Info("Starting hourly aggregation job")
	log.Println("AggregatorService.RunHourlyAggregation started")
	defer log.Println("AggregatorService.RunHourlyAggregation completed")

	s.logger.Debug("Entering RunHourlyAggregation")
	s.logger.Debug("Calling FindAllServices")
	services, err := s.logReader.FindAllServices(ctx)
	if err != nil {
		s.logger.WithError(err).Error("FindAllServices failed")
		return err
	}

	s.logger.WithField("services", services).Debug("Services retrieved")

	var aggErr error
	for _, service := range services {
		s.logger.WithField("service", service).Debug("Processing service")
		if err := s.aggregateService(ctx, service); err != nil {
			s.logger.WithError(err).WithField("service", service).Error("aggregateService failed")
			aggErr = err // Capture the error but continue processing other services
		}
	}

	s.logger.Debug("Exiting RunHourlyAggregation")
	s.logger.Info("Hourly aggregation job completed")
	return aggErr // Return the last error encountered, if any
}

func (s *AggregatorService) aggregateService(ctx context.Context, service string) error {
	log.Printf("AggregatorService.aggregateService called with service=%s", service)

	levels := []string{"info", "warn", "error"}
	end := time.Now().Truncate(time.Hour)
	start := end.Add(-1 * time.Hour)

	s.logger.WithField("service", service).Debug("Starting aggregation for service")
	// Add detailed debug logs to capture the flow and arguments passed to CountByServiceAndLevel and Upsert
	// Add detailed logs to capture levels being processed
	s.logger.WithFields(logrus.Fields{
		"service": service,
		"levels":  levels,
	}).Debug("Processing levels for service")
	for _, level := range levels {
		s.logger.WithFields(logrus.Fields{
			"service": service,
			"level":   level,
		}).Debug("Processing log level")

		count, err := s.logReader.CountByServiceAndLevel(ctx, service, level, start, end)
		if err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"service": service,
				"level":   level,
			}).Error("Failed to count logs for level")
			return err
		}

		s.logger.WithFields(logrus.Fields{
			"service": service,
			"level":   level,
			"count":   count,
		}).Debug("Log count retrieved")

		agg := &analytics_models.Aggregation{
			MetricType: analytics_models.MetricType("log_count"),
			Service:    service,
			Value:      float64(count),
			TimeBucket: start,
		}

		s.logger.WithFields(logrus.Fields{
			"aggregation": agg,
		}).Debug("Upserting aggregation")

		if err := s.aggregationRepo.Upsert(ctx, agg); err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"aggregation": agg,
			}).Error("Failed to upsert aggregation")
			return err
		}
	}

	return nil
}

// FindAllServices retrieves all services for aggregation.
// It ensures that the returned list is non-nil.
// Returns an error if the retrieval fails.
//
// Parameters:
// - ctx: The context for the operation.
//
// Returns:
// - A slice of service names.
// - An error if the operation fails.
func (s *AggregatorService) FindAllServices(ctx context.Context) ([]string, error) {
	return s.logReader.FindAllServices(ctx)
}

// Upsert inserts or updates an aggregation record.
//
// Parameters:
// - ctx: The context for the operation.
// - service: The name of the service.
// - level: The aggregation level.
//
// Returns:
// - An error if the operation fails.
func (s *AggregatorService) Upsert(ctx context.Context, service, level string, count int, timestamp time.Time) error {
	aggregation := &analytics_models.Aggregation{
		MetricType: analytics_models.MetricType("log_count"),
		Service:    service,
		Value:      float64(count),
		TimeBucket: timestamp,
	}
	return s.aggregationRepo.Upsert(ctx, aggregation)
}
