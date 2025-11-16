# End-to-End (E2E) Tests

This directory contains Playwright E2E tests that validate the complete DevSmith platform user flow and features.

## 3-Tier Testing Strategy

We use a tiered E2E testing approach to validate user experience without killing developer velocity.

### Tier 1: Smoke Tests (Pre-Push, < 30 seconds)
Fast validation that critical paths aren't catastrophically broken.

**What runs**: 6-8 essential tests
- Portal loads and dark mode renders
- Review page loads with session form
- Critical mode button triggers analysis
- Logs dashboard loads with WebSocket
- Analytics dashboard loads with filters

**When**: Before each push (catch broken features immediately)

**Run locally**:
```bash
npx playwright test --project=smoke --workers=4
```

**Purpose**: Catch "feature completely broken" before push

---

### Tier 2: Feature Tests (On-Demand, 2-3 minutes)
Comprehensive validation of specific features before creating a PR.

**What runs**: 20+ tests per feature area
- All 5 reading modes return real Ollama analysis
- Session management CRUD operations
- Dark mode persistence across navigation
- HTMX interactions and loading indicators
- Accessibility and keyboard navigation

**When**: Before creating PR (validate feature is complete)

**Run locally**:
```bash
# Test specific feature
./scripts/validate-feature.sh review
./scripts/validate-feature.sh logs
./scripts/validate-feature.sh analytics

# Test everything
./scripts/validate-feature.sh all
```

**Purpose**: Ensure feature works end-to-end before PR

---

### Tier 3: Full Suite (CI/Nightly, 5-10 minutes)
Complete cross-browser, mobile, accessibility, performance testing.

**What runs**: All tests with multiple browsers/viewports
- Chrome, Firefox, Safari (simulated)
- Mobile (375px), Tablet (768px), Desktop (1920px)
- WCAG 2.1 AA accessibility
- Performance benchmarks
- Edge cases and error scenarios

**When**: After merge (CI) or before release

**Run locally**:
```bash
npx playwright test --project=full --workers=6
```

**Purpose**: Comprehensive validation for production readiness

---

## What These Tests Validate

### Smoke Tests
- ✅ Portal service loads and renders navigation
- ✅ Dark mode toggle is visible with Alpine.js attributes
- ✅ Review page has session form with all input methods
- ✅ Reading mode buttons are present and clickable
- ✅ Critical mode button triggers AI analysis
- ✅ Logs dashboard renders with WebSocket connection
- ✅ Analytics dashboard loads with HTMX filters

### Feature Tests (Coming)
- ✅ All 5 reading modes return real Ollama analysis
- ✅ Session management: create, read, update, delete
- ✅ Session persistence across page navigation
- ✅ Dark mode persists in localStorage and across navigation
- ✅ HTMX loading indicators show during requests
- ✅ Form validation works correctly
- ✅ Error handling displays user-friendly messages
- ✅ Keyboard navigation works (Tab, Enter, Escape)
- ✅ ARIA labels present on interactive elements
- ✅ Color contrast meets WCAG 2.1 AA

### Full Suite Tests
- ✅ All Tier 1 and 2 tests
- ✅ Multiple browsers (Chrome, Firefox, Safari)
- ✅ Multiple viewports (mobile, tablet, desktop)
- ✅ Performance: page load < 5s, API response < 3s
- ✅ WebSocket reliability and reconnection
- ✅ Error recovery and graceful degradation
- ✅ Race conditions and concurrent access

---

## Running E2E Tests Locally

### Prerequisites

1. Services must be running:
```bash
docker-compose up -d
```

2. Wait for services to be healthy (check logs):
```bash
docker-compose logs -f
# Wait until you see "healthy" status for all services
```

3. Ensure Node.js and npm are installed:
```bash
node --version
npm --version
```

### Quick Tests (For Development)

Run smoke tests during development for quick feedback:

```bash
# Run smoke tests only (30 seconds)
npx playwright test --project=smoke --workers=4

# Watch mode for TDD-style development
npx playwright test --project=smoke --watch

# Debug specific test
npx playwright test tests/e2e/smoke/dark-mode-toggle.spec.ts --debug
```

### Feature Validation (Before PR)

Validate your specific feature before creating a PR:

```bash
# Validate all review features
./scripts/validate-feature.sh review

# Validate all logs features  
./scripts/validate-feature.sh logs

# Validate analytics features
./scripts/validate-feature.sh analytics

# Validate entire platform
./scripts/validate-feature.sh all
```

### Full Test Suite

Run comprehensive tests:

```bash
# Install dependencies (one time)
npm ci

# Run all tests with default project
npx playwright test

# Run specific project
npx playwright test --project=full
npx playwright test --project=quick

# Run with UI mode (interactive)
npx playwright test --ui

# View HTML report
npx playwright show-report
```

### Debugging Failed Tests

If tests fail locally:

1. Check services are healthy:
```bash
curl http://localhost:3000/health
```

2. Check nginx routing:
```bash
curl -v http://localhost:3000/
```

3. Check service logs:
```bash
docker-compose logs [service-name]
```

4. Run test with debug output:
```bash
DEBUG=pw:api npx playwright test tests/e2e/smoke/ --debug
```

5. View full test report:
```bash
npx playwright show-report
```

---

## Test Structure

Tests are organized by tier:

```
tests/e2e/
├── smoke/                      # Tier 1: Quick validation (< 30s)
│   ├── portal-loads.spec.ts
│   ├── review-loads.spec.ts
│   ├── review-critical-mode.spec.ts
│   ├── dark-mode-toggle.spec.ts
│   ├── logs-dashboard-loads.spec.ts
│   └── analytics-loads.spec.ts
│
├── features/                   # Tier 2: Comprehensive validation (2-3min)
│   ├── review-all-modes.spec.ts
│   ├── review-session-management.spec.ts
│   ├── dark-mode-complete.spec.ts
│   ├── htmx-interactions.spec.ts
│   └── accessibility.spec.ts
│
├── full_user_flow.spec.ts     # Complete platform journey
├── authentication.spec.ts      # Auth and security
├── portal_login_dashboard.spec.ts
└── README.md
```

Each test:
- Uses isolated test cases (no shared state)
- Has clear given/when/then structure
- Includes proper assertions
- Handles timeouts gracefully
- Documents what behavior it validates

---

## Configuration Details

### Playwright Projects

| Project | Tests | Timeout | Workers | Duration | Purpose |
|---------|-------|---------|---------|----------|---------|
| smoke | `smoke/*.spec.ts` | 15s | 4 | ~30s | Fast pre-push validation |
| quick | `authentication.spec.ts` | 15s | 2 | ~30s | Auth validation |
| full | All `*.spec.ts` | 30s | 6 | ~5min | Comprehensive testing |

### Test Configuration

Located in `playwright.config.ts`:

```typescript
// Browser: Chrome
// Base URL: http://localhost:3000
// Timeout: 30s per test (15s for smoke)
// Retries: 0 local, 2 in CI
// Workers: 2-6 parallel tests
// Report: HTML + JSON
// Screenshots: On failure
// Video: On failure
// Trace: On first retry
```

---

## CI/CD Integration

### Current Status

E2E tests are **NOT** run in GitHub Actions CI due to docker-compose networking constraints:

- GitHub Actions runner doesn't support full network bridge mode
- Service-to-service communication fails unpredictably  
- Timeouts and flaky failures are common

**Solution**: E2E tests run locally during development before pushing. Unit/integration tests (which don't depend on docker-compose) run in CI.

### Workflow

1. **Local Development**: Developer runs smoke tests (30s) and feature tests (2-3min) before push
2. **Pre-Push Hook**: Validates Go code (format, imports, build, lint, vet)
3. **Git Push**: Code pushed to feature branch
4. **GitHub Actions CI**: Runs unit and integration tests (no E2E)
5. **Code Review**: Reviewer merges PR
6. **Production Validation**: Full E2E suite runs nightly against production

---

## Contributing

When adding new E2E tests:

1. **Define the test in a describe block**:
```typescript
test.describe('Feature: <Feature Name>', () => {
  test('<Specific user action>', async ({ page }) => {
    // GIVEN: Setup state
    await page.goto('/path');
    
    // WHEN: Execute action
    await page.click('button');
    
    // THEN: Verify outcome
    await expect(page.locator('.result')).toBeVisible();
  });
});
```

2. **Follow naming conventions**:
   - Smoke tests: `tests/e2e/smoke/<feature>.spec.ts`
   - Feature tests: `tests/e2e/features/<feature>.spec.ts`
   - Full journey: `tests/e2e/full_user_flow.spec.ts`

3. **Use descriptive test names** that explain user behavior:
   - ✅ "User can click dark mode toggle and see DOM change"
   - ❌ "toggle works"

4. **Add proper timeouts** for slow operations:
   - Smoke tests: 15s timeout
   - Feature tests: 30s timeout
   - Ollama calls: 15-30s within test

5. **Handle async operations** with proper waits:
   - Use `page.waitForResponse()` for API calls
   - Use `page.waitForTimeout()` for UI updates
   - Use `page.waitForNavigation()` for page changes

6. **Test from user perspective**:
   - Test what user sees and does
   - Test complete workflows
   - Test error scenarios
   - Don't test implementation details

7. **Keep tests focused**:
   - One feature per test file
   - 3-5 tests per describe block
   - Each test should be independent

---

## Performance Benchmarks

Expected test execution times:

- Smoke tests (6 tests): 20-30s
- Feature tests (20 tests): 2-3 minutes  
- Full suite (92 tests): 5-10 minutes

If tests are slower:
1. Check system resources (CPU, memory)
2. Check Docker performance
3. Check Ollama availability
4. Review network latency

---

## Troubleshooting

### Tests pass locally but fail in feature validation script

**Cause**: Services running but not fully initialized

**Fix**:
```bash
docker-compose down
docker-compose up -d
# Wait 30s for services to be healthy
./scripts/validate-feature.sh smoke
```

### Timeout errors in Ollama tests

**Cause**: Ollama not responding or too slow

**Fix**:
```bash
# Check Ollama is running
curl http://localhost:11434/api/tags

# Increase test timeout temporarily
npx playwright test --timeout=45000
```

### Dark mode toggle test fails - Alpine.js not rendering

**Cause**: Templ escaping Alpine.js directives

**Fix**: Check `internal/ui/components/nav/nav.templ` for proper Alpine.js syntax and Templ rendering

### HTMX tests fail - hx-* attributes not found

**Cause**: HTMX attributes may have different selectors

**Fix**: Update test selectors to match actual HTML attributes

---

## Success Criteria

✅ Implementation is complete when:
1. All smoke tests pass (< 30s)
2. Feature tests pass for the new feature (2-3min)
3. Pre-push hook successfully validates
4. `./scripts/validate-feature.sh <feature>` exits with 0
5. Feature works when tested manually in browser
6. Code review approved
7. No flaky tests (all green on repeat runs)
