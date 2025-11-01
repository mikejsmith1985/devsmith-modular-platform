#!/bin/bash
# Run Playwright smoke tests inside Docker with access to localhost:3000
# On Linux: uses --network=host
# On macOS/Windows: uses host.docker.internal

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}🧪 Running Playwright Smoke Tests in Docker${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo ""

# Check if services are running
echo "Checking if services are running at http://localhost:3000..."
if ! curl -s http://localhost:3000/health > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  WARNING: Services may not be running at http://localhost:3000${NC}"
    echo "Start them with: docker-compose up -d"
    echo ""
fi

# Detect OS
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo -e "${GREEN}✓ Linux detected - using --network=host${NC}"
    PLAYWRIGHT_NETWORK=host docker-compose -f docker-compose.playwright.yml up --exit-code-from playwright
elif [[ "$OSTYPE" == "darwin"* ]]; then
    echo -e "${GREEN}✓ macOS detected - using host.docker.internal${NC}"
    docker-compose -f docker-compose.playwright.yml up --exit-code-from playwright
else
    echo -e "${YELLOW}⚠️  Unknown OS: $OSTYPE${NC}"
    echo "Trying with --network=host..."
    PLAYWRIGHT_NETWORK=host docker-compose -f docker-compose.playwright.yml up --exit-code-from playwright
fi

EXIT_CODE=$?

echo ""
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
else
    echo -e "${YELLOW}❌ Some tests failed${NC}"
fi
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"

exit $EXIT_CODE
