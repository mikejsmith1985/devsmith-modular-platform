#!/bin/bash
# Setup script for cross-repo logging integration tests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Cross-Repo Logging Test Environment Setup"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if Docker containers are running
echo "Checking Docker containers..."
if ! docker ps | grep -q "devsmith-modular-platform-logs"; then
    echo "❌ Logs service not running. Please start Docker:"
    echo "   cd $PROJECT_ROOT && docker-compose up -d"
    exit 1
fi
echo "✅ Docker containers running"
echo ""

# Check if database is accessible
echo "Checking database connectivity..."
if ! docker exec -i devsmith-modular-platform-postgres-1 psql -U devsmith -d devsmith -c "SELECT 1" > /dev/null 2>&1; then
    echo "❌ Cannot connect to database"
    exit 1
fi
echo "✅ Database accessible"
echo ""

# Create test project
echo "Creating test project..."
TEST_PROJECT_SQL=$(cat <<EOF
-- Delete existing test project (if exists)
DELETE FROM logs.projects WHERE slug = 'test-integration';

-- Create test project
INSERT INTO logs.projects (user_id, name, slug, description, api_key_hash)
VALUES (
    1,  -- Assuming test user ID = 1
    'Test Integration Project',
    'test-integration',
    'Automated test project for integration tests',
    '\$2a\$10\$K6FHZ8VXvXZ0KwN0Z0Z0ZuXZ0Z0Z0Z0Z0Z0Z0Z0Z0Z0Z0Z0Z0Z0'  -- Hashed "test-api-key-12345678901234567890"
);

-- Get project ID
SELECT id, name, slug FROM logs.projects WHERE slug = 'test-integration';
EOF
)

RESULT=$(docker exec -i devsmith-modular-platform-postgres-1 psql -U devsmith -d devsmith -t <<< "$TEST_PROJECT_SQL" 2>&1)
if echo "$RESULT" | grep -q "test-integration"; then
    PROJECT_ID=$(echo "$RESULT" | grep -oP '\d+' | head -1)
    echo "✅ Test project created (ID: $PROJECT_ID)"
else
    echo "❌ Failed to create test project"
    echo "$RESULT"
    exit 1
fi
echo ""

# Save test configuration
TEST_CONFIG_FILE="$SCRIPT_DIR/.test-config.json"
cat > "$TEST_CONFIG_FILE" <<EOF
{
  "apiUrl": "http://localhost:3000",
  "apiKey": "test-api-key-12345678901234567890",
  "projectSlug": "test-integration",
  "projectId": $PROJECT_ID,
  "batchEndpoint": "http://localhost:8082/api/logs/batch"
}
EOF
echo "✅ Test configuration saved: $TEST_CONFIG_FILE"
echo ""

# Install test dependencies
echo "Installing test dependencies..."
cd "$PROJECT_ROOT"
if [ ! -d "node_modules" ]; then
    npm install --silent
fi
echo "✅ Dependencies installed"
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Test Environment Ready!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Test Configuration:"
echo "  API URL:      http://localhost:3000"
echo "  API Key:      test-api-key-12345678901234567890"
echo "  Project Slug: test-integration"
echo "  Project ID:   $PROJECT_ID"
echo ""
echo "Run tests:"
echo "  npm test                    # All tests"
echo "  npm run test:logger:js      # JavaScript logger tests"
echo "  npm run test:logger:py      # Python logger tests"
echo "  npm run test:logger:go      # Go logger tests"
echo "  npm run test:express        # Express middleware tests"
echo "  npm run test:flask          # Flask extension tests"
echo "  npm run test:gin            # Gin middleware tests"
echo ""
