package logs_handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewUIHandler(t *testing.T) {
	logger := logrus.New()
	handler := NewUIHandler(logger, nil)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
}

func TestNewUIHandler_NilLogger(t *testing.T) {
	handler := NewUIHandler(nil, nil)
	assert.NotNil(t, handler)
	assert.Nil(t, handler.logger)
}

func TestUIHandler_DashboardHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "DevSmith Logs")
	assert.Contains(t, body, "logs-container")
	assert.Contains(t, body, "logs-output")
	assert.Contains(t, body, "level-filter")
	assert.Contains(t, body, "service-filter")
	assert.Contains(t, body, "search-input")
}

func TestUIHandler_DashboardHandler_ContainsControls(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "pause-btn")
	assert.Contains(t, body, "auto-scroll-btn")
	assert.Contains(t, body, "clear-btn")
	assert.Contains(t, body, "connection-status")
}

func TestUIHandler_DashboardHandler_ContainsScripts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "tailwindcss", "Should include Tailwind CSS")
	assert.Contains(t, body, "alpinejs", "Should include Alpine.js for interactivity")
	assert.Contains(t, body, "logs-output", "Should include logs output container")
	assert.Contains(t, body, "logs.css", "Should include logs stylesheet")
}

func TestUIHandler_DashboardHandler_ContainsStylesheet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "logs.css")
}

func TestUIHandler_HealthHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/health", uiHandler.HealthHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "logs")
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestUIHandler_HealthHandler_JSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/health", uiHandler.HealthHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestRegisterUIRoutes_DashboardRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	RegisterUIRoutes(router, uiHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "DevSmith Logs")
}

func TestRegisterUIRoutes_DashboardAliasRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	RegisterUIRoutes(router, uiHandler)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "DevSmith Logs")
}

func TestRegisterUIRoutes_HealthRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	RegisterUIRoutes(router, uiHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestRegisterUIRoutes_MultipleRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	RegisterUIRoutes(router, uiHandler)

	tests := []struct {
		name       string
		path       string
		expectText string
		expectCode int
	}{
		{"dashboard root", "/", "DevSmith Logs", http.StatusOK},
		{"dashboard alias", "/dashboard", "DevSmith Logs", http.StatusOK},
		{"health check", "/health", "healthy", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectCode, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectText)
		})
	}
}

func TestUIHandler_DashboardHandler_AllFilterOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Level filter options
	assert.Contains(t, body, `value="all"`)
	assert.Contains(t, body, `value="info"`)
	assert.Contains(t, body, `value="warn"`)
	assert.Contains(t, body, `value="error"`)

	// Service filter options
	assert.Contains(t, body, `value="portal"`)
	assert.Contains(t, body, `value="review"`)
	assert.Contains(t, body, `value="logs"`)
	assert.Contains(t, body, `value="analytics"`)
}

func TestUIHandler_DashboardHandler_InputElements(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()

	assert.Contains(t, body, `id="level-filter"`)
	assert.Contains(t, body, `id="service-filter"`)
	assert.Contains(t, body, `id="search-input"`)
	assert.Contains(t, body, `type="text"`)
}

func TestUIHandler_DashboardHandler_ValidHTML(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Basic HTML structure
	bodyLower := strings.ToLower(body)
	assert.Contains(t, bodyLower, "<!doctype html")
	assert.Contains(t, body, "<html")
	assert.Contains(t, body, "</html>")
	assert.Contains(t, body, "<head>")
	assert.Contains(t, body, "</head>")
	assert.Contains(t, body, "<body")  // Check for body tag with any attributes
	assert.Contains(t, body, "</body>")

	// Check tags are properly closed
	htmlOpenCount := strings.Count(body, "<html")
	htmlCloseCount := strings.Count(body, "</html>")
	assert.Greater(t, htmlOpenCount, 0)
	assert.Equal(t, htmlOpenCount, htmlCloseCount)
}

func TestUIHandler_DashboardHandler_NavbarPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "<nav")
	assert.Contains(t, body, "DevSmith", "Should have DevSmith branding")
}

func TestUIHandler_DashboardHandler_MetaTags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "charset")
	assert.Contains(t, body, "viewport")
}

func TestUIHandler_DashboardHandler_TitleTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "<title>")
	assert.Contains(t, body, "</title>")
	assert.Contains(t, body, "DevSmith Logs")
}

func TestRegisterUIRoutes_NilRouter(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()

	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	RegisterUIRoutes(nil, uiHandler)
}

func TestRegisterUIRoutes_NilLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Should not panic
	uiHandler := NewUIHandler(nil, nil)
	RegisterUIRoutes(router, uiHandler)
}

func TestUIHandler_DashboardHandler_ContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that response is HTML (Content-Type is set by templ)
	contentType := w.Header().Get("Content-Type")
	assert.True(t, contentType == "" || strings.Contains(contentType, "text/html"),
		"Content-Type should be empty or contain text/html, got: %s", contentType)
}

func TestUIHandler_HealthHandler_Status200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/health", uiHandler.HealthHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUIHandler_DashboardHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	RegisterUIRoutes(router, uiHandler)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUIHandler_DashboardHandler_OnlyGET(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUIHandler_DashboardHandler_WithQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/?test=value&foo=bar", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "DevSmith Logs")
}

func TestUIHandler_DashboardHandler_BodyNotEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Greater(t, w.Body.Len(), 0)
}

func TestUIHandler_HealthHandler_JSONFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/health", uiHandler.HealthHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Basic JSON check
	assert.True(t, strings.HasPrefix(strings.TrimSpace(w.Body.String()), "{"))
	assert.True(t, strings.HasSuffix(strings.TrimSpace(w.Body.String()), "}"))
}

func TestUIHandler_HealthHandler_ContainsServiceField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/health", uiHandler.HealthHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, `"service"`)
	assert.Contains(t, body, `"logs"`)
}

func TestUIHandler_HealthHandler_ContainsStatusField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/health", uiHandler.HealthHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, `"status"`)
	assert.Contains(t, body, `"healthy"`)
}

func BenchmarkUIHandler_DashboardHandler(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/", uiHandler.DashboardHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkUIHandler_HealthHandler(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	logger.Out = io.Discard

	uiHandler := NewUIHandler(logger, nil)
	router.GET("/health", uiHandler.HealthHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// MockResponseWriter for testing render errors
type MockResponseWriter struct {
	buffer        bytes.Buffer
	statusCode    int
	headerWritten bool
}

func (m *MockResponseWriter) Write(b []byte) (int, error) {
	return m.buffer.Write(b)
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
	m.headerWritten = true
}

func (m *MockResponseWriter) Header() http.Header {
	return http.Header{}
}
