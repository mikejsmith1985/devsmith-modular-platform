package handlers

import (
	"context"
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
