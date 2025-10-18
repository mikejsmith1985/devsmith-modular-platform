# DevSmith Modular Platform - Architecture

**Version:** 1.0
**Status:** Planning Phase
**Last Updated:** 2025-10-18

---

## Table of Contents
1. [Executive Summary](#executive-summary)
2. [Architecture Principles](#architecture-principles)
3. [System Overview](#system-overview)
4. [Technology Stack](#technology-stack)
5. [Service Architecture](#service-architecture)
6. [Data Architecture](#data-architecture)
7. [Authentication & Authorization](#authentication--authorization)
8. [API Design](#api-design)
9. [Real-Time Communication](#real-time-communication)
10. [Deployment Architecture](#deployment-architecture)
11. [Security Architecture](#security-architecture)
12. [Monitoring & Logging](#monitoring--logging)
13. [DevSmith Coding Standards](#devsmith-coding-standards)
14. [Development Workflow](#development-workflow)
15. [CI/CD & Automation](#cicd--automation)
16. [Implementation Phases](#implementation-phases)
17. [Decision Log](#decision-log)

---

## Executive Summary

### Purpose
The DevSmith Modular Platform is a comprehensive learning and development platform featuring modular apps for code review, logging, analytics, and autonomous building.

### Key Design Goals
- **True Modularity**: Apps operate independently, no forced dependencies
- **Developer Experience**: One-click installation, excellent debugging
- **AI-First**: Local LLM support via Ollama with online API fallback
- **Production-Ready**: Gateway architecture, proper auth, comprehensive logging

### Current Status
- **Phase:** Initial Planning
- **Branch:** feature/initial-setup
- **Implementation:** Not started
- **Documentation:** Complete (Requirements, Roles, TDD, Lessons Learned)

---

## Architecture Principles

### 1. Gateway-First Design
**Rationale:** Learned from previous platform that adding gateway as afterthought breaks everything.

**Implementation:**
- All services accessible through nginx reverse proxy on port 3000
- No direct port access in application code
- Gateway configured before any app development
- Single origin for shared authentication

### 2. True Modularity
**Rationale:** Apps must function independently or platform isn't truly modular.

**Implementation:**
- Each app has isolated database schema
- Clear API contracts between services
- Apps testable in complete isolation
- Optional inter-app features, not required dependencies

### 3. Build Order Discipline
**Rationale:** Foundation must be stable before building complex features.

**Build Sequence:**
1. Portal (navigation and app browser)
2. Logging (monitors all subsequent development)
3. Analytics (analyzes logs from development)
4. Review (benefits from monitoring infrastructure)
5. Build (most complex, needs stable foundation)

### 4. Never Assume, Always Verify
**Rationale:** Assumptions in previous platform caused cascading failures.

**Implementation:**
- Every integration claim requires test evidence
- Code reviews verify actual implementation
- Documentation reflects reality, not aspirations
- Three-strikes rule: 3 failed fixes = reassess approach

### 5. Configuration Over Hardcoding
**Rationale:** Hardcoded values in previous platform caused maintenance nightmare.

**Implementation:**
- All URLs, ports, keys in environment variables
- .env.example documents all required config
- Startup validation fails fast if config missing
- Single source of truth for service locations

---

## System Overview

### High-Level Architecture
```
[To be designed - Gateway-first architecture diagram]

User → Nginx Gateway (port 3000)
         ↓
    ┌────┴────┬─────────┬──────────┬──────────┐
    ↓         ↓         ↓          ↓          ↓
 Portal   Review    Logging   Analytics   Build
Frontend  Frontend  Frontend  Frontend   Frontend
    ↓         ↓         ↓          ↓          ↓
 Portal   Review    Logging   Analytics   Build
Backend   Backend   Backend   Backend    Backend
    └─────────┴─────────┴──────────┴──────────┘
                      ↓
              PostgreSQL Database
              (Isolated Schemas)
```

### Service Inventory
| Service | Purpose | Port (Dev) | Gateway Path | Status |
|---------|---------|------------|--------------|--------|
| Nginx Gateway | Reverse proxy | 3000 | / | Not implemented |
| Portal Frontend | Main UI, navigation | TBD | / | Not implemented |
| Portal Backend | Auth, user mgmt | TBD | /api/platform/ | Not implemented |
| Review Frontend | Code review UI | TBD | /review/ | Not implemented |
| Review Backend | Code analysis | TBD | /api/review/ | Not implemented |
| Logs Frontend | Log monitoring UI | TBD | /logs/ | Not implemented |
| Logs Backend | Log ingestion | TBD | /api/logs/ | Not implemented |
| Analytics Frontend | Analytics UI | TBD | /analytics/ | Not implemented |
| Analytics Backend | Data analysis | TBD | /api/analytics/ | Not implemented |
| Build Frontend | Terminal UI | TBD | /build/ | Not implemented |
| Build Backend | Code execution | TBD | /api/build/ | Not implemented |
| PostgreSQL | Database | 5432 | N/A | Not implemented |

---

## Technology Stack

### Frontend/Backend (Unified)
- **Language:** Go 1.21+
- **Web Framework:** Gin (or Echo as alternative)
- **Templating:** Templ (type-safe Go templates)
- **Interactivity:** HTMX + Alpine.js (minimal JavaScript)
- **Styling:** TailwindCSS + DaisyUI components
- **WebSocket:** Go's native net/http WebSocket support
- **Testing:** Go's built-in testing + testify

**Rationale:**
- **No Node.js = No V8 crashes** (eliminates build-time crashes from previous platform)
- Go compiles to single binary (5-20MB vs 500MB+ Node containers)
- 10-50x faster API performance than Node.js/Python
- Built-in concurrency (goroutines) perfect for WebSocket and real-time features
- Memory efficient (50-100MB runtime vs 500MB+ for Node)
- HTMX provides React-like interactivity without JavaScript framework complexity
- Templ catches template errors at compile time (type safety)
- Single language for frontend + backend reduces context switching

**Key Benefits:**
✅ **Zero V8 workarounds needed**
✅ Docker builds in 30 seconds (vs 5+ minutes with Vite)
✅ Hot reload with Air tool (same experience as HMR)
✅ Simpler deployment (copy binary, no npm install)
✅ Lower hosting costs (smaller images, less memory)

### Database
- **Primary:** PostgreSQL 15+
- **Driver:** pgx (fastest Go PostgreSQL driver)
- **Migrations:** golang-migrate/migrate
- **Caching:** Redis (for sessions, rate limiting)
- **Schema Strategy:** Isolated schemas per app, federated queries where needed

**Rationale:**
- PostgreSQL: ACID compliance, JSON support, mature
- pgx: Native Go driver, better performance than database/sql
- golang-migrate: Simple migration tool, works with any SQL
- Redis: Fast caching, pub/sub for real-time features
- Schema isolation: Maintains modularity while allowing cross-app queries

### Infrastructure
- **Containerization:** Docker + Docker Compose
- **Base Images:** golang:1.21-alpine (build), alpine:latest (runtime)
- **Reverse Proxy:** Nginx
- **CI/CD:** GitHub Actions
- **Monitoring:** (To be determined - options: Prometheus + Grafana)

**Container Strategy:**
- Multi-stage Docker builds (compile in golang image, run in alpine)
- Final images: 10-20MB per service
- Build time: ~30 seconds per service
- No npm/pip install in containers

### AI/LLM Integration
- **Local:** Ollama (for offline operation)
- **Online:** Anthropic Claude API, OpenAI API (user-provided keys)
- **Go Client:** github.com/anthropics/anthropic-sdk-go
- **HTTP Client:** Native Go http.Client with proper timeouts

**Rationale:**
- Ollama: Privacy, offline capability, no API costs
- Multiple APIs: Flexibility, no vendor lock-in
- Native Go HTTP: No SDK version compatibility issues
- Proper timeout handling: Go's context package prevents hanging requests

### Development Tools

#### Local Development
- **Hot Reload:** Air (Go file watcher, automatic rebuild)
- **Linting:** golangci-lint (comprehensive linter)
- **Formatting:** gofmt (standard Go formatter)
- **API Docs:** Swagger/OpenAPI via swaggo/swag
- **Dependency Management:** Go modules (built-in)

#### AI Development Tools (Hybrid Approach)

**Primary Implementation Agent: OpenHands + Ollama**
- **Role:** Autonomous code generation and implementation (70-80% of work)
- **Setup:**
  - OpenHands: `pip install openhands` (autonomous agent framework)
  - Ollama: Local LLM runtime (privacy, no API costs)
  - Recommended models: `deepseek-coder-v2:16b` or `codellama:34b`
- **Capabilities:**
  - Fully autonomous feature implementation
  - TDD workflow (write tests → implement → verify)
  - File creation/editing, git operations, test execution
  - Browser automation for testing
  - Checkpoint/resume on crash or interruption
- **System Requirements:**
  - Minimum: 16GB RAM, 8 CPU cores
  - Recommended: 32GB RAM, 16+ CPU cores (met by Dell G16 7630)
  - GPU: Optional but recommended (RTX 4070 ideal for 16B+ models)

**Architecture & Review: Claude (via API)**
- **Role:** High-level architecture, strategic code review (10-15% of work)
- **Interface:** Claude Code CLI (this tool)
- **Capabilities:**
  - 200K context window (can review entire codebase)
  - Architecture design and API contracts
  - Database schema design
  - Strategic PR reviews
  - Complex problem solving
- **Limitations:**
  - Subject to V8 crashes (mitigated by recovery hooks)
  - Cannot execute code directly
  - Sessions should be <30 minutes

**IDE Assistant: GitHub Copilot**
- **Role:** Real-time autocomplete during manual coding (5-10% of work)
- **Interface:** VS Code extension
- **Capabilities:**
  - Inline code suggestions
  - Boilerplate generation
  - Quick refactorings
- **Limitations:**
  - No autonomous workflow
  - Limited context (single file)

**Crash Recovery Mechanisms:**
- `.claude/hooks/` - Automated recovery scripts
  - `session-logger.sh` - Logs all actions to markdown
  - `git-recovery.sh` - Auto-commits to recovery branches
  - `recovery-helper.sh` - Interactive recovery tool
- Todo list (`.claude/todos.json`) - Persistent task tracking
- Recovery branches (`claude-recovery-YYYYMMDD`) - 7-day retention

**Benefits of Hybrid Approach:**
- ✅ 80% of work is crash-proof (OpenHands runs independently)
- ✅ No API costs for implementation (Ollama runs locally)
- ✅ No rate limits (can run 24/7)
- ✅ Claude focuses on high-value architecture work
- ✅ Parallel development (OpenHands implements while Claude reviews)

**System Requirements Met:**
- **Your System:** Dell G16 7630 (i9-13900HX, 32GB RAM, RTX 4070)
- **Assessment:** Excellent for running multiple large Ollama models simultaneously
- **Can Run:** Llama 3.1 70B quantized + CodeLlama 34B concurrently

### Why Not React/Node?

**Problems with previous platform:**
1. V8 JavaScript engine crashes during Docker builds
2. Required workarounds: `NODE_OPTIONS="--jitless"`, `DOCKER_BUILDKIT=0`
3. Large containers (500MB+) with slow build times (5+ minutes)
4. High memory usage (500MB+ per service)
5. Complex build tooling (Webpack/Vite, npm, node_modules)

**Go eliminates these issues:**
1. No V8 engine = no crashes
2. No workarounds needed
3. Tiny containers (15MB) with fast builds (30 seconds)
4. Low memory usage (50-100MB per service)
5. Simple tooling (go build, done)

---

## Service Architecture

### Portal Service
**Purpose:** Main entry point, navigation, app management

**Responsibilities:**
- User authentication (GitHub OAuth)
- App browser and launcher
- Session management
- User profile

**Dependencies:**
- PostgreSQL (users table)
- GitHub OAuth API
- No other services required

**API Endpoints:**
- `POST /api/auth/github/login` - Initiate OAuth
- `GET /api/auth/github/callback` - OAuth callback
- `GET /api/auth/me` - Get current user
- `POST /api/auth/logout` - Logout
- `GET /api/apps` - List available apps

### Review Service
**Purpose:** AI-driven code review with multiple reading modes

**Responsibilities:**
- Code import (GitHub, paste)
- Five reading modes (Previewing, Skimming, Scanning, Detailed, Critical)
- Syntax highlighting
- Real-time collaboration
- Pull request integration

**Dependencies:**
- PostgreSQL (reviews schema)
- Ollama or Claude API
- Logging service (optional, for telemetry)

**API Endpoints:**
- `POST /api/review/create` - Create review session
- `POST /api/review/{id}/analyze` - Run AI analysis
- `GET /api/review/{id}` - Get review results
- `WS /api/review/{id}/collaborate` - Real-time collaboration

### Logging Service
**Purpose:** Real-time log tracking and centralized logging

**Responsibilities:**
- Log ingestion from all services
- Real-time streaming via WebSocket
- Tag-based filtering
- Log storage and retrieval
- AI-driven context analysis (optional)

**Dependencies:**
- PostgreSQL (logs schema)
- Redis (WebSocket pub/sub)
- Ollama (optional, for log analysis)

**API Endpoints:**
- `POST /api/logs` - Ingest log entry
- `GET /api/logs` - Query logs (with filters)
- `GET /api/logs/stats` - Log statistics
- `WS /ws/logs` - Real-time log streaming

### Analytics Service
**Purpose:** Log analysis and insights

**Responsibilities:**
- Frequency analysis
- Trend detection
- Anomaly detection
- Performance metrics
- Exportable reports (CSV, JSON)

**Dependencies:**
- PostgreSQL (analytics schema, read from logs schema)
- Logging service (data source)

**API Endpoints:**
- `GET /api/analytics/trends` - Trend analysis
- `GET /api/analytics/anomalies` - Detect anomalies
- `GET /api/analytics/top-issues` - Most common issues
- `GET /api/analytics/export` - Export report

### Build Service (Phase 2)
**Purpose:** Terminal interface and autonomous coding

**Responsibilities:**
- Terminal emulation
- Cloud CLI support
- Copilot CLI integration
- OpenHands autonomous coding (Phase 2)
- Real-time collaboration

**Dependencies:**
- PostgreSQL (build sessions schema)
- Logging service (terminal output capture)
- Ollama (for autonomous coding)

**API Endpoints:**
- `POST /api/build/session` - Create terminal session
- `WS /api/build/terminal` - Terminal I/O stream
- `POST /api/build/autonomous` - Start autonomous coding task

---

## Data Architecture

### Database Design Principles
1. **Schema Isolation:** Each app has its own schema
2. **Federated Queries:** Cross-schema queries allowed via views
3. **No Shared Tables:** No tables accessed by multiple apps directly
4. **Clear Ownership:** Each schema owned by one service

### Schema Layout
```
PostgreSQL Database: devsmith_platform
├── Schema: portal
│   ├── users (id, github_id, github_username, email, created_at)
│   └── sessions (id, user_id, token, expires_at)
├── Schema: review
│   ├── reviews (id, user_id, title, code_content, status, created_at)
│   ├── review_segments (id, review_id, segment_index, line_start, line_end)
│   └── explanations (id, segment_id, content, reading_mode, created_at)
├── Schema: logs
│   ├── log_entries (id, timestamp, level, message, source, context, tags)
│   └── log_stats (id, date, level, count)
├── Schema: analytics
│   ├── trends (id, metric, value, timestamp)
│   └── anomalies (id, log_entry_id, detected_at, severity)
└── Schema: build (Phase 2)
    ├── sessions (id, user_id, status, created_at)
    └── commands (id, session_id, command, output, timestamp)
```

### Data Relationships
- **Within Schema:** Foreign keys enforced
- **Cross-Schema:** No foreign keys, joined via application logic or views
- **User References:** All schemas may reference portal.users via user_id (no FK)

### Migration Strategy
- Alembic per service
- Separate version tables per schema
- Independent migration histories
- Coordinated releases for cross-schema changes

---

## Authentication & Authorization

### Authentication Strategy
**Method:** GitHub OAuth 2.0

**Rationale:**
- No password management
- Access to user's GitHub repositories
- Industry standard, well-documented
- Previous platform proved feasibility

**Flow:**
```
1. User clicks "Login with GitHub"
2. Redirect to GitHub authorization page
3. User approves DevSmith Platform OAuth app
4. GitHub redirects to /auth/github/callback?code=...
5. Backend exchanges code for access token
6. Backend creates/updates user in database
7. Backend generates JWT with github_access_token
8. Frontend stores JWT in localStorage
9. All subsequent requests include JWT in Authorization header
```

### Token Structure
```json
{
  "user_id": 123,
  "github_id": 456789,
  "github_username": "user",
  "github_access_token": "gho_xxx",
  "exp": 1234567890
}
```

**Field Consistency Rule:** All services MUST use `github_access_token` (not `github_token`)

### Authorization Levels
- **Public:** No authentication required (Review, Logs, Analytics frontends in dev)
- **Authenticated:** Valid JWT required (Portal, user-specific data)
- **GitHub Scopes:** `read:user`, `user:email`, `repo` (for code access)

### Session Management
- JWT stored in localStorage (key: `devsmith_token`)
- Token expiry: 7 days
- Refresh tokens: Not implemented in Phase 1
- Logout: Clear localStorage, invalidate server session

### Security Considerations
- HTTPS required in production
- JWT secret in environment variable
- GitHub OAuth secrets not committed to git
- CORS configured for gateway origin only
- Rate limiting on authentication endpoints

---

## API Design

### REST Conventions
- **GET:** Retrieve resources (idempotent)
- **POST:** Create resources or trigger actions
- **PUT:** Full resource update
- **PATCH:** Partial resource update
- **DELETE:** Remove resources

### URL Structure
```
/api/{service}/{resource}/{id?}/{action?}

Examples:
GET  /api/review/reviews          - List reviews
POST /api/review/reviews          - Create review
GET  /api/review/reviews/123      - Get review by ID
POST /api/review/reviews/123/analyze - Trigger analysis
GET  /api/logs?level=ERROR        - Query logs with filter
```

### Request/Response Format
**Content-Type:** `application/json`

**Success Response:**
```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "timestamp": "2025-10-18T12:00:00Z",
    "request_id": "uuid"
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "User-friendly error message",
    "details": { ... },
    "stack_trace": "..." // Only in development
  },
  "meta": {
    "timestamp": "2025-10-18T12:00:00Z",
    "request_id": "uuid"
  }
}
```

### Status Codes
- **200 OK:** Successful GET/PUT/PATCH
- **201 Created:** Successful POST (resource created)
- **204 No Content:** Successful DELETE
- **400 Bad Request:** Client error (validation, malformed)
- **401 Unauthorized:** Missing or invalid authentication
- **403 Forbidden:** Authenticated but insufficient permissions
- **404 Not Found:** Resource doesn't exist
- **500 Internal Server Error:** Server error

### Error Handling Standards
1. **Never return error strings as data** (Lesson from old platform)
2. **Raise exceptions, don't catch and return** (Let middleware handle)
3. **Include context in logs** (request_id, user_id, resource_id)
4. **User-friendly messages** (No stack traces in production)
5. **Actionable guidance** (Tell user how to fix)

### Pagination
```
GET /api/logs?page=1&limit=50&offset=0

Response includes:
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 1234,
    "has_more": true
  }
}
```

### API Versioning
- Version in URL: `/api/v1/...`
- Current version: v1
- Breaking changes require new version

---

## Real-Time Communication

### WebSocket Architecture
**Use Cases:**
- Real-time log streaming (Logs app)
- Collaborative code review (Review app)
- Terminal I/O (Build app)
- Live notifications

### Connection Pattern
```javascript
// Frontend
const ws = new WebSocket('ws://localhost:3000/ws/logs');
ws.onopen = () => console.log('Connected');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  handleMessage(data);
};
ws.onerror = (error) => console.error('WebSocket error:', error);
ws.onclose = () => console.log('Disconnected');
```

### Message Format
```json
{
  "type": "new_log",
  "data": {
    "id": 123,
    "level": "ERROR",
    "message": "Database connection failed",
    "timestamp": "2025-10-18T12:00:00Z"
  }
}
```

### Backend Implementation
- FastAPI WebSocket endpoints
- Redis pub/sub for multi-instance support
- Heartbeat every 30 seconds
- Automatic reconnection on client side

### Error Handling
- Connection failures: Exponential backoff retry
- Message parsing errors: Log and skip
- Server errors: Graceful disconnect, notify user

---

## Deployment Architecture

### Development Environment
```yaml
# docker-compose.yml structure (to be implemented)
services:
  nginx-gateway:     # Port 3000
  portal-frontend:   # Internal only
  portal-backend:    # Internal only
  review-frontend:   # Internal only
  review-backend:    # Internal only
  logs-frontend:     # Internal only
  logs-backend:      # Internal only
  analytics-frontend: # Internal only
  analytics-backend:  # Internal only
  postgres:          # Port 5432
  redis:             # Port 6379
```

### Container Strategy
- **Frontend:** Multi-stage build (build → nginx)
- **Backend:** Python 3.11-slim base image
- **Database:** Official postgres:15-alpine
- **Redis:** Official redis:7-alpine

### Volume Management
- **postgres-data:** Database persistence
- **redis-data:** Cache persistence (optional)
- **logs:** Log file storage (optional)

### Health Checks
- All services expose `/health` endpoint
- Nginx checks before routing
- Docker health checks configured
- Startup dependencies via `depends_on`

### Environment Configuration
```bash
# .env.example structure (to be created)
# Gateway
NGINX_PORT=3000

# Portal
PORTAL_FRONTEND_PORT=5173
PORTAL_BACKEND_PORT=8000
GITHUB_CLIENT_ID=xxx
GITHUB_CLIENT_SECRET=xxx
JWT_SECRET=xxx

# Review
REVIEW_FRONTEND_PORT=5174
REVIEW_BACKEND_PORT=8001
CLAUDE_API_KEY=xxx  # Optional, for online API

# Logs
LOGS_FRONTEND_PORT=8080
LOGS_BACKEND_PORT=8002

# Analytics
ANALYTICS_FRONTEND_PORT=8081
ANALYTICS_BACKEND_PORT=8003

# Database
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=devsmith
POSTGRES_PASSWORD=xxx
POSTGRES_DB=devsmith_platform

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
```

### Production Considerations (Phase 3+)
- HTTPS with Let's Encrypt
- Separate database server
- Redis cluster for HA
- Container orchestration (K8s or ECS)
- CDN for static assets
- Log aggregation (ELK or similar)

---

## Security Architecture

### Threat Model
- **Public Code Review:** Anyone can review code, but no data persistence without auth
- **Authenticated Portal:** User accounts protected by GitHub OAuth
- **API Security:** JWT validation, rate limiting, input sanitization
- **Database Security:** Principle of least privilege, connection pooling

### Security Controls
1. **Authentication:** GitHub OAuth with proper scope limitation
2. **Authorization:** JWT validation on protected endpoints
3. **Input Validation:** Pydantic models, sanitization
4. **SQL Injection:** SQLAlchemy ORM, parameterized queries
5. **XSS Prevention:** React escapes by default, CSP headers
6. **CSRF:** SameSite cookies, CORS restrictions
7. **Rate Limiting:** Per-IP and per-user limits
8. **Secrets Management:** Environment variables, never committed

### Dependency Security
- Dependabot alerts enabled
- Regular dependency updates
- Pin major versions, allow minor/patch updates
- Security audit before major releases

### Data Privacy
- No sensitive data logged
- User data encrypted at rest (database encryption)
- GitHub tokens stored securely, never logged
- User consent for AI analysis (future consideration)

---

## Monitoring & Logging

### Logging Strategy
**Infrastructure:** Centralized logging service (built into platform)

**Log Levels:**
- **DEBUG:** Development diagnostics
- **INFO:** Normal operation events
- **WARNING:** Unexpected but handled situations
- **ERROR:** Failures requiring attention
- **CRITICAL:** System-level failures

**Structured Logging Format:**
```json
{
  "timestamp": "2025-10-18T12:00:00.000Z",
  "level": "ERROR",
  "source": "review-backend",
  "message": "Claude API call failed",
  "context": {
    "review_id": 123,
    "user_id": 456,
    "error_type": "APIConnectionError",
    "duration_ms": 5000
  },
  "stack_trace": "...",
  "tags": ["api", "claude", "timeout"]
}
```

### What to Log
✅ **DO Log:**
- API requests/responses (sanitized)
- Authentication events
- Database operations (timing)
- External API calls
- Error conditions with full context
- Performance metrics

❌ **DON'T Log:**
- Passwords or secrets
- Full GitHub tokens
- User's private code (only metadata)
- Personally identifiable information

### Metrics to Track
- Request count by endpoint
- Response times (p50, p95, p99)
- Error rates
- Database query performance
- WebSocket connection count
- GitHub API rate limit usage

### Alerting (Phase 2+)
- Error rate spikes
- Response time degradation
- Service health check failures
- Database connection pool exhaustion

---

## DevSmith Coding Standards

**Source:** Based on patterns from DevSmith Logs project

### File Organization

#### Go Service Structure
```
apps/{service}/
├── main.go              # Application entry point
├── handlers/            # HTTP request handlers
│   ├── auth.go         # Authentication handlers
│   ├── api.go          # API endpoints
│   └── health.go       # Health check endpoint
├── models/              # Data structures and database models
│   ├── user.go
│   └── session.go
├── templates/           # Templ template files
│   ├── layout.templ    # Base layout
│   ├── home.templ      # Home page
│   └── components/     # Reusable template components
├── static/              # Static assets (CSS, minimal JS, images)
│   ├── css/
│   ├── js/             # HTMX, Alpine.js, custom JS
│   └── images/
├── services/            # Business logic layer
│   ├── auth_service.go
│   └── user_service.go
├── db/                  # Database package
│   ├── db.go           # Database connection
│   ├── queries.go      # SQL queries
│   └── migrations/     # Migration files
├── middleware/          # HTTP middleware
│   ├── auth.go
│   ├── logging.go
│   └── cors.go
├── utils/               # Helper functions
│   ├── jwt.go
│   └── logger.go
├── config/              # Configuration
│   └── config.go
├── tests/               # Go test files
│   ├── handlers_test.go
│   └── services_test.go
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── Dockerfile           # Multi-stage Docker build
├── .air.toml            # Air hot reload configuration
└── README.md
```

**Key Differences from React/Python:**
- Single service combines frontend and backend (no separate -frontend/-backend)
- Templates directory instead of React components
- Static directory for CSS/minimal JS instead of node_modules
- go.mod instead of package.json/requirements.txt
- Much simpler structure (fewer directories)

---

### Naming Conventions

#### Files
| Type | Convention | Examples |
|------|------------|----------|
| Go Source Files | `snake_case.go` | `auth_handler.go`, `user_service.go`, `jwt_utils.go` |
| Templ Templates | `snake_case.templ` | `home.templ`, `login_form.templ`, `app_nav.templ` |
| Test Files | `_test.go` suffix | `auth_handler_test.go`, `user_service_test.go` |
| SQL Migrations | Timestamped | `20250118120000_create_users_table.sql` |

#### Code (Go Conventions)
| Element | Convention | Examples | Notes |
|---------|------------|----------|-------|
| Packages | `lowercase` | `auth`, `services`, `models` | Single word preferred |
| Variables | `camelCase` | `userData`, `isAuthenticated` | Unexported (private) |
| Exported Variables | `PascalCase` | `UserData`, `APIKey` | Exported (public) |
| Functions | `camelCase` | `handleLogin()`, `validateToken()` | Unexported (private) |
| Exported Functions | `PascalCase` | `HandleLogin()`, `ValidateToken()` | Exported (public) |
| Constants | `PascalCase` or `UPPER_SNAKE` | `MaxRetries`, `API_BASE_URL` | Both acceptable |
| Structs | `PascalCase` | `User`, `Session`, `LoginRequest` | Always exported |
| Interfaces | `PascalCase` | `UserService`, `AuthProvider` | Often end with -er |

**Go-specific Rules:**
- Capitalization determines visibility: `Public` (exported) vs `private` (unexported)
- Package names are always lowercase, single word
- No snake_case for identifiers (Go style is camelCase/PascalCase)
- Acronyms stay uppercase: `HTTPServer`, `JSONData`, `URLPath`

---

### Templ Template Structure

**Standard Template:**
```go
// templates/home.templ
package templates

import "github.com/mikejsmith1985/devsmith-platform/apps/portal/models"

// HomePage renders the home page with user data
templ HomePage(user *models.User, apps []models.App) {
	@Layout("Home") {
		<div class="container mx-auto p-4">
			<h1 class="text-2xl font-bold mb-4">
				Welcome, { user.Name }
			</h1>

			if len(apps) == 0 {
				<p class="text-gray-500">No apps available</p>
			} else {
				<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
					for _, app := range apps {
						@AppCard(app)
					}
				</div>
			}
		</div>
	}
}

// AppCard renders a single app card (reusable component)
templ AppCard(app models.App) {
	<div class="card bg-base-100 shadow-xl">
		<div class="card-body">
			<h2 class="card-title">{ app.Name }</h2>
			<p>{ app.Description }</p>
			<div class="card-actions justify-end">
				<button
					class="btn btn-primary"
					hx-get={ "/apps/" + app.ID + "/launch" }
					hx-target="#app-container"
					hx-swap="innerHTML"
				>
					Launch
				</button>
			</div>
		</div>
	</div>
}
```

**Key Features:**
- Type-safe: Parameters are typed Go structs
- Composable: Templates call other templates with `@TemplateName(args)`
- Compile-time checking: Errors caught during `go build`, not runtime
- HTMX integration: `hx-*` attributes for interactivity
- TailwindCSS + DaisyUI: Utility-first styling

**Templ vs React:**
| Feature | React | Templ |
|---------|-------|-------|
| Language | JSX (JavaScript) | Go |
| Type Safety | TypeScript (optional) | Built-in (Go) |
| Error Detection | Runtime | Compile-time |
| State Management | useState, Context | Server-side (Go variables) |
| Rendering | Client-side | Server-side |
| Bundle Size | Large (100KB+) | None (HTML sent) |

---

### HTMX Patterns

**HTMX handles interactivity without JavaScript frameworks:**

```html
<!-- Load more data -->
<button
  hx-get="/api/logs?page=2"
  hx-target="#log-list"
  hx-swap="beforeend"
>
  Load More
</button>

<!-- Form submission -->
<form
  hx-post="/api/auth/login"
  hx-target="#message"
  hx-swap="innerHTML"
>
  <input type="email" name="email" required />
  <button type="submit">Login</button>
</form>

<!-- WebSocket updates -->
<div
  hx-ext="ws"
  ws-connect="/ws/logs"
>
  <div id="log-stream"></div>
</div>

<!-- Polling for updates -->
<div
  hx-get="/api/status"
  hx-trigger="every 2s"
  hx-target="this"
  hx-swap="innerHTML"
>
  Checking status...
</div>
```

**Common HTMX Attributes:**
- `hx-get/post/put/delete`: HTTP method and URL
- `hx-target`: Where to put the response
- `hx-swap`: How to insert (innerHTML, outerHTML, beforeend, etc.)
- `hx-trigger`: What triggers the request (click, submit, load, every 2s)
- `hx-ext`: Extensions (ws for WebSocket, sse for Server-Sent Events)

---

### Go Handler Pattern

**Standard HTTP Handler:**
```go
// handlers/auth.go
package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-platform/apps/portal/services"
	"github.com/mikejsmith1985/devsmith-platform/apps/portal/templates"
)

// HandleLogin renders the login page
func HandleLogin(c *gin.Context) {
	component := templates.LoginPage()
	component.Render(c.Request.Context(), c.Writer)
}

// HandleLoginSubmit processes login form submission
func HandleLoginSubmit(c *gin.Context) {
	var req LoginRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Call service layer
	user, err := services.AuthenticateUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// Log error with context
		log.Error().
			Err(err).
			Str("email", req.Email).
			Msg("Authentication failed")

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Generate JWT token
	token, err := services.GenerateJWT(user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Authentication error",
		})
		return
	}

	// Return success with token
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"user":    user,
	})
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}
```

**Key Go Patterns:**
- Gin framework for routing and middleware
- Struct binding with validation tags
- Context passed through all layers
- Structured logging (zerolog or zap)
- Explicit error handling (no try-catch, Go uses `if err != nil`)
- JSON responses with `gin.H` (map shorthand)

**Handler Checklist:**
- [ ] Bind and validate input
- [ ] Call service layer (handlers don't contain business logic)
- [ ] Log errors with context
- [ ] Return user-friendly error messages
- [ ] Use appropriate HTTP status codes

---

### Go Service Pattern

**Business Logic Layer:**
```go
// services/auth_service.go
package services

import (
	"context"
	"errors"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

// AuthenticateUser verifies credentials and returns user
func AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
	// Fetch user from database
	user, err := db.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Verify password
	if !verifyPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}

	// Update last login
	user.LastLogin = time.Now()
	if err := db.UpdateUser(ctx, user); err != nil {
		log.Warn().Err(err).Msg("Failed to update last login")
		// Don't fail auth if this fails
	}

	return user, nil
}

// GenerateJWT creates a JWT token for the user
func GenerateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"email":     user.Email,
		"exp":       time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}
```

**Service Layer Best Practices:**
- Define custom error types (better than string errors)
- Use context for cancellation and timeouts
- Keep services pure (no HTTP concerns)
- Return domain models, not DTOs
- Log warnings for non-critical errors

---

### API Response Pattern

**No JavaScript fetch needed - HTMX handles it:**

When using HTMX, handlers return HTML fragments, not JSON (usually):

```go
// Return HTML for HTMX
func HandleGetApps(c *gin.Context) {
	apps, err := services.GetUserApps(c.Request.Context(), getUserID(c))
	if err != nil {
		// Return error HTML fragment
		component := templates.ErrorMessage("Failed to load apps")
		component.Render(c.Request.Context(), c.Writer)
		return
	}

	// Return apps HTML fragment
	component := templates.AppList(apps)
	component.Render(c.Request.Context(), c.Writer)
}
```

**For JSON APIs (when needed):**
```go
func HandleGetAppsJSON(c *gin.Context) {
	apps, err := services.GetUserApps(c.Request.Context(), getUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load apps",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    apps,
	})
}
```

**Key Difference from React/Node:**
- Server renders HTML, not JSON
- HTMX swaps HTML into page
- Less JavaScript, more server-side
- Simpler state management

---

### Error Handling Requirements

**Critical Rules (From Lessons Learned):**

1. **Always provide user-friendly error messages**
   ```javascript
   // ❌ BAD
   return <div>Error: {error}</div>;

   // ✅ GOOD
   return <div>Unable to load data. Please try again.</div>;
   ```

2. **Always include fallback values**
   ```javascript
   // ❌ BAD - Crashes if fetchData throws
   const data = await fetchData();

   // ✅ GOOD - Returns empty array on error
   const data = await fetchData() || [];
   ```

3. **Always include loading states**
   - Show loading indicator while fetching
   - Prevent multiple simultaneous requests
   - Disable action buttons during loading

4. **Always log errors for debugging**
   ```javascript
   console.error('Failed to fetch data:', err);
   ```

5. **NEVER return error strings as data** (Critical lesson from old platform)
   ```python
   # ❌ BAD - Error string looks like valid data
   try:
       result = process()
       return result
   except Exception as e:
       return f"Error: {e}"  # NO!

   # ✅ GOOD - Raise exception, let handler deal with it
   try:
       result = process()
       return result
   except Exception as e:
       logger.error(f"Process failed: {e}", exc_info=True)
       raise HTTPException(status_code=500, detail="Process failed")
   ```

---

### Testing Requirements

#### Test-Driven Development (TDD)
**REQUIRED:** Write tests BEFORE implementation code

**TDD Process:**
1. Read feature requirements
2. Write tests defining expected behavior
3. Run tests (should fail - Red)
4. Write minimal code to pass tests (Green)
5. Refactor if needed
6. Repeat

#### Manual Testing Checklist
Complete BEFORE creating PR:
- [ ] Feature works as expected in browser
- [ ] No console errors or warnings
- [ ] All related features still work (regression check)
- [ ] Works in both light and dark mode (if applicable)
- [ ] Responsive design works on mobile/tablet (if applicable)
- [ ] Works through nginx gateway (http://localhost:3000)
- [ ] Authentication persists across apps
- [ ] WebSocket connections work (for real-time features)
- [ ] Hot module reload (HMR) works during development

#### Automated Testing
**Coverage Requirements:**
- Unit tests: 70% minimum
- Critical paths: 90% minimum

**Test Types:**
- Unit tests for utilities and helper functions
- Component tests for React components
- API endpoint tests for backend routes
- Integration tests for critical user workflows

**Commands:**
```bash
# Frontend tests
cd apps/platform-frontend && npm test
cd apps/platform-frontend && npm run test:coverage

# Backend tests
cd apps/platform-backend && pytest
cd apps/platform-backend && pytest --cov=. --cov-report=term-missing

# Integration tests
./tests/integration-tests.sh
```

#### Gateway/Proxy Testing
When working with multiple services:
- [ ] Test through nginx gateway (http://localhost:3000)
- [ ] Verify direct access works (if supported)
- [ ] Check authentication persists across apps
- [ ] Verify WebSocket connections through gateway
- [ ] Confirm HMR/hot reload still works

---

### Git Workflow

#### Branch Strategy
- **main:** Production releases only
- **development:** Integration branch
- **feature/*:** Feature development (from GitHub issues)
- **fix/*:** Bug fixes
- **break-fix/*:** Experimental debugging (not merged)

#### Commit Message Format
**Convention:** Conventional Commits
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code formatting (no functional changes)
- `refactor:` Code restructuring (no behavior changes)
- `test:` Adding or updating tests
- `chore:` Maintenance tasks

**Examples:**
```
feat(auth): add GitHub OAuth login flow
fix(logs): resolve WebSocket connection timeout
docs(readme): update installation instructions
test(review): add unit tests for code analysis service
```

---

### Configuration Management

#### NO Hardcoded Values

**❌ BAD:**
```javascript
const API_URL = 'http://localhost:8001';
const ws = new WebSocket('ws://localhost:8003/ws/logs');
```

**✅ GOOD:**
```javascript
const API_URL = import.meta.env.VITE_API_URL;
const WS_URL = import.meta.env.VITE_WS_URL;
const ws = new WebSocket(`${WS_URL}/ws/logs`);
```

#### Requirements:
- All URLs from environment variables
- All port numbers from environment variables
- All API keys from environment variables
- All database credentials from environment variables
- .env.example updated with all new variables
- Comments explain purpose of each variable

#### .env.example Format:
```bash
# API Configuration
VITE_API_URL=http://localhost:3000/api/platform
VITE_WS_URL=ws://localhost:3000

# Authentication
GITHUB_CLIENT_ID=your_client_id_here
GITHUB_CLIENT_SECRET=your_client_secret_here
JWT_SECRET=your_secret_key_here

# Database
DATABASE_URL=postgresql://user:password@localhost:5432/dbname
```

---

## Development Workflow

### Hybrid AI Development Team

This project uses a **hybrid AI development approach** with specialized agents:

1. **OpenHands + Ollama** (70-80% of work): Autonomous implementation
2. **Claude via API** (10-15% of work): Architecture and strategic review
3. **GitHub Copilot** (5-10% of work): IDE assistance for manual coding
4. **Mike** (Always): Project orchestration and final approval

**See:** `DevSmithRoles.md` for detailed roles and workflow.

### Branch Strategy
- **main:** Production releases only
- **development:** Integration branch
- **feature/*:** Feature development (created by OpenHands or manually)
- **fix/*:** Bug fixes
- **break-fix/*:** Experimental debugging (not merged)
- **claude-recovery-YYYYMMDD:** Auto-recovery branches (7-day retention)

### Commit Standards
**Format:** Conventional Commits
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `style:` Formatting, no code change
- `refactor:` Code restructuring
- `test:` Adding tests
- `chore:` Maintenance tasks

**Example:**
```
feat(review): add scanning reading mode

Implements targeted search for specific code elements
as defined in Requirements.md section 4.3.

Closes #42
```

### Feature Development Workflow

**See `DevSmithRoles.md` for complete workflow documentation.**

**Summary:**

1. **Issue Creation** (Mike)
   - Create GitHub issue with acceptance criteria
   - Label appropriately

2. **Architecture & Spec** (Claude)
   - Short session (<30 min to avoid crashes)
   - Design architecture
   - Create detailed implementation spec
   - Save spec to issue or `.docs/specs/`

3. **Autonomous Implementation** (OpenHands + Ollama)
   - Mike triggers: `openhands --task "Implement issue #42"`
   - OpenHands works fully autonomously:
     - Creates feature branch from `development`
     - Writes tests first (TDD)
     - Implements feature per spec
     - Runs tests, fixes failures
     - Browser testing via Playwright
     - Commits with Conventional Commits
     - Creates PR with acceptance criteria checklist
   - Duration: 30 min - 2 hours (unattended)
   - **Crash-proof**: Checkpoint/resume capability

4. **Strategic Review** (Claude)
   - Short session (<30 min)
   - Reviews PR for architectural integrity
   - Verifies coding standards
   - Comments with feedback

5. **Acceptance Review** (Mike)
   - Verifies acceptance criteria 100% met
   - Reviews Claude's feedback
   - Approves or requests changes

6. **Merge** (Mike)
   - Squash merge to `development`
   - Delete feature branch
   - Issue auto-closed

7. **Release** (Mike)
   - Merge `development` to `main` when ready
   - Tag version

**Key Advantages:**
- ✅ 80% of implementation work runs unattended
- ✅ Claude crash risk reduced (only 10-15% of work time)
- ✅ OpenHands work is crash-proof (persistent state)
- ✅ No API costs for implementation (Ollama local)
- ✅ Can run overnight on complex features

**Parallel Development:**
- OpenHands can work on one feature while Claude reviews another
- Mike can trigger multiple OpenHands instances on separate features
- No conflicts as long as features are isolated

**Acceptance Criteria Gate:**
- PRs cannot be merged unless acceptance criteria 100% met
- Non-negotiable
- Partial implementations not accepted

### Testing Requirements
**Minimum Coverage:**
- Unit tests: 70%+
- Critical paths: 90%+

**Test Types:**
- Unit: Individual functions/methods
- Component: React components in isolation
- API: Backend endpoint tests
- Integration: Cross-service workflows
- E2E: Full user workflows (Cypress)

**Commands:**
```bash
# Frontend
cd apps/portal-frontend && npm test
cd apps/portal-frontend && npm run test:coverage

# Backend
cd apps/portal-backend && pytest
cd apps/portal-backend && pytest --cov=. --cov-report=term-missing

# Integration
./tests/integration-tests.sh
```

### Code Review Checklist
- [ ] Tests written and passing
- [ ] Follows DevSmith Coding Standards
- [ ] No hardcoded values (config in .env)
- [ ] Error handling implemented
- [ ] Logging added for key operations
- [ ] Documentation updated
- [ ] No "assumption" language in comments
- [ ] Single responsibility (one feature only)

---

## CI/CD & Automation

### GitHub Actions Workflows

We use GitHub Actions to automate quality checks, testing, and validation on every Pull Request.

#### Automated Checks on Every PR

**Workflow:** `.github/workflows/pr-checks.yml`

**What Gets Checked:**

1. **PR Format Validation**
   - ✅ PR title follows Conventional Commits format
   - ✅ PR links to an issue with `Closes #XX`
   - ✅ Branch name follows `feature/{issue-number}-description` format
   - ✅ Acceptance criteria section present in PR description
   - ✅ All acceptance criteria checkboxes are checked
   - ✅ AI_CHANGELOG.md was updated

2. **Automated Testing**
   - ✅ All frontend tests pass (npm test)
   - ✅ All backend tests pass (pytest)
   - ✅ Unit test coverage >= 70%
   - ✅ Critical path coverage >= 90%
   - ✅ Linting passes

3. **Code Quality**
   - ✅ No hardcoded secrets (Trufflehog scan)
   - ✅ No hardcoded URLs or localhost ports
   - ✅ PR size check (warns >1000 lines, fails >2000 lines)

4. **Docker Build**
   - ✅ All Docker images build successfully
   - ✅ Multi-stage builds work correctly

5. **Security Scan**
   - ✅ Trivy vulnerability scanner
   - ✅ Dependency security check
   - ✅ SARIF upload to GitHub Security

**All checks must pass before PR can be approved.**

---

#### Auto-Labeling

**Workflow:** `.github/workflows/auto-label.yml`

PRs are automatically labeled based on:
- **App:** `app:portal`, `app:review`, `app:logs`, `app:analytics`, `app:build`
- **Tech Stack:** `tech:frontend`, `tech:backend`, `tech:docker`, `tech:database`
- **Type:** `type:tests`, `type:docs`, `type:config`, `type:dependencies`
- **Infrastructure:** `infra:gateway`, `infra:ci-cd`
- **Size:** `size/XS`, `size/S`, `size/M`, `size/L`, `size/XL`

This helps with:
- Quick identification of what changed
- Filtering PRs by area
- Detecting scope creep (large PRs)

---

### Issue Templates

**Location:** `.github/ISSUE_TEMPLATE/`

#### Feature Request Template (`feature.yml`)
Structured form for requesting new features:
- Feature name and description
- Which app it belongs to
- User story
- Requirements and acceptance criteria (draft)
- Priority level
- Pre-submission checklist

**Usage:** Claude uses this template as starting point, then refines acceptance criteria and creates detailed implementation specs.

#### Bug Report Template (`bug.yml`)
Structured form for reporting bugs:
- Bug summary and detailed description
- Steps to reproduce
- Expected vs actual behavior
- Error logs and environment info
- Severity level

---

### Pull Request Template

**Location:** `.github/PULL_REQUEST_TEMPLATE/pull_request_template.md`

Comprehensive PR template including:
- Issue reference (Closes #XX)
- Implementation details
- Automated testing results
- Manual testing checklist
- Standards compliance checklist
- **Acceptance Criteria from issue** (all must be checked)
- Changelog update confirmation
- Screenshots (if UI changes)
- Reviewer checklist for Claude

**Key Feature:** Acceptance criteria copied from issue must ALL be checked before PR can be approved. GitHub Actions validates this automatically.

---

### Acceptance Criteria Validation

**How it works:**

1. **Claude creates issue** with acceptance criteria:
   ```markdown
   ## Acceptance Criteria
   - [ ] User can login with GitHub OAuth
   - [ ] JWT token stored in localStorage
   - [ ] Login redirects to dashboard
   ```

2. **Copilot implements** and copies criteria to PR description

3. **Copilot checks off** each criterion as completed:
   ```markdown
   ## Acceptance Criteria
   - [x] User can login with GitHub OAuth
   - [x] JWT token stored in localStorage
   - [x] Login redirects to dashboard
   ```

4. **GitHub Actions validates:**
   - Checks for "Acceptance Criteria" heading
   - Counts checkboxes
   - **Fails if ANY unchecked boxes** found
   - PR cannot be approved until all checked

5. **Claude reviews:**
   - Verifies criteria actually met (not just checked)
   - Recommends approve or request changes

6. **Mike approves:**
   - Final verification of acceptance criteria
   - Only merges if 100% complete

**This creates an automated gate preventing incomplete features from being merged.**

---

### What Can't Be Automated

While GitHub Actions handles many checks, some require human judgment:

**Claude's Role (Cannot Automate):**
- Architectural design decisions
- Complex code review (maintainability, elegance)
- Business logic validation
- Acceptance criteria creation
- Root cause diagnosis of complex bugs

**Mike's Role (Cannot Automate):**
- Final approval based on business priorities
- Release timing decisions
- Scope changes and requirement clarifications

**Copilot's Role (Partially Automated):**
- Writing quality code (automated checks help, but quality varies)
- Test design (coverage is automated, but test quality isn't)
- Problem-solving approach

---

### Status Checks Configuration

**Required status checks before merge:**
```yaml
branches:
  development:
    protection:
      required_status_checks:
        strict: true
        contexts:
          - "Validate PR Format"
          - "Frontend Tests"
          - "Backend Tests"
          - "Code Quality"
          - "Docker Build Check"
          - "Security Scan"
          - "All Checks Passed"
      required_pull_request_reviews:
        required_approving_reviews: 1
        dismiss_stale_reviews: true
      enforce_admins: false
```

**Mike (admin) can override in emergencies, but standard process requires all checks pass.**

---

### CI/CD Best Practices

1. **Fast Feedback**
   - Tests run in parallel (frontend + backend simultaneously)
   - Docker builds use layer caching
   - Most PRs get results in < 5 minutes

2. **Clear Failures**
   - Each check explains what failed and why
   - Links to relevant ARCHITECTURE.md sections
   - Suggests how to fix

3. **Security First**
   - Secrets scanning on every PR
   - Dependency vulnerability checks
   - SARIF results uploaded to GitHub Security tab

4. **Scope Control**
   - PR size warnings prevent scope creep
   - Auto-labeling makes large PRs visible
   - One feature per PR enforced by review

5. **Quality Gates**
   - 70% unit test coverage required
   - 90% critical path coverage required
   - All acceptance criteria must be checked
   - No hardcoded values allowed

---

### Future Automation Enhancements (Phase 2+)

**Potential additions:**
- **Automated deployment** to staging on merge to development
- **E2E tests** with Playwright/Cypress in CI
- **Performance benchmarks** comparing PR to main
- **Visual regression tests** for UI changes
- **Dependency update PRs** via Dependabot
- **Automatic changelog generation** from commit messages
- **Code coverage trends** tracking over time
- **AI-powered code review** suggestions (GitHub Copilot for PRs)

---

## Implementation Phases

### Phase 1: Foundation (Current - Not Started)
**Goal:** Portal + Logging infrastructure

**Deliverables:**
- [ ] Nginx gateway configuration
- [ ] Docker Compose setup
- [ ] PostgreSQL with schema initialization
- [ ] Redis setup
- [ ] Portal frontend (auth, navigation)
- [ ] Portal backend (GitHub OAuth, user management)
- [ ] Logging frontend (real-time log dashboard)
- [ ] Logging backend (ingestion, storage, WebSocket)
- [ ] Centralized logging SDK for all services
- [ ] .env.example with all configuration
- [ ] README with one-click installation

**Success Criteria:**
- User can login with GitHub
- Portal shows app browser
- Logs dashboard displays real-time logs from portal backend
- All services accessible via gateway

**Timeline:** TBD

---

### Phase 2: Core Apps
**Goal:** Review + Analytics apps

**Deliverables:**
- [ ] Review frontend (code input, reading modes UI)
- [ ] Review backend (Claude/Ollama integration, code analysis)
- [ ] Analytics frontend (dashboards, charts)
- [ ] Analytics backend (data analysis, exports)
- [ ] GitHub integration for code import
- [ ] Five reading modes implemented
- [ ] Real-time collaboration (WebSocket)

**Success Criteria:**
- User can paste code and get AI analysis
- All five reading modes functional
- Analytics shows log trends and anomalies
- GitHub repo browsing works

**Timeline:** TBD

---

### Phase 3: Build App (Phase 1)
**Goal:** Terminal interface

**Deliverables:**
- [ ] Build frontend (terminal UI)
- [ ] Build backend (terminal emulation)
- [ ] Cloud CLI support
- [ ] Copilot CLI integration
- [ ] Shared terminal sessions

**Success Criteria:**
- User can run CLI commands
- Terminal output logged in real-time
- Multiple users can co-pilot

**Timeline:** TBD

---

### Phase 4: Build App (Phase 2)
**Goal:** Autonomous coding

**Deliverables:**
- [ ] OpenHands integration
- [ ] Autonomous task execution
- [ ] Code generation and verification
- [ ] Integration with review app for validation

**Success Criteria:**
- User can request code generation
- OpenHands completes tasks autonomously
- Generated code reviewed automatically

**Timeline:** TBD

---

## Decision Log

### Template
```markdown
### Decision: [Title]
**Date:** YYYY-MM-DD
**Status:** [Proposed | Accepted | Rejected | Superseded]
**Context:** Why this decision was needed
**Decision:** What was decided
**Alternatives Considered:** Other options and why rejected
**Consequences:** Impact of this decision
**References:** Related issues, PRs, docs
```

---

### Decision: Gateway-First Architecture
**Date:** 2025-10-18
**Status:** Accepted
**Context:** Previous platform added nginx gateway as afterthought, causing authentication and routing failures. Multiple commits failed to fix issues.

**Decision:** Design and implement nginx gateway before any app development. All services will use gateway paths from day one, with no direct port access in code.

**Alternatives Considered:**
1. Direct port access with CORS - Rejected: Separate localStorage breaks auth
2. Add gateway later - Rejected: Proven to fail in previous platform
3. Service mesh (Istio) - Rejected: Overkill for project scale

**Consequences:**
- ✅ Shared authentication works across apps
- ✅ Production-ready architecture from start
- ✅ Clean URL structure
- ⚠️ Requires upfront gateway configuration
- ⚠️ Dev workflow slightly more complex (can't test services in isolation without gateway)

**References:** LESSONS_LEARNED.md Section 1.1

---

### Decision: JWT Token Field Name Standard
**Date:** 2025-10-18
**Status:** Accepted
**Context:** Previous platform had mismatch between `github_token` and `github_access_token` field names, causing authentication failures across services.

**Decision:** All services MUST use `github_access_token` as the field name in JWT payload. This is documented in architecture and enforced in code reviews.

**Alternatives Considered:**
1. `github_token` - Rejected: Less specific, unclear what type of token
2. `access_token` - Rejected: Too generic, unclear source

**Consequences:**
- ✅ No field name mismatches
- ✅ Clear and explicit naming
- ✅ Easy to grep across codebase
- ⚠️ Must be vigilant in code reviews

**References:** LESSONS_LEARNED.md Section 1.2

---

### Decision: Anthropic SDK Version >= 0.40.0
**Date:** 2025-10-18
**Status:** Accepted
**Context:** Previous platform used anthropic 0.7.7 (2023) but code expected `.messages` API from 0.40+ (2024), causing AttributeError.

**Decision:** Require anthropic SDK >= 0.40.0 in requirements.txt. Use `>=` to allow updates while ensuring minimum compatible version.

**Alternatives Considered:**
1. Pin exact version (e.g., 0.71.0) - Rejected: Prevents security updates
2. No version constraint - Rejected: Risks future breaking changes
3. Use version range (e.g., >=0.40.0,<1.0.0) - Considered for future

**Consequences:**
- ✅ `.messages.create()` API available
- ✅ Security updates allowed
- ⚠️ Must monitor for breaking changes in minor versions

**References:** LESSONS_LEARNED.md Section 3.1

---

### Decision: Single Schema Per Service
**Date:** 2025-10-18
**Status:** Accepted
**Context:** Need true modularity where apps can function independently, but also need ability to query across apps for analytics.

**Decision:** Each service owns a PostgreSQL schema. No foreign keys across schemas. Cross-schema queries via application logic or database views.

**Alternatives Considered:**
1. Shared tables - Rejected: Breaks modularity, tight coupling
2. Separate databases - Rejected: More complex, harder to query across
3. Microservices with separate DBs + API federation - Rejected: Overkill for project scale

**Consequences:**
- ✅ True modularity - apps independent
- ✅ Federated queries still possible
- ✅ Clear ownership of data
- ⚠️ No referential integrity across schemas
- ⚠️ Must manage consistency in application code

**References:** Requirements.md Section "Database"

---

### Decision: React Context API Over Redux
**Date:** 2025-10-18
**Status:** Accepted
**Context:** Need state management for authentication and theme across React apps.

**Decision:** Use React Context API for global state. No Redux or external state library.

**Alternatives Considered:**
1. Redux - Rejected: Too much boilerplate for app complexity
2. Zustand - Rejected: Adds dependency, Context API sufficient
3. Recoil - Rejected: Adds dependency, Context API sufficient

**Consequences:**
- ✅ No external dependencies
- ✅ Simpler codebase
- ✅ Built-in React feature
- ⚠️ May need Redux if state complexity grows significantly (Phase 3+)

**References:** DevSmith Coding Standards

---

### Decision: WebSocket Over Server-Sent Events
**Date:** 2025-10-18
**Status:** Accepted
**Context:** Need real-time bidirectional communication for logs, terminal, collaboration features.

**Decision:** Use WebSockets for all real-time features. Native WebSocket API on frontend, FastAPI WebSocket on backend.

**Alternatives Considered:**
1. Server-Sent Events (SSE) - Rejected: Unidirectional, doesn't fit terminal use case
2. Long polling - Rejected: Inefficient, poor latency
3. Socket.IO - Rejected: Adds dependency, native WebSocket sufficient

**Consequences:**
- ✅ Bidirectional communication
- ✅ Low latency
- ✅ Native browser support
- ⚠️ Must handle reconnection logic
- ⚠️ Redis pub/sub needed for multi-instance (Phase 3)

**References:** Requirements.md Section "Logging App"

---

### Decision: Build Order - Portal → Logging → Analytics → Review → Build
**Date:** 2025-10-18
**Status:** Accepted
**Context:** Previous platform built all apps in parallel, causing integration chaos. Need stable foundation before complex features.

**Decision:** Build apps sequentially in order: Portal (navigation) → Logging (monitors development) → Analytics (analyzes logs) → Review (benefits from monitoring) → Build (most complex).

**Alternatives Considered:**
1. Build in parallel - Rejected: Proven to fail in previous platform
2. Build Review first - Rejected: No monitoring for debugging
3. Build Build app first - Rejected: Too complex without stable foundation

**Consequences:**
- ✅ Each app builds on stable foundation
- ✅ Logging available for debugging subsequent apps
- ✅ Incremental complexity
- ⚠️ Longer time to complete all features
- ⚠️ Must resist temptation to work on multiple apps at once

**References:** Requirements.md Section "Build Order", DevSmithRoles.md

---

### Decision: TDD with 70% Unit / 90% Critical Path Coverage
**Date:** 2025-10-18
**Status:** Accepted
**Context:** Previous platform had minimal tests, causing regressions. Need quality standards.

**Decision:** Write tests before implementation. Require 70% unit test coverage and 90% critical path coverage. Tests run in CI before merge.

**Alternatives Considered:**
1. No coverage requirements - Rejected: No accountability
2. 100% coverage - Rejected: Diminishing returns, slows development
3. Different thresholds - Rejected: 70/90 is industry standard

**Consequences:**
- ✅ Catches regressions before merge
- ✅ Encourages testable design
- ✅ Living documentation
- ⚠️ Upfront time investment
- ⚠️ Coverage metrics can be gamed (quality matters too)

**References:** DevsmithTDD.md, DevSmithRoles.md

---

### Decision: Go + Templ + HTMX Over React + Node.js
**Date:** 2025-10-18
**Status:** Accepted
**Context:** Previous platform suffered from V8 JavaScript engine crashes during Docker builds, requiring workarounds (`NODE_OPTIONS="--jitless"`, `DOCKER_BUILDKIT=0`). Build times were slow (5+ minutes), containers were large (500MB+), and memory usage was high (500MB+ per service).

**Decision:** Use Go 1.21+ with Templ templates and HTMX for interactivity instead of React with Node.js/Vite build tooling.

**Technology Stack:**
- **Language:** Go 1.21+
- **Framework:** Gin (HTTP) + Templ (templates) + HTMX (interactivity)
- **Styling:** TailwindCSS + DaisyUI
- **Database:** PostgreSQL 15+ with pgx driver
- **Containerization:** Docker with multi-stage builds (golang:1.21-alpine → alpine:latest)

**Alternatives Considered:**
1. **Keep React + Node.js** - Rejected: V8 crashes unresolved, slow builds, large containers
2. **Python FastAPI + Jinja2 + Alpine.js** - Considered: Zero learning curve, but slower than Go
3. **Rust + Axum + Leptos** - Rejected: Too steep learning curve, slower development

**Consequences:**

✅ **Eliminates V8 Crashes:**
- No Node.js = No V8 engine = No crashes
- No workarounds needed (`--jitless`, `DOCKER_BUILDKIT=0`)
- Stable builds every time

✅ **Performance Benefits:**
- 10-50x faster API responses than Node.js
- Memory usage: 50-100MB (vs 500MB+ Node)
- Docker images: 10-20MB (vs 500MB+ Node)
- Build time: ~30 seconds (vs 5+ minutes)

✅ **Developer Experience:**
- Hot reload with Air tool (same as HMR)
- Single binary deployment (no `npm install`)
- Compile-time error checking (Templ templates are type-safe)
- Simpler tooling (`go build` vs Webpack/Vite/npm)

✅ **Infrastructure:**
- Smaller containers = lower hosting costs
- Faster deployments
- Built-in concurrency (goroutines) for WebSockets
- No node_modules directory

⚠️ **Trade-offs:**
- Learning curve for Go (moderate, ~1-2 weeks)
- HTMX is different paradigm than React (server-side rendering)
- Smaller ecosystem than React (but growing fast)
- Copilot less familiar with Go+Templ+HTMX combo (but still capable)

⚠️ **Risks Mitigated:**
- GitHub Copilot knows Go well (top 5 language for Copilot)
- HTMX is simpler than React state management
- Templ documentation is excellent
- Plenty of Go+HTMX examples available

**Implementation Impact:**
- Coding standards updated (Section 13) for Go conventions
- File organization simplified (single service, not frontend/backend split)
- CI/CD workflows updated (Go testing instead of Jest/Pytest mix)
- Docker builds updated (multi-stage Go builds)
- No changes to PostgreSQL, Redis, Nginx, or CI/CD tools

**References:**
- LESSONS_LEARNED.md Section 3.4 (Docker and Build System Fragility)
- Old platform CLAUDE_CHANGELOG.md: V8 crash workarounds
- Technology Stack (Section 4)

### Decision: Hybrid AI Development Team (OpenHands + Ollama + Claude + Copilot)
**Date:** 2025-10-18
**Status:** Accepted
**Context:** Claude Code (running on Node.js/V8) is prone to crashes, causing work loss. Need a development approach that:
1. Minimizes Claude crash risk
2. Automates majority of implementation work
3. Maintains high code quality
4. Operates within system resources (Dell G16 7630)

**Decision:** Implement hybrid AI team with specialized roles:
- **OpenHands + Ollama:** Primary implementation agent (70-80% of work) - fully autonomous, crash-proof
- **Claude (via API):** Architecture and strategic review (10-15% of work) - short sessions to minimize crash risk
- **GitHub Copilot:** IDE assistance for manual coding (5-10% of work)
- **Mike:** Project orchestration and final approval

**System Specs Verified:**
- Dell G16 7630: i9-13900HX (24 cores/32 threads), 32GB RAM, RTX 4070 8GB
- Assessment: Excellent for running multiple large Ollama models (16B-70B)

**Alternatives Considered:**
1. **Continue with Claude + Copilot only** - Rejected: Claude crash risk too high, Copilot not autonomous
2. **Use only OpenHands** - Rejected: Lacks Claude's architectural reasoning and 200K context
3. **Use cloud-based agents (Cursor, Replit)** - Rejected: Privacy concerns, API costs, vendor lock-in

**Consequences:**

✅ **Benefits:**
- 80% of work is crash-proof (OpenHands runs independently with checkpoint/resume)
- Claude crash risk reduced to 10-15% of work time (short architecture/review sessions)
- No API costs for implementation (Ollama runs locally)
- No rate limits (can run 24/7)
- OpenHands can work overnight on complex features
- Parallel development (OpenHands implements while Claude reviews)
- Privacy preserved (code stays local for implementation)

⚠️ **Trade-offs:**
- Learning curve for OpenHands configuration
- Ollama model management overhead
- Need to write detailed specs for OpenHands (more upfront planning)
- OpenHands quality depends on model size (16B+ recommended)

⚠️ **Risks Mitigated:**
- Claude crash recovery hooks in `.claude/hooks/` (session logging, git auto-recovery)
- Todo list persistence for tracking progress
- OpenHands checkpoint/resume handles system crashes
- Detailed specs ensure OpenHands understands requirements

**Implementation Impact:**
- DevSmithRoles.md updated with hybrid workflow
- ARCHITECTURE.md Section 5 (Development Tools) updated
- ARCHITECTURE.md Section 14 (Development Workflow) updated
- Crash recovery hooks implemented in `.claude/hooks/`
- No changes to tech stack (Go + Templ + HTMX)

**References:**
- DevSmithRoles.md - Complete workflow documentation
- `.claude/hooks/README.md` - Crash recovery documentation
- System specs screenshots (2025-10-18)

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-18 | Claude | Initial architecture document created |
| 1.1 | 2025-10-18 | Claude | Added CI/CD & Automation section (Section 15) |
| 1.2 | 2025-10-18 | Claude | Changed tech stack from React+Node to Go+Templ+HTMX to eliminate V8 crashes |
| 1.3 | 2025-10-18 | Claude | Added hybrid AI development approach (OpenHands + Ollama + Claude + Copilot) |

---

## References

- [Requirements.md](./Requirements.md) - Feature requirements and specifications
- [DevSmithRoles.md](./DevSmithRoles.md) - Team roles and responsibilities
- [DevsmithTDD.md](./DevsmithTDD.md) - Test-driven development approach
- [LESSONS_LEARNED.md](./LESSONS_LEARNED.md) - Lessons from previous platform (internal only)

---

**Next Steps:**
1. Review and approve this architecture document
2. Create initial repository structure (apps/, packages/, etc.)
3. Implement nginx gateway configuration
4. Create docker-compose.yml
5. Begin Phase 1 implementation (Portal + Logging)
