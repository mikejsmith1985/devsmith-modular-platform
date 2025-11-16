# K6 Load Test Baseline - Review Service

## Test Configuration

**Date**: 2025-11-02  
**Version**: Phase 5 - Production Ready  
**Environment**: Local Development  
**Hardware**: (To be filled after test run)

### Test Parameters
- **Virtual Users (VUs)**: 10
- **Total Iterations**: 100 (10 per VU)
- **Duration**: 2 minutes max
- **Base URL**: http://localhost:3000
- **Model**: mistral:7b-instruct

### Load Distribution (Weighted Random)
- Preview Mode: 30% (most common)
- Skim Mode: 25%
- Scan Mode: 20%
- Detailed Mode: 15%
- Critical Mode: 10% (most resource-intensive)

---

## Baseline Results

### Overall Metrics

| Metric | Target | Actual | Status |
|--------|--------|--------|--------|
| **HTTP Request Duration (P95)** | < 5000ms | TBD | ⏳ |
| **HTTP Request Duration (P99)** | < 10000ms | TBD | ⏳ |
| **HTTP Request Failure Rate** | < 10% | TBD | ⏳ |
| **Overall Error Rate** | < 10% | TBD | ⏳ |

### Reading Mode Performance

| Mode | P50 | P95 | P99 | Max | Target P95 | Status |
|------|-----|-----|-----|-----|------------|--------|
| **Preview** | TBD | TBD | TBD | TBD | < 3000ms | ⏳ |
| **Skim** | TBD | TBD | TBD | TBD | < 5000ms | ⏳ |
| **Scan** | TBD | TBD | TBD | TBD | < 4000ms | ⏳ |
| **Detailed** | TBD | TBD | TBD | TBD | < 7000ms | ⏳ |
| **Critical** | TBD | TBD | TBD | TBD | < 10000ms | ⏳ |

### Request Statistics

| Metric | Value |
|--------|-------|
| Total Requests | TBD |
| Successful Requests | TBD |
| Failed Requests | TBD |
| Requests/Second (avg) | TBD |
| Data Received | TBD |
| Data Sent | TBD |

---

## Circuit Breaker Behavior

### Observations During Load Test

**Circuit Breaker State Changes**: TBD

**Failed Requests Due to Open Circuit**: TBD

**Recovery Time**: TBD

**Expected Behavior**:
- Circuit opens after 5 consecutive failures
- Timeout: 10 seconds
- Half-open state: Allow 1 test request
- Close after 3 successful requests

**Actual Behavior**: TBD

---

## Ollama Performance

### Model: mistral:7b-instruct

| Metric | Value |
|--------|-------|
| **Concurrent Requests** | TBD |
| **Average Response Time** | TBD |
| **Max Response Time** | TBD |
| **Timeouts** | TBD |
| **Resource Usage** (CPU/RAM) | TBD |

### Observations
- TBD: Did Ollama handle concurrent requests well?
- TBD: Were there any timeouts or slowdowns?
- TBD: Did response quality degrade under load?

---

## Bottlenecks Identified

### 1. [Bottleneck Name]
**Severity**: Low/Medium/High  
**Description**: TBD  
**Recommendation**: TBD

### 2. [Bottleneck Name]
**Severity**: Low/Medium/High  
**Description**: TBD  
**Recommendation**: TBD

---

## Thresholds & Alerts

Based on baseline results, recommended production thresholds:

| Metric | Warning Threshold | Critical Threshold | Action |
|--------|------------------|-------------------|--------|
| **P95 Latency (Preview)** | > 2500ms | > 4000ms | Check Ollama, circuit breaker state |
| **P95 Latency (Critical)** | > 8000ms | > 15000ms | Scale Ollama, investigate model performance |
| **Error Rate** | > 5% | > 10% | Check circuit breaker, Ollama health |
| **Circuit Breaker Open** | N/A | State = Open | Alert on-call, check Ollama |

---

## Recommendations

### Short-Term (Next Sprint)
1. TBD: Based on test results
2. TBD: Based on test results
3. TBD: Based on test results

### Long-Term (Phase 6+)
1. **Horizontal Scaling**: Multiple Ollama instances behind load balancer
2. **Caching**: Cache common analysis results (Redis)
3. **Rate Limiting**: Per-user rate limits to prevent abuse
4. **Model Optimization**: Test smaller models for faster modes (Preview, Skim)

---

## Running the Test

### Prerequisites
```bash
# Install k6
brew install k6  # macOS
# or
sudo apt-get install k6  # Ubuntu
# or
choco install k6  # Windows

# Ensure Review service is running
docker-compose up -d review

# Ensure Ollama is running
docker-compose up -d ollama
```

### Execute Test
```bash
# Run test and see live results
k6 run tests/k6/review-load.js

# Run test and save detailed JSON report
k6 run --out json=.docs/perf/k6-results.json tests/k6/review-load.js

# Run with custom VUs/iterations
k6 run --vus 20 --iterations 200 tests/k6/review-load.js
```

### Analyze Results
```bash
# View JSON report
cat .docs/perf/k6-results.json | jq '.metrics'

# Generate HTML report (requires k6-reporter)
npm install -g k6-html-reporter
k6-html-reporter .docs/perf/k6-results.json
```

---

## Next Steps

1. ✅ Run baseline load test
2. ⏳ Fill in TBD values above
3. ⏳ Analyze bottlenecks and circuit breaker behavior
4. ⏳ Set production thresholds based on results
5. ⏳ Integrate k6 into CI/CD pipeline
6. ⏳ Monitor production metrics against baseline

---

## Appendix

### Sample k6 Output
```
(To be filled after test run)
```

### Circuit Breaker Logs
```
(To be filled after test run)
```

### Ollama Logs
```
(To be filled after test run)
```

---

**Report Generated**: 2025-11-02  
**Next Review**: After production deployment  
**Owner**: DevOps Team
