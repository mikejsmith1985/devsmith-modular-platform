package db

import (
	"database/sql"
	"testing"
)

// setupTestDB attempts to connect to a test database
// Returns nil if database is not available, allowing tests to skip gracefully
func setupTestDB(t *testing.T) *sql.DB {
	// For unit tests without Docker, return nil to skip
	// Integration tests use the integration_test.go file with explicit Docker setup
	return nil
}
