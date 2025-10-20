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
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/mikejsmith1985/devsmith-modular-platform/cmd/review/handlers"
	reviewdb "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("REVIEW_TEST_DB_URL")
	if dbURL == "" {
		// Use default test DB URL if not set
		dbURL = "postgres://devsmith:devsmith@localhost:5432/devsmith_test?sslmode=disable"
		t.Log("REVIEW_TEST_DB_URL not set, using default test DB URL")
	}
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	return db
}

func TestSkimMode_Integration(t *testing.T) {
	// Setup DB and insert a test review session
	db := setupTestDB(t)
	defer db.Close()
	_, err := db.Exec(`INSERT INTO reviews.sessions (id, user_id, title, code_source, github_repo, github_branch, pasted_code) VALUES (1001, 1, 'Test Review', 'github', 'mikejsmith1985/devsmith-modular-platform', 'main', '') ON CONFLICT (id) DO NOTHING`)
	assert.NoError(t, err)

	reviewRepo := reviewdb.NewReviewRepository(db)
	ollamaClient := &OllamaClientStub{}
	analysisRepo := &MockAnalysisRepository{}
	skimService := services.NewSkimService(ollamaClient, analysisRepo)
	reviewService := services.NewReviewService(skimService, reviewRepo)
	previewService := services.NewPreviewService()
	handler := handlers.NewReviewHandler(reviewService, previewService, skimService)

	r := gin.Default()
	r.GET("/api/reviews/:id/skim", handler.GetSkimAnalysis)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/reviews/1001/skim", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var output map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &output)
	assert.NoError(t, err)
	assert.Contains(t, output, "functions")
	assert.Contains(t, output, "interfaces")
	assert.Contains(t, output, "data_models")
	assert.Contains(t, output, "workflows")
	assert.Contains(t, output, "summary")
}

type OllamaClientStub struct{}

func (o *OllamaClientStub) Generate(ctx context.Context, prompt string) (string, error) {
	return `{"functions":[],"interfaces":[],"data_models":[],"workflows":[],"summary":"Stubbed AI output"}`, nil
}

type MockAnalysisRepository struct{}

func (m *MockAnalysisRepository) FindByReviewAndMode(ctx context.Context, reviewID int64, mode string) (*models.AnalysisResult, error) {
	return nil, fmt.Errorf("not found")
}
func (m *MockAnalysisRepository) Create(ctx context.Context, result *models.AnalysisResult) error {
	return nil
}
