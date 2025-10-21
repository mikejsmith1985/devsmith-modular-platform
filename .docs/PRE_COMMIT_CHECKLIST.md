# Pre-Commit Validation Checklist

**Purpose:** Catch issues before committing to reduce rework cycles and improve first-time quality.

**Target Audience:** All coding agents (Claude, GitHub Copilot, OpenHands) and human developers

**When to Use:** Before EVERY commit, no exceptions

---

## Quick Reference

**Time Investment:** 2-5 minutes per commit
**Rework Prevented:** 15-60 minutes per issue caught
**ROI:** 300-1200% time savings

---

## The Checklist

### Phase 1: Build & Compile ⚙️

**Purpose:** Ensure code compiles before testing

#### Go Projects
```bash
# Build specific service
go build ./cmd/{service-name}

# Or build all
go build ./...

# Expected: No compilation errors
```

**Common Issues Caught:**
- ✅ Missing imports
- ✅ Undefined variables
- ✅ Type mismatches
- ✅ Syntax errors

**Result Required:** ✅ **PASS** - No compilation errors

---

### Phase 2: Unit Tests 🧪

**Purpose:** Verify functionality works as expected

#### Run Relevant Tests
```bash
# Test specific package
go test ./internal/{package}/...

# Or test everything (slower)
make test

# With coverage
go test -cover ./internal/{package}/...
```

**What to Check:**
- ✅ All tests pass
- ✅ No panics or crashes
- ✅ Coverage doesn't decrease
- ✅ New code has tests

**Result Required:** ✅ **PASS** - All tests green

**If Tests Fail:**
1. ❌ **DO NOT COMMIT**
2. Fix the issue immediately
3. Re-run checklist from Phase 1

---

### Phase 3: Linting & Style 🎨

**Purpose:** Ensure code follows project standards

#### Run Linters
```bash
# Full linting
golangci-lint run ./...

# Specific package
golangci-lint run ./internal/{package}/...

# Auto-fix if possible
golangci-lint run --fix ./...
```

**Common Issues Caught:**
- ✅ Unused imports
- ✅ Variable shadowing
- ✅ Inefficient code patterns
- ✅ Missing error checks
- ✅ Complexity violations

**Result Required:** ✅ **PASS** - No linting errors

**Warnings:**
- 🟡 Acceptable: <5 warnings on new code
- 🔴 Unacceptable: Any critical warnings
- 🟢 Ideal: Zero warnings

---

### Phase 4: Code Review Self-Check 👀

**Purpose:** Catch issues a human reviewer would catch

#### 4.1 Review Your Diff
```bash
# See what's changing
git diff

# Or staged changes only
git diff --cached
```

**Questions to Ask:**
- ❓ Do these changes match my commit message intent?
- ❓ Did I accidentally include unrelated changes?
- ❓ Are there debug statements I should remove?
- ❓ Are there commented-out code blocks?
- ❓ Is the formatting consistent?

#### 4.2 Check for Sensitive Data
```bash
git diff | grep -i "password\|secret\|key\|token"
```

**Never Commit:**
- ❌ Passwords or API keys
- ❌ Database credentials
- ❌ Private keys or certificates
- ❌ Personal information
- ❌ `.env` files (unless `.env.example`)

#### 4.3 Verify File Changes
```bash
git status
```

**Checklist:**
- ✅ Only intended files are staged
- ✅ No accidental deletions
- ✅ No generated files (unless intentional)
- ✅ No IDE config files (.vscode, .idea)
- ✅ No temporary files (*.tmp, *.log)

---

### Phase 5: Architectural Alignment 🏗️

**Purpose:** Ensure changes follow project architecture

#### 5.1 Bounded Context Check
**Question:** Does this change respect service boundaries?

**Examples:**
- ✅ Portal code in `internal/portal/`
- ✅ Review code in `internal/review/`
- ❌ Portal code directly calling review database
- ❌ Shared business logic in multiple services

**Reference:** `ARCHITECTURE.md` Section 2 (Bounded Contexts)

#### 5.2 Dependency Direction Check
**Question:** Are dependencies flowing in the correct direction?

**Rules:**
- ✅ Core → (none)
- ✅ Domain → Core
- ✅ Application → Domain + Core
- ✅ Infrastructure → All layers
- ❌ Core → Domain (WRONG)
- ❌ Domain → Application (WRONG)

**Reference:** `ARCHITECTURE.md` Section 3 (Clean Architecture)

#### 5.3 Pattern Consistency Check
**Question:** Am I following existing patterns in the codebase?

**Check:**
```bash
# Find similar code
grep -r "similar pattern" internal/

# Read example implementation
# Match the style and structure
```

**Principle:** Consistency > Personal Preference

---

### Phase 6: Test Coverage Validation 📊

**Purpose:** Ensure new code is properly tested

#### Run Coverage Report
```bash
# Generate coverage
go test -coverprofile=coverage.out ./internal/{package}/...

# View coverage
go tool cover -func=coverage.out

# Or HTML view
go tool cover -html=coverage.out
```

**Requirements:**
- ✅ New functions have tests
- ✅ Edge cases are covered
- ✅ Error paths are tested
- ✅ Coverage ≥80% (target: ≥90%)

**Acceptable Exceptions:**
- 🟡 Pure getters/setters
- 🟡 Simple constructors
- 🟡 Generated code
- 🔴 Business logic (MUST be tested)

---

### Phase 7: Documentation Check 📚

**Purpose:** Ensure changes are understandable

#### 7.1 Code Comments
**Check:**
- ✅ Public functions have doc comments
- ✅ Complex logic has inline comments
- ✅ Non-obvious decisions are explained
- ✅ TODOs reference issue numbers

**Example:**
```go
// ✅ GOOD
// ProcessReview analyzes code changes and returns a ReviewResult.
// It uses the configured AI model (see config.yaml) and respects
// the review mode specified in the request.
func ProcessReview(ctx context.Context, req ReviewRequest) (*ReviewResult, error) {
    // ...
}

// ❌ BAD
// Process review
func ProcessReview(ctx context.Context, req ReviewRequest) (*ReviewResult, error) {
    // ...
}
```

#### 7.2 README Updates
**When to Update:**
- ✅ New service added
- ✅ New environment variable required
- ✅ New dependency added
- ✅ Setup steps changed
- ✅ API endpoints changed

**Check:**
```bash
# If you changed setup/config
git diff README.md

# Should show your updates
```

---

## Execution Flow

### Standard Workflow

```
┌─────────────────────────┐
│  1. Make Code Changes   │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│   2. Run Checklist      │
│   (Phases 1-7)          │
└───────────┬─────────────┘
            │
         ┌──┴──┐
         │Pass?│
         └──┬──┘
            │
     ┌──────┴──────┐
     │             │
    YES           NO
     │             │
     ▼             ▼
┌─────────┐  ┌──────────┐
│ Commit  │  │ Fix      │
└─────────┘  │ Issues   │
             └────┬─────┘
                  │
                  └─────┐
                        ▼
            ┌────────────────────┐
            │  Retry Checklist   │
            │  from Phase 1      │
            └────────────────────┘
```

### Time Budget per Phase

| Phase | Time | Critical? |
|-------|------|-----------|
| 1. Build | 30s-1min | 🔴 YES |
| 2. Tests | 1-3min | 🔴 YES |
| 3. Linting | 30s-1min | 🟡 HIGH |
| 4. Self-Review | 1-2min | 🟢 MEDIUM |
| 5. Architecture | 30s-1min | 🟡 HIGH |
| 6. Coverage | 30s-1min | 🟡 HIGH |
| 7. Documentation | 30s-1min | 🟢 MEDIUM |
| **TOTAL** | **5-10min** | **Required** |

**ROI Calculation:**
- Time invested: 5-10 minutes
- Average rework time prevented: 20-40 minutes
- ROI: 200-400% time savings

---

## Failure Modes & Solutions

### Failure: Build Fails

**Symptoms:**
```
./cmd/portal/main.go:15:2: undefined: processRequest
```

**Solution:**
1. Read the error message carefully
2. Fix the compilation issue
3. Re-run Phase 1
4. DO NOT proceed to Phase 2

**Common Causes:**
- Missing import
- Typo in function name
- Type mismatch

---

### Failure: Tests Fail

**Symptoms:**
```
--- FAIL: TestProcessReview (0.00s)
    review_test.go:25: expected "pass", got "fail"
```

**Solution:**
1. Identify which test failed
2. Understand why it failed
3. Fix the code (not the test, unless test is wrong)
4. Re-run Phase 1 & 2
5. DO NOT skip to Phase 3

**Common Causes:**
- Logic error in implementation
- Missing edge case handling
- Incorrect test expectations (rare)

---

### Failure: Linting Fails

**Symptoms:**
```
internal/portal/handler.go:45:2: ineffectual assignment to err (ineffassign)
```

**Solution:**
1. Read the linting error
2. Fix the issue (often a real bug!)
3. Re-run Phase 3
4. Consider if auto-fix is safe: `golangci-lint run --fix`

**Common Causes:**
- Unused variables
- Missing error handling
- Inefficient code patterns
- Code complexity too high

---

### Failure: Coverage Too Low

**Symptoms:**
```
coverage: 45.2% of statements
```

**Solution:**
1. Identify untested functions: `go tool cover -html=coverage.out`
2. Write missing tests
3. Re-run Phase 2 & 6

**Common Causes:**
- Forgot to write tests for new code
- Only tested happy path, not error paths
- Complex functions need more test cases

---

## Integration with Git Workflow

### Before First Commit
```bash
# Run full checklist
./scripts/pre-commit-check.sh  # (see: Scripts section)

# Or manually
go build ./...
go test ./...
golangci-lint run ./...
git diff
```

### Before Subsequent Commits
```bash
# Run focused checklist (changed packages only)
go build ./cmd/{changed-service}
go test ./internal/{changed-package}/...
golangci-lint run ./internal/{changed-package}/...
git diff
```

### Create Commit (After Checklist Passes)
```bash
git add {specific-files}
git commit -m "$(cat <<'EOF'
feat(scope): brief description

Longer explanation if needed.

Checklist completed:
✅ Build passes
✅ Tests pass
✅ Linting clean
✅ Self-reviewed
✅ Architecture aligned

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

---

## Automation Scripts

### Script: `pre-commit-check.sh`

**Location:** `scripts/pre-commit-check.sh`

**Usage:**
```bash
# Run full checklist
./scripts/pre-commit-check.sh

# Run for specific service
./scripts/pre-commit-check.sh portal
```

**What it does:**
1. Runs `go build`
2. Runs `go test`
3. Runs `golangci-lint`
4. Checks for sensitive data patterns
5. Validates coverage threshold
6. Generates summary report

**Output:**
```
🔨 Building...                ✅ PASS
🧪 Testing...                 ✅ PASS (98.2% coverage)
🎨 Linting...                 ✅ PASS
🔐 Security Check...          ✅ PASS
🏗️  Architecture Check...     ✅ PASS

✨ All checks passed! Safe to commit.
```

### Script: `quick-check.sh`

**Location:** `scripts/quick-check.sh`

**Usage:**
```bash
# Faster version (skips some checks)
./scripts/quick-check.sh
```

**What it does:**
1. Runs `go build` (changed packages only)
2. Runs `go test` (changed packages only)
3. Runs `golangci-lint` (changed files only)

**When to use:**
- Rapid iteration cycles
- Small changes (1-2 files)
- After full check has passed once

---

## Agent-Specific Guidance

### For Claude Code
**Limitations:**
- Cannot run interactive scripts
- Must run commands individually

**Workflow:**
```bash
# Phase 1
go build ./cmd/portal

# Phase 2
go test ./internal/portal/...

# Phase 3
golangci-lint run ./internal/portal/...

# Phase 4
git diff

# Then commit if all pass
```

### For GitHub Copilot
**Advantages:**
- Can see errors inline in IDE
- Can fix issues as you type

**Workflow:**
1. Write code (Copilot assists)
2. Save file (triggers IDE linting)
3. Fix inline errors
4. Run terminal commands for testing
5. Commit when all green

### For OpenHands
**Advantages:**
- Can run scripts directly
- Can automate checklist

**Workflow:**
```bash
# Run automated checklist
./scripts/pre-commit-check.sh

# If pass, commit
git add . && git commit -m "message"
```

---

## Metrics & Continuous Improvement

### Track These Metrics

**Per Commit:**
- ✅ Checklist completion time
- ✅ Number of issues caught
- ✅ Phase where issues found
- ✅ Rework time saved

**Per Sprint:**
- ✅ First-time commit success rate
- ✅ Average rework cycles per feature
- ✅ Most common failure phase
- ✅ ROI of checklist usage

### Adjust Workflow Based on Data

**If most failures in Phase 1 (Build):**
→ Need better type checking during coding

**If most failures in Phase 2 (Tests):**
→ Need TDD adoption or better test coverage

**If most failures in Phase 3 (Linting):**
→ Need IDE integration or auto-formatting

**If most failures in Phase 5 (Architecture):**
→ Need better architecture documentation or training

---

## Success Stories

### Before Checklist
```
Agent workflow:
1. Implement feature (45 min)
2. Attempt commit
3. Discover build fails (10 type errors)
4. Fix errors (20 min)
5. Discover tests fail (5 failures)
6. Fix tests (25 min)
7. Discover linting issues (8 warnings)
8. Fix linting (10 min)
9. Finally commit
Total: 110 minutes (45 min actual + 65 min rework)
```

### After Checklist
```
Agent workflow:
1. Implement small unit (10 min)
2. Run checklist (3 min) → 1 test fails
3. Fix test immediately (2 min)
4. Run checklist (2 min) → all pass
5. Commit
6. Repeat for next unit
Total: 17 minutes per unit (10 min actual + 7 min validation)
Rework: Minimal (caught within 2 min)
```

**Time Saved:** ~60% per feature

---

## FAQ

### Q: This checklist seems long. Do I really need all phases?

**A:** Yes, but it's faster than rework.
- Checklist: 5-10 minutes
- Rework after failed commit: 20-60 minutes
- ROI: 200-600% time savings

### Q: Can I skip phases for trivial changes?

**A:** Minimum required phases:
- 🔴 Phase 1 (Build): ALWAYS
- 🔴 Phase 2 (Tests): ALWAYS
- 🟡 Phase 3 (Linting): RECOMMENDED
- 🟢 Phase 4-7: Situation-dependent

**Rule:** When in doubt, run all phases.

### Q: What if I'm working on a spike/experiment?

**A:** Use a break-fix branch:
```bash
git checkout -b break-fix/experiment-name
# Experiment freely
# Checklist not required
# NEVER merge to development
```

### Q: Can I automate this checklist?

**A:** Yes! See `scripts/pre-commit-check.sh`
```bash
# Auto-run before every commit
cp scripts/pre-commit-check.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### Q: What if the checklist catches no issues?

**A:** Great! That means:
- ✅ You're writing quality code
- ✅ You're developing good habits
- ✅ The checklist is working

Continue using it to maintain quality.

---

## Related Documents

- `.docs/root_cause_analysis.md` - Why this checklist exists
- `.docs/AGENT_WORKFLOW_TEMPLATE.md` - Step-by-step agent workflow
- `.docs/INCREMENTAL_COMMIT_GUIDE.md` - Micro-commit strategy
- `.docs/WORKFLOW-GUIDE.md` - Git workflow fundamentals
- `ARCHITECTURE.md` - System architecture reference

---

## Conclusion

**This checklist is not bureaucracy—it's insurance.**

Every minute spent on validation prevents 3-6 minutes of rework.
Every issue caught early prevents frustration and wasted effort.
Every commit that passes first-time builds confidence and momentum.

**Adopt this checklist. Your future self will thank you.**

---

**Last Updated:** 2025-10-21
**Version:** 1.0
**Status:** Active - Required for all commits
