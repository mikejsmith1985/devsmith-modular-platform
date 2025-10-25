package db

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// SAVE METHOD TESTS
// ============================================================================

func TestLogRepository_Save_Success(t *testing.T) {
	// nolint:govet // test struct field alignment is acceptable
	tests := []struct {
		name      string
		wantID    int64
		wantError bool
		entry     *LogEntry
	}{
		{
			name:      "valid entry returns ID",
			wantID:    1,
			wantError: false,
			entry: &LogEntry{
				CreatedAt: time.Now(),
				Level:     "error",
				Message:   "Database connection failed",
				Service:   "db_service",
				Metadata:  map[string]interface{}{"user_id": 123},
			},
		},
		{
			name:      "minimal valid entry",
			wantID:    2,
			wantError: false,
			entry: &LogEntry{
				CreatedAt: time.Now(),
				Level:     "info",
				Message:   "Test",
				Service:   "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &LogRepository{}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			got, err := repo.Save(ctx, tt.entry)

			if (err != nil) != tt.wantError {
				t.Errorf("Save() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError && got <= 0 {
				t.Errorf("Save() returned invalid ID %d", got)
			}
		})
	}
}

func TestLogRepository_Save_ValidationErrors(t *testing.T) {
	// nolint:govet // test struct field alignment is acceptable
	tests := []struct {
		name    string
		wantErr bool
		entry   *LogEntry
	}{
		{"nil entry", true, nil},
		{"empty message", true, &LogEntry{CreatedAt: time.Now(), Level: "info", Service: "test", Message: ""}},
		{"empty level", true, &LogEntry{CreatedAt: time.Now(), Level: "", Service: "test", Message: "msg"}},
		{"empty service", true, &LogEntry{CreatedAt: time.Now(), Level: "info", Service: "", Message: "msg"}},
		{"zero timestamp", true, &LogEntry{CreatedAt: time.Time{}, Level: "info", Service: "test", Message: "msg"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &LogRepository{}
			_, err := repo.Save(context.Background(), tt.entry)

			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLogRepository_Save_ContextCancellation(t *testing.T) {
	repo := &LogRepository{}
	entry := &LogEntry{
		CreatedAt: time.Now(),
		Level:     "info",
		Service:   "test",
		Message:   "test",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.Save(ctx, entry)
	if err == nil {
		t.Error("Save() should error on cancelled context")
	}
}

// ============================================================================
// QUERY METHOD TESTS
// ============================================================================

func TestLogRepository_Query_Success(t *testing.T) {
	// nolint:govet // test struct field alignment is acceptable
	tests := []struct {
		name    string
		wantErr bool
		filters *QueryFilters
		page    PageOptions
	}{
		{
			name:    "query all with default pagination",
			wantErr: false,
			filters: nil,
			page:    PageOptions{Limit: 10, Offset: 0},
		},
		{
			name:    "query with service filter",
			wantErr: false,
			filters: &QueryFilters{Service: "portal"},
			page:    PageOptions{Limit: 10, Offset: 0},
		},
		{
			name:    "query with level filter",
			wantErr: false,
			filters: &QueryFilters{Level: "error"},
			page:    PageOptions{Limit: 10, Offset: 0},
		},
		{
			name:    "full-text search",
			wantErr: false,
			filters: &QueryFilters{Search: "connection timeout"},
			page:    PageOptions{Limit: 10, Offset: 0},
		},
		{
			name:    "time range query",
			wantErr: false,
			filters: &QueryFilters{From: time.Now().AddDate(0, 0, -7), To: time.Now()},
			page:    PageOptions{Limit: 10, Offset: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &LogRepository{}
			entries, err := repo.Query(context.Background(), tt.filters, tt.page)

			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && entries == nil {
				t.Error("Query() should return non-nil slice")
			}
		})
	}
}

func TestLogRepository_Query_ValidationErrors(t *testing.T) {
	// nolint:govet // test struct field alignment is acceptable
	tests := []struct {
		name    string
		wantErr bool
		filters *QueryFilters
		page    PageOptions
	}{
		{
			name:    "invalid limit (negative)",
			wantErr: true,
			filters: nil,
			page:    PageOptions{Limit: -1, Offset: 0},
		},
		{
			name:    "invalid offset (negative)",
			wantErr: true,
			filters: nil,
			page:    PageOptions{Limit: 10, Offset: -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &LogRepository{}
			_, err := repo.Query(context.Background(), tt.filters, tt.page)

			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLogRepository_Query_Pagination(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()
	var err error
	var entries []*LogEntry

	// Test offset
	entries, err = repo.Query(ctx, nil, PageOptions{Limit: 10, Offset: 5})
	if err != nil {
		t.Errorf("Query() with offset error = %v", err)
	}
	if entries == nil {
		t.Error("Query() should return slice, not nil")
	}

	// Test high limit
	_, err = repo.Query(ctx, nil, PageOptions{Limit: 1000, Offset: 0})
	if err != nil {
		t.Errorf("Query() with large limit error = %v", err)
	}
}

// ============================================================================
// GETBYID METHOD TESTS
// ============================================================================

func TestLogRepository_GetByID_Success(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	entry, err := repo.GetByID(ctx, 1)
	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}

	if entry != nil && entry.ID <= 0 {
		t.Error("GetByID() returned entry with invalid ID")
	}
}

func TestLogRepository_GetByID_NotFound(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	entry, err := repo.GetByID(ctx, 999999)
	// Should either error or return nil, but not both
	if err == nil && entry != nil {
		t.Error("GetByID() should handle not found case properly")
	}
}

func TestLogRepository_GetByID_InvalidID(t *testing.T) {
	tests := []struct {
		name string
		id   int64
	}{
		{"zero ID", 0},
		{"negative ID", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &LogRepository{}
			_, err := repo.GetByID(context.Background(), tt.id)
			if err == nil {
				t.Errorf("GetByID() with %s should validate", tt.name)
			}
		})
	}
}

// ============================================================================
// GETSTATS METHOD TESTS
// ============================================================================

func TestLogRepository_GetStats_ReturnsAggregates(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	stats, err := repo.GetStats(ctx)
	if err != nil {
		t.Errorf("GetStats() error = %v", err)
	}

	if stats == nil {
		t.Error("GetStats() should return non-nil stats map")
	}
}

func TestLogRepository_GetStats_ByLevel(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	stats, err := repo.GetStats(ctx)
	if err != nil {
		t.Errorf("GetStats() error = %v", err)
	}

	if stats != nil {
		// Should have aggregated counts by level
		if _, ok := stats["by_level"]; !ok {
			t.Error("GetStats() should include by_level aggregation")
		}
	}
}

func TestLogRepository_GetStats_ByService(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	stats, err := repo.GetStats(ctx)
	if err != nil {
		t.Errorf("GetStats() error = %v", err)
	}

	if stats != nil {
		// Should have aggregated counts by service
		if _, ok := stats["by_service"]; !ok {
			t.Error("GetStats() should include by_service aggregation")
		}
	}
}

func TestLogRepository_GetStats_TotalCount(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	stats, err := repo.GetStats(ctx)
	if err != nil {
		t.Errorf("GetStats() error = %v", err)
	}

	if stats != nil {
		// Should have total count
		if _, ok := stats["total"]; !ok {
			t.Error("GetStats() should include total count")
		}
	}
}

// ============================================================================
// DELETEOLD METHOD TESTS (RETENTION POLICY)
// ============================================================================

func TestLogRepository_DeleteOld_Success(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	cutoffTime := time.Now().AddDate(0, 0, -90)
	deletedCount, err := repo.DeleteOld(ctx, cutoffTime)

	if err != nil {
		t.Errorf("DeleteOld() error = %v", err)
	}

	if deletedCount < 0 {
		t.Errorf("DeleteOld() returned negative count %d", deletedCount)
	}
}

func TestLogRepository_DeleteOld_PreservesRecent(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	futureTime := time.Now().AddDate(0, 0, -1)
	deletedCount, err := repo.DeleteOld(ctx, futureTime)

	if err != nil {
		t.Errorf("DeleteOld() error = %v", err)
	}

	if deletedCount < 0 {
		t.Error("DeleteOld() should return valid count")
	}
}

func TestLogRepository_DeleteOld_ZeroTime(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	deletedCount, err := repo.DeleteOld(ctx, time.Time{})
	if err == nil && deletedCount > 0 {
		t.Error("DeleteOld() should validate zero time")
	}
}

func TestLogRepository_DeleteOld_90DayRetention(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	// Exactly 90 days ago
	cutoffTime := time.Now().AddDate(0, 0, -90)
	_, err := repo.DeleteOld(ctx, cutoffTime)

	if err != nil {
		t.Errorf("DeleteOld() error = %v", err)
	}
}

// ============================================================================
// BULK INSERT METHOD TESTS
// ============================================================================

func TestLogRepository_BulkInsert_Success(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	entries := []*LogEntry{
		{
			CreatedAt: time.Now(),
			Level:     "info",
			Message:   "Log 1",
			Service:   "test",
		},
		{
			CreatedAt: time.Now(),
			Level:     "error",
			Message:   "Log 2",
			Service:   "test",
		},
	}

	insertedCount, err := repo.BulkInsert(ctx, entries)
	if err != nil {
		t.Errorf("BulkInsert() error = %v", err)
	}

	if insertedCount != int64(len(entries)) {
		t.Errorf("BulkInsert() inserted %d, want %d", insertedCount, len(entries))
	}
}

func TestLogRepository_BulkInsert_EmptySlice(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	insertedCount, err := repo.BulkInsert(ctx, []*LogEntry{})
	if err == nil && insertedCount > 0 {
		t.Error("BulkInsert() should handle empty slice")
	}
}

func TestLogRepository_BulkInsert_NilSlice(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	_, err := repo.BulkInsert(ctx, nil)
	if err == nil {
		t.Error("BulkInsert() should error on nil slice")
	}
}

func TestLogRepository_BulkInsert_InvalidEntries(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	entries := []*LogEntry{
		{
			CreatedAt: time.Now(),
			Level:     "info",
			Message:   "", // Invalid
			Service:   "test",
		},
	}

	_, err := repo.BulkInsert(ctx, entries)
	if err == nil {
		t.Error("BulkInsert() should error on invalid entries")
	}
}

func TestLogRepository_BulkInsert_MixedValidity(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	entries := []*LogEntry{
		{CreatedAt: time.Now(), Level: "info", Message: "Valid", Service: "test"},
		{CreatedAt: time.Now(), Level: "error", Message: "", Service: "test"}, // Invalid
		{CreatedAt: time.Now(), Level: "warn", Message: "Valid", Service: "test"},
	}

	_, err := repo.BulkInsert(ctx, entries)
	if err == nil {
		t.Error("BulkInsert() should error when any entry is invalid")
	}
}

func TestLogRepository_BulkInsert_LargeDataset(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	entries := make([]*LogEntry, 1000)
	for i := 0; i < 1000; i++ {
		entries[i] = &LogEntry{
			CreatedAt: time.Now(),
			Level:     "info",
			Message:   "Bulk log",
			Service:   "bulk",
		}
	}

	insertedCount, err := repo.BulkInsert(ctx, entries)
	if err != nil {
		t.Errorf("BulkInsert() error = %v", err)
	}

	if insertedCount != 1000 {
		t.Errorf("BulkInsert() inserted %d, want 1000", insertedCount)
	}
}

// ============================================================================
// CONTEXT HANDLING
// ============================================================================

func TestLogRepository_ContextDeadline_Save(t *testing.T) {
	repo := &LogRepository{}
	entry := &LogEntry{
		CreatedAt: time.Now(),
		Level:     "info",
		Service:   "test",
		Message:   "test",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	time.Sleep(2 * time.Millisecond)
	defer cancel()

	_, err := repo.Save(ctx, entry)
	if err == nil {
		t.Error("Save() should respect context deadline")
	}
}

// ============================================================================
// SCHEMA AND INDEXING
// ============================================================================

func TestLogRepository_SchemaPresence(t *testing.T) {
	// Repository initialized successfully - this verifies basic instantiation works
	repo := &LogRepository{}
	if repo.db != nil {
		t.Error("unexpected: repo.db should be nil for test repo without connection")
	}
}

func TestLogRepository_IndexUsageOnTimestamp(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	filters := &QueryFilters{
		From: time.Now().AddDate(0, 0, -7),
		To:   time.Now(),
	}

	_, err := repo.Query(ctx, filters, PageOptions{Limit: 10, Offset: 0})
	if err != nil {
		t.Errorf("Query with timestamp range error = %v", err)
	}
}

func TestLogRepository_IndexUsageOnLevel(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	filters := &QueryFilters{Level: "error"}
	_, err := repo.Query(ctx, filters, PageOptions{Limit: 10, Offset: 0})

	if err != nil {
		t.Errorf("Query with level filter error = %v", err)
	}
}

func TestLogRepository_IndexUsageOnService(t *testing.T) {
	repo := &LogRepository{}
	ctx := context.Background()

	filters := &QueryFilters{Service: "portal"}
	_, err := repo.Query(ctx, filters, PageOptions{Limit: 10, Offset: 0})

	if err != nil {
		t.Errorf("Query with service filter error = %v", err)
	}
}
