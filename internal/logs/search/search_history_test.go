// +build integration

package search

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SearchHistory represents a user's search history entry.
type SearchHistory struct {
	ID        int64
	UserID    int64
	Query     string
	Results   int
	ExecutedAt time.Time
}

// TestIntegration_SearchHistory_Record tests recording a search.
func TestIntegration_SearchHistory_Record(t *testing.T) {
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
	err = createSearchHistoryTable(ctx, pg.db)
	require.NoError(t, err)

	// Record a search
	var id int64
	err = pg.db.QueryRowContext(ctx,
		`INSERT INTO search_history (user_id, query, results, executed_at)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		123, "service:auth AND level:error", 42, time.Now(),
	).Scan(&id)

	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

// TestIntegration_SearchHistory_ListByUser tests retrieving history for a user.
func TestIntegration_SearchHistory_ListByUser(t *testing.T) {
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
	err = createSearchHistoryTable(ctx, pg.db)
	require.NoError(t, err)

	now := time.Now()

	// Insert history for user 123
	for i := 0; i < 5; i++ {
		pg.db.ExecContext(ctx,
			`INSERT INTO search_history (user_id, query, results, executed_at)
			 VALUES ($1, $2, $3, $4)`,
			123, "query "+string(rune(i)), i*10, now,
		)
	}

	// Insert history for user 456
	for i := 0; i < 3; i++ {
		pg.db.ExecContext(ctx,
			`INSERT INTO search_history (user_id, query, results, executed_at)
			 VALUES ($1, $2, $3, $4)`,
			456, "query "+string(rune(i)), i*5, now,
		)
	}

	// Count history for user 123
	var count int
	pg.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM search_history WHERE user_id = $1`, 123,
	).Scan(&count)

	assert.Equal(t, 5, count)
}

// TestIntegration_SearchHistory_TimeRange tests filtering by date range.
func TestIntegration_SearchHistory_TimeRange(t *testing.T) {
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
	err = createSearchHistoryTable(ctx, pg.db)
	require.NoError(t, err)

	now := time.Now()
	pastTime := now.Add(-24 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	// Insert old and new searches
	pg.db.ExecContext(ctx,
		`INSERT INTO search_history (user_id, query, results, executed_at)
		 VALUES ($1, $2, $3, $4)`,
		123, "old query", 10, pastTime,
	)

	pg.db.ExecContext(ctx,
		`INSERT INTO search_history (user_id, query, results, executed_at)
		 VALUES ($1, $2, $3, $4)`,
		123, "new query", 20, futureTime,
	)

	// Query within range
	var count int
	pg.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM search_history
		 WHERE user_id = $1 AND executed_at >= $2 AND executed_at <= $3`,
		123, pastTime, futureTime,
	).Scan(&count)

	assert.Equal(t, 2, count)
}

// TestIntegration_SearchHistory_DeleteOldRecords tests cleanup of old records.
func TestIntegration_SearchHistory_DeleteOldRecords(t *testing.T) {
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
	err = createSearchHistoryTable(ctx, pg.db)
	require.NoError(t, err)

	now := time.Now()
	cutoff := now.Add(-7 * 24 * time.Hour) // 7 days ago

	// Insert old and recent records
	pg.db.ExecContext(ctx,
		`INSERT INTO search_history (user_id, query, results, executed_at)
		 VALUES ($1, $2, $3, $4)`,
		123, "very old", 5, cutoff.Add(-1*time.Hour),
	)

	pg.db.ExecContext(ctx,
		`INSERT INTO search_history (user_id, query, results, executed_at)
		 VALUES ($1, $2, $3, $4)`,
		123, "recent", 15, now,
	)

	// Delete records older than cutoff
	result, err := pg.db.ExecContext(ctx,
		`DELETE FROM search_history WHERE executed_at < $1`, cutoff,
	)
	require.NoError(t, err)

	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, rowsAffected, int64(1))
}

// TestIntegration_SearchHistory_ResultsCount tests recording result counts.
func TestIntegration_SearchHistory_ResultsCount(t *testing.T) {
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
	err = createSearchHistoryTable(ctx, pg.db)
	require.NoError(t, err)

	now := time.Now()

	// Insert searches with different result counts
	pg.db.ExecContext(ctx,
		`INSERT INTO search_history (user_id, query, results, executed_at)
		 VALUES ($1, $2, $3, $4)`,
		123, "high results", 1000, now,
	)

	pg.db.ExecContext(ctx,
		`INSERT INTO search_history (user_id, query, results, executed_at)
		 VALUES ($1, $2, $3, $4)`,
		123, "low results", 5, now,
	)

	// Aggregate results
	var totalResults int
	pg.db.QueryRowContext(ctx,
		`SELECT SUM(results) FROM search_history WHERE user_id = $1`, 123,
	).Scan(&totalResults)

	assert.Equal(t, 1005, totalResults)
}

// TestIntegration_SearchHistory_MostFrequent tests finding frequently searched queries.
func TestIntegration_SearchHistory_MostFrequent(t *testing.T) {
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
	err = createSearchHistoryTable(ctx, pg.db)
	require.NoError(t, err)

	now := time.Now()

	// Insert repeated queries
	for i := 0; i < 5; i++ {
		pg.db.ExecContext(ctx,
			`INSERT INTO search_history (user_id, query, results, executed_at)
			 VALUES ($1, $2, $3, $4)`,
			123, "favorite query", i*10, now,
		)
	}

	for i := 0; i < 2; i++ {
		pg.db.ExecContext(ctx,
			`INSERT INTO search_history (user_id, query, results, executed_at)
			 VALUES ($1, $2, $3, $4)`,
			123, "other query", i*5, now,
		)
	}

	// Find most frequent query
	var query string
	var count int
	pg.db.QueryRowContext(ctx,
		`SELECT query, COUNT(*) as cnt FROM search_history
		 WHERE user_id = $1
		 GROUP BY query
		 ORDER BY cnt DESC
		 LIMIT 1`,
		123,
	).Scan(&query, &count)

	assert.Equal(t, "favorite query", query)
	assert.Equal(t, 5, count)
}

// TestIntegration_SearchHistory_RecentQueries tests pagination of recent searches.
func TestIntegration_SearchHistory_RecentQueries(t *testing.T) {
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
	err = createSearchHistoryTable(ctx, pg.db)
	require.NoError(t, err)

	now := time.Now()

	// Insert 10 searches over time
	for i := 0; i < 10; i++ {
		pg.db.ExecContext(ctx,
			`INSERT INTO search_history (user_id, query, results, executed_at)
			 VALUES ($1, $2, $3, $4)`,
			123, "query "+string(rune(i)), i*10, now.Add(-time.Duration(i)*time.Hour),
		)
	}

	// Get last 5 searches
	var count int
	pg.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM (
			SELECT * FROM search_history
			WHERE user_id = $1
			ORDER BY executed_at DESC
			LIMIT 5
		) AS recent`,
		123,
	).Scan(&count)

	assert.Equal(t, 5, count)
}

// Helper function to create the search_history table
func createSearchHistoryTable(ctx context.Context, db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS search_history (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		query TEXT NOT NULL,
		results INT DEFAULT 0,
		executed_at TIMESTAMPTZ NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_search_history_user_id ON search_history(user_id);
	CREATE INDEX IF NOT EXISTS idx_search_history_executed_at ON search_history(executed_at DESC);
	`

	_, err := db.ExecContext(ctx, schema)
	return err
}
