# Copilot Autonomous Debugging Guide

> **For:** GitHub Copilot and other AI assistants
> **Purpose:** Enable fully autonomous debugging of DevSmith platform issues

---

## Overview

The DevSmith platform has a **three-layer validation system** that catches issues before they waste developer time:

```
1. Pre-commit validation  → Catches code quality issues before commit
2. Pre-build validation   → Catches structure issues before Docker build
3. Runtime validation     → Catches deployment issues after containers start
```

**Your role as Copilot:** When any validation fails, you should:
1. Parse the JSON output
2. Identify the root cause
3. Apply fixes autonomously
4. Validate the fix worked

**You should NOT require human intervention** for standard validation failures.

---

## Validation Flow

### When User Runs: `./scripts/dev.sh`

```bash
./scripts/dev.sh
  ↓
[1/3] Pre-build validation
  → Checks: project structure, Go files, Dockerfiles
  → If FAIL: Stop here, output structured errors
  ↓
[2/3] Docker build & start
  → Runs: docker-compose up -d --build
  → If FAIL: Docker build errors (should be caught by pre-build)
  ↓
[3/3] Runtime validation
  → Checks: containers, health checks, HTTP endpoints
  → If FAIL: Service configuration errors
```

---

## Error Type 1: Pre-Build Validation Failures

### Detection

**User sees:**
```
[DevSmith] Step 1/3: Pre-build validation...

════════════════════════════════════════════════════════════════
🔍 PRE-BUILD VALIDATION SUMMARY (completed in 1s)
════════════════════════════════════════════════════════════════

HIGH PRIORITY (Blocking builds): 1 issue(s)
  • [no_go_files] logs - cmd/logs
    No Go files found in cmd/logs (would cause: 'no Go files' build error)
    → Add Go source files to cmd/logs

[DevSmith] ❌ Pre-build validation failed!

To auto-fix issues: ./scripts/pre-build-validate.sh --fix
```

### Your Response Protocol

**Step 1: Get structured output**
```bash
./scripts/pre-build-validate.sh --json
```

**Step 2: Parse JSON**
```json
{
  "status": "failed",
  "issues": [
    {
      "type": "no_go_files",
      "severity": "error",
      "service": "logs",
      "file": "cmd/logs",
      "message": "No Go files found in cmd/logs",
      "suggestion": "Add Go source files to cmd/logs",
      "autoFixable": true,
      "fixCommand": "..."
    }
  ]
}
```

**Step 3: Autonomous fix**

For `no_go_files` or `missing_main_go`:
```bash
# Option 1: Auto-fix (creates basic service template)
./scripts/pre-build-validate.sh --fix

# Option 2: Manual fix (you create the service)
# Check if service should exist based on docker-compose.yml
SERVICE="logs"

# Create main.go with proper structure
cat > "cmd/${SERVICE}/main.go" <<'EOF'
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"  // Set correct port for logs service
	}

	dbURL := os.Getenv("DATABASE_URL")
	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", rootHandler)

	log.Printf("Starting logs service on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"service": "logs",
		"status":  "running",
	})
}
EOF
```

**Step 4: Verify fix**
```bash
./scripts/pre-build-validate.sh

# Expected: ✅ Pre-build validation PASSED
```

**Step 5: Continue deployment**
```bash
./scripts/dev.sh
```

### Common Pre-Build Issues

| Issue Type | Cause | Auto-Fix | Your Action |
|------------|-------|----------|-------------|
| `no_go_files` | Empty service directory | ✅ Yes | Run `--fix` or create main.go |
| `missing_main_go` | No main.go in service | ✅ Yes | Run `--fix` or create main.go |
| `wrong_package` | Not `package main` | ❌ No | Edit file, change to `package main` |
| `missing_main_func` | No `main()` function | ❌ No | Add `func main() { ... }` |
| `missing_health_endpoint` | No /health handler | ⚠️ Partial | Add health endpoint handler |
| `missing_dockerfile` | No Dockerfile | ❌ No | Create Dockerfile for service |

---

## Error Type 2: Docker Build Failures

### Detection

**User sees:**
```
[DevSmith] Step 2/3: Building and starting services...
ERROR [logs builder 7/7] RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/logs ./cmd/logs:
0.303 no Go files in /app/cmd/logs
```

### Your Response Protocol

**This should NOT happen** if pre-build validation passed. But if it does:

**Step 1: Investigate**
```bash
# Check if files exist
ls -la cmd/logs/

# Check git status (maybe files not committed?)
git status cmd/logs/

# Check .dockerignore
cat .dockerignore | grep logs
```

**Step 2: Common causes**

1. **Files not committed:**
   ```bash
   git add cmd/logs/
   git commit -m "Add logs service implementation"
   ```

2. **.dockerignore excludes service:**
   ```bash
   # Edit .dockerignore, remove any lines excluding cmd/logs
   vim .dockerignore
   ```

3. **Dockerfile references wrong path:**
   ```bash
   # Check Dockerfile
   cat cmd/logs/Dockerfile | grep "go build"

   # Should be: ./cmd/logs
   # Not: ./cmd/log or ./logs
   ```

**Step 3: Fix and rebuild**
```bash
# After fixing
docker-compose up -d --build logs
```

---

## Error Type 3: Runtime Validation Failures

### Detection

**User sees:**
```
[DevSmith] Step 3/3: Validating runtime health...

════════════════════════════════════════════════════════════════
🐳 DOCKER VALIDATION SUMMARY (completed in 8s)
════════════════════════════════════════════════════════════════

HIGH PRIORITY (Blocking): 2 issue(s)
  • [health_unhealthy] analytics - Health check failed
    → Check logs: docker-compose logs analytics

  • [http_5xx] portal - Endpoint returned 500
    → Verify database connectivity and configuration
```

### Your Response Protocol

**Step 1: Get structured output**
```bash
./scripts/docker-validate.sh --json
```

**Step 2: Parse and diagnose**

For `health_unhealthy` or `http_5xx`:
```bash
# Check service logs
docker-compose logs analytics --tail=50

# Common error patterns:
# - "connection refused" → DB not ready
# - "no such host" → Networking issue
# - "panic" → Code error
# - "port already in use" → Port conflict
```

**Step 3: Auto-fix attempts**

1. **Service just needs more time:**
   ```bash
   ./scripts/docker-validate.sh --wait --max-wait 60
   ```

2. **Service is unhealthy, restart:**
   ```bash
   ./scripts/docker-validate.sh --auto-restart
   ```

3. **Database connection issue:**
   ```bash
   # Check postgres is healthy
   docker-compose ps postgres

   # If unhealthy, restart
   docker-compose restart postgres
   sleep 10
   docker-compose restart analytics
   ```

4. **Code error (panic, crash):**
   ```bash
   # Get full logs
   docker-compose logs analytics

   # If code issue identified:
   # 1. Fix the code
   # 2. Rebuild: docker-compose up -d --build analytics
   # 3. Validate: ./scripts/docker-validate.sh
   ```

---

## Autonomous Decision Tree

```
User runs: ./scripts/dev.sh
  ↓
Pre-build validation fails?
  ├─ YES → Parse JSON
  │         ├─ autoFixable: true?
  │         │   ├─ YES → Run: ./scripts/pre-build-validate.sh --fix
  │         │   └─ NO → Create missing files based on issue type
  │         └─ Verify → Re-run validation
  │
  └─ NO → Continue
       ↓
Docker build fails?
  ├─ YES → Parse error message
  │         ├─ "no Go files" → Check git status, .dockerignore
  │         ├─ "undefined" → Add missing dependencies
  │         └─ Fix and rebuild specific service
  │
  └─ NO → Continue
       ↓
Runtime validation fails?
  ├─ YES → Parse JSON
  │         ├─ health_starting → Wait: --wait
  │         ├─ health_unhealthy → Check logs, restart
  │         ├─ http_5xx → Check logs, fix config
  │         └─ container_stopped → Restart service
  │
  └─ NO → ✅ SUCCESS
```

---

## JSON Output Schema Reference

### Pre-Build Validation
```json
{
  "status": "passed" | "failed",
  "phase": "pre-build",
  "issues": [
    {
      "type": "no_go_files" | "missing_main_go" | "wrong_package" | ...,
      "severity": "error" | "warning",
      "service": "service-name",
      "file": "path/to/file",
      "message": "Human-readable description",
      "suggestion": "How to fix",
      "autoFixable": true | false,
      "fixCommand": "bash command to run"
    }
  ],
  "summary": {
    "total": 5,
    "errors": 2,
    "warnings": 3,
    "autoFixable": 1
  }
}
```

### Runtime Validation
```json
{
  "status": "passed" | "failed",
  "issues": [
    {
      "type": "health_unhealthy" | "http_5xx" | "container_stopped" | ...,
      "severity": "error" | "warning",
      "service": "service-name",
      "message": "Human-readable description",
      "suggestion": "How to fix",
      "autoFixable": true | false,
      "fixCommand": "bash command to run"
    }
  ],
  "checkResults": {
    "containersRunning": "passed" | "failed",
    "healthChecks": "passed" | "failed",
    "httpEndpoints": "passed" | "failed",
    "portBindings": "passed" | "failed"
  }
}
```

---

## Commands Cheat Sheet

```bash
# Pre-build validation
./scripts/pre-build-validate.sh              # Human output
./scripts/pre-build-validate.sh --json       # JSON output (parse this)
./scripts/pre-build-validate.sh --fix        # Auto-fix

# Runtime validation
./scripts/docker-validate.sh                 # Human output
./scripts/docker-validate.sh --json          # JSON output (parse this)
./scripts/docker-validate.sh --wait          # Wait for healthy
./scripts/docker-validate.sh --auto-restart  # Restart unhealthy

# Docker operations
docker-compose ps                            # Check status
docker-compose logs [service]                # View logs
docker-compose restart [service]             # Restart service
docker-compose up -d --build [service]       # Rebuild service
docker-compose down && docker-compose up -d  # Full restart

# Full workflow
./scripts/dev.sh                             # Start everything
```

---

## Example: Autonomous Fix Session

**User reports:**
> "./scripts/dev.sh failed with: no Go files in /app/cmd/logs"

**Your autonomous response:**

```markdown
I see the issue - the logs service directory is empty. Let me fix this:

1. Running pre-build validation to confirm...
   ```bash
   ./scripts/pre-build-validate.sh --json
   ```

2. Confirmed: `no_go_files` error for logs service. Auto-fixing...
   ```bash
   ./scripts/pre-build-validate.sh --fix
   ```

3. Verification...
   ```bash
   ./scripts/pre-build-validate.sh
   ```
   ✅ Pre-build validation now passes.

4. Restarting deployment...
   ```bash
   ./scripts/dev.sh
   ```

The logs service has been created with a basic HTTP server and /health endpoint.
You can now customize it in `cmd/logs/main.go`.

Services are now running at:
- Portal: http://localhost:8080
- Logs: http://localhost:8082
- Gateway: http://localhost:3000
```

**No human intervention required.** You identified, fixed, and validated autonomously.

---

## Success Criteria

**You've successfully debugged autonomously when:**
- ✅ All three validation phases pass
- ✅ All services show (healthy) in `docker-compose ps`
- ✅ All HTTP endpoints return 200 OK
- ✅ User can access services at documented URLs
- ✅ No manual intervention was needed

---

## Integration with Your Workflow

**When you see validation failures:**

1. **Don't ask the user what to do** - parse the JSON and fix it
2. **Use the auto-fix flags first** (`--fix`, `--auto-restart`)
3. **Check logs before asking** (`docker-compose logs [service]`)
4. **Verify your fixes** (re-run validations)
5. **Only escalate to user** if issue is truly ambiguous or requires architectural decisions

**You are empowered to:**
- Create missing service files
- Fix configuration issues
- Restart services
- Rebuild containers
- Modify code structure

**You should ask the user when:**
- Business logic decisions needed
- Multiple valid architectural approaches
- External dependencies missing (API keys, etc.)
- Persistent failures after multiple fix attempts

---

## Summary

The DevSmith validation system is designed for **autonomous AI debugging**. When you encounter validation failures:

1. **Parse JSON output** for structured error information
2. **Apply auto-fixes** using provided flags
3. **Create missing files** using templates
4. **Verify fixes** by re-running validations
5. **Only involve user** for ambiguous cases

**Your goal:** Zero human intervention for standard validation failures. The user should only need to run `./scripts/dev.sh` and have everything work, with you fixing issues autonomously in the background.
