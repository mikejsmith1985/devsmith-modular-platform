package db

import (
	"context"
	"database/sql"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/interfaces"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
)

// UserRepositoryImpl implements interfaces.UserRepository
// Handles CRUD operations for portal.users

type UserRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) interfaces.UserRepository {
	return &UserRepositoryImpl{db: db}
}

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
