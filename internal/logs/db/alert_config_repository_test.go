// Package db provides database access and repository implementations for logs.
package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAlertConfigRepository_Create_Success tests creating a new alert configuration
func TestAlertConfigRepository_Create_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewAlertConfigRepository(db)
	ctx := context.Background()

	config := &models.AlertConfig{
		Service:                "review",
		ErrorThresholdPerMin:   10,
		WarningThresholdPerMin: 5,
		AlertEmail:             "admin@example.com",
		AlertWebhookURL:        "https://webhook.example.com/alerts",
		Enabled:                true,
	}

	err := repo.Create(ctx, config)
	require.NoError(t, err)
	assert.NotZero(t, config.ID, "config should have ID after create")
	assert.NotZero(t, config.CreatedAt, "config should have CreatedAt")
	assert.NotZero(t, config.UpdatedAt, "config should have UpdatedAt")
}

// TestAlertConfigRepository_Create_DuplicateService tests creating duplicate service config fails
func TestAlertConfigRepository_Create_DuplicateService(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewAlertConfigRepository(db)
	ctx := context.Background()

	config1 := &models.AlertConfig{
		Service:                "review",
		ErrorThresholdPerMin:   10,
		WarningThresholdPerMin: 5,
		Enabled:                true,
	}
	config2 := &models.AlertConfig{
		Service:                "review",
		ErrorThresholdPerMin:   15,
		WarningThresholdPerMin: 7,
		Enabled:                true,
	}

	err1 := repo.Create(ctx, config1)
	require.NoError(t, err1)

	err2 := repo.Create(ctx, config2)
	assert.Error(t, err2, "should fail on duplicate service")
}

// TestAlertConfigRepository_GetByService_Success tests retrieving config by service
func TestAlertConfigRepository_GetByService_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewAlertConfigRepository(db)
	ctx := context.Background()

	// Create initial config
	original := &models.AlertConfig{
		Service:                "review",
		ErrorThresholdPerMin:   10,
		WarningThresholdPerMin: 5,
		AlertEmail:             "admin@example.com",
		Enabled:                true,
	}
	err := repo.Create(ctx, original)
	require.NoError(t, err)

	// Retrieve and verify
	retrieved, err := repo.GetByService(ctx, "review")
	require.NoError(t, err)
	assert.Equal(t, original.ID, retrieved.ID)
	assert.Equal(t, "review", retrieved.Service)
	assert.Equal(t, 10, retrieved.ErrorThresholdPerMin)
}

// TestAlertConfigRepository_GetByService_NotFound tests retrieving non-existent config
func TestAlertConfigRepository_GetByService_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewAlertConfigRepository(db)
	ctx := context.Background()

	_, err := repo.GetByService(ctx, "nonexistent")
	assert.Error(t, err, "should fail for non-existent service")
}

// TestAlertConfigRepository_Update_Success tests updating an alert configuration
func TestAlertConfigRepository_Update_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewAlertConfigRepository(db)
	ctx := context.Background()

	// Create initial config
	config := &models.AlertConfig{
		Service:                "review",
		ErrorThresholdPerMin:   10,
		WarningThresholdPerMin: 5,
		AlertEmail:             "admin@example.com",
		Enabled:                true,
	}
	err := repo.Create(ctx, config)
	require.NoError(t, err)

	// Update config
	config.ErrorThresholdPerMin = 20
	config.AlertEmail = "newemail@example.com"
	err = repo.Update(ctx, config)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByService(ctx, "review")
	require.NoError(t, err)
	assert.Equal(t, 20, retrieved.ErrorThresholdPerMin)
	assert.Equal(t, "newemail@example.com", retrieved.AlertEmail)
}

// TestAlertConfigRepository_GetAll_Success tests retrieving all configs
func TestAlertConfigRepository_GetAll_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewAlertConfigRepository(db)
	ctx := context.Background()

	// Create multiple configs
	for _, service := range []string{"review", "portal", "analytics"} {
		config := &models.AlertConfig{
			Service:                service,
			ErrorThresholdPerMin:   10,
			WarningThresholdPerMin: 5,
			Enabled:                true,
		}
		err := repo.Create(ctx, config)
		require.NoError(t, err)
	}

	// Retrieve all
	configs, err := repo.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, configs, 3)
}

// TestAlertConfigRepository_Delete_Success tests deleting a config
func TestAlertConfigRepository_Delete_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewAlertConfigRepository(db)
	ctx := context.Background()

	config := &models.AlertConfig{
		Service:                "review",
		ErrorThresholdPerMin:   10,
		WarningThresholdPerMin: 5,
		Enabled:                true,
	}
	err := repo.Create(ctx, config)
	require.NoError(t, err)

	// Delete
	err = repo.Delete(ctx, config.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = repo.GetByService(ctx, "review")
	assert.Error(t, err, "should not find deleted config")
}

// TestAlertEventRepository_Create_Success tests creating alert events
func TestAlertEventRepository_Create_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	configRepo := NewAlertConfigRepository(db)
	eventRepo := NewAlertEventRepository(db)
	ctx := context.Background()

	// Create alert config first
	config := &models.AlertConfig{
		Service:                "review",
		ErrorThresholdPerMin:   10,
		WarningThresholdPerMin: 5,
		Enabled:                true,
	}
	err := configRepo.Create(ctx, config)
	require.NoError(t, err)

	// Create alert event
	event := &models.AlertEvent{
		ConfigID:       config.ID,
		TriggeredAt:    time.Now(),
		ErrorCount:     15,
		ThresholdValue: 10,
		AlertSent:      false,
		ErrorType:      "validation_error",
	}
	err = eventRepo.Create(ctx, event)
	require.NoError(t, err)
	assert.NotZero(t, event.ID)
}

// TestAlertEventRepository_GetByConfigID_Success tests retrieving events by config
func TestAlertEventRepository_GetByConfigID_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	configRepo := NewAlertConfigRepository(db)
	eventRepo := NewAlertEventRepository(db)
	ctx := context.Background()

	// Create config
	config := &models.AlertConfig{
		Service:                "review",
		ErrorThresholdPerMin:   10,
		WarningThresholdPerMin: 5,
		Enabled:                true,
	}
	err := configRepo.Create(ctx, config)
	require.NoError(t, err)

	// Create multiple events
	for i := 0; i < 3; i++ {
		event := &models.AlertEvent{
			ConfigID:       config.ID,
			TriggeredAt:    time.Now().Add(-time.Duration(i) * time.Hour),
			ErrorCount:     15 + int64(i),
			ThresholdValue: 10,
			AlertSent:      false,
		}
		err := eventRepo.Create(ctx, event)
		require.NoError(t, err)
	}

	// Retrieve events
	events, err := eventRepo.GetByConfigID(ctx, config.ID)
	require.NoError(t, err)
	assert.Len(t, events, 3)
}

// TestValidationAggregation_GetTopErrors_Success tests top errors aggregation
func TestValidationAggregation_GetTopErrors_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	agg := NewValidationAggregation(db)
	ctx := context.Background()

	// This test expects logs to exist
	// Should return top validation errors ordered by frequency
	errors, err := agg.GetTopErrors(ctx, "", 10, 7)
	require.NoError(t, err)
	// Assertion depends on test data setup
	assert.IsType(t, []models.ValidationError{}, errors)
}

// TestValidationAggregation_GetErrorTrends_Success tests error trending
func TestValidationAggregation_GetErrorTrends_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	agg := NewValidationAggregation(db)
	ctx := context.Background()

	trends, err := agg.GetErrorTrends(ctx, "", 7, "hourly")
	require.NoError(t, err)
	assert.IsType(t, []models.ErrorTrend{}, trends)
}

// TestLogExportService_ExportJSON_Success tests JSON export
func TestLogExportService_ExportJSON_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	exporter := NewLogExportService(db)
	ctx := context.Background()

	exportOpts := LogExportOptions{
		Format:    "json",
		Service:   "review",
		StartTime: time.Now().Add(-24 * time.Hour),
		EndTime:   time.Now(),
	}

	data, err := exporter.Export(ctx, exportOpts)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.IsType(t, []byte{}, data)
}

// TestLogExportService_ExportCSV_Success tests CSV export
func TestLogExportService_ExportCSV_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	exporter := NewLogExportService(db)
	ctx := context.Background()

	exportOpts := LogExportOptions{
		Format:    "csv",
		Service:   "review",
		StartTime: time.Now().Add(-24 * time.Hour),
		EndTime:   time.Now(),
	}

	data, err := exporter.Export(ctx, exportOpts)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// This is a placeholder - implementation depends on test database setup
	// In actual implementation, this would connect to a test PostgreSQL instance
	t.Skip("Test database setup not configured - to be implemented")
	return nil
}

// teardownTestDB cleans up test database
func teardownTestDB(t *testing.T, db *sql.DB) {
	if db != nil {
		_ = db.Close()
	}
}
