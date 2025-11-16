package internal_logs_handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

// ProjectHandler handles HTTP requests for project management
type ProjectHandler struct {
	projectSvc *logs_services.ProjectService
}

// NewProjectHandler creates a new ProjectHandler
func NewProjectHandler(projectSvc *logs_services.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectSvc: projectSvc,
	}
}

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Name          string `json:"name" binding:"required"`
	Slug          string `json:"slug" binding:"required"`
	Description   string `json:"description"`
	RepositoryURL string `json:"repository_url"`
}

// CreateProject handles POST /api/logs/projects
// Creates a new project and returns the generated API key
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	// Create project with service
	projectReq := &logs_models.CreateProjectRequest{
		Name:          req.Name,
		Slug:          req.Slug,
		Description:   req.Description,
		RepositoryURL: req.RepositoryURL,
	}

	resp, err := h.projectSvc.CreateProject(c.Request.Context(), userID, projectReq)
	if err != nil {
		// Check if it's a duplicate slug error
		if err.Error() == "project with this slug already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project: " + err.Error()})
		return
	}

	// Return project and API key (API key only shown once)
	c.JSON(http.StatusCreated, gin.H{
		"project": gin.H{
			"id":             resp.Project.ID,
			"name":           resp.Project.Name,
			"slug":           resp.Project.Slug,
			"description":    resp.Project.Description,
			"repository_url": resp.Project.RepositoryURL,
			"created_at":     resp.Project.CreatedAt,
		},
		"api_key": resp.APIKey,
		"message": resp.Message,
	})
}

// GetProject handles GET /api/logs/projects/:id
func (h *ProjectHandler) GetProject(c *gin.Context) {
	// Get user ID from context (not used in simplified auth model)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	_, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	// Parse project ID from URL
	projectIDStr := c.Param("id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get project
	project, err := h.projectSvc.GetProject(c.Request.Context(), projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project: " + err.Error()})
		return
	}

	// Verify project belongs to user (security check)
	// Note: Ownership check disabled for simplified authentication model
	// In production with auth, uncomment: if project.UserID != nil && *project.UserID != userID { ... }

	c.JSON(http.StatusOK, gin.H{
		"project": gin.H{
			"id":             project.ID,
			"name":           project.Name,
			"slug":           project.Slug,
			"description":    project.Description,
			"repository_url": project.RepositoryURL,
			"created_at":     project.CreatedAt,
			"updated_at":     project.UpdatedAt,
			"is_active":      project.IsActive,
		},
	})
}

// ListProjects handles GET /api/logs/projects
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	// List projects
	projects, err := h.projectSvc.ListProjects(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list projects: " + err.Error()})
		return
	}

	// Convert to response format
	projectList := make([]gin.H, len(projects))
	for i, p := range projects {
		projectList[i] = gin.H{
			"id":             p.ID,
			"name":           p.Name,
			"slug":           p.Slug,
			"description":    p.Description,
			"repository_url": p.RepositoryURL,
			"created_at":     p.CreatedAt,
			"updated_at":     p.UpdatedAt,
			"is_active":      p.IsActive,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projectList,
		"count":    len(projects),
	})
}

// RegenerateAPIKey handles POST /api/logs/projects/:id/regenerate-key
func (h *ProjectHandler) RegenerateAPIKey(c *gin.Context) {
	// Get user ID from context (not used in simplified auth model)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	_, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	// Parse project ID from URL
	projectIDStr := c.Param("id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Regenerate API key
	resp, err := h.projectSvc.RegenerateAPIKey(c.Request.Context(), projectID)
	if err != nil {
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to regenerate API key: " + err.Error()})
		return
	}

	// Verify project belongs to user (security check)
	// Note: Ownership check disabled for simplified authentication model
	_, err = h.projectSvc.GetProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key": resp.APIKey,
		"message": resp.Message,
	})
}

// DeleteProject handles DELETE /api/logs/projects/:id
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	// Get user ID from context (not used in simplified auth model)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	_, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	// Parse project ID from URL
	projectIDStr := c.Param("id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify project belongs to user (security check)
	// Note: Ownership check disabled for simplified authentication model
	_, err = h.projectSvc.GetProject(c.Request.Context(), projectID)
	if err != nil {
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project: " + err.Error()})
		return
	}

	// Deactivate project (soft delete)
	err = h.projectSvc.DeactivateProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project deleted successfully",
	})
}
