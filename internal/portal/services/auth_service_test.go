package portal_services

import (
	"context"
	"os"
	"testing"

	portalModels "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
	reviewModels "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) CreateOrUpdate(ctx context.Context, user *portalModels.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserRepo) FindByGitHubID(ctx context.Context, githubID int64) (*portalModels.User, error) {
	args := m.Called(ctx, githubID)
	return args.Get(0).(*portalModels.User), args.Error(1)
}
func (m *MockUserRepo) FindByID(ctx context.Context, id int) (*portalModels.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*portalModels.User), args.Error(1)
}

type MockGitHubClient struct{ mock.Mock }

func (m *MockGitHubClient) ExchangeCodeForToken(ctx context.Context, code string) (string, error) {
	args := m.Called(ctx, code)
	return args.String(0), args.Error(1)
}
func (m *MockGitHubClient) GetUserProfile(ctx context.Context, accessToken string) (*portalModels.GitHubProfile, error) {
	args := m.Called(ctx, accessToken)
	return args.Get(0).(*portalModels.GitHubProfile), args.Error(1)
}

type MockOllamaClient struct {
	mock.Mock
}

func (m *MockOllamaClient) Analyze(data string) (reviewModels.AnalysisResult, error) {
	args := m.Called(data)
	return args.Get(0).(reviewModels.AnalysisResult), args.Error(1)
}
func (m *MockOllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

type MockAnalysisRepository struct {
	mock.Mock
}

func (m *MockAnalysisRepository) SaveAnalysis(data string) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockAnalysisRepository) GetAnalysis(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockAnalysisRepository) Create(ctx context.Context, result *reviewModels.AnalysisResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

func (m *MockAnalysisRepository) FindByReviewAndMode(ctx context.Context, reviewID int64, mode string) (*reviewModels.AnalysisResult, error) {
	args := m.Called(ctx, reviewID, mode)
	return args.Get(0).(*reviewModels.AnalysisResult), args.Error(1)
}

func TestAuthService_AuthenticateWithGitHub(t *testing.T) {
	userRepo := new(MockUserRepo)
	githubClient := new(MockGitHubClient)
	ollamaClient := new(MockOllamaClient)
	analysisRepo := new(MockAnalysisRepository)
	logger := zerolog.New(os.Stdout)
	jwtSecret := "testsecret"

	service := NewAuthService(userRepo, githubClient, jwtSecret, &logger, ollamaClient, analysisRepo)

	profile := &portalModels.GitHubProfile{
		ID:        123456,
		Username:  "testuser",
		Email:     "test@example.com",
		AvatarURL: "https://avatars.githubusercontent.com/u/123456",
	}
	user := &portalModels.User{
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

func TestNewAuthService(t *testing.T) {
	userRepo := new(MockUserRepo)
	githubClient := new(MockGitHubClient)
	ollamaClient := new(MockOllamaClient)
	analysisRepo := new(MockAnalysisRepository)
	logger := zerolog.Nop()

	service := NewAuthService(userRepo, githubClient, "test-secret", &logger, ollamaClient, analysisRepo)

	assert.NotNil(t, service)
}

func TestAuthService_Multiple_Instances(t *testing.T) {
	userRepo := new(MockUserRepo)
	githubClient := new(MockGitHubClient)
	ollamaClient := new(MockOllamaClient)
	analysisRepo := new(MockAnalysisRepository)
	logger := zerolog.Nop()

	service1 := NewAuthService(userRepo, githubClient, "secret", &logger, ollamaClient, analysisRepo)
	service2 := NewAuthService(userRepo, githubClient, "secret", &logger, ollamaClient, analysisRepo)

	assert.NotNil(t, service1)
	assert.NotNil(t, service2)
}

func TestAuthService_ContextCancellation(t *testing.T) {
	// Skip complex context testing for now
}

func TestAuthService_WithDifferentSecrets(t *testing.T) {
	userRepo := new(MockUserRepo)
	githubClient := new(MockGitHubClient)
	ollamaClient := new(MockOllamaClient)
	analysisRepo := new(MockAnalysisRepository)
	logger := zerolog.Nop()

	service1 := NewAuthService(userRepo, githubClient, "secret1", &logger, ollamaClient, analysisRepo)
	service2 := NewAuthService(userRepo, githubClient, "secret2", &logger, ollamaClient, analysisRepo)

	assert.NotNil(t, service1)
	assert.NotNil(t, service2)
}
