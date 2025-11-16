# DevSmith Platform: Developer Onboarding Guide

Welcome to the DevSmith Modular Platform! This guide will help you get up and running as a contributor.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Architecture Overview](#architecture-overview)
4. [Development Workflow](#development-workflow)
5. [Running Tests](#running-tests)
6. [Code Standards](#code-standards)
7. [Common Tasks](#common-tasks)
8. [Troubleshooting](#troubleshooting)
9. [Resources](#resources)

---

## Prerequisites

### Required Tools

- **Docker** (v24.0+) & **Docker Compose** (v2.20+)
- **Go** (v1.21+)
- **Templ** CLI (`go install github.com/a-h/templ/cmd/templ@latest`)
- **Node.js** (v20+) & **npm** (for E2E tests)
- **Git** (v2.40+)

### Recommended Tools

- **VS Code** with extensions:
  - Go (golang.go)
  - Templ (a-h.templ)
  - Docker (ms-azuretools.vscode-docker)
  - Playwright Test (ms-playwright.playwright)
- **GitHub CLI** (`gh`) for PR management
- **Traefik Dashboard** access (http://localhost:8080)

### System Requirements

- **RAM**: 8GB minimum, 16GB recommended
- **Disk**: 10GB free space (Docker images + build artifacts)
- **OS**: Linux, macOS, or Windows (via WSL2)

---

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/mikejsmith1985/devsmith-modular-platform.git
cd devsmith-modular-platform
```

### 2. Setup Environment

```bash
# Copy example environment file
cp .env.example .env

# Edit .env and add your GitHub OAuth credentials
# Get them from: https://github.com/settings/developers
nano .env
```

**Required environment variables:**
```bash
# GitHub OAuth (create at https://github.com/settings/developers)
GITHUB_CLIENT_ID=your_client_id_here
GITHUB_CLIENT_SECRET=your_secret_here

# JWT Secret (generate with: openssl rand -hex 32)
JWT_SECRET=your_jwt_secret_here

# Redis (default works for local development)
REDIS_URL=redis://redis:6379

# PostgreSQL (default works for local development)
DATABASE_URL=postgresql://devsmith:devsmith@postgres:5432/devsmith
```

### 3. Start Services

```bash
# Start all services
docker-compose up -d

# Watch logs
docker-compose logs -f portal review logs analytics

# Check health
docker-compose ps
```

### 4. Verify Installation

```bash
# Run health check script
bash scripts/health-check-cli.sh

# Expected output:
# ✓ PostgreSQL: healthy
# ✓ Redis: healthy
# ✓ Traefik Gateway: healthy
# ✓ Portal: healthy
# ✓ Review: healthy
# ✓ Logs: healthy
# ✓ Analytics: healthy
```

### 5. Access Platform

- **Platform**: http://localhost:3000
- **Traefik Dashboard**: http://localhost:8080
- **Login**: Click "Login with GitHub"

---

## Architecture Overview

### Service Architecture

```
User → Traefik Gateway (port 3000)
         ↓
    ┌────┴────┬─────────┬──────────┬──────────┐
    ↓         ↓         ↓          ↓          ↓
 Portal    Review    Logging   Analytics   (Future)
   ↓         ↓         ↓          ↓
PostgreSQL (schemas: portal, reviews, logs, analytics)
   &
Redis (sessions, caching)
```

### Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Gateway** | Traefik | Reverse proxy, automatic HTTPS |
| **Backend** | Go 1.21+ | Service implementation |
| **Templates** | Templ | Type-safe HTML templates |
| **Interactivity** | HTMX + Alpine.js | Dynamic UI without heavy JS |
| **Styling** | TailwindCSS + DaisyUI | Unified design system |
| **Database** | PostgreSQL 15+ | Data persistence |
| **Cache/Sessions** | Redis | Session storage, pub/sub |
| **Testing** | Playwright + Percy | E2E and visual regression tests |

### Key Directories

```
devsmith-modular-platform/
├── apps/                    # Service-specific code
│   ├── portal/             # Authentication, dashboard
│   ├── review/             # Code analysis
│   ├── logs/               # Log ingestion
│   └── analytics/          # Log analysis
├── cmd/                     # Service entry points
│   ├── portal/main.go
│   ├── review/main.go
│   └── ...
├── internal/                # Shared internal packages
│   ├── middleware/         # Auth, logging middleware
│   ├── session/            # Redis session store
│   ├── security/           # JWT, crypto
│   └── ui/                 # Shared UI components
├── tests/
│   ├── e2e/                # End-to-end tests (Playwright)
│   └── integration/        # Integration tests (Go)
├── docs/                    # Documentation
│   ├── ACCESSIBILITY.md    # WCAG 2.1 AA guidelines
│   ├── openapi.yaml        # API specification
│   └── ...
└── docker-compose.yml       # Service orchestration
```

---

## Development Workflow

### Branch Strategy

```
main (production)
  ↑
development (integration)
  ↑
feature/XXX-description (your work)
```

### Creating a Feature

```bash
# 1. Sync with latest development
git checkout development
git pull origin development

# 2. Create feature branch
git checkout -b feature/123-add-oauth-login

# 3. Make changes (see Code Standards below)

# 4. Commit with conventional commits
git add -A
git commit -m "feat(portal): add GitHub OAuth login

Implemented:
- OAuth callback handler
- JWT token generation
- Session storage in Redis

Tests:
- 15/15 unit tests passing
- E2E test: login flow with GitHub

Closes #123"

# 5. Push and create PR
git push origin feature/123-add-oauth-login
gh pr create --base development --title "feat(portal): GitHub OAuth login" --body "Closes #123"
```

### Conventional Commit Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding/updating tests
- `refactor`: Code refactoring
- `style`: Formatting changes
- `chore`: Build/tooling changes

**Scopes:** `portal`, `review`, `logs`, `analytics`, `docs`, `ci`, `docker`

**Examples:**
```bash
feat(review): add critical reading mode
fix(logs): resolve WebSocket connection timeout
docs(api): update OpenAPI specification
test(portal): add OAuth E2E tests
refactor(analytics): extract query builder
```

---

## Running Tests

### Unit Tests (Go)

```bash
# Run all unit tests
go test ./...

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test ./apps/portal/handlers/...

# Run with race detection
go test -race ./...
```

### E2E Tests (Playwright)

```bash
# Install Playwright browsers (first time only)
npx playwright install

# Run all E2E tests
npm test

# Run specific test file
npm test -- tests/e2e/portal/login.spec.ts

# Run accessibility tests
npm run test:a11y

# Run with UI mode (interactive debugging)
npx playwright test --ui

# Run in headed mode (see browser)
npx playwright test --headed

# Generate test report
npx playwright show-report
```

### Visual Regression Tests (Percy)

```bash
# Local snapshots (no Percy API)
npm run test:visual:local

# Percy snapshots (requires PERCY_TOKEN)
export PERCY_TOKEN=your_percy_token
npm run test:visual
```

### Integration Tests

```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./tests/integration/...

# Cleanup
docker-compose -f docker-compose.test.yml down -v
```

---

## Code Standards

### Go Code Style

**File Organization:**
```
apps/portal/
├── main.go                 # Entry point
├── handlers/               # HTTP handlers
│   ├── auth_handler.go
│   └── auth_handler_test.go
├── services/               # Business logic
│   ├── auth_service.go
│   └── auth_service_test.go
├── templates/              # Templ templates
│   ├── layout.templ
│   └── dashboard.templ
└── static/                 # Static assets
    ├── css/
    └── js/
```

**Naming Conventions:**
```go
// Unexported (private)
func validateToken(token string) error { }
var sessionStore *RedisStore

// Exported (public)
func ValidateToken(token string) error { }
var SessionStore *RedisStore

// Constants
const MaxRetries = 3
const API_BASE_URL = "https://api.example.com"
```

**Error Handling:**
```go
// ✅ Good - explicit error checking
user, err := getUserByID(ctx, id)
if err != nil {
    log.Error().Err(err).Int("user_id", id).Msg("Failed to get user")
    return nil, fmt.Errorf("get user: %w", err)
}

// ❌ Bad - ignoring errors
user, _ := getUserByID(ctx, id)

// ❌ Bad - generic error message
if err != nil {
    return nil, errors.New("error getting user")
}
```

### Templ Templates

```go
// apps/portal/templates/dashboard.templ

package templates

templ Dashboard(user *models.User, apps []models.App) {
    @Layout("Dashboard", user) {
        <main id="main-content" class="container mx-auto p-6">
            <h1 class="text-3xl font-bold mb-6">
                Welcome, { user.Username }
            </h1>
            
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                for _, app := range apps {
                    @AppCard(app)
                }
            </div>
        </main>
    }
}

templ AppCard(app models.App) {
    <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
            <h2 class="card-title">
                <i class={ "bi", app.Icon }></i>
                { app.Name }
            </h2>
            <p>{ app.Description }</p>
            <div class="card-actions justify-end">
                if app.Status == "ready" {
                    <a href={ templ.URL(app.URL) } class="btn btn-primary">
                        Launch
                    </a>
                } else {
                    <span class="badge badge-warning">Coming Soon</span>
                }
            </div>
        </div>
    </div>
}
```

**After editing `.templ` files, regenerate:**
```bash
templ generate
```

### HTMX Patterns

```html
<!-- Partial page update -->
<button
    hx-get="/api/logs?level=ERROR"
    hx-target="#log-list"
    hx-swap="innerHTML"
    class="btn btn-primary"
>
    Filter Errors
</button>

<!-- Form submission with loading state -->
<form
    hx-post="/api/review/sessions"
    hx-target="#result"
    hx-indicator="#loading"
>
    <input type="text" name="title" required />
    <button type="submit" class="btn">Create Session</button>
    <span id="loading" class="loading loading-spinner htmx-indicator"></span>
</form>

<!-- WebSocket updates -->
<div
    hx-ext="ws"
    ws-connect="/ws/logs?service=portal"
>
    <div id="log-stream"></div>
</div>
```

---

## Common Tasks

### Adding a New API Endpoint

1. **Define handler** (`apps/SERVICE/handlers/my_handler.go`):
```go
func (h *MyHandler) GetData(c *gin.Context) {
    id := c.Param("id")
    
    data, err := h.service.GetData(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
        return
    }
    
    c.JSON(http.StatusOK, data)
}
```

2. **Register route** (`cmd/SERVICE/main.go`):
```go
router.GET("/api/my-service/data/:id", myHandler.GetData)
```

3. **Write tests** (`apps/SERVICE/handlers/my_handler_test.go`):
```go
func TestGetData_Success(t *testing.T) {
    mockService := new(MockMyService)
    handler := handlers.NewMyHandler(mockService)
    
    router := gin.Default()
    router.GET("/api/my-service/data/:id", handler.GetData)
    
    mockService.On("GetData", mock.Anything, "123").Return(&models.Data{ID: "123"}, nil)
    
    req := httptest.NewRequest("GET", "/api/my-service/data/123", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    mockService.AssertExpectations(t)
}
```

4. **Update OpenAPI** (`docs/openapi.yaml`):
```yaml
  /api/my-service/data/{id}:
    get:
      tags:
        - MyService
      summary: Get data by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Data found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Data'
```

### Adding a Database Migration

1. **Create migration file** (`apps/SERVICE/db/migrations/YYYYMMDD_NNN_description.sql`):
```sql
-- 20251105_001_add_user_preferences.sql

-- Up migration
CREATE TABLE portal.user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES portal.users(id) ON DELETE CASCADE,
    theme VARCHAR(20) DEFAULT 'auto',
    language VARCHAR(10) DEFAULT 'en',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id)
);

CREATE INDEX idx_user_preferences_user_id ON portal.user_preferences(user_id);

-- Down migration (optional, for rollback)
-- DROP TABLE portal.user_preferences;
```

2. **Run migration:**
```bash
# Via docker-compose (automatic on service start)
docker-compose up -d portal

# Or manually
docker-compose exec -T postgres psql -U devsmith -d devsmith -f /migrations/20251105_001_add_user_preferences.sql
```

3. **Verify migration:**
```bash
docker-compose exec -T postgres psql -U devsmith -d devsmith -c "\d portal.user_preferences"
```

### Rebuilding a Service

```bash
# Rebuild specific service
docker-compose up -d --build portal

# Rebuild all services
docker-compose up -d --build

# Force rebuild without cache
docker-compose build --no-cache portal
docker-compose up -d portal
```

---

## Troubleshooting

### Services Won't Start

**Problem:** `docker-compose up -d` fails

**Solutions:**
```bash
# Check logs
docker-compose logs portal review logs analytics

# Check health
docker-compose ps

# Clean restart
docker-compose down
docker-compose up -d

# Nuclear option (removes volumes)
docker-compose down -v
docker-compose up -d
```

### Database Connection Errors

**Problem:** `pq: relation does not exist`

**Solutions:**
```bash
# Check if migrations ran
docker-compose logs portal | grep migration

# Run migrations manually
bash scripts/run-migrations.sh

# Verify schema exists
docker-compose exec -T postgres psql -U devsmith -d devsmith -c "\dn"
```

### Template Compilation Errors

**Problem:** `.templ` changes not reflected

**Solutions:**
```bash
# Regenerate templates
templ generate

# Rebuild service
docker-compose up -d --build portal

# Check for Templ syntax errors
templ generate --watch  # Shows compilation errors
```

### WebSocket Connection Failures

**Problem:** Real-time features not working

**Solutions:**
```bash
# Check Traefik routing
curl -H "Connection: Upgrade" -H "Upgrade: websocket" http://localhost:3000/ws/logs

# Check Redis pub/sub
docker-compose exec redis redis-cli PUBSUB CHANNELS

# Verify WebSocket handler
docker-compose logs logs | grep WebSocket
```

### Test Failures

**Problem:** E2E tests failing

**Solutions:**
```bash
# Run in headed mode to see what's happening
npx playwright test --headed

# Debug specific test
npx playwright test --debug tests/e2e/portal/login.spec.ts

# Check if services are running
docker-compose ps

# Re-install Playwright browsers
npx playwright install --force
```

---

## Resources

### Documentation

- **Architecture**: [ARCHITECTURE.md](../ARCHITECTURE.md)
- **Accessibility**: [docs/ACCESSIBILITY.md](ACCESSIBILITY.md)
- **API Reference**: [docs/openapi.yaml](openapi.yaml)
- **Percy Setup**: [docs/PERCY_SETUP.md](PERCY_SETUP.md)

### External References

- **Go**: https://go.dev/doc/
- **Templ**: https://templ.guide/
- **HTMX**: https://htmx.org/docs/
- **Playwright**: https://playwright.dev/
- **Traefik**: https://doc.traefik.io/traefik/
- **Redis**: https://redis.io/docs/

### Getting Help

- **Slack**: #devsmith-dev (internal)
- **GitHub Issues**: https://github.com/mikejsmith1985/devsmith-modular-platform/issues
- **Wiki**: https://github.com/mikejsmith1985/devsmith-modular-platform/wiki

### Code Review Checklist

Before creating a PR, verify:

- [ ] All tests pass (`go test ./... && npm test`)
- [ ] Code follows style guide (run `gofmt`, `golangci-lint`)
- [ ] Templ templates regenerated (`templ generate`)
- [ ] API documentation updated (`docs/openapi.yaml`)
- [ ] Manual testing completed (try the feature in browser)
- [ ] Accessibility verified (`npm run test:a11y`)
- [ ] Commit messages follow conventional commits format
- [ ] Branch named correctly (`feature/XXX-description`)
- [ ] PR description includes testing evidence

---

## What's Next?

Now that you're set up, try:

1. **Fix a "good first issue"** - Check GitHub issues labeled `good-first-issue`
2. **Improve documentation** - Found something unclear? Submit a PR!
3. **Add a test** - Increase coverage for untested code paths
4. **Join the team** - Attend weekly dev sync (Fridays 2pm PST)

**Welcome to DevSmith! 🚀**
