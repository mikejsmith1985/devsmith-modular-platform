# DevSmith Modular Platform

**AI-powered code analysis platform for effective code comprehension and quality review.**

DevSmith teaches developers to read and supervise AI-generated code through five distinct reading modes, each optimized for different comprehension goals and cognitive load management.

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/docker-required-2496ED.svg)](https://docker.com)

---

## 📑 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Production Deployment](#production-deployment)
- [Reading Modes](#reading-modes)
- [API Documentation](#api-documentation)
- [Monitoring & Observability](#monitoring--observability)
- [Troubleshooting](#troubleshooting)
- [Development](#development)
- [Documentation](#documentation)

---

## Overview

DevSmith addresses the critical challenge of the "Human-in-the-Loop" era: as AI generates more code, developers must shift from *writing* code to *reading, understanding, and validating* AI output.

### The Problem

- AI coding assistants (GitHub Copilot, Claude, OpenHands) generate code 10x faster
- Developers struggle to review and validate this output effectively
- Traditional code review tools aren't designed for AI-generated code
- Reading code is harder than writing it, yet rarely taught

### The Solution

DevSmith provides **5 AI-assisted reading modes** that teach developers to:
1. **Preview** - Assess code structure in 2-3 minutes
2. **Skim** - Understand abstractions in 5-7 minutes
3. **Scan** - Find specific patterns in 3-5 minutes
4. **Detailed** - Deep-dive algorithms in 10-15 minutes
5. **Critical** - Evaluate quality and identify issues in 15-20 minutes

Each mode is optimized for **cognitive load management** and builds transferable mental frameworks.

---

## Features

### Core Capabilities

✅ **5 Reading Modes** - AI-guided code analysis for different comprehension goals  
✅ **Local AI Models** - Ollama integration (no API costs, privacy-first)  
✅ **Multiple Models** - Mistral, CodeLlama, Llama2, DeepSeek support  
✅ **Circuit Breaker** - Graceful degradation when AI unavailable  
✅ **HTMX UI** - Fast, server-rendered interface  
✅ **Docker Ready** - Single-command deployment  
✅ **Health Checks** - Comprehensive monitoring of all components  
✅ **OpenTelemetry** - Distributed tracing to Jaeger  
✅ **Graceful Shutdown** - Zero data loss during restarts

### Quality & Reliability

✅ **Production-Ready** - Circuit breakers, health checks, graceful shutdown  
✅ **Error Handling** - User-friendly error templates with retry actions  
✅ **Observability** - Prometheus metrics + Jaeger tracing  
✅ **Test Coverage** - Unit, integration, and E2E tests  
✅ **API Documentation** - OpenAPI 3.0 specification  
✅ **Incident Runbooks** - Step-by-step troubleshooting guides

---

## Architecture

### Services

```
┌─────────────────────────────────────────────────────────────┐
│                    Nginx Gateway (port 3000)                 │
│                         /api/review/*                         │
└────────────────────┬────────────────────────────────────────┘
                     │
         ┌───────────┼───────────┐
         │           │           │
    ┌────▼────┐ ┌───▼────┐ ┌───▼────────┐
    │ Portal  │ │ Review │ │ Analytics  │
    │ :8080   │ │ :8081  │ │ :8083      │
    └────┬────┘ └───┬────┘ └───┬────────┘
         │          │           │
         └──────────┼───────────┘
                    │
         ┌──────────┼──────────┐
         │          │          │
    ┌────▼─────┐ ┌─▼────────┐ ┌──────────┐
    │ Postgres │ │  Ollama  │ │  Jaeger  │
    │  :5432   │ │  :11434  │ │  :16686  │
    └──────────┘ └──────────┘ └──────────┘
```

### Review Service Components

```
Review Service
├── Circuit Breaker (5 failures → open, 60s timeout)
├── Ollama Adapter (AI model integration)
├── 5 Mode Services (Preview, Skim, Scan, Detailed, Critical)
├── Health Checker (8 components monitored)
└── OpenTelemetry (distributed tracing)
```

### Technology Stack

- **Backend**: Go 1.21+ with Gin framework
- **Frontend**: Templ templates + HTMX + TailwindCSS
- **AI**: Ollama (local LLM inference)
- **Database**: PostgreSQL 15+
- **Observability**: Prometheus + Jaeger + OpenTelemetry
- **Deployment**: Docker + Docker Compose

---

## Quick Start

### Prerequisites

```bash
# Required
- Docker 20.10+
- Docker Compose 2.0+
- 16GB RAM (for Ollama with mistral:7b-instruct)
- 10GB disk space

# Optional
- 32GB RAM (for deepseek-coder-v2:16b)
- NVIDIA GPU (faster inference)
```

### Installation (5 minutes)

```bash
# 1. Clone repository
git clone https://github.com/mikejsmith1985/devsmith-modular-platform.git
cd devsmith-modular-platform

# 2. Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# 3. Pull AI model (choose based on RAM)
ollama pull mistral:7b-instruct        # 16GB RAM (recommended)
# OR
ollama pull deepseek-coder-v2:16b      # 32GB RAM (best quality)

# 4. Start platform
docker-compose up -d

# 5. Wait for services to initialize (~30 seconds)
./scripts/health-check-cli.sh --watch

# 6. Open in browser
open http://localhost:3000
```

### Verify Installation

```bash
# Check all services healthy
curl http://localhost:8081/health | jq

# Expected output:
# {
#   "status": "healthy",
#   "components": [
#     {"name": "database", "status": "healthy"},
#     {"name": "ollama_connectivity", "status": "healthy"},
#     {"name": "ollama_model", "status": "healthy"},
#     {"name": "preview_service", "status": "healthy"},
#     ... (8 components total)
#   ]
# }

# Test analysis
curl -X POST http://localhost:8081/api/review/modes/preview \
  -H "Content-Type: application/json" \
  -d '{"pasted_code":"package main\nfunc main() {}","model":"mistral:7b-instruct"}'

# Should return HTML with code analysis in ~5 seconds
```

---

## Production Deployment

### Environment Configuration

```bash
# Copy example configuration
cp .env.example .env

# Edit required variables
vim .env
```

**Required Environment Variables:**

```bash
# Database
DATABASE_URL=postgresql://devsmith:password@postgres:5432/devsmith

# Ollama
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_DEFAULT_MODEL=mistral:7b-instruct

# Review Service
REVIEW_PORT=8081
REVIEW_LOG_LEVEL=info

# Observability (optional)
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
ENABLE_TRACING=true
```

### Docker Deployment Checklist

- [ ] All environment variables configured in `.env`
- [ ] Ollama running with model pulled: `ollama list`
- [ ] PostgreSQL data volume persisted: `docker volume ls`
- [ ] Nginx reverse proxy configured: `docker/nginx/nginx.conf`
- [ ] Health checks passing: `./scripts/health-check-cli.sh`
- [ ] SSL certificates installed (production): `docker/nginx/ssl/`
- [ ] Backup strategy configured: `pg_dump` cron job
- [ ] Monitoring alerts configured: Prometheus + Grafana
- [ ] Incident runbook reviewed: `.docs/runbooks/review-service-incidents.md`

### Production Build

```bash
# Build all services
docker-compose build

# Start in production mode
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Verify health
./scripts/health-check-cli.sh --pr

# Check logs
docker-compose logs -f review
```

### Scaling Considerations

**Vertical Scaling:**
- Review service: 512MB RAM minimum, 2GB recommended
- Ollama: 8GB RAM (mistral:7b) to 32GB (deepseek-coder-v2:16b)
- PostgreSQL: 2GB RAM, 50GB disk

**Horizontal Scaling:**
- Review service: Stateless, can replicate (add load balancer)
- Ollama: Run multiple instances on different ports, load balance
- Database: Single primary, read replicas for analytics

**Performance Targets:**
- Preview Mode: <10s response time (P95)
- Critical Mode: <30s response time (P95)
- Health check: <1s response time
- Circuit breaker overhead: <5% latency increase

---

## Reading Modes

### 1. Preview Mode (2-3 minutes)

**Purpose:** Quick structural overview  
**Use Case:** Evaluating unfamiliar code before deep dive  

**AI Analysis Provides:**
- File/folder structure with descriptions
- Identified bounded contexts (e.g., "auth domain", "data layer")
- Technology stack detection
- Architectural pattern (layered, microservices, etc.)
- Entry points and dependencies

**Example:**
```bash
curl -X POST http://localhost:8081/api/review/modes/preview \
  -H "Content-Type: application/json" \
  -d '{"pasted_code":"...","model":"mistral:7b-instruct"}'
```

---

### 2. Skim Mode (5-7 minutes)

**Purpose:** Understand abstractions without implementation details  
**Use Case:** Learning API surface before using a library  

**AI Analysis Provides:**
- Function/method signatures with descriptions
- Interface definitions and purposes
- Data models (structs, entities)
- Key workflows with diagrams
- API endpoint catalog

---

### 3. Scan Mode (3-5 minutes)

**Purpose:** Targeted search for specific patterns  
**Use Case:** Finding security issues, debugging specific errors  

**AI Analysis Provides:**
- Semantic code search (not just string matching)
- Variable/function usage tracking
- Error source identification
- Pattern matching (e.g., "find all SQL queries")
- Context-aware suggestions

**Example:**
```bash
curl -X POST 'http://localhost:8081/api/review/modes/scan?query=find+validation' \
  -H "Content-Type: application/json" \
  -d '{"pasted_code":"...","model":"mistral:7b-instruct"}'
```

---

### 4. Detailed Mode (10-15 minutes)

**Purpose:** Deep understanding of algorithms  
**Use Case:** Comprehending complex logic before modification  

**AI Analysis Provides:**
- Line-by-line explanation
- Variable state tracking at each point
- Control flow analysis (if/else paths, loops)
- Algorithm identification (e.g., "implements binary search")
- Complexity analysis (time/space)
- Edge case identification

---

### 5. Critical Mode (15-20 minutes) ⭐

**Purpose:** Quality evaluation and issue detection  
**Use Case:** **Human-in-the-Loop review of AI-generated code** (primary use case)

**AI Analysis Identifies:**

**Architecture Issues:**
- Bounded context violations
- Layer mixing (controller calling repository directly)
- Missing abstractions
- Tight coupling

**Security Issues:**
- SQL injection risks
- Unvalidated input
- Secrets in code
- Auth/authorization gaps

**Code Quality:**
- Go idiom violations
- Error handling issues
- Scope problems (unnecessary globals)
- Missing documentation

**Performance:**
- N+1 query problems
- Inefficient algorithms
- Missing database indexes

**Testing:**
- Untested code paths
- Missing error case tests

**Example:**
```bash
curl -X POST http://localhost:8081/api/review/modes/critical \
  -H "Content-Type: application/json" \
  -d '{"pasted_code":"...","model":"mistral:7b-instruct"}'
```

---

## API Documentation

### OpenAPI Specification

Full API documentation available at:
- **File:** [docs/openapi-review.yaml](docs/openapi-review.yaml)
- **UI:** Swagger UI (coming soon)

### Key Endpoints

#### Health Check
```http
GET /health

Response: 200 OK
{
  "status": "healthy",
  "timestamp": "2025-11-02T15:00:00Z",
  "components": [...]
}
```

#### Preview Mode Analysis
```http
POST /api/review/modes/preview
Content-Type: application/json

{
  "pasted_code": "package main\nfunc main() {}",
  "model": "mistral:7b-instruct"
}

Response: 200 OK (HTML for HTMX)
```

#### Critical Mode Analysis
```http
POST /api/review/modes/critical
Content-Type: application/json

{
  "pasted_code": "...",
  "model": "mistral:7b-instruct"
}

Response: 200 OK (HTML with issue list)
```

#### Debug Trace (OpenTelemetry)
```http
GET /debug/trace

Response: 200 OK
{
  "trace_id": "1234...",
  "span_id": "abcd...",
  "jaeger_query": "http://localhost:16686/..."
}
```

---

## Monitoring & Observability

### Health Checks

```bash
# Quick health check
curl http://localhost:8081/health | jq '.status'

# Detailed component health
curl http://localhost:8081/health | jq '.components[] | select(.status != "healthy")'

# Continuous monitoring
./scripts/health-check-cli.sh --watch
```

**Health Check Components (8 total):**
1. `database` - PostgreSQL connectivity and schema presence
2. `ollama_connectivity` - Ollama service reachable
3. `ollama_model` - Configured model available
4. `preview_service` - Preview mode operational
5. `skim_service` - Skim mode operational
6. `scan_service` - Scan mode operational
7. `detailed_service` - Detailed mode operational
8. `critical_service` - Critical mode operational

### Distributed Tracing (Jaeger)

```bash
# Open Jaeger UI
open http://localhost:16686

# Search for traces:
# - Service: devsmith-review
# - Operation: review.modes.preview (or skim, scan, detailed, critical)
# - Look for: high latency, errors, circuit breaker events

# Generate test trace
curl http://localhost:8081/debug/trace
```

**Trace Spans Include:**
- HTTP request handling
- AI model inference (Ollama calls)
- Circuit breaker state changes
- Database queries (if applicable)
- Error handling and retries

### Circuit Breaker Monitoring

```bash
# Check circuit breaker state
docker-compose logs review | grep -i "circuit"

# States:
# CLOSED (normal) - Requests flow through
# OPEN (protecting) - Requests fail-fast for 60s
# HALF_OPEN (testing) - One test request to check recovery

# Thresholds:
# - maxFailures: 5 consecutive failures → OPEN
# - timeout: 60s until auto-recovery attempt
```

### Logs

```bash
# View recent logs
docker-compose logs review --tail=100 -f

# Search for errors
docker-compose logs review | grep -i "error\|fatal\|panic"

# Filter by component
docker-compose logs review | grep "preview_service"
docker-compose logs review | grep "circuit_breaker"
```

---

## Troubleshooting

### Quick Diagnostics

```bash
# 1. Check service health
curl http://localhost:8081/health | jq

# 2. Check Ollama
curl http://localhost:11434/api/tags
ollama list

# 3. Check logs
docker-compose logs review --tail=50

# 4. Check circuit breaker
docker-compose logs review | grep -i "circuit"

# 5. Run health check tool
./scripts/health-check-cli.sh
```

### Common Issues

#### Issue: 503 Service Unavailable

**Symptoms:** All mode endpoints return 503  
**Likely Cause:** Ollama down or circuit breaker open  
**Fix:**
```bash
# Check Ollama
systemctl status ollama  # or: ps aux | grep ollama
ollama list

# Restart Ollama if needed
sudo systemctl restart ollama

# Wait for circuit breaker reset (60s) or restart review
docker-compose restart review
```

#### Issue: High Latency (>30s)

**Symptoms:** Requests timeout or take very long  
**Likely Cause:** Ollama overloaded or wrong model  
**Fix:**
```bash
# Check Ollama resource usage
top -p $(pgrep ollama)

# Switch to faster model in UI or:
ollama pull mistral:7b-instruct  # Faster than 13B models

# Reduce concurrent requests (nginx rate limiting)
```

#### Issue: Health Check Failing

**Symptoms:** `/health` returns "unhealthy" or "degraded"  
**Likely Cause:** Database or Ollama connectivity issue  
**Fix:**
```bash
# Identify failing component
curl http://localhost:8081/health | jq '.components[] | select(.status != "healthy")'

# Fix database
docker-compose restart postgres
sleep 10

# Fix Ollama
sudo systemctl restart ollama

# Verify
curl http://localhost:8081/health
```

### Incident Response Runbook

For detailed troubleshooting, see:
- **[Review Service Incident Runbook](.docs/runbooks/review-service-incidents.md)**

Covers:
- Health check failures
- Ollama unavailable
- Circuit breaker open
- High latency
- Container startup issues
- Memory leaks
- Common error messages
- Escalation paths

---

## Development

### Development Setup

```bash
# Clone repository
git clone https://github.com/mikejsmith1985/devsmith-modular-platform.git
cd devsmith-modular-platform

# Install dependencies
go mod download

# Install dev tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/a-h/templ/cmd/templ@latest
npm install  # For Playwright E2E tests

# Start development environment
docker-compose -f docker-compose.dev.yml up -d

# Run tests
go test ./...
go test -race -cover ./...
npx playwright test
```

### Pre-commit Hook

Every commit is automatically validated. See [Pre-Commit Hook Guide](.docs/PRE-COMMIT-HOOK.md) for details.

**Validation checks:**
- ✅ Code formatting (`go fmt`)
- ✅ Import cleanup (`goimports`)
- ✅ Static analysis (`go vet`)
- ✅ Linting (`golangci-lint`)
- ✅ Build validation (`go build ./cmd/...`)
- ✅ Tests (`go test -short ./...`)

### Running Tests

```bash
# Unit tests
go test ./internal/review/...

# With coverage
go test -cover ./internal/review/...
go test -coverprofile=coverage.out ./internal/review/...
go tool cover -html=coverage.out

# Integration tests
go test -tags=integration ./tests/integration/...

# E2E tests
npx playwright test
npx playwright test --headed  # Watch browser
npx playwright test --debug   # Debug mode

# Benchmarks
go test -bench=. -benchmem ./internal/review/...
```

### Code Generation

```bash
# Generate Templ templates
templ generate

# Watch for changes
templ generate --watch
```

### Local Development URLs

- Review Service: http://localhost:8081
- Portal Service: http://localhost:8080
- Analytics Service: http://localhost:8083
- Nginx Gateway: http://localhost:3000
- Jaeger UI: http://localhost:16686
- Ollama API: http://localhost:11434

---

## Documentation

### For Developers
- **[Architecture](ARCHITECTURE.md)** - System design and coding standards
- **[TDD Workflow](DevsmithTDD.md)** - Test-driven development approach
- **[Pre-Commit Hook Guide](.docs/PRE-COMMIT-HOOK.md)** - Validation output interpretation
- **[Workflow Guide](.docs/WORKFLOW-GUIDE.md)** - Development process

### For Operators
- **[Incident Runbook](.docs/runbooks/review-service-incidents.md)** - Troubleshooting guide
- **[OpenAPI Spec](docs/openapi-review.yaml)** - API documentation
- **[Health Check Guide](scripts/HEALTH_CHECK_GUIDE.md)** - Monitoring guide

### For AI Agents
- **[Copilot Instructions](.github/copilot-instructions.md)** - Implementation guide
- **[Issue Templates](.docs/issues/)** - Feature specifications

### Additional Resources
- **[Requirements](Requirements.md)** - Platform requirements and philosophy
- **[Troubleshooting](.docs/TROUBLESHOOTING.md)** - Common issues
- **[Activity Log](.docs/devlog/copilot-activity.md)** - Development history

---

## Contributing

We welcome contributions! Please:

1. Read [ARCHITECTURE.md](ARCHITECTURE.md) for coding standards
2. Follow [DevsmithTDD.md](DevsmithTDD.md) for test-driven development
3. Ensure pre-commit hook passes (automatic validation)
4. Create GitHub issue before major changes
5. Submit PR with tests and documentation

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

## Support

- **Issues:** [GitHub Issues](https://github.com/mikejsmith1985/devsmith-modular-platform/issues)
- **Documentation:** [.docs/](.docs/) directory
- **Runbooks:** [.docs/runbooks/](.docs/runbooks/)

---

## Acknowledgments

Built with:
- [Go](https://golang.org) - Backend language
- [Gin](https://gin-gonic.com) - HTTP framework
- [Templ](https://templ.guide) - Type-safe templates
- [HTMX](https://htmx.org) - Dynamic UI without JavaScript frameworks
- [Ollama](https://ollama.com) - Local LLM inference
- [TailwindCSS](https://tailwindcss.com) + [DaisyUI](https://daisyui.com) - Styling
- [OpenTelemetry](https://opentelemetry.io) - Observability
- [Jaeger](https://www.jaegertracing.io) - Distributed tracing

---

**Version:** 1.0.0  
**Last Updated:** 2025-11-02  
**Status:** Production Ready ✅
