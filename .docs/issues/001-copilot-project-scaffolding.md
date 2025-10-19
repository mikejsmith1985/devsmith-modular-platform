# Issue #1: [COPILOT] Project Scaffolding and Configuration

**Labels:** `copilot`, `setup`, `good first issue`
**Assignee:** Mike (with Copilot assistance)
**Estimated Time:** 30-60 minutes
**Complexity:** Low

---

## Task Description

Set up the foundational project structure and configuration files for the DevSmith Modular Platform. This includes creating the Go module structure, Docker Compose configuration, environment templates, and basic project organization.

**Why This Task for Copilot:**
- Repetitive file creation (Copilot excels at boilerplate)
- Standard configuration patterns (Copilot knows Go/Docker conventions)
- Low cognitive load (Mike can review quickly in IDE)
- No complex business logic required

---

## Files to Create

### 1. Go Module Setup
```
/home/mikej/projects/DevSmith-Modular-Platform/
├── go.mod                          # Main module file
├── go.sum                          # Dependency checksums
└── Makefile                        # Build commands
```

### 2. Docker Configuration
```
├── docker-compose.yml              # All services (Postgres, Nginx, apps)
├── docker-compose.dev.yml          # Development overrides
├── .dockerignore                   # Exclude unnecessary files
├── cmd/
│   ├── portal/
│   │   └── Dockerfile             # Portal service Docker image
│   ├── review/
│   │   └── Dockerfile             # Review service Docker image
│   ├── logs/
│   │   └── Dockerfile             # Logs service Docker image
│   └── analytics/
│       └── Dockerfile             # Analytics service Docker image
└── docker/
    ├── postgres/
    │   └── init-schemas.sql       # Create schemas: portal, reviews, logs, analytics
    └── nginx/
        └── nginx.conf             # Gateway routing configuration
```

### 3. Environment Configuration
```
├── .env.example                    # Template with all required vars
├── .env.test                       # Test environment settings
└── config/
    └── config.go                  # Config loader with validation
```

### 4. Application Structure
```
├── cmd/
│   ├── portal/
│   │   ├── main.go                # Portal service entry point (HTTP server)
│   │   └── Dockerfile             # Portal Docker image
│   ├── review/
│   │   ├── main.go                # Review service entry point (HTTP server)
│   │   └── Dockerfile             # Review Docker image
│   ├── logs/
│   │   ├── main.go                # Logs service entry point (HTTP server)
│   │   └── Dockerfile             # Logs Docker image
│   └── analytics/
│       ├── main.go                # Analytics service entry point (HTTP server)
│       └── Dockerfile             # Analytics Docker image
└── internal/
    └── portal/
        └── models/
            └── user.go            # User model placeholder
```

**IMPORTANT:** Each `main.go` must run a persistent HTTP server using Gin, NOT just print a message and exit. Otherwise Docker containers will stop immediately and Nginx cannot route to them.

### 5. Development Tools
```
├── .gitignore                      # Go, IDE, OS, Docker ignores
├── .editorconfig                   # Code formatting rules
└── scripts/
    ├── setup.sh                   # One-command setup script
    ├── dev.sh                     # Start development environment
    └── test.sh                    # Run all tests
```

---

## Detailed Specifications

### go.mod
```go
module github.com/mikejsmith1985/devsmith-modular-platform

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1           // HTTP framework
    github.com/a-h/templ v0.2.543             // Template engine
    github.com/jackc/pgx/v5 v5.5.0            // Postgres driver
    github.com/rs/zerolog v1.31.0             // Structured logging
    github.com/joho/godotenv v1.5.1           // .env loader
)
```

### docker-compose.yml
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: devsmith
      POSTGRES_USER: devsmith
      POSTGRES_PASSWORD: ${DB_PASSWORD:-devsmith_local}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./docker/postgres/init-schemas.sql:/docker-entrypoint-initdb.d/01-schemas.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U devsmith"]
      interval: 5s
      timeout: 5s
      retries: 5

  nginx:
    image: nginx:alpine
    ports:
      - "3000:80"
    volumes:
      - ./docker/nginx/nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - portal
      - review

  portal:
    build:
      context: .
      dockerfile: cmd/portal/Dockerfile
    environment:
      - DATABASE_URL=postgres://devsmith:${DB_PASSWORD:-devsmith_local}@postgres:5432/devsmith
      - GITHUB_CLIENT_ID=${GITHUB_CLIENT_ID}
      - GITHUB_CLIENT_SECRET=${GITHUB_CLIENT_SECRET}
    depends_on:
      postgres:
        condition: service_healthy

  review:
    build:
      context: .
      dockerfile: cmd/review/Dockerfile
    environment:
      - DATABASE_URL=postgres://devsmith:${DB_PASSWORD:-devsmith_local}@postgres:5432/devsmith
      - OLLAMA_URL=${OLLAMA_URL:-http://host.docker.internal:11434}
    depends_on:
      postgres:
        condition: service_healthy

volumes:
  postgres_data:
```

### .env.example
```bash
# Database
DB_PASSWORD=devsmith_local

# GitHub OAuth
GITHUB_CLIENT_ID=your_client_id_here
GITHUB_CLIENT_SECRET=your_client_secret_here
GITHUB_CALLBACK_URL=http://localhost:3000/auth/github/callback

# Ollama
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=deepseek-coder-v2:16b

# Application
ENVIRONMENT=development
LOG_LEVEL=debug
PORT=8080
```

### docker/postgres/init-schemas.sql
```sql
-- Create isolated schemas for each service
-- Follows bounded context principle from ARCHITECTURE.md

CREATE SCHEMA IF NOT EXISTS portal;
COMMENT ON SCHEMA portal IS 'Portal service: Authentication, user management, app browser';

CREATE SCHEMA IF NOT EXISTS reviews;
COMMENT ON SCHEMA reviews IS 'Review service: Code reading sessions, 5 reading modes, AI analysis';

CREATE SCHEMA IF NOT EXISTS logs;
COMMENT ON SCHEMA logs IS 'Logging service: Application logs, telemetry, real-time streaming';

CREATE SCHEMA IF NOT EXISTS analytics;
COMMENT ON SCHEMA analytics IS 'Analytics service: Log analysis, anomaly detection, reports';

CREATE SCHEMA IF NOT EXISTS builds;
COMMENT ON SCHEMA builds IS 'Build service: Terminal sessions, OpenHands integration, CLI operations';

-- Grant privileges to devsmith user
GRANT ALL PRIVILEGES ON SCHEMA portal TO devsmith;
GRANT ALL PRIVILEGES ON SCHEMA reviews TO devsmith;
GRANT ALL PRIVILEGES ON SCHEMA logs TO devsmith;
GRANT ALL PRIVILEGES ON SCHEMA analytics TO devsmith;
GRANT ALL PRIVILEGES ON SCHEMA builds TO devsmith;
```

### Makefile
```makefile
.PHONY: help setup dev test build clean

help: ## Show this help message
	@echo "DevSmith Modular Platform - Make Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Run one-time setup (dependencies, Docker, database)
	@./scripts/setup.sh

dev: ## Start development environment
	@./scripts/dev.sh

test: ## Run all tests
	@./scripts/test.sh

build: ## Build all services
	@go build -o bin/portal ./cmd/portal
	@go build -o bin/review ./cmd/review
	@go build -o bin/logs ./cmd/logs
	@go build -o bin/analytics ./cmd/analytics
	@go build -o bin/build ./cmd/build

clean: ## Clean build artifacts and stop containers
	@rm -rf bin/
	@docker-compose down -v
```

### scripts/setup.sh
```bash
#!/bin/bash
set -e

echo "🚀 DevSmith Modular Platform - One-Command Setup"
echo "=================================================="

# Check prerequisites
command -v docker >/dev/null 2>&1 || { echo "❌ Docker not found. Install: https://docs.docker.com/get-docker/"; exit 1; }
command -v go >/dev/null 2>&1 || { echo "❌ Go not found. Install: https://go.dev/doc/install"; exit 1; }

echo "✅ Prerequisites found"

# Copy environment template
if [ ! -f .env ]; then
    echo "📝 Creating .env from template..."
    cp .env.example .env
    echo "⚠️  IMPORTANT: Edit .env and add your GitHub OAuth credentials!"
fi

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go mod download

# Install templ CLI
echo "📦 Installing Templ CLI..."
go install github.com/a-h/templ/cmd/templ@latest

# Start database
echo "🐘 Starting PostgreSQL..."
docker-compose up -d postgres
sleep 5

# Wait for database
echo "⏳ Waiting for database to be ready..."
until docker-compose exec -T postgres pg_isready -U devsmith; do
    sleep 1
done

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Edit .env and add GitHub OAuth credentials"
echo "  2. Run 'make dev' to start development environment"
echo "  3. Visit http://localhost:3000"
```

### .gitignore
```gitignore
# Binaries
bin/
*.exe
*.dll
*.so
*.dylib

# Go
*.test
*.out
go.work

# Environment
.env
.env.local
.env.*.local

# IDEs
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Docker
*.log

# Build artifacts
dist/
build/

# Test coverage
coverage.out
*.coverprofile

# Temporary files
tmp/
temp/
```

---

### Dockerfile Template (All Services)

Each service needs a multi-stage Dockerfile. **Create identical Dockerfiles in each cmd/*/Dockerfile location.**

**Example: cmd/portal/Dockerfile** (replicate for review, logs, analytics)

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the service
# CGO_ENABLED=0 for static binary
# -ldflags="-w -s" to strip debug info (smaller binary)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/portal ./cmd/portal

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /home/appuser

# Copy binary from builder
COPY --from=builder /app/bin/portal ./portal

# Change ownership
RUN chown -R appuser:appuser /home/appuser

# Switch to non-root user
USER appuser

# Expose port (8080 for portal, 8081 for review, 8082 for logs, 8083 for analytics)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["./portal"]
```

**Port assignments:**
- Portal: 8080
- Review: 8081
- Logs: 8082
- Analytics: 8083

**For each Dockerfile, change:**
1. Build path: `./cmd/portal` → `./cmd/review`, etc.
2. Binary name: `portal` → `review`, etc.
3. EXPOSE port: `8080` → `8081`, `8082`, `8083`
4. Health check port in URL

---

### Service main.go Files (CRITICAL FIX)

**PROBLEM:** Current main.go files just print and exit, causing containers to stop immediately.

**SOLUTION:** Each must run a persistent HTTP server using Gin.

**Example: cmd/portal/main.go** (replicate pattern for all services)

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create Gin router
	router := gin.Default()

	// Health check endpoint (required for Docker health checks)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "portal",
			"status":  "healthy",
		})
	})

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "DevSmith Portal",
			"version": "0.1.0",
			"message": "Portal service is running",
		})
	})

	// Get port from environment or default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server (this runs forever until killed)
	fmt.Printf("Portal service starting on port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

**For other services, change:**
- Service name in JSON responses: `"portal"` → `"review"`, `"logs"`, `"analytics"`
- Default port: `"8080"` → `"8081"`, `"8082"`, `"8083"`
- Log message: `"Portal service..."` → `"Review service..."`, etc.

---

### Nginx Gateway Configuration

**File:** `docker/nginx/nginx.conf`

```nginx
events {
    worker_connections 1024;
}

http {
    # Logging
    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    # Upstream services
    upstream portal {
        server portal:8080;
    }

    upstream review {
        server review:8081;
    }

    upstream logs {
        server logs:8082;
    }

    upstream analytics {
        server analytics:8083;
    }

    # Main server block
    server {
        listen 80;
        server_name localhost;

        # Portal (default)
        location / {
            proxy_pass http://portal;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Review app
        location /review {
            proxy_pass http://review;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Logs app
        location /logs {
            proxy_pass http://logs;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            # WebSocket support for log streaming
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
        }

        # Analytics app
        location /analytics {
            proxy_pass http://analytics;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Health check endpoint
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}
```

---

### scripts/dev.sh

```bash
#!/bin/bash
set -e

echo "🚀 Starting DevSmith Development Environment"
echo "============================================="

# Check if .env exists
if [ ! -f .env ]; then
    echo "❌ .env file not found!"
    echo "Run 'make setup' first to create .env from template"
    exit 1
fi

# Start all services
echo "📦 Starting Docker Compose services..."
docker-compose up --build

# Note: This will run in foreground
# Press Ctrl+C to stop all services
```

**Make executable:** `chmod +x scripts/dev.sh`

---

### scripts/test.sh

```bash
#!/bin/bash
set -e

echo "🧪 Running DevSmith Platform Tests"
echo "==================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go not found. Install from https://go.dev/doc/install"
    exit 1
fi

# Run tests with coverage
echo "📊 Running tests with coverage..."
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Display coverage summary
echo ""
echo "📈 Coverage Summary:"
go tool cover -func=coverage.out | grep total:

# Optional: Generate HTML coverage report
if [ "$1" == "--html" ]; then
    echo ""
    echo "📄 Generating HTML coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    echo "✅ Coverage report: coverage.html"
fi

echo ""
echo "✅ All tests passed!"
```

**Make executable:** `chmod +x scripts/test.sh`

---

### Updated docker-compose.yml

Replace the existing docker-compose.yml with this complete version:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: devsmith-postgres
    environment:
      POSTGRES_DB: devsmith
      POSTGRES_USER: devsmith
      POSTGRES_PASSWORD: ${DB_PASSWORD:-devsmith_local}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./docker/postgres/init-schemas.sql:/docker-entrypoint-initdb.d/01-schemas.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U devsmith -d devsmith"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - devsmith-network

  portal:
    build:
      context: .
      dockerfile: cmd/portal/Dockerfile
    container_name: devsmith-portal
    environment:
      - PORT=8080
      - DATABASE_URL=postgres://devsmith:${DB_PASSWORD:-devsmith_local}@postgres:5432/devsmith
      - GITHUB_CLIENT_ID=${GITHUB_CLIENT_ID}
      - GITHUB_CLIENT_SECRET=${GITHUB_CLIENT_SECRET}
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    networks:
      - devsmith-network

  review:
    build:
      context: .
      dockerfile: cmd/review/Dockerfile
    container_name: devsmith-review
    environment:
      - PORT=8081
      - DATABASE_URL=postgres://devsmith:${DB_PASSWORD:-devsmith_local}@postgres:5432/devsmith
      - OLLAMA_URL=${OLLAMA_URL:-http://host.docker.internal:11434}
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8081/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    networks:
      - devsmith-network

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

  analytics:
    build:
      context: .
      dockerfile: cmd/analytics/Dockerfile
    container_name: devsmith-analytics
    environment:
      - PORT=8083
      - DATABASE_URL=postgres://devsmith:${DB_PASSWORD:-devsmith_local}@postgres:5432/devsmith
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8083/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    networks:
      - devsmith-network

  nginx:
    image: nginx:alpine
    container_name: devsmith-nginx
    ports:
      - "3000:80"
    volumes:
      - ./docker/nginx/nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - portal
      - review
      - logs
      - analytics
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:80/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    networks:
      - devsmith-network

volumes:
  postgres_data:

networks:
  devsmith-network:
    driver: bridge
```

---

### internal/portal/models/user.go

Simple placeholder model:

```go
package models

import "time"

// User represents a user in the portal system
type User struct {
	GitHubID  int64     `json:"github_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
```

---

## Acceptance Criteria

### Core Files
- [ ] `go.mod` created with correct module path and dependencies
- [ ] `docker-compose.yml` defines all 4 services + Postgres + Nginx with health checks and networks
- [ ] `init-schemas.sql` creates 4 schemas (portal, reviews, logs, analytics)
- [ ] `.env.example` contains all required environment variables
- [ ] `Makefile` has `help`, `setup`, `dev`, `test`, `build`, `clean` targets
- [ ] `.gitignore` covers Go, Docker, IDEs, and OS files

### Docker Infrastructure
- [ ] All 4 Dockerfiles created (`cmd/portal/Dockerfile`, `cmd/review/Dockerfile`, `cmd/logs/Dockerfile`, `cmd/analytics/Dockerfile`)
- [ ] Each Dockerfile uses multi-stage build (builder + runtime)
- [ ] Each Dockerfile includes health check
- [ ] Nginx configuration created (`docker/nginx/nginx.conf`)
- [ ] PostgreSQL init script created (`docker/postgres/init-schemas.sql`)

### Service Implementation
- [ ] All 4 `cmd/*/main.go` files run persistent HTTP servers using Gin
- [ ] Each service has `/health` endpoint
- [ ] Each service has `/` root endpoint with service info
- [ ] Services use correct ports (8080, 8081, 8082, 8083)
- [ ] **CRITICAL:** Containers stay running (don't exit immediately)

### Scripts
- [ ] `scripts/setup.sh` is executable and runs without errors
- [ ] `scripts/dev.sh` is executable and starts all services
- [ ] `scripts/test.sh` is executable and runs tests

### Integration Testing
- [ ] Can run `make setup` successfully
- [ ] Can run `make dev` and all services start
- [ ] All services show as "healthy" in `docker-compose ps`
- [ ] Can access http://localhost:3000/health (nginx)
- [ ] Can access http://localhost:3000/ (portal via nginx)
- [ ] Can access http://localhost:3000/review (review via nginx)
- [ ] Can access http://localhost:3000/logs (logs via nginx)
- [ ] Can access http://localhost:3000/analytics (analytics via nginx)
- [ ] Can run `make test` successfully (even with no tests yet)
- [ ] Database has 4 schemas created

---

## Testing Commands

```bash
# Clean slate
make clean

# Run setup
make setup
# Should see:
# - ✅ Prerequisites found
# - ✅ Go dependencies installed
# - ✅ PostgreSQL is ready
# - ✅ All 4 schemas created successfully

# Start development environment
make dev
# In another terminal:

# Test health checks
curl http://localhost:3000/health          # Nginx
curl http://localhost:3000/                # Portal (via nginx)
curl http://localhost:3000/review          # Review (via nginx)
curl http://localhost:3000/logs            # Logs (via nginx)
curl http://localhost:3000/analytics       # Analytics (via nginx)

# Check all services healthy
docker-compose ps
# All services should show "healthy"

# Verify database schemas
docker-compose exec postgres psql -U devsmith -d devsmith -c "\dn"
# Should show: portal, reviews, logs, analytics schemas

# Run tests
make test
# Should pass (even if no tests exist yet)

# Clean up
make clean
```

---

## Context and References

- **ARCHITECTURE.md** - See "Technology Stack" and "Service Architecture" sections
- **Requirements.md** - See "Technology Stack" section for dependency rationale
- **DevSmithRoles.md** - Copilot handles 5-10% of work (scaffolding tasks)

---

## Implementation Notes for Copilot

**Copilot Prompts to Use:**

1. In `go.mod`: "Create Go module for microservices platform with Gin, Templ, pgx, zerolog"
2. In `docker-compose.yml`: "Docker Compose with Postgres 15, Nginx gateway, health checks"
3. In `init-schemas.sql`: "Create 5 PostgreSQL schemas for microservices with comments"
4. In `Makefile`: "Makefile with help, setup, dev, test, build, clean targets"

**Expected Time:**
- Setting up files: 20 minutes (with Copilot autocomplete)
- Testing: 10 minutes
- Documentation: 5 minutes
- **Total: ~35 minutes**

---

## Implementation Steps

### 1. Create Feature Branch

**Branch Naming Convention:** See `ARCHITECTURE.md` - Branch Strategy section

```bash
git checkout development
git pull origin development
git checkout -b feature/001-project-scaffolding
```

### 2. Implement Files

Follow the file specifications above, using Copilot autocomplete to speed up boilerplate creation.

### 3. Test Setup

```bash
# Verify Go module
go mod verify

# Test Docker Compose syntax
docker-compose config

# Test database connection
docker-compose up -d postgres
sleep 5
docker-compose exec postgres psql -U devsmith -c "\dn"
# Should show 5 schemas: portal, reviews, logs, analytics, builds

# Test setup script
./scripts/setup.sh

# Clean up
make clean
```

### 4. Commit and Push

```bash
git add -A
git commit -m "feat(infra): complete project scaffolding and Docker infrastructure

- Go module with dependencies (Gin, Templ, pgx, zerolog)
- Docker Compose with health checks, networks, and all 4 services
- PostgreSQL schema initialization (4 schemas: portal, reviews, logs, analytics)
- Dockerfiles for all 4 services (multi-stage builds)
- Nginx gateway configuration with proper routing
- Persistent HTTP servers for all services (Gin with /health endpoints)
- Development scripts (setup.sh, dev.sh, test.sh)
- Makefile with common commands
- User model placeholder

CRITICAL FIX: Services now run persistent HTTP servers instead of
just printing and exiting. This fixes Nginx upstream resolution errors.

All services are now buildable, runnable, and accessible via Nginx gateway.

Implements .docs/issues/001-copilot-project-scaffolding.md

🤖 Generated with Copilot assistance

Co-Authored-By: GitHub Copilot <noreply@github.com>"

git push origin feature/001-copilot-project-scaffolding
```

### 5. Create Pull Request

```bash
gh pr create --title "feat(setup): project scaffolding and configuration" --body "
## Summary
Complete project scaffolding with Go modules, Docker Compose, database schemas, and build tools.

## Implementation
Implements \`.docs/issues/001-copilot-project-scaffolding.md\`

## Acceptance Criteria
- [x] Go module initialized with correct dependencies
- [x] Docker Compose defines all 5 services + Postgres + Nginx
- [x] init-schemas.sql creates 5 schemas (portal, reviews, logs, analytics, builds)
- [x] .env.example contains all required environment variables
- [x] Makefile has help, setup, dev, test, build, clean targets
- [x] scripts/setup.sh is executable and runs without errors
- [x] .gitignore covers Go, Docker, IDEs, and OS files
- [x] Directory structure matches specification
- [x] All 5 cmd/*/main.go files have basic package main and func main()
- [x] Can run \`make setup\` successfully
- [x] Can run \`docker-compose up postgres\` and connect to database

## Testing
Ran all verification commands successfully.

## Next Steps
After merge, OpenHands will implement Issue #2 (Portal Authentication).
"
```

### 6. After Merge

```bash
# Switch back to development and delete feature branch
git checkout development
git pull origin development
git branch -d feature/001-project-scaffolding
```

## Next Steps After Completion

1. Merge this PR to `development` branch
2. Move to Issue #2: `.docs/issues/002-openhands-portal-authentication.md`
3. OpenHands will autonomously implement Portal authentication
