# DevSmith Health Check CLI

Comprehensive health check tool for the DevSmith platform, providing real-time diagnostics for Docker containers, services, and database connectivity.

## Features

### Phase 1: Core Health Checks
- **Docker Container Validation:** Ensures all expected containers are running
- **HTTP Health Endpoint Checks:** Validates service availability through their health endpoints
- **Database Connectivity:** Tests PostgreSQL connection and responsiveness
- **Multiple Output Formats:** Human-readable and JSON output
- **Structured Reporting:** Detailed status, timing, and error information

### Phase 2: Advanced Diagnostics (Default: Enabled)
- **Gateway Routing Validation:** Parses nginx.conf and validates all configured routes respond correctly
- **Performance Metrics Collection:** Measures response times across all services, identifies slow endpoints
- **Service Interdependency Checks:** Validates dependency chains (e.g., review depends on portal + logs)
- **Performance Baselines:** Compares current performance against thresholds (fast < 100ms, slow > 1s)

## Usage

### Basic Usage (Human-Readable Output)

```bash
# Full health check with Phase 2 diagnostics (default)
go run cmd/healthcheck/main.go

# Quick check (Phase 1 only)
go run cmd/healthcheck/main.go --advanced=false
```

### JSON Output

```bash
# Full diagnostics JSON
go run cmd/healthcheck/main.go --format=json

# Quick check JSON
go run cmd/healthcheck/main.go --format=json --advanced=false
```

### Build and Run

```bash
# Build
go build -o healthcheck cmd/healthcheck/main.go

# Run
./healthcheck
./healthcheck --format=json
```

## Output Example

### Human-Readable Format

```
═══════════════════════════════════════════════════════════════
  📊 DevSmith Platform Health Check
═══════════════════════════════════════════════════════════════

Environment: docker
Hostname:    localhost
Go Version:  go1.23.5
Timestamp:   2025-01-29 14:30:00

Overall Status: ✓ pass

Summary:
  Total Checks:  10
  ✓ Passed:      9
  ⚠ Warnings:    1
  Duration:      1.2s

Detailed Results:
───────────────────────────────────────────────────────────────

✓ docker_containers
  Status:   pass
  Message:  All 6 services running
  Duration: 234ms

✓ http_gateway
  Status:   pass
  Message:  HTTP 200 OK
  Duration: 45ms
  Details:
    url: http://localhost:3000/
    status_code: 200
    response_time_ms: 43

...
```

### JSON Format

```json
{
  "status": "pass",
  "timestamp": "2025-01-29T14:30:00Z",
  "duration": 1200000000,
  "checks": [
    {
      "name": "docker_containers",
      "status": "pass",
      "message": "All 6 services running",
      "duration": 234000000,
      "details": {
        "expected": 6,
        "running": 6,
        "missing": []
      },
      "timestamp": "2025-01-29T14:30:00Z"
    }
  ],
  "summary": {
    "total": 10,
    "passed": 9,
    "warned": 1,
    "failed": 0,
    "unknown": 0
  },
  "system_info": {
    "environment": "docker",
    "hostname": "localhost",
    "go_version": "go1.23.5",
    "timestamp": "2025-01-29T14:30:00Z"
  }
}
```

## Exit Codes

- **0:** All checks passed or warnings only
- **1:** One or more checks failed

## Configuration

### Environment Variables

- `DATABASE_URL`: PostgreSQL connection string (default: `postgres://devsmith:devsmith@localhost:5432/devsmith?sslmode=disable`)
- `ENVIRONMENT`: Deployment environment (e.g., `docker`, `local`, `production`)
- `NGINX_CONFIG_PATH`: Path to nginx.conf for gateway validation (default: `docker/nginx/nginx.conf`)

## Integration with Logs Service

The health check is also integrated into the DevSmith Logs service:

- **API Endpoint:** `GET /api/logs/healthcheck` (JSON)
- **Dashboard UI:** `GET /healthcheck` (HTML)

### API Usage

```bash
# Get full health check as JSON (Phase 1 + Phase 2)
curl http://localhost:8082/api/logs/healthcheck

# Get human-readable format
curl http://localhost:8082/api/logs/healthcheck?format=human

# Quick check only (Phase 1)
curl http://localhost:8082/api/logs/healthcheck?advanced=false

# Dashboard UI with Phase 2
open http://localhost:8082/healthcheck

# Dashboard UI quick mode
open http://localhost:8082/healthcheck?advanced=false
```

## Architecture

The health check system is built with:

- **`internal/healthcheck/`**: Core health check logic
  - `types.go`: Data structures and interfaces
  - `runner.go`: Health check orchestration
  - `docker.go`: Docker container validation
  - `http.go`: HTTP endpoint checks
  - `database.go`: Database connectivity checks
  - `formatter.go`: Output formatting

- **`cmd/healthcheck/`**: Standalone CLI tool

- **Integration**: Embedded in Logs service for centralized monitoring

## Testing

```bash
# Run all tests
go test ./internal/healthcheck/... -v

# Run with coverage
go test ./internal/healthcheck/... -cover
```

## Development

### Adding a New Check

1. Implement the `Checker` interface:

```go
type CustomChecker struct {
    CheckName string
}

func (c *CustomChecker) Name() string {
    return c.CheckName
}

func (c *CustomChecker) Check() CheckResult {
    // Perform your check
    return CheckResult{
        Name:      c.CheckName,
        Status:    StatusPass,
        Message:   "Check passed",
        Duration:  time.Since(start),
        Timestamp: start,
    }
}
```

2. Add to runner in `cmd/healthcheck/main.go`:

```go
runner.AddChecker(&CustomChecker{
    CheckName: "my_custom_check",
})
```

## Future Enhancements (Phase 3)

- [ ] Historical trend analysis (store health check results over time)
- [ ] Alert integration (email/Slack notifications on failures)
- [ ] Scheduled health check runs (cron-based monitoring)
- [ ] Export to monitoring systems (Prometheus/Grafana)
- [ ] Performance regression detection (compare against historical baselines)
- [ ] Custom health check plugins
- [ ] Multi-environment support (dev/staging/prod)

