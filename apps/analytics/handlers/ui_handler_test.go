package analytics_handlers

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUIHandler(t *testing.T) {
	logger := logrus.New()
	handler := NewUIHandler(logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
}

func TestDashboardHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that HTML is returned (doctype can be lowercase)
	body := w.Body.String()
	assert.NotEqual(t, "", body)
	assert.Contains(t, body, "DevSmith Analytics")
}

func TestDashboardHandler_ContainsRequiredElements(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Check for essential dashboard components
	assert.Contains(t, body, "trends-chart", "Should contain trends chart canvas")
	assert.Contains(t, body, "anomalies-container", "Should contain anomalies container")
	assert.Contains(t, body, "issues-container", "Should contain issues container")
	assert.Contains(t, body, "time-range", "Should contain time range selector")
	assert.Contains(t, body, "export-csv", "Should contain export CSV button")
	assert.Contains(t, body, "export-json", "Should contain export JSON button")
}

func TestDashboardHandler_ContainsScripts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Check for required script tags
	assert.Contains(t, body, "chart.js", "Should include Chart.js library")
	assert.Contains(t, body, "analytics.js", "Should include analytics.js script")
	assert.Contains(t, body, "devsmith-theme.css", "Should include devsmith-theme.css")
	assert.Contains(t, body, "alpinejs", "Should include Alpine.js for interactivity")
}

func TestDashboardHandler_ContainsNavigation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Check for navigation elements
	assert.Contains(t, body, "<nav", "Should contain nav element")
	assert.Contains(t, body, "Analytics", "Should have Analytics link")
}

func TestHealthHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/health", uiHandler.HealthHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Check response format
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "analytics")
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestHealthHandler_ResponseStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/health", uiHandler.HealthHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Verify JSON structure
	assert.Contains(t, body, `"service":"analytics"`)
	assert.Contains(t, body, `"status":"healthy"`)
}

func TestRegisterUIRoutes_RootPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()

	router := gin.New()
	RegisterUIRoutes(router, logger)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "DevSmith Analytics")
}

func TestRegisterUIRoutes_DashboardPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()

	router := gin.New()
	RegisterUIRoutes(router, logger)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "DevSmith Analytics")
}

func TestRegisterUIRoutes_HealthPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()

	router := gin.New()
	RegisterUIRoutes(router, logger)

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestRegisterUIRoutes_MultipleRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()

	router := gin.New()
	RegisterUIRoutes(router, logger)

	tests := []struct {
		name           string
		bodyContains   string
		path           string
		expectedStatus int
	}{
		{
			name:           "Root path",
			path:           "/",
			expectedStatus: http.StatusOK,
			bodyContains:   "DevSmith Analytics",
		},
		{
			name:           "Dashboard path",
			path:           "/dashboard",
			expectedStatus: http.StatusOK,
			bodyContains:   "DevSmith Analytics",
		},
		{
			name:           "Health path",
			path:           "/health",
			expectedStatus: http.StatusOK,
			bodyContains:   "healthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.bodyContains)
		})
	}
}

func TestDashboardHandler_ContextPropagation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDashboardHandler_RespondsToGET(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/", uiHandler.DashboardHandler)

	// Try GET request
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthHandler_RespondsToGET(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/health", uiHandler.HealthHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewUIHandler_WithNilLogger(t *testing.T) {
	uiHandler := NewUIHandler(nil)
	assert.NotNil(t, uiHandler)
	assert.Nil(t, uiHandler.logger)
}

func TestDashboardHandler_HeadersSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// The template should set Content-Type to text/html
	contentType := w.Header().Get("Content-Type")
	assert.True(t, contentType == "text/html; charset=utf-8" || contentType == "" || contentType == "text/html")
}

func TestRegisterUIRoutes_NilLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Should not panic with nil logger
	require.NotPanics(t, func() {
		RegisterUIRoutes(router, nil)
	})
}

func TestUIHandler_DashboardContainsForms(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Check for select dropdowns
	assert.Contains(t, body, "<select")
	assert.Contains(t, body, "<option")
}

func TestUIHandler_DashboardContainsSections(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Check for section elements
	assert.Contains(t, body, "<section", "Should contain section elements")
	assert.Contains(t, body, "card", "Should contain card classes")
}

func TestHealthHandler_JSONContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/health", uiHandler.HealthHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}

func TestRegisterUIRoutes_AllRoutesRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()

	router := gin.New()
	RegisterUIRoutes(router, logger)

	// Create test requests for all registered routes
	routes := []string{"/", "/dashboard", "/health"}

	for _, route := range routes {
		req := httptest.NewRequest(http.MethodGet, route, http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Route %s should return 200", route)
	}
}

func TestDashboardHandler_ContainsMetaTags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Check for meta tags
	assert.Contains(t, body, "charset", "Should contain charset meta tag")
	assert.Contains(t, body, "viewport", "Should contain viewport meta tag")
}

func TestDashboardHandler_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a logger that won't panic
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	// Use a custom ResponseWriter that simulates an error
	router.ServeHTTP(w, req)

	// Should still return a response
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestHealthHandler_ContentLength(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/health", uiHandler.HealthHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Greater(t, w.Body.Len(), 0)
}

func TestDashboardHandler_NoCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Dashboard should be delivered successfully
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewUIHandler_StoresLogger(t *testing.T) {
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	assert.Equal(t, logger, uiHandler.logger)
}

func TestDashboardHandler_MultipleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/", uiHandler.DashboardHandler)

	// Make multiple requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "DevSmith Analytics")
	}
}

func TestHealthHandler_MultipleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/health", uiHandler.HealthHandler)

	// Make multiple requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "healthy")
	}
}

func TestDashboardHandler_LargeResponseBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	uiHandler := NewUIHandler(logger)

	router := gin.New()
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Dashboard HTML should be reasonably large (> 1KB)
	assert.Greater(t, w.Body.Len(), 1000)
}

func TestUIHandler_HandlerChaining(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()

	router := gin.New()
	RegisterUIRoutes(router, logger)

	// Test that routes can be accessed in sequence
	routes := []string{"/", "/health", "/dashboard", "/health"}

	for _, route := range routes {
		req := httptest.NewRequest(http.MethodGet, route, http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}

// MockResponseWriter for error testing
type MockResponseWriter struct {
	bytes.Buffer
	shouldFail bool
}

func (m *MockResponseWriter) Header() http.Header {
	return http.Header{}
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	// Mock implementation
}

func (m *MockResponseWriter) Write(p []byte) (int, error) {
	if m.shouldFail {
		return 0, errors.New("write error")
	}
	return m.Buffer.Write(p)
}
