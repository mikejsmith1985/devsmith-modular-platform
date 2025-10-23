# Copilot Status File Guide

> **For:** GitHub Copilot
> **Purpose:** Monitor DevSmith validation status without blocking your interaction

---

## The Problem

When you run long-running processes (like `./scripts/dev.sh` with log streaming), you lose the ability to chat with Copilot. The user needs you available to fix issues while the dev environment is running.

## The Solution

**Single status file:** `.validation/status.json`

- **Updated:** Every time a validation runs (pre-build or runtime)
- **Overwrites:** Previous status (no accumulation)
- **Contains:** Latest validation results with timestamp

---

## Workflow

### Terminal 1 (User runs manually):
```bash
./scripts/dev.sh
# Logs stream here - user monitors visually
```

### Terminal 2 (Copilot workspace):
```bash
# When user reports an issue or asks you to check:
cat .validation/status.json | jq '.'
```

---

## Status File Format

```json
{
  "timestamp": "2025-10-22T20:30:45+00:00",
  "phase": "runtime",
  "validation": {
    "status": "failed",
    "issues": [
      {
        "type": "health_unhealthy",
        "severity": "error",
        "service": "analytics",
        "message": "Health check failed - service not responding correctly",
        "suggestion": "Check logs: docker-compose logs analytics",
        "autoFixable": true,
        "fixCommand": "docker-compose restart analytics"
      }
    ],
    "checkResults": {
      "containersRunning": "passed",
      "healthChecks": "failed",
      "httpEndpoints": "passed",
      "protectedRoutes": "passed",
      "publicRoutes": "passed",
      "staticFiles": "passed",
      "portBindings": "passed"
    },
    "summary": {
      "total": 1,
      "errors": 1,
      "warnings": 0,
      "autoFixable": 1
    }
  }
}
```

### Fields Explained:

- **timestamp:** When validation last ran
- **phase:** Either `"pre-build"` or `"runtime"`
- **validation.status:** `"passed"` or `"failed"`
- **validation.issues[]:** Array of problems found
- **validation.checkResults:** Pass/fail for each check category
- **validation.summary:** Quick overview of issue counts

---

## When to Check the File

### Scenario 1: User Reports Issue
```
User: "The dev environment failed to start"
Copilot: Let me check the validation status...
```

```bash
cat .validation/status.json | jq '.validation.issues[]'
```

### Scenario 2: Proactive Monitoring
```
User: "I'm working on the portal, can you monitor for issues?"
Copilot: I'll check the validation status periodically.
```

```bash
# Check every minute (you can script this)
watch -n 60 'cat .validation/status.json | jq ".validation.status"'
```

### Scenario 3: After Making Changes
```
User: "I just updated the Dockerfile, can you verify everything still works?"
Copilot: Let me trigger validation and check the results...
```

```bash
./scripts/docker-validate.sh
cat .validation/status.json | jq '.validation.status'
```

---

## Autonomous Debugging Flow

When you detect failures in `.validation/status.json`:

### Step 1: Parse the JSON
```bash
cat .validation/status.json | jq '.validation.issues[]'
```

### Step 2: Check for Auto-Fixable Issues
```bash
cat .validation/status.json | jq '.validation.issues[] | select(.autoFixable==true)'
```

### Step 3: Execute Fix Commands
```bash
# Example: Issue says to restart analytics
docker-compose restart analytics

# Wait a bit
sleep 5

# Verify fix worked
./scripts/docker-validate.sh
cat .validation/status.json | jq '.validation.status'
```

### Step 4: Check Logs if Not Auto-Fixable
```bash
# Get the service with issues
SERVICE=$(cat .validation/status.json | jq -r '.validation.issues[0].service')

# Check its logs
docker-compose logs $SERVICE --tail=50
```

---

## Phase-Specific Handling

### Pre-Build Phase (`"phase": "pre-build"`)

**Common Issues:**
- `no_go_files` - Empty service directory
- `missing_main_go` - No main.go file
- `wrong_package` - Not `package main`
- `missing_health_endpoint` - No /health handler

**Auto-Fix:**
```bash
./scripts/pre-build-validate.sh --fix
cat .validation/status.json | jq '.validation.status'
```

### Runtime Phase (`"phase": "runtime"`)

**Common Issues:**
- `health_unhealthy` - Service not responding
- `http_5xx` - Server error
- `container_stopped` - Container crashed
- `static_file_not_found` - Missing static assets

**Auto-Fix:**
```bash
# Restart unhealthy services
./scripts/docker-validate.sh --auto-restart

# Or wait for services to become healthy
./scripts/docker-validate.sh --wait --max-wait 60

# Check results
cat .validation/status.json | jq '.validation.status'
```

---

## Examples

### Example 1: Detect and Fix Unhealthy Service

```bash
# User says: "Something's wrong with analytics"

# Check status
$ cat .validation/status.json | jq '.validation.issues[] | select(.service=="analytics")'
{
  "type": "health_unhealthy",
  "severity": "error",
  "service": "analytics",
  "message": "Health check failed",
  "suggestion": "Check logs: docker-compose logs analytics",
  "autoFixable": true,
  "fixCommand": "docker-compose restart analytics"
}

# Execute fix
$ docker-compose restart analytics
Container devsmith-analytics-1  Restarting
Container devsmith-analytics-1  Started

# Wait and verify
$ sleep 5
$ ./scripts/docker-validate.sh
✅ Docker validation PASSED

# Confirm in status file
$ cat .validation/status.json | jq '.validation.status'
"passed"
```

### Example 2: Pre-Build Issue

```bash
# User says: "Docker build is failing for logs service"

# Check status
$ cat .validation/status.json | jq '.validation'
{
  "status": "failed",
  "phase": "pre-build",
  "issues": [{
    "type": "no_go_files",
    "service": "logs",
    "autoFixable": true,
    "fixCommand": "./scripts/pre-build-validate.sh --fix"
  }]
}

# Auto-fix
$ ./scripts/pre-build-validate.sh --fix
Auto-fixing: Creating logs service structure...
✅ Pre-build validation PASSED

# Verify
$ cat .validation/status.json | jq '.validation.status'
"passed"
```

### Example 3: Monitor Status While Working

```bash
# User says: "I'm going to rebuild services, let me know if anything breaks"

# Set up monitoring (if your editor supports it)
$ watch -n 5 'cat .validation/status.json | jq -r ".validation.status // \"no-status\""'

# Or check periodically
$ cat .validation/status.json | jq '{status: .validation.status, errors: .validation.summary.errors, timestamp: .timestamp}'
{
  "status": "passed",
  "errors": 0,
  "timestamp": "2025-10-22T20:45:30+00:00"
}
```

---

## Quick Command Reference

```bash
# Check overall status
cat .validation/status.json | jq '.validation.status'

# List all issues
cat .validation/status.json | jq '.validation.issues[]'

# Get auto-fixable issues
cat .validation/status.json | jq '.validation.issues[] | select(.autoFixable==true)'

# Get error count
cat .validation/status.json | jq '.validation.summary.errors'

# Get phase (pre-build or runtime)
cat .validation/status.json | jq -r '.phase'

# Get timestamp of last validation
cat .validation/status.json | jq -r '.timestamp'

# Pretty print everything
cat .validation/status.json | jq '.'

# Check if file exists and is recent (within 5 minutes)
test -f .validation/status.json && \
  [ $(( $(date +%s) - $(date -d "$(jq -r .timestamp .validation/status.json)" +%s) )) -lt 300 ] && \
  echo "Status is current" || echo "Status is stale or missing"
```

---

## Integration with Your Workflow

### When User Runs `./scripts/dev.sh`:

1. **Pre-build validation** runs → updates `.validation/status.json`
2. **Docker build** happens
3. **Runtime validation** runs → updates `.validation/status.json`
4. **Logs stream** in Terminal 1
5. **You (Copilot)** are free in Terminal 2 to help user

### When Issues Occur:

1. User sees error in logs (Terminal 1)
2. User tells you: "Fix the analytics service"
3. You check `.validation/status.json`
4. You parse issues and execute fixes
5. You verify fix worked by checking updated status file

### The Key Benefit:

**You remain interactive** while monitoring status. No blocking on long-running processes. User gets live logs, you get structured data.

---

## Best Practices

1. **Always check the timestamp** - stale status means validation hasn't run recently
2. **Check autoFixable first** - try automated solutions before manual investigation
3. **Look at checkResults** - tells you which category failed (health checks, endpoints, etc.)
4. **Read the suggestion field** - often contains the exact command to run
5. **Verify your fixes** - always re-check status after applying fixes

---

## Limitations

- Status file only updates when validation runs (not real-time)
- For live monitoring, user should watch Terminal 1 logs
- You need to manually trigger re-validation after fixes: `./scripts/docker-validate.sh`
- File is gitignored - won't exist in fresh checkouts (created on first validation)

---

## Summary

**File location:** `.validation/status.json`
**Updated by:** `./scripts/pre-build-validate.sh` and `./scripts/docker-validate.sh`
**Format:** JSON with timestamp, phase, and validation results
**Your use:** Check status, parse issues, execute fixes, verify resolution
**User's use:** Run dev environment in separate terminal, tell you when issues occur

**Goal:** Enable you to autonomously debug without blocking on long-running processes.
