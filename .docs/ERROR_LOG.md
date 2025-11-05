# DevSmith Platform: Error Log

**Purpose**: Track all errors encountered during development to:
1. Build institutional knowledge for debugging
2. Train the Logs application's AI for intelligent error analysis
3. Help Mike debug when Copilot is offline
4. Prevent recurring issues

---

## 📝 Error Log Template

Copy this template for each new error:

```markdown
### Error: [Brief Description]
**Date**: YYYY-MM-DD HH:MM UTC  
**Context**: [What were you doing when error occurred]  
**Error Message**: 
```
[Exact error text - code block for formatting]
```

**Root Cause**: [Why did this happen - be specific]  
**Impact**: [What broke, who's affected, severity]  

**Resolution**:
```bash
# Exact commands used to fix
command1
command2
```

**Prevention**: [How to avoid this in future - process changes, validation checks]  
**Time Lost**: [Minutes/hours spent debugging]  
**Logged to Platform**: ❌ NO / ✅ YES [Log ID or location]  
**Related Issue**: #XXX (if applicable)  
**Tags**: [database, migration, ui, docker, networking, etc.]
```

---

## 🎯 Error Categories

### Database Errors
- Schema issues
- Migration failures
- Connection problems
- Query performance

### Service Errors
- Startup failures
- Crash loops
- Health check failures
- Dependency issues

### UI/UX Errors
- Template rendering issues
- Broken user workflows
- Loading spinners stuck
- Navigation problems

### Build/Deploy Errors
- Compilation failures
- Docker build issues
- Image layer problems
- Container restart loops

### Network Errors
- Service-to-service communication
- Gateway routing
- CORS issues
- WebSocket disconnections

### Testing Errors
- Flaky tests
- Mock expectation failures
- Integration test issues
- E2E test failures

---

## 2025-11-04: Missing JWT_SECRET Causes OAuth Panic

### Error: Portal OAuth Login Returns "Failed to authenticate"

**Date**: 2025-11-04 19:33 UTC  
**Context**: User completes GitHub OAuth flow, clicks authorize, gets redirected back to localhost:3000/auth/github/callback  
**Error Message**:
```
{"error":"Failed to authenticate"}

Portal logs show:
2025/11/05 00:33:33 [Recovery] 2025/11/05 - 00:33:33 panic recovered:
JWT_SECRET environment variable is not set - this is required for secure authentication
/app/internal/security/jwt.go:29
```

**Root Cause**:
OAuth flow worked perfectly - GitHub returned valid access token and user info. BUT the JWT token generation panicked because `JWT_SECRET` environment variable was not set in docker-compose.yml.

Flow that failed:
1. ✅ User clicks "Login with GitHub"
2. ✅ Redirects to GitHub OAuth
3. ✅ User authorizes
4. ✅ GitHub redirects to /auth/github/callback with code
5. ✅ Portal exchanges code for access token (got: `gho_***REDACTED***`)
6. ✅ Portal fetches user info from GitHub API (got: mikejsmith1985, id: 157150032)
7. ❌ **PANIC** when trying to create JWT token because JWT_SECRET not set

**Impact**:
- **Severity**: CRITICAL
- Complete OAuth login failure
- User cannot log in to platform
- Error message unhelpful ("Failed to authenticate" - doesn't explain JWT_SECRET missing)
- Regression tests passed because they only tested redirect behavior, not actual authentication completion

**Resolution**:
```bash
# Added JWT_SECRET to docker-compose.yml portal service environment
# Line 73 in docker-compose.yml:
- JWT_SECRET=${JWT_SECRET:-dev-secret-key-change-in-production}

# Restarted portal with new env var
docker-compose up -d portal

# Verified JWT_SECRET is now set
docker-compose exec -T portal env | grep JWT_SECRET
# Output: JWT_SECRET=dev-secret-key-change-in-production
```

**Prevention**:
1. ✅ **Add startup validation**: Portal should check for JWT_SECRET on startup and fail fast with clear error
2. ✅ **Add to .env.example**: Document JWT_SECRET requirement
3. ✅ **Add to docker-compose.yml**: Use default value with override pattern `${VAR:-default}`
4. ✅ **Improve error message**: Change panic to graceful error: "JWT_SECRET not set - check docker-compose.yml"
5. ✅ **Add E2E test**: Create OAuth visual test with screenshots (tests/e2e/oauth-visual-test.spec.ts)
6. ✅ **Container self-healing**: Add health check that validates required env vars

**Why Tests Passed**:
- Regression tests only checked:
  - ✅ Does /login return HTML?
  - ✅ Does /auth/login redirect to GitHub?
  - ✅ Does /dashboard require auth?
- Regression tests DID NOT check:
  - ❌ Does OAuth callback complete successfully?
  - ❌ Is JWT token created?
  - ❌ Can user actually log in end-to-end?

**Mike's Container Strategy Feedback**:
> "I hate docker and I think we should consider a container strategy that self heals and auto updates since we fuck that up basically every time we make a change"

**Valid concerns:**
1. Manual `docker-compose up` after every code change
2. No auto-detection of docker-compose.yml changes
3. Config changes (like missing JWT_SECRET) cause runtime panics instead of startup failures
4. No self-healing for missing env vars

**TODO - Container Improvements**:
1. Add startup validation script that checks all required env vars
2. Add docker-compose healthchecks that validate config
3. Add watch mode for docker-compose.yml changes
4. Consider Docker Compose watch feature (docker compose watch)
5. Add pre-start validation script that fails fast with helpful error messages

**Time Lost**: 45 minutes (multiple OAuth attempts, log analysis, adding debug logging)  
**Logged to Platform**: ❌ NO (panic prevented logging service call)  
**Related Issue**: Phase 2 GitHub Integration  
**Tags**: docker, environment-variables, oauth, jwt, panic-recovery, container-configuration

---

## 2025-11-04: Migration Ordering Bug

### Error: Logs Service Fails to Start - Relation Does Not Exist

**Date**: 2025-11-04 12:26 UTC  
**Context**: Running `docker-compose up -d` after implementing Phase 1 AI analysis features. Migration added AI columns to logs.entries table.  

**Error Message**:
```
logs-1  | 2025/11/04 17:06:09 Failed to run migrations: migration execution failed: 
pq: relation "logs.entries" does not exist
```

**Root Cause**: 
Migration file `009_add_ai_analysis_columns.sql` runs BEFORE `20251025_001_create_log_entries_table.sql` due to alphabetical sorting:
- Alphabetical order: `008` → `009` → `20251025_001` → `20251026_002`
- Correct order: `20251025_001` (create table) → `20251026_002` (add context) → `009` (add AI columns)

Migration 009 tried to ALTER TABLE logs.entries before the table was created.

**Impact**: 
- **Severity**: CRITICAL
- Logs service crash on startup
- Blocked all dependent services (Portal, Review, Analytics)
- Complete platform outage
- Prevented Phase 1 testing and validation

**Resolution**:
```bash
# Renamed migration to fix execution order
mv internal/logs/db/migrations/009_add_ai_analysis_columns.sql \
   internal/logs/db/migrations/20251104_003_add_ai_analysis_columns.sql

# Removed old file from git
git rm internal/logs/db/migrations/009_add_ai_analysis_columns.sql

# Committed fix
git commit -m "fix(logs): rename migration to fix execution order"

# Dropped database and restarted to run migrations fresh
docker-compose down -v
docker-compose up -d

# Verified migration success
docker-compose exec -T postgres psql -U devsmith -d devsmith -c "\d logs.entries"
# Expected: issue_type, ai_analysis, severity_score columns present
```

**Prevention**: 
1. **ALWAYS** use `YYYYMMDD_NNN_description.sql` format for migrations
2. **NEVER** use simple numeric prefixes (001, 002, etc.) - they sort incorrectly
3. Add pre-commit hook to validate migration naming:
   ```bash
   # Check all migrations follow YYYYMMDD_NNN format
   find internal/*/db/migrations -name "*.sql" | grep -v "^[0-9]\{8\}_[0-9]\{3\}_"
   ```
4. Document migration naming standard in ARCHITECTURE.md
5. Add automated test: verify migrations run in chronological order

**Time Lost**: 45 minutes debugging (3 rebuild attempts before discovering root cause)  
**Logged to Platform**: ❌ NO (Logs app not yet fully operational)  
**Related Issue**: Phase 1 AI Diagnostics (#104)  
**Tags**: database, migration, docker, startup-failure, alphabetical-sorting

---

## 2025-11-04: Container-Branch Mismatch

### Error: Review UI Showing Infinite Loading Spinner

**Date**: 2025-11-04 16:45 UTC  
**Context**: User tested Review UI after "Phase 1 complete" declaration. Clicked Review card from dashboard, got stuck on infinite loading spinner.

**Error Message**:
```
Browser: Loading spinner indefinitely visible
No console errors
Network tab: No failed requests
Behavior: Page never transitions from loading state
```

**Root Cause**:
Docker containers were running code from `feature/phase2-github-integration` branch instead of `development` branch. Phase 2 branch had removed authentication checks from `apps/review/handlers/ui_handler.go`:

```diff
// Development branch (correct):
func (h *UIHandler) HomeHandler(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.Redirect(http.StatusFound, "/auth/github/login")
        return
    }
    // ... proper session creation
}

// Phase 2 branch (broken):
func (h *UIHandler) HomeHandler(c *gin.Context) {
    // No authentication check!
    c.Redirect(http.StatusPermanentRedirect, "/review/workspace/demo")
    return
}
```

Without authentication, the redirect loop caused infinite loading state.

**Impact**:
- **Severity**: CRITICAL
- Complete Review UI failure
- User unable to access Review features
- False "complete" status for Phase 1
- No regression tests caught this

**Resolution**:
```bash
# Switched to correct branch
git checkout development

# Rebuilt services from correct branch
docker-compose down
docker-compose up -d --build

# Verified services healthy
docker-compose ps
# Expected: All services showing "Up" and "healthy"

# Tested Review UI manually
open http://localhost:3000
# Click Review card → should redirect to login (not infinite load)
```

**Prevention**:
1. **ALWAYS** verify git branch matches container code before declaring work complete
2. Add validation to deployment scripts:
   ```bash
   CURRENT_BRANCH=$(git branch --show-current)
   CONTAINER_BRANCH=$(docker-compose exec -T review git branch --show-current)
   if [ "$CURRENT_BRANCH" != "$CONTAINER_BRANCH" ]; then
       echo "ERROR: Branch mismatch!"
       exit 1
   fi
   ```
3. **MANDATORY** regression testing before declaring work complete
4. Tag Docker images with git commit SHA to ensure traceability
5. Add automated check: "Does UI show expected state?" (not just "Does service respond?")

**Time Lost**: 20 minutes debugging + 15 minutes rebuilding  
**Logged to Platform**: ❌ NO (discovered during manual testing)  
**Related Issue**: Phase 1 Finalization  
**Tags**: docker, deployment, authentication, ui-regression, branch-mismatch

---

## 2025-11-03: Portal-Review Integration Issues

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
