const assert = require('assert');
const express = require('express');
const request = require('supertest');
const fs = require('fs');
const path = require('path');

// Load test configuration
const testConfigPath = path.join(__dirname, '.test-config.json');
const testConfig = JSON.parse(fs.readFileSync(testConfigPath, 'utf8'));

// Mock logger to track calls
class MockLogger {
  constructor() {
    this.logCalls = [];
  }

  debug(message, context, tags) {
    this.logCalls.push({ level: 'DEBUG', message, context, tags });
  }

  info(message, context, tags) {
    this.logCalls.push({ level: 'INFO', message, context, tags });
  }

  warn(message, context, tags) {
    this.logCalls.push({ level: 'WARN', message, context, tags });
  }

  error(message, context, tags) {
    this.logCalls.push({ level: 'ERROR', message, context, tags });
  }

  close() {
    // Mock close
  }
}

// Import middleware (dynamically to use mock logger)
const { DevSmithMiddleware } = require('../javascript/express-middleware');

describe('Express Middleware Tests', () => {
  let mockLogger;

  beforeEach(() => {
    mockLogger = new MockLogger();
  });

  describe('Initialization', () => {
    it('should create middleware with valid logger', () => {
      const middleware = DevSmithMiddleware(mockLogger);
      assert.strictEqual(typeof middleware, 'function');
    });

    it('should throw error for missing logger', () => {
      assert.throws(() => {
        DevSmithMiddleware(null);
      }, /logger/i);
    });
  });

  describe('Request Logging', () => {
    it('should log incoming requests', async () => {
      const app = express();
      app.use(DevSmithMiddleware(mockLogger));
      app.get('/test', (req, res) => res.json({ ok: true }));

      await request(app).get('/test');

      const requestLog = mockLogger.logCalls.find(log => 
        log.message.includes('Incoming request')
      );

      assert.ok(requestLog, 'Should have request log');
      assert.strictEqual(requestLog.level, 'INFO');
      assert.strictEqual(requestLog.context.method, 'GET');
      assert.strictEqual(requestLog.context.path, '/test');
    });

    it('should log response details', async () => {
      const app = express();
      app.use(DevSmithMiddleware(mockLogger));
      app.get('/test', (req, res) => res.status(201).json({ ok: true }));

      await request(app).get('/test');

      const responseLog = mockLogger.logCalls.find(log => 
        log.message.includes('Request completed')
      );

      assert.ok(responseLog, 'Should have response log');
      assert.strictEqual(responseLog.context.statusCode, 201);
      assert.ok(responseLog.context.duration >= 0);
    });

    it('should include request timing', async () => {
      const app = express();
      app.use(DevSmithMiddleware(mockLogger));
      app.get('/test', async (req, res) => {
        await new Promise(resolve => setTimeout(resolve, 100));
        res.json({ ok: true });
      });

      await request(app).get('/test');

      const responseLog = mockLogger.logCalls.find(log => 
        log.message.includes('Request completed')
      );

      assert.ok(responseLog.context.duration >= 100);
      assert.ok(responseLog.context.duration < 200);
    });
  });

  describe('Header Redaction', () => {
    it('should redact authorization header', async () => {
      const app = express();
      app.use(DevSmithMiddleware(mockLogger));
      app.get('/test', (req, res) => res.json({ ok: true }));

      await request(app)
        .get('/test')
        .set('Authorization', 'Bearer secret-token-12345');

      const requestLog = mockLogger.logCalls.find(log => 
        log.message.includes('Incoming request')
      );

      assert.strictEqual(requestLog.context.headers.authorization, '[REDACTED]');
    });

    it('should redact cookie header', async () => {
      const app = express();
      app.use(DevSmithMiddleware(mockLogger));
      app.get('/test', (req, res) => res.json({ ok: true }));

      await request(app)
        .get('/test')
        .set('Cookie', 'session=abc123; token=xyz789');

      const requestLog = mockLogger.logCalls.find(log => 
        log.message.includes('Incoming request')
      );

      assert.strictEqual(requestLog.context.headers.cookie, '[REDACTED]');
    });

    it('should preserve non-sensitive headers', async () => {
      const app = express();
      app.use(DevSmithMiddleware(mockLogger));
      app.get('/test', (req, res) => res.json({ ok: true }));

      await request(app)
        .get('/test')
        .set('User-Agent', 'test-agent')
        .set('Accept', 'application/json');

      const requestLog = mockLogger.logCalls.find(log => 
        log.message.includes('Incoming request')
      );

      assert.strictEqual(requestLog.context.headers['user-agent'], 'test-agent');
      assert.strictEqual(requestLog.context.headers.accept, 'application/json');
    });
  });

  describe('Skip Paths', () => {
    it('should skip health check endpoint', async () => {
      const app = express();
      app.use(DevSmithMiddleware(mockLogger, { skipPaths: ['/health'] }));
      app.get('/health', (req, res) => res.json({ ok: true }));

      await request(app).get('/health');

      assert.strictEqual(mockLogger.logCalls.length, 0, 'Should not log skipped path');
    });

    it('should skip multiple paths', async () => {
      const app = express();
      app.use(DevSmithMiddleware(mockLogger, { skipPaths: ['/health', '/metrics', '/favicon.ico'] }));
      app.get('/health', (req, res) => res.json({ ok: true }));
      app.get('/metrics', (req, res) => res.json({ ok: true }));
      app.get('/test', (req, res) => res.json({ ok: true }));

      await request(app).get('/health');
      await request(app).get('/metrics');
      await request(app).get('/test');

      // Should only log /test
      assert.strictEqual(mockLogger.logCalls.length, 2); // Request + response for /test
    });

    it('should support path patterns', async () => {
      const app = express();
      app.use(DevSmithMiddleware(mockLogger, { skipPaths: ['/api/internal/*'] }));
      app.get('/api/internal/metrics', (req, res) => res.json({ ok: true }));
      app.get('/api/users', (req, res) => res.json({ ok: true }));

      await request(app).get('/api/internal/metrics');
      await request(app).get('/api/users');

      // Should only log /api/users
      const paths = mockLogger.logCalls.map(log => log.context.path);
      assert.ok(paths.includes('/api/users'));
      assert.ok(!paths.includes('/api/internal/metrics'));
    });
  });

  describe('Error Handling', () => {
    it('should log errors', async () => {
      const app = express();
      app.use(DevSmithMiddleware(mockLogger));
      app.get('/error', (req, res) => {
        throw new Error('Test error');
      });
      app.use((err, req, res, next) => {
        res.status(500).json({ error: err.message });
      });

      await request(app).get('/error');

      const errorLog = mockLogger.logCalls.find(log => 
        log.level === 'ERROR'
      );

      assert.ok(errorLog, 'Should have error log');
      assert.ok(errorLog.message.includes('Test error'));
      assert.strictEqual(errorLog.context.method, 'GET');
      assert.strictEqual(errorLog.context.path, '/error');
    });

    it('should log stack trace', async () => {
      const app = express();
      app.use(DevSmithMiddleware(mockLogger));
      app.get('/error', (req, res) => {
        const error = new Error('Test error with stack');
        throw error;
      });
      app.use((err, req, res, next) => {
        res.status(500).json({ error: err.message });
      });

      await request(app).get('/error');

      const errorLog = mockLogger.logCalls.find(log => 
        log.level === 'ERROR'
      );

      assert.ok(errorLog.context.stack, 'Should have stack trace');
      assert.ok(errorLog.context.stack.includes('Test error with stack'));
    });
  });

  describe('Custom Tags', () => {
    it('should allow custom tags', async () => {
      const app = express();
      app.use(DevSmithMiddleware(mockLogger, { tags: ['api', 'production'] }));
      app.get('/test', (req, res) => res.json({ ok: true }));

      await request(app).get('/test');

      const requestLog = mockLogger.logCalls.find(log => 
        log.message.includes('Incoming request')
      );

      assert.ok(Array.isArray(requestLog.tags));
      assert.ok(requestLog.tags.includes('api'));
      assert.ok(requestLog.tags.includes('production'));
    });

    it('should merge default and custom tags', async () => {
      const app = express();
      app.use(DevSmithMiddleware(mockLogger, { tags: ['custom'] }));
      app.get('/test', (req, res) => res.json({ ok: true }));

      await request(app).get('/test');

      const requestLog = mockLogger.logCalls.find(log => 
        log.message.includes('Incoming request')
      );

      assert.ok(requestLog.tags.includes('express'));
      assert.ok(requestLog.tags.includes('custom'));
    });
  });

  describe('Performance', () => {
    it('should handle many requests efficiently', async () => {
      const app = express();
      app.use(DevSmithMiddleware(mockLogger));
      app.get('/test', (req, res) => res.json({ ok: true }));

      const numRequests = 100;
      const start = Date.now();

      const requests = [];
      for (let i = 0; i < numRequests; i++) {
        requests.push(request(app).get('/test'));
      }
      await Promise.all(requests);

      const duration = Date.now() - start;

      // Should complete 100 requests in reasonable time (< 5 seconds)
      assert.ok(duration < 5000, `100 requests took ${duration}ms (should be < 5000ms)`);
      
      // Should have logged all requests
      assert.ok(mockLogger.logCalls.length >= numRequests * 2); // Request + response per request
    });
  });
});

console.log('Express middleware tests ready to run with: npm test');
