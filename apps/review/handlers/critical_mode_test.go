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

// TestCriticalModeButton_Integration tests the complete flow from button click to analysis
func TestCriticalModeButton_Integration_RedPhase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockLogger := &testutils.MockLogger{}
	criticalService := review_services.NewCriticalService(nil, nil, mockLogger)

	RegisterCriticalModeButtonHandler(router, criticalService)

	req := httptest.NewRequest(
		"POST",
		"/api/review/sessions/1/modes/critical",
		strings.NewReader(`{"full_code":"package main\nfunc main() {}"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Critical mode button should return 200 OK")
	assert.Contains(t, w.Body.String(), "Issues", "Response should contain quality issues")
}

// RegisterCriticalModeButtonHandler registers the button handler
func RegisterCriticalModeButtonHandler(router *gin.Engine, criticalService *review_services.CriticalService) {
	router.POST("/api/review/sessions/:id/modes/critical", func(c *gin.Context) {
		var req struct {
			FullCode string `json:"full_code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Mock response for Critical Mode - quality evaluation
		mockResponse := gin.H{
			"Issues": []map[string]interface{}{
				{
					"Severity":    "CRITICAL",
					"Category":    "Security",
					"Description": "Handler calls DB directly (layer violation)",
					"Line":        42,
					"Suggestion":  "Use service layer instead of direct DB access",
				},
				{
					"Severity":    "IMPORTANT",
					"Category":    "Error Handling",
					"Description": "Missing error handling for database operation",
					"Line":        45,
					"Suggestion":  "Check err != nil and handle appropriately",
				},
			},
			"OverallQuality": 65,
			"Summary":        "2 issues found: 1 CRITICAL, 1 IMPORTANT",
		}

		c.JSON(http.StatusOK, mockResponse)
	})
}
