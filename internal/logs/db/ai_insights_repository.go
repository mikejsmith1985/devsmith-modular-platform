package logs_db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// AIInsightsRepository handles database operations for AI insights
type AIInsightsRepository struct {
	db *sql.DB
}

// NewAIInsightsRepository creates a new AI insights repository
func NewAIInsightsRepository(db *sql.DB) *AIInsightsRepository {
	return &AIInsightsRepository{db: db}
}

// GetByLogID retrieves AI insights for a specific log entry
func (r *AIInsightsRepository) GetByLogID(ctx context.Context, logID int64) (*logs_models.AIInsight, error) {
	query := `
		SELECT id, log_id, analysis, root_cause, suggestions, model_used, generated_at
		FROM logs.ai_insights
		WHERE log_id = $1
	`

	var insight logs_models.AIInsight
	var suggestionsJSON []byte

	err := r.db.QueryRowContext(ctx, query, logID).Scan(
		&insight.ID,
		&insight.LogID,
		&insight.Analysis,
		&insight.RootCause,
		&suggestionsJSON,
		&insight.ModelUsed,
		&insight.GeneratedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No insights found (not an error)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query insights: %w", err)
	}

	// Parse suggestions JSON
	if len(suggestionsJSON) > 0 {
		if err := json.Unmarshal(suggestionsJSON, &insight.Suggestions); err != nil {
			return nil, fmt.Errorf("failed to parse suggestions: %w", err)
		}
	}

	return &insight, nil
}

// Upsert inserts or updates AI insights (replaces existing)
func (r *AIInsightsRepository) Upsert(ctx context.Context, insight *logs_models.AIInsight) (*logs_models.AIInsight, error) {
	// Convert suggestions to JSON
	suggestionsJSON, err := json.Marshal(insight.Suggestions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal suggestions: %w", err)
	}

	query := `
		INSERT INTO logs.ai_insights (log_id, analysis, root_cause, suggestions, model_used, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (log_id) 
		DO UPDATE SET 
			analysis = EXCLUDED.analysis,
			root_cause = EXCLUDED.root_cause,
			suggestions = EXCLUDED.suggestions,
			model_used = EXCLUDED.model_used,
			generated_at = EXCLUDED.generated_at,
			created_at = NOW()
		RETURNING id, log_id, analysis, root_cause, suggestions, model_used, generated_at
	`

	var result logs_models.AIInsight
	var resultSuggestionsJSON []byte

	err = r.db.QueryRowContext(
		ctx,
		query,
		insight.LogID,
		insight.Analysis,
		insight.RootCause,
		suggestionsJSON,
		insight.ModelUsed,
		insight.GeneratedAt,
	).Scan(
		&result.ID,
		&result.LogID,
		&result.Analysis,
		&result.RootCause,
		&resultSuggestionsJSON,
		&result.ModelUsed,
		&result.GeneratedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to upsert insight: %w", err)
	}

	// Parse suggestions JSON
	if len(resultSuggestionsJSON) > 0 {
		if err := json.Unmarshal(resultSuggestionsJSON, &result.Suggestions); err != nil {
			return nil, fmt.Errorf("failed to parse result suggestions: %w", err)
		}
	}

	return &result, nil
}
