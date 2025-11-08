package review_handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestHandler creates a minimal UIHandler for testing bindCodeRequest
func createTestHandler(t *testing.T) *UIHandler {
	t.Helper()

	// Create minimal logger for testing
	testLogger, err := logger.NewLogger(&logger.Config{
		ServiceName: "review-test",
		LogLevel:    "info",
	})
	require.NoError(t, err)

	return &UIHandler{
		logger: testLogger,
		// Other fields can be nil for bindCodeRequest tests
		// as it only uses logger and parses request
	}
}

// TestBindCodeRequest_JSONWithModes tests mode extraction from JSON request body
func TestBindCodeRequest_JSONWithModes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		jsonBody           string
		expectedUserMode   string
		expectedOutputMode string
	}{
		{
			name:               "Both modes provided",
			jsonBody:           `{"pasted_code": "test", "user_mode": "beginner", "output_mode": "full"}`,
			expectedUserMode:   "beginner",
			expectedOutputMode: "full",
		},
		{
			name:               "Only user_mode provided (output defaults to quick)",
			jsonBody:           `{"pasted_code": "test", "user_mode": "expert"}`,
			expectedUserMode:   "expert",
			expectedOutputMode: "quick",
		},
		{
			name:               "Only output_mode provided (user defaults to intermediate)",
			jsonBody:           `{"pasted_code": "test", "output_mode": "full"}`,
			expectedUserMode:   "intermediate",
			expectedOutputMode: "full",
		},
		{
			name:               "No modes provided (both default)",
			jsonBody:           `{"pasted_code": "test"}`,
			expectedUserMode:   "intermediate",
			expectedOutputMode: "quick",
		},
		{
			name:               "Novice + Full",
			jsonBody:           `{"pasted_code": "test", "user_mode": "novice", "output_mode": "full"}`,
			expectedUserMode:   "novice",
			expectedOutputMode: "full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("POST", "/test", bytes.NewBufferString(tt.jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			// Create handler with proper initialization
			handler := createTestHandler(t)
			req, ok := handler.bindCodeRequest(c)

			// bindCodeRequest returns false and sets error response if binding fails
			// For valid JSON, it should return true
			require.True(t, ok, "bindCodeRequest should succeed for valid JSON")
			require.NotNil(t, req, "CodeRequest should not be nil")

			// Verify modes
			assert.Equal(t, tt.expectedUserMode, req.UserMode, "UserMode mismatch")
			assert.Equal(t, tt.expectedOutputMode, req.OutputMode, "OutputMode mismatch")
		})
	}
}

// TestBindCodeRequest_FormDataWithModes tests mode extraction from form/multipart data
func TestBindCodeRequest_FormDataWithModes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		formFields         map[string]string
		expectedUserMode   string
		expectedOutputMode string
	}{
		{
			name: "Form with both modes",
			formFields: map[string]string{
				"pasted_code": "test code",
				"user_mode":   "beginner",
				"output_mode": "full",
			},
			expectedUserMode:   "beginner",
			expectedOutputMode: "full",
		},
		{
			name: "Form with only user_mode",
			formFields: map[string]string{
				"pasted_code": "test code",
				"user_mode":   "expert",
			},
			expectedUserMode:   "expert",
			expectedOutputMode: "quick",
		},
		{
			name: "Form with no modes (defaults)",
			formFields: map[string]string{
				"pasted_code": "test code",
			},
			expectedUserMode:   "intermediate",
			expectedOutputMode: "quick",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create multipart form request
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			for key, val := range tt.formFields {
				err := writer.WriteField(key, val)
				require.NoError(t, err)
			}
			writer.Close()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("POST", "/test", body)
			c.Request.Header.Set("Content-Type", writer.FormDataContentType())

			handler := createTestHandler(t)
			req, ok := handler.bindCodeRequest(c)

			require.True(t, ok, "bindCodeRequest should succeed for valid form")
			require.NotNil(t, req)

			assert.Equal(t, tt.expectedUserMode, req.UserMode)
			assert.Equal(t, tt.expectedOutputMode, req.OutputMode)
			assert.Equal(t, tt.formFields["pasted_code"], req.PastedCode)
		})
	}
}

// TestBindCodeRequest_FileUploadWithModes tests mode extraction from file upload
func TestBindCodeRequest_FileUploadWithModes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		userMode           string
		outputMode         string
		expectedUserMode   string
		expectedOutputMode string
	}{
		{
			name:               "File upload with explicit modes",
			userMode:           "expert",
			outputMode:         "full",
			expectedUserMode:   "expert",
			expectedOutputMode: "full",
		},
		{
			name:               "File upload with empty modes (defaults)",
			userMode:           "",
			outputMode:         "",
			expectedUserMode:   "intermediate",
			expectedOutputMode: "quick",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create multipart form with file
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// Add file field - use "pasted_code" as the field name (matches handler logic)
			fileWriter, err := writer.CreateFormFile("pasted_code", "test.go")
			require.NoError(t, err)
			fileWriter.Write([]byte("package main\n\nfunc main() {}"))

			// Add mode fields
			if tt.userMode != "" {
				writer.WriteField("user_mode", tt.userMode)
			}
			if tt.outputMode != "" {
				writer.WriteField("output_mode", tt.outputMode)
			}

			writer.Close()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("POST", "/test", body)
			c.Request.Header.Set("Content-Type", writer.FormDataContentType())

			handler := createTestHandler(t)
			req, ok := handler.bindCodeRequest(c)

			require.True(t, ok, "bindCodeRequest should succeed for file upload")
			require.NotNil(t, req)

			assert.Equal(t, tt.expectedUserMode, req.UserMode)
			assert.Equal(t, tt.expectedOutputMode, req.OutputMode)
			assert.NotEmpty(t, req.PastedCode, "File content should be extracted")
		})
	}
}

// TestBindCodeRequest_InvalidJSON tests error handling
func TestBindCodeRequest_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		jsonBody string
	}{
		{
			name:     "Malformed JSON",
			jsonBody: `{"pasted_code": "test"`,
		},
		{
			name:     "Empty request body",
			jsonBody: ``,
		},
		{
			name:     "Missing required field",
			jsonBody: `{"user_mode": "beginner"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("POST", "/test", bytes.NewBufferString(tt.jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler := createTestHandler(t)
			req, ok := handler.bindCodeRequest(c)

			// Should return false for invalid requests
			assert.False(t, ok, "bindCodeRequest should fail for invalid JSON")
			assert.Nil(t, req, "CodeRequest should be nil for failed binding")

			// Should have set error response
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// TestBindCodeRequest_AllExperienceLevels validates all supported experience levels
func TestBindCodeRequest_AllExperienceLevels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	levels := []string{"beginner", "novice", "intermediate", "expert"}

	for _, level := range levels {
		t.Run("ExperienceLevel_"+level, func(t *testing.T) {
			jsonBody := `{"pasted_code": "test", "user_mode": "` + level + `"}`

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("POST", "/test", bytes.NewBufferString(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler := createTestHandler(t)
			req, ok := handler.bindCodeRequest(c)

			require.True(t, ok)
			require.NotNil(t, req)
			assert.Equal(t, level, req.UserMode)
		})
	}
}

// TestBindCodeRequest_AllOutputModes validates both output modes
func TestBindCodeRequest_AllOutputModes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	modes := []string{"quick", "full"}

	for _, mode := range modes {
		t.Run("OutputMode_"+mode, func(t *testing.T) {
			jsonBody := `{"pasted_code": "test", "output_mode": "` + mode + `"}`

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("POST", "/test", bytes.NewBufferString(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler := createTestHandler(t)
			req, ok := handler.bindCodeRequest(c)

			require.True(t, ok)
			require.NotNil(t, req)
			assert.Equal(t, mode, req.OutputMode)
		})
	}
}

// TestBindCodeRequest_ModelField tests that model field is still extracted
func TestBindCodeRequest_ModelField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jsonBody := `{"pasted_code": "test", "model": "deepseek-coder-v2", "user_mode": "expert", "output_mode": "full"}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/test", bytes.NewBufferString(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := createTestHandler(t)
	req, ok := handler.bindCodeRequest(c)

	require.True(t, ok)
	require.NotNil(t, req)

	// Verify all fields extracted correctly
	assert.Equal(t, "test", req.PastedCode)
	assert.Equal(t, "deepseek-coder-v2", req.Model)
	assert.Equal(t, "expert", req.UserMode)
	assert.Equal(t, "full", req.OutputMode)
}
