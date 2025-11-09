#!/bin/bash

# OAuth Flow Tracer - Follows exact user path
# Tests the ACTUAL frontend flow, not backend endpoints

echo "==================================="
echo "OAUTH FLOW TRACE - CLIENT-SIDE PKCE"
echo "==================================="
echo ""

# Check if portal is running
echo "Step 1: Verify portal is running..."
PORTAL_CHECK=$(curl -s http://localhost:3000 2>&1 | grep -c "DevSmith" || true)
if [ "$PORTAL_CHECK" -eq 0 ]; then
    echo "❌ ERROR: Portal not responding at localhost:3000"
    exit 1
fi
echo "✅ Portal is running"
echo ""

# Enable debug logging in backend
echo "Step 2: Checking backend logs for recent OAuth activity..."
echo "Recent /auth/github/callback requests:"
docker-compose logs portal --tail=50 2>&1 | grep -E "callback|OAUTH_STATE_INVALID|401" | tail -5
echo ""

# Check what frontend is actually doing
echo "Step 3: Checking frontend OAuth implementation..."
echo "Frontend LoginPage.jsx uses:"
grep -A5 "handleGitHubLogin" /home/mikej/projects/DevSmith-Modular-Platform/frontend/src/components/LoginPage.jsx | head -10
echo ""

# Check what backend endpoints exist
echo "Step 4: Backend OAuth endpoints available:"
echo "  - POST /api/portal/auth/token (HandleTokenExchange) ← Frontend uses this"
echo "  - GET /auth/github/login (HandleGitHubOAuthLogin) ← OLD, generates server state"
echo "  - GET /auth/github/callback (HandleGitHubOAuthCallbackWithSession) ← OLD, validates server state"
echo ""

# The actual problem
echo "==================================="
echo "DIAGNOSIS"
echo "==================================="
echo ""
echo "The system has TWO OAuth implementations:"
echo ""
echo "1. CLIENT-SIDE PKCE (Modern, what frontend uses):"
echo "   - Frontend generates state + code_verifier"
echo "   - Stores in sessionStorage"
echo "   - Redirects directly to GitHub"
echo "   - GitHub callbacks to React route: /auth/github/callback"
echo "   - React validates state from sessionStorage"
echo "   - Calls backend: POST /api/portal/auth/token"
echo "   - Backend exchanges code with code_verifier"
echo ""
echo "2. SERVER-SIDE STATE (Old, not used by frontend):"
echo "   - Backend GET /auth/github/login generates state"
echo "   - Stores in Redis"
echo "   - Redirects to GitHub with server state"
echo "   - GitHub callbacks to backend: GET /auth/github/callback"
echo "   - Backend validates state from Redis"
echo ""
echo "==================================="
echo "THE BUG"
echo "==================================="
echo ""
echo "If user is seeing 'Invalid OAuth state parameter':"
echo ""
echo "POSSIBILITY 1: React routing not working"
echo "  - User visits localhost:3000/login"
echo "  - Clicks GitHub button"
echo "  - GitHub redirects to /auth/github/callback"
echo "  - But React router doesn't catch it"
echo "  - Falls through to backend route"
echo "  - Backend expects server-side state (Redis)"
echo "  - But GitHub sent client-side state (sessionStorage)"
echo "  - ❌ State mismatch → 401 Unauthorized"
echo ""
echo "POSSIBILITY 2: sessionStorage cleared/corrupted"
echo "  - Frontend stores state in sessionStorage"
echo "  - User's browser clears sessionStorage"
echo "  - OR state value corrupted"
echo "  - React can't validate state"
echo "  - Shows error"
echo ""
echo "POSSIBILITY 3: GitHub redirect URI misconfigured"
echo "  - GitHub OAuth app expects: http://localhost:3000/auth/github/callback"
echo "  - But frontend sets: window.location.origin + '/auth/github/callback'"
echo "  - If origin is different, callback fails"
echo ""
echo "==================================="
echo "NEXT STEPS TO DEBUG"
echo "==================================="
echo ""
echo "1. Check browser console during login attempt"
echo "   - Should see: [PKCE] Redirecting to GitHub with code_challenge"
echo "   - Should see: [OAuthCallback] Component mounted"
echo "   - Should see: [PKCE] State mismatch OR [PKCE] Token exchange..."
echo ""
echo "2. Check if React router handles /auth/github/callback"
echo "   - Visit: http://localhost:3000/auth/github/callback?code=test&state=test"
echo "   - Should see React component, not backend 401"
echo ""
echo "3. Check GitHub OAuth app configuration"
echo "   - Authorization callback URL must be: http://localhost:3000/auth/github/callback"
echo ""
echo "4. Watch backend logs during real login"
echo "   - Run: docker-compose logs -f portal | grep -E 'PKCE|token|callback'"
echo ""

# Test if React routes work
echo "Testing React routing..."
TEST_ROUTE=$(curl -sI http://localhost:3000/auth/github/callback?code=test&state=test | head -1)
echo "GET /auth/github/callback returns: $TEST_ROUTE"
if echo "$TEST_ROUTE" | grep -q "401"; then
    echo "⚠️  ISSUE: Backend handling /auth/github/callback instead of React!"
    echo "This means React router isn't catching the OAuth callback."
    echo "User sees backend's 401 instead of React's OAuthCallback component."
elif echo "$TEST_ROUTE" | grep -q "200"; then
    echo "✅ React router handling /auth/github/callback correctly"
else
    echo "❓ Unexpected response: $TEST_ROUTE"
fi
echo ""

echo "==================================="
echo "MANUAL TEST INSTRUCTIONS"
echo "==================================="
echo ""
echo "1. Open browser DevTools (F12)"
echo "2. Go to Console tab"
echo "3. Visit: http://localhost:3000/login"
echo "4. Click 'Login with GitHub' button"
echo "5. Watch console for:"
echo "   - [PKCE] messages (state generation)"
echo "   - [OAuthCallback] messages (callback handling)"
echo "   - Any errors"
echo "6. Check Application > Session Storage:"
echo "   - oauth_state should exist"
echo "   - pkce_code_verifier should exist"
echo "7. Check Network tab:"
echo "   - POST to /api/portal/auth/token"
echo "   - OR GET to /auth/github/callback (if this happens, routing is broken)"
echo ""
