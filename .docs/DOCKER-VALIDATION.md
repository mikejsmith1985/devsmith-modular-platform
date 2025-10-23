# Docker Validation Guide

## What It Does

The `scripts/docker-validate.sh` script validates your Docker containers are **running, healthy, and serving traffic correctly**. It catches configuration issues before you waste time manually testing broken services.

**Validation checks:**
- ✅ Container status (running/stopped/missing)
- ✅ Health checks (healthy/unhealthy/starting)
- ✅ HTTP endpoints (200 OK, not 404/500)
- ✅ Port bindings (correctly mapped to host)

**Results are logged to `.validation/status.json`** - tell Copilot to read this file and fix the issues.

---

## Runtime Route Discovery (100% Accurate)

The validation script uses **runtime discovery** to find all routes automatically by querying your running services. This ensures 100% accuracy - you'll never test a route that doesn't exist, and you'll never miss a route that does.

### How It Works

1. **Debug Endpoint**: Each service exposes `/debug/routes` (development only)
2. **Query at Runtime**: Script queries each service to get actual registered routes
3. **Discover All Routes**: Finds routes in `main.go`, handler files, and dynamically registered routes
4. **Test Everything**: Validates all discovered endpoints

**Example:**
```bash
# Portal service exposes its routes
curl http://localhost:8080/debug/routes

# Returns:
{
  "service": "portal",
  "count": 9,
  "routes": [
    {"method": "GET", "path": "/"},
    {"method": "GET", "path": "/auth/login"},
    {"method": "GET", "path": "/auth/github/dashboard"},
    {"method": "GET", "path": "/dashboard"},
    {"method": "GET", "path": "/health"},
    ...
  ]
}
```

### Benefits

- ✅ **100% Accurate** - Gets routes directly from running services
- ✅ **No Maintenance** - Automatically discovers new routes as you add them
- ✅ **Catches Missing Routes** - Won't test routes that don't exist (preventing false failures)
- ✅ **Finds Hidden Routes** - Discovers routes in handler files, not just main.go
- ✅ **Development Only** - Debug endpoint disabled in production (ENV=production)

**Discovered Sources:**
- `runtime` - Routes from `/debug/routes` endpoint (most accurate)
- `gateway` - nginx location blocks (user-facing routes)
- `static_file` - Static assets (favicon, CSS, JS)

---

## Why This Matters

### The Problem

Docker configurations often have issues:
- Containers start but serve 404s (misconfigured routes)
- Show "healthy" but return 500s (uninitialized dependencies)
- Port bindings don't work (typos, conflicts)
- Missing health check implementations

**You waste time discovering these issues during manual testing.**

### The Solution

Run validation manually before testing:
- **Pre-flight checks** - Verify everything works before you test
- **Clear diagnostics** - Know exactly what's broken and why
- **Logged results** - `.validation/status.json` file for Copilot to read and fix
- **Structured output** - JSON format for programmatic access

---

## ⚠️ CRITICAL: Preventing Multiple Instance Problems

### Why This Section Is Here

**Problem:** You or an AI assistant might accidentally run a service twice - once in Docker and once directly on your computer. This causes both copies to fail with confusing errors.

**Solution:** Follow these simple rules to avoid the problem entirely.

---

### Rule #1: Check First, Then Act

**Before running ANY service command, check if Docker is running:**

```bash
docker-compose ps
```

**Look at the output:**
- If you see services with `"Up"` status → Docker is running, use only `docker-compose` commands
- If you see nothing or `"Exit"` status → Safe to run services directly

---

### Rule #2: One Method at a Time

**Choose ONE way to run your services:**

**Option A: Using Docker (Recommended)**
```bash
# Start everything
./scripts/dev.sh

# Restart a service
docker-compose restart portal

# View logs
docker-compose logs -f portal

# Stop everything
docker-compose down
```

**Option B: Running Directly (Only when Docker is stopped)**
```bash
# Make sure Docker is stopped first!
docker-compose down

# Now it's safe to run directly
go run cmd/portal/main.go
```

**⚠️ NEVER mix these methods!** Pick one and stick with it.

---

### Rule #3: When In Doubt, Use Docker

**If you're unsure which method you're using:**

1. Stop everything:
   ```bash
   docker-compose down
   ```

2. Start with Docker:
   ```bash
   ./scripts/dev.sh
   ```

3. Use only `docker-compose` commands from now on

---

### What Errors Look Like

**If you accidentally run a service twice, you'll see errors like:**

```
Error: listen tcp :8080: bind: address already in use
Error: Failed to start server: address already in use
panic: runtime error: address already in use
```

**What this means:** Something is already using the port (probably Docker).

**How to fix it:**
```bash
# See what's using the port
lsof -i :8080

# If it shows a Docker container, use docker-compose
docker-compose restart portal

# If it shows a direct process, stop it with Ctrl+C, then use Docker
docker-compose up -d portal
```

---

### Quick Reference: Port Numbers

Each service uses a specific port. Only ONE thing can use each port at a time.

| Service   | Port | Check Command          |
|-----------|------|------------------------|
| Portal    | 8080 | `lsof -i :8080`       |
| Review    | 8081 | `lsof -i :8081`       |
| Logs      | 8082 | `lsof -i :8082`       |
| Analytics | 8083 | `lsof -i :8083`       |
| Postgres  | 5432 | `lsof -i :5432`       |
| Nginx     | 3000 | `lsof -i :3000`       |

**If any `lsof` command shows output:** That port is in use. Check `docker-compose ps` to see if it's Docker.

---

## Understanding the Output

### Human-Readable Dashboard

When validation runs, you'll see an intelligent dashboard:

```
🐳 Docker validation (standard mode)...

📦 Project: devsmith-modular-platform
🔍 Services to check: 6

[1/4] Checking container status...
[2/4] Checking health checks...
[3/4] Checking HTTP endpoints...
[4/4] Checking port bindings...

════════════════════════════════════════════════════════════════
🐳 DOCKER VALIDATION SUMMARY (completed in 8s)
════════════════════════════════════════════════════════════════

CHECK RESULTS:
  ✓ containers running      passed
  ✗ health checks           failed
  ✗ http endpoints          failed
  ✓ port bindings           passed

HIGH PRIORITY (Blocking): 2 issue(s)
  • [health_unhealthy] analytics - Health check failed - service not responding correctly
    → Check logs: docker-compose logs analytics | Review health endpoint implementation

  • [http_5xx] portal - Endpoint http://localhost:8080/health returned 500 (server error)
    → Check application logs: docker-compose logs portal. Verify database connectivity and configuration.

QUICK FIXES:
  • Auto-restart unhealthy:  ./scripts/docker-validate.sh --auto-restart
  • Wait for services:       ./scripts/docker-validate.sh --wait --max-wait 60
  • View logs:               docker-compose logs [service]
  • Restart all:             docker-compose restart
  • Rebuild service:         docker-compose up -d --build [service]

DOCUMENTATION:
  • Docker validation guide:  .docs/DOCKER-VALIDATION.md
  • Health check patterns:    .docs/DOCKER-COPILOT-GUIDE.md
════════════════════════════════════════════════════════════════
```

### Issue Types

#### Container Issues

**`[container_missing]`** - Container doesn't exist
- **Cause:** Service not defined in docker-compose.yml or never started
- **Fix:** Run `docker-compose up -d [service]`

**`[container_stopped]`** - Container is not running
- **Cause:** Container crashed or was manually stopped
- **Fix:** Run `docker-compose start [service]` or check logs for crash reason

#### Health Check Issues

**`[health_unhealthy]`** - Health check failing
- **Cause:** Service started but health endpoint returns non-200 status
- **Fix:**
  1. Check logs: `docker-compose logs [service]`
  2. Verify database connectivity
  3. Check environment variables
  4. Test health endpoint manually: `curl http://localhost:PORT/health`

**`[health_starting]`** - Still initializing
- **Cause:** Service is starting up (database migrations, etc.)
- **Fix:** Wait 30-60 seconds or use `--wait` flag

**`[health_missing]`** - No health check configured
- **Cause:** Missing HEALTHCHECK in Dockerfile or docker-compose.yml
- **Fix:** Add health check (see Health Check Patterns below)

#### HTTP Endpoint Issues

**`[http_404]`** - Endpoint returns 404 Not Found
- **Cause:** Route misconfiguration in nginx.conf or missing handler in service
- **Fix:**
  1. Verify nginx routing: check `docker/nginx/nginx.conf`
  2. Verify service implements the endpoint
  3. Check service is listening on correct port

**`[http_5xx]`** - Server error (500, 502, 503)
- **Cause:** Application error, database connection failure, or uninitialized dependencies
- **Fix:**
  1. Check service logs: `docker-compose logs [service]`
  2. Verify DATABASE_URL is correct
  3. Ensure postgres is healthy: `docker-compose ps postgres`
  4. Check for missing migrations

**`[http_timeout]`** - Connection timeout or refused
- **Cause:** Port binding issue or service not listening
- **Fix:**
  1. Verify port in docker-compose.yml: `ports: ["8080:8080"]`
  2. Check service is listening: `docker-compose exec [service] netstat -tlnp`
  3. Verify firewall rules

#### Port Binding Issues

**`[port_not_bound]`** - Port not mapped to host
- **Cause:** Missing or incorrect `ports:` section in docker-compose.yml
- **Fix:** Add proper port mapping: `ports: ["HOST_PORT:CONTAINER_PORT"]`

---

## Testing Strategy

**IMPORTANT:** The validation script follows a **gateway-first** testing strategy to prevent confusion about running services locally.

### What Gets Tested

1. **Health Checks on Direct Ports** ✅
   - `http://localhost:8080/health` (portal)
   - `http://localhost:8081/health` (review)
   - `http://localhost:8082/health` (logs)
   - `http://localhost:8083/health` (analytics)

2. **User-Facing Routes Through Gateway** ✅
   - `http://localhost:3000/` (nginx gateway)
   - `http://localhost:3000/analytics/*` (proxied to analytics service)
   - `http://localhost:3000/review/*` (proxied to review service)

3. **NOT Tested Directly** ❌
   - Service routes on ports 8080-8083 (except health checks)
   - Internal-only routes not exposed through nginx

**Why?** This prevents the "running services locally" confusion. Users should ONLY access services through the nginx gateway (port 3000), not direct service ports.

---

## Common Use Cases

### 1. Basic Validation

After starting services, validate everything is working:

```bash
./scripts/docker-validate.sh
```

### 2. Auto-Fix Issues

Automatically fix simple issues (like restarting unhealthy services):

```bash
./scripts/docker-validate.sh --auto-fix
```

**What gets auto-fixed:**
- Health check failures → Restarts the service
- Simple container issues → Restarts the service

**What doesn't get auto-fixed:**
- 5xx errors (server errors) → Need code/config fixes
- 404 errors → Need routing configuration
- Build failures → Need code changes

### 3. Re-Test Only Failed Endpoints (Fast!)

After fixing issues, re-test only what failed (5-10x faster):

```bash
./scripts/docker-validate.sh --retest-failed
```

### 4. Progressive Validation (Layer-by-Layer)

Test in layers, stopping at first failure:

```bash
./scripts/docker-validate.sh --progressive
```

**Layers:**
1. Gateway check (is nginx responding?)
2. Service health checks (are services healthy?)
3. Full endpoint testing (do all routes work?)

### 5. Combine Flags for Common Workflows

```bash
# Fix issues, then re-test only what failed
./scripts/docker-validate.sh --auto-fix --retest-failed

# Progressive validation with auto-fix
./scripts/docker-validate.sh --progressive --auto-fix

# Quick validation (containers + health only)
./scripts/docker-validate.sh --quick

# Thorough validation (with port bindings)
./scripts/docker-validate.sh --thorough
```

### 6. JSON Output for Tool Integration

Get structured JSON output for parsing by tools:

```bash
./scripts/docker-validate.sh --json | jq '.summary'
```

Example output:
```json
{
  "status": "failed",
  "duration": 12,
  "summary": {
    "total": 3,
    "errors": 2,
    "warnings": 1,
    "autoFixable": 1
  },
  "checkResults": {
    "containersRunning": "passed",
    "healthChecks": "failed",
    "httpEndpoints": "failed",
    "portBindings": "passed"
  }
}
```

---

## Health Check Patterns

### Go Service Health Endpoint

Every service should implement a `/health` endpoint:

```go
// cmd/[service]/main.go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    // Check database connectivity
    if err := db.Ping(); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "unhealthy",
            "error":  err.Error(),
        })
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
    })
}

// In main():
http.HandleFunc("/health", healthHandler)
```

### Docker Compose Health Check

Add to `docker-compose.yml`:

```yaml
services:
  myservice:
    # ... other config ...
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 5
      start_period: 40s  # Time for app to boot
    depends_on:
      postgres:
        condition: service_healthy  # Wait for DB
```

### Dockerfile Health Check

Add to `Dockerfile`:

```dockerfile
# Install curl for health checks
RUN apk add --no-cache curl

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1
```

---

## Integration with Development Workflow

### Two Scripts - When to Use Which

**Script 1: `./scripts/dev.sh` (Full Startup + Validation)**

Use when:
- Starting from scratch (no containers running)
- After making code changes that need rebuild
- Want to restart everything fresh

```bash
./scripts/dev.sh
```

This:
1. Stops all services (`docker-compose down`)
2. Rebuilds containers (`docker-compose up -d --build`)
3. Waits for services to be healthy (max 120s)
4. Runs validation and saves to `.validation/status.json`
5. Shows logs if successful, or tells you to fix issues

---

**Script 2: `./scripts/docker-validate.sh` (Validation Only)**

Use when:
- Containers are already running
- You just want to check current status
- After making config changes (without code rebuild)
- In the fix/test loop with Copilot

```bash
./scripts/docker-validate.sh
```

This:
1. Discovers all endpoints dynamically
2. Tests everything
3. Updates `.validation/status.json`
4. Shows results

---

### Typical Development Workflow

**Initial startup:**
```bash
./scripts/dev.sh
```

**If validation fails, fix loop:**
```bash
# 1. Check what's broken
cat .validation/status.json | jq '.validation.issues[]'

# 2. Tell Copilot to fix it
# "Read .validation/status.json and fix the issues"

# 3. If Copilot made code changes, rebuild
docker-compose up -d --build [service]

# 4. Re-validate
./scripts/docker-validate.sh

# 5. Repeat until green
```

**Quick status check anytime:**
```bash
./scripts/docker-validate.sh
```

### Validation Modes

```bash
# Standard mode (containers + health + endpoints)
./scripts/docker-validate.sh

# Quick mode (skip endpoint testing)
./scripts/docker-validate.sh --quick

# Thorough mode (everything including port bindings)
./scripts/docker-validate.sh --thorough

# Wait for services to become healthy
./scripts/docker-validate.sh --wait --max-wait 120

# Auto-restart unhealthy containers
./scripts/docker-validate.sh --auto-restart
```

### CI/CD Integration

Add to your CI/CD pipeline:

```yaml
# .github/workflows/test.yml
- name: Start services
  run: docker-compose up -d

- name: Validate services
  run: ./scripts/docker-validate.sh --wait --max-wait 60

- name: Run integration tests
  run: go test -tags=integration ./...
```

---

## Optional: Continuous Monitoring with Uptime Kuma

For persistent monitoring and alerting, enable the monitoring stack:

### 1. Start Monitoring Services

```bash
docker-compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d
```

This adds:
- **Uptime Kuma** (http://localhost:3001) - Visual dashboard with HTTP monitoring
- **Autoheal** - Automatically restarts unhealthy containers

### 2. Configure Uptime Kuma

1. Open http://localhost:3001
2. Create an admin account
3. Add monitors for each service:
   - Portal: `http://portal:8080/health`
   - Review: `http://review:8081/health`
   - Logs: `http://logs:8082/health`
   - Analytics: `http://analytics:8083/health`
   - Gateway: `http://nginx:80/`

4. Set up notifications (email, Slack, Discord, etc.)

### 3. Autoheal Behavior

Autoheal monitors all containers and:
- Checks every 10 seconds
- Restarts containers marked as "unhealthy"
- Waits 30 seconds after startup before monitoring
- Logs all restart actions

View autoheal logs:
```bash
docker-compose logs autoheal
```

---

## Troubleshooting

### "No containers running"

**Problem:** `docker-compose ps` shows no services

**Solutions:**
1. Start services: `docker-compose up -d`
2. Check docker daemon: `docker ps`
3. Verify docker-compose.yml exists

### All health checks show "starting"

**Problem:** Services never become healthy

**Solutions:**
1. Check if health endpoints are implemented
2. View logs: `docker-compose logs [service]`
3. Check database is healthy: `docker-compose ps postgres`
4. Increase `start_period` in health check config

### Validation passes but manual curl fails

**Problem:** Validation says OK but `curl http://localhost:8080` fails

**Solutions:**
1. Check port mapping: `docker-compose ps`
2. Verify service is listening: `docker-compose exec [service] netstat -tlnp`
3. Check firewall rules
4. Ensure using correct host port (not container port)

### nginx returns 502 Bad Gateway

**Problem:** Gateway is healthy but returns 502 for service routes

**Solutions:**
1. Verify upstream services are healthy
2. Check nginx.conf upstream definitions
3. Ensure service names match docker-compose service names
4. Verify network connectivity: `docker-compose exec nginx ping portal`

---

## Advanced: Customizing Validation

### Add New Services

Edit `scripts/docker-validate.sh`:

```bash
# Service definitions (port inside container)
declare -A SERVICES=(
    [postgres]="5432"
    [portal]="8080"
    [mynewservice]="8084"  # Add here
)

# Endpoint definitions for HTTP checks
declare -A ENDPOINTS=(
    [portal]="http://localhost:8080/health"
    [mynewservice]="http://localhost:8084/health"  # Add here
)
```

### Custom Health Check Logic

For complex health checks, create a custom endpoint:

```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    health := checkSystemHealth()

    if !health.Healthy {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(health)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(health)
}

func checkSystemHealth() HealthStatus {
    status := HealthStatus{Healthy: true, Checks: make(map[string]bool)}

    // Database check
    if err := db.Ping(); err != nil {
        status.Healthy = false
        status.Checks["database"] = false
        status.Error = fmt.Sprintf("DB error: %v", err)
    } else {
        status.Checks["database"] = true
    }

    // Redis check (if applicable)
    if err := redis.Ping(); err != nil {
        status.Healthy = false
        status.Checks["redis"] = false
    } else {
        status.Checks["redis"] = true
    }

    return status
}
```

---

## Working with Copilot

When validation fails:

1. **Read the status file:**
   ```
   Tell Copilot: "Read .validation/status.json and show me what's broken"
   ```

2. **Let Copilot fix issues:**
   ```
   Tell Copilot: "Fix the issues in .validation/status.json"
   ```

3. **Re-validate:**
   ```bash
   ./scripts/docker-validate.sh
   ```

The status file is overwritten on each validation run, so it always reflects current state.

---

## Best Practices

### 1. Always Implement Health Endpoints

Every HTTP service should have a `/health` endpoint that:
- Returns 200 OK when healthy
- Returns 503 Service Unavailable when unhealthy
- Checks critical dependencies (database, redis, etc.)
- Returns JSON with detailed status

### 2. Use Proper start_period

Set `start_period` to cover:
- Application boot time
- Database connection establishment
- Migrations (if run on startup)
- Cache warming

Typical values:
- Simple services: 10s
- Services with migrations: 40s
- Complex initialization: 60s

### 3. Set depends_on with service_healthy

Always use `condition: service_healthy`:

```yaml
services:
  myservice:
    depends_on:
      postgres:
        condition: service_healthy  # Wait for DB
      redis:
        condition: service_healthy  # Wait for cache
```

This ensures services start in the correct order.

### 4. Monitor in Production

Use Uptime Kuma or similar tools in production:
- Set up alerts for downtime
- Monitor response times
- Track availability metrics
- Configure automatic restarts

### 5. Test Health Endpoints

Add integration tests for health endpoints:

```go
func TestHealthEndpoint(t *testing.T) {
    resp, err := http.Get("http://localhost:8080/health")
    require.NoError(t, err)

    assert.Equal(t, http.StatusOK, resp.StatusCode)

    var health HealthStatus
    json.NewDecoder(resp.Body).Decode(&health)
    assert.True(t, health.Healthy)
}
```

---

## Command Line Flags Reference

### All Available Flags

```bash
./scripts/docker-validate.sh [FLAGS]
```

| Flag | Description | Example |
|------|-------------|---------|
| `--auto-fix` | Automatically run fixable commands (restarts, simple fixes) | `--auto-fix` |
| `--retest-failed` | Only re-test endpoints that failed in previous run (5-10x faster) | `--retest-failed` |
| `--progressive` | Layer-by-layer validation, stops at first failure | `--progressive` |
| `--diff` | Show progress compared to previous run (enabled by default) | `--diff` |
| `--quick` | Containers + health checks only (skip endpoint testing) | `--quick` |
| `--thorough` | All checks including port bindings | `--thorough` |
| `--json` | Output JSON format instead of human-readable | `--json` |
| `--wait` | Wait for services to become healthy before testing | `--wait` |
| `--max-wait N` | Maximum seconds to wait (default: 120) | `--max-wait 300` |

### Common Flag Combinations

**After making fixes:**
```bash
./scripts/docker-validate.sh --auto-fix --retest-failed
```
- Auto-fixes simple issues
- Re-tests only what failed
- Fastest iteration loop

**For debugging:**
```bash
./scripts/docker-validate.sh --progressive --thorough
```
- Tests layer-by-layer (fail fast)
- Comprehensive checks including ports
- Clear failure point

**In CI/CD:**
```bash
./scripts/docker-validate.sh --wait --json | jq -e '.validation.status == "passed"'
```
- Waits for services to be ready
- Structured JSON output
- Exits with error code if failed

**Quick health check:**
```bash
./scripts/docker-validate.sh --quick
```
- Skips endpoint testing
- Just containers + health
- ~1 second vs ~2 seconds

---

## Summary

### Quick Reference

**Starting fresh?**
```bash
./scripts/dev.sh
```

**Containers already running?**
```bash
./scripts/docker-validate.sh
```

**After Copilot makes fixes?**
```bash
docker-compose up -d --build [service]  # If code changed
./scripts/docker-validate.sh            # Re-validate
```

**Check what's broken?**
```bash
cat .validation/status.json | jq '.validation.issues[]'
```

---

### Complete Workflow

1. **Start services:** `./scripts/dev.sh`
2. **If validation fails:** Check `.validation/status.json`
3. **Fix issues:** Tell Copilot to read the file and fix
4. **Rebuild if needed:** `docker-compose up -d --build [service]`
5. **Re-validate:** `./scripts/docker-validate.sh`
6. **Repeat** until green

---

### Best Practices

**For Developers:**
- Use `dev.sh` for first startup of the day
- Use `docker-validate.sh` for quick checks
- Check `.validation/status.json` when debugging
- Implement `/health` endpoints in all services
- Use `depends_on` with `service_healthy` in docker-compose

**For CI/CD:**
- Use `--json` for structured output
- Use `--wait` to block until services are ready
- Parse JSON output to fail builds on errors
- Example: `./scripts/docker-validate.sh --wait --json | jq -e '.status == "passed"'`
