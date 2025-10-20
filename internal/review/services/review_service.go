package services

import (
	"context"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
)

// ReviewService is a placeholder for wiring SkimService (expand as needed)
type ReviewService struct {
	skimService *SkimService
	reviewRepo  *db.ReviewRepository
}

// NewReviewService creates a new ReviewService with SkimService and ReviewRepository
func NewReviewService(skimService *SkimService, reviewRepo *db.ReviewRepository) *ReviewService {
	return &ReviewService{
		skimService: skimService,
		reviewRepo:  reviewRepo,
	}
}

// GetReview fetches a review session by ID from the database
func (r *ReviewService) GetReview(ctx context.Context, id int64) (*db.Review, error) {
	return r.reviewRepo.GetByID(ctx, id)
}

// CreateReview creates a new review session
func (r *ReviewService) CreateReview(ctx context.Context, review *db.Review) (*db.Review, error) {
	return r.reviewRepo.Create(ctx, review)
}
