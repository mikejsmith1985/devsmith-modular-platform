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
	getRepoTreeFunc     func(ctx context.Context, owner, repo, branch, token string) (*RepoTree, error)
	getFileContentFunc  func(ctx context.Context, owner, repo, path, branch, token string) (*FileContent, error)
	getPullRequestFunc  func(ctx context.Context, owner, repo string, prNumber int, token string) (*PullRequest, error)
	getPRFilesFunc      func(ctx context.Context, owner, repo string, prNumber int, token string) ([]PRFile, error)
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

func (m *MockClient) GetRepoTree(ctx context.Context, owner, repo, branch, token string) (*RepoTree, error) {
	if m.getRepoTreeFunc != nil {
		return m.getRepoTreeFunc(ctx, owner, repo, branch, token)
	}
	return nil, nil
}

func (m *MockClient) GetFileContent(ctx context.Context, owner, repo, path, branch, token string) (*FileContent, error) {
	if m.getFileContentFunc != nil {
		return m.getFileContentFunc(ctx, owner, repo, path, branch, token)
	}
	return nil, nil
}

func (m *MockClient) GetPullRequest(ctx context.Context, owner, repo string, prNumber int, token string) (*PullRequest, error) {
	if m.getPullRequestFunc != nil {
		return m.getPullRequestFunc(ctx, owner, repo, prNumber, token)
	}
	return nil, nil
}

func (m *MockClient) GetPRFiles(ctx context.Context, owner, repo string, prNumber int, token string) ([]PRFile, error) {
	if m.getPRFilesFunc != nil {
		return m.getPRFilesFunc(ctx, owner, repo, prNumber, token)
	}
	return nil, nil
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

// ===== Phase 2: Repository Tree Tests =====

func TestGetRepoTree_Success(t *testing.T) {
	// GIVEN: Mock client returning repository tree
	mock := &MockClient{
		getRepoTreeFunc: func(ctx context.Context, owner, repo, branch, token string) (*RepoTree, error) {
			return &RepoTree{
				Owner:  owner,
				Repo:   repo,
				Branch: branch,
				RootNodes: []TreeNode{
					{
						Path: "README.md",
						Type: "file",
						SHA:  "abc123",
						Size: 1024,
					},
					{
						Path: "src",
						Type: "dir",
						SHA:  "def456",
						Children: []TreeNode{
							{
								Path: "src/main.go",
								Type: "file",
								SHA:  "ghi789",
								Size: 2048,
							},
						},
					},
				},
			}, nil
		},
	}

	// WHEN: Getting repository tree
	tree, err := mock.GetRepoTree(context.Background(), "golang", "go", "main", "token")

	// THEN: Should succeed with tree structure
	assert.NoError(t, err)
	require.NotNil(t, tree)
	assert.Equal(t, "golang", tree.Owner)
	assert.Equal(t, "go", tree.Repo)
	assert.Equal(t, "main", tree.Branch)
	assert.Len(t, tree.RootNodes, 2)
	assert.Equal(t, "README.md", tree.RootNodes[0].Path)
	assert.Equal(t, "file", tree.RootNodes[0].Type)
	assert.Equal(t, "src", tree.RootNodes[1].Path)
	assert.Equal(t, "dir", tree.RootNodes[1].Type)
	assert.Len(t, tree.RootNodes[1].Children, 1)
	assert.Equal(t, "src/main.go", tree.RootNodes[1].Children[0].Path)
}

func TestGetRepoTree_NotFound(t *testing.T) {
	// GIVEN: Mock client with not found error
	mock := &MockClient{
		getRepoTreeFunc: func(ctx context.Context, owner, repo, branch, token string) (*RepoTree, error) {
			return nil, &NotFoundError{Owner: owner, Repo: repo}
		},
	}

	// WHEN: Getting tree for non-existent repository
	tree, err := mock.GetRepoTree(context.Background(), "notfound", "repo", "main", "token")

	// THEN: Should return not found error
	assert.Error(t, err)
	assert.Nil(t, tree)
	var notFoundErr *NotFoundError
	assert.True(t, errors.As(err, &notFoundErr))
}

func TestGetRepoTree_Unauthorized(t *testing.T) {
	// GIVEN: Mock client with auth error
	mock := &MockClient{
		getRepoTreeFunc: func(ctx context.Context, owner, repo, branch, token string) (*RepoTree, error) {
			return nil, &AuthError{Message: "invalid token"}
		},
	}

	// WHEN: Getting tree with invalid token
	tree, err := mock.GetRepoTree(context.Background(), "private", "repo", "main", "bad_token")

	// THEN: Should return auth error
	assert.Error(t, err)
	assert.Nil(t, tree)
	var authErr *AuthError
	assert.True(t, errors.As(err, &authErr))
}

// ===== Phase 2: File Content Tests =====

func TestGetFileContent_Success(t *testing.T) {
	// GIVEN: Mock client returning file content
	mock := &MockClient{
		getFileContentFunc: func(ctx context.Context, owner, repo, path, branch, token string) (*FileContent, error) {
			return &FileContent{
				Path:    path,
				Content: "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}",
				SHA:     "abc123",
				Size:    50,
			}, nil
		},
	}

	// WHEN: Getting file content
	content, err := mock.GetFileContent(context.Background(), "golang", "go", "main.go", "main", "token")

	// THEN: Should succeed with file content
	assert.NoError(t, err)
	require.NotNil(t, content)
	assert.Equal(t, "main.go", content.Path)
	assert.Contains(t, content.Content, "package main")
	assert.Equal(t, "abc123", content.SHA)
	assert.Equal(t, int64(50), content.Size)
}

func TestGetFileContent_NotFound(t *testing.T) {
	// GIVEN: Mock client with not found error
	mock := &MockClient{
		getFileContentFunc: func(ctx context.Context, owner, repo, path, branch, token string) (*FileContent, error) {
			return nil, &NotFoundError{Owner: owner, Repo: repo}
		},
	}

	// WHEN: Getting non-existent file
	content, err := mock.GetFileContent(context.Background(), "golang", "go", "missing.go", "main", "token")

	// THEN: Should return not found error
	assert.Error(t, err)
	assert.Nil(t, content)
	var notFoundErr *NotFoundError
	assert.True(t, errors.As(err, &notFoundErr))
}

func TestGetFileContent_BinaryFile(t *testing.T) {
	// GIVEN: Mock client handling binary file
	mock := &MockClient{
		getFileContentFunc: func(ctx context.Context, owner, repo, path, branch, token string) (*FileContent, error) {
			return &FileContent{
				Path:    path,
				Content: "[binary content]",
				SHA:     "def456",
				Size:    10240,
			}, nil
		},
	}

	// WHEN: Getting binary file content
	content, err := mock.GetFileContent(context.Background(), "test", "repo", "image.png", "main", "token")

	// THEN: Should succeed with binary indicator
	assert.NoError(t, err)
	require.NotNil(t, content)
	assert.Equal(t, "image.png", content.Path)
	assert.Contains(t, content.Content, "[binary content]")
}

// ===== Phase 2: Pull Request Tests =====

func TestGetPullRequest_Success(t *testing.T) {
	// GIVEN: Mock client returning PR metadata
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()
	mock := &MockClient{
		getPullRequestFunc: func(ctx context.Context, owner, repo string, prNumber int, token string) (*PullRequest, error) {
			return &PullRequest{
				Number:      prNumber,
				Title:       "Add new feature",
				Description: "This PR adds a new feature",
				Author:      "testuser",
				State:       "open",
				HeadBranch:  "feature/new-feature",
				BaseBranch:  "main",
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			}, nil
		},
	}

	// WHEN: Getting pull request
	pr, err := mock.GetPullRequest(context.Background(), "test", "repo", 42, "token")

	// THEN: Should succeed with PR metadata
	assert.NoError(t, err)
	require.NotNil(t, pr)
	assert.Equal(t, 42, pr.Number)
	assert.Equal(t, "Add new feature", pr.Title)
	assert.Equal(t, "testuser", pr.Author)
	assert.Equal(t, "open", pr.State)
	assert.Equal(t, "feature/new-feature", pr.HeadBranch)
	assert.Equal(t, "main", pr.BaseBranch)
	assert.Equal(t, createdAt, pr.CreatedAt)
	assert.Equal(t, updatedAt, pr.UpdatedAt)
}

func TestGetPullRequest_NotFound(t *testing.T) {
	// GIVEN: Mock client with not found error
	mock := &MockClient{
		getPullRequestFunc: func(ctx context.Context, owner, repo string, prNumber int, token string) (*PullRequest, error) {
			return nil, &NotFoundError{Owner: owner, Repo: repo}
		},
	}

	// WHEN: Getting non-existent PR
	pr, err := mock.GetPullRequest(context.Background(), "test", "repo", 999, "token")

	// THEN: Should return not found error
	assert.Error(t, err)
	assert.Nil(t, pr)
	var notFoundErr *NotFoundError
	assert.True(t, errors.As(err, &notFoundErr))
}

func TestGetPullRequest_Merged(t *testing.T) {
	// GIVEN: Mock client returning merged PR
	mock := &MockClient{
		getPullRequestFunc: func(ctx context.Context, owner, repo string, prNumber int, token string) (*PullRequest, error) {
			return &PullRequest{
				Number:     prNumber,
				Title:      "Merged PR",
				State:      "merged",
				HeadBranch: "feature/done",
				BaseBranch: "main",
			}, nil
		},
	}

	// WHEN: Getting merged PR
	pr, err := mock.GetPullRequest(context.Background(), "test", "repo", 10, "token")

	// THEN: Should succeed with merged state
	assert.NoError(t, err)
	require.NotNil(t, pr)
	assert.Equal(t, "merged", pr.State)
}

// ===== Phase 2: PR Files Tests =====

func TestGetPRFiles_Success(t *testing.T) {
	// GIVEN: Mock client returning PR files
	mock := &MockClient{
		getPRFilesFunc: func(ctx context.Context, owner, repo string, prNumber int, token string) ([]PRFile, error) {
			return []PRFile{
				{
					Filename:  "main.go",
					Status:    "modified",
					Additions: 10,
					Deletions: 5,
					Changes:   15,
					Patch:     "@@ -1,3 +1,8 @@\n package main\n+\n+func newFunc() {}",
				},
				{
					Filename:  "test.go",
					Status:    "added",
					Additions: 20,
					Deletions: 0,
					Changes:   20,
					Patch:     "@@ -0,0 +1,20 @@\n+package main\n+\n+func Test() {}",
				},
				{
					Filename:  "old.go",
					Status:    "removed",
					Additions: 0,
					Deletions: 30,
					Changes:   30,
					Patch:     "@@ -1,30 +0,0 @@\n-package old",
				},
			}, nil
		},
	}

	// WHEN: Getting PR files
	files, err := mock.GetPRFiles(context.Background(), "test", "repo", 42, "token")

	// THEN: Should succeed with file list
	assert.NoError(t, err)
	require.Len(t, files, 3)

	// Check modified file
	assert.Equal(t, "main.go", files[0].Filename)
	assert.Equal(t, "modified", files[0].Status)
	assert.Equal(t, 10, files[0].Additions)
	assert.Equal(t, 5, files[0].Deletions)
	assert.Equal(t, 15, files[0].Changes)
	assert.Contains(t, files[0].Patch, "func newFunc()")

	// Check added file
	assert.Equal(t, "test.go", files[1].Filename)
	assert.Equal(t, "added", files[1].Status)
	assert.Equal(t, 20, files[1].Additions)
	assert.Equal(t, 0, files[1].Deletions)

	// Check removed file
	assert.Equal(t, "old.go", files[2].Filename)
	assert.Equal(t, "removed", files[2].Status)
	assert.Equal(t, 0, files[2].Additions)
	assert.Equal(t, 30, files[2].Deletions)
}

func TestGetPRFiles_Empty(t *testing.T) {
	// GIVEN: Mock client returning empty file list
	mock := &MockClient{
		getPRFilesFunc: func(ctx context.Context, owner, repo string, prNumber int, token string) ([]PRFile, error) {
			return []PRFile{}, nil
		},
	}

	// WHEN: Getting files for PR with no changes
	files, err := mock.GetPRFiles(context.Background(), "test", "repo", 1, "token")

	// THEN: Should succeed with empty list
	assert.NoError(t, err)
	assert.Empty(t, files)
}

func TestGetPRFiles_NotFound(t *testing.T) {
	// GIVEN: Mock client with not found error
	mock := &MockClient{
		getPRFilesFunc: func(ctx context.Context, owner, repo string, prNumber int, token string) ([]PRFile, error) {
			return nil, &NotFoundError{Owner: owner, Repo: repo}
		},
	}

	// WHEN: Getting files for non-existent PR
	files, err := mock.GetPRFiles(context.Background(), "test", "repo", 999, "token")

	// THEN: Should return not found error
	assert.Error(t, err)
	assert.Nil(t, files)
	var notFoundErr *NotFoundError
	assert.True(t, errors.As(err, &notFoundErr))
}
