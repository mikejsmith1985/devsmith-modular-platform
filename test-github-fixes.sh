#!/bin/bash
# Test GitHub Integration Fixes
# Tests Quick Scan, Full Browser tree, and file fetching

set -e

REVIEW_URL="http://localhost:3000/api/review"
TEST_REPO="github.com/mikejsmith1985/devsmith-modular-platform"
TEST_BRANCH="development"

echo "🧪 Testing GitHub Integration Fixes"
echo "======================================"
echo ""

# Get session token (assuming authenticated)
# For now, test without auth to check endpoint availability

echo "1️⃣  Testing Quick Scan (should find core files)"
echo "   Fetching: $TEST_REPO @ $TEST_BRANCH"
QUICK_SCAN_URL="$REVIEW_URL/github/quick-scan?url=$TEST_REPO&branch=$TEST_BRANCH"
echo "   URL: $QUICK_SCAN_URL"

# Note: This will fail with 401 if not authenticated, but that's expected
# The real test is whether backend is receiving the request properly
curl -s -w "\n   HTTP Status: %{http_code}\n" "$QUICK_SCAN_URL" | head -20

echo ""
echo "2️⃣  Testing Full Browser Tree"
TREE_URL="$REVIEW_URL/github/tree?url=$TEST_REPO&branch=$TEST_BRANCH"
echo "   URL: $TREE_URL"
curl -s -w "\n   HTTP Status: %{http_code}\n" "$TREE_URL" | head -20

echo ""
echo "3️⃣  Testing File Fetch (normalized path)"
FILE_URL="$REVIEW_URL/github/file?url=$TEST_REPO&branch=$TEST_BRANCH&path=README.md"
echo "   URL: $FILE_URL"
echo "   Path: README.md (no leading slash)"
curl -s -w "\n   HTTP Status: %{http_code}\n" "$FILE_URL" | head -20

echo ""
echo "4️⃣  Testing File Fetch (with leading slash - should normalize)"
FILE_URL_SLASH="$REVIEW_URL/github/file?url=$TEST_REPO&branch=$TEST_BRANCH&path=/README.md"
echo "   URL: $FILE_URL_SLASH"
echo "   Path: /README.md (with leading slash)"
curl -s -w "\n   HTTP Status: %{http_code}\n" "$FILE_URL_SLASH" | head -20

echo ""
echo "✅ Endpoint tests complete!"
echo ""
echo "Note: If you see 401 errors, you need to authenticate first."
echo "The important thing is that endpoints are responding and not returning 404."
echo ""
echo "To test with authentication:"
echo "1. Go to http://localhost:3000"
echo "2. Login with GitHub"
echo "3. Open Review app"
echo "4. Try importing the test repo with Quick Scan and Full Browser"
