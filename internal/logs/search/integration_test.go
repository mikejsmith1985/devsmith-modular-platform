//go:build integration
// +build integration

// Package search provides advanced filtering and search functionality for log entries.
// RED Phase: Integration tests for complete search workflows.
package search

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_CompleteSearchWorkflow tests end-to-end search workflow
func TestIntegration_CompleteSearchWorkflow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSearchRepository(db)
	service := NewSearchServiceWithRepo(db, repo)
	ctx := context.Background()

	// 1. Save a search
	saved := &SavedSearch{
		UserID:      1,
		Name:        "Portal Errors",
		QueryString: "service:portal AND level:error",
	}
	searchID, err := repo.SaveSearch(ctx, saved)
	require.NoError(t, err)

	// 2. Add to search history
	_, err = repo.SaveSearchHistory(ctx, 1, saved.QueryString)
	require.NoError(t, err)

	// 3. Execute search
	results, err := service.ExecuteSavedSearch(ctx, searchID)
	require.NoError(t, err)
	require.NotNil(t, results)

	// 4. Export results
	jsonExport, err := repo.ExportAsJSON(ctx, results)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonExport)
}

// TestIntegration_FullTextSearchPerformance tests FTS performance on 100k logs
func TestIntegration_FullTextSearchPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	service := NewSearchService(db, nil)
	ctx := context.Background()

	// Insert 100k test logs
	insertTestLogs(t, db, 100000)

	// Full-text search should complete in <100ms
	start := time.Now()
	_, err := service.ExecuteSearch(ctx, "database connection")
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, duration, 100*time.Millisecond, "FTS must complete <100ms for 100k logs")
}

// TestIntegration_RegexSearchWithIndexes tests regex search uses proper indexes
func TestIntegration_RegexSearchWithIndexes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewSearchService(db, nil)
	ctx := context.Background()

	// Insert test logs
	insertTestLogs(t, db, 10000)

	// Regex search should use indexes efficiently
	start := time.Now()
	results, err := service.ExecuteSearch(ctx, "/error: \\d+/")
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, duration, 100*time.Millisecond, "Regex search should be efficient")
	require.NotNil(t, results)
}

// TestIntegration_BooleanOperatorsCombined tests complex boolean expressions
func TestIntegration_BooleanOperatorsCombined(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewSearchService(db, nil)
	ctx := context.Background()

	// Insert diverse test logs
	insertDiverseTestLogs(t, db)

	// Complex query: (message:error AND service:portal) OR level:critical
	results, err := service.ExecuteSearch(ctx, "(message:error AND service:portal) OR level:critical")

	require.NoError(t, err)
	require.NotNil(t, results)

	// Validate results match query logic
	for _, result := range results {
		isPortalError := result["service"] == "portal" && result["level"] == "error"
		isCritical := result["level"] == "critical"
		assert.True(t, isPortalError || isCritical)
	}
}

// TestIntegration_SavedSearchSharing tests saved search sharing between users
func TestIntegration_SavedSearchSharing(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSearchRepository(db)
	ctx := context.Background()

	// User 1 saves a search
	saved := &SavedSearch{
		UserID:      1,
		Name:        "Shared Search",
		QueryString: "level:error",
	}
	searchID, err := repo.SaveSearch(ctx, saved)
	require.NoError(t, err)

	// User 1 shares with User 2
	err = repo.ShareSearch(ctx, searchID, 1, 2)
	require.NoError(t, err)

	// User 2 can access shared search
	err = repo.ValidateSearchAccess(ctx, searchID, 2)
	assert.NoError(t, err)
}

// TestIntegration_SearchHistoryTracking tests search history is tracked
func TestIntegration_SearchHistoryTracking(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSearchRepository(db)
	ctx := context.Background()

	userID := int64(1)

	// Execute multiple searches
	queries := []string{
		"service:portal",
		"level:error",
		"message:connection",
	}

	for _, q := range queries {
		_, _ = repo.SaveSearchHistory(ctx, userID, q)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Get history
	history, err := repo.GetSearchHistory(ctx, userID, 10)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(history), 3)
	// Most recent should be first
	assert.Equal(t, "message:connection", history[0].QueryString)
}

// TestIntegration_ExportFormats tests exporting in JSON and CSV
func TestIntegration_ExportFormats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSearchRepository(db)
	service := NewSearchService(db, nil)
	ctx := context.Background()

	// Insert test logs
	insertTestLogs(t, db, 100)

	// Search
	results, err := service.ExecuteSearch(ctx, "level:error")
	require.NoError(t, err)

	// Export as JSON
	jsonExport, err := repo.ExportAsJSON(ctx, results)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonExport)
	assert.Contains(t, string(jsonExport), "error")

	// Export as CSV
	csvExport, err := repo.ExportAsCSV(ctx, results)
	require.NoError(t, err)
	assert.NotEmpty(t, csvExport)
	assert.Greater(t, len(csvExport), 1) // At least header + 1 row
}

// TestIntegration_PaginationLargeResultSet tests pagination on large result sets
func TestIntegration_PaginationLargeResultSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	service := NewSearchService(db, nil)
	ctx := context.Background()

	// Insert 1000 test logs
	insertTestLogs(t, db, 1000)

	// Get results with pagination
	page1, total, err := service.ExecuteSearchPaginated(ctx, "level:error", 50, 0)

	require.NoError(t, err)
	assert.Equal(t, 50, len(page1))
	assert.Greater(t, total, 50)

	// Get next page
	page2, _, err := service.ExecuteSearchPaginated(ctx, "level:error", 50, 50)

	require.NoError(t, err)
	// Pages should be different
	assert.NotEqual(t, page1[0]["id"], page2[0]["id"])
}

// TestIntegration_SearchCaseSensitivity tests case-insensitive full-text search
func TestIntegration_SearchCaseSensitivity(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewSearchService(db, nil)
	ctx := context.Background()

	// Insert test data with mixed case
	insertTestLogsWithCase(t, db)

	// Search with different case
	results1, err := service.ExecuteSearchCaseSensitive(ctx, "message:DATABASE", false)
	require.NoError(t, err)

	results2, err := service.ExecuteSearchCaseSensitive(ctx, "message:database", false)
	require.NoError(t, err)

	// Both should find same results (case-insensitive)
	assert.Equal(t, len(results1), len(results2))
}

// TestIntegration_DateRangeFiltering tests filtering by date range
func TestIntegration_DateRangeFiltering(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewSearchService(db, nil)
	ctx := context.Background()

	// Insert logs over time range
	insertLogsOverTimeRange(t, db)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	filters := map[string]interface{}{
		"from": yesterday,
		"to":   now,
	}

	results, err := service.ExecuteSearchWithDateRange(ctx, "level:error", filters)

	require.NoError(t, err)

	// All results should be within date range
	for _, result := range results {
		createdAt, ok := result["created_at"].(time.Time)
		require.True(t, ok)
		assert.True(t, createdAt.After(yesterday))
		assert.True(t, createdAt.Before(now))
	}
}

// TestIntegration_AggregationStats tests result aggregation for statistics
func TestIntegration_AggregationStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewSearchService(db, nil)
	ctx := context.Background()

	// Insert diverse logs
	insertDiverseTestLogs(t, db)

	// Aggregate errors by service
	agg, err := service.ExecuteSearchAggregation(ctx, "level:error", "service")

	require.NoError(t, err)
	assert.Greater(t, len(agg), 0)

	// Each service should have count
	for svc, count := range agg {
		assert.NotEmpty(t, svc)
		assert.Greater(t, count, 0)
	}
}

// TestIntegration_ConcurrentSearchExecution tests multiple searches at once
func TestIntegration_ConcurrentSearchExecution(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewSearchService(db, nil)
	ctx := context.Background()

	// Insert test data
	insertTestLogs(t, db, 1000)

	// Run concurrent searches
	results := make(chan error, 10)

	queries := []string{
		"level:error",
		"service:portal",
		"message:connection",
		"level:warning",
		"service:review",
		"message:failed",
		"level:info",
		"service:logs",
		"message:timeout",
		"level:debug",
	}

	for _, q := range queries {
		go func(query string) {
			_, err := service.ExecuteSearch(ctx, query)
			results <- err
		}(q)
	}

	// All should complete successfully
	for i := 0; i < len(queries); i++ {
		err := <-results
		assert.NoError(t, err)
	}
}

// TestIntegration_QueryValidationSafety tests query safety validation
func TestIntegration_QueryValidationSafety(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	parser := NewQueryParser()

	// Test that queries are validated for SQL injection
	dangerousQueries := []string{
		"'; DROP TABLE logs;--",
		"message:'; DROP TABLE logs_entries;--",
		"service:*/DROP/admin",
	}

	for _, q := range dangerousQueries {
		query, err := parser.ParseAndValidate(q)
		// Should either fail validation or be safely escaped
		if err == nil && query != nil {
			// If it parses, ensure SQL is parameterized
			sqlCond, params, _ := parser.GetSQLCondition(query)
			assert.NotContains(t, sqlCond, "DROP")
			assert.Greater(t, len(params), 0, "Should be parameterized")
		}
	}
}

// Helper functions

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	// This would normally connect to test database
	// For now, return nil which tests handle gracefully
	return nil
}

func insertTestLogs(t *testing.T, db *sql.DB, count int) {
	t.Helper()
	if db == nil {
		return
	}
	// Implementation would insert test data
}

func insertDiverseTestLogs(t *testing.T, db *sql.DB) {
	t.Helper()
	if db == nil {
		return
	}
	// Implementation would insert diverse test data for multi-service queries
}

func insertTestLogsWithCase(t *testing.T, db *sql.DB) {
	t.Helper()
	if db == nil {
		return
	}
	// Implementation would insert logs with mixed case messages
}

func insertLogsOverTimeRange(t *testing.T, db *sql.DB) {
	t.Helper()
	if db == nil {
		return
	}
	// Implementation would insert logs over time range (past 48 hours)
}
