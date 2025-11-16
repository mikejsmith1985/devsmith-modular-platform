# Review Service Incident Response Runbook

**Service:** DevSmith Review Service  
**Owner:** Platform Team  
**Last Updated:** 2025-11-02  
**Escalation:** Check #devsmith-alerts Slack channel

---

## Quick Reference

| Symptom | Likely Cause | First Action | Runbook Section |
|---------|--------------|--------------|-----------------|
| 503 errors on all modes | Ollama down | Check Ollama connectivity | [Ollama Unavailable](#ollama-unavailable) |
| Circuit breaker open | Repeated Ollama failures | Wait 60s or restart Ollama | [Circuit Breaker Open](#circuit-breaker-open) |
| Health check failing | Database or Ollama issue | Check `/health` details | [Health Check Failures](#health-check-failures) |
| High latency (>30s) | Ollama overloaded | Check Ollama resource usage | [High Latency](#high-latency) |
| Container not starting | Missing env vars or DB | Check logs and config | [Container Issues](#container-startup-issues) |
| Memory leak | Service memory growing | Restart service, check for leak | [Memory Leak](#memory-leak-detection) |

---

## Health Check Failures

### Symptoms
- `/health` returns 503 status
- Component status shows "unhealthy" or "degraded"
- Logs show health check failures

### Diagnosis

**Step 1: Check health endpoint**
```bash
# Get health status (local)
curl -v http://localhost:8081/health | jq

# Via nginx gateway
curl -v http://localhost:3000/api/review/health | jq
```

**Step 2: Identify failing component**
```bash
# Parse health response
curl -s http://localhost:8081/health | jq '.components[] | select(.status != "healthy")'

# Example output:
# {
#   "name": "ollama_connectivity",
#   "status": "unhealthy",
#   "message": "Failed to connect to Ollama: dial tcp 127.0.0.1:11434: connect: connection refused"
# }
```

**Step 3: Check service logs**
```bash
# Docker logs (last 50 lines)
docker-compose logs review --tail=50

# Search for errors
docker-compose logs review | grep -i "error\|fatal\|panic"
```

### Resolution by Component

#### Database Unhealthy
```bash
# Check postgres container
docker-compose ps postgres

# Check connection
docker-compose exec postgres psql -U devsmith -d devsmith -c "\dt reviews.*"

# Restart database if needed
docker-compose restart postgres

# Wait for health
sleep 10
curl http://localhost:8081/health | jq '.components[] | select(.name == "database")'
```

#### Ollama Connectivity Unhealthy
See [Ollama Unavailable](#ollama-unavailable) section.

#### Service Status Unhealthy
```bash
# Check if services are initialized
curl -s http://localhost:8081/health | jq '.components[] | select(.name | contains("service"))'

# Restart review service
docker-compose restart review

# Verify health
sleep 5
curl http://localhost:8081/health
```

---

## Ollama Unavailable

### Symptoms
- All mode endpoints return 503
- Error message: "Ollama service unavailable"
- Health check shows "ollama_connectivity: unhealthy"

### Diagnosis

**Step 1: Check Ollama service**
```bash
# Check if Ollama is running
systemctl status ollama   # Linux with systemd
ps aux | grep ollama       # macOS/Linux

# Check Ollama endpoint
curl -v http://localhost:11434/api/tags

# Expected: 200 OK with JSON list of models
# If connection refused: Ollama not running
# If timeout: Network issue
```

**Step 2: Check Ollama logs**
```bash
# Journalctl (Linux with systemd)
journalctl -u ollama -n 100 --no-pager

# Docker logs (if Ollama in container)
docker logs ollama

# Look for: crashes, OOM kills, model loading failures
```

**Step 3: Check model availability**
```bash
# List installed models
ollama list

# Expected output:
# NAME                         ID              SIZE      MODIFIED
# mistral:7b-instruct          1234abcd        4.1 GB    2 days ago

# If model missing, pull it:
ollama pull mistral:7b-instruct
```

### Resolution

**Option A: Restart Ollama**
```bash
# Systemd
sudo systemctl restart ollama
sudo systemctl status ollama

# Docker
docker-compose restart ollama  # If using docker-compose for Ollama

# Verify
sleep 5
curl http://localhost:11434/api/tags
```

**Option B: Reinstall Model**
```bash
# If model corrupted or missing
ollama rm mistral:7b-instruct
ollama pull mistral:7b-instruct

# Verify model works
curl http://localhost:11434/api/generate -d '{
  "model": "mistral:7b-instruct",
  "prompt": "Hello",
  "stream": false
}'
```

**Option C: Check Resource Usage**
```bash
# Ollama requires significant RAM
free -h  # Linux
vm_stat  # macOS

# If low memory:
# 1. Close other applications
# 2. Use smaller model (mistral:7b instead of codellama:13b)
# 3. Add swap space (Linux)
```

---

## Circuit Breaker Open

### Symptoms
- Requests fail with "Circuit breaker is OPEN"
- Multiple consecutive Ollama failures logged
- Service still returns 503 even after Ollama fixed

### Diagnosis

**Step 1: Check circuit breaker status**
```bash
# Circuit breaker opens after 5 consecutive failures
docker-compose logs review | grep -i "circuit"

# Example:
# [WARN] Preview circuit breaker opened after 5 failures
# [INFO] Circuit breaker will attempt reset in 60s
```

**Step 2: Verify underlying issue resolved**
```bash
# Check Ollama is working
curl http://localhost:11434/api/tags

# Check health endpoint
curl http://localhost:8081/health | jq '.components[] | select(.name | contains("ollama"))'
```

### Resolution

**Option A: Wait for auto-recovery (60 seconds)**
```bash
# Circuit breaker automatically attempts reset after 60s
# Monitor logs:
docker-compose logs -f review | grep -i "circuit"

# Expected:
# [INFO] Circuit breaker attempting reset
# [INFO] Preview service: test request successful
# [INFO] Circuit breaker transitioned to HALF_OPEN
# [INFO] Circuit breaker transitioned to CLOSED
```

**Option B: Restart service to reset immediately**
```bash
# Only if underlying issue is fixed
docker-compose restart review

# Verify circuit closed
curl -s http://localhost:8081/health | jq '.components[] | select(.name | contains("service"))'
```

**Option C: Adjust circuit breaker thresholds (if recurring)**
```bash
# Edit internal/review/services/circuit_breaker.go
# Current: maxFailures=5, timeout=60s
# Consider: maxFailures=10 for flaky Ollama instances

# Rebuild
docker-compose build review
docker-compose up -d review
```

---

## High Latency

### Symptoms
- Requests take >30 seconds
- Users report timeouts
- CPU usage high on Ollama

### Diagnosis

**Step 1: Check request timings**
```bash
# Test Preview mode latency
time curl -X POST http://localhost:8081/api/review/modes/preview \
  -H "Content-Type: application/json" \
  -d '{"pasted_code":"package main\nfunc main() {}","model":"mistral:7b-instruct"}'

# Expected: <10s for small code
# If >30s: Performance issue
```

**Step 2: Check Ollama resource usage**
```bash
# CPU and memory
top -p $(pgrep ollama)  # Linux
top | grep ollama       # macOS

# GPU usage (if applicable)
nvidia-smi  # NVIDIA GPU

# Look for:
# - CPU >90%: Ollama overloaded
# - Memory near limit: Model too large
# - GPU 100%: GPU bottleneck
```

**Step 3: Check concurrent requests**
```bash
# Count in-flight requests
docker-compose logs review | grep "Starting.*mode analysis" | tail -20

# If >5 concurrent: Review service overloaded
```

### Resolution

**Option A: Reduce concurrent requests**
```bash
# Add rate limiting (nginx)
# Edit docker/nginx/nginx.conf

location /api/review/ {
    limit_req zone=review_api burst=5;
    proxy_pass http://review:8081;
}

# Restart nginx
docker-compose restart nginx
```

**Option B: Use faster model**
```bash
# Switch to smaller model in UI
# mistral:7b-instruct (default, 4GB RAM)
# or llama2:13b (slower but more accurate)

# Update default in session_form.templ if needed
```

**Option C: Scale Ollama horizontally**
```bash
# Run multiple Ollama instances (advanced)
# Edit docker-compose.yml to add:
#   ollama2:
#     image: ollama/ollama
#     ports: ["11435:11434"]

# Update review service to load balance between instances
```

---

## Container Startup Issues

### Symptoms
- `docker-compose up` fails
- Container exits immediately
- Health check never passes

### Diagnosis

**Step 1: Check container status**
```bash
docker-compose ps

# Look for:
# - Exit code 1: Configuration error
# - Exit code 137: OOM killed
# - Restarting loop: Crash loop
```

**Step 2: Check logs**
```bash
# Get last 100 lines
docker-compose logs review --tail=100

# Common errors:
# "failed to connect to database": DB not ready
# "missing environment variable": .env incomplete
# "panic": Code bug
```

**Step 3: Check configuration**
```bash
# Verify .env file
cat .env | grep -E "REVIEW_|OLLAMA_|DATABASE_"

# Required variables:
# OLLAMA_BASE_URL=http://localhost:11434
# DATABASE_URL=postgresql://devsmith:...
# REVIEW_PORT=8081
```

### Resolution

**Option A: Wait for dependencies**
```bash
# Database may take time to initialize
docker-compose up -d postgres
sleep 10  # Wait for postgres ready

# Then start review
docker-compose up -d review
```

**Option B: Fix environment variables**
```bash
# Copy example if missing
cp .env.example .env

# Edit required values
vim .env

# Restart with new config
docker-compose down
docker-compose up -d
```

**Option C: Check disk space**
```bash
df -h  # Linux/macOS

# If low (<1GB free):
# - Remove old Docker images: docker image prune -a
# - Remove old logs: truncate -s 0 /var/lib/docker/containers/*/*-json.log
```

---

## Memory Leak Detection

### Symptoms
- Memory usage grows over time
- Eventually OOM killed
- Performance degrades

### Diagnosis

**Step 1: Monitor memory usage**
```bash
# Container memory
docker stats review --no-stream

# Expected: <500MB for review service
# If >1GB: Potential leak
```

**Step 2: Check for goroutine leaks**
```bash
# Enable pprof in review service
curl http://localhost:8081/debug/pprof/goroutine?debug=2 > goroutines.txt

# Analyze
grep -c "goroutine" goroutines.txt

# Expected: <100 goroutines
# If >1000: Goroutine leak
```

**Step 3: Check for memory leaks**
```bash
# Take heap dump
curl http://localhost:8081/debug/pprof/heap > heap.prof

# Analyze with pprof
go tool pprof heap.prof
# In pprof: top10, list <function>
```

### Resolution

**Option A: Restart service (immediate)**
```bash
docker-compose restart review

# Monitor memory after restart
watch -n 5 'docker stats review --no-stream'
```

**Option B: Investigate leak (long-term)**
```bash
# Enable memory profiling
# Add to cmd/review/main.go:
# import _ "net/http/pprof"
# go func() { http.ListenAndServe("localhost:6060", nil) }()

# Rebuild and profile
docker-compose build review
docker-compose up -d review

# Generate profile after load
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap
```

**Option C: Add memory limits**
```bash
# Edit docker-compose.yml
services:
  review:
    deploy:
      resources:
        limits:
          memory: 2G
        reservations:
          memory: 512M

# Restart
docker-compose up -d review
```

---

## Common Error Messages

### "Context key 'ollama-client' not found"

**Cause:** Ollama client not initialized in request context.

**Fix:**
```bash
# Check middleware order in cmd/review/main.go
# OllamaMiddleware must run before handlers

# Verify logs
docker-compose logs review | grep "initialized Ollama client"

# If missing, restart service
docker-compose restart review
```

### "Failed to execute Ollama request: circuit breaker is OPEN"

**Cause:** Circuit breaker protecting from cascading failures.

**Fix:** See [Circuit Breaker Open](#circuit-breaker-open) section.

### "Database schema 'reviews' not found"

**Cause:** Database migrations not run.

**Fix:**
```bash
# Run migrations
docker-compose exec postgres psql -U devsmith -d devsmith < db/migrations/001_create_reviews_schema.sql

# Verify
docker-compose exec postgres psql -U devsmith -d devsmith -c "\dn reviews"

# Restart service
docker-compose restart review
```

### "Failed to parse analysis result: unexpected end of JSON"

**Cause:** Ollama returned incomplete response or service crashed mid-stream.

**Fix:**
```bash
# Check Ollama stability
curl -X POST http://localhost:11434/api/generate \
  -d '{"model":"mistral:7b-instruct","prompt":"test","stream":false}'

# If Ollama works, check review service logs
docker-compose logs review | grep -A10 "Failed to parse"

# Likely timeout or OOM - check resources
```

---

## Graceful Shutdown

The review service supports graceful shutdown (SIGTERM handling).

### During Deployment

```bash
# Send SIGTERM (docker-compose does this automatically)
docker-compose stop review

# Service will:
# 1. Stop accepting new requests
# 2. Wait up to 30s for in-flight requests
# 3. Close database connections
# 4. Exit

# Monitor logs
docker-compose logs review | tail -20
# Expected: "Graceful shutdown complete"
```

### During Emergency

```bash
# Forceful shutdown (SIGKILL) - last resort
docker-compose kill review

# Note: In-flight requests will be lost
# Use only when graceful shutdown hangs
```

---

## Escalation Path

### Level 1: Self-Service (This Runbook)
- Health check failures → Check database and Ollama
- Circuit breaker open → Wait or restart
- High latency → Check resources

### Level 2: On-Call Engineer
Contact if:
- Issue persists after following runbook
- Data corruption suspected
- Multiple services affected

**Contact:** #devsmith-oncall Slack channel

### Level 3: Platform Team
Escalate if:
- Security incident
- Data breach
- Architecture changes needed

**Contact:** @platform-team in Slack

---

## Monitoring & Alerts

### Key Metrics

**Health Check:**
```bash
# Continuous health monitoring
watch -n 10 'curl -s http://localhost:8081/health | jq ".status"'
```

**Jaeger Traces:**
```bash
# View traces in Jaeger UI
open http://localhost:16686

# Search for: service=devsmith-review
# Look for: high latency, errors, retries
```

**Circuit Breaker State:**
```bash
# Monitor circuit breaker state changes
docker-compose logs -f review | grep -i "circuit.*transition"
```

### Alerting Thresholds

**Critical Alerts (Immediate Action):**
- Health check returns 503 for >5 minutes
- Circuit breaker open for >10 minutes
- Memory usage >90%
- Request latency P95 >60s

**Warning Alerts (Monitor):**
- Health check degraded
- Request latency P95 >30s
- Error rate >5%
- Circuit breaker flapping (open/close within 5 min)

---

## Post-Incident Review

After resolving an incident:

1. **Document what happened**
   - Timeline of events
   - Root cause analysis
   - Actions taken

2. **Update runbook**
   - Add new error patterns
   - Improve diagnostics
   - Clarify resolution steps

3. **File issue**
   - GitHub issue with "incident" label
   - Link to logs and traces
   - Propose improvements

4. **Share learnings**
   - Post-mortem in #devsmith channel
   - Update on-call rotation knowledge

---

## Useful Commands Cheatsheet

```bash
# Health checks
curl http://localhost:8081/health | jq
curl http://localhost:3000/api/review/health | jq  # Via nginx

# Ollama checks
curl http://localhost:11434/api/tags
ollama list

# Logs
docker-compose logs review --tail=100 -f
docker-compose logs review | grep ERROR

# Container management
docker-compose ps
docker-compose restart review
docker-compose up -d --build review

# Database checks
docker-compose exec postgres psql -U devsmith -d devsmith
\dt reviews.*
\q

# Traces
open http://localhost:16686  # Jaeger UI
curl http://localhost:8081/debug/trace  # Generate test trace

# Performance
docker stats review --no-stream
curl -o- http://localhost:8081/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

---

**Document Version:** 1.0  
**Last Updated:** 2025-11-02  
**Maintained By:** Platform Team  
**Feedback:** Create issue with "runbook" label
