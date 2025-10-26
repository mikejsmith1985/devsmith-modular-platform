// Package jobs provides archive storage implementations.
package jobs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LocalArchiveStorage implements ArchiveStorage using the local filesystem.
type LocalArchiveStorage struct {
	basePath string
}

// NewLocalArchiveStorage creates a new local archive storage.
func NewLocalArchiveStorage(basePath string) (*LocalArchiveStorage, error) {
	if err := os.MkdirAll(basePath, archiveDirectoryPermissions); err != nil {
		return nil, fmt.Errorf("failed to create archive directory: %w", err)
	}

	return &LocalArchiveStorage{
		basePath: basePath,
	}, nil
}

// SaveArchive saves an archive file to local storage.
func (s *LocalArchiveStorage) SaveArchive(ctx context.Context, filename string, data []byte) error {
	if filename == "" {
		return fmt.Errorf("filename is required")
	}

	if len(data) == 0 {
		return fmt.Errorf("data is required")
	}

	fullPath := filepath.Join(s.basePath, filename)
	// Ensure we're not writing outside the base path using Clean
	cleanedPath := filepath.Clean(fullPath)
	if !isWithinDirectory(cleanedPath, s.basePath) {
		return fmt.Errorf("invalid filename: path traversal detected")
	}

	// Write file
	if err := os.WriteFile(fullPath, data, archiveFilePermissions); err != nil {
		return fmt.Errorf("failed to write archive file: %w", err)
	}

	return nil
}

// ListArchives returns a list of all archive filenames.
func (s *LocalArchiveStorage) ListArchives(ctx context.Context) ([]string, error) {
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read archive directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".gz" || filepath.Ext(name) == ".json" {
			files = append(files, name)
		}
	}

	sort.Strings(files)
	return files, nil
}

// GetArchive reads an archive file from local storage.
func (s *LocalArchiveStorage) GetArchive(ctx context.Context, filename string) ([]byte, error) {
	if filename == "" {
		return nil, fmt.Errorf("filename is required")
	}

	fullPath := filepath.Join(s.basePath, filename)
	cleanedPath := filepath.Clean(fullPath)
	if !isWithinDirectory(cleanedPath, s.basePath) {
		return nil, fmt.Errorf("invalid filename: path traversal detected")
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read archive file: %w", err)
	}

	return data, nil
}

// DeleteArchive removes an archive file from local storage.
func (s *LocalArchiveStorage) DeleteArchive(ctx context.Context, filename string) error {
	if filename == "" {
		return fmt.Errorf("filename is required")
	}

	fullPath := filepath.Join(s.basePath, filename)
	cleanedPath := filepath.Clean(fullPath)
	if !isWithinDirectory(cleanedPath, s.basePath) {
		return fmt.Errorf("invalid filename: path traversal detected")
	}

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete archive file: %w", err)
	}

	return nil
}

// GetStorageMetrics returns storage metrics for local archives.
func (s *LocalArchiveStorage) GetStorageMetrics(ctx context.Context) (StorageMetrics, error) {
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		return StorageMetrics{}, fmt.Errorf("failed to read archive directory: %w", err)
	}

	metrics := StorageMetrics{
		TotalArchives: 0,
		TotalSize:     0,
	}

	var oldestTime, newestTime int64

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}

		metrics.TotalArchives++
		metrics.TotalSize += info.Size()

		fileTime := info.ModTime().Unix()
		if oldestTime == 0 || fileTime < oldestTime {
			oldestTime = fileTime
		}
		if newestTime == 0 || fileTime > newestTime {
			newestTime = fileTime
		}
	}

	return metrics, nil
}

// isWithinDirectory checks if a path is within a directory.
func isWithinDirectory(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	// Check if relative path is absolute or tries to escape directory
	if filepath.IsAbs(rel) {
		return false
	}
	if rel == ".." {
		return false
	}
	// Check if path starts with .. to prevent directory traversal
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}
