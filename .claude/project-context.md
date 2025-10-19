# DevSmith Modular Platform - Project Context

**Quick Reference for Claude.ai Sessions**

---

## Repository

**URL:** https://github.com/mikejsmith1985/devsmith-modular-platform

**Main Branch:** `development` (integration branch, PRs merge here)
**Production Branch:** `main` (releases only)

---

## Project Mission

**Core Goal:** Build a platform that teaches developers how to effectively read and understand code - the critical "Human-in-the-Loop" skill for supervising AI-generated code.

**Key Insight:** As AI generates more code (10x+ increase since 2024), the primary developer responsibility shifts from *writing* code to *reading, understanding, and validating* AI output.

**Centerpiece:** Review app with 5 reading modes (Preview, Skim, Scan, Detailed, Critical)

---

## Current Phase

**Status:** Foundation (Documentation Complete, Implementation Starting)

**Active Work:**
- Issue #1: Project scaffolding (Copilot + Mike, 30-60 min)
- Issue #2: Portal authentication (OpenHands, 1.5-2 hr) - Next

**Completion:**
- [x] ARCHITECTURE.md (3,035 lines) - Mental models, service specs, workflow
- [x] Requirements.md (815 lines) - Feature requirements, tech stack
- [x] DevSmithRoles.md (498 lines) - Hybrid AI team workflow
- [x] DevsmithTDD.md (1,122 lines) - Test strategy with mental models
- [x] Issue specs created (#1 and #2)
- [x] Crash recovery system implemented
- [ ] Project scaffolding (Issue #1)
- [ ] Portal authentication (Issue #2)

---

## Key Documentation

### Primary References
1. **ARCHITECTURE.md** - Single source of truth
   - Mental models (bounded contexts, layering, abstractions, scope)
   - Technology stack rationale
   - Service architecture (5 services)
   - Development workflow (6 phases)
   - Branch naming strategy

2. **Requirements.md** - Feature specifications
   - Executive summary
   - Five reading modes (detailed specs)
   - Service requirements
   - Success metrics

3. **DevSmithRoles.md** - Team workflow
   - Hybrid AI team (OpenHands 70-80%, Claude 10-15%, Copilot 5-10%)
   - Role responsibilities
   - Reading modes per team member

4. **DevsmithTDD.md** - Test strategy
   - Mental models as test organization
   - Go test patterns
   - Critical mode tests (HITL centerpiece)

### Implementation Specs
- `.docs/issues/001-copilot-project-scaffolding.md` (407 lines)
- `.docs/issues/002-openhands-portal-authentication.md` (1,280 lines)
- `.docs/specs/TEMPLATE.md` - OpenHands spec template (653 lines)

---

## Technology Stack

**Backend:** Go 1.21+
**Templates:** Templ (type-safe Go templates)
**Frontend:** HTMX + TailwindCSS + DaisyUI
**Database:** PostgreSQL 15+ (schema isolation per service)
**Gateway:** Nginx reverse proxy
**AI:** Ollama with `deepseek-coder-v2:16b` (local, no API costs)
**Auth:** GitHub OAuth (only auth provider for MVP)

**Why Go + Templ + HTMX:**
- No V8 crashes (previous Node.js platform failed)
- Explicit error handling (Go idioms)
- Server-side rendering (simpler, faster)
- Lower cognitive load

---

## Mental Models (Core Architecture Principles)

### 1. Bounded Contexts (Horizontal Separation)
**Concept:** Same entity means different things in different domains

**Example:**
- `User` in Portal context = GitHub authentication identity
- `User` in Review context = Code reviewer with review history
- NO cross-context leakage (Portal doesn't know about reviews)

### 2. Layering (Vertical Separation)
**Three Layers:**
- **Controller Layer:** HTTP handlers, Templ templates (request/response)
- **Orchestration Layer:** Business logic, services (workflows)
- **Data Layer:** Repositories, SQL queries (persistence)

**Rules:**
- ✅ Controllers call Services
- ✅ Services call Repositories
- ❌ Controllers MUST NOT call Repositories directly
- ❌ No circular dependencies

### 3. Abstractions vs Implementation
**Pattern:** Interfaces first, concrete implementations second

**Example:**
```go
// Abstraction (interface)
type AuthService interface {
    AuthenticateWithGitHub(ctx context.Context, code string) (*User, string, error)
}

// Implementation (struct)
type AuthServiceImpl struct {
    repo UserRepository
}
```

**Benefits:** Testing with mocks, future flexibility

### 4. Scope and Context
**Principle:** Minimize variable visibility, avoid global state

**Pattern:**
- Package-level: Only read-only config/constants
- Struct-level: Dependencies via constructor
- Function-level: Keep variables local, pass context explicitly

---

## Five Reading Modes

### 1. Preview Mode
**Purpose:** Quick structure assessment
**Output:** File tree, bounded contexts, tech stack, entry points
**Use Case:** "Is this code interesting? Should I look deeper?"

### 2. Skim Mode
**Purpose:** Understand high-level abstractions
**Output:** Interfaces, service boundaries, key types
**Use Case:** "What are the main components?"

### 3. Scan Mode
**Purpose:** Targeted search for specific elements
**Output:** Variable locations, function calls, imports
**Use Case:** "Where is X used?"

### 4. Detailed Mode
**Purpose:** Line-by-line comprehension
**Output:** Step-by-step explanation of logic
**Use Case:** "How does this algorithm work?"

### 5. Critical Mode (MOST IMPORTANT)
**Purpose:** Quality evaluation and issue identification
**Output:** Architecture issues, security flaws, performance problems
**Use Case:** "How can I make this better? What's broken?"

**This is the HITL centerpiece** - teaches developers to supervise AI output

---

## Branch Strategy

**Format:** `feature/{issue-number}-{short-description}`

**Examples:**
- `feature/001-project-scaffolding`
- `feature/002-portal-authentication`
- `feature/003-review-preview-mode`

**Other Types:**
- `fix/{issue-number}-{description}` - Bug fixes
- `break-fix/*` - Experimental debugging (not merged)
- `claude-recovery-YYYYMMDD` - Auto-recovery (7-day retention)

**Lifecycle:**
1. Create from `development`
2. Implement feature in isolation
3. Push and create PR
4. Review (Claude strategic, Mike acceptance)
5. Merge to `development`
6. Delete feature branch

---

## Development Workflow (6 Phases)

### 1. Issue Spec Creation
**Who:** Claude (OpenHands specs) or Copilot (simple tasks)

**Outputs:**
- `.docs/issues/{XXX}-openhands-{feature}.md` (800-1500 lines, complete implementation spec)
- `.docs/issues/{XXX}-copilot-{task}.md` (50-200 lines, task description)

### 2. Implementation
**Copilot Tasks:** Mike drives, Copilot assists (30-60 min)
**OpenHands Features:** Fully autonomous (1.5-3 hr, unattended)

### 3. Strategic Review (Claude)
**Duration:** <30 min
**Checklist:** Mental models (bounded context, layering, abstractions, scope)

### 4. Acceptance Review (Mike)
**Reading Modes:** Preview (2 min) → Skim (5 min) → Critical (10 min)
**Gate:** 100% acceptance criteria met (non-negotiable)

### 5. Merge (Mike)
**Command:** `git merge --no-ff feature/{XXX}-{description}`
**Template:** Includes acceptance criteria checklist, reviewers, test coverage

### 6. Release (When Ready)
**Flow:** `development` → `main` with version tag

---

## Hybrid AI Team

**Mike (Project Orchestrator):**
- Triggers work
- Final acceptance review
- Merges to development

**OpenHands + Ollama (70-80%):**
- Autonomous feature implementation
- 1.5-3 hour sessions (unattended, crash-proof)
- Reads complete specs (800-1500 lines)
- Implements TDD (tests first)

**Claude via API (10-15%):**
- Architecture decisions
- Creates OpenHands specs
- Strategic PR reviews (Critical mode)
- <30 min sessions (avoid crashes)

**GitHub Copilot (5-10%):**
- IDE autocomplete
- Scaffolding tasks
- Simple utilities

---

## Service Architecture (5 Services)

### 1. Portal Service
**Bounded Context:** Authentication and app management
**Responsibilities:** GitHub OAuth, session management, app browser
**Schema:** `portal` (users, sessions tables)
**Port:** 8001

### 2. Review Service (Platform Centerpiece)
**Bounded Context:** Code reading and analysis
**Responsibilities:** 5 reading modes, AI integration, code import
**Schema:** `reviews` (sessions, reading_sessions, code_files tables)
**Port:** 8002

### 3. Logging Service
**Bounded Context:** Application telemetry
**Responsibilities:** Real-time log streaming, AI-driven analysis
**Schema:** `logs` (log_entries, log_analysis tables)
**Port:** 8003

### 4. Analytics Service
**Bounded Context:** Data analysis and insights
**Responsibilities:** Log aggregation, anomaly detection, reports
**Schema:** `analytics` (metrics, anomalies, reports tables)
**Port:** 8004

### 5. Build Service
**Bounded Context:** Development workflow
**Responsibilities:** Terminal sessions, OpenHands integration, CLI operations
**Schema:** `builds` (terminal_sessions, commands tables)
**Port:** 8005

**All services accessible through Nginx gateway on port 3000**

---

## System Requirements

**Verified Configuration (Dell G16 7630):**
- CPU: i9-13900HX (24 cores) - Excellent
- RAM: 32GB - Good (7GB headroom with current workload)
- GPU: RTX 4070 8GB - Perfect for 16B models
- Storage: 512GB+ SSD - Sufficient

**Ollama Model:** `deepseek-coder-v2:16b` (~10GB RAM)

---

## Critical Constraints

**MVP Constraints:**
- GitHub OAuth only (no other auth providers)
- English only (no i18n)
- PostgreSQL only (no other databases)
- Ollama required (local AI, no cloud fallback for MVP)
- Desktop/laptop only (no mobile optimization)

**Non-Negotiable:**
- 70%+ unit test coverage
- 90%+ critical path coverage
- TDD workflow (Red → Green → Refactor)
- Mental models respected in all code
- No hardcoded values (config in .env)

---

## Common Review Scenarios

### Reviewing OpenHands PR:

**Critical Reading Mode Checklist:**

**Bounded Context:**
- [ ] No cross-context leakage (e.g., Portal doesn't know about Reviews)
- [ ] Entities defined within correct context
- [ ] Schema isolation maintained (no cross-schema FKs)

**Layering:**
- [ ] No handler → repository direct calls
- [ ] Services call repositories, not handlers
- [ ] No circular dependencies between layers

**Abstractions:**
- [ ] Interfaces defined before implementations
- [ ] Dependency injection used (constructor parameters)
- [ ] Tests use mocks, not concrete types

**Scope:**
- [ ] No global mutable state
- [ ] Variables kept as local as possible
- [ ] Dependencies passed explicitly via constructors

**Code Quality:**
- [ ] Error handling with context: `fmt.Errorf("...%w", err)`
- [ ] Tests achieve 70%+ coverage (90%+ for critical paths)
- [ ] Follows Go idioms (explicit errors, no exceptions)
- [ ] Clear naming (no abbreviations)

**Security:**
- [ ] No SQL concatenation (parameterized queries only: `$1`, `$2`)
- [ ] No secrets in code
- [ ] Input validation present (Gin binding tags)

---

## Quick Commands Reference

### Git Workflow
```bash
# Start new feature
git checkout development
git pull origin development
git checkout -b feature/{XXX}-{description}

# Commit with spec reference
git commit -m "feat(scope): description

Implements .docs/issues/{XXX}-{agent}-{feature}.md"

# Create PR
gh pr create --title "feat: description" --body "Implements .docs/issues/{XXX}"
```

### Testing
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Development
```bash
# One-time setup
make setup

# Start development environment
make dev

# Run tests
make test

# Build all services
make build
```

---

## Recovery Hooks (Crash Protection)

**Location:** `.claude/hooks/`

**System:**
- `session-logger.sh` - Logs all actions to daily markdown
- `git-recovery.sh` - Auto-commits to recovery branches
- `recovery-helper.sh` - Interactive recovery tool
- `user-prompt-submit.sh` - Captures user context

**Recovery Branches:** `claude-recovery-YYYYMMDD` (7-day retention)

**Usage:**
```bash
# Check recovery status
.claude/hooks/recovery-helper.sh status

# Restore from crash
.claude/hooks/recovery-helper.sh restore
```

---

## Next Steps for Claude.ai Sessions

**When You Start a New Conversation:**

1. Share this context file
2. Provide specific task or question
3. Reference relevant documentation section

**Example:**
```
I'm working on this project: [paste project-context.md]

Please review this PR using Critical Reading Mode:
https://github.com/mikejsmith1985/devsmith-modular-platform/pull/2

Focus on bounded context and layering principles from ARCHITECTURE.md.
```

---

## Project Philosophy

**For Mike (ADHD-Friendly):**
- All specs in markdown (no UI context switching)
- Clear branch naming shows what's in progress
- Quick reviews using 5 reading modes
- Minimal process overhead
- Focus on momentum, not perfection

**For Platform:**
- Teach effective code reading (not just code writing)
- Human-in-the-Loop as core skill
- Cognitive load management throughout
- Mental models as foundation
- AI as teaching tool, human as learner

**For Development:**
- Autonomous agents do implementation (80%)
- Humans do architecture and review (20%)
- Crash-proof workflow (recovery hooks)
- Test-driven (70%+ coverage minimum)
- Documentation reflects reality, not aspirations

---

**Last Updated:** 2025-10-19
**Document Version:** 1.0
