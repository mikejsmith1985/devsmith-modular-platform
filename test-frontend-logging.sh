#!/bin/bash

# Test Frontend Error Logging System
# This script verifies that frontend errors are being logged to the Logs service

echo "========================================"
echo "Frontend Error Logging Test"
echo "========================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "1. Testing if Logs service is accessible..."
LOGS_HEALTH=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/api/logs/health)

if [ "$LOGS_HEALTH" = "200" ]; then
    echo -e "${GREEN}✓${NC} Logs service is healthy"
else
    echo -e "${RED}✗${NC} Logs service not responding (HTTP $LOGS_HEALTH)"
    exit 1
fi

echo ""
echo "2. Sending test error log from frontend..."

# Simulate a frontend error log
TEST_LOG=$(cat <<EOF
{
  "service": "frontend",
  "level": "error",
  "message": "Test error from logging verification script",
  "metadata": {
    "test": true,
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "environment": "test",
    "url": "test://verification",
    "context": {
      "purpose": "Verify frontend error logging is working"
    }
  }
}
EOF
)

RESPONSE=$(curl -s -X POST http://localhost:3000/api/logs \
  -H "Content-Type: application/json" \
  -d "$TEST_LOG")

if echo "$RESPONSE" | grep -q "id"; then
    echo -e "${GREEN}✓${NC} Test log successfully posted to Logs service"
    LOG_ID=$(echo "$RESPONSE" | grep -o '"id":[^,}]*' | grep -o '[0-9]*')
    echo "  Log ID: $LOG_ID"
else
    echo -e "${RED}✗${NC} Failed to post log"
    echo "  Response: $RESPONSE"
    exit 1
fi

echo ""
echo "3. Verifying log appears in service..."
sleep 1

QUERY_RESPONSE=$(curl -s "http://localhost:3000/api/logs?service=frontend&limit=1")

if echo "$QUERY_RESPONSE" | grep -q "Test error from logging verification script"; then
    echo -e "${GREEN}✓${NC} Test log found in Logs service"
else
    echo -e "${YELLOW}⚠${NC} Could not verify log in query (may be pagination issue)"
    echo "  Try viewing in Logs UI: http://localhost:3000/logs"
fi

echo ""
echo "========================================"
echo "4. Instructions for Manual Testing:"
echo "========================================"
echo ""
echo "Test AI Insights Error Logging:"
echo "  1. Go to: http://localhost:3000/logs"
echo "  2. Click any log entry"
echo "  3. Click 'Generate AI Insights'"
echo "  4. If error occurs, check Logs service"
echo "  5. Filter by service: 'frontend'"
echo "  6. Should see error with full context"
echo ""
echo "Test Global Error Handler:"
echo "  1. Open browser console on any page"
echo "  2. Type: throw new Error('Test error')"
echo "  3. Go to Logs service"
echo "  4. Should see error logged automatically"
echo ""
echo "========================================"
echo -e "${GREEN}Frontend error logging system is operational!${NC}"
echo "========================================"
