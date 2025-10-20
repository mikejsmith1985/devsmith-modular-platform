# Issue #016: [COPILOT] End-to-End Integration & One-Command Setup

**Type:** Feature (Copilot Implementation)
**Service:** All services (Platform-wide)
**Depends On:** Issues #001-#015 (All previous issues)
**Estimated Duration:** 90-120 minutes

---

## Summary

Create the final integration layer that ties all services together and provides a one-command setup script for deploying the entire DevSmith Modular Platform. This includes environment configuration, database migrations, service health checks, and automated testing of the complete system.

**User Story:**
> As a new user of DevSmith Platform, I want to run a single command that sets up the entire platform (databases, services, dependencies), so I can start using the platform immediately without manual configuration.

---

## Bounded Context

**Platform Integration Context:**
- **Responsibility:** Platform-wide orchestration, deployment, health monitoring
- **Does NOT:** Implement service features (those live in individual services)
- **Boundaries:** Integration scripts coordinate services but don't contain business logic

**Why This Matters:**
- One command (`./setup.sh`) should get platform running
- Each service remains independent but connected via gateway
- Health checks verify end-to-end functionality

---

## Success Criteria

### Must Have (MVP)
- [ ] One-command setup script (`./setup.sh`) that:
  - [ ] Checks prerequisites (Go, Docker, PostgreSQL, Ollama)
  - [ ] Creates all databases and schemas
  - [ ] Runs database migrations for all services
  - [ ] Sets up .env files from templates
  - [ ] Builds all service binaries
  - [ ] Starts all services in correct order
  - [ ] Runs health checks on all services
  - [ ] Reports setup status (success/failure)
- [ ] Integration tests that verify:
  - [ ] Portal authentication flow
  - [ ] Cross-service communication
  - [ ] Database connectivity for all services
  - [ ] WebSocket connections (Logs service)
  - [ ] AI analysis (Review service with Ollama)
- [ ] Health check endpoint for each service
- [ ] Docker Compose file for containerized deployment
- [ ] Setup verification script (`./verify-setup.sh`)
- [ ] Teardown script (`./teardown.sh`)

### Nice to Have (Post-MVP)
- Kubernetes deployment manifests
- Automated backup/restore scripts
- Performance testing suite
- Load testing harness

---

## File Structure

```
/ (root)
├── setup.sh                      # NEW - One-command setup
├── teardown.sh                   # NEW - Clean teardown
├── verify-setup.sh               # NEW - Verify installation
├── docker-compose.yml            # UPDATE - Full platform
├── docker-compose.dev.yml        # NEW - Dev environment
├── .env.example                  # NEW - Platform-wide template
├── scripts/
│   ├── check-prerequisites.sh    # NEW - Prereq checker
│   ├── create-databases.sh       # NEW - Database setup
│   ├── run-migrations.sh         # NEW - Run all migrations
│   ├── build-services.sh         # NEW - Build all Go binaries
│   ├── start-services.sh         # NEW - Start in order
│   └── health-checks.sh          # NEW - Verify all services
├── tests/
│   └── integration/
│       ├── auth_flow_test.go     # NEW - E2E auth test
│       ├── review_flow_test.go   # NEW - E2E review test
│       ├── logs_flow_test.go     # NEW - E2E logs test
│       └── analytics_flow_test.go # NEW - E2E analytics test
└── migrations/
    ├── portal/                   # Existing from Issue #003
    ├── review/                   # Existing from Issues #004-#008
    ├── logs/                     # Existing from Issue #009
    └── analytics/                # Existing from Issue #011
```

---

## Implementation Details

### 1. Main Setup Script

**File:** `setup.sh`

```bash
#!/bin/bash
set -e

echo "🚀 DevSmith Modular Platform - One-Command Setup"
echo "=================================================="
echo ""

# Colors for output
GREEN='\033[0.32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Step 1: Check prerequisites
echo "📋 Step 1/8: Checking prerequisites..."
./scripts/check-prerequisites.sh
echo -e "${GREEN}✓ Prerequisites verified${NC}\n"

# Step 2: Create databases
echo "🗄️  Step 2/8: Creating databases and schemas..."
./scripts/create-databases.sh
echo -e "${GREEN}✓ Databases created${NC}\n"

# Step 3: Run migrations
echo "📊 Step 3/8: Running database migrations..."
./scripts/run-migrations.sh
echo -e "${GREEN}✓ Migrations completed${NC}\n"

# Step 4: Setup environment files
echo "⚙️  Step 4/8: Setting up environment configuration..."
if [ ! -f .env ]; then
  cp .env.example .env
  echo -e "${YELLOW}⚠️  Please edit .env with your GitHub OAuth credentials${NC}"
  echo "   Then run ./setup.sh again"
  exit 1
fi
echo -e "${GREEN}✓ Environment configured${NC}\n"

# Step 5: Build services
echo "🔨 Step 5/8: Building all service binaries..."
./scripts/build-services.sh
echo -e "${GREEN}✓ Services built${NC}\n"

# Step 6: Start Ollama and select model based on RAM
echo "🤖 Step 6/8: Setting up Ollama and AI model..."
if ! pgrep -x "ollama" > /dev/null; then
  echo "Starting Ollama..."
  ollama serve > /dev/null 2>&1 &
  sleep 3
fi

# Detect RAM and recommend model
TOTAL_RAM=$(free -g 2>/dev/null | awk '/^Mem:/{print $2}' || sysctl -n hw.memsize 2>/dev/null | awk '{print int($1/1024/1024/1024)}')

if [ -z "$TOTAL_RAM" ] || [ "$TOTAL_RAM" -lt 8 ]; then
  echo -e "${RED}⚠️  Unable to detect RAM or less than 8GB${NC}"
  echo "   Recommend: deepseek-coder:1.5b (minimal)"
  DEFAULT_MODEL="deepseek-coder:1.5b"
elif [ "$TOTAL_RAM" -lt 24 ]; then
  echo "✓ ${TOTAL_RAM}GB RAM detected"
  echo "   Recommend: deepseek-coder:6.7b (good balance)"
  DEFAULT_MODEL="deepseek-coder:6.7b"
else
  echo "✓ ${TOTAL_RAM}GB RAM detected"
  echo "   Recommend: deepseek-coder-v2:16b (best quality)"
  DEFAULT_MODEL="deepseek-coder-v2:16b"
fi

echo ""
echo "Available models:"
echo "  1) deepseek-coder:1.5b (8GB RAM, ~1GB download, fastest)"
echo "  2) deepseek-coder:6.7b (16GB RAM, ~4GB download, recommended)"
echo "  3) deepseek-coder-v2:16b (32GB RAM, ~9GB download, best quality)"
echo "  4) qwen2.5-coder:7b (16GB RAM, ~4GB download, alternative)"
echo ""

read -p "Select model [2]: " MODEL_CHOICE
MODEL_CHOICE=${MODEL_CHOICE:-2}

case $MODEL_CHOICE in
  1) CHOSEN_MODEL="deepseek-coder:1.5b" ;;
  2) CHOSEN_MODEL="deepseek-coder:6.7b" ;;
  3) CHOSEN_MODEL="deepseek-coder-v2:16b" ;;
  4) CHOSEN_MODEL="qwen2.5-coder:7b" ;;
  *) CHOSEN_MODEL=$DEFAULT_MODEL ;;
esac

echo "Selected model: $CHOSEN_MODEL"

# Pull model if not already present
if ! ollama list | grep -q "$CHOSEN_MODEL"; then
  echo "Pulling $CHOSEN_MODEL (this may take 5-15 minutes depending on model size)..."
  ollama pull $CHOSEN_MODEL
else
  echo "Model $CHOSEN_MODEL already downloaded"
fi

# Update .env with chosen model
if [ -f .env ]; then
  if grep -q "OLLAMA_MODEL=" .env; then
    sed -i.bak "s|OLLAMA_MODEL=.*|OLLAMA_MODEL=$CHOSEN_MODEL|" .env
  else
    echo "OLLAMA_MODEL=$CHOSEN_MODEL" >> .env
  fi
  echo "✓ Updated .env with OLLAMA_MODEL=$CHOSEN_MODEL"
fi

echo -e "${GREEN}✓ Ollama ready${NC}\n"

# Step 7: Start services
echo "🚀 Step 7/8: Starting all services..."
./scripts/start-services.sh
echo -e "${GREEN}✓ Services started${NC}\n"

# Step 8: Health checks
echo "🏥 Step 8/8: Running health checks..."
sleep 5  # Give services time to start
./scripts/health-checks.sh
echo -e "${GREEN}✓ All services healthy${NC}\n"

echo "=================================================="
echo -e "${GREEN}✅ Setup complete!${NC}"
echo ""
echo "🌐 Platform URLs:"
echo "   Portal:    http://localhost:8080"
echo "   Review:    http://localhost:8081"
echo "   Logs:      http://localhost:8082"
echo "   Analytics: http://localhost:8083"
echo ""
echo "📝 Next steps:"
echo "   1. Open http://localhost:8080"
echo "   2. Log in with GitHub OAuth"
echo "   3. Start reviewing code!"
echo ""
echo "🛑 To stop all services: ./teardown.sh"
echo "🔍 To verify setup: ./verify-setup.sh"
```

---

### 2. Prerequisites Checker

**File:** `scripts/check-prerequisites.sh`

```bash
#!/bin/bash
set -e

MISSING=0

# Check Go
if ! command -v go &> /dev/null; then
  echo "❌ Go is not installed (required: >= 1.23)"
  MISSING=1
else
  GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
  echo "✓ Go $GO_VERSION"
fi

# Check PostgreSQL
if ! command -v psql &> /dev/null; then
  echo "❌ PostgreSQL is not installed (required: >= 14)"
  MISSING=1
else
  PG_VERSION=$(psql --version | awk '{print $3}')
  echo "✓ PostgreSQL $PG_VERSION"
fi

# Check Docker
if ! command -v docker &> /dev/null; then
  echo "❌ Docker is not installed (optional but recommended)"
else
  DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
  echo "✓ Docker $DOCKER_VERSION"
fi

# Check Ollama
if ! command -v ollama &> /dev/null; then
  echo "❌ Ollama is not installed (required for AI features)"
  MISSING=1
else
  echo "✓ Ollama installed"
fi

# Check Node.js (for Templ if needed)
if ! command -v node &> /dev/null; then
  echo "⚠️  Node.js not found (optional for development)"
else
  NODE_VERSION=$(node --version)
  echo "✓ Node.js $NODE_VERSION"
fi

# Check Templ CLI
if ! command -v templ &> /dev/null; then
  echo "📦 Installing Templ CLI..."
  go install github.com/a-h/templ/cmd/templ@latest
  echo "✓ Templ CLI installed"
else
  echo "✓ Templ CLI installed"
fi

if [ $MISSING -eq 1 ]; then
  echo ""
  echo "❌ Missing required dependencies. Please install them first."
  echo ""
  echo "Installation instructions:"
  echo "  Go: https://go.dev/dl/"
  echo "  PostgreSQL: https://www.postgresql.org/download/"
  echo "  Ollama: https://ollama.ai/"
  exit 1
fi

echo "✓ All prerequisites met"
```

---

### 3. Database Creation Script

**File:** `scripts/create-databases.sh`

```bash
#!/bin/bash
set -e

echo "Creating databases and schemas..."

# Database names
PORTAL_DB="devsmith_portal"
REVIEW_DB="devsmith_review"
LOGS_DB="devsmith_logs"
ANALYTICS_DB="devsmith_analytics"

# Create databases
psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$PORTAL_DB'" | grep -q 1 || \
  psql -U postgres -c "CREATE DATABASE $PORTAL_DB"

psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$REVIEW_DB'" | grep -q 1 || \
  psql -U postgres -c "CREATE DATABASE $REVIEW_DB"

psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$LOGS_DB'" | grep -q 1 || \
  psql -U postgres -c "CREATE DATABASE $LOGS_DB"

psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$ANALYTICS_DB'" | grep -q 1 || \
  psql -U postgres -c "CREATE DATABASE $ANALYTICS_DB"

echo "✓ Databases created"

# Create schemas
psql -U postgres -d $PORTAL_DB -c "CREATE SCHEMA IF NOT EXISTS portal"
psql -U postgres -d $REVIEW_DB -c "CREATE SCHEMA IF NOT EXISTS review"
psql -U postgres -d $LOGS_DB -c "CREATE SCHEMA IF NOT EXISTS logs"
psql -U postgres -d $ANALYTICS_DB -c "CREATE SCHEMA IF NOT EXISTS analytics"

echo "✓ Schemas created"

# Create users (if not exist)
psql -U postgres -tc "SELECT 1 FROM pg_roles WHERE rolname='portal_user'" | grep -q 1 || \
  psql -U postgres -c "CREATE USER portal_user WITH PASSWORD 'portal_pass'"

psql -U postgres -tc "SELECT 1 FROM pg_roles WHERE rolname='review_user'" | grep -q 1 || \
  psql -U postgres -c "CREATE USER review_user WITH PASSWORD 'review_pass'"

psql -U postgres -tc "SELECT 1 FROM pg_roles WHERE rolname='logs_user'" | grep -q 1 || \
  psql -U postgres -c "CREATE USER logs_user WITH PASSWORD 'logs_pass'"

psql -U postgres -tc "SELECT 1 FROM pg_roles WHERE rolname='analytics_user'" | grep -q 1 || \
  psql -U postgres -c "CREATE USER analytics_user WITH PASSWORD 'analytics_pass'"

echo "✓ Users created"

# Grant permissions
psql -U postgres -d $PORTAL_DB -c "GRANT ALL PRIVILEGES ON SCHEMA portal TO portal_user"
psql -U postgres -d $REVIEW_DB -c "GRANT ALL PRIVILEGES ON SCHEMA review TO review_user"
psql -U postgres -d $LOGS_DB -c "GRANT ALL PRIVILEGES ON SCHEMA logs TO logs_user"
psql -U postgres -d $ANALYTICS_DB -c "GRANT ALL PRIVILEGES ON SCHEMA analytics TO analytics_user"

# Grant analytics READ-ONLY access to logs
psql -U postgres -d $LOGS_DB -c "GRANT USAGE ON SCHEMA logs TO analytics_user"
psql -U postgres -d $LOGS_DB -c "GRANT SELECT ON ALL TABLES IN SCHEMA logs TO analytics_user"
psql -U postgres -d $LOGS_DB -c "ALTER DEFAULT PRIVILEGES IN SCHEMA logs GRANT SELECT ON TABLES TO analytics_user"

echo "✓ Permissions granted"
```

---

### 4. Run Migrations Script

**File:** `scripts/run-migrations.sh`

```bash
#!/bin/bash
set -e

echo "Running database migrations..."

# Install golang-migrate if not present
if ! command -v migrate &> /dev/null; then
  echo "Installing golang-migrate..."
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

# Portal migrations
echo "→ Portal service migrations..."
migrate -path migrations/portal -database "postgresql://portal_user:portal_pass@localhost:5432/devsmith_portal?sslmode=disable" up

# Review migrations
echo "→ Review service migrations..."
migrate -path migrations/review -database "postgresql://review_user:review_pass@localhost:5432/devsmith_review?sslmode=disable" up

# Logs migrations
echo "→ Logs service migrations..."
migrate -path migrations/logs -database "postgresql://logs_user:logs_pass@localhost:5432/devsmith_logs?sslmode=disable" up

# Analytics migrations
echo "→ Analytics service migrations..."
migrate -path migrations/analytics -database "postgresql://analytics_user:analytics_pass@localhost:5432/devsmith_analytics?sslmode=disable" up

echo "✓ All migrations completed"
```

---

### 5. Build Services Script

**File:** `scripts/build-services.sh`

```bash
#!/bin/bash
set -e

echo "Building Go services..."

# Generate Templ templates first
echo "→ Generating Templ templates..."
find apps -name "*.templ" -exec dirname {} \; | sort -u | while read dir; do
  echo "  Generating templates in $dir"
  (cd "$dir" && templ generate)
done

# Build each service
echo "→ Building Portal service..."
go build -o bin/portal ./cmd/portal

echo "→ Building Review service..."
go build -o bin/review ./cmd/review

echo "→ Building Logs service..."
go build -o bin/logs ./cmd/logs

echo "→ Building Analytics service..."
go build -o bin/analytics ./cmd/analytics

echo "✓ All services built successfully"
```

---

### 6. Start Services Script

**File:** `scripts/start-services.sh`

```bash
#!/bin/bash
set -e

echo "Starting services..."

# Create logs directory
mkdir -p logs

# Start services in background
echo "→ Starting Portal service (port 8080)..."
./bin/portal > logs/portal.log 2>&1 &
echo $! > .pid_portal

echo "→ Starting Review service (port 8081)..."
./bin/review > logs/review.log 2>&1 &
echo $! > .pid_review

echo "→ Starting Logs service (port 8082)..."
./bin/logs > logs/logs.log 2>&1 &
echo $! > .pid_logs

echo "→ Starting Analytics service (port 8083)..."
./bin/analytics > logs/analytics.log 2>&1 &
echo $! > .pid_analytics

echo "✓ All services started"
echo "  PIDs saved to .pid_* files"
echo "  Logs available in logs/ directory"
```

---

### 7. Health Checks Script

**File:** `scripts/health-checks.sh`

```bash
#!/bin/bash

FAILED=0

check_service() {
  SERVICE=$1
  URL=$2

  if curl -f -s "$URL" > /dev/null; then
    echo "✓ $SERVICE is healthy"
  else
    echo "❌ $SERVICE is NOT responding at $URL"
    FAILED=1
  fi
}

echo "Checking service health..."

check_service "Portal" "http://localhost:8080/health"
check_service "Review" "http://localhost:8081/health"
check_service "Logs" "http://localhost:8082/health"
check_service "Analytics" "http://localhost:8083/health"

if [ $FAILED -eq 1 ]; then
  echo ""
  echo "❌ Some services failed health checks"
  echo "   Check logs in logs/ directory"
  exit 1
fi

echo ""
echo "✓ All services are healthy"
```

---

### 8. Teardown Script

**File:** `teardown.sh`

```bash
#!/bin/bash

echo "🛑 Stopping DevSmith Platform services..."

# Kill services by PID
for pidfile in .pid_*; do
  if [ -f "$pidfile" ]; then
    SERVICE=$(basename "$pidfile" | sed 's/.pid_//')
    PID=$(cat "$pidfile")

    if kill -0 "$PID" 2>/dev/null; then
      echo "→ Stopping $SERVICE (PID $PID)..."
      kill "$PID"
    fi

    rm "$pidfile"
  fi
done

echo "✓ All services stopped"
echo ""
echo "Note: Databases and data are preserved"
echo "      To clean everything: ./teardown.sh --clean"

if [ "$1" == "--clean" ]; then
  echo ""
  echo "🗑️  Cleaning databases..."
  psql -U postgres -c "DROP DATABASE IF EXISTS devsmith_portal"
  psql -U postgres -c "DROP DATABASE IF EXISTS devsmith_review"
  psql -U postgres -c "DROP DATABASE IF EXISTS devsmith_logs"
  psql -U postgres -c "DROP DATABASE IF EXISTS devsmith_analytics"
  echo "✓ Databases dropped"
fi
```

---

### 9. Environment Template

**File:** `.env.example`

```bash
# ==============================================================================
# DevSmith Modular Platform - Environment Configuration
# ==============================================================================
# Copy this file to .env and fill in your values
# DO NOT commit .env to version control

# ------------------------------------------------------------------------------
# Portal Service (Authentication)
# ------------------------------------------------------------------------------
PORTAL_PORT=8080
PORTAL_DATABASE_URL=postgresql://portal_user:portal_pass@localhost:5432/devsmith_portal?sslmode=disable

# GitHub OAuth (https://github.com/settings/developers)
GITHUB_CLIENT_ID=your_github_oauth_client_id_here
GITHUB_CLIENT_SECRET=your_github_oauth_client_secret_here
GITHUB_REDIRECT_URL=http://localhost:8080/auth/callback

# JWT Secret (generate with: openssl rand -base64 32)
JWT_SECRET=your_jwt_secret_min_32_characters_here

# ------------------------------------------------------------------------------
# Review Service (Code Analysis)
# ------------------------------------------------------------------------------
REVIEW_PORT=8081
REVIEW_DATABASE_URL=postgresql://review_user:review_pass@localhost:5432/devsmith_review?sslmode=disable

# Ollama Configuration
OLLAMA_URL=http://localhost:11434

# AI Model Selection (auto-configured by setup script based on RAM)
# Options:
#   deepseek-coder:1.5b    - 8GB RAM min, fastest, basic quality
#   deepseek-coder:6.7b    - 16GB RAM, recommended, good balance
#   deepseek-coder-v2:16b  - 32GB RAM, best quality, slower
#   qwen2.5-coder:7b       - 16GB RAM, alternative
OLLAMA_MODEL=deepseek-coder:6.7b

# Model settings (optional, defaults shown)
# OLLAMA_TEMPERATURE=0.7
# OLLAMA_TOP_P=0.9
# OLLAMA_CONTEXT_LENGTH=4096

# GitHub API (for fetching repositories)
GITHUB_TOKEN=your_github_personal_access_token

# ------------------------------------------------------------------------------
# Logs Service (Log Aggregation)
# ------------------------------------------------------------------------------
LOGS_PORT=8082
LOGS_DATABASE_URL=postgresql://logs_user:logs_pass@localhost:5432/devsmith_logs?sslmode=disable

# WebSocket Configuration
WEBSOCKET_PING_INTERVAL=30s
WEBSOCKET_MAX_CONNECTIONS=100

# ------------------------------------------------------------------------------
# Analytics Service (Trend Analysis)
# ------------------------------------------------------------------------------
ANALYTICS_PORT=8083
ANALYTICS_DATABASE_URL=postgresql://analytics_user:analytics_pass@localhost:5432/devsmith_analytics?sslmode=disable

# Cross-schema read-only access to Logs
LOGS_DATABASE_URL=postgresql://analytics_user:readonly_pass@localhost:5432/devsmith_logs?sslmode=disable

# Analytics Configuration
AGGREGATION_INTERVAL=1h
ANOMALY_THRESHOLD=2.0

# ------------------------------------------------------------------------------
# Platform Configuration
# ------------------------------------------------------------------------------
ENVIRONMENT=development
LOG_LEVEL=info
```

---

### 10. Integration Test Example

**File:** `tests/integration/auth_flow_test.go`

```go
package integration

import (
	"net/http"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAuthFlow(t *testing.T) {
	// Test Portal health
	resp, err := http.Get("http://localhost:8080/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test login page loads
	resp, err = http.Get("http://localhost:8080/login")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test GitHub OAuth initiation
	resp, err = http.Get("http://localhost:8080/auth/github")
	assert.NoError(t, err)
	// Should redirect to GitHub
	assert.Equal(t, http.StatusFound, resp.StatusCode)
}

func TestCrossServiceAccess(t *testing.T) {
	services := []struct {
		name string
		url  string
	}{
		{"Portal", "http://localhost:8080/health"},
		{"Review", "http://localhost:8081/health"},
		{"Logs", "http://localhost:8082/health"},
		{"Analytics", "http://localhost:8083/health"},
	}

	for _, service := range services {
		t.Run(service.name, func(t *testing.T) {
			resp, err := http.Get(service.url)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
```

---

## Testing Requirements

### Manual Testing Checklist

- [ ] Clone fresh repository
- [ ] Run `./setup.sh`
- [ ] Verify all prerequisites checked
- [ ] Verify databases created
- [ ] Verify migrations run successfully
- [ ] Verify services built without errors
- [ ] Verify all services start
- [ ] Verify health checks pass
- [ ] Open `http://localhost:8080`
- [ ] Complete GitHub OAuth login
- [ ] Navigate to Review service from dashboard
- [ ] Run a code review analysis
- [ ] Navigate to Logs service - verify logs appear
- [ ] Navigate to Analytics service - verify charts render
- [ ] Run `./verify-setup.sh` - all checks pass
- [ ] Run `./teardown.sh` - services stop cleanly
- [ ] Run `./setup.sh` again - should restart successfully

---

## Configuration

**No service-specific config - uses existing .env files from previous issues.**

---

## Acceptance Criteria

Before marking this issue complete, verify:

- [x] `./setup.sh` completes successfully on fresh install
- [x] All databases created with correct schemas
- [x] All migrations run without errors
- [x] All services build without errors
- [x] All services start and respond to health checks
- [x] Portal authentication works end-to-end
- [x] Review service can analyze repositories
- [x] Logs service displays real-time logs
- [x] Analytics service shows trends and charts
- [x] Integration tests pass
- [x] `./verify-setup.sh` reports all systems operational
- [x] `./teardown.sh` stops all services cleanly
- [x] Documentation includes setup instructions
- [x] Docker Compose file works for containerized deployment

---

## Branch Naming

```bash
feature/016-e2e-integration-setup
```

---

## Notes

- Setup script is idempotent (can run multiple times)
- Each service has health endpoint at `/health`
- Logs written to `logs/` directory for debugging
- PIDs saved to `.pid_*` files for clean shutdown
- Ollama model pull takes 10-15 minutes on first run
- PostgreSQL must be running before setup
- Docker Compose is alternative to manual setup

---

**Created:** 2025-10-20
**For:** Copilot Implementation
**Estimated Time:** 90-120 minutes
**Priority:** HIGH - This is the final MVP integration piece
