package logs_models

import "time"

// Project represents an external application/repository that sends logs to DevSmith
type Project struct {
	ID            int       `json:"id" db:"id"`
	UserID        int       `json:"user_id" db:"user_id"`
	Name          string    `json:"name" db:"name"`
	Slug          string    `json:"slug" db:"slug"`
	Description   string    `json:"description" db:"description"`
	RepositoryURL string    `json:"repository_url" db:"repository_url"`
	APIKeyHash    string    `json:"-" db:"api_key_hash"` // Never expose in JSON
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	IsActive      bool      `json:"is_active" db:"is_active"`

	// Computed fields (from joins/aggregations)
	LogCount     int        `json:"log_count,omitempty" db:"total_logs"`
	ErrorCount   int        `json:"error_count,omitempty" db:"error_count"`
	WarnCount    int        `json:"warn_count,omitempty" db:"warn_count"`
	InfoCount    int        `json:"info_count,omitempty" db:"info_count"`
	DebugCount   int        `json:"debug_count,omitempty" db:"debug_count"`
	LastLogAt    *time.Time `json:"last_log_at,omitempty" db:"last_log_at"`
	ServiceCount int        `json:"service_count,omitempty" db:"service_count"`
}

// CreateProjectRequest is the request body for creating a new project
type CreateProjectRequest struct {
	Name          string `json:"name" binding:"required,min=1,max=255"`
	Slug          string `json:"slug" binding:"required,min=3,max=100,alphanum_hyphen"`
	Description   string `json:"description" binding:"max=1000"`
	RepositoryURL string `json:"repository_url" binding:"omitempty,url"`
}

// CreateProjectResponse includes the plain API key (shown only once!)
type CreateProjectResponse struct {
	Project
	APIKey  string `json:"api_key"` // Plain key, shown ONLY on creation
	Message string `json:"message"`
}

// UpdateProjectRequest is the request body for updating a project
type UpdateProjectRequest struct {
	Name          *string `json:"name" binding:"omitempty,min=1,max=255"`
	Description   *string `json:"description" binding:"omitempty,max=1000"`
	RepositoryURL *string `json:"repository_url" binding:"omitempty,url"`
	IsActive      *bool   `json:"is_active"`
}

// RegenerateKeyResponse includes the new API key
type RegenerateKeyResponse struct {
	APIKey  string `json:"api_key"`
	Message string `json:"message"`
}
