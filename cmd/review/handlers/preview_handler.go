// Package handlers contains HTTP handlers for the review service.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
)

// RegisterPreviewRoutes registers the API routes for preview mode analysis in the review service.
func RegisterPreviewRoutes(router *gin.Engine, previewService *services.PreviewService) {
	router.POST("/api/review/sessions/:id/analyze", func(c *gin.Context) {
		var req struct {
			ReadingMode string `json:"reading_mode"`
			TargetPath  string `json:"target_path"`
			ScanQuery   string `json:"scan_query"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if req.ReadingMode != "preview" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported reading mode"})
			return
		}

		// For now, use a mock codebase path
		result, err := previewService.AnalyzePreview(c.Request.Context(), "testdata/sample_project")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Analysis failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"analysis": result, "cached": false})
	})
}
