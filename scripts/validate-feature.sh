#!/bin/bash
# Feature Validation Script - Comprehensive E2E Testing
# Usage: ./scripts/validate-feature.sh [review|logs|analytics|portal|all]
# 
# This script runs feature-specific E2E tests to validate implementations
# before creating pull requests. Tests run in parallel (4-6 workers).
#
# Examples:
#   ./scripts/validate-feature.sh review          # Test review features
#   ./scripts/validate-feature.sh all              # Full feature validation

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Default to all if no argument provided
FEATURE="${1:-all}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}🧪 Feature Validation Suite${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo ""
echo "Feature: $FEATURE"
echo ""

# Check if docker-compose services are running
echo "Checking Docker services..."
if ! curl -s http://localhost:3000/health > /dev/null 2>&1; then
    echo "❌ ERROR: Services not running at http://localhost:3000"
    echo ""
    echo "Start services with:"
    echo "  docker-compose up -d"
    echo ""
    exit 1
fi
echo "✅ Services are running"
echo ""

# Check if Playwright is installed
if ! command -v npx &> /dev/null; then
    echo "❌ ERROR: npx not found. Install Node.js and npm."
    exit 1
fi

# Run tests based on feature
case "$FEATURE" in
  smoke)
    echo "Running smoke tests (quick validation)..."
    npx playwright test --project=smoke --workers=4
    ;;
  
  review)
    echo "Running review feature tests..."
    npx playwright test tests/e2e/smoke/review-*.spec.ts tests/e2e/features/review-*.spec.ts --workers=4
    ;;
  
  logs)
    echo "Running logs feature tests..."
    npx playwright test tests/e2e/smoke/logs-*.spec.ts tests/e2e/features/logs-*.spec.ts --workers=4
    ;;
  
  analytics)
    echo "Running analytics feature tests..."
    npx playwright test tests/e2e/smoke/analytics-*.spec.ts tests/e2e/features/analytics-*.spec.ts --workers=4
    ;;
  
  portal)
    echo "Running portal feature tests..."
    npx playwright test tests/e2e/smoke/portal-*.spec.ts tests/e2e/features/portal-*.spec.ts --workers=4
    ;;
  
  all)
    echo "Running all smoke tests + feature tests..."
    echo "This validates: Portal, Review, Logs, Analytics"
    echo ""
    npx playwright test --project=full --workers=6
    ;;
  
  *)
    echo "❌ Unknown feature: $FEATURE"
    echo ""
    echo "Valid options:"
    echo "  smoke      - Quick smoke tests (< 30s)"
    echo "  review     - Review feature tests"
    echo "  logs       - Logs feature tests"
    echo "  analytics  - Analytics feature tests"
    echo "  portal     - Portal feature tests"
    echo "  all        - All tests (full suite)"
    echo ""
    exit 1
    ;;
esac

TEST_EXIT_CODE=$?

echo ""
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}✅ FEATURE VALIDATION PASSED${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo "Ready to create PR!"
else
    echo -e "${YELLOW}════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}❌ FEATURE VALIDATION FAILED${NC}"
    echo -e "${YELLOW}════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo "Review test failures above and fix:"
    echo "  1. Check specific test failure messages"
    echo "  2. Review implementation code"
    echo "  3. Fix the issue"
    echo "  4. Re-run: ./scripts/validate-feature.sh $FEATURE"
    echo ""
fi

exit $TEST_EXIT_CODE
