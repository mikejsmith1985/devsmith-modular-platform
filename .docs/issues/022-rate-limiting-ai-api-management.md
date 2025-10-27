# Issue #022: Rate Limiting & AI API Management

**Priority:** High (cost control)  
**Service:** Review  
**Sprint:** TBD  
**Status:** Not Started

---

## Summary

Implement rate limiting, request queuing, circuit breaking, and cost tracking for AI API calls to prevent excessive costs and API quota exhaustion. This feature protects the platform from runaway AI API consumption while providing visibility into usage patterns.

---

## Problem Statement

Currently, there are no controls on AI API calls in the Review service. This creates two risks:

1. **Cost Risk**: Without rate limiting, a single user could make unlimited AI requests, resulting in unexpected costs
2. **Quota Risk**: API providers (Claude, OpenAI) enforce rate limits. Uncontrolled requests could exhaust quotas and cause service outages

### Acceptance Criteria

- [x] Rate limiting middleware enforces per-user limit (10 requests/minute)
- [x] Rate limiting enforces per-IP limit for unauthenticated users
- [x] Queue system manages AI requests with FIFO ordering
- [x] Retry logic implements exponential backoff (configurable)
- [x] Circuit breaker pattern prevents cascading failures
- [x] Cost tracking stores usage per user in database
- [x] Rate limit errors return HTTP 429 (Too Many Requests)
- [x] Unit tests cover all components (70%+ coverage)
- [x] Integration tests verify end-to-end flows

---

## Architecture & Design

### 1. Rate Limiting Middleware

**Location:** `internal/review/middleware/rate_limiter.go`

**Features:**
- Redis-backed rate limiting using token bucket algorithm
- Per-user limit: 10 requests per minute (configurable via env var)
- Per-IP limit: 20 requests per minute for unauthenticated users
- Returns HTTP 429 with `Retry-After` header
- Respects existing user quotas in database

**Interface:**
```go
type RateLimiter interface {
    CheckLimit(ctx context.Context, identifier string) error
    GetRemainingQuota(ctx context.Context, identifier string) (int, time.Time, error)
}
```

**Implementation:**
```go
type RedisRateLimiter struct {
    client    redis.Client
    defaultLimit int        // 10 per minute
    windowSize  time.Duration  // 1 minute
}
```

### 2. Queue System for AI Requests

**Location:** `internal/review/queue/ai_request_queue.go`

**Features:**
- In-memory FIFO queue with Redis persistence option
- Automatic retry for failed requests
- Request deduplication (prevent duplicate large analyses)
- Priority levels (user tier, request complexity)
- Graceful shutdown with in-flight tracking

**Entities:**
```go
type AIRequest struct {
    ID              string
    UserID          int64
    ReadingMode     string
    CodeContent     string
    EnqueuedAt      time.Time
    MaxRetries      int
    CurrentAttempt  int
}

type AIResponse struct {
    RequestID   string
    Result      interface{}
    Duration    time.Duration
    TokensUsed  int
    Error       error
    CompletedAt time.Time
}
```

**Interface:**
```go
type RequestQueue interface {
    Enqueue(ctx context.Context, req *AIRequest) error
    Dequeue(ctx context.Context) (*AIRequest, error)
    MarkComplete(ctx context.Context, requestID string, response *AIResponse) error
    GetStatus(ctx context.Context, requestID string) (*RequestStatus, error)
}
```

### 3. Retry Logic with Exponential Backoff

**Location:** `internal/review/retry/backoff.go`

**Features:**
- Configurable retry strategy (max retries, initial delay, backoff multiplier)
- Exponential backoff: `delay = baseDelay × (multiplier ^ attempt)`
- Jitter to prevent thundering herd
- Context-aware (respects cancellation)
- Idempotency support (safe to retry)

**Configuration:**
```go
type RetryConfig struct {
    MaxRetries        int           // Default: 3
    InitialDelay      time.Duration // Default: 100ms
    BackoffMultiplier float64       // Default: 2.0
    MaxDelay          time.Duration // Default: 30s
    JitterFraction    float64       // Default: 0.1 (10% jitter)
}
```

**Default Values:**
- Attempt 1: 100ms
- Attempt 2: 200ms
- Attempt 3: 400ms
- Max: 30s

### 4. Circuit Breaker Pattern

**Location:** `internal/review/circuit/circuit_breaker.go`

**Features:**
- Three states: CLOSED (healthy) → OPEN (failing) → HALF_OPEN (testing)
- Automatic state transitions based on error rates
- Configurable thresholds (error count, success count)
- Metrics tracking (failures, successes, state changes)

**States:**
```
CLOSED
  ↓ (5 consecutive errors)
OPEN (reject all requests immediately)
  ↓ (after 30s timeout)
HALF_OPEN (allow 1 test request)
  ↓ (test succeeds or fails)
CLOSED or OPEN
```

**Configuration:**
```go
type CircuitBreakerConfig struct {
    OpenThreshold      int           // Errors before opening
    HalfOpenThreshold  int           // Successes to close
    Timeout            time.Duration // Duration in OPEN state
    MetricsWindow      time.Duration // For calculating error rate
}
```

### 5. Cost Tracking

**Location:** `internal/review/models/cost_tracking.go`

**Database Schema:**

```sql
CREATE TABLE reviews.ai_api_usage (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES portal.users(id),
    request_id      VARCHAR(255) UNIQUE,
    api_provider    VARCHAR(50),      -- 'claude', 'openai', 'ollama'
    model_name      VARCHAR(100),
    reading_mode    VARCHAR(20),
    input_tokens    INT,
    output_tokens   INT,
    total_cost      DECIMAL(10, 6),   -- USD
    status          VARCHAR(20),      -- 'queued', 'processing', 'success', 'failed'
    error_message   TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    INDEX (user_id, created_at),
    INDEX (status)
);

CREATE TABLE reviews.user_quotas (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES portal.users(id),
    monthly_limit   DECIMAL(10, 2),   -- USD
    monthly_used    DECIMAL(10, 6),   -- USD
    requests_limit  INT,              -- Per minute
    requests_used   INT,
    reset_date      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
```

**Tracking Service:**
```go
type CostTracker interface {
    RecordRequest(ctx context.Context, req *AIRequest) error
    RecordResponse(ctx context.Context, requestID string, resp *AIResponse) error
    GetUserUsage(ctx context.Context, userID int64, period time.Duration) (*UsageStats, error)
    CheckQuota(ctx context.Context, userID int64) (bool, error)
}

type UsageStats struct {
    TotalRequests  int
    TotalTokens    int
    TotalCost      float64
    ByProvider     map[string]*ProviderStats
    ByMode         map[string]*ModeStats
}
```

---

## Implementation Phases

### Phase 1: Core Infrastructure (Days 1-2)

1. **Rate Limiter** 
   - Redis token bucket implementation
   - Per-user and per-IP checks
   - HTTP 429 response with Retry-After header
   - Unit tests + integration test with Redis

2. **Request Queue**
   - Basic FIFO queue
   - Enqueue/dequeue operations
   - Status tracking
   - Unit tests

### Phase 2: Retry & Circuit Breaker (Days 3-4)

3. **Retry Logic**
   - Exponential backoff implementation
   - Jitter support
   - Context-aware cancellation
   - Integration with queue

4. **Circuit Breaker**
   - State machine (CLOSED/OPEN/HALF_OPEN)
   - Error rate tracking
   - Metrics collection
   - Integration with AI service calls

### Phase 3: Cost Tracking & Monitoring (Days 5-6)

5. **Cost Tracking Service**
   - Database schema and migrations
   - Request/response recording
   - Quota checking
   - Usage analytics

6. **Dashboard Integration** (Phase 3.1)
   - Display user cost dashboard
   - Quota warnings
   - Usage trends

---

## Database Migrations

**File:** `migrations/{timestamp}_create_rate_limiting_tables.sql`

```sql
-- See Schema above for full definition
-- Creates:
-- - reviews.ai_api_usage (tracks every request)
-- - reviews.user_quotas (quota enforcement)
-- - Indexes for query performance
```

---

## API Endpoints

### Rate Limit Status
```
GET /api/review/rate-limit/status
Response:
{
  "user_id": 123,
  "requests_remaining": 8,
  "requests_limit": 10,
  "window_reset_at": "2025-10-20T12:05:00Z"
}
```

### Usage Statistics
```
GET /api/review/usage/stats?period=7d
Response:
{
  "total_requests": 45,
  "total_cost": 2.34,
  "by_mode": {
    "preview": { "count": 10, "cost": 0.12 },
    "critical": { "count": 5, "cost": 1.23 }
  }
}
```

### Circuit Breaker Status
```
GET /api/review/circuit-breaker/status
Response:
{
  "state": "CLOSED",
  "failures": 2,
  "successes": 145,
  "last_failure": "2025-10-20T12:00:00Z"
}
```

---

## Environment Variables

**New Variables:**

```bash
# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=10
RATE_LIMIT_IP_REQUESTS_PER_MINUTE=20
RATE_LIMIT_WINDOW_DURATION=1m

# Queue
QUEUE_MAX_SIZE=1000
QUEUE_WORKER_COUNT=5

# Retry
RETRY_MAX_ATTEMPTS=3
RETRY_INITIAL_DELAY=100ms
RETRY_BACKOFF_MULTIPLIER=2.0
RETRY_MAX_DELAY=30s

# Circuit Breaker
CB_OPEN_THRESHOLD=5
CB_HALF_OPEN_THRESHOLD=3
CB_TIMEOUT=30s
CB_METRICS_WINDOW=5m

# Cost Tracking
COST_TRACKING_ENABLED=true
USER_MONTHLY_QUOTA=50.00
```

---

## Testing Strategy

### Unit Tests

**Rate Limiter Tests:**
- ✅ Test allowed request passes through
- ✅ Test rejected request after limit reached
- ✅ Test Retry-After header set correctly
- ✅ Test limit resets after window

**Queue Tests:**
- ✅ Test enqueue/dequeue FIFO ordering
- ✅ Test status tracking
- ✅ Test concurrent operations

**Retry Tests:**
- ✅ Test exponential backoff calculation
- ✅ Test jitter application
- ✅ Test context cancellation
- ✅ Test max retries limit

**Circuit Breaker Tests:**
- ✅ Test state transitions
- ✅ Test error rate calculation
- ✅ Test recovery after timeout

**Cost Tracker Tests:**
- ✅ Test request/response recording
- ✅ Test quota checking
- ✅ Test usage aggregation

### Integration Tests

- ✅ End-to-end request flow with rate limiting
- ✅ Queue with actual AI service calls
- ✅ Circuit breaker prevents cascading failures
- ✅ Cost tracking records all API usage
- ✅ Retry logic recovers from transient failures

---

## References

- **ARCHITECTURE.md** Section: AI/LLM Integration (line 571+)
- **Requirements.md** Section: Review Service (line 705+)
- **DevsmithTDD.md** Section: TDD Workflow
- **Security Architecture** Section: Rate Limiting (line 1529)

---

## Acceptance Criteria Checklist

Implementation Complete When:

- [ ] **Rate Limiting**
  - [ ] Redis-backed token bucket working
  - [ ] Per-user limit enforced (10 req/min)
  - [ ] Per-IP limit enforced (20 req/min)
  - [ ] HTTP 429 returned with Retry-After
  - [ ] Tests: 70%+ coverage

- [ ] **Request Queue**
  - [ ] FIFO queue operational
  - [ ] Enqueue/dequeue working
  - [ ] Status tracking functional
  - [ ] Concurrent operations safe
  - [ ] Tests: 70%+ coverage

- [ ] **Retry Logic**
  - [ ] Exponential backoff implemented
  - [ ] Jitter applied correctly
  - [ ] Max retries enforced
  - [ ] Context cancellation respected
  - [ ] Tests: 70%+ coverage

- [ ] **Circuit Breaker**
  - [ ] State transitions working (CLOSED/OPEN/HALF_OPEN)
  - [ ] Error thresholds enforced
  - [ ] Metrics tracked accurately
  - [ ] Integration with AI service
  - [ ] Tests: 70%+ coverage

- [ ] **Cost Tracking**
  - [ ] Database schema created
  - [ ] Request/response recording working
  - [ ] Usage stats accurate
  - [ ] Quota enforcement functional
  - [ ] Tests: 70%+ coverage

- [ ] **API Endpoints**
  - [ ] Rate limit status endpoint working
  - [ ] Usage stats endpoint working
  - [ ] Circuit breaker status endpoint working

- [ ] **Testing**
  - [ ] All unit tests passing
  - [ ] Integration tests passing
  - [ ] No hardcoded values
  - [ ] No pre-commit violations
  - [ ] Full service builds successfully

---

## TDD Workflow

### RED Phase: Write Failing Tests

Start with tests for each component:
1. `internal/review/middleware/rate_limiter_test.go` - Rate limiter tests
2. `internal/review/queue/ai_request_queue_test.go` - Queue tests
3. `internal/review/retry/backoff_test.go` - Retry logic tests
4. `internal/review/circuit/circuit_breaker_test.go` - Circuit breaker tests
5. `internal/review/services/cost_tracker_test.go` - Cost tracking tests

### GREEN Phase: Minimal Implementation

Implement each component with minimal code to pass tests:
1. `rate_limiter.go` - Token bucket implementation
2. `ai_request_queue.go` - FIFO queue
3. `backoff.go` - Exponential backoff calculator
4. `circuit_breaker.go` - State machine
5. `cost_tracker.go` - Cost tracking service

### REFACTOR Phase: Improve Quality

- Add comprehensive error messages
- Optimize Redis operations
- Add metrics/monitoring
- Improve documentation
- Extract helper methods

---

## Common Pitfalls to Avoid

❌ **WRONG:** Hardcoding rate limit values in code
✅ **CORRECT:** All limits in environment variables

❌ **WRONG:** Synchronous API calls in queue
✅ **CORRECT:** Async queue with background workers

❌ **WRONG:** Returning error strings as data
✅ **CORRECT:** Proper error types with context

❌ **WRONG:** Circuit breaker without metrics
✅ **CORRECT:** Track successes and failures

❌ **WRONG:** Cost tracking only on success
✅ **CORRECT:** Track all requests (success + failure)

---

## Success Metrics

After implementation, these metrics should be trackable:

1. **Adoption**: % of AI requests going through rate limiter
2. **Cost Control**: Average cost per user per month
3. **Reliability**: Circuit breaker state transitions over time
4. **Queue Performance**: Average queue wait time
5. **Retry Success**: % of requests succeeding after retry

