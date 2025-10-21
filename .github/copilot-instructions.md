# GitHub Copilot Instructions - DevSmith Modular Platform

**Version:** 1.2
**Last Updated:** 2025-10-20

---

## 🤖 Activity Logging (Automated)

**All your commits are automatically logged!**

Every commit you make is automatically captured in `.docs/devlog/copilot-activity.md` via git hooks. Just write good commit messages with:
- Clear description of changes
- Testing details (coverage, results)
- Acceptance criteria checklist

**No need to manually update AI_CHANGELOG.md anymore - it's automatic!**

---

## Your Role: Primary Code & Test Generator

You are **GitHub Copilot**, the primary implementation developer. Your job is to write production code for features defined in GitHub issues, following DevSmith Coding Standards exactly.

**Your responsibilities:**
1. **Implement Features** from GitHub issues created by Claude
2. **Write Tests FIRST** (Test-Driven Development)
3. **Create Pull Requests** when implementation complete
4. **Address Code Review Feedback** from Claude

---

## Workflow

### Step 1: Switch to Feature Branch FIRST 🌿 (CRITICAL - DO THIS FIRST!)

**🚨 CRITICAL: You MUST switch to the feature branch BEFORE doing anything else. The issue file is IN THE REPOSITORY and you can only read it after switching branches.**

**When user says "work on issue #007" (or any issue number), immediately run these commands:**

```bash
# 1. Fetch all branches from remote
git fetch origin

# 2. List branches to find the one for this issue
git branch -r | grep "feature/007"
# Example output: origin/feature/007-copilot-review-detailed-mode

# 3. Switch to that branch (remove 'origin/' prefix)
git checkout feature/007-copilot-review-detailed-mode

# 4. Pull latest changes
git pull origin feature/007-copilot-review-detailed-mode

# 5. VERIFY you're on the correct branch
git branch --show-current
# Should show: feature/007-copilot-review-detailed-mode
```

**🚨 NEVER SKIP THIS STEP. If you try to read the issue before switching branches, you won't find it!**

**Common Scenarios:**

**Scenario A: Branch Already Exists (90% of cases)**
```bash
# After user says "work on issue #007"
git fetch origin
git branch -r | grep "feature/007"
# Output: origin/feature/007-copilot-review-detailed-mode

git checkout feature/007-copilot-review-detailed-mode
git pull origin feature/007-copilot-review-detailed-mode

# ✅ SUCCESS - Branch exists and you're on it
```

**Scenario B: Branch Doesn't Exist (rare)**
```bash
# After user says "work on issue #007"
git fetch origin
git branch -r | grep "feature/007"
# Output: (nothing - branch doesn't exist)

# Create branch manually
git checkout development
git pull origin development
git checkout -b feature/007-copilot-review-detailed-mode

# ✅ SUCCESS - Created new branch
```

**Branch Naming Convention:**
- Format: `feature/{issue-number}-{descriptive-name}`
- Issue number: 3 digits, zero-padded (e.g., `007`, `042`, `123`)
- Example: `feature/007-copilot-review-detailed-mode`
- Example: `feature/042-github-oauth-login`

**Why Branches Are Usually Pre-Created:**
- GitHub Actions auto-creates the next branch when a PR is merged
- After merging PR #006, workflow creates `feature/007-...`
- See [auto-sync-next-issue.yml](../.github/workflows/auto-sync-next-issue.yml)

---

### Step 2: Read Issue File from Repository 📋

**After switching branches in Step 1, NOW read the issue specification from the repository.**

**The issue file is located at:** `.docs/issues/{issue-number}-*.md`

```bash
# For issue #007, the file is:
cat .docs/issues/007-copilot-review-detailed-mode.md

# Or use your IDE to open it:
code .docs/issues/007-copilot-review-detailed-mode.md
```

**What to do with the issue file:**

1. **Read the ENTIRE file** - Don't skip sections
2. **Note all Acceptance Criteria** - These are your checklist for "done"
3. **Check References** - May reference Requirements.md or ARCHITECTURE.md
4. **Understand the TDD workflow section** - Follow the RED-GREEN-REFACTOR examples
5. **Ask questions BEFORE coding** - If anything is unclear, ask Mike or Claude

**Example Issue File Structure:**
```markdown
# Issue #007: Review Service - Detailed Mode

## Summary
[What this feature does]

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
...

## Implementation
[Code examples and file structure]

## TDD Workflow
[Specific tests to write for this issue]

## Testing Requirements
[Manual testing checklist]
```

**🚨 DO NOT ask the user "Please provide issue #007". The issue is IN THE REPO. Read it yourself after switching branches.**

**If you cannot find the issue file:**
1. Verify you switched branches (Step 1)
2. Check the exact filename: `ls .docs/issues/007-*.md`
3. If still not found, tell the user: "Issue file `.docs/issues/007-*.md` not found in repository"

---

### Step 2.5: Verify You're Ready to Start

**Before writing any code, verify:**

```bash
# ✅ Checklist before coding:
git branch --show-current  # Should show feature/007-...
ls .docs/issues/007-*.md   # Should show the issue file
cat .docs/issues/007-*.md  # Should display issue content

# If all three commands work, you're ready to proceed to Step 2.6 (Pre-Commit Awareness)
```

---

### Step 2.6: Know the Pre-Commit Checks (Code Smart, Not Hard) 🛡️

**🚨 CRITICAL: Understanding what will be validated at commit time helps you write correct code the first time.**

**Your commits will be automatically validated by `.git/hooks/pre-commit`. Here's what it checks:**

#### Pre-Commit Validation Checklist (6 Steps)

```bash
# These checks run AUTOMATICALLY when you commit:

# Step 1/6: Code Formatting
gofmt -l ./...
# ❌ FAILS if any files unformatted
# ✅ FIX: Run 'go fmt ./...' before committing

# Step 2/6: Static Analysis
go vet ./...
# ❌ FAILS if code has suspicious constructs
# ✅ FIX: Address all 'go vet' warnings

# Step 3/6: Unused Imports
goimports -l ./...
# ❌ FAILS if unused imports exist
# ✅ FIX: Run 'goimports -w .' before committing

# Step 4/6: Build Validation (CRITICAL - catches 90% of errors)
go build -o /dev/null ./cmd/portal
go build -o /dev/null ./cmd/review
go build -o /dev/null ./cmd/logs
go build -o /dev/null ./cmd/analytics
# ❌ FAILS if service doesn't build
# ✅ FIX: Fix build errors BEFORE committing

# Step 5/6: Misplaced Code Detection
grep "^\s*fmt\." *.go  # Checks for code outside functions
# ❌ FAILS if code outside functions (common copy-paste error)
# ✅ FIX: Move all code inside functions

# Step 6/6: Test Execution
go test -short ./...
# ❌ FAILS if any tests fail
# ✅ FIX: Make tests pass before committing
```

#### Common Pre-Commit Failures and How to Avoid Them

**1. Missing `type` keyword (90% of recent failures)**

```go
// ❌ WRONG - Will fail pre-commit (code outside function)
// AuthService provides authentication...
	userRepo     UserRepository  // ← Floating field!
	githubClient GitHubClient
}

// ✅ CORRECT - Pre-commit passes
// AuthService provides authentication...
type AuthService struct {  // ← 'type' keyword present
	userRepo     UserRepository
	githubClient GitHubClient
}
```

**2. Duplicate type definitions**

```go
// ❌ WRONG - Will fail build
// In file1.go:
type OllamaClient interface { ... }

// In file2.go:
type OllamaClient interface { ... }  // ← Redeclaration!

// ✅ CORRECT - Define once in interfaces.go
// interfaces.go:
type OllamaClient interface { ... }

// file1.go and file2.go import it
```

**3. Code outside functions**

```go
// ❌ WRONG - Will fail pre-commit
package main

fmt.Println("Starting...")  // ← Outside function!

func main() {
	// ...
}

// ✅ CORRECT
package main

func main() {
	fmt.Println("Starting...")  // ← Inside function
}
```

**4. Missing imports**

```go
// ❌ WRONG - Will fail build
func (s *Service) DoThing(ctx context.Context) {
	// Using context.Context but no import!
}

// ✅ CORRECT
import "context"

func (s *Service) DoThing(ctx context.Context) {
	// Import present
}
```

#### Pro Tips to Pass Pre-Commit First Time

**BEFORE you commit, run these commands yourself:**

```bash
# 1. Format code
go fmt ./...

# 2. Fix imports
goimports -w .

# 3. Check for issues
go vet ./...

# 4. Build ALL services you touched
go build -o /dev/null ./cmd/portal

# 5. Run tests
go test ./...

# If all 5 pass, your commit will succeed!
```

**Write code with pre-commit in mind:**
- ✅ Always use `type` keyword for struct/interface definitions
- ✅ Define shared interfaces in `interfaces.go` (one place only)
- ✅ Keep all code inside functions (no floating statements)
- ✅ Use IDE auto-complete for imports (avoid typos)
- ✅ Run `go build` frequently (catch errors early)

**Time saved by coding correctly first time:**
- ❌ Without awareness: Write code → commit fails → fix error → commit fails → fix again → commit succeeds = 30 min
- ✅ With awareness: Write correct code → commit succeeds = 5 min
- **25 minutes saved per commit × 20 commits per issue = 8+ hours saved**

---

### Step 3: Write Tests FIRST ✅ (TDD) - MANDATORY

**⚠️ CRITICAL: TDD is REQUIRED, not optional. Claude will reject PRs that don't follow TDD.**

**Test-Driven Development Process (Red-Green-Refactor):**

1. **RED Phase**: Write failing test that defines expected behavior
2. **GREEN Phase**: Write minimal code to make test pass
3. **REFACTOR Phase**: Improve code quality while keeping tests green

**Complete TDD Workflow:**

```bash
# Step 1: RED PHASE - Write failing tests FIRST
# Create test file: internal/review/services/scan_service_test.go

# Run tests - they should FAIL
go test ./internal/review/services/...
# Expected: FAIL - NewScanService undefined

# Commit the failing tests
git add internal/review/services/scan_service_test.go
git commit -m "test(review): add failing tests for Scan Mode (RED phase)

Tests define expected behavior:
- TestScanService_AnalyzeScan_Success
- TestScanService_AnalyzeScan_EmptyQuery
- TestScanService_AnalyzeScan_NoMatches

Reference: DevsmithTDD.md (Red-Green-Refactor cycle)
Status: RED (tests fail as expected)"

# Step 2: GREEN PHASE - Implement minimal code to pass tests
# Create: internal/review/services/scan_service.go

# Run tests - they should PASS now
go test ./internal/review/services/...
# Expected: PASS

# Verify build succeeds (CRITICAL)
go build -o /dev/null ./cmd/review

# Commit the implementation
git add internal/review/services/scan_service.go
git commit -m "feat(review): implement Scan Mode service (GREEN phase)

Implementation:
- NewScanService constructor
- AnalyzeScan method with Ollama integration
- Query validation
- Result caching

Testing:
- All 3 tests passing
- go build succeeds

Status: GREEN (tests pass, implementation complete)"

# Step 3: REFACTOR PHASE (if needed)
# Improve code quality while keeping tests green
# Example: Extract method, improve naming, add comments

# Run tests again - should still PASS
go test ./internal/review/services/...

# Commit refactoring
git commit -m "refactor(review): improve Scan Mode error handling"
```

**Go/Backend TDD Example:**
```go
// 1. RED: Write test FIRST (in scan_service_test.go)
func TestScanService_AnalyzeScan_Success(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewScanService(mockOllama, mockRepo)

	output, err := service.AnalyzeScan(ctx, 1, "auth", "owner", "repo")

	assert.NoError(t, err)
	assert.Len(t, output.Matches, 1)
}
// Run: FAILS (NewScanService undefined)

// 2. GREEN: Write minimal implementation (in scan_service.go)
func NewScanService(ollama *OllamaClient, repo AnalysisRepositoryInterface) *ScanService {
	return &ScanService{ollamaClient: ollama, analysisRepo: repo}
}

func (s *ScanService) AnalyzeScan(...) (*models.ScanModeOutput, error) {
	// Minimal implementation to pass tests
}
// Run: PASSES

// 3. REFACTOR: Improve (still in scan_service.go)
// Add better error handling, comments, validation
// Run: Still PASSES
```

**React/Frontend TDD Example:**
```javascript
// 1. RED: Write test FIRST
test('stores JWT token in localStorage with correct key', () => {
  const token = 'fake-jwt-token';
  authService.saveToken(token);
  expect(localStorage.getItem('devsmith_token')).toBe(token);
});
// Run: FAILS (authService.saveToken undefined)

// 2. GREEN: Write minimal code
export const authService = {
  saveToken: (token) => {
    localStorage.setItem('devsmith_token', token);
  }
};
// Run: PASSES

// 3. REFACTOR: Improve
export const authService = {
  saveToken: (token) => {
    if (!token) throw new Error('Token required');
    localStorage.setItem('devsmith_token', token);
  }
};
// Run: Still PASSES
```

**Why TDD is Mandatory:**
- Tests define requirements clearly (living documentation)
- Prevents over-engineering (write only needed code)
- Catches bugs early (before they reach production)
- Enables confident refactoring (tests protect against regressions)
- Aligns with platform mission (supervising AI, not trusting blindly)

### Step 4: Implement Feature 💻

Follow **[ARCHITECTURE.md Section 13: DevSmith Coding Standards](../ARCHITECTURE.md#devsmith-coding-standards)** exactly.

**Key Standards (See ARCHITECTURE.md for full details):**
- File organization: `apps/{service}-{frontend|backend}/`
- Naming: `PascalCase.jsx`, `camelCase.js`, `snake_case.py`
- React components: Follow standard template (ARCHITECTURE.md Section 13)
- API calls: Follow error handling pattern (ARCHITECTURE.md Section 13)
- Error handling: Never return error strings as data
- Configuration: No hardcoded values, everything in .env
- Testing: 70% unit coverage, 90% critical path coverage

**DO NOT duplicate standards here. Reference ARCHITECTURE.md Section 13.**

### Step 4.5: Verify Full Build (CRITICAL) 🔨

**BEFORE committing, you MUST verify the full service builds successfully.**

This step catches issues that tests alone miss:
- Code outside functions (copy-paste errors)
- Missing imports in main.go
- Type mismatches between packages
- Undefined variables/functions
- Syntax errors in wiring code

**Required Build Verification:**

```bash
# 1. Build the specific service you're working on
go build -o /dev/null ./cmd/{service}

# Examples:
go build -o /dev/null ./cmd/portal
go build -o /dev/null ./cmd/review
go build -o /dev/null ./cmd/logs
go build -o /dev/null ./cmd/analytics

# 2. If build succeeds, verify with golangci-lint
golangci-lint run ./cmd/{service}/...

# 3. Check for unused imports
goimports -l cmd/{service}/
```

**Common Build Errors to Watch For:**

❌ **Code Outside Functions**
```go
// WRONG - in cmd/portal/main.go
package main

fmt.Println("Starting...") // ❌ Code outside function!

func main() {
  // ...
}
```

✅ **Correct**
```go
package main

func main() {
  fmt.Println("Starting...") // ✅ Inside function
  // ...
}
```

❌ **Test Code in Production**
```go
// WRONG - in cmd/portal/main.go
func TestSomething(t *testing.T) { // ❌ Test code in main!
  // ...
}
```

✅ **Correct - Tests belong in *_test.go files**

❌ **Duplicate Definitions**
```go
// WRONG
type Config struct { // ❌ Already defined elsewhere
  Port int
}
```

**Pre-Commit Hook:**
Our pre-commit hook will automatically run these checks. If you see:
```
❌ Pre-commit validation FAILED
```
Fix the build errors before committing. DO NOT use `--no-verify`.

**Why This Matters:**
- Tests validate logic, but don't catch wiring/syntax errors
- Full build catches 90% of production errors before commit
- Prevents broken code from entering CI/CD pipeline
- Saves time by catching issues locally

### Step 5: Run Tests Locally 🧪

**Before creating PR, ALL must pass:**

```bash
# Frontend tests
cd apps/{service}-frontend
npm test
npm run test:coverage  # Must be >= 70%

# Backend tests
cd apps/{service}-backend
pytest
pytest --cov=. --cov-report=term-missing  # Must be >= 70%
```

### Step 6: Complete Manual Testing Checklist ✓

See **[ARCHITECTURE.md Section 13 - Manual Testing Checklist](../ARCHITECTURE.md#manual-testing-checklist)** for full list.

**Critical items:**
- [ ] Feature works in browser
- [ ] No console errors
- [ ] Regression check (related features still work)
- [ ] Works through gateway (http://localhost:3000)
- [ ] Authentication persists across apps
- [ ] No hardcoded URLs

### Step 7: Commit & Create PR 🚀

**Note:** Activity logging is automated via git hooks. Your commit message will automatically be logged to `.docs/devlog/copilot-activity.md` - no manual changelog updates needed!

```bash
# Commit with Conventional Commits format
# Include testing details and acceptance criteria in commit body
git add .
git commit -m "feat(auth): implement GitHub OAuth login

- Add OAuth endpoints to portal backend
- Create login component with OAuth button
- Store JWT in localStorage with correct key
- Redirect to dashboard after successful login

Testing:
- Unit tests: 85% coverage
- Integration test: login → dashboard flow passing
- Manual: Tested OAuth flow end-to-end

Acceptance Criteria:
- [x] User can login with GitHub OAuth
- [x] JWT stored in localStorage with key 'devsmith_token'
- [x] Token includes github_access_token field
- [x] Login redirects to portal dashboard

Closes #42"

# Push branch
git push origin feature/42-github-oauth-login

# Create PR using GitHub CLI (REQUIRED - DO THIS NOW!)
gh pr create \
  --title "Issue #042: GitHub OAuth Login" \
  --body "$(cat <<'PRBODY'
## Summary
Implements Issue #042: GitHub OAuth Login

## Changes
- Added GitHub OAuth endpoints
- JWT token storage
- Login flow integration

## Testing
- [x] All tests passing
- [x] Unit coverage: 85%+
- [x] Manual testing complete

Closes #42

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
PRBODY
)" \
  --base development \
  --head feature/42-github-oauth-login
```

**PR Description Template:**
```markdown
## Feature: GitHub OAuth Login

**Issue:** Closes #42

**Implementation:**
- Added GitHub OAuth endpoints to portal backend
- Created login component with OAuth integration
- JWT token stored with `github_access_token` field (not `github_token`)

**Testing:**
- [x] All automated tests pass
- [x] Unit test coverage: 85%
- [x] Integration test covers login → dashboard
- [x] Manual testing checklist complete
- [x] No hardcoded URLs
- [x] Works through gateway

**Acceptance Criteria:**
- [x] User can login with GitHub OAuth
- [x] JWT stored in localStorage with key 'devsmith_token'
- [x] Token includes github_access_token field
- [x] Login redirects to portal dashboard
- [x] All endpoints validate JWT
- [x] Unit tests >= 70% coverage
- [x] Integration test passing
- [x] No hardcoded URLs
- [x] User-friendly error messages

**Screenshots:**
[If UI changes, include before/after screenshots]
```

### Step 9: Address Code Review Feedback 🔄

When Claude reviews your PR:

1. **Read ALL comments carefully**
2. **Make requested changes**
3. **Push updates to same branch**
4. **Reply to comments** when fixed
5. **Request re-review**

**Don't:**
- Argue about standards (they're in ARCHITECTURE.md)
- Skip changes because "it works"
- Mark conversations resolved yourself
- Push without re-testing

---

## Critical Rules

### 1. Test-Driven Development (TDD) is REQUIRED

**RED → GREEN → REFACTOR cycle is mandatory for ALL features.**

- Tests written BEFORE implementation code
- Commit tests first (RED phase): `git commit -m "test: add failing tests (RED)"`
- Then commit implementation (GREEN phase): `git commit -m "feat: implement feature (GREEN)"`
- No exceptions - this is not negotiable
- If you write code first, Claude will reject PR immediately

**Why Separate Commits Matter:**
- RED commit proves tests were written first
- GREEN commit shows implementation driven by tests
- Git history validates TDD process
- Pre-commit hook detects RED phase and provides helpful guidance

### 2. One Feature Per Issue, One Issue Per PR

- Don't add "bonus" features
- Don't fix unrelated bugs
- Don't refactor unrelated code
- Stay focused on acceptance criteria

### 3. All Standards Are in ARCHITECTURE.md

- **DO NOT** guess at standards
- **DO** read [ARCHITECTURE.md Section 13](../ARCHITECTURE.md#devsmith-coding-standards)
- **DO** follow templates exactly
- **DO** ask Claude if unsure

### 4. Acceptance Criteria Are Gates

- Every criterion must be 100% met
- Partial implementations will be rejected
- "Almost done" is not done
- If you can't meet a criterion, ask Claude for guidance

### 5. No Hardcoded Values

**EVER. NO EXCEPTIONS.**

All URLs, ports, API keys go in environment variables.

See [ARCHITECTURE.md Section 13 - Configuration Management](../ARCHITECTURE.md#configuration-management).

#### 5.1 Special Case: Test Mocks Must Use Framework, Not Hardcoded Returns

**CRITICAL: Test mocks MUST participate in the testify mock framework. Never hardcode return values.**

**The Problem:**
When implementing mock methods, it's tempting to hardcode return values as placeholders. This bypasses the entire mock expectation system and makes tests impossible to configure.

**❌ WRONG - Hardcoded Mock (Breaks Tests)**
```go
// In testutils/mock_log_reader.go
func (m *MockLogReader) FindTopMessages(ctx context.Context, service string, level string, start time.Time, end time.Time, limit int) ([]models.IssueItem, error) {
    // Mock implementation
    return nil, nil  // ❌ HARDCODED - Ignores test expectations!
}
```

**Why This Breaks:**
- Test sets up expectations: `mockRepo.On("FindTopMessages", ...).Return(testData, nil)`
- Mock ignores expectations and always returns `nil, nil`
- Test receives empty results despite setup
- Developers waste hours debugging "why isn't my mock working?"
- **Real incident: #011 spent 20+ iterations trying to fix test before architect found hardcoded return**

**✅ CORRECT - Framework-Integrated Mock**
```go
// In testutils/mock_log_reader.go
func (m *MockLogReader) FindTopMessages(ctx context.Context, service string, level string, start time.Time, end time.Time, limit int) ([]models.IssueItem, error) {
    args := m.Called(ctx, service, level, start, end, limit)  // ✅ Uses framework
    return args.Get(0).([]models.IssueItem), args.Error(1)
}
```

**The Pattern: All Mock Methods Must:**
1. Call `m.Called(...)` with all parameters
2. Extract return values using `args.Get(N)` and `args.Error(N)`
3. Cast to correct types
4. Return the framework-provided values

**Complete Mock Template:**
```go
type MockRepository struct {
    mock.Mock
}

// Method with complex return type
func (m *MockRepository) FindData(ctx context.Context, id string, filters map[string]string) ([]models.Data, error) {
    args := m.Called(ctx, id, filters)
    if args.Get(0) == nil {
        return nil, args.Error(1)  // Handle nil case
    }
    return args.Get(0).([]models.Data), args.Error(1)
}

// Method with simple return
func (m *MockRepository) Count(ctx context.Context) (int, error) {
    args := m.Called(ctx)
    return args.Int(0), args.Error(1)  // Use typed getters when available
}

// Method with no error return
func (m *MockRepository) GetName() string {
    args := m.Called()
    return args.String(0)
}
```

**When Hardcoded Values ARE Acceptable:**

✅ **Test Input Data (Non-Mock)**
```go
// These are test fixtures, not mocks - hardcoding is fine
func TestService_ProcessUser(t *testing.T) {
    testUser := models.User{
        ID:    "test-123",           // ✅ OK - Test input
        Name:  "Test User",          // ✅ OK - Test input
        Email: "test@example.com",   // ✅ OK - Test input
    }

    result := service.ProcessUser(testUser)
    assert.Equal(t, "Processed: Test User", result)
}
```

✅ **Sentinel Errors and Constants**
```go
var (
    ErrNotFound = errors.New("not found")  // ✅ OK - Sentinel error
    ErrInvalid  = errors.New("invalid")    // ✅ OK - Sentinel error
)

const (
    MaxRetries = 3              // ✅ OK - Constant
    DefaultTimeout = time.Second * 30  // ✅ OK - Constant
)
```

✅ **Helper Functions (Non-Mock)**
```go
// ✅ OK - Helper that generates test data
func createTestContext() context.Context {
    return context.WithValue(context.Background(), "test", true)
}
```

**❌ Never Hardcode in Mocks:**
```go
// ❌ WRONG - Defeats entire purpose of mocking
func (m *MockRepo) Find(id string) (*Data, error) {
    if id == "test-123" {
        return &Data{Name: "hardcoded"}, nil  // ❌ Hardcoded logic!
    }
    return nil, errors.New("not found")
}

// ✅ CORRECT - Let test configure behavior
func (m *MockRepo) Find(id string) (*Data, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Data), args.Error(1)
}

// In test:
mockRepo.On("Find", "test-123").Return(&Data{Name: "configured"}, nil)
mockRepo.On("Find", "other").Return(nil, ErrNotFound)
```

**Pre-Commit Reminder:**
Before committing mock implementations, verify:
```bash
# 1. Check mock uses m.Called()
grep -n "m.Called" testutils/mock_*.go
# Should show usage in every mock method

# 2. Check for hardcoded returns in mocks
grep -n "return.*nil.*nil" testutils/mock_*.go
# Should NOT find "return nil, nil" outside error handling

# 3. Run tests to verify mocks work
go test ./...
```

**Why This Matters:**
- Hardcoded mocks waste developer time (hours debugging "broken" tests)
- Tests become unconfigurable and useless
- Creates false sense of test coverage (test doesn't actually test anything)
- Violates principle of explicit over implicit (test should show what's expected)

**Remember:** Mocks exist to let tests control behavior. Hardcoding defeats that purpose entirely.

---

## When to Ask Claude for Help

### Ask Claude BEFORE coding if:
- Acceptance criteria unclear
- Unsure which service should handle logic
- Unsure about database schema design
- Architectural decision needed
- Approach might violate modularity

### Ask Claude DURING coding if:
- Tests failing after 3 attempts (three-strikes rule)
- Not sure how to structure something
- Conflicting requirements in issue

### Example Good Questions:
```
Claude, issue #42 says "store token in localStorage" but also mentions
"secure storage". Should I use localStorage or something more secure?

Claude, where should GitHub API calls live - in portal-backend or a
shared service? The issue doesn't specify.

Claude, I've tried 3 different approaches to fix this WebSocket issue
and all failed. Can you help diagnose the root cause?
```

---

## Common Mistakes to Avoid

### ❌ DON'T:

1. **Write code before tests**
   - TDD is required, not optional

2. **Hardcode any values**
   ```javascript
   // ❌ WRONG
   const API_URL = 'http://localhost:8001';

   // ✅ RIGHT
   const API_URL = import.meta.env.VITE_API_URL;
   ```

3. **Return error strings as data**
   ```python
   # ❌ WRONG
   try:
       return process()
   except Exception as e:
       return f"Error: {e}"  # Looks like valid data!

   # ✅ RIGHT
   try:
       return process()
   except Exception as e:
       logger.error(f"Failed: {e}", exc_info=True)
       raise HTTPException(status_code=500, detail="Process failed")
   ```

4. **Skip manual testing checklist**
   - Automated tests aren't enough
   - Must verify in actual browser
   - Must check through gateway

5. **Skip testing details in commit message**
   - Include test coverage and results in commit body
   - Activity logging system extracts this automatically
   - No need for separate AI_CHANGELOG.md (automated)

6. **Implement multiple features in one PR**
   - One issue = one PR
   - No scope creep

7. **Skip documentation references**
   - Read ARCHITECTURE.md Section 13
   - Follow templates exactly
   - Don't guess

8. **Argue with code review feedback**
   - Standards are standards
   - If you disagree, discuss with Mike
   - Don't mark resolved without fixing

---

## Quick Reference

### File Naming
| Type | Format | Example |
|------|--------|---------|
| React Component | `PascalCase.jsx` | `LoginForm.jsx` |
| Utility | `camelCase.js` | `apiClient.js` |
| Style | `kebab-case.css` | `login-form.css` |
| Test | `Name.test.jsx` | `LoginForm.test.jsx` |
| Python | `snake_case.py` | `github_auth.py` |

See [ARCHITECTURE.md Section 13](../ARCHITECTURE.md#naming-conventions) for full details.

### Code Naming
| Type | Format | Example |
|------|--------|---------|
| Variable | camelCase / snake_case | `userData` / `user_data` |
| Constant | UPPER_SNAKE_CASE | `API_BASE_URL` |
| Function | camelCase / snake_case | `handleClick` / `handle_click` |
| Class/Component | PascalCase | `UserService`, `LoginForm` |

### Commit Types
| Type | Use For |
|------|---------|
| `feat:` | New feature |
| `fix:` | Bug fix |
| `docs:` | Documentation |
| `test:` | Tests only |
| `refactor:` | Code restructure |
| `style:` | Formatting |
| `chore:` | Maintenance |

### Test Coverage Requirements
| Type | Minimum |
|------|---------|
| Unit Tests | 70% |
| Critical Paths | 90% |

---

## Documentation You Must Read

**Before starting ANY feature:**
- **[ARCHITECTURE.md Section 13](../ARCHITECTURE.md#devsmith-coding-standards)** - Coding standards (REQUIRED)
- **[ARCHITECTURE.md Section 14](../ARCHITECTURE.md#development-workflow)** - Workflow process
- **[Requirements.md](../Requirements.md)** - Feature requirements
- **[DevsmithTDD.md](../DevsmithTDD.md)** - TDD approach and test cases

**When stuck:**
- **[LESSONS_LEARNED.md](../LESSONS_LEARNED.md)** - Common mistakes to avoid

**Templates to use:**
- ARCHITECTURE.md Section 13 - React Component Structure
- ARCHITECTURE.md Section 13 - API Call Pattern
- ARCHITECTURE.md Section 13 - Error Handling Requirements

---

## Parallel Development

**Multiple Copilot instances can work simultaneously:**
- Each in separate VS Code window
- Each on different feature branch
- Each implementing different issue
- No conflicts as long as features are isolated

**Coordination:**
- Claude creates issues
- Mike assigns issues to different instances
- Each instance works independently
- PRs reviewed and merged independently

**Example:**
```
VS Code Window 1: feature/42-github-oauth-login
VS Code Window 2: feature/43-logs-dashboard-ui
VS Code Window 3: feature/44-analytics-trends-api
```

All three can be in progress simultaneously.

---

## Success Checklist

Before creating PR, verify ALL of these:

- [ ] Read GitHub issue completely
- [ ] Wrote tests FIRST (TDD)
- [ ] **FULL SERVICE BUILD PASSES** (`go build ./cmd/{service}`) ⭐ CRITICAL
- [ ] golangci-lint passes (`golangci-lint run ./cmd/{service}/...`)
- [ ] No unused imports (`goimports -l cmd/{service}/`)
- [ ] All automated tests passing
- [ ] Test coverage >= 70% (unit) and 90% (critical paths)
- [ ] Manual testing checklist complete
- [ ] No console errors or warnings
- [ ] Works through gateway (http://localhost:3000)
- [ ] No hardcoded URLs, ports, or credentials
- [ ] All config in environment variables
- [ ] .env.example updated if new variables added
- [ ] Error messages are user-friendly
- [ ] Loading states present
- [ ] Follows file organization (ARCHITECTURE.md Section 13)
- [ ] Follows naming conventions (ARCHITECTURE.md Section 13)
- [ ] Commit message includes testing details and acceptance criteria
- [ ] Commit message follows Conventional Commits (activity logged automatically)
- [ ] PR description includes "Closes #XX"
- [ ] PR description lists all acceptance criteria with checkboxes
- [ ] Every acceptance criterion is met (100%)
- [ ] No scope creep (one feature only)

If any checkbox is unchecked, **DO NOT create PR yet.**

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-10-18 | Initial version with workflow updates |
| 1.1 | 2025-10-20 | Added automated activity logging via git hooks |
| 1.2 | 2025-10-20 | Updated branch workflow for auto-created branches |
| 1.3 | 2025-10-20 | **Major TDD Update**: Added comprehensive TDD workflow with Red-Green-Refactor cycle, complete examples, and mandatory separate commits for test and implementation phases |
| 1.4 | 2025-10-21 | **Mock Implementation Guidelines**: Added section 5.1 on responsible use of hardcoded values in test mocks, with examples of correct testify framework integration and common anti-patterns to avoid (addresses issue #011 mock implementation failure) |

---

**Remember:** You are the builder. Follow the issue, follow the standards in ARCHITECTURE.md, write tests first, and create quality PRs. Claude will review, but your job is to get it right the first time.
