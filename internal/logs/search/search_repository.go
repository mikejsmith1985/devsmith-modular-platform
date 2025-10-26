// Package search provides advanced filtering and search functionality for log entries.
//
//nolint:revive // SearchRepository name is intentional for public API clarity
package search

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// SearchRepository handles persistence of saved searches and search history.
//
//nolint:govet // minor field alignment optimization not worth restructuring for this implementation
type SearchRepository struct {
	mu            sync.RWMutex
	searches      map[int64]*SavedSearch
	history       map[int64][]*SearchHistory
	shares        map[int64][]int64 // search_id -> list of user_ids
	nextSearchID  int64
	nextHistoryID int64
	db            *sql.DB
}

// NewSearchRepository creates a new search repository.
func NewSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{
		db:            db,
		searches:      make(map[int64]*SavedSearch),
		history:       make(map[int64][]*SearchHistory),
		shares:        make(map[int64][]int64),
		nextSearchID:  1,
		nextHistoryID: 1,
	}
}

// SaveSearch saves a new search query for a user.
func (r *SearchRepository) SaveSearch(ctx context.Context, search *SavedSearch) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate name per user
	for _, existing := range r.searches {
		if existing.UserID == search.UserID && existing.Name == search.Name {
			return 0, fmt.Errorf("search name already exists for this user")
		}
	}

	id := r.nextSearchID
	r.nextSearchID++

	search.ID = id
	search.CreatedAt = time.Now()
	search.UpdatedAt = time.Now()

	r.searches[id] = search

	return id, nil
}

// GetSavedSearch retrieves a saved search by ID.
func (r *SearchRepository) GetSavedSearch(ctx context.Context, searchID int64) (*SavedSearch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	search, ok := r.searches[searchID]
	if !ok {
		return nil, fmt.Errorf("search not found")
	}

	return search, nil
}

// ListUserSearches lists all saved searches for a user.
func (r *SearchRepository) ListUserSearches(ctx context.Context, userID int64) ([]*SavedSearch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*SavedSearch
	for _, search := range r.searches {
		if search.UserID == userID {
			results = append(results, search)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	return results, nil
}

// UpdateSavedSearch updates an existing saved search.
func (r *SearchRepository) UpdateSavedSearch(ctx context.Context, search *SavedSearch) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.searches[search.ID]
	if !ok {
		return fmt.Errorf("search not found")
	}

	// Check for duplicate name with different searches
	for _, s := range r.searches {
		if s.UserID == search.UserID && s.Name == search.Name && s.ID != search.ID {
			return fmt.Errorf("search name already exists for this user")
		}
	}

	search.UpdatedAt = time.Now()
	search.CreatedAt = existing.CreatedAt // preserve creation time

	r.searches[search.ID] = search

	return nil
}

// DeleteSavedSearch deletes a saved search.
func (r *SearchRepository) DeleteSavedSearch(ctx context.Context, searchID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.searches, searchID)
	delete(r.shares, searchID)

	return nil
}

// SaveSearchHistory records a search execution in history.
func (r *SearchRepository) SaveSearchHistory(ctx context.Context, userID int64, queryString string) (*SearchHistory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.nextHistoryID
	r.nextHistoryID++

	entry := &SearchHistory{
		ID:          id,
		UserID:      userID,
		QueryString: queryString,
		SearchedAt:  time.Now(),
	}

	r.history[userID] = append(r.history[userID], entry)

	return entry, nil
}

// GetSearchHistory retrieves search history for a user.
func (r *SearchRepository) GetSearchHistory(ctx context.Context, userID int64, limit int) ([]*SearchHistory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := r.history[userID]

	// Sort by most recent first
	sorted := make([]*SearchHistory, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SearchedAt.After(sorted[j].SearchedAt)
	})

	if limit > 0 && len(sorted) > limit {
		sorted = sorted[:limit]
	}

	return sorted, nil
}

// GetRecentSearches retrieves unique recent searches for a user.
func (r *SearchRepository) GetRecentSearches(ctx context.Context, userID int64, limit int) ([]*SearchHistory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := r.history[userID]

	// Deduplicate by query string (keep most recent)
	seen := make(map[string]*SearchHistory)
	for _, entry := range entries {
		if existing, ok := seen[entry.QueryString]; !ok {
			seen[entry.QueryString] = entry
		} else if entry.SearchedAt.After(existing.SearchedAt) {
			seen[entry.QueryString] = entry
		}
	}

	// Convert back to list and sort
	var unique []*SearchHistory
	for _, entry := range seen {
		unique = append(unique, entry)
	}

	sort.Slice(unique, func(i, j int) bool {
		return unique[i].SearchedAt.After(unique[j].SearchedAt)
	})

	if limit > 0 && len(unique) > limit {
		unique = unique[:limit]
	}

	return unique, nil
}

// ClearSearchHistory clears all search history for a user.
func (r *SearchRepository) ClearSearchHistory(ctx context.Context, userID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.history, userID)

	return nil
}

// ShareSearch shares a saved search with another user.
func (r *SearchRepository) ShareSearch(ctx context.Context, searchID, ownerID, userID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	search, ok := r.searches[searchID]
	if !ok {
		return fmt.Errorf("search not found")
	}

	if search.UserID != ownerID {
		return fmt.Errorf("cannot share search you don't own")
	}

	// Add to shares list
	shares := r.shares[searchID]
	for _, id := range shares {
		if id == userID {
			return nil // already shared
		}
	}
	r.shares[searchID] = append(shares, userID)

	return nil
}

// GetSharedSearches retrieves searches shared with a user.
func (r *SearchRepository) GetSharedSearches(ctx context.Context, userID int64) ([]*SavedSearch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*SavedSearch
	for searchID, sharedWithIDs := range r.shares {
		for _, id := range sharedWithIDs {
			if id == userID {
				if search, ok := r.searches[searchID]; ok {
					results = append(results, search)
				}
				break
			}
		}
	}

	return results, nil
}

// ValidateSearchAccess checks if a user can access a search.
func (r *SearchRepository) ValidateSearchAccess(ctx context.Context, searchID, userID int64) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	search, ok := r.searches[searchID]
	if !ok {
		return fmt.Errorf("search not found")
	}

	// Owner can always access
	if search.UserID == userID {
		return nil
	}

	// Check if shared
	if shares, ok := r.shares[searchID]; ok {
		for _, id := range shares {
			if id == userID {
				return nil
			}
		}
	}

	return fmt.Errorf("access denied")
}

// ExportAsJSON exports results as JSON bytes.
func (r *SearchRepository) ExportAsJSON(ctx context.Context, results []interface{}) ([]byte, error) {
	jsonBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return jsonBytes, nil
}

// ExportAsCSV exports results as CSV bytes.
func (r *SearchRepository) ExportAsCSV(ctx context.Context, results []interface{}) ([]byte, error) {
	if len(results) == 0 {
		return []byte(""), nil
	}

	// Extract headers from first result
	var headers []string
	var rows [][]string

	for _, result := range results {
		if m, ok := result.(map[string]interface{}); ok {
			if len(headers) == 0 {
				// First row: extract headers
				for k := range m {
					headers = append(headers, k)
				}
				sort.Strings(headers)
			}

			// Add data row
			var row []string
			for _, h := range headers {
				if v, ok := m[h]; ok {
					row = append(row, fmt.Sprintf("%v", v))
				} else {
					row = append(row, "")
				}
			}
			rows = append(rows, row)
		}
	}

	// Build CSV
	var csv strings.Builder
	csv.WriteString(strings.Join(headers, ",") + "\n")
	for _, row := range rows {
		csv.WriteString(strings.Join(row, ",") + "\n")
	}

	return []byte(csv.String()), nil
}

// GetSearchMetadata retrieves metadata about a saved search.
func (r *SearchRepository) GetSearchMetadata(ctx context.Context, searchID int64) (*SearchMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	search, ok := r.searches[searchID]
	if !ok {
		return nil, fmt.Errorf("search not found")
	}

	return &SearchMetadata{
		ID:          search.ID,
		QueryString: search.QueryString,
		CreatedAt:   search.CreatedAt,
	}, nil
}

// ListUserSearchesPaginated retrieves paginated saved searches for a user.
func (r *SearchRepository) ListUserSearchesPaginated(ctx context.Context, userID int64, limit, offset int) ([]*SavedSearch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*SavedSearch
	for _, search := range r.searches {
		if search.UserID == userID {
			results = append(results, search)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	// Apply pagination
	if offset >= len(results) {
		return []*SavedSearch{}, nil
	}

	end := offset + limit
	if end > len(results) {
		end = len(results)
	}

	return results[offset:end], nil
}
