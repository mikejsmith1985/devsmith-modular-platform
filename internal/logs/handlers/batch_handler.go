// Package internal_logs_handlers provides HTTP handlers for the logs service.
package internal_logs_handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	logs_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

// BatchHandler handles batch log ingestion for cross-repo logging.
type BatchHandler struct {
	logRepo     *logs_db.LogEntryRepository
	projectRepo *logs_db.ProjectRepository
	projectSvc  *logs_services.ProjectService
}

// NewBatchHandler creates a new BatchHandler.
func NewBatchHandler(
	logRepo *logs_db.LogEntryRepository,
	projectRepo *logs_db.ProjectRepository,
	projectSvc *logs_services.ProjectService,
) *BatchHandler {
	return &BatchHandler{
		logRepo:     logRepo,
		projectRepo: projectRepo,
		projectSvc:  projectSvc,
	}
}

// BatchLogEntry represents a single log entry in a batch request.
type BatchLogEntry struct {
	Timestamp   string                 `json:"timestamp"`              // ISO 8601 timestamp
	Level       string                 `json:"level"`                  // debug, info, warn, error
	Message     string                 `json:"message"`                // Log message
	ServiceName string                 `json:"service_name,omitempty"` // Microservice identifier
	Context     map[string]interface{} `json:"context,omitempty"`      // Additional context
}

// BatchLogRequest represents the batch ingestion request payload.
type BatchLogRequest struct {
	ProjectSlug string          `json:"project_slug" binding:"required"` // Project identifier
	Logs        []BatchLogEntry `json:"logs" binding:"required,min=1"`   // Array of log entries
}

// BatchLogResponse represents the batch ingestion response.
type BatchLogResponse struct {
	Accepted int    `json:"accepted"` // Number of logs accepted
	Message  string `json:"message"`
}

// IngestBatch handles POST /api/logs/batch for batch log ingestion.
// This endpoint is designed for internal services to send logs to DevSmith.
//
// Performance: 100 logs in ~50ms (vs 3000ms for individual requests)
//
// Authentication: None (designed for internal service communication)
// Future: Add authentication when needed for external services
func (h *BatchHandler) IngestBatch(c *gin.Context) {
	// Step 1: Parse request body
	var req BatchLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// Validate batch size (max 1000 logs per request)
	if len(req.Logs) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Batch size exceeds maximum of 1000 logs",
		})
		return
	}

	// Step 2: Get or create project by slug
	ctx := c.Request.Context()
	project, err := h.projectRepo.GetBySlugGlobal(ctx, req.ProjectSlug)
	if err != nil {
		// Auto-create project if it doesn't exist (simplified for internal use)
		// UserID is nil for auto-created projects (no authentication required)
		newProject := &logs_models.Project{
			Name:     req.ProjectSlug,
			Slug:     req.ProjectSlug,
			IsActive: true,
			UserID:   nil, // No user ID for auto-created projects
		}
		project, err = h.projectRepo.Create(ctx, newProject)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create project",
			})
			return
		}
	}

	// Check if project is active
	if !project.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Project is inactive",
		})
		return
	}

	// Step 5: Verify project slug matches (additional validation)
	if project.Slug != req.ProjectSlug {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Project slug mismatch. API key belongs to project '%s', not '%s'", project.Slug, req.ProjectSlug),
		})
		return
	}

	// Step 6: Convert batch entries to LogEntry models
	entries := make([]*logs_models.LogEntry, 0, len(req.Logs))
	projectID := int64(project.ID)

	for i, logEntry := range req.Logs {
		// Parse timestamp
		timestamp, err := time.Parse(time.RFC3339, logEntry.Timestamp)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Invalid timestamp format at index %d: %v", i, err),
			})
			return
		}

		// Validate level
		level := strings.ToUpper(logEntry.Level)
		validLevels := map[string]bool{
			"DEBUG": true,
			"INFO":  true,
			"WARN":  true,
			"ERROR": true,
		}
		if !validLevels[level] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Invalid log level '%s' at index %d. Must be: debug, info, warn, error", logEntry.Level, i),
			})
			return
		}

		// Convert context map to JSON bytes
		var metadataBytes []byte
		if logEntry.Context != nil {
			metadataBytes, err = json.Marshal(logEntry.Context)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Invalid context at index %d: %v", i, err),
				})
				return
			}
		} else {
			metadataBytes = []byte("{}")
		}

		// Create LogEntry model
		entry := &logs_models.LogEntry{
			ProjectID:   &projectID,
			Service:     "external", // Mark as external log source
			ServiceName: logEntry.ServiceName,
			Level:       level,
			Message:     logEntry.Message,
			Metadata:    metadataBytes,
			Tags:        []string{}, // Empty tags for now
			Timestamp:   timestamp,
		}

		entries = append(entries, entry)
	}

	// Step 7: Insert batch using optimized CreateBatch method
	if err := h.logRepo.CreateBatch(ctx, entries); err != nil {
		fmt.Printf("ERROR: Failed to insert batch logs - project_id=%d, entry_count=%d, error=%v\n", project.ID, len(entries), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to insert logs: %v", err),
		})
		return
	}

	// Step 8: Return success response
	c.JSON(http.StatusCreated, BatchLogResponse{
		Accepted: len(entries),
		Message:  fmt.Sprintf("Successfully ingested %d log entries", len(entries)),
	})
}
