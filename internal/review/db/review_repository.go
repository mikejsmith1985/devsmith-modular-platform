// Package review_db provides database access for review sessions and repositories.
package review_db

import (
	"context"
	"database/sql"
	"fmt"
)

// Review represents a code review session in the database.
type Review struct {
	Title        string
	CodeSource   string
	GithubRepo   string
	GithubBranch string
	PastedCode   string
	CreatedAt    string
	LastAccessed string
	ID           int64
	UserID       int64
}

// ReviewRepository handles CRUD operations for Review sessions.
type ReviewRepository struct {
	DB *sql.DB
}

// NewReviewRepository creates a new ReviewRepository with the given database connection.
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

// GetByID retrieves a Review by its ID.
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

// ListByUserID retrieves all review sessions for a user with pagination support
func (r *ReviewRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*Review, int, error) {
	// Get total count
	var total int
	countErr := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM reviews.sessions WHERE user_id = $1`, userID).Scan(&total)
	if countErr != nil {
		return nil, 0, fmt.Errorf("db: failed to count reviews: %w", countErr)
	}

	// Get paginated results
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, user_id, title, code_source, github_repo, github_branch, pasted_code, created_at, last_accessed 
		 FROM reviews.sessions WHERE user_id = $1 
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("db: failed to query reviews: %w", err)
	}
	defer func() {
		_ = rows.Close() //nolint:errcheck // explicitly ignore close error
	}()

	var reviews []*Review
	for rows.Next() {
		var review Review
		scanErr := rows.Scan(&review.ID, &review.UserID, &review.Title, &review.CodeSource,
			&review.GithubRepo, &review.GithubBranch, &review.PastedCode,
			&review.CreatedAt, &review.LastAccessed)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("db: failed to scan review: %w", scanErr)
		}
		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("db: iteration error: %w", err)
	}

	return reviews, total, nil
}

// DeleteByID removes a review session from the database
func (r *ReviewRepository) DeleteByID(ctx context.Context, id int64) error {
	result, err := r.DB.ExecContext(ctx, `DELETE FROM reviews.sessions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("db: failed to delete review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("db: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("db: review not found")
	}

	return nil
}

// UpdateLastAccessed updates the last_accessed timestamp for a session
func (r *ReviewRepository) UpdateLastAccessed(ctx context.Context, id int64) error {
	result, err := r.DB.ExecContext(ctx,
		`UPDATE reviews.sessions SET last_accessed = NOW() WHERE id = $1`,
		id)
	if err != nil {
		return fmt.Errorf("db: failed to update last_accessed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("db: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("db: review not found")
	}

	return nil
}

// FindExpiredSessions returns sessions older than the specified number of days
func (r *ReviewRepository) FindExpiredSessions(ctx context.Context, daysOld int) ([]*Review, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, user_id, title, code_source, github_repo, github_branch, pasted_code, created_at, last_accessed 
		 FROM reviews.sessions 
		 WHERE created_at < NOW() - INTERVAL '1 day' * $1
		 ORDER BY created_at ASC`,
		daysOld)
	if err != nil {
		return nil, fmt.Errorf("db: failed to query expired sessions: %w", err)
	}
	defer func() {
		_ = rows.Close() //nolint:errcheck // explicitly ignore close error
	}()

	var sessions []*Review
	for rows.Next() {
		var session Review
		scanErr := rows.Scan(&session.ID, &session.UserID, &session.Title, &session.CodeSource,
			&session.GithubRepo, &session.GithubBranch, &session.PastedCode,
			&session.CreatedAt, &session.LastAccessed)
		if scanErr != nil {
			return nil, fmt.Errorf("db: failed to scan session: %w", scanErr)
		}
		sessions = append(sessions, &session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("db: iteration error: %w", err)
	}

	return sessions, nil
}
