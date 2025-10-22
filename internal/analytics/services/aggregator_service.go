// Package services provides the implementation of analytics services.
package services

import (
	"context"
	"log"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/sirupsen/logrus"
)

type AggregatorService struct {
	aggregationRepo analytics.AggregationRepositoryInterface
	logReader       analytics.LogReaderInterface
	logger          *logrus.Logger
}

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
	for _, level := range levels {
		s.logger.WithFields(logrus.Fields{
			"service": service,
			"level":   level,
		}).Debug("Counting logs for level")
		log.Printf("CountByServiceAndLevel called with service=%s, level=%s, start=%v, end=%v", service, level, start, end)
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

		agg := &models.Aggregation{
			MetricType: models.MetricType("log_count"),
			Service:    service,
			Value:      float64(count),
			TimeBucket: start,
		}

		s.logger.WithFields(logrus.Fields{
			"aggregation": agg,
		}).Debug("Upserting aggregation")
		log.Printf("Upsert called with aggregation=%+v", agg)
		if err := s.aggregationRepo.Upsert(ctx, agg); err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"aggregation": agg,
			}).Error("Failed to upsert aggregation")
			return err
		}
	}

	return nil
}

// Define FindAllServices as part of logReader
func (s *AggregatorService) FindAllServices(ctx context.Context) ([]string, error) {
	return s.logReader.FindAllServices(ctx)
}

// Define Upsert as part of aggregationRepo
func (s *AggregatorService) Upsert(ctx context.Context, service string, level string, count int, timestamp time.Time) error {
	aggregation := &models.Aggregation{
		MetricType: models.MetricType("log_count"),
		Service:    service,
		Value:      float64(count),
		TimeBucket: timestamp,
	}
	return s.aggregationRepo.Upsert(ctx, aggregation)
}
