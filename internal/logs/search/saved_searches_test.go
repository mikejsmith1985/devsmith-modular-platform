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

// SavedSearch represents a saved search query.
type SavedSearch struct {
	ID        int64
	Name      string
	QueryStr  string
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    int64
	IsPublic  bool
}

// TestIntegration_SavedSearches_Create tests creating a saved search.
func TestIntegration_SavedSearches_Create(t *testing.T) {
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
	err = createSavedSearchesTable(ctx, pg.db)
	require.NoError(t, err)

	// Create a saved search
	var id int64
	err = pg.db.QueryRowContext(ctx,
		`INSERT INTO saved_searches (name, query, user_id, is_public, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		"Auth Errors", "service:auth AND level:error", 123, false, time.Now(), time.Now(),
	).Scan(&id)

	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

// TestIntegration_SavedSearches_Read tests retrieving a saved search.
func TestIntegration_SavedSearches_Read(t *testing.T) {
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
	err = createSavedSearchesTable(ctx, pg.db)
	require.NoError(t, err)

	// Insert a saved search
	now := time.Now()
	var id int64
	pg.db.QueryRowContext(ctx,
		`INSERT INTO saved_searches (name, query, user_id, is_public, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		"Auth Errors", "service:auth AND level:error", 123, false, now, now,
	).Scan(&id)

	// Retrieve the saved search
	var name, query string
	var userID int64
	err = pg.db.QueryRowContext(ctx,
		`SELECT name, query, user_id FROM saved_searches WHERE id = $1`, id,
	).Scan(&name, &query, &userID)

	require.NoError(t, err)
	assert.Equal(t, "Auth Errors", name)
	assert.Equal(t, "service:auth AND level:error", query)
	assert.Equal(t, int64(123), userID)
}

// TestIntegration_SavedSearches_Update tests updating a saved search.
func TestIntegration_SavedSearches_Update(t *testing.T) {
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
	err = createSavedSearchesTable(ctx, pg.db)
	require.NoError(t, err)

	// Insert a saved search
	now := time.Now()
	var id int64
	pg.db.QueryRowContext(ctx,
		`INSERT INTO saved_searches (name, query, user_id, is_public, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		"Auth Errors", "service:auth AND level:error", 123, false, now, now,
	).Scan(&id)

	// Update the saved search
	later := now.Add(1 * time.Second)
	_, err = pg.db.ExecContext(ctx,
		`UPDATE saved_searches SET query = $1, updated_at = $2 WHERE id = $3`,
		"service:auth AND level:warn", later, id,
	)
	require.NoError(t, err)

	// Verify update
	var query string
	pg.db.QueryRowContext(ctx,
		`SELECT query FROM saved_searches WHERE id = $1`, id,
	).Scan(&query)

	assert.Equal(t, "service:auth AND level:warn", query)
}

// TestIntegration_SavedSearches_Delete tests deleting a saved search.
func TestIntegration_SavedSearches_Delete(t *testing.T) {
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
	err = createSavedSearchesTable(ctx, pg.db)
	require.NoError(t, err)

	// Insert a saved search
	now := time.Now()
	var id int64
	pg.db.QueryRowContext(ctx,
		`INSERT INTO saved_searches (name, query, user_id, is_public, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		"Auth Errors", "service:auth", 123, false, now, now,
	).Scan(&id)

	// Delete the saved search
	result, err := pg.db.ExecContext(ctx,
		`DELETE FROM saved_searches WHERE id = $1`, id,
	)
	require.NoError(t, err)

	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)
}

// TestIntegration_SavedSearches_ListByUser tests listing searches for a user.
func TestIntegration_SavedSearches_ListByUser(t *testing.T) {
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
	err = createSavedSearchesTable(ctx, pg.db)
	require.NoError(t, err)

	now := time.Now()

	// Insert multiple searches for user 123
	for i := 0; i < 3; i++ {
		pg.db.ExecContext(ctx,
			`INSERT INTO saved_searches (name, query, user_id, is_public, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			"Search "+string(rune(i)), "service:test", 123, false, now, now,
		)
	}

	// Insert searches for user 456
	for i := 0; i < 2; i++ {
		pg.db.ExecContext(ctx,
			`INSERT INTO saved_searches (name, query, user_id, is_public, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			"Other "+string(rune(i)), "service:other", 456, false, now, now,
		)
	}

	// List searches for user 123
	var count int
	pg.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM saved_searches WHERE user_id = $1`, 123,
	).Scan(&count)

	assert.Equal(t, 3, count)
}

// TestIntegration_SavedSearches_PublicQueries tests listing public searches.
func TestIntegration_SavedSearches_PublicQueries(t *testing.T) {
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
	err = createSavedSearchesTable(ctx, pg.db)
	require.NoError(t, err)

	now := time.Now()

	// Insert public and private searches
	pg.db.ExecContext(ctx,
		`INSERT INTO saved_searches (name, query, user_id, is_public, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		"Public Search", "service:auth", 123, true, now, now,
	)

	pg.db.ExecContext(ctx,
		`INSERT INTO saved_searches (name, query, user_id, is_public, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		"Private Search", "service:portal", 123, false, now, now,
	)

	// Count public searches
	var count int
	pg.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM saved_searches WHERE is_public = true`,
	).Scan(&count)

	assert.Equal(t, 1, count)
}

// TestIntegration_SavedSearches_DuplicateName tests handling duplicate names.
func TestIntegration_SavedSearches_DuplicateName(t *testing.T) {
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
	err = createSavedSearchesTable(ctx, pg.db)
	require.NoError(t, err)

	now := time.Now()

	// Insert first search
	var id1 int64
	pg.db.QueryRowContext(ctx,
		`INSERT INTO saved_searches (name, query, user_id, is_public, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		"Same Name", "service:auth", 123, false, now, now,
	).Scan(&id1)

	// Insert second search with same name (different user or query is allowed)
	var id2 int64
	pg.db.QueryRowContext(ctx,
		`INSERT INTO saved_searches (name, query, user_id, is_public, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		"Same Name", "service:portal", 123, false, now, now,
	).Scan(&id2)

	assert.NotEqual(t, id1, id2)
}

// Helper function to create the saved_searches table
func createSavedSearchesTable(ctx context.Context, db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS saved_searches (
		id BIGSERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		query TEXT NOT NULL,
		user_id BIGINT NOT NULL,
		is_public BOOLEAN DEFAULT false,
		created_at TIMESTAMPTZ NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_saved_searches_user_id ON saved_searches(user_id);
	CREATE INDEX IF NOT EXISTS idx_saved_searches_is_public ON saved_searches(is_public);
	`

	_, err := db.ExecContext(ctx, schema)
	return err
}
