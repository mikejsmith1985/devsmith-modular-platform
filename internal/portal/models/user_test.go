package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_BasicStruct(t *testing.T) {
	now := time.Now()
	user := User{
		CreatedAt:         now,
		UpdatedAt:         now,
		Username:          "testuser",
		Email:             "test@example.com",
		AvatarURL:         "https://example.com/avatar.jpg",
		GitHubAccessToken: "token123",
		ID:                1,
		GitHubID:          12345,
	}

	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "https://example.com/avatar.jpg", user.AvatarURL)
	assert.Equal(t, "token123", user.GitHubAccessToken)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, int64(12345), user.GitHubID)
}

func TestUser_ZeroValues(t *testing.T) {
	user := User{}

	assert.Equal(t, time.Time{}, user.CreatedAt)
	assert.Equal(t, time.Time{}, user.UpdatedAt)
	assert.Equal(t, "", user.Username)
	assert.Equal(t, "", user.Email)
	assert.Equal(t, "", user.AvatarURL)
	assert.Equal(t, "", user.GitHubAccessToken)
	assert.Equal(t, 0, user.ID)
	assert.Equal(t, int64(0), user.GitHubID)
}

func TestUser_FieldIndependence(t *testing.T) {
	user := User{}

	user.Username = "john"
	assert.Equal(t, "john", user.Username)
	assert.Equal(t, "", user.Email)

	user.Email = "john@example.com"
	assert.Equal(t, "john", user.Username)
	assert.Equal(t, "john@example.com", user.Email)

	user.ID = 42
	assert.Equal(t, "john", user.Username)
	assert.Equal(t, "john@example.com", user.Email)
	assert.Equal(t, 42, user.ID)
}

func TestUser_MultipleInstances(t *testing.T) {
	user1 := User{Username: "user1", ID: 1}
	user2 := User{Username: "user2", ID: 2}
	user3 := User{Username: "user3", ID: 3}

	assert.NotEqual(t, user1.ID, user2.ID)
	assert.NotEqual(t, user2.ID, user3.ID)
	assert.NotEqual(t, user1.Username, user2.Username)
}

func TestUser_Timestamps(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)

	user := User{
		CreatedAt: now,
		UpdatedAt: later,
	}

	assert.True(t, user.UpdatedAt.After(user.CreatedAt))
	assert.Equal(t, time.Hour, user.UpdatedAt.Sub(user.CreatedAt))
}

func TestGitHubProfile_BasicStruct(t *testing.T) {
	profile := GitHubProfile{
		Username:  "octocat",
		Email:     "octocat@github.com",
		AvatarURL: "https://avatars.githubusercontent.com/u/1?v=4",
		ID:        1,
	}

	assert.Equal(t, "octocat", profile.Username)
	assert.Equal(t, "octocat@github.com", profile.Email)
	assert.Equal(t, "https://avatars.githubusercontent.com/u/1?v=4", profile.AvatarURL)
	assert.Equal(t, int64(1), profile.ID)
}

func TestGitHubProfile_ZeroValues(t *testing.T) {
	profile := GitHubProfile{}

	assert.Equal(t, "", profile.Username)
	assert.Equal(t, "", profile.Email)
	assert.Equal(t, "", profile.AvatarURL)
	assert.Equal(t, int64(0), profile.ID)
}

func TestGitHubProfile_FieldIndependence(t *testing.T) {
	profile := GitHubProfile{}

	profile.Username = "alice"
	assert.Equal(t, "alice", profile.Username)
	assert.Equal(t, "", profile.Email)

	profile.ID = 999
	assert.Equal(t, "alice", profile.Username)
	assert.Equal(t, int64(999), profile.ID)
}

func TestGitHubProfile_MultipleInstances(t *testing.T) {
	p1 := GitHubProfile{Username: "user1", ID: 1}
	p2 := GitHubProfile{Username: "user2", ID: 2}

	assert.NotEqual(t, p1.ID, p2.ID)
	assert.NotEqual(t, p1.Username, p2.Username)
}

func TestUser_EmailValidation(t *testing.T) {
	user := User{Email: "test@example.com"}
	assert.NotEmpty(t, user.Email)
	assert.Contains(t, user.Email, "@")
}

func TestGitHubProfile_AvatarURLFormat(t *testing.T) {
	profile := GitHubProfile{AvatarURL: "https://avatars.githubusercontent.com/u/123?v=4"}
	assert.NotEmpty(t, profile.AvatarURL)
	assert.True(t, len(profile.AvatarURL) > 0)
}
