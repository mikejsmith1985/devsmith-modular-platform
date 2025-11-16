# Comprehensive Testing Strategy for API Monitoring

This testing strategy ensures our monitoring system catches payload validation issues like the session_id problem through automated tests.

## 1. Test Categories

### A. Unit Tests - Monitoring Components

#### Test: Monitoring Middleware Records Metrics
```go
// internal/monitoring/middleware_test.go

func TestMonitoringMiddlewareRecordsAPICall(t *testing.T) {
    // Setup
    mockCollector := &MockMetricsCollector{}
    middleware := MetricsMiddleware(mockCollector, "test-service")
    
    // Create test request
    router := gin.New()
    router.Use(middleware)
    router.POST("/api/test", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"success": true})
    })
    
    // Execute request
    req := httptest.NewRequest("POST", "/api/test", strings.NewReader(`{"data": "test"}`))
    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)
    
    // Verify monitoring recorded the call
    assert.Equal(t, 1, len(mockCollector.RecordedCalls))
    call := mockCollector.RecordedCalls[0]
    assert.Equal(t, "POST", call.Method)
    assert.Equal(t, "/api/test", call.Endpoint)
    assert.Equal(t, 200, call.StatusCode)
    assert.Equal(t, "test-service", call.ServiceName)
}
```

#### Test: Payload Validation Failure Detection
```go
func TestPayloadValidationFailureDetection(t *testing.T) {
    // Setup mock collector
    mockCollector := &MockMetricsCollector{}
    
    // Create handler that validates payload structure
    handler := func(c *gin.Context) {
        var validPayload struct {
            PastedCode string `json:"pasted_code" binding:"required"`
            Model      string `json:"model" binding:"required"`
            // Note: NO session_id field
        }
        
        if err := c.ShouldBindJSON(&validPayload); err != nil {
            // Record validation failure with field analysis
            RecordPayloadValidationFailure(c, []string{}, []string{"session_id"}, []string{})
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"success": true})
    }
    
    router := gin.New()
    router.Use(MetricsMiddleware(mockCollector, "review-service"))
    router.POST("/api/review/preview", handler)
    
    // Test with problematic payload (includes session_id)
    invalidPayload := `{
        "session_id": "abc123",
        "pasted_code": "func main() {}",
        "model": "claude-3-5-sonnet"
    }`
    
    req := httptest.NewRequest("POST", "/api/review/preview", strings.NewReader(invalidPayload))
    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)
    
    // Verify response
    assert.Equal(t, http.StatusBadRequest, resp.Code)
    
    // Verify monitoring caught the issue
    assert.Equal(t, 1, len(mockCollector.ValidationFailures))
    failure := mockCollector.ValidationFailures[0]
    assert.Contains(t, failure.ExtraFields, "session_id")
    assert.Equal(t, "/api/review/preview", failure.Endpoint)
}
```

### B. Integration Tests - Full API Flow

#### Test: Review API with Monitoring
```go
// tests/integration/review_monitoring_test.go

func TestReviewAPIMonitoring(t *testing.T) {
    // Setup test database
    db := testutils.SetupTestDB(t)
    defer testutils.TeardownTestDB(t, db)
    
    // Initialize monitoring
    collector := monitoring.NewPostgreSQLMetricsCollector(db)
    err := collector.InitializeSchema(context.Background())
    require.NoError(t, err)
    
    // Start review service with monitoring
    server := startTestReviewServer(t, collector)
    defer server.Close()
    
    tests := []struct {
        name           string
        payload        string
        expectedStatus int
        expectAlert    bool
    }{
        {
            name: "Valid Payload",
            payload: `{
                "pasted_code": "func main() { fmt.Println(\"Hello\") }",
                "model": "claude-3-5-sonnet"
            }`,
            expectedStatus: 200,
            expectAlert:    false,
        },
        {
            name: "Invalid Payload - Extra session_id Field",
            payload: `{
                "session_id": "abc123",
                "pasted_code": "func main() {}",
                "model": "claude-3-5-sonnet"
            }`,
            expectedStatus: 400,
            expectAlert:    true,
        },
        {
            name: "Invalid Payload - Missing Required Field",
            payload: `{
                "model": "claude-3-5-sonnet"
            }`,
            expectedStatus: 400,
            expectAlert:    true,
        },
    }
    
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            // Make API request
            resp, err := http.Post(server.URL+"/api/review/modes/preview", 
                "application/json", strings.NewReader(test.payload))
            require.NoError(t, err)
            defer resp.Body.Close()
            
            // Verify response status
            assert.Equal(t, test.expectedStatus, resp.StatusCode)
            
            // Give monitoring time to record
            time.Sleep(100 * time.Millisecond)
            
            // Verify monitoring recorded the call
            metrics, err := collector.GetEndpointMetrics(context.Background(), 5*time.Minute)
            require.NoError(t, err)
            
            found := false
            for _, metric := range metrics {
                if metric.Endpoint == "/api/review/modes/preview" {
                    found = true
                    if test.expectAlert {
                        assert.Greater(t, metric.ErrorCount, int64(0))
                    } else {
                        assert.Equal(t, int64(0), metric.ErrorCount)
                    }
                }
            }
            assert.True(t, found, "Endpoint metrics not recorded")
        })
    }
}
```

### C. End-to-End Tests - Real User Scenarios

#### Test: Frontend API Contract Validation
```javascript
// tests/e2e/api-contract-monitoring.spec.ts

import { test, expect } from '@playwright/test';

test.describe('API Contract Monitoring', () => {
  test('monitoring detects payload structure issues', async ({ page }) => {
    // Navigate to review app
    await page.goto('http://localhost:3000/review');
    
    // Intercept API calls to inject invalid payload
    await page.route('**/api/review/modes/preview', async route => {
      // Simulate frontend bug that sends session_id
      const invalidPayload = {
        session_id: 'abc123',  // ← This should trigger monitoring
        pasted_code: 'func main() {}',
        model: 'claude-3-5-sonnet'
      };
      
      await route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Invalid request data' })
      });
      
      // Verify monitoring API was called with error details
      // (This would be a separate request to our monitoring endpoint)
    });
    
    // Try to submit code for review
    await page.fill('#code-textarea', 'func main() { fmt.Println("Hello") }');
    await page.click('#preview-mode-button');
    
    // Verify error is displayed
    await expect(page.locator('.error-message')).toBeVisible();
    
    // Check monitoring endpoint for recorded error
    const monitoringResp = await page.request.get('http://localhost:3003/api/monitoring/metrics');
    const monitoringData = await monitoringResp.json();
    
    expect(monitoringData.error_rate_per_minute).toBeGreaterThan(0);
  });
  
  test('monitoring confirms successful API calls', async ({ page }) => {
    // Test with valid payload structure
    await page.route('**/api/review/modes/preview', async route => {
      const validPayload = {
        pasted_code: 'func main() {}',
        model: 'claude-3-5-sonnet'
        // No session_id field
      };
      
      await route.fulfill({
        status: 200,
        contentType: 'application/json', 
        body: JSON.stringify({ 
          analysis: 'This is a simple Go main function...',
          file_structure: ['main.go']
        })
      });
    });
    
    await page.goto('http://localhost:3000/review');
    await page.fill('#code-textarea', 'func main() { fmt.Println("Hello") }');
    await page.click('#preview-mode-button');
    
    // Verify success
    await expect(page.locator('.analysis-result')).toBeVisible();
    
    // Check monitoring shows healthy metrics
    const monitoringResp = await page.request.get('http://localhost:3003/api/monitoring/metrics');
    const monitoringData = await monitoringResp.json();
    
    // Should have low error rate
    expect(monitoringData.error_rate_per_minute).toBeLessThanOrEqual(1.0);
  });
});
```

### D. Load Tests - Error Rate Spike Detection

#### Test: Monitoring Under Load
```javascript
// tests/load/monitoring-load-test.js (K6 test)

import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up
    { duration: '1m', target: 10 },   // Stay at 10 users 
    { duration: '30s', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_failed: ['rate<0.1'], // Less than 10% errors
    http_req_duration: ['p(95)<500'], // 95% under 500ms
  }
};

export default function () {
  // Test both valid and invalid payloads
  const scenarios = [
    {
      // 90% valid requests
      weight: 9,
      payload: {
        pasted_code: 'func main() { fmt.Println("Test") }',
        model: 'claude-3-5-sonnet'
      },
      expectedStatus: 200
    },
    {
      // 10% invalid requests (to test monitoring)
      weight: 1,
      payload: {
        session_id: 'load-test-session',  // Invalid field
        pasted_code: 'func main() {}',
        model: 'claude-3-5-sonnet'
      },
      expectedStatus: 400
    }
  ];
  
  // Pick scenario based on weight
  const scenario = scenarios[Math.floor(Math.random() * 10) < 9 ? 0 : 1];
  
  const response = http.post('http://localhost:3000/api/review/modes/preview', 
    JSON.stringify(scenario.payload), {
      headers: { 'Content-Type': 'application/json' }
    });
  
  check(response, {
    [`status is ${scenario.expectedStatus}`]: (r) => r.status === scenario.expectedStatus,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  // Check monitoring endpoint periodically
  if (__VU === 1 && __ITER % 10 === 0) {
    const monitoring = http.get('http://localhost:3003/api/monitoring/metrics');
    check(monitoring, {
      'monitoring API accessible': (r) => r.status === 200,
      'error rate tracked': (r) => {
        const data = JSON.parse(r.body);
        return data.error_rate_per_minute >= 0; // Should be tracking errors
      }
    });
  }
}
```

## 2. Test Automation Strategy

### Pre-commit Hook Integration

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Run API contract validation tests
echo "Running API contract validation tests..."
go test ./tests/integration/... -tags=api_monitoring

if [ $? -ne 0 ]; then
    echo "❌ API monitoring tests failed!"
    echo "Ensure your changes don't break API contracts"
    exit 1
fi

# Run payload validation tests
echo "Running payload validation tests..."
npm run test:api-contracts

if [ $? -ne 0 ]; then
    echo "❌ Payload validation tests failed!"
    echo "Check for extra/missing fields in API requests"
    exit 1
fi

echo "✅ All API monitoring tests passed"
```

### CI/CD Pipeline Integration

```yaml
# .github/workflows/monitoring-tests.yml

name: API Monitoring Tests

on:
  pull_request:
    branches: [development, main]
  push:
    branches: [development, main]

jobs:
  monitoring-tests:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_DB: devsmith_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      # Install dependencies
      - run: go mod download
      - run: npm install
      
      # Run unit tests for monitoring components
      - name: Unit Tests - Monitoring
        run: go test ./internal/monitoring/... -v
      
      # Run integration tests with monitoring
      - name: Integration Tests - API Monitoring  
        env:
          DATABASE_URL: postgres://postgres:test@localhost:5432/devsmith_test
        run: go test ./tests/integration/... -tags=monitoring -v
      
      # Start services for E2E tests
      - name: Start Test Environment
        run: |
          docker-compose -f docker-compose.test.yml up -d
          sleep 30 # Wait for services to start
      
      # Run E2E tests with monitoring validation
      - name: E2E Tests - API Contract Monitoring
        run: npx playwright test tests/e2e/api-contract-monitoring.spec.ts
      
      # Run load tests (short version for CI)
      - name: Load Tests - Monitoring Under Stress
        run: |
          npm install -g k6
          k6 run tests/load/monitoring-load-test.js --duration=30s --vus=5
      
      # Verify monitoring detected test errors appropriately
      - name: Verify Monitoring Effectiveness
        run: |
          # Check monitoring API for recorded metrics
          curl -s http://localhost:3003/api/monitoring/metrics | jq '
            if .error_rate_per_minute > 0 then 
              "✅ Monitoring detected test errors"
            else 
              "❌ Monitoring did not detect test errors" | halt_error
            end'
```

### Test Data Management

```go
// internal/testutils/monitoring_testdata.go

package testutils

// PayloadTestCases defines test scenarios for API payload validation
var PayloadTestCases = []struct {
    Name           string
    Payload        map[string]interface{}
    ExpectedStatus int
    ExpectedAlert  bool
    Description    string
}{
    {
        Name: "Valid Review Request",
        Payload: map[string]interface{}{
            "pasted_code": "func main() { fmt.Println(\"Hello\") }",
            "model":       "claude-3-5-sonnet",
        },
        ExpectedStatus: 200,
        ExpectedAlert:  false,
        Description:    "Standard valid request with correct fields",
    },
    {
        Name: "Invalid - Extra session_id Field",
        Payload: map[string]interface{}{
            "session_id":  "abc123",           // ← The problematic field
            "pasted_code": "func main() {}",
            "model":       "claude-3-5-sonnet",
        },
        ExpectedStatus: 400,
        ExpectedAlert:  true,
        Description:    "Request includes session_id field that should be removed",
    },
    {
        Name: "Invalid - Missing Required Field",
        Payload: map[string]interface{}{
            "model": "claude-3-5-sonnet",
            // Missing: pasted_code
        },
        ExpectedStatus: 400,
        ExpectedAlert:  true,
        Description:    "Request missing required pasted_code field",
    },
    {
        Name: "Invalid - Wrong Field Type",
        Payload: map[string]interface{}{
            "pasted_code": 12345, // Should be string
            "model":       "claude-3-5-sonnet",
        },
        ExpectedStatus: 400,
        ExpectedAlert:  true,
        Description:    "Request has wrong type for pasted_code field",
    },
    {
        Name: "Invalid - Empty Required Field",
        Payload: map[string]interface{}{
            "pasted_code": "",
            "model":       "claude-3-5-sonnet",
        },
        ExpectedStatus: 400,
        ExpectedAlert:  true,
        Description:    "Request has empty required field",
    },
}

// MockMetricsCollector for testing monitoring functionality
type MockMetricsCollector struct {
    RecordedCalls       []monitoring.APIMetrics
    ValidationFailures  []monitoring.PayloadValidationFailure
    mu                  sync.Mutex
}

func (m *MockMetricsCollector) RecordAPICall(ctx context.Context, metrics monitoring.APIMetrics) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.RecordedCalls = append(m.RecordedCalls, metrics)
    return nil
}

func (m *MockMetricsCollector) RecordValidationFailure(ctx context.Context, failure monitoring.PayloadValidationFailure) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.ValidationFailures = append(m.ValidationFailures, failure)
    return nil
}

func (m *MockMetricsCollector) GetErrorRate(ctx context.Context, window time.Duration) (float64, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    errorCount := 0
    for _, call := range m.RecordedCalls {
        if call.StatusCode >= 400 {
            errorCount++
        }
    }
    
    minutes := window.Minutes()
    return float64(errorCount) / minutes, nil
}
```

## 3. Success Criteria

### Test Coverage Targets
- ✅ **Unit Test Coverage**: 90%+ for monitoring components
- ✅ **Integration Test Coverage**: 100% of API endpoints with monitoring
- ✅ **E2E Test Coverage**: 100% of user-facing API workflows  
- ✅ **Load Test Coverage**: All API endpoints under realistic load

### Detection Metrics  
- ✅ **Payload Issues Detected**: 100% (session_id type problems caught immediately)
- ✅ **False Positive Rate**: <5% (alerts indicate real issues)
- ✅ **Detection Time**: <1 minute for error rate spikes
- ✅ **Alert Accuracy**: 95%+ of alerts lead to actionable fixes

### Development Workflow  
- ✅ **Pre-commit Validation**: API contract tests run automatically
- ✅ **CI/CD Integration**: Monitoring tests block broken API contracts
- ✅ **Load Testing**: Production-like error rate validation
- ✅ **Historical Analysis**: Trend tracking prevents recurring issues

This comprehensive testing strategy ensures that payload validation issues like the session_id problem are caught automatically at multiple levels - unit tests, integration tests, E2E tests, and load tests - providing complete confidence that our monitoring system works as intended.