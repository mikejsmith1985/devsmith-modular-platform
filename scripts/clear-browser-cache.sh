#!/bin/bash
# Script to force clear browser cache after frontend rebuild
# This addresses the persistent 404 issue with hashed asset filenames

echo "🔄 Clearing browser cache for DevSmith Platform..."

# Get the current JavaScript bundle hash
BUNDLE_HASH=$(docker exec devsmith-frontend find /usr/share/nginx/html/assets -name "index-*.js" -type f | head -1 | sed 's/.*index-\(.*\)\.js/\1/')

if [ -n "$BUNDLE_HASH" ]; then
    echo "✅ Current bundle hash: $BUNDLE_HASH"
else
    echo "⚠️  Could not determine bundle hash"
fi

# Verify cache headers are correct
echo ""
echo "📋 Verifying cache headers..."
CACHE_HEADER=$(curl -s -I http://localhost:3000/ 2>/dev/null | grep -i "cache-control")
if echo "$CACHE_HEADER" | grep -q "no-store"; then
    echo "✅ Cache-Control headers are correct: $CACHE_HEADER"
else
    echo "❌ Cache-Control headers missing or incorrect!"
    echo "   Found: $CACHE_HEADER"
fi

echo ""
echo "🌐 To clear your browser cache:"
echo "   Chrome/Edge: Ctrl+Shift+Delete → Clear cache"
echo "   Firefox:     Ctrl+Shift+Delete → Clear cache"
echo "   Or: Hard refresh with Ctrl+Shift+R (Linux) or Cmd+Shift+R (Mac)"
echo ""
echo "💡 After clearing cache, visit: http://localhost:3000/"
echo "   The new bundle (index-${BUNDLE_HASH}.js) should load"
