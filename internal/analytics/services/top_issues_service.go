// Package services provides functionality for analytics services, including top issues analysis.
package services

import (
	"context"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/sirupsen/logrus"
)

// TopIssuesService provides methods to retrieve top issues.
type TopIssuesService struct {
	logReader db.LogReaderInterface
	logger    *logrus.Logger
}

// NewTopIssuesService initializes a new TopIssuesService.
//
// Parameters:
// - logReader: The log reader interface for accessing log data.
// - logger: The logger instance for logging operations.
//
// Returns:
// - A pointer to the initialized TopIssuesService.
func NewTopIssuesService(logReader db.LogReaderInterface, logger *logrus.Logger) *TopIssuesService {
	return &TopIssuesService{
		logReader: logReader,
		logger:    logger,
	}
}

// GetTopIssues retrieves the most frequent errors and warnings for a service
func (s *TopIssuesService) GetTopIssues(ctx context.Context, service, level string, start, end time.Time, limit int) ([]models.IssueItem, error) {
	s.logger.WithFields(logrus.Fields{
		"service": service,
		"level":   level,
		"start":   start,
		"end":     end,
		"limit":   limit,
	}).Info("Fetching top issues")
	s.logger.Debugf("Arguments passed to FindTopMessages: service=%s, level=%s, start=%v, end=%v, limit=%d", service, level, start, end, limit)

	s.logger.Debug("Entering GetTopIssues method")
	s.logger.Debugf("Input parameters: service=%s, level=%s, start=%v, end=%v, limit=%d", service, level, start, end, limit)

	// Add log to verify service and level
	s.logger.Debugf("Service: %s, Level: %s", service, level)

	// Add log to verify time range
	s.logger.Debugf("Time Range: Start=%v, End=%v", start, end)

	// Add log to verify limit
	s.logger.Debugf("Limit: %d", limit)

	// Log before calling FindTopMessages
	s.logger.Debug("Calling FindTopMessages")

	// Add detailed logs to trace arguments passed to FindTopMessages
	s.logger.Debug("Tracing arguments passed to FindTopMessages")
	s.logger.Debugf("Service: %s, Level: %s, Start: %v, End: %v, Limit: %d", service, level, start, end, limit)

	// Log the context being passed
	s.logger.Debugf("Context: %+v", ctx)

	issues, err := s.logReader.FindTopMessages(ctx, service, level, start, end, limit)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch top issues")
		return nil, err
	}

	// Add log to verify if services are fetched
	s.logger.Debug("Fetching services")
	services, err := s.logReader.FindAllServices(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch services")
		return nil, err
	}
	s.logger.Debugf("Services fetched: %v", services)

	// Add log to verify if FindTopMessages is called for each service
	for _, service := range services {
		s.logger.Debugf("Calling FindTopMessages for service: %s", service)
	}

	// Add log to verify the type of Count in issues
	for _, issue := range issues {
		s.logger.Debugf("Issue: %+v, Count Type: %T", issue, issue.Count)
	}

	// Add log to verify the type of Count after FindTopMessages
	for _, issue := range issues {
		s.logger.Debugf("Post-FindTopMessages Issue: %+v, Count Type: %T", issue, issue.Count)
	}

	// Add log to verify the type of Count after processing
	for _, issue := range issues {
		s.logger.Debugf("Processed Issue: %+v, Count Type: %T", issue, issue.Count)
	}

	// Add detailed logs to trace Count during processing
	for i, issue := range issues {
		s.logger.Debugf("Processing Issue[%d]: %+v, Count Type: %T", i, issue, issue.Count)
	}

	s.logger.WithField("count", len(issues)).Info("Top issues fetched successfully")
	return issues, nil
}
