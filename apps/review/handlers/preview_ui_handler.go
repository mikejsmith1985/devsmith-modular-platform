package review_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
)

// RegisterPreviewUIRoutes registers the UI routes for preview mode in the review app.
func RegisterPreviewUIRoutes(router *gin.Engine, previewService *review_services.PreviewService) {
	router.GET("/review/preview", func(c *gin.Context) {
		// For demo, use mock session ID and codebase with default modes
		result, err := previewService.AnalyzePreview(c.Request.Context(), "testdata/sample_project", "intermediate", "quick")
		if err != nil {
			c.String(http.StatusInternalServerError, "Analysis failed")
			return
		}
		c.HTML(http.StatusOK, "layout", gin.H{
			"Title":                "Preview Mode",
			"FileTree":             result.FileTree,
			"BoundedContexts":      result.BoundedContexts,
			"TechStack":            result.TechStack,
			"ArchitecturePattern":  result.ArchitectureStyle,
			"EntryPoints":          result.EntryPoints,
			"ExternalDependencies": result.ExternalDeps,
			"Summary":              result.Summary,
			"SessionID":            "demo-session",
		})
	})
}
