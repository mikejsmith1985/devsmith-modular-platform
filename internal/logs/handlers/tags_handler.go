package internal_logs_handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	logs_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
)

// TagsHandler handles tag-related operations
type TagsHandler struct {
	repo *logs_db.LogRepository
}

// NewTagsHandler creates a new tags handler
func NewTagsHandler(repo *logs_db.LogRepository) *TagsHandler {
	return &TagsHandler{
		repo: repo,
	}
}

// GetAvailableTags returns all unique tags from the database
// GET /api/logs/tags
func (h *TagsHandler) GetAvailableTags(c *gin.Context) {
	tags, err := h.repo.GetAllTags(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch tags",
		})
		return
	}

	// Return tags with counts
	tagCounts := make(map[string]int)
	for _, tag := range tags {
		tagCounts[tag]++
	}

	c.JSON(http.StatusOK, gin.H{
		"tags":   tags,
		"counts": tagCounts,
	})
}

// AddTagToLog adds a manual tag to a log entry
// POST /api/logs/:id/tags
func (h *TagsHandler) AddTagToLog(c *gin.Context) {
	logIDStr := c.Param("id")
	logID, err := strconv.ParseInt(logIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid log ID",
		})
		return
	}

	var req struct {
		Tag string `json:"tag" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tag is required",
		})
		return
	}

	// Add tag to log
	if err := h.repo.AddTag(c.Request.Context(), logID, req.Tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add tag",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "tag_added",
		"tag":    req.Tag,
	})
}

// RemoveTagFromLog removes a tag from a log entry
// DELETE /api/logs/:id/tags/:tag
func (h *TagsHandler) RemoveTagFromLog(c *gin.Context) {
	logIDStr := c.Param("id")
	logID, err := strconv.ParseInt(logIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid log ID",
		})
		return
	}

	tag := c.Param("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tag parameter is required",
		})
		return
	}

	// Remove tag from log
	if err := h.repo.RemoveTag(c.Request.Context(), logID, tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove tag",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "tag_removed",
		"tag":    tag,
	})
}
