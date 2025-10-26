// Package services provides service implementations for logs operations.
package services

import (
	"context"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/sirupsen/logrus"
)

// DashboardService implements dashboard statistics retrieval.
type DashboardService struct {
	logReader LogReaderInterface
	logger    *logrus.Logger
}

// LogReaderInterface defines the contract for reading logs.
type LogReaderInterface interface {
	FindAllServices(ctx context.Context) ([]string, error)
	CountByServiceAndLevel(ctx context.Context, service, level string, start, end time.Time) (int64, error)
	FindTopMessages(ctx context.Context, service, level string, start, end time.Time, limit int) ([]models.LogMessage, error)
}

// NewDashboardService creates a new DashboardService.
func NewDashboardService(logReader LogReaderInterface, logger *logrus.Logger) *DashboardService {
	return &DashboardService{
		logReader: logReader,
		logger:    logger,
	}
}

// GetDashboardStats returns aggregated statistics for the dashboard.
func (s *DashboardService) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
	now := time.Now()
	stats := &models.DashboardStats{
		GeneratedAt:      now,
		ServiceStats:     make(map[string]*models.LogStats),
		ServiceHealth:    make(map[string]*models.ServiceHealth),
		TopErrors:        []models.TopErrorMessage{},
		Violations:       []models.AlertThresholdViolation{},
		TimestampOne:     now.Add(-1 * time.Hour),
		TimestampOneDay:  now.Add(-24 * time.Hour),
		TimestampOneWeek: now.Add(-7 * 24 * time.Hour),
	}

	// Get all services
	services, err := s.logReader.FindAllServices(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get all services")
		return stats, nil // Return partial stats
	}

	// Aggregate stats for each service
	oneHourAgo := now.Add(-1 * time.Hour)
	for _, service := range services {
		// Get service stats
		svcStats, svcErr := s.GetServiceStats(ctx, service, time.Hour)
		if svcErr != nil {
			s.logger.WithError(svcErr).Warnf("Failed to get stats for service %s", service)
			continue
		}
		if svcStats != nil {
			stats.ServiceStats[service] = svcStats
		}

		// Get service health
		health := s.calculateServiceHealth(ctx, service, oneHourAgo, now)
		stats.ServiceHealth[service] = health
	}

	// Get top errors
	topErrors, err := s.GetTopErrors(ctx, 10, time.Hour)
	if err == nil && topErrors != nil {
		stats.TopErrors = topErrors
	}

	return stats, nil
}

// GetServiceStats returns statistics for a specific service.
func (s *DashboardService) GetServiceStats(ctx context.Context, service string, timeRange time.Duration) (*models.LogStats, error) {
	now := time.Now()
	start := now.Add(-timeRange)

	stats := &models.LogStats{
		Timestamp:    now,
		Service:      service,
		CountByLevel: make(map[string]int64),
	}

	// Count logs by level
	levels := []string{"error", "warning", "info", "debug"}
	totalCount := int64(0)
	errorCount := int64(0)

	for _, level := range levels {
		count, err := s.logReader.CountByServiceAndLevel(ctx, service, level, start, now)
		if err != nil {
			s.logger.WithError(err).Warnf("Failed to count logs for service %s level %s", service, level)
			continue
		}
		stats.CountByLevel[level] = count
		totalCount += count

		if level == "error" {
			errorCount = count
		}
	}

	stats.TotalCount = totalCount
	if totalCount > 0 {
		stats.ErrorRate = float64(errorCount) / float64(totalCount)
	}

	return stats, nil
}

// GetTopErrors returns the top error messages within a time range.
func (s *DashboardService) GetTopErrors(ctx context.Context, limit int, timeRange time.Duration) ([]models.TopErrorMessage, error) {
	now := time.Now()
	start := now.Add(-timeRange)

	// Get all services
	services, err := s.logReader.FindAllServices(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get services for top errors")
		return []models.TopErrorMessage{}, nil
	}

	allErrors := make([]models.TopErrorMessage, 0)

	// Collect errors from all services
	for _, service := range services {
		issues, err := s.logReader.FindTopMessages(ctx, service, "error", start, now, limit*2)
		if err != nil {
			s.logger.WithError(err).Warnf("Failed to get top messages for service %s", service)
			continue
		}

		for _, issue := range issues {
			allErrors = append(allErrors, models.TopErrorMessage{
				Message:   issue.Message,
				Service:   issue.Service,
				Level:     issue.Level,
				Count:     int64(issue.Count),
				FirstSeen: issue.LastSeen.Add(-1 * time.Hour), // Approximate
				LastSeen:  issue.LastSeen,
			})
		}
	}

	// Sort by count and limit results
	if len(allErrors) > limit {
		allErrors = allErrors[:limit]
	}

	return allErrors, nil
}

// GetServiceHealth returns health status for all services.
func (s *DashboardService) GetServiceHealth(ctx context.Context) (map[string]*models.ServiceHealth, error) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)

	health := make(map[string]*models.ServiceHealth)

	// Get all services
	services, err := s.logReader.FindAllServices(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get services for health check")
		return health, nil
	}

	// Calculate health for each service
	for _, service := range services {
		serviceHealth := s.calculateServiceHealth(ctx, service, oneHourAgo, now)
		health[service] = serviceHealth
	}

	return health, nil
}

// calculateServiceHealth determines the health status of a service.
func (s *DashboardService) calculateServiceHealth(ctx context.Context, service string, start, end time.Time) *models.ServiceHealth {
	health := &models.ServiceHealth{
		Service:       service,
		Status:        "OK",
		LastCheckedAt: time.Now(),
	}

	// Count errors and warnings
	errorCount, err := s.logReader.CountByServiceAndLevel(ctx, service, "error", start, end)
	if err != nil {
		s.logger.WithError(err).Warnf("Failed to count errors for service %s", service)
		errorCount = 0
	}

	warningCount, err := s.logReader.CountByServiceAndLevel(ctx, service, "warning", start, end)
	if err != nil {
		s.logger.WithError(err).Warnf("Failed to count warnings for service %s", service)
		warningCount = 0
	}

	infoCount, err := s.logReader.CountByServiceAndLevel(ctx, service, "info", start, end)
	if err != nil {
		s.logger.WithError(err).Warnf("Failed to count info for service %s", service)
		infoCount = 0
	}

	health.ErrorCount = errorCount
	health.WarningCount = warningCount
	health.InfoCount = infoCount

	// Determine status based on error count
	switch {
	case errorCount > 50:
		health.Status = "Error"
	case errorCount > 10 || warningCount > 50:
		health.Status = "Warning"
	default:
		health.Status = "OK"
	}

	return health
}
