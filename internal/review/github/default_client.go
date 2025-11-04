package github

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// DefaultClient implements ClientInterface using GitHub REST API
type DefaultClient struct {
	baseURL string
}

// NewDefaultClient creates a new GitHub API client
func NewDefaultClient() *DefaultClient {
	return &DefaultClient{
		baseURL: "https://api.github.com",
	}
}

// ValidateURL parses and validates a GitHub repository URL
func (c *DefaultClient) ValidateURL(urlStr string) (owner, repo string, err error) {
	if urlStr == "" {
		return "", "", &URLParseError{URL: urlStr, Reason: "empty url"}
	}

	// Handle different GitHub URL formats
	// Format 1: https://github.com/owner/repo
	// Format 2: https://github.com/owner/repo.git
	// Format 3: git@github.com:owner/repo.git

	if strings.HasPrefix(urlStr, "git@github.com:") {
		// SSH format: git@github.com:owner/repo.git
		parts := strings.TrimPrefix(urlStr, "git@github.com:")
		parts = strings.TrimSuffix(parts, ".git")
		segments := strings.Split(parts, "/")
		if len(segments) != 2 || segments[0] == "" || segments[1] == "" {
			return "", "", &URLParseError{URL: urlStr, Reason: "invalid ssh format"}
		}
		return segments[0], segments[1], nil
	}

	if strings.HasPrefix(urlStr, "https://") || strings.HasPrefix(urlStr, "http://") {
		// HTTPS format
		u, err := url.Parse(urlStr)
		if err != nil {
			return "", "", &URLParseError{URL: urlStr, Reason: err.Error()}
		}

		if u.Host != "github.com" {
			return "", "", &URLParseError{URL: urlStr, Reason: "not a github.com url"}
		}

		// Extract path segments
		path := strings.Trim(u.Path, "/")
		segments := strings.Split(path, "/")

		if len(segments) < 2 || segments[0] == "" || segments[1] == "" {
			return "", "", &URLParseError{URL: urlStr, Reason: "invalid url format"}
		}

		owner := segments[0]
		repo := strings.TrimSuffix(segments[1], ".git")

		return owner, repo, nil
	}

	return "", "", &URLParseError{URL: urlStr, Reason: "unsupported url format"}
}

// FetchCode retrieves code from a GitHub repository (stub implementation)
func (c *DefaultClient) FetchCode(ctx context.Context, owner, repo, branch, token string) (*CodeFetch, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	if owner == "" || repo == "" {
		return nil, &URLParseError{URL: "", Reason: "owner or repo is empty"}
	}

	// Stub implementation - returns mock data
	// In production, this would make actual GitHub API calls
	return &CodeFetch{
		Code:      "// Code would be fetched from GitHub API",
		Branch:    branch,
		CommitSHA: "stub_commit_sha",
		Metadata: &RepoMetadata{
			Owner:       owner,
			Name:        repo,
			Description: "Repository fetched from GitHub",
			StarsCount:  0,
			IsPrivate:   false,
			DefaultURL:  fmt.Sprintf("https://github.com/%s/%s", owner, repo),
		},
		FetchedAt: time.Now(),
	}, nil
}

// GetRepoMetadata retrieves repository metadata (stub implementation)
func (c *DefaultClient) GetRepoMetadata(ctx context.Context, owner, repo, token string) (*RepoMetadata, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	if owner == "" || repo == "" {
		return nil, &URLParseError{URL: "", Reason: "owner or repo is empty"}
	}

	// Stub implementation - returns mock metadata
	// In production, this would call GitHub API
	return &RepoMetadata{
		Owner:       owner,
		Name:        repo,
		Description: "Repository metadata from GitHub",
		StarsCount:  0,
		IsPrivate:   false,
		DefaultURL:  fmt.Sprintf("https://github.com/%s/%s", owner, repo),
	}, nil
}

// GetRateLimit retrieves GitHub API rate limit information (stub implementation)
func (c *DefaultClient) GetRateLimit(ctx context.Context, token string) (remaining int, resetTime time.Time, err error) {
	if ctx.Err() != nil {
		return 0, time.Time{}, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	if token == "" {
		return 0, time.Time{}, &AuthError{StatusCode: 401, Message: "token is empty"}
	}

	// Stub implementation - returns mock data
	// In production, this would call GitHub API /rate_limit endpoint
	return 5000, time.Now().Add(1 * time.Hour), nil
}

// GetRepoTree retrieves the complete file tree structure for a repository
func (c *DefaultClient) GetRepoTree(ctx context.Context, owner, repo, branch, token string) (*RepoTree, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	if owner == "" || repo == "" {
		return nil, &URLParseError{URL: "", Reason: "owner or repo is empty"}
	}

	if branch == "" {
		branch = "main"
	}

	// Stub implementation - returns mock tree structure
	// In production, this would call GitHub API /repos/{owner}/{repo}/git/trees/{sha}?recursive=1
	return &RepoTree{
		Owner:  owner,
		Repo:   repo,
		Branch: branch,
		RootNodes: []TreeNode{
			{
				Path: "README.md",
				Type: "file",
				SHA:  "stub_readme_sha",
				Size: 1024,
			},
			{
				Path: "go.mod",
				Type: "file",
				SHA:  "stub_gomod_sha",
				Size: 512,
			},
			{
				Path: "internal",
				Type: "dir",
				SHA:  "stub_internal_sha",
				Children: []TreeNode{
					{
						Path: "internal/main.go",
						Type: "file",
						SHA:  "stub_main_sha",
						Size: 2048,
					},
				},
			},
		},
	}, nil
}

// GetFileContent retrieves the content of a specific file from the repository
func (c *DefaultClient) GetFileContent(ctx context.Context, owner, repo, path, branch, token string) (*FileContent, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	if owner == "" || repo == "" {
		return nil, &URLParseError{URL: "", Reason: "owner or repo is empty"}
	}

	if path == "" {
		return nil, &URLParseError{URL: "", Reason: "path is empty"}
	}

	if branch == "" {
		branch = "main"
	}

	// Stub implementation - returns mock file content
	// In production, this would call GitHub API /repos/{owner}/{repo}/contents/{path}?ref={branch}
	// and decode the base64 content
	return &FileContent{
		Path:    path,
		Content: fmt.Sprintf("// Stub content for file: %s\npackage main\n\nfunc main() {\n\t// Implementation\n}", path),
		SHA:     "stub_file_sha",
		Size:    256,
	}, nil
}

// GetPullRequest retrieves metadata for a specific pull request
func (c *DefaultClient) GetPullRequest(ctx context.Context, owner, repo string, prNumber int, token string) (*PullRequest, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	if owner == "" || repo == "" {
		return nil, &URLParseError{URL: "", Reason: "owner or repo is empty"}
	}

	if prNumber <= 0 {
		return nil, &URLParseError{URL: "", Reason: "invalid PR number"}
	}

	// Stub implementation - returns mock PR metadata
	// In production, this would call GitHub API /repos/{owner}/{repo}/pulls/{prNumber}
	return &PullRequest{
		Number:      prNumber,
		Title:       fmt.Sprintf("Pull Request #%d", prNumber),
		Description: "Stub PR description",
		Author:      "stubauthor",
		State:       "open",
		HeadBranch:  "feature/stub",
		BaseBranch:  "main",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now(),
	}, nil
}

// GetPRFiles retrieves the list of files changed in a pull request
func (c *DefaultClient) GetPRFiles(ctx context.Context, owner, repo string, prNumber int, token string) ([]PRFile, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	if owner == "" || repo == "" {
		return nil, &URLParseError{URL: "", Reason: "owner or repo is empty"}
	}

	if prNumber <= 0 {
		return nil, &URLParseError{URL: "", Reason: "invalid PR number"}
	}

	// Stub implementation - returns mock PR files
	// In production, this would call GitHub API /repos/{owner}/{repo}/pulls/{prNumber}/files
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
	}, nil
}
