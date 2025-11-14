#!/bin/bash
# Feature Validation Script - Modular Smoke Test Execution
# Usage: ./scripts/validate-feature.sh [ollama|ui|all|review|logs|analytics|portal]
# 
# Focused smoke test suites for rapid validation during development.
# Run only the tests you need based on what you're working on.
#
# Examples:
#   ./scripts/validate-feature.sh ollama          # Test Ollama integration (preview, skim, scan, detailed, critical modes)
#   ./scripts/validate-feature.sh ui              # Test UI rendering (dark mode, navigation, Alpine.js)
#   ./scripts/validate-feature.sh all             # Run all smoke tests
#   ./scripts/validate-feature.sh review          # Shorthand for ollama
#   ./scripts/validate-feature.sh logs            # Shorthand for full-suite

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Default to all if no argument provided
FEATURE="${1:-all}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Map feature names to test paths and worker counts
declare -A FEATURE_MAP=(
	[ollama]="tests/e2e/smoke/ollama-integration 4"
	[ui]="tests/e2e/smoke/ui-rendering 2"
	[all]="tests/e2e/smoke 6"
	[review]="tests/e2e/smoke/ollama-integration 4"  # review = ollama tests
	[logs]="tests/e2e/smoke/full-suite 6"             # logs = full-suite tests
	[analytics]="tests/e2e/smoke/full-suite 6"        # analytics = full-suite tests
	[portal]="tests/e2e/smoke/ui-rendering 2"         # portal = ui rendering tests
)

# Validate feature argument
if [[ ! ${FEATURE_MAP[$FEATURE]+_} ]]; then
	echo -e "${RED}❌ Unknown feature: $FEATURE${NC}"
	echo ""
	echo "Available options:"
	echo "  ollama     - Test Ollama integration (Preview, Skim, Scan, Detailed, Critical modes)"
	echo "  ui         - Test UI rendering (Dark mode, Navigation, Alpine.js)"
	echo "  all        - Run all smoke tests (default)"
	echo "  review     - Alias for 'ollama'"
	echo "  logs       - Alias for 'full-suite'"
	echo "  analytics  - Alias for 'full-suite'"
	echo "  portal     - Alias for 'ui'"
	exit 1
fi

# Parse feature map
IFS=' ' read -r TEST_PATH WORKERS <<< "${FEATURE_MAP[$FEATURE]}"

echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}🧪 Feature Validation: $FEATURE${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${CYAN}Test Suite:${NC} $TEST_PATH"
echo -e "${CYAN}Workers:${NC} $WORKERS"
echo -e "${CYAN}Duration:${NC} <15 seconds${NC}"
echo ""

# Check if docker-compose services are running
echo "Checking Docker services..."
if ! curl -s http://localhost:3000/api/portal/health > /dev/null 2>&1; then
	echo "❌ ERROR: Services not running at http://localhost:3000"
	echo ""
	echo "Start services with:"
	echo "  docker-compose up -d"
	echo ""
	exit 1
fi
echo -e "${GREEN}✅ Services are running${NC}"
echo ""

# Check if Playwright is installed
if ! command -v npx &> /dev/null; then
	echo "❌ ERROR: npx not found. Install Node.js and npm."
	exit 1
fi

# Run tests
echo -e "${YELLOW}Running tests...${NC}"
echo ""

START_TIME=$(date +%s)

if npx playwright test "$TEST_PATH" --project=smoke --workers="$WORKERS" --timeout=15000; then
	END_TIME=$(date +%s)
	DURATION=$((END_TIME - START_TIME))
	
	echo ""
	echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
	echo -e "${GREEN}✅ All tests passed!${NC}"
	echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
	echo ""
	echo "Duration: ${DURATION}s"
	echo ""
	echo "Next steps:"
	case "$FEATURE" in
		ollama|review)
			echo "  • Validate UI still works: ./scripts/validate-feature.sh ui"
			echo "  • Run full validation: ./scripts/validate-feature.sh all"
			echo "  • Push to feature branch when ready"
			;;
		ui|portal)
			echo "  • Test Ollama integration: ./scripts/validate-feature.sh ollama"
			echo "  • Run full validation: ./scripts/validate-feature.sh all"
			echo "  • Push to feature branch when ready"
			;;
		all)
			echo "  • All tests passed!"
			echo "  • Ready to push: git push origin feature/xxx"
			;;
	esac
	echo ""
	exit 0
else
	END_TIME=$(date +%s)
	DURATION=$((END_TIME - START_TIME))
	
	echo ""
	echo -e "${RED}════════════════════════════════════════════════════════════${NC}"
	echo -e "${RED}❌ Tests failed${NC}"
	echo -e "${RED}════════════════════════════════════════════════════════════${NC}"
	echo ""
	echo "Duration: ${DURATION}s"
	echo ""
	echo "Troubleshooting:"
	echo "  • Check services: docker-compose ps"
	echo "  • View logs: docker-compose logs <service>"
	echo "  • Verify Ollama running (for ollama tests): ollama serve"
	echo ""
	exit 1
fi
