#!/bin/bash

# Test script for Claude API Integration (Phase 5)
# Usage: ./test-claude-api-integration.sh [session_token]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="${BASE_URL:-http://localhost:3000}"
SESSION_TOKEN="${1:-}"

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Claude API Integration Test Suite (Phase 5)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Check if session token provided
if [ -z "$SESSION_TOKEN" ]; then
    echo -e "${YELLOW}⚠ No session token provided${NC}"
    echo -e "${YELLOW}  To test authenticated endpoints, run:${NC}"
    echo -e "${YELLOW}  ./test-claude-api-integration.sh YOUR_SESSION_TOKEN${NC}"
    echo ""
    echo -e "${BLUE}ℹ Running unauthenticated tests only...${NC}"
    echo ""
fi

# Test counter
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run test
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_pattern="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "${BLUE}TEST $TOTAL_TESTS: $test_name${NC}"
    
    # Run command and capture output
    output=$(eval "$command" 2>&1)
    exit_code=$?
    
    # Check result
    if echo "$output" | grep -q "$expected_pattern"; then
        echo -e "${GREEN}✓ PASS${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}"
        echo -e "${RED}Expected pattern: $expected_pattern${NC}"
        echo -e "${RED}Got: ${output:0:200}${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

echo -e "${BLUE}━━━ Phase 5.1: LLM Configuration Management ━━━${NC}"
echo ""

# Test 1: GET /api/portal/llm-configs (unauthenticated)
run_test "GET /api/portal/llm-configs (no auth)" \
    "curl -s $BASE_URL/api/portal/llm-configs" \
    "Authentication required"

if [ -n "$SESSION_TOKEN" ]; then
    echo ""
    
    # Test 2: GET /api/portal/llm-configs (authenticated)
    run_test "GET /api/portal/llm-configs (authenticated)" \
        "curl -s -H 'Cookie: devsmith_token=$SESSION_TOKEN' $BASE_URL/api/portal/llm-configs" \
        "configs"
    
    # Test 3: POST /api/portal/llm-configs (create Claude config)
    echo ""
    echo -e "${YELLOW}Creating Claude API config (requires valid API key)...${NC}"
    
    # Check if ANTHROPIC_API_KEY env var exists
    if [ -n "$ANTHROPIC_API_KEY" ]; then
        API_KEY="$ANTHROPIC_API_KEY"
        echo -e "${GREEN}Using ANTHROPIC_API_KEY from environment${NC}"
    else
        echo -e "${YELLOW}No ANTHROPIC_API_KEY found in environment${NC}"
        echo -e "${YELLOW}Enter Claude API key (or press Enter to skip):${NC}"
        read -r API_KEY
    fi
    
    if [ -n "$API_KEY" ]; then
        run_test "POST /api/portal/llm-configs (create Claude config)" \
            "curl -s -X POST -H 'Cookie: devsmith_token=$SESSION_TOKEN' -H 'Content-Type: application/json' -d '{\"provider\":\"claude\",\"model\":\"claude-3-5-sonnet-20241022\",\"api_key\":\"$API_KEY\",\"enabled\":true}' $BASE_URL/api/portal/llm-configs" \
            "id"
        
        # Store config ID for later tests
        CONFIG_RESPONSE=$(curl -s -H "Cookie: devsmith_token=$SESSION_TOKEN" $BASE_URL/api/portal/llm-configs)
        CONFIG_ID=$(echo "$CONFIG_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
        
        if [ -n "$CONFIG_ID" ]; then
            echo ""
            echo -e "${GREEN}Created config with ID: $CONFIG_ID${NC}"
            
            # Test 4: GET /api/portal/llm-configs/:id
            run_test "GET /api/portal/llm-configs/$CONFIG_ID" \
                "curl -s -H 'Cookie: devsmith_token=$SESSION_TOKEN' $BASE_URL/api/portal/llm-configs/$CONFIG_ID" \
                "claude"
            
            # Test 5: PUT /api/portal/llm-configs/:id
            echo ""
            run_test "PUT /api/portal/llm-configs/$CONFIG_ID (update)" \
                "curl -s -X PUT -H 'Cookie: devsmith_token=$SESSION_TOKEN' -H 'Content-Type: application/json' -d '{\"enabled\":false}' $BASE_URL/api/portal/llm-configs/$CONFIG_ID" \
                "success"
            
            # Test 6: POST /api/portal/llm-configs/:id/test
            echo ""
            echo -e "${YELLOW}Testing Claude API connection...${NC}"
            run_test "POST /api/portal/llm-configs/$CONFIG_ID/test" \
                "curl -s -X POST -H 'Cookie: devsmith_token=$SESSION_TOKEN' $BASE_URL/api/portal/llm-configs/$CONFIG_ID/test" \
                "success\|working\|valid"
            
            # Test 7: DELETE /api/portal/llm-configs/:id
            echo ""
            echo -e "${YELLOW}Cleaning up test config...${NC}"
            run_test "DELETE /api/portal/llm-configs/$CONFIG_ID" \
                "curl -s -X DELETE -H 'Cookie: devsmith_token=$SESSION_TOKEN' $BASE_URL/api/portal/llm-configs/$CONFIG_ID" \
                "success"
        fi
    else
        echo -e "${YELLOW}⚠ Skipping Claude API config creation tests (no API key provided)${NC}"
    fi
fi

echo ""
echo -e "${BLUE}━━━ Phase 5.2: Review Service Integration ━━━${NC}"
echo ""

# Test 8: Check if Review service has models endpoint
run_test "Review service models endpoint availability" \
    "curl -s $BASE_URL/api/review/models" \
    "models\|deepseek\|claude\|ollama"

echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "Total Tests:  $TOTAL_TESTS"
echo -e "Passed:       ${GREEN}$PASSED_TESTS ✓${NC}"
echo -e "Failed:       ${RED}$FAILED_TESTS ✗${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ ALL TESTS PASSED${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}❌ SOME TESTS FAILED${NC}"
    exit 1
fi
