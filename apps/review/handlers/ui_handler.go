package handlers

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/a-h/templ"
	"devsmith/apps/review/templates"
	"github.com/google/uuid"
)

// HomeHandler serves the main Review UI (mode selector + repo input)
func HomeHandler(c *gin.Context) {
	component := templates.Home()
	renderTemplate(c, component)
}

// AnalysisResultHandler displays analysis results
func AnalysisResultHandler(c *gin.Context) {
	mode := c.Query("mode")
	repo := c.Query("repo")
	branch := c.Query("branch")
	analysisMarkdown := c.Query("analysis")

	result := templates.AnalysisResult{
		AnalysisID:   generateAnalysisID(),
		Mode:         mode,
		Repository:   repo,
		Branch:       branch,
		AnalysisHTML: analysisMarkdown,
		CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
	}

	component := templates.Analysis(result)
	renderTemplate(c, component)
}

// Helper function to render Templ components
func renderTemplate(c *gin.Context, component templ.Component) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Template rendering failed"})
	}
}

// Generate unique analysis ID
func generateAnalysisID() string {
	return uuid.New().String()
}
