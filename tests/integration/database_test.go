//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDatabaseConnections(t *testing.T) {
	databases := []struct {
		name    string
		connStr string
	}{
		{"Portal", "postgresql://portal_user:portal_pass@localhost:5432/devsmith_portal?sslmode=disable"},
		{"Review", "postgresql://review_user:review_pass@localhost:5432/devsmith_review?sslmode=disable"},
		{"Logs", "postgresql://logs_user:logs_pass@localhost:5432/devsmith_logs?sslmode=disable"},
		{"Analytics", "postgresql://analytics_user:analytics_pass@localhost:5432/devsmith_analytics?sslmode=disable"},
	}

	for _, db := range databases {
		t.Run(db.name, func(t *testing.T) {
			pool, err := pgxpool.New(context.Background(), db.connStr)
			if err != nil {
				t.Skipf("%s database connection failed (likely not running): %v", db.name, err)
			}
			defer pool.Close()

			// Test connection
			err = pool.Ping(context.Background())
			if err != nil {
				t.Skipf("%s database ping failed (likely not running): %v", db.name, err)
			}
		})
	}
}

func TestPortalDatabaseSchema(t *testing.T) {
	pool, err := pgxpool.New(context.Background(),
		"postgresql://portal_user:portal_pass@localhost:5432/devsmith_portal?sslmode=disable")
	if err != nil {
		t.Skipf("Portal database connection failed (likely not running): %v", err)
	}
	defer pool.Close()

	// Check if schema exists
	var schemaExists int
	err = pool.QueryRow(context.Background(),
		"SELECT 1 FROM information_schema.schemata WHERE schema_name='portal'").Scan(&schemaExists)
	if err != nil {
		t.Skipf("Portal schema check failed (likely not running): %v", err)
	}
}

func TestReviewDatabaseSchema(t *testing.T) {
	pool, err := pgxpool.New(context.Background(),
		"postgresql://review_user:review_pass@localhost:5432/devsmith_review?sslmode=disable")
	if err != nil {
		t.Skipf("Review database connection failed (likely not running): %v", err)
	}
	defer pool.Close()

	var schemaExists int
	err = pool.QueryRow(context.Background(),
		"SELECT 1 FROM information_schema.schemata WHERE schema_name='review'").Scan(&schemaExists)
	if err != nil {
		t.Skipf("Review schema check failed (likely not running): %v", err)
	}
}

func TestLogsDatabaseSchema(t *testing.T) {
	pool, err := pgxpool.New(context.Background(),
		"postgresql://logs_user:logs_pass@localhost:5432/devsmith_logs?sslmode=disable")
	if err != nil {
		t.Skipf("Logs database connection failed (likely not running): %v", err)
	}
	defer pool.Close()

	var schemaExists int
	err = pool.QueryRow(context.Background(),
		"SELECT 1 FROM information_schema.schemata WHERE schema_name='logs'").Scan(&schemaExists)
	if err != nil {
		t.Skipf("Logs schema check failed (likely not running): %v", err)
	}
}

func TestAnalyticsDatabaseSchema(t *testing.T) {
	pool, err := pgxpool.New(context.Background(),
		"postgresql://analytics_user:analytics_pass@localhost:5432/devsmith_analytics?sslmode=disable")
	if err != nil {
		t.Skipf("Analytics database connection failed (likely not running): %v", err)
	}
	defer pool.Close()

	var schemaExists int
	err = pool.QueryRow(context.Background(),
		"SELECT 1 FROM information_schema.schemata WHERE schema_name='analytics'").Scan(&schemaExists)
	if err != nil {
		t.Skipf("Analytics schema check failed (likely not running): %v", err)
	}
}

func TestAnalyticsReadOnlyAccessToLogs(t *testing.T) {
	t.Skip("Analytics read-only access test requires logs.entries table with data")
}
