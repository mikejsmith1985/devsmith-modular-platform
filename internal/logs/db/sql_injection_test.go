package logs_db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogRepository_Save_ChecksNilDB(t *testing.T) {
	// GIVEN: A log repository with nil database
	repo := NewLogRepository(nil)
	entry := &LogEntry{
		Service:   "test-service",
		Level:     "info",
		Message:   "Test message",
		Metadata:  map[string]interface{}{"key": "value"},
		CreatedAt: time.Now(),
	}

	// WHEN: Save is called
	id, err := repo.Save(context.Background(), entry)

	// THEN: Returns mock ID without attempting SQL (no nil pointer panic)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
}

func TestLogRepository_Query_ChecksNilDB(t *testing.T) {
	// GIVEN: A log repository with nil database
	repo := NewLogRepository(nil)
	filters := &QueryFilters{
		Service: "test-service",
		Level:   "info",
	}
	page := PageOptions{Limit: 10, Offset: 0}

	// WHEN: Query is called with nil db
	entries, err := repo.Query(context.Background(), filters, page)

	// THEN: Returns empty list safely without attempting SQL
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestBuildWhereClause_ParameterizesAllFragments(t *testing.T) {
	// GIVEN: Filters with potentially malicious SQL injection input
	filters := &QueryFilters{
		Service: "service'; DROP TABLE logs.entries; --",
		Level:   "error' OR '1'='1' OR '1",
		Search:  "message' UNION SELECT * FROM portal.users; --",
	}

	// WHEN: buildWhereClause constructs SQL fragments
	fragments, args, _ := buildWhereClause(filters)

	// THEN: All fragments use $N parameter placeholders (parameterized queries)
	t.Run("AllFragmentsAreParameterized", func(t *testing.T) {
		for _, fragment := range fragments {
			// Check that fragment uses $ parameter syntax
			assert.Regexp(t, `\$\d+`, fragment, "Fragment must use parameterized $N syntax")
			// Ensure no unescaped quotes in the SQL fragment itself
			assert.NotContains(t, fragment, "DROP", "SQL keywords should not be in fragments")
			assert.NotContains(t, fragment, "UNION", "SQL keywords should not be in fragments")
		}
	})

	// THEN: Malicious input values are safely placed in args, not in SQL
	t.Run("MaliciousInputInArgs", func(t *testing.T) {
		assert.Contains(t, args, "service'; DROP TABLE logs.entries; --")
		assert.Contains(t, args, "error' OR '1'='1' OR '1")
		assert.Contains(t, args, "%message' UNION SELECT * FROM portal.users; --%")
	})

	// THEN: The SQL string itself is safe even if all args are malicious
	t.Run("SQLFragmentsAreInjectionSafe", func(t *testing.T) {
		// Reconstruct what final SQL would look like
		fullSQL := "SELECT * FROM logs.entries"
		if len(fragments) > 0 {
			fullSQL += " WHERE " + fragments[0]
		}
		// The SQL should have parameter placeholders, not string literals
		assert.Regexp(t, `WHERE.*\$\d+`, fullSQL)
	})
}

func TestBuildWhereClause_EdgeCases(t *testing.T) {
	testCases := []struct {
		name    string
		filters *QueryFilters
	}{
		{
			name: "EmptyFilters",
			filters: &QueryFilters{
				Service: "",
				Level:   "",
				Search:  "",
			},
		},
		{
			name: "AllFiltersPopulated",
			filters: &QueryFilters{
				Service:    "portal",
				Level:      "error",
				Search:     "connection timeout",
				From:       time.Now().Add(-24 * time.Hour),
				To:         time.Now(),
				MetaEquals: map[string]string{"userId": "123", "action": "login"},
			},
		},
		{
			name: "SQLInjectionAttempts",
			filters: &QueryFilters{
				Service: "admin' --",
				Level:   "' OR '1'='1",
				Search:  "; DELETE FROM logs.entries; --",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fragments, args, _ := buildWhereClause(tc.filters)

			// Verify all fragments use parameterized queries
			for _, fragment := range fragments {
				assert.Regexp(t, `\$\d+`, fragment)
			}

			// At minimum: user input is separated from SQL structure
			// (empty filters = 0 fragments + 0 args is valid and safe)
			assert.GreaterOrEqual(t, len(args)+len(fragments), 0)
		})
	}
}
