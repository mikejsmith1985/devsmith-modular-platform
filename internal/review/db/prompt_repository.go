package review_db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// PromptTemplateRepository handles database operations for prompt templates
type PromptTemplateRepository struct {
	DB *sql.DB
}

// NewPromptTemplateRepository creates a new PromptTemplateRepository
func NewPromptTemplateRepository(db *sql.DB) *PromptTemplateRepository {
	return &PromptTemplateRepository{DB: db}
}

// FindByUserAndMode finds a custom prompt for a specific user and mode combination
// Returns sql.ErrNoRows if not found
func (r *PromptTemplateRepository) FindByUserAndMode(ctx context.Context, userID int, mode, userLevel, outputMode string) (*review_models.PromptTemplate, error) {
	query := `
		SELECT id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version, created_at, updated_at
		FROM review.prompt_templates
		WHERE user_id = $1 AND mode = $2 AND user_level = $3 AND output_mode = $4
	`

	var pt review_models.PromptTemplate
	var variablesJSON []byte

	err := r.DB.QueryRowContext(ctx, query, userID, mode, userLevel, outputMode).Scan(
		&pt.ID, &pt.UserID, &pt.Mode, &pt.UserLevel, &pt.OutputMode,
		&pt.PromptText, &variablesJSON, &pt.IsDefault, &pt.Version,
		&pt.CreatedAt, &pt.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse variables JSON
	if err := json.Unmarshal(variablesJSON, &pt.Variables); err != nil {
		return nil, fmt.Errorf("failed to parse variables JSON: %w", err)
	}

	return &pt, nil
}

// FindDefaultByMode finds the system default prompt for a mode combination
func (r *PromptTemplateRepository) FindDefaultByMode(ctx context.Context, mode, userLevel, outputMode string) (*review_models.PromptTemplate, error) {
	query := `
		SELECT id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version, created_at, updated_at
		FROM review.prompt_templates
		WHERE user_id IS NULL AND mode = $1 AND user_level = $2 AND output_mode = $3
	`

	var pt review_models.PromptTemplate
	var variablesJSON []byte

	err := r.DB.QueryRowContext(ctx, query, mode, userLevel, outputMode).Scan(
		&pt.ID, &pt.UserID, &pt.Mode, &pt.UserLevel, &pt.OutputMode,
		&pt.PromptText, &variablesJSON, &pt.IsDefault, &pt.Version,
		&pt.CreatedAt, &pt.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse variables JSON
	if err := json.Unmarshal(variablesJSON, &pt.Variables); err != nil {
		return nil, fmt.Errorf("failed to parse variables JSON: %w", err)
	}

	return &pt, nil
}

// CreateCustom creates a new custom prompt for a user (deprecated - use Upsert)
func (r *PromptTemplateRepository) CreateCustom(ctx context.Context, pt *review_models.PromptTemplate) error {
	return fmt.Errorf("CreateCustom is deprecated, use Upsert instead")
}

// UpdateCustom updates an existing custom prompt (deprecated - use Upsert)
func (r *PromptTemplateRepository) UpdateCustom(ctx context.Context, pt *review_models.PromptTemplate) error {
	return fmt.Errorf("UpdateCustom is deprecated, use Upsert instead")
}

// Upsert creates or updates a custom prompt (implements the interface)
func (r *PromptTemplateRepository) Upsert(ctx context.Context, template *review_models.PromptTemplate) (*review_models.PromptTemplate, error) {
	// Serialize variables to JSON
	variablesJSON, err := json.Marshal(template.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal variables: %w", err)
	}

	query := `
		INSERT INTO review.prompt_templates 
		(id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, mode, user_level, output_mode)
		DO UPDATE SET
			prompt_text = EXCLUDED.prompt_text,
			variables = EXCLUDED.variables,
			version = review.prompt_templates.version + 1,
			updated_at = NOW()
		RETURNING id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version, created_at, updated_at
	`

	var result review_models.PromptTemplate
	var variablesJSONResult []byte

	err = r.DB.QueryRowContext(ctx, query,
		template.ID, template.UserID, template.Mode, template.UserLevel, template.OutputMode,
		template.PromptText, variablesJSON, template.IsDefault, template.Version,
	).Scan(
		&result.ID, &result.UserID, &result.Mode, &result.UserLevel, &result.OutputMode,
		&result.PromptText, &variablesJSONResult, &result.IsDefault, &result.Version,
		&result.CreatedAt, &result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to upsert prompt: %w", err)
	}

	// Parse variables JSON from result
	if err := json.Unmarshal(variablesJSONResult, &result.Variables); err != nil {
		return nil, fmt.Errorf("failed to parse variables JSON: %w", err)
	}

	return &result, nil
}

// DeleteUserCustom deletes a user's custom prompt (implements the interface)
func (r *PromptTemplateRepository) DeleteUserCustom(ctx context.Context, userID int, mode, userLevel, outputMode string) error {
	query := `
		DELETE FROM review.prompt_templates
		WHERE user_id = $1 AND mode = $2 AND user_level = $3 AND output_mode = $4 AND is_default = false
	`

	result, err := r.DB.ExecContext(ctx, query, userID, mode, userLevel, outputMode)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// SaveExecution logs a prompt execution for analytics (implements the interface)
func (r *PromptTemplateRepository) SaveExecution(ctx context.Context, exec *review_models.PromptExecution) error {
	query := `
		INSERT INTO review.prompt_executions 
		(template_id, user_id, rendered_prompt, response, model_used, latency_ms, tokens_used, user_rating)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	return r.DB.QueryRowContext(ctx, query,
		exec.TemplateID, exec.UserID, exec.RenderedPrompt, exec.Response,
		exec.ModelUsed, exec.LatencyMs, exec.TokensUsed, exec.UserRating,
	).Scan(&exec.ID, &exec.CreatedAt)
}

// UpdateExecutionRating updates the user rating for an execution (implements the interface)
func (r *PromptTemplateRepository) UpdateExecutionRating(ctx context.Context, executionID int64, userID int, rating int) error {
	query := `
		UPDATE review.prompt_executions
		SET user_rating = $1
		WHERE id = $2 AND user_id = $3
	`

	result, err := r.DB.ExecContext(ctx, query, rating, executionID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetExecutionHistory retrieves recent prompt executions for a user (implements the interface)
func (r *PromptTemplateRepository) GetExecutionHistory(ctx context.Context, userID int, limit int) ([]*review_models.PromptExecution, error) {
	query := `
		SELECT id, template_id, user_id, rendered_prompt, response, model_used, latency_ms, tokens_used, user_rating, created_at
		FROM review.prompt_executions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.DB.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []*review_models.PromptExecution

	for rows.Next() {
		var exec review_models.PromptExecution
		err := rows.Scan(
			&exec.ID, &exec.TemplateID, &exec.UserID, &exec.RenderedPrompt,
			&exec.Response, &exec.ModelUsed, &exec.LatencyMs, &exec.TokensUsed,
			&exec.UserRating, &exec.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		executions = append(executions, &exec)
	}

	return executions, rows.Err()
}
