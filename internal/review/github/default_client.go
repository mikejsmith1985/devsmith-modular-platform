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
