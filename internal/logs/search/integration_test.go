//go:build integration
// +build integration

package search

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

// PostgreSQLContainer manages a test PostgreSQL database.
type PostgreSQLContainer struct {
	container testcontainers.Container
	db        *sql.DB
	connStr   string
}

// setupPostgres creates and starts a PostgreSQL container for testing.
func setupPostgres(ctx context.Context) (*PostgreSQLContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_USER":     "user",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, err
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, err
	}

	connStr := fmt.Sprintf("postgres://user:password@%s:%s/testdb?sslmode=disable", host, port.Port())

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		container.Terminate(ctx)
		return nil, err
	}

	// Wait for DB to be ready
	for i := 0; i < 30; i++ {
		if err := db.PingContext(ctx); err == nil {
			break
		}
		time.Sleep(time.Second)
		if i == 29 {
			container.Terminate(ctx)
			return nil, fmt.Errorf("database not ready after 30 seconds")
		}
	}

	return &PostgreSQLContainer{
		container: container,
		db:        db,
		connStr:   connStr,
	}, nil
}

// cleanup terminates the container and closes the database connection.
func (pc *PostgreSQLContainer) cleanup(ctx context.Context) error {
	if pc.db != nil {
		pc.db.Close()
	}
	if pc.container != nil {
		return pc.container.Terminate(ctx)
	}
	return nil
}

// createSchema creates the logs table with full-text search support.
func (pc *PostgreSQLContainer) createSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS logs (
		id BIGSERIAL PRIMARY KEY,
		service TEXT NOT NULL,
		level TEXT NOT NULL,
		message TEXT NOT NULL,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMPTZ DEFAULT NOW(),
		search_vector tsvector GENERATED ALWAYS AS (
			setweight(to_tsvector('english', COALESCE(service, '')), 'A') ||
			setweight(to_tsvector('english', COALESCE(level, '')), 'B') ||
			setweight(to_tsvector('english', COALESCE(message, '')), 'C')
		) STORED
	);

	CREATE INDEX IF NOT EXISTS idx_logs_search ON logs USING GIN(search_vector);
	CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_logs_service ON logs(service);
	CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
	CREATE INDEX IF NOT EXISTS idx_logs_metadata ON logs USING GIN(metadata);
	`

	_, err := pc.db.ExecContext(ctx, schema)
	return err
}

// TestIntegration_BasicInsertAndQuery tests inserting and retrieving logs.
func TestIntegration_BasicInsertAndQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pg, err := setupPostgres(ctx)
	require.NoError(t, err, "failed to setup postgres")
	defer pg.cleanup(ctx)

	err = pg.createSchema(ctx)
	require.NoError(t, err, "failed to create schema")

	// Insert test data
	insertSQL := `INSERT INTO logs (service, level, message) VALUES ($1, $2, $3) RETURNING id`
	var id int64
	err = pg.db.QueryRowContext(ctx, insertSQL, "auth", "error", "authentication failed").Scan(&id)
	require.NoError(t, err, "failed to insert log")
	assert.Greater(t, id, int64(0))

	// Query the log back
	querySQL := `SELECT id, service, level, message FROM logs WHERE id = $1`
	var service, level, message string
	err = pg.db.QueryRowContext(ctx, querySQL, id).Scan(&id, &service, &level, &message)
	require.NoError(t, err, "failed to query log")
	assert.Equal(t, "auth", service)
	assert.Equal(t, "error", level)
	assert.Equal(t, "authentication failed", message)
}

// TestIntegration_FullTextSearch tests full-text search with ts_vector.
func TestIntegration_FullTextSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pg, err := setupPostgres(ctx)
	require.NoError(t, err)
	defer pg.cleanup(ctx)

	err = pg.createSchema(ctx)
	require.NoError(t, err)

	// Insert test data
	logs := []struct {
		service string
		level   string
		message string
	}{
		{"auth", "error", "database connection failed"},
		{"auth", "error", "authentication timeout"},
		{"portal", "warn", "slow database query"},
		{"api", "info", "request received"},
	}

	for _, log := range logs {
		_, err := pg.db.ExecContext(ctx,
			`INSERT INTO logs (service, level, message) VALUES ($1, $2, $3)`,
			log.service, log.level, log.message)
		require.NoError(t, err)
	}

	// Full-text search for "database"
	searchSQL := `SELECT COUNT(*) FROM logs WHERE search_vector @@ plainto_tsquery('english', $1)`
	var count int
	err = pg.db.QueryRowContext(ctx, searchSQL, "database").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 3, count, "should find 3 logs with 'database'")
}

// TestIntegration_ServiceFilter tests filtering by service field.
func TestIntegration_ServiceFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pg, err := setupPostgres(ctx)
	require.NoError(t, err)
	defer pg.cleanup(ctx)

	err = pg.createSchema(ctx)
	require.NoError(t, err)

	// Insert logs for different services
	for i := 0; i < 5; i++ {
		_, err := pg.db.ExecContext(ctx,
			`INSERT INTO logs (service, level, message) VALUES ($1, $2, $3)`,
			"auth", "info", fmt.Sprintf("request %d", i))
		require.NoError(t, err)
	}

	for i := 0; i < 3; i++ {
		_, err := pg.db.ExecContext(ctx,
			`INSERT INTO logs (service, level, message) VALUES ($1, $2, $3)`,
			"portal", "warn", fmt.Sprintf("warning %d", i))
		require.NoError(t, err)
	}

	// Filter by service
	var count int
	err = pg.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM logs WHERE service = $1`, "auth").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 5, count)

	err = pg.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM logs WHERE service = $1`, "portal").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

// TestIntegration_LevelFilter tests filtering by log level.
func TestIntegration_LevelFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pg, err := setupPostgres(ctx)
	require.NoError(t, err)
	defer pg.cleanup(ctx)

	err = pg.createSchema(ctx)
	require.NoError(t, err)

	// Insert logs with different levels
	for _, level := range []string{"error", "error", "error", "warn", "warn", "info"} {
		_, err := pg.db.ExecContext(ctx,
			`INSERT INTO logs (service, level, message) VALUES ($1, $2, $3)`,
			"test", level, "message")
		require.NoError(t, err)
	}

	// Count by level
	var errorCount, warnCount int
	pg.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM logs WHERE level = $1`, "error").Scan(&errorCount)
	pg.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM logs WHERE level = $1`, "warn").Scan(&warnCount)

	assert.Equal(t, 3, errorCount)
	assert.Equal(t, 2, warnCount)
}

// TestIntegration_TimeRangeFilter tests filtering by timestamp range.
func TestIntegration_TimeRangeFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pg, err := setupPostgres(ctx)
	require.NoError(t, err)
	defer pg.cleanup(ctx)

	err = pg.createSchema(ctx)
	require.NoError(t, err)

	now := time.Now()
	pastTime := now.Add(-24 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	// Insert logs
	_, err = pg.db.ExecContext(ctx,
		`INSERT INTO logs (service, level, message, created_at) VALUES ($1, $2, $3, $4)`,
		"test", "info", "old log", pastTime)
	require.NoError(t, err)

	_, err = pg.db.ExecContext(ctx,
		`INSERT INTO logs (service, level, message, created_at) VALUES ($1, $2, $3, $4)`,
		"test", "info", "current log", now)
	require.NoError(t, err)

	// Query within time range
	var count int
	err = pg.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM logs WHERE created_at >= $1 AND created_at <= $2`,
		pastTime, futureTime).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

// TestIntegration_MetadataJSONBFilter tests filtering by JSONB metadata.
func TestIntegration_MetadataJSONBFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pg, err := setupPostgres(ctx)
	require.NoError(t, err)
	defer pg.cleanup(ctx)

	err = pg.createSchema(ctx)
	require.NoError(t, err)

	// Insert logs with metadata
	_, err = pg.db.ExecContext(ctx,
		`INSERT INTO logs (service, level, message, metadata) VALUES ($1, $2, $3, $4)`,
		"auth", "error", "failed login", `{"user_id": 123, "attempts": 5}`)
	require.NoError(t, err)

	_, err = pg.db.ExecContext(ctx,
		`INSERT INTO logs (service, level, message, metadata) VALUES ($1, $2, $3, $4)`,
		"auth", "info", "successful login", `{"user_id": 456, "attempts": 1}`)
	require.NoError(t, err)

	// Query by metadata
	var count int
	err = pg.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM logs WHERE metadata @> jsonb_build_object('attempts', 5)::jsonb`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestIntegration_PaginationWithOffset tests pagination using limit and offset.
func TestIntegration_PaginationWithOffset(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pg, err := setupPostgres(ctx)
	require.NoError(t, err)
	defer pg.cleanup(ctx)

	err = pg.createSchema(ctx)
	require.NoError(t, err)

	// Insert 10 logs
	for i := 0; i < 10; i++ {
		_, err := pg.db.ExecContext(ctx,
			`INSERT INTO logs (service, level, message) VALUES ($1, $2, $3)`,
			"test", "info", fmt.Sprintf("log %d", i))
		require.NoError(t, err)
	}

	// Query page 1 (limit 3, offset 0)
	var count int
	err = pg.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM logs LIMIT 3 OFFSET 0`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	// Query page 2 (limit 3, offset 3)
	err = pg.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM logs LIMIT 3 OFFSET 3`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

// TestIntegration_DeleteByID tests deleting a log entry by ID.
func TestIntegration_DeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pg, err := setupPostgres(ctx)
	require.NoError(t, err)
	defer pg.cleanup(ctx)

	err = pg.createSchema(ctx)
	require.NoError(t, err)

	// Insert log
	var id int64
	err = pg.db.QueryRowContext(ctx,
		`INSERT INTO logs (service, level, message) VALUES ($1, $2, $3) RETURNING id`,
		"test", "info", "to be deleted").Scan(&id)
	require.NoError(t, err)

	// Delete log
	result, err := pg.db.ExecContext(ctx, `DELETE FROM logs WHERE id = $1`, id)
	require.NoError(t, err)

	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)

	// Verify deletion
	var count int
	pg.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM logs WHERE id = $1`, id).Scan(&count)
	assert.Equal(t, 0, count)
}

// TestIntegration_DeleteBefore tests deleting logs before a specific timestamp.
func TestIntegration_DeleteBefore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pg, err := setupPostgres(ctx)
	require.NoError(t, err)
	defer pg.cleanup(ctx)

	err = pg.createSchema(ctx)
	require.NoError(t, err)

	now := time.Now()
	pastTime := now.Add(-24 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	// Insert old and new logs
	pg.db.ExecContext(ctx,
		`INSERT INTO logs (service, level, message, created_at) VALUES ($1, $2, $3, $4)`,
		"test", "info", "old", pastTime)
	pg.db.ExecContext(ctx,
		`INSERT INTO logs (service, level, message, created_at) VALUES ($1, $2, $3, $4)`,
		"test", "info", "new", futureTime)

	// Delete logs before now
	result, err := pg.db.ExecContext(ctx, `DELETE FROM logs WHERE created_at < $1`, now)
	require.NoError(t, err)

	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, rowsAffected, int64(1))

	// Verify future log still exists
	var count int
	pg.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM logs WHERE message = $1`, "new").Scan(&count)
	assert.Equal(t, 1, count)
}

// TestIntegration_CombinedFilters tests combining multiple filters in one query.
func TestIntegration_CombinedFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pg, err := setupPostgres(ctx)
	require.NoError(t, err)
	defer pg.cleanup(ctx)

	err = pg.createSchema(ctx)
	require.NoError(t, err)

	// Insert test data
	pg.db.ExecContext(ctx,
		`INSERT INTO logs (service, level, message) VALUES ($1, $2, $3)`,
		"auth", "error", "database connection failed")
	pg.db.ExecContext(ctx,
		`INSERT INTO logs (service, level, message) VALUES ($1, $2, $3)`,
		"auth", "error", "timeout occurred")
	pg.db.ExecContext(ctx,
		`INSERT INTO logs (service, level, message) VALUES ($1, $2, $3)`,
		"auth", "info", "login successful")
	pg.db.ExecContext(ctx,
		`INSERT INTO logs (service, level, message) VALUES ($1, $2, $3)`,
		"portal", "error", "database connection failed")

	// Query: service=auth AND level=error
	var count int
	err = pg.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM logs WHERE service = $1 AND level = $2`,
		"auth", "error").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

// TestIntegration_PerformanceBenchmark tests that queries complete within performance target.
func TestIntegration_PerformanceBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pg, err := setupPostgres(ctx)
	require.NoError(t, err)
	defer pg.cleanup(ctx)

	err = pg.createSchema(ctx)
	require.NoError(t, err)

	// Insert 1000 logs for performance testing
	for i := 0; i < 1000; i++ {
		service := []string{"auth", "portal", "api", "admin"}[i%4]
		level := []string{"info", "warn", "error", "debug"}[i%4]
		_, err := pg.db.ExecContext(ctx,
			`INSERT INTO logs (service, level, message) VALUES ($1, $2, $3)`,
			service, level, fmt.Sprintf("log message %d", i))
		require.NoError(t, err)
	}

	// Perform query and measure time
	start := time.Now()
	var count int
	err = pg.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM logs WHERE service = $1 AND level = $2`,
		"auth", "error").Scan(&count)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, elapsed, 100*time.Millisecond, "query should complete in under 100ms for 1000 logs")
}
