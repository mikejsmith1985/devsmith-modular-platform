# DevSmith Team Roles & Responsibilities

**Version:** 1.0
**Last Updated:** 2025-10-20

---

## Overview

DevSmith Platform uses a hybrid AI development team with **Mike as Product Manager and Quality Gate**. Each team member (human + AI agents) has specific responsibilities that maximize efficiency while ensuring quality.

---

## 1. Mike (Human - Product Manager & Quality Gate)

**Primary Role:** Project orchestration and final approval

**Responsibilities:**
- ✅ **Final PR approval** - Reviews and approves/rejects all PRs
- ✅ **Business priorities** - Decides what to build and when
- ✅ **Release timing** - Controls when features ship to production
- ✅ **Scope management** - Clarifies requirements and handles scope changes
- ✅ **Quality gate** - Ensures acceptance criteria are met before merge
- ✅ **Strategic oversight** - Guides platform direction and architecture

**NOT Responsible For:**
- ❌ Writing code (AI agents do this)
- ❌ Creating Git branches (automated by workflow)
- ❌ Switching branches (automated by Copilot)
- ❌ Creating PRs (automated by AI agents)
- ❌ Running tests (automated by CI/CD)
- ❌ Manual Git operations (automated where possible)

**Workflow Position:**
```
Issue Created → AI Implementation → PR Auto-Created → [MIKE REVIEWS] → Approve/Reject → Merge
                                                            ↑
                                                      QUALITY GATE
```

**Time Investment:**
- ~5-10 minutes per PR review
- ~30-60 minutes per planning session with Claude
- Focus on strategic decisions, not tactical implementation

**Success Metrics:**
- PRs reviewed within 24 hours
- <5% of PRs require changes after review
- Platform velocity maintained (consistent PR throughput)

---

## 2. Claude (AI - Strategic Architect & Code Reviewer)

**Primary Role:** Architecture design and code review

**Responsibilities:**
- ✅ **Architectural design** - System structure, service boundaries, data models
- ✅ **Requirement decomposition** - Breaks down features into implementable issues
- ✅ **Code review** - Reviews PRs for maintainability, correctness, standards compliance
- ✅ **Documentation** - Maintains ARCHITECTURE.md, Requirements.md, issue specs
- ✅ **Root cause analysis** - Diagnoses complex bugs and design flaws
- ✅ **Acceptance criteria** - Defines what "done" means for each issue
- ✅ **Strategic guidance** - Advises Mike on technical trade-offs

**Workflow Position:**
```
Planning Session → Create Issues → [AI IMPLEMENTS] → Review PRs → Recommend Approve/Reject
      ↑                                                    ↓
    MIKE                                              MIKE DECIDES
```

**Session Duration:** <30 minutes (focused, high-value interactions)

**Claude's Review Checklist:**
1. Code follows DevSmith standards (ARCHITECTURE.md)
2. Tests written first (TDD compliance)
3. All acceptance criteria met
4. No architectural violations
5. Documentation updated
6. Mental models preserved (bounded contexts, layering, etc.)

---

## 3. GitHub Copilot (AI - Primary Implementation Agent)

**Primary Role:** Fast, IDE-integrated code implementation

**Responsibilities:**
- ✅ **Feature implementation** - Writes production code for defined issues
- ✅ **Test-first development** - Follows TDD (RED-GREEN-REFACTOR)
- ✅ **Branch detection** - Auto-switches to correct feature branch
- ✅ **PR creation** - Auto-creates PR when work complete
- ✅ **Quick iterations** - Handles small-to-medium complexity tasks
- ✅ **IDE assistance** - Provides inline suggestions and completions

**Best For:**
- UI implementation (Templ templates, HTMX)
- API endpoints (Gin handlers)
- Database migrations
- Test writing
- Quick fixes

**Session Duration:** 30-60 minutes

**Workflow:**
1. Mike says: "Copilot, work on issue #006"
2. Copilot auto-switches to `feature/006-review-scan-mode`
3. Copilot reads issue spec from `.docs/issues/006-*.md`
4. Copilot writes tests FIRST (RED phase)
5. Copilot implements feature (GREEN phase)
6. Copilot creates PR with `gh pr create`
7. Mike reviews PR

---

## 4. OpenHands (AI - Autonomous Long-Running Implementation)

**Primary Role:** Complex, autonomous feature implementation

**Responsibilities:**
- ✅ **Complex features** - Multi-file, multi-service implementations
- ✅ **Autonomous execution** - Runs unattended for 30min-2hr
- ✅ **End-to-end features** - Complete user stories from DB to UI
- ✅ **Overnight work** - Can run while Mike sleeps (no API costs)
- ✅ **Recovery from errors** - Self-corrects build/test failures

**Best For:**
- Service foundations (database + handlers + services)
- Reading modes (complex prompt engineering)
- Integration work (cross-service features)
- E2E test scenarios

**Session Duration:** 30 minutes - 2 hours (autonomous)

**Workflow:**
1. Mike triggers: OpenHands on issue #009 (async)
2. OpenHands works autonomously (no Mike interaction needed)
3. OpenHands creates PR when done
4. Mike reviews PR next morning

**Key Advantage:** Mike can work with Claude/Copilot while OpenHands runs in background.

---

## 5. GitHub Actions (Automation - CI/CD & Workflow)

**Primary Role:** Automated quality checks and workflow orchestration

**Responsibilities:**
- ✅ **Branch auto-creation** - Creates next feature branch after PR merge
- ✅ **Pre-merge validation** - Runs tests, build checks, linting
- ✅ **PR format validation** - Ensures acceptance criteria checkboxes present
- ✅ **Acceptance gate** - Blocks merge if criteria unchecked
- ✅ **Test coverage reporting** - Validates 70%+ coverage
- ✅ **Build artifact creation** - Generates binaries for release

**Workflows:**
1. **auto-sync-next-issue.yml** - Creates next feature branch automatically
2. **test-and-build.yml** - Validates all Go code on push
3. **validate-pr.yml** - Checks PR format and acceptance criteria
4. **coverage-check.yml** - Enforces test coverage minimums

**Prevents:**
- ❌ Merging code that doesn't build
- ❌ Merging PRs with failing tests
- ❌ Merging incomplete features (unchecked acceptance criteria)
- ❌ Merging code below coverage threshold

---

## Team Interaction Patterns

### Pattern 1: New Feature Planning (Mike + Claude)
```
1. Mike: "I want users to export analytics to CSV"
2. Claude: Analyzes requirements, references ARCHITECTURE.md
3. Claude: Creates issue #017 with acceptance criteria
4. Mike: Approves scope
5. GitHub Actions: Auto-creates feature/017-analytics-csv-export branch
```

### Pattern 2: Quick Implementation (Mike + Copilot)
```
1. Mike: "Copilot, work on issue #012"
2. Copilot: Auto-switches to feature/012-portal-dashboard-ui
3. Copilot: Implements dashboard (30-45 min)
4. Copilot: Creates PR
5. Mike: Reviews and approves
```

### Pattern 3: Complex Implementation (Mike + OpenHands)
```
1. Mike: Triggers OpenHands on issue #009
2. OpenHands: Works autonomously (1-2 hours)
3. OpenHands: Creates PR
4. Mike: Reviews next morning
5. Claude: Performs deep code review
6. Mike: Approves after Claude's review
```

### Pattern 4: PR Review Flow
```
1. AI Agent: Creates PR (auto or manual trigger)
2. GitHub Actions: Validates build, tests, format
3. Claude: Reviews code quality and standards
4. Mike: Reviews acceptance criteria
5. Mike: Approves/Requests Changes
6. GitHub Actions: Auto-creates next branch on merge
```

---

## Decision Authority Matrix

| Decision Type | Authority | Can Delegate? |
|--------------|-----------|---------------|
| **What to build** | Mike | No |
| **When to build** | Mike | No |
| **How to build** | Claude + AI Agents | Yes |
| **Architecture** | Claude | Mike reviews |
| **Code quality** | Claude | Mike trusts |
| **PR approval** | Mike | No |
| **Scope changes** | Mike | No |
| **Technical trade-offs** | Claude recommends, Mike decides | Partial |
| **Release timing** | Mike | No |
| **Test strategy** | Claude | Yes |
| **Implementation details** | AI Agents | Yes |

---

## Communication Protocols

### Mike ↔ Claude
- **Medium:** Claude Code CLI
- **Session Length:** <30 minutes
- **Frequency:** As needed (planning, reviews)
- **Topics:** Architecture, requirements, strategy

### Mike ↔ Copilot
- **Medium:** GitHub Copilot Chat in VS Code
- **Session Length:** 30-60 minutes
- **Frequency:** Daily for active development
- **Topics:** Feature implementation, bug fixes

### Mike ↔ OpenHands
- **Medium:** OpenHands CLI (async)
- **Session Length:** Trigger once, check back later
- **Frequency:** For complex tasks
- **Topics:** Multi-file features, integrations

### Claude ↔ Copilot
- **Medium:** `.github/copilot-instructions.md`
- **Communication:** One-way (Claude writes instructions)
- **Purpose:** Guide Copilot on standards and workflow

### GitHub Actions ↔ All
- **Medium:** GitHub webhooks, status checks
- **Communication:** Automated (no human intervention)
- **Purpose:** Quality gates, workflow orchestration

---

## Role Boundaries (What NOT to Do)

### Mike Should NOT:
- ❌ Write code directly (exception: emergency fixes)
- ❌ Create Git branches manually (workflow automates this)
- ❌ Switch branches for AI agents (they auto-switch)
- ❌ Create PRs manually (agents do this)
- ❌ Micromanage implementation details

### Claude Should NOT:
- ❌ Make business decisions (Mike's role)
- ❌ Approve PRs (Mike's role)
- ❌ Write production code (Copilot/OpenHands role)
- ❌ Change requirements without Mike's approval

### Copilot Should NOT:
- ❌ Make architectural decisions (Claude's role)
- ❌ Work on development branch (always feature branches)
- ❌ Skip tests (TDD is mandatory)
- ❌ Merge PRs (Mike's role)

### OpenHands Should NOT:
- ❌ Make scope decisions (Mike's role)
- ❌ Skip acceptance criteria (defined in issue)
- ❌ Work on multiple issues simultaneously

### GitHub Actions Should NOT:
- ❌ Auto-merge PRs (Mike must approve)
- ❌ Auto-approve PRs (quality gate is Mike)
- ❌ Override test failures (must be fixed)

---

## Success Criteria for Team

**Mike's Success:**
- PRs reviewed quickly (<24 hours)
- Informed decisions based on Claude's recommendations
- Not becoming a bottleneck in workflow

**Claude's Success:**
- Architecture stays coherent and maintainable
- Issues are clear and implementable
- Code reviews catch problems early

**Copilot's Success:**
- Features implemented quickly (30-60 min)
- TDD followed (tests first, then code)
- PRs meet acceptance criteria on first submission

**OpenHands' Success:**
- Complex features completed autonomously
- Build/test failures self-corrected
- PRs ready for review after first run

**GitHub Actions' Success:**
- No broken code merged to development
- Branches auto-created reliably
- Quality gates prevent incomplete work

---

## Escalation Paths

**Technical Issue:**
```
Copilot/OpenHands encounters problem → Claude analyzes → Claude recommends solution → Mike decides if acceptable
```

**Scope Clarification:**
```
AI agent unsure of requirement → Ask Mike → Mike clarifies → Update issue spec
```

**Architecture Decision:**
```
New pattern needed → Claude designs → Claude presents to Mike → Mike approves/rejects → Document in ARCHITECTURE.md
```

**Build Failure:**
```
GitHub Actions fails → AI agent investigates → Fix and push → Rerun checks
```

**PR Rejection:**
```
Mike rejects PR → Explains reason → AI agent fixes → Submit new version → Mike re-reviews
```

---

## References

- **ARCHITECTURE.md** - Technical architecture and patterns
- **Requirements.md** - Platform requirements and specifications
- **DevsmithTDD.md** - Test-Driven Development workflow
- **.github/copilot-instructions.md** - Copilot workflow guide
- **.docs/WORKFLOW-GUIDE.md** - Git workflow and branching strategy

---

**Key Insight:** Mike is the **Quality Gate and Product Manager**, not a **Bottleneck**. Automation handles the mechanics, Mike handles the strategy and approval.
