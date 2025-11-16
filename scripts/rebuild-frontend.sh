#!/bin/bash
# rebuild-frontend.sh - Rebuild frontend with NO CACHE to ensure fresh deployment
#
# This script:
# 1. Forces Docker to rebuild without using cached layers
# 2. Removes old container completely
# 3. Restarts Traefik to clear routing cache
# 4. Verifies the new code is actually deployed

set -e

echo "🔨 Rebuilding frontend with NO CACHE..."

# Step 1: Clean old build artifacts in frontend/dist
echo "  1️⃣  Cleaning old builds..."
rm -rf frontend/dist

# Step 2: Build frontend fresh
echo "  2️⃣  Building frontend..."
cd frontend && npm run build && cd ..

# Step 3: Stop and remove old container
echo "  3️⃣  Stopping old container..."
docker-compose down frontend

# Step 4: Force rebuild with no cache using dev-nocache override
echo "  4️⃣  Rebuilding Docker container (no cache)..."
export BUILD_TIMESTAMP=$(date +%s)
docker-compose -f docker-compose.yml -f docker-compose.dev-nocache.yml build frontend

# Step 5: Start the new container
echo "  5️⃣  Starting new container..."
docker-compose up -d frontend

# Give it a moment to start
sleep 3

# Restart Traefik to clear any routing cache
echo "  6️⃣  Restarting Traefik to clear routing cache..."
docker-compose restart traefik
sleep 2

# Verify deployment
echo ""
echo "✅ Verifying deployment..."
echo "Frontend container:"
docker-compose ps frontend

echo ""
echo "JavaScript bundle hash:"
DIRECT_HASH=$(curl -s http://localhost:5173/ | grep -o 'index-[^"]*\.js' | head -1)
GATEWAY_HASH=$(curl -s http://localhost:3000/ | grep -o 'index-[^"]*\.js' | head -1)

echo "  Direct (port 5173): $DIRECT_HASH"
echo "  Gateway (port 3000): $GATEWAY_HASH"

if [ "$DIRECT_HASH" = "$GATEWAY_HASH" ]; then
    echo "✅ SUCCESS! Both hashes match - deployment is fresh!"
else
    echo "❌ WARNING! Hash mismatch - gateway may be cached"
    echo "   Try: curl -H 'Cache-Control: no-cache' http://localhost:3000/"
fi

echo ""
echo "🌐 Frontend is available at:"
echo "   http://localhost:3000 (through gateway - ALWAYS USE THIS)"
echo "   http://localhost:5173 (direct - for debugging only)"
echo ""
echo "⚠️  IMPORTANT: If you still see old code in browser:"
echo "   1. Hard refresh: Ctrl+Shift+R (Windows/Linux) or Cmd+Shift+R (Mac)"
echo "   2. Or clear browser cache: Ctrl+Shift+Delete"
echo "   3. Or use incognito/private window"
echo "   4. Or add cache buster: http://localhost:3000/?cb=$(date +%s)"
echo ""
