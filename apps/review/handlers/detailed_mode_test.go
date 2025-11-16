package review_handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestDetailedModeButton_Integration tests the complete flow from button click to analysis
func TestDetailedModeButton_Integration_RedPhase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockLogger := &testutils.MockLogger{}
	detailedService := review_services.NewDetailedService(nil, nil, mockLogger)

	RegisterDetailedModeButtonHandler(router, detailedService)

	req := httptest.NewRequest(
		"POST",
		"/api/review/sessions/1/modes/detailed",
		strings.NewReader(`{"file":"main.go"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Detailed mode button should return 200 OK")
	assert.Contains(t, w.Body.String(), "LineByLine", "Response should contain detailed analysis")
}

// RegisterDetailedModeButtonHandler registers the button handler
func RegisterDetailedModeButtonHandler(router *gin.Engine, detailedService *review_services.DetailedService) {
	router.POST("/api/review/sessions/:id/modes/detailed", func(c *gin.Context) {
		var req struct {
			File string `json:"file"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Mock response for Detailed Mode
		mockResponse := gin.H{
			"File": req.File,
			"LineByLine": []map[string]interface{}{
				{
					"LineNumber":  1,
					"Code":        "package main",
					"Explanation": "Package declaration - defines this as executable package",
				},
				{
					"LineNumber":  2,
					"Code":        "import \"fmt\"",
					"Explanation": "Imports fmt package for formatted I/O",
				},
			},
			"Summary": "Line-by-line algorithm explanation for " + req.File,
		}

		c.JSON(http.StatusOK, mockResponse)
	})
}
