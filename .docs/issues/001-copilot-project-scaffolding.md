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
└── docker/
    ├── postgres/
    │   └── init-schemas.sql       # Create schemas: portal, reviews, logs, analytics, builds
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
│   │   └── main.go                # Portal service entry point
│   ├── review/
│   │   └── main.go                # Review service entry point
│   ├── logs/
│   │   └── main.go                # Logging service entry point
│   ├── analytics/
│   │   └── main.go                # Analytics service entry point
│   └── build/
│       └── main.go                # Build service entry point
└── internal/
    └── shared/
        ├── middleware/
        │   ├── auth.go            # GitHub OAuth middleware
        │   └── logging.go         # Request logging
        └── database/
            └── connection.go      # Postgres connection pool
```

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

## Acceptance Criteria

- [ ] `go.mod` created with correct module path and dependencies
- [ ] `docker-compose.yml` defines all 5 services + Postgres + Nginx
- [ ] `init-schemas.sql` creates 5 schemas (portal, reviews, logs, analytics, builds)
- [ ] `.env.example` contains all required environment variables
- [ ] `Makefile` has `help`, `setup`, `dev`, `test`, `build`, `clean` targets
- [ ] `scripts/setup.sh` is executable and runs without errors
- [ ] `.gitignore` covers Go, Docker, IDEs, and OS files
- [ ] Directory structure matches specification
- [ ] All 5 `cmd/*/main.go` files have basic `package main` and `func main()`
- [ ] Can run `make setup` successfully
- [ ] Can run `docker-compose up postgres` and connect to database

---

## Testing Commands

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

## Next Steps After Completion

1. Run `make setup` to verify everything works
2. Commit with message: `feat(setup): project scaffolding and configuration`
3. Create PR to `development` branch
4. Move to Issue #2 (OpenHands will implement Portal service)
