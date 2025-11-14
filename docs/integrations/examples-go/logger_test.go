//go:build integration
// +build integration

package devsmith

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// Test configuration loaded from .test-config.json
type TestConfig struct {
	APIKey       string `json:"apiKey"`
	ProjectSlug  string `json:"projectSlug"`
	ProjectID    int    `json:"projectId"`
	APIUrl       string `json:"apiUrl"`
	BatchEndpoint string `json:"batchEndpoint"`
}

var testConfig TestConfig
var receivedRequests []ReceivedRequest
var requestsMutex sync.Mutex

type ReceivedRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    BatchRequest
}

// Load test configuration
func loadTestConfig() error {
	configPath := filepath.Join(".", ".test-config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read test config: %w", err)
	}
	
	if err := json.Unmarshal(data, &testConfig); err != nil {
		return fmt.Errorf("failed to parse test config: %w", err)
	}
	
	return nil
}

// Mock HTTP server handler
func mockServerHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	
	var batchReq BatchRequest
	json.Unmarshal(body, &batchReq)
	
	request := ReceivedRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Headers: map[string]string{
			"X-Api-Key": r.Header.Get("X-Api-Key"),
		},
		Body: batchReq,
	}
	
	requestsMutex.Lock()
	receivedRequests = append(receivedRequests, request)
	requestsMutex.Unlock()
	
	// Validate API key
	if r.Header.Get("X-Api-Key") != testConfig.APIKey {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid API key"})
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"received": len(batchReq.Logs),
	})
}

func TestMain(m *testing.M) {
	// Load test configuration
	if err := loadTestConfig(); err != nil {
		fmt.Printf("Failed to load test config: %v\n", err)
		os.Exit(1)
	}
	
	// Run tests
	os.Exit(m.Run())
}

func resetReceivedRequests() {
	requestsMutex.Lock()
	receivedRequests = []ReceivedRequest{}
	requestsMutex.Unlock()
}

func getReceivedRequests() []ReceivedRequest {
	requestsMutex.Lock()
	defer requestsMutex.Unlock()
	return append([]ReceivedRequest{}, receivedRequests...)
}

func TestInitializationValid(t *testing.T) {
	logger := NewLogger(
		testConfig.APIKey,
		"http://localhost:8997",
		testConfig.ProjectSlug,
		"test-service",
	)
	
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}
	
	if logger.projectSlug != testConfig.ProjectSlug {
		t.Errorf("Expected project slug %s, got %s", testConfig.ProjectSlug, logger.projectSlug)
	}
	
	if logger.serviceName != "test-service" {
		t.Errorf("Expected service name 'test-service', got %s", logger.serviceName)
	}
	
	logger.Close()
}

func TestInitializationMissingAPIKey(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for missing API key")
		}
	}()
	
	NewLogger("", "http://localhost:8997", "test", "service")
}

func TestInitializationMissingProjectSlug(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for missing project slug")
		}
	}()
	
	NewLogger("key", "http://localhost:8997", "", "service")
}

func TestBufferManagement(t *testing.T) {
	logger := NewLogger(
		testConfig.APIKey,
		"http://localhost:8997",
		testConfig.ProjectSlug,
		"test",
	)
	defer logger.Close()
	
	logger.Info("Message 1", nil, nil)
	logger.Info("Message 2", nil, nil)
	
	logger.mutex.Lock()
	bufferLen := len(logger.buffer)
	logger.mutex.Unlock()
	
	if bufferLen != 2 {
		t.Errorf("Expected buffer length 2, got %d", bufferLen)
	}
}

func TestBufferSizeLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(mockServerHandler))
	defer server.Close()
	resetReceivedRequests()
	
	logger := NewLogger(
		testConfig.APIKey,
		server.URL,
		testConfig.ProjectSlug,
		"test",
	)
	logger.bufferSize = 5
	
	for i := 0; i < 10; i++ {
		logger.Info(fmt.Sprintf("Message %d", i), nil, nil)
	}
	
	time.Sleep(500 * time.Millisecond)
	
	logger.mutex.Lock()
	bufferLen := len(logger.buffer)
	logger.mutex.Unlock()
	
	if bufferLen >= 10 {
		t.Errorf("Buffer should have been flushed, got length %d", bufferLen)
	}
	
	logger.Close()
}

func TestBufferFlushOnFull(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(mockServerHandler))
	defer server.Close()
	resetReceivedRequests()
	
	logger := NewLogger(
		testConfig.APIKey,
		server.URL,
		testConfig.ProjectSlug,
		"test",
	)
	logger.bufferSize = 3
	
	logger.Info("Message 1", nil, nil)
	logger.Info("Message 2", nil, nil)
	logger.Info("Message 3", nil, nil) // Should trigger flush
	
	time.Sleep(500 * time.Millisecond)
	
	requests := getReceivedRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	
	if len(requests[0].Body.Logs) != 3 {
		t.Errorf("Expected 3 logs in batch, got %d", len(requests[0].Body.Logs))
	}
	
	logger.Close()
}

func TestLogLevels(t *testing.T) {
	logger := NewLogger(
		testConfig.APIKey,
		"http://localhost:8997",
		testConfig.ProjectSlug,
		"test",
	)
	defer logger.Close()
	
	logger.Debug("Debug message", nil, nil)
	logger.Info("Info message", nil, nil)
	logger.Warn("Warning message", nil, nil)
	logger.Error("Error message", nil, nil)
	
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	
	if len(logger.buffer) != 4 {
		t.Fatalf("Expected 4 logs, got %d", len(logger.buffer))
	}
	
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	for i, expected := range levels {
		if logger.buffer[i].Level != expected {
			t.Errorf("Expected level %s at index %d, got %s", expected, i, logger.buffer[i].Level)
		}
	}
}

func TestContextAndTags(t *testing.T) {
	logger := NewLogger(
		testConfig.APIKey,
		"http://localhost:8997",
		testConfig.ProjectSlug,
		"test",
	)
	defer logger.Close()
	
	context := map[string]interface{}{
		"user_id": 123,
		"action":  "login",
	}
	tags := []string{"auth", "user"}
	
	logger.Info("Message", context, tags)
	
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	
	entry := logger.buffer[0]
	if entry.Context == nil {
		t.Fatal("Context should not be nil")
	}
	
	if entry.Context["user_id"] != float64(123) { // JSON unmarshals numbers as float64
		t.Errorf("Expected user_id 123, got %v", entry.Context["user_id"])
	}
	
	if len(entry.Tags) != 2 || entry.Tags[0] != "auth" || entry.Tags[1] != "user" {
		t.Errorf("Expected tags ['auth', 'user'], got %v", entry.Tags)
	}
}

func TestBatchFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(mockServerHandler))
	defer server.Close()
	resetReceivedRequests()
	
	logger := NewLogger(
		testConfig.APIKey,
		server.URL,
		testConfig.ProjectSlug,
		"test-service",
	)
	logger.bufferSize = 2
	
	logger.Info("Message 1", nil, nil)
	logger.Info("Message 2", nil, nil)
	
	time.Sleep(500 * time.Millisecond)
	
	requests := getReceivedRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	
	req := requests[0]
	if req.Method != "POST" {
		t.Errorf("Expected POST method, got %s", req.Method)
	}
	
	if req.Headers["X-Api-Key"] != testConfig.APIKey {
		t.Errorf("Expected API key %s, got %s", testConfig.APIKey, req.Headers["X-Api-Key"])
	}
	
	if req.Body.ProjectSlug != testConfig.ProjectSlug {
		t.Errorf("Expected project slug %s, got %s", testConfig.ProjectSlug, req.Body.ProjectSlug)
	}
	
	if len(req.Body.Logs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(req.Body.Logs))
	}
	
	logger.Close()
}

func TestBatchRequiredFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(mockServerHandler))
	defer server.Close()
	resetReceivedRequests()
	
	logger := NewLogger(
		testConfig.APIKey,
		server.URL,
		testConfig.ProjectSlug,
		"test-service",
	)
	logger.bufferSize = 1
	
	context := map[string]interface{}{"key": "value"}
	tags := []string{"tag1"}
	logger.Info("Test message", context, tags)
	
	time.Sleep(500 * time.Millisecond)
	
	requests := getReceivedRequests()
	if len(requests) == 0 {
		t.Fatal("No requests received")
	}
	
	entry := requests[0].Body.Logs[0]
	
	if entry.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
	
	if entry.Level != "INFO" {
		t.Errorf("Expected level INFO, got %s", entry.Level)
	}
	
	if entry.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got %s", entry.Message)
	}
	
	if entry.Service != "test-service" {
		t.Errorf("Expected service 'test-service', got %s", entry.Service)
	}
	
	logger.Close()
}

func TestTimeBasedFlush(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(mockServerHandler))
	defer server.Close()
	resetReceivedRequests()
	
	logger := NewLogger(
		testConfig.APIKey,
		server.URL,
		testConfig.ProjectSlug,
		"test",
	)
	logger.flushInterval = 2 * time.Second
	
	logger.Info("Message 1", nil, nil)
	
	// Should not have sent yet
	time.Sleep(500 * time.Millisecond)
	requests := getReceivedRequests()
	if len(requests) != 0 {
		t.Errorf("Expected 0 requests before flush interval, got %d", len(requests))
	}
	
	// Wait for flush interval
	time.Sleep(2 * time.Second)
	requests = getReceivedRequests()
	if len(requests) != 1 {
		t.Errorf("Expected 1 request after flush interval, got %d", len(requests))
	}
	
	logger.Close()
}

func TestCleanupOnClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(mockServerHandler))
	defer server.Close()
	resetReceivedRequests()
	
	logger := NewLogger(
		testConfig.APIKey,
		server.URL,
		testConfig.ProjectSlug,
		"test",
	)
	
	logger.Info("Message 1", nil, nil)
	logger.Info("Message 2", nil, nil)
	
	logger.Close()
	
	time.Sleep(500 * time.Millisecond)
	
	requests := getReceivedRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	
	if len(requests[0].Body.Logs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(requests[0].Body.Logs))
	}
}

func TestConcurrentLogging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(mockServerHandler))
	defer server.Close()
	resetReceivedRequests()
	
	logger := NewLogger(
		testConfig.APIKey,
		server.URL,
		testConfig.ProjectSlug,
		"test",
	)
	logger.bufferSize = 100
	
	var wg sync.WaitGroup
	numGoroutines := 5
	messagesPerGoroutine := 10
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Info(fmt.Sprintf("Goroutine %d - Message %d", id, j), nil, nil)
			}
		}(i)
	}
	
	wg.Wait()
	logger.Close()
	
	time.Sleep(500 * time.Millisecond)
	
	// Count total logs received
	requests := getReceivedRequests()
	totalLogs := 0
	for _, req := range requests {
		totalLogs += len(req.Body.Logs)
	}
	
	expectedLogs := numGoroutines * messagesPerGoroutine
	if totalLogs != expectedLogs {
		t.Errorf("Expected %d logs total, got %d", expectedLogs, totalLogs)
	}
}
