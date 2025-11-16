# Traefik Gateway Violations - CRITICAL ARCHITECTURAL ISSUE

**Date**: 2025-11-15  
**Severity**: CRITICAL  
**Principle Violated**: ALL platform access MUST go through Traefik gateway (localhost:3000)

## Violations Found

### Category 1: Test Scripts (HIGH PRIORITY)
1. **scripts/regression-test.sh** - Uses ports 8081, 8082, 8083 directly
2. **scripts/verify-review-fixes.sh** - Uses port 8081 directly
3. **scripts/health-checks.sh** - Uses ports 8081, 8082, 8083 directly
4. **scripts/quick-test.sh** - Uses port 8081 directly
5. **scripts/test-logs-ingestion.sh** - Uses port 8082 directly
6. **scripts/simple-load-test.sh** - Uses port 8082 directly
7. **scripts/security-test-batch.sh** - Uses port 8082 directly
8. **scripts/load-test-batch.js** - Uses port 8082 directly
9. **scripts/dev.sh** - Documentation shows ports 8081, 8082, 8083
10. **scripts/deploy-portal.sh** - Uses port 3001 directly
11. **scripts/test-oauth-enhancements.sh** - Uses port 3001 directly

### Category 2: Docker Health Checks (ACCEPTABLE - Internal)
- docker-compose.yml health checks use localhost:3001, 8081, 8082, 8083
- **VERDICT**: ACCEPTABLE - These run INSIDE containers, not through Traefik

### Category 3: Documentation (NEEDS UPDATE)
1. **QUICK_START.md** - Shows direct port access examples
2. **README.md** - Shows Ollama direct access
3. **DEPLOYMENT.md** - Shows Ollama direct access
4. **cmd/healthcheck/README.md** - Shows direct port 8082 access
5. **setup.sh** - Shows ports 8081, 8082, 8083 in output

### Category 4: Integration Examples (NEEDS UPDATE)
1. **docs/integration-examples/tests/setup-test-env.sh** - Uses port 8082
2. **docs/openapi-review.yaml** - Uses port 8081 as server URL
3. **apps/review/static/assets/index-CUchRKLW.js** - Hardcoded port 8082

### Category 5: GitHub Workflows (NOT YET CHECKED)
- Need to verify smoke-test.yml and other CI workflows

## Correct Architecture

ALL external access should use:
- **Traefik Gateway**: http://localhost:3000
- **Portal**: http://localhost:3000/ (or /api/portal/*)
- **Review**: http://localhost:3000/api/review/*
- **Logs**: http://localhost:3000/api/logs/*
- **Analytics**: http://localhost:3000/api/analytics/*

## Fix Priority

1. **IMMEDIATE**: Fix all test scripts to use localhost:3000 with proper paths
2. **IMMEDIATE**: Fix CI workflows
3. **IMMEDIATE**: Update documentation
4. **IMMEDIATE**: Verify NO code uses direct ports

## Impact

- Current tests are bypassing Traefik routing
- May be testing wrong endpoints
- CI may be passing when it shouldn't
- Architecture violations could cause production issues
