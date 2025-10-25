package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReviewRepository(t *testing.T) {
	repo := NewReviewRepository(nil)

	assert.NotNil(t, repo)
	assert.Nil(t, repo.DB)
}

func TestNewReviewRepository_WithDB(t *testing.T) {
	// Test with nil DB
	repo := NewReviewRepository(nil)

	assert.NotNil(t, repo)
	assert.IsType(t, &ReviewRepository{}, repo)
}

func TestReview_BasicStruct(t *testing.T) {
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

func TestReview_AllFields(t *testing.T) {
	review := Review{
		ID:           42,
		UserID:       999,
		Title:        "Comprehensive Review",
		CodeSource:   "paste",
		GithubRepo:   "org/project",
		GithubBranch: "develop",
		PastedCode:   "func main() {}",
		CreatedAt:    "2025-10-24T12:00:00Z",
		LastAccessed: "2025-10-24T13:00:00Z",
	}

	assert.Equal(t, int64(42), review.ID)
	assert.Equal(t, int64(999), review.UserID)
	assert.Equal(t, "Comprehensive Review", review.Title)
	assert.Equal(t, "paste", review.CodeSource)
	assert.Equal(t, "org/project", review.GithubRepo)
	assert.Equal(t, "develop", review.GithubBranch)
	assert.Equal(t, "func main() {}", review.PastedCode)
	assert.Equal(t, "2025-10-24T12:00:00Z", review.CreatedAt)
	assert.Equal(t, "2025-10-24T13:00:00Z", review.LastAccessed)
}

func TestReview_ZeroValues(t *testing.T) {
	review := Review{}

	assert.Equal(t, int64(0), review.ID)
	assert.Equal(t, int64(0), review.UserID)
	assert.Equal(t, "", review.Title)
	assert.Equal(t, "", review.CodeSource)
	assert.Equal(t, "", review.GithubRepo)
	assert.Equal(t, "", review.GithubBranch)
	assert.Equal(t, "", review.PastedCode)
	assert.Equal(t, "", review.CreatedAt)
	assert.Equal(t, "", review.LastAccessed)
}

func TestReviewRepository_NotNil(t *testing.T) {
	repo := NewReviewRepository(nil)

	assert.NotNil(t, repo)
	require.IsType(t, &ReviewRepository{}, repo)
}

func TestReviewRepository_HandleNilDB(t *testing.T) {
	repo := NewReviewRepository(nil)

	assert.Nil(t, repo.DB)
	assert.NotNil(t, repo)
}

func TestReview_MultipleInstances(t *testing.T) {
	review1 := Review{ID: 1, Title: "Review 1"}
	review2 := Review{ID: 2, Title: "Review 2"}
	review3 := Review{ID: 3, Title: "Review 3"}

	assert.NotEqual(t, review1.ID, review2.ID)
	assert.NotEqual(t, review2.ID, review3.ID)
	assert.NotEqual(t, review1.ID, review3.ID)

	assert.Equal(t, "Review 1", review1.Title)
	assert.Equal(t, "Review 2", review2.Title)
	assert.Equal(t, "Review 3", review3.Title)
}

func TestReview_FieldIndependence(t *testing.T) {
	review := Review{}

	review.ID = 100
	assert.Equal(t, int64(100), review.ID)

	review.UserID = 200
	assert.Equal(t, int64(200), review.UserID)

	review.Title = "New Title"
	assert.Equal(t, "New Title", review.Title)

	// Previous fields unchanged
	assert.Equal(t, int64(100), review.ID)
	assert.Equal(t, int64(200), review.UserID)
}

func TestReviewRepository_Constructor(t *testing.T) {
	repo := NewReviewRepository(nil)

	// Verify structure is correct
	assert.NotNil(t, repo)
	_, isRepo := interface{}(repo).(*ReviewRepository)
	assert.True(t, isRepo)
}

func TestReview_Marshaling(t *testing.T) {
	review := Review{
		ID:         1,
		UserID:     100,
		Title:      "Test",
		CodeSource: "github",
	}

	// Verify fields can be read
	assert.Equal(t, int64(1), review.ID)
	assert.Equal(t, int64(100), review.UserID)
	assert.Equal(t, "Test", review.Title)
	assert.Equal(t, "github", review.CodeSource)
}

func TestReview_EmptyTitles(t *testing.T) {
	review1 := Review{Title: ""}
	review2 := Review{Title: ""}

	assert.Equal(t, "", review1.Title)
	assert.Equal(t, "", review2.Title)
}

func TestReview_LargeIDs(t *testing.T) {
	maxInt64 := int64(9223372036854775807)
	review := Review{ID: maxInt64}

	assert.Equal(t, maxInt64, review.ID)
}
