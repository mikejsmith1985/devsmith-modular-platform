# DevSmith Modular Platform: Roles and Responsibilities

## Project Overview
The DevSmith Modular Platform, hosted at [github.com/mikejsmith1985/devsmith-modular-platform](https://github.com/mikejsmith1985/devsmith-modular-platform), is a modular, AI-driven platform for learning, debugging, and building code. This document defines the roles of the **hybrid AI development team**, adhering to the DevSmith Coding Standards and Test-Driven Development (TDD) principles. The goal is to maintain a clean, recoverable repo with high-quality code and robust testing.

## System Architecture
- **Tech Stack**: Go 1.21+ with Templ templates, HTMX for interactivity, TailwindCSS + DaisyUI
- **Database**: PostgreSQL 15+ with pgx driver, schema isolation per app
- **Infrastructure**: Docker + Docker Compose, Nginx gateway, GitHub Actions CI/CD
- **AI Integration**: Ollama (local), OpenHands (autonomous coding), Claude (architecture), GitHub Copilot (IDE assist)

## Roles and Responsibilities

### 1. Project Orchestrator and Manager (Mike)
- **Role**: Oversees the project, manages the AI agent team, and ensures alignment with project goals.
- **Responsibilities**:
  - Define and prioritize features based on `Requirements.md`.
  - Create GitHub issues with clear, single-feature tasks and acceptance criteria using issue templates.
  - **Trigger OpenHands** with implementation specs from Claude.
  - Review and approve pull requests (PRs) after Claude's architectural review.
  - Merge approved PRs into the `development` branch and manage releases to `main`.
  - Monitor project progress and ensure adherence to TDD and coding standards.
  - Validate backups (logs, code states, model configurations) to ensure recoverability.
  - Coordinate sprints (e.g., Sprint 1: Portal + Logging) and track milestones.
  - Configure GitHub branch protection rules to enforce tests, approvals, and changelog updates.
  - **Manage Ollama models** (install, update, configure).
- **Tools**:
  - GitHub for issue tracking, PR approvals, and repo management.
  - GitHub Projects for sprint planning.
  - OpenHands CLI for triggering autonomous tasks.
  - Ollama for local LLM management.

### 2. Primary Architect and Strategic Reviewer (Claude via API)
- **Role**: Designs high-level architecture, reviews PRs strategically, and solves complex problems.

- **Reading Mode**: **Critical Mode** (evaluative review)
  - Operates in "Critical Reading" mode during PR reviews
  - Identifies architectural issues, security concerns, quality problems
  - Provides actionable improvement suggestions
  - See: ARCHITECTURE.md - Mental Models - Application to Review App

- **Responsibilities**:
  - Design the modular architecture, ensuring apps (logging, analytics, review, build) are isolated yet interoperable.
  - Define database schemas (PostgreSQL with schema isolation) and API contracts.
  - **Create detailed implementation specs** for OpenHands to execute autonomously.
    - Specs follow template in `.docs/specs/TEMPLATE.md`
    - Explicitly state bounded contexts, layering, and abstractions
    - Optimize for cognitive load management

  - **Review OpenHands-generated PRs using mental models:**
    - ✅ **Bounded Context:** No cross-context leakage (e.g., Portal User vs Review User)
    - ✅ **Layering:** Controllers don't call repositories directly, clear layer separation
    - ✅ **Abstractions:** Interfaces used appropriately, implementations follow contracts
    - ✅ **Scope:** Variables kept local, minimal global state
    - ✅ **Coding Standards:** File organization, naming, error handling
    - ✅ **TDD:** Test coverage (70%+), critical paths tested
    - ✅ **Security:** No SQL injection, input validation, no exposed secrets
    - ✅ **Performance:** No N+1 queries, efficient algorithms

  - Provide detailed feedback on PRs, suggesting improvements or refactoring.
  - Validate AI-driven features (e.g., Ollama integration, Review app's 5 reading modes).
  - Ensure WebSocket implementation for real-time logging is robust.
  - Root cause analysis of complex bugs.
  - Recommend optimizations for the one-click installation process.

- **Tools**:
  - Claude Code CLI (this interface).
  - GitHub for PR reviews and comments.
  - Go + Templ + HTMX for architecture decisions.
  - Mental models: Bounded Context, Layering, Abstractions, Scope

- **Limitations**:
  - Cannot execute code directly (relies on OpenHands for implementation).
  - Subject to V8 crashes (mitigated by recovery hooks in `.claude/hooks/`).
  - Sessions should be kept short (< 30 minutes) to reduce crash risk.

### 3. Primary Implementation Agent (OpenHands + Ollama)
- **Role**: **Autonomous code generator and implementer** - executes 70-80% of development work.
- **Responsibilities**:
  - **Implement features autonomously** based on specs from Claude.
  - Follow DevSmith Coding Standards:
    - **File Organization** (Go service structure):
      - `apps/{service}/{main.go, handlers/, models/, templates/, static/, services/, db/, middleware/, utils/, config/, tests/}`.
      - Templates: Templ files (`.templ`) for server-side rendering.
      - Static assets: CSS (TailwindCSS), minimal JS (HTMX, Alpine.js), images.
    - **Naming Conventions** (Go conventions):
      - Files: `snake_case.go` for source files, `snake_case.templ` for templates, `*_test.go` for tests.
      - Code: `camelCase` for unexported, `PascalCase` for exported, `UPPER_SNAKE` for constants.
      - Acronyms: Keep uppercase (`HTTPServer`, `JSONData`, `URLPath`).
    - **Go Handler Pattern**:
      ```go
      func HandleFeature(c *gin.Context) {
        var req FeatureRequest
        if err := c.ShouldBindJSON(&req); err != nil {
          c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
          return
        }
        result, err := services.ProcessFeature(c.Request.Context(), req)
        if err != nil {
          log.Error().Err(err).Msg("Feature processing failed")
          c.JSON(http.StatusInternalServerError, gin.H{"error": "Processing failed"})
          return
        }
        c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
      }
      ```
    - **Templ Template Pattern**:
      ```go
      templ FeaturePage(data FeatureData) {
        @Layout("Feature") {
          <div class="container mx-auto p-4">
            <h1 class="text-2xl font-bold">{data.Title}</h1>
            if len(data.Items) == 0 {
              <p>No items found</p>
            } else {
              for _, item := range data.Items {
                @ItemCard(item)
              }
            }
          </div>
        }
      }
      ```
    - **Error Handling**: Provide user-friendly messages, explicit error checking (`if err != nil`), structured logging.
  - Write TDD-compliant tests (unit, integration) before coding, targeting 70%+ unit test coverage and 90%+ critical path coverage.
  - Run tests locally (`go test ./...`) before committing.
  - **Autonomous workflow** (no human intervention needed):
    - Create feature branch from `development`.
    - Write tests first (TDD).
    - Implement feature following specs.
    - Run tests, fix failures iteratively.
    - Commit with Conventional Commit messages (e.g., `feat(portal): add GitHub OAuth login`).
    - Update `AI_CHANGELOG.md`.
    - Push code and create PR to `development`.
  - Perform manual testing (browser-based via Playwright/Cypress integration):
    - Feature works through nginx gateway (`http://localhost:3000`).
    - No console errors/warnings.
    - Regression check for related features.
    - Light/dark mode compatibility (if applicable).
    - Responsive design for mobile/tablet (if applicable).
- **Tools**:
  - OpenHands CLI (autonomous agent framework).
  - Ollama with models: `deepseek-coder-v2:16b` or `codellama:34b`.
  - Go toolchain: `go test`, `go build`, `air` (hot reload).
  - Git for version control.
  - Browser automation for testing.
- **Strengths**:
  - **Fully autonomous** - can work overnight on complex tasks.
  - **Local execution** - no API costs, no rate limits.
  - **Persistent state** - checkpoint/resume on crash or reboot.
  - **Direct execution** - bash, file editing, git operations.
- **Limitations**:
  - Smaller context window than Claude (but sufficient for single features).
  - May need guidance on complex architectural decisions (defers to Claude).
  - Quality depends on Ollama model (use 16B+ models for best results).

### 4. IDE Coding Assistant (GitHub Copilot)
- **Role**: Real-time autocomplete and quick generation during manual coding.
- **When to use**:
  - Quick code snippets while Mike is manually coding.
  - Real-time autocomplete in VS Code.
  - Small refactorings.
  - Boilerplate generation.
- **Responsibilities**:
  - Provide intelligent autocomplete suggestions.
  - Generate function/struct boilerplate.
  - Assist with quick edits.
- **Tools**:
  - GitHub Copilot extension in VS Code.
- **Strengths**:
  - Instant feedback in IDE.
  - Great for boilerplate.
  - Works well with Go, Templ, HTMX.
- **Limitations**:
  - No autonomous workflow.
  - Limited context (single file).
  - Requires manual intervention.

## Hybrid Workflow

### Standard Feature Development (80% of work)

1. **Issue Creation** (Mike):
   - Create GitHub issue with clear acceptance criteria using issue templates.
   - Label with appropriate tags (`feature`, `app:portal`, etc.).

2. **Architecture & Spec Creation** (Claude):
   - Mike triggers Claude session (<30 minutes to avoid crash risk).
   - Claude designs high-level architecture.
   - Claude creates **detailed implementation spec** with:
     - File structure (which Go files, templates, handlers).
     - Function signatures and interfaces.
     - Database schema changes (if any).
     - Test requirements (unit, integration).
     - Acceptance criteria from issue.
   - Spec saved to issue comment or `.docs/specs/` directory.
   - **At end of session:** Claude creates/updates devlog entry in `.docs/devlog/YYYY-MM-DD.md` with:
     - Problems discovered
     - Decisions made
     - Solutions implemented
     - Action items for next agent

3. **Autonomous Implementation** (OpenHands + Ollama):
   - Mike triggers OpenHands: `openhands --task "Implement feature from spec in issue #42"`.
   - **Before starting:** OpenHands reads:
     - Latest devlog entry (`.docs/devlog/YYYY-MM-DD.md`) for context and action items
     - Issue spec
     - Architecture docs
   - OpenHands works **fully autonomously**:
     - Creates feature branch from `development`.
     - Writes tests first (TDD per `DevsmithTDD.md`).
     - Implements feature following Claude's spec.
     - Runs tests (`go test ./...`), fixes failures iteratively.
     - Performs browser testing via Playwright integration.
     - Commits with Conventional Commit messages.
     - Updates `AI_CHANGELOG.md`.
     - Pushes code and creates PR to `development`.
   - **After completion:** Updates devlog with:
     - Implementation notes
     - Issues encountered and solutions
     - Test results
   - **Duration**: 30 minutes - 2 hours (runs unattended).
   - **Crash-proof**: OpenHands checkpoint/resume if interrupted.

4. **PR Creation** (Automatic via GitHub Actions):
   - **When**: Automatically triggered when code is pushed to a `feature/**` branch
   - **How**: `.github/workflows/auto-create-pr.yml` workflow runs automatically
   - **What it does**:
     - Detects the feature branch (e.g., `feature/003-copilot-portal-auth`)
     - Extracts issue number from branch name (e.g., `003`)
     - Finds corresponding issue file (`.docs/issues/003-*.md`)
     - Extracts PR title from issue file (first line)
     - Extracts PR description from issue template
     - Creates PR automatically with:
       - Base: `development`
       - Title: From issue spec
       - Body: From issue spec (includes "Closes #N", summary, testing checklist)
   - **Benefits**:
     - No manual PR creation needed
     - PR details always match issue spec
     - Consistent PR format across all agents
     - Reduces human error
   - **When skipped**: If PR already exists for the branch (prevents duplicates)
   - GitHub Actions run automated checks after PR creation:
     - Go tests and coverage
     - Linting (golangci-lint)
     - Docker build verification
     - Security scan (Trivy)

5. **Strategic Review** (Claude):
   - Mike triggers Claude for PR review (<30 minutes).
   - **Before starting:** Claude reads latest devlog for context
   - Claude reviews for:
     - Architectural integrity (modularity, scalability).
     - Adherence to DevSmith Coding Standards.
     - TDD compliance (test coverage, quality).
     - Security and error handling.
   - Claude comments on PR with detailed feedback.
   - **After review:** Updates devlog with review findings and recommendations

6. **Acceptance Review** (Mike):
   - Verifies acceptance criteria from issue are 100% met.
   - Reviews Claude's feedback.
   - Approves or requests changes.

7. **Merge and Release** (Mike):
   - Merge PR to `development` (squash merge).
   - Delete feature branch.
   - Issue automatically closed.
   - When ready for release: merge `development` to `main` with version tag.

### Complex Problems / Architecture Changes (20% of work)

For complex issues that require deep architectural thinking:

1. **Problem Analysis** (Claude):
   - Mike describes problem in detail.
   - Claude performs root cause analysis.
   - Claude proposes multiple solutions with trade-offs.

2. **Decision & Spec** (Mike + Claude):
   - Mike selects approach.
   - Claude creates detailed spec (as above).

3. **Implementation** (OpenHands):
   - Same autonomous workflow as standard features.

4. **Review** (Claude + Mike):
   - Extra scrutiny on architectural changes.
   - May require multiple review rounds.

### Manual Coding (5-10% of work)

When Mike codes manually:

1. **IDE Assistance** (Copilot):
   - Real-time autocomplete in VS Code.
   - Quick boilerplate generation.

2. **Testing** (Manual):
   - Run `go test ./...`.
   - Manual browser testing.

3. **PR** (Manual):
   - Create PR, same review process as OpenHands PRs.

### Work Distribution

| Type | Agent | % of Work | Crash Risk |
|------|-------|-----------|------------|
| Architecture & Specs | Claude | 10-15% | Low (short sessions) |
| Implementation | OpenHands | 70-80% | **None** (checkpoint/resume) |
| PR Reviews | Claude | 5-10% | Low (short sessions) |
| Manual Coding | Mike + Copilot | 5-10% | None |

### Crash Recovery

If Claude crashes during spec creation:

1. Check `.claude/todos.json` to see what was in progress.
2. Check `.claude/recovery-logs/session-YYYYMMDD.md` for recent actions.
3. Run `.claude/hooks/recovery-helper.sh restore` to recover work from git.
4. Resume from where Claude left off.

**Key Insight**: 80% of work (implementation) is now crash-proof because OpenHands runs as a separate process with persistent state.

## Testing Requirements

### Automated Testing (OpenHands)

- **Unit tests** for utilities and services (70%+ coverage).
- **Handler tests** for HTTP endpoints.
- **Integration tests** for critical paths (e.g., login → portal → app launch).
- **Template tests** for Templ components (compile-time validation).

**Commands**:
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./apps/portal/handlers/...

# Run integration tests
go test -tags=integration ./tests/integration/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Manual Testing Checklist (OpenHands + Mike)

- [ ] Feature works in browser through nginx gateway (`http://localhost:3000`).
- [ ] No JavaScript errors in browser console.
- [ ] Regression check for related features.
- [ ] Light/dark mode compatibility (via DaisyUI themes).
- [ ] Responsive design for mobile/tablet (TailwindCSS breakpoints).
- [ ] HTMX interactions work correctly (partial updates, form submissions).
- [ ] WebSocket connections stable (for real-time features).
- [ ] Hot reload works with Air (Go file changes trigger rebuild).

### CI Pipeline (GitHub Actions)

**Automated Checks**:
- Go tests and coverage (`go test -cover ./...`).
- Linting (`golangci-lint run`).
- Security scan (Trivy for Docker images).
- Docker build verification (multi-stage builds).
- PR format validation (title, branch name, acceptance criteria).

**Branch Protection**:
- All checks must pass before merge.
- One approval required (from Mike).
- Enforce conventional commit messages.

## Workflow Improvements
- **Automated PR Checks**: GitHub Actions run tests, linting, and coverage checks on PRs to catch issues early.
- **Branch Protection**: Require passing tests, one approval (Orchestrator), and updated changelogs for `development` and `main`.
- **Conventional Commits**: Use `feat:`, `fix:`, `docs:`, etc., for clear commit history and automated changelog generation.
- **Pre-Commit Hooks**: Use Husky (frontend) and pre-commit (backend) to run linting and tests locally before commits.
- **Issue Templates**: Standardize feature/bug reports with acceptance criteria to ensure Copilot focuses on single features.
- **Backup Tests**: Add automated tests for backup system to verify recoverability.
- **Sprints**: Organize development into sprints (e.g., Sprint 1: Portal + Logging) with milestones for tracking.

---

## Code Reading Modes for Team Members

The platform recognizes five distinct modes of reading code, each appropriate for different situations. Understanding when to use each mode is critical for effective AI supervision.

### For Mike (Project Orchestrator)

**Primary Modes: Preview, Skim, and Critical**

#### Preview Mode
**When:** First time reviewing OpenHands output or exploring a new service
- Quick scan of file structure
- Understand what was implemented at high level
- Decide if deeper review is needed

**How:**
- Look at file tree in GitHub PR
- Read OpenHands' PR description
- Check which bounded contexts and layers were touched

#### Skim Mode
**When:** Understanding OpenHands implementation before critical review
- See what functions/interfaces were created
- Understand data models added
- Get high-level flow of the feature

**How:**
- Read function signatures (don't dive into implementations yet)
- Check struct definitions
- Review API endpoint contracts
- Look at database schema changes

#### Critical Mode
**When:** Final acceptance review before merging
- Verify acceptance criteria from issue are 100% met
- Spot check implementations for quality
- Ensure no obvious security or performance issues

**How:**
- Use Claude's review as primary filter
- Focus on business logic correctness
- Verify tests exist and make sense
- Check that feature actually solves the issue

**Mike's Review Checklist:**
```markdown
- [ ] Acceptance criteria from issue #[number] are 100% met
- [ ] Claude's review has been addressed (no unresolved critical issues)
- [ ] Tests exist and pass
- [ ] Feature works when tested manually
- [ ] No obvious security issues (secrets, SQL injection, etc.)
- [ ] Documentation updated if needed
```

### For Claude (Architect & Reviewer)

**Primary Modes: Skim and Critical**

#### Skim Mode
**When:** Creating implementation specs for OpenHands
- Need to understand existing patterns in codebase
- Want to see how similar features were implemented
- Building mental model before designing new feature

#### Critical Mode
**When:** Reviewing OpenHands PRs (90% of Claude's work)
- Full architectural review
- Security analysis
- Performance evaluation
- Code quality assessment

**Claude's Review Process:**
1. **Preview context**: What service? What bounded context?
2. **Skim abstractions**: What interfaces/contracts were created?
3. **Critical review**: Apply mental models checklist
4. **Provide feedback**: Specific, actionable improvements

### For OpenHands (Implementation Agent)

**Primary Modes: Skim, Scan, and Detailed**

#### Skim Mode
**When:** Starting implementation from Claude's spec
- Understanding existing code patterns
- Finding similar implementations to follow
- Building mental model of codebase

#### Scan Mode
**When:** Looking for specific information during implementation
- "Where do other services call the logging API?"
- "How are database connections initialized?"
- "What's the pattern for error handling in handlers?"

#### Detailed Mode
**When:** Understanding complex existing logic to extend
- Need to integrate with non-trivial algorithm
- Extending complex business logic
- Understanding subtle edge cases

**OpenHands should prefer Skim over Detailed** - implementation specs from Claude should minimize need for detailed reading.

### Cognitive Load Strategy by Mode

| Mode | Intrinsic Load | Extraneous Load | Germane Load | Best For |
|------|---------------|-----------------|--------------|----------|
| **Preview** | Minimal | Reduce | Build map | Quick assessment |
| **Skim** | Low | Reduce | Build framework | Understanding abstractions |
| **Scan** | Target | Minimize | Context only | Finding specific info |
| **Detailed** | High | Provide context | Complete model | Algorithm understanding |
| **Critical** | High | Focus | Patterns/anti-patterns | Quality evaluation |

---

## Platform Implementation Note

These five reading modes will be **directly implemented** in the Review Service application. Users will be able to:

1. Upload code (GitHub repo, paste, upload)
2. Select reading mode based on their goal
3. Receive AI-guided analysis appropriate for that mode
4. Transition between modes fluidly

**This platform teaches users how to read code effectively**, which is the critical skill for supervising AI-generated code (Human in the Loop).

See: `ARCHITECTURE.md` - Section "Service Architecture → Review Service" for complete implementation specification.

---

## Notes
- OpenHands must focus on one feature per issue to avoid scope creep and maintain repo clarity.
- Claude's architectural reviews ensure modularity and scalability, reducing technical debt.
- Mike's oversight and approval process ensures alignment with project goals and acceptance criteria.
- All team members apply appropriate reading modes for their role.
- The Review app is the **centerpiece** of the platform - it teaches the fundamentals of code reading.
- All workflows adhere to TDD principles per `DevsmithTDD.md` and DevSmith Coding Standards.
- Mental models (bounded context, layering, abstractions, scope) are foundational to all development.
