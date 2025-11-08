// Package review_services provides business logic services for the Review Service
package review_services

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/repositories"
)

// Constants for reading modes and variables
const (
	ModeScan     = "scan"
	VarCode      = "{{code}}"
	VarQuery     = "{{query}}"
	VarFile      = "{{file}}"
	VarUserLevel = "{{user_level}}"
)

// Error messages
const (
	ErrMissingVariable     = "missing variable value for %s"
	ErrMissingRequired     = "missing required variable %s for mode %s"
	ErrNoDefault           = "no default prompt found for mode=%s, userLevel=%s, outputMode=%s"
	ErrTemplateIDRequired  = "template_id is required"
	ErrUserIDRequired      = "user_id is required"
	ErrModelUsedRequired   = "model_used is required"
)

// Variable extraction regex pattern
var variablePattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)

// PromptTemplateService provides business logic for prompt template management
type PromptTemplateService struct {
	repo repositories.PromptTemplateRepositoryInterface
}

// NewPromptTemplateService creates a new prompt template service
func NewPromptTemplateService(repo repositories.PromptTemplateRepositoryInterface) *PromptTemplateService {
	return &PromptTemplateService{
		repo: repo,
	}
}

// GetEffectivePrompt returns the effective prompt for a user.
// It first attempts to retrieve a user-specific custom prompt. If none exists,
// it falls back to the system default prompt for the given parameters.
// Returns an error if neither user custom nor system default is found.
func (s *PromptTemplateService) GetEffectivePrompt(ctx context.Context, userID int, mode, userLevel, outputMode string) (*review_models.PromptTemplate, error) {
	// Try to get user custom first
	userPrompt, err := s.repo.FindByUserAndMode(ctx, userID, mode, userLevel, outputMode)
	if err != nil {
		return nil, fmt.Errorf("error fetching user prompt: %w", err)
	}

	if userPrompt != nil {
		return userPrompt, nil
	}

	// Fall back to system default
	defaultPrompt, err := s.repo.FindDefaultByMode(ctx, mode, userLevel, outputMode)
	if err != nil {
		return nil, fmt.Errorf("error fetching default prompt: %w", err)
	}

	if defaultPrompt == nil {
		return nil, fmt.Errorf(ErrNoDefault, mode, userLevel, outputMode)
	}

	return defaultPrompt, nil
}

// SaveCustomPrompt validates and saves a user's custom prompt.
// It extracts variables from the prompt text, validates that all required
// variables for the given mode are present, and saves the custom prompt
// to the database. Returns the saved template or an error.
func (s *PromptTemplateService) SaveCustomPrompt(ctx context.Context, userID int, mode, userLevel, outputMode, promptText string) (*review_models.PromptTemplate, error) {
	// Extract variables from prompt text
	variables := s.ExtractVariables(promptText)

	// Validate required variables are present
	if err := s.validateRequiredVariables(mode, variables); err != nil {
		return nil, err
	}

	// Create template
	template := &review_models.PromptTemplate{
		ID:         fmt.Sprintf("custom-%d-%s-%s-%s", userID, mode, userLevel, outputMode),
		UserID:     &userID,
		Mode:       mode,
		UserLevel:  userLevel,
		OutputMode: outputMode,
		PromptText: promptText,
		Variables:  variables,
		IsDefault:  false,
		Version:    1,
	}

	// Save to database
	saved, err := s.repo.Upsert(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("error saving custom prompt: %w", err)
	}

	return saved, nil
}

// validateRequiredVariables checks if all required variables for a mode are present
func (s *PromptTemplateService) validateRequiredVariables(mode string, variables []string) error {
	requiredVars := getRequiredVariables(mode)
	for _, required := range requiredVars {
		found := false
		for _, v := range variables {
			if v == required {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf(ErrMissingRequired, required, mode)
		}
	}
	return nil
}

// FactoryReset deletes a user's custom prompt and restores the system default.
// This removes the user's customization for the specified mode, user level,
// and output mode combination.
func (s *PromptTemplateService) FactoryReset(ctx context.Context, userID int, mode, userLevel, outputMode string) error {
	if err := s.repo.DeleteUserCustom(ctx, userID, mode, userLevel, outputMode); err != nil {
		return fmt.Errorf("error deleting custom prompt: %w", err)
	}
	return nil
}

// RenderPrompt substitutes all template variables with their actual values.
// Returns an error if any template variable is missing from the variables map.
func (s *PromptTemplateService) RenderPrompt(template *review_models.PromptTemplate, variables map[string]string) (string, error) {
	rendered := template.PromptText

	// Check all template variables have values
	for _, variable := range template.Variables {
		value, exists := variables[variable]
		if !exists {
			return "", fmt.Errorf(ErrMissingVariable, variable)
		}
		rendered = strings.ReplaceAll(rendered, variable, value)
	}

	return rendered, nil
}

// LogExecution records a prompt execution with validation of required fields.
// Returns an error if template_id, user_id, or model_used is missing.
func (s *PromptTemplateService) LogExecution(ctx context.Context, execution *review_models.PromptExecution) error {
	// Validate required fields
	if execution.TemplateID == "" {
		return fmt.Errorf(ErrTemplateIDRequired)
	}
	if execution.UserID == 0 {
		return fmt.Errorf(ErrUserIDRequired)
	}
	if execution.ModelUsed == "" {
		return fmt.Errorf(ErrModelUsedRequired)
	}

	return s.repo.SaveExecution(ctx, execution)
}

// ExtractVariables finds all {{variable}} patterns in text.
// Variables are deduplicated and returned in a slice.
func (s *PromptTemplateService) ExtractVariables(text string) []string {
	matches := variablePattern.FindAllStringSubmatch(text, -1)

	// Use map to deduplicate
	uniqueVars := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 0 {
			uniqueVars[match[0]] = true // match[0] is the full match including {{}}
		}
	}

	// Convert to slice
	result := make([]string, 0, len(uniqueVars))
	for v := range uniqueVars {
		result = append(result, v)
	}

	return result
}

// getRequiredVariables returns the required variables for a given reading mode.
// All modes require {{code}}, and scan mode additionally requires {{query}}.
func getRequiredVariables(mode string) []string {
	switch mode {
	case ModeScan:
		return []string{VarCode, VarQuery}
	default:
		return []string{VarCode}
	}
}
