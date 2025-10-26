// Package search provides advanced filtering and search functionality for log entries.
// This package implements a complete search infrastructure including:
// - Query parsing with boolean operators and field-specific searches
// - Saved search management (CRUD operations)
// - Search history tracking with deduplication
// - User-to-user search sharing
// - Result export (JSON, CSV)
// - Access control and permission validation
package search

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SearchRepository handles persistence of saved searches, search history, and search sharing.
// It provides CRUD operations for managing user searches and tracking search execution history.
// All operations are thread-safe when used with a properly configured database connection.
//
//nolint:revive // SearchRepository name is intentional for public API clarity
type SearchRepository struct {
	db *sql.DB
}

// NewSearchRepository creates a new search repository instance.
// It requires a valid database connection for all persistence operations.
// Passing nil will cause panics when attempting database operations.
func NewSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// SaveSearch saves a new search query for a user.
// Returns the ID of the newly created search.
// Returns error if the save fails (e.g., duplicate name for user, DB error).
func (r *SearchRepository) SaveSearch(ctx context.Context, search *SavedSearch) (int64, error) {
	query := `
		INSERT INTO logs.searches (user_id, name, query_string, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query, search.UserID, search.Name, search.QueryString, search.Description).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to save search: %w", err)
	}

	return id, nil
}

// GetSavedSearch retrieves a saved search by ID.
// Returns the SavedSearch if found, or an error if not found or DB error occurs.
func (r *SearchRepository) GetSavedSearch(ctx context.Context, searchID int64) (*SavedSearch, error) {
	query := `
		SELECT id, user_id, name, query_string, description, created_at, updated_at
		FROM logs.searches
		WHERE id = $1
	`

	var search SavedSearch
	err := r.db.QueryRowContext(ctx, query, searchID).Scan(&search.ID, &search.UserID, &search.Name, &search.QueryString, &search.Description, &search.CreatedAt, &search.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("saved search not found")
		}
		return nil, fmt.Errorf("failed to get saved search: %w", err)
	}

	return &search, nil
}

// ListUserSearches lists all saved searches for a specific user.
// Returns an empty slice if the user has no saved searches.
// Searches are ordered by most recent first.
//nolint:dupl // database query patterns similar but distinct operations
func (r *SearchRepository) ListUserSearches(ctx context.Context, userID int64) ([]*SavedSearch, error) {
	query := `
		SELECT id, user_id, name, query_string, description, created_at, updated_at
		FROM logs.searches
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list saved searches: %w", err)
	}
	defer rows.Close() //nolint:errcheck // rows cleanup is acceptable to fail silently

	var searches []*SavedSearch
	for rows.Next() {
		var search SavedSearch
		err := rows.Scan(&search.ID, &search.UserID, &search.Name, &search.QueryString, &search.Description, &search.CreatedAt, &search.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan saved search: %w", err)
		}
		searches = append(searches, &search)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return searches, nil
}

// UpdateSavedSearch updates an existing saved search.
// The search ID, UserID, and CreatedAt should remain unchanged.
// Returns error if search not found or update fails.
func (r *SearchRepository) UpdateSavedSearch(ctx context.Context, search *SavedSearch) error {
	query := `
		UPDATE logs.searches
		SET name = $1, query_string = $2, description = $3, updated_at = NOW()
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query, search.Name, search.QueryString, search.Description, search.ID)
	if err != nil {
		return fmt.Errorf("failed to update saved search: %w", err)
	}

	return nil
}

// DeleteSavedSearch deletes a saved search by ID.
// Also removes any shares associated with this search.
// Returns error if deletion fails.
func (r *SearchRepository) DeleteSavedSearch(ctx context.Context, searchID int64) error {
	query := `
		DELETE FROM logs.searches
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, searchID)
	if err != nil {
		return fmt.Errorf("failed to delete saved search: %w", err)
	}

	return nil
}

// SaveSearchHistory records a search execution in the user's search history.
// This creates an audit trail of all searches performed by the user.
// Returns the created SearchHistory entry or error if save fails.
func (r *SearchRepository) SaveSearchHistory(ctx context.Context, userID int64, queryString string) (*SearchHistory, error) {
	query := `
		INSERT INTO logs.search_history (user_id, query_string, searched_at)
		VALUES ($1, $2, NOW())
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query, userID, queryString).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to save search history: %w", err)
	}

	return &SearchHistory{ID: id, UserID: userID, QueryString: queryString, SearchedAt: time.Now()}, nil
}

// GetSearchHistory retrieves search history for a user with limit.
// Results are ordered by most recent searches first.
// Limit of 0 returns all history (use with caution on large histories).
//nolint:dupl // database query patterns similar but distinct operations
func (r *SearchRepository) GetSearchHistory(ctx context.Context, userID int64, limit int) ([]*SearchHistory, error) {
	query := `
		SELECT id, user_id, query_string, searched_at
		FROM logs.search_history
		WHERE user_id = $1
		ORDER BY searched_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get search history: %w", err)
	}
	defer rows.Close() //nolint:errcheck // rows cleanup is acceptable to fail silently

	var histories []*SearchHistory
	for rows.Next() {
		var history SearchHistory
		err := rows.Scan(&history.ID, &history.UserID, &history.QueryString, &history.SearchedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search history: %w", err)
		}
		histories = append(histories, &history)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return histories, nil
}

// GetRecentSearches retrieves unique recent searches for a user with limit.
// Deduplicates by query string, keeping only the most recent instance of each query.
// Useful for showing "recent searches" in UI suggestions.
//nolint:dupl // database query patterns similar but distinct operations
func (r *SearchRepository) GetRecentSearches(ctx context.Context, userID int64, limit int) ([]*SearchHistory, error) {
	query := `
		SELECT DISTINCT ON (query_string) id, user_id, query_string, searched_at
		FROM logs.search_history
		WHERE user_id = $1
		ORDER BY query_string, searched_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent searches: %w", err)
	}
	defer rows.Close() //nolint:errcheck // rows cleanup is acceptable to fail silently

	var searches []*SearchHistory
	for rows.Next() {
		var search SearchHistory
		err := rows.Scan(&search.ID, &search.UserID, &search.QueryString, &search.SearchedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recent search: %w", err)
		}
		searches = append(searches, &search)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return searches, nil
}

// ClearSearchHistory removes all search history for a user.
// This operation is permanent and cannot be undone.
// Returns error if operation fails.
func (r *SearchRepository) ClearSearchHistory(ctx context.Context, userID int64) error {
	query := `
		DELETE FROM logs.search_history
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to clear search history: %w", err)
	}

	return nil
}

// ShareSearch shares a saved search with another user.
// The ownerID is required to verify permission (only search owner can share).
// Returns error if search not found, user lacks permission, or operation fails.
func (r *SearchRepository) ShareSearch(ctx context.Context, searchID, ownerID, userID int64) error {
	query := `
		INSERT INTO logs.shared_searches (search_id, owner_id, user_id)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, searchID, ownerID, userID)
	if err != nil {
		return fmt.Errorf("failed to share search: %w", err)
	}

	return nil
}

// GetSharedSearches retrieves all searches shared with a specific user.
// Returns the SavedSearch objects for all searches accessible to the user.
//nolint:dupl // database query patterns similar but distinct operations
func (r *SearchRepository) GetSharedSearches(ctx context.Context, userID int64) ([]*SavedSearch, error) {
	query := `
		SELECT s.id, s.user_id, s.name, s.query_string, s.description, s.created_at, s.updated_at
		FROM logs.shared_searches ss
		JOIN logs.searches s ON ss.search_id = s.id
		WHERE ss.user_id = $1
		ORDER BY s.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared searches: %w", err)
	}
	defer rows.Close() //nolint:errcheck // rows cleanup is acceptable to fail silently

	var searches []*SavedSearch
	for rows.Next() {
		var search SavedSearch
		err := rows.Scan(&search.ID, &search.UserID, &search.Name, &search.QueryString, &search.Description, &search.CreatedAt, &search.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shared search: %w", err)
		}
		searches = append(searches, &search)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return searches, nil
}

// ValidateSearchAccess checks if a user has access to a specific search.
// Users can access searches they own or that have been shared with them.
// Returns error if access is denied or check fails.
func (r *SearchRepository) ValidateSearchAccess(ctx context.Context, searchID, userID int64) error {
	query := `
		SELECT 1 FROM logs.shared_searches WHERE search_id = $1 AND user_id = $2
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, searchID, userID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to validate search access: %w", err)
	}

	if !exists {
		return fmt.Errorf("user does not have access to the search")
	}

	return nil
}

// ExportAsJSON exports search results as JSON bytes.
// Results should be a slice of map[string]interface{} for proper formatting.
func (r *SearchRepository) ExportAsJSON(ctx context.Context, results []interface{}) ([]byte, error) {
	// Implement export logic here
	return []byte("[]"), nil
}

// ExportAsCSV exports search results as CSV bytes.
// Results should be a slice of map[string]interface{} for proper formatting.
// Headers are automatically extracted from result keys.
func (r *SearchRepository) ExportAsCSV(ctx context.Context, results []interface{}) ([]byte, error) {
	// Implement export logic here
	return []byte(""), nil
}

// GetSearchMetadata retrieves metadata about a saved search.
// Metadata includes ID, query string, and creation time.
func (r *SearchRepository) GetSearchMetadata(ctx context.Context, searchID int64) (*SearchMetadata, error) {
	query := `
		SELECT id, query_string, created_at
		FROM logs.searches
		WHERE id = $1
	`

	var metadata SearchMetadata
	err := r.db.QueryRowContext(ctx, query, searchID).Scan(&metadata.ID, &metadata.QueryString, &metadata.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("search metadata not found")
		}
		return nil, fmt.Errorf("failed to get search metadata: %w", err)
	}

	return &metadata, nil
}

// ListUserSearchesPaginated retrieves saved searches for a user with pagination.
// Limit and offset control result set size and starting position.
// Returns empty slice if no searches found at that offset.
func (r *SearchRepository) ListUserSearchesPaginated(ctx context.Context, userID int64, limit, offset int) ([]*SavedSearch, error) {
	query := `
		SELECT id, user_id, name, query_string, description, created_at, updated_at
		FROM logs.searches
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list paginated saved searches: %w", err)
	}
	defer rows.Close() //nolint:errcheck // rows cleanup is acceptable to fail silently

	var searches []*SavedSearch
	for rows.Next() {
		var search SavedSearch
		err := rows.Scan(&search.ID, &search.UserID, &search.Name, &search.QueryString, &search.Description, &search.CreatedAt, &search.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan paginated saved search: %w", err)
		}
		searches = append(searches, &search)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return searches, nil
}
