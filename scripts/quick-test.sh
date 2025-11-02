#!/bin/bash

# Quick test script for immediate feedback after service restart

echo "=== Quick Review Service Test ==="
echo ""

# Test 1: Health check
echo "1. Testing service health..."
if curl -s http://localhost:8081/health | grep -q "healthy"; then
    echo "   ✓ Service is healthy"
else
    echo "   ✗ Service health check failed"
    echo "   Run: docker-compose logs review"
    exit 1
fi

# Test 2: Simple form submission
echo ""
echo "2. Testing form submission..."
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8081/api/review/modes/preview \
  -d "pasted_code=test")

HTTP_CODE=$(echo "$RESPONSE" | tail -1)
if [ "$HTTP_CODE" = "200" ]; then
    echo "   ✓ Form submission returns 200 OK"
else
    echo "   ✗ Form submission failed (HTTP $HTTP_CODE)"
    echo "   Response:"
    echo "$RESPONSE" | head -n -1 | head -20
    exit 1
fi

# Test 3: Models endpoint
echo ""
echo "3. Testing models endpoint..."
if curl -s http://localhost:8081/api/review/models | grep -q "mistral"; then
    echo "   ✓ Models endpoint working"
else
    echo "   ✗ Models endpoint failed"
    exit 1
fi

# Test 4: Database connections
echo ""
echo "4. Checking database connections..."
CONN_COUNT=$(docker exec devsmith-postgres psql -U devsmith -d devsmith -t -c \
  "SELECT count(*) FROM pg_stat_activity WHERE datname='devsmith';" 2>/dev/null | tr -d ' ')

if [ -n "$CONN_COUNT" ]; then
    echo "   Current connections: $CONN_COUNT"
    if [ "$CONN_COUNT" -lt 50 ]; then
        echo "   ✓ Connection count is healthy"
    else
        echo "   ⚠ Connection count is high"
    fi
else
    echo "   ⚠ Could not check connections (PostgreSQL not ready?)"
fi

echo ""
echo "✓ Quick tests passed!"
echo ""
echo "Next: Test in browser at http://localhost:3000/review"
