# Copilot Autonomous Fix Patterns

> **Quick reference for common validation failures and autonomous fixes**

---

## Exit Code Handling

**The validation scripts now exit gracefully with clear guidance.**

When you see output ending with:
```
[DevSmith] ❌ Runtime validation failed!

For autonomous debugging, Copilot can:
  1. ./scripts/docker-validate.sh --json | jq '.issues[]'
  2. docker-compose logs nginx
  3. ./scripts/docker-validate.sh --auto-restart
  4. ./scripts/dev.sh
```

**This is NOT an interruption** - it's a controlled exit with actionable steps.

---

## Common Issue Patterns

### Pattern 1: Missing Go Files (Pre-Build)

**Error:**
```json
{
  "type": "no_go_files",
  "severity": "error",
  "service": "logs",
  "file": "cmd/logs",
  "message": "No Go files found in cmd/logs",
  "autoFixable": true
}
```

**Autonomous Fix:**
```bash
# Step 1: Confirm issue
./scripts/pre-build-validate.sh --json | jq '.issues[] | select(.type=="no_go_files")'

# Step 2: Auto-fix
./scripts/pre-build-validate.sh --fix

# Step 3: Verify
./scripts/pre-build-validate.sh

# Step 4: Continue
./scripts/dev.sh
```

---

### Pattern 2: Health Check Failures (Runtime)

**Error:**
```json
{
  "type": "health_unhealthy",
  "severity": "error",
  "service": "analytics",
  "message": "Health check failed - service not responding correctly"
}
```

**Autonomous Fix:**
```bash
# Step 1: Check logs
docker-compose logs analytics --tail=50

# Step 2: Common causes:
# - Service still starting → Wait longer
# - Database not ready → Check postgres
# - Missing curl → Add to Dockerfile

# Step 3: If just needs time:
./scripts/docker-validate.sh --wait --max-wait 120

# Step 4: If truly unhealthy, restart:
docker-compose restart analytics
sleep 10
./scripts/docker-validate.sh
```

---

### Pattern 3: Nginx 502 Bad Gateway ⭐ **Common After Rebuilds**

**Error:**
```json
{
  "type": "http_5xx",
  "severity": "error",
  "service": "nginx",
  "message": "Endpoint http://localhost:3000/ returned 502"
}
```

**Root Cause:** Container IPs changed after rebuild, nginx cached old IPs.

**Autonomous Fix:**
```bash
# Step 1: Verify backend services are healthy
docker-compose ps | grep healthy

# Step 2: If backends are healthy, restart nginx
docker-compose restart nginx

# Step 3: Wait for nginx to pick up new IPs
sleep 5

# Step 4: Validate
./scripts/docker-validate.sh

# Should now pass ✅
```

**Why this happens:**
1. Services get rebuilt → new container IPs assigned
2. Nginx starts with old IP mappings
3. Backend services are healthy, but nginx can't reach them
4. Restarting nginx forces DNS resolution to new IPs

**Prevention:** Nginx now has `depends_on: service: condition: service_healthy` which helps, but after rebuilds you may need to restart nginx.

---

### Pattern 4: Port Already in Use

**Error:**
```bash
Error response from daemon: driver failed programming external connectivity:
Bind for 0.0.0.0:8080 failed: port is already allocated
```

**Autonomous Fix:**
```bash
# Step 1: Find what's using the port
lsof -i :8080

# Step 2: If it's an old container:
docker-compose down
docker-compose up -d --build

# Step 3: If it's another process, kill it or use different port
```

---

### Pattern 5: Database Connection Failures

**Error (in logs):**
```
Failed to connect to database: dial tcp: lookup postgres: no such host
```

**Autonomous Fix:**
```bash
# Step 1: Check postgres is healthy
docker-compose ps postgres

# Step 2: If postgres is starting:
./scripts/docker-validate.sh --wait --max-wait 60

# Step 3: If postgres is unhealthy:
docker-compose restart postgres
sleep 15
docker-compose restart <failing-service>

# Step 4: Validate
./scripts/docker-validate.sh
```

---

## Decision Tree for Autonomous Debugging

```
./scripts/dev.sh exits with error
    ↓
Check which step failed:
    ↓
[1/3] Pre-build validation failed?
    ├─ autoFixable: true → Run --fix
    └─ autoFixable: false → Create missing files

[2/3] Docker build failed?
    ├─ Check for type errors in code
    └─ This SHOULD have been caught by pre-build

[3/3] Runtime validation failed?
    ├─ Parse JSON: ./scripts/docker-validate.sh --json
    │
    ├─ Issue: health_unhealthy
    │   ├─ Check logs
    │   ├─ Wait longer: --wait
    │   └─ Restart: docker-compose restart [service]
    │
    ├─ Issue: http_5xx (nginx)
    │   ├─ Check backend services healthy
    │   └─ Restart nginx: docker-compose restart nginx
    │
    ├─ Issue: http_5xx (service)
    │   ├─ Check logs
    │   ├─ Verify database connection
    │   └─ Restart service
    │
    └─ Issue: container_stopped
        └─ docker-compose up -d [service]
```

---

## Full Autonomous Fix Session Example

**User runs:**
```bash
./scripts/dev.sh
```

**Output:**
```
[DevSmith] ❌ Runtime validation failed!

HIGH PRIORITY: 1 issue(s)
  • [http_5xx] nginx - Endpoint returned 502

For autonomous debugging, Copilot can:
  1. ./scripts/docker-validate.sh --json | jq '.issues[]'
  2. docker-compose logs nginx
  3. docker-compose restart nginx
  4. ./scripts/docker-validate.sh
```

**Copilot's autonomous response:**

```bash
# 1. Parse JSON to understand issue
./scripts/docker-validate.sh --json | jq '.issues[]'

# Output shows: nginx 502, all backends healthy

# 2. Check nginx logs
docker-compose logs nginx --tail=20

# Output shows: "Connection refused" to backends

# 3. Diagnosis: Container IPs changed, nginx needs restart
docker-compose restart nginx

# 4. Wait for nginx to start
sleep 5

# 5. Revalidate
./scripts/docker-validate.sh

# ✅ All checks passed!
```

**Report to user:**
```
Fixed! The issue was nginx using stale container IPs after rebuild.
I restarted nginx and all services are now healthy.

Services available at:
  • Portal:    http://localhost:8080
  • Review:    http://localhost:8081
  • Logs:      http://localhost:8082
  • Analytics: http://localhost:8083
  • Gateway:   http://localhost:3000
```

---

## Key Principles for Autonomous Debugging

### 1. Scripts Exit Gracefully, Not Abruptly

**Before (with `set -e`):**
- Validation fails → script exits immediately
- Copilot sees "interrupted"
- No clear guidance

**After (graceful exit):**
- Validation fails → script catches error
- Prints clear guidance for Copilot
- Exits with code 1 (failure) but gracefully

### 2. Always Parse JSON First

```bash
# Human-readable output is for humans
./scripts/docker-validate.sh

# JSON output is for Copilot
./scripts/docker-validate.sh --json | jq '.issues[]'
```

### 3. Follow the Suggested Commands

The error output literally tells you what to run:
```
For autonomous debugging, Copilot can:
  1. ./scripts/docker-validate.sh --json | jq '.issues[]'
  2. docker-compose logs nginx
  3. docker-compose restart nginx
  4. ./scripts/docker-validate.sh
```

**Run these in order** - they're sequenced for success.

### 4. Verify After Each Fix

```bash
# After any fix, always validate
./scripts/docker-validate.sh

# If still failing, check JSON again
./scripts/docker-validate.sh --json
```

### 5. Don't Ask User Unless Truly Stuck

**Only ask user when:**
- Multiple fix attempts failed
- Ambiguous errors (not in patterns)
- External dependencies needed (API keys, etc.)
- Architectural decision required

**Otherwise, fix autonomously** using the patterns above.

---

## Testing Your Autonomous Debugging

### Simulate Issues:

```bash
# Test 1: Empty service directory
rm -rf cmd/logs/*
./scripts/dev.sh
# Expected: Pre-build catches it, suggests --fix

# Test 2: Missing health check
# Remove curl from Dockerfile
./scripts/dev.sh
# Expected: Runtime catches unhealthy, suggests restart

# Test 3: Nginx stale IPs
docker-compose up -d --build portal
./scripts/docker-validate.sh
# Expected: Nginx 502, suggests restart nginx
```

### Expected Behavior:

Each test should:
1. ✅ Exit with clear error message
2. ✅ Provide JSON with structured data
3. ✅ Suggest specific fix commands
4. ✅ Allow autonomous resolution

---

## Summary

**The validation system is designed for autonomous debugging.**

**When validation fails:**
1. Scripts exit gracefully with guidance
2. JSON output provides structured data
3. Suggested commands are sequenced for success
4. You (Copilot) can fix without asking user

**Common fix pattern:**
```bash
# 1. Parse issue
./scripts/docker-validate.sh --json | jq '.issues[]'

# 2. Check logs
docker-compose logs [service]

# 3. Apply fix (restart/rebuild/--fix)
docker-compose restart [service]

# 4. Verify
./scripts/docker-validate.sh

# 5. Report success to user
```

**Your goal:** User runs `./scripts/dev.sh` once, you handle any issues autonomously, they start coding with everything working.
