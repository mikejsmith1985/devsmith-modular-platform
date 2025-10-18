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
- **Responsibilities**:
  - Design the modular architecture, ensuring apps (logging, analytics, review, build) are isolated yet interoperable.
  - Define database schemas (PostgreSQL with schema isolation) and API contracts.
  - **Create detailed implementation specs** for OpenHands to execute autonomously.
  - Review OpenHands-generated PRs for:
    - Adherence to DevSmith Coding Standards (file organization, naming, error handling).
    - Architectural integrity (modularity, scalability, performance).
    - Alignment with TDD principles (test coverage, test quality).
    - Security and debugging best practices (e.g., friendly error messages, logging).
  - Provide detailed feedback on PRs, suggesting improvements or refactoring.
  - Validate AI-driven features (e.g., Ollama integration, review app reading modes).
  - Ensure WebSocket implementation for real-time logging is robust.
  - Root cause analysis of complex bugs.
  - Recommend optimizations for the one-click installation process.
- **Tools**:
  - Claude Code CLI (this interface).
  - GitHub for PR reviews and comments.
  - Go + Templ + HTMX for architecture decisions.
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

3. **Autonomous Implementation** (OpenHands + Ollama):
   - Mike triggers OpenHands: `openhands --task "Implement feature from spec in issue #42"`.
   - OpenHands works **fully autonomously**:
     - Creates feature branch from `development`.
     - Writes tests first (TDD per `DevsmithTDD.md`).
     - Implements feature following Claude's spec.
     - Runs tests (`go test ./...`), fixes failures iteratively.
     - Performs browser testing via Playwright integration.
     - Commits with Conventional Commit messages.
     - Updates `AI_CHANGELOG.md`.
     - Pushes code and creates PR to `development`.
   - **Duration**: 30 minutes - 2 hours (runs unattended).
   - **Crash-proof**: OpenHands checkpoint/resume if interrupted.

4. **PR Creation** (OpenHands):
   - PR includes:
     - Link to issue (`Closes #42`).
     - Implementation summary.
     - Test results (automated + manual).
     - Screenshots (if UI changes).
     - Acceptance criteria checklist (all checked).
   - GitHub Actions run automated checks:
     - Go tests and coverage.
     - Linting (golangci-lint).
     - Docker build verification.
     - Security scan (Trivy).

5. **Strategic Review** (Claude):
   - Mike triggers Claude for PR review (<30 minutes).
   - Claude reviews for:
     - Architectural integrity (modularity, scalability).
     - Adherence to DevSmith Coding Standards.
     - TDD compliance (test coverage, quality).
     - Security and error handling.
   - Claude comments on PR with detailed feedback.

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

## Notes
- Copilot must focus on one feature per issue to avoid scope creep and maintain repo clarity.
- Claude’s architectural reviews ensure modularity and scalability, reducing technical debt.
- Orchestrator’s oversight and approval process ensures alignment with project goals and recoverability.
- All team members adhere to TDD principles per `TDD.md` and DevSmith Coding Standards.
