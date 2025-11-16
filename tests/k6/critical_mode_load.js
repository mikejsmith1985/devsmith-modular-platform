/**
 * Critical Mode Performance Test
 * 
 * Tests the Critical Mode reading endpoint under load.
 * Measures response times, throughput, and error rates.
 * 
 * Usage:
 *   k6 run tests/k6/critical_mode_load.js
 *   k6 run tests/k6/critical_mode_load.js -v       # Verbose
 *   k6 run tests/k6/critical_mode_load.js --stage=peak
 */

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const duration = new Trend('request_duration');
const throughput = new Counter('requests');
const successRate = new Rate('success');

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';
const REVIEW_ID = '1';

// Test code for analysis
const TEST_CODE = `package main

import (
    "database/sql"
    "log"
)

func getUserData(userInput string) map[string]interface{} {
    // SECURITY: SQL injection vulnerability
    db, _ := sql.Open("postgres", "user=admin password=secret dbname=prod")
    rows, _ := db.Query("SELECT * FROM users WHERE id = " + userInput)
    
    // PERFORMANCE: N+1 query in loop
    users := make([]map[string]interface{}, 0)
    for rows.Next() {
        var id int
        var name string
        rows.Scan(&id, &name)
        
        // This runs a query for EVERY row (N+1 problem)
        details, _ := db.Query("SELECT * FROM user_details WHERE user_id = ?", id)
        details.Close()
        
        users = append(users, map[string]interface{}{
            "id":   id,
            "name": name,
        })
    }
    
    // ERROR HANDLING: Missing error check
    rows.Close()
    
    return map[string]interface{}{
        "users": users,
    }
}`;

export const options = {
    stages: [
        { duration: '10s', target: 5 },    // Ramp up to 5 VUs
        { duration: '30s', target: 10 },   // Ramp up to 10 VUs
        { duration: '20s', target: 10 },   // Stay at 10 VUs (stable)
        { duration: '10s', target: 0 },    // Ramp down to 0
    ],
    thresholds: {
        'http_req_duration': ['p(95)<2000', 'p(99)<3000'], // 95th percentile under 2s
        'errors': ['rate<0.1'],                            // Error rate under 10%
        'requests': ['count>50'],                          // At least 50 requests
    },
};

export default function () {
    group('Critical Mode Analysis', () => {
        const url = `${BASE_URL}/api/review/sessions/${REVIEW_ID}/modes/critical`;
        const payload = JSON.stringify({
            code: TEST_CODE,
        });

        const params = {
            headers: {
                'Content-Type': 'application/json',
            },
            timeout: '30s',
        };

        const response = http.post(url, payload, params);
        
        // Record metrics
        throughput.add(1);
        duration.add(response.timings.duration);
        
        const success = check(response, {
            'status is 200': (r) => r.status === 200,
            'response time < 2s': (r) => r.timings.duration < 2000,
            'response has issues array': (r) => r.json('issues') !== undefined,
            'response has overall_grade': (r) => r.json('overall_grade') !== undefined,
            'response has summary': (r) => r.json('summary') !== undefined,
        });
        
        successRate.add(success);
        errorRate.add(!success);
        
        sleep(1); // Wait 1s between requests
    });
}

export function handleSummary(data) {
    console.log('=== Critical Mode Performance Test Summary ===');
    console.log(`Requests: ${data.metrics.requests.values.count}`);
    console.log(`Errors: ${data.metrics.errors.values.count}`);
    console.log(`Success Rate: ${(data.metrics.success.values.rate * 100).toFixed(2)}%`);
    console.log(`Avg Duration: ${data.metrics.request_duration.values.avg.toFixed(0)}ms`);
    console.log(`P95 Duration: ${data.metrics.request_duration.values['p(95)'].toFixed(0)}ms`);
    console.log(`P99 Duration: ${data.metrics.request_duration.values['p(99)'].toFixed(0)}ms`);
    
    return {
        stdout: textSummary(data, { indent: ' ', enableColors: true }),
    };
}
