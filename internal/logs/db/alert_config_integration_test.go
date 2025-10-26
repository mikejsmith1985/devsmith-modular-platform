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
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupTestPostgres creates a test PostgreSQL container
func setupTestPostgres(t *testing.T) (*sql.DB, testcontainers.Container) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "Failed to create test container")

	// Get connection string
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := "postgres://testuser:testpass@" + host + ":" + port.Port() + "/testdb?sslmode=disable"

	// Wait for database to be ready
	time.Sleep(2 * time.Second)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "Failed to connect to test database")

	err = db.PingContext(ctx)
	require.NoError(t, err, "Failed to ping test database")

	// Create schema and tables
	schema := `
	CREATE SCHEMA IF NOT EXISTS logs;

	CREATE TABLE logs.alert_configs (
		id BIGSERIAL PRIMARY KEY,
		service TEXT NOT NULL UNIQUE,
		error_threshold_per_min INT NOT NULL DEFAULT 10,
		warning_threshold_per_min INT NOT NULL DEFAULT 5,
		alert_email TEXT,
		alert_webhook_url TEXT,
		enabled BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE logs.alert_events (
		id BIGSERIAL PRIMARY KEY,
		config_id BIGINT NOT NULL REFERENCES logs.alert_configs(id) ON DELETE CASCADE,
		triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		error_count INT NOT NULL,
		threshold_value INT NOT NULL,
		alert_sent BOOLEAN NOT NULL DEFAULT false,
		sent_at TIMESTAMPTZ,
		error_type TEXT
	);

	CREATE INDEX idx_alert_configs_service ON logs.alert_configs(service);
	CREATE INDEX idx_alert_events_config_id ON logs.alert_events(config_id);
	CREATE INDEX idx_alert_events_triggered_at ON logs.alert_events(triggered_at DESC);
	CREATE INDEX idx_alert_events_alert_sent ON logs.alert_events(alert_sent);
	`

	_, err = db.ExecContext(ctx, schema)
	require.NoError(t, err, "Failed to create schema")

	return db, container
}

// cleanupTestPostgres cleans up the test container
func cleanupTestPostgres(t *testing.T, container testcontainers.Container) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := container.Terminate(ctx)
	require.NoError(t, err, "Failed to terminate container")
}

// TestAlertConfigRepository_Create_WritesToDatabase tests that creating an alert config writes to database
func TestAlertConfigRepository_Create_WritesToDatabase(t *testing.T) {
	db, container := setupTestPostgres(t)
	defer db.Close()
	defer cleanupTestPostgres(t, container)

	repo := NewAlertConfigRepository(db)
	ctx := context.Background()

	// Create alert config
	config := &models.AlertConfig{
		Service:                "review",
		ErrorThresholdPerMin:   10,
		WarningThresholdPerMin: 5,
		AlertEmail:             "admin@example.com",
		AlertWebhookURL:        "https://webhook.example.com/alerts",
		Enabled:                true,
	}

	// WRITE to database
	err := repo.Create(ctx, config)
	require.NoError(t, err, "Failed to write alert config to database")
	assert.NotZero(t, config.ID, "Config should have ID after write")
	t.Logf("✅ Successfully WROTE to database: config ID = %d", config.ID)

	// VERIFY: Read back from database
	retrieved, err := repo.GetByService(ctx, "review")
	require.NoError(t, err, "Failed to read from database")
	assert.Equal(t, config.ID, retrieved.ID, "Retrieved config ID should match")
	assert.Equal(t, "review", retrieved.Service, "Retrieved service should match")
	assert.Equal(t, 10, retrieved.ErrorThresholdPerMin, "Retrieved threshold should match")
	t.Logf("✅ Successfully READ from database: %+v", retrieved)
}

// TestAlertConfigRepository_Update_WritesToDatabase tests that updating an alert config writes to database
func TestAlertConfigRepository_Update_WritesToDatabase(t *testing.T) {
	db, container := setupTestPostgres(t)
	defer db.Close()
	defer cleanupTestPostgres(t, container)

	repo := NewAlertConfigRepository(db)
	ctx := context.Background()

	// Create initial config
	config := &models.AlertConfig{
		Service:                "portal",
		ErrorThresholdPerMin:   5,
		WarningThresholdPerMin: 2,
		Enabled:                true,
	}
	err := repo.Create(ctx, config)
	require.NoError(t, err)
	originalID := config.ID

	// Update the config
	config.ErrorThresholdPerMin = 20
	config.WarningThresholdPerMin = 10
	config.AlertEmail = "ops@example.com"
	err = repo.Update(ctx, config)
	require.NoError(t, err, "Failed to update alert config")
	t.Logf("✅ Successfully WROTE update to database: config ID = %d", config.ID)

	// VERIFY: Read back updated values
	retrieved, err := repo.GetByService(ctx, "portal")
	require.NoError(t, err)
	assert.Equal(t, originalID, retrieved.ID, "ID should not change")
	assert.Equal(t, 20, retrieved.ErrorThresholdPerMin, "Error threshold should be updated")
	assert.Equal(t, 10, retrieved.WarningThresholdPerMin, "Warning threshold should be updated")
	assert.Equal(t, "ops@example.com", retrieved.AlertEmail, "Email should be updated")
	t.Logf("✅ Successfully READ updated config from database: %+v", retrieved)
}

// TestAlertEventRepository_Create_WritesToDatabase tests that creating an alert event writes to database
func TestAlertEventRepository_Create_WritesToDatabase(t *testing.T) {
	db, container := setupTestPostgres(t)
	defer db.Close()
	defer cleanupTestPostgres(t, container)

	configRepo := NewAlertConfigRepository(db)
	eventRepo := NewAlertEventRepository(db)
	ctx := context.Background()

	// First, create a config
	config := &models.AlertConfig{
		Service:                "analytics",
		ErrorThresholdPerMin:   15,
		WarningThresholdPerMin: 8,
		Enabled:                true,
	}
	err := configRepo.Create(ctx, config)
	require.NoError(t, err)

	// Create alert event
	event := &models.AlertEvent{
		ConfigID:       config.ID,
		ErrorCount:     20,
		ThresholdValue: 15,
		AlertSent:      false,
		ErrorType:      "validation_error",
	}

	// WRITE to database
	err = eventRepo.Create(ctx, event)
	require.NoError(t, err, "Failed to write alert event to database")
	assert.NotZero(t, event.ID, "Event should have ID after write")
	t.Logf("✅ Successfully WROTE alert event to database: event ID = %d", event.ID)

	// VERIFY: Read back from database
	retrieved, err := eventRepo.GetByID(ctx, event.ID)
	require.NoError(t, err, "Failed to read event from database")
	assert.Equal(t, event.ID, retrieved.ID, "Retrieved event ID should match")
	assert.Equal(t, config.ID, retrieved.ConfigID, "Retrieved config ID should match")
	assert.Equal(t, 20, retrieved.ErrorCount, "Error count should match")
	assert.Equal(t, "validation_error", retrieved.ErrorType, "Error type should match")
	t.Logf("✅ Successfully READ alert event from database: %+v", retrieved)
}

// TestAlertEventRepository_GetByConfigID_WritesToDatabase tests querying multiple alert events
func TestAlertEventRepository_GetByConfigID_WritesToDatabase(t *testing.T) {
	db, container := setupTestPostgres(t)
	defer db.Close()
	defer cleanupTestPostgres(t, container)

	configRepo := NewAlertConfigRepository(db)
	eventRepo := NewAlertEventRepository(db)
	ctx := context.Background()

	// Create config
	config := &models.AlertConfig{
		Service:                "logs",
		ErrorThresholdPerMin:   25,
		WarningThresholdPerMin: 12,
		Enabled:                true,
	}
	err := configRepo.Create(ctx, config)
	require.NoError(t, err)

	// Create multiple alert events
	eventIDs := make([]int64, 3)
	for i := 0; i < 3; i++ {
		event := &models.AlertEvent{
			ConfigID:       config.ID,
			ErrorCount:     30 + i*5,
			ThresholdValue: 25,
			AlertSent:      i == 0, // First one marked as sent
			ErrorType:      "security_violation",
		}
		// WRITE multiple events to database
		err := eventRepo.Create(ctx, event)
		require.NoError(t, err, "Failed to write event %d", i)
		eventIDs[i] = event.ID
		t.Logf("✅ Successfully WROTE alert event %d to database: event ID = %d", i+1, event.ID)
	}

	// VERIFY: Read all events for this config
	retrieved, err := eventRepo.GetByConfigID(ctx, config.ID)
	require.NoError(t, err, "Failed to retrieve events by config ID")
	assert.Equal(t, 3, len(retrieved), "Should retrieve 3 events")

	// Verify all retrieved
	for i, event := range retrieved {
		assert.Equal(t, config.ID, event.ConfigID, "Event %d should reference correct config", i)
		assert.Equal(t, "security_violation", event.ErrorType, "Event %d should have correct type", i)
	}
	t.Logf("✅ Successfully READ all %d alert events from database", len(retrieved))
}

// TestAlertConfigRepository_Delete_RemovesFromDatabase tests that deleting removes from database
func TestAlertConfigRepository_Delete_RemovesFromDatabase(t *testing.T) {
	db, container := setupTestPostgres(t)
	defer db.Close()
	defer cleanupTestPostgres(t, container)

	repo := NewAlertConfigRepository(db)
	ctx := context.Background()

	// Create config
	config := &models.AlertConfig{
		Service:              "temp-service",
		ErrorThresholdPerMin: 10,
		Enabled:              true,
	}
	err := repo.Create(ctx, config)
	require.NoError(t, err)
	configID := config.ID

	// Verify it was written
	_, err = repo.GetByService(ctx, "temp-service")
	require.NoError(t, err, "Config should exist after creation")

	// DELETE from database
	err = repo.Delete(ctx, configID)
	require.NoError(t, err, "Failed to delete alert config")
	t.Logf("✅ Successfully DELETED from database: config ID = %d", configID)

	// VERIFY: Confirm deletion
	_, err = repo.GetByService(ctx, "temp-service")
	assert.Error(t, err, "Config should not exist after deletion")
	t.Logf("✅ Successfully VERIFIED deletion from database")
}
