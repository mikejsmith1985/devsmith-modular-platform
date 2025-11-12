// Package devsmith provides Go integration for DevSmith Logs API
/*
DevSmith Logger - Go Integration

Copy this file into your project and customize the configuration.

Usage:

	package main

	import (
		"os"
		"github.com/yourorg/yourproject/logger"
	)

	func main() {
		log := logger.NewLogger(
			os.Getenv("DEVSMITH_API_KEY"),
			os.Getenv("DEVSMITH_API_URL"),
			"my-project",
			"api-server",
		)
		defer log.Close()

		log.Info("User logged in", map[string]interface{}{"userId": 123})
		log.Error("Database error", map[string]interface{}{"code": "ECONNREFUSED"})
	}
*/

package devsmithlogger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Context   map[string]interface{} `json:"context"`
}

// BatchRequest is the request body for batch log ingestion
type BatchRequest struct {
	ProjectSlug string     `json:"project_slug"`
	Logs        []LogEntry `json:"logs"`
}

// DevSmithLogger handles buffering and batch sending of logs
type DevSmithLogger struct {
	apiKey        string
	apiURL        string
	projectSlug   string
	serviceName   string
	batchSize     int
	flushInterval time.Duration

	buffer     []LogEntry
	mutex      sync.Mutex
	ticker     *time.Ticker
	httpClient *http.Client
	done       chan bool
}

// NewLogger creates a new DevSmith logger instance
func NewLogger(apiKey, apiURL, projectSlug, serviceName string) *DevSmithLogger {
	return NewLoggerWithOptions(apiKey, apiURL, projectSlug, serviceName, 100, 5*time.Second)
}

// NewLoggerWithOptions creates a logger with custom batch size and flush interval
func NewLoggerWithOptions(
	apiKey, apiURL, projectSlug, serviceName string,
	batchSize int,
	flushInterval time.Duration,
) *DevSmithLogger {
	// Validate required config
	if apiKey == "" {
		panic("DevSmithLogger: apiKey is required")
	}
	if projectSlug == "" {
		panic("DevSmithLogger: projectSlug is required")
	}
	if serviceName == "" {
		panic("DevSmithLogger: serviceName is required")
	}
	if apiURL == "" {
		apiURL = "http://localhost:3000"
	}

	logger := &DevSmithLogger{
		apiKey:        apiKey,
		apiURL:        apiURL,
		projectSlug:   projectSlug,
		serviceName:   serviceName,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		buffer:        make([]LogEntry, 0, batchSize),
		ticker:        time.NewTicker(flushInterval),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		done: make(chan bool),
	}

	// Start flush timer
	go logger.flushPeriodically()

	return logger
}

// flushPeriodically runs in background and flushes logs periodically
func (l *DevSmithLogger) flushPeriodically() {
	for {
		select {
		case <-l.ticker.C:
			l.Flush()
		case <-l.done:
			return
		}
	}
}

// Log adds an entry to the buffer
func (l *DevSmithLogger) Log(level, message string, context map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Service:   l.serviceName,
		Context:   context,
	}

	l.mutex.Lock()
	l.buffer = append(l.buffer, entry)
	shouldFlush := len(l.buffer) >= l.batchSize
	l.mutex.Unlock()

	if shouldFlush {
		l.Flush()
	}
}

// Flush sends buffered logs to DevSmith API
func (l *DevSmithLogger) Flush() {
	l.mutex.Lock()
	if len(l.buffer) == 0 {
		l.mutex.Unlock()
		return
	}

	logs := make([]LogEntry, len(l.buffer))
	copy(logs, l.buffer)
	l.buffer = l.buffer[:0]
	l.mutex.Unlock()

	payload := BatchRequest{
		ProjectSlug: l.projectSlug,
		Logs:        logs,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("DevSmith Logger: Failed to marshal logs: %v\n", err)
		// Re-add logs to buffer
		l.mutex.Lock()
		l.buffer = append(l.buffer, logs...)
		l.mutex.Unlock()
		return
	}

	req, err := http.NewRequest("POST", l.apiURL+"/api/logs/batch", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("DevSmith Logger: Failed to create request: %v\n", err)
		// Re-add logs to buffer
		l.mutex.Lock()
		l.buffer = append(l.buffer, logs...)
		l.mutex.Unlock()
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.apiKey)

	resp, err := l.httpClient.Do(req)
	if err != nil {
		fmt.Printf("DevSmith Logger: Network error: %v\n", err)
		// Re-add logs to buffer
		l.mutex.Lock()
		l.buffer = append(l.buffer, logs...)
		l.mutex.Unlock()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		fmt.Printf("DevSmith Logger: Failed to send logs (%d)\n", resp.StatusCode)
		// Re-add logs to buffer
		l.mutex.Lock()
		l.buffer = append(l.buffer, logs...)
		l.mutex.Unlock()
	}
}

// Close flushes remaining logs and stops the logger
func (l *DevSmithLogger) Close() {
	l.ticker.Stop()
	l.done <- true
	l.Flush()
}

// Convenience methods
func (l *DevSmithLogger) Debug(message string, context map[string]interface{}) {
	l.Log("DEBUG", message, context)
}

func (l *DevSmithLogger) Info(message string, context map[string]interface{}) {
	l.Log("INFO", message, context)
}

func (l *DevSmithLogger) Warn(message string, context map[string]interface{}) {
	l.Log("WARN", message, context)
}

func (l *DevSmithLogger) Error(message string, context map[string]interface{}) {
	l.Log("ERROR", message, context)
}
