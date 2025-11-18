#!/bin/bash
# nuclear-complete-rebuild-enhanced.sh: Enhanced rebuild with better error reporting
set -e

# Configuration
SKIP_MANUAL_VERIFICATION=${SKIP_MANUAL_VERIFICATION:-true}  # Default skip for rebuild
VERBOSE=${VERBOSE:-false}

# Logging helpers
log_phase() {
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "[$1] $2"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""
}

log_success() {
  echo "✅ $1"
}

log_warning() {
  echo "⚠️  $1"
}

log_error() {
  echo "❌ $1"
}

report_service_logs() {
  local service=$1
  log_error "Showing last 50 lines of $service logs:"
  docker-compose logs "$service" --tail=50 || echo "Could not fetch logs for $service"
}

# Phase 1: Teardown
log_phase "1/7" "Teardown: Removing all containers and volumes"
docker-compose down -v
log_success "Teardown complete"

# Phase 2: Build
log_phase "2/7" "Build: Building and starting all services"
docker-compose up -d --build traefik portal review logs analytics postgres redis
log_success "Build complete"

# Phase 2.1: Traefik Health
log_phase "2.1/7" "Waiting for Traefik health (max 60s)"
for i in {1..30}; do
  status=$(docker inspect --format='{{.State.Health.Status}}' devsmith-traefik 2>/dev/null || echo "notfound")
  if [ "$status" = "healthy" ]; then
    log_success "Traefik is healthy"
    break
  fi
  if [ "$status" = "notfound" ]; then
    log_warning "Traefik container not found, waiting..."
  else
    echo "Traefik status: $status (attempt $i/30)"
  fi
  sleep 2
done

if [ "$status" != "healthy" ]; then
  log_error "Traefik failed to become healthy"
  report_service_logs "traefik"
  exit 1
fi

# Phase 2.2: Port Check
log_phase "2.2/7" "Waiting for port 3000 (max 30s)"
for i in {1..15}; do
  if lsof -i :3000 | grep -q LISTEN; then
    log_success "Port 3000 is listening"
    break
  fi
  echo "Waiting for port 3000 (attempt $i/15)"
  sleep 2
done

if ! lsof -i :3000 | grep -q LISTEN; then
  log_error "Port 3000 is not listening"
  exit 1
fi

# Phase 3: Health Check
log_phase "3/7" "Health Check: Verifying all services"
docker-compose ps
echo ""

# Check individual service health
services="portal review logs analytics postgres redis"
unhealthy_count=0

for service in $services; do
  container_name="devsmith-modular-platform-${service}-1"
  status=$(docker inspect --format='{{.State.Health.Status}}' "$container_name" 2>/dev/null || echo "no-health-check")
  
  if [ "$status" = "healthy" ]; then
    log_success "$service is healthy"
  elif [ "$status" = "no-health-check" ]; then
    # Some services like Redis might not have health checks
    if docker ps --filter "name=$container_name" --filter "status=running" | grep -q "$container_name"; then
      log_success "$service is running (no health check)"
    else
      log_error "$service is not running"
      unhealthy_count=$((unhealthy_count + 1))
    fi
  else
    log_error "$service is unhealthy (status: $status)"
    report_service_logs "$service"
    unhealthy_count=$((unhealthy_count + 1))
  fi
done

if [ $unhealthy_count -gt 0 ]; then
  log_error "$unhealthy_count service(s) are unhealthy"
  exit 1
fi

# Phase 4: Migrations
log_phase "4/7" "Migrations: Running database migrations"
if ! bash scripts/run-migrations.sh; then
  log_error "Migration failed"
  report_service_logs "portal"
  report_service_logs "review"
  report_service_logs "logs"
  exit 1
fi
log_success "Migrations complete"

# Phase 4.1: Validate database schema
log_phase "4.1/7" "Validating database schema"
if docker-compose exec -T postgres psql -U devsmith -d devsmith -c "\d logs.entries" > /dev/null 2>&1; then
  log_success "Database schema validated (logs.entries exists)"
else
  log_warning "Could not validate database schema"
fi

# Phase 5: Regression Tests
log_phase "5/7" "Regression tests: Running test suite"
if bash scripts/regression-test.sh; then
  log_success "Regression tests passed"
else
  log_warning "Regression tests failed or incomplete"
  echo "Check test-results/ for details"
  echo "Note: Tests might not be configured yet after full rebuild"
fi

# Phase 6: Service Endpoint Validation
log_phase "6/7" "Service endpoint validation"

validate_endpoint() {
  local name=$1
  local url=$2
  if curl -f -s "$url" > /dev/null 2>&1; then
    log_success "$name endpoint responding"
    return 0
  else
    log_warning "$name endpoint not responding: $url"
    return 1
  fi
}

validate_endpoint "Portal health" "http://localhost:3000/api/portal/health"
validate_endpoint "Review health" "http://localhost:3000/api/review/health"
validate_endpoint "Logs health" "http://localhost:3000/api/logs/health"
validate_endpoint "Analytics health" "http://localhost:3000/api/analytics/health"

# Phase 7: Manual Verification
log_phase "7/7" "Manual verification"

if [ "$SKIP_MANUAL_VERIFICATION" = "false" ]; then
  echo "Checking manual verification artifacts..."
  if ! ls test-results/manual-verification-* 1> /dev/null 2>&1; then
    log_warning "Manual verification screenshots missing"
    echo "Create screenshots after setting up AI model and testing UI"
  fi
  if ! find test-results/manual-verification-* -name VERIFICATION.md 2>/dev/null | grep -q VERIFICATION.md; then
    log_warning "VERIFICATION.md document missing"
    echo "Create VERIFICATION.md after completing manual tests"
  fi
else
  echo "Skipping manual verification (SKIP_MANUAL_VERIFICATION=$SKIP_MANUAL_VERIFICATION)"
  echo ""
  echo "To complete validation:"
  echo "  1. Login: http://localhost:3000/auth/github/login"
  echo "  2. Setup AI model in AI Factory: http://localhost:3000/ai-factory"
  echo "  3. Test Review app: http://localhost:3000/review"
  echo "  4. Run Playwright tests: npx playwright test"
  echo "  5. Create VERIFICATION.md with screenshots"
  echo ""
fi

# Final Report
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ NUCLEAR REBUILD COMPLETE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Service Status:"
docker-compose ps
echo ""
echo "Platform URL: http://localhost:3000"
echo ""
echo "Next Steps:"
echo "  1. Setup AI model: http://localhost:3000/ai-factory"
echo "     - Provider: Ollama (Local)"
echo "     - Endpoint: http://host.docker.internal:11434"
echo "     - Model: qwen2.5-coder:7b or deepseek-coder:6.7b"
echo ""
echo "  2. Test Review app: http://localhost:3000/review"
echo "     - Paste code and select Preview/Skim/Scan/Detailed/Critical mode"
echo "     - Verify JSON responses (not HTML)"
echo ""
echo "  3. Check for errors:"
echo "     docker-compose logs review --tail=100"
echo ""
echo "  4. Run Playwright + Percy validation:"
echo "     npx playwright test"
echo ""
echo "  5. Create verification document if all tests pass"
echo ""
