// Package db provides database access and repository implementations for the portal service.
package db

import (
	"context"
	"database/sql"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
)

// UserRepositoryImpl implements interfaces.UserRepository and handles CRUD operations for portal.users.
// It provides methods to create, update, and retrieve users from the database.
// UserRepositoryImpl implements interfaces.UserRepository and handles CRUD operations for portal.users.
// It provides methods to create, update, and retrieve users from the database.
	db *sql.DB
}

// NewUserRepository creates a new UserRepositoryImpl with the given database connection.
func NewUserRepository(db *sql.DB) authifaces.UserRepository {
	return &UserRepositoryImpl{db: db}
}

// CreateOrUpdate inserts or updates a user in the portal.users table by GitHub ID.
func (r *UserRepositoryImpl) CreateOrUpdate(ctx context.Context, user *models.User) error {
	// Upsert user by GitHub ID
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO portal.users (github_id, username, email, avatar_url, github_access_token, created_at, updated_at)
         VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
         ON CONFLICT (github_id) DO UPDATE SET
           username = EXCLUDED.username,
           email = EXCLUDED.email,
           avatar_url = EXCLUDED.avatar_url,
           github_access_token = EXCLUDED.github_access_token,
           updated_at = NOW();`,
		user.GitHubID, user.Username, user.Email, user.AvatarURL, user.GitHubAccessToken)
	return err
}

// FindByGitHubID retrieves a user by GitHub ID from the portal.users table.
func (r *UserRepositoryImpl) FindByGitHubID(ctx context.Context, githubID int64) (*models.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, github_id, username, email, avatar_url, github_access_token, created_at, updated_at
         FROM portal.users WHERE github_id = $1`, githubID)
	var user models.User
	err := row.Scan(&user.ID, &user.GitHubID, &user.Username, &user.Email, &user.AvatarURL, &user.GitHubAccessToken, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID retrieves a user by ID from the portal.users table.
func (r *UserRepositoryImpl) FindByID(ctx context.Context, id int) (*models.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, github_id, username, email, avatar_url, github_access_token, created_at, updated_at
         FROM portal.users WHERE id = $1`, id)
	var user models.User
	err := row.Scan(&user.ID, &user.GitHubID, &user.Username, &user.Email, &user.AvatarURL, &user.GitHubAccessToken, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
