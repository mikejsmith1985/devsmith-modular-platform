# Docker Configuration Guide for AI Assistants

> **For:** GitHub Copilot, Claude Code, and other AI coding assistants
> **Purpose:** Ensure Docker configurations are production-ready and validated

---

## Critical Rules

### 🚨 ALWAYS Follow These Rules

1. **Health checks are MANDATORY** for all HTTP services
2. **Use `condition: service_healthy`** in all `depends_on` declarations
3. **Implement `/health` endpoints** in all Go services
4. **Add `start_period`** to all health checks (minimum 40s for services with DB)
5. **Validate with `./scripts/docker-validate.sh`** after any Docker changes

### ❌ NEVER Do These

1. ❌ NEVER create a service without a health check
2. ❌ NEVER use `depends_on: - service` without `condition: service_healthy`
3. ❌ NEVER skip the `/health` endpoint implementation
4. ❌ NEVER commit Docker changes without running validation
5. ❌ NEVER use `HEALTHCHECK NONE` unless explicitly required

---

## Standard Patterns

### Pattern 1: New Go HTTP Service

When creating a new Go HTTP service, follow this EXACT pattern:

#### 1. Service Code with Health Endpoint

```go
// cmd/[service]/main.go
package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "os"

    _ "github.com/lib/pq"
)

var db *sql.DB

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    // Initialize database
    dbURL := os.Getenv("DATABASE_URL")
    var err error
    db, err = sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

    // Verify connection
    if err := db.Ping(); err != nil {
        log.Fatal("Failed to ping database:", err)
    }

    // Register handlers
    http.HandleFunc("/health", healthHandler)
    // ... other handlers ...

    log.Printf("Starting %s service on port %s", "[SERVICE]", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

// REQUIRED: Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
    // Check database connectivity
    if err := db.Ping(); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
            "checks": map[string]bool{
                "database": false,
            },
        })
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "healthy",
        "checks": map[string]bool{
            "database": true,
        },
    })
}
```

#### 2. Dockerfile

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/[service] ./cmd/[service]

# Runtime stage
FROM alpine:latest

# Install required tools for health checks
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN adduser -D -u 1000 appuser

# Copy binary from builder
COPY --from=builder /app/[service] /[service]

# Set ownership
RUN chown appuser:appuser /[service]

USER appuser

# Expose port
EXPOSE [PORT]

# REQUIRED: Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:[PORT]/health || exit 1

CMD ["/[service]"]
```

#### 3. docker-compose.yml Entry

```yaml
services:
  [service]:
    build:
      context: .
      dockerfile: cmd/[service]/Dockerfile
    ports:
      - "[HOST_PORT]:[CONTAINER_PORT]"
    environment:
      - PORT=[CONTAINER_PORT]
      - DATABASE_URL=postgres://devsmith:devsmith@postgres:5432/devsmith?sslmode=disable
    depends_on:
      postgres:
        condition: service_healthy  # REQUIRED: Wait for DB
    networks:
      - devsmith-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:[CONTAINER_PORT]/health"]
      interval: 30s
      timeout: 10s
      retries: 5
      start_period: 40s  # REQUIRED: Allow time for DB connection + initialization
```

#### 4. Update nginx.conf (if proxied)

```nginx
# Add upstream
upstream [service] {
    server [service]:[CONTAINER_PORT];
}

# Add location block
location /[service]/ {
    proxy_pass http://[service]/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

#### 5. Update docker-validate.sh

```bash
# Add to SERVICES array (line ~40)
declare -A SERVICES=(
    [postgres]="5432"
    [portal]="8080"
    # ... existing services ...
    [[service]]="[PORT]"  # Add this line
)

# Add to ENDPOINTS array (line ~50)
declare -A ENDPOINTS=(
    [portal]="http://localhost:8080/health"
    # ... existing endpoints ...
    [[service]]="http://localhost:[HOST_PORT]/health"  # Add this line
)
```

#### 6. Validate

```bash
# Start services
docker-compose up -d --build

# Run validation
./scripts/docker-validate.sh --wait --max-wait 120

# Check logs if validation fails
docker-compose logs [service]
```

---

## Pattern 2: Database-Dependent Service

Services that require database connectivity MUST:

### 1. Wait for Database

```yaml
depends_on:
  postgres:
    condition: service_healthy  # MANDATORY
```

### 2. Verify Connection in Health Check

```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    // MANDATORY: Check DB connectivity
    if err := db.Ping(); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "unhealthy",
            "error":  err.Error(),
        })
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
```

### 3. Use Adequate start_period

```yaml
healthcheck:
  # ... other settings ...
  start_period: 40s  # Minimum 40s for DB-dependent services
```

**Reasoning:** Service needs time to:
- Boot up (5-10s)
- Connect to database (5-10s)
- Run migrations if applicable (10-20s)
- Initialize application state (5-10s)

---

## Pattern 3: Database Service (PostgreSQL)

PostgreSQL service MUST have health check:

```yaml
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: devsmith
      POSTGRES_PASSWORD: devsmith
      POSTGRES_DB: devsmith
    volumes:
      - ./docker/postgres/init-schemas.sql:/docker-entrypoint-initdb.d/init-schemas.sql
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    networks:
      - devsmith-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U devsmith -d devsmith"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 10s  # DB initializes quickly
```

**Key points:**
- Use `pg_isready` with username and database
- Shorter intervals (10s) for faster dependency startup
- Short start_period (10s) - databases initialize quickly

---

## Pattern 4: Nginx Gateway

Nginx MUST wait for all backend services:

```yaml
services:
  nginx:
    image: nginx:latest
    volumes:
      - ./docker/nginx/nginx.conf:/etc/nginx/nginx.conf
    ports:
      - "3000:80"
    depends_on:
      portal:
        condition: service_healthy  # REQUIRED
      review:
        condition: service_healthy  # REQUIRED
      logs:
        condition: service_healthy  # REQUIRED
      analytics:
        condition: service_healthy  # REQUIRED
    networks:
      - devsmith-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:80/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s  # Nginx starts fast
```

**Key points:**
- List ALL backend services in depends_on
- Each with `condition: service_healthy`
- Nginx starts quickly (10s start_period is enough)

---

## Common Mistakes & Fixes

### Mistake 1: Health Check Returns 404

**Problem:**
```bash
[http_404] portal - Endpoint http://localhost:8080/health returned 404 Not Found
```

**Causes & Fixes:**

1. ❌ Forgot to implement `/health` endpoint
   ✅ Add `http.HandleFunc("/health", healthHandler)` to main.go

2. ❌ Route registered at wrong path (e.g., `/api/health`)
   ✅ Use exactly `/health` (no prefix)

3. ❌ Mux/router not including the route
   ✅ Ensure router includes health endpoint

### Mistake 2: Health Check Returns 500

**Problem:**
```bash
[http_5xx] portal - Endpoint http://localhost:8080/health returned 500 (server error)
```

**Causes & Fixes:**

1. ❌ Database connection failed but returned 200
   ✅ Check `db.Ping()` and return 503 on error

2. ❌ Panic in health handler
   ✅ Add error handling and recover

3. ❌ Missing environment variable (DATABASE_URL)
   ✅ Verify env vars in docker-compose.yml

### Mistake 3: Container Unhealthy Forever

**Problem:**
```bash
[health_starting] portal - Health check still starting
```

**Causes & Fixes:**

1. ❌ `start_period` too short
   ✅ Increase to 40s for DB-dependent services

2. ❌ Database not ready when service starts
   ✅ Add `depends_on: postgres: condition: service_healthy`

3. ❌ Service crashing during startup
   ✅ Check logs: `docker-compose logs [service]`

### Mistake 4: Port Binding Issues

**Problem:**
```bash
[port_not_bound] portal - Port 8080 is not bound to host
```

**Causes & Fixes:**

1. ❌ Missing `ports:` section
   ✅ Add `ports: - "8080:8080"`

2. ❌ Wrong port mapping (e.g., `- "80:8080"` when should be `- "8080:8080"`)
   ✅ Verify both ports match expected values

3. ❌ Port already in use
   ✅ Check: `docker ps` and `lsof -i :8080`

### Mistake 5: Nginx 502 Bad Gateway

**Problem:**
```bash
Nginx returns 502 when accessing http://localhost:3000/portal/
```

**Causes & Fixes:**

1. ❌ Upstream service not healthy
   ✅ Verify: `docker-compose ps portal` shows "healthy"

2. ❌ Upstream definition uses wrong service name
   ✅ Use docker-compose service name: `server portal:8080;`

3. ❌ Nginx started before backends
   ✅ Add `depends_on` with `service_healthy` for all backends

---

## Validation Checklist

Before committing Docker changes, verify:

- [ ] All services have health checks in docker-compose.yml
- [ ] All services have HEALTHCHECK in Dockerfile
- [ ] All services implement `/health` endpoint (if HTTP)
- [ ] Health endpoints check database connectivity (if applicable)
- [ ] All `depends_on` use `condition: service_healthy`
- [ ] `start_period` is adequate (40s for DB services, 10s for others)
- [ ] Ports are correctly mapped in docker-compose.yml
- [ ] Service added to `docker-validate.sh` arrays
- [ ] Nginx upstream configured (if proxied)
- [ ] Nginx depends_on includes new service
- [ ] `./scripts/docker-validate.sh --wait` passes

---

## Testing Your Changes

### Step 1: Clean Start

```bash
# Stop and remove everything
docker-compose down -v

# Rebuild and start
docker-compose up -d --build
```

### Step 2: Validate

```bash
# Wait for services to be healthy
./scripts/docker-validate.sh --wait --max-wait 120
```

**Expected output:**
```
✅ Docker validation PASSED
```

### Step 3: Manual Verification

```bash
# Check container status
docker-compose ps

# All services should show (healthy)
```

```bash
# Test endpoints manually
curl http://localhost:8080/health  # Should return 200 OK
curl http://localhost:3000/        # Gateway should work
```

### Step 4: Check Logs

```bash
# Verify no errors in logs
docker-compose logs | grep -i error
```

---

## Quick Reference

### Essential Commands

```bash
# Start with validation
./scripts/dev.sh

# Validate manually
./scripts/docker-validate.sh

# Validate with wait
./scripts/docker-validate.sh --wait --max-wait 120

# Auto-restart unhealthy
./scripts/docker-validate.sh --auto-restart

# JSON output
./scripts/docker-validate.sh --json

# Quick validation (no HTTP checks)
./scripts/docker-validate.sh --quick

# Thorough validation (all checks)
./scripts/docker-validate.sh --thorough

# View logs
docker-compose logs [service]
docker-compose logs -f  # Follow all logs

# Check status
docker-compose ps

# Restart service
docker-compose restart [service]

# Rebuild service
docker-compose up -d --build [service]

# Clean restart
docker-compose down && docker-compose up -d --build
```

### Health Endpoint Template

```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    // Create response structure
    health := map[string]interface{}{
        "status": "healthy",
        "checks": make(map[string]bool),
    }

    healthy := true

    // Check database
    if db != nil {
        if err := db.Ping(); err != nil {
            healthy = false
            health["checks"].(map[string]bool)["database"] = false
            health["error"] = err.Error()
        } else {
            health["checks"].(map[string]bool)["database"] = true
        }
    }

    // Check other dependencies here...

    // Set status and return
    if !healthy {
        health["status"] = "unhealthy"
        w.WriteHeader(http.StatusServiceUnavailable)
    } else {
        w.WriteHeader(http.StatusOK)
    }

    json.NewEncoder(w).Encode(health)
}
```

---

## Integration with AI Workflow

### When Creating a New Service

1. **Generate service code** with health endpoint
2. **Create Dockerfile** with HEALTHCHECK
3. **Add to docker-compose.yml** with health check and proper depends_on
4. **Update nginx.conf** if service is proxied
5. **Update docker-validate.sh** arrays
6. **Run validation** to verify

### When Modifying Existing Service

1. **Make changes** to code/config
2. **Rebuild**: `docker-compose up -d --build [service]`
3. **Validate**: `./scripts/docker-validate.sh --wait`
4. **Check logs** if validation fails

### When Troubleshooting

1. **Run validation**: `./scripts/docker-validate.sh`
2. **Check logs**: `docker-compose logs [service]`
3. **Verify health endpoint**: `curl http://localhost:PORT/health`
4. **Check dependencies**: `docker-compose ps`
5. **Review configuration**: Ensure health checks are properly configured

---

## Summary for AI Assistants

**When you create or modify Docker configurations:**

1. ✅ Always implement `/health` endpoints in services
2. ✅ Always add `healthcheck:` to docker-compose.yml
3. ✅ Always use `depends_on: service: condition: service_healthy`
4. ✅ Always add HEALTHCHECK to Dockerfile
5. ✅ Always set adequate `start_period` (40s for DB services)
6. ✅ Always update docker-validate.sh when adding services
7. ✅ Always run `./scripts/docker-validate.sh` to verify
8. ✅ Always check logs if validation fails

**Critical validations to perform:**
- Health endpoint returns 200 OK when healthy
- Health endpoint checks database connectivity
- Container becomes "healthy" within start_period
- HTTP endpoints return expected status codes
- Ports are correctly bound to host

**Documentation to reference:**
- This guide: `.docs/DOCKER-COPILOT-GUIDE.md`
- User guide: `.docs/DOCKER-VALIDATION.md`
- Example services: `cmd/{portal,review,logs,analytics}/main.go`

By following these patterns, you ensure Docker configurations are robust, validated, and production-ready. The validation system will catch configuration errors before they impact development.
