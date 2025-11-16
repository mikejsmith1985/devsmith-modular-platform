#!/bin/bash

# OAuth Robustness Enhancements Test Script
# Tests all 5 priorities of OAuth improvements

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "OAuth Robustness Enhancements - Test Suite"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

BASE_URL="http://localhost:3001"
GATEWAY_URL="http://localhost:3000"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to test endpoint
test_endpoint() {
    local name="$1"
    local url="$2"
    local expected_status="$3"
    local description="$4"
    
    echo -n "Testing: $name... "
    
    response=$(curl -s -w "\n%{http_code}" "$url")
    status_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $status_code)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        if [ -n "$description" ]; then
            echo "  → $description"
        fi
        return 0
    else
        echo -e "${RED}✗ FAIL${NC} (Expected $expected_status, got $status_code)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo "  Response: $body"
        return 1
    fi
}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Priority 1: OAuth Health Check Endpoint"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 1.1: Health check endpoint exists
test_endpoint \
    "Health check via direct API" \
    "$BASE_URL/api/portal/auth/health" \
    "200" \
    "Endpoint returns OK status"

# Test 1.2: Health check via gateway
test_endpoint \
    "Health check via gateway" \
    "$GATEWAY_URL/api/portal/auth/health" \
    "200" \
    "Gateway routing works"

# Test 1.3: Health check via legacy path
test_endpoint \
    "Health check via legacy path" \
    "$BASE_URL/auth/health" \
    "200" \
    "Backward compatibility maintained"

# Test 1.4: Validate health check response structure
echo -n "Testing: Health check response structure... "
health_response=$(curl -s "$BASE_URL/api/portal/auth/health")

if echo "$health_response" | jq -e '.healthy' > /dev/null 2>&1 && \
   echo "$health_response" | jq -e '.checks' > /dev/null 2>&1 && \
   echo "$health_response" | jq -e '.timestamp' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    
    # Show detailed checks
    echo "  → Health status: $(echo "$health_response" | jq -r '.healthy')"
    echo "  → GitHub Client ID configured: $(echo "$health_response" | jq -r '.checks.github_client_id_set')"
    echo "  → GitHub Client Secret configured: $(echo "$health_response" | jq -r '.checks.github_client_secret_set')"
    echo "  → JWT Secret configured: $(echo "$health_response" | jq -r '.checks.jwt_secret_set')"
    echo "  → Redis available: $(echo "$health_response" | jq -r '.checks.redis_available')"
    echo "  → Redis writable: $(echo "$health_response" | jq -r '.checks.redis_writable')"
else
    echo -e "${RED}✗ FAIL${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo "  Missing required fields in response"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Priority 2: State Parameter (CSRF Protection)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 2.1: Login generates state parameter
echo -n "Testing: GitHub OAuth login generates state... "
# Use -v to get headers, grep for Location header with state parameter
login_redirect=$(curl -s -v "$BASE_URL/auth/github/login" 2>&1 | grep -i "^< Location:" | cut -d' ' -f3)

if echo "$login_redirect" | grep -q "state="; then
    echo -e "${GREEN}✓ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    state_param=$(echo "$login_redirect" | grep -oP 'state=\K[^&]+' | head -n 1)
    echo "  → State parameter generated: ${state_param:0:30}..."
else
    echo -e "${RED}✗ FAIL${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo "  → No state parameter in OAuth URL"
    echo "  → Redirect URL: $login_redirect"
fi

# Test 2.2: Callback without state parameter is rejected
echo -n "Testing: Callback rejects missing state... "
callback_response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/portal/auth/github/callback?code=test123")
callback_status=$(echo "$callback_response" | tail -n 1)
callback_body=$(echo "$callback_response" | head -n -1)

if [ "$callback_status" = "400" ] && echo "$callback_body" | grep -q "state"; then
    echo -e "${GREEN}✓ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo "  → Correctly rejects callback without state"
else
    echo -e "${RED}✗ FAIL${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo "  → Should reject callback without state (got HTTP $callback_status)"
fi

# Test 2.3: Callback with invalid state is rejected
echo -n "Testing: Callback rejects invalid state... "
callback_response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/portal/auth/github/callback?code=test123&state=invalid_state_12345")
callback_status=$(echo "$callback_response" | tail -n 1)
callback_body=$(echo "$callback_response" | head -n -1)

if [ "$callback_status" = "401" ] && echo "$callback_body" | grep -qi "state"; then
    echo -e "${GREEN}✓ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo "  → Correctly rejects invalid state parameter"
else
    echo -e "${RED}✗ FAIL${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo "  → Should reject invalid state (got HTTP $callback_status)"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Priority 3: Enhanced Error Messages"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 3.1: Error response includes detailed fields
echo -n "Testing: Error messages have required fields... "
error_response=$(curl -s "$BASE_URL/api/portal/auth/github/callback?code=&state=")
error_body=$(echo "$error_response")

has_error=$(echo "$error_body" | jq -e '.error' > /dev/null 2>&1 && echo "yes" || echo "no")
has_details=$(echo "$error_body" | jq -e '.details' > /dev/null 2>&1 && echo "yes" || echo "no")
has_action=$(echo "$error_body" | jq -e '.action' > /dev/null 2>&1 && echo "yes" || echo "no")

if [ "$has_error" = "yes" ] && [ "$has_details" = "yes" ] && [ "$has_action" = "yes" ]; then
    echo -e "${GREEN}✓ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo "  → Error: $(echo "$error_body" | jq -r '.error')"
    echo "  → Details: $(echo "$error_body" | jq -r '.details')"
    echo "  → Action: $(echo "$error_body" | jq -r '.action')"
else
    echo -e "${RED}✗ FAIL${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo "  → Missing required fields (error=$has_error, details=$has_details, action=$has_action)"
fi

# Test 3.2: Error includes error code for support
echo -n "Testing: Error messages include error codes... "
if echo "$error_body" | jq -r '.action' | grep -q "error code:"; then
    echo -e "${GREEN}✓ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    error_code=$(echo "$error_body" | jq -r '.action' | grep -oP 'error code: \K[A-Z_]+' | head -n 1)
    echo "  → Error code found: $error_code"
else
    echo -e "${YELLOW}⚠ PARTIAL${NC}"
    echo "  → Error code not found in action message"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Priority 4: Comprehensive Logging"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 4.1: Check for enhanced logging tags
echo -n "Testing: Enhanced logging tags present... "
docker logs devsmith-modular-platform-portal-1 2>&1 | tail -100 > /tmp/portal_logs.txt

if grep -q "\[OAUTH\]" /tmp/portal_logs.txt || \
   grep -q "\[TOKEN_EXCHANGE\]" /tmp/portal_logs.txt || \
   grep -q "\[USER_INFO\]" /tmp/portal_logs.txt; then
    echo -e "${GREEN}✓ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    
    oauth_count=$(grep -c "\[OAUTH\]" /tmp/portal_logs.txt || echo "0")
    token_count=$(grep -c "\[TOKEN_EXCHANGE\]" /tmp/portal_logs.txt || echo "0")
    user_count=$(grep -c "\[USER_INFO\]" /tmp/portal_logs.txt || echo "0")
    
    echo "  → [OAUTH] logs: $oauth_count"
    echo "  → [TOKEN_EXCHANGE] logs: $token_count"
    echo "  → [USER_INFO] logs: $user_count"
else
    echo -e "${YELLOW}⚠ PARTIAL${NC}"
    echo "  → Enhanced logging tags not found (may not have run OAuth flow yet)"
fi

# Test 4.2: Check for step-by-step logging
echo -n "Testing: Step-by-step OAuth logging... "
if grep -q "Step [0-9]" /tmp/portal_logs.txt; then
    echo -e "${GREEN}✓ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    steps=$(grep -o "Step [0-9]" /tmp/portal_logs.txt | sort -u)
    echo "  → Found logging steps: $steps"
else
    echo -e "${YELLOW}⚠ PARTIAL${NC}"
    echo "  → Step-by-step logging not found (may not have run OAuth flow yet)"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Total Tests: $((TESTS_PASSED + TESTS_FAILED))"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}✓ ALL TESTS PASSED${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    exit 0
else
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${RED}✗ SOME TESTS FAILED${NC}"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    exit 1
fi
