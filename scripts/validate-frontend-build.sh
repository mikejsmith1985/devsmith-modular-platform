#!/bin/bash
# Validate Frontend Build Script
# Prevents deployment of builds with incorrect API_URL
# 
# This script is called:
# - Before docker-compose build portal (in Makefile)
# - In pre-commit hook (optional)
# - In CI/CD pipeline (GitHub Actions)

set -e

FRONTEND_DIR="/home/mikej/projects/DevSmith-Modular-Platform/frontend"
EXPECTED_API_URL="http://localhost:3000"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Frontend Build Validation"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check 1: Verify .env.production exists and has correct value
echo ""
echo "✓ Check 1: Validate .env.production file"
if [ ! -f "$PROJECT_ROOT/frontend/.env.production" ]; then
    echo "❌ ERROR: frontend/.env.production not found!"
    exit 1
fi

PROD_API_URL=$(grep "^VITE_API_URL=" "$PROJECT_ROOT/frontend/.env.production" | cut -d'=' -f2)
if [ "$PROD_API_URL" != "$EXPECTED_API_URL" ]; then
    echo "❌ ERROR: .env.production has wrong VITE_API_URL!"
    echo "   Expected: $EXPECTED_API_URL"
    echo "   Got:      $PROD_API_URL"
    echo ""
    echo "   This will cause double /api/api in URLs!"
    echo "   Fix: Update frontend/.env.production line 17"
    exit 1
fi
echo "   ✓ VITE_API_URL=$PROD_API_URL (correct)"

# Check 2: Verify dist directory exists
echo ""
echo "✓ Check 2: Validate dist directory exists"
if [ ! -d "$PROJECT_ROOT/frontend/dist" ]; then
    echo "❌ ERROR: frontend/dist/ not found!"
    echo "   Run: cd frontend && npm run build"
    exit 1
fi
echo "   ✓ dist/ directory exists"

# Check 3: Verify built JavaScript has correct API_URL
echo ""
echo "✓ Check 3: Validate built JavaScript bundle"
BUILT_JS=$(ls "$PROJECT_ROOT/frontend/dist/assets/index-"*.js 2>/dev/null | head -1)
if [ -z "$BUILT_JS" ]; then
    echo "❌ ERROR: No JavaScript bundle found in dist/assets/"
    exit 1
fi

# Extract API_URL from minified JavaScript
# Looking for pattern: u="http://localhost:3000"
if strings "$BUILT_JS" | grep -q 'u="http://localhost:3000"'; then
    echo "   ✓ Built bundle has correct API_URL (http://localhost:3000)"
elif strings "$BUILT_JS" | grep -q 'u="/api"'; then
    echo "❌ ERROR: Built bundle has WRONG API_URL!"
    echo "   Found: u=\"/api\""
    echo "   Expected: u=\"http://localhost:3000\""
    echo ""
    echo "   This will cause double /api/api in URLs (404 errors)!"
    echo ""
    echo "   Root cause: .env.production has VITE_API_URL=/api"
    echo "   Fix: Update frontend/.env.production and rebuild:"
    echo "        cd frontend && npm run build"
    exit 1
else
    echo "⚠️  WARNING: Could not find API_URL in built bundle"
    echo "   Bundle: $BUILT_JS"
    echo "   Manual verification recommended"
fi

# Check 4: Verify index.html exists
echo ""
echo "✓ Check 4: Validate index.html exists"
if [ ! -f "$PROJECT_ROOT/frontend/dist/index.html" ]; then
    echo "❌ ERROR: frontend/dist/index.html not found!"
    exit 1
fi
echo "   ✓ index.html exists"

# Check 5: Verify apps/portal/static has been updated
echo ""
echo "✓ Check 5: Validate portal static files"
if [ ! -f "$PROJECT_ROOT/apps/portal/static/index.html" ]; then
    echo "⚠️  WARNING: apps/portal/static/index.html not found"
    echo "   You may need to copy dist files:"
    echo "   cp -r frontend/dist/* apps/portal/static/"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ All validation checks passed!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
