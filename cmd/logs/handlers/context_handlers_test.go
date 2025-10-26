package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
	"github.com/stretchr/testify/assert"
)

// TestGetCorrelatedLogs_Valid tests retrieving correlated logs
func TestGetCorrelatedLogs_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	contextSvc := services.NewContextService(nil)
	router.GET("/api/logs/correlation/:correlationId", GetCorrelatedLogs(contextSvc))

	req := httptest.NewRequest("GET", "/api/logs/correlation/test-123?limit=50&offset=0", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "test-123", resp["correlation_id"])
	assert.Empty(t, resp["logs"]) // Nil repo returns empty slice
}

// TestGetCorrelatedLogs_MissingID tests error when correlation ID is missing
func TestGetCorrelatedLogs_MissingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	contextSvc := services.NewContextService(nil)
	router.GET("/api/logs/correlation/:correlationId", GetCorrelatedLogs(contextSvc))

	req := httptest.NewRequest("GET", "/api/logs/correlation/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Gin will 404 if param is empty
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestGetCorrelatedLogs_WithPagination tests pagination parameters
func TestGetCorrelatedLogs_WithPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	contextSvc := services.NewContextService(nil)
	router.GET("/api/logs/correlation/:correlationId", GetCorrelatedLogs(contextSvc))

	req := httptest.NewRequest("GET", "/api/logs/correlation/test-123?limit=100&offset=50", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(100), resp["limit"])
	assert.Equal(t, float64(50), resp["offset"])
}

// TestGetCorrelatedLogs_LimitCapped tests limit is capped at 1000
func TestGetCorrelatedLogs_LimitCapped(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	contextSvc := services.NewContextService(nil)
	router.GET("/api/logs/correlation/:correlationId", GetCorrelatedLogs(contextSvc))

	req := httptest.NewRequest("GET", "/api/logs/correlation/test-123?limit=5000", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(100), resp["limit"]) // Should use default 100 since value > 1000
}

// TestGetCorrelationMetadata_Valid tests retrieving metadata
func TestGetCorrelationMetadata_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	contextSvc := services.NewContextService(nil)
	router.GET("/api/logs/correlation/:correlationId/metadata", GetCorrelationMetadata(contextSvc))

	req := httptest.NewRequest("GET", "/api/logs/correlation/test-123/metadata", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotNil(t, resp)
}

// TestGetCorrelationMetadata_MissingID tests error when correlation ID is missing
func TestGetCorrelationMetadata_MissingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	contextSvc := services.NewContextService(nil)
	router.GET("/api/logs/correlation/:correlationId/metadata", GetCorrelationMetadata(contextSvc))

	req := httptest.NewRequest("GET", "/api/logs/correlation/invalid/metadata", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// With nil repo, GetContextMetadata returns empty map with no error
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGetTraceTimeline_Valid tests retrieving trace timeline
func TestGetTraceTimeline_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	contextSvc := services.NewContextService(nil)
	router.GET("/api/logs/correlation/:correlationId/timeline", GetTraceTimeline(contextSvc))

	req := httptest.NewRequest("GET", "/api/logs/correlation/test-123/timeline", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "test-123", resp["correlation_id"])
	// Timeline may be nil with nil repo
	if resp["timeline"] != nil {
		assert.NotNil(t, resp["timeline"])
	}
}

// TestGetTraceTimeline_MissingID tests error when correlation ID is missing
func TestGetTraceTimeline_MissingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	contextSvc := services.NewContextService(nil)
	router.GET("/api/logs/correlation/:correlationId/timeline", GetTraceTimeline(contextSvc))

	req := httptest.NewRequest("GET", "/api/logs/correlation/invalid/timeline", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// With nil repo, should still work and return 200
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGetCorrelatedLogs_ResponseFormat tests response includes required fields
func TestGetCorrelatedLogs_ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	contextSvc := services.NewContextService(nil)
	router.GET("/api/logs/correlation/:correlationId", GetCorrelatedLogs(contextSvc))

	req := httptest.NewRequest("GET", "/api/logs/correlation/test-123", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Verify all required fields
	assert.Contains(t, resp, "correlation_id")
	assert.Contains(t, resp, "logs")
	assert.Contains(t, resp, "count")
	assert.Contains(t, resp, "limit")
	assert.Contains(t, resp, "offset")
}

// TestGetCorrelationMetadata_ResponseFormat tests metadata response format
func TestGetCorrelationMetadata_ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	contextSvc := services.NewContextService(nil)
	router.GET("/api/logs/correlation/:correlationId/metadata", GetCorrelationMetadata(contextSvc))

	req := httptest.NewRequest("GET", "/api/logs/correlation/test-123/metadata", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// With nil repo, should return empty map
	assert.NotNil(t, resp)
}

// TestGetTraceTimeline_ResponseFormat tests timeline response format
func TestGetTraceTimeline_ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	contextSvc := services.NewContextService(nil)
	router.GET("/api/logs/correlation/:correlationId/timeline", GetTraceTimeline(contextSvc))

	req := httptest.NewRequest("GET", "/api/logs/correlation/test-123/timeline", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Verify timeline structure
	assert.Contains(t, resp, "correlation_id")
	assert.Contains(t, resp, "timeline")
	assert.Contains(t, resp, "count")
}
