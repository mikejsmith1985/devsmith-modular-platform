package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikejsmith1985/devsmith-modular-platform/cmd/logs/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
	"github.com/stretchr/testify/assert"
)

func setupTestPool(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	config, err := pgxpool.ParseConfig("postgres://devsmith:devsmith@localhost:5432/devsmith_test?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to parse database configuration: %v", err)
	}
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return db
}

func createTestGinContext(req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c, w
}

func decodeResponseBody(body io.ReadCloser, v interface{}) {
	defer body.Close()
	json.NewDecoder(body).Decode(v)
}

func TestLogsService_Integration(t *testing.T) {
	// Setup test database
	testDB := setupTestPool(t)
	repo := db.NewLogRepository(testDB)
	service := services.NewLogService(repo)
	handler := handlers.NewLogHandler(service)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		c.Set("user_id", int64(1)) // Inject user_id into context
		handler.IngestLog(c)
	}))
	defer srv.Close()

	// Test creating a log entry
	resp, err := http.Post(srv.URL, "application/json", strings.NewReader(`{
		"user_id": 1,
		"service": "portal",
		"level": "info",
		"message": "Integration test log",
		"metadata": "{\"key\": \"value\"}"
	}`))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "id")
}

func TestLogHandler_IngestLog(t *testing.T) {
	testPool := setupTestPool(t)
	defer testPool.Close()

	logRepo := db.NewLogRepository(testPool)
	logService := services.NewLogService(logRepo)
	logHandler := handlers.NewLogHandler(logService)

	// Create a test HTTP request
	requestBody := `{"service": "test-service", "level": "info", "message": "Test log message"}`
	req, err := http.NewRequest("POST", "/api/logs", strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a test Gin context
	c, w := createTestGinContext(req)
	c.Set("user_id", int64(1)) // Inject user_id into context

	// Call the handler
	logHandler.IngestLog(c)

	// Assert the response
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "id")
}

func TestLogsService_RetrieveLogs(t *testing.T) {
	testDB := setupTestPool(t)
	repo := db.NewLogRepository(testDB)
	service := services.NewLogService(repo)
	handler := handlers.NewLogHandler(service)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		handler.GetLogs(c)
	}))
	defer srv.Close()

	// Insert a log entry
	repo.Create(context.Background(), &models.LogEntry{
		UserID:   1,
		Service:  "test-service",
		Level:    "info",
		Message:  "Integration test log",
		Metadata: []byte(`{"key":"value"}`),
	})

	// Test retrieving log entries
	resp, err := http.Get(srv.URL + "/logs?service=test-service")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	var logs []models.LogEntry
	decodeResponseBody(resp.Body, &logs)
	assert.NotEmpty(t, logs)
	assert.Equal(t, "test-service", logs[0].Service)
}
