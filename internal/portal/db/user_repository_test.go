package db

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", "postgres://devsmith:devsmith@localhost:5432/devsmith_test?sslmode=disable")
	if err != nil {
		t.Skipf("skipping: failed to connect to test db: %v", err)
	}

	// Ping to verify connection
	if err := db.Ping(); err != nil {
		t.Skipf("skipping: test database not available: %v", err)
	}

	return db
}

func TestUserRepository_CreateOrUpdateAndFind(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		GitHubID:          123456,
		Username:          "testuser",
		Email:             "test@example.com",
		AvatarURL:         "https://avatars.githubusercontent.com/u/123456",
		GitHubAccessToken: "fake-token",
	}

	err := repo.CreateOrUpdate(ctx, user)
	if err != nil {
		t.Fatalf("CreateOrUpdate failed: %v", err)
	}

	fetched, err := repo.FindByGitHubID(ctx, 123456)
	if err != nil {
		t.Fatalf("FindByGitHubID failed: %v", err)
	}
	if fetched.Username != user.Username {
		t.Errorf("expected username %s, got %s", user.Username, fetched.Username)
	}
	if fetched.GitHubAccessToken != user.GitHubAccessToken {
		t.Errorf("expected token %s, got %s", user.GitHubAccessToken, fetched.GitHubAccessToken)
	}
}
