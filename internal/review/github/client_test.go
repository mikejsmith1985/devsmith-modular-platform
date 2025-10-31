package github

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockClient implements ClientInterface for testing
type MockClient struct {
	fetchCodeFunc       func(ctx context.Context, owner, repo, branch string, token string) (*CodeFetch, error)
	getRepoMetadataFunc func(ctx context.Context, owner, repo string, token string) (*RepoMetadata, error)
	validateURLFunc     func(url string) (string, string, error)
	getRateLimitFunc    func(ctx context.Context, token string) (int, time.Time, error)
}

func (m *MockClient) FetchCode(ctx context.Context, owner, repo, branch, token string) (*CodeFetch, error) {
	if m.fetchCodeFunc != nil {
		return m.fetchCodeFunc(ctx, owner, repo, branch, token)
	}
	return nil, nil
}

func (m *MockClient) GetRepoMetadata(ctx context.Context, owner, repo, token string) (*RepoMetadata, error) {
	if m.getRepoMetadataFunc != nil {
		return m.getRepoMetadataFunc(ctx, owner, repo, token)
	}
	return nil, nil
}

func (m *MockClient) ValidateURL(url string) (owner, repo string, err error) {
	if m.validateURLFunc != nil {
		return m.validateURLFunc(url)
	}
	return "", "", nil
}

func (m *MockClient) GetRateLimit(ctx context.Context, token string) (int, time.Time, error) {
	if m.getRateLimitFunc != nil {
		return m.getRateLimitFunc(ctx, token)
	}
	return 0, time.Time{}, nil
}

func TestValidateURL_Success(t *testing.T) {
	// GIVEN: Valid GitHub URLs
	tests := []struct {
		url           string
		expectedOwner string
		expectedRepo  string
	}{
		{"https://github.com/mikejsmith1985/devsmith-modular-platform", "mikejsmith1985", "devsmith-modular-platform"},
		{"https://github.com/golang/go", "golang", "go"},
		{"git@github.com:mikejsmith1985/test-repo.git", "mikejsmith1985", "test-repo"},
	}

	client := &DefaultClient{}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			// WHEN: Validating URL
			owner, repo, err := client.ValidateURL(tt.url)

			// THEN: Should succeed with correct values
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOwner, owner)
			assert.Equal(t, tt.expectedRepo, repo)
		})
	}
}

func TestValidateURL_InvalidURL(t *testing.T) {
	// GIVEN: Invalid GitHub URLs
	tests := []string{
		"not-a-url",
		"https://example.com/repo",
		"https://github.com/only-owner",
		"",
	}

	client := &DefaultClient{}

	for _, url := range tests {
		t.Run(url, func(t *testing.T) {
			// WHEN: Validating invalid URL
			_, _, err := client.ValidateURL(url)

			// THEN: Should return error
			assert.Error(t, err)
		})
	}
}

func TestFetchCode_Success(t *testing.T) {
	// GIVEN: Mock client with successful fetch
	mock := &MockClient{
		fetchCodeFunc: func(ctx context.Context, owner, repo, branch string, token string) (*CodeFetch, error) {
			return &CodeFetch{
				Code:      "package main\n\nfunc main() {}",
				Branch:    branch,
				CommitSHA: "abc123def456",
				Metadata: &RepoMetadata{
					Owner:      owner,
					Name:       repo,
					IsPrivate:  false,
					StarsCount: 100,
				},
				FetchedAt: time.Now(),
			}, nil
		},
	}

	// WHEN: Fetching code
	result, err := mock.FetchCode(context.Background(), "golang", "go", "master", "token")

	// THEN: Should succeed
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Code)
	assert.Equal(t, "golang", result.Metadata.Owner)
	assert.Equal(t, "go", result.Metadata.Name)
}

func TestFetchCode_AuthError(t *testing.T) {
	// GIVEN: Mock client with auth error
	mock := &MockClient{
		fetchCodeFunc: func(ctx context.Context, owner, repo, branch string, token string) (*CodeFetch, error) {
			return nil, &AuthError{
				StatusCode: 401,
				Message:    "Invalid token",
			}
		},
	}

	// WHEN: Fetching code with invalid token
	result, err := mock.FetchCode(context.Background(), "golang", "go", "master", "invalid")

	// THEN: Should return auth error
	assert.Error(t, err)
	assert.Nil(t, result)
	var authErr *AuthError
	assert.True(t, errors.As(err, &authErr))
	assert.Equal(t, 401, authErr.StatusCode)
}

func TestFetchCode_NotFound(t *testing.T) {
	// GIVEN: Mock client with not found error
	mock := &MockClient{
		fetchCodeFunc: func(ctx context.Context, owner, repo, branch string, token string) (*CodeFetch, error) {
			return nil, &NotFoundError{
				Owner: owner,
				Repo:  repo,
			}
		},
	}

	// WHEN: Fetching from non-existent repo
	result, err := mock.FetchCode(context.Background(), "invalid", "repo", "master", "token")

	// THEN: Should return not found error
	assert.Error(t, err)
	assert.Nil(t, result)
	var notFoundErr *NotFoundError
	assert.True(t, errors.As(err, &notFoundErr))
	assert.Equal(t, "invalid", notFoundErr.Owner)
}

func TestGetRepoMetadata_Success(t *testing.T) {
	// GIVEN: Mock client returning metadata
	mock := &MockClient{
		getRepoMetadataFunc: func(ctx context.Context, owner, repo string, token string) (*RepoMetadata, error) {
			return &RepoMetadata{
				Owner:       owner,
				Name:        repo,
				Description: "Go language repository",
				StarsCount:  120000,
				IsPrivate:   false,
				DefaultURL:  "https://github.com/golang/go",
			}, nil
		},
	}

	// WHEN: Getting repo metadata
	metadata, err := mock.GetRepoMetadata(context.Background(), "golang", "go", "token")

	// THEN: Should succeed with correct metadata
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, "golang", metadata.Owner)
	assert.Equal(t, "go", metadata.Name)
	assert.Equal(t, 120000, metadata.StarsCount)
	assert.False(t, metadata.IsPrivate)
}

func TestGetRateLimit_Success(t *testing.T) {
	// GIVEN: Mock client returning rate limit
	resetTime := time.Now().Add(1 * time.Hour)
	mock := &MockClient{
		getRateLimitFunc: func(ctx context.Context, token string) (int, time.Time, error) {
			return 4999, resetTime, nil
		},
	}

	// WHEN: Getting rate limit
	remaining, reset, err := mock.GetRateLimit(context.Background(), "token")

	// THEN: Should succeed
	assert.NoError(t, err)
	assert.Equal(t, 4999, remaining)
	assert.Equal(t, resetTime, reset)
}

func TestGetRateLimit_Exceeded(t *testing.T) {
	// GIVEN: Mock client with rate limit exceeded
	resetTime := time.Now().Add(1 * time.Hour)
	mock := &MockClient{
		getRateLimitFunc: func(ctx context.Context, token string) (int, time.Time, error) {
			return 0, resetTime, &RateLimitError{ResetTime: resetTime}
		},
	}

	// WHEN: Getting rate limit when exceeded
	remaining, _, err := mock.GetRateLimit(context.Background(), "token")

	// THEN: Should return rate limit error
	assert.Error(t, err)
	assert.Equal(t, 0, remaining)
	var rateLimitErr *RateLimitError
	assert.True(t, errors.As(err, &rateLimitErr))
	assert.Equal(t, resetTime, rateLimitErr.ResetTime)
}

func TestCodeFetch_Validation(t *testing.T) {
	// GIVEN: CodeFetch structure
	codeFetch := &CodeFetch{
		Code:      "package main",
		Branch:    "main",
		CommitSHA: "abc123",
		Metadata: &RepoMetadata{
			Owner: "test",
			Name:  "repo",
		},
		FetchedAt: time.Now(),
	}

	// THEN: All fields should be accessible
	require.NotNil(t, codeFetch)
	assert.NotEmpty(t, codeFetch.Code)
	assert.Equal(t, "main", codeFetch.Branch)
	assert.NotEmpty(t, codeFetch.CommitSHA)
	assert.NotNil(t, codeFetch.Metadata)
}
