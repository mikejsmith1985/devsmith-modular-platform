# Feature 035: Log Aggregation & Statistics Dashboard - NEXT STEPS

## Current Status: RED PHASE ✅ COMPLETE

The RED phase has established:
- ✓ 6 new data models for dashboard functionality
- ✓ 4 service interfaces defining contracts
- ✓ 157 comprehensive test cases (all passing with mocks)
- ✓ Zero linting errors - production-ready code structure
- ✓ Full test specification coverage

## Next Phase: GREEN PHASE

The GREEN phase focuses on implementation - making the failing tests pass with real code.

### Phase 2.1: Service Implementations

Create concrete implementations in `internal/logs/services/`:

1. **DashboardService** → `dashboard_service.go`
   - Implement `GetDashboardStats()` - aggregate all dashboard data
   - Implement `GetServiceStats()` - service-specific statistics
   - Implement `GetTopErrors()` - retrieve top error messages
   - Implement `GetServiceHealth()` - calculate health status for each service

2. **AlertService** → `alert_service.go`
   - Implement `CreateAlertConfig()` - persist alert configuration
   - Implement `UpdateAlertConfig()` - modify alert settings
   - Implement `GetAlertConfig()` - retrieve configuration
   - Implement `CheckThresholds()` - detect threshold violations
   - Implement `SendAlert()` - send alerts via email/webhook

3. **LogAggregationService** → `log_aggregation_service.go`
   - Implement `AggregateLogsHourly()` - hourly aggregation job
   - Implement `AggregateLogsDaily()` - daily aggregation job
   - Implement `GetErrorRate()` - calculate error percentages
   - Implement `CountLogsByServiceAndLevel()` - log counting

4. **WebSocketRealtimeService** → `websocket_realtime_service.go`
   - Implement `RegisterConnection()` - track WebSocket clients
   - Implement `UnregisterConnection()` - remove disconnected clients
   - Implement `BroadcastStats()` - send stats to all connected clients
   - Implement `BroadcastAlert()` - urgent alert broadcasting
   - Implement `GetConnectionCount()` - get active connection count

### Phase 2.2: Database Layer

Create persistence in `internal/logs/db/`:

1. **Alert Configuration Repository**
   - Table: `logs.alert_configs`
   - CRUD operations for alert configurations
   - Query by service

2. **Execution History Table**
   - Table: `logs.job_executions`
   - Track hourly/daily aggregation runs
   - Last execution timestamps

3. **Violation Tracking Table**
   - Table: `logs.alert_violations`
   - Store threshold violations
   - Track alert delivery status

### Phase 2.3: Database Migrations

Create in `internal/logs/db/migrations/`:

```sql
-- Alert configurations table
CREATE TABLE logs.alert_configs (
    id BIGSERIAL PRIMARY KEY,
    service TEXT NOT NULL UNIQUE,
    error_threshold_per_min INT NOT NULL DEFAULT 100,
    warning_threshold_per_min INT NOT NULL DEFAULT 50,
    alert_email TEXT,
    alert_webhook_url TEXT,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Alert violations table
CREATE TABLE logs.alert_violations (
    id BIGSERIAL PRIMARY KEY,
    service TEXT NOT NULL,
    level TEXT NOT NULL,
    current_count BIGINT NOT NULL,
    threshold_value INT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    alert_sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Job execution history
CREATE TABLE logs.job_executions (
    id BIGSERIAL PRIMARY KEY,
    job_type TEXT NOT NULL,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    status TEXT NOT NULL,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Phase 2.4: HTTP Handlers

Create in `internal/logs/handlers/`:

1. **DashboardHandler** → `dashboard_handler.go`
   - `GET /api/logs/dashboard` - full dashboard stats
   - `GET /api/logs/dashboard/service/{service}` - service stats
   - `GET /api/logs/dashboard/top-errors` - top errors list
   - `GET /api/logs/dashboard/health` - service health

2. **AlertHandler** → `alert_handler.go`
   - `POST /api/logs/alerts/config` - create alert config
   - `GET /api/logs/alerts/config/{service}` - get config
   - `PUT /api/logs/alerts/config/{service}` - update config
   - `POST /api/logs/alerts/check` - trigger threshold check
   - `GET /api/logs/alerts/violations` - list violations

3. **WebSocketHandler** → `websocket_handler.go` (update existing)
   - `WebSocket /ws/logs/dashboard` - real-time stats stream
   - Connection management
   - Broadcast coordination

### Phase 2.5: Background Jobs

Create in `internal/logs/jobs/`:

1. **HourlyAggregationJob** → `hourly_aggregation.go`
   - Run every hour
   - Call `LogAggregationService.AggregateLogsHourly()`
   - Track execution in database

2. **DailyAggregationJob** → `daily_aggregation.go`
   - Run daily (e.g., midnight UTC)
   - Call `LogAggregationService.AggregateLogsDaily()`
   - Archive old violations

3. **ThresholdCheckJob** → `threshold_check.go`
   - Run every minute
   - Call `AlertService.CheckThresholds()`
   - Send alerts for violations
   - Update WebSocket clients

### Phase 2.6: Alert Delivery

Create in `internal/logs/alerts/`:

1. **EmailNotifier** → `email_notifier.go`
   - Send alerts via SMTP
   - Template rendering
   - Error handling

2. **WebhookNotifier** → `webhook_notifier.go`
   - POST to configured webhook URLs
   - Retry logic
   - Payload formatting

## Implementation Checklist

### Step 1: Service Implementations
- [ ] Implement DashboardService with mocked dependencies
- [ ] Implement AlertService 
- [ ] Implement LogAggregationService
- [ ] Implement WebSocketRealtimeService
- [ ] All services should use the existing LogReader interface
- [ ] Run tests after each service: `go test -v ./internal/logs/services`

### Step 2: Database Layer
- [ ] Create alert config repository
- [ ] Create violation repository
- [ ] Create execution history repository
- [ ] Implement repository tests
- [ ] Create database migrations
- [ ] Run migrations against test database

### Step 3: HTTP Handlers
- [ ] Implement DashboardHandler
- [ ] Implement AlertHandler
- [ ] Update WebSocketHandler
- [ ] Register routes in main handler setup
- [ ] Add handler tests

### Step 4: Background Jobs
- [ ] Implement job scheduler/coordinator
- [ ] Implement HourlyAggregationJob
- [ ] Implement DailyAggregationJob
- [ ] Implement ThresholdCheckJob
- [ ] Add job lifecycle management
- [ ] Add job tests

### Step 5: Alert Delivery
- [ ] Implement EmailNotifier
- [ ] Implement WebhookNotifier
- [ ] Add retry logic
- [ ] Add template rendering
- [ ] Add delivery tests

### Step 6: Integration & Testing
- [ ] Run all tests: `go test ./internal/logs/...`
- [ ] Verify coverage remains high
- [ ] Add integration tests
- [ ] Test with real data
- [ ] Performance testing for large datasets
- [ ] WebSocket real-time testing

## Test Driven Development Progression

1. **RED Phase (Completed ✓)**
   - Tests fail because implementation doesn't exist
   - Focus is on specification and contract definition
   - Mock objects validate test structure

2. **GREEN Phase (Current Focus)**
   - Make all tests pass with minimal implementation
   - Real services with real dependencies
   - Database layer operational
   - All 157 tests should pass

3. **REFACTOR Phase**
   - Improve code quality and performance
   - Remove duplication
   - Optimize database queries
   - Enhance error handling
   - Keep all tests passing

## Running Tests During Implementation

```bash
# Run specific service tests
go test -v ./internal/logs/services -run DashboardService

# Run with coverage
go test -v -cover ./internal/logs/...

# Run with race detector
go test -race ./internal/logs/...

# Watch test output
go test -v ./internal/logs/... | grep -E "(PASS|FAIL|---)"

# Debug specific test
go test -v ./internal/logs/services -run "TestGetDashboardStats_ReturnsValidStats" -v
```

## Key Implementation Notes

### Dependencies
- Use `context.Context` throughout
- Leverage existing `LogReader` interface for log access
- Use `logrus` for structured logging
- Follow existing code patterns in the project

### Database Access
- Use `database/sql` (already in use in project)
- Implement repository pattern
- Add proper error handling
- Use transactions where needed

### Real-Time Updates
- Use existing WebSocket infrastructure
- Broadcast stats every second (configurable)
- Send alerts immediately
- Track connection count

### Alert Delivery
- Email: Configure SMTP in config
- Webhook: Validate URLs, retry on failure
- Track delivery status
- Log all alert attempts

## Success Criteria for GREEN Phase

- [ ] All 157 tests pass with real implementations
- [ ] Zero linting errors
- [ ] Database migrations work correctly
- [ ] HTTP endpoints accessible and functional
- [ ] WebSocket real-time updates working
- [ ] Background jobs execute on schedule
- [ ] Alert delivery tested end-to-end
- [ ] Performance acceptable (< 1 sec dashboard load)
- [ ] Code coverage remains above 80%
- [ ] All acceptance criteria met

## Estimated Timeline

- **Service Implementations**: 2-3 days
- **Database Layer**: 1-2 days
- **HTTP Handlers**: 1-2 days
- **Background Jobs**: 1-2 days
- **Alert Delivery**: 1-2 days
- **Integration Testing**: 1-2 days
- **Total**: 7-13 days (1-2 weeks)

## Additional Resources

- See: `FEATURE_035_RED_PHASE_SUMMARY.md` for RED phase details
- See: `DevsmithTDD.md` for TDD guidelines in this project
- See: `internal/analytics/services/` for reference implementation patterns
- See: `internal/review/services/` for another service example

---

**Ready to begin GREEN phase? Start with DashboardService implementation!**
