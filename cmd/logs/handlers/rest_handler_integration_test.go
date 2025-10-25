//go:build integration
// +build integration

package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/service"
)

// setupPostgresContainer starts a PostgreSQL container for testing
func setupPostgresContainer(t *testing.T) (string, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "logs_test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	connStr := fmt.Sprintf("postgres://test:test@%s:%s/logs_test?sslmode=disable", host, port.Port())

	// Wait for database to be ready
	var dbConn *sql.DB
	for i := 0; i < 30; i++ {
		dbConn, err = sql.Open("postgres", connStr)
		if err == nil {
			err = dbConn.Ping()
			if err == nil {
				break
			}
		}
		time.Sleep(time.Second)
	}
	require.NoError(t, err)
	dbConn.Close()

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		container.Terminate(ctx)
	}

	return connStr, cleanup
}

// setupTestDatabase creates the schema and returns a database connection
func setupTestDatabase(t *testing.T, connStr string) *sql.DB {
	dbConn, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	// Create schema
	schema := `
		CREATE SCHEMA IF NOT EXISTS logs;

		CREATE TABLE IF NOT EXISTS logs.entries (
			id BIGSERIAL PRIMARY KEY,
			service TEXT NOT NULL,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			metadata JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_logs_entries_created_at ON logs.entries(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_logs_entries_service ON logs.entries(service);
		CREATE INDEX IF NOT EXISTS idx_logs_entries_level ON logs.entries(level);
		CREATE INDEX IF NOT EXISTS idx_logs_entries_metadata ON logs.entries USING GIN(metadata);
	`

	_, err = dbConn.Exec(schema)
	require.NoError(t, err)

	return dbConn
}

// TestIntegration_PostLogs_InsertsIntoDatabase tests POST /api/logs stores in DB
func TestIntegration_PostLogs_InsertsIntoDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	connStr, cleanup := setupPostgresContainer(t)
	defer cleanup()

	dbConn := setupTestDatabase(t, connStr)
	defer dbConn.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	logRepo := db.NewLogRepository(dbConn)
	logService := service.NewLogService(logRepo)

	router.POST("/api/logs", PostLogs(logService))

	body := map[string]interface{}{
		"service": "portal",
		"level":   "info",
		"message": "User logged in",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/logs", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify in database
	var count int
	err := dbConn.QueryRow("SELECT COUNT(*) FROM logs.entries WHERE service = 'portal' AND level = 'info'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestIntegration_GetLogs_ReturnsFromDatabase tests GET /api/logs retrieves from DB
func TestIntegration_GetLogs_ReturnsFromDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	connStr, cleanup := setupPostgresContainer(t)
	defer cleanup()

	dbConn := setupTestDatabase(t, connStr)
	defer dbConn.Close()

	// Insert test data
	_, err := dbConn.Exec(
		"INSERT INTO logs.entries (service, level, message) VALUES ($1, $2, $3)",
		"review", "error", "Review analysis failed",
	)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	logRepo := db.NewLogRepository(dbConn)
	logService := service.NewLogService(logRepo)

	router.GET("/api/logs", GetLogs(logService))

	req := httptest.NewRequest("GET", "/api/logs?service=review", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1), resp["count"])

	entries := resp["entries"].([]interface{})
	assert.Greater(t, len(entries), 0)
	entry := entries[0].(map[string]interface{})
	assert.Equal(t, "review", entry["service"])
	assert.Equal(t, "error", entry["level"])
}

// TestIntegration_GetLogByID_ReturnsFromDatabase tests GET /api/logs/:id retrieves from DB
func TestIntegration_GetLogByID_ReturnsFromDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	connStr, cleanup := setupPostgresContainer(t)
	defer cleanup()

	dbConn := setupTestDatabase(t, connStr)
	defer dbConn.Close()

	// Insert test data and get ID
	var id int64
	err := dbConn.QueryRow(
		"INSERT INTO logs.entries (service, level, message) VALUES ($1, $2, $3) RETURNING id",
		"analytics", "warn", "Slow query detected",
	).Scan(&id)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	logRepo := db.NewLogRepository(dbConn)
	logService := service.NewLogService(logRepo)

	router.GET("/api/logs/:id", GetLogByID(logService))

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/logs/%d", id), http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var entry map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &entry)
	assert.Equal(t, float64(id), entry["id"])
	assert.Equal(t, "analytics", entry["service"])
	assert.Equal(t, "warn", entry["level"])
}

// TestIntegration_GetStats_AggregatesFromDatabase tests GET /api/logs/stats
func TestIntegration_GetStats_AggregatesFromDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	connStr, cleanup := setupPostgresContainer(t)
	defer cleanup()

	dbConn := setupTestDatabase(t, connStr)
	defer dbConn.Close()

	// Insert test data
	testData := []struct {
		service string
		level   string
		message string
	}{
		{"portal", "info", "Login success"},
		{"portal", "info", "Logout success"},
		{"portal", "error", "Auth failed"},
		{"review", "warn", "Slow analysis"},
		{"review", "error", "Timeout"},
	}

	for _, data := range testData {
		_, err := dbConn.Exec(
			"INSERT INTO logs.entries (service, level, message) VALUES ($1, $2, $3)",
			data.service, data.level, data.message,
		)
		require.NoError(t, err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()

	logRepo := db.NewLogRepository(dbConn)
	logService := service.NewLogService(logRepo)

	router.GET("/api/logs/stats", GetStats(logService))

	req := httptest.NewRequest("GET", "/api/logs/stats", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &stats)

	assert.Equal(t, float64(5), stats["total"])

	byLevel := stats["by_level"].(map[string]interface{})
	assert.Equal(t, float64(2), byLevel["info"])
	assert.Equal(t, float64(2), byLevel["error"])
	assert.Equal(t, float64(1), byLevel["warn"])

	byService := stats["by_service"].(map[string]interface{})
	assert.Equal(t, float64(3), byService["portal"])
	assert.Equal(t, float64(2), byService["review"])
}

// TestIntegration_GetLogs_FiltersByService tests filtering by service
func TestIntegration_GetLogs_FiltersByService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	connStr, cleanup := setupPostgresContainer(t)
	defer cleanup()

	dbConn := setupTestDatabase(t, connStr)
	defer dbConn.Close()

	// Insert mixed data
	for i := 0; i < 3; i++ {
		dbConn.Exec("INSERT INTO logs.entries (service, level, message) VALUES ($1, $2, $3)",
			"portal", "info", fmt.Sprintf("Portal log %d", i))
	}
	for i := 0; i < 2; i++ {
		dbConn.Exec("INSERT INTO logs.entries (service, level, message) VALUES ($1, $2, $3)",
			"review", "error", fmt.Sprintf("Review log %d", i))
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()

	logRepo := db.NewLogRepository(dbConn)
	logService := service.NewLogService(logRepo)

	router.GET("/api/logs", GetLogs(logService))

	req := httptest.NewRequest("GET", "/api/logs?service=portal", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(3), resp["count"])
}

// TestIntegration_GetLogs_FiltersByLevel tests filtering by level
func TestIntegration_GetLogs_FiltersByLevel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	connStr, cleanup := setupPostgresContainer(t)
	defer cleanup()

	dbConn := setupTestDatabase(t, connStr)
	defer dbConn.Close()

	// Insert mixed data
	for i := 0; i < 4; i++ {
		dbConn.Exec("INSERT INTO logs.entries (service, level, message) VALUES ($1, $2, $3)",
			"portal", "info", fmt.Sprintf("Info %d", i))
	}
	for i := 0; i < 3; i++ {
		dbConn.Exec("INSERT INTO logs.entries (service, level, message) VALUES ($1, $2, $3)",
			"portal", "error", fmt.Sprintf("Error %d", i))
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()

	logRepo := db.NewLogRepository(dbConn)
	logService := service.NewLogService(logRepo)

	router.GET("/api/logs", GetLogs(logService))

	req := httptest.NewRequest("GET", "/api/logs?level=error", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(3), resp["count"])
}

// TestIntegration_DeleteLogs_DeletesFromDatabase tests DELETE /api/logs
func TestIntegration_DeleteLogs_DeletesFromDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	connStr, cleanup := setupPostgresContainer(t)
	defer cleanup()

	dbConn := setupTestDatabase(t, connStr)
	defer dbConn.Close()

	// Insert old logs
	oldTime := time.Now().Add(-48 * time.Hour)
	dbConn.Exec(
		"INSERT INTO logs.entries (service, level, message, created_at) VALUES ($1, $2, $3, $4)",
		"portal", "info", "Old log", oldTime,
	)

	// Insert recent log
	recentTime := time.Now().Add(-1 * time.Hour)
	dbConn.Exec(
		"INSERT INTO logs.entries (service, level, message, created_at) VALUES ($1, $2, $3, $4)",
		"portal", "info", "Recent log", recentTime,
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	logRepo := db.NewLogRepository(dbConn)
	logService := service.NewLogService(logRepo)

	router.DELETE("/api/logs", DeleteLogs(logService))

	body := map[string]interface{}{"before": oldTime.Add(1 * time.Hour).Format(time.RFC3339)}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("DELETE", "/api/logs", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify old log deleted, recent log remains
	var count int
	dbConn.QueryRow("SELECT COUNT(*) FROM logs.entries").Scan(&count)
	assert.Equal(t, 1, count)
}
