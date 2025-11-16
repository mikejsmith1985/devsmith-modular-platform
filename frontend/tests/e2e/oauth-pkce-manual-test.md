# OAuth PKCE Manual Test Plan

## Test 1: PKCE Parameters Generation

1. Open browser DevTools (F12)
2. Navigate to `http://localhost:3000/`
3. Click "Login with GitHub" button
4. **Expected**: Browser redirects to GitHub OAuth page
5. **Verify** in the URL bar:
   - `code_challenge` parameter is present (43 chars, base64URL)
   - `code_challenge_method=S256`
   - `state` parameter is present (random string)
   - `client_id=YOUR_GITHUB_CLIENT_ID` (from GitHub OAuth App settings)

## Test 2: SessionStorage

Before clicking login:
```javascript
// In browser console:
sessionStorage.getItem('pkce_code_verifier') // Should be null
sessionStorage.getItem('oauth_state') // Should be null
```

After clicking login (before GitHub loads):
```javascript
sessionStorage.getItem('pkce_code_verifier') // Should be 43-char string
sessionStorage.getItem('oauth_state') // Should match state in URL
```

## Test 3: Token Exchange Endpoint

Test with curl:
```bash
curl -X POST http://localhost:3000/api/portal/auth/token \
  -H "Content-Type: application/json" \
  -d '{
    "code": "test",
    "state": "test",
    "code_verifier": "test123456789012345678901234567890123"
  }'
```

**Expected**: `{"error":"Failed to authenticate"}` (401 status)
**Means**: Endpoint is accessible and validating requests

## Test 4: Complete Flow (Manual)

1. Click "Login with GitHub" on `http://localhost:3000/`
2. Authorize the DevSmith app on GitHub
3. GitHub redirects to `/auth/github/callback?code=XXX&state=YYY`
4. Frontend OAuthCallback component:
   - Validates state matches sessionStorage
   - Retrieves code_verifier from sessionStorage
   - Calls `/api/portal/auth/token` with code, state, code_verifier
   - Stores returned JWT in localStorage
   - Redirects to dashboard

5. **Expected final state**:
   - `localStorage.getItem('devsmith_token')` contains JWT
   - User is on dashboard page
   - sessionStorage PKCE data is cleared

## Verification Commands

```bash
# 1. Check OAuth login route
curl -s -D - http://localhost:3000/api/portal/auth/github/login -o /dev/null | grep Location
# Expected: Location: https://github.com/login/oauth/authorize?client_id=...

# 2. Check token endpoint
curl -X POST http://localhost:3000/api/portal/auth/token \
  -H "Content-Type: application/json" \
  -d '{"code":"test","state":"test","code_verifier":"test"}'
# Expected: {"error":"Failed to authenticate"}

# 3. Check Traefik routing
curl -s http://localhost:8090/api/http/routers | jq '.[] | select(.name | contains("portal"))'
# Expected: portal-api and portal-auth routers

# 4. Check portal service health
docker-compose exec -T portal curl -f http://localhost:3001/health
# Expected: {"service":"portal","status":"healthy"}
```

## Success Criteria

✅ OAuth login redirects to GitHub with PKCE parameters  
✅ code_challenge is SHA-256 hash of code_verifier  
✅ code_challenge_method is S256  
✅ state parameter prevents CSRF attacks  
✅ code_verifier stored in sessionStorage  
✅ Token exchange endpoint accessible  
✅ Full OAuth flow completes successfully

