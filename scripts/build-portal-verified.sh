#!/bin/bash
# Build Portal service with MANDATORY verification
# This script FAILS if the container doesn't contain what it should

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "VERIFIED PORTAL BUILD"
echo "=========================================="
echo ""

# Step 1: Get version info
echo "📋 Step 1: Collecting build metadata..."
VERSION=$(cat VERSION 2>/dev/null || echo "dev")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_TIMESTAMP=$(date +%s)

echo "  VERSION: $VERSION"
echo "  GIT_COMMIT: $GIT_COMMIT"
echo "  BUILD_TIME: $BUILD_TIME"
echo ""

# Step 2: Verify frontend assets exist BEFORE building
echo "📋 Step 2: Pre-build verification - checking host filesystem..."
REQUIRED_FILES=(
    "apps/portal/static/index.html"
    "apps/portal/static/assets/index-*.js"
    "apps/portal/static/assets/index-*.css"
)

for pattern in "${REQUIRED_FILES[@]}"; do
    files=$(ls $pattern 2>/dev/null || echo "")
    if [ -z "$files" ]; then
        echo -e "${RED}❌ FAILED: Required file missing: $pattern${NC}"
        echo ""
        echo "Frontend assets not found. Did you run:"
        echo "  cd frontend && npm run build"
        echo "  cp -r frontend/dist/* apps/portal/static/"
        exit 1
    fi
    echo -e "${GREEN}✓${NC} Found: $pattern"
done

# Count assets for later verification
HOST_ASSET_COUNT=$(find apps/portal/static/assets -type f 2>/dev/null | wc -l)
echo ""
echo "  Host has $HOST_ASSET_COUNT asset files"
echo ""

# Step 3: Build Docker image
echo "📋 Step 3: Building Docker image..."
docker-compose build --no-cache \
    --build-arg VERSION="$VERSION" \
    --build-arg GIT_COMMIT="$GIT_COMMIT" \
    --build-arg BUILD_TIME="$BUILD_TIME" \
    --build-arg BUILD_TIMESTAMP="$BUILD_TIMESTAMP" \
    portal

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Docker build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker build completed${NC}"
echo ""

# Step 4: CRITICAL - Verify container contents BEFORE starting it
echo "📋 Step 4: Post-build verification - checking container image..."
echo "  Creating temporary container to inspect contents..."

TEMP_CONTAINER="portal-verify-$$"
docker create --name "$TEMP_CONTAINER" devsmith-modular-platform-portal:latest >/dev/null 2>&1

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Failed to create verification container${NC}"
    exit 1
fi

# Verify critical files exist in container
echo "  Checking for required files in container..."
CONTAINER_CHECKS=(
    "/app/static/index.html:index.html"
    "/app/static/assets/:assets directory"
)

VERIFICATION_FAILED=0
for check in "${CONTAINER_CHECKS[@]}"; do
    path="${check%%:*}"
    desc="${check##*:}"
    
    if docker exec "$TEMP_CONTAINER" test -e "$path" 2>/dev/null; then
        echo -e "  ${GREEN}✓${NC} Found: $desc"
    else
        echo -e "  ${RED}✗${NC} Missing: $desc (path: $path)"
        VERIFICATION_FAILED=1
    fi
done

# Count assets in container
CONTAINER_ASSET_COUNT=$(docker exec "$TEMP_CONTAINER" find /app/static/assets -type f 2>/dev/null | wc -l || echo "0")
echo ""
echo "  Container has $CONTAINER_ASSET_COUNT asset files"

if [ "$CONTAINER_ASSET_COUNT" -eq 0 ]; then
    echo -e "${RED}❌ CRITICAL: No asset files found in container!${NC}"
    VERIFICATION_FAILED=1
elif [ "$CONTAINER_ASSET_COUNT" -lt "$HOST_ASSET_COUNT" ]; then
    echo -e "${YELLOW}⚠ WARNING: Container has fewer assets than host ($CONTAINER_ASSET_COUNT vs $HOST_ASSET_COUNT)${NC}"
fi

# Verify version endpoint will work
echo ""
echo "  Checking version info in container..."
docker exec "$TEMP_CONTAINER" grep -q "$GIT_COMMIT" /app/main 2>/dev/null
if [ $? -eq 0 ]; then
    echo -e "  ${GREEN}✓${NC} Version info embedded correctly"
else
    echo -e "  ${YELLOW}⚠${NC} Could not verify version info in binary"
fi

# Cleanup temp container
docker rm "$TEMP_CONTAINER" >/dev/null 2>&1

echo ""
if [ $VERIFICATION_FAILED -eq 1 ]; then
    echo -e "${RED}=========================================="
    echo "❌ VERIFICATION FAILED"
    echo "==========================================${NC}"
    echo ""
    echo "The container image was built but does NOT contain required files."
    echo "This would cause the same problem you just experienced."
    echo ""
    echo "Common causes:"
    echo "  1. Docker build context issue"
    echo "  2. .dockerignore excluding files"
    echo "  3. Files copied after build started"
    echo ""
    echo "The image has been built but NOT started."
    echo "Manual investigation required."
    exit 1
fi

# Step 5: Start container
echo "📋 Step 5: Starting Portal container..."
docker-compose up -d portal

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Failed to start container${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Container started${NC}"
echo ""

# Step 6: Wait for health check
echo "📋 Step 6: Waiting for health check..."
for i in {1..30}; do
    if curl -sf http://localhost:3000/api/portal/health >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Health check passed${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}❌ Health check timeout${NC}"
        echo "Container logs:"
        docker-compose logs --tail=20 portal
        exit 1
    fi
    echo -n "."
    sleep 1
done
echo ""

# Step 7: Verify version endpoint
echo "📋 Step 7: Verifying version endpoint..."
VERSION_RESPONSE=$(curl -sf http://localhost:3000/api/portal/version 2>/dev/null || echo "")
if [ -z "$VERSION_RESPONSE" ]; then
    echo -e "${RED}❌ Version endpoint not responding${NC}"
    exit 1
fi

REPORTED_VERSION=$(echo "$VERSION_RESPONSE" | grep -o '"version":"[^"]*"' | cut -d'"' -f4 || echo "")
REPORTED_COMMIT=$(echo "$VERSION_RESPONSE" | grep -o '"commit":"[^"]*"' | cut -d'"' -f4 || echo "")

if [ "$REPORTED_COMMIT" = "$GIT_COMMIT" ]; then
    echo -e "${GREEN}✓ Version endpoint reports correct commit: $GIT_COMMIT${NC}"
else
    echo -e "${RED}❌ Version mismatch!${NC}"
    echo "  Expected: $GIT_COMMIT"
    echo "  Got: $REPORTED_COMMIT"
    exit 1
fi
echo ""

# Step 8: Verify frontend assets are served
echo "📋 Step 8: Verifying frontend assets are served..."
HTTP_CODE=$(curl -sf -o /dev/null -w "%{http_code}" http://localhost:3000/ 2>/dev/null || echo "000")
if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Root path returns 200${NC}"
else
    echo -e "${RED}❌ Root path returns $HTTP_CODE${NC}"
    exit 1
fi

# Try to fetch a JS asset
JS_FILE=$(ls apps/portal/static/assets/index-*.js 2>/dev/null | head -1 | xargs basename || echo "")
if [ -n "$JS_FILE" ]; then
    HTTP_CODE=$(curl -sf -o /dev/null -w "%{http_code}" "http://localhost:3000/assets/$JS_FILE" 2>/dev/null || echo "000")
    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓ JavaScript asset served correctly${NC}"
    else
        echo -e "${RED}❌ JavaScript asset returns $HTTP_CODE${NC}"
        exit 1
    fi
fi
echo ""

# Success!
echo -e "${GREEN}=========================================="
echo "✅ BUILD VERIFIED AND DEPLOYED"
echo "==========================================${NC}"
echo ""
echo "Portal version: $VERSION-$GIT_COMMIT"
echo "Container assets: $CONTAINER_ASSET_COUNT files"
echo "Health: ✓ Passing"
echo "Version endpoint: ✓ Working"
echo "Frontend: ✓ Serving"
echo ""
echo "🔍 Test the app:"
echo "   http://localhost:3000/"
echo ""
echo "🔍 Check version:"
echo "   curl http://localhost:3000/api/portal/version | jq"
echo ""
