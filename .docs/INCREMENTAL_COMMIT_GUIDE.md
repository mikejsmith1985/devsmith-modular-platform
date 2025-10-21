# Incremental Commit Strategy Guide

**Purpose:** Master the art of micro-commits to reduce work loss, improve code quality, and build momentum

**Target Audience:** All developers and coding agents

**Core Philosophy:** "Commit the smallest unit that adds value"

---

## The Problem with Traditional Commits

### Traditional Approach (Problematic)
```
❌ Work for 2 hours
❌ Implement entire feature
❌ Write all tests
❌ Fix all bugs
❌ Make ONE commit
❌ Push

Problems:
- Lost 2 hours if crash occurs
- Can't isolate which change broke tests
- Difficult to review (too much at once)
- Can't revert part of the work
- No sense of progress until end
```

### Incremental Approach (Recommended)
```
✅ Work for 10 minutes
✅ Implement one function
✅ Write test for that function
✅ Make commit (5 minutes of work saved!)
✅ Repeat

Benefits:
- Lost maximum 10 minutes if crash
- Can bisect to find breaking change
- Easy to review (atomic changes)
- Can revert individual commits
- Constant sense of progress
```

---

## What is a "Logical Unit"?

### Definition
> A **logical unit** is the smallest change that:
> 1. Makes sense in isolation
> 2. Can be described in one sentence
> 3. Doesn't break existing functionality
> 4. Adds or modifies exactly one concept

### Examples of Logical Units

#### ✅ GOOD - These are logical units
- "Add User struct with JSON tags"
- "Implement email validation function"
- "Write test for email validation"
- "Add database migration for users table"
- "Update README with setup instructions"
- "Fix null pointer in GetUser handler"
- "Refactor parseRequest to use helper"
- "Add error handling to ProcessReview"

#### ❌ BAD - These are too large
- "Implement entire authentication system"
- "Add review service with tests and docs"
- "Refactor all handlers"
- "Fix all bugs in portal service"

#### ❌ BAD - These are too small
- "Add comma to line 45"
- "Fix typo"
- "Add newline"

**Exception:** Typo fixes are okay if that's ALL the commit does.

---

## The Commit Spectrum

### Granularity Scale

```
Too Small          Ideal Range               Too Large
    ↓          ↓              ↓                   ↓
    •          •━━━━━━━━━━━━━━•                   •
  Single    Single    Single      Multiple    Entire
   line    function   file         files      feature

 Not useful  ✅ GOOD  ✅ GOOD      🟡 OK     ❌ BAD
```

### Ideal Granularity by Change Type

| Change Type | Ideal Granularity | Example |
|-------------|------------------|---------|
| New struct/type | 1 commit | Add ReviewRequest struct |
| New function | 1 commit per function | Add ValidateEmail function |
| Test | 1 commit per test suite | Add tests for ReviewRequest |
| Bug fix | 1 commit (small fixes) | Fix null check in Handler |
| Refactor | 1 commit per concept | Extract validation to helper |
| Documentation | 1 commit per doc | Update API documentation |
| Configuration | 1 commit per config | Add Redis cache config |

---

## Commit Patterns

### Pattern 1: Test-Driven Development (TDD)

**Workflow:**
```bash
# Commit 1: Write failing test
git add internal/review/handler_test.go
git commit -m "test(review): add failing test for ProcessReview (TDD RED)"

# Commit 2: Implement minimal code to pass
git add internal/review/handler.go
git commit -m "feat(review): implement ProcessReview handler (TDD GREEN)"

# Commit 3: Refactor if needed
git add internal/review/handler.go
git commit -m "refactor(review): extract validation to helper (TDD REFACTOR)"
```

**Benefits:**
- Clear progression (RED → GREEN → REFACTOR)
- Each commit is a checkpoint
- Easy to see what changed at each stage

### Pattern 2: Struct-Then-Test

**Workflow:**
```bash
# Commit 1: Define struct
git add internal/review/models.go
git commit -m "feat(review): add ReviewRequest struct"

# Commit 2: Add validation
git add internal/review/validation.go
git commit -m "feat(review): add ReviewRequest validation"

# Commit 3: Write tests
git add internal/review/validation_test.go
git commit -m "test(review): add ReviewRequest validation tests"
```

### Pattern 3: Layer by Layer

**Workflow:**
```bash
# Commit 1: Database layer
git add internal/review/repository.go
git commit -m "feat(review): add ReviewRepository interface"

# Commit 2: Database implementation
git add internal/review/postgres_repository.go
git commit -m "feat(review): implement PostgreSQL ReviewRepository"

# Commit 3: Service layer
git add internal/review/service.go
git commit -m "feat(review): add ReviewService with business logic"

# Commit 4: Handler layer
git add internal/review/handler.go
git commit -m "feat(review): add HTTP handler for reviews"
```

### Pattern 4: Feature + Fix

**Workflow:**
```bash
# Commit 1: Implement feature
git add internal/portal/dashboard.go
git commit -m "feat(portal): add dashboard handler"

# Commit 2: Discover bug while testing
git add internal/portal/dashboard.go
git commit -m "fix(portal): handle empty user session in dashboard"

# Commit 3: Add test for bug fix
git add internal/portal/dashboard_test.go
git commit -m "test(portal): add test for empty session handling"
```

---

## Commit Messages

### Anatomy of a Good Commit Message

**Format:**
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Example:**
```
feat(review): add ReviewRequest validation

Implement validation for ReviewRequest struct:
- Required: repo_url, branch, mode
- Optional: file_filter
- Mode must be one of: preview, skim, scan, detailed, critical

Uses validator/v10 library consistent with other services.
```

### Type Prefixes

**Common types:**
- `feat`: New feature or enhancement
- `fix`: Bug fix
- `test`: Add or update tests
- `refactor`: Code change that neither fixes bug nor adds feature
- `docs`: Documentation only
- `chore`: Maintenance tasks (dependencies, config)
- `ci`: CI/CD changes
- `perf`: Performance improvement
- `style`: Code style/formatting (no logic change)

### Scope Guidelines

**Scopes should be:**
- Service name: `portal`, `review`, `logs`, `analytics`
- Layer name: `api`, `db`, `auth`, `cache`
- Component: `handler`, `service`, `repository`, `middleware`

**Examples:**
- `feat(review): ...`
- `fix(portal-auth): ...`
- `refactor(review-service): ...`
- `test(logs-handler): ...`

### Description Guidelines

**Good descriptions:**
- ✅ Start with verb (add, implement, fix, update, refactor)
- ✅ Use imperative mood ("add feature" not "added feature")
- ✅ Keep under 50 characters
- ✅ Don't end with period
- ✅ Be specific

**Examples:**
```
✅ feat(review): add ReviewRequest struct with validation
✅ fix(portal): handle nil user in session middleware
✅ refactor(logs): extract WebSocket logic to service
✅ test(analytics): add tests for metrics calculation
```

**Bad examples:**
```
❌ "updates"
❌ "fixed stuff"
❌ "WIP"
❌ "more changes"
❌ "feat: did some work on the review service and also fixed a bug in portal and updated some tests"
```

---

## Practical Examples

### Example 1: Adding a New Feature

**Task:** Add code review service endpoint

#### Traditional Approach (1 commit)
```bash
# Work for 2 hours
# Implement: models, validation, handler, tests, docs
git add .
git commit -m "feat: add code review service"
git push

# Lost 2 hours if crash before commit
# Hard to review 500+ lines
```

#### Incremental Approach (8 commits)
```bash
# Commit 1 (5 min work)
git add internal/review/models.go
git commit -m "feat(review): add ReviewRequest struct"

# Commit 2 (5 min work)
git add internal/review/models.go
git commit -m "feat(review): add ReviewResponse struct"

# Commit 3 (8 min work)
git add internal/review/models_test.go
git commit -m "test(review): add tests for review models"

# Commit 4 (10 min work)
git add internal/review/service.go
git commit -m "feat(review): implement ReviewService interface"

# Commit 5 (12 min work)
git add internal/review/ai_service.go
git commit -m "feat(review): implement AI-powered review logic"

# Commit 6 (8 min work)
git add internal/review/service_test.go
git commit -m "test(review): add ReviewService tests"

# Commit 7 (10 min work)
git add internal/review/handler.go
git commit -m "feat(review): add HTTP handler for reviews"

# Commit 8 (7 min work)
git add internal/review/handler_test.go
git commit -m "test(review): add handler integration tests"

# Push all at once (or incrementally)
git push

# Lost maximum 12 minutes if crash
# Easy to review commit-by-commit
```

### Example 2: Bug Fix

**Task:** Fix null pointer panic in review handler

#### Traditional Approach
```bash
# Fix bug + add test + refactor related code
git add .
git commit -m "fix: review handler issues"
```

#### Incremental Approach
```bash
# Commit 1: Fix the immediate bug
git add internal/review/handler.go
git commit -m "fix(review): add nil check for request body"

# Commit 2: Add test to prevent regression
git add internal/review/handler_test.go
git commit -m "test(review): add test for nil request body"

# Commit 3: Refactor error handling (if needed)
git add internal/review/handler.go
git commit -m "refactor(review): use consistent error handling pattern"
```

**Benefits:**
- Can cherry-pick just the bug fix (commit 1) to hotfix branch
- Can see exactly what test was added (commit 2)
- Can discuss refactoring separately (commit 3)

### Example 3: Documentation Update

**Task:** Update setup documentation

#### Incremental Approach
```bash
# Commit 1: Update installation steps
git add README.md
git commit -m "docs: update installation steps for Go 1.21"

# Commit 2: Add configuration examples
git add README.md
git commit -m "docs: add example .env configuration"

# Commit 3: Update API documentation
git add docs/API.md
git commit -m "docs: document review service endpoints"
```

---

## When to Combine vs Split

### Combine When:
- ✅ Changes are tightly coupled
- ✅ One without the other breaks functionality
- ✅ Both changes describe same concept

**Example:**
```bash
# These should be ONE commit
git add internal/review/models.go
git add internal/review/models_test.go  # (if test is simple)
git commit -m "feat(review): add ReviewRequest struct with tests"
```

### Split When:
- ✅ Changes are independent
- ✅ Each change makes sense alone
- ✅ Changes touch different concerns

**Example:**
```bash
# These should be TWO commits

# Commit 1
git add internal/review/handler.go
git commit -m "feat(review): add ProcessReview handler"

# Commit 2
git add internal/portal/dashboard.go
git commit -m "feat(portal): add link to review service"
```

### Judgment Call Examples

#### Scenario 1: New Function + Test
```bash
# Option A: Combined (if test is simple)
git add handler.go handler_test.go
git commit -m "feat(review): add ValidateRequest function with tests"

# Option B: Separate (if test is complex)
git add handler.go
git commit -m "feat(review): add ValidateRequest function"
git add handler_test.go
git commit -m "test(review): add comprehensive ValidateRequest tests"

# Use Option A for unit tests
# Use Option B for complex integration tests
```

#### Scenario 2: Struct + Multiple Tests
```bash
# Split if tests are substantial

# Commit 1
git add models.go
git commit -m "feat(review): add ReviewRequest struct"

# Commit 2
git add validation_test.go
git commit -m "test(review): add validation tests for ReviewRequest"

# Commit 3
git add serialization_test.go
git commit -m "test(review): add JSON serialization tests"
```

---

## Commit Timing Strategies

### Strategy 1: Time-Boxed (Recommended)

**Rule:** Commit every 10-15 minutes, regardless

**Benefits:**
- Regular checkpoints
- Reduces crash risk
- Forces atomic changes

**Implementation:**
```bash
# Set timer
timer 15m "Time to commit!"

# When timer goes off
git add <changed-files>
git commit -m "..."

# Reset timer
timer 15m "Time to commit!"
```

### Strategy 2: Test-Driven

**Rule:** Commit after each TDD cycle

**Cycle:**
1. Write test → Commit (RED)
2. Implement → Commit (GREEN)
3. Refactor → Commit (REFACTOR)

**Benefits:**
- Natural checkpoints
- Clear progression
- Testable increments

### Strategy 3: Validation-Driven

**Rule:** Commit when validation passes

**Workflow:**
```bash
# Make change
# Run validation
go build && go test && golangci-lint run

# If all pass → commit
# If any fail → fix and retry

# Only commit when GREEN
git commit -m "..."
```

### Strategy 4: Feature-Slice

**Rule:** Commit vertical slices of features

**Example:**
```bash
# Slice 1: Happy path only
git commit -m "feat(review): implement basic review flow (happy path)"

# Slice 2: Error handling
git commit -m "feat(review): add error handling for review flow"

# Slice 3: Edge cases
git commit -m "feat(review): handle edge cases in review flow"

# Slice 4: Performance
git commit -m "perf(review): add caching to review results"
```

---

## Advanced Techniques

### Technique 1: Commit Squashing

**When:** Multiple WIP commits need cleanup before PR

**Workflow:**
```bash
# During development (make lots of small commits)
git commit -m "WIP: trying approach A"
git commit -m "WIP: approach A didn't work"
git commit -m "WIP: trying approach B"
git commit -m "WIP: approach B works!"
git commit -m "WIP: added tests"

# Before PR (squash related commits)
git rebase -i development

# In interactive rebase, mark commits:
pick abc1234 WIP: trying approach A
squash def5678 WIP: approach A didn't work
squash ghi9012 WIP: trying approach B
pick jkl3456 WIP: approach B works!
squash mno7890 WIP: added tests

# Results in 2 clean commits:
# 1. "feat(review): experiment with different approaches"
# 2. "feat(review): implement working solution with tests"
```

**IMPORTANT:** Only squash before pushing or on unshared branches!

### Technique 2: Commit Splitting

**When:** Accidentally committed too much

**Workflow:**
```bash
# Oops, committed 2 unrelated changes
git log -1
# "feat: add review handler and fix portal bug"

# Split into 2 commits
git reset HEAD~1  # Undo commit, keep changes

# Commit 1: Review handler
git add internal/review/handler.go
git commit -m "feat(review): add review handler"

# Commit 2: Portal bug fix
git add internal/portal/auth.go
git commit -m "fix(portal): fix session timeout bug"
```

### Technique 3: Commit Amending

**When:** Need to fix last commit (typo, forgot file, etc.)

**Workflow:**
```bash
# Made commit
git commit -m "feat(review): add validation"

# Oops, forgot to add test file
git add validation_test.go
git commit --amend --no-edit

# Or fix commit message
git commit --amend -m "feat(review): add request validation"
```

**WARNING:** Only amend unpushed commits or you'll need force push!

### Technique 4: Partial File Staging

**When:** File has multiple unrelated changes

**Workflow:**
```bash
# File has 2 changes: feature + bug fix
git add -p handler.go

# Git will prompt for each change:
# Stage this hunk [y,n,q,a,d,e,?]?

# Stage only feature changes
y  # yes to feature hunks
n  # no to bug fix hunks

git commit -m "feat(review): add new handler method"

# Now stage bug fix
git add handler.go
git commit -m "fix(review): fix validation logic"
```

---

## Common Mistakes

### Mistake 1: "WIP" Commits Left in PR

**Problem:**
```
❌ git log:
abc1234 WIP
def5678 WIP more stuff
ghi9012 WIP fixed it
jkl3456 feat(review): add review service
```

**Solution:**
```bash
# Squash WIP commits before PR
git rebase -i development
# Squash all WIP into meaningful commits
```

### Mistake 2: Mixing Refactor with Features

**Problem:**
```
❌ One commit:
- Adds new feature (100 lines)
- Refactors unrelated code (200 lines)
- Updates documentation (50 lines)
```

**Solution:**
```bash
# Commit 1: Refactor first
git add <refactored-files>
git commit -m "refactor(review): extract helper functions"

# Commit 2: Then feature
git add <feature-files>
git commit -m "feat(review): add new validation logic"

# Commit 3: Then docs
git add README.md
git commit -m "docs: update validation documentation"
```

### Mistake 3: Committing Broken Code

**Problem:**
```
❌ git commit -m "feat: add handler" # (but tests fail)
```

**Solution:**
```bash
# Always validate before commit
go build && go test

# Only commit if green
✅ git commit -m "feat: add handler"
```

**Exception:** TDD RED commits (explicitly marked)
```
✅ git commit -m "test(review): add failing test for handler (TDD RED)"
```

### Mistake 4: Vague Commit Messages

**Problem:**
```
❌ "updates"
❌ "fixes"
❌ "changes"
❌ "WIP"
```

**Solution:**
```
✅ "feat(review): add validation for ReviewRequest"
✅ "fix(portal): handle nil user session"
✅ "refactor(logs): extract WebSocket logic"
✅ "test(analytics): add metrics calculation tests"
```

---

## Metrics & Goals

### Track These Metrics

**Per Day:**
- Number of commits made
- Average commit size (lines changed)
- Commits per hour of work
- First-time validation pass rate

**Per Week:**
- Most common commit types
- Largest commit (identify outliers)
- Time between commits (average)
- Commits reverted (should be low)

### Target Goals

| Metric | Target | Excellent |
|--------|--------|-----------|
| Commits per hour | 3-6 | 6+ |
| Lines per commit | <100 | <50 |
| Time between commits | <15 min | <10 min |
| Validation pass rate | >80% | >95% |
| Commits reverted | <5% | <2% |

### Warning Signs

**Too few commits:**
- <2 commits per hour
- Average commit >200 lines
- >30 min between commits
→ **Risk:** Work loss, difficult review, hard to debug

**Too many commits:**
- >10 commits per hour
- Average commit <10 lines
- Commits every 2-3 minutes
→ **Risk:** Noise, meaningless history, review fatigue

**Sweet spot:** 3-6 meaningful commits per hour

---

## Integration with DevSmith Workflow

### Fits into Pre-Commit Checklist

**Reference:** `.docs/PRE_COMMIT_CHECKLIST.md`

**Enhanced workflow:**
```
1. Implement smallest unit (10 min)
2. Run pre-commit checklist (3 min)
   - Build
   - Test
   - Lint
   - Self-review
3. Commit if all pass (1 min)
4. Repeat
```

### Fits into Agent Workflow

**Reference:** `.docs/AGENT_WORKFLOW_TEMPLATE.md`

**Phase 2: Implement → Phase 4: Commit**
- Implement one unit
- Validate
- **Commit immediately** (this guide)
- Move to next unit

---

## Quick Reference

### Decision Tree: "Should I Commit?"

```
Have I made changes?
  ├─ No → Keep working
  └─ Yes → Do changes represent a logical unit?
       ├─ No → Break down further
       └─ Yes → Does it build?
            ├─ No → Fix compilation
            └─ Yes → Do tests pass?
                 ├─ No → Fix tests
                 └─ Yes → Is linting clean?
                      ├─ No → Fix linting
                      └─ Yes → ✅ COMMIT NOW
```

### Commit Frequency Guidelines

| Situation | Commit Frequency |
|-----------|-----------------|
| Adding new struct | Immediately after struct definition |
| Writing function | After function + basic test |
| Bug fix | After fix + regression test |
| Refactoring | After each refactor concept |
| Documentation | After each logical section |
| Configuration | After each config change |

### Emergency Commit

**If crash seems imminent:**
```bash
# Commit ANYTHING over losing work
git add .
git commit -m "WIP: saving progress before potential crash"
git push

# Clean up later if needed
```

---

## Success Stories

### Before: Traditional Commits
```
Work session: 120 minutes
Commits: 1
Largest commit: 847 lines
Crash at 115 min → Lost everything
Frustration: HIGH
```

### After: Incremental Commits
```
Work session: 120 minutes
Commits: 8
Largest commit: 142 lines
Average commit: 67 lines
Crash at 115 min → Lost 12 minutes of work
Frustration: LOW
Recovery: EASY (just redo last unit)
```

---

## Related Documents

- `.docs/root_cause_analysis.md` - Why incremental commits matter
- `.docs/PRE_COMMIT_CHECKLIST.md` - What to check before commit
- `.docs/AGENT_WORKFLOW_TEMPLATE.md` - How to integrate commits into workflow
- `.docs/WORKFLOW-GUIDE.md` - Overall git workflow
- `ARCHITECTURE.md` - System architecture

---

## Conclusion

**Incremental commits are not overhead—they're insurance.**

**Key Benefits:**
1. ✅ **Reduce work loss** - Maximum 10-15 min at risk
2. ✅ **Easier debugging** - `git bisect` finds breaking changes fast
3. ✅ **Better reviews** - Reviewers can follow your logic
4. ✅ **Cleaner history** - Each commit tells a story
5. ✅ **Build momentum** - Constant sense of progress
6. ✅ **Enable collaboration** - Others can build on your work sooner

**Anti-Benefits of Large Commits:**
1. ❌ **High work loss** - Hours of work at risk
2. ❌ **Debugging nightmare** - Which of 847 lines broke tests?
3. ❌ **Review fatigue** - "LGTM" without actually reviewing
4. ❌ **Messy history** - "Update stuff" commits everywhere
5. ❌ **Discouragement** - No progress visible for hours
6. ❌ **Blocked collaboration** - Others wait for your "big" commit

**Remember:**
> Commit early, commit often.
> Each commit is a save point.
> Small commits, big wins.

**Adopt this strategy. Your future self will thank you.**

---

**Last Updated:** 2025-10-21
**Version:** 1.0
**Status:** Active - Required for all development work
