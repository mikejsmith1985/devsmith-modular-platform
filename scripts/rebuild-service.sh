#!/bin/bash
# Mandatory rebuild script - GUARANTEES fresh deployment
# Usage: ./scripts/rebuild-service.sh <service-name>
# Example: ./scripts/rebuild-service.sh logs

set -e  # Exit on any error

SERVICE=$1

if [ -z "$SERVICE" ]; then
    echo "❌ ERROR: Service name required"
    echo "Usage: $0 <service-name>"
    echo "Available services: portal, review, logs, analytics"
    exit 1
fi

# Validate service name
VALID_SERVICES="portal review logs analytics"
if ! echo "$VALID_SERVICES" | grep -wq "$SERVICE"; then
    echo "❌ ERROR: Invalid service: $SERVICE"
    echo "Valid services: $VALID_SERVICES"
    exit 1
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔨 MANDATORY REBUILD: $SERVICE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Step 1: Stop container
echo "⏹️  Step 1/6: Stopping container..."
docker-compose stop "$SERVICE" 2>&1 | head -5

# Step 2: Remove container
echo "🗑️  Step 2/6: Removing container..."
docker-compose rm -f "$SERVICE" 2>&1 | head -5

# Step 3: Remove image (CRITICAL - guarantees fresh build)
echo "🗑️  Step 3/6: Removing image..."
IMAGE_NAME="devsmith-modular-platform-${SERVICE}"
if docker images -q "$IMAGE_NAME" 2>/dev/null | grep -q .; then
    docker rmi -f "$IMAGE_NAME" 2>&1 | head -5
    echo "   ✅ Image removed"
else
    echo "   ⚠️  No existing image found (fresh build)"
fi

# Step 4: Build with cache-busting timestamp
echo "🔨 Step 4/6: Building with --no-cache and BUILD_TIMESTAMP..."
BUILD_START=$(date +%s)
export BUILD_TIMESTAMP=$BUILD_START
docker-compose build --no-cache "$SERVICE" 2>&1 | tail -20
BUILD_END=$(date +%s)
BUILD_DURATION=$((BUILD_END - BUILD_START))
echo "   ✅ Build completed in ${BUILD_DURATION}s"

# Step 5: Start container
echo "🚀 Step 5/6: Starting container..."
docker-compose up -d "$SERVICE"
CONTAINER_START=$(date +%s)

# Wait for container to be ready
echo "⏳ Waiting for container to start..."
sleep 3

# Step 6: Verify deployment
echo "✅ Step 6/6: Verifying deployment..."
./scripts/verify-deployment.sh "$SERVICE"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ REBUILD COMPLETE: $SERVICE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Summary:"
echo "   - Build time: ${BUILD_DURATION}s"
echo "   - Build timestamp: $BUILD_TIMESTAMP"
echo "   - Container start: $(date -d @$CONTAINER_START '+%Y-%m-%d %H:%M:%S')"
echo ""
echo "🎯 Next steps:"
echo "   1. Test the service manually"
echo "   2. Run regression tests: bash scripts/regression-test.sh"
echo "   3. Check container logs: docker-compose logs --tail=50 $SERVICE"
echo ""
