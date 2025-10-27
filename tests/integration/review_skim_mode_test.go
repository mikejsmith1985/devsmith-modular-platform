package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mikejsmith1985/devsmith-modular-platform/cmd/review/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/instrumentation"
	reviewdb "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// createTestInstrumentationLogger creates a dummy instrumentation logger for tests
func createTestInstrumentationLogger() *instrumentation.ServiceInstrumentationLogger {
	return instrumentation.NewServiceInstrumentationLogger("review-test", "http://localhost:8082")
}

func setupTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("REVIEW_TEST_DB_URL")
	if dbURL == "" {
		// Use default test DB URL if not set
		dbURL = "postgres://devsmith:devsmith@localhost:5432/devsmith_test?sslmode=disable"
		t.Log("REVIEW_TEST_DB_URL not set, using default test DB URL")
	}
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		t.Skipf("skipping: failed to open test db: %v", err)
	}

	// Ping to verify connection
	if err := db.Ping(); err != nil {
		t.Skipf("skipping: test database not available: %v", err)
	}

	return db
}

func TestSkimMode_Integration(t *testing.T) {
	// Setup DB and insert a test review session
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Error closing test DB: %v", err)
		}
	}()
	_, err := db.Exec(`INSERT INTO reviews.sessions (id, user_id, title, code_source, github_repo, github_branch, pasted_code) VALUES (1001, 1, 'Test Review', 'github', 'mikejsmith1985/devsmith-modular-platform', 'main', '') ON CONFLICT (id) DO NOTHING`)
	assert.NoError(t, err)

	reviewRepo := reviewdb.NewReviewRepository(db)
	ollamaClient := &OllamaClientStub{}
	analysisRepo := &MockAnalysisRepository{}
	mockLogger := &testutils.MockLogger{}
	skimService := services.NewSkimService(ollamaClient, analysisRepo, mockLogger)
	reviewService := services.NewReviewService(skimService, reviewRepo)
	previewService := services.NewPreviewService(mockLogger)
	scanService := services.NewScanService(ollamaClient, analysisRepo, mockLogger)

	// Create a dummy instrumentation logger for testing
	instrLogger := createTestInstrumentationLogger()

	handler := handlers.NewReviewHandler(reviewService, previewService, skimService, scanService, instrLogger)

	r := gin.Default()
	r.GET("/api/reviews/:id/skim", handler.GetSkimAnalysis)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/reviews/1001/skim", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var output map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &output)
	assert.NoError(t, err)
	assert.Contains(t, output, "Summary") // Updated to match the capitalized field name
	assert.Contains(t, output, "functions")
	assert.Contains(t, output, "interfaces")
	assert.Contains(t, output, "data_models")
	assert.Contains(t, output, "workflows")
}

type OllamaClientStub struct{}

func (o *OllamaClientStub) Generate(_ context.Context, _ string) (string, error) {
	return `{"functions":[],"interfaces":[],"data_models":[],"workflows":[],"summary":"Stubbed AI output"}`, nil
}

type MockAnalysisRepository struct{}

func (m *MockAnalysisRepository) FindByReviewAndMode(_ context.Context, reviewID int64, mode string) (*models.AnalysisResult, error) {
	if reviewID == 1001 && mode == models.SkimMode {
		return &models.AnalysisResult{
			ReviewID: reviewID,
			Mode:     mode,
			Summary:  "Cached summary for test",
			Metadata: `{"functions":[],"interfaces":[],"data_models":[],"workflows":[],"summary":"Cached summary for test"}`,
		}, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *MockAnalysisRepository) Create(_ context.Context, _ *models.AnalysisResult) error {
	return nil
}

func TestLoginFlow_EndToEnd(t *testing.T) {
	// This is an end-to-end test that requires the portal service to be running
	// Check if service is available first
	healthResp, err := http.Get("http://localhost:3000/health")
	if err != nil || healthResp.StatusCode != http.StatusOK {
		t.Skip("Portal service not running on localhost:3000, skipping E2E test")
	}
	defer healthResp.Body.Close()

	// Create HTTP client that doesn't follow redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	// Act - Make request to GitHub login endpoint
	resp, err := client.Get("http://localhost:3000/auth/github/login")
	assert.NoError(t, err, "Request to /auth/github/login should not error")
	defer resp.Body.Close()

	// Assert
	assert.Equal(t, http.StatusFound, resp.StatusCode, "Should redirect to GitHub OAuth")

	location := resp.Header.Get("Location")
	assert.Contains(t, location, "https://github.com/login/oauth/authorize", "Should redirect to GitHub OAuth URL")
	assert.Contains(t, location, "client_id=", "Should include client ID in redirect URL")
}
