# Issue #38: Log Context & Correlation

**Labels:** `copilot`, `logs`, `feature`
**Assignee:** Mike (with Copilot assistance)
**Created:** 2025-10-26
**Issue:** #38
**Estimated Complexity:** Medium
**Target Service:** Logs
**Estimated Time:** 120-150 minutes
**Depends On:** Issue #009 (Logs Service Foundation)

---

# 🚨 CRITICAL: FIRST STEP - FEATURE BRANCH ✅

## STEP 0: Verify Feature Branch

✅ **Already completed:**
```bash
git branch --show-current
# Output: feature/038-log-context-correlation
```

---

# ⚠️ READ THIS BEFORE CODING ⚠️

**DO NOT work on the `development` branch directly!**

**Workflow Order (MANDATORY):**
1. ✅ Feature branch created
2. ✅ Read this entire spec
3. → **Write tests FIRST (RED phase - 100% complete)**
4. → **Implement code (GREEN phase - 100% complete)**
5. → **Improve quality (REFACTOR phase - 100% complete)**
6. → Commit after EACH phase
7. → Create PR with "closes #38"

---

## Overview

### Feature Description

Add correlation ID and request context tracking to the Logs service. This enables tracing requests across multiple services and correlating all related logs. Critical for debugging distributed system issues and understanding request flows through the platform.

### User Story

As a developer, I want to see all logs related to a specific request so that I can trace request flows across services and debug issues efficiently.

### Success Criteria

- [ ] Correlation ID generated and propagated via middleware
- [ ] Request context (user_id, session_id, request_id) stored in log entries
- [ ] GET /api/logs?correlation_id=xyz shows all related logs
- [ ] UI groups logs by correlation ID
- [ ] OpenTelemetry trace ID support integrated
- [ ] Automatic context enrichment (env, host, version)
- [ ] 70%+ unit test coverage for context propagation
- [ ] Integration tests for cross-service correlation
- [ ] All tests passing

---

## Context for Cognitive Load Management

### Bounded Context

**Service:** Logs Service
**Domain:** Request Tracing and Context Propagation
**Related Entities:**
- `LogEntry` - Individual log message (already exists)
- `CorrelationContext` - Request context metadata
- `ContextMiddleware` - HTTP middleware for propagation

**Context Boundaries:**
- ✅ **Within scope:** Generating/propagating correlation IDs, storing context in logs, querying by correlation ID
- ❌ **Out of scope:** External tracing infrastructure (OpenTelemetry collection), application-specific business logic

**Why This Separation:**
Context is a cross-cutting concern. Logs service provides the collection point. Each service uses the context middleware to propagate correlation IDs.

---

### Layering

**All three layers required:**

#### Controller/Middleware Layer
```
cmd/logs/middleware/
├── correlation_middleware.go
├── correlation_middleware_test.go
```

#### Orchestration Layer
```
internal/logs/services/
├── context_service.go
├── context_service_test.go
```

#### Data Layer
```
internal/logs/db/
├── context_repository.go
├── context_repository_test.go
├── migrations/
    └── 20251026000000_add_correlation_context_to_logs.sql
```

**Cross-Layer Rules:**
- ✅ Middleware calls services
- ✅ Services call repositories
- ❌ Middleware MUST NOT call repositories directly
- ❌ Repositories MUST NOT know about HTTP

---

## Implementation Specification

### Phase 1: Data Models & Database

**Create correlation context types and database schema.**

#### 1.1 Create Models

**File:** `internal/logs/models/context.go`

```go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// CorrelationContext stores request context for tracing
type CorrelationContext struct {
	// Identifiers
	CorrelationID  string `json:"correlation_id"`   // Unique per request
	TraceID        string `json:"trace_id"`         // OpenTelemetry trace ID
	SpanID         string `json:"span_id"`          // OpenTelemetry span ID
	RequestID      string `json:"request_id"`       // HTTP request ID
	
	// User Context
	UserID         *int   `json:"user_id,omitempty"`
	SessionID      string `json:"session_id,omitempty"`
	
	// Service Context
	Service        string `json:"service"`          // Service that generated log
	Hostname       string `json:"hostname"`         // Server hostname
	Environment    string `json:"environment"`      // dev, staging, prod
	Version        string `json:"version"`          // Service version
	
	// Request Context
	Method         string `json:"method,omitempty"` // HTTP method
	Path           string `json:"path,omitempty"`   // HTTP path
	RemoteAddr     string `json:"remote_addr,omitempty"`
	
	// Timing
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Value implements driver.Valuer for database storage
func (cc CorrelationContext) Value() (driver.Value, error) {
	return json.Marshal(cc)
}

// Scan implements sql.Scanner for database retrieval
func (cc *CorrelationContext) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion failed")
	}
	return json.Unmarshal(bytes, &cc)
}
```

**File:** `internal/logs/models/log_entry.go` (UPDATE existing)

Add to LogEntry struct:
```go
type LogEntry struct {
	// ... existing fields ...
	
	// Correlation context (NEW)
	Context *CorrelationContext `json:"context,omitempty"`
	
	// ... rest of fields ...
}
```

#### 1.2 Database Migration

**File:** `internal/logs/db/migrations/20251026000000_add_correlation_context_to_logs.sql`

```sql
-- Add correlation context to log_entries table
ALTER TABLE logs.log_entries ADD COLUMN context JSONB;
ALTER TABLE logs.log_entries ADD COLUMN correlation_id TEXT;

-- Create index for faster correlation ID queries
CREATE INDEX idx_log_entries_correlation_id ON logs.log_entries(correlation_id);

-- Create partial index for active traces (last 24 hours)
CREATE INDEX idx_log_entries_correlation_recent 
ON logs.log_entries(correlation_id) 
WHERE created_at > NOW() - INTERVAL '24 hours';

-- Create GIN index for JSON context queries
CREATE INDEX idx_log_entries_context_gin ON logs.log_entries USING GIN(context);

-- View for correlated logs
CREATE OR REPLACE VIEW logs.v_correlated_logs AS
SELECT 
    le.id,
    le.timestamp,
    le.level,
    le.message,
    le.service,
    le.context->>'correlation_id' as correlation_id,
    le.context->>'user_id' as user_id,
    le.context->>'trace_id' as trace_id,
    le.context
FROM logs.log_entries le
WHERE le.context IS NOT NULL
ORDER BY le.timestamp DESC;
```

**Commit after Phase 1:**
```bash
git add internal/logs/models/ internal/logs/db/migrations/
git commit -m "feat(logs): add correlation context models and migrations (Phase 1)"
```

---

### Phase 2: Repository Layer

**File:** `internal/logs/db/context_repository.go`

```go
package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
	
	"devsmith/internal/logs/models"
)

// ContextRepository handles correlation context persistence
type ContextRepository struct {
	db *sql.DB
}

// NewContextRepository creates a new context repository
func NewContextRepository(db *sql.DB) *ContextRepository {
	return &ContextRepository{db: db}
}

// GetCorrelatedLogs retrieves all logs for a correlation ID
func (r *ContextRepository) GetCorrelatedLogs(
	ctx context.Context,
	correlationID string,
	limit int,
	offset int,
) ([]models.LogEntry, error) {
	if correlationID == "" {
		return nil, errors.New("correlation_id required")
	}
	
	query := `
		SELECT id, timestamp, level, message, service, context, created_at
		FROM logs.log_entries
		WHERE correlation_id = $1 OR context->>'correlation_id' = $2
		ORDER BY timestamp DESC
		LIMIT $3 OFFSET $4
	`
	
	rows, err := r.db.QueryContext(ctx, query, correlationID, correlationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var logs []models.LogEntry
	for rows.Next() {
		var log models.LogEntry
		var contextJSON sql.NullString
		
		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.Level,
			&log.Message,
			&log.Service,
			&contextJSON,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Parse context if present
		if contextJSON.Valid {
			ctx := &models.CorrelationContext{}
			if err := ctx.Scan([]byte(contextJSON.String)); err != nil {
				return nil, err
			}
			log.Context = ctx
		}
		
		logs = append(logs, log)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return logs, nil
}

// GetCorrelationCount returns count of logs for a correlation ID
func (r *ContextRepository) GetCorrelationCount(
	ctx context.Context,
	correlationID string,
) (int, error) {
	if correlationID == "" {
		return 0, errors.New("correlation_id required")
	}
	
	query := `
		SELECT COUNT(*)
		FROM logs.log_entries
		WHERE correlation_id = $1 OR context->>'correlation_id' = $2
	`
	
	var count int
	err := r.db.QueryRowContext(ctx, query, correlationID, correlationID).Scan(&count)
	return count, err
}

// GetRecentCorrelations returns active correlation IDs from last N minutes
func (r *ContextRepository) GetRecentCorrelations(
	ctx context.Context,
	minutes int,
	limit int,
) ([]string, error) {
	query := `
		SELECT DISTINCT correlation_id
		FROM logs.log_entries
		WHERE created_at > NOW() - INTERVAL '1 minute' * $1
		  AND correlation_id IS NOT NULL
		ORDER BY created_at DESC
		LIMIT $2
	`
	
	rows, err := r.db.QueryContext(ctx, query, minutes, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var correlationIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		correlationIDs = append(correlationIDs, id)
	}
	
	return correlationIDs, rows.Err()
}

// GetContextMetadata retrieves metadata for a correlation
func (r *ContextRepository) GetContextMetadata(
	ctx context.Context,
	correlationID string,
) (map[string]interface{}, error) {
	query := `
		SELECT jsonb_object_agg(key, value) as metadata
		FROM (
			SELECT DISTINCT
				context->>'service' as key,
				context->>'trace_id' as value
			FROM logs.log_entries
			WHERE correlation_id = $1 OR context->>'correlation_id' = $2
		) t
	`
	
	var metadataJSON sql.NullString
	err := r.db.QueryRowContext(ctx, query, correlationID, correlationID).
		Scan(&metadataJSON)
	if err != nil || !metadataJSON.Valid {
		return make(map[string]interface{}), err
	}
	
	// Parse metadata
	metadata := make(map[string]interface{})
	if err := json.Unmarshal([]byte(metadataJSON.String), &metadata); err != nil {
		return nil, err
	}
	
	return metadata, nil
}
```

**Commit after Phase 2:**
```bash
git add internal/logs/db/context_repository.go
git commit -m "feat(logs): implement context repository layer (Phase 2)"
```

---

### Phase 3: Service Layer

**File:** `internal/logs/services/context_service.go`

```go
package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"
	
	"devsmith/internal/logs/db"
	"devsmith/internal/logs/models"
)

// ContextService manages correlation context
type ContextService struct {
	repo *db.ContextRepository
}

// NewContextService creates a new context service
func NewContextService(repo *db.ContextRepository) *ContextService {
	return &ContextService{repo: repo}
}

// GenerateCorrelationID creates a new unique correlation ID
func (s *ContextService) GenerateCorrelationID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// EnrichContext adds automatic metadata to context
func (s *ContextService) EnrichContext(
	ctx *models.CorrelationContext,
) *models.CorrelationContext {
	if ctx == nil {
		ctx = &models.CorrelationContext{}
	}
	
	// Generate correlation ID if missing
	if ctx.CorrelationID == "" {
		ctx.CorrelationID = s.GenerateCorrelationID()
	}
	
	// Add automatic enrichment
	if ctx.Hostname == "" {
		if host, err := os.Hostname(); err == nil {
			ctx.Hostname = host
		}
	}
	
	if ctx.Environment == "" {
		ctx.Environment = os.Getenv("ENVIRONMENT")
		if ctx.Environment == "" {
			ctx.Environment = "development"
		}
	}
	
	if ctx.Version == "" {
		ctx.Version = os.Getenv("SERVICE_VERSION")
		if ctx.Version == "" {
			ctx.Version = "dev"
		}
	}
	
	// Set timestamps
	now := time.Now()
	if ctx.CreatedAt.IsZero() {
		ctx.CreatedAt = now
	}
	ctx.UpdatedAt = now
	
	return ctx
}

// GetCorrelatedLogs retrieves all logs for a correlation ID
func (s *ContextService) GetCorrelatedLogs(
	ctx context.Context,
	correlationID string,
	limit int,
	offset int,
) ([]models.LogEntry, error) {
	// Validate limit (max 1000)
	if limit > 1000 {
		limit = 1000
	}
	if limit <= 0 {
		limit = 50
	}
	
	return s.repo.GetCorrelatedLogs(ctx, correlationID, limit, offset)
}

// GetCorrelationMetadata returns summary of a correlation
func (s *ContextService) GetCorrelationMetadata(
	ctx context.Context,
	correlationID string,
) (map[string]interface{}, error) {
	count, err := s.repo.GetCorrelationCount(ctx, correlationID)
	if err != nil {
		return nil, err
	}
	
	metadata, err := s.repo.GetContextMetadata(ctx, correlationID)
	if err != nil {
		return nil, err
	}
	
	metadata["total_logs"] = count
	metadata["correlation_id"] = correlationID
	
	return metadata, nil
}

// GetTraceTimeline returns timeline of events for a correlation
func (s *ContextService) GetTraceTimeline(
	ctx context.Context,
	correlationID string,
) ([]map[string]interface{}, error) {
	logs, err := s.repo.GetCorrelatedLogs(ctx, correlationID, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	timeline := make([]map[string]interface{}, 0, len(logs))
	for _, log := range logs {
		entry := map[string]interface{}{
			"timestamp": log.Timestamp,
			"level":     log.Level,
			"service":   log.Service,
			"message":   log.Message,
		}
		if log.Context != nil {
			entry["trace_id"] = log.Context.TraceID
			entry["span_id"] = log.Context.SpanID
		}
		timeline = append(timeline, entry)
	}
	
	return timeline, nil
}
```

**Commit after Phase 3:**
```bash
git add internal/logs/services/context_service.go
git commit -m "feat(logs): implement context service layer (Phase 3)"
```

---

### Phase 4: Middleware

**File:** `cmd/logs/middleware/correlation_middleware.go`

```go
package middleware

import (
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
	"devsmith/internal/logs/models"
	"devsmith/internal/logs/services"
)

// CorrelationMiddleware adds correlation context to requests
func CorrelationMiddleware(contextService *services.ContextService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract or generate correlation ID
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = contextService.GenerateCorrelationID()
		}
		
		// Extract trace ID (OpenTelemetry)
		traceID := c.GetHeader("traceparent") // W3C Trace Context
		if traceID == "" {
			traceID = c.GetHeader("X-Trace-ID") // Custom header
		}
		
		// Extract OpenTelemetry span ID
		spanID := extractSpanID(traceID)
		
		// Build correlation context
		ctx := &models.CorrelationContext{
			CorrelationID: correlationID,
			TraceID:       traceID,
			SpanID:        spanID,
			RequestID:     c.GetHeader("X-Request-ID"),
			Method:        c.Request.Method,
			Path:          c.Request.URL.Path,
			RemoteAddr:    c.ClientIP(),
			Service:       os.Getenv("SERVICE_NAME"),
		}
		
		// Extract user context if authenticated
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(int); ok {
				ctx.UserID = &uid
			}
		}
		
		if sessionID, exists := c.Get("session_id"); exists {
			if sid, ok := sessionID.(string); ok {
				ctx.SessionID = sid
			}
		}
		
		// Enrich with automatic metadata
		ctx = contextService.EnrichContext(ctx)
		
		// Store in request context
		c.Set("correlation_context", ctx)
		
		// Add response headers for tracing
		c.Header("X-Correlation-ID", correlationID)
		if traceID != "" {
			c.Header("X-Trace-ID", traceID)
		}
		
		c.Next()
	}
}

// extractSpanID extracts span ID from W3C traceparent format
// Format: traceparent: version-traceid-spanid-traceflags
func extractSpanID(traceparent string) string {
	if traceparent == "" {
		return ""
	}
	
	parts := strings.Split(traceparent, "-")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

// GetCorrelationContext retrieves correlation context from request
func GetCorrelationContext(c *gin.Context) *models.CorrelationContext {
	if ctx, exists := c.Get("correlation_context"); exists {
		if correlationCtx, ok := ctx.(*models.CorrelationContext); ok {
			return correlationCtx
		}
	}
	
	// Fallback: create minimal context
	return &models.CorrelationContext{
		CorrelationID: c.GetHeader("X-Correlation-ID"),
	}
}
```

**Commit after Phase 4:**
```bash
git add cmd/logs/middleware/correlation_middleware.go
git commit -m "feat(logs): implement correlation middleware (Phase 4)"
```

---

### Phase 5: Handler Endpoints

**File:** `cmd/logs/handlers/context_handlers.go`

```go
package handlers

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"devsmith/cmd/logs/middleware"
	"devsmith/internal/logs/services"
)

// ContextHandlers handles correlation context endpoints
type ContextHandlers struct {
	contextService *services.ContextService
}

// NewContextHandlers creates new context handlers
func NewContextHandlers(contextService *services.ContextService) *ContextHandlers {
	return &ContextHandlers{
		contextService: contextService,
	}
}

// GetCorrelatedLogs returns all logs for a correlation ID
// GET /api/logs?correlation_id=xyz
func (h *ContextHandlers) GetCorrelatedLogs(c *gin.Context) {
	correlationID := c.Query("correlation_id")
	if correlationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "correlation_id parameter required",
		})
		return
	}
	
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	
	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}
	
	logs, err := h.contextService.GetCorrelatedLogs(c.Request.Context(), correlationID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve correlated logs",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"meta": map[string]interface{}{
			"correlation_id": correlationID,
			"count":          len(logs),
			"limit":          limit,
			"offset":         offset,
		},
	})
}

// GetCorrelationMetadata returns metadata for a correlation
// GET /api/logs/correlation/:id
func (h *ContextHandlers) GetCorrelationMetadata(c *gin.Context) {
	correlationID := c.Param("id")
	if correlationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "correlation_id required",
		})
		return
	}
	
	metadata, err := h.contextService.GetCorrelationMetadata(c.Request.Context(), correlationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve metadata",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metadata,
	})
}

// GetTraceTimeline returns timeline of events
// GET /api/logs/trace/:id/timeline
func (h *ContextHandlers) GetTraceTimeline(c *gin.Context) {
	correlationID := c.Param("id")
	if correlationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "correlation_id required",
		})
		return
	}
	
	timeline, err := h.contextService.GetTraceTimeline(c.Request.Context(), correlationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve timeline",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    timeline,
	})
}
```

**Commit after Phase 5:**
```bash
git add cmd/logs/handlers/context_handlers.go
git commit -m "feat(logs): add context handlers (Phase 5)"
```

---

## TDD Testing Strategy - MANDATORY FIRST

### RED PHASE: Write Failing Tests First ✅

**Before implementing ANY code, write comprehensive tests that define expected behavior.**

#### Test 1: Correlation ID Generation
```go
func TestContextService_GenerateCorrelationID_Unique(t *testing.T) {
	service := services.NewContextService(nil)
	
	id1 := service.GenerateCorrelationID()
	id2 := service.GenerateCorrelationID()
	
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 32) // Hex encoded 16 bytes
}
```

#### Test 2: Context Enrichment
```go
func TestContextService_EnrichContext_AddsMetadata(t *testing.T) {
	service := services.NewContextService(nil)
	ctx := &models.CorrelationContext{}
	
	enriched := service.EnrichContext(ctx)
	
	assert.NotEmpty(t, enriched.CorrelationID)
	assert.NotEmpty(t, enriched.Hostname)
	assert.NotEmpty(t, enriched.Environment)
	assert.NotEmpty(t, enriched.Version)
	assert.False(t, enriched.CreatedAt.IsZero())
}
```

#### Test 3: Correlation ID Storage
```go
func TestContextRepository_StoreAndRetrieve(t *testing.T) {
	db := setupTestDB(t)
	repo := db.NewContextRepository(db)
	
	log := &models.LogEntry{
		Level:   "info",
		Message: "Test log",
		Context: &models.CorrelationContext{
			CorrelationID: "test-123",
		},
	}
	
	// Store
	id, err := repo.Create(context.Background(), log)
	assert.NoError(t, err)
	assert.NotZero(t, id)
	
	// Retrieve
	logs, err := repo.GetCorrelatedLogs(context.Background(), "test-123", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "test-123", logs[0].Context.CorrelationID)
}
```

#### Test 4: Middleware Adds Context
```go
func TestCorrelationMiddleware_GeneratesID(t *testing.T) {
	service := services.NewContextService(nil)
	middleware := middleware.CorrelationMiddleware(service)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	
	middleware(c)
	
	ctx := middleware.GetCorrelationContext(c)
	assert.NotNil(t, ctx)
	assert.NotEmpty(t, ctx.CorrelationID)
}
```

#### Test 5: Correlation ID Propagation
```go
func TestCorrelationMiddleware_PropagatesID(t *testing.T) {
	service := services.NewContextService(nil)
	middleware := middleware.CorrelationMiddleware(service)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("X-Correlation-ID", "incoming-123")
	
	middleware(c)
	
	ctx := middleware.GetCorrelationContext(c)
	assert.Equal(t, "incoming-123", ctx.CorrelationID)
	assert.Equal(t, "incoming-123", w.Header().Get("X-Correlation-ID"))
}
```

#### Test 6: OpenTelemetry Trace ID
```go
func TestCorrelationMiddleware_ExtractsTraceID(t *testing.T) {
	service := services.NewContextService(nil)
	middleware := middleware.CorrelationMiddleware(service)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	
	middleware(c)
	
	ctx := middleware.GetCorrelationContext(c)
	assert.NotEmpty(t, ctx.TraceID)
	assert.NotEmpty(t, ctx.SpanID)
}
```

#### Test 7: Handler Returns Correlated Logs
```go
func TestContextHandlers_GetCorrelatedLogs(t *testing.T) {
	// Mock service
	mockService := &MockContextService{}
	mockService.On("GetCorrelatedLogs", mock.Anything, "test-123", 50, 0).
		Return([]models.LogEntry{
			{ID: 1, Message: "Log 1", Level: "info"},
			{ID: 2, Message: "Log 2", Level: "error"},
		}, nil)
	
	handlers := handlers.NewContextHandlers(mockService)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/logs?correlation_id=test-123", nil)
	
	handlers.GetCorrelatedLogs(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, 2, len(response["data"].([]interface{})))
}
```

#### Test 8: Integration - Cross-Service Logging
```go
func TestIntegration_LogsCorrelatedAcrossServices(t *testing.T) {
	// Setup test servers
	logsServer := setupTestLogsServer(t)
	portalServer := setupTestPortalServer(t)
	
	correlationID := "trace-123"
	
	// 1. Portal service makes request with correlation ID
	req, _ := http.NewRequest("POST", portalServer.URL+"/api/auth/login", nil)
	req.Header.Set("X-Correlation-ID", correlationID)
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()
	
	// 2. Query logs service for correlated logs
	logs, _ := http.Get(logsServer.URL + "/api/logs?correlation_id=" + correlationID)
	defer logs.Body.Close()
	
	var result map[string]interface{}
	json.NewDecoder(logs.Body).Decode(&result)
	
	// 3. Verify logs from both services are present
	logsList := result["data"].([]interface{})
	assert.Greater(t, len(logsList), 0)
	
	services := make(map[string]bool)
	for _, log := range logsList {
		logMap := log.(map[string]interface{})
		services[logMap["service"].(string)] = true
	}
	
	assert.True(t, services["portal"])
	assert.True(t, services["logs"])
}
```

---

## Implementation Checklist

### Phase 0: Branch Setup ✅
- [x] Verified on development branch
- [x] Created feature branch: `feature/038-log-context-correlation`
- [x] Verified on feature branch

### Phase 1: Data Models & Database
- [ ] Create `internal/logs/models/context.go`
- [ ] Update `internal/logs/models/log_entry.go`
- [ ] Create migration file
- [ ] Run migration: `go run ./cmd/migrate up`
- [ ] Verify schema: `psql -c "\d logs.log_entries"`
- [ ] Commit: `git add internal/logs/models/ internal/logs/db/migrations/ && git commit -m "feat(logs): add correlation context models and migrations (Phase 1)"`

### Phase 2: Repository Layer
- [ ] Create `internal/logs/db/context_repository.go`
- [ ] Implement `GetCorrelatedLogs()`
- [ ] Implement `GetCorrelationCount()`
- [ ] Implement `GetRecentCorrelations()`
- [ ] Implement `GetContextMetadata()`
- [ ] Unit tests: 75%+ coverage
- [ ] Run: `go test ./internal/logs/db/... -v`
- [ ] Commit: `git add internal/logs/db/context_repository.go && git commit -m "feat(logs): implement context repository layer (Phase 2)"`

### Phase 3: Service Layer
- [ ] Create `internal/logs/services/context_service.go`
- [ ] Implement `GenerateCorrelationID()`
- [ ] Implement `EnrichContext()`
- [ ] Implement `GetCorrelatedLogs()`
- [ ] Implement `GetCorrelationMetadata()`
- [ ] Implement `GetTraceTimeline()`
- [ ] Unit tests: 80%+ coverage
- [ ] Run: `go test ./internal/logs/services/... -v`
- [ ] Commit: `git add internal/logs/services/context_service.go && git commit -m "feat(logs): implement context service layer (Phase 3)"`

### Phase 4: Middleware
- [ ] Create `cmd/logs/middleware/correlation_middleware.go`
- [ ] Implement `CorrelationMiddleware()`
- [ ] Implement `GetCorrelationContext()`
- [ ] Implement `extractSpanID()`
- [ ] Unit tests: 80%+ coverage
- [ ] Run: `go test ./cmd/logs/middleware/... -v`
- [ ] Commit: `git add cmd/logs/middleware/correlation_middleware.go && git commit -m "feat(logs): implement correlation middleware (Phase 4)"`

### Phase 5: Handler Endpoints
- [ ] Create `cmd/logs/handlers/context_handlers.go`
- [ ] Implement `GetCorrelatedLogs()`
- [ ] Implement `GetCorrelationMetadata()`
- [ ] Implement `GetTraceTimeline()`
- [ ] Handler tests: 70%+ coverage
- [ ] Run: `go test ./cmd/logs/handlers/... -v`
- [ ] Commit: `git add cmd/logs/handlers/context_handlers.go && git commit -m "feat(logs): add context handlers (Phase 5)"`

### Phase 6: Integration
- [ ] Update main logs service to register middleware
- [ ] Register handlers in router
- [ ] Integration tests: Cross-service correlation
- [ ] Manual testing through gateway
- [ ] Run: `go test -tags=integration ./tests/integration/... -v`
- [ ] Commit: `git add cmd/logs/main.go && git commit -m "feat(logs): integrate correlation context (Phase 6)"`

### Phase 7: Full Test Suite
- [ ] Run: `go test ./...`
- [ ] Run: `go test -cover ./...` (verify 70%+ coverage)
- [ ] Run: `golangci-lint run ./...`
- [ ] Run: `go fmt ./...`
- [ ] Run: `goimports -w .`
- [ ] Verify: `docker-compose up -d`
- [ ] Test endpoints: Manual verification

### Phase 8: Final Push
- [ ] Review all commits: `git log development..HEAD --oneline`
- [ ] Push: `git push -u origin feature/038-log-context-correlation`
- [ ] Wait for automatic PR creation (GitHub Actions)
- [ ] Verify CI passes on PR
- [ ] Create PR with description including "closes #38"

---

## Environment Variables

Add to `.env.example`:

```bash
# Logs Service - Correlation Context
LOGS_CORRELATION_ID_LENGTH=32     # Length of generated correlation IDs
LOGS_TRACE_RETENTION_HOURS=24     # How long to keep trace logs
LOGS_CONTEXT_ENRICHMENT=true      # Enable automatic context enrichment
```

---

## Testing Requirements

### Unit Tests (70%+ coverage required)
- Models: 80%+ (JSON marshaling, validation)
- Repository: 75%+ (database queries)
- Service: 80%+ (business logic)
- Middleware: 80%+ (context extraction)
- Handlers: 70%+ (HTTP responses)

### Integration Tests
- Correlation ID propagation across services
- OpenTelemetry trace ID extraction
- Cross-service log correlation
- Context persistence in database

### Manual Testing Checklist
- [ ] Single service logging with correlation ID
- [ ] Multi-service request with correlation propagation
- [ ] GET /api/logs?correlation_id=xyz returns all related logs
- [ ] WebSocket receives correlated logs
- [ ] Analytics queries work with correlation ID
- [ ] Trace timeline visualization works

---

## Success Metrics

This issue is complete when:

1. ✅ All database migrations run successfully
2. ✅ Correlation context stored in log entries (JSONB)
3. ✅ GET /api/logs?correlation_id=xyz works
4. ✅ Middleware generates/propagates correlation IDs
5. ✅ OpenTelemetry trace ID support integrated
6. ✅ Automatic context enrichment (env, host, version)
7. ✅ All acceptance criteria met
8. ✅ All unit tests pass with 70%+ coverage
9. ✅ Integration tests for cross-service correlation
10. ✅ No linting errors
11. ✅ CI/CD pipeline passes
12. ✅ PR created with "closes #38"

---

## References

- `ARCHITECTURE.md` - Monitoring & logging (lines 1545-1605)
- `ARCHITECTURE.md` - Mental Models (lines 99-475)
- `.cursorrules` - Budget mode and cost optimization
- `copilot-instructions.md` - TDD workflow and standards
- `DevsmithTDD.md` - Complete TDD guide
- OpenTelemetry Go SDK: https://opentelemetry.io/docs/instrumentation/go/

---

**Next Steps (For Copilot):**
1. ✅ Feature branch created
2. ✅ Read this spec completely
3. → **Write tests FIRST (RED phase)**
4. → Implement code (GREEN phase)
5. → Improve quality (REFACTOR phase)
6. → Commit after each phase
7. → Push regularly
8. → Create PR with "closes #38"

**Estimated Time:** 120-150 minutes
**Test Coverage Target:** 70%+ (aim for 75%+)
**Success Metric:** Correlation IDs enable tracing requests across services
**Depends On:** Issue #009 (Logs Service Foundation)
