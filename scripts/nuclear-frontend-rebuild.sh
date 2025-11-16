#!/bin/bash
# Nuclear Frontend Rebuild - 4-Layer Cache Invalidation Strategy
# Solves persistent Vite/Rollup cache issues by clearing ALL possible cache locations

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
FRONTEND_DIR="$PROJECT_ROOT/frontend"
PORTAL_STATIC="$PROJECT_ROOT/apps/portal/static"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔥 NUCLEAR FRONTEND REBUILD - 4-Layer Cache Invalidation"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd "$FRONTEND_DIR"

# LAYER 1: Vite's Build Cache (.vite directory)
echo "🧹 LAYER 1: Clearing Vite build cache (.vite)..."
rm -rf .vite
echo "   ✓ Removed .vite/"

# LAYER 2: Node Modules Cache (transforms, esbuild cache)
echo "🧹 LAYER 2: Clearing node_modules cache..."
rm -rf node_modules/.vite
rm -rf node_modules/.cache
find . -name "*.cache" -type d -exec rm -rf {} + 2>/dev/null || true
echo "   ✓ Removed node_modules caches"

# LAYER 3: Dist Output (old bundles)
echo "🧹 LAYER 3: Clearing dist output..."
rm -rf dist
echo "   ✓ Removed dist/"

# LAYER 4: Deployed Static Files (Docker/Portal)
echo "🧹 LAYER 4: Clearing deployed static files..."
rm -rf "$PORTAL_STATIC/assets"/*.js 2>/dev/null || true
rm -rf "$PORTAL_STATIC/assets"/*.css 2>/dev/null || true
rm -rf "$PORTAL_STATIC/index.html" 2>/dev/null || true
echo "   ✓ Purged old static files from portal"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔨 Building Frontend from Scratch..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Build with explicit cache busting
npm run build

# Verify build succeeded
if [ ! -d "dist" ] || [ ! -f "dist/index.html" ]; then
  echo "❌ ERROR: Build failed - dist directory not created"
  exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📦 Deploying to Portal Static Directory..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Copy with verification
cp -r dist/* "$PORTAL_STATIC/"

# Verify deployment
NEW_BUNDLE=$(ls -t "$PORTAL_STATIC/assets"/index-*.js 2>/dev/null | head -1)
if [ -n "$NEW_BUNDLE" ]; then
  BUNDLE_HASH=$(basename "$NEW_BUNDLE" | sed 's/index-\(.*\)\.js/\1/')
  BUNDLE_SIZE=$(du -h "$NEW_BUNDLE" | cut -f1)
  echo "   ✓ Deployed bundle: index-$BUNDLE_HASH.js ($BUNDLE_SIZE)"
  
  # Check for new variable name
  if strings "$NEW_BUNDLE" | grep -q "unfilteredStats"; then
    echo "   ✅ Bundle contains 'unfilteredStats' - FIX CONFIRMED!"
  else
    echo "   ⚠️  WARNING: Bundle does NOT contain 'unfilteredStats'"
    echo "   Checking for old 'stats' references..."
    if strings "$NEW_BUNDLE" | grep -E "stats\s*is\s*not\s*defined" >/dev/null 2>&1; then
      echo "   ❌ ERROR: Bundle still contains old code!"
      exit 1
    fi
  fi
else
  echo "   ❌ ERROR: No bundle found after deployment"
  exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🚀 Restarting Portal Container..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd "$PROJECT_ROOT"
docker-compose up -d --build --force-recreate portal

# Wait for portal to be healthy
echo "⏳ Waiting for portal to be healthy..."
sleep 3

if docker-compose ps portal | grep -q "Up"; then
  echo "   ✅ Portal is running"
else
  echo "   ❌ ERROR: Portal failed to start"
  docker-compose logs portal --tail=20
  exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ NUCLEAR REBUILD COMPLETE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📋 Next Steps:"
echo "   1. Open: http://localhost:3000/health"
echo "   2. Check browser console for errors (should be NONE)"
echo "   3. Apply ERROR filter - stats should STAY at total counts"
echo "   4. Run: npx playwright test tests/e2e/health/stats-filtering-visual.spec.ts --reporter=list"
echo ""
echo "🔍 Bundle Hash: $BUNDLE_HASH"
echo "📍 Location: $NEW_BUNDLE"
echo ""
