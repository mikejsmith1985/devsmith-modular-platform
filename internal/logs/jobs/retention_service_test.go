// Package jobs provides tests for the retention and archival service.
package jobs

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock interfaces for testing
type MockLogRepository struct {
	mock.Mock
}

func (m *MockLogRepository) DeleteEntriesOlderThan(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLogRepository) GetEntriesForArchival(ctx context.Context, before time.Time, limit int) ([]map[string]interface{}, error) {
	args := m.Called(ctx, before, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockLogRepository) CountEntriesOlderThan(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

type MockArchiveStorage struct {
	mock.Mock
}

func (m *MockArchiveStorage) SaveArchive(ctx context.Context, filename string, data []byte) error {
	args := m.Called(ctx, filename, data)
	return args.Error(0)
}

func (m *MockArchiveStorage) ListArchives(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockArchiveStorage) GetArchive(ctx context.Context, filename string) ([]byte, error) {
	args := m.Called(ctx, filename)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockArchiveStorage) DeleteArchive(ctx context.Context, filename string) error {
	args := m.Called(ctx, filename)
	return args.Error(0)
}

func (m *MockArchiveStorage) GetStorageMetrics(ctx context.Context) (StorageMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(StorageMetrics), args.Error(1)
}

// RED Phase Tests - These should fail until implementation is complete

func TestRetentionService_NewRetentionService_Success(t *testing.T) {
	// GIVEN: Valid configuration parameters
	mockRepo := new(MockLogRepository)
	mockStorage := new(MockArchiveStorage)
	config := RetentionConfig{
		RetentionDays:             90,
		ArchiveEnabled:            true,
		ArchiveCompressionEnabled: true,
		StorageType:               "local",
		LocalArchivePath:          "/tmp/archives",
	}

	// WHEN: Creating a retention service
	service, err := NewRetentionService(&config, mockRepo, mockStorage)

	// THEN: Service should be created successfully
	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, 90, service.config.RetentionDays)
}

func TestRetentionService_LoadConfig_FromEnvironment(t *testing.T) {
	// GIVEN: Environment variables are set
	t.Setenv("LOG_RETENTION_DAYS", "60")
	t.Setenv("LOG_ARCHIVE_ENABLED", "true")
	t.Setenv("LOG_ARCHIVE_COMPRESSION", "true")
	t.Setenv("LOG_ARCHIVE_STORAGE_TYPE", "local")
	t.Setenv("LOG_ARCHIVE_LOCAL_PATH", "/tmp/logs-archive")

	// WHEN: Loading configuration from environment
	config, err := LoadRetentionConfig()

	// THEN: Config should be loaded correctly with environment values
	require.NoError(t, err)
	assert.Equal(t, 60, config.RetentionDays)
	assert.True(t, config.ArchiveEnabled)
	assert.True(t, config.ArchiveCompressionEnabled)
	assert.Equal(t, "local", config.StorageType)
	assert.Equal(t, "/tmp/logs-archive", config.LocalArchivePath)
}

func TestRetentionService_LoadConfig_DefaultValues(t *testing.T) {
	// GIVEN: No environment variables set
	t.Setenv("LOG_RETENTION_DAYS", "")
	t.Setenv("LOG_ARCHIVE_ENABLED", "")

	// WHEN: Loading configuration from environment
	config, err := LoadRetentionConfig()

	// THEN: Should use default values
	require.NoError(t, err)
	assert.Equal(t, 90, config.RetentionDays)
	assert.False(t, config.ArchiveEnabled)
}

func TestRetentionService_CleanupOldLogs_Success(t *testing.T) {
	// GIVEN: Retention service with mock repository
	mockRepo := new(MockLogRepository)
	mockStorage := new(MockArchiveStorage)
	config := RetentionConfig{
		RetentionDays:  90,
		ArchiveEnabled: false,
	}
	service, _ := NewRetentionService(&config, mockRepo, mockStorage)

	// Setup expectations
	before := time.Now().AddDate(0, 0, -90)
	mockRepo.On("DeleteEntriesOlderThan", mock.MatchedBy(func(ctx context.Context) bool {
		return true
	}), mock.MatchedBy(func(t time.Time) bool {
		return t.Before(before.Add(24*time.Hour)) && t.After(before.Add(-24*time.Hour))
	})).Return(int64(1000), nil)

	// WHEN: Running cleanup
	deleted, err := service.CleanupOldLogs(context.Background())

	// THEN: Should delete old logs successfully
	require.NoError(t, err)
	assert.Equal(t, int64(1000), deleted)
	mockRepo.AssertExpectations(t)
}

func TestRetentionService_CleanupOldLogs_WithArchive(t *testing.T) {
	// GIVEN: Retention service with archival enabled
	mockRepo := new(MockLogRepository)
	mockStorage := new(MockArchiveStorage)
	config := RetentionConfig{
		RetentionDays:             90,
		ArchiveEnabled:            true,
		ArchiveCompressionEnabled: true,
	}
	service, _ := NewRetentionService(&config, mockRepo, mockStorage)

	// Setup expectations
	logData := []map[string]interface{}{
		{"id": 1, "message": "test1", "created_at": "2025-07-01"},
		{"id": 2, "message": "test2", "created_at": "2025-07-02"},
	}
	mockRepo.On("GetEntriesForArchival", mock.Anything, mock.Anything, mock.Anything).
		Return(logData, nil)
	mockStorage.On("SaveArchive", mock.Anything, mock.MatchedBy(func(s string) bool {
		return s != ""
	}), mock.Anything).Return(nil)
	mockRepo.On("DeleteEntriesOlderThan", mock.Anything, mock.Anything).Return(int64(2), nil)

	// WHEN: Running cleanup with archival
	deleted, err := service.CleanupOldLogs(context.Background())

	// THEN: Should archive before deleting
	require.NoError(t, err)
	assert.Equal(t, int64(2), deleted)
	mockStorage.AssertCalled(t, "SaveArchive", mock.Anything, mock.Anything, mock.Anything)
	mockRepo.AssertExpectations(t)
}

func TestRetentionService_RestoreLogsFromArchive_Success(t *testing.T) {
	// GIVEN: Archived logs available
	mockRepo := new(MockLogRepository)
	mockStorage := new(MockArchiveStorage)
	config := RetentionConfig{
		RetentionDays:  90,
		ArchiveEnabled: true,
	}
	service, _ := NewRetentionService(&config, mockRepo, mockStorage)

	// Create sample compressed archive data
	archiveData := []byte{0x1f, 0x8b, 0x08, 0x00}

	mockStorage.On("GetArchive", mock.Anything, "logs-archive-20250101.json.gz").
		Return(archiveData, nil)

	// WHEN: Restoring logs from archive
	logs, err := service.RestoreFromArchive(context.Background(), "logs-archive-20250101.json.gz")

	// THEN: Logs should be restored
	require.NoError(t, err)
	assert.NotNil(t, logs)
	mockStorage.AssertExpectations(t)
}

func TestRetentionService_SearchArchives_ByDateRange(t *testing.T) {
	// GIVEN: Multiple archives available
	mockRepo := new(MockLogRepository)
	mockStorage := new(MockArchiveStorage)
	config := RetentionConfig{
		RetentionDays:  90,
		ArchiveEnabled: true,
	}
	service, _ := NewRetentionService(&config, mockRepo, mockStorage)

	mockStorage.On("ListArchives", mock.Anything).Return([]string{
		"logs-archive-20250101.json.gz",
		"logs-archive-20250102.json.gz",
		"logs-archive-20250105.json.gz",
	}, nil)

	// WHEN: Searching archives by date range
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)
	results, err := service.SearchArchives(context.Background(), startDate, endDate)

	// THEN: Should return archives within date range
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Contains(t, results, "logs-archive-20250101.json.gz")
	assert.Contains(t, results, "logs-archive-20250102.json.gz")
}

func TestRetentionService_GetStorageMetrics_Success(t *testing.T) {
	// GIVEN: Retention service configured
	mockRepo := new(MockLogRepository)
	mockStorage := new(MockArchiveStorage)
	config := RetentionConfig{
		RetentionDays:  90,
		ArchiveEnabled: true,
	}
	service, _ := NewRetentionService(&config, mockRepo, mockStorage)

	mockStorage.On("GetStorageMetrics", mock.Anything).Return(StorageMetrics{
		TotalArchives: 5,
		TotalSize:     1073741824,
		OldestArchive: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		NewestArchive: time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
	}, nil)

	// WHEN: Getting storage metrics
	metrics, err := service.GetMetrics(context.Background())

	// THEN: Should return storage metrics
	require.NoError(t, err)
	assert.Equal(t, 5, metrics.TotalArchives)
	assert.Equal(t, int64(1073741824), metrics.TotalSize)
}

func TestRetentionService_CreateRetentionJob_Success(t *testing.T) {
	// GIVEN: Retention service configured
	mockRepo := new(MockLogRepository)
	mockStorage := new(MockArchiveStorage)
	config := RetentionConfig{
		RetentionDays:  90,
		ArchiveEnabled: true,
	}
	service, _ := NewRetentionService(&config, mockRepo, mockStorage)

	mockRepo.On("DeleteEntriesOlderThan", mock.Anything, mock.Anything).Return(int64(100), nil)

	// WHEN: Creating retention job
	job := service.CreateRetentionJob()

	// THEN: Job should be created with correct properties
	require.NotNil(t, job)
	assert.Equal(t, "log-retention", job.Name)
	assert.Equal(t, 24*time.Hour, job.Interval)

	// WHEN: Executing the job
	err := job.Fn(context.Background())

	// THEN: Job should execute successfully
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRetentionService_ValidateRetentionConfig_InvalidDays(t *testing.T) {
	// GIVEN: Invalid retention days
	config := RetentionConfig{
		RetentionDays: -1,
	}

	// WHEN: Validating config
	err := config.Validate()

	// THEN: Should return error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RetentionDays")
}

func TestRetentionService_ValidateRetentionConfig_Valid(t *testing.T) {
	// GIVEN: Valid retention configuration
	config := RetentionConfig{
		RetentionDays:    90,
		ArchiveEnabled:   true,
		StorageType:      "local",
		LocalArchivePath: "/tmp/archives",
	}

	// WHEN: Validating config
	err := config.Validate()

	// THEN: Should not return error
	require.NoError(t, err)
}

func TestRetentionService_CleanupOldLogs_DatabaseError(t *testing.T) {
	// GIVEN: Repository returns error
	mockRepo := new(MockLogRepository)
	mockStorage := new(MockArchiveStorage)
	config := RetentionConfig{
		RetentionDays:  90,
		ArchiveEnabled: false,
	}
	service, _ := NewRetentionService(&config, mockRepo, mockStorage)

	mockRepo.On("DeleteEntriesOlderThan", mock.Anything, mock.Anything).
		Return(int64(0), errors.New("database error"))

	// WHEN: Running cleanup
	_, err := service.CleanupOldLogs(context.Background())

	// THEN: Should return error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestRetentionService_ArchiveLogs_CompressionError(t *testing.T) {
	// GIVEN: Archive storage fails
	mockRepo := new(MockLogRepository)
	mockStorage := new(MockArchiveStorage)
	config := RetentionConfig{
		RetentionDays:             90,
		ArchiveEnabled:            true,
		ArchiveCompressionEnabled: true,
	}
	service, _ := NewRetentionService(&config, mockRepo, mockStorage)

	logData := []map[string]interface{}{
		{"id": 1, "message": "test"},
	}
	mockStorage.On("SaveArchive", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("storage error"))

	// WHEN: Archiving logs
	_, err := service.ArchiveLogs(context.Background(), logData)

	// THEN: Should return error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "storage error")
}

func TestRetentionService_SearchArchives_NoMatches(t *testing.T) {
	// GIVEN: Archives outside search date range
	mockRepo := new(MockLogRepository)
	mockStorage := new(MockArchiveStorage)
	config := RetentionConfig{
		RetentionDays:  90,
		ArchiveEnabled: true,
	}
	service, _ := NewRetentionService(&config, mockRepo, mockStorage)

	mockStorage.On("ListArchives", mock.Anything).Return([]string{
		"logs-archive-20250110.json.gz",
		"logs-archive-20250115.json.gz",
	}, nil)

	// WHEN: Searching archives outside available range
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)
	results, err := service.SearchArchives(context.Background(), startDate, endDate)

	// THEN: Should return empty results
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestRetentionService_ArchiveLocalStorage_CreatesDirectory(t *testing.T) {
	// GIVEN: Archive path that doesn't exist
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "subdir", "archives")
	mockRepo := new(MockLogRepository)
	mockStorage := new(MockArchiveStorage)
	config := RetentionConfig{
		RetentionDays:    90,
		ArchiveEnabled:   true,
		LocalArchivePath: archivePath,
	}
	service, _ := NewRetentionService(&config, mockRepo, mockStorage)

	logData := []map[string]interface{}{
		{"id": 1, "message": "test"},
	}
	mockStorage.On("SaveArchive", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// WHEN: Archiving logs
	_, err := service.ArchiveLogs(context.Background(), logData)

	// THEN: Directory should be created
	require.NoError(t, err)
	assert.DirExists(t, archivePath)
}
