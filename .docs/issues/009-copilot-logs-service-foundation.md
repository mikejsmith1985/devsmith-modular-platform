# Issue #009: [COPILOT] Logs Service - Foundation & Ingestion

**Labels:** `copilot`, `logs`, `websocket`
**Created:** 2025-10-19
**Issue:** #9
**Estimated Time:** 90-120 minutes
**Depends On:** Issue #003 (Portal Auth)

---

# 🚨 STEP 0: CREATE FEATURE BRANCH FIRST 🚨

```bash
git checkout development && git pull origin development
git checkout -b feature/009-copilot-logs-service-foundation
git branch --show-current
```

---

## Task Description

Build Logs Service foundation - ingestion API, storage, basic retrieval. Real-time streaming (WebSocket) comes in Issue #010.

---

## Success Criteria
- [ ] POST /api/logs endpoint accepts log entries
- [ ] Logs stored in logs.log_entries table
- [ ] GET /api/logs retrieves logs with filters (tag, level, service)
- [ ] Pagination support
- [ ] Health check endpoint
- [ ] 70%+ test coverage

---

## ⚠️ CRITICAL: Test-Driven Development (TDD) Required

**YOU MUST WRITE TESTS FIRST, THEN IMPLEMENTATION.**

Follow the Red-Green-Refactor cycle from DevsmithTDD.md.

### TDD Workflow for This Issue

**Step 1: RED PHASE (Write Failing Tests) - DO THIS FIRST!**

Create test files BEFORE implementation:

```go
// internal/logs/db/log_repository_test.go
package db

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestLogRepository_Create_Success(t *testing.T) {
	// Test creating log entry
	repo := NewLogRepository(testDB)
	entry := &models.LogEntry{
		UserID:  1,
		Service: "portal",
		Level:   "info",
		Message: "Test log",
		Tags:    []string{"test"},
	}

	err := repo.Create(context.Background(), entry)

	assert.NoError(t, err)
	assert.NotZero(t, entry.ID)
}

func TestLogRepository_Find_WithFilters(t *testing.T) {
	// Test retrieving with filters
	repo := NewLogRepository(testDB)

	entries, err := repo.Find(context.Background(), LogFilters{
		Service: "portal",
		Level:   "error",
	})

	assert.NoError(t, err)
	// All entries should match filters
	for _, entry := range entries {
		assert.Equal(t, "portal", entry.Service)
		assert.Equal(t, "error", entry.Level)
	}
}

func TestLogRepository_Find_Pagination(t *testing.T) {
	// Test pagination works correctly
	repo := NewLogRepository(testDB)

	page1, _ := repo.Find(ctx, LogFilters{Limit: 10, Offset: 0})
	page2, _ := repo.Find(ctx, LogFilters{Limit: 10, Offset: 10})

	assert.Len(t, page1, 10)
	assert.Len(t, page2, 10)
	assert.NotEqual(t, page1[0].ID, page2[0].ID)
}
```

**Run tests (should FAIL):**
```bash
go test ./internal/logs/...
# Expected: FAIL - NewLogRepository undefined
```

**Commit failing tests:**
```bash
git add internal/logs/db/log_repository_test.go
git commit -m "test(logs): add failing tests for log repository (RED phase)"
```

**Step 2: GREEN PHASE - Implement to Pass Tests**

Now implement `log_repository.go`. See Implementation section below.

**Step 3: Verify Build**
```bash
go build -o /dev/null ./cmd/logs
```

**Step 4: Commit Implementation**
```bash
git add internal/logs/db/log_repository.go
git commit -m "feat(logs): implement log repository (GREEN phase)"
```

**Reference:** DevsmithTDD.md lines 15-36

---

## Implementation

**IMPORTANT: Follow TDD workflow above. Write tests FIRST (shown above), then implement.**

### Phase 1: Database Migrations

**File:** `internal/logs/db/migrations/20251019_001_create_logs_schema.sql`
```sql
CREATE SCHEMA IF NOT EXISTS logs;

CREATE TABLE IF NOT EXISTS logs.log_entries (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,  -- Who generated the log
    service VARCHAR(50) NOT NULL,  -- portal, review, logs, analytics
    level VARCHAR(20) NOT NULL,  -- debug, info, warn, error, fatal
    message TEXT NOT NULL,
    tags TEXT[],  -- Array of tags for filtering
    metadata JSONB,  -- Additional context
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_logs_user_id ON logs.log_entries(user_id);
CREATE INDEX idx_logs_service ON logs.log_entries(service);
CREATE INDEX idx_logs_level ON logs.log_entries(level);
CREATE INDEX idx_logs_created_at ON logs.log_entries(created_at DESC);
CREATE INDEX idx_logs_tags ON logs.log_entries USING gin(tags);
CREATE INDEX idx_logs_metadata ON logs.log_entries USING gin(metadata);
```

**Commit:** `git add internal/logs/db/migrations/ && git commit -m "feat(logs): add database migrations"`

---

### Phase 2: Models

**File:** `internal/logs/models/log.go`
```go
package models

import "time"

type LogEntry struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Service   string    `json:"service" db:"service"`
	Level     string    `json:"level" db:"level"`
	Message   string    `json:"message" db:"message"`
	Tags      []string  `json:"tags" db:"tags"`
	Metadata  string    `json:"metadata" db:"metadata"`  // JSON string
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

**Commit:** `git add internal/logs/models/ && git commit -m "feat(logs): add LogEntry model"`

---

### Phase 3: Repository

**File:** `internal/logs/db/log_repository.go`
```go
package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

type LogRepository struct {
	db *pgxpool.Pool
}

func NewLogRepository(db *pgxpool.Pool) *LogRepository {
	return &LogRepository{db: db}
}

func (r *LogRepository) Create(ctx context.Context, log *models.LogEntry) error {
	query := `
		INSERT INTO logs.log_entries (user_id, service, level, message, tags, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, log.UserID, log.Service, log.Level, log.Message, log.Tags, log.Metadata).
		Scan(&log.ID, &log.CreatedAt)
}

func (r *LogRepository) FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.LogEntry, error) {
	query := `SELECT id, user_id, service, level, message, tags, metadata, created_at
	          FROM logs.log_entries
	          WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if service, ok := filters["service"]; ok {
		query += fmt.Sprintf(" AND service = $%d", argIndex)
		args = append(args, service)
		argIndex++
	}

	if level, ok := filters["level"]; ok {
		query += fmt.Sprintf(" AND level = $%d", argIndex)
		args = append(args, level)
		argIndex++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.LogEntry
	for rows.Next() {
		log := &models.LogEntry{}
		rows.Scan(&log.ID, &log.UserID, &log.Service, &log.Level, &log.Message, &log.Tags, &log.Metadata, &log.CreatedAt)
		logs = append(logs, log)
	}
	return logs, nil
}
```

**Commit:** `git add internal/logs/db/ && git commit -m "feat(logs): add log repository"`

---

### Phase 4: Service Layer

**File:** `internal/logs/services/log_service.go`
```go
package services

import (
	"context"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

type LogService struct {
	logRepo LogRepositoryInterface
}

type LogRepositoryInterface interface {
	Create(ctx context.Context, log *models.LogEntry) error
	FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.LogEntry, error)
}

func NewLogService(logRepo LogRepositoryInterface) *LogService {
	return &LogService{logRepo: logRepo}
}

func (s *LogService) IngestLog(ctx context.Context, log *models.LogEntry) error {
	return s.logRepo.Create(ctx, log)
}

func (s *LogService) QueryLogs(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.LogEntry, error) {
	return s.logRepo.FindAll(ctx, filters, limit, offset)
}
```

**Commit:** `git add internal/logs/services/ && git commit -m "feat(logs): add log service"`

---

### Phase 5: Handlers

**File:** `cmd/logs/handlers/log_handler.go`
```go
package handlers

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

type LogHandler struct {
	logService *services.LogService
}

func NewLogHandler(logService *services.LogService) *LogHandler {
	return &LogHandler{logService: logService}
}

func (h *LogHandler) IngestLog(c *gin.Context) {
	var log models.LogEntry
	if err := c.ShouldBindJSON(&log); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	log.UserID = userID.(int64)

	if err := h.logService.IngestLog(c.Request.Context(), &log); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, log)
}

func (h *LogHandler) QueryLogs(c *gin.Context) {
	filters := map[string]interface{}{}
	if service := c.Query("service"); service != "" {
		filters["service"] = service
	}
	if level := c.Query("level"); level != "" {
		filters["level"] = level
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, _ := h.logService.QueryLogs(c.Request.Context(), filters, limit, offset)
	c.JSON(http.StatusOK, logs)
}
```

**Commit:** `git add cmd/logs/handlers/ && git commit -m "feat(logs): add log handlers"`

---

### Phase 6: Main Entry Point

**File:** `cmd/logs/main.go`
```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikejsmith1985/devsmith-modular-platform/cmd/logs/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	dbPool, _ := pgxpool.New(context.Background(), databaseURL)
	defer dbPool.Close()

	logRepo := db.NewLogRepository(dbPool)
	logService := services.NewLogService(logRepo)
	logHandler := handlers.NewLogHandler(logService)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "logs", "status": "healthy"})
	})

	api := router.Group("/api")
	{
		api.POST("/logs", logHandler.IngestLog)
		api.GET("/logs", logHandler.QueryLogs)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	fmt.Printf("Logs service starting on port %s...\\n", port)
	router.Run(":" + port)
}
```

**Commit:** `git add cmd/logs/main.go && git commit -m "feat(logs): add service entry point"`

---

### Phase 7: Dockerfile

**File:** `cmd/logs/Dockerfile`
```dockerfile
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/logs ./cmd/logs

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata wget
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
WORKDIR /home/appuser
COPY --from=builder /app/bin/logs ./logs
RUN chown -R appuser:appuser /home/appuser
USER appuser
EXPOSE 8082
HEALTHCHECK --interval=30s --timeout=3s CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1
CMD ["./logs"]
```

**Commit:** `git add cmd/logs/Dockerfile && git commit -m "feat(logs): add Dockerfile"`

---

### Phase 8: Update docker-compose.yml

Add logs service to `docker-compose.yml`:
```yaml
  logs:
    build:
      context: .
      dockerfile: cmd/logs/Dockerfile
    container_name: devsmith-logs
    environment:
      - PORT=8082
      - DATABASE_URL=postgres://devsmith:${DB_PASSWORD:-devsmith_local}@postgres:5432/devsmith
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8082/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    networks:
      - devsmith-network
```

Update nginx config to route `/logs`:
```nginx
location /logs {
    proxy_pass http://logs:8082;
    proxy_set_header Host $host;
}
```

**Commit:** `git add docker-compose.yml docker/nginx/nginx.conf && git commit -m "feat(logs): add to docker-compose and nginx"`

---

### Phase 9: Push

```bash
git push -u origin feature/009-copilot-logs-service-foundation
```

---

## References
- `ARCHITECTURE.md` lines 1126-1145 (Logs Service spec)

**Time:** 90-120 minutes
