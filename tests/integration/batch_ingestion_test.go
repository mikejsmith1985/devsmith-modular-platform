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
	logs_middleware "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/middleware"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

// TestBatchIngestion_ValidBatch tests successful batch log ingestion
func TestBatchIngestion_ValidBatch(t *testing.T) {
	// Setup test database and services
	testDB := setupTestDatabase(t)
	defer teardownTestDatabase(t, testDB)

	projectRepo := logs_db.NewProjectRepository(testDB)
	logRepo := logs_db.NewLogEntryRepository(testDB)
	projectService := logs_services.NewProjectService(projectRepo)
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
	require.NotEmpty(t, createResp.APIKey)

	apiKey := createResp.APIKey
	t.Logf("🔑 Generated API key: %s", apiKey)
	t.Logf("📋 Project created with ID: %d, Slug: %s", createResp.Project.ID, createResp.Project.Slug)

	// Create test batch of 100 logs using correct handler types
	batchEntries := make([]internal_logs_handlers.BatchLogEntry, 100)
	for i := 0; i < 100; i++ {
		batchEntries[i] = internal_logs_handlers.BatchLogEntry{
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			Level:       "info",
			Message:     fmt.Sprintf("Test log message %d", i),
			ServiceName: "test-service",
			Context:     map[string]interface{}{"index": i},
		}
	}

	batchRequest := internal_logs_handlers.BatchLogRequest{
		ProjectSlug: "test-project",
		Logs:        batchEntries,
	}

	// Marshal request
	body, err := json.Marshal(batchRequest)
	require.NoError(t, err)

	// Create HTTP request with X-API-Key header (matches SimpleAPITokenAuth middleware)
	req := httptest.NewRequest("POST", "/api/logs/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	// Create router with middleware (matches production setup)
	router := gin.New()
	router.POST("/api/logs/batch", logs_middleware.SimpleAPITokenAuth(projectRepo), batchHandler.IngestBatch)

	// Execute request through router (middleware included)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response internal_logs_handlers.BatchLogResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 100, response.Accepted)
	assert.Contains(t, response.Message, "Successfully ingested")

	t.Logf("✅ Successfully ingested %d logs", response.Accepted)
}

// TestBatchIngestion_InvalidAPIKey tests rejection of invalid API keys
func TestBatchIngestion_InvalidAPIKey(t *testing.T) {
	testDB := setupTestDatabase(t)
	defer teardownTestDatabase(t, testDB)

	projectRepo := logs_db.NewProjectRepository(testDB)
	logRepo := logs_db.NewLogEntryRepository(testDB)
	projectService := logs_services.NewProjectService(projectRepo)
	batchHandler := internal_logs_handlers.NewBatchHandler(logRepo, projectRepo, projectService)

	// Create batch with single log using correct handler types
	batchRequest := internal_logs_handlers.BatchLogRequest{
		ProjectSlug: "test-project",
		Logs: []internal_logs_handlers.BatchLogEntry{
			{
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
				Level:       "info",
				Message:     "Test message",
				ServiceName: "test-service",
				Context:     map[string]interface{}{},
			},
		},
	}

	body, err := json.Marshal(batchRequest)
	require.NoError(t, err)

	// Request with invalid API key using X-API-Key header (matches middleware)
	req := httptest.NewRequest("POST", "/api/logs/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "invalid_api_key_12345")

	// Create router with middleware to test auth rejection
	router := gin.New()
	router.POST("/api/logs/batch", logs_middleware.SimpleAPITokenAuth(projectRepo), batchHandler.IngestBatch)

	// Execute request through router
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid API key")
}

// TestBatchIngestion_MissingAuthHeader tests rejection when no auth header provided
func TestBatchIngestion_MissingAuthHeader(t *testing.T) {
	testDB := setupTestDatabase(t)
	defer teardownTestDatabase(t, testDB)

	projectRepo := logs_db.NewProjectRepository(testDB)
	logRepo := logs_db.NewLogEntryRepository(testDB)
	projectService := logs_services.NewProjectService(projectRepo)
	batchHandler := internal_logs_handlers.NewBatchHandler(logRepo, projectRepo, projectService)

	batchRequest := internal_logs_handlers.BatchLogRequest{
		ProjectSlug: "test-project",
		Logs: []internal_logs_handlers.BatchLogEntry{
			{
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
				Level:       "info",
				Message:     "Test message",
				ServiceName: "test-service",
				Context:     map[string]interface{}{},
			},
		},
	}

	body, err := json.Marshal(batchRequest)
	require.NoError(t, err)

	// Request WITHOUT X-API-Key header
	req := httptest.NewRequest("POST", "/api/logs/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Create router with middleware to test missing header rejection
	router := gin.New()
	router.POST("/api/logs/batch", logs_middleware.SimpleAPITokenAuth(projectRepo), batchHandler.IngestBatch)

	// Execute request through router
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Missing X-API-Key header")
}

// TestBatchIngestion_DeactivatedProject tests rejection when project is deactivated
func TestBatchIngestion_DeactivatedProject(t *testing.T) {
	testDB := setupTestDatabase(t)
	defer teardownTestDatabase(t, testDB)

	projectRepo := logs_db.NewProjectRepository(testDB)
	logRepo := logs_db.NewLogEntryRepository(testDB)
	projectService := logs_services.NewProjectService(projectRepo)
	batchHandler := internal_logs_handlers.NewBatchHandler(logRepo, projectRepo, projectService)

	// Create project with correct signature (userID, request)
	testUserID := 1
	createResp, err := projectService.CreateProject(context.Background(), testUserID, &logs_models.CreateProjectRequest{
		Name: "Test Project",
		Slug: "test-project",
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)

	apiKey := createResp.APIKey

	// Deactivate project (assuming DeactivateProject exists - may need userID parameter too)
	err = projectService.DeactivateProject(context.Background(), createResp.ID)
	require.NoError(t, err)

	// Try to ingest logs with deactivated project's API key
	batchRequest := internal_logs_handlers.BatchLogRequest{
		ProjectSlug: "test-project",
		Logs: []internal_logs_handlers.BatchLogEntry{
			{
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
				Level:       "info",
				Message:     "Test message",
				ServiceName: "test-service",
				Context:     map[string]interface{}{},
			},
		},
	}

	body, err := json.Marshal(batchRequest)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/logs/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	// Create router with middleware to test deactivated project rejection
	router := gin.New()
	router.POST("/api/logs/batch", logs_middleware.SimpleAPITokenAuth(projectRepo), batchHandler.IngestBatch)

	// Execute request through router
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify 403 Forbidden (deactivated project returns forbidden error from middleware)
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Project disabled")
}

// TestBatchIngestion_MaxBatchSize tests rejection when batch exceeds 1000 logs
func TestBatchIngestion_MaxBatchSize(t *testing.T) {
	testDB := setupTestDatabase(t)
	defer teardownTestDatabase(t, testDB)

	projectRepo := logs_db.NewProjectRepository(testDB)
	logRepo := logs_db.NewLogEntryRepository(testDB)
	projectService := logs_services.NewProjectService(projectRepo)
	batchHandler := internal_logs_handlers.NewBatchHandler(logRepo, projectRepo, projectService)

	// Create project with correct signature
	testUserID := 1
	createResp, err := projectService.CreateProject(context.Background(), testUserID, &logs_models.CreateProjectRequest{
		Name: "Test Project",
		Slug: "test-project",
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	apiKey := createResp.APIKey

	// Create batch with 1001 logs (exceeds max of 1000)
	batchEntries := make([]internal_logs_handlers.BatchLogEntry, 1001)
	for i := 0; i < 1001; i++ {
		batchEntries[i] = internal_logs_handlers.BatchLogEntry{
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			Level:       "info",
			Message:     fmt.Sprintf("Test log %d", i),
			ServiceName: "test-service",
			Context:     map[string]interface{}{"index": i},
		}
	}

	batchRequest := internal_logs_handlers.BatchLogRequest{
		ProjectSlug: "test-project",
		Logs:        batchEntries,
	}

	body, err := json.Marshal(batchRequest)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/logs/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	batchHandler.IngestBatch(c)

	// Verify 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Batch size exceeds maximum of 1000 logs")
}

// TestBatchIngestion_InvalidJSON tests rejection of malformed JSON
func TestBatchIngestion_InvalidJSON(t *testing.T) {
	testDB := setupTestDatabase(t)
	defer teardownTestDatabase(t, testDB)

	projectRepo := logs_db.NewProjectRepository(testDB)
	logRepo := logs_db.NewLogEntryRepository(testDB)
	projectService := logs_services.NewProjectService(projectRepo)
	batchHandler := internal_logs_handlers.NewBatchHandler(logRepo, projectRepo, projectService)

	testUserID := 1
	createResp, err := projectService.CreateProject(context.Background(), testUserID, &logs_models.CreateProjectRequest{
		Name: "Test Project",
		Slug: "test-project",
	})
	require.NoError(t, err)
	apiKey := createResp.APIKey

	// Send invalid JSON
	invalidJSON := []byte(`{"project_slug": "test-project", "logs": [{"level": "INFO", "message": "test"`)

	req := httptest.NewRequest("POST", "/api/logs/batch", bytes.NewReader(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	batchHandler.IngestBatch(c)

	// Verify 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBatchIngestion_Performance tests throughput for 1000-log batch
func TestBatchIngestion_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	testDB := setupTestDatabase(t)
	defer teardownTestDatabase(t, testDB)

	projectRepo := logs_db.NewProjectRepository(testDB)
	logRepo := logs_db.NewLogEntryRepository(testDB)
	projectService := logs_services.NewProjectService(projectRepo)
	batchHandler := internal_logs_handlers.NewBatchHandler(logRepo, projectRepo, projectService)

	testUserID := 1
	createResp, err := projectService.CreateProject(context.Background(), testUserID, &logs_models.CreateProjectRequest{
		Name: "Test Project",
		Slug: "test-project",
	})
	require.NoError(t, err)
	apiKey := createResp.APIKey

	// Create 1000-log batch
	batchEntries := make([]internal_logs_handlers.BatchLogEntry, 1000)
	for i := 0; i < 1000; i++ {
		batchEntries[i] = internal_logs_handlers.BatchLogEntry{
			Level:       "INFO",
			Message:     fmt.Sprintf("Performance test log %d", i),
			ServiceName: "perf-test",
			Context:     map[string]interface{}{"index": i, "batch": "perf"},
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}
	}

	batchRequest := internal_logs_handlers.BatchLogRequest{
		ProjectSlug: "test-project",
		Logs:        batchEntries,
	}

	body, err := json.Marshal(batchRequest)
	require.NoError(t, err)

	// Measure time to ingest
	start := time.Now()

	req := httptest.NewRequest("POST", "/api/logs/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	batchHandler.IngestBatch(c)

	duration := time.Since(start)

	// Verify success
	assert.Equal(t, http.StatusCreated, w.Code)

	// Performance target: < 500ms for 1000 logs
	assert.Less(t, duration.Milliseconds(), int64(500),
		"Batch ingestion of 1000 logs should complete in <500ms (actual: %dms)", duration.Milliseconds())

	t.Logf("Performance: Ingested 1000 logs in %dms (%.0f logs/sec)",
		duration.Milliseconds(),
		1000.0/(duration.Seconds()))
}

// Helper functions for test setup/teardown
func setupTestDatabase(t *testing.T) *sql.DB {
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

	// Configure connection pool
	testDB.SetMaxOpenConns(5)
	testDB.SetMaxIdleConns(2)

	// Run migrations to set up schema
	if err := runTestMigrations(testDB); err != nil {
		testDB.Close()
		t.Fatalf("Failed to run migrations: %v", err)
		return nil
	}

	// Clean up any existing test data
	if err := cleanupTestData(testDB); err != nil {
		testDB.Close()
		t.Fatalf("Failed to clean up test data: %v", err)
		return nil
	}

	return testDB
}

func teardownTestDatabase(t *testing.T, testDB *sql.DB) {
	if testDB == nil {
		return
	}

	// Clean up test data
	if err := cleanupTestData(testDB); err != nil {
		t.Logf("Warning: failed to clean up test data: %v", err)
	}

	// Close database connection
	if err := testDB.Close(); err != nil {
		t.Logf("Warning: failed to close database: %v", err)
	}
}

// runTestMigrations runs the necessary migrations for testing
func runTestMigrations(db *sql.DB) error {
	// Create logs schema if it doesn't exist
	_, err := db.Exec("CREATE SCHEMA IF NOT EXISTS logs")
	if err != nil {
		return fmt.Errorf("failed to create logs schema: %w", err)
	}

	// Drop existing tables to ensure clean schema (test isolation)
	_, _ = db.Exec("DROP TABLE IF EXISTS logs.entries CASCADE")
	_, _ = db.Exec("DROP TABLE IF EXISTS logs.projects CASCADE")

	// Create projects table with CORRECT schema (api_key_hash, not api_token!)
	_, err = db.Exec(`
		CREATE TABLE logs.projects (
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
		CREATE TABLE logs.entries (
			id BIGSERIAL PRIMARY KEY,
			project_id INTEGER REFERENCES logs.projects(id) ON DELETE CASCADE,
			level VARCHAR(20) NOT NULL,
			message TEXT NOT NULL,
			service_name VARCHAR(255) NOT NULL,
			timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
			metadata JSONB,
			context JSONB,
			tags TEXT[],
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

// cleanupTestData removes all test data from the database
func cleanupTestData(db *sql.DB) error {
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
