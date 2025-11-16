// Package logs_services provides business logic services for the DevSmith Logs application.
//
// This package contains service implementations for:
// - Project management and API key generation
// - Cross-repository log ingestion and batching
// - Integration with external logging systems
//
// The services maintain separation of concerns by handling business logic
// while delegating data access to repository interfaces.
package logs_services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"

	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// ProjectService handles project management operations
type ProjectService struct {
	repo ProjectRepository
}

// ProjectRepository defines the interface for project data access
type ProjectRepository interface {
	Create(ctx context.Context, project *logs_models.Project) (*logs_models.Project, error)
	GetByID(ctx context.Context, id int, userID int) (*logs_models.Project, error)
	GetByIDGlobal(ctx context.Context, id int) (*logs_models.Project, error) // Without user constraint
	GetBySlug(ctx context.Context, slug string, userID int) (*logs_models.Project, error)
	GetBySlugGlobal(ctx context.Context, slug string) (*logs_models.Project, error) // For batch API validation
	FindByAPIToken(ctx context.Context, token string) (*logs_models.Project, error)
	ListByUserID(ctx context.Context, userID int) ([]logs_models.Project, error)
	Update(ctx context.Context, project *logs_models.Project) error
	UpdateAPIToken(ctx context.Context, projectID int, newAPIToken string) error
	Delete(ctx context.Context, id int) error
}

// NewProjectService creates a new project service
func NewProjectService(repo ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

// GenerateAPIKey generates a new API key and returns both the plain key and bcrypt hash
// Format: dsk_<32 random bytes base64url encoded> = dsk_abc123xyz... (47 chars total)
func GenerateAPIKey() (plainKey, hash string, err error) {
	// Generate 32 random bytes
	randomBytes := make([]byte, 32)
	if _, readErr := rand.Read(randomBytes); readErr != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", readErr)
	}

	// Encode as base64url (URL-safe, no padding)
	encoded := base64.RawURLEncoding.EncodeToString(randomBytes)

	// Format: dsk_ prefix + encoded bytes
	plainKey = "dsk_" + encoded

	// Hash the key with bcrypt (cost 10 = ~100ms)
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(plainKey), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash API key: %w", err)
	}

	return plainKey, string(hashBytes), nil
}

// ValidateAPIKey checks if the provided API key matches the stored hash
func ValidateAPIKey(providedKey, storedHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(providedKey))
	return err == nil
}

// ValidateSlug checks if a slug is valid (lowercase alphanumeric + hyphens, 3-100 chars)
func ValidateSlug(slug string) error {
	if len(slug) < 3 || len(slug) > 100 {
		return fmt.Errorf("slug must be between 3 and 100 characters")
	}

	// Must start and end with alphanumeric, can contain hyphens in middle
	validSlug := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
	if !validSlug.MatchString(slug) {
		return fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens (cannot start/end with hyphen)")
	}

	// Disallow consecutive hyphens
	if strings.Contains(slug, "--") {
		return fmt.Errorf("slug cannot contain consecutive hyphens")
	}

	// Reserved slugs
	reserved := []string{"devsmith-platform", "admin", "api", "health", "logs", "analytics"}
	for _, r := range reserved {
		if slug == r {
			return fmt.Errorf("slug '%s' is reserved", r)
		}
	}

	return nil
}

// ValidateAPIKeyForSlug validates an API key for a given project slug.
// This is used by the batch ingestion endpoint to authenticate external requests.
// Returns the project if validation succeeds, error if project not found or key invalid.
func (s *ProjectService) ValidateAPIKeyForSlug(ctx context.Context, slug, apiKey string) (*logs_models.Project, error) {
	// Validate API key format
	if !strings.HasPrefix(apiKey, "dsk_") {
		return nil, fmt.Errorf("invalid API key format")
	}

	// Get project by slug (global lookup, no userID constraint)
	project, err := s.repo.GetBySlugGlobal(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("project not found or inactive: %w", err)
	}

	// Check if project is active
	if !project.IsActive {
		return nil, fmt.Errorf("project is inactive")
	}

	// Validate API key against stored hash using bcrypt
	if !ValidateAPIKey(apiKey, project.APIKeyHash) {
		return nil, fmt.Errorf("invalid API key")
	}

	return project, nil
}

// CreateProject creates a new project with a generated API key
func (s *ProjectService) CreateProject(ctx context.Context, userID int, req *logs_models.CreateProjectRequest) (*logs_models.CreateProjectResponse, error) {
	// Validate slug format
	if err := ValidateSlug(req.Slug); err != nil {
		return nil, fmt.Errorf("invalid slug: %w", err)
	}

	// Generate API key
	plainKey, hash, err := GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Create project model
	project := &logs_models.Project{
		UserID:        &userID, // Convert int to *int
		Name:          req.Name,
		Slug:          req.Slug,
		Description:   req.Description,
		RepositoryURL: req.RepositoryURL,
		APIKeyHash:    hash,
		IsActive:      true,
	}

	// Save to database
	createdProject, err := s.repo.Create(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Return response with plain API key (ONLY TIME IT'S SHOWN!)
	return &logs_models.CreateProjectResponse{
		Project: *createdProject,
		APIKey:  plainKey,
		Message: "Project created successfully. Save your API key - it will not be shown again!",
	}, nil
}

// GetProjectByAPIKey validates an API key and returns the project
func (s *ProjectService) GetProjectByAPIKey(ctx context.Context, apiKey string) (*logs_models.Project, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if !strings.HasPrefix(apiKey, "dsk_") {
		return nil, fmt.Errorf("invalid API key format")
	}

	// Try to find project by hashing the key
	// Note: We can't hash first then lookup - bcrypt generates different hashes each time
	// So we need to iterate through active projects and compare (inefficient, but secure)
	// Better approach: Cache API key hashes in Redis for fast lookup

	// For now, we'll query by a prefix or use a different approach
	// This is a simplified version - in production, use Redis cache
	return nil, fmt.Errorf("API key validation not implemented - TODO: add Redis cache")
}

// ListProjects returns all projects for a user with statistics
func (s *ProjectService) ListProjects(ctx context.Context, userID int) ([]logs_models.Project, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// GetProject returns a project by ID with statistics
func (s *ProjectService) GetProject(ctx context.Context, projectID int) (*logs_models.Project, error) {
	// TODO: Add log count statistics to Project model
	return s.repo.GetByIDGlobal(ctx, projectID)
}

// UpdateProject updates a project's metadata
func (s *ProjectService) UpdateProject(ctx context.Context, projectID int, req *logs_models.UpdateProjectRequest) (*logs_models.Project, error) {
	// Get existing project
	project, err := s.repo.GetByIDGlobal(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.RepositoryURL != nil {
		project.RepositoryURL = *req.RepositoryURL
	}
	if req.IsActive != nil {
		project.IsActive = *req.IsActive
	}

	// Save changes
	if err := s.repo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

// RegenerateAPIKey generates a new API key for a project
func (s *ProjectService) RegenerateAPIKey(ctx context.Context, projectID int) (*logs_models.RegenerateKeyResponse, error) {
	// Verify project exists
	_, err := s.repo.GetByIDGlobal(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Generate new API token (plain, no hashing)
	plainKey, _, err := GenerateAPIKey() // Still use same generation, just don't use hash
	if err != nil {
		return nil, fmt.Errorf("failed to generate API token: %w", err)
	}

	// Update project with new token (plain, not hashed)
	if err := s.repo.UpdateAPIToken(ctx, projectID, plainKey); err != nil {
		return nil, fmt.Errorf("failed to update API token: %w", err)
	}

	return &logs_models.RegenerateKeyResponse{
		APIKey:  plainKey,
		Message: "API key regenerated successfully. Update your applications with the new key.",
	}, nil
}

// DeactivateProject soft-deletes a project
func (s *ProjectService) DeactivateProject(ctx context.Context, projectID int) error {
	project, err := s.repo.GetByIDGlobal(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	project.IsActive = false
	if err := s.repo.Update(ctx, project); err != nil {
		return fmt.Errorf("failed to deactivate project: %w", err)
	}

	return nil
}
