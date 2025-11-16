#!/bin/bash
set -e

echo "=== DevSmith Review Service - Fix Verification ==="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function for test results
test_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $2"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: $2"
        ((TESTS_FAILED++))
    fi
}

echo "Phase 1: Testing Core Functionality Fixes"
echo "=========================================="
echo ""

# Test 1: Form Data Binding
echo "Test 1: Form data binding with code submission"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "http://localhost:8081/api/review/modes/preview" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "pasted_code=package main

func main() {
    println(\"hello world\")
}")

HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "200" ]; then
    test_result 0 "Form binding returns 200 OK"
    
    if echo "$BODY" | grep -q "Analysis" || echo "$BODY" | grep -q "Structure"; then
        test_result 0 "Response contains analysis content"
    else
        test_result 1 "Response missing analysis content"
        echo "   Response body: $BODY"
    fi
else
    test_result 1 "Form binding failed with HTTP $HTTP_CODE"
    echo "   Response: $BODY"
fi
echo ""

# Test 2: Model Selection
echo "Test 2: Model parameter accepted in request"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "http://localhost:8081/api/review/modes/preview" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "pasted_code=test code&model=deepseek-coder:6.7b")

HTTP_CODE=$(echo "$RESPONSE" | tail -1)
if [ "$HTTP_CODE" = "200" ]; then
    test_result 0 "Model parameter accepted (HTTP 200)"
else
    test_result 1 "Model parameter not working (HTTP $HTTP_CODE)"
fi
echo ""

# Test 3: All 5 Modes Work
echo "Test 3: All 5 reading modes functional"
MODES=("preview" "skim" "scan" "detailed" "critical")
for mode in "${MODES[@]}"; do
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
      "http://localhost:8081/api/review/modes/$mode" \
      -d "pasted_code=test")
    
    if [ "$HTTP_CODE" = "200" ]; then
        test_result 0 "Mode '$mode' returns 200"
    else
        test_result 1 "Mode '$mode' failed (HTTP $HTTP_CODE)"
    fi
done
echo ""

# Test 4: Models Endpoint
echo "Test 4: Models endpoint returns available models"
RESPONSE=$(curl -s http://localhost:8081/api/review/models)
if echo "$RESPONSE" | grep -q "mistral" && echo "$RESPONSE" | grep -q "codellama"; then
    test_result 0 "Models endpoint returns model list"
else
    test_result 1 "Models endpoint not working"
    echo "   Response: $RESPONSE"
fi
echo ""

# Test 5: Database Connection Check
echo "Test 5: Database connection pool configured"
echo "   Checking PostgreSQL connections..."
CONN_COUNT=$(docker exec devsmith-postgres psql -U devsmith -d devsmith -t -c \
  "SELECT count(*) FROM pg_stat_activity WHERE datname='devsmith';" 2>/dev/null || echo "0")

echo "   Current connections: $CONN_COUNT"
if [ "$CONN_COUNT" -lt 50 ]; then
    test_result 0 "Connection count is healthy (<50)"
else
    test_result 1 "Connection count is high (>50)"
fi
echo ""

# Test 6: Empty Code Validation
echo "Test 6: Empty code shows validation error"
RESPONSE=$(curl -s -X POST "http://localhost:8081/api/review/modes/preview" \
  -d "pasted_code=")
  
if echo "$RESPONSE" | grep -iq "code required" || echo "$RESPONSE" | grep -iq "required"; then
    test_result 0 "Empty code validation works"
else
    test_result 1 "Empty code validation not working"
    echo "   Response: $RESPONSE"
fi
echo ""

# Test 7: Service Health
echo "Test 7: Service health check"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/health)
if [ "$HTTP_CODE" = "200" ]; then
    test_result 0 "Service is healthy"
else
    test_result 1 "Service health check failed"
fi
echo ""

echo "=========================================="
echo "Summary:"
echo "  Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo "  Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    echo ""
    echo "Next Steps:"
    echo "1. Test in browser at http://localhost:3000/review"
    echo "2. Paste code in textarea"
    echo "3. Select a model from dropdown"
    echo "4. Click a mode button"
    echo "5. Verify analysis appears"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    echo ""
    echo "Debug Steps:"
    echo "1. Check service logs: docker-compose logs review"
    echo "2. Check database: docker exec devsmith-postgres psql -U devsmith"
    echo "3. Verify Ollama: curl http://localhost:11434/api/tags"
    exit 1
fi
