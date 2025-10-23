# Docker Validation Quick Start

> **TL;DR:** Run `./scripts/dev.sh` to start services with automatic validation. All services now have health checks and proper startup ordering.

---

## What Changed?

### ✅ Enhanced docker-compose.yml

1. **Added postgres health check** - Database now validates it's ready before services start
2. **All services use `condition: service_healthy`** - Proper startup ordering guaranteed
3. **Added `start_period` to all health checks** - Services have time to initialize
4. **Nginx waits for all backends** - Gateway only starts when services are healthy

### ✅ New Validation Script

`scripts/docker-validate.sh` - Comprehensive validation of:
- Container status (running/stopped)
- Health checks (healthy/unhealthy)
- HTTP endpoints (200 OK, not 404/500)
- Port bindings (correct mappings)

### ✅ Integrated Workflow

`scripts/dev.sh` now:
1. Starts all services
2. Waits for health checks to pass
3. Validates HTTP endpoints are working
4. Shows clear error messages if anything fails

### ✅ Optional Monitoring

`docker-compose.monitoring.yml` adds:
- **Uptime Kuma** - Visual monitoring dashboard
- **Autoheal** - Automatically restarts unhealthy containers

---

## Quick Commands

```bash
# Start everything with validation (recommended)
./scripts/dev.sh

# Validate running containers
./scripts/docker-validate.sh

# Wait for services to be ready
./scripts/docker-validate.sh --wait --max-wait 120

# Auto-restart unhealthy services
./scripts/docker-validate.sh --auto-restart

# Start with monitoring
docker-compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f [service]
```

---

## ⚠️ IMPORTANT: Avoiding Multiple Instance Problems

### What This Means (In Plain English)

When you start services with `./scripts/dev.sh`, your applications run inside Docker containers. Think of containers like separate computers running on your machine.

**The Problem:** If you (or an AI assistant like Copilot) try to run the same service again directly on your computer (outside of Docker), you'll have **two copies of the same service trying to use the same network port**. This causes errors and neither will work properly.

**Example:** It's like trying to have two people answer the same phone line at the same time - it just doesn't work.

---

### How to Tell If Services Are Already Running

**Before running any service commands, ALWAYS check if Docker containers are already running:**

```bash
# See all running Docker containers
docker-compose ps
```

**What you'll see if services are running:**

```
NAME                                    STATUS
devsmith-modular-platform-portal-1      Up 30 minutes (healthy)
devsmith-modular-platform-review-1      Up 30 minutes (healthy)
devsmith-modular-platform-logs-1        Up 30 minutes (healthy)
devsmith-modular-platform-analytics-1   Up 30 minutes (healthy)
```

**Key indicator:** Look for `"Up"` in the STATUS column. If you see this, **DO NOT** run the service again using other commands.

---

### What NOT to Do When Docker Is Running

**❌ NEVER do these when containers are running:**

```bash
# DON'T run these if docker-compose ps shows services "Up"
go run cmd/portal/main.go       # ❌ Will conflict with Docker portal
go run cmd/review/main.go       # ❌ Will conflict with Docker review
go run cmd/logs/main.go         # ❌ Will conflict with Docker logs
go run cmd/analytics/main.go    # ❌ Will conflict with Docker analytics

# DON'T build and run manually
go build -o portal cmd/portal/main.go && ./portal  # ❌ Port conflict
```

**Why?** Each service needs a specific port:
- Portal uses port 8080
- Review uses port 8081
- Logs uses port 8082
- Analytics uses port 8083

If Docker is already using these ports, running the service again will fail with errors like:
```
Error: bind: address already in use
```

---

### What TO Do Instead

**✅ If containers are running and you need to:**

**Restart a service:**
```bash
# Restart just the portal service
docker-compose restart portal

# Restart all services
docker-compose restart
```

**See what a service is doing (view logs):**
```bash
# Watch portal logs in real-time
docker-compose logs -f portal

# See last 50 lines of portal logs
docker-compose logs portal --tail=50
```

**Stop and rebuild a service after code changes:**
```bash
# Rebuild and restart portal service
docker-compose up -d --build portal
```

**Stop all services:**
```bash
# Stop all containers (doesn't delete them)
docker-compose stop

# Stop and remove all containers
docker-compose down
```

---

### Step-by-Step: Making Changes to a Service

**When you edit code and want to see your changes:**

1. **Check what's running:**
   ```bash
   docker-compose ps
   ```

2. **Stop the service you're changing:**
   ```bash
   docker-compose stop portal
   ```

3. **Rebuild with your changes:**
   ```bash
   docker-compose up -d --build portal
   ```

4. **Verify it's working:**
   ```bash
   ./scripts/docker-validate.sh
   ```

**Important:** You don't need to stop all services, just the one you're changing!

---

### For AI Assistants (Copilot, Claude, etc.)

**⚠️ CRITICAL: Check Before Running Services**

Before executing ANY command that starts a service, you MUST run:

```bash
docker-compose ps | grep -q "Up" && echo "DOCKER IS RUNNING - Use docker-compose commands only" || echo "Safe to run directly"
```

**If Docker is running (you see "DOCKER IS RUNNING"):**
- ✅ Use: `docker-compose restart [service]`
- ✅ Use: `docker-compose logs [service]`
- ✅ Use: `docker-compose up -d --build [service]`
- ❌ NEVER use: `go run cmd/[service]/main.go`
- ❌ NEVER use: `./[service]` or any direct service execution

**Why this matters:**
1. Prevents port conflicts (multiple services on same port)
2. Prevents database connection issues (multiple connections)
3. Prevents confusion about which instance is responding
4. Prevents "works for me but not in Docker" bugs

**Check port availability:**
```bash
# Check if a port is in use (example: port 8080)
lsof -i :8080 || echo "Port 8080 is free"

# Check all DevSmith service ports
for port in 8080 8081 8082 8083; do
  lsof -i :$port >/dev/null 2>&1 && echo "Port $port: IN USE ⚠️" || echo "Port $port: free ✓"
done
```

**Safe workflow:**
1. ALWAYS check: `docker-compose ps`
2. If services are "Up", use `docker-compose` commands only
3. If services are "Down" or not listed, direct execution is safe
4. When in doubt, ask the user: "I see Docker containers running. Should I restart the service in Docker, or stop Docker first?"

---

## Files Created/Modified

### New Files
- `scripts/docker-validate.sh` - Validation script
- `docker-compose.monitoring.yml` - Optional monitoring stack
- `.docs/DOCKER-VALIDATION.md` - Complete user guide
- `.docs/DOCKER-COPILOT-GUIDE.md` - AI assistant guide
- `.docs/DOCKER-QUICKSTART.md` - This file

### Modified Files
- `docker-compose.yml` - Added health checks, improved depends_on
- `scripts/dev.sh` - Integrated validation

---

## For Developers

**Starting services:**
```bash
./scripts/dev.sh
```

That's it! The script handles everything.

**If validation fails:**
1. Check the error messages - they tell you exactly what's wrong
2. View logs: `docker-compose logs [service]`
3. See `.docs/DOCKER-VALIDATION.md` for detailed troubleshooting

---

## For AI Assistants (Copilot, etc.)

When creating or modifying Docker configurations:

1. **Always implement `/health` endpoints** in all HTTP services
2. **Always add health checks** to docker-compose.yml
3. **Always use `depends_on` with `service_healthy`**
4. **Always run validation** after changes: `./scripts/docker-validate.sh`

See `.docs/DOCKER-COPILOT-GUIDE.md` for complete patterns and examples.

---

## What This Solves

### Before
- Services started but served 404s
- Health checks said "healthy" but returned 500s
- Copilot missed port bindings and dependency ordering
- Manual debugging required

### After
- ✅ Automatic validation on startup
- ✅ Clear error messages with fix suggestions
- ✅ Proper dependency ordering
- ✅ Guaranteed service health before accepting traffic
- ✅ Optional continuous monitoring

---

## Architecture

```
docker-compose up -d
        ↓
    postgres starts
        ↓
    postgres becomes healthy (pg_isready)
        ↓
    portal, review, logs, analytics start (parallel)
        ↓
    services connect to DB and initialize
        ↓
    services become healthy (/health returns 200)
        ↓
    nginx starts
        ↓
    nginx becomes healthy
        ↓
    docker-validate.sh checks:
        ✓ All containers running
        ✓ All health checks passing
        ✓ All HTTP endpoints responding
        ✓ All ports correctly bound
        ↓
    ✅ Ready for development!
```

---

## Next Steps

1. **Read** `.docs/DOCKER-VALIDATION.md` for complete documentation
2. **Try** `./scripts/dev.sh` to see it in action
3. **Enable monitoring** with `docker-compose.monitoring.yml` (optional)
4. **Use** the validation script during development

---

## Support

- **Full guide:** `.docs/DOCKER-VALIDATION.md`
- **AI assistant guide:** `.docs/DOCKER-COPILOT-GUIDE.md`
- **Troubleshooting:** `.docs/DOCKER-VALIDATION.md#troubleshooting`
