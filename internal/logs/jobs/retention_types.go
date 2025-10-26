// Package jobs provides types for log retention and archival.
package jobs

import (
	"context"
	"time"
)

const (
	// storageTypeLocal is the local filesystem storage type
	storageTypeLocal = "local"
	// defaultRetentionDays is the default log retention period
	defaultRetentionDays = 90
	// archiveDirectoryPermissions is the permission mode for archive directories
	archiveDirectoryPermissions = 0o750
	// archiveFilePermissions is the permission mode for created archive files
	archiveFilePermissions = 0o600
)

// RetentionConfig holds configuration for log retention and archival.
type RetentionConfig struct { //nolint:govet // struct alignment optimized for readability
	RetentionDays             int
	ArchiveEnabled            bool
	ArchiveCompressionEnabled bool
	StorageType               string // "local" or "s3"
	LocalArchivePath          string
	S3Bucket                  string
	S3Region                  string
}

// StorageMetrics represents storage usage statistics.
type StorageMetrics struct { //nolint:govet // struct alignment optimized for readability
	TotalArchives int64
	TotalSize     int64
	OldestArchive time.Time
	NewestArchive time.Time
}

// LogRepository defines the interface for log data access.
type LogRepository interface {
	DeleteEntriesOlderThan(ctx context.Context, before time.Time) (int64, error)
	GetEntriesForArchival(ctx context.Context, before time.Time, limit int) ([]map[string]interface{}, error)
	CountEntriesOlderThan(ctx context.Context, before time.Time) (int64, error)
}

// ArchiveStorage defines the interface for archive storage operations.
type ArchiveStorage interface {
	SaveArchive(ctx context.Context, filename string, data []byte) error
	ListArchives(ctx context.Context) ([]string, error)
	GetArchive(ctx context.Context, filename string) ([]byte, error)
	DeleteArchive(ctx context.Context, filename string) error
	GetStorageMetrics(ctx context.Context) (StorageMetrics, error)
}

// RetentionService manages log retention and archival.
type RetentionService struct { //nolint:govet // struct alignment optimized for readability
	config  *RetentionConfig
	repo    LogRepository  //nolint:unused // Used in service implementation
	storage ArchiveStorage //nolint:unused // Used in service implementation
}
