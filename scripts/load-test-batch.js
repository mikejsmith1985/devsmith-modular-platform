// Load Testing Script for Batch Log Ingestion
// Target: 14,000-33,000 logs/second (1M logs/hour)
// Usage: k6 run scripts/load-test-batch.js
// 
// Install k6: https://k6.io/docs/getting-started/installation/
// - Ubuntu: sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
//           echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
//           sudo apt-get update
//           sudo apt-get install k6
// - macOS: brew install k6

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const batchDuration = new Trend('batch_duration');
const logsIngested = new Counter('logs_ingested');

// Configuration from environment variables
const LOGS_API_URL = __ENV.LOGS_API_URL || 'http://localhost:8082';
const SERVICE_NAME = __ENV.SERVICE_NAME || 'load-test';

// Test configuration options
export const options = {
  scenarios: {
    // Scenario 1: Ramp up to target load (14K logs/sec)
    ramp_up: {
      executor: 'ramping-vus',
      startVUs: 1,
      stages: [
        { duration: '30s', target: 10 },   // Warm up: 10 VUs
        { duration: '1m', target: 50 },    // Ramp up to 50 VUs
        { duration: '2m', target: 100 },   // Sustain at 100 VUs
        { duration: '30s', target: 0 },    // Ramp down
      ],
      gracefulRampDown: '10s',
    },
    
    // Scenario 2: Constant load test (optional - comment out if not needed)
    // constant_load: {
    //   executor: 'constant-vus',
    //   vus: 50,
    //   duration: '5m',
    //   startTime: '4m', // Start after ramp_up finishes
    // },
    
    // Scenario 3: Spike test (optional)
    // spike_test: {
    //   executor: 'ramping-vus',
    //   startVUs: 0,
    //   stages: [
    //     { duration: '10s', target: 200 }, // Spike to 200 VUs
    //     { duration: '30s', target: 200 }, // Hold
    //     { duration: '10s', target: 0 },   // Drop
    //   ],
    //   startTime: '10m',
    // },
  },
  
  thresholds: {
    // Response time thresholds
    'http_req_duration': ['p(95)<500', 'p(99)<1000'], // 95% < 500ms, 99% < 1s
    'batch_duration': ['p(95)<500'],
    
    // Error rate threshold
    'errors': ['rate<0.01'], // Error rate < 1%
    
    // HTTP failures
    'http_req_failed': ['rate<0.01'],
  },
};

// Generate batch of log entries
function generateBatch(batchSize, vu = 0, iter = 0) {
  const entries = [];
  const timestamp = new Date().toISOString();
  const levels = ['debug', 'info', 'warn', 'error']; // Lowercase per API spec
  
  for (let i = 0; i < batchSize; i++) {
    entries.push({
      level: levels[Math.floor(Math.random() * levels.length)],
      message: `Load test message ${i} - ${Math.random().toString(36).substring(7)}`,
      service_name: SERVICE_NAME,
      context: {
        test_id: vu,
        iteration: iter,
        index: i,
        timestamp_ms: Date.now(),
      },
      timestamp: timestamp,
    });
  }
  
  return { project_slug: 'load-test', logs: entries };
}

// Test with different batch sizes
export default function () {
  const batchSizes = [100, 500, 1000]; // Test different batch sizes
  const batchSize = batchSizes[Math.floor(Math.random() * batchSizes.length)];
  
  const batch = generateBatch(batchSize, __VU, __ITER);
  const payload = JSON.stringify(batch);
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    timeout: '10s',
  };
  
  const startTime = Date.now();
  const response = http.post(`${LOGS_API_URL}/api/logs/batch`, payload, params);
  const duration = Date.now() - startTime;
  
  // Record metrics
  batchDuration.add(duration);
  
  // Check response
  const success = check(response, {
    'status is 200': (r) => r.status === 200,
    'response has accepted count': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.accepted === batchSize;
      } catch (e) {
        return false;
      }
    },
    'response time < 1s': (r) => r.timings.duration < 1000,
  });
  
  if (success) {
    logsIngested.add(batchSize);
  } else {
    errorRate.add(1);
    console.error(`Batch failed: status=${response.status}, batch_size=${batchSize}, duration=${duration}ms`);
  }
  
  // Small delay between requests (adjust based on target throughput)
  sleep(0.1); // 100ms delay = ~10 requests/sec per VU
}

// Setup function - runs once before test
export function setup() {
  console.log('🚀 Starting load test...');
  console.log(`   API URL: ${LOGS_API_URL}`);
  console.log(`   Service: ${SERVICE_NAME}`);
  console.log(`   Target: 14,000-33,000 logs/second`);
  console.log('');
  
  // Verify service is available
  const testBatch = generateBatch(1);
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  const response = http.post(`${LOGS_API_URL}/api/logs/batch`, JSON.stringify(testBatch), params);
  
  if (response.status !== 200 && response.status !== 201) { // Accept both 200 and 201
    console.error(`❌ Service validation failed: ${response.status} ${response.body}`);
    throw new Error('Service unavailable');
  }
  
  console.log('✅ Service validated successfully');
  return {};
}

// Teardown function - runs once after test
export function teardown(data) {
  console.log('');
  console.log('📊 Load test completed');
}

// Handle summary - custom output
export function handleSummary(data) {
  const logsPerSecond = data.metrics.logs_ingested.values.count / (data.state.testRunDurationMs / 1000);
  const errorPercentage = (data.metrics.errors.values.rate * 100).toFixed(2);
  
  console.log('');
  console.log('═══════════════════════════════════════════════════════════');
  console.log('📈 PERFORMANCE SUMMARY');
  console.log('═══════════════════════════════════════════════════════════');
  console.log(`Total Logs Ingested: ${data.metrics.logs_ingested.values.count.toLocaleString()}`);
  console.log(`Total Requests: ${data.metrics.http_reqs.values.count}`);
  console.log(`Throughput: ${logsPerSecond.toFixed(0)} logs/second`);
  console.log(`Error Rate: ${errorPercentage}%`);
  console.log('');
  console.log('Response Times (batch ingestion):');
  console.log(`  p50: ${data.metrics.batch_duration.values['p(50)'].toFixed(0)}ms`);
  console.log(`  p95: ${data.metrics.batch_duration.values['p(95)'].toFixed(0)}ms`);
  console.log(`  p99: ${data.metrics.batch_duration.values['p(99)'].toFixed(0)}ms`);
  console.log(`  max: ${data.metrics.batch_duration.values.max.toFixed(0)}ms`);
  console.log('');
  
  // Target validation
  const targetMet = logsPerSecond >= 14000;
  console.log(`Target (14K logs/sec): ${targetMet ? '✅ MET' : '❌ NOT MET'}`);
  
  if (logsPerSecond >= 33000) {
    console.log('🏆 EXCELLENT: Exceeded stretch goal (33K logs/sec)!');
  } else if (logsPerSecond >= 14000) {
    console.log('✅ GOOD: Met minimum target (14K logs/sec)');
  } else {
    console.log('⚠️  WARNING: Below target throughput');
  }
  
  console.log('═══════════════════════════════════════════════════════════');
  
  return {
    'stdout': '', // Suppress default summary
    'summary.json': JSON.stringify(data, null, 2),
  };
}
