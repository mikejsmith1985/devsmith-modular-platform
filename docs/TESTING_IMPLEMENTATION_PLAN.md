# Testing Implementation Plan

## Overview
This plan implements comprehensive testing to ensure API issues like the session_id field mismatch get caught automatically before reaching production.

## 1. ✅ COMPLETED: API Validation Tests

**File**: `tests/api/review_api_test.go`
**Purpose**: Catch payload structure mismatches between frontend and backend

**Tests Implemented**:
- `TestReviewAPIPayloadValidation`: Tests all 5 API modes with various payload combinations
- `TestReviewAPIContentTypeHandling`: Tests content-type validation
- `TestReviewAPIErrorMessages`: Tests error message quality
- `TestFrontendBackendAPIContract`: Tests exact frontend payload structures

**Results**: ✅ All tests pass - would have caught the session_id issue

## 2. 🚧 IN PROGRESS: Pre-commit Testing Hooks

**Files to Create**:
- `.githooks/pre-commit` - Git hook that runs tests before commit
- `scripts/validate-api-contracts.sh` - Script to validate API contracts
- `Makefile` targets for easy test execution

**Implementation**:
```bash
#!/bin/bash
# .githooks/pre-commit

echo "Running API validation tests..."
go test ./tests/api/... -v
if [ $? -ne 0 ]; then
    echo "❌ API tests failed - commit blocked"
    exit 1
fi

echo "Running frontend build test..."
cd frontend && npm run build
if [ $? -ne 0 ]; then
    echo "❌ Frontend build failed - commit blocked"
    exit 1
fi

echo "✅ All tests passed - commit allowed"
```

**Commands to Setup**:
```bash
# Install the pre-commit hook
chmod +x .githooks/pre-commit
git config core.hooksPath .githooks

# Test the hook
git add . && git commit -m "test: validate pre-commit hook"
```

## 3. 📋 TODO: CI/CD Pipeline Integration

**File**: `.github/workflows/api-validation.yml`
**Purpose**: Run tests on every PR and block merge if tests fail

**Pipeline Steps**:
1. Checkout code
2. Setup Go environment
3. Run API validation tests
4. Setup Node.js environment  
5. Run frontend build tests
6. Run integration tests with running services
7. Block PR merge if any step fails

## 4. 📋 TODO: Integration Tests

**File**: `tests/integration/review_api_integration_test.go`
**Purpose**: Test real API calls with running services

**Tests to Implement**:
```go
func TestReviewAPIIntegration(t *testing.T) {
    // Start review service in test mode
    // Make real HTTP calls to API endpoints
    // Verify responses match expected structure
    // Test authentication flows
    // Test error scenarios with real services
}
```

## 5. 📋 TODO: Frontend Contract Tests

**File**: `frontend/src/tests/api-contract.test.js`
**Purpose**: Test that frontend API calls match backend expectations

**Tests to Implement**:
```javascript
describe('API Contract Tests', () => {
  test('preview mode sends correct payload structure', async () => {
    const payload = { pasted_code: 'test', model: 'mistral:7b-instruct' };
    // Validate payload matches backend expectations
    // No session_id field should be present
  });
});
```

## 6. 📋 TODO: E2E Tests with Playwright

**File**: `tests/e2e/review-workflow.spec.ts`
**Purpose**: Test complete user workflows end-to-end

**Tests to Implement**:
```typescript
test('review workflow end-to-end', async ({ page }) => {
  await page.goto('/review');
  await page.fill('[data-testid=code-input]', 'package main\nfunc main() {}');
  await page.click('[data-testid=preview-button]');
  
  // Should not see 400 error
  await expect(page.locator('[data-testid=error]')).not.toBeVisible();
  // Should see analysis results
  await expect(page.locator('[data-testid=analysis-result]')).toBeVisible();
});
```

## 7. 📋 TODO: API Response Validation

**File**: `internal/middleware/response_validator.go`
**Purpose**: Validate API responses in development mode

**Implementation**:
```go
func ResponseValidatorMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if gin.Mode() == gin.DebugMode {
            // Capture response
            // Validate against OpenAPI spec
            // Log warnings for invalid responses
        }
        c.Next()
    }
}
```

## 8. 📋 TODO: Monitoring Integration

**File**: `internal/middleware/api_metrics.go`
**Purpose**: Track API usage and errors for monitoring dashboard

**Metrics to Track**:
- Request count by endpoint
- Response times
- Error rates (400, 500)
- Payload validation failures
- Authentication failures

## Success Criteria

✅ **Phase 1**: API validation tests created and passing
📋 **Phase 2**: Pre-commit hooks prevent bad commits  
📋 **Phase 3**: CI/CD pipeline blocks bad PRs
📋 **Phase 4**: E2E tests catch UI regressions
📋 **Phase 5**: Monitoring catches production issues

## Estimated Timeline

- **Phase 1**: ✅ Complete (2 hours)
- **Phase 2**: 3-4 hours (pre-commit hooks + scripts)
- **Phase 3**: 2-3 hours (GitHub Actions workflow)
- **Phase 4**: 4-5 hours (Playwright E2E tests)
- **Phase 5**: 2-3 hours (monitoring middleware)

**Total**: 13-17 hours over 2-3 days

## Commands for Development

```bash
# Run API validation tests
go test ./tests/api/... -v

# Run all tests
make test

# Run tests with coverage
make test-coverage

# Validate API contracts
./scripts/validate-api-contracts.sh

# Install pre-commit hooks
make install-hooks
```