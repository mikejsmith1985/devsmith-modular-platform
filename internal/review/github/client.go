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
	Size     int64      `json:"size,omitempty"`     // 8 bytes
	Path     string     `json:"path"`               // 16 bytes
	Type     string     `json:"type"`               // 16 bytes - "file" or "dir"
	SHA      string     `json:"sha"`                // 16 bytes
	Children []TreeNode `json:"children,omitempty"` // 24 bytes
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
	Number      int       // 8 bytes
	CreatedAt   time.Time // 24 bytes
	UpdatedAt   time.Time // 24 bytes
	Title       string    // 16 bytes
	Description string    // 16 bytes
	Author      string    // 16 bytes
	State       string    // 16 bytes - "open", "closed", "merged"
	HeadBranch  string    // 16 bytes
	BaseBranch  string    // 16 bytes
}

// PRFile represents a file changed in a pull request
type PRFile struct {
	Additions int    // 8 bytes
	Deletions int    // 8 bytes
	Changes   int    // 8 bytes
	Filename  string // 16 bytes
	Status    string // 16 bytes - "added", "modified", "removed", "renamed"
	Patch     string // 16 bytes - Diff patch
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
