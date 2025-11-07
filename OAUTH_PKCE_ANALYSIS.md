# OAuth 2.0 PKCE Implementation Analysis

**Date**: 2025-11-06  
**Status**: OAuth Currently Broken - Needs Immediate Fix  
**Current Branch**: feature/ui-fixes  
**Issue**: Routes registered as `/auth/*` but frontend calls `/api/portal/auth/*`

---

## 🚨 Quick Fix Required Before PKCE

### Current Problem
```bash
# Frontend calls:
fetch('/api/portal/auth/github/login')  # 404 Not Found

# Backend routes registered as:
router.GET("/auth/github/login", handler)  # Wrong path!

# Should be:
router.GET("/api/portal/auth/github/login", handler)
```

### Immediate Fix (5 minutes)

**File**: `cmd/portal/main.go`
```go
// After line ~60, register auth routes with correct prefix:
authGroup := router.Group("/api/portal/auth")
{
    authGroup.GET("/github/login", handlers.HandleGitHubOAuthLogin)
    authGroup.GET("/github/callback", handlers.HandleGitHubOAuthCallbackWithSession)
    authGroup.POST("/logout", handlers.HandleLogout)
    authGroup.GET("/me", handlers.HandleAuthMe)  // If exists
}
```

Then:
```bash
docker-compose up -d --build portal
curl -I http://localhost:3000/api/portal/auth/github/login
# Should return: 302 Found (redirect to GitHub)
```

---

## Executive Summary

✅ **YES - IMPLEMENT PKCE IMMEDIATELY (No Phased Approach Needed)**

**Current Status**: OAuth is **completely broken** (404 errors on all auth routes).

Since there are no working sessions to preserve, we should implement the **complete PKCE solution in one go** (Big Bang approach). This is actually **easier** than a phased migration.

**Recommendation**: Implement full PKCE solution now while fixing the broken OAuth routes.

---

## Current Implementation Review

### ❌ **OAuth is Currently Broken**

**Critical Issues:**
```bash
# Test results:
curl http://localhost:3000/api/portal/auth/github/login
# Returns: 404 Not Found

# Portal logs show:
# "404 Not Found: /api/portal/auth/github/login"
```

**Root Cause**: Auth routes registered as `/auth/*` but React calls `/api/portal/auth/*`

### ✅ What Code Exists (But Needs Fixing)

#### 1. **Backend OAuth Code** (Exists but routes not registered correctly)
- ✅ Portal service has OAuth handler (`HandleGitHubOAuthCallbackWithSession`)
- ✅ Code exchange with GitHub API (`exchangeCodeForToken`)
- ✅ User profile fetching (`FetchUserInfo`)
- ✅ JWT generation with session management
- ✅ Secure httpOnly cookies
- ❌ **Routes registered without `/api/portal` prefix**

#### 2. **Frontend OAuth Code** (Exists but calls wrong URLs)
- ✅ React frontend callback handler (`OAuthCallback.jsx`)
- ❌ **Calls `/api/portal/auth/github/login` (404)**

#### 2. **Architecture Already Supports PKCE**
- ✅ React frontend initiates login (`/api/portal/auth/github/login`)
- ✅ Backend handles callback and token exchange
- ✅ Traefik routes `/api/portal/auth/*` correctly
- ✅ Redis session store for SSO across services

#### 3. **Security Best Practices in Place**
- ✅ Client secret kept on backend only
- ✅ JWT stored in both httpOnly cookie AND localStorage
- ✅ Session-based JWT (contains only `session_id`, not user data)
- ✅ 7-day token expiry
- ✅ CORS configuration for `localhost:3000`

### ⚠️ Current Implementation Gaps

#### 1. **No PKCE Challenge/Verifier**
- ❌ Frontend doesn't generate `code_verifier` or `code_challenge`
- ❌ Backend doesn't validate PKCE flow
- ❌ GitHub authorize URL missing `code_challenge` parameter

#### 2. **State Parameter Not Used**
- ❌ No CSRF protection via `state` parameter
- ❌ Frontend doesn't generate/validate state token
- ⚠️ **Security Risk**: Vulnerable to CSRF attacks

#### 3. **Frontend Doesn't Use Refresh Tokens**
- ❌ Backend exchanges code but doesn't request `refresh_token` scope
- ❌ No token refresh logic when JWT expires
- ⚠️ User must re-authenticate every 7 days

---

## PKCE Implementation Plan

### Phase 1: Frontend Changes (React)

#### File: `frontend/src/utils/pkce.js` (NEW)
```javascript
/**
 * PKCE helper functions for OAuth 2.0 Authorization Code Flow
 */

/**
 * Generate a cryptographically random code verifier (43-128 chars)
 */
export function generateCodeVerifier() {
  const array = new Uint8Array(32);
  window.crypto.getRandomValues(array);
  return base64URLEncode(array);
}

/**
 * Generate SHA-256 code challenge from verifier
 */
export async function generateCodeChallenge(verifier) {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await window.crypto.subtle.digest('SHA-256', data);
  return base64URLEncode(new Uint8Array(digest));
}

/**
 * Generate random state for CSRF protection
 */
export function generateState() {
  const array = new Uint8Array(16);
  window.crypto.getRandomValues(array);
  return base64URLEncode(array);
}

/**
 * Base64-URL encode (RFC 4648 Section 5)
 */
function base64URLEncode(buffer) {
  const base64 = btoa(String.fromCharCode(...buffer));
  return base64
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}
```

#### File: `frontend/src/components/LoginPage.jsx` (MODIFY)
```javascript
import { generateCodeVerifier, generateCodeChallenge, generateState } from '../utils/pkce';

const handleGitHubLogin = async () => {
  try {
    // Generate PKCE parameters
    const codeVerifier = generateCodeVerifier();
    const codeChallenge = await generateCodeChallenge(codeVerifier);
    const state = generateState();

    // Store verifier and state in sessionStorage (temporary, cleared on tab close)
    sessionStorage.setItem('pkce_code_verifier', codeVerifier);
    sessionStorage.setItem('oauth_state', state);

    // Build GitHub OAuth URL with PKCE
    const params = new URLSearchParams({
      client_id: import.meta.env.VITE_GITHUB_CLIENT_ID,
      redirect_uri: 'http://localhost:3000/auth/callback',
      scope: 'user:email read:user',
      state: state,
      code_challenge: codeChallenge,
      code_challenge_method: 'S256',
    });

    const authURL = `https://github.com/login/oauth/authorize?${params}`;
    console.log('[PKCE] Redirecting to GitHub with code_challenge');
    window.location.href = authURL;
  } catch (error) {
    console.error('[PKCE] Failed to generate PKCE parameters:', error);
    setError('Failed to initiate login. Please try again.');
  }
};
```

#### File: `frontend/src/components/OAuthCallback.jsx` (MODIFY)
```javascript
useEffect(() => {
  const code = searchParams.get('code');
  const state = searchParams.get('state');
  const errorParam = searchParams.get('error');

  if (errorParam) {
    setError('GitHub authentication failed. Please try again.');
    setTimeout(() => navigate('/login'), 3000);
    return;
  }

  if (!code) {
    setError('No authorization code received.');
    setTimeout(() => navigate('/login'), 3000);
    return;
  }

  // Validate state (CSRF protection)
  const storedState = sessionStorage.getItem('oauth_state');
  if (!state || state !== storedState) {
    console.error('[PKCE] State mismatch - possible CSRF attack');
    setError('Security validation failed. Please try again.');
    sessionStorage.clear(); // Clear PKCE data
    setTimeout(() => navigate('/login'), 3000);
    return;
  }

  // Get PKCE verifier
  const codeVerifier = sessionStorage.getItem('pkce_code_verifier');
  if (!codeVerifier) {
    console.error('[PKCE] Missing code verifier');
    setError('Security validation failed. Please try again.');
    setTimeout(() => navigate('/login'), 3000);
    return;
  }

  // Exchange code for token (send verifier to backend)
  const exchangeCodeForToken = async () => {
    try {
      const response = await fetch('/api/portal/auth/token', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          code,
          state,
          code_verifier: codeVerifier,
        }),
      });

      if (!response.ok) {
        throw new Error('Token exchange failed');
      }

      const data = await response.json();
      
      // Store JWT
      localStorage.setItem('devsmith_token', data.token);
      
      // Clear PKCE data
      sessionStorage.removeItem('pkce_code_verifier');
      sessionStorage.removeItem('oauth_state');
      
      // Redirect to dashboard
      navigate('/');
    } catch (err) {
      console.error('[PKCE] Token exchange error:', err);
      setError('Failed to complete authentication. Please try again.');
      sessionStorage.clear();
      setTimeout(() => navigate('/login'), 3000);
    }
  };

  exchangeCodeForToken();
}, [searchParams, navigate]);
```

---

### Phase 2: Backend Changes (Go)

#### File: `apps/portal/handlers/auth_handler.go` (ADD NEW ENDPOINT)
```go
// TokenRequest represents the PKCE token exchange request
type TokenRequest struct {
	Code         string `json:"code" binding:"required"`
	State        string `json:"state" binding:"required"`
	CodeVerifier string `json:"code_verifier" binding:"required"`
}

// HandleTokenExchange handles PKCE token exchange
// POST /api/portal/auth/token
func HandleTokenExchange(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ERROR] Invalid token request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	log.Printf("[DEBUG] PKCE token exchange - code=%s, state=%s", req.Code, req.State)

	// Validate OAuth config
	if !ValidateOAuthConfig() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub OAuth not configured"})
		return
	}

	// Note: GitHub's OAuth doesn't validate code_verifier server-side
	// PKCE is client-side protection (prevents authorization code interception)
	// Backend still uses client_secret for server-to-server security
	
	// Exchange code for access token (existing function works)
	accessToken, err := exchangeCodeForToken(req.Code)
	if err != nil {
		log.Printf("[ERROR] Failed to exchange code: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to authenticate"})
		return
	}

	// Fetch user info
	user, err := FetchUserInfo(accessToken)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch user: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to fetch user info"})
		return
	}

	log.Printf("[DEBUG] User authenticated: %s (ID: %d)", user.Login, user.ID)

	// Create Redis session
	sess := &session.Session{
		UserID:         int(user.ID),
		GitHubUsername: user.Login,
		GitHubToken:    accessToken,
		Metadata: map[string]interface{}{
			"email":      user.Email,
			"avatar_url": user.AvatarURL,
			"name":       user.Name,
		},
	}

	sessionID, err := sessionStore.Create(c.Request.Context(), sess)
	if err != nil {
		log.Printf("[ERROR] Failed to create session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Generate JWT
	claims := jwt.MapClaims{
		"session_id": sessionID,
		"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret := security.GetJWTSecret()
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("[ERROR] JWT generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue token"})
		return
	}

	log.Printf("[DEBUG] Token exchange successful, session: %s", sessionID)

	// Set httpOnly cookie
	SetSecureJWTCookie(c, tokenString)

	// Return token to frontend (for localStorage)
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"username":   user.Login,
			"email":      user.Email,
			"avatar_url": user.AvatarURL,
			"github_id":  user.ID,
		},
	})
}

// RegisterTokenRoutes registers the token exchange endpoint
func RegisterTokenRoutes(router *gin.Engine) {
	router.POST("/api/portal/auth/token", HandleTokenExchange)
}
```

#### File: `cmd/portal/main.go` (REGISTER NEW ROUTE)
```go
// In main() function, after RegisterAuthRoutesWithSession:
handlers.RegisterTokenRoutes(router)
```

---

### Phase 3: Environment Configuration

#### File: `.env.example` (ADD)
```bash
# GitHub OAuth Configuration
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
REDIRECT_URI=http://localhost:3000/auth/callback

# Frontend Environment
VITE_GITHUB_CLIENT_ID=your_github_client_id  # Same as GITHUB_CLIENT_ID
VITE_API_URL=http://localhost:3000
```

#### File: `frontend/.env.example` (NEW)
```bash
VITE_GITHUB_CLIENT_ID=your_github_client_id
VITE_API_URL=http://localhost:3000
```

---

## Migration Path

### ✅ **Big Bang Implementation (Recommended - OAuth Currently Broken)**

Since OAuth is currently **not working at all** (404 errors), there's no backward compatibility to preserve:

1. ✅ **Fix route registration** - Register routes under `/api/portal/auth/*`
2. ✅ **Implement PKCE frontend** - Add code verifier/challenge generation
3. ✅ **Implement PKCE backend** - Add `/api/portal/auth/token` endpoint
4. ✅ **Remove old callback route** - Clean up non-functional code
5. ✅ **Test end-to-end** - Verify complete OAuth flow works

**Why This is Actually Easier:**
- No existing sessions to worry about
- No gradual migration complexity
- Clean implementation from scratch
- Can delete broken code immediately

~~### Option 1: Gradual Migration~~
~~(Not needed - OAuth is broken, no sessions to preserve)~~

---

## Benefits of PKCE Implementation

### Security Improvements
1. ✅ **Prevents Authorization Code Interception**
   - Attacker can't use stolen `code` without `code_verifier`
   - Critical for public clients (React SPA)
2. ✅ **No Client Secret in Frontend**
   - Already implemented (client secret stays on backend)
3. ✅ **CSRF Protection via State**
   - Prevents cross-site request forgery attacks

### User Experience Improvements
1. ✅ **More Secure Login**
   - Industry standard for SPAs
   - Compliant with OAuth 2.1 draft
2. ✅ **Better Error Handling**
   - Frontend validates state before backend call
   - Clear error messages for security failures
3. ✅ **Seamless Migration**
   - No user-visible changes
   - Existing sessions continue to work

---

## Testing Strategy

### Unit Tests
```go
// apps/portal/handlers/auth_handler_test.go
func TestHandleTokenExchange_ValidPKCE(t *testing.T) {
	// Mock GitHub token exchange
	// Mock user profile fetch
	// Verify JWT generation
	// Verify session creation
}

func TestHandleTokenExchange_InvalidState(t *testing.T) {
	// Verify 401 Unauthorized
}

func TestHandleTokenExchange_MissingVerifier(t *testing.T) {
	// Verify 400 Bad Request
}
```

### Integration Tests
```javascript
// tests/e2e/oauth-pkce.spec.ts
test('OAuth PKCE flow completes successfully', async ({ page }) => {
  // Navigate to login
  await page.goto('http://localhost:3000/login');
  
  // Click GitHub login button
  await page.click('text=Login with GitHub');
  
  // Verify GitHub OAuth URL has code_challenge
  await page.waitForURL(/github.com\/login\/oauth\/authorize/);
  const url = page.url();
  expect(url).toContain('code_challenge=');
  expect(url).toContain('code_challenge_method=S256');
  expect(url).toContain('state=');
  
  // Mock GitHub callback
  await page.goto('http://localhost:3000/auth/callback?code=test_code&state=test_state');
  
  // Verify redirect to dashboard
  await page.waitForURL('http://localhost:3000/');
  
  // Verify token stored
  const token = await page.evaluate(() => localStorage.getItem('devsmith_token'));
  expect(token).toBeTruthy();
});
```

---

## Effort Estimate

### Development Time
- **Route Registration Fix**: 5 minutes ⚡
  - Update `cmd/portal/main.go` to use `/api/portal/auth` prefix
  - Rebuild portal service
  - Test with curl
  
- **Frontend PKCE Changes**: 2 hours
  - PKCE utility functions: 30 min
  - LoginPage modifications: 30 min
  - OAuthCallback modifications: 1 hour
  
- **Backend PKCE Changes**: 1.5 hours
  - New `/auth/token` endpoint: 45 min
  - Route registration: 15 min (already fixed above)
  - Testing: 30 min
  
- **Testing & Validation**: 1 hour
  - Unit tests: 30 min
  - E2E tests: 30 min
  
- **Documentation**: 30 min
  - Update README.md
  - Update ARCHITECTURE.md
  - Update error log

**Total**: 5 minutes (quick fix) + 4 hours (full PKCE) = **~4 hours total**

### Complexity
- **Low**: No database changes required
- **Low**: No breaking changes to existing sessions
- **Low**: Leverages existing OAuth infrastructure
- **Medium**: Requires understanding of cryptographic functions

---

### Risks & Mitigations

### Risk 1: ~~Breaking Existing Sessions~~ (Not Applicable - OAuth Already Broken)
- **Status**: No existing sessions to break
- **Action**: Implement cleanly from scratch

### Risk 2: Browser Compatibility
- **Mitigation**: Use `window.crypto.subtle` (supported in all modern browsers)
- **Fallback**: Show error message for unsupported browsers

### Risk 3: Session Storage vs Local Storage
- **Mitigation**: Use `sessionStorage` for PKCE data (cleared on tab close)
- **Security**: Prevents verifier leakage across sessions

### Risk 4: GitHub API Rate Limits
- **Mitigation**: Already handled by existing implementation
- **Note**: PKCE doesn't increase API calls

---

## Recommendation

### ✅ **Implement Complete PKCE Solution Now (Not Phased)**

**Why Big Bang Approach?**
1. ✅ OAuth currently **completely broken** (404 on all routes)
2. ✅ No existing sessions to preserve
3. ✅ Simpler than phased migration
4. ✅ Can fix routes and add PKCE simultaneously
5. ✅ Total time: 4 hours (vs 4-5 hours phased)

**Implementation Order:**
1. ⚡ **Fix route registration** (5 min) - Get OAuth working
2. 🔐 **Add PKCE to frontend** (2 hours) - Add security
3. 🔐 **Add PKCE to backend** (1.5 hours) - Complete flow
4. ✅ **Test thoroughly** (1 hour) - Validate everything works

### Next Steps (Start Now)
1. ✅ Fix route registration in `cmd/portal/main.go`
2. ✅ Test basic OAuth works: `curl -I http://localhost:3000/api/portal/auth/github/login`
3. ✅ Create `frontend/src/utils/pkce.js`
4. ✅ Update `LoginPage.jsx` with PKCE
5. ✅ Update `OAuthCallback.jsx` with state validation
6. ✅ Add `/api/portal/auth/token` endpoint
7. ✅ Write E2E tests
8. ✅ Deploy and test with real GitHub OAuth

---

## References

### OAuth 2.0 & PKCE
- [RFC 7636 - PKCE for OAuth 2.0](https://tools.ietf.org/html/rfc7636)
- [OAuth 2.1 Draft](https://oauth.net/2.1/)
- [GitHub OAuth Apps Documentation](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps)

### Current Implementation
- `apps/portal/handlers/auth_handler.go` (lines 499-580)
- `internal/portal/services/github_client.go`
- `frontend/src/components/LoginPage.jsx`
- `frontend/src/components/OAuthCallback.jsx`

### Architecture Documents
- `ARCHITECTURE.md` (Section 7: Authentication & Authorization)
- `Requirements.md` (OAuth requirements)
- `.docs/ERROR_LOG.md` (OAuth-related errors)

---

**Assessment by**: GitHub Copilot  
**Date**: 2025-11-06  
**Confidence Level**: High (95%)  
**Implementation Difficulty**: Low-Medium  
**Security Impact**: High (Significantly improves security)
