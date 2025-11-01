# k6 Performance Testing Suite

Performance testing suite for DevSmith Platform using [k6](https://k6.io/).

## Overview

This suite validates:
- **Reading Mode Performance**: Critical, Preview, Skim, Scan, Detailed modes
- **WebSocket Reliability**: Log streaming, real-time message delivery
- **Health Check Integration**: Auto-repair and monitoring performance
- **Load Capacity**: Throughput under concurrent load

## Installation

### Prerequisites
- k6 installed: https://k6.io/docs/getting-started/installation/
- DevSmith services running locally (or configure BASE_URL)
- Ollama service running (for AI analysis modes)

### Install k6

**macOS (Homebrew):**
```bash
brew install k6
```

**Linux:**
```bash
sudo apt-get install k6
```

**Windows (Chocolatey):**
```bash
choco install k6
```

Verify installation:
```bash
k6 version
```

## Test Files

### `critical_mode_load.js`
**Purpose:** Load test for Critical Mode (quality evaluation)

**What it tests:**
- HTTP POST to `/api/review/sessions/:id/modes/critical`
- Code analysis with quality scoring
- Response time under concurrent load
- JSON response validation

**Performance Targets:**
- P95 response time: < 2 seconds
- P99 response time: < 3 seconds
- Error rate: < 10%
- Minimum 50 requests during test

**Run:**
```bash
k6 run tests/k6/critical_mode_load.js
```

### `websocket_load.js`
**Purpose:** Performance test for WebSocket log streaming

**What it tests:**
- WebSocket connection establishment
- Message streaming latency
- JSON message validation
- Connection stability under load

**Performance Targets:**
- Connection time P95: < 1 second
- Message latency P95: < 500ms
- Error rate: < 10%
- Minimum 10 messages per connection

**Run:**
```bash
k6 run tests/k6/websocket_load.js
```

## Running Tests

### Run Single Test
```bash
k6 run tests/k6/critical_mode_load.js
```

### Run with Verbosity
```bash
k6 run tests/k6/critical_mode_load.js -v
```

### Custom Base URL
```bash
k6 run tests/k6/critical_mode_load.js --env BASE_URL=http://production:3000
```

### Run All k6 Tests
```bash
for test in tests/k6/*.js; do
  echo "Running $test..."
  k6 run "$test"
  echo ""
done
```

## Metrics

### Custom Metrics Tracked

**Critical Mode:**
- `errors`: Error rate (%)
- `request_duration`: Response time (ms)
- `requests`: Total request count
- `success`: Success rate (%)

**WebSocket:**
- `ws_errors`: WebSocket error rate (%)
- `ws_message_latency`: Message latency (ms)
- `ws_messages_received`: Total messages
- `ws_connection_time`: Connection establishment (ms)

### Thresholds

Tests automatically fail if thresholds are exceeded:

```
Critical Mode:
  - P95 duration < 2000ms ✓
  - P99 duration < 3000ms ✓
  - Error rate < 10% ✓
  
WebSocket:
  - Error rate < 10% ✓
  - Message latency P95 < 500ms ✓
  - Connection time P95 < 1000ms ✓
```

## Load Stages

### Standard Load Profile (Default)
```
Stage 1: 10s  → Ramp up to 5 VUs (virtual users)
Stage 2: 30s  → Ramp up to 10 VUs
Stage 3: 20s  → Sustain at 10 VUs
Stage 4: 10s  → Ramp down to 0 VUs
```

**Total Duration:** ~70 seconds per test

### High-Load Profile
```bash
k6 run tests/k6/critical_mode_load.js \
  --stage "20s:50" \
  --stage "30s:50" \
  --stage "20s:0"
```

## Integration with Health Checks

### Auto-Repair Monitoring
The k6 tests can trigger and monitor auto-repair actions:

```bash
# Run Critical Mode while health checks auto-repair
# Monitor repair frequency in health check API
curl http://localhost:3000/api/health/repairs?window=1h
```

### Expected Auto-Repair Patterns
- High error rates trigger auto-repair
- Repair events logged to logs service
- Performance metrics tracked before/after repair

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Performance Tests

on: [push]

jobs:
  k6:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
      ollama:
        image: ollama/ollama:latest
    
    steps:
      - uses: actions/checkout@v3
      - uses: grafana/setup-k6-action@v1
      
      - run: docker-compose up -d
      - run: sleep 10  # Wait for services
      
      - run: k6 run tests/k6/critical_mode_load.js
      - run: k6 run tests/k6/websocket_load.js
```

## Troubleshooting

### Connection Refused
```
ERROR WebSocket dial error: connection refused
```

**Solution:** Ensure services are running:
```bash
docker-compose ps  # Check service status
docker-compose up -d  # Start services
```

### Threshold Failures
```
ERROR Thresholds have been breached
```

**Diagnostics:**
1. Check Ollama health: `curl http://localhost:11434/api/tags`
2. Check database: `docker-compose logs postgres`
3. Monitor logs: `docker-compose logs review`

### Timeout Errors
```
ERROR Request timeout after 30s
```

**Solution:** Increase timeout in test:
```javascript
timeout: '60s'  // Increase from 30s
```

## Performance Baselines

### Expected Performance (Local Development)

| Mode | Avg (ms) | P95 (ms) | P99 (ms) |
|------|----------|----------|----------|
| Critical | 1200 | 1800 | 2500 |
| Preview | 800 | 1200 | 1600 |
| WebSocket | 250 | 400 | 600 |

### Performance Degradation Indicators

| Issue | Symptom |
|-------|---------|
| Ollama overloaded | P95 > 3000ms |
| Database slow | Error rate > 5% |
| Network issues | P95 variance > 50% |
| Memory pressure | Timeouts increase |

## Advanced Usage

### Save Results to JSON
```bash
k6 run tests/k6/critical_mode_load.js \
  --out json=results.json
```

### InfluxDB Export
```bash
k6 run tests/k6/critical_mode_load.js \
  --out influxdb=http://localhost:8086/myk6db
```

### Grafana Dashboard
View real-time metrics during test:
```bash
# Terminal 1: Start Grafana
docker run -d -p 3001:3000 grafana/grafana

# Terminal 2: Run k6 with Grafana
k6 run tests/k6/critical_mode_load.js \
  --out influxdb \
  --vus 10 --duration 5m
```

## Best Practices

1. **Run during off-peak hours** - Avoid production traffic
2. **Start small** - Test with 5-10 VUs first
3. **Ramp gradually** - Use stages to simulate real load
4. **Monitor resources** - Watch CPU/memory during tests
5. **Baseline first** - Establish baseline before optimizations
6. **Compare changes** - Run before/after code changes
7. **Document results** - Save results for trend analysis

## Maintenance

### Adding New Tests
1. Create `tests/k6/new_feature_load.js`
2. Define load stages in `options`
3. Set realistic thresholds
4. Add to this README
5. Test locally before committing

### Updating Thresholds
When performance improves:
1. Run baseline test: `k6 run tests/k6/critical_mode_load.js`
2. Note new metrics
3. Update threshold values
4. Document in commit message

## References

- [k6 Documentation](https://k6.io/docs/)
- [k6 API Reference](https://k6.io/docs/javascript-api/)
- [Performance Testing Best Practices](https://k6.io/docs/test-types/load-test/)
- [DevSmith Architecture](../../ARCHITECTURE.md)
