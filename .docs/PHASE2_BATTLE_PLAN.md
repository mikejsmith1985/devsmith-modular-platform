# Phase 2: Complexity Refactoring - Battle Plan

**Status**: READY TO START  
**Branch**: Review-App-Beta-Ready  
**Completion Target**: All gocognit violations <20, nestif <5, funlen compliance  
**Estimated Duration**: 2-3 hours for full resolution

---

## 1. Critical Violations (MUST FIX FIRST)

### VIOLATION #1: bindCodeRequest - Cognitive Complexity 48 (Limit 20)

**File**: `apps/review/handlers/ui_handler.go`  
**Function**: `bindCodeRequest` (line 68)  
**Severity**: 🔴 CRITICAL (2.4x over limit)  

**Current Structure** (~50 lines):
```go
func (h *UIHandler) bindCodeRequest(c *gin.Context, targetHandler string) (*review_models.CodeRequest, error) {
    var req review_models.CodeRequest
    
    // Multiple conditional branches checking different input types:
    // 1. Check if paste form data exists
    // 2. Check if upload exists
    // 3. Check if GitHub URL provided
    // 4. Validate and parse each
    // 5. Error handling for each path
    // 6. Deep nesting of if-else chains
}
```

**Extraction Plan**:
```
bindCodeRequest (complexity 48)
├─ extractPastedCode() → complexity 10
├─ extractUploadedFile() → complexity 12
├─ extractGitHubSource() → complexity 10
└─ validateCodeRequest() → complexity 5
Result: Max function complexity drops to ~8
```

**Implementation Steps**:
1. Extract `extractPastedCode()` - handles form data parsing
2. Extract `extractUploadedFile()` - handles multipart file parsing
3. Extract `extractGitHubSource()` - handles GitHub URL parsing
4. Extract `validateCodeRequest()` - validates final request
5. Refactor `bindCodeRequest()` to call these in sequence

**Expected Reduction**: 48 → 8 ✅

---

### VIOLATION #2: ui_handler.go Line 72 - Nested Block Complexity 33 (Limit 5)

**File**: `apps/review/handlers/ui_handler.go`  
**Location**: Line 72 (if err != nil block)  
**Severity**: 🔴 CRITICAL (6.6x over limit)  

**Current Structure**:
```go
if err != nil {
    // Complex error handling with multiple nested conditions:
    // 1. Check error type
    // 2. Log different error details
    // 3. Return different responses for each error
    // Multiple levels of nesting
}
```

**Extraction Plan**:
```
Line 72 if err != nil block (complexity 33)
├─ handleValidationError(err) → handle validation issues
├─ handleBindError(err) → handle binding failures
├─ handleQueryError(err) → handle query parsing errors
└─ handleGenericError(err) → default error handler
Result: Max nesting complexity drops to ~2
```

**Implementation Steps**:
1. Extract error type checking to `classifyError(err error) string`
2. Extract error logging to `logError(err error, context string)`
3. Extract response generation to `errorResponse(err error) (status int, body interface{})`
4. Simplify main block to: classify → log → respond

**Expected Reduction**: 33 → 3 ✅

---

## 2. High-Priority Violations (SHOULD FIX)

### VIOLATION #3: HandleGitHubOAuthCallbackWithSession - 79 Lines (Limit 50)

**File**: `apps/portal/handlers/auth_handler.go`  
**Function**: `HandleGitHubOAuthCallbackWithSession` (line 1023)  
**Severity**: 🟠 HIGH (1.6x over limit)  

**Current Structure** (~79 lines):
```go
func (h *AuthHandler) HandleGitHubOAuthCallbackWithSession(...) {
    // 1. Extract code from query
    // 2. Exchange code for token
    // 3. Get user info from GitHub
    // 4. Create/update user in DB
    // 5. Create session
    // 6. Return success
}
```

**Extraction Plan**:
```
HandleGitHubOAuthCallbackWithSession (79 lines)
├─ extractOAuthCode(c *gin.Context) → 5 lines
├─ exchangeCodeForToken(code string) → 8 lines
├─ fetchGitHubUser(token string) → 8 lines
├─ createOrUpdateUser(ghUser *GitHubUser) → 12 lines
├─ createUserSession(user *User) → 10 lines
└─ sendSuccessResponse(c, user, session) → 8 lines
Result: Main function ~12 lines, each helper <15 lines
```

**Implementation Steps**:
1. Extract OAuth code extraction to separate function
2. Extract token exchange to `exchangeCodeForToken()`
3. Extract GitHub user fetch to `fetchGitHubUser()`
4. Extract user creation to `createOrUpdateUser()`
5. Extract session creation to `createUserSession()`
6. Extract response building to `sendSuccessResponse()`

**Expected Reduction**: 79 → ~15 ✅

---

### VIOLATION #4: HandleTokenExchange - 108 Lines (Limit 100)

**File**: `apps/portal/handlers/auth_handler.go`  
**Function**: `HandleTokenExchange` (line 547)  
**Severity**: 🟠 MEDIUM (1.08x over limit)  

**Current Structure** (~108 lines):
```go
func (h *AuthHandler) HandleTokenExchange(...) {
    // 1. Parse request body
    // 2. Validate token
    // 3. Get user from context
    // 4. Refresh token if needed
    // 5. Return new token
}
```

**Extraction Plan**:
```
HandleTokenExchange (108 lines)
├─ parseTokenRequest(c) → 10 lines
├─ validateAndParseJWT(token) → 12 lines
├─ refreshTokenIfNeeded(token) → 15 lines
└─ sendTokenResponse(c, token) → 8 lines
Result: Main function ~10 lines, each helper <20 lines
```

**Expected Reduction**: 108 → ~12 ✅

---

### VIOLATION #5: HandleTestLogin - 105 Lines (Limit 100)

**File**: `apps/portal/handlers/auth_handler.go`  
**Function**: `HandleTestLogin` (line 218)  
**Severity**: 🟠 MEDIUM (1.05x over limit)  

**Extraction Plan**:
```
HandleTestLogin (105 lines)
├─ parseTestUserRequest(c) → 10 lines
├─ createTestUser(req) → 25 lines
├─ generateTestJWT(user) → 10 lines
└─ sendTestAuthResponse(c, user, token) → 8 lines
Result: Main function ~10 lines, each helper <30 lines
```

**Expected Reduction**: 105 → ~12 ✅

---

## 3. Medium-Priority Violations (NICE TO FIX)

### HandleScanMode - Cognitive Complexity 23 (Limit 20)
- **Current**: 23 (just over)
- **Extract**: AI analysis logic to service method
- **Reduction**: 23 → ~15

### renderError - Cognitive Complexity 23 (Limit 20)
- **Current**: 23 (just over)
- **Extract**: Error type checking to helper function
- **Reduction**: 23 → ~12

### TestConnection - Cognitive Complexity 24 (Limit 20)
- **Current**: 24
- **Extract**: Connection testing logic
- **Reduction**: 24 → ~15

---

## 4. Minor Violations (OPTIONAL)

- **goconst**: Magic strings duplicated (low priority)
- **gocritic**: paramTypeCombine, typeDefFirst (style fixes)
- **errcheck**: Response body not closed (usually warning)
- **errorlint**: Missing %w in fmt.Errorf (auto-fixable)

---

## 5. Execution Strategy

### PHASE 2A: CRITICAL FIXES (Highest Impact)
**Estimated Time**: 90 minutes  
**Expected Result**: Eliminate 2 critical violations (bindCodeRequest, line 72 nested block)

1. **Extract bindCodeRequest helpers** (20 min)
   - Create: `extractPastedCode()`, `extractUploadedFile()`, `extractGitHubSource()`
   - Verify: Each helper <20 complexity
   - Test: Unit tests for each extraction

2. **Simplify error handling at line 72** (20 min)
   - Create: Error classification and response helpers
   - Verify: Nesting reduced to ~3 levels
   - Test: Error handling paths

3. **Verify Compilation & Tests** (10 min)
   - Run: `go build ./apps/review/handlers`
   - Run: `go test -short ./apps/review/handlers`
   - Verify: No new errors introduced

4. **Verify gocognit Reduction** (10 min)
   - Run: `golangci-lint run --disable-all -E gocognit`
   - Verify: bindCodeRequest < 20, nested blocks < 5

5. **Update Commit** (10 min)
   - Commit Phase 2A with detailed message

---

### PHASE 2B: HIGH-PRIORITY FIXES (Function Extraction)
**Estimated Time**: 60 minutes  
**Expected Result**: Fix 3 funlen violations

1. **Refactor HandleGitHubOAuthCallbackWithSession** (20 min)
2. **Refactor HandleTokenExchange** (15 min)
3. **Refactor HandleTestLogin** (15 min)
4. **Verify & Commit** (10 min)

---

### PHASE 2C: MEDIUM-PRIORITY FIXES (Final Polish)
**Estimated Time**: 30 minutes  
**Expected Result**: Bring remaining gocognit violations under 20

1. **HandleScanMode extraction** (10 min)
2. **renderError refactoring** (10 min)
3. **TestConnection simplification** (5 min)
4. **Final Verification** (5 min)

---

## 6. Success Criteria

✅ **Phase 2A Complete When**:
- bindCodeRequest complexity < 20
- Line 72 nested block complexity < 5
- All helpers properly tested
- Compilation passes

✅ **Phase 2B Complete When**:
- All 3 funlen violations resolved (< limits)
- Each function properly decomposed
- All extraction tested

✅ **Phase 2C Complete When**:
- All gocognit violations < 20
- All nestif violations < 5
- All funlen violations within limits
- Pre-push gate: linting PASSES

---

## 7. Validation Checkpoints

**After Phase 2A**:
```bash
golangci-lint run --disable-all -E gocognit,gocritic | grep -E "bindCodeRequest|ui_handler.go:72"
# Expected: No results (violations cleared)
```

**After Phase 2B**:
```bash
golangci-lint run --disable-all -E funlen | grep -E "HandleGitHub|HandleToken|HandleTestLogin"
# Expected: No results
```

**After Phase 2C**:
```bash
bash scripts/hooks/pre-push 2>&1 | grep -E "^error|failed"
# Expected: No errors (all linting passes)
```

---

## 8. Risk Mitigation

**Risk**: Breaking existing functionality during refactoring
- **Mitigation**: Keep comprehensive unit tests passing throughout
- **Check**: `go test -short ./apps/portal/handlers ./apps/review/handlers` after each major change

**Risk**: Introducing new linting errors while fixing others
- **Mitigation**: Use targeted linting checks between phases
- **Check**: Run specific rule before and after changes

**Risk**: Over-extracting and creating too many small functions
- **Mitigation**: Aim for ~15-20 line helper functions (readable, not tiny)
- **Check**: Review each extracted function for readability

---

## 9. Committed Status

✅ **Phase 1 Commit**: c463977 (3/3 violations fixed)  
⏳ **Phase 2 Commit**: (ready to create after Phase 2A-2C complete)  
⏳ **Phase 3 Commit**: (E2E testing and verification)  

---

## Next Steps

1. ✅ Read this battle plan carefully
2. ⏳ Begin with Phase 2A (Critical Fixes) - highest impact
3. ⏳ Proceed through 2B and 2C systematically
4. ⏳ Run pre-push validation gate after each phase
5. ⏳ Commit completed phases with clear messages
6. ⏳ Prepare for Phase 3 (E2E testing)

**Ready to proceed? Start with Phase 2A by extracting bindCodeRequest helpers.**

