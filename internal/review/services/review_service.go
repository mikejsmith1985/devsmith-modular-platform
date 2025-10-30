package review_services

import (
	"context"

	review_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// ReviewService is a placeholder for wiring SkimService (expand as needed)
type ReviewService struct {
	skimService *SkimService
	reviewRepo  *review_db.ReviewRepository
}

// NewReviewService creates a new ReviewService with SkimService and ReviewRepository
func NewReviewService(skimService *SkimService, reviewRepo *review_db.ReviewRepository) *ReviewService {
	return &ReviewService{
		skimService: skimService,
		reviewRepo:  reviewRepo,
	}
}

// GetReview fetches a review session by ID from the database
func (r *ReviewService) GetReview(ctx context.Context, id int64) (*review_models.Review, error) {
	dbReview, err := r.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Map review_db.Review to review_models.Review
	return &review_models.Review{
		ID:           dbReview.ID,
		Title:        dbReview.Title,
		CodeSource:   dbReview.CodeSource,
		CreatedAt:    dbReview.CreatedAt,
		LastAccessed: dbReview.LastAccessed,
	}, nil
}

// CreateReview creates a new review session
func (r *ReviewService) CreateReview(ctx context.Context, review *review_db.Review) (*review_db.Review, error) {
	return r.reviewRepo.Create(ctx, review)
}
