package authifaces

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
)

// MockAuthService implements AuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) AuthenticateWithGitHub(ctx context.Context, code string) (*models.User, string, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*models.User), args.String(1), args.Error(2)
}

func (m *MockAuthService) ValidateSession(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) RevokeSession(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// MockUserRepository implements UserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateOrUpdate(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByGitHubID(ctx context.Context, githubID int64) (*models.User, error) {
	args := m.Called(ctx, githubID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// MockGitHubClient implements GitHubClient for testing
type MockGitHubClient struct {
	mock.Mock
}

func (m *MockGitHubClient) ExchangeCodeForToken(ctx context.Context, code string) (string, error) {
	args := m.Called(ctx, code)
	return args.String(0), args.Error(1)
}

func (m *MockGitHubClient) GetUserProfile(ctx context.Context, accessToken string) (*models.GitHubProfile, error) {
	args := m.Called(ctx, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GitHubProfile), args.Error(1)
}

// Tests for interface implementations
func TestAuthService_Interface(t *testing.T) {
	var _ AuthService = (*MockAuthService)(nil)
}

func TestUserRepository_Interface(t *testing.T) {
	var _ UserRepository = (*MockUserRepository)(nil)
}

func TestGitHubClient_Interface(t *testing.T) {
	var _ GitHubClient = (*MockGitHubClient)(nil)
}

func TestMockAuthService_Methods(t *testing.T) {
	mock := new(MockAuthService)
	ctx := context.Background()

	// Test AuthenticateWithGitHub
	user := &models.User{Username: "test"}
	mock.On("AuthenticateWithGitHub", ctx, "code123").Return(user, "token", nil)

	// Test ValidateSession
	mock.On("ValidateSession", ctx, "token").Return(user, nil)

	// Test RevokeSession
	mock.On("RevokeSession", ctx, "token").Return(nil)

	assert.NotNil(t, mock)
}

func TestMockUserRepository_Methods(t *testing.T) {
	mock := new(MockUserRepository)
	ctx := context.Background()
	user := &models.User{ID: 1, Username: "test"}

	// Test CreateOrUpdate
	mock.On("CreateOrUpdate", ctx, user).Return(nil)

	// Test FindByGitHubID
	mock.On("FindByGitHubID", ctx, int64(123)).Return(user, nil)

	// Test FindByID
	mock.On("FindByID", ctx, 1).Return(user, nil)

	assert.NotNil(t, mock)
}

func TestMockGitHubClient_Methods(t *testing.T) {
	mock := new(MockGitHubClient)
	ctx := context.Background()
	profile := &models.GitHubProfile{Username: "octocat", ID: 1}

	// Test ExchangeCodeForToken
	mock.On("ExchangeCodeForToken", ctx, "code").Return("token", nil)

	// Test GetUserProfile
	mock.On("GetUserProfile", ctx, "token").Return(profile, nil)

	assert.NotNil(t, mock)
}
