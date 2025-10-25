package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HomeHandler serves the main Review UI (mode selector + repo input)
func HomeHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "home.html", nil)
}

// AnalysisResultHandler displays analysis results
func AnalysisResultHandler(c *gin.Context) {
	mode := c.Query("mode")
	repo := c.Query("repo")
	branch := c.Query("branch")
	analysisMarkdown := c.Query("analysis")

	data := map[string]interface{}{
		"AnalysisID":   generateAnalysisID(),
		"Mode":         mode,
		"Repository":   repo,
		"Branch":       branch,
		"AnalysisHTML": analysisMarkdown,
		"CreatedAt":    time.Now().Format("2006-01-02 15:04:05"),
	}

	c.HTML(http.StatusOK, "analysis.html", data)
}

// Generate unique analysis ID
func generateAnalysisID() string {
	return uuid.New().String()
}

// GenerateAnalysisID is the exported version for testing
func GenerateAnalysisID() string {
	return generateAnalysisID()
}
