//go:build integration
// +build integration

package logs_db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupLogsIntegrationDB(ctx context.Context, t *testing.T) *sql.DB {
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
	require.NoError(t, err)
	t.Cleanup(func() { container.Terminate(context.Background()) })

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	for i := 0; i < 30; i++ {
		if err := db.PingContext(ctx); err == nil {
			break
		}
		if i == 29 {
			require.Fail(t, "database failed to become ready")
		}
		time.Sleep(time.Second)
	}

	_, err = db.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS logs")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS logs.log_entries (
			id BIGSERIAL PRIMARY KEY,
			user_id INT,
			service VARCHAR(50),
			level VARCHAR(20),
			message TEXT,
			metadata JSONB,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_logs_service_level ON logs.log_entries(service, level, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_logs_user ON logs.log_entries(user_id, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_logs_created ON logs.log_entries(created_at DESC)
	`)
	require.NoError(t, err)

	return db
}

func TestIntegration_LogEntryRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	entry := &logs_models.LogEntry{
		UserID:   1,
		Service:  "portal",
		Level:    "info",
		Message:  "User logged in",
		Metadata: []byte(`{"ip": "192.168.1.1"}`),
	}

	created, err := repo.Create(ctx, entry)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "portal", created.Service)
	assert.Equal(t, "info", created.Level)
	assert.NotZero(t, created.CreatedAt)
}

func TestIntegration_LogEntryRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	entry := &logs_models.LogEntry{
		UserID:   2,
		Service:  "review",
		Level:    "warn",
		Message:  "Review timeout",
		Metadata: []byte(`{}`),
	}

	created, err := repo.Create(ctx, entry)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, "review", retrieved.Service)
	assert.Equal(t, "warn", retrieved.Level)
}

func TestIntegration_LogEntryRepository_GetByID_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	retrieved, err := repo.GetByID(ctx, 999999)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestIntegration_LogEntryRepository_GetByService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	for i := 0; i < 5; i++ {
		_, err := repo.Create(ctx, &logs_models.LogEntry{
			UserID:   int64(i),
			Service:  "portal",
			Level:    "info",
			Message:  fmt.Sprintf("Entry %d", i),
			Metadata: []byte(`{}`),
		})
		require.NoError(t, err)
	}

	entries, err := repo.GetByService(ctx, "portal", 10, 0)
	require.NoError(t, err)
	assert.Len(t, entries, 5)
	assert.Equal(t, "portal", entries[0].Service)
}

func TestIntegration_LogEntryRepository_GetByLevel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	_, err := repo.Create(ctx, &logs_models.LogEntry{
		UserID:   1,
		Service:  "analytics",
		Level:    "error",
		Message:  "Database error",
		Metadata: []byte(`{}`),
	})
	require.NoError(t, err)

	entries, err := repo.GetByLevel(ctx, "error", 10, 0)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "error", entries[0].Level)
}

func TestIntegration_LogEntryRepository_GetByUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	for i := 0; i < 3; i++ {
		_, err := repo.Create(ctx, &logs_models.LogEntry{
			UserID:   5,
			Service:  "review",
			Level:    "debug",
			Message:  fmt.Sprintf("Debug %d", i),
			Metadata: []byte(`{}`),
		})
		require.NoError(t, err)
	}

	entries, err := repo.GetByUser(ctx, 5, 10, 0)
	require.NoError(t, err)
	assert.Len(t, entries, 3)
	assert.Equal(t, int64(5), entries[0].UserID)
}

func TestIntegration_LogEntryRepository_GetRecent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	for i := 0; i < 5; i++ {
		_, err := repo.Create(ctx, &logs_models.LogEntry{
			UserID:   int64(i),
			Service:  "portal",
			Level:    "info",
			Message:  fmt.Sprintf("Message %d", i),
			Metadata: []byte(`{}`),
		})
		require.NoError(t, err)
	}

	entries, err := repo.GetRecent(ctx, 3)
	require.NoError(t, err)
	assert.Len(t, entries, 3)
}

func TestIntegration_LogEntryRepository_GetStats(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	_, err := repo.Create(ctx, &logs_models.LogEntry{
		UserID:   1,
		Service:  "portal",
		Level:    "info",
		Message:  "Test",
		Metadata: []byte(`{}`),
	})
	require.NoError(t, err)

	_, err = repo.Create(ctx, &logs_models.LogEntry{
		UserID:   2,
		Service:  "review",
		Level:    "error",
		Message:  "Test",
		Metadata: []byte(`{}`),
	})
	require.NoError(t, err)

	stats, err := repo.GetStats(ctx)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "by_level")
	assert.Contains(t, stats, "by_service")
}

func TestIntegration_LogEntryRepository_Count(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	for i := 0; i < 5; i++ {
		_, err := repo.Create(ctx, &logs_models.LogEntry{
			UserID:   1,
			Service:  "portal",
			Level:    "info",
			Message:  "Test",
			Metadata: []byte(`{}`),
		})
		require.NoError(t, err)
	}

	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestIntegration_LogEntryRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	created, err := repo.Create(ctx, &logs_models.LogEntry{
		UserID:   1,
		Service:  "portal",
		Level:    "info",
		Message:  "To delete",
		Metadata: []byte(`{}`),
	})
	require.NoError(t, err)

	err = repo.Delete(ctx, created.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestIntegration_LogEntryRepository_DeleteByService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	for i := 0; i < 3; i++ {
		_, err := repo.Create(ctx, &logs_models.LogEntry{
			UserID:   1,
			Service:  "portal",
			Level:    "info",
			Message:  "Test",
			Metadata: []byte(`{}`),
		})
		require.NoError(t, err)
	}

	count, err := repo.DeleteByService(ctx, "portal")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestIntegration_LogEntryRepository_DeleteOlderThan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	_, err := repo.Create(ctx, &logs_models.LogEntry{
		UserID:   1,
		Service:  "portal",
		Level:    "info",
		Message:  "Recent",
		Metadata: []byte(`{}`),
	})
	require.NoError(t, err)

	count, err := repo.DeleteOlderThan(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestIntegration_LogEntryRepository_GetMetadataValue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	metadata := map[string]interface{}{
		"ip":        "192.168.1.1",
		"timestamp": 1234567890,
		"error":     "connection refused",
	}
	metadataJSON, err := json.Marshal(metadata)
	require.NoError(t, err)

	_, err = repo.Create(ctx, &logs_models.LogEntry{
		UserID:   1,
		Service:  "portal",
		Level:    "error",
		Message:  "Connection error",
		Metadata: metadataJSON,
	})
	require.NoError(t, err)

	ip, err := repo.GetMetadataValue(metadataJSON, "ip")
	require.NoError(t, err)
	assert.Equal(t, "192.168.1.1", ip)
}

func TestIntegration_LogEntryRepository_Pagination(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	for i := 0; i < 10; i++ {
		_, err := repo.Create(ctx, &logs_models.LogEntry{
			UserID:   1,
			Service:  "portal",
			Level:    "info",
			Message:  fmt.Sprintf("Message %d", i),
			Metadata: []byte(`{}`),
		})
		require.NoError(t, err)
	}

	page1, err := repo.GetByService(ctx, "portal", 5, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 5)

	page2, err := repo.GetByService(ctx, "portal", 5, 5)
	require.NoError(t, err)
	assert.Len(t, page2, 5)

	assert.NotEqual(t, page1[0].ID, page2[0].ID)
}

func TestIntegration_LogEntryRepository_MultipleServices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupLogsIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogEntryRepository(db)

	services := []string{"portal", "review", "analytics", "logs"}
	for _, service := range services {
		_, err := repo.Create(ctx, &logs_models.LogEntry{
			UserID:   1,
			Service:  service,
			Level:    "info",
			Message:  fmt.Sprintf("%s log", service),
			Metadata: []byte(`{}`),
		})
		require.NoError(t, err)
	}

	for _, service := range services {
		entries, err := repo.GetByService(ctx, service, 10, 0)
		require.NoError(t, err)
		assert.Len(t, entries, 1)
		assert.Equal(t, service, entries[0].Service)
	}
}
