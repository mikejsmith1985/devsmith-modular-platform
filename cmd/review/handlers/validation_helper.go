// Package handlers provides HTTP request handlers for the review service.
// The validation_helper module provides composite validation functions for use in handlers.
package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
)

// ValidateRequest runs a series of validator functions and returns HTTP error if any fail.
// This is a helper to streamline validation in handlers and provide consistent error responses.
//
// Parameters:
//   - c: The Gin context for writing HTTP responses
//   - validators: Variable number of validator functions to run in sequence
//
// Returns:
//   - bool: true if all validators pass, false if any fails (response already written)
func ValidateRequest(c *gin.Context, validators ...func() error) bool {
	for _, validator := range validators {
		if err := validator(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "validation error",
				"detail": err.Error(),
			})
			return false
		}
	}
	return true
}

// ValidateCreateReviewRequest validates all fields for a review creation request.
// Performs conditional validation based on code source:
//   - title: Always validated (required, max length)
//   - codeSource: Always validated (must be paste/github/upload)
//   - pastedCode: Validated if codeSource is "paste" (size limit)
//   - githubRepo: Validated if codeSource is "github" (URL format)
//   - githubBranch: Validated if codeSource is "github" (naming rules)
//
// Parameters:
//   - title: The review session title
//   - codeSource: The source type (paste, github, upload)
//   - pastedCode: The code content (if source is paste)
//   - githubRepo: The GitHub repository URL (if source is github)
//   - githubBranch: The GitHub branch name (if source is github)
//
// Returns:
//   - error: nil if all applicable validations pass, wrapped error with field name on failure
func ValidateCreateReviewRequest(title, codeSource, pastedCode, githubRepo, githubBranch string) error {
	// Validate title (always required)
	if err := services.ValidateTitle(title); err != nil {
		return fmt.Errorf("title: %w", err)
	}

	// Validate code source (always required)
	if err := services.ValidateCodeSource(codeSource); err != nil {
		return fmt.Errorf("code_source: %w", err)
	}

	// Conditional validation: pasted code
	if codeSource == "paste" && pastedCode != "" {
		if err := services.ValidateCodeContent(pastedCode); err != nil {
			return fmt.Errorf("pasted_code: %w", err)
		}
	}

	// Conditional validation: GitHub repository
	if codeSource == "github" && githubRepo != "" {
		if err := services.ValidateGitHubURL(githubRepo); err != nil {
			return fmt.Errorf("github_repo: %w", err)
		}
	}

	// Conditional validation: GitHub branch
	if codeSource == "github" && githubBranch != "" {
		if err := services.ValidateGitHubBranch(githubBranch); err != nil {
			return fmt.Errorf("github_branch: %w", err)
		}
	}

	return nil
}

// ValidateScanRequest validates parameters for a code scan analysis request.
// Validates both the reading mode and the semantic search query.
//
// Parameters:
//   - readingMode: The analysis mode (preview/skim/scan/detailed/critical)
//   - query: The semantic search query (2-1000 characters)
//
// Returns:
//   - error: nil if all validations pass, wrapped error with field name on failure
func ValidateScanRequest(readingMode, query string) error {
	// Validate reading mode
	if err := services.ValidateReadingMode(readingMode); err != nil {
		return fmt.Errorf("reading_mode: %w", err)
	}

	// Validate query
	if err := services.ValidateScanQuery(query); err != nil {
		return fmt.Errorf("query: %w", err)
	}

	return nil
}

// ValidateReadingModeRequest validates a reading mode enum value.
// Ensures the mode is one of the five supported analysis modes.
//
// Parameters:
//   - mode: The reading mode to validate
//
// Returns:
//   - error: nil if mode is valid, error otherwise
func ValidateReadingModeRequest(mode string) error {
	return services.ValidateReadingMode(mode)
}
