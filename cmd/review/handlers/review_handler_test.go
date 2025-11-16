package cmd_review_handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	review_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockReviewService struct {
	mock.Mock
}

func (m *MockReviewService) GetReview(ctx context.Context, id int64) (*review_models.Review, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*review_models.Review), args.Error(1)
}

func (m *MockReviewService) CreateReview(ctx context.Context, review *review_db.Review) (*review_db.Review, error) {
	args := m.Called(ctx, review)
	return args.Get(0).(*review_db.Review), args.Error(1)
}

type MockScanService struct {
	mock.Mock
}

func (m *MockScanService) AnalyzeScan(ctx context.Context, query string, code string) (*review_models.ScanModeOutput, error) {
	args := m.Called(ctx, query, code)
	return args.Get(0).(*review_models.ScanModeOutput), args.Error(1)
}

func TestGetScanAnalysis(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reviewService := new(MockReviewService)
	scanService := new(MockScanService)
	handler := &ReviewHandler{
		reviewService: reviewService,
		scanService:   scanService,
	}

	reviewService.On("GetReview", mock.Anything, int64(1)).Return(&review_models.Review{ID: 1}, nil)
	scanService.On("AnalyzeScan", mock.Anything, "auth", "").Return(&review_models.ScanModeOutput{}, nil)

	r := gin.Default()
	r.GET("/api/reviews/:id/scan", handler.GetScanAnalysis)

	req, _ := http.NewRequest("GET", "/api/reviews/1/scan?q=auth", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetScanAnalysis_WithDifferentQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reviewService := new(MockReviewService)
	scanService := new(MockScanService)
	handler := &ReviewHandler{
		reviewService: reviewService,
		scanService:   scanService,
	}

	reviewService.On("GetReview", mock.Anything, int64(1)).Return(&review_models.Review{ID: 1}, nil)
	scanService.On("AnalyzeScan", mock.Anything, "database", "").Return(&review_models.ScanModeOutput{}, nil)

	r := gin.Default()
	r.GET("/api/reviews/:id/scan", handler.GetScanAnalysis)

	req, _ := http.NewRequest("GET", "/api/reviews/1/scan?q=database", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMockReviewService_GetReview(t *testing.T) {
	mockService := new(MockReviewService)
	mockService.On("GetReview", mock.Anything, int64(1)).Return(&review_models.Review{ID: 1}, nil)

	review, err := mockService.GetReview(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, review)
	assert.Equal(t, int64(1), review.ID)
}

func TestMockReviewService_CreateReview(t *testing.T) {
	mockService := new(MockReviewService)
	testReview := &review_db.Review{UserID: 1, Title: "Test"}
	mockService.On("CreateReview", mock.Anything, testReview).Return(testReview, nil)

	created, err := mockService.CreateReview(context.Background(), testReview)

	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, "Test", created.Title)
}

func TestMockScanService_AnalyzeScan(t *testing.T) {
	mockService := new(MockScanService)
	mockService.On("AnalyzeScan", mock.Anything, "query", "").Return(&review_models.ScanModeOutput{}, nil)

	result, err := mockService.AnalyzeScan(context.Background(), "query", "")

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGetScanAnalysis_ScanServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reviewService := new(MockReviewService)
	scanService := new(MockScanService)
	handler := &ReviewHandler{
		reviewService: reviewService,
		scanService:   scanService,
	}

	reviewService.On("GetReview", mock.Anything, int64(1)).Return(&review_models.Review{ID: 1}, nil)
	scanService.On("AnalyzeScan", mock.Anything, "test", "").Return((*review_models.ScanModeOutput)(nil), errors.New("analysis error"))

	r := gin.Default()
	r.GET("/api/reviews/:id/scan", handler.GetScanAnalysis)

	req, _ := http.NewRequest("GET", "/api/reviews/1/scan?q=test", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestMockServices_MultipleInteractions(t *testing.T) {
	reviewService := new(MockReviewService)
	scanService := new(MockScanService)

	// Set up multiple expectations
	reviewService.On("GetReview", mock.Anything, int64(1)).Return(&review_models.Review{ID: 1}, nil)
	reviewService.On("GetReview", mock.Anything, int64(2)).Return(&review_models.Review{ID: 2}, nil)
	scanService.On("AnalyzeScan", mock.Anything, "auth", "").Return(&review_models.ScanModeOutput{}, nil)
	scanService.On("AnalyzeScan", mock.Anything, "db", "").Return(&review_models.ScanModeOutput{}, nil)

	// Call multiple times
	r1, _ := reviewService.GetReview(context.Background(), 1)
	r2, _ := reviewService.GetReview(context.Background(), 2)
	s1, _ := scanService.AnalyzeScan(context.Background(), "auth", "")
	s2, _ := scanService.AnalyzeScan(context.Background(), "db", "")

	assert.Equal(t, int64(1), r1.ID)
	assert.Equal(t, int64(2), r2.ID)
	assert.NotNil(t, s1)
	assert.NotNil(t, s2)
}
