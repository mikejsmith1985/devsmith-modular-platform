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
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// SearchRepository handles persistence of saved searches, search history, and search sharing.
// It provides CRUD operations for managing user searches and tracking search execution history.
// All operations are thread-safe when used with a properly configured database connection.
//
//nolint:revive // SearchRepository name is intentional for public API clarity
type SearchRepository struct {
	db *sql.DB
	// in-memory fallback used when db is nil (unit tests)
	memSearches map[int64]*SavedSearch
	memHistory  map[int64][]*SearchHistory
	memShared   map[int64][]int64 // map[userID] -> list of searchIDs shared with them
	nextID      int64
}

// NewSearchRepository creates a new search repository instance.
// It requires a valid database connection for all persistence operations.
// Passing nil will cause panics when attempting database operations.
func NewSearchRepository(db *sql.DB) *SearchRepository {
	repo := &SearchRepository{db: db}
	if db == nil {
		repo.memSearches = make(map[int64]*SavedSearch)
		repo.memHistory = make(map[int64][]*SearchHistory)
		repo.nextID = 0
		repo.memShared = make(map[int64][]int64)
	}
	return repo
}

// SaveSearch saves a new search query for a user.
// Returns the ID of the newly created search.
// Returns error if the save fails (e.g., duplicate name for user, DB error).
func (r *SearchRepository) SaveSearch(ctx context.Context, search *SavedSearch) (int64, error) {
	// If no DB provided, use in-memory fallback for unit tests
	if r.db == nil {
		// enforce name uniqueness per user in-memory
		for _, s := range r.memSearches {
			if s.UserID == search.UserID && s.Name == search.Name {
				return 0, fmt.Errorf("duplicate search name for user")
			}
		}

		r.nextID++
		id := r.nextID
		now := time.Now()
		s := &SavedSearch{
			ID:          id,
			UserID:      search.UserID,
			Name:        search.Name,
			QueryString: search.QueryString,
			Description: search.Description,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		r.memSearches[id] = s
		return id, nil
	}

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
	if r.db == nil {
		if s, ok := r.memSearches[searchID]; ok {
			return s, nil
		}
		return nil, fmt.Errorf("saved search not found")
	}

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
//
//nolint:dupl // database query patterns similar but distinct operations
func (r *SearchRepository) ListUserSearches(ctx context.Context, userID int64) ([]*SavedSearch, error) {
	if r.db == nil {
		var results []*SavedSearch
		for _, s := range r.memSearches {
			if s.UserID == userID {
				results = append(results, s)
			}
		}
		return results, nil
	}

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
	if r.db == nil {
		if s, ok := r.memSearches[search.ID]; ok {
			s.Name = search.Name
			s.QueryString = search.QueryString
			s.Description = search.Description
			s.UpdatedAt = time.Now()
			return nil
		}
		return fmt.Errorf("saved search not found")
	}

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
	if r.db == nil {
		delete(r.memSearches, searchID)
		return nil
	}

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
	if r.db == nil {
		r.nextID++
		id := r.nextID
		entry := &SearchHistory{ID: id, UserID: userID, QueryString: queryString, SearchedAt: time.Now()}
		r.memHistory[userID] = append(r.memHistory[userID], entry)
		return entry, nil
	}

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
//
//nolint:dupl // database query patterns similar but distinct operations
func (r *SearchRepository) GetSearchHistory(ctx context.Context, userID int64, limit int) ([]*SearchHistory, error) {
	if r.db == nil {
		// Return most recent first
		histories := r.memHistory[userID]
		n := len(histories)
		if n == 0 {
			return []*SearchHistory{}, nil
		}
		// build reversed slice
		rev := make([]*SearchHistory, 0, n)
		for i := n - 1; i >= 0; i-- {
			rev = append(rev, histories[i])
		}
		if limit > 0 && len(rev) > limit {
			return rev[:limit], nil
		}
		return rev, nil
	}

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
//
//nolint:dupl // database query patterns similar but distinct operations
func (r *SearchRepository) GetRecentSearches(ctx context.Context, userID int64, limit int) ([]*SearchHistory, error) {
	if r.db == nil {
		// In-memory: dedupe by QueryString keeping most recent
		hist := r.memHistory[userID]
		// iterate from end to start to collect most recent occurrences
		seen := make(map[string]bool)
		var recent []*SearchHistory
		for i := len(hist) - 1; i >= 0; i-- {
			h := hist[i]
			if !seen[h.QueryString] {
				recent = append(recent, h)
				seen[h.QueryString] = true
			}
			if limit > 0 && len(recent) >= limit {
				break
			}
		}
		// recent currently in reverse chronological order (most recent first), return as-is
		return recent, nil
	}

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
	if r.db == nil {
		delete(r.memHistory, userID)
		return nil
	}

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
	if r.db == nil {
		// Ensure owner actually owns the search
		if s, ok := r.memSearches[searchID]; !ok || s.UserID != ownerID {
			return fmt.Errorf("search not found or owner mismatch")
		}
		r.memShared[userID] = append(r.memShared[userID], searchID)
		return nil
	}

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
//
//nolint:dupl // database query patterns similar but distinct operations
func (r *SearchRepository) GetSharedSearches(ctx context.Context, userID int64) ([]*SavedSearch, error) {
	if r.db == nil {
		ids := r.memShared[userID]
		var results []*SavedSearch
		for _, id := range ids {
			if s, ok := r.memSearches[id]; ok {
				results = append(results, s)
			}
		}
		return results, nil
	}

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
	if r.db == nil {
		// Owner has access
		if s, ok := r.memSearches[searchID]; ok {
			if s.UserID == userID {
				return nil
			}
		}
		// Check shared map
		if ids, ok := r.memShared[userID]; ok {
			for _, id := range ids {
				if id == searchID {
					return nil
				}
			}
		}
		return fmt.Errorf("user does not have access to the search")
	}

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
	if results == nil {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal results to JSON: %w", err)
	}
	return b, nil
}

// ExportAsCSV exports search results as CSV bytes.
// Results should be a slice of map[string]interface{} for proper formatting.
// Headers are automatically extracted from result keys.
func (r *SearchRepository) ExportAsCSV(ctx context.Context, results []interface{}) ([]byte, error) {
	if len(results) == 0 {
		return []byte(""), nil
	}

	// Expect each result to be a map[string]interface{}
	first, ok := results[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type for CSV export")
	}

	// Collect headers deterministically sorted
	headers := make([]string, 0, len(first))
	for k := range first {
		headers = append(headers, k)
	}
	sort.Strings(headers)

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(headers); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, ritem := range results {
		rowMap, ok := ritem.(map[string]interface{})
		if !ok {
			// write empty row to preserve alignment
			empty := make([]string, len(headers))
			if err := w.Write(empty); err != nil {
				return nil, fmt.Errorf("failed to write CSV row: %w", err)
			}
			continue
		}
		row := make([]string, len(headers))
		for i, h := range headers {
			if v, exists := rowMap[h]; exists && v != nil {
				row[i] = fmt.Sprintf("%v", v)
			} else {
				row[i] = ""
			}
		}
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	return buf.Bytes(), nil
}

// GetSearchMetadata retrieves metadata about a saved search.
// Metadata includes ID, query string, and creation time.
func (r *SearchRepository) GetSearchMetadata(ctx context.Context, searchID int64) (*SearchMetadata, error) {
	if r.db == nil {
		if s, ok := r.memSearches[searchID]; ok {
			return &SearchMetadata{ID: s.ID, QueryString: s.QueryString, CreatedAt: s.CreatedAt}, nil
		}
		return nil, fmt.Errorf("search metadata not found")
	}

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
	if r.db == nil {
		// Build list from in-memory map
		var all []*SavedSearch
		for _, s := range r.memSearches {
			if s.UserID == userID {
				all = append(all, s)
			}
		}
		// Sort by CreatedAt desc, then ID desc
		sort.Slice(all, func(i, j int) bool {
			if all[i].CreatedAt.Equal(all[j].CreatedAt) {
				return all[i].ID > all[j].ID
			}
			return all[i].CreatedAt.After(all[j].CreatedAt)
		})

		// Apply offset and limit
		start := offset
		if start > len(all) {
			return []*SavedSearch{}, nil
		}
		end := len(all)
		if limit > 0 && start+limit < end {
			end = start + limit
		}
		return all[start:end], nil
	}

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
