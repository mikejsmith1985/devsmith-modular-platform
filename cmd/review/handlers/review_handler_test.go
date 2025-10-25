package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockReviewService struct {
	mock.Mock
}

func (m *MockReviewService) GetReview(ctx context.Context, id int64) (*models.Review, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Review), args.Error(1)
}

func (m *MockReviewService) CreateReview(ctx context.Context, review *db.Review) (*db.Review, error) {
	args := m.Called(ctx, review)
	return args.Get(0).(*db.Review), args.Error(1)
}

type MockScanService struct {
	mock.Mock
}

func (m *MockScanService) AnalyzeScan(ctx context.Context, reviewID int64, query string) (*models.ScanModeOutput, error) {
	args := m.Called(ctx, reviewID, query)
	return args.Get(0).(*models.ScanModeOutput), args.Error(1)
}

func TestGetScanAnalysis(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reviewService := new(MockReviewService)
	scanService := new(MockScanService)
	handler := &ReviewHandler{
		reviewService: reviewService,
		scanService:   scanService,
	}

	reviewService.On("GetReview", mock.Anything, int64(1)).Return(&models.Review{ID: 1}, nil)
	scanService.On("AnalyzeScan", mock.Anything, int64(1), "auth").Return(&models.ScanModeOutput{}, nil)

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

	reviewService.On("GetReview", mock.Anything, int64(1)).Return(&models.Review{ID: 1}, nil)
	scanService.On("AnalyzeScan", mock.Anything, int64(1), "database").Return(&models.ScanModeOutput{}, nil)

	r := gin.Default()
	r.GET("/api/reviews/:id/scan", handler.GetScanAnalysis)

	req, _ := http.NewRequest("GET", "/api/reviews/1/scan?q=database", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMockReviewService_GetReview(t *testing.T) {
	mockService := new(MockReviewService)
	mockService.On("GetReview", mock.Anything, int64(1)).Return(&models.Review{ID: 1}, nil)

	review, err := mockService.GetReview(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, review)
	assert.Equal(t, int64(1), review.ID)
}

func TestMockReviewService_CreateReview(t *testing.T) {
	mockService := new(MockReviewService)
	testReview := &db.Review{UserID: 1, Title: "Test"}
	mockService.On("CreateReview", mock.Anything, testReview).Return(testReview, nil)

	created, err := mockService.CreateReview(context.Background(), testReview)

	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, "Test", created.Title)
}

func TestMockScanService_AnalyzeScan(t *testing.T) {
	mockService := new(MockScanService)
	mockService.On("AnalyzeScan", mock.Anything, int64(1), "query").Return(&models.ScanModeOutput{}, nil)

	result, err := mockService.AnalyzeScan(context.Background(), 1, "query")

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

	reviewService.On("GetReview", mock.Anything, int64(1)).Return(&models.Review{ID: 1}, nil)
	scanService.On("AnalyzeScan", mock.Anything, int64(1), "test").Return((*models.ScanModeOutput)(nil), errors.New("analysis error"))

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
	reviewService.On("GetReview", mock.Anything, int64(1)).Return(&models.Review{ID: 1}, nil)
	reviewService.On("GetReview", mock.Anything, int64(2)).Return(&models.Review{ID: 2}, nil)
	scanService.On("AnalyzeScan", mock.Anything, int64(1), "auth").Return(&models.ScanModeOutput{}, nil)
	scanService.On("AnalyzeScan", mock.Anything, int64(2), "db").Return(&models.ScanModeOutput{}, nil)

	// Call multiple times
	r1, _ := reviewService.GetReview(context.Background(), 1)
	r2, _ := reviewService.GetReview(context.Background(), 2)
	s1, _ := scanService.AnalyzeScan(context.Background(), 1, "auth")
	s2, _ := scanService.AnalyzeScan(context.Background(), 2, "db")

	assert.Equal(t, int64(1), r1.ID)
	assert.Equal(t, int64(2), r2.ID)
	assert.NotNil(t, s1)
	assert.NotNil(t, s2)
}
