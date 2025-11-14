#!/bin/bash
# Post-deployment verification script
# Usage: ./scripts/verify-deployment.sh <service-name>
# Example: ./scripts/verify-deployment.sh logs

set -e

SERVICE=$1

if [ -z "$SERVICE" ]; then
    echo "❌ ERROR: Service name required"
    echo "Usage: $0 <service-name>"
    exit 1
fi

# Service port mapping
declare -A SERVICE_PORTS
SERVICE_PORTS=(
    ["portal"]=3001
    ["review"]=8081
    ["logs"]=8082
    ["analytics"]=8083
)

PORT=${SERVICE_PORTS[$SERVICE]}
if [ -z "$PORT" ]; then
    echo "❌ ERROR: Unknown service: $SERVICE"
    exit 1
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔍 VERIFYING DEPLOYMENT: $SERVICE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check 1: Container is running
echo "1️⃣  Checking container status..."
CONTAINER_NAME="devsmith-modular-platform-${SERVICE}-1"
if ! docker ps --filter "name=$CONTAINER_NAME" --filter "status=running" | grep -q "$CONTAINER_NAME"; then
    echo "   ❌ FAIL: Container not running"
    docker ps -a --filter "name=$CONTAINER_NAME"
    exit 1
fi
echo "   ✅ PASS: Container is running"

# Check 2: Container age (must be < 2 minutes to be fresh)
echo "2️⃣  Checking container age..."
CONTAINER_START=$(docker inspect -f '{{.State.StartedAt}}' "$CONTAINER_NAME" 2>/dev/null)
if [ -z "$CONTAINER_START" ]; then
    echo "   ❌ FAIL: Cannot determine container start time"
    exit 1
fi

# Convert to epoch seconds
CONTAINER_EPOCH=$(date -d "$CONTAINER_START" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%S" "$CONTAINER_START" +%s 2>/dev/null)
CURRENT_EPOCH=$(date +%s)
CONTAINER_AGE=$((CURRENT_EPOCH - CONTAINER_EPOCH))

if [ "$CONTAINER_AGE" -gt 120 ]; then
    echo "   ❌ FAIL: Container is ${CONTAINER_AGE}s old (expected < 120s)"
    echo "   This suggests container was not rebuilt - old code may be running"
    exit 1
fi
echo "   ✅ PASS: Container is ${CONTAINER_AGE}s old (fresh deployment)"

# Check 3: Health endpoint responds
echo "3️⃣  Checking health endpoint..."
MAX_RETRIES=10
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -sf "http://localhost:${PORT}/health" > /dev/null 2>&1; then
        echo "   ✅ PASS: Health endpoint responding"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
        echo "   ❌ FAIL: Health endpoint not responding after ${MAX_RETRIES} attempts"
        echo "   Showing recent logs:"
        docker-compose logs --tail=20 "$SERVICE"
        exit 1
    fi
    echo "   ⏳ Attempt $RETRY_COUNT/$MAX_RETRIES - waiting 2s..."
    sleep 2
done

# Check 4: No recent errors in logs
echo "4️⃣  Checking for errors in recent logs..."
ERROR_COUNT=$(docker-compose logs --tail=50 "$SERVICE" 2>&1 | grep -i "error\|panic\|fatal" | grep -v "no error" | wc -l || echo "0")
if [ "$ERROR_COUNT" -gt 0 ]; then
    echo "   ⚠️  WARNING: Found $ERROR_COUNT error(s) in recent logs:"
    docker-compose logs --tail=50 "$SERVICE" 2>&1 | grep -i "error\|panic\|fatal" | grep -v "no error" | head -10
    echo ""
    echo "   ⚠️  Review logs manually: docker-compose logs --tail=100 $SERVICE"
    # Don't fail - errors might be expected during startup
else
    echo "   ✅ PASS: No errors in recent logs"
fi

# Check 5: Verify BUILD_TIMESTAMP was used (if available in logs)
echo "5️⃣  Checking if BUILD_TIMESTAMP was used..."
if docker-compose logs --tail=100 "$SERVICE" 2>&1 | grep -q "Build timestamp:"; then
    TIMESTAMP=$(docker-compose logs --tail=100 "$SERVICE" 2>&1 | grep "Build timestamp:" | tail -1)
    echo "   ✅ PASS: $TIMESTAMP"
else
    echo "   ⚠️  WARNING: BUILD_TIMESTAMP not found in logs"
    echo "   This is expected for older builds"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ VERIFICATION COMPLETE: $SERVICE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Deployment Summary:"
echo "   - Container: $CONTAINER_NAME"
echo "   - Age: ${CONTAINER_AGE}s (started at $(date -d "$CONTAINER_START" '+%H:%M:%S'))"
echo "   - Health: ✅ Responding on port $PORT"
echo "   - Errors: $ERROR_COUNT in recent logs"
echo ""
echo "✅ Deployment verified - service is running fresh code"
echo ""
