# Traefik Router Priority Fix

**Date**: 2025-01-XX  
**Issue**: Gateway (port 3000) served stale JavaScript while frontend container had new code  
**Root Cause**: Traefik router priority misconfiguration

## Problem Details

### Symptoms
- Frontend container (http://localhost:5173) served NEW code: `index-CxhvLpzd.js`
- Gateway (http://localhost:3000) served OLD code: `index-D-ZYrNfr.js`
- Restarting Traefik didn't help
- Cache-busting headers didn't help
- Container inspection showed correct configuration

### Investigation Path
1. ✅ Verified frontend container has new code (direct access worked)
2. ✅ Verified Traefik routing labels correct
3. ✅ Tested cache-busting strategies (all failed)
4. ✅ Checked Traefik routes - found `dashboard@internal` with PathPrefix(`/`)

### Root Cause
Traefik router priorities:
- `frontend@docker`: priority `1` (LOW)
- `dashboard@internal`: priority `2147483645` (NEARLY MAX - Traefik dashboard)

In Traefik, **HIGHER priority number = HIGHER precedence**. The internal Traefik dashboard was intercepting requests to `/` before they reached our frontend!

## Solution

Changed frontend router priority in docker-compose.yml:

```yaml
# BEFORE:
- "traefik.http.routers.frontend.priority=1"

# AFTER:
- "traefik.http.routers.frontend.priority=2147483647"  # Max priority to override Traefik dashboard
```

Then recreated the container:
```bash
docker-compose up -d --force-recreate frontend
```

## Verification

```bash
# Before fix:
curl -s http://localhost:3000/ | grep -o 'index-[^"]*\.js'
# Output: index-D-ZYrNfr.js (OLD)

# After fix:
curl -s http://localhost:3000/ | grep -o 'index-[^"]*\.js'
# Output: index-CxhvLpzd.js (NEW) ✅
```

## Prevention

1. **Always set explicit router priorities** for custom routes
2. **Use priority > 2147483645** for application routes to override Traefik internals
3. **Check Traefik dashboard** routes when debugging routing issues:
   ```bash
   curl http://localhost:8090/api/http/routers | jq '.[] | {name, rule, priority}'
   ```

## Related Files
- `/home/mikej/projects/DevSmith-Modular-Platform/docker-compose.yml` (line 99)
- `/home/mikej/projects/DevSmith-Modular-Platform/scripts/rebuild-frontend.sh` (diagnostic script)

## Testing Status

### ✅ Fixed
- Gateway serves new JavaScript code
- Direct container access works
- Traefik routing priority corrected

### ⚠️ Remaining Work
- E2E tests need authentication setup (tests fail because pages require login)
- User needs to verify UI changes work correctly
- May need to clear browser cache for testing

## Lessons Learned

1. **Traefik priority system**: Higher number = higher priority (counter-intuitive)
2. **Internal routes**: Traefik dashboard creates its own high-priority routes
3. **Debugging approach**: Check ALL routes in Traefik API, not just custom ones
4. **Container caching wasn't the issue**: The container had correct code all along

## Commands for Future Reference

```bash
# Check all Traefik routes with priorities
curl -s http://localhost:8090/api/http/routers | jq '[.[] | {name: .name, rule: .rule, priority: .priority}]'

# Check frontend container labels
docker inspect devsmith-frontend | jq '.[0].Config.Labels' | grep traefik

# Quick deployment verification
curl -s http://localhost:3000/ | grep -o 'index-[^"]*\.js'
curl -s http://localhost:5173/ | grep -o 'index-[^"]*\.js'
```
