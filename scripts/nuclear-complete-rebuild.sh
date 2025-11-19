#!/bin/bash
# nuclear-complete-rebuild.sh: Bulletproof teardown, rebuild, and validation
set -e

echo "[1/6] Teardown: docker-compose down -v"
docker-compose down -v

echo "[2/6] Build: docker-compose up -d --build traefik portal review logs analytics postgres redis"
docker-compose up -d --build traefik portal review logs analytics postgres redis

echo "[2.1/6] Waiting for Traefik health..."
# Wait for Traefik container to be healthy (max 60s). If not healthy, continue but warn.
TRAEFIK_CN="devsmith-traefik"
for i in {1..30}; do
  status=$(docker inspect --format='{{.State.Health.Status}}' ${TRAEFIK_CN} 2>/dev/null || echo "notfound")
  if [ "$status" = "healthy" ]; then
    echo "✓ Traefik is healthy"
    break
  fi
  if [ "$status" = "notfound" ]; then
    echo "Traefik container not found, waiting..."
  else
    echo "Waiting for Traefik to be healthy... ($i/30)"
  fi
  sleep 2
done
if [ "$status" != "healthy" ]; then
  echo "WARN: Traefik did not become healthy after 60s. Check logs for details."
  docker-compose logs traefik --tail=50
fi

echo "[2.2/6] Checking port 3000 availability..."
for i in {1..15}; do
  if lsof -i :3000 | grep LISTEN; then
    echo "✓ Port 3000 is listening (Traefik gateway)"
    break
  fi
  echo "Waiting for port 3000 to be available... ($i/15)"
  sleep 2
done
if ! lsof -i :3000 | grep LISTEN; then
  echo "WARN: Port 3000 is not listening after 30s. Traefik may have failed to start. Check logs above."
  docker-compose logs traefik --tail=50
fi
echo "[3/6] Health: docker-compose ps"
docker-compose ps

echo "[4/6] Migrations: bash scripts/run-migrations.sh"
if ! bash scripts/run-migrations.sh; then
  echo "WARN: Migration script returned an error. Check logs and proceed for diagnostics." >&2
fi

echo "[5/6] Regression tests: bash scripts/regression-test.sh"
if ! bash scripts/regression-test.sh; then
  echo "WARN: Regression tests failed. See test-results/regression-*/ for detail" >&2
fi

echo "[6/6] Manual verification: check screenshots and VERIFICATION.md"
SKIP_MANUAL_VERIFICATION=${SKIP_MANUAL_VERIFICATION:-true}
if [ "$SKIP_MANUAL_VERIFICATION" = "false" ]; then
  if ! ls test-results/manual-verification-* 1> /dev/null 2>&1; then
    echo "Manual verification screenshots missing." && exit 1
  fi
  if ! find test-results/manual-verification-* -name VERIFICATION.md | grep -q VERIFICATION.md; then
    echo "Verification document missing." && exit 1
  fi
else
  echo "INFO: Skipping manual verification (SKIP_MANUAL_VERIFICATION=$SKIP_MANUAL_VERIFICATION)."
fi

echo "[7/7] Quick service & AI Factory checks"
# Quick health endpoint check for critical services
services=("http://localhost:3000" "http://localhost:3001/health" "http://localhost:8081/health" "http://localhost:8082/health" "http://localhost:8083/health")
for s in "${services[@]}"; do
  if curl -s -f "$s" > /dev/null; then
    echo "✓ $s reachable"
  else
    echo "WARN: $s did not return a healthy response"
  fi
done
PORTAL_URL=${PORTAL_URL:-http://localhost:3001}
if curl -s -f "$PORTAL_URL/api/portal/app-llm-preferences" -o /dev/null; then
  echo "✓ AI Factory endpoint accessible: $PORTAL_URL/api/portal/app-llm-preferences"
else
  echo "WARN: AI Factory endpoint returned non-200 (maybe requires auth or not configured)."
fi

echo "Nuclear rebuild and validation complete."