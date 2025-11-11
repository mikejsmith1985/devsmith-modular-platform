# Testing Sample Integration Files

Validate that sample loggers work correctly against DevSmith API.

---

## Prerequisites

1. DevSmith platform running: `docker-compose up -d`
2. Test project created in database
3. API key generated

---

## Setup Test Project

### Option A: Via SQL (Quick)

```bash
# Connect to database
docker exec -it devsmith-modular-platform-postgres-1 psql -U devsmith -d devsmith

# Create test project
INSERT INTO logs.projects (user_id, name, slug, description, api_key_hash, is_active)
VALUES (
  1,  -- Replace with your user_id from portal.users
  'Test Project',
  'test-project',
  'Testing cross-repo logging integration',
  '$2a$10$EXAMPLE_HASH_WILL_BE_GENERATED',  -- We'll update this
  true
);

# Get project ID
SELECT id FROM logs.projects WHERE slug = 'test-project';
-- Remember this ID (e.g., 1)
```

### Option B: Via API (Future - Week 3)

```bash
# Create project via API
curl -X POST http://localhost:3000/api/projects \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Project",
    "slug": "test-project",
    "description": "Testing cross-repo logging"
  }'

# Response includes API key (save it!)
# {
#   "id": 1,
#   "name": "Test Project",
#   "slug": "test-project",
#   "api_key": "dsk_abc123xyz789..."
# }
```

---

## Generate API Key Manually

For testing, we need to generate an API key and hash it:

```bash
# Generate random API key (32 bytes base64)
API_KEY="dsk_$(openssl rand -base64 32 | tr -d '=' | tr '+/' '-_')"
echo "Generated API Key: $API_KEY"
echo "Save this for testing!"

# Hash the key with bcrypt (requires htpasswd or Python)
# Method 1: Using htpasswd (if available)
API_KEY_HASH=$(htpasswd -bnBC 10 "" "${API_KEY:4}" | tr -d ':\n')

# Method 2: Using Python
API_KEY_HASH=$(python3 -c "import bcrypt; print(bcrypt.hashpw('${API_KEY:4}'.encode(), bcrypt.gensalt(10)).decode())")

echo "API Key Hash: $API_KEY_HASH"

# Update database with hashed key
docker exec -it devsmith-modular-platform-postgres-1 psql -U devsmith -d devsmith -c \
  "UPDATE logs.projects SET api_key_hash = '$API_KEY_HASH' WHERE slug = 'test-project';"
```

**Important**: The API key format is `dsk_<random>` but we hash only the `<random>` part (after `dsk_`).

---

## Test 1: JavaScript Logger

**Create test file:**
```bash
cd docs/integrations/javascript
cat > test.js << 'EOF'
const DevSmithLogger = require('./logger');

const logger = new DevSmithLogger({
  apiKey: process.env.DEVSMITH_API_KEY,
  apiUrl: process.env.DEVSMITH_API_URL || 'http://localhost:3000',
  projectSlug: 'test-project',
  serviceName: 'nodejs-test',
  batchSize: 5  // Small batch for quick testing
});

console.log('Sending test logs...');

logger.debug('Debug message', { test: true, iteration: 1 });
logger.info('Info message', { test: true, iteration: 2 });
logger.warn('Warning message', { test: true, iteration: 3 });
logger.error('Error message', { test: true, iteration: 4, error: 'Test error' });

// Send one more to trigger batch (5 logs)
logger.info('Batch trigger', { test: true, iteration: 5 });

console.log('Logs sent! Check DevSmith Health dashboard.');

// Wait for final flush
setTimeout(() => {
  console.log('Test complete!');
  process.exit(0);
}, 6000);  // Wait for flush interval
EOF
```

**Run test:**
```bash
export DEVSMITH_API_KEY="dsk_abc123xyz789..."  # Your generated key
node test.js
```

**Expected output:**
```
Sending test logs...
Logs sent! Check DevSmith Health dashboard.
[DevSmithLogger] Successfully sent 5 logs
Test complete!
```

**Verify in database:**
```bash
docker exec -it devsmith-modular-platform-postgres-1 psql -U devsmith -d devsmith -c \
  "SELECT COUNT(*) FROM logs.entries WHERE service = 'nodejs-test';"
# Expected: 5
```

---

## Test 2: Python Logger

**Create test file:**
```bash
cd docs/integrations/python
cat > test.py << 'EOF'
import os
import time
from logger import DevSmithLogger

logger = DevSmithLogger(
    api_key=os.getenv('DEVSMITH_API_KEY'),
    api_url=os.getenv('DEVSMITH_API_URL', 'http://localhost:3000'),
    project_slug='test-project',
    service_name='python-test',
    batch_size=5  # Small batch for quick testing
)

print('Sending test logs...')

logger.debug('Debug message', test=True, iteration=1)
logger.info('Info message', test=True, iteration=2)
logger.warn('Warning message', test=True, iteration=3)
logger.error('Error message', test=True, iteration=4, error='Test error')

# Send one more to trigger batch (5 logs)
logger.info('Batch trigger', test=True, iteration=5)

print('Logs sent! Check DevSmith Health dashboard.')

# Wait for final flush
time.sleep(6)
print('Test complete!')
EOF
```

**Run test:**
```bash
export DEVSMITH_API_KEY="dsk_abc123xyz789..."  # Your generated key
python3 test.py
```

**Expected output:**
```
Sending test logs...
Logs sent! Check DevSmith Health dashboard.
[DevSmithLogger] Successfully sent 5 logs
Test complete!
```

**Verify in database:**
```bash
docker exec -it devsmith-modular-platform-postgres-1 psql -U devsmith -d devsmith -c \
  "SELECT COUNT(*) FROM logs.entries WHERE service = 'python-test';"
# Expected: 5
```

---

## Test 3: Go Logger

**Create test file:**
```bash
cd docs/integrations/go
cat > test.go << 'EOF'
package main

import (
    "fmt"
    "os"
    "time"
)

func main() {
    logger := NewLoggerWithOptions(
        os.Getenv("DEVSMITH_API_KEY"),
        getEnvOrDefault("DEVSMITH_API_URL", "http://localhost:3000"),
        "test-project",
        "go-test",
        5,             // Small batch for quick testing
        2*time.Second, // Faster flush for testing
    )
    defer logger.Close()

    fmt.Println("Sending test logs...")

    logger.Debug("Debug message", map[string]interface{}{"test": true, "iteration": 1})
    logger.Info("Info message", map[string]interface{}{"test": true, "iteration": 2})
    logger.Warn("Warning message", map[string]interface{}{"test": true, "iteration": 3})
    logger.Error("Error message", map[string]interface{}{
        "test":      true,
        "iteration": 4,
        "error":     "Test error",
    })

    // Send one more to trigger batch (5 logs)
    logger.Info("Batch trigger", map[string]interface{}{"test": true, "iteration": 5})

    fmt.Println("Logs sent! Check DevSmith Health dashboard.")

    // Wait for final flush
    time.Sleep(3 * time.Second)
    fmt.Println("Test complete!")
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
EOF
```

**Run test:**
```bash
export DEVSMITH_API_KEY="dsk_abc123xyz789..."  # Your generated key
go run test.go logger.go
```

**Expected output:**
```
Sending test logs...
Logs sent! Check DevSmith Health dashboard.
[DevSmithLogger] Successfully sent 5 logs
Test complete!
```

**Verify in database:**
```bash
docker exec -it devsmith-modular-platform-postgres-1 psql -U devsmith -d devsmith -c \
  "SELECT COUNT(*) FROM logs.entries WHERE service = 'go-test';"
# Expected: 5
```

---

## Test 4: Error Scenarios

### Invalid API Key
```bash
export DEVSMITH_API_KEY="dsk_invalid_key"
node test.js
```

**Expected output:**
```
[DevSmithLogger] Error sending logs: 401 Unauthorized - Invalid API key
```

### Network Error (Service Down)
```bash
docker-compose stop logs
export DEVSMITH_API_KEY="dsk_abc123xyz789..."
node test.js
docker-compose start logs
```

**Expected output:**
```
[DevSmithLogger] Error sending logs: ECONNREFUSED
[DevSmithLogger] Logs re-added to buffer for retry
```

### Rate Limiting (Future)
```bash
# Send 1000 requests in 1 minute (exceeds 1000/min limit)
for i in {1..1000}; do
  curl -X POST http://localhost:3000/api/logs/batch \
    -H "Authorization: Bearer $DEVSMITH_API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"project_slug":"test-project","logs":[{"timestamp":"2025-11-11T12:00:00Z","level":"INFO","message":"Test","service":"test","context":{}}]}'
done
```

**Expected**: After ~1000 requests, receive `429 Too Many Requests`

---

## Test 5: Batch Performance

Test that batching is significantly faster than individual requests:

**Create benchmark file:**
```bash
cd docs/integrations/javascript
cat > benchmark.js << 'EOF'
const DevSmithLogger = require('./logger');

const NUM_LOGS = 1000;

// Test 1: Batched (using logger)
console.time('Batched');
const logger = new DevSmithLogger({
  apiKey: process.env.DEVSMITH_API_KEY,
  apiUrl: process.env.DEVSMITH_API_URL || 'http://localhost:3000',
  projectSlug: 'test-project',
  serviceName: 'benchmark-batch',
  batchSize: 100
});

for (let i = 0; i < NUM_LOGS; i++) {
  logger.info(`Log ${i}`, { iteration: i });
}

setTimeout(() => {
  console.timeEnd('Batched');
  console.log(`Sent ${NUM_LOGS} logs in batches of 100`);
  process.exit(0);
}, 10000);
EOF
```

**Run benchmark:**
```bash
export DEVSMITH_API_KEY="dsk_abc123xyz789..."
node benchmark.js
```

**Expected output:**
```
Batched: ~500ms
Sent 1000 logs in batches of 100
```

**Compare with individual requests** (don't actually run this - would take ~30 seconds):
```
Individual: ~30,000ms (30 seconds)
Batched:    ~500ms

Speedup: 60x faster!
```

---

## Cleanup

After testing, remove test data:

```bash
docker exec -it devsmith-modular-platform-postgres-1 psql -U devsmith -d devsmith << EOF
DELETE FROM logs.entries WHERE service IN ('nodejs-test', 'python-test', 'go-test', 'benchmark-batch');
DELETE FROM logs.projects WHERE slug = 'test-project';
EOF
```

---

## Success Criteria

- ✅ All 3 sample loggers (JS, Python, Go) successfully send logs
- ✅ Logs appear in database with correct project_id, service, level, message
- ✅ Invalid API key returns 401 error
- ✅ Network errors are handled gracefully (retry logic works)
- ✅ Batch performance is 50x+ faster than individual requests
- ✅ No memory leaks or crashes during 1000-log benchmark

---

## Next Steps

After successful testing:
1. Update `CROSS_REPO_LOGGING_ARCHITECTURE.md` with test results
2. Create PR for Week 1 + Week 2 implementation
3. Begin Week 3: Project Management UI
