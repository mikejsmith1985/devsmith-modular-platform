# Monitoring & Alerting Dashboard Implementation Plan

## Overview
Implement comprehensive monitoring and alerting that visualizes through the logs app to catch issues like the 400 error before they impact users.

## 1. 📊 Monitoring Architecture

### Data Collection Points
```go
// internal/middleware/monitoring.go
type APIMetrics struct {
    Timestamp     time.Time
    Method        string
    Endpoint      string
    StatusCode    int
    ResponseTime  time.Duration
    PayloadSize   int64
    UserID        string
    ErrorType     string
    ErrorMessage  string
}
```

### Storage Strategy
- **Time-series data** in PostgreSQL with efficient indexing
- **Real-time metrics** in Redis for dashboard updates
- **Historical data** retained for 30 days (configurable)
- **Alert state** tracked in dedicated tables

## 2. 🔍 Metrics to Track

### API Performance Metrics
- **Request Rate**: Requests per minute by endpoint
- **Error Rate**: 400/500 errors per minute
- **Response Time**: P50, P95, P99 percentiles
- **Payload Validation Failures**: Field-level error analysis
- **Authentication Failures**: Login/token validation issues

### Service Health Metrics  
- **Service Availability**: Health check response rates
- **Resource Usage**: CPU, memory, disk usage
- **Database Performance**: Query times, connection pool status
- **Container Health**: Docker container status and restarts

### Business Metrics
- **User Activity**: Active sessions, feature usage
- **Code Analysis Volume**: Reviews per hour, analysis types
- **Error Patterns**: Most common API errors

## 3. 🚨 Alert Thresholds

### Critical Alerts (Immediate Action)
```yaml
alerts:
  api_500_rate:
    threshold: 1.0  # per minute
    action: "Send immediate notification + log to ERROR_LOG.md"
  
  service_down:
    threshold: 2  # consecutive failed health checks
    action: "Attempt auto-restart + immediate notification"
  
  session_id_mismatch:
    threshold: 0  # Any occurrence
    action: "Critical alert + stop deployments"
```

### Warning Alerts (Monitor Closely)  
```yaml
  api_400_rate:
    threshold: 5.0  # per minute for 3 consecutive minutes
    action: "Warning notification + trend analysis"
  
  slow_response_time:
    threshold: 2000  # ms for P95
    action: "Performance investigation notification"
  
  payload_validation_spike:
    threshold: 10.0  # validation failures per minute
    action: "Field analysis report + developer notification"
```

## 4. 📈 Dashboard Features

### Real-time Overview Page
```html
<!-- logs app: /monitoring/overview -->
<div class="monitoring-grid">
  <div class="metric-card error-rate">
    <h3>API Error Rate</h3>
    <canvas id="error-rate-chart"></canvas>
    <div class="threshold-line" data-threshold="5.0"></div>
  </div>
  
  <div class="metric-card response-times">
    <h3>Response Times</h3>
    <canvas id="response-time-heatmap"></canvas>
  </div>
  
  <div class="metric-card service-health">
    <h3>Service Health</h3>
    <div class="service-grid">
      <div class="service-status portal" data-status="healthy"></div>
      <div class="service-status review" data-status="healthy"></div>
      <div class="service-status logs" data-status="healthy"></div>
      <div class="service-status analytics" data-status="warning"></div>
    </div>
  </div>
</div>
```

### Alert Management Page
```html
<!-- logs app: /monitoring/alerts -->
<div class="alerts-dashboard">
  <div class="active-alerts">
    <h2>Active Alerts (2)</h2>
    <div class="alert critical">
      <span class="alert-icon">🚨</span>
      <span class="alert-message">Review API: 400 error rate > 5.0/min</span>
      <span class="alert-time">2 minutes ago</span>
      <button class="acknowledge-btn">Acknowledge</button>
    </div>
  </div>
  
  <div class="alert-history">
    <h2>Alert History (24h)</h2>
    <!-- Timeline of resolved alerts -->
  </div>
</div>
```

### API Endpoint Analysis Page
```html
<!-- logs app: /monitoring/endpoints -->
<table class="endpoints-table">
  <thead>
    <tr>
      <th>Endpoint</th>
      <th>Request Rate</th>
      <th>Error Rate</th>
      <th>Avg Response Time</th>
      <th>Status</th>
    </tr>
  </thead>
  <tbody>
    <tr class="endpoint-row" data-status="healthy">
      <td>/api/review/modes/preview</td>
      <td>12.3/min</td>
      <td>0.2%</td>
      <td>145ms</td>
      <td><span class="status-badge healthy">Healthy</span></td>
    </tr>
    <tr class="endpoint-row" data-status="warning">
      <td>/api/review/modes/critical</td>
      <td>3.1/min</td>
      <td>8.1%</td>
      <td>2,340ms</td>
      <td><span class="status-badge warning">Slow</span></td>
    </tr>
  </tbody>
</table>
```

## 5. 🔧 Implementation Steps

### Step 1: Add Monitoring Middleware (3-4 hours)
```go
// internal/middleware/metrics.go
func MetricsMiddleware(metricsClient *MetricsClient) gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        metrics := APIMetrics{
            Timestamp:    param.TimeStamp,
            Method:       param.Method,
            Endpoint:     param.Path,
            StatusCode:   param.StatusCode,
            ResponseTime: param.Latency,
            UserID:       getUserID(param.Keys),
        }
        
        // Store metrics asynchronously
        go metricsClient.RecordAPICall(metrics)
        
        return param.DefaultFormat
    })
}
```

### Step 2: Create Metrics Storage (2-3 hours)
```sql
-- Database schema for metrics
CREATE TABLE monitoring.api_metrics (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL,
    method VARCHAR(10) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    status_code INT NOT NULL,
    response_time_ms INT NOT NULL,
    payload_size_bytes BIGINT,
    user_id VARCHAR(50),
    error_type VARCHAR(100),
    error_message TEXT
);

CREATE TABLE monitoring.alerts (
    id SERIAL PRIMARY KEY,
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    threshold_value DECIMAL,
    actual_value DECIMAL,
    triggered_at TIMESTAMPTZ DEFAULT NOW(),
    acknowledged_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ
);

-- Indexes for performance
CREATE INDEX idx_api_metrics_timestamp ON monitoring.api_metrics(timestamp DESC);
CREATE INDEX idx_api_metrics_endpoint ON monitoring.api_metrics(endpoint, timestamp DESC);
CREATE INDEX idx_api_metrics_status ON monitoring.api_metrics(status_code, timestamp DESC);
```

### Step 3: Build Alert Engine (4-5 hours)
```go
// internal/monitoring/alert_engine.go
type AlertEngine struct {
    rules     []AlertRule
    storage   AlertStorage
    notifier  NotificationService
}

type AlertRule struct {
    Name        string
    Condition   string  // "api_400_rate > 5.0"
    Window      time.Duration
    Threshold   float64
    Severity    AlertSeverity
}

func (e *AlertEngine) EvaluateRules(ctx context.Context) {
    for _, rule := range e.rules {
        current := e.calculateMetric(rule.Condition, rule.Window)
        if current > rule.Threshold {
            e.triggerAlert(rule, current)
        }
    }
}
```

### Step 4: Create Dashboard Frontend (6-8 hours)
```typescript
// logs frontend: components/MonitoringDashboard.tsx
interface MetricData {
  timestamp: string;
  value: number;
  threshold?: number;
}

function ErrorRateChart({ data }: { data: MetricData[] }) {
  return (
    <div className="chart-container">
      <canvas ref={chartRef}></canvas>
      <div className="threshold-line" style={{top: `${thresholdPosition}px`}}>
        Alert Threshold: 5.0/min
      </div>
    </div>
  );
}
```

### Step 5: Real-time Updates (2-3 hours)
```go
// WebSocket endpoint for real-time metrics
func (h *MonitoringHandler) MetricsWebSocket(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        metrics := h.getLatestMetrics()
        conn.WriteJSON(metrics)
    }
}
```

## 6. 🎯 Success Metrics

### Monitoring Effectiveness
- **Alert Accuracy**: 90% of alerts indicate real issues
- **Mean Time to Detection**: < 5 minutes for critical issues
- **False Positive Rate**: < 10% of alerts
- **Dashboard Response Time**: < 500ms for real-time updates

### Issue Prevention
- **API Contract Issues**: 0 occurrences (like session_id mismatch)
- **Performance Degradation**: Detected before user impact
- **Service Outages**: Auto-recovery success rate > 80%
- **Error Rate Trends**: Early detection of increasing error patterns

## 7. 🚀 Deployment Plan

### Development Phase (Week 1)
- Set up metrics collection middleware
- Create database schema and storage layer  
- Build basic dashboard with error rate charts
- Implement simple alerting rules

### Testing Phase (Week 2)  
- Add comprehensive alert rules
- Test alert accuracy and false positive rates
- Performance test dashboard with high metric volume
- Validate real-time update performance

### Production Phase (Week 3)
- Deploy monitoring to production
- Configure alert notifications
- Train team on dashboard usage
- Monitor monitoring system performance

## 8. 🛠️ Development Commands

```bash
# Run monitoring tests
go test ./internal/monitoring/... -v

# Start monitoring dashboard locally
cd logs && npm run dev

# Test alert engine
go test ./internal/monitoring/alert_engine_test.go -v

# Generate test metrics
./scripts/generate-test-metrics.sh

# View monitoring logs
docker-compose logs logs | grep monitoring
```

## 9. 🔮 Future Enhancements

### Phase 2 Features
- **Anomaly Detection**: ML-based pattern recognition
- **Custom Dashboards**: User-configurable monitoring views
- **Mobile Alerts**: Push notifications for critical issues
- **Slack Integration**: Alert notifications to Slack channels
- **Historical Analysis**: Trends over weeks/months
- **Capacity Planning**: Resource usage predictions

### Integration Opportunities
- **CI/CD Pipeline**: Block deployments if metrics degrade
- **Load Testing**: Automated performance testing triggers
- **Documentation**: Auto-generate runbooks from alert patterns
- **Team Metrics**: Developer productivity and code quality metrics