# DevSmith Modular Platform: Roles and Responsibilities

## Project Overview
The DevSmith Modular Platform, hosted at [github.com/mikejsmith1985/devsmith-modular-platform](https://github.com/mikejsmith1985/devsmith-modular-platform), is a modular, AI-driven platform for learning, debugging, and building code. This document defines the roles of the **hybrid AI development team**, adhering to the DevSmith Coding Standards and Test-Driven Development (TDD) principles. The goal is to maintain a clean, recoverable repo with high-quality code and robust testing.

## System Architecture
- **Tech Stack**: React 18 + TypeScript frontend with Vite, Go 1.21+ microservices backend, devsmith-theme.css styling
- **Database**: PostgreSQL 15+ with pgx driver, schema isolation per app
- **Infrastructure**: Docker + Docker Compose, Traefik v2.10+ gateway (port 3000), GitHub Actions CI/CD
- **AI Integration**: Claude Code for architecture/planning, Cursor/Copilot for implementation

## Roles and Responsibilities

### 1. Project Orchestrator and Supervisor (Mike)
- **Role**: Oversees the project, supervises AI development work, and ensures alignment with project goals.
- **Responsibilities**:
  - Define and prioritize features based on `Requirements.md`.
  - Create GitHub issues with clear, single-feature tasks and acceptance criteria using issue templates.
  - **Trigger Claude Haiku for documentation cleanup and workflow creation and planning sessions.
  - **Implement features** using Cursor/Copilot based on Claude's specs.
  - Review and test code as it's built.
  - Create pull requests and manage code reviews.
  - Merge approved PRs into the `development` branch and manage releases to `main`.
  - Monitor project progress and ensure adherence to TDD and coding standards.
  - Coordinate sprints (e.g., Sprint 1: Portal + Logging) and track milestones.
  - Configure GitHub branch protection rules to enforce tests, approvals, and changelog updates.
- **Tools**:
  - GitHub for issue tracking, PR management, and repo operations.
  - GitHub Projects for sprint planning.
  - Claude Haiku for documentation cleanup and workflow creation sessions.
  - Cursor/Copilot in VS Code for implementation.
  - Git for version control.

### 2. Primary Architect and Planner (Claude Code)
- **Role**: Designs high-level architecture, creates implementation plans, and provides strategic guidance.

- **Reading Mode**: **Critical Mode** (evaluative review)
  - Operates in "Critical Reading" mode during planning and reviews
  - Identifies architectural issues, security concerns, quality problems
  - Provides actionable improvement suggestions
  - See: ARCHITECTURE.md - Mental Models - Application to Review App

- **Responsibilities**:
  - Design the modular architecture, ensuring apps (logging, analytics, review, build) are isolated yet interoperable.
  - Define database schemas (PostgreSQL with schema isolation) and API contracts.
  - **Create detailed implementation plans** for Mike to execute with Copilot.
    - Plans include: file structure, function signatures, interfaces, test requirements
    - Explicitly state bounded contexts, layering, and abstractions
    - Optimize for cognitive load management
    - Provide code patterns and examples

  - **Review code using mental models:**
    - ✅ **Bounded Context:** No cross-context leakage (e.g., Portal User vs Review User)
    - ✅ **Layering:** Controllers don't call repositories directly, clear layer separation
    - ✅ **Abstractions:** Interfaces used appropriately, implementations follow contracts
    - ✅ **Scope:** Variables kept local, minimal global state
    - ✅ **Coding Standards:** File organization, naming, error handling
    - ✅ **TDD:** Test coverage (70%+), critical paths tested
    - ✅ **Security:** No SQL injection, input validation, no exposed secrets
    - ✅ **Performance:** No N+1 queries, efficient algorithms

  - Provide detailed architectural guidance during implementation.
  - Validate feature designs before implementation begins.
  - Ensure WebSocket implementation for real-time logging is robust.
  - Root cause analysis of complex bugs.
  - Recommend optimizations and refactorings.

- **Tools**:
  - Claude Code CLI (this interface)
  - Direct file read/write/edit capabilities
  - Bash for running tests and builds
  - Mental models: Bounded Context, Layering, Abstractions, Scope

- **Strengths**:
  - Can read, write, and edit files directly
  - Can run tests and validate changes
  - Large context window for understanding complex systems
  - Strong architectural reasoning capabilities

- **Crash Recovery**:
  - V8 crash recovery hooks in `.claude/hooks/` for automatic recovery
  - Todo list persistence (`.claude/todos.json`) tracks progress across crashes
  - Recovery branches (`claude-recovery-YYYYMMDD`) with auto-commits
  - Session logs (`.claude/recovery-logs/`) for resuming work

### 3. Primary Implementation Assistant (Cursor/Copilot)
- **Role**: **AI-powered code completion and generation** - assists Mike with 70-80% of development work.
- **Responsibilities**:
  - **Generate code** based on specs from Claude.
  - Follow DevSmith Coding Standards:
    - **File Organization** (Go service structure):
      - `apps/{service}/{main.go, handlers/, models/, services/, db/, middleware/, utils/, config/, tests/}`.
      - Frontend: Separate React app in `frontend/` directory with package.json.
      - Static assets: devsmith-theme.css, Bootstrap Icons.
    - **Naming Conventions** (Go conventions):
      - Files: `snake_case.go` for source files, `*_test.go` for tests.
      - React: `PascalCase.jsx` for components, `camelCase.js` for utilities.
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
    - **React Component Pattern**:
      ```jsx
      // FeaturePage.jsx
      import React from 'react';
      import Layout from '../components/Layout';
      import ItemCard from '../components/ItemCard';

      function FeaturePage({ data }) {
        return (
          <Layout title="Feature">
            <div className="container mx-auto p-4">
              <h1 className="text-2xl font-bold">{data.title}</h1>
              {data.items.length === 0 ? (
                <p>No items found</p>
              ) : (
                data.items.map(item => <ItemCard key={item.id} item={item} />)
              )}
            </div>
          </Layout>
        );
      }
      export default FeaturePage;
      ```
    - **Error Handling**: Provide user-friendly messages, explicit error checking (`if err != nil`), structured logging.
  - Suggest TDD-compliant test implementations targeting 70%+ unit test coverage and 90%+ critical path coverage.
  - Assist Mike with running tests locally (`go test ./...`) before committing.
  - **Supervised workflow** (Mike drives, Copilot assists):
    - Mike creates feature branch from `development`.
    - Copilot suggests test implementations (TDD).
    - Copilot generates feature code following Claude's specs.
    - Mike reviews suggestions, accepts/modifies as needed.
    - Mike runs tests, Copilot helps fix failures.
    - Mike commits with Conventional Commit messages (e.g., `feat(portal): add GitHub OAuth login`).
    - Mike updates `AI_CHANGELOG.md`.
    - Mike pushes code and creates PR to `development`.
  - Assist with manual testing reminders:
    - Feature works through nginx gateway (`http://localhost:3000`).
    - No console errors/warnings.
    - Regression check for related features.
    - Light/dark mode compatibility (if applicable).
    - Responsive design for mobile/tablet (if applicable).
- **Tools**:
  - Cursor/Copilot extension in VS Code.
  - Cursor/Copilot Chat for explanations and refactoring.
  - Go toolchain: `go test`, `go build`, `air` (hot reload).
- **Strengths**:
  - **Real-time assistance** - instant suggestions as Mike types.
  - **Context-aware** - understands current file and surrounding code.
  - **Multi-language** - excellent with Go, React/JSX, TypeScript, SQL.
  - **Chat interface** - can explain code, suggest improvements, write tests.
  - **Fast** - no latency, no rate limits.
- **Limitations**:
  - Requires Mike's supervision and decision-making.
  - Limited to file-level context (doesn't see entire codebase).
  - May suggest code that doesn't follow architectural patterns without guidance.

## Hybrid Workflow

### Standard Feature Development (80% of work)

1. **Issue Creation** (Mike):
   - Create GitHub issue with clear acceptance criteria using issue templates.
   - Label with appropriate tags (`feature`, `app:portal`, etc.).

2. **Architecture & Planning** (Claude Code):
   - Mike triggers Claude Code session.
   - Claude designs high-level architecture.
   - Claude creates **detailed implementation plan** with:
     - File structure (which Go files, React components, handlers).
     - Function signatures and interfaces.
     - Database schema changes (if any).
     - Test requirements (unit, integration).
     - Code examples and patterns to follow.
     - Acceptance criteria from issue.
   - Plan saved to issue comment or `.docs/specs/` directory.
   - **Note**: Devlog updated POST-MERGE (step 7), not during planning session

3. **Supervised Implementation** (Mike + Copilot):
   - Mike reads Claude's implementation plan.
   - Mike creates feature branch from `development`.
   - **Mike writes tests first** (TDD per `DevsmithTDD.md`), with Copilot assisting:
     - Copilot suggests test structure and assertions.
     - Mike reviews and accepts/modifies suggestions.
   - **Mike implements feature** following Claude's spec, with Copilot generating code:
     - Mike provides context (comments, function signatures).
     - Copilot generates implementation code.
     - Mike reviews each suggestion before accepting.
   - Mike runs tests (`go test ./...`), fixes failures with Copilot's help.
   - Mike performs manual browser testing.
   - Mike commits with Conventional Commit messages.
   - Mike updates `AI_CHANGELOG.md`.
   - Mike pushes code to remote.
   - **Duration**: Varies by feature complexity (Mike works at own pace).

4. **PR Creation** (Mike + Copilot):
   - Mike uses Copilot to assist with PR creation.
   - Copilot suggests PR title and description based on commits.
   - Mike runs `gh pr create` with Copilot-generated content:
     - PR title: "Issue #XXX: Title from issue"
     - PR description includes:
       - Link to issue: "Closes #N"
       - Summary of changes
       - Testing checklist completed
   - Base branch: `development`
   - GitHub Actions run automated checks:
     - Go tests and coverage
     - Linting (golangci-lint)
     - Docker build verification
     - Security scan (Trivy)

5. **Optional Review** (Claude Code):
   - For complex features or when Mike wants architectural validation.
   - Mike triggers Claude Code for PR review.
   - Claude reviews for:
     - Architectural integrity (modularity, scalability).
     - Adherence to DevSmith Coding Standards.
     - TDD compliance (test coverage, quality).
     - Security and error handling.
     - Performance considerations.
   - Claude provides detailed feedback.
   - Mike addresses feedback if needed.

6. **Acceptance and Merge** (Mike):
   - Verifies acceptance criteria from issue are 100% met.
   - Ensures all automated checks pass.
   - Reviews Claude's feedback (if review was requested).
   - Merges PR to `development` (squash merge).
   - Deletes feature branch.
   - Issue automatically closed.
   - When ready for release: merge `development` to `main` with version tag.

7. **Post-Merge Documentation** (Mike + Copilot or Claude):
   - **After merge**, update devlog entry in `.docs/devlog/YYYY-MM-DD.md`:
     - What was implemented
     - Decisions made during implementation
     - Issues encountered and solutions
     - Any architectural insights
   - **Purpose**: Maintains project history and context for future sessions
   - **Who writes**: Mike with Copilot assistance, or Claude Code if session is active

### Complex Problems / Architecture Changes (20% of work)

For complex issues that require deep architectural thinking:

1. **Problem Analysis** (Claude Code):
   - Mike describes problem in detail.
   - Claude performs root cause analysis.
   - Claude proposes multiple solutions with trade-offs.
   - Claude may prototype or validate approaches.

2. **Decision & Planning** (Mike + Claude):
   - Mike selects approach.
   - Claude creates detailed implementation plan (as above).
   - Extra detail on complex areas and potential pitfalls.

3. **Implementation** (Mike + Copilot):
   - Same supervised workflow as standard features.
   - Mike may consult Claude more frequently during implementation.
   - Extra testing and validation steps.

4. **Review** (Claude + Mike):
   - Extra scrutiny on architectural changes.
   - May require multiple review rounds.
   - Claude may suggest refactorings or improvements.

### Quick Fixes / Small Changes (5-10% of work)

When Mike makes small changes without architecture planning:

1. **IDE Assistance** (Copilot):
   - Real-time autocomplete in VS Code.
   - Quick boilerplate generation.
   - Bug fixes, typos, small refactorings.

2. **Testing** (Mike):
   - Run `go test ./...`.
   - Manual browser testing.

3. **PR** (Mike):
   - Create PR, same process as standard features.
   - May skip Claude review for trivial changes.

### Work Distribution

| Type | Agent | % of Work | Notes |
|------|-------|-----------|-------|
| Architecture & Planning | Claude Code | 15-20% | Deep design sessions |
| Implementation | Mike + Copilot | 70-80% | Supervised coding |
| Code Review (optional) | Claude Code | 5-10% | Complex features only |
| Quick Fixes | Mike + Copilot | 5-10% | Small changes |

**Key Insight**: Mike supervises all implementation work, maintaining quality and learning the codebase deeply. Claude provides architectural guidance and strategic reviews when needed.

## Testing Requirements

### Automated Testing (Mike + Copilot)

- **Unit tests** for utilities and services (70%+ coverage).
- **Handler tests** for HTTP endpoints.
- **Integration tests** for critical paths (e.g., login → portal → app launch).
- **Component tests** for React components (React Testing Library).

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

### Manual Testing Checklist (Mike)

- [ ] Feature works in browser through Traefik gateway (`http://localhost:3000`).
- [ ] No JavaScript errors in browser console.
- [ ] Regression check for related features.
- [ ] Light/dark mode compatibility (Alpine.js theme toggle).
- [ ] Responsive design for mobile/tablet (devsmith-theme.css breakpoints).
- [ ] React components render correctly (state management, props, hooks).
- [ ] WebSocket connections stable (for real-time features).
- [ ] Hot reload works with Vite HMR (frontend) and Air (backend).

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
**When:** First time reviewing implemented code or exploring a new service
- Quick scan of file structure
- Understand what was implemented at high level
- Decide if deeper review is needed

**How:**
- Look at file tree in GitHub PR
- Read PR description
- Check which bounded contexts and layers were touched

#### Skim Mode
**When:** Understanding implementation before critical review or testing
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
**When:** Creating implementation plans for Mike
- Need to understand existing patterns in codebase
- Want to see how similar features were implemented
- Building mental model before designing new feature

#### Critical Mode
**When:** Reviewing code or providing architectural guidance
- Full architectural review
- Security analysis
- Performance evaluation
- Code quality assessment

**Claude's Review Process:**
1. **Preview context**: What service? What bounded context?
2. **Skim abstractions**: What interfaces/contracts were created?
3. **Critical review**: Apply mental models checklist
4. **Provide feedback**: Specific, actionable improvements

### For Mike + Copilot (Implementation Team)

**Primary Modes: Skim, Scan, and Detailed**

#### Skim Mode
**When:** Starting implementation from Claude's plan
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

**Prefer Skim over Detailed** - Claude's implementation plans should minimize need for detailed reading.

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
- Each implementation focuses on one feature per issue to avoid scope creep and maintain repo clarity.
- Claude's architectural planning ensures modularity and scalability, reducing technical debt.
- Mike's supervision ensures quality, alignment with project goals, and deep codebase knowledge.
- All team members apply appropriate reading modes for their role.
- The Review app is the **centerpiece** of the platform - it teaches the fundamentals of code reading.
- All workflows adhere to TDD principles per `DevsmithTDD.md` and DevSmith Coding Standards.
- Mental models (bounded context, layering, abstractions, scope) are foundational to all development.
- This simplified workflow (Claude + Copilot + Mike) eliminates local LLM complexity while maintaining high quality.
