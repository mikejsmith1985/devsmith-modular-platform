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
- **AI-Assisted Development**: Claude Haiku for documentation, Cursor/Copilot for implementation
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

## Mental Models for Understanding This Platform

### Overview: Cognitive Load and Code Comprehension

This platform is designed around **managing cognitive load** - the mental effort required to understand and work with code. Our architecture, tooling, and workflows all aim to:

1. **Minimize unnecessary complexity** (reduce wasted mental effort)
2. **Simplify inherent complexity** (make hard things approachable)
3. **Build strong mental frameworks** (enable reasoning and transfer of learning)

### The Three Types of Mental Effort

#### Intrinsic Complexity
The unavoidable difficulty inherent in a task itself.

**Example:** Understanding how GitHub OAuth works has inherent complexity - it's not simple by nature.

**Our Strategy:**
- Use Go's explicit error handling (clearer than exceptions)
- Templ's compile-time checks (catch errors early)
- Clear naming conventions
- Modular services (tackle one problem at a time)

#### Wasted Effort
Mental energy spent on confusion, poor documentation, or unclear architecture.

**Example:** Debugging why a variable is undefined because scope wasn't clear.

**Our Strategy:**
- Configuration over hardcoding (single source of truth)
- Gateway-first design (no mysterious port conflicts)
- Crash recovery hooks (reduce stress of work loss)
- Clear bounded contexts (avoid mixing concerns)

#### Framework-Building Effort
Mental work that helps you build transferable understanding.

**Example:** Learning layered architecture once, then applying it everywhere.

**Our Strategy:**
- Consistent patterns across all services
- Explicit documentation of mental models (this section!)
- Comprehensive GitHub issues with acceptance criteria
- Architecture documents that explain "why" not just "how"

---

### Core Mental Models

These four frameworks are essential for understanding any part of this platform:

#### 1. Bounded Contexts (Horizontal Separation)

**Concept:** The same word can mean different things in different business domains.

**Real-World Example:**
- "Customer" in Sales = someone with a territory and sales pipeline
- "Customer" in Support = someone with support tickets and assigned agents
- Same word, completely different data and behaviors

**In Our Platform:**

```
Portal Service:
├── User (authentication context)
│   ├── github_id, github_username
│   ├── login(), logout()
│   └── Concerns: Identity, sessions

Review Service:
├── User (code review context)
│   ├── reviews_created, reviews_participated
│   ├── submitReview(), requestReview()
│   └── Concerns: Review ownership, permissions

Logging Service:
├── User (audit context)
│   ├── log_entries_created
│   ├── logAction(), queryLogs()
│   └── Concerns: Audit trail, activity tracking
```

**Why This Matters:**
- Prevents "God objects" that try to be everything to everyone
- Enables independent service development
- Reduces coupling between services
- Makes code easier to reason about within its context

**When Reading Code:**
Always ask: "Which bounded context am I in?" The answer changes what entities mean and what behaviors are valid.

---

#### 2. Layered Architecture (Vertical Separation)

**Concept:** Software is organized in layers, each responsible for different concerns.

**The Three Layers:**

```
┌─────────────────────────────────────┐
│   CONTROLLER LAYER (handlers/)     │  ← User interaction
│   - HTTP handlers                  │  ← Request/response
│   - Templ templates                │  ← UI rendering
│   - Input validation               │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   ORCHESTRATION LAYER (services/)  │  ← Business logic
│   - Business rules                 │  ← Service coordination
│   - Data transformation            │  ← Error handling
│   - External API calls             │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   DATA LAYER (db/)                 │  ← Persistence
│   - Database queries               │  ← Transaction management
│   - Schema definitions             │  ← Data integrity
│   - Migrations                     │
└─────────────────────────────────────┘
```

**Same Entity, Different Concerns:**

The "Review" entity exists in all three layers:

```go
// CONTROLLER LAYER (handlers/review_handler.go)
// Concern: HTTP request/response, user input
type ReviewRequest struct {
    CodeContent string `json:"code_content" binding:"required"`
    ReadingMode string `json:"reading_mode" binding:"required,oneof=preview skim scan detailed critical"`
}

// ORCHESTRATION LAYER (services/review_service.go)
// Concern: Business logic, AI interaction
func (s *ReviewService) AnalyzeCode(ctx context.Context, review *models.Review) error {
    // Call AI service (Claude API, OpenAI, etc.)
    // Parse results
    // Apply business rules
}

// DATA LAYER (db/reviews.go)
// Concern: Database persistence
func (r *ReviewRepository) Save(ctx context.Context, review *models.Review) error {
    query := `INSERT INTO reviews.reviews (user_id, code_content, status) VALUES ($1, $2, $3)`
    // Database interaction
}
```

**Why This Matters:**
- Each layer has ONE responsibility
- Teams can specialize in layers
- Changes in one layer don't break others
- Testing becomes easier (test each layer independently)

**When Reading Code:**
Always ask: "Which layer am I in?" This tells you what concerns are appropriate and what dependencies you can expect.

---

#### 3. Abstraction vs Implementation

**Concept:** Understand the "what" before diving into the "how."

**Real-World Example:**
- Abstraction: A car has an "accelerate" function
- Implementation: Internal combustion engine with fuel injection, spark plugs, etc.
- You can drive without knowing engine internals

**In Our Platform:**

```go
// ABSTRACTION (interfaces/auth_provider.go)
// The "contract" - what behavior exists
type AuthProvider interface {
    // Authenticate user and return token
    Authenticate(ctx context.Context, code string) (*User, error)

    // Validate token is still valid
    ValidateToken(ctx context.Context, token string) (bool, error)
}

// IMPLEMENTATION (services/github_auth.go)
// The "how" - actual OAuth dance with GitHub
type GitHubAuthProvider struct {
    clientID     string
    clientSecret string
    httpClient   *http.Client
}

func (g *GitHubAuthProvider) Authenticate(ctx context.Context, code string) (*User, error) {
    // Step 1: Exchange code for access token
    // Step 2: Call GitHub API to get user info
    // Step 3: Create or update user in database
    // ... 50+ lines of OAuth details ...
}
```

**Why This Matters:**
- **Most code reading happens at the abstraction level**
- You only dive into implementation when debugging or extending
- Understanding abstractions lets you reason about the whole system
- Implementations can change without breaking your mental model

**When Reading Code:**
Start with interfaces and abstract types. Only read concrete implementations when you need to understand specific behavior or fix a bug.

---

#### 4. Scope and Context

**Concept:** Variables and functions have limited "lifespans" and visibility.

**The Scope Hierarchy:**

```go
// PACKAGE-LEVEL SCOPE
// Visible throughout the entire package
var GlobalConfig *Config

// STRUCT SCOPE
// Visible to methods on this struct
type ReviewService struct {
    aiClient AIProvider  // Accessible to all methods (interface for Claude, OpenAI, etc.)
    repo     *ReviewRepository
}

// FUNCTION SCOPE
// Only exists during function execution
func (s *ReviewService) AnalyzeCode(ctx context.Context, review *models.Review) error {
    // Local variable - dies when function returns
    result := s.aiClient.Generate(review.CodeContent)

    // Block scope - only exists in this if statement
    if result.Error != nil {
        tempError := fmt.Errorf("AI call failed: %w", result.Error)
        return tempError  // tempError doesn't exist outside this block
    }

    return nil
}
```

**Why This Matters:**
- **Limits where you need to look** when tracking down a variable
- Prevents naming conflicts (same name can exist in different scopes)
- Makes code easier to test (limited dependencies)
- Reduces cognitive load (smaller context windows)

**When Reading Code:**
When you see a variable, ask: "What scope is this?" This immediately limits where it could be defined and what could affect it.

---

### How These Models Work Together

**Example: Reading the Portal Authentication Handler**

```go
// 1. BOUNDED CONTEXT: Portal Service, Authentication Domain
// 2. LAYER: Controller Layer (handles HTTP)
// 3. SCOPE: Function-level

func HandleGitHubCallback(c *gin.Context) {  // Function scope begins
    code := c.Query("code")  // Local variable

    // 4. ABSTRACTION: Using interface, not concrete implementation
    user, err := authProvider.Authenticate(c.Request.Context(), code)

    if err != nil {
        // Error handling in controller layer (appropriate for this layer)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
        return
    }

    // Generate JWT (orchestration concern, but acceptable in handler for simple cases)
    token := generateJWT(user)

    c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}  // Function scope ends - code, user, token no longer exist
```

**Mental Model Analysis:**
1. ✅ **Bounded Context:** Portal/Auth - only deals with authentication
2. ✅ **Layer:** Controller - handles HTTP, delegates to services
3. ✅ **Abstraction:** Uses `AuthProvider` interface - don't need to know OAuth details
4. ✅ **Scope:** All variables local - easy to track, no side effects

This handler is **easy to understand** because it respects all four mental models.

---

### Application to the Review App

The **Review application** will explicitly implement **five reading modes**, each designed to balance cognitive load appropriately:

#### Preview Mode
**Purpose:** Quick assessment of code structure
**Cognitive Strategy:** Minimal intrinsic load, maximum speed
**Use Case:** "Is this code interesting? Should I look deeper?"

#### Skim Mode
**Purpose:** Understand general functionality and flow
**Cognitive Strategy:** Build high-level mental framework
**Use Case:** "What does this codebase do overall?"

#### Scan Mode
**Purpose:** Find specific information (variables, functions, patterns)
**Cognitive Strategy:** Targeted search, minimize extraneous load
**Use Case:** "Where is this error coming from?"

#### Detailed Mode
**Purpose:** Deep understanding of algorithms and logic
**Cognitive Strategy:** High intrinsic load, maximum comprehension
**Use Case:** "How exactly does this algorithm work?"

#### Critical Mode
**Purpose:** Evaluate quality and identify improvements
**Cognitive Strategy:** Evaluative reasoning, all models active
**Use Case:** "How can I make this better? What's broken?"

**See Section: Service Architecture → Review Service** for implementation details of these modes.

---

### Using Mental Models in Development

#### For Claude (Architecture & Review):
- ✅ Design with bounded contexts clearly defined
- ✅ Ensure layering is respected in architecture
- ✅ Create abstractions before implementations
- ✅ Review PRs using these models as checklist

#### For AI-Assisted Code Review (Implementation):
- ✅ Use these models when reading AI-Assisted Code Review output
- ✅ Specs will explicitly state bounded context and layer
- ✅ Interface definitions provided before asking for implementations
- ✅ Scope kept minimal (function-level when possible)
- ✅ Follow existing patterns (maximize germane load)

#### For Mike (Oversight):
- ✅ Use these models when reading AI-Assisted Code Review output
- ✅ Verify bounded contexts haven't leaked across services
- ✅ Check that layering is clean
- ✅ Ensure abstractions are used appropriately

---

### Key Principles for This Platform

1. **Context is Everything**
   - Same word? Check the bounded context.
   - Same file? Check the layer.
   - Same function? Check the scope.

2. **Abstractions First**
   - Define interfaces before implementations
   - Read at the abstraction level when possible
   - Dive into implementations only when necessary

3. **Layers Stay Separate**
   - Controllers handle HTTP, not business logic
   - Services handle logic, not database details
   - Data layer handles persistence, nothing else

4. **Minimize Scope**
   - Keep variables as local as possible
   - Avoid package-level mutable state
   - Pass dependencies explicitly

5. **Reduce Cognitive Load**
   - Choose clarity over cleverness
   - Explicit over implicit
   - Boring and predictable over novel and surprising

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

### AI/LLM Integration (Platform Features)
- **API Support:** Anthropic Claude API, OpenAI API (user-provided keys)
- **Go Client:** github.com/anthropics/anthropic-sdk-go
- **HTTP Client:** Native Go http.Client with proper timeouts
- **Interface-based:** AIProvider interface for multiple backend support

**Rationale:**
- Multiple APIs: Flexibility, no vendor lock-in
- Native Go HTTP: No SDK version compatibility issues
- Proper timeout handling: Go's context package prevents hanging requests
- Interface pattern: Easy to add new AI providers

### Development Tools

#### Local Development
- **Hot Reload:** Air (Go file watcher, automatic rebuild)
- **Linting:** golangci-lint (comprehensive linter)
- **Formatting:** gofmt (standard Go formatter)
- **API Docs:** Swagger/OpenAPI via swaggo/swag
- **Dependency Management:** Go modules (built-in)

#### AI Development Tools (Supervised Approach)

**Architect & Planner: Claude Code**
- **Role:** High-level architecture, planning, strategic guidance (15-20% of work)
- **Interface:** Claude Code CLI (this tool)
- **Capabilities:**
  - 200K context window (can understand entire codebase)
  - Direct file read/write/edit operations
  - Architecture design and API contracts
  - Database schema design
  - Test execution and validation
  - Implementation planning with code examples
  - Complex problem solving
- **Workflow:**
  - Designs architecture and creates implementation plans
  - Provides detailed specs with file structure, function signatures, patterns
  - Reviews code when requested
  - Assists with debugging and problem-solving

**Primary Implementation: Cursor/Copilot**
- **Role:** AI-assisted code generation during supervised implementation (70-80% of work)
- **Interface:** VS Code extension
- **Capabilities:**
  - Real-time code suggestions as developer types
  - Full function/struct generation from comments
  - Test generation assistance
  - Refactoring suggestions
  - Multi-language support (Go, Templ, SQL, HTMX)
  - Chat interface for explanations and guidance
- **Workflow:**
  - Developer implements features following Claude's plans
  - Copilot provides suggestions, developer reviews and accepts/modifies
  - Maintains human oversight and quality control

**Project Orchestrator: Mike**
- **Role:** Supervises all development, maintains quality (100% oversight)
- **Responsibilities:**
  - Triggers Claude Code for architecture sessions
  - Implements features with Copilot assistance
  - Reviews all code before committing
  - Runs tests and validates functionality
  - Creates PRs and manages merges
  - Ensures adherence to standards and TDD principles

**Development Log (Devlog):**
- `.docs/devlog/` - Human-readable session summaries
  - Date-based entries (`YYYY-MM-DD.md`)
  - Tracks decisions, problems, solutions across sessions
  - **Purpose:** Shared memory between development sessions
  - **Timing:** Updated POST-MERGE after features are completed
  - **Who writes:** Mike with Copilot assistance, or Claude Code if session is active
  - See: `.docs/devlog/README.md` for complete guide

**Crash Recovery (Claude Code V8 Crashes):**
- `.claude/hooks/` - Automated recovery scripts
  - `session-logger.sh` - Logs all actions to markdown
  - `git-recovery.sh` - Auto-commits to recovery branches
  - `recovery-helper.sh` - Interactive recovery tool
- Todo list (`.claude/todos.json`) - Persistent task tracking across crashes
- Recovery branches (`claude-recovery-YYYYMMDD`) - 7-day retention
- Session logs (`.claude/recovery-logs/`) - For resuming interrupted work

**Benefits of Supervised Approach:**
- ✅ Human oversight ensures quality and deep codebase understanding
- ✅ No local LLM complexity or management overhead
- ✅ Claude provides architectural guidance when needed
- ✅ Copilot accelerates implementation without sacrificing control
- ✅ Simple tool chain: Claude Code + Copilot + Git
- ✅ Developer learns codebase through hands-on implementation
- ✅ Crash recovery hooks handle Claude Code V8 crashes gracefully
- ✅ Copilot assists with PR creation for streamlined workflow

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
**Purpose:** AI-driven code review platform with five distinct reading modes, each optimized for managing cognitive load

**Core Philosophy:**
The Review service is the **centerpiece of the platform**, designed to teach users how to effectively read and understand code by providing AI-assisted analysis at different depths. Each mode balances the three types of cognitive load differently to support different reading goals.

---

#### The Five Reading Modes

**1. Preview Mode**

**Purpose:** Rapid assessment of code structure and organization

**Cognitive Load Strategy:**
- **Minimize Intrinsic:** Show only high-level structure (files, folders, imports)
- **Reduce Extraneous:** No implementation details, no line-by-line analysis
- **Maximize Germane:** Build mental map of codebase organization

**What AI Provides:**
- File structure tree with descriptions
- Primary bounded contexts identified
- Technology stack detection
- Architectural pattern recognition (layered, microservices, etc.)
- Entry points and main flows
- External dependencies summary

**Use Cases:**
- Evaluating a new GitHub repo
- Quick assessment before deeper dive
- Understanding project organization
- Determining if code is relevant to your needs

**UI/UX:**
- Tree view of file structure
- Color-coded by layer (controller/service/data)
- Collapsible sections
- Quick filter by file type
- AI summary panel: "This is a Go web service using Gin framework..."

---

**2. Skim Mode**

**Purpose:** Understand overall functionality and key flows without implementation details

**Cognitive Load Strategy:**
- **Minimize Intrinsic:** Focus on abstractions (interfaces, function signatures)
- **Reduce Extraneous:** Skip implementation bodies, show contracts only
- **Maximize Germane:** Build mental model of what the system does

**What AI Provides:**
- Function/method signatures with natural language descriptions
- Interface definitions and their purposes
- Data model overview (struct definitions, primary entities)
- Key workflows identified (e.g., "User authentication flow")
- API endpoint catalog with descriptions
- Integration points with external systems

**Use Cases:**
- Understanding what a codebase does overall
- Preparing to contribute to a project
- Architectural review
- Documentation generation

**UI/UX:**
- Collapsible function list with AI descriptions
- Interface viewer showing contracts
- Workflow diagrams (mermaid.js)
- Entity relationship diagrams
- Click to expand for implementation (transitions to Detailed mode)

---

**3. Scan Mode**

**Purpose:** Find specific information quickly (targeted search)

**Cognitive Load Strategy:**
- **Minimize Intrinsic:** Direct path to target information
- **Reduce Extraneous:** Filter out irrelevant code
- **Maximize Germane:** Understand context around the finding

**What AI Provides:**
- Semantic search (not just keyword matching)
  - "Where is authentication validated?" → Finds relevant functions even if they don't say "validate"
- Variable/function usage tracking
- Error source identification
- Pattern matching ("Find all database queries")
- Related code discovery ("Show me all callers of this function")
- Context-aware suggestions

**Use Cases:**
- Debugging: "Where does this error come from?"
- Understanding data flow: "Where is this variable modified?"
- Security audit: "Find all SQL queries"
- Refactoring prep: "What calls this deprecated function?"

**UI/UX:**
- Search bar with natural language support
- Results with surrounding context (3 lines before/after)
- Jump-to-definition
- Highlight matches in code
- Related results panel
- Filters: by layer, by bounded context, by file type

---

**4. Detailed Mode**

**Purpose:** Deep understanding of specific algorithms and logic

**Cognitive Load Strategy:**
- **Accept High Intrinsic:** This is unavoidably complex
- **Reduce Extraneous:** Provide maximum context to aid understanding
- **Maximize Germane:** Explain step-by-step, build complete mental model

**What AI Provides:**
- Line-by-line explanation of selected code block
- Variable state tracking ("At this point, `user` is...")
- Control flow analysis (if/else branches, loops)
- Algorithm explanation ("This implements binary search...")
- Complexity analysis (time/space complexity if applicable)
- Edge cases identified
- Potential bugs or issues
- Related documentation (links to Go docs, Stack Overflow, etc.)

**Use Cases:**
- Understanding a complex algorithm
- Debugging subtle logic errors
- Learning from well-written code
- Preparing to modify intricate logic
- Code review of critical path

**UI/UX:**
- Split view: code on left, AI explanation on right
- Synchronized scrolling
- Click any line for detailed explanation
- Variable hover shows current state/type
- Execution flow visualization
- Step-through simulation for logic
- Annotation mode (add notes)

---

**5. Critical Mode**

**Purpose:** Evaluate code quality and identify improvements (Human-in-the-Loop review mode)

**Cognitive Load Strategy:**
- **Accept High Intrinsic:** Evaluative thinking is demanding
- **Reduce Extraneous:** Focus on actionable feedback
- **Maximize Germane:** Teach patterns and anti-patterns

**What AI Provides:**
- **Architecture Review:**
  - Bounded context violations
  - Layer mixing (controller logic in data layer, etc.)
  - Missing abstractions
  - Tight coupling issues

- **Code Quality:**
  - Go idiom violations
  - Error handling issues
  - Scope problems (unnecessary global variables)
  - Naming convention violations
  - Missing comments/documentation

- **Security Concerns:**
  - SQL injection risks
  - Unvalidated input
  - Secrets in code
  - Authentication/authorization gaps

- **Performance:**
  - N+1 query problems
  - Unnecessary allocations
  - Missing indexes
  - Inefficient algorithms

- **Testing:**
  - Untested code paths
  - Missing error case tests
  - Test coverage gaps

- **Improvement Suggestions:**
  - Specific refactoring recommendations
  - Before/after code examples
  - Priority ranking (critical/important/nice-to-have)
  - Estimated effort

**Use Cases:**
- **Pre-merge PR review** (human-in-the-loop)
- **Reviewing AI-generated code** before production
- Learning from mistakes (educational)
- Architectural refactoring planning
- Security audit preparation

**UI/UX:**
- Issue list (categorized by severity)
- Click issue to jump to code location
- Issue explanation with context
- Suggested fix (diff view)
- Accept/reject/modify suggestions
- Add to refactoring backlog
- Generate PR comment for team review
- Track issue history (which issues keep appearing?)

---

#### Reading Mode Selection

**The Platform Helps Users Choose:**

When uploading code, AI suggests starting mode based on:
- **First time seeing this code?** → Start with Preview
- **Need to understand overall purpose?** → Start with Skim
- **Looking for something specific?** → Start with Scan
- **Trying to understand complex logic?** → Start with Detailed
- **Reviewing for quality/security?** → Start with Critical

**Fluid Transitions:**
- Click "Go Deeper" in Preview → transitions to Skim
- Click function in Skim → transitions to Detailed for that function
- Click "Find Usages" in Detailed → transitions to Scan
- Click "Review This" in any mode → transitions to Critical

---

#### Technical Implementation

**Database Schema (reviews.* schema):**

```sql
CREATE TABLE reviews.sessions (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES portal.users(id),
    title VARCHAR(255),
    code_source TEXT, -- 'github', 'paste', 'upload'
    github_repo VARCHAR(255),  -- if github
    github_branch VARCHAR(100), -- if github
    pasted_code TEXT,           -- if paste
    created_at TIMESTAMP DEFAULT NOW(),
    last_accessed TIMESTAMP DEFAULT NOW()
);

CREATE TABLE reviews.reading_sessions (
    id SERIAL PRIMARY KEY,
    session_id INT REFERENCES reviews.sessions(id),
    reading_mode VARCHAR(20) CHECK (reading_mode IN ('preview', 'skim', 'scan', 'detailed', 'critical')),
    target_path VARCHAR(500),  -- file or function being analyzed
    ai_response JSONB,          -- AI analysis results
    user_annotations TEXT,      -- user notes
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE reviews.critical_issues (
    id SERIAL PRIMARY KEY,
    reading_session_id INT REFERENCES reviews.reading_sessions(id),
    issue_type VARCHAR(50),     -- 'architecture', 'security', 'performance', 'quality'
    severity VARCHAR(20),       -- 'critical', 'important', 'minor'
    file_path VARCHAR(500),
    line_number INT,
    description TEXT,
    suggested_fix TEXT,
    status VARCHAR(20) DEFAULT 'open', -- 'open', 'accepted', 'rejected', 'fixed'
    created_at TIMESTAMP DEFAULT NOW()
);
```

**API Endpoints:**

```
POST   /api/review/sessions              - Create new review session
GET    /api/review/sessions              - List user's sessions
GET    /api/review/sessions/:id          - Get session details
DELETE /api/review/sessions/:id          - Delete session

POST   /api/review/sessions/:id/analyze  - Run AI analysis
  Body: {
    "reading_mode": "preview|skim|scan|detailed|critical",
    "target_path": "/path/to/file.go",    // optional for preview
    "scan_query": "find authentication",  // for scan mode
    "options": {}
  }

GET    /api/review/sessions/:id/results/:mode  - Get cached results for mode
POST   /api/review/sessions/:id/annotate       - Add user annotations
GET    /api/review/sessions/:id/issues         - Get all critical issues
PATCH  /api/review/issues/:id                  - Update issue status

WS     /ws/review/sessions/:id/collaborate     - Real-time collaboration
```

**AI Integration:**

```go
// services/review_ai_service.go

type ReviewAIService struct {
    aiClient AIProvider
    model    string // From env: AI_MODEL (e.g., "claude-3-5-sonnet-20241022")
}

func (s *ReviewAIService) AnalyzeInMode(
    ctx context.Context,
    code string,
    mode ReadingMode,
    options AnalysisOptions,
) (*AnalysisResult, error) {

    prompt := s.buildPromptForMode(mode, code, options)

    response, err := s.aiClient.Generate(ctx, &AIRequest{
        Model:  s.model,
        Prompt: prompt,
        Options: map[string]interface{}{
            "temperature": s.getTemperatureForMode(mode),
        },
    })

    return s.parseResponse(response, mode)
}

func (s *ReviewAIService) buildPromptForMode(mode ReadingMode, code string, opts AnalysisOptions) string {
    switch mode {
    case ModePreview:
        return fmt.Sprintf(`Analyze this codebase at a high level. Provide:
1. File structure overview
2. Identified bounded contexts
3. Technology stack
4. Architectural patterns
5. Entry points

Code:
%s

Format response as JSON with keys: file_structure, bounded_contexts, tech_stack, architecture_pattern, entry_points`, code)

    case ModeSkim:
        return fmt.Sprintf(`Analyze this code's abstractions. Provide:
1. All function signatures with brief descriptions
2. Interface definitions and purposes
3. Key data structures
4. Major workflows
5. API endpoints

Focus on WHAT, not HOW.

Code:
%s

Format response as JSON.`, code)

    case ModeScan:
        return fmt.Sprintf(`Search this code for: "%s"

Provide:
1. All relevant matches with context
2. Line numbers
3. Surrounding code (3 lines before/after)
4. Explanation of why each match is relevant

Code:
%s

Format response as JSON array of matches.`, opts.ScanQuery, code)

    case ModeDetailed:
        return fmt.Sprintf(`Provide detailed line-by-line analysis of this code:

%s

For each significant line, explain:
1. What it does
2. Why it's needed
3. Variable states
4. Control flow
5. Edge cases

Format response as JSON array indexed by line number.`, code)

    case ModeCritical:
        return fmt.Sprintf(`Review this code critically. Identify issues in:

1. Architecture (bounded context violations, layer mixing, missing abstractions)
2. Code Quality (idiom violations, error handling, scope issues, naming)
3. Security (injection risks, unvalidated input, exposed secrets)
4. Performance (N+1 queries, inefficient algorithms)
5. Testing (missing tests, uncovered paths)

For each issue provide:
- Severity (critical/important/minor)
- Location (file:line)
- Description
- Suggested fix with code example
- Rationale

Code:
%s

Format response as JSON array of issues.`, code)
    }

    return ""
}
```

---

**Dependencies:**
- PostgreSQL (reviews schema)
- Claude API or user-provided AI models (default: `deepseek-coder:6.7b`, or Claude API fallback)
- Logging service (for telemetry and AI performance tracking)
- Database caching (AI responses expensive to regenerate)

**Integration with Other Services:**
- **Logging:** All AI calls logged for performance analysis
- **Analytics:** Usage patterns (which modes used most, success metrics)
- **Build:** Can trigger review of code before merge
- **Portal:** Authentication, session management

### Logging Service
**Purpose:** Real-time log tracking and centralized logging

**Responsibilities:**
- Log ingestion from all services
- Real-time streaming via WebSocket
- Tag-based filtering
- Log storage and retrieval
- AI-driven context analysis (optional)
- **System health check monitoring** (integrated)

**Dependencies:**
- PostgreSQL (logs schema)
- Redis (WebSocket pub/sub)
- AI API (optional, for log analysis)

**API Endpoints:**
- `POST /api/logs` - Ingest log entry
- `GET /api/logs` - Query logs (with filters)
- `GET /api/logs/stats` - Log statistics
- `WS /ws/logs` - Real-time log streaming
- `GET /api/logs/healthcheck` - System-wide health diagnostics (JSON)
- `GET /healthcheck` - Health check dashboard (UI)

**Health Check Integration:**

The Logs service includes an integrated health check system (`internal/healthcheck/`) that validates:
- Docker container status for all services
- HTTP health endpoints for each service
- Database connectivity and responsiveness
- Gateway routing and availability

Available as both a standalone CLI tool (`cmd/healthcheck/`) and integrated into the Logs service API and dashboard.

**Phase 3: Health Intelligence (NEW)**

Extended with intelligent monitoring and auto-repair capabilities:

**Core Components:**
1. **Historical Trend Analysis** (`internal/logs/services/health_storage_service.go`)
   - 30-day retention of health check results
   - Response time trending and analysis
   - Per-service performance metrics
   - SQL-based querying for historical data

2. **Intelligent Auto-Repair** (`internal/logs/services/auto_repair_service.go`)
   - Issue classification (timeout, crash, dependency, security)
   - Adaptive repair strategies:
     - Timeout → `restart` (quick recovery)
     - Crash → `rebuild` (fresh image)
     - Dependency → `none` (can't repair this service)
     - Security CRITICAL → `rebuild` (patch needed)
   - Outcome tracking and logging

3. **Security Scanning** (`internal/healthcheck/trivy.go`)
   - Trivy integration for container image scanning
   - Vulnerability count by severity (CRITICAL/HIGH/MEDIUM/LOW)
   - Status determination based on findings
   - Scheduled scanning every 5 minutes

4. **Custom Health Policies** (`internal/logs/services/health_policy_service.go`)
   - Per-service configuration
   - Max response time thresholds
   - Repair strategy selection
   - Alert behavior settings
   - Default policies for all services

5. **Scheduled Monitoring** (`internal/logs/services/health_scheduler.go`)
   - Background health checks every 5 minutes
   - Includes Phase 1, Phase 2, and Phase 3 checks
   - Automatic repair triggering based on policies
   - Thread-safe concurrent execution

**New REST API Endpoints:**
```
GET  /api/health/history?limit=50        # Recent health checks
GET  /api/health/trends/:service?hours=24  # Service trend data
GET  /api/health/policies                 # All service policies
GET  /api/health/policies/:service        # Single service policy
PUT  /api/health/policies/:service        # Update policy configuration
GET  /api/health/repairs?limit=50         # Repair action history
POST /api/health/repair/:service          # Manual repair trigger
```

**Dashboard Enhancements:**
- **Historical Trends Tab** - 7-day performance charts, statistics, per-service analysis
- **Security Scans Tab** - Trivy results, vulnerability heatmap, detailed listing
- **Policies Tab** - Editable per-service policies with live updates

**Database Schema (logs schema):**
```sql
-- New tables for Phase 3
health_checks              -- Full health reports with retention
health_check_details       -- Individual checker results
security_scans             -- Trivy scan results
auto_repairs               -- Repair action history
health_policies            -- Per-service configuration
```

**Key Architectural Decisions:**

1. **Integration into Logs Service (NOT Separate)**
   - Single source of truth for observability
   - Reuses existing database, auth, and UI
   - Cross-correlation between health events and application logs
   - No duplicate infrastructure

2. **Intelligent Repair Strategy**
   - Not just "restart" - analyzes issue type first
   - Policy-based configuration per service
   - Timeout vs. crash vs. security requires different fixes
   - Dependency failures skip repair (dependencies must be fixed first)

3. **Trivy Integration (NOT Custom Implementation)**
   - Wraps existing `scripts/trivy-scan.sh`
   - Leverages 20K+ starred open-source tool
   - Maintains by Trivy team, not DevSmith
   - Parses JSON output, counts by severity

**Configuration (Environment Variables):**
```bash
HEALTH_CHECK_INTERVAL=5m              # Scheduler interval
HEALTH_AUTO_REPAIR_ENABLED=true       # Global toggle
HEALTH_RETENTION_DAYS=30              # Data retention
TRIVY_PATH=scripts/trivy-scan.sh      # Trivy binary/script
```

**Default Repair Policies:**
```go
"portal":    {MaxResponseTime: 500ms, AutoRepair: true, Strategy: "restart"}
"review":    {MaxResponseTime: 1000ms, AutoRepair: true, Strategy: "restart"}
"logs":      {MaxResponseTime: 500ms, AutoRepair: false, Strategy: "none"}
"analytics": {MaxResponseTime: 2000ms, AutoRepair: true, Strategy: "restart"}
```

**Future Enhancements (Phase 4+):**
- WebSocket real-time health updates
- Alert integrations (email, Slack)
- Performance regression detection
- Custom health check plugins
- ML-based anomaly detection
- Multi-environment support

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
**Purpose:** Terminal interface and collaborative coding

**Responsibilities:**
- Terminal emulation
- Cloud CLI support
- Copilot CLI integration
- Real-time collaboration
- Session recording and playback

**Dependencies:**
- PostgreSQL (build sessions schema)
- Logging service (terminal output capture)
- AI API (optional, for code assistance)

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

### Cross-service logging configuration

The platform uses a centralized Logs service reachable via the environment variable `LOGS_SERVICE_URL`.

- Default values:
  - In Docker: `http://logs:8082/api/logs`
  - Local development: `http://localhost:8082/api/logs`

- Per-service overrides: a service may set a per-service environment variable to override the default location. Example:
  - `REVIEW_LOGS_URL` will take precedence for the Review service
  - `PORTAL_LOGS_URL` will take precedence for the Portal service

- Startup policy (`LOGS_STRICT`):
  - `true` (default): startup validates `LOGS_SERVICE_URL` (or per-service override) and fails fast on invalid configuration.
  - `false`: startup logs a warning and proceeds with logging disabled (best-effort instrumentation will no-op).

Instrumented services should use the platform helper `internal/logging.NewClient(endpoint)` and the config helpers `internal/config.LoadLogsConfigFor(service)` or `LoadLogsConfigWithFallbackFor(service)` to resolve the effective endpoint and honor `LOGS_STRICT`.

Usage example (pseudo):
```
url, enabled, err := config.LoadLogsConfigWithFallbackFor("review")
if enabled {
    client := logging.NewClient(url)
    instrumentation := instrumentation.New(client)
} else {
    instrumentation := instrumentation.NewNoop()
}
```

Documented precedence:
1. Per-service override: `<SERVICE>_LOGS_URL` (uppercase service name)
2. `LOGS_SERVICE_URL`
3. Default based on `ENVIRONMENT` (`docker` vs local)


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

1. **GitHub Issues workflow): Autonomous implementation
2. **Claude via API** (10-15% of work): Architecture and strategic review
3. **Cursor/Copilot** (5-10% of work): IDE assistance for manual coding
4. **Mike** (Always): Project orchestration and final approval

**See:** `DevSmithRoles.md` for detailed roles and workflow.

### Branch Strategy

**Main Branches:**
- **main:** Production releases only (tagged versions)
- **development:** Integration branch (all PRs merge here first)

**Feature Branches:**
```
feature/{issue-number}-{short-description}
```

**Examples:**
- `feature/001-project-scaffolding`
- `feature/002-portal-authentication`
- `feature/003-review-preview-mode`
- `feature/015-critical-reading-mode`

**Why This Format:**
- ✅ Issue number provides traceability to `.docs/issues/` spec
- ✅ Short description makes purpose immediately clear
- ✅ Agents (Git workflow automation) know exactly what branch to create
- ✅ Easy to identify what work is in progress
- ✅ Merge commits reference specific implementation specs

**Other Branch Types:**
- **fix/{issue-number}-{description}:** Bug fixes (e.g., `fix/042-session-timeout`)
- **break-fix/*:** Experimental debugging (NOT merged to development)
- **claude-recovery-YYYYMMDD:** Auto-recovery branches (7-day retention)

**Branch Lifecycle:**
1. Create from `development`: `git checkout -b feature/XXX-description`
2. Work in isolation (commits, tests, implementation)
3. Push to origin: `git push origin feature/XXX-description`
4. Create PR to `development`
5. Review → Merge → Delete branch

**Branch Protection:**
- `main`: Requires PR from `development`, all checks must pass
- `development`: Requires PR from feature branch, 1 approval minimum

### Commit Standards
**Format:** Conventional Commits
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `