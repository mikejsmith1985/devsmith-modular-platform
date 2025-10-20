package services

import (
	"context"
	"os"
	"testing"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) CreateOrUpdate(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserRepo) FindByGitHubID(ctx context.Context, githubID int64) (*models.User, error) {
	args := m.Called(ctx, githubID)
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserRepo) FindByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

type MockGitHubClient struct{ mock.Mock }

func (m *MockGitHubClient) ExchangeCodeForToken(ctx context.Context, code string) (string, error) {
	args := m.Called(ctx, code)
	return args.String(0), args.Error(1)
}
func (m *MockGitHubClient) GetUserProfile(ctx context.Context, accessToken string) (*models.GitHubProfile, error) {
	args := m.Called(ctx, accessToken)
	return args.Get(0).(*models.GitHubProfile), args.Error(1)
}

func TestAuthService_AuthenticateWithGitHub(t *testing.T) {
	userRepo := new(MockUserRepo)
	githubClient := new(MockGitHubClient)
	logger := zerolog.New(os.Stdout)
	jwtSecret := "testsecret"
	service := NewAuthService(userRepo, githubClient, jwtSecret, &logger)

	profile := &models.GitHubProfile{
		ID:        123456,
		Username:  "testuser",
		Email:     "test@example.com",
		AvatarURL: "https://avatars.githubusercontent.com/u/123456",
	}
	user := &models.User{
		GitHubID:          profile.ID,
		Username:          profile.Username,
		Email:             profile.Email,
		AvatarURL:         profile.AvatarURL,
		GitHubAccessToken: "fake-token",
	}
	githubClient.On("ExchangeCodeForToken", mock.Anything, "code123").Return("fake-token", nil)
	githubClient.On("GetUserProfile", mock.Anything, "fake-token").Return(profile, nil)
	userRepo.On("CreateOrUpdate", mock.Anything, user).Return(nil)

	gotUser, token, err := service.AuthenticateWithGitHub(context.Background(), "code123")
	assert.NoError(t, err)
	assert.Equal(t, user.Username, gotUser.Username)
	assert.NotEmpty(t, token)
}
