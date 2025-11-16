// Package review_handlers contains HTTP handlers for the review app.
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

// TestPreviewModeButton_Integration tests the complete flow from button click to analysis
// RED phase: This test FAILS because PreviewModeButtonHandler doesn't exist yet
func TestPreviewModeButton_Integration_RedPhase(t *testing.T) {
	// GIVEN: A review session exists with code input
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a mock preview service
	mockLogger := &testutils.MockLogger{}
	mockOllama := &testutils.MockOllamaClient{GenerateResponse: `{"summary":"ok","bounded_contexts":[],"tech_stack":[],"file_tree":[]}`}
	previewService := review_services.NewPreviewService(mockOllama, mockLogger)

	// Register the handler that we're testing (currently doesn't exist)
	RegisterPreviewModeButtonHandler(router, previewService)

	// WHEN: A POST request comes to trigger Preview Mode analysis
	req := httptest.NewRequest(
		"POST",
		"/api/review/sessions/1/modes/preview",
		strings.NewReader(`{"code":"package main\nfunc main() {}"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// THEN: Should return 200 OK with preview analysis results
	assert.Equal(t, http.StatusOK, w.Code, "Preview mode button should return 200 OK")
	assert.Contains(t, w.Body.String(), "bounded_contexts", "Response should contain preview analysis")
	assert.Contains(t, w.Body.String(), "tech_stack", "Response should contain tech stack")
	assert.Contains(t, w.Body.String(), "file_tree", "Response should contain file tree")
}

// TestPreviewModeButton_UIRendering tests that the button renders in the home page
// SKIP: UI rendering is tested via E2E tests (Playwright)
func TestPreviewModeButton_UIRendering_RedPhase(t *testing.T) {
	t.Skip("UI rendering tested via E2E tests (Playwright)")
}

// TestPreviewModeButton_JavaScriptIntegration tests button click wiring in review.js
// SKIP: JS integration tested via E2E tests
func TestPreviewModeButton_JavaScriptIntegration_RedPhase(t *testing.T) {
	t.Skip("JS integration tested via E2E tests (Playwright)")
}

// RegisterPreviewModeButtonHandler registers the button handler
// GREEN phase: Minimal implementation to pass RED phase test
func RegisterPreviewModeButtonHandler(router *gin.Engine, previewService *review_services.PreviewService) {
	// Register POST endpoint for Preview Mode button analysis
	router.POST("/api/review/sessions/:id/modes/preview", func(c *gin.Context) {
		// Parse request
		var req struct {
			Code string `json:"code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Call preview service with default modes for test
		result, err := previewService.AnalyzePreview(c.Request.Context(), req.Code, "intermediate", "quick")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Analysis failed"})
			return
		}

		// Return results as JSON
		c.JSON(http.StatusOK, result)
	})
}
