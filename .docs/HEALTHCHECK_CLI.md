# Enhanced Healthcheck CLI - Comprehensive Diagnostics

## Overview

The healthcheck CLI now provides **actionable diagnostics** for container and service health issues instead of just reporting failures.

## What's New

### 1. Container Health.Log Inspection

Shows detailed health check history from `docker inspect`:

```
maildev: RUNNING
  Last health check: 2025-01-20T10:30:40Z (exit 1)
  Output: exec: "curl": executable file not found
```

**What this shows:**
- Timestamp of last health check
- Exit code (0 = healthy, non-zero = unhealthy)
- Exact error message from the health check command
- Helps identify why container failed (missing binary, connection refused, etc.)

### 2. Fallback Host-Side Probes

When a container's health check fails, the CLI tries reaching the service from the host:

```
maildev: RUNNING
  Last health check: 2025-01-20T10:30:40Z (exit 1)
  Output: exec: "curl": executable file not found
  ✓ Accessible from host (HTTP 200 on :1080)
```

**What this means:**
- Container health check is broken (missing curl binary)
- BUT the service IS running and reachable from host
- The issue is the healthcheck configuration, not the service itself
- Fix: Update healthcheck to use a tool that exists in the image

### 3. Pattern-Based Remediation Suggestions

Detects common failure patterns and suggests specific fixes:

#### Pattern: Executable Not Found
```
Suggested fix:
  Healthcheck command failed (missing tool). Options:
  1. Add tool to image: docker-compose up -d --build maildev
  2. Use host-side check instead of in-container exec
  3. For Alpine: docker exec maildev apk add curl
```

#### Pattern: Connection Refused
```
Suggested fix:
  Service connection failed. Try:
  1. Check service logs: docker-compose logs nginx --tail=20
  2. Restart service: docker-compose restart nginx
  3. Check port bindings: docker-compose port nginx
```

#### Pattern: Unknown Exit Code
```
Suggested fix:
  Healthcheck exited with code 127. Container logs:
  docker-compose logs postgres --tail=10
```

### 4. Recent Container Logs

For missing containers, shows why they failed to start:

```
nginx: MISSING
  Suggested fix:
    docker-compose up -d nginx
  Recent logs:
    error: image 'nginx:latest' not found
    [check image availability]
```

## Usage

### Basic Health Check

```bash
./healthcheck
```

Shows all services status with diagnostics.

### With Specific Format

```bash
# JSON output (for parsing/integration)
./healthcheck -format json

# Human-readable (default)
./healthcheck -format human
```

### Advanced Diagnostics

```bash
# Include Phase 2 advanced checks (default: enabled)
./healthcheck -advanced=true
```

## Example Output

### Healthy System

```
🔍 Running pre-push validation checks...

Portal: RUNNING
  Last health check: 2025-01-20T10:35:22Z (exit 0)
  ✓ Accessible from host
  
Review: RUNNING
  Last health check: 2025-01-20T10:35:21Z (exit 0)
  ✓ Accessible from host

Logs: RUNNING
  Last health check: 2025-01-20T10:35:20Z (exit 0)
  ✓ Accessible from host
```

### System with Issues

```
nginx: MISSING
  Suggested fix:
    docker-compose up -d nginx

maildev: RUNNING
  Last health check: 2025-01-20T10:35:10Z (exit 1)
  Output: exec: "curl": executable file not found
  ✓ Accessible from host (HTTP 200 on :1080)
  Suggested fix:
    Healthcheck command failed (missing tool). Options:
    1. Add tool to image: docker-compose up -d --build maildev
    2. Use host-side check instead of in-container exec
```

## How It Works

### Step 1: Check Container Status
- Runs `docker-compose ps` to see which services are running
- Marks services as RUNNING or MISSING

### Step 2: Get Health Logs (for running services)
- Runs `docker inspect` on each container
- Extracts `State.Health.Log` entries
- Shows last health check timestamp, exit code, and output

### Step 3: Try Host-Side Probe (if health check failed)
- Tests TCP connection to service port
- Attempts HTTP GET to web services
- Shows if service is reachable despite healthcheck failure

### Step 4: Pattern Detection
- Analyzes error output for known failure patterns
- Suggests specific remediation commands
- Provides multiple options for different scenarios

### Step 5: Display Summary
- Shows status, health logs, accessibility, and suggestions
- One place to see everything needed for troubleshooting

## Troubleshooting Examples

### Issue: "maildev unhealthy"

**Old Output:**
```
maildev: unhealthy
```

**New Output:**
```
maildev: RUNNING
  Last health check: 2025-01-20T10:30:40Z (exit 1)
  Output: exec: "curl": executable file not found
  ✓ Accessible from host (HTTP 200 on :1080)
  Suggested fix:
    Healthcheck command failed (missing tool). Options:
    1. Add tool to image: docker-compose up -d --build maildev
    2. Use host-side check instead of in-container exec
    3. For Alpine: docker exec maildev apk add curl
```

**What Changed:**
- Shows the service IS running and reachable
- Identifies the REAL problem: missing `curl` binary
- Suggests 3 specific solutions to try

### Issue: "nginx gateway connection refused"

**Old Output:**
```
gateway: connection refused
```

**New Output:**
```
nginx: MISSING
  Suggested fix:
    docker-compose up -d nginx

gateway: RUNNING
  Last health check: 2025-01-20T10:30:35Z (exit 1)
  Output: Connection refused
  ✗ Not accessible from host
  Suggested fix:
    Service connection failed. Try:
    1. Check service logs: docker-compose logs gateway --tail=20
    2. Restart service: docker-compose restart gateway
    3. Check port bindings: docker-compose port gateway
```

**What Changed:**
- Shows nginx is missing AND gateway is misconfigured
- Suggests starting nginx first
- Provides exact commands to debug gateway

### Issue: "postgres unhealthy"

**Old Output:**
```
postgres: unhealthy
```

**New Output:**
```
postgres: RUNNING
  Last health check: 2025-01-20T10:30:15Z (exit 2)
  ✓ Accessible from host (TCP successful on :5432)
  Suggested fix:
    Healthcheck exited with code 2. Container logs:
    docker-compose logs postgres --tail=10
```

**What Changed:**
- Shows port IS listening (TCP successful)
- Suggests checking logs for actual error
- Indicates it's a database issue, not connectivity

## Implementation Details

### Service Port Configuration

The CLI knows which ports each service uses for fallback checks:

```go
ServicePorts: map[string]int{
    "nginx":      3000,
    "portal":     8080,
    "review":     8081,
    "logs":       8082,
    "analytics":  8083,
    "postgres":   5432,
    "maildev":    1080,
}
```

### Container Health Log Structure

```json
{
  "Start": "2025-01-20T10:30:35.123456789Z",
  "End": "2025-01-20T10:30:36.456789123Z",
  "ExitCode": 1,
  "Output": "exec: \"curl\": executable file not found"
}
```

### Fallback Check Results

```json
{
  "tcp_connection": "success" | "failed: ...",
  "http_check": "HTTP 200" | "failed: ...",
  "accessible": true | false
}
```

## Benefits

✅ **Faster Troubleshooting** - See the actual problem, not just "unhealthy"
✅ **Actionable Guidance** - Get exact commands to fix issues
✅ **Smart Fallbacks** - Know if service works despite failed healthcheck
✅ **Pattern Detection** - Common problems → immediate solutions
✅ **Context Included** - Logs, timestamps, exit codes all shown
✅ **No Manual Debugging** - No need for docker inspect, docker logs separately

## Related Documentation

- `docker-compose.yml` - Service definitions and healthchecks
- `internal/healthcheck/` - Checker implementations
- `cmd/healthcheck/main.go` - CLI entry point

