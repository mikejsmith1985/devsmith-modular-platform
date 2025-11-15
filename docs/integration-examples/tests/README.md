# Cross-Repo Logging Integration Tests

Automated test suite for Week 2 sample files and framework integrations.

## Test Coverage (102 tests, ~2,780 lines)

### Unit Tests (48 tests)
- ✅ JavaScript logger (16 tests) - buffer, batch, retry, cleanup
- ✅ Python logger (17 tests) - threading, batch, retry, cleanup
- ✅ Go logger (15 tests) - goroutines, mutex, batch, retry, cleanup

### Integration Tests (37 tests)
- ✅ Express.js middleware (13 tests) - request/response logging, timing, header redaction
- ✅ Flask extension (13 tests) - hooks, decorator, exception tracking
- ✅ Gin middleware (11 tests) - request/response logging, panic recovery

### API Tests (17 tests)
- ✅ Batch endpoint validation (4 tests) - single log, 100 logs, all levels, rich metadata
- ✅ API key authentication (2 tests) - invalid key, missing key
- ✅ Request validation (5 tests) - missing slug, missing logs, empty logs, missing fields, invalid level
- ✅ Performance (2 tests) - 500 logs <5s, concurrent batches
- ✅ Rate limiting (1 test) - 100 rapid requests
- ✅ Error handling (3 tests) - malformed JSON, partial failures, network errors

## Running Tests

### All Tests
```bash
npm run test:all  # Runs Playwright E2E + cross-repo integration tests
```

### Cross-Repo Tests Only
```bash
npm run test:cross-repo  # Runs logger, express, and API tests
```

### Specific Test Suites
```bash
# JavaScript logger unit tests
npm run test:logger:js

# Python logger unit tests
npm run test:logger:py

# Go logger unit tests
npm run test:logger:go

# Express middleware integration tests
npm run test:express

# Flask extension integration tests
npm run test:flask

# Gin middleware integration tests
npm run test:gin

# API endpoint tests
npm run test:api
```

## Test Environment Setup

Tests require:
1. DevSmith platform running (docker-compose up)
2. Test project created in database
3. Valid API key generated

Setup script:
```bash
bash docs/integrations/tests/setup-test-env.sh
```

## CI/CD Integration

GitHub Actions workflow runs all tests on:
- Push to feature branch
- Pull request creation
- Pull request updates

See: `.github/workflows/cross-repo-logging-tests.yml`

## Manual Testing

For external app integration testing (requires beta users):
- See `docs/integrations/tests/MANUAL_TEST_GUIDE.md`
- Sample apps in `docs/integrations/tests/sample-apps/`
