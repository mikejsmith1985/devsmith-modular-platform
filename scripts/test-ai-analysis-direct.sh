#!/bin/bash
# Direct API test for AI analysis endpoints
# This script tests both Review and Logs AI analysis with a real authenticated session

set -e

echo "=== Setting up test environment ==="

# 1. Create test user and session
echo "Creating test user session..."
RESPONSE=$(curl -s -X POST http://localhost:3000/auth/test-login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","avatar_url":"https://example.com/avatar.png","github_id":"123456"}' \
  -c /tmp/test-cookies-direct.txt)

echo "Test login response: $RESPONSE"

# Extract session cookie
SESSION_COOKIE=$(grep devsmith_token /tmp/test-cookies-direct.txt | awk '{print $NF}')
echo "Session cookie: ${SESSION_COOKIE:0:50}..."

if [ -z "$SESSION_COOKIE" ]; then
  echo "ERROR: No session cookie found!"
  exit 1
fi

# 2. Ensure LLM config exists for test user
echo ""
echo "=== Ensuring LLM config for test user ==="
docker-compose exec -T postgres psql -U devsmith -d devsmith << 'EOF'
-- Delete existing config and recreate
DELETE FROM portal.llm_configs WHERE user_id = 2;
DELETE FROM portal.users WHERE id = 2;

-- Create test user
INSERT INTO portal.users (id, github_id, username, email, avatar_url, created_at, updated_at)
VALUES (2, 123456, 'testuser', 'test@example.com', 'https://example.com/avatar.png', NOW(), NOW());

-- Create LLM config
INSERT INTO portal.llm_configs (id, user_id, provider, model_name, api_endpoint, is_default, max_tokens, temperature, created_at, updated_at)
VALUES ('test-user-ollama', 2, 'ollama', 'qwen2.5-coder:7b-instruct-q5_K_M', 'http://host.docker.internal:11434', true, 2000, 0.7, NOW(), NOW());

-- Verify configuration
SELECT 'User and LLM config:' as info;
SELECT u.id, u.username, c.provider, c.model_name, c.is_default, c.api_endpoint
FROM portal.users u
LEFT JOIN portal.llm_configs c ON u.id = c.user_id
WHERE u.id = 2;
EOF

# 3. Test Review API (Preview mode)
echo ""
echo "=== Testing Review API (Preview Mode) ==="
echo "Request: POST /api/review/modes/preview"

REVIEW_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}\n" \
  -X POST http://localhost:3000/api/review/modes/preview \
  -H "Content-Type: application/json" \
  -H "Cookie: devsmith_token=$SESSION_COOKIE" \
  -d '{
    "pasted_code": "function greet(name) { console.log(\"Hello, \" + name); }",
    "model": "qwen2.5-coder:7b-instruct-q5_K_M",
    "user_mode": "intermediate",
    "output_mode": "quick"
  }')

echo "$REVIEW_RESPONSE"

HTTP_STATUS=$(echo "$REVIEW_RESPONSE" | grep "HTTP_STATUS:" | cut -d':' -f2)
RESPONSE_BODY=$(echo "$REVIEW_RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$HTTP_STATUS" = "200" ]; then
  echo "✅ Review API SUCCESS"
else
  echo "❌ Review API FAILED with status $HTTP_STATUS"
  echo "Response: $RESPONSE_BODY"
fi

# 4. Create a test log entry for Logs AI insights
echo ""
echo "=== Creating test log entry ==="
LOG_INSERT=$(docker-compose exec -T postgres psql -U devsmith -d devsmith -c "
INSERT INTO logs.entries (service, level, message, metadata, created_at)
VALUES ('test-service', 'ERROR', 'Test error for AI analysis', '{\"error\":\"sample error\"}', NOW())
RETURNING id;
" -t -A)

LOG_ID=$(echo "$LOG_INSERT" | tail -1)
echo "Created log entry with ID: $LOG_ID"

# 5. Test Logs API (AI Insights)
echo ""
echo "=== Testing Logs API (AI Insights) ==="
echo "Request: POST /api/logs/$LOG_ID/insights"

LOGS_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}\n" \
  -X POST "http://localhost:3000/api/logs/$LOG_ID/insights" \
  -H "Content-Type: application/json" \
  -H "Cookie: devsmith_token=$SESSION_COOKIE" \
  -d '{
    "model": "qwen2.5-coder:7b-instruct-q5_K_M"
  }')

echo "$LOGS_RESPONSE"

HTTP_STATUS=$(echo "$LOGS_RESPONSE" | grep "HTTP_STATUS:" | cut -d':' -f2)
RESPONSE_BODY=$(echo "$LOGS_RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$HTTP_STATUS" = "200" ]; then
  echo "✅ Logs AI Insights SUCCESS"
else
  echo "❌ Logs AI Insights FAILED with status $HTTP_STATUS"
  echo "Response: $RESPONSE_BODY"
fi

# 6. Summary
echo ""
echo "=== TEST SUMMARY ==="
echo "Review API: $([ "$HTTP_STATUS" = "200" ] && echo '✅ PASS' || echo '❌ FAIL')"
echo "Logs AI Insights: $([ "$HTTP_STATUS" = "200" ] && echo '✅ PASS' || echo '❌ FAIL')"

echo ""
echo "Check service logs for detailed error messages:"
echo "  docker-compose logs review --tail=100"
echo "  docker-compose logs logs --tail=100"
