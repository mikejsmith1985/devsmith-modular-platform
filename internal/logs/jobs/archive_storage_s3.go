// Package jobs provides S3 archive storage implementation.
package jobs

import (
	"context"
	"fmt"
)

// S3ArchiveStorage implements ArchiveStorage using AWS S3.
type S3ArchiveStorage struct {
	bucket string
	region string
	// TODO: Add AWS SDK client when implementing S3 support
}

// NewS3ArchiveStorage creates a new S3 archive storage.
// Currently returns a stub - full implementation in future version.
func NewS3ArchiveStorage(bucket, region string) (*S3ArchiveStorage, error) {
	if bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required")
	}

	if region == "" {
		return nil, fmt.Errorf("S3 region is required")
	}

	return &S3ArchiveStorage{
		bucket: bucket,
		region: region,
	}, nil
}

// SaveArchive saves an archive file to S3.
func (s *S3ArchiveStorage) SaveArchive(ctx context.Context, filename string, data []byte) error {
	return fmt.Errorf("S3 storage not yet implemented")
}

// ListArchives returns a list of all archive filenames in S3.
func (s *S3ArchiveStorage) ListArchives(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("S3 storage not yet implemented")
}

// GetArchive reads an archive file from S3.
func (s *S3ArchiveStorage) GetArchive(ctx context.Context, filename string) ([]byte, error) {
	return nil, fmt.Errorf("S3 storage not yet implemented")
}

// DeleteArchive removes an archive file from S3.
func (s *S3ArchiveStorage) DeleteArchive(ctx context.Context, filename string) error {
	return fmt.Errorf("S3 storage not yet implemented")
}

// GetStorageMetrics returns storage metrics for S3 archives.
func (s *S3ArchiveStorage) GetStorageMetrics(ctx context.Context) (StorageMetrics, error) {
	return StorageMetrics{}, fmt.Errorf("S3 storage not yet implemented")
}
