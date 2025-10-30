//go:build integration
// +build integration

package logs_db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupIntegrationDB(ctx context.Context, t *testing.T) *sql.DB {
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
		CREATE TABLE IF NOT EXISTS logs.entries (
			id BIGSERIAL PRIMARY KEY,
			service TEXT NOT NULL,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			metadata JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_logs_entries_created_at ON logs.entries(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_logs_entries_service ON logs.entries(service);
		CREATE INDEX IF NOT EXISTS idx_logs_entries_level ON logs.entries(level);
		CREATE INDEX IF NOT EXISTS idx_logs_entries_metadata ON logs.entries USING GIN(metadata)
	`)
	require.NoError(t, err)

	return db
}

// TestIntegration_InsertAndQueryBasic tests basic CRUD operations
func TestIntegration_InsertAndQueryBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogRepository(db)

	entry := &LogEntry{
		Service:  "portal",
		Level:    "info",
		Message:  "User logged in successfully",
		Metadata: map[string]interface{}{"ip": "192.168.1.1"},
	}

	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)
	assert.NotZero(t, id)

	results, err := repo.Query(ctx, &QueryFilters{Service: "portal"}, PageOptions{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, "portal", results[0].Service)
	assert.Equal(t, "User logged in successfully", results[0].Message)
}

// TestIntegration_FilterByService tests filtering logs by service with multiple services
func TestIntegration_FilterByService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogRepository(db)

	services := []string{"portal", "review", "analytics"}
	for _, service := range services {
		for j := 0; j < 3; j++ {
			_, err := repo.Insert(ctx, &LogEntry{
				Service:  service,
				Level:    "info",
				Message:  fmt.Sprintf("%s message %d", service, j),
				Metadata: map[string]interface{}{},
			})
			require.NoError(t, err)
		}
	}

	portalEntries, err := repo.Query(ctx, &QueryFilters{Service: "portal"}, PageOptions{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.Len(t, portalEntries, 3)
	for _, e := range portalEntries {
		assert.Equal(t, "portal", e.Service)
	}

	reviewEntries, err := repo.Query(ctx, &QueryFilters{Service: "review"}, PageOptions{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.Len(t, reviewEntries, 3)
	for _, e := range reviewEntries {
		assert.Equal(t, "review", e.Service)
	}
}

// TestIntegration_FilterByLevel tests filtering logs by severity level
func TestIntegration_FilterByLevel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogRepository(db)

	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		for j := 0; j < 2; j++ {
			_, err := repo.Insert(ctx, &LogEntry{
				Service:  "portal",
				Level:    level,
				Message:  fmt.Sprintf("Message at %s level %d", level, j),
				Metadata: map[string]interface{}{},
			})
			require.NoError(t, err)
			time.Sleep(10 * time.Millisecond)
		}
	}

	errorEntries, err := repo.Query(ctx, &QueryFilters{Level: "error"}, PageOptions{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.Len(t, errorEntries, 2)
	for _, e := range errorEntries {
		assert.Equal(t, "error", e.Level)
	}

	warnEntries, err := repo.Query(ctx, &QueryFilters{Level: "warn"}, PageOptions{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.Len(t, warnEntries, 2)
	for _, e := range warnEntries {
		assert.Equal(t, "warn", e.Level)
	}
}

// TestIntegration_FilterByTimeRange tests filtering by created_at timestamp range
func TestIntegration_FilterByTimeRange(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogRepository(db)

	startTime := time.Now().Add(-1 * time.Second)

	for i := 0; i < 5; i++ {
		_, err := repo.Insert(ctx, &LogEntry{
			Service:  "portal",
			Level:    "info",
			Message:  fmt.Sprintf("Time-based message %d", i),
			Metadata: map[string]interface{}{},
		})
		require.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}

	endTime := time.Now().Add(1 * time.Second)

	results, err := repo.Query(ctx, &QueryFilters{From: startTime, To: endTime}, PageOptions{Limit: 100, Offset: 0})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 5)
}

// TestIntegration_SearchByMessageSubstring tests searching logs by message substring
func TestIntegration_SearchByMessageSubstring(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogRepository(db)

	messages := []string{
		"SQL injection detected in query",
		"Buffer overflow vulnerability found",
		"Authentication failed for user",
		"SQL syntax error in migration",
	}

	for _, msg := range messages {
		_, err := repo.Insert(ctx, &LogEntry{
			Service:  "review",
			Level:    "error",
			Message:  msg,
			Metadata: map[string]interface{}{},
		})
		require.NoError(t, err)
	}

	results, err := repo.Query(ctx, &QueryFilters{Search: "SQL"}, PageOptions{Limit: 100, Offset: 0})
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

// TestIntegration_FilterByJSONBMetadata tests JSONB metadata equality filtering
func TestIntegration_FilterByJSONBMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogRepository(db)

	for i := 0; i < 3; i++ {
		_, err := repo.Insert(ctx, &LogEntry{
			Service:  "portal",
			Level:    "error",
			Message:  "Error with metadata",
			Metadata: map[string]interface{}{"error_code": "E001", "severity": "high"},
		})
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		_, err := repo.Insert(ctx, &LogEntry{
			Service:  "portal",
			Level:    "error",
			Message:  "Error with metadata",
			Metadata: map[string]interface{}{"error_code": "E002", "severity": "low"},
		})
		require.NoError(t, err)
	}

	results, err := repo.Query(ctx, &QueryFilters{MetaEquals: map[string]string{"error_code": "E001"}}, PageOptions{Limit: 100, Offset: 0})
	require.NoError(t, err)
	assert.Equal(t, 3, len(results))
}

// TestIntegration_PaginationDeterministic tests pagination with deterministic ordering
func TestIntegration_PaginationDeterministic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogRepository(db)

	for i := 0; i < 15; i++ {
		_, err := repo.Insert(ctx, &LogEntry{
			Service:  "portal",
			Level:    "info",
			Message:  fmt.Sprintf("Message %d", i),
			Metadata: map[string]interface{}{},
		})
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	page1, err := repo.Query(ctx, &QueryFilters{}, PageOptions{Limit: 5, Offset: 0})
	require.NoError(t, err)
	assert.Len(t, page1, 5)

	page2, err := repo.Query(ctx, &QueryFilters{}, PageOptions{Limit: 5, Offset: 5})
	require.NoError(t, err)
	assert.Len(t, page2, 5)

	page3, err := repo.Query(ctx, &QueryFilters{}, PageOptions{Limit: 5, Offset: 10})
	require.NoError(t, err)
	assert.Len(t, page3, 5)

	assert.NotEqual(t, page1[0].ID, page2[0].ID)
	assert.NotEqual(t, page2[0].ID, page3[0].ID)

	allInOrder := append(append(page1, page2...), page3...)
	for i := 1; i < len(allInOrder); i++ {
		assert.Greater(t, allInOrder[i-1].CreatedAt, allInOrder[i].CreatedAt)
	}
}

// TestIntegration_DeleteByID tests deleting individual log entries
func TestIntegration_DeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogRepository(db)

	id, err := repo.Insert(ctx, &LogEntry{
		Service:  "portal",
		Level:    "info",
		Message:  "To be deleted",
		Metadata: map[string]interface{}{},
	})
	require.NoError(t, err)

	err = repo.DeleteByID(ctx, id)
	require.NoError(t, err)

	results, err := repo.Query(ctx, &QueryFilters{}, PageOptions{Limit: 100, Offset: 0})
	require.NoError(t, err)
	assert.Empty(t, results)
}

// TestIntegration_DeleteByTimeRange tests deleting entries older than specified time
func TestIntegration_DeleteByTimeRange(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogRepository(db)

	for i := 0; i < 3; i++ {
		_, err := repo.Insert(ctx, &LogEntry{
			Service:  "portal",
			Level:    "info",
			Message:  fmt.Sprintf("Recent message %d", i),
			Metadata: map[string]interface{}{},
		})
		require.NoError(t, err)
	}

	count, err := repo.DeleteBefore(ctx, time.Now().Add(-1*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	all, err := repo.Query(ctx, &QueryFilters{}, PageOptions{Limit: 100, Offset: 0})
	require.NoError(t, err)
	assert.Len(t, all, 3)
}

// TestIntegration_ErrorHandling_DeleteNonExistent tests deleting non-existent entry
func TestIntegration_ErrorHandling_DeleteNonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogRepository(db)

	err := repo.DeleteByID(ctx, 999999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestIntegration_MultipleEntriesSameService tests multiple entries from same service
func TestIntegration_MultipleEntriesSameService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogRepository(db)

	for i := 0; i < 10; i++ {
		_, err := repo.Insert(ctx, &LogEntry{
			Service:  "review",
			Level:    "info",
			Message:  fmt.Sprintf("Review log %d", i),
			Metadata: map[string]interface{}{"index": float64(i)},
		})
		require.NoError(t, err)
	}

	entries, err := repo.Query(ctx, &QueryFilters{Service: "review"}, PageOptions{Limit: 100, Offset: 0})
	require.NoError(t, err)
	assert.Len(t, entries, 10)
}

// TestIntegration_CombinedFilters tests combining multiple filters
func TestIntegration_CombinedFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewLogRepository(db)

	for i := 0; i < 5; i++ {
		_, err := repo.Insert(ctx, &LogEntry{
			Service:  "portal",
			Level:    "error",
			Message:  fmt.Sprintf("Portal error %d", i),
			Metadata: map[string]interface{}{},
		})
		require.NoError(t, err)
	}

	for i := 0; i < 3; i++ {
		_, err := repo.Insert(ctx, &LogEntry{
			Service:  "review",
			Level:    "error",
			Message:  fmt.Sprintf("Review error %d", i),
			Metadata: map[string]interface{}{},
		})
		require.NoError(t, err)
	}

	results, err := repo.Query(ctx, &QueryFilters{Service: "portal", Level: "error"}, PageOptions{Limit: 100, Offset: 0})
	require.NoError(t, err)
	assert.Len(t, results, 5)
	for _, e := range results {
		assert.Equal(t, "portal", e.Service)
		assert.Equal(t, "error", e.Level)
	}
}
