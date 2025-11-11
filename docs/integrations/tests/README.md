# Cross-Repo Logging Integration Tests

Automated test suite for Week 2 sample files and framework integrations.

## Test Coverage

### Unit Tests
- ✅ JavaScript logger (buffer, batch, retry, cleanup)
- ✅ Python logger (threading, batch, retry, cleanup)
- ✅ Go logger (goroutines, mutex, batch, retry, cleanup)

### Integration Tests
- ✅ Express.js middleware (request/response logging, timing, header redaction)
- ✅ Flask extension (hooks, decorator, exception tracking)
- ✅ Gin middleware (request/response logging, panic recovery)

### API Tests
- ✅ Batch endpoint validation
- ✅ API key authentication
- ✅ Rate limiting
- ✅ Invalid request handling

## Running Tests

### All Tests
```bash
npm test
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
