# DevSmith Modular Platform: Test-Driven Development (TDD)

## Document Purpose

This TDD document ensures the DevSmith Modular Platform delivers on its core mission: **teaching developers to effectively read and understand code** through AI-assisted analysis in five distinct reading modes. Every test validates that the platform helps users develop the critical skill of supervising AI-generated code (Human in the Loop).

**Repository**: [github.com/mikejsmith1985/devsmith-modular-platform](https://github.com/mikejsmith1985/devsmith-modular-platform)

---

## Test-Driven Development Philosophy

### Core Principles

1. **Red → Green → Refactor**
   - Write failing test first (Red)
   - Implement minimal code to pass (Green)
   - Improve code quality while keeping tests green (Refactor)

2. **Tests as Living Documentation**
   - Tests explain what the system does
   - Tests validate requirements are met
   - Tests guide implementation

3. **Mental Models as Test Categories**
   - Tests organized by bounded context
   - Tests validate layering (controller, service, data)
   - Tests verify abstractions work correctly
   - Tests check scope boundaries

4. **Cognitive Load in Test Design**
   - Tests should be simple to understand (reduce extraneous load)
   - Tests should build understanding of system (maximize germane load)
   - Complex tests broken into smaller, focused tests

---

## TDD Workflow Best Practices (Reduce Iterations)

### Pre-RED Phase Checklist (Before Writing Tests)

**Problem:** Writing tests that fail due to structural issues (missing interfaces, wrong imports) wastes time.

**Solution:** Validate structure BEFORE writing tests.

#### Step 0: Structure Validation (2 minutes, saves 20+ minutes)

```bash
# 1. Verify package structure exists
ls -la internal/review/services/

# 2. Check if shared interfaces exist
grep -l "interface" internal/review/services/*.go

# 3. If interfaces scattered, consolidate FIRST:
# Create: internal/review/services/interfaces.go
# Move all shared interfaces there
# Remove duplicates from service files

# 4. Verify imports compile
go build ./internal/review/services/...
# ✅ MUST pass before writing tests

# 5. Run gofmt and goimports
gofmt -w internal/review/services/
goimports -w internal/review/services/
```

**Pre-RED Checklist:**
- [ ] Package directory exists
- [ ] Shared interfaces in `interfaces.go` (one location only)
- [ ] No duplicate type/interface definitions
- [ ] `go build` succeeds (even with empty functions)
- [ ] Imports formatted with `goimports`
- [ ] Package declaration matches directory name

**Why This Matters:**
- ❌ **Without checklist:** Write test → import error → fix import → redeclaration error → fix duplicate → finally test runs = 30 minutes
- ✅ **With checklist:** Validate structure → write test → test fails as expected = 5 minutes

---

### RED Phase: Write Failing Tests

**Common Pitfalls (Go-Specific):**

1. **Import Path Errors**
   ```go
   // ❌ WRONG: Relative imports
   import "../models"

   // ✅ CORRECT: Absolute from module root
   import "devsmith/internal/review/models"
   ```

2. **Duplicate Interface Definitions**
   ```go
   // ❌ WRONG: Interface in skim_service.go AND scan_service.go
   type OllamaClientInterface interface { ... }

   // ✅ CORRECT: Interface in interfaces.go (one place)
   // Both services import from interfaces.go
   ```

3. **Missing Package Declaration**
   ```go
   // ❌ WRONG: No package at top
   import "testing"

   // ✅ CORRECT: Package first
   package services

   import "testing"
   ```

4. **Copy-Paste Corruption**
   ```go
   // ❌ WRONG: Copied from markdown with invisible characters
   import​ "context"  // ← invisible Unicode character

   // ✅ CORRECT: Type it fresh or use IDE auto-complete
   import "context"
   ```

**RED Phase Workflow:**
```bash
# 1. Create test file
touch internal/review/services/scan_service_test.go

# 2. Write test (use IDE autocomplete for imports)
# Let IDE suggest imports instead of typing them

# 3. Run test - SHOULD FAIL with "undefined: NewScanService"
go test ./internal/review/services/... -v -run TestScanService
# Expected: FAIL - function doesn't exist yet

# 4. Commit RED phase
git add internal/review/services/scan_service_test.go
git commit -m "test(review): add failing test for ScanService (RED phase)"
```

**RED Phase Commit:** Tests exist, they fail, that's expected.

---

### GREEN Phase: Minimal Implementation

**GREEN Phase Workflow:**
```bash
# 1. Create implementation file
touch internal/review/services/scan_service.go

# 2. Add minimal code to pass test
# - Package declaration
# - Import interfaces from interfaces.go (NOT redefine)
# - Struct definition
# - Constructor
# - Method stubs (return nil or zero values)

# 3. Build BEFORE running tests
go build ./internal/review/services/...
# ✅ MUST pass (catches syntax errors)

# 4. Run tests - SHOULD PASS
go test ./internal/review/services/... -v
# Expected: PASS - minimal implementation works

# 5. Commit GREEN phase
git add internal/review/services/scan_service.go
git commit -m "feat(review): implement ScanService (GREEN phase)"
```

**GREEN Phase Commit:** Tests pass, implementation is minimal but correct.

---

### REFACTOR Phase: Improve Quality

**REFACTOR Phase Workflow:**
```bash
# 1. Improve code quality (while keeping tests green)
# - Extract constants
# - Add error handling
# - Improve naming
# - Add documentation comments

# 2. Run tests after EACH change
go test ./internal/review/services/... -v
# ✅ MUST stay green

# 3. Format code
gofmt -w internal/review/services/
goimports -w internal/review/services/

# 4. Commit REFACTOR phase
git add internal/review/services/scan_service.go
git commit -m "refactor(review): improve ScanService error handling and docs"
```

**REFACTOR Phase Commit:** Tests still pass, code is cleaner.

---

### Anti-Patterns (What NOT to Do)

1. **Don't skip `go build` before tests**
   ```bash
   # ❌ WRONG: Jump straight to tests
   go test ./internal/review/services/...
   # Gets cryptic errors about missing imports

   # ✅ CORRECT: Build first to catch syntax errors
   go build ./internal/review/services/...
   go test ./internal/review/services/...
   ```

2. **Don't define interfaces in service files**
   ```go
   // ❌ WRONG: scan_service.go
   type OllamaClientInterface interface { ... }
   type ScanService struct { ... }

   // ✅ CORRECT: interfaces.go
   type OllamaClientInterface interface { ... }

   // scan_service.go just imports it
   type ScanService struct {
       ollamaClient OllamaClientInterface // From interfaces.go
   }
   ```

3. **Don't commit without running tests**
   ```bash
   # ❌ WRONG: Commit without verification
   git add . && git commit -m "fix: stuff"

   # ✅ CORRECT: Test before commit
   go test ./... && git add . && git commit -m "fix: resolve import errors"
   ```

4. **Don't mix multiple phases in one commit**
   ```bash
   # ❌ WRONG: One commit with tests + implementation
   git commit -m "add ScanService with tests"

   # ✅ CORRECT: Three commits (RED → GREEN → REFACTOR)
   git commit -m "test(review): add failing test (RED)"
   git commit -m "feat(review): implement ScanService (GREEN)"
   git commit -m "refactor(review): improve error handling"
   ```

---

### Quick Reference: TDD Phases

| Phase | Action | Expected Outcome | Commit Message |
|-------|--------|------------------|----------------|
| **PRE-RED** | Validate structure | `go build` passes | N/A (no commit) |
| **RED** | Write failing test | Test fails (function undefined) | `test: add failing test (RED)` |
| **GREEN** | Minimal implementation | Test passes | `feat: implement feature (GREEN)` |
| **REFACTOR** | Improve quality | Tests still pass | `refactor: improve code quality` |

---

### Time Savings

**Without Pre-RED Validation:**
- Write test (5 min)
- Import error (5 min to fix)
- Redeclaration error (10 min to fix)
- Build error (5 min to fix)
- Test finally runs (25 min total)

**With Pre-RED Validation:**
- Validate structure (2 min)
- Write test (5 min)
- Test runs immediately (7 min total)

**Time Saved:** 18 minutes per test file × 16 issues = **4.8 hours saved**

---

## Test Framework & Tools

### Backend (Go)
- **Unit Tests**: Go's built-in `testing` package
- **Mocking**: `testify/mock` for interfaces
- **Database Tests**: `dockertest` for PostgreSQL integration tests
- **HTTP Tests**: `httptest` for handler testing
- **Coverage**: `go test -cover` (target: 70%+ unit, 90%+ critical path)

### Frontend (Templ + HTMX)
- **Template Tests**: Templ compile-time validation
- **Integration Tests**: Playwright for browser automation
- **Visual Regression**: Percy or Chromatic (Phase 2)
- **Accessibility**: axe-core automated checks

### AI Integration
- **Ollama Mocking**: Mock responses for deterministic tests
- **Prompt Validation**: Verify prompts contain required context
- **Response Parsing**: Validate JSON structure from AI

### End-to-End
- **Framework**: Playwright
- **Browsers**: Chrome, Firefox, Safari
- **Test Environment**: Docker Compose with test database

---

## Test Coverage Targets

| Component | Unit Tests | Integration Tests | E2E Tests |
|-----------|------------|-------------------|-----------|
| Portal Service | 70%+ | Critical paths | Login flow |
| Review Service | 70%+ | All 5 modes | All 5 modes |
| Logging Service | 70%+ | WebSocket flow | Real-time logs |
| Analytics Service | 70%+ | Query aggregation | Report generation |
| Build Service (P2) | 70%+ | OpenHands integration | Terminal session |

**Critical Path Coverage**: 90%+ (authentication, Review app modes, data persistence)

---

## Test Organization

### Directory Structure
```
apps/
├── portal/
│   ├── handlers/
│   │   ├── auth_handler.go
│   │   └── auth_handler_test.go
│   ├── services/
│   │   ├── auth_service.go
│   │   └── auth_service_test.go
│   └── db/
│       ├── users.go
│       └── users_test.go
├── review/
│   ├── handlers/
│   ├── services/
│   │   ├── review_ai_service.go
│   │   └── review_ai_service_test.go  # Critical: 5 modes
│   └── db/
└── ... (other services)

tests/
├── integration/
│   ├── portal_auth_test.go
│   ├── review_modes_test.go         # Critical: 5 modes end-to-end
│   └── logging_websocket_test.go
└── e2e/
    ├── playwright.config.ts
    ├── auth.spec.ts
    ├── review_preview_mode.spec.ts
    ├── review_skim_mode.spec.ts
    ├── review_scan_mode.spec.ts
    ├── review_detailed_mode.spec.ts
    ├── review_critical_mode.spec.ts  # Most important test
    └── ...
```

---

## Test Categories by Mental Model

### 1. Bounded Context Tests

**Purpose**: Verify entities have single, clear meaning within their context

**Portal Context Tests**:
```go
// Test: User in Portal context means authenticated identity
func TestPortalUser_HasAuthenticationFields(t *testing.T) {
    user := &models.PortalUser{
        GitHubID:   12345,
        Username:   "testuser",
        AvatarURL:  "https://...",
    }

    assert.NotZero(t, user.GitHubID, "Portal User must have GitHub ID")
    assert.NotEmpty(t, user.Username, "Portal User must have username")
}
```

**Review Context Tests**:
```go
// Test: User in Review context means code reviewer
func TestReviewUser_HasReviewFields(t *testing.T) {
    user := &models.ReviewUser{
        UserID:          1,
        ReviewsCreated:  5,
        IssuesIdentified: 23,
    }

    assert.NotZero(t, user.ReviewsCreated, "Review User tracks review count")
    // Note: No GitHubID field - that's Portal's concern
}
```

**Cross-Context Violation Test**:
```go
// Test: Portal User should NOT appear in Review service
func TestReviewService_DoesNotImportPortalModels(t *testing.T) {
    // Static analysis or import check
    // This test ensures bounded context boundaries are respected
}
```

---

### 2. Layering Tests

**Purpose**: Verify clean separation of controller, service, and data layers

**Controller Layer Test** (HTTP only, no business logic):
```go
func TestAuthHandler_GitHubCallback_ValidatesInput(t *testing.T) {
    // Arrange
    mockService := new(MockAuthService)
    handler := handlers.NewAuthHandler(mockService)

    req := httptest.NewRequest("GET", "/auth/github/callback?code=", nil)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req

    // Act
    handler.GitHubCallback(c)

    // Assert
    assert.Equal(t, http.StatusBadRequest, w.Code, "Handler should validate code param")
    assert.Contains(t, w.Body.String(), "error", "Should return error JSON")
    mockService.AssertNotCalled(t, "Authenticate") // Service not called with invalid input
}
```

**Service Layer Test** (Business logic, no HTTP):
```go
func TestAuthService_Authenticate_ValidCode(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    mockGitHub := new(MockGitHubClient)
    service := services.NewAuthService(mockRepo, mockGitHub)

    mockGitHub.On("ExchangeCode", mock.Anything, "valid_code").
        Return(&github.AccessToken{Token: "token123"}, nil)
    mockGitHub.On("GetUser", mock.Anything, "token123").
        Return(&github.User{ID: 12345, Login: "testuser"}, nil)
    mockRepo.On("FindOrCreateByGitHubID", mock.Anything, 12345, "testuser").
        Return(&models.User{ID: 1, GitHubID: 12345}, nil)

    // Act
    user, token, err := service.Authenticate(context.Background(), "valid_code")

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 12345, user.GitHubID)
    assert.NotEmpty(t, token, "Should return JWT token")
    mockRepo.AssertExpectations(t)
}
```

**Data Layer Test** (SQL only, no business logic):
```go
func TestUserRepository_FindOrCreateByGitHubID_CreatesNewUser(t *testing.T) {
    // Arrange
    db := setupTestDB(t)
    defer teardownTestDB(t, db)
    repo := db.NewUserRepository(db)

    // Act
    user, err := repo.FindOrCreateByGitHubID(context.Background(), 99999, "newuser")

    // Assert
    assert.NoError(t, err)
    assert.NotZero(t, user.ID, "Should assign ID")
    assert.Equal(t, 99999, user.GitHubID)
    assert.Equal(t, "newuser", user.Username)

    // Verify in DB
    var count int
    db.QueryRow("SELECT COUNT(*) FROM portal.users WHERE github_id = $1", 99999).Scan(&count)
    assert.Equal(t, 1, count, "User should be in database")
}
```

**Layer Violation Test**:
```go
func TestAuthHandler_DoesNotCallRepository(t *testing.T) {
    // This is a design test - handlers should never import db package
    // Enforced via static analysis or import checks
}
```

---

### 3. Abstraction Tests

**Purpose**: Verify interfaces work and implementations are swappable

**Interface Definition**:
```go
// interfaces/auth_provider.go
type AuthProvider interface {
    Authenticate(ctx context.Context, code string) (*User, string, error)
    ValidateToken(ctx context.Context, token string) (*User, error)
}
```

**Test Against Interface** (not concrete type):
```go
func TestGitHubAuthProvider_ImplementsAuthProvider(t *testing.T) {
    var _ interfaces.AuthProvider = (*services.GitHubAuthProvider)(nil)
    // Compile-time check that GitHubAuthProvider implements interface
}

func TestMockAuthProvider_CanReplaceReal(t *testing.T) {
    // Test that mock provider works in handler
    mockProvider := new(MockAuthProvider)
    handler := handlers.NewAuthHandler(mockProvider) // Takes interface, not concrete

    mockProvider.On("Authenticate", mock.Anything, "test_code").
        Return(&models.User{ID: 1}, "jwt_token", nil)

    // Test handler with mock provider
    // ...
}
```

**Swappable Implementation Test**:
```go
// Future: If we add GitLabAuthProvider, this test ensures it works
func TestAuthHandler_WorksWithDifferentProviders(t *testing.T) {
    providers := []interfaces.AuthProvider{
        services.NewGitHubAuthProvider(...),
        // services.NewGitLabAuthProvider(...), // Future
    }

    for _, provider := range providers {
        handler := handlers.NewAuthHandler(provider)
        // Test handler works with any provider
    }
}
```

---

### 4. Scope Tests

**Purpose**: Verify variables have minimal, appropriate scope

**Function Scope Test**:
```go
func TestReviewService_AnalyzeInMode_NoLeakedVariables(t *testing.T) {
    // Test that temp variables don't leak outside function
    service := setupReviewService(t)

    // Call function multiple times
    _, err1 := service.AnalyzeInMode(ctx, code, ModePreview, opts)
    _, err2 := service.AnalyzeInMode(ctx, code, ModeSkim, opts)

    // Each call should be independent (no shared state)
    assert.NoError(t, err1)
    assert.NoError(t, err2)
    // Calls don't affect each other
}
```

**No Global State Test**:
```go
func TestReviewService_ThreadSafe(t *testing.T) {
    service := setupReviewService(t)

    // Run multiple goroutines calling service
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _, err := service.AnalyzeInMode(ctx, code, ModePreview, opts)
            assert.NoError(t, err)
        }()
    }
    wg.Wait()
    // No race conditions or shared mutable state
}
```

---

## Core Feature Tests

### 1. Authentication (Portal Service)

#### Test 1.1: GitHub OAuth - Complete Flow
```go
func TestAuth_GitHubOAuthFlow_Success(t *testing.T) {
    // GIVEN: Valid GitHub OAuth setup
    config := loadTestConfig(t)
    server := setupTestServer(t, config)
    defer server.Close()

    // WHEN: User initiates GitHub login
    resp, err := http.Get(server.URL + "/auth/github/login")
    require.NoError(t, err)

    // THEN: Redirect to GitHub OAuth page
    assert.Equal(t, http.StatusFound, resp.StatusCode)
    location := resp.Header.Get("Location")
    assert.Contains(t, location, "github.com/login/oauth/authorize")
    assert.Contains(t, location, "client_id="+config.GitHubClientID)

    // WHEN: GitHub redirects back with code
    callbackResp, err := http.Get(server.URL + "/auth/github/callback?code=test_code")
    require.NoError(t, err)

    // THEN: User authenticated with JWT token
    assert.Equal(t, http.StatusOK, callbackResp.StatusCode)
    var result map[string]interface{}
    json.NewDecoder(callbackResp.Body).Decode(&result)
    assert.NotEmpty(t, result["token"], "Should return JWT token")
    assert.NotNil(t, result["user"], "Should return user object")
}
```

#### Test 1.2: Unauthorized Access Blocked
```go
func TestAuth_UnauthorizedAccess_Redirects(t *testing.T) {
    server := setupTestServer(t, config)

    // WHEN: Accessing protected endpoint without token
    resp, _ := http.Get(server.URL + "/api/apps")

    // THEN: Redirect to login
    assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
    var result map[string]string
    json.NewDecoder(resp.Body).Decode(&result)
    assert.Equal(t, "Please log in with GitHub", result["error"])
}
```

#### Test 1.3: JWT Token Validation
```go
func TestAuth_JWTValidation_ValidToken(t *testing.T) {
    // GIVEN: Valid JWT token
    token := generateTestJWT(t, &models.User{ID: 1, Username: "testuser"})

    // WHEN: Making request with token
    req, _ := http.NewRequest("GET", "/api/auth/me", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    resp := makeRequest(t, req)

    // THEN: User info returned
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    var user models.User
    json.NewDecoder(resp.Body).Decode(&user)
    assert.Equal(t, "testuser", user.Username)
}
```

---

### 2. Review Service - Five Reading Modes (CRITICAL)

**These are the most important tests in the system - the core value proposition**

#### Test 2.1: Preview Mode - Quick Structural Assessment
```go
func TestReviewAI_PreviewMode_ReturnsStructure(t *testing.T) {
    // GIVEN: Sample Go codebase
    codebase := loadTestCodebase(t, "testdata/sample_go_project")
    service := setupReviewAIService(t)

    // WHEN: Analyzing in Preview mode
    result, err := service.AnalyzeInMode(
        context.Background(),
        codebase,
        models.ModePreview,
        models.AnalysisOptions{},
    )

    // THEN: Returns high-level structure
    require.NoError(t, err)
    assert.NotNil(t, result.FileStructure, "Must return file structure")
    assert.NotEmpty(t, result.BoundedContexts, "Must identify bounded contexts")
    assert.NotEmpty(t, result.TechStack, "Must detect tech stack")
    assert.Contains(t, result.TechStack, "Go", "Should detect Go")
    assert.Contains(t, result.ArchitecturePattern, "layered", "Should identify layering")

    // THEN: Should NOT contain implementation details
    assert.Empty(t, result.FunctionImplementations, "Preview doesn't show implementations")

    // THEN: Cognitive load managed
    assert.Less(t, len(result.Summary), 500, "Summary should be brief (reduce intrinsic load)")
}
```

#### Test 2.2: Skim Mode - Abstractions Only
```go
func TestReviewAI_SkimMode_ReturnsAbstractions(t *testing.T) {
    // GIVEN: Go service with interfaces and implementations
    code := `
    package services

    type UserService interface {
        GetUser(ctx context.Context, id int) (*User, error)
        CreateUser(ctx context.Context, user *User) error
    }

    type userServiceImpl struct {
        repo UserRepository
    }

    func (s *userServiceImpl) GetUser(ctx context.Context, id int) (*User, error) {
        // ... 50 lines of implementation ...
    }
    `

    // WHEN: Analyzing in Skim mode
    result, err := service.AnalyzeInMode(ctx, code, models.ModeSkim, opts)

    // THEN: Returns function signatures, not implementations
    require.NoError(t, err)
    assert.Len(t, result.Functions, 2, "Should list both functions")
    assert.Equal(t, "GetUser", result.Functions[0].Name)
    assert.Contains(t, result.Functions[0].Signature, "GetUser(ctx context.Context, id int)")
    assert.NotEmpty(t, result.Functions[0].Description, "Should describe what it does")

    // THEN: Implementation details not included
    assert.Empty(t, result.Functions[0].ImplementationLines, "Skim mode skips implementation")

    // THEN: Interfaces identified
    assert.Len(t, result.Interfaces, 1, "Should identify UserService interface")
    assert.Equal(t, "UserService", result.Interfaces[0].Name)
}
```

#### Test 2.3: Scan Mode - Targeted Search
```go
func TestReviewAI_ScanMode_FindsSpecificCode(t *testing.T) {
    // GIVEN: Codebase with authentication logic
    codebase := loadTestCodebase(t, "testdata/sample_project")

    // WHEN: Scanning for authentication validation
    result, err := service.AnalyzeInMode(ctx, codebase, models.ModeScan, models.AnalysisOptions{
        ScanQuery: "Where is authentication validated?",
    })

    // THEN: Finds relevant code sections
    require.NoError(t, err)
    assert.NotEmpty(t, result.Matches, "Should find authentication code")

    // THEN: Each match has context
    for _, match := range result.Matches {
        assert.NotEmpty(t, match.FilePath, "Should specify file")
        assert.NotZero(t, match.LineNumber, "Should specify line")
        assert.NotEmpty(t, match.Context, "Should provide surrounding context")
        assert.NotEmpty(t, match.Explanation, "Should explain why it matches")
    }

    // THEN: Semantic search works (not just keyword)
    foundValidation := false
    for _, match := range result.Matches {
        if strings.Contains(match.Code, "ValidateToken") ||
           strings.Contains(match.Code, "jwt.Parse") {
            foundValidation = true
        }
    }
    assert.True(t, foundValidation, "Should find token validation even without 'authentication' keyword")
}
```

#### Test 2.4: Detailed Mode - Line-by-Line Explanation
```go
func TestReviewAI_DetailedMode_ExplainsAlgorithm(t *testing.T) {
    // GIVEN: Complex algorithm
    code := `
    func BinarySearch(arr []int, target int) int {
        left, right := 0, len(arr)-1
        for left <= right {
            mid := left + (right-left)/2
            if arr[mid] == target {
                return mid
            } else if arr[mid] < target {
                left = mid + 1
            } else {
                right = mid - 1
            }
        }
        return -1
    }
    `

    // WHEN: Analyzing in Detailed mode
    result, err := service.AnalyzeInMode(ctx, code, models.ModeDetailed, models.AnalysisOptions{
        TargetPath: "binary_search.go:BinarySearch",
    })

    // THEN: Line-by-line explanation provided
    require.NoError(t, err)
    assert.NotEmpty(t, result.LineExplanations, "Should explain each significant line")

    // THEN: Variable states tracked
    assert.Contains(t, result.LineExplanations[2].Explanation, "left=0, right=", "Should show variable states")

    // THEN: Algorithm identified
    assert.Contains(t, result.AlgorithmSummary, "binary search", "Should recognize algorithm")
    assert.Contains(t, result.Complexity, "O(log n)", "Should analyze complexity")

    // THEN: Edge cases noted
    assert.NotEmpty(t, result.EdgeCases, "Should identify edge cases")
    assert.Contains(t, result.EdgeCases[0], "empty array", "Should note empty array case")
}
```

#### Test 2.5: Critical Mode - Quality Evaluation (MOST IMPORTANT)
```go
func TestReviewAI_CriticalMode_IdentifiesIssues(t *testing.T) {
    // GIVEN: Code with multiple issues
    code := `
    package handlers

    import "database/sql"

    var db *sql.DB // Global variable (scope issue)

    func GetUser(w http.ResponseWriter, r *http.Request) {
        id := r.URL.Query().Get("id")

        // SQL injection vulnerability
        query := "SELECT * FROM users WHERE id = " + id
        rows, _ := db.Query(query) // Error ignored

        // Missing input validation
        // No bounded context check
        // Handler calling database directly (layer violation)
    }
    `

    // WHEN: Analyzing in Critical mode
    result, err := service.AnalyzeInMode(ctx, code, models.ModeCritical, opts)

    // THEN: Architecture issues identified
    require.NoError(t, err)
    architectureIssues := filterIssuesByType(result.Issues, "architecture")
    assert.NotEmpty(t, architectureIssues, "Should find layer violation")
    assert.Contains(t, architectureIssues[0].Description, "Handler should not call database directly")

    // THEN: Security issues identified
    securityIssues := filterIssuesByType(result.Issues, "security")
    assert.NotEmpty(t, securityIssues, "Should find SQL injection")
    assert.Equal(t, "critical", securityIssues[0].Severity, "SQL injection is critical")
    assert.Contains(t, securityIssues[0].Description, "SQL injection")
    assert.NotEmpty(t, securityIssues[0].SuggestedFix, "Should provide fix")
    assert.Contains(t, securityIssues[0].SuggestedFix, "parameterized query")

    // THEN: Code quality issues identified
    qualityIssues := filterIssuesByType(result.Issues, "quality")
    assert.NotEmpty(t, qualityIssues, "Should find error handling issue")
    assert.Contains(t, qualityIssues[0].Description, "Error ignored")

    // THEN: Scope issues identified
    scopeIssues := filterIssuesByType(result.Issues, "scope")
    assert.NotEmpty(t, scopeIssues, "Should find global variable issue")
    assert.Contains(t, scopeIssues[0].Description, "Global variable")

    // THEN: All issues have location and fix
    for _, issue := range result.Issues {
        assert.NotEmpty(t, issue.FilePath, "Issue must have file path")
        assert.NotZero(t, issue.LineNumber, "Issue must have line number")
        assert.NotEmpty(t, issue.Description, "Issue must have description")
        assert.NotEmpty(t, issue.SuggestedFix, "Issue must have suggested fix")
        assert.NotEmpty(t, issue.Severity, "Issue must have severity")
    }
}
```

#### Test 2.6: Mode Transitions
```go
func TestReviewUI_ModeTransitions_Fluid(t *testing.T) {
    // E2E test with Playwright
    page := setupBrowser(t)

    // GIVEN: User uploads code
    page.Goto(baseURL + "/review/sessions/new")
    page.Fill("#code-input", sampleCode)
    page.Click("#create-session")

    // WHEN: Starting in Preview mode
    page.SelectOption("#reading-mode", "preview")
    page.Click("#analyze")
    page.WaitForSelector(".ai-result")

    // THEN: Can transition to Skim
    page.Click("#go-deeper")  // Preview → Skim transition
    page.WaitForSelector(".function-list")

    // THEN: Can transition to Detailed on specific function
    page.Click(".function-item:first-child") // Skim → Detailed
    page.WaitForSelector(".line-by-line")

    // THEN: Can transition to Scan
    page.Click("#find-usages")  // Detailed → Scan
    page.WaitForSelector(".search-results")

    // THEN: Can transition to Critical from any mode
    page.Click("#review-this")  // Any → Critical
    page.WaitForSelector(".issue-list")

    // All transitions work without errors
}
```

---

### 3. Logging Service - Real-Time Monitoring

#### Test 3.1: WebSocket Log Streaming
```go
func TestLogging_WebSocketStream_RealTime(t *testing.T) {
    // GIVEN: Logging service running
    server := setupTestServer(t)
    wsURL := "ws://localhost:3003/ws/logs"

    // WHEN: Client connects to WebSocket
    conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
    require.NoError(t, err)
    defer conn.Close()

    // WHEN: Log entry created
    logEntry := models.LogEntry{
        Service: "portal",
        Level:   "error",
        Message: "Test error message",
    }
    createLogEntry(t, server, logEntry)

    // THEN: Client receives log within 1 second
    conn.SetReadDeadline(time.Now().Add(1 * time.Second))
    var receivedLog models.LogEntry
    err = conn.ReadJSON(&receivedLog)
    require.NoError(t, err)

    assert.Equal(t, "portal", receivedLog.Service)
    assert.Equal(t, "error", receivedLog.Level)
    assert.Equal(t, "Test error message", receivedLog.Message)
    assert.NotZero(t, receivedLog.CreatedAt, "Should have timestamp")
}
```

#### Test 3.2: Log Filtering
```go
func TestLogging_QueryLogs_FilteredByService(t *testing.T) {
    // GIVEN: Logs from multiple services
    createTestLogs(t, []models.LogEntry{
        {Service: "portal", Level: "info", Message: "Portal log"},
        {Service: "review", Level: "error", Message: "Review error"},
        {Service: "portal", Level: "error", Message: "Portal error"},
    })

    // WHEN: Querying for portal logs only
    resp := makeRequest(t, "GET", "/api/logs?service=portal", nil)

    // THEN: Only portal logs returned
    var logs []models.LogEntry
    json.NewDecoder(resp.Body).Decode(&logs)
    assert.Len(t, logs, 2, "Should return 2 portal logs")
    for _, log := range logs {
        assert.Equal(t, "portal", log.Service)
    }
}
```

---

### 4. Analytics Service - Pattern Detection

#### Test 4.1: Trend Analysis
```go
func TestAnalytics_TrendAnalysis_DetectsIncrease(t *testing.T) {
    // GIVEN: Error logs increasing over time
    baseTime := time.Now().Add(-24 * time.Hour)
    for hour := 0; hour < 24; hour++ {
        errorCount := hour / 4  // Increasing trend
        for i := 0; i < errorCount; i++ {
            createLogEntry(t, models.LogEntry{
                Service:   "review",
                Level:     "error",
                Message:   "AI analysis timeout",
                CreatedAt: baseTime.Add(time.Duration(hour) * time.Hour),
            })
        }
    }

    // WHEN: Running trend analysis
    resp := makeRequest(t, "GET", "/api/analytics/trends?metric=error_rate&service=review", nil)

    // THEN: Increasing trend detected
    var result models.TrendAnalysis
    json.NewDecoder(resp.Body).Decode(&result)
    assert.Equal(t, "increasing", result.Direction)
    assert.Greater(t, result.ChangePercent, 50.0, "Should show significant increase")
}
```

---

### 5. Integration Tests - Cross-Service Flows

#### Test 5.1: Review Session Logged to Logging Service
```go
func TestIntegration_ReviewSession_LogsActivity(t *testing.T) {
    // GIVEN: User authenticated
    token := authenticateTestUser(t)

    // WHEN: Creating review session
    reviewResp := makeAuthenticatedRequest(t, "POST", "/api/review/sessions", token, map[string]interface{}{
        "code_source": "paste",
        "pasted_code": "package main\nfunc main() {}",
        "title":       "Test Review",
    })
    assert.Equal(t, http.StatusOK, reviewResp.StatusCode)

    var session models.ReviewSession
    json.NewDecoder(reviewResp.Body).Decode(&session)

    // WHEN: Running AI analysis in Critical mode
    analysisResp := makeAuthenticatedRequest(t, "POST",
        fmt.Sprintf("/api/review/sessions/%d/analyze", session.ID),
        token,
        map[string]interface{}{
            "reading_mode": "critical",
        },
    )
    assert.Equal(t, http.StatusOK, analysisResp.StatusCode)

    // THEN: Activity logged to Logging service
    time.Sleep(100 * time.Millisecond) // Allow async logging
    logsResp := makeRequest(t, "GET", "/api/logs?service=review", nil)

    var logs []models.LogEntry
    json.NewDecoder(logsResp.Body).Decode(&logs)

    // Find the AI analysis log
    found := false
    for _, log := range logs {
        if strings.Contains(log.Message, "AI analysis completed") &&
           strings.Contains(log.Message, "critical") {
            found = true
            assert.Equal(t, "info", log.Level)
            metadata := log.Metadata.(map[string]interface{})
            assert.Equal(t, float64(session.ID), metadata["session_id"])
            assert.Equal(t, "critical", metadata["reading_mode"])
        }
    }
    assert.True(t, found, "AI analysis should be logged")
}
```

---

## Enhanced Pre-commit Hook Tests (Phase 2)

### Overview
Tests for the enhanced pre-commit validation system that integrates with Logging and Analytics services. These tests ensure AI agents and developers receive intelligent, actionable feedback on code quality issues.

### Test Category: Pre-commit Hook Core Features

#### Test 1: JSON Output Generation
```bash
# test/hooks/test_json_output.sh
test_json_output_contains_all_fields() {
    # GIVEN: Code with known issues
    echo 'package test\nimport "unused"\nfunc Test() {}' > test.go
    git add test.go

    # WHEN: Running hook with --json flag
    output=$(.git/hooks/pre-commit --json 2>&1 || true)

    # THEN: Output is valid JSON with required fields
    echo "$output" | jq -e '.status' > /dev/null
    echo "$output" | jq -e '.duration' > /dev/null
    echo "$output" | jq -e '.issues' > /dev/null
    echo "$output" | jq -e '.grouped' > /dev/null
    echo "$output" | jq -e '.summary' > /dev/null

    assert_equal "$(echo "$output" | jq -r '.status')" "failed"
    assert_greater_than "$(echo "$output" | jq '.issues | length')" "0"
}
```

#### Test 2: Issue Prioritization
```bash
test_issues_grouped_by_priority() {
    # GIVEN: Code with high, medium, and low priority issues
    cat > test_file.go <<EOF
package test
import "unused"  // Medium: unused import
func Test() {
    undefined_func()  // High: build error
}
// Missing comment  // Low: style issue
type ExportedType struct{}
EOF
    git add test_file.go

    # WHEN: Running validation
    output=$(.git/hooks/pre-commit --json 2>&1 || true)

    # THEN: Issues are properly grouped
    high_count=$(echo "$output" | jq '.grouped.high | length')
    medium_count=$(echo "$output" | jq '.grouped.medium | length')
    low_count=$(echo "$output" | jq '.grouped.low | length')

    assert_greater_than "$high_count" "0"  # Build errors
    assert_greater_than "$medium_count" "0"  # Unused import
    assert_greater_than "$low_count" "0"  # Style issues
}
```

#### Test 3: Context Extraction
```bash
test_context_includes_surrounding_lines() {
    # GIVEN: File with error on line 10
    for i in {1..20}; do echo "// Line $i" >> context_test.go; done
    sed -i '10s/.*/undefined_func()  \/\/ Error here/' context_test.go
    git add context_test.go

    # WHEN: Running validation
    output=$(.git/hooks/pre-commit --json 2>&1 || true)

    # THEN: Context includes ±3 lines around error
    context=$(echo "$output" | jq -r '.issues[0].context')
    echo "$context" | grep -q "Line 7"
    echo "$context" | grep -q "Line 10"
    echo "$context" | grep -q "Line 13"
}
```

#### Test 4: Auto-fix Mode
```bash
test_auto_fix_corrects_formatting() {
    # GIVEN: Unformatted Go code
    cat > unformatted.go <<EOF
package test
import( "fmt"
"strings" )
func Test( ){fmt.Println("test")}
EOF
    git add unformatted.go

    # WHEN: Running with --fix flag
    .git/hooks/pre-commit --fix

    # THEN: Code is properly formatted
    go fmt unformatted.go
    diff <(cat unformatted.go) <(gofmt unformatted.go)
    assert_equal "$?" "0"
}
```

#### Test 5: Parallel Execution Performance
```bash
test_parallel_execution_faster_than_sequential() {
    # GIVEN: Repository with multiple Go files
    for i in {1..10}; do
        echo "package test$i" > "test$i.go"
        git add "test$i.go"
    done

    # WHEN: Running in standard mode (parallel)
    start=$(date +%s)
    .git/hooks/pre-commit 2>&1 || true
    parallel_duration=$(($(date +%s) - start))

    # AND: Simulating sequential execution
    start=$(date +%s)
    go fmt ./... && go vet ./... && golangci-lint run ./... && go test ./...
    sequential_duration=$(($(date +%s) - start))

    # THEN: Parallel is at least 2x faster
    assert_greater_than "$sequential_duration" "$((parallel_duration * 2))"
}
```

#### Test 6: Dependency Graph Generation
```bash
test_dependency_graph_shows_fix_order() {
    # GIVEN: Code with build error (blocks tests)
    cat > broken.go <<EOF
package test
func Test() {
    undefined_func()  // Build error
}
EOF
    git add broken.go

    # WHEN: Running validation
    output=$(.git/hooks/pre-commit --json 2>&1 || true)

    # THEN: Dependency graph shows correct fix order
    fix_order=$(echo "$output" | jq -r '.dependencyGraph.fix_order[]')
    echo "$fix_order" | head -1 | grep -q "build_errors"
    echo "$fix_order" | tail -1 | grep -q "style"
}
```

#### Test 7: Progressive Validation Modes
```bash
test_quick_mode_faster_than_standard() {
    # GIVEN: Repository with test files
    create_test_files

    # WHEN: Running in quick mode
    start=$(date +%s)
    .git/hooks/pre-commit --quick 2>&1 || true
    quick_duration=$(($(date +%s) - start))

    # AND: Running in standard mode
    start=$(date +%s)
    .git/hooks/pre-commit 2>&1 || true
    standard_duration=$(($(date +%s) - start))

    # THEN: Quick mode is significantly faster
    assert_less_than "$quick_duration" "$((standard_duration / 2))"
}
```

#### Test 8: Interactive Explain Mode
```bash
test_explain_mode_provides_test_details() {
    # GIVEN: Failing test
    cat > failing_test.go <<EOF
package test
import "testing"
func TestExample(t *testing.T) {
    t.Error("This test fails")
}
EOF
    git add failing_test.go

    # WHEN: Running explain mode
    output=$(.git/hooks/pre-commit --explain TestExample 2>&1 || true)

    # THEN: Output contains test-specific details
    echo "$output" | grep -q "TestExample"
    echo "$output" | grep -q "Issue:"
    echo "$output" | grep -q "Suggestion:"
}
```

#### Test 9: LSP Output Format
```bash
test_lsp_output_valid_format() {
    # GIVEN: Code with linting issues
    cat > lsp_test.go <<EOF
package test
func unexported() {}  // Missing comment
EOF
    git add lsp_test.go

    # WHEN: Generating LSP output
    output=$(.git/hooks/pre-commit --output-lsp 2>&1 || true)

    # THEN: Output is valid LSP diagnostic format
    echo "$output" | jq -e '.[].uri' > /dev/null
    echo "$output" | jq -e '.[].range.start.line' > /dev/null
    echo "$output" | jq -e '.[].severity' > /dev/null
    echo "$output" | jq -e '.[].message' > /dev/null

    # AND: File URIs are properly formatted
    uri=$(echo "$output" | jq -r '.[0].uri')
    echo "$uri" | grep -q "^file://"
}
```

#### Test 10: Agent Guide Integration
```bash
test_agent_guide_provides_fix_steps() {
    # GIVEN: Agent guide exists with mock setup pattern
    test -f .git/hooks/pre-commit-agent-guide.json
    grep -q "missing_mock_setup" .git/hooks/pre-commit-agent-guide.json

    # WHEN: Encountering mock expectation failure
    cat > mock_test.go <<EOF
package test
import "testing"
func TestMock(t *testing.T) {
    // Mock expectations not met error
    t.Error("0 out of 5 expectation(s) were met")
}
EOF
    git add mock_test.go
    output=$(.git/hooks/pre-commit --json 2>&1 || true)

    # THEN: Suggestion includes guide steps
    suggestion=$(echo "$output" | jq -r '.issues[0].suggestion')
    echo "$suggestion" | grep -q "Mock.On()"
    echo "$suggestion" | grep -q ".docs/copilot-instructions.md"
}
```

### Test Category: Logging Service Integration

#### Test 11: Validation Results Ingestion
```go
// internal/logs/handlers/validation_handler_test.go
func TestSubmitValidationResults(t *testing.T) {
    // GIVEN: Validation results from pre-commit hook
    validationData := models.ValidationRun{
        UserID:     1,
        Repository: "devsmith-modular-platform",
        Branch:     "feature/test",
        CommitSHA:  "abc123",
        Mode:       "standard",
        Duration:   45,
        Status:     "failed",
        IssuesJSON: `{"total": 25, "errors": 2}`,
    }

    // WHEN: Submitting via API
    req := httptest.NewRequest("POST", "/api/logs/validation", toJSON(validationData))
    w := httptest.NewRecorder()
    handler.SubmitValidation(w, req)

    // THEN: Results are stored
    assert.Equal(t, http.StatusCreated, w.Code)

    // AND: Can be retrieved
    stored, err := repo.FindValidationByID(1)
    assert.NoError(t, err)
    assert.Equal(t, "feature/test", stored.Branch)
    assert.Equal(t, 2, stored.ErrorCount())
}
```

#### Test 12: Validation History Query
```go
func TestGetValidationHistory(t *testing.T) {
    // GIVEN: Multiple validation runs for a user
    createValidationRuns(userID, 10)

    // WHEN: Querying history
    req := httptest.NewRequest("GET", "/api/logs/validation/history?user_id=1&limit=5", nil)
    w := httptest.NewRecorder()
    handler.GetValidationHistory(w, req)

    // THEN: Returns most recent runs
    assert.Equal(t, http.StatusOK, w.Code)
    var results []models.ValidationRun
    json.Unmarshal(w.Body.Bytes(), &results)
    assert.Equal(t, 5, len(results))

    // AND: Results are sorted by created_at DESC
    assert.True(t, results[0].CreatedAt.After(results[1].CreatedAt))
}
```

#### Test 13: WebSocket Validation Streaming
```go
func TestValidationWebSocketStream(t *testing.T) {
    // GIVEN: WebSocket connection established
    conn := connectToWebSocket("/ws/logs/validation")

    // WHEN: New validation run is submitted
    submitValidation(models.ValidationRun{
        Status: "failed",
        IssuesJSON: `{"total": 5}`,
    })

    // THEN: Event is broadcast via WebSocket
    msg, err := conn.ReadMessage()
    assert.NoError(t, err)

    var event ValidationEvent
    json.Unmarshal(msg, &event)
    assert.Equal(t, "validation_completed", event.Type)
    assert.Equal(t, "failed", event.Data.Status)
}
```

### Test Category: Analytics Service Integration

#### Test 14: Validation Trend Analysis
```go
// internal/analytics/services/validation_analytics_test.go
func TestValidationTrendAnalysis(t *testing.T) {
    // GIVEN: Validation runs over 7 days
    for day := 0; day < 7; day++ {
        date := time.Now().AddDate(0, 0, -day)
        createValidationRuns(date, 10, 8) // 10 total, 8 passed
    }

    // WHEN: Analyzing trends
    trends, err := service.GetValidationTrends(7)

    // THEN: Returns daily pass rates
    assert.NoError(t, err)
    assert.Equal(t, 7, len(trends))
    assert.Equal(t, 80.0, trends[0].PassRate) // 8/10 = 80%
}
```

#### Test 15: Top Issues Aggregation
```go
func TestTopIssuesAggregation(t *testing.T) {
    // GIVEN: Validation runs with various issue types
    createValidationWithIssues("missing_mock_setup", 25)
    createValidationWithIssues("missing_godoc", 15)
    createValidationWithIssues("unused_import", 10)

    // WHEN: Querying top issues
    topIssues, err := service.GetTopIssues(5)

    // THEN: Returns most common issues
    assert.NoError(t, err)
    assert.Equal(t, "missing_mock_setup", topIssues[0].IssueType)
    assert.Equal(t, 25, topIssues[0].Count)
}
```

#### Test 16: Auto-fix Effectiveness Rate
```go
func TestAutoFixEffectivenessRate(t *testing.T) {
    // GIVEN: Validations with auto-fixable issues
    createValidation(ValidationRun{
        IssuesJSON: `{"total": 20, "autoFixable": 15}`,
    })

    // WHEN: Calculating auto-fix rate
    rate, err := service.GetAutoFixRate()

    // THEN: Returns correct percentage
    assert.NoError(t, err)
    assert.Equal(t, 75.0, rate) // 15/20 = 75%
}
```

#### Test 17: Agent Fix Success Rate
```go
func TestAgentFixSuccessRate(t *testing.T) {
    // GIVEN: Validations with agent attribution
    createAgentValidation("OpenHands", "passed", 8)
    createAgentValidation("OpenHands", "failed", 2)

    // WHEN: Calculating agent success rate
    rate, err := service.GetAgentSuccessRate("OpenHands")

    // THEN: Returns correct rate
    assert.NoError(t, err)
    assert.Equal(t, 80.0, rate) // 8/10 = 80%
}
```

### Test Category: Portal Dashboard Integration

#### Test 18: Validation Dashboard Widget
```go
// internal/portal/handlers/dashboard_test.go
func TestValidationDashboardWidget(t *testing.T) {
    // GIVEN: User with validation history
    createValidationRuns(userID, 10, 7) // 10 total, 7 passed

    // WHEN: Loading dashboard
    req := httptest.NewRequest("GET", "/dashboard", nil)
    req = addAuth(req, userID)
    w := httptest.NewRecorder()
    handler.Dashboard(w, req)

    // THEN: Dashboard shows validation stats
    body := w.Body.String()
    assert.Contains(t, body, "Validation Pass Rate")
    assert.Contains(t, body, "70%") // 7/10
    assert.Contains(t, body, "Recent Validations")
}
```

### Test Category: End-to-End Workflows

#### Test 19: OpenHands Integration Workflow
```bash
# tests/e2e/openhands_validation.sh
test_openhands_validation_workflow() {
    # GIVEN: OpenHands implementing a feature
    # (Simulated - actual OpenHands integration in Phase 2)

    # WHEN: Running pre-commit validation
    output=$(.git/hooks/pre-commit --json 2>&1 || true)

    # THEN: JSON is parseable by agent
    issues=$(echo "$output" | jq '.issues')
    assert_not_empty "$issues"

    # AND: Auto-fix handles simple issues
    .git/hooks/pre-commit --fix
    output2=$(.git/hooks/pre-commit --json 2>&1 || true)
    count1=$(echo "$output" | jq '.issues | length')
    count2=$(echo "$output2" | jq '.issues | length')
    assert_less_than "$count2" "$count1"

    # AND: Remaining issues have fix guidance
    for issue in $(echo "$output2" | jq -c '.issues[]'); do
        suggestion=$(echo "$issue" | jq -r '.suggestion')
        assert_not_empty "$suggestion"
    done
}
```

#### Test 20: Developer Workflow with Dashboard
```bash
test_developer_sees_validation_stats() {
    # GIVEN: Developer has made multiple commits
    for i in {1..5}; do
        create_commit_with_issues $i
        .git/hooks/pre-commit --json | curl -X POST http://localhost:3003/api/logs/validation -d @-
    done

    # WHEN: Viewing Portal dashboard
    response=$(curl http://localhost:3000/dashboard -H "Cookie: session=$SESSION")

    # THEN: Dashboard shows validation trends
    echo "$response" | grep -q "Validation Statistics"
    echo "$response" | grep -q "Pass Rate"
    echo "$response" | grep -q "Top Issues"
}
```

### Performance Requirements

#### Perf Test 1: JSON Output Generation Speed
```go
func BenchmarkJSONOutputGeneration(b *testing.B) {
    // GIVEN: Pre-commit results with 100 issues
    issues := generateIssues(100)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // WHEN: Generating JSON output
        output := generateJSONOutput(issues)

        // THEN: Completes in <10ms
        _ = output
    }
    // Target: <10ms per call
}
```

#### Perf Test 2: Parallel Execution Scaling
```bash
test_parallel_execution_scales_linearly() {
    # GIVEN: Increasing numbers of Go files
    for count in 10 50 100 200; do
        create_test_files $count

        # WHEN: Running validation
        start=$(date +%s)
        .git/hooks/pre-commit 2>&1 || true
        duration=$(($(date +%s) - start))

        # THEN: Duration scales sub-linearly (due to parallelization)
        # Record: count=$count, duration=$duration
        echo "$count,$duration" >> scaling_results.csv
    done

    # Verify: 200 files takes <2x time of 100 files
}
```

### Acceptance Criteria

All tests must pass before marking Phase 2 complete:

**Core Features** (Tests 1-10):
- [ ] JSON output is valid and contains all required fields
- [ ] Issues are correctly grouped by priority
- [ ] Code context is extracted for all issues
- [ ] Auto-fix corrects at least 60% of common issues
- [ ] Parallel execution is 2x+ faster than sequential
- [ ] Dependency graph shows correct fix order
- [ ] All three validation modes work correctly
- [ ] Interactive modes provide useful information
- [ ] LSP output format is valid
- [ ] Agent guide provides actionable fix steps

**Service Integration** (Tests 11-18):
- [ ] Logging service ingests validation results
- [ ] Validation history is queryable
- [ ] WebSocket streaming works
- [ ] Analytics calculates correct trends
- [ ] Top issues are identified
- [ ] Auto-fix and agent success rates are tracked
- [ ] Portal dashboard displays validation stats

**End-to-End** (Tests 19-20):
- [ ] OpenHands workflow completes successfully
- [ ] Developer can view stats in dashboard

**Performance** (Perf Tests 1-2):
- [ ] JSON generation completes in <10ms
- [ ] Parallel execution scales sub-linearly

---

## End-to-End User Workflows

### E2E 1: New User Onboarding
```typescript
// tests/e2e/onboarding.spec.ts
test('New user can complete full review workflow', async ({ page }) => {
  // GIVEN: User visits platform for first time
  await page.goto('http://localhost:3000');

  // WHEN: Clicking login
  await page.click('text=Login with GitHub');

  // THEN: Redirected to GitHub OAuth (mocked in test)
  await expect(page).toHaveURL(/github.com\/login/);

  // WHEN: GitHub redirects back (simulated)
  await page.goto('http://localhost:3000/auth/github/callback?code=test_code');

  // THEN: Lands on portal dashboard
  await expect(page).toHaveURL('http://localhost:3000/portal');
  await expect(page.locator('.welcome-message')).toBeVisible();

  // WHEN: Clicking Review app
  await page.click('text=Code Review');

  // THEN: Review app opens
  await expect(page).toHaveURL(/\/review/);

  // WHEN: Pasting code and selecting Critical mode
  await page.fill('#code-input', `
    func GetUser(id string) (*User, error) {
      query := "SELECT * FROM users WHERE id = " + id
      // SQL injection vulnerability
    }
  `);
  await page.selectOption('#reading-mode', 'critical');
  await page.click('text=Analyze Code');

  // THEN: AI identifies SQL injection
  await expect(page.locator('.issue-list')).toBeVisible();
  await expect(page.locator('.issue-item')).toContainText('SQL injection');
  await expect(page.locator('.severity-critical')).toBeVisible();

  // WHEN: Clicking suggested fix
  await page.click('.show-fix');

  // THEN: Fix shown with before/after
  await expect(page.locator('.suggested-fix')).toContainText('parameterized query');

  // User has successfully learned to identify critical issue!
});
```

### E2E 2: Developer Reviews OpenHands Output
```typescript
test('Developer reviews OpenHands PR in Critical mode', async ({ page, context }) => {
  // GIVEN: OpenHands created PR
  const prCode = await fetchPRCode('feature/user-auth'); // Simulated

  await page.goto('http://localhost:3000/review/sessions/new');

  // WHEN: Loading code from GitHub PR
  await page.selectOption('#code-source', 'github');
  await page.fill('#github-pr', 'mikejsmith1985/devsmith-platform/pull/42');
  await page.click('text=Load PR');

  // THEN: Code loaded
  await expect(page.locator('.code-display')).toBeVisible();

  // WHEN: Running Critical review
  await page.selectOption('#reading-mode', 'critical');
  await page.click('text=Review Code');

  // THEN: Issues categorized by type
  await expect(page.locator('.architecture-issues')).toBeVisible();
  await expect(page.locator('.security-issues')).toBeVisible();
  await expect(page.locator('.quality-issues')).toBeVisible();

  // WHEN: Accepting a suggested fix
  await page.click('.issue-item:first-child .accept-fix');

  // THEN: Fix applied to code
  await expect(page.locator('.code-display')).toContainText('// Fixed:');

  // WHEN: Generating PR comment
  await page.click('text=Generate PR Comment');

  // THEN: Comment formatted for GitHub
  await expect(page.locator('.pr-comment-preview')).toContainText('## Code Review');
  await expect(page.locator('.pr-comment-preview')).toContainText('- [ ] Architecture');

  // Developer successfully performed Human-in-the-Loop review!
});
```

---

## Performance Tests

### Perf 1: AI Analysis Response Times
```go
func TestPerformance_PreviewMode_Under3Seconds(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test in short mode")
    }

    codebase := loadTestCodebase(t, "testdata/medium_project") // ~1000 lines
    service := setupReviewAIService(t)

    start := time.Now()
    _, err := service.AnalyzeInMode(ctx, codebase, models.ModePreview, opts)
    duration := time.Since(start)

    assert.NoError(t, err)
    assert.Less(t, duration, 3*time.Second, "Preview mode must complete in <3s")
}

func TestPerformance_CriticalMode_Under30Seconds(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test in short mode")
    }

    code := loadTestCode(t, "testdata/handler_500_lines.go")
    service := setupReviewAIService(t)

    start := time.Now()
    _, err := service.AnalyzeInMode(ctx, code, models.ModeCritical, opts)
    duration := time.Since(start)

    assert.NoError(t, err)
    assert.Less(t, duration, 30*time.Second, "Critical mode must complete in <30s for 500 lines")
}
```

### Perf 2: WebSocket Latency
```go
func TestPerformance_WebSocket_Under100ms(t *testing.T) {
    conn := setupWebSocket(t)
    defer conn.Close()

    // Send log entry
    entry := models.LogEntry{Message: "test"}
    sendTime := time.Now()
    sendLogEntry(t, entry)

    // Receive via WebSocket
    var received models.LogEntry
    conn.ReadJSON(&received)
    latency := time.Since(sendTime)

    assert.Less(t, latency, 100*time.Millisecond, "WebSocket latency must be <100ms")
}
```

---

## Test Execution Strategy

### Local Development
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./apps/review/services/...

# Run integration tests only
go test -tags=integration ./tests/integration/...

# Skip slow tests
go test -short ./...
```

### CI/CD Pipeline (GitHub Actions)
```yaml
jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test -cover ./...
      - run: go test -coverprofile=coverage.out ./...
      - run: go tool cover -html=coverage.out -o coverage.html
      - uses: actions/upload-artifact@v3
        with:
          name: coverage-report
          path: coverage.html

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
      redis:
        image: redis:7
    steps:
      - uses: actions/checkout@v3
      - run: go test -tags=integration ./tests/integration/...

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: docker-compose up -d
      - uses: actions/setup-node@v3
      - run: npx playwright install
      - run: npx playwright test
```

---

## Test Maintenance

### Red Flags (Tests Need Attention)
- ❌ Test passes when code is wrong (false positive)
- ❌ Test fails intermittently (flaky test)
- ❌ Test takes >10 seconds (slow test)
- ❌ Test requires manual setup (not automated)
- ❌ Test doesn't match requirements (stale test)

### When to Update Tests
1. **Requirements Change**: Update tests first (TDD)
2. **Bug Found**: Write failing test, then fix
3. **Refactoring**: Tests should still pass (if not, tests are too coupled to implementation)
4. **New Feature**: Write tests before implementation

### Test Review Checklist
- [ ] Test name clearly describes what is tested
- [ ] Test has GIVEN/WHEN/THEN structure
- [ ] Test tests one thing (atomic)
- [ ] Test is independent (no order dependency)
- [ ] Test uses meaningful assertions
- [ ] Test cleans up resources
- [ ] Test is fast (<1s for unit, <10s for integration)

---

## Success Criteria

The platform's TDD approach is successful when:

### For Users (Learning Outcomes)
- ✅ User can identify critical issues in AI-generated code (Critical mode)
- ✅ User understands bounded contexts after using Preview mode
- ✅ User can navigate codebase confidently after Skim mode
- ✅ User finds bugs faster with Scan mode
- ✅ User comprehends complex algorithms with Detailed mode

### For Development (Quality Metrics)
- ✅ 70%+ unit test coverage
- ✅ 90%+ critical path coverage (5 reading modes, auth, logging)
- ✅ Zero flaky tests in CI
- ✅ All E2E tests pass on every commit
- ✅ PRs cannot merge without passing tests

### For Platform (Reliability Metrics)
- ✅ All reading modes complete within performance targets
- ✅ WebSocket logs delivered <100ms latency
- ✅ No SQL injection vulnerabilities (tested)
- ✅ No layer violations (tested)
- ✅ Bounded contexts respected (tested)

---

## Appendix: Test Data

### Sample Code for Testing Review Modes
Located in `testdata/`:
- `simple_go_handler.go` - Basic handler (100 lines)
- `medium_service.go` - Service with interfaces (500 lines)
- `complex_algorithm.go` - Algorithm for Detailed mode
- `vulnerable_code.go` - Code with multiple issues for Critical mode
- `sample_project/` - Complete project for Preview/Skim modes

### Mock Ollama Responses
Located in `tests/mocks/ollama_responses.json`:
- Responses for each reading mode
- Ensures deterministic tests
- Updated when prompt templates change

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-18 | Claude | Initial TDD document (React/Node stack) |
| 2.0 | 2025-10-18 | Claude | Complete rewrite for Go+Templ+HTMX, mental models, 5 reading modes centerpiece |

---

## References
- Requirements.md - Complete platform requirements
- ARCHITECTURE.md - System design and mental models
- DevSmithRoles.md - Team roles and workflows
- .docs/specs/TEMPLATE.md - Implementation spec template
