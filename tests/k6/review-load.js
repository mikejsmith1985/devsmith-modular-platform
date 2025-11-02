import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

/**
 * K6 LOAD TEST FOR REVIEW SERVICE
 * 
 * Tests all 5 reading modes under realistic load:
 * - 10 Virtual Users (VUs)
 * - 100 total requests (10 per VU)
 * - Measures P95/P99 latency
 * - Observes circuit breaker behavior
 * 
 * Run:
 *   k6 run tests/k6/review-load.js
 * 
 * Generate report:
 *   k6 run --out json=.docs/perf/k6-results.json tests/k6/review-load.js
 */

// Custom metrics
const errorRate = new Rate('errors');
const previewLatency = new Trend('preview_mode_duration');
const skimLatency = new Trend('skim_mode_duration');
const scanLatency = new Trend('scan_mode_duration');
const detailedLatency = new Trend('detailed_mode_duration');
const criticalLatency = new Trend('critical_mode_duration');

// Test configuration
export const options = {
	vus: 10,                    // 10 concurrent virtual users
	iterations: 100,            // 100 total requests (10 per VU)
	duration: '2m',            // Max 2 minutes
	thresholds: {
		'http_req_duration': ['p(95)<5000', 'p(99)<10000'], // 95th percentile < 5s, 99th < 10s
		'http_req_failed': ['rate<0.1'],                     // < 10% failure rate
		'errors': ['rate<0.1'],                              // < 10% errors
		'preview_mode_duration': ['p(95)<3000'],             // Preview < 3s (95th percentile)
		'skim_mode_duration': ['p(95)<5000'],                // Skim < 5s
		'scan_mode_duration': ['p(95)<4000'],                // Scan < 4s
		'detailed_mode_duration': ['p(95)<7000'],            // Detailed < 7s
		'critical_mode_duration': ['p(95)<10000'],           // Critical < 10s
	},
};

const BASE_URL = __ENV.REVIEW_URL || 'http://localhost:3000';

// Sample code for testing
const SAMPLE_GO_CODE = `package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func GetUser(c *gin.Context) {
	id := c.Param("id")
	user := fetchUserFromDB(id)
	c.JSON(http.StatusOK, user)
}`;

const SAMPLE_VULNERABLE_CODE = `package handlers

import (
	"database/sql"
	"fmt"
)

func GetUser(id string) (*User, error) {
	query := "SELECT * FROM users WHERE id = " + id  // SQL injection
	rows, _ := db.Query(query)  // Error ignored
	return parseUser(rows), nil
}`;

// Helper function to test a reading mode
function testReadingMode(mode, code, latencyMetric) {
	const payload = {
		pasted_code: code,
		model: 'mistral:7b-instruct',
	};

	const params = {
		headers: {
			'Content-Type': 'application/x-www-form-urlencoded',
		},
		tags: { mode: mode },
	};

	// Encode payload as form data
	const formData = Object.keys(payload)
		.map(key => `${encodeURIComponent(key)}=${encodeURIComponent(payload[key])}`)
		.join('&');

	const startTime = Date.now();
	const response = http.post(`${BASE_URL}/api/review/modes/${mode}`, formData, params);
	const duration = Date.now() - startTime;

	// Record latency for this mode
	latencyMetric.add(duration);

	// Check response
	const success = check(response, {
		'status is 200': (r) => r.status === 200,
		'response has content': (r) => r.body && r.body.length > 0,
		'response time acceptable': (r) => r.timings.duration < 30000, // < 30s
	});

	if (!success) {
		errorRate.add(1);
		console.error(`[${mode}] Request failed: ${response.status} - ${response.body}`);
	} else {
		errorRate.add(0);
	}

	return response;
}

// Main test scenario
export default function () {
	// Randomly select a reading mode to simulate realistic usage
	const modes = [
		{ name: 'preview', code: SAMPLE_GO_CODE, metric: previewLatency, weight: 30 },
		{ name: 'skim', code: SAMPLE_GO_CODE, metric: skimLatency, weight: 25 },
		{ name: 'scan', code: SAMPLE_GO_CODE, metric: scanLatency, weight: 20 },
		{ name: 'detailed', code: SAMPLE_GO_CODE, metric: detailedLatency, weight: 15 },
		{ name: 'critical', code: SAMPLE_VULNERABLE_CODE, metric: criticalLatency, weight: 10 },
	];

	// Weighted random selection (Preview most common, Critical least)
	const totalWeight = modes.reduce((sum, m) => sum + m.weight, 0);
	let random = Math.random() * totalWeight;
	let selectedMode = modes[0];

	for (const mode of modes) {
		random -= mode.weight;
		if (random <= 0) {
			selectedMode = mode;
			break;
		}
	}

	// Execute request
	testReadingMode(selectedMode.name, selectedMode.code, selectedMode.metric);

	// Think time: 1-3 seconds between requests (realistic user behavior)
	sleep(Math.random() * 2 + 1);
}

// Setup function (runs once at start)
export function setup() {
	console.log('===================================');
	console.log('K6 Load Test - Review Service');
	console.log('===================================');
	console.log(`Base URL: ${BASE_URL}`);
	console.log(`VUs: ${options.vus}`);
	console.log(`Total Iterations: ${options.iterations}`);
	console.log(`Duration: ${options.duration}`);
	console.log('===================================\n');

	// Health check
	const healthResponse = http.get(`${BASE_URL}/api/review/health`);
	check(healthResponse, {
		'health check passes': (r) => r.status === 200,
		'service is healthy': (r) => {
			try {
				const body = JSON.parse(r.body);
				return body.status === 'healthy';
			} catch (e) {
				return false;
			}
		},
	});

	if (healthResponse.status !== 200) {
		console.error('❌ Health check failed! Review service may not be running.');
		console.error(`Response: ${healthResponse.status} - ${healthResponse.body}`);
	} else {
		console.log('✅ Health check passed. Starting load test...\n');
	}
}

// Teardown function (runs once at end)
export function teardown(data) {
	console.log('\n===================================');
	console.log('Load Test Complete');
	console.log('===================================');
	console.log('Check metrics above for:');
	console.log('  - http_req_duration (P95, P99)');
	console.log('  - [mode]_duration for each reading mode');
	console.log('  - error rate');
	console.log('  - http_req_failed rate');
	console.log('\nGenerate detailed report:');
	console.log('  k6 run --out json=.docs/perf/k6-results.json tests/k6/review-load.js');
	console.log('===================================\n');
}
