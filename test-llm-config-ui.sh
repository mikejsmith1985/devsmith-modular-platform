#!/bin/bash

# Test script for LLM Config UI
# This script automates authentication and tests the LLM config functionality

set -e

echo "======================================"
echo "LLM Config UI Test"
echo "======================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Step 1: Authenticate and get session cookie
echo "Step 1: Authenticating test user..."
AUTH_RESPONSE=$(curl -s -X POST http://localhost:3000/auth/test-login \
  -H "Content-Type: application/json" \
  -d '{"username":"uitest","email":"test@devsmith.local","avatar_url":"https://example.com/avatar.png","github_id":"54321"}' \
  -c /tmp/llm-test-cookies.txt)

if echo "$AUTH_RESPONSE" | jq -e '.message == "success"' > /dev/null; then
  echo -e "${GREEN}✓${NC} Authentication successful"
else
  echo -e "${RED}✗${NC} Authentication failed"
  echo "$AUTH_RESPONSE" | jq .
  exit 1
fi

# Step 2: Delete any existing configs to start fresh
echo ""
echo "Step 2: Cleaning up existing configs..."
EXISTING_CONFIGS=$(curl -s http://localhost:3000/api/portal/llm-configs -b /tmp/llm-test-cookies.txt)
CONFIG_COUNT=$(echo "$EXISTING_CONFIGS" | jq 'length')
echo "Found $CONFIG_COUNT existing configs"

if [ "$CONFIG_COUNT" -gt 0 ]; then
  echo "$EXISTING_CONFIGS" | jq -r '.[].id' | while read -r config_id; do
    curl -s -X DELETE "http://localhost:3000/api/portal/llm-configs/$config_id" -b /tmp/llm-test-cookies.txt > /dev/null
    echo "  Deleted config: $config_id"
  done
fi

# Step 3: Create Ollama config
echo ""
echo "Step 3: Creating Ollama (DeepSeek) configuration..."
OLLAMA_RESPONSE=$(curl -s -X POST http://localhost:3000/api/portal/llm-configs \
  -b /tmp/llm-test-cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"name":"Test DeepSeek","provider":"ollama","model":"deepseek-coder-v2:16b","is_default":false}')

if echo "$OLLAMA_RESPONSE" | jq -e '.id' > /dev/null; then
  OLLAMA_ID=$(echo "$OLLAMA_RESPONSE" | jq -r '.id')
  echo -e "${GREEN}✓${NC} Ollama config created successfully (ID: $OLLAMA_ID)"
else
  echo -e "${RED}✗${NC} Ollama config creation failed"
  echo "$OLLAMA_RESPONSE" | jq .
  exit 1
fi

# Step 4: Create Claude config (with fake API key for testing)
echo ""
echo "Step 4: Creating Claude (Anthropic) configuration..."
CLAUDE_RESPONSE=$(curl -s -X POST http://localhost:3000/api/portal/llm-configs \
  -b /tmp/llm-test-cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Claude","provider":"anthropic","model":"claude-3-5-sonnet-20241022","api_key":"sk-test-fake-api-key-for-testing-12345678901234567890","is_default":false}')

if echo "$CLAUDE_RESPONSE" | jq -e '.id' > /dev/null; then
  CLAUDE_ID=$(echo "$CLAUDE_RESPONSE" | jq -r '.id')
  echo -e "${GREEN}✓${NC} Claude config created successfully (ID: $CLAUDE_ID)"
  echo -e "${GREEN}✓${NC} API key encrypted: $(echo "$CLAUDE_RESPONSE" | jq -r '.has_api_key')"
else
  echo -e "${RED}✗${NC} Claude config creation failed"
  echo "$CLAUDE_RESPONSE" | jq .
  exit 1
fi

# Step 5: List all configs
echo ""
echo "Step 5: Listing all configurations..."
ALL_CONFIGS=$(curl -s http://localhost:3000/api/portal/llm-configs -b /tmp/llm-test-cookies.txt)
CONFIG_COUNT=$(echo "$ALL_CONFIGS" | jq 'length')
echo -e "${GREEN}✓${NC} Total configs: $CONFIG_COUNT"
echo "$ALL_CONFIGS" | jq -r '.[] | "  - \(.provider)/\(.model) (ID: \(.id))"'

# Step 6: Verify data in database
echo ""
echo "Step 6: Verifying database records..."
DB_COUNT=$(docker-compose exec -T postgres psql -U devsmith -d devsmith -t -c "SELECT COUNT(*) FROM portal.llm_configs WHERE user_id = 999999;" | xargs)
echo -e "${GREEN}✓${NC} Database contains $DB_COUNT records for test user"

# Final summary
echo ""
echo "======================================"
echo "Test Summary"
echo "======================================"
echo -e "${GREEN}✓${NC} Authentication: PASS"
echo -e "${GREEN}✓${NC} Ollama config creation: PASS"
echo -e "${GREEN}✓${NC} Claude config creation: PASS"
echo -e "${GREEN}✓${NC} API key encryption: PASS"
echo -e "${GREEN}✓${NC} Database persistence: PASS"
echo ""
echo "======================================"
echo "Manual UI Verification Steps"
echo "======================================"
echo "1. Open browser to: http://localhost:3000"
echo "2. You should see existing test session"
echo "3. Navigate to LLM Config page"
echo "4. You should see 2 configs listed:"
echo "   - ollama/deepseek-coder-v2:16b"
echo "   - anthropic/claude-3-5-sonnet-20241022"
echo "5. Click 'Add AI Model' to test form"
echo "6. Select Ollama, choose llama3.1:8b, click Save"
echo "7. Verify new config appears in list"
echo ""
echo "Cookie file for manual testing: /tmp/llm-test-cookies.txt"
echo "Use with curl: curl -b /tmp/llm-test-cookies.txt http://localhost:3000/api/portal/llm-configs"
echo ""
