# Quick Start Guide - Testing Critical Fixes

## 🚀 Immediate Actions

### 1. Rebuild and Restart Services
```bash
cd /home/mikej/projects/DevSmith-Modular-Platform

# Stop services
docker-compose down

# Rebuild with all changes
docker-compose up -d --build

# Watch logs for startup
docker-compose logs -f review portal logs
# Press Ctrl+C when you see "healthy" messages
```

### 2. Wait for Services to be Ready
```bash
# Check service status (wait for all to show "healthy")
docker-compose ps

# Or use health check
./scripts/quick-test.sh
```

### 3. Test in Browser
Open: http://localhost:3000/review

**Test Steps:**
1. ✅ Paste this code:
   ```go
   package main
   
   func main() {
       println("Hello DevSmith")
   }
   ```

2. ✅ Select a model from dropdown (try "DeepSeek Coder 6.7B")

3. ✅ Click "Select Preview" button

4. ✅ Wait 3-5 seconds

5. ✅ Verify analysis appears below form

6. ✅ Try other modes (Skim, Scan, Detailed, Critical)

7. ✅ Try different models and compare output

---

## 📊 What Was Fixed

| Fix | Status | Evidence |
|-----|--------|----------|
| Form binding | ✅ Fixed | `CodeRequest` struct with proper binding |
| Model selector | ✅ Added | Dropdown in UI with 5 models |
| DB connections | ✅ Fixed | Pool size 10 per service, max 200 total |
| Model override | ✅ Wired | Context-based model selection |
| All 5 modes | ✅ Updated | Accept code + model parameters |

---

## 🔍 Quick Debugging

### If form returns "Code required":
```bash
# Check review logs
docker-compose logs review --tail=50 | grep -i "bind"

# Should see: "Code request bound successfully"
```

### If no analysis appears:
```bash
# Check Ollama is running
curl http://localhost:11434/api/tags

# Check review service can reach Ollama
docker-compose exec review curl http://host.docker.internal:11434/api/tags

# Check logs for Ollama errors
docker-compose logs review | grep -i ollama
```

### If "too many clients" error:
```bash
# Check connection count
docker exec devsmith-postgres psql -U devsmith -d devsmith -c \
  "SELECT count(*) FROM pg_stat_activity WHERE datname='devsmith';"

# Should be < 50
# If > 50, restart services
docker-compose restart
```

---

## 📝 Files That Changed

**Backend:**
- `apps/review/handlers/ui_handler.go` - Form binding + model support
- `internal/review/services/ollama_adapter.go` - Context model override
- `cmd/review/main.go` - Connection pool + routes
- `cmd/portal/main.go` - Connection pool
- `cmd/logs/main.go` - Connection pool

**Frontend:**
- `apps/review/templates/session_form.templ` - Model selector dropdown

**Infrastructure:**
- `docker-compose.yml` - PostgreSQL max_connections

---

## ✅ Success Checklist

After restarting services, verify:

- [ ] `docker-compose ps` shows all services healthy
- [ ] `./scripts/quick-test.sh` passes all checks
- [ ] Browser loads http://localhost:3000/review
- [ ] Model dropdown has 5 options
- [ ] Pasting code + clicking mode = analysis appears
- [ ] Different modes return different results
- [ ] Different models produce variations
- [ ] Connection count stays under 50

---

## 🆘 If Something Breaks

**Rollback:**
```bash
git checkout HEAD -- apps/review/ internal/review/ cmd/ docker-compose.yml
docker-compose down
docker-compose up -d --build
```

**Get Help:**
```bash
# Show recent errors
docker-compose logs --tail=100 | grep -i error

# Check database
docker exec -it devsmith-postgres psql -U devsmith

# Check Ollama
curl http://localhost:11434/api/tags
```

---

## 📚 Full Details

See `.docs/IMPLEMENTATION_SUMMARY.md` for complete implementation details.

See `.docs/CRITICAL_FIXES_NEEDED.md` for original problem statement.

---

**Ready to test? Run:** `docker-compose up -d --build && ./scripts/quick-test.sh`
