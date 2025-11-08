package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// PromptTemplateRepositoryInterface defines the interface for prompt template operations
type PromptTemplateRepositoryInterface interface {
	FindByUserAndMode(ctx context.Context, userID int, mode, userLevel, outputMode string) (*review_models.PromptTemplate, error)
	FindDefaultByMode(ctx context.Context, mode, userLevel, outputMode string) (*review_models.PromptTemplate, error)
	Upsert(ctx context.Context, template *review_models.PromptTemplate) (*review_models.PromptTemplate, error)
	DeleteUserCustom(ctx context.Context, userID int, mode, userLevel, outputMode string) error
	SaveExecution(ctx context.Context, execution *review_models.PromptExecution) error
	GetExecutionHistory(ctx context.Context, userID int, limit int) ([]*review_models.PromptExecution, error)
}

// SQL query constants for maintainability
const (
	selectPromptFields = `id, user_id, mode, user_level, output_mode, prompt_text, 
	                      variables, is_default, version, created_at, updated_at`

	queryFindByUserAndMode = `
		SELECT ` + selectPromptFields + `
		FROM review.prompt_templates
		WHERE user_id = $1 
		  AND mode = $2 
		  AND user_level = $3 
		  AND output_mode = $4`

	queryFindDefaultByMode = `
		SELECT ` + selectPromptFields + `
		FROM review.prompt_templates
		WHERE user_id IS NULL 
		  AND is_default = true
		  AND mode = $1 
		  AND user_level = $2 
		  AND output_mode = $3`

	queryUpsertPrompt = `
		INSERT INTO review.prompt_templates 
		(id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, mode, user_level, output_mode)
		DO UPDATE SET
			prompt_text = EXCLUDED.prompt_text,
			variables = EXCLUDED.variables,
			version = EXCLUDED.version,
			updated_at = NOW()
		RETURNING ` + selectPromptFields

	queryDeleteUserCustom = `
		DELETE FROM review.prompt_templates
		WHERE user_id = $1 
		  AND mode = $2 
		  AND user_level = $3 
		  AND output_mode = $4
		  AND is_default = false`

	querySaveExecution = `
		INSERT INTO review.prompt_executions
		(template_id, user_id, rendered_prompt, response, model_used, 
		 latency_ms, tokens_used, user_rating)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	queryGetExecutionHistory = `
		SELECT id, template_id, user_id, rendered_prompt, response, 
		       model_used, latency_ms, tokens_used, user_rating, created_at
		FROM review.prompt_executions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`
)

// PromptTemplateRepository handles database operations for prompt templates
type PromptTemplateRepository struct {
	db *sql.DB
}

// NewPromptTemplateRepository creates a new prompt template repository
func NewPromptTemplateRepository(db *sql.DB) *PromptTemplateRepository {
	return &PromptTemplateRepository{db: db}
}

// scanPromptTemplate is a helper to reduce code duplication in scanning prompt templates
func scanPromptTemplate(scanner interface {
	Scan(dest ...interface{}) error
}) (*review_models.PromptTemplate, error) {
	var template review_models.PromptTemplate
	var variablesJSON []byte

	err := scanner.Scan(
		&template.ID,
		&template.UserID,
		&template.Mode,
		&template.UserLevel,
		&template.OutputMode,
		&template.PromptText,
		&variablesJSON,
		&template.IsDefault,
		&template.Version,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse variables JSON
	if err := json.Unmarshal(variablesJSON, &template.Variables); err != nil {
		return nil, fmt.Errorf("failed to parse variables: %w", err)
	}

	return &template, nil
}

// FindByUserAndMode finds a user's custom prompt for a specific mode combination
// Returns nil if no user custom exists
func (r *PromptTemplateRepository) FindByUserAndMode(
	ctx context.Context,
	userID int,
	mode, userLevel, outputMode string,
) (*review_models.PromptTemplate, error) {
	row := r.db.QueryRowContext(ctx, queryFindByUserAndMode, userID, mode, userLevel, outputMode)

	template, err := scanPromptTemplate(row)
	if err == sql.ErrNoRows {
		return nil, nil // No user custom exists
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user prompt: %w", err)
	}

	return template, nil
}

// FindDefaultByMode finds the system default prompt for a mode combination
func (r *PromptTemplateRepository) FindDefaultByMode(
	ctx context.Context,
	mode, userLevel, outputMode string,
) (*review_models.PromptTemplate, error) {
	row := r.db.QueryRowContext(ctx, queryFindDefaultByMode, mode, userLevel, outputMode)

	template, err := scanPromptTemplate(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no default prompt found for mode=%s, level=%s, output=%s", mode, userLevel, outputMode)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find default prompt: %w", err)
	}

	return template, nil
}

// Upsert creates or updates a prompt template
func (r *PromptTemplateRepository) Upsert(
	ctx context.Context,
	template *review_models.PromptTemplate,
) (*review_models.PromptTemplate, error) {
	// Convert variables to JSON
	variablesJSON, err := json.Marshal(template.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal variables: %w", err)
	}

	row := r.db.QueryRowContext(ctx, queryUpsertPrompt,
		template.ID,
		template.UserID,
		template.Mode,
		template.UserLevel,
		template.OutputMode,
		template.PromptText,
		variablesJSON,
		template.IsDefault,
		template.Version,
	)

	result, err := scanPromptTemplate(row)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert prompt template: %w", err)
	}

	return result, nil
}

// DeleteUserCustom deletes a user's custom prompt for a specific mode
func (r *PromptTemplateRepository) DeleteUserCustom(
	ctx context.Context,
	userID int,
	mode, userLevel, outputMode string,
) error {
	result, err := r.db.ExecContext(ctx, queryDeleteUserCustom, userID, mode, userLevel, outputMode)
	if err != nil {
		return fmt.Errorf("failed to delete user custom prompt: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user custom prompt found to delete")
	}

	return nil
}

// SaveExecution logs a prompt execution
func (r *PromptTemplateRepository) SaveExecution(
	ctx context.Context,
	execution *review_models.PromptExecution,
) error {
	err := r.db.QueryRowContext(ctx, querySaveExecution,
		execution.TemplateID,
		execution.UserID,
		execution.RenderedPrompt,
		execution.Response,
		execution.ModelUsed,
		execution.LatencyMs,
		execution.TokensUsed,
		execution.UserRating,
	).Scan(&execution.ID, &execution.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to save execution: %w", err)
	}

	return nil
}

// GetExecutionHistory retrieves a user's recent prompt executions
func (r *PromptTemplateRepository) GetExecutionHistory(
	ctx context.Context,
	userID int,
	limit int,
) ([]*review_models.PromptExecution, error) {
	rows, err := r.db.QueryContext(ctx, queryGetExecutionHistory, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query execution history: %w", err)
	}
	defer rows.Close()

	var executions []*review_models.PromptExecution
	for rows.Next() {
		var exec review_models.PromptExecution
		err := rows.Scan(
			&exec.ID,
			&exec.TemplateID,
			&exec.UserID,
			&exec.RenderedPrompt,
			&exec.Response,
			&exec.ModelUsed,
			&exec.LatencyMs,
			&exec.TokensUsed,
			&exec.UserRating,
			&exec.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution: %w", err)
		}
		executions = append(executions, &exec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating execution rows: %w", err)
	}

	return executions, nil
}
