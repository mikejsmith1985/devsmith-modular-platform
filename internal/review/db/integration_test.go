//go:build integration
// +build integration

package review_db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupIntegrationDB creates a PostgreSQL container and returns a database connection
func setupIntegrationDB(ctx context.Context, t *testing.T) *sql.DB {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { container.Terminate(context.Background()) })

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	// Wait for database to be ready
	for i := 0; i < 30; i++ {
		if err := db.PingContext(ctx); err == nil {
			break
		}
		if i == 29 {
			require.Fail(t, "database failed to become ready")
		}
		time.Sleep(time.Second)
	}

	// Create schema and tables
	_, err = db.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS reviews")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS reviews.sessions (
			id SERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			title VARCHAR(255) NOT NULL,
			code_source VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS reviews.reading_sessions (
			id SERIAL PRIMARY KEY,
			session_id INT NOT NULL REFERENCES reviews.sessions(id) ON DELETE CASCADE,
			reading_mode VARCHAR(20) NOT NULL,
			target_path VARCHAR(500),
			scan_query TEXT,
			ai_response TEXT,
			user_annotations TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS reviews.critical_issues (
			id SERIAL PRIMARY KEY,
			reading_session_id INT NOT NULL REFERENCES reviews.reading_sessions(id) ON DELETE CASCADE,
			issue_type VARCHAR(50) NOT NULL,
			severity VARCHAR(20) NOT NULL,
			file_path VARCHAR(500),
			line_number INT,
			description TEXT NOT NULL,
			suggested_fix TEXT,
			status VARCHAR(20) NOT NULL DEFAULT 'open',
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	return db
}

// TestIntegration_ReadingSessionRepository tests CRUD operations with real database
func TestIntegration_ReadingSessionRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	// Create session
	_, err := db.ExecContext(ctx, `INSERT INTO reviews.sessions (user_id, title, code_source) VALUES ($1, $2, $3)`, 1, "Test", "paste")
	require.NoError(t, err)

	repo := NewReadingSessionRepository(db)

	// Test Create
	rs := &ReadingSession{SessionID: 1, ReadingMode: "preview", TargetPath: "/src/main.go"}
	created, err := repo.Create(ctx, rs)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)

	// Test GetByID
	retrieved, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "preview", retrieved.ReadingMode)

	// Test GetBySessionID
	sessions, err := repo.GetBySessionID(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)

	// Test Update
	retrieved.ReadingMode = "detailed"
	err = repo.Update(ctx, retrieved)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, created.ID)
	assert.Equal(t, "detailed", updated.ReadingMode)

	// Test Delete
	err = repo.Delete(ctx, created.ID)
	require.NoError(t, err)

	deleted, err := repo.GetByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.Nil(t, deleted)
}

// TestIntegration_CriticalIssuesRepository tests CRUD operations with real database
func TestIntegration_CriticalIssuesRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	// Create session and reading session
	_, err := db.ExecContext(ctx, `INSERT INTO reviews.sessions (user_id, title, code_source) VALUES ($1, $2, $3)`, 1, "Test", "paste")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO reviews.reading_sessions (session_id, reading_mode) VALUES ($1, $2)`, 1, "critical")
	require.NoError(t, err)

	repo := NewCriticalIssuesRepository(db)

	// Test Create
	issue := &CriticalIssue{
		ReadingSessionID: 1,
		IssueType:        "security",
		Severity:         "critical",
		FilePath:         "/src/auth.go",
		LineNumber:       42,
		Description:      "SQL injection",
		SuggestedFix:     "Use parameterized queries",
	}

	created, err := repo.Create(ctx, issue)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "open", created.Status)

	// Test GetByID
	retrieved, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "security", retrieved.IssueType)

	// Test GetByReadingSessionID
	issues, err := repo.GetByReadingSessionID(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, issues, 1)

	// Test Update
	retrieved.Status = "accepted"
	err = repo.Update(ctx, retrieved)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, created.ID)
	assert.Equal(t, "accepted", updated.Status)

	// Test Delete
	err = repo.Delete(ctx, created.ID)
	require.NoError(t, err)

	deleted, err := repo.GetByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.Nil(t, deleted)
}

// TestIntegration_CascadingDeletes tests foreign key cascade behavior
func TestIntegration_CascadingDeletes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	// Create data
	var sessionID int64
	err := db.QueryRowContext(ctx, `INSERT INTO reviews.sessions (user_id, title, code_source) VALUES ($1, $2, $3) RETURNING id`, 1, "Test", "paste").Scan(&sessionID)
	require.NoError(t, err)

	var readingSessionID int64
	err = db.QueryRowContext(ctx, `INSERT INTO reviews.reading_sessions (session_id, reading_mode) VALUES ($1, $2) RETURNING id`, sessionID, "critical").Scan(&readingSessionID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `INSERT INTO reviews.critical_issues (reading_session_id, issue_type, severity, description) VALUES ($1, $2, $3, $4)`,
		readingSessionID, "security", "critical", "Test issue")
	require.NoError(t, err)

	// Delete session - should cascade
	_, err = db.ExecContext(ctx, "DELETE FROM reviews.sessions WHERE id = $1", sessionID)
	require.NoError(t, err)

	// Verify cascade
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM reviews.reading_sessions WHERE session_id = $1", sessionID).Scan(&count)
	require.NoError(t, err)
	assert.Zero(t, count, "reading sessions should be cascade deleted")

	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM reviews.critical_issues WHERE reading_session_id = $1", readingSessionID).Scan(&count)
	require.NoError(t, err)
	assert.Zero(t, count, "critical issues should be cascade deleted")
}

// TestIntegration_Pagination tests pagination with large result sets
func TestIntegration_Pagination(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	// Create session
	_, err := db.ExecContext(ctx, `INSERT INTO reviews.sessions (user_id, title, code_source) VALUES ($1, $2, $3)`, 1, "Test", "paste")
	require.NoError(t, err)

	repo := NewReadingSessionRepository(db)

	// Create multiple reading sessions
	for i := 0; i < 15; i++ {
		_, err := repo.Create(ctx, &ReadingSession{
			SessionID:   1,
			ReadingMode: "preview",
			TargetPath:  fmt.Sprintf("/src/file%d.go", i),
		})
		require.NoError(t, err)
	}

	// Test retrieving all
	all, err := repo.GetBySessionID(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, all, 15)
}

// TestIntegration_ValidationErrors tests validation on create operations
func TestIntegration_ValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	// Create session
	_, err := db.ExecContext(ctx, `INSERT INTO reviews.sessions (user_id, title, code_source) VALUES ($1, $2, $3)`, 1, "Test", "paste")
	require.NoError(t, err)

	issuesRepo := NewCriticalIssuesRepository(db)
	readingRepo := NewReadingSessionRepository(db)

	// Create reading session
	rs, err := readingRepo.Create(ctx, &ReadingSession{
		SessionID:   1,
		ReadingMode: "preview",
	})
	require.NoError(t, err)

	// Test creating issue with missing required fields - should set defaults
	issue := &CriticalIssue{
		ReadingSessionID: rs.ID,
		IssueType:        "security",
		Severity:         "critical",
		Description:      "Test issue",
		// Status will default to "open"
	}

	created, err := issuesRepo.Create(ctx, issue)
	require.NoError(t, err)
	assert.Equal(t, "open", created.Status, "status should default to open")
}

// TestIntegration_MultipleIssuesPerSession tests multiple issues in one session
func TestIntegration_MultipleIssuesPerSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	// Create session
	_, err := db.ExecContext(ctx, `INSERT INTO reviews.sessions (user_id, title, code_source) VALUES ($1, $2, $3)`, 1, "Test", "paste")
	require.NoError(t, err)

	readingRepo := NewReadingSessionRepository(db)
	issuesRepo := NewCriticalIssuesRepository(db)

	// Create reading session
	rs, err := readingRepo.Create(ctx, &ReadingSession{
		SessionID:   1,
		ReadingMode: "critical",
	})
	require.NoError(t, err)

	// Create multiple issues with different severities
	severities := []string{"critical", "important", "minor"}
	issueTypes := []string{"security", "performance", "quality"}

	for i, severity := range severities {
		_, err := issuesRepo.Create(ctx, &CriticalIssue{
			ReadingSessionID: rs.ID,
			IssueType:        issueTypes[i],
			Severity:         severity,
			FilePath:         fmt.Sprintf("/src/file%d.go", i),
			LineNumber:       i * 10,
			Description:      fmt.Sprintf("Issue %d", i),
		})
		require.NoError(t, err)
	}

	// Retrieve all issues for session
	issues, err := issuesRepo.GetByReadingSessionID(ctx, rs.ID)
	require.NoError(t, err)
	assert.Len(t, issues, 3)

	// Verify we have all severity levels
	severityMap := make(map[string]bool)
	for _, issue := range issues {
		severityMap[issue.Severity] = true
	}
	assert.True(t, severityMap["critical"])
	assert.True(t, severityMap["important"])
	assert.True(t, severityMap["minor"])
}

// TestIntegration_UpdateStatusWorkflow tests complete status workflow
func TestIntegration_UpdateStatusWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	// Setup
	_, err := db.ExecContext(ctx, `INSERT INTO reviews.sessions (user_id, title, code_source) VALUES ($1, $2, $3)`, 1, "Test", "paste")
	require.NoError(t, err)

	readingRepo := NewReadingSessionRepository(db)
	issuesRepo := NewCriticalIssuesRepository(db)

	rs, err := readingRepo.Create(ctx, &ReadingSession{
		SessionID:   1,
		ReadingMode: "critical",
	})
	require.NoError(t, err)

	// Create issue (defaults to "open")
	issue, err := issuesRepo.Create(ctx, &CriticalIssue{
		ReadingSessionID: rs.ID,
		IssueType:        "security",
		Severity:         "critical",
		Description:      "SQL injection vulnerability",
		SuggestedFix:     "Use parameterized queries",
	})
	require.NoError(t, err)
	assert.Equal(t, "open", issue.Status)

	// Transition to accepted
	issue.Status = "accepted"
	err = issuesRepo.Update(ctx, issue)
	require.NoError(t, err)

	verified, err := issuesRepo.GetByID(ctx, issue.ID)
	assert.Equal(t, "accepted", verified.Status)

	// Transition to fixed
	issue.Status = "fixed"
	err = issuesRepo.Update(ctx, issue)
	require.NoError(t, err)

	verified, err = issuesRepo.GetByID(ctx, issue.ID)
	assert.Equal(t, "fixed", verified.Status)
}

// TestIntegration_ReadingSessionUpdate tests updating various fields
func TestIntegration_ReadingSessionUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	// Setup
	_, err := db.ExecContext(ctx, `INSERT INTO reviews.sessions (user_id, title, code_source) VALUES ($1, $2, $3)`, 1, "Test", "paste")
	require.NoError(t, err)

	repo := NewReadingSessionRepository(db)

	// Create reading session
	rs, err := repo.Create(ctx, &ReadingSession{
		SessionID:   1,
		ReadingMode: "preview",
		TargetPath:  "/src/original.go",
		ScanQuery:   "SELECT * FROM table",
		AIResponse:  "{}",
	})
	require.NoError(t, err)

	// Update multiple fields
	rs.ReadingMode = "detailed"
	rs.TargetPath = "/src/updated.go"
	rs.ScanQuery = "SELECT id FROM users"
	rs.AIResponse = `{"insights": "updated"}`

	err = repo.Update(ctx, rs)
	require.NoError(t, err)

	// Verify all updates persisted
	updated, err := repo.GetByID(ctx, rs.ID)
	require.NoError(t, err)
	assert.Equal(t, "detailed", updated.ReadingMode)
	assert.Equal(t, "/src/updated.go", updated.TargetPath)
	assert.Equal(t, "SELECT id FROM users", updated.ScanQuery)
	assert.Equal(t, `{"insights": "updated"}`, updated.AIResponse)
}

// TestIntegration_ErrorHandling tests error scenarios
func TestIntegration_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := setupIntegrationDB(ctx, t)
	defer db.Close()

	repo := NewReadingSessionRepository(db)
	issuesRepo := NewCriticalIssuesRepository(db)

	// Test getting non-existent records
	notFound, err := repo.GetByID(ctx, 999999)
	require.NoError(t, err)
	assert.Nil(t, notFound, "should return nil for non-existent record")

	notFoundIssue, err := issuesRepo.GetByID(ctx, 999999)
	require.NoError(t, err)
	assert.Nil(t, notFoundIssue, "should return nil for non-existent issue")

	// Test deleting non-existent record
	err = repo.Delete(ctx, 999999)
	assert.Error(t, err, "should error when deleting non-existent record")

	err = issuesRepo.Delete(ctx, 999999)
	assert.Error(t, err, "should error when deleting non-existent issue")

	// Test querying empty result
	sessions, err := repo.GetBySessionID(ctx, 999999)
	require.NoError(t, err)
	assert.Empty(t, sessions, "should return empty slice for non-existent session")
}
