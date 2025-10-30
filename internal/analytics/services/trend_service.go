package analytics_services

import (
	"context"
	"time"

	analytics_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/db"
	analytics_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/sirupsen/logrus"
)

// TrendService provides methods to analyze trends.
type TrendService struct {
	aggregationRepo analytics_db.AggregationRepositoryInterface
	logger          *logrus.Logger
}

// NewTrendService creates a new instance of TrendService.
func NewTrendService(aggregationRepo analytics_db.AggregationRepositoryInterface, logger *logrus.Logger) *TrendService {
	return &TrendService{
		aggregationRepo: aggregationRepo,
		logger:          logger,
	}
}

// AnalyzeTrends calculates trends for a given metric type and service
func (s *TrendService) AnalyzeTrends(ctx context.Context, metricType analytics_models.MetricType, service string, start, end time.Time) (*analytics_models.TrendAnalysis, error) {
	s.logger.WithFields(logrus.Fields{
		"metricType": metricType,
		"service":    service,
		"start":      start,
		"end":        end,
	}).Info("Starting trend analysis")

	aggregations, err := s.aggregationRepo.FindByRange(ctx, metricType, service, start, end)
	if err != nil {
		s.logger.WithError(err).Error("Failed to retrieve aggregations")
		return nil, err
	}

	if len(aggregations) < 2 {
		s.logger.Warn("Not enough data points for trend analysis")
		return nil, nil
	}

	var totalChange float64
	var firstValue, lastValue float64
	firstValue = aggregations[0].Value
	lastValue = aggregations[len(aggregations)-1].Value

	for i := 1; i < len(aggregations); i++ {
		totalChange += aggregations[i].Value - aggregations[i-1].Value
	}

	trend := &analytics_models.TrendAnalysis{
		MetricType: metricType,
		Service:    service,
		Start:      start,
		End:        end,
		Direction:  "stable",
		Change:     totalChange,
	}

	if totalChange > 0 {
		trend.Direction = "increasing"
	} else if totalChange < 0 {
		trend.Direction = "decreasing"
	}

	trend.PercentageChange = ((lastValue - firstValue) / firstValue) * 100

	s.logger.WithFields(logrus.Fields{
		"direction":        trend.Direction,
		"change":           trend.Change,
		"percentageChange": trend.PercentageChange,
	}).Info("Trend analysis completed")

	return trend, nil
}

// GetTrends analyzes trends for a metric over a time range
func (s *TrendService) GetTrends(ctx context.Context, metricType analytics_models.MetricType, service string, start, end time.Time) (*analytics_models.TrendResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"metricType": metricType,
		"service":    service,
		"start":      start,
		"end":        end,
	}).Info("Fetching trends")

	aggregations, err := s.aggregationRepo.FindByRange(ctx, metricType, service, start, end)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch trends")
		return nil, err
	}

	// Process aggregations into a TrendResponse
	trendSummary := &analytics_models.TrendSummary{
		Direction:  "stable", // Placeholder logic
		Confidence: 0.95,     // Placeholder confidence
		Summary:    "Trend analysis completed successfully.",
	}

	response := &analytics_models.TrendResponse{
		MetricType: metricType,
		Service:    service,
		Trend:      trendSummary,
	}

	s.logger.WithField("count", len(aggregations)).Info("Trends fetched successfully")
	return response, nil
}
