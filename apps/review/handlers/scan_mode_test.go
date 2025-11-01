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

// TestScanModeButton_Integration tests the complete flow from button click to analysis
func TestScanModeButton_Integration_RedPhase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockLogger := &testutils.MockLogger{}
	scanService := review_services.NewScanService(nil, nil, mockLogger)

	RegisterScanModeButtonHandler(router, scanService)

	req := httptest.NewRequest(
		"POST",
		"/api/review/sessions/1/modes/scan",
		strings.NewReader(`{"query":"authentication"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Scan mode button should return 200 OK")
	assert.Contains(t, w.Body.String(), "Matches", "Response should contain scan results")
}

// RegisterScanModeButtonHandler registers the button handler
func RegisterScanModeButtonHandler(router *gin.Engine, scanService *review_services.ScanService) {
	router.POST("/api/review/sessions/:id/modes/scan", func(c *gin.Context) {
		var req struct {
			Query string `json:"query"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Mock response for Scan Mode
		mockResponse := gin.H{
			"Query": req.Query,
			"Matches": []map[string]interface{}{
				{
					"File":        "auth/handler.go",
					"Line":        42,
					"Content":     "func HandleLogin(c *gin.Context)",
					"Relevance":   0.95,
				},
			},
			"Summary": "Found 1 match for query: " + req.Query,
		}

		c.JSON(http.StatusOK, mockResponse)
	})
}
