#!/bin/bash
# nuclear-complete-rebuild.sh: Bulletproof teardown, rebuild, and validation
set -e

echo "[1/6] Teardown: docker-compose down -v"
docker-compose down -v

echo "[2/6] Build: docker-compose up -d --build"
docker-compose up -d --build

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