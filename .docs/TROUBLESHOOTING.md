# Troubleshooting Guide

Common issues and solutions for DevSmith Modular Platform development.

---

## Build Issues

### ❌ "undefined: SomeType" in main.go

**Cause:** Type definition missing or import path wrong

**Solution:**
```bash
# 1. Check if type exists in the package
grep -r "type SomeType" .

# 2. Verify import path in main.go
# 3. Run go mod tidy
go mod tidy

# 4. Rebuild
go build ./cmd/{service}
```

---

### ❌ "code outside function" error

**Cause:** Copy-paste error - code placed at package level instead of inside function

**Solution:**
```bash
# Find the problem
grep -Hn "^\s*fmt\." cmd/portal/main.go | grep -v "//"

# Move the code inside main() or appropriate function
```

---

### ❌ "Test code in production file"

**Cause:** Test functions in main.go or non-test files

**Solution:**
- Move all `func Test*` to `*_test.go` files
- Never put test code in cmd/*/main.go

---

## Docker Issues

### ❌ Service can't connect to database

**Symptoms:**
```
dial tcp: lookup postgres: no such host
```

**Solution:**
```yaml
# In docker-compose.yml, use service name as host
DATABASE_URL=postgresql://user:pass@postgres:5432/dbname
#                                    ^^^^^^^^ service name, not localhost
```

---

### ❌ "Connection refused" on startup

**Cause:** Service starts before database is ready

**Solution:**
Add healthcheck and depends_on:
```yaml
services:
  portal:
    depends_on:
      postgres:
        condition: service_healthy
```

---

## Pre-Commit Hook Issues

### ❌ Pre-commit hook not running

**Solution:**
```bash
# Make hook executable
chmod +x .git/hooks/pre-commit

# Verify
ls -la .git/hooks/pre-commit
```

---

### ❌ Pre-commit fails but I need to commit urgently

**NOT RECOMMENDED, but if absolutely necessary:**
```bash
git commit --no-verify -m "message"
```

**Better approach:**
```bash
# Fix the actual issue first
go build ./cmd/{service}

# Then commit normally
git commit -m "message"
```

---

## golangci-lint Issues

### ❌ "unused variable/import" errors

**Solution:**
```bash
# Auto-fix imports
goimports -w .

# Remove unused variables
# golangci-lint will show you which ones to remove
golangci-lint run
```

---

### ❌ "function too long" warnings

**Cause:** Function exceeds 100 lines or 50 statements

**Solution:**
- Extract helper functions
- Break complex logic into smaller pieces
- Consider if function is doing too much (SRP violation)

---

## GitHub Actions CI Failures

### ❌ Build passes locally but fails in CI

**Common causes:**
1. Forgot to commit a file
2. Environment variable missing in CI
3. Go version mismatch

**Solution:**
```bash
# Run same checks as CI locally
go test ./...
go build ./cmd/portal
go build ./cmd/review
go build ./cmd/logs
go build ./cmd/analytics
golangci-lint run
```

---

### ❌ Tests pass locally but fail in CI

**Cause:** Test depends on local state or timing

**Solution:**
- Use `t.TempDir()` for file operations
- Mock external dependencies
- Don't rely on specific timing
- Check for race conditions: `go test -race ./...`

---

## Development Workflow Issues

### ❌ Branch already exists (auto-created by workflow)

**Cause:** GitHub Actions auto-created branch after PR merge

**Solution:**
```bash
# Don't create new branch, checkout existing one
git checkout development
git pull origin development
git checkout feature/005-description  # Already exists!
```

**Reference:** See copilot-instructions.md Step 2

---

### ❌ Activity log merge conflict

**Cause:** Multiple people committing causes `.docs/devlog/copilot-activity.md` conflicts

**Solution:**
```bash
# Use sync script
./scripts/sync-and-start-issue.sh 005 description

# Or manually
git stash
git pull origin development
git stash pop
git add .docs/devlog/copilot-activity.md
git commit -m "chore: merge activity log"
```

---

## Service-Specific Issues

### Portal Service

**❌ "GitHub OAuth callback failed"**

Check:
1. `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` in .env
2. Callback URL matches in GitHub OAuth app settings
3. Callback URL: `http://localhost:8080/auth/callback`

---

### Review Service

**❌ "Ollama connection refused"**

**Solution:**
```bash
# Start Ollama
ollama serve &

# Pull model
ollama pull deepseek-coder-v2:16b

# Verify
curl http://localhost:11434/api/version
```

---

### Logs Service

**❌ "WebSocket connection failed"**

Check:
1. Service running on correct port (8082)
2. WebSocket URL: `ws://localhost:8082/ws/logs`
3. CORS settings if accessing from different origin

---

### Analytics Service

**❌ "Permission denied on logs.entries"**

**Cause:** analytics_user doesn't have READ permissions on logs schema

**Solution:**
```sql
-- Run as postgres user
GRANT USAGE ON SCHEMA logs TO analytics_user;
GRANT SELECT ON ALL TABLES IN SCHEMA logs TO analytics_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA logs GRANT SELECT ON TABLES TO analytics_user;
```

---

## Performance Issues

### ❌ Slow test execution

**Solution:**
```bash
# Run tests in parallel
go test -parallel 4 ./...

# Run only short tests
go test -short ./...

# Skip integration tests locally
go test -short -v ./...
```

---

### ❌ Slow Docker builds

**Solution:**
```yaml
# Use BuildKit cache
docker-compose build --no-cache  # Only when needed

# Use layer caching
# Ensure go.mod and go.sum are copied before code
```

---

## Getting Help

1. **Check logs:**
   ```bash
   # Service logs (if using scripts)
   tail -f logs/portal.log
   tail -f logs/review.log

   # Docker logs
   docker-compose logs -f portal
   ```

2. **Run health checks:**
   ```bash
   ./scripts/health-checks.sh
   ```

3. **Verify setup:**
   ```bash
   ./verify-setup.sh
   ```

4. **Check GitHub Actions logs:**
   - Go to Actions tab in GitHub
   - Click on failed workflow
   - Check "Build Services" step for errors

5. **Ask for help:**
   - Include error messages
   - Include relevant logs
   - Describe what you've tried

---

**Last Updated:** 2025-10-20
