package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
)

func RegisterPreviewUIRoutes(router *gin.Engine, previewService *services.PreviewService) {
	router.GET("/review/preview", func(c *gin.Context) {
		// For demo, use mock session ID and codebase
		result, err := previewService.AnalyzePreview(c.Request.Context(), "testdata/sample_project")
		if err != nil {
			c.String(http.StatusInternalServerError, "Analysis failed")
			return
		}
		c.HTML(http.StatusOK, "layout", gin.H{
			"Title":                "Preview Mode",
			"FileTree":             result.FileTree,
			"BoundedContexts":      result.BoundedContexts,
			"TechStack":            result.TechStack,
			"ArchitecturePattern":  result.ArchitecturePattern,
			"EntryPoints":          result.EntryPoints,
			"ExternalDependencies": result.ExternalDependencies,
			"Summary":              result.Summary,
			"SessionID":            "demo-session",
		})
	})
}
