#!/bin/bash
# Test script for GitHub Integration Phase 1 endpoints
# Usage: ./test-github-integration.sh [github_token]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:3000"
TEST_REPO="github.com/mikejsmith1985/devsmith-modular-platform"
TEST_BRANCH="main"
TEST_FILE="README.md"

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}GitHub Integration Phase 1 Test Suite${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Check if review service is running
echo -e "${YELLOW}→ Checking if Review service is running...${NC}"
if ! curl -sf "${BASE_URL}/health" > /dev/null 2>&1; then
    echo -e "${RED}✗ Review service is not running at ${BASE_URL}${NC}"
    echo -e "${YELLOW}  Start it with: docker-compose up -d review${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Review service is running${NC}"
echo ""

# Note about authentication
echo -e "${YELLOW}→ Note: These endpoints require authentication${NC}"
echo -e "${YELLOW}  You need to be logged in via Portal's GitHub OAuth${NC}"
echo -e "${YELLOW}  1. Visit: http://localhost:3000/auth/github/login${NC}"
echo -e "${YELLOW}  2. Authorize with GitHub${NC}"
echo -e "${YELLOW}  3. Copy your session cookie and pass it to these tests${NC}"
echo ""

# If token provided as argument, use it
if [ -n "$1" ]; then
    TOKEN="$1"
    echo -e "${GREEN}✓ Using provided session token${NC}"
else
    echo -e "${YELLOW}  No token provided, tests will show auth required errors${NC}"
    TOKEN=""
fi
echo ""

# Test 1: Repository Tree Endpoint
echo -e "${YELLOW}→ Test 1: GET /api/review/github/tree${NC}"
TREE_RESPONSE=$(curl -s -w "\n%{http_code}" \
    -H "Cookie: devsmith_token=${TOKEN}" \
    "${BASE_URL}/api/review/github/tree?url=${TEST_REPO}&branch=${TEST_BRANCH}")

TREE_STATUS=$(echo "$TREE_RESPONSE" | tail -n 1)
TREE_BODY=$(echo "$TREE_RESPONSE" | head -n -1)

if [ "$TREE_STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Status: 200 OK${NC}"
    
    # Parse JSON response
    OWNER=$(echo "$TREE_BODY" | jq -r '.owner // "N/A"')
    REPO=$(echo "$TREE_BODY" | jq -r '.repo // "N/A"')
    FILE_COUNT=$(echo "$TREE_BODY" | jq -r '.file_count // 0')
    ENTRY_POINTS=$(echo "$TREE_BODY" | jq -r '.entry_points | length // 0')
    
    echo -e "${GREEN}  Owner: ${OWNER}${NC}"
    echo -e "${GREEN}  Repo: ${REPO}${NC}"
    echo -e "${GREEN}  Files: ${FILE_COUNT}${NC}"
    echo -e "${GREEN}  Entry Points: ${ENTRY_POINTS}${NC}"
elif [ "$TREE_STATUS" = "401" ]; then
    echo -e "${YELLOW}⚠ Status: 401 Unauthorized${NC}"
    echo -e "${YELLOW}  Please provide authentication token${NC}"
else
    echo -e "${RED}✗ Status: ${TREE_STATUS}${NC}"
    echo -e "${RED}  Response: ${TREE_BODY}${NC}"
fi
echo ""

# Test 2: File Content Endpoint
echo -e "${YELLOW}→ Test 2: GET /api/review/github/file${NC}"
FILE_RESPONSE=$(curl -s -w "\n%{http_code}" \
    -H "Cookie: devsmith_token=${TOKEN}" \
    "${BASE_URL}/api/review/github/file?url=${TEST_REPO}&path=${TEST_FILE}&branch=${TEST_BRANCH}")

FILE_STATUS=$(echo "$FILE_RESPONSE" | tail -n 1)
FILE_BODY=$(echo "$FILE_RESPONSE" | head -n -1)

if [ "$FILE_STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Status: 200 OK${NC}"
    
    PATH=$(echo "$FILE_BODY" | jq -r '.path // "N/A"')
    LANGUAGE=$(echo "$FILE_BODY" | jq -r '.language // "N/A"')
    SIZE=$(echo "$FILE_BODY" | jq -r '.size // 0')
    CONTENT_LENGTH=$(echo "$FILE_BODY" | jq -r '.content | length // 0')
    
    echo -e "${GREEN}  Path: ${PATH}${NC}"
    echo -e "${GREEN}  Language: ${LANGUAGE}${NC}"
    echo -e "${GREEN}  Size: ${SIZE} bytes${NC}"
    echo -e "${GREEN}  Content Length: ${CONTENT_LENGTH} chars${NC}"
elif [ "$FILE_STATUS" = "401" ]; then
    echo -e "${YELLOW}⚠ Status: 401 Unauthorized${NC}"
    echo -e "${YELLOW}  Please provide authentication token${NC}"
else
    echo -e "${RED}✗ Status: ${FILE_STATUS}${NC}"
    echo -e "${RED}  Response: ${FILE_BODY}${NC}"
fi
echo ""

# Test 3: Quick Scan Endpoint
echo -e "${YELLOW}→ Test 3: GET /api/review/github/quick-scan${NC}"
SCAN_RESPONSE=$(curl -s -w "\n%{http_code}" \
    -H "Cookie: devsmith_token=${TOKEN}" \
    "${BASE_URL}/api/review/github/quick-scan?url=${TEST_REPO}&branch=${TEST_BRANCH}")

SCAN_STATUS=$(echo "$SCAN_RESPONSE" | tail -n 1)
SCAN_BODY=$(echo "$SCAN_RESPONSE" | head -n -1)

if [ "$SCAN_STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Status: 200 OK${NC}"
    
    OWNER=$(echo "$SCAN_BODY" | jq -r '.owner // "N/A"')
    REPO=$(echo "$SCAN_BODY" | jq -r '.repo // "N/A"')
    FILES_FETCHED=$(echo "$SCAN_BODY" | jq -r '.files_fetched // 0')
    
    echo -e "${GREEN}  Owner: ${OWNER}${NC}"
    echo -e "${GREEN}  Repo: ${REPO}${NC}"
    echo -e "${GREEN}  Files Fetched: ${FILES_FETCHED}${NC}"
    
    # Show which files were fetched
    echo -e "${GREEN}  Core Files:${NC}"
    echo "$SCAN_BODY" | jq -r '.files[].path' | while read -r file; do
        echo -e "${GREEN}    - ${file}${NC}"
    done
elif [ "$SCAN_STATUS" = "401" ]; then
    echo -e "${YELLOW}⚠ Status: 401 Unauthorized${NC}"
    echo -e "${YELLOW}  Please provide authentication token${NC}"
else
    echo -e "${RED}✗ Status: ${SCAN_STATUS}${NC}"
    echo -e "${RED}  Response: ${SCAN_BODY}${NC}"
fi
echo ""

# Summary
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Test Summary${NC}"
echo -e "${YELLOW}========================================${NC}"

if [ "$TREE_STATUS" = "200" ] && [ "$FILE_STATUS" = "200" ] && [ "$SCAN_STATUS" = "200" ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    echo -e "${GREEN}  Phase 1 GitHub Integration is working correctly${NC}"
    exit 0
elif [ "$TREE_STATUS" = "401" ] || [ "$FILE_STATUS" = "401" ] || [ "$SCAN_STATUS" = "401" ]; then
    echo -e "${YELLOW}⚠ Authentication required${NC}"
    echo -e "${YELLOW}  Please log in via Portal and provide session token${NC}"
    echo -e "${YELLOW}  Usage: ./test-github-integration.sh YOUR_SESSION_TOKEN${NC}"
    exit 1
else
    echo -e "${RED}✗ Some tests failed${NC}"
    echo -e "${RED}  Check the output above for details${NC}"
    exit 1
fi
