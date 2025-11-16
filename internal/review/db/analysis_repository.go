package review_db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// AnalysisRepository implements services.AnalysisRepositoryInterface
// Stores and retrieves analysis results for review sessions
// Used by ScanService, SkimService, etc.
type AnalysisRepository struct {
	DB *sql.DB
}

// NewAnalysisRepository creates a new AnalysisRepository with the given DB connection.
func NewAnalysisRepository(db *sql.DB) *AnalysisRepository {
	return &AnalysisRepository{DB: db}
}

// FindByReviewAndMode retrieves an analysis result by review ID and mode.
func (r *AnalysisRepository) FindByReviewAndMode(ctx context.Context, reviewID int64, mode string) (*review_models.AnalysisResult, error) {
	row := r.DB.QueryRowContext(ctx, `SELECT review_id, mode, prompt, summary, metadata, model_used, raw_output FROM reviews.analysis_results WHERE review_id = $1 AND mode = $2`, reviewID, mode)
	var result review_models.AnalysisResult
	if err := row.Scan(&result.ReviewID, &result.Mode, &result.Prompt, &result.Summary, &result.Metadata, &result.ModelUsed, &result.RawOutput); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("not found")
		}
		return nil, fmt.Errorf("db: failed to get analysis result: %w", err)
	}
	return &result, nil
}

// Create inserts a new analysis result into the database.
func (r *AnalysisRepository) Create(ctx context.Context, result *review_models.AnalysisResult) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO reviews.analysis_results (review_id, mode, prompt, summary, metadata, model_used, raw_output) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		result.ReviewID, result.Mode, result.Prompt, result.Summary, result.Metadata, result.ModelUsed, result.RawOutput)
	if err != nil {
		return fmt.Errorf("db: failed to create analysis result: %w", err)
	}
	return nil
}

// DeleteOlderThan removes analysis results older than the provided cutoff time.
func (r *AnalysisRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) error {
	// NOTE: The table is expected to have a created_at column.
	_, err := r.DB.ExecContext(ctx, `DELETE FROM reviews.analysis_results WHERE created_at < $1`, cutoff)
	if err != nil {
		return fmt.Errorf("db: failed to delete old analysis results: %w", err)
	}
	return nil
}
