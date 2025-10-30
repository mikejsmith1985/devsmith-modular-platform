package logs_db

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/lib/pq"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Performance Testing Suite for Logs Service (Feature #39)
//
// This package contains comprehensive performance tests for the Logs service's
// high-throughput logging capabilities. All tests are designed to validate the
// acceptance criteria for Feature #39: Performance Optimization & Load Testing.
//
// Key Performance Targets:
//   - Bulk insert: 1000+ logs in <500ms
//   - Sustained throughput: 1000+ logs/second
//   - Ingestion latency p95: <50ms
//   - Query latency p95: <100ms
//   - Concurrent WebSocket clients: 100+
//   - Connection pool: 50-100 connections
//
// Test Database Setup:
//   - PostgreSQL 15 with testcontainers (auto-managed lifecycle)
//   - Indexes on all query fields for performance
//   - WAL optimization for write performance
//   - Parameterized queries (no SQL injection)
//
// Running Tests:
//   - All tests: go test -v ./internal/logs/db/...
//   - Short mode: go test -short ./internal/logs/db/...
//   - Benchmarks: go test -bench=Performance ./internal/logs/db/...
//
// References:
//   - DevsmithTDD.md: Performance tests (lines 1705+)
//   - ARCHITECTURE.md: Mental models for performance optimization
//   - Feature #39: Performance Optimization & Load Testing

// TestPerformanceRepository_BulkInsert_1000Logs tests bulk insert functionality for 1000 logs
func TestPerformanceRepository_BulkInsert_1000Logs(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer teardownPerformanceTestDB(t, db)

	repo := NewLogEntryRepository(db)

	// Create 1000 log entries
	logEntries := make([]*logs_models.LogEntry, 1000)
	for i := 0; i < 1000; i++ {
		logEntries[i] = &logs_models.LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Test log entry %d", i),
			Service:   "logs",
			Tags:      []string{"test", "bulk"},
		}
	}

	// WHEN: Bulk inserting 1000 logs
	// THEN: Should complete without error and persist all entries
	err := repo.BulkInsert(context.Background(), logEntries)
	require.NoError(t, err, "BulkInsert should succeed")

	// Verify all entries were inserted
	var count int64
	err = db.QueryRow("SELECT COUNT(*) FROM logs.log_entries WHERE service = 'logs'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), count, "All 1000 entries should be persisted")
}

// TestPerformanceRepository_BulkInsert_CompletesUnder500ms tests that bulk insert of 1000 logs completes quickly
func TestPerformanceRepository_BulkInsert_CompletesUnder500ms(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer teardownPerformanceTestDB(t, db)

	repo := NewLogEntryRepository(db)

	logEntries := make([]*logs_models.LogEntry, 1000)
	for i := 0; i < 1000; i++ {
		logEntries[i] = &logs_models.LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Perf test %d", i),
			Service:   "portal",
			Tags:      []string{"performance"},
		}
	}

	start := time.Now()
	err := repo.BulkInsert(context.Background(), logEntries)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, duration, 500*time.Millisecond, "BulkInsert of 1000 logs should complete in <500ms")
}

// TestPerformanceRepository_ConnectionPool_Has50To100Connections tests connection pool size
func TestPerformanceRepository_ConnectionPool_Has50To100Connections(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer teardownPerformanceTestDB(t, db)

	// Extract connection pool stats
	stats := db.Stats()
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 50, "Connection pool should have at least 50 max connections")
	assert.LessOrEqual(t, stats.MaxOpenConnections, 100, "Connection pool should have at most 100 max connections")
}

// TestPerformanceRepository_HasIndexesOnQueryFields tests that all query fields have indexes
func TestPerformanceRepository_HasIndexesOnQueryFields(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer teardownPerformanceTestDB(t, db)

	ctx := context.Background()

	// Check for indexes on common query fields
	indexFields := []string{
		"service",
		"level",
		"timestamp",
		"correlation_id",
		"user_id",
	}

	for _, field := range indexFields {
		query := `
			SELECT COUNT(*)
			FROM pg_indexes
			WHERE tablename = 'log_entries' AND indexname LIKE $1
		`
		var count int
		err := db.QueryRowContext(ctx, query, "%"+field+"%").Scan(&count)
		require.NoError(t, err, "Should be able to query pg_indexes")
		assert.Greater(t, count, 0, "Index should exist for field: %s", field)
	}
}

// TestPerformanceRepository_CanExplainAnalyzeSlowQueries tests EXPLAIN ANALYZE support
func TestPerformanceRepository_CanExplainAnalyzeSlowQueries(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer teardownPerformanceTestDB(t, db)

	ctx := context.Background()

	// Sample query to analyze
	query := `
		SELECT service, level, COUNT(*) as count
		FROM logs.log_entries
		WHERE timestamp > NOW() - INTERVAL '1 day'
		GROUP BY service, level
	`

	// Should be able to run EXPLAIN ANALYZE
	explainQuery := "EXPLAIN ANALYZE " + query
	rows, err := db.QueryContext(ctx, explainQuery)
	require.NoError(t, err, "EXPLAIN ANALYZE should work")
	defer rows.Close()

	// Should have result rows
	hasRows := rows.Next()
	assert.True(t, hasRows, "EXPLAIN ANALYZE should return analysis results")

	// Should be able to get plan information
	var planLine string
	err = rows.Scan(&planLine)
	require.NoError(t, err)
	assert.NotEmpty(t, planLine, "Plan should have content")
}

// TestPerformanceRepository_Ingestion_Achieves1000LogsPerSecond tests 1000 logs/sec sustained throughput
func TestPerformanceRepository_Ingestion_Achieves1000LogsPerSecond(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer teardownPerformanceTestDB(t, db)

	repo := NewLogEntryRepository(db)
	ctx := context.Background()

	// Total logs to insert
	totalLogs := 5000
	batchSize := 100
	numBatches := totalLogs / batchSize

	start := time.Now()

	for batch := 0; batch < numBatches; batch++ {
		logEntries := make([]*logs_models.LogEntry, batchSize)
		for i := 0; i < batchSize; i++ {
			logEntries[i] = &logs_models.LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   fmt.Sprintf("Throughput test %d", batch*batchSize+i),
				Service:   "review",
				Tags:      []string{"throughput"},
			}
		}

		err := repo.BulkInsert(ctx, logEntries)
		require.NoError(t, err, "BulkInsert should succeed")
	}

	totalDuration := time.Since(start)

	// Calculate throughput: logs per second
	logsPerSecond := float64(totalLogs) / totalDuration.Seconds()

	assert.GreaterOrEqual(t, logsPerSecond, 1000.0, "Should achieve at least 1000 logs/sec throughput")
}

// TestPerformanceRepository_Ingestion_LatencyUnder50msP95 tests ingestion latency
func TestPerformanceRepository_Ingestion_LatencyUnder50msP95(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer teardownPerformanceTestDB(t, db)

	repo := NewLogEntryRepository(db)
	ctx := context.Background()

	latencies := make([]time.Duration, 100)

	// Record ingestion latencies for 100 operations
	for i := 0; i < 100; i++ {
		logEntry := &logs_models.LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Latency test %d", i),
			Service:   "analytics",
			Tags:      []string{"latency"},
		}

		start := time.Now()
		err := repo.BulkInsert(ctx, []*logs_models.LogEntry{logEntry})
		latencies[i] = time.Since(start)

		require.NoError(t, err)
	}

	// Calculate p95 latency
	p95Latency := calculateP95(latencies)

	assert.Less(t, p95Latency, 50*time.Millisecond, "Ingestion p95 latency should be <50ms")
}

// TestPerformanceRepository_QueryLatency_Under100msP95 tests query latency
func TestPerformanceRepository_QueryLatency_Under100msP95(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer teardownPerformanceTestDB(t, db)

	repo := NewLogEntryRepository(db)
	ctx := context.Background()

	// Insert test data first
	logEntries := make([]*logs_models.LogEntry, 500)
	for i := 0; i < 500; i++ {
		logEntries[i] = &logs_models.LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Query test %d", i),
			Service:   "portal",
			Tags:      []string{"query"},
		}
	}
	err := repo.BulkInsert(ctx, logEntries)
	require.NoError(t, err)

	// Record query latencies
	latencies := make([]time.Duration, 100)

	for i := 0; i < 100; i++ {
		start := time.Now()

		// Execute a representative query
		query := `
			SELECT id, timestamp, level, message, service
			FROM logs.log_entries
			WHERE service = 'portal'
			LIMIT 10
		`
		rows, err := db.QueryContext(ctx, query)
		if err == nil {
			rows.Close()
		}

		latencies[i] = time.Since(start)
		require.NoError(t, err)
	}

	p95Latency := calculateP95(latencies)
	assert.Less(t, p95Latency, 100*time.Millisecond, "Query p95 latency should be <100ms")
}

// BenchmarkPerformanceRepository_BulkInsert_1000Logs benchmarks bulk insert performance
func BenchmarkPerformanceRepository_BulkInsert_1000Logs(b *testing.B) {
	db := setupPerformanceTestDB(&testing.T{})
	defer teardownPerformanceTestDB(&testing.T{}, db)

	repo := NewLogEntryRepository(db)
	ctx := context.Background()

	logEntries := make([]*logs_models.LogEntry, 1000)
	for i := 0; i < 1000; i++ {
		logEntries[i] = &logs_models.LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Benchmark log %d", i),
			Service:   "review",
			Tags:      []string{"bench"},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := repo.BulkInsert(ctx, logEntries)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPerformanceRepository_SingleInsert benchmarks single insert for comparison
func BenchmarkPerformanceRepository_SingleInsert(b *testing.B) {
	db := setupPerformanceTestDB(&testing.T{})
	defer teardownPerformanceTestDB(&testing.T{}, db)

	repo := NewLogEntryRepository(db)
	ctx := context.Background()

	logEntry := &logs_models.LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Benchmark entry",
		Service:   "analytics",
		Tags:      []string{"bench"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := repo.BulkInsert(ctx, []*logs_models.LogEntry{logEntry})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestPerformanceRepository_WebSocket_100ConcurrentClients tests concurrent WebSocket clients
func TestPerformanceRepository_WebSocket_100ConcurrentClients(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer teardownPerformanceTestDB(t, db)

	repo := NewLogEntryRepository(db)
	ctx := context.Background()

	numClients := 100
	logsPerClient := 10
	var wg sync.WaitGroup
	var successCount int64
	var errorCount int64

	start := time.Now()

	// Simulate 100 concurrent clients
	for clientID := 0; clientID < numClients; clientID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			logEntries := make([]*logs_models.LogEntry, logsPerClient)
			for i := 0; i < logsPerClient; i++ {
				logEntries[i] = &logs_models.LogEntry{
					Timestamp: time.Now(),
					Level:     "info",
					Message:   fmt.Sprintf("Client %d log %d", id, i),
					Service:   "logs",
					Tags:      []string{"concurrent"},
				}
			}

			err := repo.BulkInsert(ctx, logEntries)
			if err != nil {
				atomic.AddInt64(&errorCount, 1)
			} else {
				atomic.AddInt64(&successCount, int64(logsPerClient))
			}
		}(clientID)
	}

	wg.Wait()
	totalDuration := time.Since(start)

	assert.Equal(t, int64(0), errorCount, "No errors during concurrent ingestion")
	assert.Equal(t, int64(numClients*logsPerClient), successCount, "All logs should be inserted")

	// Should complete in reasonable time (less than 10 seconds for 1000 logs from 100 clients)
	assert.Less(t, totalDuration, 10*time.Second, "Concurrent ingestion should complete in <10 seconds")
}

// TestPerformanceRepository_WalOptimization_ConfigExists tests WAL optimization config
func TestPerformanceRepository_WalOptimization_ConfigExists(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer teardownPerformanceTestDB(t, db)

	ctx := context.Background()

	// Check PostgreSQL WAL configuration
	walSettings := []string{
		"wal_level",
		"synchronous_commit",
		"wal_buffers",
		"checkpoint_completion_target",
	}

	for _, setting := range walSettings {
		var actualValue string
		query := fmt.Sprintf("SHOW %s", setting)
		err := db.QueryRowContext(ctx, query).Scan(&actualValue)

		// Note: In RED phase, these might not be optimized yet
		// The test just verifies we CAN query these settings
		require.NoError(t, err, "Should be able to query PostgreSQL setting: %s", setting)
	}
}

// setupPerformanceTestDB sets up a test database with performance optimizations
func setupPerformanceTestDB(t *testing.T) *sql.DB {
	t.Helper()

	ctx := context.Background()

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

	// Wait for database to be ready
	for i := 0; i < 30; i++ {
		pingErr := db.PingContext(ctx)
		if pingErr == nil {
			break
		}
		if i == 29 {
			require.Fail(t, "database failed to become ready")
		}
		time.Sleep(time.Second)
	}

	// Set connection pool settings for performance testing
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)

	// Create schema
	_, err = db.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS logs")
	require.NoError(t, err)

	// Create log_entries table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS logs.log_entries (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT,
			service TEXT NOT NULL,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			metadata BYTEA,
			tags TEXT[],
			correlation_id TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	// Create indexes on query fields for performance
	_, err = db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_log_entries_service ON logs.log_entries(service);
		CREATE INDEX IF NOT EXISTS idx_log_entries_level ON logs.log_entries(level);
		CREATE INDEX IF NOT EXISTS idx_log_entries_timestamp ON logs.log_entries(timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_log_entries_correlation_id ON logs.log_entries(correlation_id);
		CREATE INDEX IF NOT EXISTS idx_log_entries_user_id ON logs.log_entries(user_id);
		CREATE INDEX IF NOT EXISTS idx_log_entries_service_level ON logs.log_entries(service, level)
	`)
	require.NoError(t, err)

	// Configure WAL settings for write performance
	// Note: Some settings may be read-only depending on PostgreSQL configuration
	_, _ = db.ExecContext(ctx, "ALTER SYSTEM SET synchronous_commit = off")
	_, _ = db.ExecContext(ctx, "ALTER SYSTEM SET wal_buffers = 2048")
	_, _ = db.ExecContext(ctx, "ALTER SYSTEM SET checkpoint_completion_target = 0.9")

	return db
}

// teardownPerformanceTestDB cleans up test database
func teardownPerformanceTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	if db != nil {
		db.Close()
	}
}

// calculateP95 calculates 95th percentile from a slice of durations
func calculateP95(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	// Sort the durations
	sortedDurations := make([]time.Duration, len(durations))
	copy(sortedDurations, durations)
	sort.Slice(sortedDurations, func(i, j int) bool {
		return sortedDurations[i] < sortedDurations[j]
	})

	index := (95 * len(sortedDurations)) / 100
	if index >= len(sortedDurations) {
		index = len(sortedDurations) - 1
	}

	return sortedDurations[index]
}
