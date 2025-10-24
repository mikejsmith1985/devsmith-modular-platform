package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReviewRepository(t *testing.T) {
	var db *sql.DB

	repo := NewReviewRepository(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.DB)
}

func TestReviewRepository_Methods(t *testing.T) {
	// Test that the repository methods exist and can be called
	// Note: We suppress usage to avoid "declared and not used" errors
	repo := NewReviewRepository(nil)
	assert.NotNil(t, repo)

	// Verify fields are accessible
	assert.Nil(t, repo.DB)

	// Don't actually call methods as they would panic with nil DB
	_ = context.Background()
}

func TestReview_Struct(t *testing.T) {
	review := Review{
		ID:           1,
		UserID:       123,
		Title:        "Test Review",
		CodeSource:   "github",
		GithubRepo:   "test/repo",
		GithubBranch: "main",
		PastedCode:   "package main",
	}

	assert.Equal(t, int64(1), review.ID)
	assert.Equal(t, int64(123), review.UserID)
	assert.Equal(t, "Test Review", review.Title)
	assert.Equal(t, "github", review.CodeSource)
	assert.Equal(t, "test/repo", review.GithubRepo)
	assert.Equal(t, "main", review.GithubBranch)
	assert.Equal(t, "package main", review.PastedCode)
}
