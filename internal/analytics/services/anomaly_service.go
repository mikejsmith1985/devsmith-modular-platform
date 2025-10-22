// Package services provides the implementation of analytics services.
package services

import (
	"context"
	"math"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/sirupsen/logrus"
)

type AnomalyService struct {
	aggregationRepo db.AggregationRepositoryInterface
	logger          *logrus.Logger
}

func NewAnomalyService(aggregationRepo db.AggregationRepositoryInterface, logger *logrus.Logger) *AnomalyService {
	return &AnomalyService{
		aggregationRepo: aggregationRepo,
		logger:          logger,
	}
}

// DetectAnomalies identifies unusual spikes or dips in the data
func (s *AnomalyService) DetectAnomalies(ctx context.Context, metricType models.MetricType, service string, start, end time.Time) ([]*models.Anomaly, error) {
	s.logger.WithFields(logrus.Fields{
		"metricType": metricType,
		"service":    service,
		"start":      start,
		"end":        end,
	}).Info("Starting anomaly detection")

	aggregations, err := s.aggregationRepo.FindByRange(ctx, metricType, service, start, end)
	if err != nil {
		s.logger.WithError(err).Error("Failed to retrieve aggregations")
		return nil, err
	}

	if len(aggregations) < 3 {
		s.logger.Warn("Not enough data points for anomaly detection")
		return nil, nil
	}

	if len(aggregations) == 0 {
		s.logger.Warn("No aggregations found, skipping anomaly detection")
		return nil, nil
	}

	var anomalies []*models.Anomaly
	s.logger.Debug("Calculating mean and standard deviation")
	mean, stddev := calculateStats(aggregations)
	s.logger.WithFields(logrus.Fields{
		"mean":   mean,
		"stddev": stddev,
	}).Debug("Stats calculated")

	s.logger.Infof("Mean: %f, StdDev: %f", mean, stddev)
	for _, agg := range aggregations {
		zScore := (agg.Value - mean) / stddev
		s.logger.Infof("Value: %f, Z-Score: %f", agg.Value, zScore)
	}

	for _, agg := range aggregations {
		zScore := (agg.Value - mean) / stddev
		s.logger.WithFields(logrus.Fields{
			"value":     agg.Value,
			"mean":      mean,
			"stddev":    stddev,
			"zScore":    zScore,
			"isAnomaly": math.Abs(zScore) > 1.5,
		}).Info("Evaluating aggregation for anomaly")

		if math.Abs(zScore) > 1.5 { // Revert Z-score threshold to original value
			anomaly := &models.Anomaly{
				MetricType: metricType,
				Service:    service,
				TimeBucket: agg.TimeBucket,
				Value:      agg.Value,
				ZScore:     zScore,
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	s.logger.WithField("count", len(anomalies)).Info("Anomaly detection completed")
	return anomalies, nil
}

func calculateStats(aggregations []*models.Aggregation) (mean, stddev float64) {
	if len(aggregations) == 0 {
		return 0, 0 // Avoid division by zero
	}

	var sum, sumSquares float64
	count := float64(len(aggregations))

	for _, agg := range aggregations {
		sum += agg.Value
		sumSquares += agg.Value * agg.Value
	}

	mean = sum / count
	variance := (sumSquares / count) - (mean * mean)
	stddev = math.Sqrt(variance)
	return
}
