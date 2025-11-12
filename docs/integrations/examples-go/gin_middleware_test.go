//go:build integration
// +build integration

package devsmith

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// Mock logger for testing
type MockLogger struct {
	LogCalls []LogCall
}

type LogCall struct {
	Level   string
	Message string
	Context map[string]interface{}
	Tags    []string
}

func (m *MockLogger) Debug(message string, context map[string]interface{}, tags []string) {
	m.LogCalls = append(m.LogCalls, LogCall{"DEBUG", message, context, tags})
}

func (m *MockLogger) Info(message string, context map[string]interface{}, tags []string) {
	m.LogCalls = append(m.LogCalls, LogCall{"INFO", message, context, tags})
}

func (m *MockLogger) Warn(message string, context map[string]interface{}, tags []string) {
	m.LogCalls = append(m.LogCalls, LogCall{"WARN", message, context, tags})
}

func (m *MockLogger) Error(message string, context map[string]interface{}, tags []string) {
	m.LogCalls = append(m.LogCalls, LogCall{"ERROR", message, context, tags})
}

func (m *MockLogger) Close() {
	// Mock close
}

func TestInitializationValid(t *testing.T) {
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// Test configuration
type TestConfig struct {
	APIKey       string `json:"apiKey"`
	ProjectSlug  string `json:"projectSlug"`
	ProjectID    int    `json:"projectId"`
	APIUrl       string `json:"apiUrl"`
	BatchEndpoint string `json:"batchEndpoint"`
}

// Mock logger
type MockLogger struct {
	LogCalls []LogCall
}

type LogCall struct {
	Level   string
	Message string
	Context map[string]interface{}
	Tags    []string
}

func (m *MockLogger) Debug(message string, context map[string]interface{}, tags []string) {
	m.LogCalls = append(m.LogCalls, LogCall{"DEBUG", message, context, tags})
}

func (m *MockLogger) Info(message string, context map[string]interface{}, tags []string) {
	m.LogCalls = append(m.LogCalls, LogCall{"INFO", message, context, tags})
}

func (m *MockLogger) Warn(message string, context map[string]interface{}, tags []string) {
	m.LogCalls = append(m.LogCalls, LogCall{"WARN", message, context, tags})
}

func (m *MockLogger) Error(message string, context map[string]interface{}, tags []string) {
	m.LogCalls = append(m.LogCalls, LogCall{"ERROR", message, context, tags})
}

func (m *MockLogger) Close() {
	// Mock close
}

var testConfig TestConfig

func TestMain(m *testing.M) {
	// Load test configuration
	configPath := filepath.Join(".", ".test-config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Failed to load test config: %v\n", err)
		os.Exit(1)
	}
	
	if err := json.Unmarshal(data, &testConfig); err != nil {
		fmt.Printf("Failed to parse test config: %v\n", err)
		os.Exit(1)
	}
	
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestInitializationValid(t *testing.T) {
	mockLogger := &MockLogger{}
	middleware := DevSmithMiddleware(mockLogger, nil)
	
	if middleware == nil {
		t.Fatal("Middleware should not be nil")
	}
}

func TestRequestLogging(t *testing.T) {
	mockLogger := &MockLogger{}
	
	router := gin.New()
	router.Use(DevSmithMiddleware(mockLogger, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	
	// Find request log
	var requestLog *LogCall
	for i := range mockLogger.LogCalls {
		if strings.Contains(mockLogger.LogCalls[i].Message, "Incoming request") {
			requestLog = &mockLogger.LogCalls[i]
			break
		}
	}
	
	if requestLog == nil {
		t.Fatal("Should have request log")
	}
	
	if requestLog.Level != "INFO" {
		t.Errorf("Expected INFO level, got %s", requestLog.Level)
	}
	
	if requestLog.Context["method"] != "GET" {
		t.Errorf("Expected method GET, got %v", requestLog.Context["method"])
	}
	
	if requestLog.Context["path"] != "/test" {
		t.Errorf("Expected path /test, got %v", requestLog.Context["path"])
	}
}

func TestResponseLogging(t *testing.T) {
	mockLogger := &MockLogger{}
	
	router := gin.New()
	router.Use(DevSmithMiddleware(mockLogger, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	
	// Find response log
	var responseLog *LogCall
	for i := range mockLogger.LogCalls {
		if strings.Contains(mockLogger.LogCalls[i].Message, "Request completed") {
			responseLog = &mockLogger.LogCalls[i]
			break
		}
	}
	
	if responseLog == nil {
		t.Fatal("Should have response log")
	}
	
	statusCode := int(responseLog.Context["status_code"].(float64))
	if statusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", statusCode)
	}
	
	duration := responseLog.Context["duration"]
	if duration == nil {
		t.Error("Should have duration")
	}
}

func TestRequestTiming(t *testing.T) {
	mockLogger := &MockLogger{}
	
	router := gin.New()
	router.Use(DevSmithMiddleware(mockLogger, nil))
	router.GET("/slow", func(c *gin.Context) {
		time.Sleep(100 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/slow", nil)
	router.ServeHTTP(w, req)
	
	var responseLog *LogCall
	for i := range mockLogger.LogCalls {
		if strings.Contains(mockLogger.LogCalls[i].Message, "Request completed") {
			responseLog = &mockLogger.LogCalls[i]
			break
		}
	}
	
	duration := responseLog.Context["duration"].(float64)
	if duration < 100 {
		t.Errorf("Expected duration >= 100ms, got %.2fms", duration)
	}
	if duration >= 200 {
		t.Errorf("Expected duration < 200ms, got %.2fms", duration)
	}
}

func TestHeaderRedaction(t *testing.T) {
	mockLogger := &MockLogger{}
	
	router := gin.New()
	router.Use(DevSmithMiddleware(mockLogger, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer secret-token-12345")
	req.Header.Set("Cookie", "session=abc123")
	req.Header.Set("User-Agent", "test-agent")
	router.ServeHTTP(w, req)
	
	var requestLog *LogCall
	for i := range mockLogger.LogCalls {
		if strings.Contains(mockLogger.LogCalls[i].Message, "Incoming request") {
			requestLog = &mockLogger.LogCalls[i]
			break
		}
	}
	
	headers := requestLog.Context["headers"].(map[string]interface{})
	
	if headers["Authorization"] != "[REDACTED]" {
		t.Errorf("Expected Authorization to be redacted, got %v", headers["Authorization"])
	}
	
	if headers["Cookie"] != "[REDACTED]" {
		t.Errorf("Expected Cookie to be redacted, got %v", headers["Cookie"])
	}
	
	if headers["User-Agent"] != "test-agent" {
		t.Errorf("Expected User-Agent to be preserved, got %v", headers["User-Agent"])
	}
}

func TestSkipPaths(t *testing.T) {
	mockLogger := &MockLogger{}
	
	options := &MiddlewareOptions{
		SkipPaths: []string{"/health", "/metrics"},
	}
	
	router := gin.New()
	router.Use(DevSmithMiddleware(mockLogger, options))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	router.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	
	// Request skipped paths
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)
	
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)
	
	// Request normal path
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	
	// Should only log /test
	paths := []string{}
	for _, log := range mockLogger.LogCalls {
		if path, ok := log.Context["path"].(string); ok {
			paths = append(paths, path)
		}
	}
	
	hasTest := false
	hasHealth := false
	hasMetrics := false
	
	for _, path := range paths {
		if path == "/test" {
			hasTest = true
		}
		if path == "/health" {
			hasHealth = true
		}
		if path == "/metrics" {
			hasMetrics = true
		}
	}
	
	if !hasTest {
		t.Error("Should log /test")
	}
	if hasHealth {
		t.Error("Should not log /health")
	}
	if hasMetrics {
		t.Error("Should not log /metrics")
	}
}

func TestPanicRecovery(t *testing.T) {
	mockLogger := &MockLogger{}
	
	router := gin.New()
	router.Use(DevSmithMiddleware(mockLogger, nil))
	router.GET("/panic", func(c *gin.Context) {
		panic("Test panic")
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)
	
	// Find error log
	var errorLog *LogCall
	for i := range mockLogger.LogCalls {
		if mockLogger.LogCalls[i].Level == "ERROR" {
			errorLog = &mockLogger.LogCalls[i]
			break
		}
	}
	
	if errorLog == nil {
		t.Fatal("Should have error log for panic")
	}
	
	if !strings.Contains(errorLog.Message, "Test panic") {
		t.Errorf("Error message should contain panic text, got: %s", errorLog.Message)
	}
	
	if errorLog.Context["stack"] == nil {
		t.Error("Should have stack trace")
	}
}

func TestCustomTags(t *testing.T) {
	mockLogger := &MockLogger{}
	
	options := &MiddlewareOptions{
		Tags: []string{"api", "production"},
	}
	
	router := gin.New()
	router.Use(DevSmithMiddleware(mockLogger, options))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	
	var requestLog *LogCall
	for i := range mockLogger.LogCalls {
		if strings.Contains(mockLogger.LogCalls[i].Message, "Incoming request") {
			requestLog = &mockLogger.LogCalls[i]
			break
		}
	}
	
	tags := requestLog.Tags
	hasAPI := false
	hasProduction := false
	hasGin := false
	
	for _, tag := range tags {
		if tag == "api" {
			hasAPI = true
		}
		if tag == "production" {
			hasProduction = true
		}
		if tag == "gin" {
			hasGin = true
		}
	}
	
	if !hasAPI {
		t.Error("Should have 'api' tag")
	}
	if !hasProduction {
		t.Error("Should have 'production' tag")
	}
	if !hasGin {
		t.Error("Should have default 'gin' tag")
	}
}

func TestPostRequest(t *testing.T) {
	mockLogger := &MockLogger{}
	
	router := gin.New()
	router.Use(DevSmithMiddleware(mockLogger, nil))
	router.POST("/post", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"received": true})
	})
	
	w := httptest.NewRecorder()
	body := strings.NewReader(`{"data": "test"}`)
	req, _ := http.NewRequest("POST", "/post", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	var requestLog *LogCall
	for i := range mockLogger.LogCalls {
		if strings.Contains(mockLogger.LogCalls[i].Message, "Incoming request") {
			requestLog = &mockLogger.LogCalls[i]
			break
		}
	}
	
	if requestLog.Context["method"] != "POST" {
		t.Errorf("Expected method POST, got %v", requestLog.Context["method"])
	}
}

func TestPerformanceManyRequests(t *testing.T) {
	mockLogger := &MockLogger{}
	
	router := gin.New()
	router.Use(DevSmithMiddleware(mockLogger, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	
	numRequests := 100
	start := time.Now()
	
	for i := 0; i < numRequests; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
	}
	
	duration := time.Since(start)
	
	// Should complete 100 requests in reasonable time (< 5 seconds)
	if duration > 5*time.Second {
		t.Errorf("100 requests took %v (should be < 5s)", duration)
	}
	
	// Should have logged all requests
	if len(mockLogger.LogCalls) < numRequests*2 {
		t.Errorf("Expected at least %d log calls, got %d", numRequests*2, len(mockLogger.LogCalls))
	}
}
