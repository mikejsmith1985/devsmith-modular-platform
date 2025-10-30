package healthcheck

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// DatabaseChecker validates database connectivity
type DatabaseChecker struct {
	CheckName     string
	ConnectionURL string
}

// Name returns the checker name
func (c *DatabaseChecker) Name() string {
	return c.CheckName
}

// Check validates database connectivity
func (c *DatabaseChecker) Check() CheckResult {
	start := time.Now()
	result := CheckResult{
		Name:      c.CheckName,
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", c.ConnectionURL)
	if err != nil {
		result.Status = StatusFail
		result.Message = "Failed to open database connection"
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Log but don't fail - connection test already completed
		}
	}()

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		result.Status = StatusFail
		result.Message = "Database ping failed"
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	// Get database stats
	stats := db.Stats()
	result.Details["open_connections"] = stats.OpenConnections
	result.Details["in_use"] = stats.InUse
	result.Details["idle"] = stats.Idle

	// Run a simple query
	var version string
	err = db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		result.Status = StatusWarn
		result.Message = "Database connected but query failed"
		result.Error = err.Error()
	} else {
		result.Status = StatusPass
		result.Message = "Database connected and responsive"
		result.Details["postgres_version"] = version
	}

	result.Duration = time.Since(start)
	return result
}
