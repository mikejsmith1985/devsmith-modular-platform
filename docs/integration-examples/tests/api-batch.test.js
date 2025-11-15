/**
 * API Batch Endpoint Tests
 * Tests POST /api/logs/batch endpoint directly against real API
 * 
 * Run: mocha api-batch.test.js
 * Requires: DevSmith platform running, test project created
 */

const assert = require('assert');
const http = require('http');
const fs = require('fs');
const path = require('path');

// Load test configuration
const configPath = path.join(__dirname, '.test-config.json');
let config;

try {
  config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
} catch (error) {
  console.error('ERROR: Could not load .test-config.json');
  console.error('Run: bash setup-test-env.sh');
  process.exit(1);
}

/**
 * Make HTTP POST request to batch endpoint
 */
function postBatch(data, apiKey = config.apiKey) {
  return new Promise((resolve, reject) => {
    const url = new URL(config.batchEndpoint);
    const payload = JSON.stringify(data);
    
    const options = {
      hostname: url.hostname,
      port: url.port,
      path: url.pathname,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(payload),
        'Authorization': `Bearer ${apiKey}`
      }
    };

    const req = http.request(options, (res) => {
      let body = '';
      res.on('data', chunk => body += chunk);
      res.on('end', () => {
        try {
          const parsed = JSON.parse(body);
          resolve({ status: res.statusCode, body: parsed, headers: res.headers });
        } catch (e) {
          resolve({ status: res.statusCode, body, headers: res.headers });
        }
      });
    });

    req.on('error', reject);
    req.write(payload);
    req.end();
  });
}

/**
 * Query database to verify logs were inserted
 */
function queryDatabase(projectSlug) {
  return new Promise((resolve, reject) => {
    // This would require pg module, simplified for now
    // In real implementation, query logs.entries table
    // SELECT COUNT(*) FROM logs.entries WHERE project_slug = $1
    resolve({ count: 0 }); // Placeholder
  });
}

describe('API Batch Endpoint Tests', () => {
  
  describe('Valid Batch Requests', () => {
    
    it('should accept valid batch with single log', async () => {
      const batch = {
        project_slug: config.projectSlug,
        logs: [
          {
            timestamp: new Date().toISOString(),
            level: 'INFO',
            message: 'Test log from API test',
            service: 'api-test',
            context: { test: true },
            tags: ['api-test']
          }
        ]
      };

      const response = await postBatch(batch);
      
      assert.strictEqual(response.status, 200, 'Should return 200 OK');
      assert.ok(response.body.success, 'Response should indicate success');
      assert.ok(response.body.inserted >= 1, 'Should insert at least 1 log');
    });

    it('should accept batch with multiple logs (100)', async () => {
      const logs = [];
      for (let i = 0; i < 100; i++) {
        logs.push({
          timestamp: new Date().toISOString(),
          level: i % 4 === 0 ? 'ERROR' : i % 3 === 0 ? 'WARN' : i % 2 === 0 ? 'INFO' : 'DEBUG',
          message: `Batch test log ${i}`,
          service: 'api-test',
          context: { index: i },
          tags: ['batch-test', `batch-${Math.floor(i / 10)}`]
        });
      }

      const batch = {
        project_slug: config.projectSlug,
        logs
      };

      const response = await postBatch(batch);
      
      assert.strictEqual(response.status, 200, 'Should return 200 OK');
      assert.strictEqual(response.body.inserted, 100, 'Should insert all 100 logs');
    });

    it('should handle batch with all log levels', async () => {
      const batch = {
        project_slug: config.projectSlug,
        logs: [
          { timestamp: new Date().toISOString(), level: 'DEBUG', message: 'Debug message', service: 'test' },
          { timestamp: new Date().toISOString(), level: 'INFO', message: 'Info message', service: 'test' },
          { timestamp: new Date().toISOString(), level: 'WARN', message: 'Warning message', service: 'test' },
          { timestamp: new Date().toISOString(), level: 'ERROR', message: 'Error message', service: 'test' }
        ]
      };

      const response = await postBatch(batch);
      
      assert.strictEqual(response.status, 200, 'Should return 200 OK');
      assert.strictEqual(response.body.inserted, 4, 'Should insert all 4 logs');
    });

    it('should handle batch with context and tags', async () => {
      const batch = {
        project_slug: config.projectSlug,
        logs: [
          {
            timestamp: new Date().toISOString(),
            level: 'INFO',
            message: 'Log with rich metadata',
            service: 'test',
            context: {
              user_id: 123,
              request_id: 'abc-123',
              duration_ms: 456,
              nested: { key: 'value' }
            },
            tags: ['production', 'api', 'critical']
          }
        ]
      };

      const response = await postBatch(batch);
      
      assert.strictEqual(response.status, 200, 'Should return 200 OK');
      assert.strictEqual(response.body.inserted, 1, 'Should insert log with metadata');
    });
  });

  describe('Authentication', () => {
    
    it('should reject request with invalid API key', async () => {
      const batch = {
        project_slug: config.projectSlug,
        logs: [
          { timestamp: new Date().toISOString(), level: 'INFO', message: 'Test', service: 'test' }
        ]
      };

      const response = await postBatch(batch, 'invalid-api-key-12345');
      
      assert.strictEqual(response.status, 401, 'Should return 401 Unauthorized');
      assert.ok(response.body.error, 'Should include error message');
    });

    it('should reject request with missing API key', async () => {
      const batch = {
        project_slug: config.projectSlug,
        logs: [
          { timestamp: new Date().toISOString(), level: 'INFO', message: 'Test', service: 'test' }
        ]
      };

      const response = await postBatch(batch, '');
      
      assert.strictEqual(response.status, 401, 'Should return 401 Unauthorized');
    });
  });

  describe('Validation', () => {
    
    it('should reject batch with missing project_slug', async () => {
      const batch = {
        logs: [
          { timestamp: new Date().toISOString(), level: 'INFO', message: 'Test', service: 'test' }
        ]
      };

      const response = await postBatch(batch);
      
      assert.strictEqual(response.status, 400, 'Should return 400 Bad Request');
      assert.ok(response.body.error, 'Should include validation error');
    });

    it('should reject batch with missing logs array', async () => {
      const batch = {
        project_slug: config.projectSlug
      };

      const response = await postBatch(batch);
      
      assert.strictEqual(response.status, 400, 'Should return 400 Bad Request');
      assert.ok(response.body.error, 'Should indicate missing logs array');
    });

    it('should reject batch with empty logs array', async () => {
      const batch = {
        project_slug: config.projectSlug,
        logs: []
      };

      const response = await postBatch(batch);
      
      assert.strictEqual(response.status, 400, 'Should return 400 Bad Request');
    });

    it('should reject log entry with missing required fields', async () => {
      const batch = {
        project_slug: config.projectSlug,
        logs: [
          { 
            timestamp: new Date().toISOString(),
            // Missing: level, message, service
          }
        ]
      };

      const response = await postBatch(batch);
      
      assert.strictEqual(response.status, 400, 'Should return 400 Bad Request');
      assert.ok(response.body.error, 'Should indicate missing required fields');
    });

    it('should reject log entry with invalid level', async () => {
      const batch = {
        project_slug: config.projectSlug,
        logs: [
          {
            timestamp: new Date().toISOString(),
            level: 'INVALID_LEVEL',
            message: 'Test',
            service: 'test'
          }
        ]
      };

      const response = await postBatch(batch);
      
      assert.strictEqual(response.status, 400, 'Should return 400 Bad Request');
      assert.ok(response.body.error, 'Should indicate invalid level');
    });
  });

  describe('Performance', () => {
    
    it('should handle large batch (500 logs) efficiently', async function() {
      this.timeout(10000); // Allow up to 10 seconds
      
      const logs = [];
      for (let i = 0; i < 500; i++) {
        logs.push({
          timestamp: new Date().toISOString(),
          level: 'INFO',
          message: `Performance test log ${i}`,
          service: 'perf-test',
          context: { batch: 'large', index: i },
          tags: ['performance']
        });
      }

      const batch = {
        project_slug: config.projectSlug,
        logs
      };

      const startTime = Date.now();
      const response = await postBatch(batch);
      const duration = Date.now() - startTime;
      
      assert.strictEqual(response.status, 200, 'Should return 200 OK');
      assert.strictEqual(response.body.inserted, 500, 'Should insert all 500 logs');
      assert.ok(duration < 5000, `Should complete in < 5 seconds (took ${duration}ms)`);
    });

    it('should handle concurrent batches', async function() {
      this.timeout(10000);
      
      const createBatch = (batchId) => ({
        project_slug: config.projectSlug,
        logs: Array.from({ length: 50 }, (_, i) => ({
          timestamp: new Date().toISOString(),
          level: 'INFO',
          message: `Concurrent batch ${batchId} log ${i}`,
          service: 'concurrent-test',
          context: { batchId, logIndex: i },
          tags: ['concurrent']
        }))
      });

      // Send 5 batches concurrently
      const promises = [
        postBatch(createBatch(1)),
        postBatch(createBatch(2)),
        postBatch(createBatch(3)),
        postBatch(createBatch(4)),
        postBatch(createBatch(5))
      ];

      const responses = await Promise.all(promises);
      
      // All should succeed
      responses.forEach((response, index) => {
        assert.strictEqual(response.status, 200, `Batch ${index + 1} should return 200 OK`);
        assert.strictEqual(response.body.inserted, 50, `Batch ${index + 1} should insert 50 logs`);
      });
    });
  });

  describe('Rate Limiting', () => {
    
    it('should respect rate limits for authenticated requests', async function() {
      this.timeout(15000);
      
      const batch = {
        project_slug: config.projectSlug,
        logs: [
          { timestamp: new Date().toISOString(), level: 'INFO', message: 'Rate limit test', service: 'test' }
        ]
      };

      // Send many requests rapidly
      const promises = [];
      for (let i = 0; i < 100; i++) {
        promises.push(postBatch(batch));
      }

      const responses = await Promise.all(promises);
      
      // Check if any were rate limited (429)
      const rateLimited = responses.filter(r => r.status === 429);
      const successful = responses.filter(r => r.status === 200);
      
      // Most should succeed (we have high limits), but rate limiting mechanism should exist
      assert.ok(successful.length > 0, 'Some requests should succeed');
      
      // If rate limited, response should indicate retry-after
      if (rateLimited.length > 0) {
        assert.ok(rateLimited[0].headers['retry-after'], 'Rate limited response should include Retry-After header');
      }
    });
  });

  describe('Error Handling', () => {
    
    it('should handle malformed JSON gracefully', async () => {
      return new Promise((resolve) => {
        const url = new URL(config.batchEndpoint);
        const options = {
          hostname: url.hostname,
          port: url.port,
          path: url.pathname,
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${config.apiKey}`
          }
        };

        const req = http.request(options, (res) => {
          let body = '';
          res.on('data', chunk => body += chunk);
          res.on('end', () => {
            assert.strictEqual(res.statusCode, 400, 'Should return 400 Bad Request');
            resolve();
          });
        });

        req.write('{ malformed json');
        req.end();
      });
    });

    it('should return detailed error for partial batch failure', async () => {
      // Some logs valid, some invalid
      const batch = {
        project_slug: config.projectSlug,
        logs: [
          { timestamp: new Date().toISOString(), level: 'INFO', message: 'Valid log', service: 'test' },
          { timestamp: 'invalid-timestamp', level: 'INFO', message: 'Invalid timestamp', service: 'test' },
          { timestamp: new Date().toISOString(), level: 'INFO', message: 'Another valid log', service: 'test' }
        ]
      };

      const response = await postBatch(batch);
      
      // Depending on implementation, might be 200 with partial success or 400 with error
      if (response.status === 200) {
        assert.ok(response.body.inserted < 3, 'Should insert fewer than all logs');
        assert.ok(response.body.errors, 'Should include error details');
      } else {
        assert.strictEqual(response.status, 400, 'Should return 400 for invalid data');
      }
    });
  });
});
