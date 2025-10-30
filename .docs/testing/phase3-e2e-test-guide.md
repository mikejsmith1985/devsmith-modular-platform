# Phase 3: Health Intelligence - E2E Testing Guide

**Status:** Test Plan Ready for Execution  
**Date:** October 30, 2025  
**Scope:** Complete validation of intelligent health check, auto-repair, and security scanning

---

## Test Environment Setup

### Prerequisites
```bash
# Ensure all services are running
docker-compose up -d

# Verify database migrations applied
go run cmd/logs/migrate.go up

# Check health scheduler started
curl http://localhost:8082/api/health/history
```

### Expected Service Status
- ✅ Portal (8080) - Running
- ✅ Review (8081) - Running
- ✅ Logs (8082) - Running
- ✅ Analytics (8083) - Running
- ✅ PostgreSQL (5432) - Running
- ✅ Nginx (3000) - Running

---

## Unit Test Execution

### Phase 3 Decision Logic Tests
```bash
# Auto-repair strategy determination
go test -v ./internal/logs/services -run TestDetermineRepairStrategy
# Expected: 5/5 pass (timeout→restart, crash→rebuild, dependency→none, etc.)

# Issue classification
go test -v ./internal/logs/services -run TestClassifyIssue
# Expected: 6/6 pass (timeout, crash, dependency, security, unknown)

# Service name extraction
go test -v ./internal/logs/services -run TestExtractServiceName
# Expected: 6/6 pass

# Trivy security scanning
go test -v ./internal/healthcheck -run TestTrivyCheckerParsing
# Expected: 2/2 pass (empty output, vulnerable output)

# Health policies
go test -v ./internal/logs/services -run TestDefaultPolicies
# Expected: 4/4 pass (all services have defaults)
```

**Success Criteria:** All unit tests pass with 100% coverage of decision logic

---

## API Integration Tests

### 1. Health History API
```bash
# Test 1: Get recent health checks
curl -s http://localhost:8082/api/health/history?limit=10 | jq '.'
# Expected: Returns array of health checks ordered by timestamp DESC

# Test 2: Verify data structure
curl -s http://localhost:8082/api/health/history?limit=1 | jq '.data[0]'
# Expected: {id, timestamp, overall_status, duration_ms, check_count, triggered_by}

# Test 3: Test limit parameter
curl -s http://localhost:8082/api/health/history?limit=100 | jq '.count'
# Expected: 100 (or less if fewer checks exist)

# Test 4: Verify timestamp ordering
curl -s http://localhost:8082/api/health/history?limit=5 | jq '.data[].timestamp'
# Expected: Descending order (newest first)
```

### 2. Trends API
```bash
# Test 1: Get trend data for a service
curl -s "http://localhost:8082/api/health/trends/portal?hours=24" | jq '.'
# Expected: {service, time_period, data_points[], average, peak}

# Test 2: Verify data points
curl -s "http://localhost:8082/api/health/trends/review?hours=72" | jq '.data.data_points | length'
# Expected: >0 (depends on scheduler activity)

# Test 3: Test different time ranges
for hours in 24 72 168; do
  curl -s "http://localhost:8082/api/health/trends/logs?hours=$hours" | jq '.data.service'
done
# Expected: All return "logs"

# Test 4: Missing service fallback
curl -s "http://localhost:8082/api/health/trends/nonexistent?hours=24" | jq '.data.data_points'
# Expected: Empty array (graceful)
```

### 3. Policies API
```bash
# Test 1: Get all policies
curl -s http://localhost:8082/api/health/policies | jq '.data | length'
# Expected: >=4 (portal, review, logs, analytics)

# Test 2: Get specific policy
curl -s http://localhost:8082/api/health/policies/portal | jq '.data'
# Expected: {service_name: "portal", max_response_time_ms: 500, auto_repair_enabled: true, ...}

# Test 3: Verify all required fields
curl -s http://localhost:8082/api/health/policies | jq '.data[0] | keys'
# Expected: includes id, service_name, max_response_time_ms, repair_strategy, etc.

# Test 4: Update policy
curl -X PUT http://localhost:8082/api/health/policies/portal \
  -H "Content-Type: application/json" \
  -d '{"max_response_time_ms": 750, "auto_repair_enabled": false}'
# Expected: 200 OK, returns updated policy

# Test 5: Verify policy persisted
curl -s http://localhost:8082/api/health/policies/portal | jq '.data.max_response_time_ms'
# Expected: 750
```

### 4. Repair History API
```bash
# Test 1: Get repair history
curl -s "http://localhost:8082/api/health/repairs?limit=20" | jq '.data | length'
# Expected: Array of repair actions (may be empty initially)

# Test 2: Verify repair fields
curl -s "http://localhost:8082/api/health/repairs?limit=1" | jq '.data[0]'
# Expected: {id, timestamp, service_name, issue_type, repair_action, status, ...}

# Test 3: Filter by status
curl -s "http://localhost:8082/api/health/repairs?limit=50" | jq '.data[] | select(.status=="success") | .service_name'
# Expected: List of successfully repaired services
```

---

## Scheduler Integration Tests

### Verify Scheduler is Running
```bash
# Test 1: Check scheduler status
curl -s http://localhost:8082/api/health/history | jq '.data[0].triggered_by'
# Expected: "scheduled"

# Test 2: Verify checks run every 5 minutes
# Wait 6 minutes then check timestamps
sleep 360
FIRST=$(curl -s http://localhost:8082/api/health/history?limit=1 | jq '.data[0].timestamp')
sleep 300
SECOND=$(curl -s http://localhost:8082/api/health/history?limit=1 | jq '.data[0].timestamp')
# Expected: Timestamps differ by ~5 minutes

# Test 3: Verify all check types included
curl -s http://localhost:8082/api/health/history?limit=1 | jq '.data[0].report_json.checks[].name'
# Expected: Contains docker, http_*, database, gateway_routing, performance_metrics, etc.
```

---

## Intelligent Auto-Repair Verification

### Simulate Service Failure and Monitor Repair
```bash
# Test 1: Stop a service (e.g., review)
docker-compose stop review

# Test 2: Wait for scheduler to detect (max 5 min)
sleep 310

# Test 3: Check health history shows failure
curl -s http://localhost:8082/api/health/history?limit=1 | jq '.data[0].overall_status'
# Expected: "fail"

# Test 4: Check if auto-repair was triggered
curl -s "http://localhost:8082/api/health/repairs?limit=10" | jq '.data[] | select(.service_name=="review")'
# Expected: Shows repair action with status=success or pending

# Test 5: Verify service was restarted
docker ps | grep review
# Expected: Review container running

# Test 6: Check health recovered
curl -s http://localhost:8082/api/health/history?limit=1 | jq '.data[0].overall_status'
# Expected: "pass"
```

---

## Dashboard UI Testing

### Health Trends Tab
```bash
# Test 1: Navigate to trends
open http://localhost:8082/healthcheck
# Click "Health Status" → "Historical Trends"

# Test 2: Verify charts render
# Expected: Response time chart loads with data

# Test 3: Test time range selector
# Change "Last 72 Hours" to "Last 24 Hours"
# Expected: Chart updates

# Test 4: Test service filter
# Select "portal" from service dropdown
# Expected: Shows only portal trends

# Test 5: Verify statistics
# Check "Avg Response Time", "Peak Response Time", "Success Rate"
# Expected: All display reasonable numbers
```

### Security Scans Tab
```bash
# Test 1: Navigate to security scans
# Click "Security Scans" tab

# Test 2: Verify vulnerability counts display
# Expected: Shows Critical, High, Medium, Low badges with counts

# Test 3: Verify heatmap renders
# Expected: Table showing services with vulnerability breakdown

# Test 4: Click "Scan Now"
# Expected: Triggers Trivy scan, updates display

# Test 5: Verify vulnerability details
# Click on CRITICAL tab
# Expected: Shows CVE IDs, titles, packages
```

### Health Policies Tab
```bash
# Test 1: Navigate to policies
# Click "Policies" tab

# Test 2: Verify all services shown
# Expected: portal, review, logs, analytics cards visible

# Test 3: Test policy update
# Change portal's max_response_time to 750ms
# Click "Save Changes"
# Expected: Success notification, value persists on refresh

# Test 4: Test toggle auto-repair
# Disable auto-repair for review
# Click "Save Changes"
# Expected: Badge updates to "Auto-Repair Disabled"

# Test 5: Change repair strategy
# For analytics, change from "restart" to "rebuild"
# Click "Save Changes"
# Expected: Badge updates to "rebuild"
```

---

## Security Scanning End-to-End

### Trivy Integration
```bash
# Test 1: Verify Trivy is called
# Check logs for Trivy execution
docker logs devsmith-logs-1 | grep -i trivy
# Expected: Shows Trivy execution logs

# Test 2: Verify vulnerability detection
# If services have known vulnerabilities:
curl -s "http://localhost:8082/api/health/history?limit=1" | jq '.data[0].report_json.checks[] | select(.name=="security_scan")'
# Expected: Shows critical/high/medium/low counts

# Test 3: Verify status based on vulnerability
# If CRITICAL vulns found:
curl -s "http://localhost:8082/api/health/history?limit=1" | jq '.data[0].overall_status'
# Expected: "fail"

# Test 4: Database storage
psql -U devsmith -d devsmith -c "SELECT count(*) FROM logs.security_scans;"
# Expected: >0 (scans stored)
```

---

## Data Persistence Tests

### Database Validation
```bash
# Test 1: Verify health_checks table populated
psql -U devsmith -d devsmith -c "SELECT count(*) FROM logs.health_checks;"
# Expected: >0

# Test 2: Verify health_check_details
psql -U devsmith -d devsmith -c "SELECT count(*) FROM logs.health_check_details;"
# Expected: >0

# Test 3: Verify policies initialized
psql -U devsmith -d devsmith -c "SELECT count(*) FROM logs.health_policies;"
# Expected: >=4

# Test 4: Check data retention (30 days)
psql -U devsmith -d devsmith -c "
  SELECT 
    COUNT(*) as old_checks,
    MIN(timestamp) as oldest
  FROM logs.health_checks 
  WHERE timestamp < NOW() - INTERVAL '30 days';"
# Expected: Should be 0 (cleanup working)
```

---

## Load/Stress Testing

### Continuous Monitoring
```bash
# Test 1: Run health checks rapidly
for i in {1..100}; do
  curl -s http://localhost:8082/api/health/history?limit=1 > /dev/null
done
# Expected: All requests succeed, no timeouts

# Test 2: Parallel API calls
(curl -s http://localhost:8082/api/health/policies &
 curl -s http://localhost:8082/api/health/history &
 curl -s "http://localhost:8082/api/health/trends/portal" &
 wait)
# Expected: All complete successfully

# Test 3: Monitor database performance
# Watch query times grow (should stay <100ms)
for i in {1..10}; do
  time curl -s http://localhost:8082/api/health/history?limit=50 > /dev/null
done
# Expected: Consistent query times <500ms
```

---

## Failure Scenario Testing

### Cascading Failures
```bash
# Test 1: Stop postgres
docker-compose stop postgres

# Test 2: Verify health check fails gracefully
curl -s http://localhost:8082/api/health/history 2>&1 | jq '.error'
# Expected: Error message (not crash)

# Test 3: Verify API returns 500
curl -w "\n%{http_code}\n" http://localhost:8082/api/health/history
# Expected: 500

# Test 4: Restart postgres
docker-compose start postgres

# Test 5: Verify recovery
curl -s http://localhost:8082/api/health/history | jq '.data | length'
# Expected: >0
```

---

## Success Criteria Checklist

- [ ] All unit tests pass (100% decision logic coverage)
- [ ] All API endpoints respond correctly
- [ ] Health trends show accurate data
- [ ] Security scans integrate with Trivy
- [ ] Policies persist and apply correctly
- [ ] Auto-repair triggers on simulated failures
- [ ] Scheduler runs every 5 minutes
- [ ] Dashboard UI renders all tabs
- [ ] Database stores all data correctly
- [ ] System recovers from failures gracefully
- [ ] No memory leaks under load
- [ ] Repair decisions match expected logic
- [ ] 30-day data retention works
- [ ] Cross-service correlation visible

---

## Regression Testing (Post-Deployment)

### Weekly Validation
- [ ] Run all E2E tests
- [ ] Check database disk usage (trends growing)
- [ ] Verify repair success rate trending upward
- [ ] Monitor scheduler interval (should be consistent 5m)
- [ ] Check for any hung checks or timeouts
- [ ] Validate new vulnerabilities detected
- [ ] Confirm alert thresholds still appropriate

---

## Known Issues & Workarounds

**None documented yet - will add as testing progresses**
