// Package db provides database access and repository implementations for logs.
package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAlertConfigRepository_CreateSimple_Success tests creating and reading alert config from database
//
//nolint:funlen // Test function requires multiple setup and verification steps
func TestAlertConfigRepository_CreateSimple_Success(t *testing.T) {
	// Skip if database not available
	dsn := "postgres://devsmith:devsmith@localhost:5432/devsmith_test?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("Database not available: %v", err)
	}
	pingErr := db.Ping()
	if pingErr != nil {
		closeErr := db.Close()
		if closeErr != nil {
			t.Logf("Failed to close database: %v", closeErr)
		}
		t.Skipf("Database not reachable: %v", pingErr)
	}
	defer func() {
		closeErr := db.Close()
		if closeErr != nil {
			t.Logf("Failed to close database: %v", closeErr)
		}
	}()

	// Setup: Create schema and tables
	schema := `
	DROP TABLE IF EXISTS logs.alert_events CASCADE;
	DROP TABLE IF EXISTS logs.alert_configs CASCADE;
	DROP SCHEMA IF EXISTS logs CASCADE;

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
	`
	_, err = db.Exec(schema)
	require.NoError(t, err, "Failed to create schema")

	// Test: Create alert config
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

	// Write to database
	err = repo.Create(ctx, config)
	require.NoError(t, err, "Failed to create alert config")
	assert.NotZero(t, config.ID, "Config should have ID after creation")
	t.Logf("✅ Successfully wrote to database: config ID = %d", config.ID)

	// Verify: Read back from database
	retrieved, err := repo.GetByService(ctx, "review")
	require.NoError(t, err, "Failed to retrieve config")
	assert.Equal(t, config.ID, retrieved.ID, "Retrieved config should have same ID")
	assert.Equal(t, "review", retrieved.Service, "Retrieved config should have correct service")
	assert.Equal(t, 10, retrieved.ErrorThresholdPerMin, "Retrieved config should have correct threshold")
	t.Logf("✅ Successfully read from database: config = %+v", retrieved)

	// Test: Create alert event
	eventRepo := NewAlertEventRepository(db)
	event := &models.AlertEvent{
		ConfigID:       config.ID,
		ErrorCount:     15,
		ThresholdValue: 10,
		AlertSent:      false,
		ErrorType:      "validation_error",
	}

	err = eventRepo.Create(ctx, event)
	require.NoError(t, err, "Failed to create alert event")
	assert.NotZero(t, event.ID, "Event should have ID after creation")
	t.Logf("✅ Successfully wrote alert event: event ID = %d", event.ID)

	// Verify: Read back event from database
	retrievedEvent, err := eventRepo.GetByID(ctx, event.ID)
	require.NoError(t, err, "Failed to retrieve event")
	assert.Equal(t, event.ID, retrievedEvent.ID, "Retrieved event should have same ID")
	assert.Equal(t, config.ID, retrievedEvent.ConfigID, "Event should reference correct config")
	t.Logf("✅ Successfully read alert event from database: event = %+v", retrievedEvent)

	// Cleanup: Drop schema
	_, cleanupErr := db.Exec("DROP SCHEMA IF EXISTS logs CASCADE")
	if cleanupErr != nil {
		t.Logf("Warning: Failed to cleanup schema: %v", cleanupErr)
	}
}
