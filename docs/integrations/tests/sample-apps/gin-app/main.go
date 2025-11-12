package main
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// DevSmith Logger (simplified from docs/integrations/go/logger.go)
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Context   map[string]interface{} `json:"context"`
	Tags      []string               `json:"tags"`
}

type DevSmithLogger struct {
	apiURL       string
	apiKey       string
	projectSlug  string
	serviceName  string
	buffer       []LogEntry
	bufferSize   int
	flushTimer   *time.Timer
	mu           sync.Mutex
	httpClient   *http.Client
	stopChan     chan struct{}
	flushPending bool
}

func NewLogger(apiURL, apiKey, projectSlug, serviceName string, bufferSize int, flushInterval time.Duration) *DevSmithLogger {
	logger := &DevSmithLogger{
		apiURL:      apiURL,
		apiKey:      apiKey,
		projectSlug: projectSlug,
		serviceName: serviceName,
		buffer:      make([]LogEntry, 0, bufferSize),
		bufferSize:  bufferSize,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		stopChan:    make(chan struct{}),
	}
	
	logger.scheduleFlush(flushInterval)
	return logger
}

func (l *DevSmithLogger) scheduleFlush(interval time.Duration) {
	l.flushTimer = time.AfterFunc(interval, func() {
		l.Flush()
		l.scheduleFlush(interval)
	})
}

func (l *DevSmithLogger) Flush() {
	l.mu.Lock()
	if len(l.buffer) == 0 {
		l.mu.Unlock()
		return
	}
	
	batch := make([]LogEntry, len(l.buffer))
	copy(batch, l.buffer)
	l.buffer = l.buffer[:0]
	l.mu.Unlock()
	
	payload := map[string]interface{}{
		"project_slug": l.projectSlug,
		"logs":         batch,
	}
	
	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", l.apiURL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.apiKey)
	
	resp, err := l.httpClient.Do(req)
	if err != nil {
		log.Printf("DevSmith flush error: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

func (l *DevSmithLogger) log(level, message string, context map[string]interface{}, tags []string) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Service:   l.serviceName,
		Context:   context,
		Tags:      tags,
	}
	
	l.mu.Lock()
	l.buffer = append(l.buffer, entry)
	shouldFlush := len(l.buffer) >= l.bufferSize
	l.mu.Unlock()
	
	if shouldFlush {
		if l.flushTimer != nil {
			l.flushTimer.Stop()
		}
		l.Flush()
	}
}

func (l *DevSmithLogger) Debug(message string, context map[string]interface{}, tags []string) {
	l.log("DEBUG", message, context, tags)
}

func (l *DevSmithLogger) Info(message string, context map[string]interface{}, tags []string) {
	l.log("INFO", message, context, tags)
}

func (l *DevSmithLogger) Warn(message string, context map[string]interface{}, tags []string) {
	l.log("WARN", message, context, tags)
}

func (l *DevSmithLogger) Error(message string, context map[string]interface{}, tags []string) {
	l.log("ERROR", message, context, tags)
}

func (l *DevSmithLogger) Close() {
	if l.flushTimer != nil {
		l.flushTimer.Stop()
	}
	l.Flush()
	close(l.stopChan)
}

// Gin Middleware (simplified from docs/integrations/go/gin_middleware.go)
func DevSmithMiddleware(logger *DevSmithLogger, skipPaths []string, tags []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip paths
		for _, path := range skipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}
		
		start := time.Now()
		
		// Log incoming request
		logger.Info("Incoming request", map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"ip":     c.ClientIP(),
		}, append([]string{"request"}, tags...))
		
		// Recover from panics
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered", map[string]interface{}{
					"error": fmt.Sprintf("%v", err),
					"path":  c.Request.URL.Path,
				}, append([]string{"panic"}, tags...))
				
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			}
		}()
		
		c.Next()
		
		// Log response
		duration := time.Since(start).Milliseconds()
		logger.Info("Request completed", map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status_code": c.Writer.Status(),
			"duration_ms": duration,
		}, append([]string{"response"}, tags...))
	}
}

// User struct for examples
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	// Load .env
	godotenv.Load()
	
	// Initialize logger
	logger := NewLogger(
		os.Getenv("DEVSMITH_API_URL"),
		os.Getenv("DEVSMITH_API_KEY"),
		os.Getenv("DEVSMITH_PROJECT_SLUG"),
		os.Getenv("DEVSMITH_SERVICE_NAME"),
		100,
		5*time.Second,
	)
	defer logger.Close()
	
	// Setup Gin
	gin.SetMode(os.Getenv("GIN_MODE"))
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(DevSmithMiddleware(logger, []string{"/health", "/metrics"}, []string{"production", "api"}))
	
	// Routes
	router.GET("/", func(c *gin.Context) {
		logger.Info("Root endpoint accessed", map[string]interface{}{
			"ip": c.ClientIP(),
		}, []string{"endpoint", "public"})
		
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "DevSmith Gin Sample App",
			"endpoints": []string{
				"GET / - This page",
				"GET /health - Health check (not logged)",
				"GET /api/users - Get users list",
				"POST /api/users - Create user",
				"GET /api/error - Trigger error for testing",
				"GET /api/panic - Trigger panic for recovery testing",
			},
		})
	})
	
	router.GET("/health", func(c *gin.Context) {
		// Health check endpoint - skipped by middleware
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	
	router.GET("/metrics", func(c *gin.Context) {
		// Metrics endpoint - skipped by middleware
		c.JSON(http.StatusOK, gin.H{"metrics": "data"})
	})
	
	router.GET("/api/users", func(c *gin.Context) {
		logger.Debug("Fetching users list", map[string]interface{}{
			"page":  c.DefaultQuery("page", "1"),
			"limit": c.DefaultQuery("limit", "10"),
		}, []string{"users", "api"})
		
		// Simulate database query
		users := []User{
			{ID: 1, Username: "alice", Email: "alice@example.com"},
			{ID: 2, Username: "bob", Email: "bob@example.com"},
		}
		
		c.JSON(http.StatusOK, gin.H{
			"users": users,
			"count": len(users),
		})
	})
	
	router.POST("/api/users", func(c *gin.Context) {
		var userData User
		if err := c.ShouldBindJSON(&userData); err != nil {
			logger.Warn("User creation failed - invalid JSON", map[string]interface{}{
				"error": err.Error(),
			}, []string{"validation", "error"})
			
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}
		
		logger.Info("Creating new user", map[string]interface{}{
			"username": userData.Username,
			"email":    userData.Email,
		}, []string{"users", "create"})
		
		// Simulate validation
		if strings.TrimSpace(userData.Username) == "" {
			logger.Warn("User creation failed - missing username", map[string]interface{}{
				"provided_fields": []string{"email"},
			}, []string{"validation", "error"})
			
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username required"})
			return
		}
		
		// Simulate user creation
		userData.ID = 1000 + int(time.Now().Unix()%1000)
		userData.CreatedAt = time.Now()
		
		logger.Info("User created successfully", map[string]interface{}{
			"user_id":  userData.ID,
			"username": userData.Username,
		}, []string{"users", "success"})
		
		c.JSON(http.StatusCreated, gin.H{"user": userData})
	})
	
	router.GET("/api/error", func(c *gin.Context) {
		logger.Warn("Error endpoint called - simulating error", map[string]interface{}{
			"ip": c.ClientIP(),
		}, []string{"error", "test"})
		
		logger.Error("Application error occurred", map[string]interface{}{
			"error":    "Simulated database connection error",
			"endpoint": "/api/error",
		}, []string{"error", "exception"})
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Simulated database connection error",
		})
	})
	
	router.GET("/api/panic", func(c *gin.Context) {
		logger.Warn("Panic endpoint called", map[string]interface{}{
			"ip": c.ClientIP(),
		}, []string{"panic", "test"})
		
		// This will be caught by middleware recovery
		panic("Simulated panic for testing recovery")
	})
	
	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		logger.Warn("404 Not Found", map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"ip":     c.ClientIP(),
		}, []string{"404", "routing"})
		
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	})
	
	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	logger.Info("Gin server starting", map[string]interface{}{
		"port": port,
		"mode": gin.Mode(),
	}, []string{"startup", "server"})
	
	fmt.Printf("Server running on http://localhost:%s\n", port)
	fmt.Println("DevSmith logging enabled")
	
	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	
	<-sigChan
	logger.Info("Server shutting down - flushing logs", map[string]interface{}{}, []string{"shutdown"})
	logger.Close()
	fmt.Println("Server stopped")
}
