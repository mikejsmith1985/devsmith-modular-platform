package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	review_handlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/review/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logging"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/testutils"
)

// TestReviewAPIPayloadValidation tests that API endpoints validate payloads correctly
// This test would have caught the session_id field mismatch issue
func TestReviewAPIPayloadValidation(t *testing.T) {
	// Setup test router with Review handlers
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create mock services
	mockLogger := &testutils.MockLogger{}
	mockLogClient := &logging.Client{} // Simplified for testing

	// Mock analyzer services (nil for this test - we're testing validation, not analysis)
	handler := review_handlers.NewUIHandler(
		mockLogger,
		mockLogClient,
		nil, // previewService
		nil, // skimService
		nil, // scanService
		nil, // detailedService
		nil, // criticalService
		nil, // modelService
	)

	// Register API routes
	router.POST("/api/review/modes/preview", handler.HandlePreviewMode)
	router.POST("/api/review/modes/skim", handler.HandleSkimMode)
	router.POST("/api/review/modes/scan", handler.HandleScanMode)
	router.POST("/api/review/modes/detailed", handler.HandleDetailedMode)
	router.POST("/api/review/modes/critical", handler.HandleCriticalMode)

	testCases := []struct {
		name           string
		endpoint       string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  string
		description    string
	}{
		{
			name:     "Valid payload with required fields",
			endpoint: "/api/review/modes/preview",
			payload: map[string]interface{}{
				"pasted_code": "package main\n\nfunc main() {\n    fmt.Println(\"Hello\")\n}",
				"model":       "mistral:7b-instruct",
			},
			expectedStatus: 503, // Service unavailable due to nil analyzer
			description:    "Should accept correct payload structure",
		},
		{
			name:     "Missing required pasted_code field",
			endpoint: "/api/review/modes/preview",
			payload: map[string]interface{}{
				"model": "mistral:7b-instruct",
			},
			expectedStatus: 400,
			expectedError:  "required",
			description:    "Should reject payload missing required field",
		},
		{
			name:     "Empty pasted_code field",
			endpoint: "/api/review/modes/preview",
			payload: map[string]interface{}{
				"pasted_code": "",
				"model":       "mistral:7b-instruct",
			},
			expectedStatus: 400,
			expectedError:  "required",
			description:    "Should reject empty pasted_code",
		},
		{
			name:     "Wrong field name - should use pasted_code not code",
			endpoint: "/api/review/modes/preview",
			payload: map[string]interface{}{
				"code":  "package main\nfunc main() {}", // WRONG field name
				"model": "mistral:7b-instruct",
			},
			expectedStatus: 400,
			expectedError:  "required",
			description:    "Should reject 'code' field name - must use 'pasted_code'",
		},
		{
			name:     "Extra session_id field should be ignored",
			endpoint: "/api/review/modes/preview",
			payload: map[string]interface{}{
				"session_id":  "test123", // Extra field that should be ignored
				"pasted_code": "package main\nfunc main() {}",
				"model":       "mistral:7b-instruct",
			},
			expectedStatus: 503, // Should work despite extra field
			description:    "Should ignore extra session_id field",
		},
		{
			name:     "Model field is optional",
			endpoint: "/api/review/modes/preview",
			payload: map[string]interface{}{
				"pasted_code": "package main\nfunc main() {}",
				// No model field
			},
			expectedStatus: 503, // Should work without model
			description:    "Should accept payload without model field",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			payloadBytes, err := json.Marshal(tc.payload)
			require.NoError(t, err, "Failed to marshal payload")

			req := httptest.NewRequest("POST", tc.endpoint, bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, w.Code,
				"Expected status %d but got %d for case: %s\nResponse body: %s",
				tc.expectedStatus, w.Code, tc.description, w.Body.String())

			// Check error message if expected
			if tc.expectedError != "" {
				responseBody := w.Body.String()
				assert.Contains(t, strings.ToLower(responseBody), strings.ToLower(tc.expectedError),
					"Expected error message containing '%s' but got: %s", tc.expectedError, responseBody)
			}

			// Log the test case for debugging
			t.Logf("✓ %s: %s (Status: %d)", tc.name, tc.description, w.Code)
		})
	}
}

// TestReviewAPIContentTypeHandling tests different content types
func TestReviewAPIContentTypeHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockLogger := &testutils.MockLogger{}
	mockLogClient := &logging.Client{}

	handler := review_handlers.NewUIHandler(mockLogger, mockLogClient, nil, nil, nil, nil, nil, nil)
	router.POST("/api/review/modes/preview", handler.HandlePreviewMode)

	testCases := []struct {
		name           string
		contentType    string
		body           string
		expectedStatus int
		description    string
	}{
		{
			name:           "Valid JSON content type",
			contentType:    "application/json",
			body:           `{"pasted_code":"package main\nfunc main() {}","model":"mistral:7b-instruct"}`,
			expectedStatus: 503, // Should work
			description:    "Should accept application/json",
		},
		{
			name:           "Missing content type",
			contentType:    "",
			body:           `{"pasted_code":"package main\nfunc main() {}","model":"mistral:7b-instruct"}`,
			expectedStatus: 400, // Should fail validation
			description:    "Should reject missing content type",
		},
		{
			name:           "Wrong content type",
			contentType:    "text/plain",
			body:           `{"pasted_code":"package main\nfunc main() {}","model":"mistral:7b-instruct"}`,
			expectedStatus: 400, // Should fail validation
			description:    "Should reject wrong content type",
		},
		{
			name:           "Invalid JSON",
			contentType:    "application/json",
			body:           `{"pasted_code":"package main\nfunc main() {}"`, // Missing closing brace
			expectedStatus: 400,
			description:    "Should reject malformed JSON",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/review/modes/preview", strings.NewReader(tc.body))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code,
				"Expected status %d but got %d for case: %s\nResponse: %s",
				tc.expectedStatus, w.Code, tc.description, w.Body.String())

			t.Logf("✓ %s: %s (Status: %d)", tc.name, tc.description, w.Code)
		})
	}
}

// TestReviewAPIErrorMessages tests that error messages are helpful
func TestReviewAPIErrorMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockLogger := &testutils.MockLogger{}
	mockLogClient := &logging.Client{}

	handler := review_handlers.NewUIHandler(mockLogger, mockLogClient, nil, nil, nil, nil, nil, nil)
	router.POST("/api/review/modes/preview", handler.HandlePreviewMode)

	// Test that missing required field gives helpful error
	req := httptest.NewRequest("POST", "/api/review/modes/preview", strings.NewReader(`{"model":"mistral:7b-instruct"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)

	responseBody := w.Body.String()

	// Should mention the missing field
	assert.Contains(t, strings.ToLower(responseBody), "required",
		"Error message should mention 'required' field. Got: %s", responseBody)

	t.Logf("✓ Error message for missing field: %s", responseBody)
}

// TestFrontendBackendAPIContract ensures frontend API calls match backend expectations
func TestFrontendBackendAPIContract(t *testing.T) {
	t.Log("Testing Frontend-Backend API Contract")

	// This test simulates the exact payloads that our frontend sends
	// to ensure they match what the backend expects

	frontendPayloads := []struct {
		mode    string
		payload map[string]interface{}
	}{
		{
			mode: "preview",
			payload: map[string]interface{}{
				"pasted_code": "package main\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}",
				"model":       "mistral:7b-instruct",
			},
		},
		{
			mode: "skim",
			payload: map[string]interface{}{
				"pasted_code": "package main\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}",
				"model":       "mistral:7b-instruct",
			},
		},
		{
			mode: "scan",
			payload: map[string]interface{}{
				"pasted_code": "package main\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}",
				"model":       "mistral:7b-instruct",
				"query":       "find main function",
			},
		},
		{
			mode: "detailed",
			payload: map[string]interface{}{
				"pasted_code": "package main\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}",
				"model":       "mistral:7b-instruct",
			},
		},
		{
			mode: "critical",
			payload: map[string]interface{}{
				"pasted_code": "package main\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}",
				"model":       "mistral:7b-instruct",
			},
		},
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockLogger := &testutils.MockLogger{}
	mockLogClient := &logging.Client{}

	handler := review_handlers.NewUIHandler(mockLogger, mockLogClient, nil, nil, nil, nil, nil, nil)

	// Register all mode endpoints
	router.POST("/api/review/modes/preview", handler.HandlePreviewMode)
	router.POST("/api/review/modes/skim", handler.HandleSkimMode)
	router.POST("/api/review/modes/scan", handler.HandleScanMode)
	router.POST("/api/review/modes/detailed", handler.HandleDetailedMode)
	router.POST("/api/review/modes/critical", handler.HandleCriticalMode)

	for _, test := range frontendPayloads {
		t.Run(fmt.Sprintf("Frontend_%s_payload", test.mode), func(t *testing.T) {
			payloadBytes, err := json.Marshal(test.payload)
			require.NoError(t, err)

			endpoint := fmt.Sprintf("/api/review/modes/%s", test.mode)
			req := httptest.NewRequest("POST", endpoint, bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should not be 400 (bad request) - payload structure should be valid
			assert.NotEqual(t, 400, w.Code,
				"Frontend payload for %s mode should not return 400 Bad Request. "+
					"Got status %d with response: %s",
				test.mode, w.Code, w.Body.String())

			// Should be 503 (service unavailable) due to nil analyzer services, which means payload validation passed
			assert.Equal(t, 503, w.Code,
				"Expected 503 Service Unavailable (nil analyzer) for %s mode, got %d: %s",
				test.mode, w.Code, w.Body.String())

			t.Logf("✓ Frontend %s payload validated successfully", test.mode)
		})
	}
}
