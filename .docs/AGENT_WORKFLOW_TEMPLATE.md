# Agent Implementation Workflow Template

**Purpose:** Step-by-step workflow for coding agents to maximize effectiveness and minimize rework

**Target Audience:** Claude Code, GitHub Copilot, OpenHands, and human developers

**Design Goals:**
- ✅ Reduce V8 crash risk by 60-70%
- ✅ Cut rework cycles by 50-60%
- ✅ Maintain or improve code quality
- ✅ Build confidence through incremental progress

---

## Quick Start

**Before starting ANY task:**

1. ✅ Read this template
2. ✅ Review relevant issue spec (`.docs/issues/###-*.md`)
3. ✅ Check latest devlog entry (`.docs/devlog/YYYY-MM-DD.md`)
4. ✅ Set 25-minute timer
5. ✅ Start workflow

---

## The Workflow (TL;DR)

```
1. Understand    → Read minimal context
2. Plan         → Break into micro-tasks
3. Implement    → Smallest testable unit
4. Validate     → Run pre-commit checklist
5. Commit       → Save progress
6. Repeat       → Next micro-task
```

**Session Duration:** 15-25 minutes maximum
**Commit Frequency:** Every 10-15 minutes
**Context Target:** <100K tokens

---

## Phase 0: Session Setup (2-3 minutes)

### 0.1 Check Current Branch
```bash
git branch --show-current
```

**Expected:** `feature/###-description`

**If on wrong branch:**
```bash
# Create/switch to correct branch
git checkout development
git pull origin development
git checkout -b feature/###-description
```

### 0.2 Understand the Task

**Read the issue spec:**
```bash
# List available issues
ls -la .docs/issues/

# Read YOUR issue
cat .docs/issues/###-your-issue.md
```

**Extract:**
- ✅ What needs to be done?
- ✅ Acceptance criteria
- ✅ Technical constraints
- ✅ Related files/components

**DON'T:**
- ❌ Read 10 files "for context"
- ❌ Try to understand entire codebase
- ❌ Over-research before doing

**Principle:** Just-in-time learning > Comprehensive research

### 0.3 Check Previous Progress

**Read latest devlog:**
```bash
# Find today's or most recent log
ls -lt .docs/devlog/ | head -5
cat .docs/devlog/2025-10-21.md
```

**Look for:**
- ✅ What was done previously?
- ✅ Any blockers or challenges?
- ✅ Decisions made?
- ✅ Next steps suggested?

### 0.4 Set Timer
```bash
# Set 25-minute checkpoint
# When timer goes off → COMMIT or PAUSE
```

**Why:** Forces incremental progress, reduces crash risk window

---

## Phase 1: Task Breakdown (3-5 minutes)

### 1.1 Identify Micro-Tasks

**Break the feature into smallest logical units:**

**❌ BAD (too large):**
- "Implement review service"

**✅ GOOD (bite-sized):**
1. Create review request struct
2. Create review response struct
3. Add validation for request
4. Implement ReviewHandler function
5. Write unit test for validation
6. Implement review processing logic
7. Write unit test for processing
8. Add error handling
9. Write integration test
10. Update documentation

### 1.2 Prioritize by Risk

**Order:**
1. **Low-hanging fruit first** (build confidence)
2. **Core logic** (highest value)
3. **Edge cases** (after core works)
4. **Polish** (nice-to-haves)

### 1.3 Estimate Token Budget

**Example breakdown:**
```
Task: Implement ReviewHandler

Context needed:
- Read ReviewRequest struct definition (500 tokens)
- Read existing handler example (1000 tokens)
- Read error handling pattern (300 tokens)
Total: ~1800 tokens

Implementation:
- Write handler (500 tokens)
- Write test (800 tokens)
- Tool results (1000 tokens)
Total: ~2300 tokens

Grand Total: ~4100 tokens
Buffer: ~1000 tokens

Estimated: 5K tokens (well within budget)
```

**Decision:**
- ✅ <10K tokens → Safe to proceed
- 🟡 10-20K tokens → Proceed with caution
- 🔴 >20K tokens → Break down further

### 1.4 Create Todo List (Optional but Recommended)

**Using TodoWrite tool (Claude Code):**
```json
{
  "todos": [
    {"content": "Create ReviewRequest struct", "status": "pending"},
    {"content": "Create ReviewResponse struct", "status": "pending"},
    {"content": "Add validation logic", "status": "pending"},
    {"content": "Write validation tests", "status": "pending"},
    {"content": "Implement handler", "status": "pending"},
    {"content": "Write handler tests", "status": "pending"}
  ]
}
```

---

## Phase 2: Implement First Unit (10-15 minutes)

### 2.1 Read Minimal Context

**Only read what you need for THIS task:**

**❌ BAD:**
```bash
# Read entire codebase "to understand context"
Read internal/review/service.go
Read internal/review/handler.go
Read internal/review/repository.go
Read internal/review/models.go
Read internal/portal/handler.go  # (unrelated!)
Read ARCHITECTURE.md  # (2243 lines!)
```

**✅ GOOD:**
```bash
# Read only what's needed for current micro-task
Read internal/review/models.go  # (to see existing structs)
# That's it. Start implementing.
```

**Principle:** Read < Implement < Validate

### 2.2 Implement Smallest Testable Unit

**Example Task:** "Create ReviewRequest struct"

**Implementation:**
```go
// File: internal/review/models.go

// ReviewRequest represents a code review request
type ReviewRequest struct {
    RepoURL    string   `json:"repo_url" validate:"required,url"`
    Branch     string   `json:"branch" validate:"required"`
    Mode       string   `json:"mode" validate:"required,oneof=preview skim scan detailed critical"`
    FileFilter []string `json:"file_filter,omitempty"`
}
```

**Time:** 5 minutes

**Tokens used:** ~1000 tokens

### 2.3 Write Test IMMEDIATELY

**Don't wait. Test now.**

```go
// File: internal/review/models_test.go

func TestReviewRequest_Validation(t *testing.T) {
    tests := []struct {
        name    string
        req     ReviewRequest
        wantErr bool
    }{
        {
            name: "valid request",
            req: ReviewRequest{
                RepoURL: "https://github.com/user/repo",
                Branch:  "main",
                Mode:    "preview",
            },
            wantErr: false,
        },
        {
            name: "missing repo url",
            req: ReviewRequest{
                Branch: "main",
                Mode:   "preview",
            },
            wantErr: true,
        },
        // ... more cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validate.Struct(tt.req)
            if (err != nil) != tt.wantErr {
                t.Errorf("validation error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Time:** 5-8 minutes

**Tokens used:** ~1500 tokens

**Total so far:** ~2500 tokens (safe!)

---

## Phase 3: Validate Before Commit (3-5 minutes)

### 3.1 Run Pre-Commit Checklist

**Reference:** `.docs/PRE_COMMIT_CHECKLIST.md`

#### Step 1: Build
```bash
go build ./internal/review/...
```

**Expected:** ✅ No errors

**If errors:**
- Fix immediately
- Don't proceed to next step

#### Step 2: Test
```bash
go test ./internal/review/...
```

**Expected:** ✅ All tests pass

**If failures:**
- Fix immediately
- Understand why test failed
- Don't skip or ignore

#### Step 3: Lint
```bash
golangci-lint run ./internal/review/...
```

**Expected:** ✅ No errors, <5 warnings

**If issues:**
- Fix real issues (unused vars, missing errors)
- Consider auto-fix: `golangci-lint run --fix`

#### Step 4: Self-Review
```bash
git diff
```

**Check:**
- ✅ Changes match intent
- ✅ No debug statements
- ✅ No commented code
- ✅ No accidental changes

### 3.2 Validation Summary

**All green?** → Proceed to commit
**Any red?** → Fix issues, re-run validation

**Time invested:** 3-5 minutes
**Rework prevented:** 20-40 minutes

---

## Phase 4: Commit the Unit (1-2 minutes)

### 4.1 Stage Changes
```bash
# Stage specific files (preferred)
git add internal/review/models.go
git add internal/review/models_test.go

# Check what's staged
git status
```

### 4.2 Write Clear Commit Message
```bash
git commit -m "$(cat <<'EOF'
feat(review): add ReviewRequest struct with validation

Implement ReviewRequest model with JSON tags and validation rules:
- Required: repo_url, branch, mode
- Optional: file_filter
- Mode validation: preview, skim, scan, detailed, critical

Tests cover:
- Valid requests
- Missing required fields
- Invalid mode values

Part of issue #011 (Review Service Foundation)

✅ Build passes
✅ Tests pass (100% coverage)
✅ Linting clean

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

### 4.3 Confirm Commit
```bash
git log -1 --stat
```

**Verify:**
- ✅ Commit message is clear
- ✅ Files changed are correct
- ✅ No unexpected changes

### 4.4 Update Todo (if using)
```json
{
  "todos": [
    {"content": "Create ReviewRequest struct", "status": "completed"},
    {"content": "Create ReviewResponse struct", "status": "in_progress"},
    ...
  ]
}
```

**Time for commit:** 1-2 minutes

---

## Phase 5: Checkpoint Decision (1 minute)

### 5.1 Check Session Status

**Questions to ask:**

#### Time Check
- ⏰ How long have I been working?
  - <15 min → Continue to next unit
  - 15-25 min → Continue but be cautious
  - >25 min → **STOP, PUSH, TAKE BREAK**

#### Token Check
- 📊 How much context have I accumulated?
  - <50K → Safe to continue
  - 50-100K → Continue with caution
  - >100K → **STOP, PUSH, RESTART SESSION**

#### Complexity Check
- 🧠 How complex is the next unit?
  - Simple → Continue
  - Moderate → Continue if time allows
  - Complex → **STOP, COMMIT WHAT YOU HAVE**

### 5.2 Decision Matrix

| Time | Tokens | Next Complexity | Decision |
|------|--------|----------------|----------|
| <15min | <50K | Simple | ✅ Continue |
| <15min | <50K | Complex | 🟡 Proceed with caution |
| 15-25min | 50-100K | Simple | 🟡 Proceed with caution |
| 15-25min | 50-100K | Complex | 🔴 Stop, push, break |
| >25min | Any | Any | 🔴 Stop, push, break |
| Any | >100K | Any | 🔴 Stop, push, restart |

### 5.3 If Stopping

#### Push Your Work
```bash
# First time
git push -u origin feature/###-description

# Subsequent pushes
git push
```

#### Create Session Summary

**Update devlog:**
```markdown
## Session: 2025-10-21 14:00-14:25

**Task:** Issue #011 - Review Service Foundation

**Completed:**
- ✅ Created ReviewRequest struct with validation
- ✅ Wrote comprehensive tests (100% coverage)
- ✅ Validated and committed

**Next Steps:**
1. Create ReviewResponse struct
2. Implement ReviewHandler function
3. Add error handling patterns

**Decisions Made:**
- Used struct tags for validation (consistent with Portal service)
- Chose enum validation for Mode field (prevents invalid values)

**Blockers:** None

**Context for Next Session:**
- Review models are in: internal/review/models.go
- Follow Portal handler pattern in: internal/portal/handler.go
```

**Time:** 2-3 minutes

### 5.4 If Continuing

**Proceed to next micro-task:**
1. Mark current task complete in todo
2. Mark next task as "in_progress"
3. Read minimal context for next task
4. Implement → Validate → Commit
5. Repeat

---

## Phase 6: Iteration (Repeat 2-5)

### 6.1 Next Micro-Task

**Example:** "Create ReviewResponse struct"

**Context needed:**
```bash
# Already have models.go in context (from previous task)
# Just add to it
```

**Implementation:**
```go
// ReviewResponse represents the result of a code review
type ReviewResponse struct {
    ReviewID    string         `json:"review_id"`
    Status      string         `json:"status"`
    Summary     string         `json:"summary"`
    Issues      []ReviewIssue  `json:"issues"`
    Metrics     ReviewMetrics  `json:"metrics"`
    CreatedAt   time.Time      `json:"created_at"`
}

type ReviewIssue struct {
    File     string `json:"file"`
    Line     int    `json:"line"`
    Severity string `json:"severity"`
    Message  string `json:"message"`
    Category string `json:"category"`
}

type ReviewMetrics struct {
    FilesScanned    int     `json:"files_scanned"`
    IssuesFound     int     `json:"issues_found"`
    CriticalIssues  int     `json:"critical_issues"`
    CodeQualityScore float64 `json:"code_quality_score"`
}
```

### 6.2 Write Tests
```go
func TestReviewResponse_JSON(t *testing.T) {
    // Test JSON marshaling/unmarshaling
}

func TestReviewIssue_Severity(t *testing.T) {
    // Test severity validation
}
```

### 6.3 Validate
```bash
go build ./internal/review/...
go test ./internal/review/...
golangci-lint run ./internal/review/...
git diff
```

### 6.4 Commit
```bash
git add internal/review/models.go internal/review/models_test.go
git commit -m "feat(review): add ReviewResponse structs"
```

### 6.5 Check Time/Tokens

**After 2nd commit:**
- Time: ~25 minutes
- Tokens: ~60K
- Decision: **PUSH and BREAK**

---

## Phase 7: Session Completion (5 minutes)

### 7.1 Push All Commits
```bash
# Check unpushed commits
git log origin/feature/###-description..HEAD

# Push
git push
```

### 7.2 Update Devlog

**Create or update today's log:**

**File:** `.docs/devlog/2025-10-21.md`

```markdown
# DevLog: 2025-10-21

## Session 1: 14:00-14:25 (25 min)

**Issue:** #011 - Review Service Foundation

**Branch:** feature/011-review-service-foundation

**Completed:**
1. ✅ ReviewRequest struct + validation + tests
2. ✅ ReviewResponse structs (main, Issue, Metrics) + tests

**Commits:** 2
- feat(review): add ReviewRequest struct with validation
- feat(review): add ReviewResponse structs

**Tests:** All passing (100% coverage on models)

**Next Session:**
- Implement ReviewHandler function
- Follow Portal handler pattern
- Add middleware integration

**Blockers:** None

**Context Preserved:**
- Models defined in: internal/review/models.go
- Test patterns in: internal/review/models_test.go
```

### 7.3 Clean Exit

**If this was a good stopping point:**
```bash
# Switch to development for safety
git checkout development

# Or stay on feature branch if continuing soon
```

**If you'll continue in same session:**
- Take 5-10 minute break
- Check time budget (don't go >60 min total)
- Start fresh with Phase 2

---

## Special Scenarios

### Scenario A: Bug Fix Mid-Implementation

**Situation:** You discover a bug while implementing

**Workflow:**
1. **STOP current work**
2. **COMMIT current work** (even if incomplete)
   ```bash
   git add .
   git commit -m "WIP: partial implementation of X (pausing for bug fix)"
   ```
3. **Fix the bug** in separate commit
4. **Resume original work**
5. **Clean up WIP commit** if needed:
   ```bash
   # Squash WIP commits before PR
   git rebase -i development
   ```

### Scenario B: Unexpected Complexity

**Situation:** Task is 3x more complex than estimated

**Workflow:**
1. **STOP and COMMIT** what you have
2. **PUSH** your progress
3. **RE-PLAN** the task:
   - Break into smaller units
   - Update todo list
   - Estimate token/time budget
4. **Ask for help** if needed:
   - Update devlog with blocker
   - Tag issue for review
   - Consult architecture docs

**DON'T:**
- ❌ Push through and hope
- ❌ Skip validation to save time
- ❌ Accumulate massive context

### Scenario C: Crash Recovery

**Situation:** Claude Code crashes mid-session

**Workflow:**
1. **Check git status:**
   ```bash
   git status
   ```
2. **Read recovery logs** (if available):
   ```bash
   cat .claude/recovery/session-YYYYMMDD-HHMMSS.md
   ```
3. **Review uncommitted changes:**
   ```bash
   git diff
   ```
4. **Decide:**
   - ✅ Changes look good → Validate and commit
   - ❌ Changes incomplete → Discard and restart
5. **Resume from last good commit:**
   ```bash
   git log -1 --stat
   # Read commit to understand where you were
   ```

### Scenario D: Test Fails Repeatedly

**Situation:** Can't get test to pass after 3 attempts

**Workflow:**
1. **COMMIT the test** (even failing):
   ```bash
   git add internal/review/handler_test.go
   git commit -m "test(review): add failing test for ReviewHandler (TDD RED)"
   ```
2. **PUSH and ASK FOR HELP:**
   - Update devlog with blocker
   - Provide test code
   - Describe expected vs actual behavior
3. **OR SIMPLIFY:**
   - Break test into smaller tests
   - Test one thing at a time
   - Remove complex assertions

**DON'T:**
- ❌ Skip the test
- ❌ Change test to match wrong behavior
- ❌ Spend >20 minutes debugging alone

---

## Anti-Patterns to Avoid

### ❌ Anti-Pattern 1: Context Hoarding
```
Bad workflow:
1. Read 15 files "to understand"
2. Context at 120K tokens
3. Haven't written any code yet
4. Crash risk: HIGH
```

**Fix:** Read one file, implement one thing, commit.

### ❌ Anti-Pattern 2: Test Procrastination
```
Bad workflow:
1. Implement 5 functions
2. "I'll write tests at the end"
3. Write all tests at once
4. 7 tests fail
5. Can't remember what you were thinking
```

**Fix:** Write test immediately after each function.

### ❌ Anti-Pattern 3: Validation Avoidance
```
Bad workflow:
1. Implement feature
2. Skip build check
3. Skip test run
4. Commit directly
5. CI fails
6. Spend 30 min fixing
```

**Fix:** Run pre-commit checklist ALWAYS.

### ❌ Anti-Pattern 4: Marathon Sessions
```
Bad workflow:
1. Start implementing
2. Hour 1: Making progress
3. Hour 2: Still going
4. Hour 2.5: Crash
5. Lost 2.5 hours of work
```

**Fix:** 25-minute time-boxes with mandatory commits.

### ❌ Anti-Pattern 5: Batch Commits
```
Bad workflow:
1. Implement 10 features
2. Write 20 tests
3. Update 5 docs
4. Make ONE giant commit
5. Can't isolate issues
6. Can't revert cleanly
```

**Fix:** One logical unit = one commit.

---

## Success Metrics

### Track Your Performance

**After each session, record:**

#### Effectiveness Metrics
- ✅ Micro-tasks completed
- ✅ Commits created
- ✅ Tests written
- ✅ First-time validation pass rate

#### Efficiency Metrics
- ⏱️ Session duration
- 📊 Context tokens used
- 🔄 Rework cycles needed
- 💥 Crashes experienced

#### Quality Metrics
- ✅ Test coverage achieved
- ✅ Linting issues found
- ✅ Code review feedback
- ✅ Bugs found later

### Weekly Review

**Every Friday, analyze:**

**What went well:**
- Sessions that stayed <25 min
- High first-time validation pass rate
- Smooth incremental progress

**What needs improvement:**
- Sessions that went >30 min
- Multiple validation failures
- Complex tasks not broken down

**Adjustments:**
- Break tasks smaller
- Read less context
- Validate more frequently
- Commit more often

---

## Quick Reference Card

### Session Checklist

**Start:**
- ✅ Correct branch?
- ✅ Issue understood?
- ✅ Devlog read?
- ✅ Timer set (25 min)?

**During:**
- ✅ Reading minimal context?
- ✅ Implementing small units?
- ✅ Testing immediately?
- ✅ Validating before commit?

**After Each Unit:**
- ✅ Build passes?
- ✅ Tests pass?
- ✅ Linting clean?
- ✅ Committed?

**Session End:**
- ✅ All commits pushed?
- ✅ Devlog updated?
- ✅ Todo list current?
- ✅ Clean exit?

### Red Flags

**STOP if:**
- 🔴 >25 minutes without commit
- 🔴 >100K tokens accumulated
- 🔴 3+ validation failures in a row
- 🔴 Feeling lost or confused
- 🔴 Task more complex than estimated

### Green Lights

**Continue if:**
- 🟢 <20 minutes elapsed
- 🟢 <80K tokens used
- 🟢 Validation passing first-time
- 🟢 Clear on next steps
- 🟢 Making steady progress

---

## Tool-Specific Guidance

### Claude Code (CLI)

**Strengths:**
- 200K context window
- Multi-file understanding
- Architecture reasoning

**Limitations:**
- V8 crash risk
- Cannot run interactive scripts
- Context accumulates

**Optimal workflow:**
1. Short sessions (15-25 min)
2. Minimal file reading
3. Frequent commits
4. Manual command execution

**Commands to run separately:**
```bash
# Don't chain commands - run one at a time
go build ./internal/review/...
# Wait for result, then:
go test ./internal/review/...
# Wait for result, then:
golangci-lint run ./internal/review/...
```

### GitHub Copilot (IDE)

**Strengths:**
- Inline suggestions
- Real-time feedback
- IDE integration

**Limitations:**
- Limited context
- No autonomous workflow
- Requires human guidance

**Optimal workflow:**
1. Write test first
2. Let Copilot suggest implementation
3. Review and adjust suggestion
4. Run validation in terminal
5. Commit

### OpenHands (Autonomous)

**Strengths:**
- Autonomous execution
- Checkpoint/resume
- Can run scripts

**Limitations:**
- Model quality varies
- Needs detailed specs
- Less architectural reasoning

**Optimal workflow:**
1. Provide detailed issue spec
2. Let it implement autonomously
3. Review PR output
4. Provide feedback via issues

---

## Integration with Existing Workflow

### Fits into DevSmith Workflow

**Reference:** `.docs/WORKFLOW-GUIDE.md`

**This template enhances Step 3: "Work on the Issue"**

```
Standard workflow:
1. Pick issue
2. Create feature branch
3. Work on issue          ← THIS TEMPLATE GOES HERE
4. Check your work
5. Stage and commit
6. Create PR
```

**Detailed flow:**
```
1. Pick issue (.docs/issues/###)
2. Create feature branch (feature/###-description)
3. FOR EACH micro-task:
   a. Read minimal context
   b. Implement unit
   c. Write test
   d. Validate (pre-commit checklist)
   e. Commit
   f. Check time/tokens
   g. Continue or break
4. Push all commits
5. Update devlog
6. Create PR (auto-created on push)
```

---

## Related Documents

### Foundation
- `.docs/root_cause_analysis.md` - Why this workflow exists
- `ARCHITECTURE.md` - System architecture
- `DevSmithRoles.md` - Agent roles

### Templates & Guides
- `.docs/PRE_COMMIT_CHECKLIST.md` - Validation checklist
- `.docs/INCREMENTAL_COMMIT_GUIDE.md` - Commit strategy
- `.docs/WORKFLOW-GUIDE.md` - Git workflow fundamentals

### Issue Tracking
- `.docs/issues/TEMPLATE-COPILOT.md` - Issue template
- `.docs/devlog/README.md` - Devlog system

---

## Conclusion

**This workflow is battle-tested and designed for the realities of AI-assisted development.**

**Key Principles:**
1. 🎯 **Small batches** - 10-15 minute increments
2. ⚡ **Fast feedback** - Validate immediately
3. 💾 **Commit frequently** - Every logical unit
4. 📊 **Manage context** - Stay under 100K tokens
5. ⏰ **Time-box sessions** - 25 minutes maximum

**Expected Results:**
- ✅ 60-70% reduction in crash risk
- ✅ 50-60% reduction in rework time
- ✅ Maintained or improved quality
- ✅ Increased confidence and momentum
- ✅ Better work-life balance (less frustration)

**Remember:**
> Perfect is the enemy of done.
> Done is the friend of shipped.
> Shipped is the friend of feedback.
> Feedback is the friend of improvement.

**Commit early. Commit often. Ship with confidence.**

---

**Last Updated:** 2025-10-21
**Version:** 1.0
**Status:** Active - Recommended for all agent workflows
