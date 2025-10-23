# Complete Validation Workflow

## Three-Layer Defense System

DevSmith now has a **comprehensive three-layer validation system** that catches issues before they waste your time:

```
Layer 1: Pre-Commit        → Code quality (tests, lint, build)
Layer 2: Pre-Build         → Structure & dependencies
Layer 3: Runtime           → Service health & connectivity
```

---

## Workflow Diagram

```
Developer commits code
    ↓
[Layer 1] .git/hooks/pre-commit
    ├─ Go fmt/vet
    ├─ golangci-lint
    ├─ go test
    └─ go build
    ↓
✅ Commit accepted
    ↓
Developer runs: ./scripts/dev.sh
    ↓
[Layer 2] scripts/pre-build-validate.sh
    ├─ Check project structure
    ├─ Check Go modules
    ├─ Check Dockerfiles exist
    ├─ Check service files exist
    └─ Check main.go has package main + main()
    ↓
✅ Pre-build validation passed
    ↓
[Layer 3a] docker-compose up -d --build
    ├─ Build all services
    └─ Start containers
    ↓
[Layer 3b] scripts/docker-validate.sh --wait
    ├─ Wait for containers running
    ├─ Wait for health checks passing
    ├─ Check HTTP endpoints (200 OK)
    └─ Verify port bindings
    ↓
✅ Runtime validation passed
    ↓
🎉 Developer starts coding
```

---

## What Each Layer Catches

### Layer 1: Pre-Commit (Code Quality)

**Purpose:** Catch code quality issues before they enter git history

**Checks:**
- ✅ Code formatting (`go fmt`)
- ✅ Static analysis (`go vet`)
- ✅ Linting (`golangci-lint`)
- ✅ Unit tests (`go test`)
- ✅ Build errors (type errors, imports)

**When it runs:** On every `git commit`

**Output:** Structured JSON with issue priority, fix suggestions, auto-fix capabilities

**Example failure:**
```
HIGH PRIORITY (Blocking): 2 issue(s)
  • [test_mock_panic] Test 'TestAggregator' - missing mock expectation
    → Add Mock.On("FindAllServices").Return(...)

  • [build_typecheck] Error return value is not checked
    → Fix type error - this blocks tests from running
```

**Bypass (not recommended):** `git commit --no-verify`

**Documentation:** `.docs/PRE-COMMIT-HOOK.md`

---

### Layer 2: Pre-Build (Structure & Dependencies)

**Purpose:** Catch structural issues BEFORE wasting time on Docker builds

**Checks:**
- ✅ Project structure (cmd/, internal/, docker/ exist)
- ✅ Go modules (go.mod, go.sum valid)
- ✅ Dockerfiles exist for all services
- ✅ Service directories have Go files
- ✅ main.go has `package main` and `func main()`
- ✅ Health endpoints implemented

**When it runs:**
- Automatically in `./scripts/dev.sh` (first step)
- Manually: `./scripts/pre-build-validate.sh`

**Output:** Structured JSON with auto-fix capabilities

**Example failure:**
```
HIGH PRIORITY (Blocking builds): 1 issue(s)
  • [no_go_files] logs - cmd/logs
    No Go files found in cmd/logs (would cause: 'no Go files' build error)
    → Add Go source files to cmd/logs

QUICK FIXES:
  • Auto-fix issues: ./scripts/pre-build-validate.sh --fix
```

**Auto-fix:** `./scripts/pre-build-validate.sh --fix`
- Creates missing service directories
- Generates basic main.go with health endpoint
- Sets up proper package structure

**Documentation:** `.docs/COPILOT-AUTONOMOUS-DEBUGGING.md`

---

### Layer 3: Runtime (Service Health)

**Purpose:** Ensure services are actually working and serving traffic

**Checks:**
- ✅ Containers running (not stopped/crashed)
- ✅ Health checks passing (not unhealthy)
- ✅ HTTP endpoints responding 200 OK (not 404/500)
- ✅ Ports correctly bound to host

**When it runs:**
- Automatically in `./scripts/dev.sh` (after build)
- Manually: `./scripts/docker-validate.sh`

**Output:** Structured JSON with service status

**Example failure:**
```
HIGH PRIORITY (Blocking): 2 issue(s)
  • [health_unhealthy] analytics - Health check failed
    → Check logs: docker-compose logs analytics

  • [http_5xx] portal - Endpoint returned 500
    → Verify database connectivity and configuration

QUICK FIXES:
  • Auto-restart unhealthy: ./scripts/docker-validate.sh --auto-restart
  • Wait for services:      ./scripts/docker-validate.sh --wait --max-wait 60
```

**Auto-fix:** `./scripts/docker-validate.sh --auto-restart`
- Restarts unhealthy containers
- Waits for services to become healthy

**Documentation:** `.docs/DOCKER-VALIDATION.md`

---

## Autonomous Debugging (Copilot Integration)

All three layers output **structured JSON** designed for AI assistant parsing:

```json
{
  "status": "failed",
  "phase": "pre-build",
  "issues": [
    {
      "type": "no_go_files",
      "severity": "error",
      "service": "logs",
      "file": "cmd/logs",
      "message": "No Go files found in cmd/logs",
      "suggestion": "Add Go source files to cmd/logs",
      "autoFixable": true,
      "fixCommand": "mkdir -p cmd/logs && touch cmd/logs/main.go"
    }
  ],
  "summary": {
    "total": 1,
    "errors": 1,
    "warnings": 0,
    "autoFixable": 1
  }
}
```

**Copilot can:**
1. Parse JSON output
2. Identify root cause
3. Run auto-fix commands
4. Verify fixes worked
5. Continue without human intervention

**See:** `.docs/COPILOT-AUTONOMOUS-DEBUGGING.md` for complete Copilot integration guide

---

## Command Reference

### Developer Commands

```bash
# Full startup with all validations
./scripts/dev.sh

# Individual validations
./scripts/pre-build-validate.sh           # Structure check
./scripts/docker-validate.sh              # Runtime check

# Auto-fix modes
./scripts/pre-build-validate.sh --fix     # Create missing files
./scripts/docker-validate.sh --auto-restart  # Restart unhealthy

# JSON output (for tools/AI)
./scripts/pre-build-validate.sh --json
./scripts/docker-validate.sh --json
.git/hooks/pre-commit --json
```

### Quick Fixes

```bash
# Pre-build issues
./scripts/pre-build-validate.sh --fix

# Runtime issues - wait for services
./scripts/docker-validate.sh --wait --max-wait 120

# Runtime issues - restart unhealthy
./scripts/docker-validate.sh --auto-restart

# View logs for debugging
docker-compose logs [service]

# Rebuild specific service
docker-compose up -d --build [service]

# Full restart
docker-compose down && docker-compose up -d --build
```

---

## Expected Behavior

### ✅ Perfect Run

```
$ ./scripts/dev.sh

[DevSmith] Starting development environment...

[DevSmith] Step 1/3: Pre-build validation...
🔍 Pre-build validation...

✅ Pre-build validation PASSED

[DevSmith] Step 2/3: Building and starting services...
[+] Building ... (services build successfully)
[+] Running 6/6
 ✔ Container devsmith-postgres-1   Started
 ✔ Container devsmith-portal-1     Started
 ✔ Container devsmith-review-1     Started
 ✔ Container devsmith-logs-1       Started
 ✔ Container devsmith-analytics-1  Started
 ✔ Container devsmith-nginx-1      Started

[DevSmith] Step 3/3: Validating runtime health...

🐳 Docker validation (standard mode)...
[1/4] Checking container status...
[2/4] Checking health checks...
[3/4] Checking HTTP endpoints...

✅ Docker validation PASSED

[DevSmith] All services are healthy! 🎉

Services available at:
  • Portal:    http://localhost:8080
  • Review:    http://localhost:8081
  • Logs:      http://localhost:8082
  • Analytics: http://localhost:8083
  • Gateway:   http://localhost:3000
```

### ❌ Pre-Build Failure (Your Case)

```
$ ./scripts/dev.sh

[DevSmith] Starting development environment...

[DevSmith] Step 1/3: Pre-build validation...

════════════════════════════════════════════════════════════════
🔍 PRE-BUILD VALIDATION SUMMARY (completed in 1s)
════════════════════════════════════════════════════════════════

CHECK RESULTS:
  ✓ project structure      passed
  ✓ go modules             passed
  ✓ docker files           passed
  ✗ service files          failed

HIGH PRIORITY (Blocking builds): 1 issue(s)
  • [no_go_files] logs - cmd/logs
    No Go files found in cmd/logs (would cause: 'no Go files' build error)
    → Add Go source files to cmd/logs

QUICK FIXES:
  • Auto-fix issues: ./scripts/pre-build-validate.sh --fix

════════════════════════════════════════════════════════════════
✗ Pre-build validation FAILED
════════════════════════════════════════════════════════════════

[DevSmith] ❌ Pre-build validation failed!

To auto-fix issues: ./scripts/pre-build-validate.sh --fix
To see JSON output:  ./scripts/pre-build-validate.sh --json
```

**Copilot sees this and autonomously runs:**
```bash
./scripts/pre-build-validate.sh --fix
./scripts/dev.sh  # Try again
```

**No human intervention needed.**

---

## Benefits

### For Developers

1. **Faster debugging** - Clear error messages with fix suggestions
2. **No wasted time** - Catches issues before Docker builds
3. **Confidence** - All services validated before coding
4. **Clear errors** - No more cryptic Docker error messages

### For AI Assistants (Copilot)

1. **Structured data** - JSON output for easy parsing
2. **Auto-fix guidance** - Clear commands to run
3. **Autonomous fixing** - No human needed for standard issues
4. **Verification** - Re-run validations to confirm fixes

### For the Team

1. **Consistent quality** - All code validated before commit
2. **Fast onboarding** - New developers see clear errors
3. **Less noise** - Issues caught early, not in production
4. **Documentation** - Issues link to docs for learning

---

## Troubleshooting

### Pre-Build Validation Keeps Failing

**Check:**
```bash
# View detailed JSON output
./scripts/pre-build-validate.sh --json | jq '.issues[]'

# Try auto-fix
./scripts/pre-build-validate.sh --fix

# Check if files are gitignored
git status cmd/
```

### Docker Build Still Fails After Pre-Build Passes

**Possible causes:**
1. Files not committed to git
2. .dockerignore excludes service files
3. Dockerfile references wrong path

**Debug:**
```bash
git status
cat .dockerignore
cat cmd/[service]/Dockerfile
```

### Runtime Validation Shows Unhealthy

**Check:**
```bash
# View service logs
docker-compose logs [service] --tail=50

# Check database health
docker-compose ps postgres

# Wait longer
./scripts/docker-validate.sh --wait --max-wait 180

# Restart and try again
docker-compose restart [service]
```

---

## Documentation Index

- **Pre-Commit Hook:** `.docs/PRE-COMMIT-HOOK.md`
- **Docker Validation:** `.docs/DOCKER-VALIDATION.md`
- **Copilot Debugging:** `.docs/COPILOT-AUTONOMOUS-DEBUGGING.md`
- **Docker for Copilot:** `.docs/DOCKER-COPILOT-GUIDE.md`
- **Quick Start:** `.docs/DOCKER-QUICKSTART.md`

---

## Summary

DevSmith's **three-layer validation** ensures:

1. ✅ **Code quality** before commit (pre-commit hook)
2. ✅ **Structure validity** before build (pre-build-validate)
3. ✅ **Service health** before development (docker-validate)

**For you:** Run `./scripts/dev.sh` and everything is validated automatically.

**For Copilot:** Parse JSON, auto-fix issues, no human intervention needed.

**Result:** **Zero debugging time for standard configuration issues.**
