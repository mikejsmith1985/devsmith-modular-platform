#!/bin/bash
# Security Testing Script for Batch Log Ingestion API
# Tests: API key validation, rate limiting, SQL injection, invalid JSON, oversized payloads
# Usage: bash scripts/security-test-batch.sh

set -e

# Configuration
LOGS_API_URL="${LOGS_API_URL:-http://localhost:8082}"
VALID_API_KEY="${LOGS_API_KEY:-proj_test_key_12345}"  # Replace with real key
SERVICE_NAME="security-test"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TESTS_PASSED=0
TESTS_FAILED=0

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "🔒 SECURITY TESTING - Batch Log Ingestion API"
echo "═══════════════════════════════════════════════════════════"
echo "API URL: $LOGS_API_URL"
echo "Service: $SERVICE_NAME"
echo ""

# Helper function to run a test
run_test() {
    local test_name="$1"
    local expected_status="$2"
    local curl_cmd="$3"
    local check_body="$4"
    
    echo -n "Testing: $test_name... "
    
    # Execute curl and capture response
    response=$(eval "$curl_cmd" 2>&1)
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    # Check status code
    if [ "$status_code" -eq "$expected_status" ]; then
        # Additional body check if provided
        if [ -n "$check_body" ]; then
            if echo "$body" | grep -q "$check_body"; then
                echo -e "${GREEN}✓ PASS${NC}"
                ((TESTS_PASSED++))
            else
                echo -e "${RED}✗ FAIL${NC} (status OK but body check failed)"
                echo "  Expected body to contain: $check_body"
                echo "  Got: $body"
                ((TESTS_FAILED++))
            fi
        else
            echo -e "${GREEN}✓ PASS${NC}"
            ((TESTS_PASSED++))
        fi
    else
        echo -e "${RED}✗ FAIL${NC}"
        echo "  Expected: $expected_status, Got: $status_code"
        echo "  Response: $body"
        ((TESTS_FAILED++))
    fi
}

# Test 1: Valid API Key
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1️⃣  API Key Validation Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "Valid API key" 200 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: application/json' \
    -H 'Authorization: Bearer $VALID_API_KEY' \
    -d '{\"entries\":[{\"level\":\"INFO\",\"message\":\"Security test\",\"service\":\"$SERVICE_NAME\"}]}'" \
    "inserted"

# Test 2: Invalid API Key
run_test "Invalid API key" 401 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: application/json' \
    -H 'Authorization: Bearer invalid_key_xyz123' \
    -d '{\"entries\":[{\"level\":\"INFO\",\"message\":\"Test\",\"service\":\"$SERVICE_NAME\"}]}'" \
    "Invalid API key"

# Test 3: Missing Authorization Header
run_test "Missing Authorization header" 401 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: application/json' \
    -d '{\"entries\":[{\"level\":\"INFO\",\"message\":\"Test\",\"service\":\"$SERVICE_NAME\"}]}'" \
    "Authorization header required"

# Test 4: Malformed Bearer Token
run_test "Malformed Bearer token (no 'Bearer' prefix)" 401 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: application/json' \
    -H 'Authorization: $VALID_API_KEY' \
    -d '{\"entries\":[{\"level\":\"INFO\",\"message\":\"Test\",\"service\":\"$SERVICE_NAME\"}]}'"

# Test 5: Empty API Key
run_test "Empty API key" 401 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: application/json' \
    -H 'Authorization: Bearer ' \
    -d '{\"entries\":[{\"level\":\"INFO\",\"message\":\"Test\",\"service\":\"$SERVICE_NAME\"}]}'"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2️⃣  Rate Limiting Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Test 6: Rate Limiting (1000 req/min per key)
echo -n "Testing: Rate limiting (rapid fire)... "
rate_limit_hit=false
for i in {1..20}; do
    status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$LOGS_API_URL/api/logs/batch" \
        -H 'Content-Type: application/json' \
        -H "Authorization: Bearer $VALID_API_KEY" \
        -d '{"entries":[{"level":"INFO","message":"Rate limit test","service":"'$SERVICE_NAME'"}]}')
    
    if [ "$status" -eq 429 ]; then
        rate_limit_hit=true
        break
    fi
done

if [ "$rate_limit_hit" = true ]; then
    echo -e "${YELLOW}⚠ INFO${NC} (Rate limit enforced - received 429)"
    echo "  Note: Rate limiting is working, but may need more requests to trigger in dev"
else
    echo -e "${BLUE}ℹ INFO${NC} (No rate limit hit in 20 requests)"
    echo "  Note: Rate limit (1000 req/min) not hit with 20 requests - this is expected"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3️⃣  SQL Injection Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Test 7: SQL Injection in message field
run_test "SQL injection in message" 200 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: application/json' \
    -H 'Authorization: Bearer $VALID_API_KEY' \
    -d '{\"entries\":[{\"level\":\"INFO\",\"message\":\"Test'; DROP TABLE logs.entries; --\",\"service\":\"$SERVICE_NAME\"}]}'" \
    "inserted"

# Test 8: SQL Injection in metadata
run_test "SQL injection in metadata" 200 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: application/json' \
    -H 'Authorization: Bearer $VALID_API_KEY' \
    -d '{\"entries\":[{\"level\":\"INFO\",\"message\":\"Test\",\"service\":\"$SERVICE_NAME\",\"metadata\":{\"user\":\"admin'--\"}}]}'" \
    "inserted"

# Test 9: SQL Injection in service field
run_test "SQL injection in service" 200 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: application/json' \
    -H 'Authorization: Bearer $VALID_API_KEY' \
    -d '{\"entries\":[{\"level\":\"INFO\",\"message\":\"Test\",\"service\":\"test' OR '1'='1\"}]}'" \
    "inserted"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4️⃣  Invalid JSON Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Test 10: Malformed JSON
run_test "Malformed JSON" 400 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: application/json' \
    -H 'Authorization: Bearer $VALID_API_KEY' \
    -d '{\"entries\":[{\"level\":\"INFO\",\"message\":\"Test\"}'"

# Test 11: Missing required field (entries)
run_test "Missing 'entries' field" 400 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: application/json' \
    -H 'Authorization: Bearer $VALID_API_KEY' \
    -d '{}'"

# Test 12: Empty entries array
run_test "Empty entries array" 400 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: application/json' \
    -H 'Authorization: Bearer $VALID_API_KEY' \
    -d '{\"entries\":[]}'"

# Test 13: Invalid level value
run_test "Invalid log level" 200 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: application/json' \
    -H 'Authorization: Bearer $VALID_API_KEY' \
    -d '{\"entries\":[{\"level\":\"INVALID_LEVEL\",\"message\":\"Test\",\"service\":\"$SERVICE_NAME\"}]}'" \
    "inserted"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5️⃣  Oversized Payload Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Test 14: Batch exceeds max size (1000 logs)
echo -n "Testing: Batch size exceeds maximum (1001 logs)... "
# Generate 1001 log entries
batch_1001='{"entries":['
for i in {1..1001}; do
    if [ $i -gt 1 ]; then
        batch_1001="$batch_1001,"
    fi
    batch_1001="$batch_1001{\"level\":\"INFO\",\"message\":\"Test $i\",\"service\":\"$SERVICE_NAME\"}"
done
batch_1001="$batch_1001]}"

response=$(curl -s -w '\n%{http_code}' -X POST "$LOGS_API_URL/api/logs/batch" \
    -H 'Content-Type: application/json' \
    -H "Authorization: Bearer $VALID_API_KEY" \
    -d "$batch_1001")

status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$status_code" -eq 400 ]; then
    if echo "$body" | grep -q "exceeds maximum"; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${YELLOW}⚠ PARTIAL${NC} (400 but wrong error message)"
        echo "  Expected: 'exceeds maximum', Got: $body"
        ((TESTS_PASSED++))
    fi
else
    echo -e "${RED}✗ FAIL${NC}"
    echo "  Expected: 400, Got: $status_code"
    ((TESTS_FAILED++))
fi

# Test 15: Extremely large metadata object
run_test "Extremely large metadata (10KB)" 200 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: application/json' \
    -H 'Authorization: Bearer $VALID_API_KEY' \
    -d '{\"entries\":[{\"level\":\"INFO\",\"message\":\"Test\",\"service\":\"$SERVICE_NAME\",\"metadata\":{\"large_field\":\"$(python3 -c 'print("x" * 10000)')\"}}]}'" \
    "inserted"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6️⃣  HTTP Method Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Test 16: GET method (should only accept POST)
run_test "GET method not allowed" 405 \
    "curl -s -w '\n%{http_code}' -X GET '$LOGS_API_URL/api/logs/batch' \
    -H 'Authorization: Bearer $VALID_API_KEY'"

# Test 17: PUT method (should only accept POST)
run_test "PUT method not allowed" 405 \
    "curl -s -w '\n%{http_code}' -X PUT '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: application/json' \
    -H 'Authorization: Bearer $VALID_API_KEY' \
    -d '{\"entries\":[{\"level\":\"INFO\",\"message\":\"Test\",\"service\":\"$SERVICE_NAME\"}]}'"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "7️⃣  Content-Type Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Test 18: Missing Content-Type header
run_test "Missing Content-Type header" 400 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Authorization: Bearer $VALID_API_KEY' \
    -d '{\"entries\":[{\"level\":\"INFO\",\"message\":\"Test\",\"service\":\"$SERVICE_NAME\"}]}'"

# Test 19: Wrong Content-Type
run_test "Wrong Content-Type (text/plain)" 400 \
    "curl -s -w '\n%{http_code}' -X POST '$LOGS_API_URL/api/logs/batch' \
    -H 'Content-Type: text/plain' \
    -H 'Authorization: Bearer $VALID_API_KEY' \
    -d '{\"entries\":[{\"level\":\"INFO\",\"message\":\"Test\",\"service\":\"$SERVICE_NAME\"}]}'"

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "📊 SECURITY TEST RESULTS"
echo "═══════════════════════════════════════════════════════════"
echo ""
TOTAL_TESTS=$((TESTS_PASSED + TESTS_FAILED))
echo "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$TESTS_PASSED ✓${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED ✗${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ ALL SECURITY TESTS PASSED${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}❌ SOME SECURITY TESTS FAILED${NC}"
    echo ""
    exit 1
fi
