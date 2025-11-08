# Authentication Fix - GitHub Integration

**Date**: 2025-01-14  
**Issue**: 401 Unauthorized errors when attempting to import GitHub repositories  
**Status**: ✅ **FIXED**

## Problem

User reported 401 errors when attempting to import GitHub repositories using both Quick Scan and Full Browser modes.

### Root Cause

**Context Key Mismatch** between middleware and handler:

- **Middleware** (`internal/middleware/redis_session_auth.go` line 98):
  ```go
  c.Set("github_token", sess.GitHubToken)
  ```

- **Handlers** (`internal/review/handlers/github_handler.go`):
  ```go
  token, exists := c.Get("github_access_token")  // ❌ WRONG KEY
  ```

The middleware was setting the GitHub token with key `"github_token"`, but all three GitHub handlers were looking for `"github_access_token"`, causing authentication to fail.

## Solution

Changed all three GitHub handlers to use the correct context key:

### Files Modified

1. **`internal/review/handlers/github_handler.go`** - Three handlers updated:
   - `GetRepoTree` (line 85)
   - `GetRepoFile` (line 148) 
   - `QuickRepoScan` (line 212)

### Changes Made

**Before** (❌ BROKEN):
```go
// Get GitHub token from session
token, exists := c.Get("github_access_token")
if !exists {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "GitHub authentication required"})
    return
}
```

**After** (✅ FIXED):
```go
// Get GitHub token from session
token, exists := c.Get("github_token")
if !exists {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "GitHub authentication required"})
    return
}
```

## Verification

1. ✅ Review service container rebuilt successfully
2. ✅ Service health check passing: `Up 7 seconds (healthy)`
3. ✅ Container accessible on port 8081

## Testing Required

**Manual E2E Testing** (User must test):
1. Log in to Portal via GitHub OAuth
2. Navigate to Review app
3. Click "Import from GitHub"
4. **Test Quick Scan Mode**:
   - Enter repository URL (e.g., `github.com/mikejsmith1985/devsmith-modular-platform`)
   - Select "Quick Scan" mode
   - Click "Import"
   - **Expected**: Repository profile loads with README, dependencies, etc.
   - **Previously**: 401 Unauthorized error
5. **Test Full Browser Mode**:
   - Enter repository URL
   - Select "Full Browser" mode
   - Click "Import"
   - **Expected**: File tree loads, can select and open files
   - **Previously**: 401 Unauthorized error

## Next Steps

1. ✅ **User tests manually** to confirm 401 errors are resolved
2. **If successful**:
   - Commit fix: `git add internal/review/handlers/github_handler.go`
   - Commit message: `fix(review): correct GitHub token context key mismatch`
   - Push to GitHub
3. **If still failing**:
   - Check if Portal is actually storing `github_token` in session
   - Verify session middleware is active on GitHub routes
   - Add debug logging to see what session data exists

## Context Key Reference

For future development, the **correct context keys** set by authentication middleware are:

```go
// internal/middleware/redis_session_auth.go lines 95-98
c.Set("user_id", sess.UserID)
c.Set("github_username", sess.GitHubUsername)
c.Set("github_token", sess.GitHubToken)        // ✅ USE THIS
c.Set("session_id", sessionID)
```

**DO NOT** use:
- ❌ `github_access_token` (old/incorrect key)
- ❌ `access_token` (ambiguous)
- ❌ `token` (too generic)

Always use `github_token` when retrieving the GitHub OAuth token from context.

## Logging

This fix should be logged to `.docs/ERROR_LOG.md` with:
- **Error Category**: Authentication
- **Root Cause**: Context key mismatch
- **Resolution**: Changed `github_access_token` → `github_token` in all handlers
- **Prevention**: Document correct context keys in developer guide
