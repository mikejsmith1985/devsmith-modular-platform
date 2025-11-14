// Package logs_db provides database access for project management (cross-repo logging).
package logs_db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// ProjectRepository handles CRUD operations for projects.
type ProjectRepository struct {
	db *sql.DB
}

// NewProjectRepository creates a new ProjectRepository with the given database connection.
func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create inserts a new project and returns the created project with ID.
func (r *ProjectRepository) Create(ctx context.Context, project *logs_models.Project) (*logs_models.Project, error) {
	query := `
		INSERT INTO logs.projects (user_id, name, slug, description, repository_url, api_token, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		project.UserID,
		project.Name,
		project.Slug,
		project.Description,
		project.RepositoryURL,
		project.APIKeyHash,
		project.IsActive,
	).Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("db: failed to create project: %w", err)
	}

	return project, nil
}

// GetByID retrieves a project by its ID and user ID.
func (r *ProjectRepository) GetByID(ctx context.Context, id int, userID int) (*logs_models.Project, error) {
	query := `
		SELECT id, user_id, name, slug, description, repository_url, api_token, 
		       created_at, updated_at, is_active
		FROM logs.projects
		WHERE id = $1 AND user_id = $2
	`

	var project logs_models.Project
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&project.ID,
		&project.UserID,
		&project.Name,
		&project.Slug,
		&project.Description,
		&project.RepositoryURL,
		&project.APIKeyHash,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("db: failed to get project by id: %w", err)
	}

	return &project, nil
}

// GetByIDGlobal retrieves a project by ID without userID constraint.
// Used for service operations that don't have user context.
func (r *ProjectRepository) GetByIDGlobal(ctx context.Context, id int) (*logs_models.Project, error) {
	query := `
		SELECT id, user_id, name, slug, description, repository_url, api_token, 
		       created_at, updated_at, is_active
		FROM logs.projects
		WHERE id = $1
	`

	var project logs_models.Project
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.UserID,
		&project.Name,
		&project.Slug,
		&project.Description,
		&project.RepositoryURL,
		&project.APIKeyHash,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("db: failed to get project by id: %w", err)
	}

	return &project, nil
}

// GetBySlug retrieves a project by its slug and user ID.
func (r *ProjectRepository) GetBySlug(ctx context.Context, slug string, userID int) (*logs_models.Project, error) {
	query := `
		SELECT id, user_id, name, slug, description, repository_url, api_token, 
		       created_at, updated_at, is_active
		FROM logs.projects
		WHERE slug = $1 AND user_id = $2
	`

	var project logs_models.Project
	err := r.db.QueryRowContext(ctx, query, userID, slug).Scan(
		&project.ID,
		&project.UserID,
		&project.Name,
		&project.Slug,
		&project.Description,
		&project.RepositoryURL,
		&project.APIKeyHash,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("db: failed to get project by slug: %w", err)
	}

	return &project, nil
}

// GetBySlugGlobal retrieves a project by slug without userID constraint.
// This is used by the batch ingestion endpoint for API key validation.
// Only returns active projects.
func (r *ProjectRepository) GetBySlugGlobal(ctx context.Context, slug string) (*logs_models.Project, error) {
	query := `
		SELECT id, user_id, name, slug, description, repository_url, api_token, 
		       created_at, updated_at, is_active
		FROM logs.projects
		WHERE slug = $1 AND is_active = true
	`

	var project logs_models.Project
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&project.ID,
		&project.UserID,
		&project.Name,
		&project.Slug,
		&project.Description,
		&project.RepositoryURL,
		&project.APIKeyHash,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("db: failed to get project by slug: %w", err)
	}

	return &project, nil
}

// FindByAPIToken retrieves a project by its plain API token (for authentication).
// Uses indexed lookup for O(1) performance. Only returns active projects.
func (r *ProjectRepository) FindByAPIToken(ctx context.Context, token string) (*logs_models.Project, error) {
	query := `
		SELECT id, user_id, name, slug, description, repository_url, api_token, 
		       created_at, updated_at, is_active
		FROM logs.projects
		WHERE api_token = $1 AND is_active = true
	`

	var project logs_models.Project
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&project.ID,
		&project.UserID,
		&project.Name,
		&project.Slug,
		&project.Description,
		&project.RepositoryURL,
		&project.APIKeyHash, // Model field still named APIKeyHash, but now stores plain token
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("db: project not found for api token")
		}
		return nil, fmt.Errorf("db: failed to find project by api token: %w", err)
	}

	return &project, nil
}

// ListByUserID retrieves all projects for a specific user.
func (r *ProjectRepository) ListByUserID(ctx context.Context, userID int) ([]logs_models.Project, error) {
	query := `
		SELECT id, user_id, name, slug, description, repository_url, api_token, 
		       created_at, updated_at, is_active
		FROM logs.projects
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("db: failed to list projects: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Error closing rows: %v\n", closeErr)
		}
	}()

	var projects []logs_models.Project
	for rows.Next() {
		var project logs_models.Project
		err := rows.Scan(
			&project.ID,
			&project.UserID,
			&project.Name,
			&project.Slug,
			&project.Description,
			&project.RepositoryURL,
			&project.APIKeyHash,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("db: failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db: rows iteration error: %w", err)
	}

	return projects, nil
}

// Update updates an existing project.
func (r *ProjectRepository) Update(ctx context.Context, project *logs_models.Project) error {
	query := `
		UPDATE logs.projects
		SET name = $1, description = $2, repository_url = $3, is_active = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		project.Name,
		project.Description,
		project.RepositoryURL,
		project.IsActive,
		time.Now(),
		project.ID,
	)

	if err != nil {
		return fmt.Errorf("db: failed to update project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("db: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("db: project not found")
	}

	return nil
}

// UpdateAPIToken updates the API token for a project (for token regeneration).
func (r *ProjectRepository) UpdateAPIToken(ctx context.Context, projectID int, newAPIToken string) error {
	query := `
		UPDATE logs.projects
		SET api_token = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, newAPIToken, time.Now(), projectID)
	if err != nil {
		return fmt.Errorf("db: failed to update api token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("db: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("db: project not found")
	}

	return nil
}

// Delete soft-deletes a project by setting is_active to false.
func (r *ProjectRepository) Delete(ctx context.Context, id int) error {
	query := `
		UPDATE logs.projects
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("db: failed to delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("db: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("db: project not found")
	}

	return nil
}
