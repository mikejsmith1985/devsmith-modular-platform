// Package internal_logs_handlers provides HTTP handlers for logs operations.
package internal_logs_handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAnalysisService mocks the analysis service for testing
type MockAnalysisService struct {
	mock.Mock
}

func (m *MockAnalysisService) AnalyzeLogEntry(ctx context.Context, entry *logs_models.LogEntry) (*logs_services.AnalysisResult, error) {
	args := m.Called(ctx, entry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*logs_services.AnalysisResult), args.Error(1)
}

func (m *MockAnalysisService) ClassifyLogEntry(ctx context.Context, entry *logs_models.LogEntry) (string, error) {
	args := m.Called(ctx, entry)
	return args.String(0), args.Error(1)
}

// TestAnalyzeLog_Success tests successful log analysis
func TestAnalyzeLog_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockAnalysisService)
	logger := logrus.New()
	handler := NewAnalysisHandler(mockService, logger)
	
	// Create test request
	logEntry := logs_models.LogEntry{
		ID:        1,
		Service:   "portal",
		Level:     "error",
		Message:   "database connection refused",
		Metadata:  []byte(`{"correlation_id":"req-123"}`),
		CreatedAt: time.Now(),
	}
	
	expectedAnalysis := &logs_services.AnalysisResult{
		RootCause:    "PostgreSQL connection refused",
		SuggestedFix: "Check database service status",
		Severity:     5,
		RelatedLogs:  []string{"req-123"},
		FixSteps:     []string{"Verify PostgreSQL is running", "Check connection string"},
	}
	
	mockService.On("AnalyzeLogEntry", mock.Anything, mock.MatchedBy(func(entry *logs_models.LogEntry) bool {
		return entry.Message == "database connection refused"
	})).Return(expectedAnalysis, nil)
	
	// Create request
	reqBody := AnalyzeLogRequest{
		LogEntry: logEntry,
	}
	jsonData, _ := json.Marshal(reqBody)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/logs/analyze", bytes.NewReader(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	// Execute
	handler.AnalyzeLog(c)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	
	data := response.Data.(map[string]interface{})
	assert.Equal(t, "PostgreSQL connection refused", data["root_cause"])
	assert.Equal(t, float64(5), data["severity"])
	
	mockService.AssertExpectations(t)
}

// TestAnalyzeLog_InvalidRequest tests handling of invalid JSON
func TestAnalyzeLog_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockAnalysisService)
	logger := logrus.New()
	handler := NewAnalysisHandler(mockService, logger)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/logs/analyze", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	
	handler.AnalyzeLog(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response Response
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)
	assert.NotEmpty(t, response.Error)
}

// TestAnalyzeLog_ServiceError tests handling of service errors
func TestAnalyzeLog_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockAnalysisService)
	logger := logrus.New()
	handler := NewAnalysisHandler(mockService, logger)
	
	logEntry := logs_models.LogEntry{
		ID:      1,
		Service: "portal",
		Level:   "error",
		Message: "some error",
	}
	
	mockService.On("AnalyzeLogEntry", mock.Anything, mock.Anything).Return(nil, assert.AnError)
	
	reqBody := AnalyzeLogRequest{LogEntry: logEntry}
	jsonData, _ := json.Marshal(reqBody)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/logs/analyze", bytes.NewReader(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	handler.AnalyzeLog(c)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	mockService.AssertExpectations(t)
}

// TestClassifyLog_Success tests successful log classification
func TestClassifyLog_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockAnalysisService)
	logger := logrus.New()
	handler := NewAnalysisHandler(mockService, logger)
	
	logEntry := logs_models.LogEntry{
		ID:      1,
		Service: "portal",
		Level:   "error",
		Message: "connection refused to database",
	}
	
	mockService.On("ClassifyLogEntry", mock.Anything, mock.Anything).Return("db_connection", nil)
	
	reqBody := ClassifyLogRequest{LogEntry: logEntry}
	jsonData, _ := json.Marshal(reqBody)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/logs/classify", bytes.NewReader(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	handler.ClassifyLog(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response Response
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	
	data := response.Data.(map[string]interface{})
	assert.Equal(t, "db_connection", data["issue_type"])
	
	mockService.AssertExpectations(t)
}
