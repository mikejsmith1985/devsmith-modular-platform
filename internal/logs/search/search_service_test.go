// Package search provides advanced filtering and search functionality for log entries.
// RED Phase: Test-driven development with comprehensive failing tests.
package search

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSearchService_ExecuteSearch tests basic search execution
func TestSearchService_ExecuteSearch(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	results, err := service.ExecuteSearch(ctx, "message:error")

	require.NoError(t, err)
	require.NotNil(t, results)
	// Results should match the query
}

// TestSearchService_FullTextSearch tests PostgreSQL full-text search
// Uses ts_vector for efficient searching
func TestSearchService_FullTextSearch(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	// Search for natural language phrase
	results, err := service.ExecuteSearch(ctx, "database connection failed")

	require.NoError(t, err)
	require.NotNil(t, results)
	// Should find related logs using full-text search
}

// TestSearchService_RegexSearch tests regex pattern matching
func TestSearchService_RegexSearch(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	// Search with regex pattern
	results, err := service.ExecuteSearch(ctx, "/error: \\d+/")

	require.NoError(t, err)
	require.NotNil(t, results)
	// Should find logs matching pattern
}

// TestSearchService_BooleanAND tests AND operator matching
func TestSearchService_BooleanAND(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	results, err := service.ExecuteSearch(ctx, "service:portal AND level:error")

	require.NoError(t, err)
	require.NotNil(t, results)

	// All results should have both conditions
	for _, result := range results {
		assert.Equal(t, "portal", result["service"])
		assert.Equal(t, "error", result["level"])
	}
}

// TestSearchService_BooleanOR tests OR operator matching
func TestSearchService_BooleanOR(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	results, err := service.ExecuteSearch(ctx, "level:error OR level:critical")

	require.NoError(t, err)
	require.NotNil(t, results)

	// All results should have either condition
	for _, result := range results {
		level := result["level"]
		assert.True(t, level == "error" || level == "critical")
	}
}

// TestSearchService_BooleanNOT tests NOT operator
func TestSearchService_BooleanNOT(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	results, err := service.ExecuteSearch(ctx, "NOT level:debug")

	require.NoError(t, err)
	require.NotNil(t, results)

	// No result should have level:debug
	for _, result := range results {
		assert.NotEqual(t, "debug", result["level"])
	}
}

// TestSearchService_ComplexBooleanExpression tests complex query
// (message:error AND service:portal) OR level:critical
func TestSearchService_ComplexBooleanExpression(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	results, err := service.ExecuteSearch(ctx, "(message:error AND service:portal) OR level:critical")

	require.NoError(t, err)
	require.NotNil(t, results)

	// Each result should match one of the conditions
	for _, result := range results {
		isPortalError := result["service"] == "portal" && result["level"] == "error"
		isCritical := result["level"] == "critical"
		assert.True(t, isPortalError || isCritical, "Result should match complex query")
	}
}

// TestSearchService_PerformanceUnder100ms tests search completes in <100ms
// For 100k logs
func TestSearchService_PerformanceUnder100ms(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	start := time.Now()
	_, err := service.ExecuteSearch(ctx, "message:error")
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, duration, 100*time.Millisecond, "Search must complete in <100ms for 100k logs")
}

// TestSearchService_SearchWithFilters tests combining search with filters
func TestSearchService_SearchWithFilters(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	filters := map[string]string{
		"service": "portal",
		"level":   "error",
	}

	results, err := service.ExecuteSearchWithFilters(ctx, "message:connection", filters)

	require.NoError(t, err)
	require.NotNil(t, results)

	// Results should match both search and filters
	for _, result := range results {
		assert.Equal(t, "portal", result["service"])
		assert.Equal(t, "error", result["level"])
	}
}

// TestSearchService_DateRangeFilter tests filtering by date range
func TestSearchService_DateRangeFilter(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	filters := map[string]interface{}{
		"from": yesterday,
		"to":   now,
	}

	results, err := service.ExecuteSearchWithDateRange(ctx, "message:error", filters)

	require.NoError(t, err)
	require.NotNil(t, results)

	// All results should be within date range
	for _, result := range results {
		createdAt, ok := result["created_at"].(time.Time)
		require.True(t, ok)
		assert.True(t, createdAt.After(yesterday) && createdAt.Before(now))
	}
}

// TestSearchService_CaseSensitiveOption tests case sensitivity control
func TestSearchService_CaseSensitiveOption(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	// Case-insensitive (default)
	resultsInsensitive, err := service.ExecuteSearchCaseSensitive(ctx, "message:ERROR", false)
	require.NoError(t, err)

	// Case-sensitive
	resultsSensitive, err := service.ExecuteSearchCaseSensitive(ctx, "message:ERROR", true)
	require.NoError(t, err)

	// Case-insensitive should find more or equal results
	assert.GreaterOrEqual(t, len(resultsInsensitive), len(resultsSensitive))
}

// TestSearchService_HighlightMatches tests highlighting search matches in results
func TestSearchService_HighlightMatches(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	results, err := service.ExecuteSearchWithHighlight(ctx, "connection")

	require.NoError(t, err)

	// Results should have highlighted matches
	for _, result := range results {
		if message, ok := result["message"].(string); ok {
			// Message should contain highlight markers or highlighted text
			_ = message // Use in assertion
		}
	}
}

// TestSearchService_Pagination tests paginated search results
func TestSearchService_Pagination(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	// Get first page
	page1, total, err := service.ExecuteSearchPaginated(ctx, "message:error", 10, 0)

	require.NoError(t, err)
	assert.Greater(t, total, 0, "Should find at least some results")
	assert.LessOrEqual(t, len(page1), 10)

	// Get second page
	page2, _, err := service.ExecuteSearchPaginated(ctx, "message:error", 10, 10)

	require.NoError(t, err)
	// Different pages should have different results
	if len(page1) > 0 && len(page2) > 0 {
		assert.NotEqual(t, page1[0]["id"], page2[0]["id"])
	}
}

// TestSearchService_Sorting tests sorting search results
func TestSearchService_Sorting(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	// Sort by created_at DESC (newest first)
	results, err := service.ExecuteSearchSorted(ctx, "message:error", "created_at", "DESC")

	require.NoError(t, err)
	require.NotNil(t, results)

	// Verify sorting (if results exist)
	if len(results) > 1 {
		prev := results[0]["created_at"].(time.Time)
		for i := 1; i < len(results); i++ {
			curr := results[i]["created_at"].(time.Time)
			assert.True(t, prev.After(curr) || prev.Equal(curr))
			prev = curr
		}
	}
}

// TestSearchService_Aggregation tests aggregating results
func TestSearchService_Aggregation(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	// Count errors by service
	agg, err := service.ExecuteSearchAggregation(ctx, "level:error", "service")

	require.NoError(t, err)
	require.NotNil(t, agg)

	// Should have counts for each service
	for service, count := range agg {
		assert.NotEmpty(t, service)
		assert.Greater(t, count, 0)
	}
}

// TestSearchService_SaveAndExecute tests saving and re-executing searches
func TestSearchService_SaveAndExecute(t *testing.T) {
	repo := NewSearchRepository(nil)
	service := NewSearchServiceWithRepo(repo)
	ctx := context.Background()

	// Save search
	saved := &SavedSearch{
		UserID:      1,
		Name:        "Portal Errors",
		QueryString: "service:portal AND level:error",
	}
	searchID, err := repo.SaveSearch(ctx, saved)
	require.NoError(t, err)

	// Execute it
	results, err := service.ExecuteSavedSearch(ctx, searchID)

	require.NoError(t, err)
	require.NotNil(t, results)
}

// TestSearchService_InvalidQueryRejection tests rejection of invalid queries
func TestSearchService_InvalidQueryRejection(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	invalidQueries := []string{
		"message:",       // Missing value
		"/invalid regex", // Invalid regex
		"(unclosed",      // Unmatched paren
	}

	for _, query := range invalidQueries {
		_, err := service.ExecuteSearch(ctx, query)
		assert.Error(t, err, "Should reject invalid query: %s", query)
	}
}

// TestSearchService_SearchCaching tests caching of recent search results
func TestSearchService_SearchCaching(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	query := "message:error"

	// First search
	start := time.Now()
	results1, err := service.ExecuteSearch(ctx, query)
	duration1 := time.Since(start)
	require.NoError(t, err)

	// Second search (should be cached)
	start = time.Now()
	results2, err := service.ExecuteSearch(ctx, query)
	duration2 := time.Since(start)
	require.NoError(t, err)

	// Cached result should be faster
	assert.Less(t, duration2, duration1)
	// Results should be identical
	assert.Equal(t, len(results1), len(results2))
}

// TestSearchService_CacheTTL tests cache expiration
func TestSearchService_CacheTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cache test in short mode")
	}

	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	query := "message:error"

	// First search
	_, _ = service.ExecuteSearch(ctx, query)

	// Wait for cache to expire (should be < 1 second in tests)
	time.Sleep(2 * time.Second)

	// Should not be cached anymore
	cached := service.GetCachedResult(ctx, query)
	assert.Nil(t, cached, "Cache should have expired")
}

// TestSearchService_ConcurrentSearches tests thread-safe concurrent searches
func TestSearchService_ConcurrentSearches(t *testing.T) {
	service := NewSearchService(NewSearchRepository(nil))
	ctx := context.Background()

	results := make(chan error, 5)

	// Execute 5 concurrent searches
	for i := 0; i < 5; i++ {
		go func() {
			_, err := service.ExecuteSearch(ctx, "message:error")
			results <- err
		}()
	}

	// All should complete successfully
	for i := 0; i < 5; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}
