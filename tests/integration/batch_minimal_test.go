package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	logs_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	internal_logs_handlers "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/handlers"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

// TestBatchIngestion_Minimal tests the basic batch ingestion flow with correct API
func TestBatchIngestion_Minimal(t *testing.T) {
	// Setup test database
	testDB := setupMinimalTestDatabase(t)
	defer teardownMinimalTestDatabase(t, testDB)

	// Create repositories and services
	projectRepo := logs_db.NewProjectRepository(testDB)
	logRepo := logs_db.NewLogEntryRepository(testDB)
	projectService := logs_services.NewProjectService(projectRepo)

	// Create batch handler with all 3 required parameters
	batchHandler := internal_logs_handlers.NewBatchHandler(logRepo, projectRepo, projectService)

	// Create test project with API key (correct signature: userID, request)
	testUserID := 1
	createResp, err := projectService.CreateProject(context.Background(), testUserID, &logs_models.CreateProjectRequest{
		Name:        "Test Project",
		Slug:        "test-project",
		Description: "Integration test project",
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotEmpty(t, createResp.APIKey, "API key should be returned")

	apiKey := createResp.APIKey

	// Create batch request using actual handler types
	batchRequest := internal_logs_handlers.BatchLogRequest{
		ProjectSlug: "test-project",
		Logs: []internal_logs_handlers.BatchLogEntry{
			{
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
				Level:       "info",
				Message:     "Test log message 1",
				ServiceName: "test-service",
				Context:     map[string]interface{}{"test": "data"},
			},
			{
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
				Level:       "error",
				Message:     "Test error message",
				ServiceName: "test-service",
				Context:     map[string]interface{}{"error_code": 500},
			},
		},
	}

	// Marshal request
	body, err := json.Marshal(batchRequest)
	require.NoError(t, err)

	// Create HTTP request with Bearer token
	req := httptest.NewRequest("POST", "/api/logs/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	batchHandler.IngestBatch(c)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code, "Expected 201 Created status")

	var response internal_logs_handlers.BatchLogResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 2, response.Accepted, "Should accept 2 logs")
	assert.Contains(t, response.Message, "Successfully ingested", "Should have success message")

	t.Logf("✅ Minimal batch ingestion test passed - ingested %d logs", response.Accepted)
}

// setupMinimalTestDatabase creates a test database connection with minimal schema
func setupMinimalTestDatabase(t *testing.T) *sql.DB {
	// Get test database URL from environment or use default
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://devsmith:devsmith@localhost:5432/devsmith_test?sslmode=disable"
	}

	// Open database connection
	testDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to test database: %v", err)
		return nil
	}

	// Verify connection
	if err := testDB.Ping(); err != nil {
		t.Skipf("Skipping test: test database not available: %v", err)
		return nil
	}

	// Configure connection pool (lighter for tests)
	testDB.SetMaxOpenConns(5)
	testDB.SetMaxIdleConns(2)

	// Run migrations to set up schema
	if err := runMinimalMigrations(testDB); err != nil {
		testDB.Close()
		t.Fatalf("Failed to run migrations: %v", err)
		return nil
	}

	// Clean up any existing test data
	if err := cleanupMinimalTestData(testDB); err != nil {
		testDB.Close()
		t.Fatalf("Failed to clean up test data: %v", err)
		return nil
	}

	return testDB
}

// teardownMinimalTestDatabase cleans up test database
func teardownMinimalTestDatabase(t *testing.T, testDB *sql.DB) {
	if testDB == nil {
		return
	}

	// Clean up test data
	if err := cleanupMinimalTestData(testDB); err != nil {
		t.Logf("Warning: failed to clean up test data: %v", err)
	}

	// Close database connection
	if err := testDB.Close(); err != nil {
		t.Logf("Warning: failed to close database: %v", err)
	}
}

// runMinimalMigrations creates the minimal schema needed for batch ingestion tests
func runMinimalMigrations(db *sql.DB) error {
	// Create logs schema if it doesn't exist
	_, err := db.Exec("CREATE SCHEMA IF NOT EXISTS logs")
	if err != nil {
		return fmt.Errorf("failed to create logs schema: %w", err)
	}

	// Create projects table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS logs.projects (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) NOT NULL UNIQUE,
			description TEXT,
			repository_url VARCHAR(500),
			api_key_hash VARCHAR(255) NOT NULL,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create projects table: %w", err)
	}

	// Create log entries table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS logs.entries (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			project_id INTEGER REFERENCES logs.projects(id) ON DELETE CASCADE,
			service VARCHAR(255) NOT NULL,
			service_name VARCHAR(255),
			level VARCHAR(20) NOT NULL,
			message TEXT NOT NULL,
			metadata JSONB,
			context JSONB,
			tags TEXT[],
			timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create entries table: %w", err)
	}

	// Create index on project_id for faster queries
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_entries_project_id ON logs.entries(project_id)
	`)
	if err != nil {
		return fmt.Errorf("failed to create project_id index: %w", err)
	}

	return nil
}

// cleanupMinimalTestData removes all test data from database
func cleanupMinimalTestData(db *sql.DB) error {
	// Delete in order (entries first due to foreign key)
	_, err := db.Exec("DELETE FROM logs.entries")
	if err != nil {
		return fmt.Errorf("failed to delete entries: %w", err)
	}

	_, err = db.Exec("DELETE FROM logs.projects")
	if err != nil {
		return fmt.Errorf("failed to delete projects: %w", err)
	}

	return nil
}
