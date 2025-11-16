#!/bin/bash
# nuclear-complete-rebuild.sh: Bulletproof teardown, rebuild, and validation
set -e

echo "[1/6] Teardown: docker-compose down -v"
docker-compose down -v

echo "[2/6] Build: docker-compose up -d --build traefik portal review logs analytics postgres redis"
docker-compose up -d --build traefik portal review logs analytics postgres redis

echo "[2.1/6] Waiting for Traefik health..."
# Wait for Traefik container to be healthy (max 60s)
for i in {1..30}; do
  status=$(docker inspect --format='{{.State.Health.Status}}' devsmith-traefik 2>/dev/null || echo "notfound")
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
  echo "ERROR: Traefik did not become healthy after 60s. Check logs."
  docker-compose logs traefik --tail=50
  exit 1
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
  echo "ERROR: Port 3000 is not listening after 30s. Traefik may have failed to start."
  docker-compose logs traefik --tail=50
  exit 1
fi
echo "[3/6] Health: docker-compose ps"
docker-compose ps

echo "[4/6] Migrations: bash scripts/run-migrations.sh"
bash scripts/run-migrations.sh

echo "[5/6] Regression tests: bash scripts/regression-test.sh"
bash scripts/regression-test.sh

echo "[6/6] Manual verification: check screenshots and VERIFICATION.md"
if ! ls test-results/manual-verification-* 1> /dev/null 2>&1; then
  echo "Manual verification screenshots missing." && exit 1
fi
if ! find test-results/manual-verification-* -name VERIFICATION.md | grep -q VERIFICATION.md; then
  echo "Verification document missing." && exit 1
fi

echo "Nuclear rebuild and validation complete."