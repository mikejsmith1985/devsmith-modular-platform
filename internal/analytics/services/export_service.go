// Package services provides the implementation of analytics services.
package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"path/filepath"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/sirupsen/logrus"
)

// ExportService provides methods for exporting analytics data to various formats.
type ExportService struct {
	aggregationRepo db.AggregationRepositoryInterface
	logger          *logrus.Logger
}

// NewExportService creates a new instance of ExportService with the provided dependencies.
func NewExportService(aggregationRepo db.AggregationRepositoryInterface, logger *logrus.Logger) *ExportService {
	return &ExportService{
		aggregationRepo: aggregationRepo,
		logger:          logger,
	}
}

// ExportToCSV exports aggregation data to a CSV file. It retrieves data from the repository
// and writes it to the specified file path in CSV format.
// func (s *ExportService) ExportToCSV(ctx context.Context, metricType models.MetricType, service string, filePath string) error {
func (s *ExportService) ExportToCSV(ctx context.Context, metricType models.MetricType, service, filePath string) error {
	s.logger.WithFields(logrus.Fields{
		"metricType": metricType,
		"service":    service,
		"filePath":   filePath,
	}).Info("Exporting data to CSV")

	aggregations, err := s.aggregationRepo.FindByRange(ctx, metricType, service, models.MinTime, models.MaxTime)
	if err != nil {
		s.logger.WithError(err).Error("Failed to retrieve aggregations")
		return err
	}

	// Add validation for file paths to prevent potential file inclusion vulnerabilities.
	if !isValidFilePath(filePath) || strings.Contains(filePath, "..") {
		s.logger.WithField("filePath", filePath).Error("Invalid or unsafe file path")
		return fmt.Errorf("invalid or unsafe file path: %s", filePath)
	}

	// Ensure the directory exists before creating the file.
	dir := filepath.Dir(filePath)
	// Resolve variable shadowing by renaming inner variables.
	if errCreateDir := os.MkdirAll(dir, 0o700); errCreateDir != nil {
		s.logger.WithField("dir", dir).Error("Failed to create directory")
		return fmt.Errorf("failed to create directory: %s", dir)
	}

	// Use a secure method to create the file.
	file, err := os.OpenFile(filepath.Clean(filePath), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		s.logger.WithField("filePath", filePath).Error("Failed to open file")
		return fmt.Errorf("failed to open file: %s", filePath)
	}
	defer func() {
		if err := file.Close(); err != nil {
			s.logger.WithError(err).Error("Failed to close file")
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"MetricType", "Service", "Value", "TimeBucket", "CreatedAt"}); err != nil {
		return err
	}

	// Write rows
	for _, agg := range aggregations {
		row := []string{
			string(agg.MetricType), // Convert MetricType to string
			agg.Service,
			strconv.FormatFloat(agg.Value, 'f', -1, 64), // Use strconv for better precision
			agg.TimeBucket.Format(time.RFC3339),
			agg.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	s.logger.Info("Data exported to CSV successfully")
	return nil
}

// ExportToJSON writes the provided data to a JSON file at the specified file path.
// ExportToJSON exports aggregations to a JSON file
func (s *ExportService) ExportToJSON(ctx context.Context, metricType models.MetricType, serviceAndPath ...string) error {
	service, filePath := serviceAndPath[0], serviceAndPath[1]
	s.logger.WithFields(logrus.Fields{
		"metricType": metricType,
		"service":    service,
		"filePath":   filePath,
	}).Info("Exporting data to JSON")

	aggregations, err := s.aggregationRepo.FindByRange(ctx, metricType, service, models.MinTime, models.MaxTime)
	if err != nil {
		s.logger.WithError(err).Error("Failed to retrieve aggregations")
		return err
	}

	// Add validation for file paths to prevent potential file inclusion vulnerabilities.
	if !isValidFilePath(filePath) || strings.Contains(filePath, "..") {
		s.logger.WithField("filePath", filePath).Error("Invalid or unsafe file path")
		return fmt.Errorf("invalid or unsafe file path: %s", filePath)
	}

	// Ensure the directory exists before creating the file.
	dir := filepath.Dir(filePath)
	// Resolve variable shadowing by renaming inner variables.
	if errCreateDir := os.MkdirAll(dir, 0o700); errCreateDir != nil {
		s.logger.WithField("dir", dir).Error("Failed to create directory")
		return fmt.Errorf("failed to create directory: %s", dir)
	}

	// Use a secure method to create the file.
	file, err := os.OpenFile(filepath.Clean(filePath), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		s.logger.WithField("filePath", filePath).Error("Failed to open file")
		return fmt.Errorf("failed to open file: %s", filePath)
	}
	defer func() {
		if err := file.Close(); err != nil {
			s.logger.WithError(err).Error("Failed to close file")
		}
	}()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(aggregations); err != nil {
		s.logger.WithError(err).Error("Failed to write JSON data")
		return err
	}

	s.logger.Info("Data exported to JSON successfully")
	return nil
}

// ExportData exports aggregations to a file (CSV or JSON) based on the file extension.
func (s *ExportService) ExportData(ctx context.Context, metricType models.MetricType, service, filePath string) error {
	s.logger.WithFields(logrus.Fields{
		"metricType": metricType,
		"service":    service,
		"filePath":   filePath,
	}).Info("Exporting data")

	if len(filePath) > 4 && filePath[len(filePath)-4:] == ".csv" {
		return s.ExportToCSV(ctx, metricType, service, filePath)
	} else if len(filePath) > 5 && filePath[len(filePath)-5:] == ".json" {
		return s.ExportToJSON(ctx, metricType, service, filePath)
	}

	s.logger.Error("Unsupported file extension")
	return fmt.Errorf("unsupported file extension: %s", filePath)
}

// isValidFilePath validates the file path to prevent potential file inclusion vulnerabilities.
func isValidFilePath(filePath string) bool {
	// Example validation: Ensure the file path is within a specific directory
	allowedDir := "/safe/export/directory"
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}
	return strings.HasPrefix(absPath, allowedDir)
}
