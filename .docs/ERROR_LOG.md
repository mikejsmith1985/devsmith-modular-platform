# DevSmith Platform: Error Log

**Purpose**: Track all errors encountered during development and ensure they're logged to the Logs service for future debugging.

**Format**: Each error entry should include:
- **Date/Time**: When the error occurred
- **Context**: What was being attempted
- **Error Message**: Exact error text
- **Log Location**: Where this error should appear in Logs app
- **Root Cause**: Why it happened
- **Resolution**: How it was fixed
- **Prevention**: How to avoid in future

---

## 2025-11-03: Portal-Review Integration Issues

### Error 1: Dashboard Showing All Cards as "Ready"

**Date**: 2025-11-03 22:00 UTC  
**Context**: After modifying `dashboard.templ` to show "Coming Soon" badges, dashboard still showed all cards as "Ready"  
**Error Message**: Runtime UI showed all cards with green "Ready" badges despite source code having "Coming Soon"  

**Log Location**: Should appear in Logs app as:
```
Service: portal
Level: WARN
Message: Template mismatch detected - compiled template differs from source
Context: {
  "source_file": "apps/portal/templates/dashboard.templ",
  "compiled_file": "apps/portal/templates/dashboard_templ.go",
  "badge_state_source": "Coming Soon",
  "badge_state_compiled": "Ready"
}
```

**Root Cause**:  
1. Templ templates are compiled to Go files (`*_templ.go`)
2. Modified `.templ` source file but didn't run `templ generate`
3. Docker rebuild used old compiled `_templ.go` files
4. No warning system to detect source/compiled mismatch

**Resolution**:
```bash
# Regenerate all Templ templates
templ generate

# Verify compilation
grep -A 5 "Development Logs" apps/portal/templates/dashboard_templ.go

# Rebuild portal with correct templates
docker-compose up -d --build portal
```

**Prevention**:
1. ✅ **Add to copilot-instructions.md**: Always run `templ generate` before committing `.templ` changes
2. ✅ **Add pre-commit hook**: Validate `.templ` files match `*_templ.go` files
3. ✅ **Add build validation**: Check template consistency before Docker build
4. ✅ **Add runtime check**: Portal startup should validate template versions

**Logged to Platform**: ❌ NOT YET  
**Action Item**: Add template validation check that logs warnings

---

### Error 2: Review Service Returns "Authentication required"

**Date**: 2025-11-03 22:30 UTC  
**Context**: User logged in via GitHub OAuth, has valid JWT cookie, clicks "Open Review" button  
**Error Message**: `HTTP/1.1 401 Unauthorized - Authentication required. Please log in via Portal.`  

**Log Location**: Should appear in Logs app as:
```
Service: review
Level: INFO
Message: User not authenticated, returning 401 on public route
Context: {
  "endpoint": "/review",
  "method": "GET",
  "handler": "HomeHandler",
  "user_id_in_context": false,
  "expected_behavior": "redirect to login"
}
```

**Root Cause**:  
1. Review service route `/review` registered as **public** (no JWT middleware)
2. HomeHandler **manually checks** for `user_id` in context
3. Handler returns **401 error** when `user_id` not found
4. **Mismatch**: Public routes should redirect to login, not return 401
5. Standard web practice: 401 = "protected resource", 302 = "please authenticate"

**Resolution**:
```go
// apps/review/handlers/ui_handler.go - Line 445-449
// OLD CODE (returns 401 on public route):
if !exists {
    h.logger.Warn("User not authenticated, cannot create session")
    c.String(http.StatusUnauthorized, "Authentication required. Please log in via Portal.")
    return
}

// NEW CODE (redirects to login):
if !exists {
    h.logger.Info("User not authenticated, redirecting to portal login")
    c.Redirect(http.StatusFound, "/auth/github/login")
    return
}
```

Steps taken:
1. Modified `apps/review/handlers/ui_handler.go` to redirect instead of 401
2. Rebuilt Review service: `docker-compose up -d --build review`
3. Tested with curl: `curl -I http://localhost:3000/review` → `302 Found`
4. Validated with Playwright: Tests confirm 302 redirect (not 401)

**Prevention**:
1. ✅ **Design principle**: Public routes MUST redirect to login (never return 401)
2. ✅ **Code review**: Check handler logic matches route middleware
3. ✅ **Testing**: Add Playwright test for unauthenticated access (✅ DONE - `tests/e2e/review-auth.spec.ts`)
4. ✅ **Documentation**: Update ARCHITECTURE.md with public vs protected route patterns

**Validation Results**:
```bash
# Playwright test results:
✅ PASS: Review returns 302 redirect (not 401)
   Location: /auth/github/login
✅ PASS: Review does not return 401 (bug fixed!)
   Actual status: 302
```

**Logged to Platform**: ❌ NOT YET  
**Action Item**: Add authentication attempt logging to Review service

**Status**: ✅ RESOLVED - 2025-11-03 23:00 UTC

---
  "cookie_name": "devsmith_token",
  "jwt_secret_configured": true,
  "validation_error": "specific error from jwt.Parse",
  "nginx_forwarded_headers": ["Cookie", "Authorization", "Host", ...]
}
```

**Root Cause**: TBD - Need to investigate:
1. Is JWT cookie being forwarded by nginx?
2. Is Review service reading cookie correctly?
3. Is JWT_SECRET the same in both Portal and Review?
4. Is JWT format correct (HS256 algorithm)?

**Resolution**: IN PROGRESS  

**Prevention**: TBD  

**Logged to Platform**: ❌ NOT YET  
**Action Item**: 
- Add detailed JWT validation logging to Review service
- Log all incoming headers in Review middleware
- Add JWT secret validation check at startup
- Create health check endpoint that validates JWT flow

---

### Error 3: Nginx Not Forwarding Authorization Headers

**Date**: 2025-11-03 20:30 UTC  
**Context**: Review service couldn't validate JWT because nginx wasn't passing Authorization headers  
**Error Message**: None visible (silent failure)  

**Log Location**: Should appear in Logs app as:
```
Service: nginx
Level: WARN
Message: Authorization header not forwarded to backend service
Context: {
  "upstream": "review",
  "path": "/review",
  "client_ip": "...",
  "headers_forwarded": ["Cookie", "Host", ...],
  "headers_missing": ["Authorization"]
}
```

**Root Cause**:
1. Nginx default config doesn't forward `Authorization` header
2. No logging to indicate header was dropped
3. Review service only logs "auth required", not "header missing"

**Resolution**:
```nginx
# Added to docker/nginx/conf.d/default.conf
location /review {
    proxy_pass http://review:8081;
    proxy_set_header Authorization $http_authorization;  # CRITICAL
    proxy_set_header Cookie $http_cookie;
    # ... other headers
}
```

**Prevention**:
1. ✅ **Document requirement**: nginx must forward auth headers
2. ✅ **Add validation**: nginx startup should verify proxy_set_header directives
3. ✅ **Add logging**: Log when auth headers are present/missing
4. ✅ **Add health check**: Validate header forwarding in docker-validate.sh

**Logged to Platform**: ❌ NOT YET  
**Action Item**: Add nginx access log parsing to Logs service

---

## Template for Future Errors

```markdown
### Error N: [Brief Description]

**Date**: YYYY-MM-DD HH:MM UTC  
**Context**: [What was being attempted]  
**Error Message**: [Exact error text or symptom]  

**Log Location**: Should appear in Logs app as:
```
Service: [service_name]
Level: [ERROR|WARN|INFO]
Message: [log message]
Context: {
  "field": "value",
  ...
}
```

**Root Cause**: [Why it happened]  

**Resolution**: [How it was fixed with code/commands]  

**Prevention**:  
1. [Step to prevent recurrence]
2. [Additional measures]

**Logged to Platform**: [YES ✅ | NO ❌ | PARTIAL ⚠️]  
**Action Item**: [What needs to be implemented]
```

---

## Error Categories

### Template Errors (Category: TEMPLATE)
- Source/compiled mismatch
- Missing template regeneration
- Template syntax errors

### Authentication Errors (Category: AUTH)
- JWT validation failures
- Missing credentials
- Token expiration
- Header forwarding issues

### Routing Errors (Category: ROUTE)
- Nginx misconfiguration
- Service route registration
- CORS issues

### Database Errors (Category: DB)
- Connection failures
- Query errors
- Migration issues

### Build Errors (Category: BUILD)
- Docker build failures
- Dependency issues
- Compilation errors

---

## Logs App Integration Requirements

When implementing the Logs application, ensure it can:

1. **Display Error Context**:
   - Show full error with all context fields
   - Link to this ERROR_LOG.md for known issues
   - Highlight critical fields (service, level, timestamp)

2. **Search by Category**:
   - Filter by error category (TEMPLATE, AUTH, ROUTE, etc.)
   - Search by service name
   - Filter by date range

3. **Error Frequency**:
   - Show how many times each error occurred
   - Trending errors (increasing/decreasing)
   - Alert on new error patterns

4. **Root Cause Linking**:
   - Link log entries to ERROR_LOG.md entries
   - Show "Known Issue" badge if error matches documented case
   - Provide quick link to resolution steps

5. **Prevention Tracking**:
   - Show which prevention measures are implemented
   - Track if an error recurs after being "fixed"
   - Alert if preventable error happens again

---

## Maintenance

- **Update Frequency**: Add entry immediately when error is encountered
- **Review Cycle**: Weekly review to identify patterns
- **Cleanup**: Archive resolved errors after 90 days (move to ERROR_LOG_ARCHIVE.md)
- **Ownership**: All team members (OpenHands, Claude, Copilot, Mike) must log errors here
