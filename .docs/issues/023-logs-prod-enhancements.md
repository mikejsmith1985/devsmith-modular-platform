# Issue #023: Logs Service Production Enhancements

**Priority:** High (Production Readiness)  
**Type:** Feature  
**Complexity:** High (Multiple Interconnected Features)  
**Estimated Effort:** 4-5 days  

---

## Summary

Implement production-ready enhancements to the Logs service to provide visibility into system health, validation errors, and security events. Includes real-time dashboard, alert thresholds, log aggregation queries, and export capabilities.

**Why This Matters:**
- Current logs service ingests data but has no visibility layer
- Validation errors from Issue #21 are being logged but not analyzed
- No alerts when error rates spike
- No way to identify security patterns (e.g., repeated path traversal attempts)
- Need audit trail capability for compliance

---

## Acceptance Criteria

### Core Features (All Required)
- [ ] **Database Schema**: Alert config table with indexes for efficient queries
- [ ] **Data Layer**: Repositories for alert configs, alert events, and aggregations
- [ ] **Service Layer**: 
  - Validation aggregation (top errors, trends)
  - Alert threshold management
  - Log export functionality
- [ ] **API Endpoints** (Logs Service):
  - `GET /api/logs/dashboard/stats` - Real-time validation stats
  - `GET /api/logs/validations/top-errors` - Top errors by frequency
  - `GET /api/logs/validations/trends` - Error rate trending
  - `GET /api/logs/export?format=json|csv&service=review&start=ISO&end=ISO` - Export logs
  - `POST /api/logs/alert-config` - Create alert threshold
  - `GET /api/logs/alert-config/:service` - Get alert config
  - `PUT /api/logs/alert-config/:service` - Update alert config
  - `GET /api/logs/alert-events` - Get triggered alerts
- [ ] **Portal UI** (Portal Service):
  - New page: `/portal/dashboard/logs` - Validation stats dashboard
  - Shows: Total errors, error breakdown by type, error rate graph
  - Time range filters (last hour, 24h, 7d, custom)
  - Service filter
  - Error type filter
  - Admin section to manage alert thresholds (configurable per service)
  - Display triggered alerts
  - Link to export logs
- [ ] **Dashboard Data Queries**:
  - Real-time error count (last 5 min, 1 hour, 24 hours)
  - Error breakdown by type (validation_error, security_violation, etc.)
  - Error rate trending (% of total requests)
  - Top 10 errors by frequency
  - Error distribution by service
- [ ] **Testing**:
  - 70%+ unit test coverage
  - 90%+ critical path coverage (dashboard queries, aggregations)
  - Handler tests for all new endpoints
  - Integration tests for alert logic
- [ ] **Documentation**:
  - ARCHITECTURE.md updated with new schema and queries
  - API documentation for new endpoints
  - Dashboard usage guide

---

## Implementation Strategy

### Phase 1: Database & Data Layer (RED → GREEN)
1. Create migration for alert_configs and alert_events tables
2. Create indexes for efficient aggregation queries
3. Implement repositories for new tables
4. Write tests for database layer

### Phase 2: Service Layer (RED → GREEN)
1. Implement ValidationAggregation service (top errors, trends)
2. Implement AlertThreshold service (CRUD operations)
3. Implement LogExport service (JSON/CSV formatting)
4. Write comprehensive tests

### Phase 3: API Endpoints (RED → GREEN)
1. Implement dashboard stats endpoint
2. Implement top-errors endpoint
3. Implement trends endpoint
4. Implement export endpoint
5. Implement alert config management endpoints
6. Write handler tests

### Phase 4: Portal UI (RED → GREEN)
1. Create dashboard template with stats display
2. Add filters (time range, service, error type)
3. Add admin threshold management interface
4. Add alert display
5. Add export button
6. Write UI tests

### Phase 5: Integration (RED → GREEN)
1. Alert triggering logic (check error rates)
2. Webhook/email notifications (placeholder for Phase 2)
3. Integration tests across services

---

## Database Schema Changes

### New Tables

```sql
-- Alert Configuration (configurable per service)
CREATE TABLE logs.alert_configs (
    id BIGSERIAL PRIMARY KEY,
    service TEXT NOT NULL UNIQUE,
    error_threshold_per_min INT NOT NULL DEFAULT 10,
    warning_threshold_per_min INT NOT NULL DEFAULT 5,
    alert_email TEXT,
    alert_webhook_url TEXT,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Alert Events (triggers when thresholds exceeded)
CREATE TABLE logs.alert_events (
    id BIGSERIAL PRIMARY KEY,
    config_id BIGINT NOT NULL REFERENCES logs.alert_configs(id) ON DELETE CASCADE,
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    error_count INT NOT NULL,
    threshold_value INT NOT NULL,
    alert_sent BOOLEAN NOT NULL DEFAULT false,
    sent_at TIMESTAMPTZ,
    error_type TEXT  -- validation_error, security_violation, etc.
);

-- Indexes for efficient queries
CREATE INDEX idx_alert_configs_service ON logs.alert_configs(service);
CREATE INDEX idx_alert_events_config_id ON logs.alert_events(config_id);
CREATE INDEX idx_alert_events_triggered_at ON logs.alert_events(triggered_at DESC);
CREATE INDEX idx_alert_events_alert_sent ON logs.alert_events(alert_sent);

-- For validation error aggregations
CREATE INDEX idx_logs_validation_errors ON logs.entries(created_at DESC, service)
    WHERE metadata->>'error_type' IN ('validation_error', 'security_violation');
CREATE INDEX idx_logs_error_type ON logs.entries USING GIN((metadata->>'error_type'));
```

### Existing Table Enhancement

```sql
-- Add to logs.entries if not already present (for correlation tracking)
-- Verify request_id is stored in metadata.request_id
-- Verify error_type is stored in metadata.error_type
```

---

## API Endpoints

### Dashboard Stats (Real-Time Validation Metrics)

**Endpoint:** `GET /api/logs/dashboard/stats`

**Query Parameters:**
- `service` (optional): Filter by service (default: all)
- `time_range` (optional): last_5m, last_hour, last_24h (default: last_hour)

**Response:**
```json
{
  "total_errors": 1250,
  "error_rate_percent": 2.5,
  "errors_by_type": {
    "validation_error": 800,
    "security_violation": 350,
    "other": 100
  },
  "errors_by_service": {
    "review": 900,
    "portal": 250,
    "logs": 100
  },
  "time_window": {
    "start": "2025-10-26T10:00:00Z",
    "end": "2025-10-26T11:00:00Z"
  }
}
```

### Top Validation Errors

**Endpoint:** `GET /api/logs/validations/top-errors`

**Query Parameters:**
- `service` (optional): Filter by service
- `limit` (optional): Max results (default: 10, max: 50)
- `days` (optional): Look back days (default: 7)

**Response:**
```json
{
  "errors": [
    {
      "error_type": "validation_error",
      "message": "code exceeds maximum size",
      "count": 245,
      "last_occurrence": "2025-10-26T10:45:00Z",
      "affected_services": ["review"]
    }
  ]
}
```

### Error Rate Trends

**Endpoint:** `GET /api/logs/validations/trends`

**Query Parameters:**
- `service` (optional): Filter by service
- `days` (optional): Look back days (default: 7)
- `interval` (optional): hourly, daily (default: hourly)

**Response:**
```json
{
  "trend": [
    {
      "timestamp": "2025-10-26T00:00:00Z",
      "error_count": 120,
      "error_rate_percent": 0.5,
      "by_type": {
        "validation_error": 80,
        "security_violation": 40
      }
    }
  ],
  "summary": {
    "total_24h": 2560,
    "avg_per_hour": 106.7,
    "trend_direction": "increasing"
  }
}
```

### Export Logs

**Endpoint:** `GET /api/logs/export`

**Query Parameters:**
- `format`: json, csv (required)
- `service` (optional): Filter by service
- `error_type` (optional): validation_error, security_violation, etc.
- `start`: ISO datetime (optional, default: 7 days ago)
- `end`: ISO datetime (optional, default: now)

**Response:** File download (JSON or CSV)

### Alert Config Management

**Endpoints:**
- `POST /api/logs/alert-config` - Create config
- `GET /api/logs/alert-config/:service` - Get config
- `PUT /api/logs/alert-config/:service` - Update config
- `GET /api/logs/alert-events` - Get triggered alerts

**POST Body:**
```json
{
  "service": "review",
  "error_threshold_per_min": 10,
  "warning_threshold_per_min": 5,
  "alert_email": "admin@example.com",
  "alert_webhook_url": "https://webhook.example.com/alerts",
  "enabled": true
}
```

---

## Portal Dashboard UI

### Page: `/portal/dashboard/logs`

**Components:**
1. **Header**
   - Page title: "Logs Dashboard"
   - Time range selector (Last 5m, 1h, 24h, 7d, Custom)
   - Service filter dropdown
   - Error type filter (All, Validation Errors, Security Violations)
   - Refresh button

2. **Stats Cards** (Real-time)
   - Total Errors (last period)
   - Error Rate (%)
   - Most Common Error
   - Alerts Triggered

3. **Error Breakdown Chart**
   - Pie chart: Errors by type
   - Bar chart: Errors by service

4. **Error Trend Graph**
   - Line chart: Error rate over time
   - X-axis: Time
   - Y-axis: Error count & rate %

5. **Top Errors Table**
   - Error message
   - Frequency
   - Last occurrence
   - Affected services

6. **Recent Alerts**
   - Alert timestamp
   - Service
   - Threshold that triggered
   - Error count

7. **Admin Section** (Expandable)
   - Service selector
   - Error threshold slider (1-100 per min)
   - Warning threshold slider
   - Alert email input
   - Alert webhook URL input
   - Save button
   - Status message

8. **Actions**
   - Export logs button (opens dialog for format/date selection)
   - View detailed logs button (links to logs page with filters)

**Tech Stack:**
- Template: Templ (Go)
- Styling: TailwindCSS + DaisyUI
- Interactivity: HTMX + Alpine.js
- Charts: Chart.js or similar (via CDN)
- Real-time updates: WebSocket or polling (HTMX hx-trigger="every 30s")

---

## Testing Requirements

### Unit Tests (70%+ coverage)

1. **Database Layer Tests**
   - AlertConfigRepository CRUD operations
   - AlertEventRepository create and query
   - Index efficiency tests
   - Query performance benchmarks

2. **Service Layer Tests**
   - ValidationAggregation: top errors, trends
   - AlertThreshold: create, update, retrieve
   - LogExport: JSON/CSV formatting
   - Error boundary conditions

3. **Handler Tests**
   - Dashboard stats endpoint (various time ranges)
   - Top errors endpoint (filters, limits)
   - Trends endpoint (date ranges)
   - Export endpoint (both formats)
   - Alert config CRUD
   - Error handling

### Integration Tests (90%+ critical path)

1. **Dashboard Workflow**
   - Create logs → Retrieve stats → Display on dashboard
   - Filter by service → Verify results
   - Filter by error type → Verify results
   - Time range selections → Verify date queries

2. **Alert Workflow**
   - Create alert config → Save to database
   - Log errors → Trigger alert if threshold exceeded
   - Retrieve alert events → Display in UI
   - Update config → Verify changes apply

3. **Export Workflow**
   - Export as JSON → Parse and validate structure
   - Export as CSV → Verify headers and format
   - Export with filters → Verify only matching logs included

### Manual Testing Checklist

- [ ] Dashboard loads without errors
- [ ] Time range filters work correctly
- [ ] Service filter updates results in real-time
- [ ] Error type filter updates results
- [ ] Stats cards display correct numbers
- [ ] Charts render and update
- [ ] Top errors table sorts by frequency
- [ ] Admin section loads alert thresholds
- [ ] Can update alert thresholds
- [ ] Export button works for JSON and CSV
- [ ] Exported files contain correct data
- [ ] Recent alerts display current events
- [ ] No console errors or warnings
- [ ] Responsive design works on mobile/tablet
- [ ] WebSocket updates (or polling) work

---

## References

- **ARCHITECTURE.md:** Service architecture, mental models, layering
- **DevsmithTDD.md:** TDD workflow (RED → GREEN → REFACTOR)
- **Issue #21:** Input Validation & Sanitization (generates logs we're analyzing)
- **copilot-instructions.md:** Development workflow and standards

---

## TDD Workflow

### RED Phase: Write Failing Tests FIRST
1. Database layer tests (migrations not applied yet)
2. Service layer tests (services don't exist)
3. Handler tests (endpoints not implemented)
4. UI tests (template not created)

**Commit:** `test(logs): add failing tests for production enhancements (RED phase)`

### GREEN Phase: Implement to Pass Tests
1. Create migrations
2. Implement repositories
3. Implement services
4. Implement handlers
5. Create UI templates

**Commits:** Separate commits for each layer:
- `feat(logs): add alert config schema (GREEN phase)`
- `feat(logs): implement aggregation services (GREEN phase)`
- `feat(logs): implement API endpoints (GREEN phase)`
- `feat(portal): add logs dashboard UI (GREEN phase)`

### REFACTOR Phase: Improve Code Quality
- Extract constants
- Add documentation
- Optimize queries
- Improve error handling
- Add comments

**Commits:** 
- `refactor(logs): optimize aggregation queries`
- `refactor(portal): improve dashboard UX`

---

## Success Criteria

When this issue is complete:

1. ✅ Dashboard shows real-time validation error stats
2. ✅ Can see error trends over time
3. ✅ Can identify top errors by frequency
4. ✅ Admin can configure alert thresholds per service
5. ✅ Alerts trigger when error rates spike
6. ✅ Can export logs for audit/compliance
7. ✅ 70%+ unit test coverage
8. ✅ All acceptance criteria met
9. ✅ No hardcoded values
10. ✅ Clean git history with separate RED/GREEN/REFACTOR commits

---

## Notes

- **Phase 2 (Future):** Email/webhook notifications when alerts trigger
- **Phase 3 (Future):** Machine learning for anomaly detection
- **Phase 4 (Future):** Custom alert rule builder
- Alert thresholds stored in database (configurable via admin UI, not hardcoded)
- Dashboard updates every 30s (via HTMX polling or WebSocket)
- Export respects date range and service filters
