#!/bin/bash
# Run Playwright smoke tests locally
# Tests execute in < 60 seconds locally, avoiding Docker hanging issues

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}🧪 Running Playwright Smoke Tests (Local)${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo ""

# Default to all if no argument provided
FEATURE="${1:-all}"

# Map feature names to test paths
declare -A FEATURE_MAP=(
	[ollama]="tests/e2e/smoke/ollama-integration/"
	[ui]="tests/e2e/smoke/ui-rendering/"
	[all]="tests/e2e/smoke/"
	[review]="tests/e2e/smoke/ollama-integration/"
	[logs]="tests/e2e/smoke/full-suite/"
	[analytics]="tests/e2e/smoke/full-suite/"
	[portal]="tests/e2e/smoke/ui-rendering/"
)

# Validate feature argument
if [[ ! ${FEATURE_MAP[$FEATURE]+_} ]]; then
	echo -e "${YELLOW}❌ Unknown feature: $FEATURE${NC}"
	echo ""
	echo "Available options:"
	echo "  ollama     - Ollama integration tests"
	echo "  ui         - UI rendering tests"
	echo "  all        - All smoke tests (default)"
	echo "  review     - Alias for 'ollama'"
	echo "  logs       - Logs tests"
	echo "  analytics  - Analytics tests"
	echo "  portal     - Portal tests"
	exit 1
fi

TEST_PATH="${FEATURE_MAP[$FEATURE]}"

echo -e "${BLUE}Test Suite:${NC} $TEST_PATH"
echo ""

# Check if services are running
echo "Checking if services are running at http://localhost:3000..."
if ! curl -s http://localhost:3000/health > /dev/null 2>&1; then
	echo -e "${YELLOW}⚠️  WARNING: Services may not be running at http://localhost:3000${NC}"
	echo "Start them with: docker-compose up -d"
	echo ""
fi

# Run tests locally (< 60 seconds)
echo -e "${YELLOW}Running tests locally...${NC}"
echo ""

START_TIME=$(date +%s)

if npx playwright test "$TEST_PATH" --project=smoke --workers=6; then
	END_TIME=$(date +%s)
	DURATION=$((END_TIME - START_TIME))
	
	echo ""
	echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
	echo -e "${GREEN}✅ All tests passed!${NC}"
	echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
	echo ""
	echo "Duration: ${DURATION}s"
	echo ""
	exit 0
else
	END_TIME=$(date +%s)
	DURATION=$((END_TIME - START_TIME))
	
	echo ""
	echo -e "${YELLOW}════════════════════════════════════════════════════════════${NC}"
	echo -e "${YELLOW}❌ Tests failed${NC}"
	echo -e "${YELLOW}════════════════════════════════════════════════════════════${NC}"
	echo ""
	echo "Duration: ${DURATION}s"
	echo ""
	exit 1
fi
