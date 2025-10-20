package interfaces

import (
    "context"
    "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
)

// AuthService defines authentication operations
// Why interface: Allows mocking in handler tests, enables future auth providers
type AuthService interface {
    // AuthenticateWithGitHub completes OAuth flow and creates user session
    // Returns user and JWT token on success
    AuthenticateWithGitHub(ctx context.Context, code string) (*models.User, string, error)

    // ValidateSession checks if JWT token is valid and returns user
    ValidateSession(ctx context.Context, token string) (*models.User, error)

    // RevokeSession invalidates the given JWT token
    RevokeSession(ctx context.Context, token string) error
}

// UserRepository defines database operations for users
// Why interface: Enables testing services without real database
type UserRepository interface {
    // CreateOrUpdate inserts new user or updates existing (by GitHub ID)
    CreateOrUpdate(ctx context.Context, user *models.User) error

    // FindByGitHubID retrieves user by their GitHub ID
    FindByGitHubID(ctx context.Context, githubID int64) (*models.User, error)

    // FindByID retrieves user by internal database ID
    FindByID(ctx context.Context, id int) (*models.User, error)
}

// GitHubClient defines operations for GitHub API integration
// Why interface: Enables testing without hitting real GitHub API
type GitHubClient interface {
    // ExchangeCodeForToken exchanges OAuth code for access token
    ExchangeCodeForToken(ctx context.Context, code string) (string, error)

    // GetUserProfile fetches authenticated user's GitHub profile
    GetUserProfile(ctx context.Context, accessToken string) (*models.GitHubProfile, error)
}
