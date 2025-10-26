// Package search provides advanced filtering and search functionality for log entries.
// RED Phase: Test-driven development with comprehensive failing tests.
package search

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSearchRepository_SaveSearch tests saving a search query
func TestSearchRepository_SaveSearch(t *testing.T) {
	repo := NewSearchRepository(nil) // nil DB for unit test
	ctx := context.Background()

	savedSearch := &SavedSearch{
		UserID:      1,
		Name:        "Portal Errors",
		QueryString: "service:portal AND level:error",
		Description: "All errors in portal service",
	}

	id, err := repo.SaveSearch(ctx, savedSearch)

	require.NoError(t, err)
	assert.Greater(t, id, int64(0), "Should return valid ID")
}

// TestSearchRepository_GetSavedSearch tests retrieving a saved search
func TestSearchRepository_GetSavedSearch(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	// First save a search
	savedSearch := &SavedSearch{
		UserID:      1,
		Name:        "Portal Errors",
		QueryString: "service:portal AND level:error",
	}
	id, err := repo.SaveSearch(ctx, savedSearch)
	require.NoError(t, err)

	// Then retrieve it
	retrieved, err := repo.GetSavedSearch(ctx, id)

	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, "Portal Errors", retrieved.Name)
	assert.Equal(t, savedSearch.QueryString, retrieved.QueryString)
}

// TestSearchRepository_ListUserSearches tests listing searches for a user
func TestSearchRepository_ListUserSearches(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	userID := int64(1)
	// Save multiple searches
	for i := 0; i < 3; i++ {
		repo.SaveSearch(ctx, &SavedSearch{
			UserID:      userID,
			Name:        "Search " + string(rune('A'+i)),
			QueryString: "message:error",
		})
	}

	searches, err := repo.ListUserSearches(ctx, userID)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(searches), 3, "Should return at least 3 searches")
}

// TestSearchRepository_DeleteSavedSearch tests deleting a saved search
func TestSearchRepository_DeleteSavedSearch(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	// Save a search
	savedSearch := &SavedSearch{
		UserID:      1,
		Name:        "To Delete",
		QueryString: "message:error",
	}
	id, err := repo.SaveSearch(ctx, savedSearch)
	require.NoError(t, err)

	// Delete it
	err = repo.DeleteSavedSearch(ctx, id)

	require.NoError(t, err)

	// Should not be retrievable
	_, err = repo.GetSavedSearch(ctx, id)
	assert.Error(t, err, "Should not find deleted search")
}

// TestSearchRepository_UpdateSavedSearch tests updating a saved search
func TestSearchRepository_UpdateSavedSearch(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	// Save initial search
	savedSearch := &SavedSearch{
		UserID:      1,
		Name:        "Portal Errors",
		QueryString: "service:portal AND level:error",
	}
	id, err := repo.SaveSearch(ctx, savedSearch)
	require.NoError(t, err)

	// Update it
	savedSearch.ID = id
	savedSearch.QueryString = "service:portal AND (level:error OR level:warn)"
	err = repo.UpdateSavedSearch(ctx, savedSearch)

	require.NoError(t, err)

	// Verify update
	retrieved, _ := repo.GetSavedSearch(ctx, id)
	assert.Equal(t, "service:portal AND (level:error OR level:warn)", retrieved.QueryString)
}

// TestSearchRepository_SaveSearchHistory tests recording search history
func TestSearchRepository_SaveSearchHistory(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	userID := int64(1)
	const testQueryString = "message:error"

	entry, err := repo.SaveSearchHistory(ctx, userID, testQueryString)

	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, userID, entry.UserID)
	assert.Equal(t, testQueryString, entry.QueryString)
	assert.NotZero(t, entry.SearchedAt)
}

// TestSearchRepository_GetSearchHistory tests retrieving search history
func TestSearchRepository_GetSearchHistory(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	userID := int64(1)
	queries := []string{
		"message:error",
		"service:portal",
		"level:critical",
	}

	// Add to history
	for _, q := range queries {
		repo.SaveSearchHistory(ctx, userID, q)
	}

	// Retrieve history (most recent first)
	history, err := repo.GetSearchHistory(ctx, userID, 10)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(history), 3, "Should have at least 3 history entries")
	// Most recent should be first
	assert.Equal(t, "level:critical", history[0].QueryString)
}

// TestSearchRepository_GetRecentSearches tests getting unique recent searches
func TestSearchRepository_GetRecentSearches(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	userID := int64(1)

	// Save same query multiple times
	for i := 0; i < 5; i++ {
		repo.SaveSearchHistory(ctx, userID, "message:error")
	}

	// Add different query
	repo.SaveSearchHistory(ctx, userID, "service:portal")

	// Get recent (should deduplicate)
	recent, err := repo.GetRecentSearches(ctx, userID, 5)

	require.NoError(t, err)
	// Should have 2 unique queries, not 6
	assert.Equal(t, 2, len(recent), "Should return unique queries only")
}

// TestSearchRepository_ClearSearchHistory tests clearing history
func TestSearchRepository_ClearSearchHistory(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	userID := int64(1)

	// Add history
	repo.SaveSearchHistory(ctx, userID, "message:error")
	repo.SaveSearchHistory(ctx, userID, "service:portal")

	// Clear
	err := repo.ClearSearchHistory(ctx, userID)

	require.NoError(t, err)

	// Should be empty
	history, _ := repo.GetSearchHistory(ctx, userID, 10)
	assert.Empty(t, history, "History should be cleared")
}

// TestSearchRepository_SearchHistoryLimit tests that history respects limit
func TestSearchRepository_SearchHistoryLimit(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	userID := int64(1)

	// Add many searches
	for i := 0; i < 100; i++ {
		repo.SaveSearchHistory(ctx, userID, "query "+string(rune('A'+(i%26))))
	}

	// Request only 10
	history, err := repo.GetSearchHistory(ctx, userID, 10)

	require.NoError(t, err)
	assert.LessOrEqual(t, len(history), 10, "Should respect limit")
}

// TestSearchRepository_SharedSearches tests sharing saved searches
func TestSearchRepository_ShareSearch(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	// User 1 saves a search
	saved := &SavedSearch{
		UserID:      1,
		Name:        "Portal Errors",
		QueryString: "service:portal AND level:error",
	}
	searchID, err := repo.SaveSearch(ctx, saved)
	require.NoError(t, err)

	// User 1 shares with User 2
	err = repo.ShareSearch(ctx, searchID, 1, 2)

	require.NoError(t, err)

	// User 2 should see it
	sharedSearches, err := repo.GetSharedSearches(ctx, 2)
	require.NoError(t, err)
	assert.Greater(t, len(sharedSearches), 0, "User 2 should see shared search")
}

// TestSearchRepository_ExportSearchResults tests exporting search results
func TestSearchRepository_ExportResults(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	results := []interface{}{
		map[string]interface{}{"id": 1, "message": "error 1"},
		map[string]interface{}{"id": 2, "message": "error 2"},
	}

	// Export as JSON
	jsonExport, err := repo.ExportAsJSON(ctx, results)

	require.NoError(t, err)
	assert.NotEmpty(t, jsonExport)
	assert.Contains(t, string(jsonExport), "error")
}

// TestSearchRepository_ExportAsCSV tests CSV export format
func TestSearchRepository_ExportAsCSV(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	results := []interface{}{
		map[string]interface{}{"id": 1, "message": "error 1", "level": "ERROR"},
		map[string]interface{}{"id": 2, "message": "error 2", "level": "ERROR"},
	}

	csvExport, err := repo.ExportAsCSV(ctx, results)

	require.NoError(t, err)
	assert.NotEmpty(t, csvExport)
	// Should have header and data rows
	assert.Greater(t, len(csvExport), 1)
}

// TestSearchRepository_ValidateSavedSearchPermission tests permission validation
func TestSearchRepository_ValidateSavedSearchPermission(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	// User 1 saves a search
	saved := &SavedSearch{
		UserID:      1,
		Name:        "Portal Errors",
		QueryString: "service:portal",
	}
	searchID, _ := repo.SaveSearch(ctx, saved)

	// User 1 can access their search
	err := repo.ValidateSearchAccess(ctx, searchID, 1)
	assert.NoError(t, err)

	// User 2 cannot access User 1's search (unless shared)
	err = repo.ValidateSearchAccess(ctx, searchID, 2)
	assert.Error(t, err, "User 2 should not access User 1's private search")
}

// TestSearchRepository_SearchNameUniquenessPerUser tests name uniqueness
func TestSearchRepository_SearchNameUniqueness(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	userID := int64(1)

	// Save first search
	repo.SaveSearch(ctx, &SavedSearch{
		UserID:      userID,
		Name:        "Portal Errors",
		QueryString: "service:portal",
	})

	// Try to save with same name for same user
	_, err := repo.SaveSearch(ctx, &SavedSearch{
		UserID:      userID,
		Name:        "Portal Errors",
		QueryString: "service:portal AND level:error",
	})

	assert.Error(t, err, "Should reject duplicate search names for same user")
}

// TestSearchRepository_SearchMetadata tests search metadata (execution stats)
func TestSearchRepository_GetSearchMetadata(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	// Save a search
	saved := &SavedSearch{
		UserID:      1,
		Name:        "Test",
		QueryString: "message:error",
	}
	searchID, _ := repo.SaveSearch(ctx, saved)

	// Get metadata
	metadata, err := repo.GetSearchMetadata(ctx, searchID)

	require.NoError(t, err)
	require.NotNil(t, metadata)
	assert.NotZero(t, metadata.CreatedAt)
	assert.Equal(t, "message:error", metadata.QueryString)
}

// TestSearchRepository_SearchPagination tests paginated search results
func TestSearchRepository_SearchWithPagination(t *testing.T) {
	repo := NewSearchRepository(nil)
	ctx := context.Background()

	userID := int64(1)

	// Create many searches
	for i := 0; i < 25; i++ {
		repo.SaveSearch(ctx, &SavedSearch{
			UserID:      userID,
			Name:        "Search " + string(rune('A'+(i%26))),
			QueryString: "query",
		})
	}

	// Get paginated results
	page1, err := repo.ListUserSearchesPaginated(ctx, userID, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 10, len(page1))

	// Get second page
	page2, err := repo.ListUserSearchesPaginated(ctx, userID, 10, 10)

	require.NoError(t, err)
	assert.Greater(t, len(page2), 0)
}
