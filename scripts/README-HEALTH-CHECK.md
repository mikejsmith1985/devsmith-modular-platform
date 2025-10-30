# DevSmith Health Check CLI

**Fast system diagnostics tool for monitoring and troubleshooting services during development.**

## Quick Start

```bash
# Basic health check
./scripts/health-check-cli.sh

# JSON output for parsing
./scripts/health-check-cli.sh --json

# Continuous monitoring (while developing)
./scripts/health-check-cli.sh --watch

# Get help
./scripts/health-check-cli.sh --help
```

## For Copilot

**Use this instead of `docker-validate.sh` for quick diagnostics while implementing features.**

The health check CLI provides:
- ✅ **Fast diagnostics** (< 1 second)
- ✅ **JSON output** for programmatic parsing
- ✅ **Watch mode** for real-time monitoring during development
- ✅ **Phase 1, 2, 3 checks** (containers, routing, security, auto-repair)
- ✅ **No frontend required** (works standalone from CLI)

## Usage Examples

### Example 1: Quick Health Check Before Starting Work

```bash
./scripts/health-check-cli.sh
```

**Output:**
```
═══════════════════════════════════════════════════════════════
  DevSmith Platform Health Check
═══════════════════════════════════════════════════════════════

Overall Status: ✓ pass

Summary:
  Total Checks:  9
  ✓ Passed:      9
  Duration:      890ms
```

### Example 2: Monitor Health While Developing

**Terminal 1 - Start monitoring:**
```bash
./scripts/health-check-cli.sh --watch
# Outputs health status every 5 seconds
```

**Terminal 2 - Make changes:**
```bash
vim internal/review/services/scan_service.go
```

**Terminal 3 - Rebuild and test:**
```bash
docker-compose up -d --build review
go test ./internal/review/services/...
```

Terminal 1 shows real-time health updates.

### Example 3: Troubleshoot Specific Service

```bash
# Get JSON output
./scripts/health-check-cli.sh --json

# Parse specific service status
./scripts/health-check-cli.sh --json | jq '.Checks[] | select(.Name=="review")'

# Check all failed services
./scripts/health-check-cli.sh --json | jq '.Checks[] | select(.Status!="pass")'

# Get summary only
./scripts/health-check-cli.sh --json | jq '.Summary'
```

### Example 4: Verify System Ready for PR

```bash
# Check if all systems healthy
STATUS=$(./scripts/health-check-cli.sh --json | jq -r '.Status')
if [[ "$STATUS" == "pass" ]]; then
  echo "✅ System healthy - ready for PR"
else
  echo "❌ System has issues - run: ./scripts/docker-validate.sh"
fi
```

## Options

| Option | Description | Example |
|--------|-------------|---------|
| `--json` | Output in JSON format | `./scripts/health-check-cli.sh --json` |
| `--watch` | Continuous monitoring mode (5s interval) | `./scripts/health-check-cli.sh --watch` |
| `--store` | Store results to database (TODO) | `./scripts/health-check-cli.sh --store` |
| `--advanced false` | Skip Phase 2 advanced checks | `./scripts/health-check-cli.sh --advanced false` |
| `--db-url URL` | Override database URL | `./scripts/health-check-cli.sh --db-url "postgres://..."` |
| `-h, --help` | Show help | `./scripts/health-check-cli.sh --help` |

## Health Check Phases

### Phase 1: Core Infrastructure (Always Runs)
- Docker container status
- HTTP health endpoints
- Database connectivity
- Response times

### Phase 2: Advanced Diagnostics (Default)
- Gateway routing validation (nginx.conf)
- Performance metrics
- Service interdependencies

### Phase 3: Intelligence (Integrated)
- Security scans (Trivy)
- Historical trends
- Auto-repair policies
- Health policies

## JSON Output Structure

```json
{
  "status": "pass|fail",
  "timestamp": "2025-10-30T07:20:35Z",
  "duration": 890000000,
  "summary": {
    "total_checks": 9,
    "passed": 9,
    "warnings": 0,
    "failed": 0
  },
  "checks": [
    {
      "name": "docker_containers",
      "status": "pass",
      "message": "All services running",
      "duration": 215000000
    }
    // ... more checks
  ]
}
```

## Common Issues & Solutions

### Error: "healthcheck binary not found"

```bash
# Build the binary
go build -o healthcheck ./cmd/healthcheck

# Verify
./healthcheck --format human
```

### Error: "Services may be down"

```bash
# Start containers
docker-compose up -d

# Wait a moment
sleep 5

# Try again
./scripts/health-check-cli.sh
```

### Error: "Connection refused"

```bash
# Check Docker is running
docker ps

# Check containers are healthy
docker-compose ps

# Rebuild if needed
docker-compose up -d --build
```

## Integration with Development Workflow

**Complete TDD + Health Check workflow:**

```bash
# 1. Start health monitor
./scripts/health-check-cli.sh --watch &
MONITOR_PID=$!

# 2. Write failing tests (RED)
vim internal/review/services/scan_service_test.go
go test ./internal/review/services/...

# 3. Implement feature (GREEN)
vim internal/review/services/scan_service.go
go test ./internal/review/services/...

# 4. Rebuild Docker
docker-compose up -d --build review

# 5. Health status updates in real-time (Terminal 1)
# (Monitor running in background)

# 6. Full validation before PR
./scripts/docker-validate.sh

# 7. Stop monitor
kill $MONITOR_PID
```

## For Copilot: When to Use Each Command

### Use `./scripts/health-check-cli.sh` When:
- ✅ Starting a new feature implementation
- ✅ Rebuilding a service with `docker-compose up -d --build`
- ✅ Verifying services before running tests
- ✅ Diagnosing why a specific service is failing
- ✅ Monitoring real-time health during development

### Use `./scripts/docker-validate.sh` When:
- ✅ Before creating a PR
- ✅ Full infrastructure validation
- ✅ Complete endpoint discovery
- ✅ Pre-commit comprehensive checks

## Troubleshooting with JSON

**List only failed checks:**
```bash
./scripts/health-check-cli.sh --json | jq '.Checks[] | select(.Status=="fail")'
```

**Get error messages:**
```bash
./scripts/health-check-cli.sh --json | jq '.Checks[] | select(.Status=="fail") | .Error'
```

**Check response times:**
```bash
./scripts/health-check-cli.sh --json | jq '.Checks[] | {Name: .Name, Duration: .Duration}'
```

**Export to file for later review:**
```bash
./scripts/health-check-cli.sh --json > /tmp/health-report.json
```

## Watch Mode Examples

**Basic continuous monitoring:**
```bash
./scripts/health-check-cli.sh --watch
# Outputs status every 5 seconds
# Press Ctrl+C to stop
```

**Watch with JSON (for parsing):**
```bash
./scripts/health-check-cli.sh --watch --json
```

**Monitor in background while working:**
```bash
./scripts/health-check-cli.sh --watch > /tmp/health-monitor.log 2>&1 &
# ... make changes ...
# tail -f /tmp/health-monitor.log  # View in another terminal
```

## For the Review Feature Implementation

**Copilot should follow this pattern:**

```bash
# 1. Verify system is healthy before starting
./scripts/health-check-cli.sh

# 2. Monitor while implementing
./scripts/health-check-cli.sh --watch &

# 3. Write tests (RED)
# ...

# 4. Implement feature (GREEN)
# ...

# 5. Rebuild Review service
docker-compose up -d --build review

# 6. Verify health check shows green
./scripts/health-check-cli.sh

# 7. If any issues, get details
./scripts/health-check-cli.sh --json | jq '.Checks[] | select(.Status!="pass")'
```

## Related Commands

```bash
# Quick Docker validation (full)
./scripts/docker-validate.sh

# Docker validation with JSON output
./scripts/docker-validate.sh --json

# Check container status
docker-compose ps

# View service logs
docker-compose logs <service> --tail=20

# Rebuild specific service
docker-compose up -d --build <service>

# Start services
docker-compose up -d

# Stop services
docker-compose down
```

## Performance Notes

- **Basic check:** ~900ms
- **With Phase 2 (advanced):** ~1s
- **Watch interval:** 5 seconds between checks
- **JSON parsing:** Instant (< 100ms)

## Future Enhancements

- [ ] `--store` flag to save results to database
- [ ] Real-time WebSocket dashboard updates
- [ ] Multi-check report comparison
- [ ] Performance trend analysis
- [ ] Custom alert thresholds

---

**For more information, see:**
- `.github/copilot-instructions.md` - Section 3.5 (Health Check CLI)
- `ARCHITECTURE.md` - Section 12 (Health Check Integration)
- `health-check-phase-3.plan.md` - Detailed Phase 3 implementation
