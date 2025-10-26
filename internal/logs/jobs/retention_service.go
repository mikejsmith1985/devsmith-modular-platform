// Package jobs provides retention and archival service implementation.
package jobs

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// NewRetentionService creates a fully initialized retention service.
// It validates the configuration and ensures all dependencies are provided.
// Returns error if config is invalid or dependencies are nil.
func NewRetentionService(cfg *RetentionConfig, repo LogRepository, storage ArchiveStorage) (*RetentionService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("retention config is required")
	}

	if repo == nil {
		return nil, fmt.Errorf("log repository is required")
	}

	if storage == nil {
		return nil, fmt.Errorf("archive storage is required")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid retention config: %w", err)
	}

	return &RetentionService{
		config:  cfg,
		repo:    repo,
		storage: storage,
	}, nil
}

// LoadRetentionConfig loads retention configuration from environment variables.
// It reads LOG_RETENTION_DAYS, LOG_ARCHIVE_ENABLED, LOG_ARCHIVE_COMPRESSION, LOG_ARCHIVE_STORAGE_TYPE,
// LOG_ARCHIVE_LOCAL_PATH, LOG_ARCHIVE_S3_BUCKET, and LOG_ARCHIVE_S3_REGION.
// Uses sensible defaults if environment variables are not set.
func LoadRetentionConfig() (RetentionConfig, error) {
	cfg := RetentionConfig{
		RetentionDays:             defaultRetentionDays,
		ArchiveEnabled:            false,
		ArchiveCompressionEnabled: true,
		StorageType:               storageTypeLocal,
		LocalArchivePath:          "./logs-archive",
	}

	// Load from environment
	if days := os.Getenv("LOG_RETENTION_DAYS"); days != "" {
		d, err := strconv.Atoi(days)
		if err != nil {
			return cfg, fmt.Errorf("invalid LOG_RETENTION_DAYS: %w", err)
		}
		cfg.RetentionDays = d
	}

	if archive := os.Getenv("LOG_ARCHIVE_ENABLED"); archive != "" {
		cfg.ArchiveEnabled = archive == "true" || archive == "1"
	}

	if compress := os.Getenv("LOG_ARCHIVE_COMPRESSION"); compress != "" {
		cfg.ArchiveCompressionEnabled = compress == "true" || compress == "1"
	}

	if storageType := os.Getenv("LOG_ARCHIVE_STORAGE_TYPE"); storageType != "" {
		cfg.StorageType = storageType
	}

	if localPath := os.Getenv("LOG_ARCHIVE_LOCAL_PATH"); localPath != "" {
		cfg.LocalArchivePath = localPath
	}

	if bucket := os.Getenv("LOG_ARCHIVE_S3_BUCKET"); bucket != "" {
		cfg.S3Bucket = bucket
	}

	if region := os.Getenv("LOG_ARCHIVE_S3_REGION"); region != "" {
		cfg.S3Region = region
	}

	return cfg, nil
}

// Validate checks if retention config is valid.
// It ensures RetentionDays is positive and storage configuration is consistent.
// Sets default values for StorageType and LocalArchivePath if ArchiveEnabled is true.
func (r *RetentionConfig) Validate() error {
	if r.RetentionDays <= 0 {
		return fmt.Errorf("RetentionDays must be positive, got %d", r.RetentionDays)
	}

	if !r.ArchiveEnabled {
		return nil
	}

	// Use defaults if not specified
	if r.StorageType == "" {
		r.StorageType = storageTypeLocal
	}

	if r.StorageType != storageTypeLocal && r.StorageType != "s3" {
		return fmt.Errorf("StorageType must be '%s' or 's3', got '%s'", storageTypeLocal, r.StorageType)
	}

	if r.StorageType == storageTypeLocal && r.LocalArchivePath == "" {
		r.LocalArchivePath = "./logs-archive"
	}

	if r.StorageType == "s3" && r.S3Bucket == "" {
		return fmt.Errorf("S3Bucket is required when using S3 storage")
	}

	return nil
}

// CleanupOldLogs deletes logs older than the retention period, optionally archiving first.
// If archiving is enabled, it fetches and archives old logs before deletion.
// Returns the number of log entries deleted.
func (rs *RetentionService) CleanupOldLogs(ctx context.Context) (int64, error) {
	if rs == nil {
		return 0, fmt.Errorf("retention service not initialized")
	}

	if rs.config == nil {
		return 0, fmt.Errorf("retention config not initialized")
	}

	if rs.repo == nil {
		return 0, fmt.Errorf("repository not initialized")
	}

	before := time.Now().AddDate(0, 0, -rs.config.RetentionDays)

	// Archive before deletion if enabled
	if rs.config.ArchiveEnabled { //nolint:nestif //nolint:nestif
		logData, err := rs.repo.GetEntriesForArchival(ctx, before, 10000)
		if err != nil {
			return 0, fmt.Errorf("failed to fetch logs for archival: %w", err)
		}

		if len(logData) > 0 {
			if _, err := rs.ArchiveLogs(ctx, logData); err != nil {
				return 0, fmt.Errorf("failed to archive logs: %w", err)
			}
		}
	}

	// Delete old logs
	deleted, err := rs.repo.DeleteEntriesOlderThan(ctx, before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old logs: %w", err)
	}

	return deleted, nil
}

// ArchiveLogs compresses and stores logs to archive storage.
// Data is compressed to gzip format if compression is enabled.
// Returns the filename of the created archive.
func (rs *RetentionService) ArchiveLogs(ctx context.Context, logData []map[string]interface{}) (string, error) {
	if rs.storage == nil {
		return "", fmt.Errorf("storage not initialized")
	}

	if len(logData) == 0 {
		return "", fmt.Errorf("no logs to archive")
	}

	// Marshal logs to JSON
	jsonData, err := json.Marshal(logData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal logs to JSON: %w", err)
	}

	// Compress if enabled
	var archiveData []byte
	if rs.config.ArchiveCompressionEnabled {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(jsonData); err != nil {
			return "", fmt.Errorf("failed to compress logs: %w", err)
		}
		if err := gz.Close(); err != nil {
			return "", fmt.Errorf("failed to finalize gzip: %w", err)
		}
		archiveData = buf.Bytes()
	} else {
		archiveData = jsonData
	}

	// Generate filename
	now := time.Now().UTC()
	filename := fmt.Sprintf("logs-archive-%s.json", now.Format("20060102-150405"))
	if rs.config.ArchiveCompressionEnabled {
		filename += ".gz"
	}

	// Create directory if needed
	if rs.config.StorageType == storageTypeLocal {
		if err := os.MkdirAll(rs.config.LocalArchivePath, archiveDirectoryPermissions); err != nil {
			return "", fmt.Errorf("failed to create archive directory: %w", err)
		}
	}

	// Save archive
	if err := rs.storage.SaveArchive(ctx, filename, archiveData); err != nil {
		return "", fmt.Errorf("failed to save archive: %w", err)
	}

	return filename, nil
}

// RestoreFromArchive reads and decompresses logs from an archive file.
// Automatically detects and decompresses gzip-compressed data.
func (rs *RetentionService) RestoreFromArchive(ctx context.Context, filename string) ([]map[string]interface{}, error) {
	if rs == nil {
		return nil, fmt.Errorf("retention service not initialized")
	}

	if rs.storage == nil {
		return nil, fmt.Errorf("storage not initialized")
	}

	// Fetch archive data
	archiveData, err := rs.storage.GetArchive(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch archive: %w", err)
	}

	// Decompress if needed
	var jsonData []byte
	if bytes.HasPrefix(archiveData, []byte{0x1f, 0x8b}) { // gzip magic bytes
		gz, err := gzip.NewReader(bytes.NewReader(archiveData))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gz.Close() //nolint:errcheck // error ignored per defer pattern

		var buf bytes.Buffer
		if _, err := buf.ReadFrom(gz); err != nil {
			return nil, fmt.Errorf("failed to decompress archive: %w", err)
		}
		jsonData = buf.Bytes()
	} else {
		jsonData = archiveData
	}

	// Unmarshal JSON
	var logs []map[string]interface{}
	if err := json.Unmarshal(jsonData, &logs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal logs: %w", err)
	}

	return logs, nil
}

// SearchArchives finds archive files within a date range.
// Returns sorted list of archive filenames that were created between startDate and endDate.
func (rs *RetentionService) SearchArchives(ctx context.Context, startDate, endDate time.Time) ([]string, error) {
	if rs == nil {
		return nil, fmt.Errorf("retention service not initialized")
	}

	if rs.storage == nil {
		return nil, fmt.Errorf("storage not initialized")
	}

	// List all archives
	all, err := rs.storage.ListArchives(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list archives: %w", err)
	}

	// Filter by date range
	var results []string
	for _, fname := range all {
		date, err := parseArchiveDate(fname)
		if err != nil {
			continue // Skip files with unparseable dates
		}

		if !date.Before(startDate) && !date.After(endDate) {
			results = append(results, fname)
		}
	}

	// Sort results
	sort.Strings(results)

	return results, nil
}

// GetMetrics returns storage metrics for archives.
// Includes total archive count, total size, and oldest/newest archive timestamps.
func (rs *RetentionService) GetMetrics(ctx context.Context) (StorageMetrics, error) {
	if rs == nil {
		return StorageMetrics{}, fmt.Errorf("retention service not initialized")
	}

	if rs.storage == nil {
		return StorageMetrics{}, fmt.Errorf("storage not initialized")
	}

	return rs.storage.GetStorageMetrics(ctx)
}

// CreateRetentionJob creates a background job for the retention cleanup.
// The job runs daily and logs the number of deleted entries.
func (rs *RetentionService) CreateRetentionJob() Job {
	return Job{
		Name:     "log-retention",
		Interval: 24 * time.Hour,
		Fn: func(ctx context.Context) error {
			deleted, err := rs.CleanupOldLogs(ctx)
			if err != nil {
				return err
			}

			if deleted > 0 {
				if logger, ok := ctx.Value("logger").(*logrus.Logger); ok {
					logger.Infof("Deleted %d old log entries", deleted)
				}
			}

			return nil
		},
	}
}

// Helper function to parse archive date from filename
// Expected format: logs-archive-20250101-150405.json.gz
func parseArchiveDate(filename string) (time.Time, error) {
	// Extract date part from filename
	// logs-archive-YYYYMMDD-HHMMSS.json.gz
	parts := filepath.Base(filename)
	if len(parts) < 21 {
		return time.Time{}, fmt.Errorf("invalid archive filename format: %s", filename)
	}

	// Extract date-time string (should be at position 13-28: YYYYMMDD-HHMMSS)
	dateStr := parts[13:21] // YYYYMMDD

	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse archive date: %w", err)
	}

	return date, nil
}
