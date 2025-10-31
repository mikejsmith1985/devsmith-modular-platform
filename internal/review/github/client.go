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
