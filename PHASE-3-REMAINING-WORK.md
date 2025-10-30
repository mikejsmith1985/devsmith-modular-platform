# Phase 3 Implementation - Remaining Work

**Status:** Core services complete and compiling ✅

**Current Phase:** Infrastructure complete, now need API/UI integration and testing

---

## What's Done ✅

### Database & Services (Core Infrastructure)
- [x] Database migration: 5 new tables in `logs` schema
  - `health_checks` - store health check reports
  - `health_check_details` - individual check results
  - `security_scans` - Trivy vulnerability scan results
  - `auto_repairs` - repair action history
  - `health_policies` - per-service configuration

- [x] **HealthStorageService** - Store/retrieve health check history, trend analysis
- [x] **HealthPolicyService** - Per-service policy management with defaults
- [x] **AutoRepairService** - Intelligent restart/rebuild logic
- [x] **HealthScheduler** - Background 5-minute health checks with auto-repair

**All services compile successfully and have no linter errors.**

---

## What Remains ⏳

### 1. Wire into `cmd/logs/main.go` (15-20 min)
**File:** `cmd/logs/main.go`

Initialize and start the Phase 3 services:

```go
// After database connection is established
storageService := services.NewHealthStorageService(sqlDB)
policyService := services.NewHealthPolicyService(sqlDB)
autoRepairService := services.NewAutoRepairService(sqlDB, policyService)
scheduler := services.NewHealthScheduler(5*time.Minute, storageService, autoRepairService)

// Initialize default policies
policyService.InitializeDefaultPolicies(context.Background())

// Start background scheduler
scheduler.Start()

// Ensure scheduler stops gracefully on shutdown
defer scheduler.Stop()
```

### 2. Create API Handlers (20-30 min)
**File:** `cmd/logs/handlers/health_history_handler.go` (already exists, needs impl)

Implement endpoints:
```
GET  /api/health/history?limit=50         # Recent health checks
GET  /api/health/trends/:service?hours=24 # Trend data for service
GET  /api/health/policies                 # All policies
PUT  /api/health/policies/:service        # Update policy
GET  /api/health/repairs?limit=50         # Repair history
POST /api/health/repair/:service          # Manual repair trigger
```

### 3. Wire Handlers into Router (10 min)
**File:** `cmd/logs/main.go` (same main.go)

Register routes in Gin router:
```go
healthHandler := handlers.NewHealthHistoryHandler(
    storageService, 
    policyService, 
    autoRepairService,
)

health := router.Group("/api/health")
health.GET("/history", healthHandler.GetHistory)
health.GET("/trends/:service", healthHandler.GetTrends)
health.GET("/policies", healthHandler.GetPolicies)
health.PUT("/policies/:service", healthHandler.UpdatePolicy)
health.GET("/repairs", healthHandler.GetRepairs)
health.POST("/repair/:service", healthHandler.ManualRepair)
```

### 4. Update Dashboard UI (30-40 min)
**Files:** 
- `apps/logs/templates/healthcheck.templ` (already exists, needs tabs)
- `apps/logs/templates/health_trends.templ` (created, needs impl)
- `apps/logs/templates/security_scans.templ` (created, needs impl)
- `apps/logs/templates/health_policies.templ` (created, needs impl)

**Tasks:**
- Add tab navigation to main healthcheck template
- Implement trends tab with Chart.js graphs
- Implement security scans tab with vulnerability display
- Implement policies tab with editable policy cards
- Wire up HTMX calls to new API endpoints

### 5. Run Database Migrations (5 min)
```bash
# Migration files are in: internal/logs/db/migrations/008_health_intelligence.sql
cd /home/mikej/projects/DevSmith-Modular-Platform
go run cmd/logs/main.go  # Will run migrations on startup, or use explicit migration tool
```

### 6. Write Tests (30-60 min)
Create comprehensive tests for all Phase 3 components:

- [x] `auto_repair_service_test.go` (simplified, needs integration tests)
- [x] `health_policy_service_test.go` (basic unit tests, needs DB integration tests)
- [ ] `health_storage_service_test.go` (NEW - needs creation)
- [ ] `health_scheduler_test.go` (NEW - needs creation)

### 7. Test WebSocket Integration (Optional Phase 4)
Real-time health updates via WebSocket are documented but not yet implemented.

---

## Key Files to Modify/Create

### Modify:
- `cmd/logs/main.go` - Wire up all Phase 3 services
- `apps/logs/templates/healthcheck.templ` - Add tab navigation
- `cmd/logs/handlers/ui_handler.go` - Pass storage/policy services to templates

### Create/Complete:
- `cmd/logs/handlers/health_history_handler.go` - API endpoint implementations
- `apps/logs/templates/health_trends.templ` - Tab content
- `apps/logs/templates/security_scans.templ` - Tab content  
- `apps/logs/templates/health_policies.templ` - Tab content
- Tests for all Phase 3 components

---

## Testing Checklist

Before committing Phase 3:

- [ ] Services compile: `go build ./cmd/logs`
- [ ] All tests pass: `go test ./internal/logs/...`
- [ ] Database migrations run: `docker-compose up -d logs postgres`
- [ ] Health checks run every 5 minutes (verify in logs)
- [ ] API endpoints respond: `curl http://localhost:8082/api/health/history`
- [ ] Dashboard displays trends/security/policies tabs
- [ ] Auto-repair executes when services fail
- [ ] Manual repair endpoint works: `POST /api/health/repair/review`

---

## Architecture Notes

### Design Decisions (Already Implemented)
1. **Integrated into Logs Service** - Not separate app
   - Single source of truth for observability
   - Shares database, auth, UI stack
   - Cross-correlation: health events + logs + scans

2. **Stateless Health Scheduler**
   - Runs in-process as background goroutine
   - Can be horizontally scaled (Redis pub/sub future)
   - No external coordination needed

3. **Intelligent Auto-Repair**
   - Analyzes issue type before action
   - Policy-based per-service configuration
   - Timeouts → restart, Crashes → rebuild
   - Dependencies → no repair (fix dependencies first)

4. **Security Scanning**
   - Wraps existing `scripts/trivy-scan.sh`
   - Runs as part of health check cycle
   - No separate infrastructure needed

---

## Next Steps for Implementer

1. Start with `cmd/logs/main.go` - Wire up services (easiest, unblocks others)
2. Create `health_history_handler.go` - API layer (mechanical)
3. Update templates for UI - Visual feedback (most user-facing)
4. Write tests - Validation and future-proofing
5. Run integration tests - Full system verification

---

## Questions for Developer

- Should auto-repair be disabled by default for logs service? (Currently: false in policy)
- Should we implement WebSocket real-time updates in Phase 3 or Phase 4?
- Any custom health policies needed beyond defaults?
- Should manual repair require confirmation prompt?

