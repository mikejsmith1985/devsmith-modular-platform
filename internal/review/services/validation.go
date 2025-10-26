// Package services provides business logic for code review analysis.
// The validation subpackage ensures all inputs are properly validated for security and correctness.
package services

import (
	"errors"
	"fmt"
	"html"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// Constants for validation limits and bounds
const (
	// MaxCodeSize is the maximum allowed code size in bytes (10MB)
	MaxCodeSize = 10 * 1024 * 1024

	// MinQueryLength is the minimum query string length for Scan mode
	MinQueryLength = 2

	// MaxQueryLength is the maximum query string length for Scan mode
	MaxQueryLength = 1000

	// MaxTitleLength is the maximum review session title length
	MaxTitleLength = 255
)

// ValidateCodeContent validates code content input for size and presence.
// It ensures the code is not empty and does not exceed the 10MB limit.
// This protects against memory exhaustion and DoS attacks.
//
// Parameters:
//   - code: The code content to validate
//
// Returns:
//   - error: nil if valid, otherwise describes the validation failure
func ValidateCodeContent(code string) error {
	if code == "" {
		return errors.New("code content is empty")
	}

	if len(code) > MaxCodeSize {
		return fmt.Errorf("code exceeds maximum size of %d bytes (%.1f MB)", MaxCodeSize, float64(MaxCodeSize)/1024/1024)
	}

	return nil
}

// ValidateScanQuery validates search query input for length and presence.
// Scan queries drive semantic code search, so they must be meaningful (minimum length)
// but not excessively long to prevent performance issues.
//
// Parameters:
//   - query: The search query to validate
//
// Returns:
//   - error: nil if valid, otherwise describes the validation failure
func ValidateScanQuery(query string) error {
	if query == "" {
		return errors.New("query cannot be empty")
	}

	if len(query) < MinQueryLength {
		return fmt.Errorf("query must be at least %d characters", MinQueryLength)
	}

	if len(query) > MaxQueryLength {
		return fmt.Errorf("query exceeds maximum length of %d characters", MaxQueryLength)
	}

	return nil
}

// ValidateReadingMode validates that the reading mode is one of the five supported modes.
// Reading modes control how AI analysis is performed:
//   - preview: High-level structural overview
//   - skim: Function signatures and abstractions only
//   - scan: Targeted semantic search
//   - detailed: Line-by-line explanation
//   - critical: Quality evaluation and issue detection
//
// Parameters:
//   - mode: The reading mode to validate
//
// Returns:
//   - error: nil if valid, otherwise describes the validation failure
func ValidateReadingMode(mode string) error {
	validModes := map[string]bool{
		"preview":  true,
		"skim":     true,
		"scan":     true,
		"detailed": true,
		"critical": true,
	}

	if !validModes[mode] {
		return fmt.Errorf("invalid reading mode: %s (valid: preview, skim, scan, detailed, critical)", mode)
	}

	return nil
}

// ValidateGitHubURL validates GitHub repository URLs for format and domain.
// Accepts both full HTTPS URLs and shorthand github.com/user/repo format.
// Prevents non-GitHub sources and ensures proper URL structure.
//
// Parameters:
//   - urlStr: The GitHub URL to validate (https://github.com/user/repo or github.com/user/repo)
//
// Returns:
//   - error: nil if valid, otherwise describes the validation failure
func ValidateGitHubURL(urlStr string) error {
	if urlStr == "" {
		return errors.New("GitHub URL cannot be empty")
	}

	// Normalize URL format
	urlStr = strings.TrimSpace(urlStr)

	// Add https:// if missing
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		if strings.HasPrefix(urlStr, "github.com/") {
			urlStr = "https://" + urlStr
		} else {
			return fmt.Errorf("invalid GitHub URL format: must be github.com or full HTTPS URL")
		}
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Verify it's GitHub (prevent redirect to malicious sites)
	if !strings.Contains(parsedURL.Host, "github.com") {
		return fmt.Errorf("URL must be from github.com, got: %s", parsedURL.Host)
	}

	// Verify HTTPS or http only (no ftp, gopher, etc.)
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTP(S) URLs supported, got: %s", parsedURL.Scheme)
	}

	// Verify path has owner/repo (prevents incomplete URLs like github.com/user)
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return fmt.Errorf("GitHub URL must include owner/repo: %s", urlStr)
	}

	return nil
}

// ValidateFilePath validates file paths to prevent directory traversal attacks.
// Rejects absolute paths, path traversal attempts (..), and Windows absolute paths.
// Ensures file paths stay within the intended code review scope.
//
// Security considerations:
//   - Prevents ../../../etc/passwd attacks
//   - Prevents C:\Windows\System32 attacks
//   - Allows relative paths like src/handler.go
//
// Parameters:
//   - path: The file path to validate
//
// Returns:
//   - error: nil if valid, otherwise describes the validation failure
func ValidateFilePath(path string) error {
	if path == "" {
		return errors.New("file path cannot be empty")
	}

	// Reject absolute paths
	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute paths not allowed: %s", path)
	}

	// Reject Windows absolute paths (e.g., C:\)
	if len(path) > 2 && path[1] == ':' {
		return fmt.Errorf("absolute paths not allowed: %s", path)
	}

	// Reject path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed: %s", path)
	}

	// Normalize and check again after normalization (catch obfuscated traversals)
	normalized := filepath.Clean(path)
	if strings.Contains(normalized, "..") {
		return fmt.Errorf("path traversal not allowed: %s", path)
	}

	// Ensure path doesn't start with / (absolute path detection for Unix)
	if strings.HasPrefix(normalized, "/") {
		return fmt.Errorf("absolute paths not allowed: %s", path)
	}

	return nil
}

// SanitizeCodeForDisplay sanitizes code for safe display in HTML contexts.
// Escapes HTML special characters to prevent XSS attacks.
// This is crucial because user-provided code is displayed in the browser.
//
// Security note:
// This function escapes HTML entities (&, <, >, ", ') to their HTML entity equivalents.
// This prevents JavaScript execution while preserving code readability.
//
// Parameters:
//   - code: The code string to sanitize
//
// Returns:
//   - string: The sanitized code safe for HTML display
func SanitizeCodeForDisplay(code string) string {
	// html.EscapeString converts:
	// & -> &amp;
	// < -> &lt;
	// > -> &gt;
	// " -> &#34;
	// ' -> &#39;
	escaped := html.EscapeString(code)
	return escaped
}

// ValidateTitle validates review session titles for length and presence.
// Titles are user-provided identifiers for code review sessions and should
// be meaningful but reasonably sized.
//
// Parameters:
//   - title: The review session title to validate
//
// Returns:
//   - error: nil if valid, otherwise describes the validation failure
func ValidateTitle(title string) error {
	if title == "" {
		return errors.New("title cannot be empty")
	}

	if len(title) > MaxTitleLength {
		return fmt.Errorf("title exceeds maximum length of %d characters", MaxTitleLength)
	}

	return nil
}

// ValidateCodeSource validates the code source type enum.
// Code can come from three sources, each handled differently:
//   - paste: Direct code paste from user
//   - github: GitHub repository URL
//   - upload: File upload (future feature)
//
// Parameters:
//   - source: The code source type to validate
//
// Returns:
//   - error: nil if valid, otherwise describes the validation failure
func ValidateCodeSource(source string) error {
	validSources := map[string]bool{
		"paste":   true,
		"github":  true,
		"upload":  true,
	}

	if !validSources[source] {
		return fmt.Errorf("invalid code source: %s (valid: paste, github, upload)", source)
	}

	return nil
}

// ValidateGitHubBranch validates GitHub branch names for GitHub branch naming rules.
// Prevents path traversal in branch names and enforces GitHub's branch naming conventions.
// Branch names can contain alphanumerics, dots, dashes, underscores, and forward slashes.
//
// Security considerations:
//   - Prevents ../../../malicious attacks in branch names
//   - Prevents /etc/passwd attacks
//   - Allows common patterns like feature/user-auth, release/v1.0.0
//
// Parameters:
//   - branch: The GitHub branch name to validate
//
// Returns:
//   - error: nil if valid, otherwise describes the validation failure
func ValidateGitHubBranch(branch string) error {
	if branch == "" {
		return errors.New("GitHub branch cannot be empty")
	}

	// Reject path traversal attempts
	if strings.Contains(branch, "..") {
		return fmt.Errorf("path traversal in branch name not allowed: %s", branch)
	}

	// Reject absolute paths
	if strings.HasPrefix(branch, "/") || (len(branch) > 2 && branch[1] == ':') {
		return fmt.Errorf("absolute paths in branch name not allowed: %s", branch)
	}

	// Valid branch names: alphanumeric, -, _, / for feature/xxx format
	// Based on GitHub's branch naming rules: https://github.com/blog/821-tree-branch-name-validation
	validBranchPattern := regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
	if !validBranchPattern.MatchString(branch) {
		return fmt.Errorf("invalid branch name format: %s (allowed: alphanumeric, dash, underscore, dot, slash)", branch)
	}

	return nil
}
