// Package github provides GitHub API integration for the Review Service
package github

import (
	"context"
	"time"
)

// RepoMetadata contains GitHub repository information
type RepoMetadata struct {
	Owner       string
	Name        string
	Description string
	DefaultURL  string
	StarsCount  int
	IsPrivate   bool
}

// CodeFetch contains fetched code and metadata
type CodeFetch struct {
	FetchedAt time.Time
	Metadata  *RepoMetadata
	Code      string
	CommitSHA string
	Branch    string
}

// TreeNode represents a file or directory in the repository tree
type TreeNode struct {
	Path     string     `json:"path"`
	Type     string     `json:"type"` // "file" or "dir"
	SHA      string     `json:"sha"`
	Size     int64      `json:"size,omitempty"`
	Children []TreeNode `json:"children,omitempty"`
}

// RepoTree represents the complete repository file structure
type RepoTree struct {
	Owner     string
	Repo      string
	Branch    string
	RootNodes []TreeNode
}

// FileContent represents the content of a single file
type FileContent struct {
	Path    string
	Content string // Base64 decoded content
	SHA     string
	Size    int64
}

// PullRequest represents GitHub PR metadata
type PullRequest struct {
	Number      int
	Title       string
	Description string
	Author      string
	State       string // "open", "closed", "merged"
	HeadBranch  string
	BaseBranch  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PRFile represents a file changed in a pull request
type PRFile struct {
	Filename  string
	Status    string // "added", "modified", "removed", "renamed"
	Additions int
	Deletions int
	Changes   int
	Patch     string // Diff patch
}

// ClientInterface defines GitHub API operations
type ClientInterface interface {
	// FetchCode retrieves code from a GitHub repository
	FetchCode(ctx context.Context, owner, repo, branch string, token string) (*CodeFetch, error)

	// GetRepoMetadata retrieves repository information
	GetRepoMetadata(ctx context.Context, owner, repo string, token string) (*RepoMetadata, error)

	// ValidateURL parses and validates a GitHub URL
	ValidateURL(url string) (owner string, repo string, err error)

	// GetRateLimit returns remaining API calls for the token
	GetRateLimit(ctx context.Context, token string) (remaining int, resetTime time.Time, err error)

	// GetRepoTree retrieves the complete file tree structure for a repository
	GetRepoTree(ctx context.Context, owner, repo, branch, token string) (*RepoTree, error)

	// GetFileContent retrieves the content of a specific file from the repository
	GetFileContent(ctx context.Context, owner, repo, path, branch, token string) (*FileContent, error)

	// GetPullRequest retrieves metadata for a specific pull request
	GetPullRequest(ctx context.Context, owner, repo string, prNumber int, token string) (*PullRequest, error)

	// GetPRFiles retrieves the list of files changed in a pull request
	GetPRFiles(ctx context.Context, owner, repo string, prNumber int, token string) ([]PRFile, error)
}

// URLParseError indicates URL parsing failed
type URLParseError struct {
	URL    string
	Reason string
}

func (e *URLParseError) Error() string {
	return "invalid github url: " + e.Reason + " (" + e.URL + ")"
}

// AuthError indicates authentication failed
type AuthError struct {
	Message    string
	StatusCode int
}

func (e *AuthError) Error() string {
	return "github auth failed: " + e.Message
}

// NotFoundError indicates repository not found
type NotFoundError struct {
	Owner string
	Repo  string
}

func (e *NotFoundError) Error() string {
	return "github repository not found: " + e.Owner + "/" + e.Repo
}

// RateLimitError indicates rate limit exceeded
type RateLimitError struct {
	ResetTime time.Time
}

func (e *RateLimitError) Error() string {
	return "github api rate limit exceeded, resets at " + e.ResetTime.String()
}
