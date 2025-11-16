# Phase 2 Session 6: Manual Testing Plan

**Status:** Route registration complete (commit e8f022d)  
**Branch:** feature/phase2-github-integration  
**Date:** 2025-11-04

## Routes Registered

All routes protected with JWT authentication:

1. `POST /api/review/sessions/github` - CreateSession
2. `GET /api/review/sessions/:id/github` - GetSession
3. `GET /api/review/sessions/:id/tree` - GetTree
4. `POST /api/review/sessions/:id/files` - OpenFile
5. `GET /api/review/sessions/:id/files` - GetOpenFiles
6. `DELETE /api/review/files/:tab_id` - CloseFile
7. `PATCH /api/review/sessions/:id/files/activate` - SetActiveTab
8. `POST /api/review/sessions/:id/analyze` - AnalyzeMultipleFiles

## Prerequisites

1. **Environment Variables:**
   ```bash
   export GITHUB_TOKEN="your_github_personal_access_token"
   export REVIEW_DB_URL="postgresql://devsmith:devsmith@localhost:5432/devsmith?sslmode=disable"
   export PORT="8081"
   ```

2. **Start Services:**
   ```bash
   docker-compose up -d postgres
   docker-compose up review
   ```

3. **Get JWT Token:**
   ```bash
   # Authenticate via Portal to get JWT token
   curl -X POST http://localhost:3000/auth/github/login
   # Follow OAuth flow, extract token from response
   export JWT_TOKEN="your_jwt_token_here"
   ```

## Test Scenarios

### Test 1: Create GitHub Session
```bash
curl -X POST http://localhost:3000/api/review/sessions/github \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "repository_url": "https://github.com/mikejsmith1985/devsmith-modular-platform",
    "branch": "main"
  }'
```

**Expected Response:**
```json
{
  "session_id": 1,
  "repository_url": "https://github.com/mikejsmith1985/devsmith-modular-platform",
  "branch": "main",
  "tree_sha": "abc123...",
  "created_at": "2025-11-04T..."
}
```

### Test 2: Get Session Details
```bash
SESSION_ID=1
curl -X GET http://localhost:3000/api/review/sessions/$SESSION_ID/github \
  -H "Authorization: Bearer $JWT_TOKEN"
```

**Expected Response:**
```json
{
  "session_id": 1,
  "repository_url": "https://github.com/...",
  "branch": "main",
  "tree_sha": "abc123...",
  "cache_expires_at": "2025-11-04T...",
  "created_at": "2025-11-04T..."
}
```

### Test 3: Get File Tree
```bash
curl -X GET http://localhost:3000/api/review/sessions/$SESSION_ID/tree \
  -H "Authorization: Bearer $JWT_TOKEN"
```

**Expected Response:**
```json
{
  "tree": [
    {
      "path": "README.md",
      "type": "file",
      "size": 1234
    },
    {
      "path": "cmd",
      "type": "directory"
    }
  ],
  "cached": false
}
```

### Test 4: Open File
```bash
curl -X POST http://localhost:3000/api/review/sessions/$SESSION_ID/files \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "README.md"
  }'
```

**Expected Response:**
```json
{
  "tab_id": 1,
  "file_path": "README.md",
  "content": "# DevSmith Platform...",
  "language": "markdown",
  "is_active": true
}
```

### Test 5: List Open Files
```bash
curl -X GET http://localhost:3000/api/review/sessions/$SESSION_ID/files \
  -H "Authorization: Bearer $JWT_TOKEN"
```

**Expected Response:**
```json
{
  "files": [
    {
      "tab_id": 1,
      "file_path": "README.md",
      "is_active": true,
      "opened_at": "2025-11-04T..."
    }
  ]
}
```

### Test 6: Set Active Tab
```bash
curl -X PATCH http://localhost:3000/api/review/sessions/$SESSION_ID/files/activate \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tab_id": 1
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "active_tab_id": 1
}
```

### Test 7: Close File
```bash
TAB_ID=1
curl -X DELETE http://localhost:3000/api/review/files/$TAB_ID \
  -H "Authorization: Bearer $JWT_TOKEN"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "File closed"
}
```

### Test 8: Multi-File Analysis (Placeholder)
```bash
curl -X POST http://localhost:3000/api/review/sessions/$SESSION_ID/analyze \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "file_paths": ["README.md", "ARCHITECTURE.md"],
    "reading_mode": "skim"
  }'
```

**Expected Response:**
```json
{
  "analysis_id": 1,
  "status": "pending",
  "message": "Multi-file analysis queued (AI integration pending)"
}
```

## Error Cases to Test

### Test E1: Unauthenticated Request
```bash
curl -X GET http://localhost:3000/api/review/sessions/1/github
# Should return 401 Unauthorized
```

### Test E2: Invalid Session ID
```bash
curl -X GET http://localhost:3000/api/review/sessions/99999/github \
  -H "Authorization: Bearer $JWT_TOKEN"
# Should return 404 Not Found
```

### Test E3: Invalid Repository URL
```bash
curl -X POST http://localhost:3000/api/review/sessions/github \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "repository_url": "not-a-url",
    "branch": "main"
  }'
# Should return 400 Bad Request with validation error
```

### Test E4: GitHub API Rate Limit (without token)
```bash
# Stop service, remove GITHUB_TOKEN, restart
unset GITHUB_TOKEN
docker-compose restart review
# Should see warning in logs: "GITHUB_TOKEN not set - rate limited to 60/hour"
```

## Success Criteria

- ✅ All endpoints return expected status codes
- ✅ Response schemas match documentation
- ✅ Authentication errors return 401
- ✅ Session not found returns 404
- ✅ Invalid input returns 400 with clear error
- ✅ File tree caching works (second request faster, `cached: true`)
- ✅ Multi-tab state persists (open/close/activate)
- ✅ Service logs show GitHub API calls
- ✅ No panics or server crashes

## Notes

- **Multi-file analysis AI integration** deferred to Phase 3 (returns placeholder response)
- **Multi-tab UI frontend** deferred to Phase 3 (backend ready)
- **Integration tests** deferred to future integration phase
- **E2E Playwright tests** deferred until multi-tab UI is implemented

## Next Steps

1. Manual testing with curl (this document)
2. Update PR #106 with Session 6 progress
3. Decision point: Merge Phase 2 to development or continue with Phase 3?
4. Phase 3: Multi-tab UI frontend + multi-file AI analysis

## References

- **Handler Implementation:** `internal/review/handlers/github_session_handler.go`
- **Route Registration:** `cmd/review/main.go` (lines 296-306)
- **GitHub Client:** `internal/review/github/default_client.go`
- **Repository:** `internal/review/db/github_repository.go`
