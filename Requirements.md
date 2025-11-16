# DevSmith Modular Platform: Requirements

## Repository
- **Hosted at**: [github.com/mikejsmith1985/devsmith-modular-platform](https://github.com/mikejsmith1985/devsmith-modular-platform)
- **Architecture**: Modular platform with isolated, interoperable services
- **Development Approach**: Hybrid AI team (OpenHands + Ollama + Claude + Copilot)

---

## Executive Summary

**Core Mission**: Build a platform that teaches developers how to effectively read and understand code - the critical skill for supervising AI-generated code in the "Human in the Loop" era.

**Key Insight**: As AI generates more code (10x+ increase since 2024), the primary developer responsibility shifts from *writing* code to *reading, understanding, and validating* AI output. This platform trains users in effective code reading through five distinct reading modes, each optimized for managing cognitive load.

**Central Philosophy**: The Review app is the **centerpiece** - all other apps exist to support the code reading, evaluation, and iteration workflow.

---

## Platform Principles

### 1. Cognitive Load Management
Everything is designed to optimize mental effort:
- **Minimize Intrinsic Load**: Simplify inherent complexity
- **Reduce Extraneous Load**: Eliminate wasted mental effort
- **Maximize Germane Load**: Build transferable mental frameworks

### 2. Mental Models as Foundation
Four core models underpin all architecture:
- **Bounded Contexts**: Same entity, different meanings in different domains
- **Layered Architecture**: Controller → Orchestration → Data separation
- **Abstraction vs Implementation**: Understand "what" before "how"
- **Scope & Context**: Minimize variable lifespans and visibility

### 3. Human in the Loop (HITL)
Platform prepares users for the new developer responsibility:
- **Old Model**: Write code
- **New Model**: Supervise and validate AI-generated code
- **Required Skill**: Effective code reading (not code writing)

### 4. Five Reading Modes
All code interaction happens through one of five modes:
1. **Preview**: Quick structural assessment
2. **Skim**: Understand abstractions and flow
3. **Scan**: Find specific information
4. **Detailed**: Deep algorithm comprehension
5. **Critical**: Quality evaluation and improvement

---

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin (HTTP routing)
- **Rationale**:
  - No V8/Node.js crashes
  - Explicit error handling (no hidden exceptions)
  - Fast compilation and execution
  - Strong concurrency primitives

### Frontend
- **Templates**: Templ (type-safe, compile-time checked)
- **Interactivity**: HTMX (no heavy JavaScript frameworks)
- **Styling**: TailwindCSS + DaisyUI
- **Rationale**:
  - Server-side rendering (simple deployment)
  - Minimal client-side JavaScript (reduce complexity)
  - Progressive enhancement (works without JS)

### Database
- **Primary**: PostgreSQL 15+
- **Driver**: pgx (native Go, high performance)
- **Architecture**: Schema isolation per service
  - `portal.*` - Authentication, app management
  - `reviews.*` - Code review sessions
  - `logs.*` - Log storage
  - `analytics.*` - Aggregated statistics
  - `builds.*` - Build session data (Phase 2)

**No cross-schema foreign keys** - services communicate via APIs, not direct DB coupling.

### Infrastructure
- **Gateway**: Nginx (reverse proxy, single entry point)
- **Containerization**: Docker + Docker Compose
- **CI/CD**: GitHub Actions
- **Deployment**: Single-command setup via Docker Compose

### AI Integration
- **Primary**: Ollama (local, private, no API costs)
- **Model Selection**: Configurable based on available RAM (see System Requirements)
- **Fallback**: Claude API for complex architectural tasks
- **Configuration**: Environment variables, toggled via UI

---

## System Requirements

### AI Model Selection

The platform supports multiple DeepSeek-Coder models with different resource requirements:

| Model | RAM Required | Performance | Best For | Download Size |
|-------|--------------|-------------|----------|---------------|
| `deepseek-coder:1.5b` | 8GB | Fastest | Low-end systems, quick responses | ~1GB |
| `deepseek-coder:6.7b` | 16GB | **Recommended** | Most users, good balance | ~4GB |
| `deepseek-coder-v2:16b` | 32GB | Best quality | Complex analysis, Critical Mode | ~9GB |
| `qwen2.5-coder:7b` | 16GB | Alternative | Similar to 6.7b | ~4GB |

**Default**: `deepseek-coder:6.7b` (setup script auto-detects RAM and suggests best model)

**Model Capabilities by Reading Mode**:
- **Preview Mode** (2-3 min): All models adequate (structure analysis)
- **Skim Mode** (5-7 min): 6.7b+ recommended (pattern recognition)
- **Scan Mode** (3-5 min): All models adequate (targeted search)
- **Detailed Mode** (10-15 min): 6.7b+ recommended, 16b better (deep analysis)
- **Critical Mode** (15-20 min): 16b preferred (quality analysis, but 6.7b adequate)

### Minimum Configuration
- **RAM**: 16GB (for `deepseek-coder:6.7b` - recommended default)
- **CPU**: 8 cores (Intel/AMD) or Apple M1+
- **Storage**: 50GB (models + Docker images)
- **OS**: Linux, macOS, Windows (via WSL2)

**Low-end Systems (8GB RAM)**:
- Use `deepseek-coder:1.5b` model
- Expect slower inference and less accurate analysis
- All reading modes functional but quality reduced

### Recommended Configuration
- **RAM**: 32GB (for `deepseek-coder-v2:16b` - best quality)
- **CPU**: 16+ cores or Apple M1 Pro+
- **GPU**: Optional but beneficial (8GB+ VRAM)
  - NVIDIA RTX 4070+ ideal for 16B+ models
  - Apple Silicon uses unified memory (no separate GPU needed)

### Verified Configurations

**Budget-Friendly (16GB RAM)**:
- Model: `deepseek-coder:6.7b`
- Performance: Good for 90% of use cases
- Inference: ~2-5 seconds per response
- Assessment: ✅ Recommended for most users

**High-Performance (32GB RAM)**:
- **Dell G16 7630**: i9-13900HX (24 cores), 32GB RAM, RTX 4070 8GB
- Model: `deepseek-coder-v2:16b`
- Performance: Excellent - can run 70B quantized models
- Inference: ~5-10 seconds per response
- Assessment: ✅ Best for complex codebases and Critical Mode

---

## Core Applications

### Portal Service
**Purpose**: Authentication gateway and app launcher

**Responsibilities**:
- GitHub OAuth authentication
- Session management (JWT tokens)
- App directory and launcher
- User profile management
- Service health dashboard

**Bounded Context**: Authentication and app orchestration
- `User` = authenticated identity
- `Session` = active login
- `App` = launchable service

**Key Features**:
- One-click GitHub login
- App cards showing status (running/stopped)
- Recent activity feed
- Quick navigation to other services

**Tech Stack**: Go + Gin + Templ + HTMX

---

### Review Service
**Purpose**: The **centerpiece** - teaches effective code reading through AI-assisted analysis

**Core Philosophy**:
This is the primary value proposition of the platform. The Review service implements the five reading modes, teaching users how to supervise AI-generated code effectively.

#### The Five Reading Modes (Detailed Requirements)

**1. Preview Mode**
- **Purpose**: Rapid assessment of code structure
- **Cognitive Strategy**: Minimal intrinsic load, maximum speed
- **AI Provides**:
  - File/folder tree with descriptions
  - Identified bounded contexts (e.g., "Auth domain", "Review domain")
  - Technology stack detection (Go, Python, etc.)
  - Architectural pattern (layered, microservices, monolith)
  - Entry points (main.go, handlers, etc.)
  - External dependencies (APIs, databases)
- **UI/UX**:
  - Tree view with expandable folders
  - Color coding by layer (controller=blue, service=green, data=orange)
  - AI summary panel: "This is a Go microservice using Gin and PostgreSQL..."
  - Filter by file type (.go, .templ, .sql)
- **Use Cases**:
  - Evaluating GitHub repo before cloning
  - Quick assessment of OpenHands output
  - Determining project relevance

**2. Skim Mode**
- **Purpose**: Understand functionality without implementation details
- **Cognitive Strategy**: Focus on abstractions, reduce extraneous load
- **AI Provides**:
  - Function/method signatures with descriptions
  - Interface definitions and purposes
  - Data models (structs, entities)
  - Key workflows with diagrams
  - API endpoint catalog
  - Integration points (external services)
- **UI/UX**:
  - Collapsible function list with AI descriptions
  - Interface viewer (contracts only, no implementations)
  - Auto-generated workflow diagrams (Mermaid.js)
  - Entity relationship diagram for data models
  - Click to expand → transitions to Detailed mode
- **Use Cases**:
  - Understanding what a codebase does overall
  - Preparing spec for OpenHands
  - Architectural review by Claude

**3. Scan Mode**
- **Purpose**: Targeted information search
- **Cognitive Strategy**: Direct path to target, filter noise
- **AI Provides**:
  - Semantic search (not just string matching)
    - "Where is auth validated?" → finds relevant code even without exact words
  - Variable/function usage tracking
  - Error source identification
  - Pattern matching ("Find all SQL queries")
  - Related code discovery ("Show all callers")
  - Context-aware suggestions
- **UI/UX**:
  - Natural language search bar
  - Results with 3 lines before/after context
  - Jump-to-definition
  - Syntax highlighting of matches
  - Related results panel
  - Filters: layer, bounded context, file type
- **Use Cases**:
  - Debugging: "Where does this null pointer come from?"
  - Security audit: "Find all database queries"
  - Refactoring prep: "What calls this deprecated function?"

**4. Detailed Mode**
- **Purpose**: Deep understanding of algorithms
- **Cognitive Strategy**: Accept high intrinsic load, maximize context
- **AI Provides**:
  - Line-by-line explanation
  - Variable state at each point ("Here, `user` is nil because...")
  - Control flow analysis (if/else paths, loops)
  - Algorithm explanation ("This implements Dijkstra's algorithm...")
  - Complexity analysis (time/space if applicable)
  - Edge case identification
  - Bug/issue detection
  - Links to docs (Go docs, Stack Overflow, etc.)
- **UI/UX**:
  - Split view: code left, AI explanation right
  - Synchronized scrolling
  - Click any line for detailed breakdown
  - Variable hover shows type and current state
  - Execution flow visualization
  - Step-through simulation
  - Annotation mode (user notes)
- **Use Cases**:
  - Understanding complex algorithm before modifying
  - Debugging subtle logic error
  - Learning from well-written code

**5. Critical Mode** (Human-in-the-Loop Review)
- **Purpose**: Evaluate quality and identify improvements
- **Cognitive Strategy**: Evaluative thinking, teach patterns/anti-patterns
- **AI Provides**:
  - **Architecture Issues**:
    - Bounded context violations (e.g., Portal User logic in Review service)
    - Layer mixing (controller calling repository directly)
    - Missing abstractions (should be interface, not concrete type)
    - Tight coupling (too many dependencies)
  - **Code Quality**:
    - Go idiom violations (not following conventions)
    - Error handling issues (errors ignored, not wrapped)
    - Scope problems (unnecessary globals)
    - Naming violations
    - Missing documentation
  - **Security**:
    - SQL injection risks (string concatenation in queries)
    - Unvalidated input
    - Secrets in code
    - Auth/authorization gaps
  - **Performance**:
    - N+1 query problems
    - Unnecessary allocations
    - Missing database indexes
    - Inefficient algorithms
  - **Testing**:
    - Untested code paths
    - Missing error case tests
    - Low coverage
  - **Improvement Suggestions**:
    - Specific refactoring with before/after examples
    - Priority (critical/important/minor)
    - Estimated effort
- **UI/UX**:
  - Issue list categorized by type and severity
  - Click issue → jump to code location
  - Issue explanation with context
  - Suggested fix (diff view)
  - Accept/reject/modify buttons
  - Add to refactoring backlog
  - Generate PR comment
  - Track issue history (recurring problems)
- **Use Cases**:
  - **PRIMARY**: Reviewing OpenHands output before merge
  - Claude reviewing PRs
  - Mike's final acceptance check
  - Security audit
  - Refactoring planning

#### Reading Mode Selection & Transitions
- **Intelligent Suggestions**: AI recommends starting mode based on:
  - First time seeing code? → Preview
  - Need overall understanding? → Skim
  - Looking for something specific? → Scan
  - Complex logic to understand? → Detailed
  - Quality review needed? → Critical

- **Fluid Transitions**:
  - Preview → "Go Deeper" button → Skim
  - Skim → Click function → Detailed (for that function)
  - Detailed → "Find Usages" → Scan
  - Any mode → "Review This" → Critical

#### Technical Implementation
**Database Schema** (`reviews.*`):
```sql
-- Review sessions (one per code upload)
CREATE TABLE reviews.sessions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR(255),
    code_source VARCHAR(20) CHECK (code_source IN ('github', 'paste', 'upload')),
    github_repo VARCHAR(255),
    github_branch VARCHAR(100),
    pasted_code TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    last_accessed TIMESTAMP DEFAULT NOW()
);

-- Reading sessions (one per mode analysis)
CREATE TABLE reviews.reading_sessions (
    id SERIAL PRIMARY KEY,
    session_id INT REFERENCES reviews.sessions(id) ON DELETE CASCADE,
    reading_mode VARCHAR(20) CHECK (reading_mode IN ('preview', 'skim', 'scan', 'detailed', 'critical')),
    target_path VARCHAR(500),  -- specific file/function
    scan_query TEXT,           -- for scan mode
    ai_response JSONB,         -- cached AI results
    user_annotations TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Issues found in Critical mode
CREATE TABLE reviews.critical_issues (
    id SERIAL PRIMARY KEY,
    reading_session_id INT REFERENCES reviews.reading_sessions(id) ON DELETE CASCADE,
    issue_type VARCHAR(50),     -- 'architecture', 'security', 'performance', 'quality', 'testing'
    severity VARCHAR(20),       -- 'critical', 'important', 'minor'
    file_path VARCHAR(500),
    line_number INT,
    description TEXT,
    suggested_fix TEXT,
    status VARCHAR(20) DEFAULT 'open',  -- 'open', 'accepted', 'rejected', 'fixed'
    created_at TIMESTAMP DEFAULT NOW()
);
```

**API Endpoints**:
```
POST   /api/review/sessions              - Create review session (GitHub URL or paste)
GET    /api/review/sessions              - List user's sessions
GET    /api/review/sessions/:id          - Get session details
DELETE /api/review/sessions/:id          - Delete session

POST   /api/review/sessions/:id/analyze  - Run AI analysis for a mode
  Request: {
    "reading_mode": "preview|skim|scan|detailed|critical",
    "target_path": "/path/to/file.go",    // optional, required for detailed
    "scan_query": "find authentication",  // required for scan mode
  }
  Response: { "analysis": {...}, "cached": true/false }

GET    /api/review/sessions/:id/results/:mode  - Get cached analysis
POST   /api/review/sessions/:id/annotate       - Save user notes
GET    /api/review/sessions/:id/issues         - Get all critical issues
PATCH  /api/review/issues/:id                  - Update issue status

WS     /ws/review/sessions/:id/collaborate     - Real-time collaboration
```

**AI Integration**:
- Ollama endpoint: `http://localhost:11434`
- Model: Configurable (default: `deepseek-coder:6.7b`, see System Requirements)
- Temperature varies by mode (0.1 for Preview, 0.7 for Critical)
- Responses cached in database (expensive to regenerate)
- Fallback to Claude API if Ollama unavailable

**Integration with Other Services**:
- **Logging**: All AI calls logged (performance metrics)
- **Analytics**: Usage patterns (which modes used most, success rate)
- **Build**: Can auto-trigger Critical review of OpenHands output
- **Portal**: Auth, session management

---

### Logging Service
**Purpose**: Real-time log capture and monitoring

**Bounded Context**: Audit and monitoring
- `LogEntry` = single log line with metadata
- `User` = actor who triggered log (audit trail)

**Responsibilities**:
- Ingest logs from all services via REST API
- Real-time streaming via WebSocket
- Tag-based filtering and search
- Severity-level tracking (debug, info, warn, error)
- Optional AI-driven analysis (patterns, anomalies)

**Database Schema** (`logs.*`):
```sql
CREATE TABLE logs.entries (
    id BIGSERIAL PRIMARY KEY,
    user_id INT,
    service VARCHAR(50),      -- 'portal', 'review', 'logging', etc.
    level VARCHAR(20),        -- 'debug', 'info', 'warn', 'error'
    message TEXT,
    metadata JSONB,           -- Generic metadata for flexibility
    created_at TIMESTAMP DEFAULT NOW()
);

-- NEW: Browser debugging data (network tab + console output)
CREATE TABLE logs.browser_debug_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id INT,
    session_name VARCHAR(255),
    user_action TEXT,         -- "Clicked Review card from dashboard"
    network_log JSONB,        -- Full DevTools Network tab output
    console_log JSONB,        -- Full DevTools Console output
    page_errors JSONB,        -- JavaScript errors
    navigation_events JSONB,  -- URL changes, redirects
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_logs_service_level ON logs.entries(service, level, created_at DESC);
CREATE INDEX idx_logs_user ON logs.entries(user_id, created_at DESC);
CREATE INDEX idx_logs_created ON logs.entries(created_at DESC);
CREATE INDEX idx_browser_debug_user ON logs.browser_debug_sessions(user_id, created_at DESC);
```

**API Endpoints**:
```
POST   /api/logs              - Ingest log entry
GET    /api/logs              - Query logs (with filters)
GET    /api/logs/stats        - Statistics (count by level, service)
WS     /ws/logs               - Real-time log stream

-- NEW: Browser debugging endpoints
POST   /api/logs/browser-debug          - Submit browser debug session
GET    /api/logs/browser-debug/:id      - Retrieve debug session
GET    /api/logs/browser-debug/user/:id - Get all debug sessions for user
```

**Features**:
- WebSocket pub/sub via Redis
- Filters: service, level, date range, keyword
- Export logs (JSON, CSV)
- AI analysis: "Explain this error pattern"

**Integration**:
- All services send logs here
- Analytics reads from logs.* schema
- Review app logs AI call performance

---

### Analytics Service
**Purpose**: Aggregate and visualize log patterns

**Bounded Context**: Statistical analysis
- `Trend` = pattern over time
- `Anomaly` = unusual spike or dip

**Responsibilities**:
- Frequency analysis (most common errors)
- Trend detection (error rate over time)
- Anomaly detection (sudden spikes)
- Performance metrics (avg response time)
- Exportable reports (CSV, JSON, PDF)

**Database Schema** (`analytics.*`):
```sql
CREATE TABLE analytics.aggregations (
    id SERIAL PRIMARY KEY,
    metric_type VARCHAR(50),   -- 'error_frequency', 'response_time', etc.
    service VARCHAR(50),
    time_bucket TIMESTAMP,     -- hourly buckets
    value NUMERIC,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**API Endpoints**:
```
GET    /api/analytics/trends         - Trend analysis
GET    /api/analytics/anomalies      - Detect anomalies
GET    /api/analytics/top-issues     - Most common issues
GET    /api/analytics/export         - Export report
```

**Features**:
- Time-series charts (Chart.js)
- Heatmaps (severity by hour/day)
- Comparative analysis (this week vs last week)
- AI insights: "Error rate spiked because..."

**Integration**:
- Reads from `logs.*` schema (no writes)
- Can export to Review app for code investigation

---

### Build Service (Phase 2)
**Purpose**: Terminal interface and autonomous coding

**Phase 1 Features**:
- Web-based terminal (xterm.js)
- Run shell commands
- Cloud CLI support (AWS, GCP, Azure)
- GitHub CLI integration
- Terminal output captured in Logging service

**Phase 2 Features**:
- **OpenHands integration** (autonomous coding)
- Give OpenHands a spec → it implements feature autonomously
- Runs tests, fixes failures, creates PR
- All activity logged to Logging service
- Output auto-reviewed in Critical mode before PR

**Bounded Context**: Development environment
- `Session` = active terminal or OpenHands task
- `Command` = individual command execution

**Database Schema** (`builds.*`):
```sql
CREATE TABLE builds.sessions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    session_type VARCHAR(20),  -- 'terminal', 'openhands'
    status VARCHAR(20),        -- 'active', 'completed', 'failed'
    created_at TIMESTAMP DEFAULT NOW(),
    ended_at TIMESTAMP
);

CREATE TABLE builds.commands (
    id BIGSERIAL PRIMARY KEY,
    session_id INT REFERENCES builds.sessions(id) ON DELETE CASCADE,
    command TEXT,
    output TEXT,
    exit_code INT,
    executed_at TIMESTAMP DEFAULT NOW()
);
```

**API Endpoints**:
```
POST   /api/build/sessions            - Create terminal session
WS     /ws/build/terminal/:id         - Terminal I/O stream
POST   /api/build/openhands           - Start OpenHands task (Phase 2)
GET    /api/build/openhands/:id       - Get task status (Phase 2)
```

**Integration**:
- Logs sent to Logging service
- OpenHands output auto-reviewed in Review service
- Analytics tracks build success rate

---

## Development Workflow (Hybrid AI Team)

### Roles
1. **Mike**: Project orchestrator, final approval
2. **Claude** (via API): Architecture, strategic review (Critical mode)
3. **OpenHands + Ollama**: Autonomous implementation (70-80% of work)
4. **GitHub Copilot**: IDE assistance for manual coding (5-10%)

### Standard Feature Development Process

1. **Issue Creation** (Mike)
   - Create GitHub issue with acceptance criteria
   - Label appropriately

2. **Architecture & Spec** (Claude, <30 min)
   - Design high-level architecture
   - Create detailed spec using `.docs/specs/TEMPLATE.md`
   - Specify bounded context, layering, abstractions
   - Save spec to issue or `.docs/specs/`

3. **Autonomous Implementation** (OpenHands + Ollama)
   - Mike triggers: `openhands --task "Implement issue #42"`
   - OpenHands works fully autonomously:
     - Creates feature branch
     - Writes tests first (TDD)
     - Implements feature per spec
     - Runs tests, fixes failures
     - Commits with conventional messages
     - Creates PR
   - Duration: 30 min - 2 hours (unattended)
   - **Crash-proof**: Checkpoint/resume capability

4. **Strategic Review** (Claude, <30 min)
   - Review PR in Critical mode
   - Check mental models (bounded context, layering, abstractions, scope)
   - Verify coding standards, security, performance
   - Comment with specific, actionable feedback

5. **Acceptance Review** (Mike)
   - Use Preview/Skim modes to understand changes
   - Verify acceptance criteria 100% met
   - Review Claude's feedback
   - Test feature manually
   - Approve or request changes

6. **Merge** (Mike)
   - Squash merge to `development`
   - Delete feature branch
   - Issue auto-closed

7. **Release** (Mike)
   - Merge `development` to `main` when ready
   - Tag with version

### Benefits of Hybrid Approach
- ✅ 80% of work runs unattended (OpenHands)
- ✅ Claude crash risk minimal (only 10-15% of work time)
- ✅ No API costs for implementation (local Ollama)
- ✅ Can work overnight on complex features
- ✅ Mike focuses on high-value orchestration

---

## Installation & Deployment

### One-Command Setup
```bash
git clone https://github.com/mikejsmith1985/devsmith-modular-platform.git
cd devsmith-modular-platform
./setup.sh
```

**Setup script handles**:
1. Docker and Docker Compose check
2. Ollama installation (if not present)
3. RAM detection and model recommendation
4. Model download (auto-selects based on available RAM)
5. PostgreSQL initialization (via Docker)
6. Database migrations
7. GitHub OAuth app creation wizard
8. Environment variable configuration
9. Service health checks
9. Launch platform at `http://localhost:3000`

### Docker Compose Architecture
```yaml
services:
  nginx:        # Gateway, port 3000
  postgres:     # Database
  redis:        # WebSocket pub/sub, caching
  portal:       # Port 3001 (internal)
  review:       # Port 3002 (internal)
  logging:      # Port 3003 (internal)
  analytics:    # Port 3004 (internal)
  # build:      # Port 3005 (Phase 2)
```

All services behind Nginx gateway - users only access `localhost:3000`.

---

## Non-Functional Requirements

### Performance
- Page load: <500ms (server-side rendering)
- AI analysis (Preview/Skim): <3 seconds
- AI analysis (Critical): <30 seconds for 500-line file
- WebSocket latency: <100ms
- Log ingestion: 1000+ entries/second

### Scalability
- Handle 100+ concurrent users (single instance)
- Database: Designed for 10M+ log entries
- Services independently scalable via Docker

### Security
- GitHub OAuth only (no custom password storage)
- JWT tokens with 24-hour expiry
- HTTPS required in production
- No secrets in code (all in environment variables)
- SQL injection prevention (parameterized queries)
- Input validation on all endpoints

### Reliability
- Health checks every 30 seconds
- Automatic restart on service crash (Docker)
- Database backups (pg_dump daily)
- Recovery hooks for Claude sessions (`.claude/hooks/`)

### Observability
- All services log to Logging service
- Request tracing (correlation IDs)
- Performance metrics in Analytics
- Service health dashboard in Portal

---

## Testing Requirements

### Unit Tests
- 70%+ coverage required
- Mock external dependencies
- Run on every commit (GitHub Actions)

### Integration Tests
- Test cross-layer flows (handler → service → repository)
- Require test database (Docker)
- Run before merge to `development`

### End-to-End Tests
- Test full user workflows through gateway
- Use Playwright (Phase 2)
- Run before release to `main`

### Manual Testing Checklist
- [ ] Feature works through nginx gateway
- [ ] No browser console errors
- [ ] Responsive design (mobile/desktop)
- [ ] Light/dark mode compatible
- [ ] HTMX interactions work
- [ ] WebSocket connections stable

---

## Documentation Requirements

### For Users
- README.md: Project overview, installation, quick start
- User Guide: How to use each reading mode
- Video tutorials: Common workflows

### For Developers
- ARCHITECTURE.md: Complete system design (already exists)
- DevSmithRoles.md: Team roles and workflows (already exists)
- DevsmithTDD.md: TDD approach
- .docs/specs/TEMPLATE.md: Spec template for OpenHands (already exists)
- API documentation: Swagger/OpenAPI for all endpoints

### For Contributors
- CONTRIBUTING.md: How to contribute
- CODE_OF_CONDUCT.md: Community standards
- Issue templates: Feature request, bug report (already exist)

---

## Success Metrics

### Platform Usage
- Active users per month
- Review sessions created per user
- Most-used reading mode (expect Critical for HITL)
- User retention (30-day, 90-day)

### Educational Impact
- Time to complete first Critical review
- Improvement in issue detection over time
- User-reported confidence in reviewing AI code

### Development Efficiency
- Time from issue to merged PR
- OpenHands success rate (% of PRs merged without major changes)
- Claude review turnaround time
- Test coverage across services

### Platform Stability
- Uptime (target: 99.5%)
- Mean time between failures
- Mean time to recovery
- V8 crash recovery success rate

---

## Future Enhancements (Not MVP)

### Phase 2
- Build service with OpenHands integration
- Real-time collaboration in all apps (not just Review)
- Mobile app (React Native or Progressive Web App)
- VS Code extension (launch Review from editor)
- **Enhanced Pre-commit Hook Integration** (Developer Experience Enhancement)

### Phase 3
- Team features (shared workspaces, org accounts)
- Integration with Jira, Linear, etc.
- Custom LLM model fine-tuning
- Enterprise deployment (Kubernetes)

### Phase 4
- Marketplace for custom reading modes
- Plugin system for language-specific analyzers
- Integration with CI/CD pipelines
- Automated code improvement suggestions

---

## Enhanced Pre-commit Hook Integration (Phase 2)

### Overview
Integrate the enhanced pre-commit validation system into the DevSmith platform's Logging and Analytics services, providing developers and AI agents with intelligent, actionable feedback on code quality issues before commits.

### Core Features

#### 1. Machine-Readable Output (JSON API)
Enable programmatic access to validation results for AI agents and tools:

**Output Format**:
```json
{
  "status": "failed|passed",
  "duration": 45,
  "mode": "standard|quick|thorough",
  "issues": [
    {
      "type": "test_mock_expectation|style|security|...",
      "severity": "error|warning",
      "file": "path/to/file.go",
      "line": 42,
      "message": "Test failed - 18 mock expectations not met",
      "suggestion": "Add Mock.On() expectations - see docs §5.1",
      "autoFixable": false,
      "fixCommand": "go fmt file.go",
      "context": "...code snippet..."
    }
  ],
  "grouped": {
    "high": [...],    // Blocking errors
    "medium": [...],  // Should fix
    "low": [...]      // Can defer
  },
  "dependencyGraph": {
    "nodes": ["build_errors", "tests", "style"],
    "edges": [...],
    "fix_order": ["build_errors", "tests", "style"]
  },
  "summary": {
    "total": 25,
    "errors": 2,
    "warnings": 23,
    "autoFixable": 15
  }
}
```

**API Endpoints** (Logging Service):
```
POST   /api/validation/submit     - Submit pre-commit results
GET    /api/validation/history    - Get validation history
GET    /api/validation/stats      - Aggregate statistics
WS     /ws/validation             - Real-time validation stream
```

#### 2. Issue Prioritization & Grouping
Automatically categorize validation issues by priority:

- **High Priority (Blocking)**: Build errors, test failures, critical security issues
- **Medium Priority (Should Fix)**: Security warnings, error handling gaps, unused imports
- **Low Priority (Can Defer)**: Style issues, missing comments, code quality suggestions

**Benefits**:
- AI agents know what to fix first
- Humans see most important issues up front
- Reduces cognitive load during code review

#### 3. Context-Aware Suggestions
Provide actionable guidance with code context:

**Enhanced Issue Display**:
```
ERROR: Test 'TestAggregatorService' - 18 mock expectations not met
  File: internal/analytics/services/aggregator_service_test.go:45

  43: func TestAggregatorService_RunHourlyAggregation(t *testing.T) {
  44:     mockRepo := &testutils.MockAggregationRepository{}
  45:     service := NewAggregatorService(mockRepo, logger)
  46:     // Missing: mockRepo.On("FindByRange", ...).Return(...)

  Suggestion: Add mock expectations before service call
  Template: mockRepo.On("FindByRange", mock.Anything, ...).Return([]*models.Aggregation{}, nil)

  See: .docs/copilot-instructions.md §5.1
  Similar fixes: git log -S "FindByRange" --oneline | head -1
```

#### 4. Parallel Execution
Speed up validation with concurrent checks:

**Performance Gains**:
- Sequential: ~60 seconds
- Parallel: ~15 seconds (4x faster)

**Implementation**:
```bash
# Run independent checks in parallel
{
  go fmt ./... &
  go vet ./... &
  golangci-lint run ./... &
  go test -short ./... &
  wait
}
```

#### 5. Auto-Fix Mode
Automatically correct simple issues:

**Supported Fixes**:
- Code formatting (`go fmt`, `goimports`)
- Unused imports removal
- Basic comment templates
- Parameter type combinations

**Usage**:
```bash
.git/hooks/pre-commit --fix
# ✓ Auto-fixed 12 issue(s)
# ⚠ 3 issues require manual attention
```

#### 6. Smart Caching
Skip validation for unchanged files:

**Cache Strategy**:
- Store MD5 hash of each file
- Compare before running checks
- Only validate modified files
- Clear cache on branch switch

**Performance Impact**:
- 50-80% faster for incremental commits
- Especially beneficial for large codebases

#### 7. Issue Context Extraction
Show code snippets around problems:

**Context Window**: ±3 lines around issue
**Benefits**:
- AI agents understand problem without reading entire file
- Faster diagnosis and fixes
- Reduces file I/O operations

#### 8. Dependency Graph
Visualize issue relationships:

```
Issue Dependencies:
  Build Error (logs service)
    ↳ Blocks: All logs service tests
    ↳ Blocks: Integration tests

  Fix Priority:
    1. Fix build → 2. Run tests → 3. Fix style issues
```

#### 9. Progressive Validation Modes

**Quick Mode** (~5 seconds):
- Formatting checks
- Critical build errors only
- Use during rapid development

**Standard Mode** (~15 seconds):
- All checks in parallel
- Default for pre-commit hook
- Balanced speed/thoroughness

**Thorough Mode** (~60 seconds):
- Includes race detection
- More comprehensive linting
- Use before creating PR

**Usage**:
```bash
.git/hooks/pre-commit --quick     # Fast feedback
.git/hooks/pre-commit              # Normal (default)
.git/hooks/pre-commit --thorough   # Comprehensive
```

#### 10. Agent-Specific Guide
Provide AI agents with fix patterns:

**Guide File**: `.git/hooks/pre-commit-agent-guide.json`

**Contents**:
- Common error patterns
- Step-by-step fix instructions
- Code examples (before/after)
- Auto-fixable flags
- Priority recommendations

**Example Pattern**:
```json
{
  "missing_mock_setup": {
    "pattern": "mock expectation(s) not met",
    "severity": "error",
    "fix_steps": [
      "1. Read test file to identify test function",
      "2. Locate mock object (type Mock*)",
      "3. Add mockObj.On(\"Method\", ...).Return(...)",
      "4. Ensure m.Called() is used in mock"
    ],
    "example_code": "mockRepo.On(\"FindByRange\", ...).Return(...)",
    "auto_fixable": false
  }
}
```

#### 11. Interactive Query Mode
Allow agents to query specific issues:

**Commands**:
```bash
# Explain a specific test failure
.git/hooks/pre-commit --explain TestAggregatorService

# Get fix suggestion for specific line
.git/hooks/pre-commit --suggest-fix file.go:42

# Check only specific tool
.git/hooks/pre-commit --check-only golangci-lint
```

**Benefits**:
- Targeted investigation
- Faster debugging
- Better integration with AI workflows

#### 12. LSP-Compatible Diagnostics
Export validation results for IDEs:

**Output Format** (Language Server Protocol):
```json
[
  {
    "uri": "file:///path/to/file.go",
    "range": {
      "start": {"line": 41, "character": 0},
      "end": {"line": 41, "character": 999}
    },
    "severity": 1,  // 1=error, 2=warning
    "source": "pre-commit",
    "message": "Test failed - 18 mock expectations not met",
    "code": "test_mock_expectation"
  }
]
```

**Usage**:
```bash
.git/hooks/pre-commit --output-lsp > diagnostics.json
```

**Integration**:
- VS Code can consume LSP diagnostics
- Display issues inline in editor
- Click to jump to problem location

### Integration with DevSmith Services

#### Logging Service Integration
**New Database Schema** (`logs.validations`):
```sql
CREATE TABLE logs.validation_runs (
    id BIGSERIAL PRIMARY KEY,
    user_id INT,
    repository VARCHAR(255),
    branch VARCHAR(100),
    commit_sha VARCHAR(40),
    mode VARCHAR(20),  -- 'quick', 'standard', 'thorough'
    duration INT,      -- seconds
    status VARCHAR(20), -- 'passed', 'failed'
    issues_json JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_validation_user ON logs.validation_runs(user_id, created_at DESC);
CREATE INDEX idx_validation_repo ON logs.validation_runs(repository, created_at DESC);
CREATE INDEX idx_validation_status ON logs.validation_runs(status, created_at DESC);
```

**API Enhancements**:
```
POST   /api/logs/validation        - Submit validation results
GET    /api/logs/validation/:id    - Get specific run details
GET    /api/logs/validation/recent - Get recent validations
WS     /ws/logs/validation         - Stream validation events
```

#### Analytics Service Integration
**New Metrics**:
- Validation success rate over time
- Most common issue types
- Average fix time per issue type
- Auto-fix effectiveness rate
- Validation performance trends

**New API Endpoints**:
```
GET    /api/analytics/validation/trends      - Trend analysis
GET    /api/analytics/validation/top-issues  - Most common problems
GET    /api/analytics/validation/fix-time    - Time to fix by issue type
GET    /api/analytics/validation/agent-stats - AI agent fix success rate
```

**Dashboard Visualizations**:
- Validation pass/fail rate (time series)
- Issue type distribution (pie chart)
- Fix priority heatmap
- Auto-fix vs manual fix ratio
- Agent fix success rate by issue type

#### Portal Service Integration
**Validation Dashboard**:
- Recent validation runs (last 10)
- Overall pass rate (7-day trend)
- Top 5 recurring issues
- Quick links to detailed analytics

**User Profile Enhancements**:
- Validation statistics
- Achievement badges ("No failures for 30 days")
- Comparison with team average

### Benefits for AI Agents

#### For OpenHands
1. **Structured Feedback**: JSON output is easily parsed
2. **Priority Guidance**: Knows what to fix first
3. **Fix Instructions**: Step-by-step remediation
4. **Code Context**: No need to re-read entire files
5. **Auto-fix Option**: Handle simple issues automatically

**Workflow**:
```
1. OpenHands implements feature
2. Runs: .git/hooks/pre-commit --json
3. Parses JSON to identify issues
4. Runs: .git/hooks/pre-commit --fix (handles 60% of issues)
5. Uses agent guide to fix remaining issues
6. Re-runs validation until passed
7. Creates PR
```

#### For Claude/Copilot
1. **Quick Assessment**: --quick mode for fast feedback
2. **Explain Feature**: Get detailed issue explanations
3. **Suggest-Fix**: Get targeted fix recommendations
4. **LSP Integration**: Issues visible in IDE
5. **Historical Context**: See similar fixes from git log

### User Benefits

#### For Developers
1. **Faster Feedback**: 4x faster with parallel execution
2. **Clear Priorities**: Know what's critical vs. optional
3. **Actionable Guidance**: Specific fix instructions
4. **Less Context Switching**: See code snippets inline
5. **Learning Tool**: Understand common patterns

#### For Teams
1. **Consistent Standards**: Enforced via pre-commit hook
2. **Quality Metrics**: Track improvement over time
3. **Knowledge Sharing**: Common patterns documented
4. **Reduced Review Time**: Fewer trivial issues in PRs
5. **Better Collaboration**: Agents and humans use same system

### Implementation Timeline

**Week 1-2**: Core infrastructure
- Logging service schema and API
- Validation result ingestion
- Basic analytics queries

**Week 3-4**: Enhanced features
- Dependency graph computation
- Agent guide JSON finalization
- Auto-fix improvements

**Week 5-6**: Integration & UI
- Portal dashboard widgets
- Analytics visualizations
- WebSocket streaming

**Week 7-8**: Testing & Documentation
- End-to-end testing
- User documentation
- Agent integration guides

### Success Metrics

**Adoption**:
- % of commits using enhanced hook
- % of teams using validation dashboard

**Quality**:
- Reduction in PR review cycles
- Decrease in production bugs
- Improvement in test coverage

**Performance**:
- Average validation time
- Auto-fix success rate
- Agent fix success rate

**User Satisfaction**:
- Developer feedback scores
- Time saved per week
- Learning effectiveness ratings

---

## Open Questions (To Be Resolved)

1. **LLM Response Caching**: Redis or PostgreSQL JSONB? (Lean Redis for speed)
2. **WebSocket Scaling**: Redis pub/sub sufficient or need dedicated message broker?
3. **AI Model Versioning**: How to handle Ollama model updates without breaking sessions?
4. **Collaboration Persistence**: Store collaboration events in database or ephemeral?
5. **Rate Limiting**: Per-user or per-service? How to handle OpenHands long-running tasks?

---

## Constraints & Assumptions

### Constraints
- GitHub OAuth only (no Google, email/password)
- English language only (MVP)
- Desktop browsers only (mobile in Phase 2)
- Ollama models <70B (hardware limitations for most users)

### Assumptions
- Users have GitHub account
- Users understand basic programming concepts
- Users have stable internet (for GitHub OAuth, optional Claude API)
- Users willing to download 10GB+ for Ollama models

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-18 | Claude | Initial requirements document |
| 2.0 | 2025-10-18 | Claude | Complete rewrite with mental models, cognitive load, hybrid AI workflow, Go stack |
| 2.1 | 2025-11-16 | Copilot | Updated to reflect template improvements and review routing fixes. See commit ef3b0c4. |

---

## GitHub Repository Integration - Phased Implementation Plan

### Overview
Multi-phase approach to GitHub repository integration with the Review service, balancing immediate value delivery with long-term scalability and performance optimization.

---

### Phase 1: Lazy-Load MVP (Week 1) 🚀 **START HERE**

**Goal**: Deliver working GitHub integration with minimal bandwidth usage and instant setup.

**Architecture**: Lazy-load tree structure, fetch files on-demand.

#### Features

**1. Simple Repo Scan Mode** (NEW!)
- **Purpose**: Instant repository profiling without downloading all files
- **What It Fetches**:
  - `README.md` (or README.rst, README.txt) - Project overview
  - `package.json` / `go.mod` / `requirements.txt` / `Cargo.toml` - Dependencies
  - `LICENSE` - Licensing information
  - `CONTRIBUTING.md` - Contribution guidelines (if exists)
  - Root-level config: `.gitignore`, `docker-compose.yml`, `Makefile`, `.github/workflows/*`
  - Entry point files (auto-detected by language):
    - Go: `main.go`, `cmd/*/main.go`
    - JavaScript: `index.js`, `src/index.js`, `app.js`
    - Python: `main.py`, `app.py`, `__init__.py`
    - Rust: `main.rs`, `lib.rs`
- **AI Analysis**: Runs in **Preview Mode** on fetched files
- **Output**:
  - Project description (extracted from README)
  - Technology stack (detected from dependency files)
  - Architecture pattern (inferred from structure)
  - Setup instructions (parsed from README)
  - Key entry points identified
  - Dependencies summary
- **Performance**: ~5-8 files, ~50-100KB total, **<2 seconds**
- **UI**: Single button "Quick Repo Scan" → instant results

**2. Full Repository Browser**
- **Tree Structure**: Fetch GitHub tree API (~100KB for 1000-file repo)
- **On-Demand File Fetch**: Only fetch file contents when user clicks/selects
- **Multi-Select**: Select multiple files → batch fetch → analyze
- **Search/Filter**: Client-side tree filtering (no API calls)
- **Progress Indicator**: "Fetching repository structure..." during tree load

#### Technical Implementation

**Authentication**:
- Reuse Portal's GitHub OAuth token from Redis session
- Token passed to Review service via session store
- No separate GitHub token input required

**Backend Endpoints**:
```go
// Fetch repository tree structure (no file contents)
GET /api/review/github/tree
  Query: ?url=github.com/owner/repo&branch=main
  Response: { tree: [...], entry_points: [...] }

// Fetch single file contents
GET /api/review/github/file
  Query: ?url=github.com/owner/repo&path=src/main.go&branch=main
  Response: { content: "...", language: "go", size: 1234 }

// Quick repo scan (Simple Repo Scan Mode)
GET /api/review/github/quick-scan
  Query: ?url=github.com/owner/repo&branch=main
  Response: { 
    readme: "...",
    dependencies: {...},
    entry_points: [...],
    config_files: [...],
    ai_analysis: {...}  // Preview mode analysis
  }
```

**Frontend Components** (Already Created):
- ✅ `FileTabs.jsx` - Tab management for multiple open files
- ✅ `FileTreeBrowser.jsx` - Hierarchical tree with search, multi-select
- ⏳ `RepoImportModal.jsx` - GitHub URL input, branch selection, scan mode choice

**UI Flow**:
```
1. User clicks "Import from GitHub" button
2. Modal opens with:
   - GitHub URL input (e.g., github.com/owner/repo)
   - Branch dropdown (auto-populated from API)
   - Two buttons:
     a) "Quick Repo Scan" (Simple Mode) - Instant profile
     b) "Full Repository Browser" - Explore all files
3a. Quick Scan:
   - Fetches ~5-8 core files (~100KB)
   - Runs Preview Mode AI analysis
   - Shows results in Analysis pane
   - ~2 seconds total
3b. Full Browser:
   - Fetches tree structure (~100KB)
   - Shows FileTreeBrowser in sidebar
   - User selects files → fetched on-demand
   - Opens in FileTabs for editing/analysis
```

**Performance Benefits**:
- **Simple Scan**: 100KB vs 50MB full clone (500x smaller)
- **Full Browser**: Fetch 5 files = 25KB vs 50MB (2000x smaller)
- **Time Savings**: 2s vs 30-60s for clone
- **Rate Limits**: 5,000 API calls/hour (authenticated) vs 60/hour (unauthenticated)

#### Limitations (Accept for MVP)
- ❌ No file caching (re-fetch on page refresh)
- ❌ No offline support
- ❌ No semantic search across files
- ❌ Max 100 files per analysis (prevent API overload)

#### Deliverables
- [ ] Backend: `/api/review/github/tree` endpoint
- [ ] Backend: `/api/review/github/file` endpoint
- [ ] Backend: `/api/review/github/quick-scan` endpoint
- [ ] Frontend: `RepoImportModal.jsx` component
- [ ] Frontend: Integrate FileTreeBrowser into ReviewPage sidebar
- [ ] Frontend: Connect file selection to FileTabs
- [ ] UI: Loading states and progress indicators
- [ ] UI: "Quick Repo Scan" vs "Full Browser" mode selection
- [ ] Testing: E2E test for full workflow
- [ ] Docs: Update user guide with GitHub integration

**Acceptance Criteria**:
- ✅ User can paste GitHub URL and see repo structure in <3 seconds
- ✅ Quick Repo Scan completes in <2 seconds with meaningful analysis
- ✅ User can select files and open them in tabs
- ✅ Multiple files can be analyzed together
- ✅ GitHub rate limits respected (authenticated token)
- ✅ Error handling for private repos (403) and rate limits (429)

---

### Phase 2: Performance Optimization (Week 2-3)

**Goal**: Improve user experience with caching and smarter prefetching.

#### Features

**1. Browser-Side Caching (IndexedDB)**
- Cache fetched file contents keyed by `${repoUrl}:${filePath}:${commitSHA}`
- Persist across page refreshes
- Auto-expire after 7 days or on new commits
- Cache size limit: 50MB per repository

**2. Intelligent Prefetching**
- When user opens folder in tree, prefetch immediate children (if <10 files)
- When user opens file, prefetch files in same directory
- Prefetch entry point files automatically (main.go, index.js, etc.)

**3. Batch API Optimization**
- Combine multiple file requests into single GitHub API call
- Use GitHub's blob API with multiple SHAs
- Reduces API calls by 80% for multi-file workflows

#### Technical Implementation

**IndexedDB Schema**:
```javascript
// Database: devsmith-review-cache
// Store: github-files
{
  key: "github.com/owner/repo:src/main.go:abc123",  // repoUrl:path:commitSHA
  value: {
    content: "...",
    language: "go",
    size: 1234,
    cached_at: "2025-11-07T12:00:00Z",
    commit_sha: "abc123"
  }
}
```

**Cache Wrapper**:
```javascript
class GitHubCache {
  async getFile(repoUrl, path, commitSHA) {
    const key = `${repoUrl}:${path}:${commitSHA}`;
    
    // Check cache first
    const cached = await this.db.get(key);
    if (cached && !this.isExpired(cached)) {
      return cached;
    }
    
    // Fetch from API
    const file = await api.fetchFile(repoUrl, path);
    
    // Cache result
    await this.db.put(key, { ...file, cached_at: new Date() });
    
    return file;
  }
  
  isExpired(cached) {
    const age = Date.now() - new Date(cached.cached_at);
    return age > 7 * 24 * 60 * 60 * 1000;  // 7 days
  }
}
```

#### Deliverables
- [ ] IndexedDB cache implementation
- [ ] Cache invalidation on new commits
- [ ] Prefetching heuristics
- [ ] Batch API optimization
- [ ] Cache statistics in UI (hit rate, storage used)

**Acceptance Criteria**:
- ✅ Previously viewed files load instantly from cache
- ✅ Cache persists across browser sessions
- ✅ Cache invalidates on repo updates
- ✅ Prefetching reduces wait time for common workflows
- ✅ Batch API reduces GitHub API calls by 80%

---

### Phase 3: Backend Indexing (Optional - Month 2+)

**Goal**: Enable advanced features like semantic search and offline analysis.

**⚠️ Note**: Only implement if Phase 1-2 prove insufficient for user needs.

#### Features

**1. Background Repository Indexing**
- User triggers: "Index this repository for offline use"
- Background job fetches entire repo asynchronously
- Stores in PostgreSQL `reviews.repo_cache` table
- Shows progress: "Indexing... 234/1000 files (23%)"

**2. Semantic Search Across Repository**
- "Find all error handling patterns" → searches indexed content
- "Show all database queries" → finds SQL across all files
- AI-powered search (not just text matching)

**3. Offline Analysis**
- Once indexed, no GitHub API calls needed
- Instant file access from database cache
- Analysis works even if GitHub is down

**4. Cache Management**
- Auto-update: Re-index on webhook events (new commits)
- Manual trigger: "Re-index now" button
- Cache expiry: Remove repos not accessed in 30 days
- Storage limits: Max 100 repos or 10GB per user

#### Technical Implementation

**Database Schema**:
```sql
-- Indexed repositories
CREATE TABLE reviews.repo_index (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    repo_url VARCHAR(255) NOT NULL,
    branch VARCHAR(100) DEFAULT 'main',
    last_commit_sha VARCHAR(40),
    file_count INT,
    total_size_bytes BIGINT,
    indexed_at TIMESTAMP,
    last_accessed TIMESTAMP,
    status VARCHAR(20),  -- 'indexing', 'complete', 'failed'
    UNIQUE(user_id, repo_url, branch)
);

-- Cached file contents
CREATE TABLE reviews.repo_files (
    id BIGSERIAL PRIMARY KEY,
    repo_id INT REFERENCES reviews.repo_index(id) ON DELETE CASCADE,
    file_path VARCHAR(500) NOT NULL,
    content TEXT,
    content_hash VARCHAR(64),  -- SHA256 for deduplication
    language VARCHAR(50),
    size_bytes INT,
    ast_data JSONB,  -- Abstract Syntax Tree for advanced search
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(repo_id, file_path)
);

CREATE INDEX idx_repo_files_path ON reviews.repo_files(repo_id, file_path);
CREATE INDEX idx_repo_files_hash ON reviews.repo_files(content_hash);
CREATE INDEX idx_repo_index_user ON reviews.repo_index(user_id, last_accessed DESC);
```

**Background Indexing Worker**:
```go
// Background job triggered by user or webhook
func IndexRepository(ctx context.Context, repoURL, branch string, userID int) error {
    // Update status to 'indexing'
    repo := createRepoIndex(userID, repoURL, branch, "indexing")
    
    // Fetch tree
    tree, err := github.GetTree(ctx, repoURL, branch, recursive=true)
    repo.FileCount = len(tree.Entries)
    repo.Save()
    
    // Fetch files in batches of 10
    for i := 0; i < len(tree.Entries); i += 10 {
        batch := tree.Entries[i:min(i+10, len(tree.Entries))]
        
        // Parallel fetch
        files := fetchFilesBatch(ctx, repoURL, branch, batch)
        
        // Store in database
        for _, file := range files {
            saveRepoFile(repo.ID, file)
        }
        
        // Update progress
        progress := float64(i) / float64(len(tree.Entries)) * 100
        notifyProgress(userID, repo.ID, progress)
    }
    
    // Mark complete
    repo.Status = "complete"
    repo.IndexedAt = time.Now()
    repo.Save()
    
    return nil
}
```

**API Endpoints**:
```
POST   /api/review/repos/index
  Body: { repo_url: "...", branch: "main" }
  Response: { job_id: 123, status: "indexing" }

GET    /api/review/repos/:id/status
  Response: { status: "indexing", progress: 45, files: 450/1000 }

GET    /api/review/repos/:id/search
  Query: ?q=error handling patterns&mode=semantic
  Response: { matches: [...] }

DELETE /api/review/repos/:id
  Response: { deleted: true }
```

#### Deliverables
- [ ] Database schema for repo indexing
- [ ] Background indexing worker
- [ ] Progress tracking (WebSocket or polling)
- [ ] Semantic search implementation
- [ ] Cache invalidation (webhook or TTL)
- [ ] Storage quota management
- [ ] UI for indexed repos management

**Acceptance Criteria**:
- ✅ User can trigger repository indexing
- ✅ Indexing shows progress (234/1000 files)
- ✅ Once indexed, files load instantly (no API calls)
- ✅ Semantic search works across entire repo
- ✅ Cache invalidates on new commits (webhook)
- ✅ Storage limits enforced (100 repos or 10GB)

---

### Phase 4: Enterprise Features (Future)

**Goal**: Scale to team and enterprise use cases.

#### Features
- **Shared Repository Cache**: Team members share indexed repos
- **Webhook Integration**: Auto-update cache on GitHub push events
- **Custom Indexing Rules**: Skip node_modules, vendor, etc.
- **Multi-Repository Analysis**: Compare patterns across repos
- **AI Model Fine-Tuning**: Train on team's codebase patterns
- **Private GitHub Enterprise**: Support GHE instances

---

### Implementation Roadmap

**Week 1** (Phase 1 MVP):
- Day 1-2: Backend tree/file/quick-scan endpoints
- Day 3-4: RepoImportModal component + sidebar integration
- Day 5: Testing + documentation

**Week 2-3** (Phase 2 Optimization):
- Week 2: IndexedDB caching implementation
- Week 3: Prefetching + batch API optimization

**Month 2+** (Phase 3 - If Needed):
- Week 1-2: Database schema + indexing worker
- Week 3: Semantic search
- Week 4: UI + cache management

---

### Success Metrics

**Phase 1**:
- ⏱️ Time to repo profile: <2 seconds (Quick Scan)
- ⏱️ Time to tree load: <3 seconds
- 📦 Bandwidth saved: 95%+ vs full clone
- 🎯 User satisfaction: "Feels like GitHub" experience

**Phase 2**:
- 📈 Cache hit rate: >60% for repeat visits
- ⚡ File load time: <100ms for cached files
- 📉 API calls reduced: 80% via batching

**Phase 3** (If implemented):
- 🔍 Semantic search accuracy: >90%
- 💾 Storage efficiency: <100MB per 1000-file repo
- ⚙️ Indexing speed: >10 files/second

---

### Decision Gates

**Before Phase 2**: Ask users:
- "Do you return to the same repos frequently?"
- "Do you want files to load instantly on revisits?"
- If YES → Implement caching

**Before Phase 3**: Analyze metrics:
- Cache hit rate >60%?
- Users requesting "search all files" feature?
- Users hitting GitHub rate limits?
- If YES to 2+ → Implement indexing

---

### Technical Constraints

**GitHub API Rate Limits**:
- Authenticated: 5,000 requests/hour
- Per-user: ~83 requests/minute
- Strategy: Batch requests, cache aggressively

**Browser Storage**:
- IndexedDB: ~50MB default quota
- Can request more (usually up to 1GB)
- Strategy: LRU eviction when quota reached

**Database Storage**:
- PostgreSQL: Unlimited (within server capacity)
- Estimate: 100KB per file average
- Strategy: Compression + deduplication by content hash

---

## References
- ARCHITECTURE.md - Complete system design
- DevSmithRoles.md - Hybrid AI team roles
- DevsmithTDD.md - Test-driven development approach
- .docs/specs/TEMPLATE.md - OpenHands spec template
- .claude/hooks/ - Crash recovery system
