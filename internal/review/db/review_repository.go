package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Review struct {
	ID           int64
	UserID       int64
	Title        string
	CodeSource   string
	GithubRepo   string
	GithubBranch string
	PastedCode   string
	CreatedAt    string
	LastAccessed string
}

type ReviewRepository struct {
	DB *sql.DB
}

func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{DB: db}
}

// Create inserts a new review session and returns the created Review with ID
func (r *ReviewRepository) Create(ctx context.Context, review *Review) (*Review, error) {
	query := `INSERT INTO reviews.sessions (user_id, title, code_source, github_repo, github_branch, pasted_code) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, last_accessed`
	err := r.DB.QueryRowContext(ctx, query,
		review.UserID,
		review.Title,
		review.CodeSource,
		review.GithubRepo,
		review.GithubBranch,
		review.PastedCode,
	).Scan(&review.ID, &review.CreatedAt, &review.LastAccessed)
	if err != nil {
		return nil, fmt.Errorf("db: failed to create review: %w", err)
	}
	return review, nil
}

func (r *ReviewRepository) GetByID(ctx context.Context, id int64) (*Review, error) {
	row := r.DB.QueryRowContext(ctx, `SELECT id, user_id, title, code_source, github_repo, github_branch, pasted_code, created_at, last_accessed FROM reviews.sessions WHERE id = $1`, id)
	var review Review
	err := row.Scan(&review.ID, &review.UserID, &review.Title, &review.CodeSource, &review.GithubRepo, &review.GithubBranch, &review.PastedCode, &review.CreatedAt, &review.LastAccessed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("db: failed to get review by id: %w", err)
	}
	return &review, nil
}
