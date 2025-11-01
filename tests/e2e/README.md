# End-to-End (E2E) Tests

This directory contains Playwright E2E tests that validate the complete DevSmith platform user flow.

## What These Tests Validate

- ✅ Portal service accessibility via nginx proxy
- ✅ Health check endpoints respond correctly
- ✅ Cross-service routing through nginx
- ✅ Error handling for non-existent pages
- ✅ Concurrent service access
- ✅ Performance (page load times)

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

### Run Tests

The new Playwright configuration provides two test projects for flexible testing:

#### Quick Tests (Fast Feedback)
For rapid validation during development:
```bash
# Run only authentication tests (~15 tests, ~30 seconds)
npx playwright test --project=quick

# Watch mode for quick iteration
npx playwright test --project=quick --watch
```

#### Full Tests (Comprehensive Coverage)
For complete validation before pushing:
```bash
# Run all E2E tests (~92 tests, ~2-3 minutes)
npx playwright test --project=full

# Or simply:
npx playwright test
```

#### All Tests
```bash
# Install dependencies (one time)
npm ci

# Run all tests with all projects
npx playwright test

# Run specific test file
npx playwright test tests/e2e/authentication.spec.ts

# Run with UI mode (interactive)
npx playwright test --ui

# View test report
npx playwright show-report
```

### Configuration Details

**Quick Project** (`--project=quick`):
- Runs: `authentication.spec.ts` only
- Timeout: 15 seconds per test
- Workers: 2 (local) / 1 (CI)
- Best for: Development feedback loop

**Full Project** (`--project=full`):
- Runs: All `*.spec.ts` files
- Timeout: 30 seconds per test
- Workers: 2 (local) / 1 (CI)
- Best for: Pre-push validation

## Why E2E Tests Are Not in CI

E2E tests require:
- Full docker-compose network with multiple services
- Nginx reverse proxy routing
- Database connections
- WebSocket support

GitHub Actions CI environment has constraints that make reliable docker-compose networking problematic:
- Docker daemon doesn't support full network bridge mode
- Service-to-service communication requires special configuration
- Timeouts and connection failures are unpredictable

**Solution**: E2E tests run locally as part of development validation before pushing. Unit and integration tests (which don't depend on docker-compose) run in CI.

## Test Structure

Tests are organized by feature area:
- `full_user_flow.spec.ts` - Complete user journey validation

Each test:
- Uses isolated test cases
- Has clear given/when/then structure
- Includes proper assertions
- Handles timeouts gracefully

## Debugging Failed Tests

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
DEBUG=pw:api npx playwright test tests/e2e/ --debug
```

5. View full test report:
```bash
npx playwright show-report
```

## Contributing

When adding new E2E tests:
- Keep tests focused on user-facing behavior
- Use page object model for complex interactions
- Add timeouts for slow CI-like environments
- Document what the test validates
- Ensure test passes locally before committing
