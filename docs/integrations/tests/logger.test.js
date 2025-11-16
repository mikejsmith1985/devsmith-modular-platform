/**
 * Unit tests for JavaScript logger (logger.js)
 * 
 * Tests buffer management, batch sending, retry logic, and cleanup.
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');
const http = require('http');

// Load test configuration
const testConfigPath = path.join(__dirname, '.test-config.json');
const testConfig = JSON.parse(fs.readFileSync(testConfigPath, 'utf8'));

// Mock server for testing
let mockServer;
let receivedRequests = [];

// Start mock server before tests
function startMockServer(port = 8999) {
    return new Promise((resolve) => {
        mockServer = http.createServer((req, res) => {
            let body = '';
            req.on('data', chunk => body += chunk);
            req.on('end', () => {
                const request = {
                    method: req.method,
                    url: req.url,
                    headers: req.headers,
                    body: body ? JSON.parse(body) : null
                };
                receivedRequests.push(request);
                
                // Simulate API responses
                if (req.headers['x-api-key'] !== testConfig.apiKey) {
                    res.writeHead(401, { 'Content-Type': 'application/json' });
                    res.end(JSON.stringify({ error: 'Invalid API key' }));
                } else {
                    res.writeHead(200, { 'Content-Type': 'application/json' });
                    res.end(JSON.stringify({ success: true, received: request.body?.logs?.length || 0 }));
                }
            });
        });
        mockServer.listen(port, () => {
            console.log(`Mock server started on port ${port}`);
            resolve();
        });
    });
}

// Stop mock server after tests
function stopMockServer() {
    return new Promise((resolve) => {
        if (mockServer) {
            mockServer.close(() => {
                console.log('Mock server stopped');
                resolve();
            });
        } else {
            resolve();
        }
    });
}

// Import logger after mock server is ready
let Logger;

describe('JavaScript Logger Unit Tests', function() {
    this.timeout(10000);

    before(async () => {
        await startMockServer();
        
        // Dynamically import logger
        const loggerPath = path.join(__dirname, '../javascript/logger.js');
        const loggerCode = fs.readFileSync(loggerPath, 'utf8');
        
        // Create a module context
        const moduleExports = {};
        const moduleFunc = new Function('module', 'exports', loggerCode);
        moduleFunc({ exports: moduleExports }, moduleExports);
        Logger = moduleExports.DevSmithLogger || moduleExports;
    });

    after(async () => {
        await stopMockServer();
    });

    beforeEach(() => {
        receivedRequests = [];
    });

    describe('Initialization', () => {
        it('should create logger with valid configuration', () => {
            const logger = new Logger(
                testConfig.apiKey,
                'http://localhost:8999',
                testConfig.projectSlug,
                'test-service'
            );
            
            assert.ok(logger);
            assert.strictEqual(logger.projectSlug, testConfig.projectSlug);
            assert.strictEqual(logger.serviceName, 'test-service');
        });

        it('should throw error with missing API key', () => {
            assert.throws(() => {
                new Logger(null, 'http://localhost:8999', 'test', 'service');
            }, /API key is required/);
        });

        it('should throw error with missing project slug', () => {
            assert.throws(() => {
                new Logger('key', 'http://localhost:8999', null, 'service');
            }, /Project slug is required/);
        });
    });

    describe('Buffer Management', () => {
        it('should add logs to buffer', () => {
            const logger = new Logger(testConfig.apiKey, 'http://localhost:8999', testConfig.projectSlug, 'test');
            
            logger.info('Test message 1');
            logger.info('Test message 2');
            
            assert.strictEqual(logger.buffer.length, 2);
        });

        it('should respect custom buffer size', () => {
            const logger = new Logger(
                testConfig.apiKey,
                'http://localhost:8999',
                testConfig.projectSlug,
                'test',
                { bufferSize: 5 }
            );
            
            for (let i = 0; i < 10; i++) {
                logger.info(`Message ${i}`);
            }
            
            // Buffer should be cleared after reaching size
            assert.ok(logger.buffer.length < 10);
        });

        it('should trigger flush when buffer is full', (done) => {
            const logger = new Logger(
                testConfig.apiKey,
                'http://localhost:8999',
                testConfig.projectSlug,
                'test',
                { bufferSize: 3 }
            );
            
            logger.info('Message 1');
            logger.info('Message 2');
            logger.info('Message 3'); // Should trigger flush
            
            setTimeout(() => {
                assert.strictEqual(receivedRequests.length, 1);
                assert.strictEqual(receivedRequests[0].body.logs.length, 3);
                done();
            }, 500);
        });
    });

    describe('Log Levels', () => {
        it('should log DEBUG level', () => {
            const logger = new Logger(testConfig.apiKey, 'http://localhost:8999', testConfig.projectSlug, 'test');
            logger.debug('Debug message', { detail: 'value' });
            
            assert.strictEqual(logger.buffer.length, 1);
            assert.strictEqual(logger.buffer[0].level, 'DEBUG');
            assert.strictEqual(logger.buffer[0].message, 'Debug message');
        });

        it('should log INFO level', () => {
            const logger = new Logger(testConfig.apiKey, 'http://localhost:8999', testConfig.projectSlug, 'test');
            logger.info('Info message');
            
            assert.strictEqual(logger.buffer[0].level, 'INFO');
        });

        it('should log WARN level', () => {
            const logger = new Logger(testConfig.apiKey, 'http://localhost:8999', testConfig.projectSlug, 'test');
            logger.warn('Warning message');
            
            assert.strictEqual(logger.buffer[0].level, 'WARN');
        });

        it('should log ERROR level', () => {
            const logger = new Logger(testConfig.apiKey, 'http://localhost:8999', testConfig.projectSlug, 'test');
            logger.error('Error message', { error: 'details' });
            
            assert.strictEqual(logger.buffer[0].level, 'ERROR');
        });
    });

    describe('Context and Tags', () => {
        it('should include context in log entry', () => {
            const logger = new Logger(testConfig.apiKey, 'http://localhost:8999', testConfig.projectSlug, 'test');
            logger.info('Message', { userId: 123, action: 'login' });
            
            assert.deepStrictEqual(logger.buffer[0].context, { userId: 123, action: 'login' });
        });

        it('should include tags in log entry', () => {
            const logger = new Logger(testConfig.apiKey, 'http://localhost:8999', testConfig.projectSlug, 'test');
            logger.info('Message', {}, ['auth', 'user']);
            
            assert.deepStrictEqual(logger.buffer[0].tags, ['auth', 'user']);
        });
    });

    describe('Batch Sending', () => {
        it('should send batch with correct format', (done) => {
            const logger = new Logger(
                testConfig.apiKey,
                'http://localhost:8999',
                testConfig.projectSlug,
                'test',
                { bufferSize: 2 }
            );
            
            logger.info('Message 1');
            logger.info('Message 2');
            
            setTimeout(() => {
                assert.strictEqual(receivedRequests.length, 1);
                const request = receivedRequests[0];
                
                assert.strictEqual(request.method, 'POST');
                assert.strictEqual(request.headers['x-api-key'], testConfig.apiKey);
                assert.strictEqual(request.body.project_slug, testConfig.projectSlug);
                assert.strictEqual(request.body.logs.length, 2);
                
                done();
            }, 500);
        });

        it('should include all required fields in batch', (done) => {
            const logger = new Logger(
                testConfig.apiKey,
                'http://localhost:8999',
                testConfig.projectSlug,
                'test-service',
                { bufferSize: 1 }
            );
            
            logger.info('Test message', { key: 'value' }, ['tag1']);
            
            setTimeout(() => {
                const log = receivedRequests[0].body.logs[0];
                
                assert.ok(log.timestamp);
                assert.strictEqual(log.level, 'INFO');
                assert.strictEqual(log.message, 'Test message');
                assert.strictEqual(log.service, 'test-service');
                assert.deepStrictEqual(log.context, { key: 'value' });
                assert.deepStrictEqual(log.tags, ['tag1']);
                
                done();
            }, 500);
        });
    });

    describe('Time-based Flush', () => {
        it('should flush after flush interval', function(done) {
            this.timeout(7000);
            
            const logger = new Logger(
                testConfig.apiKey,
                'http://localhost:8999',
                testConfig.projectSlug,
                'test',
                { flushInterval: 2000 } // 2 seconds
            );
            
            logger.info('Message 1');
            
            // Should not have sent yet
            assert.strictEqual(receivedRequests.length, 0);
            
            // Wait for flush interval
            setTimeout(() => {
                assert.strictEqual(receivedRequests.length, 1);
                done();
            }, 2500);
        });
    });

    describe('Retry Logic', () => {
        it('should retry on network failure', function(done) {
            this.timeout(5000);
            
            // Create logger with invalid URL
            const logger = new Logger(
                testConfig.apiKey,
                'http://localhost:9999', // Non-existent port
                testConfig.projectSlug,
                'test',
                { bufferSize: 1, flushInterval: 1000 }
            );
            
            logger.info('Test message');
            
            // After failed attempt, logs should remain in buffer
            setTimeout(() => {
                assert.ok(logger.buffer.length > 0, 'Logs should be retained after failure');
                done();
            }, 1500);
        });
    });

    describe('Cleanup', () => {
        it('should flush buffer on close', (done) => {
            const logger = new Logger(
                testConfig.apiKey,
                'http://localhost:8999',
                testConfig.projectSlug,
                'test'
            );
            
            logger.info('Message 1');
            logger.info('Message 2');
            
            logger.close();
            
            setTimeout(() => {
                assert.strictEqual(receivedRequests.length, 1);
                assert.strictEqual(receivedRequests[0].body.logs.length, 2);
                done();
            }, 500);
        });
    });
});

// Run tests if executed directly
if (require.main === module) {
    const Mocha = require('mocha');
    const mocha = new Mocha();
    mocha.suite.emit('pre-require', global, null, mocha);
    
    // Add this file
    mocha.addFile(__filename);
    
    // Run tests
    mocha.run((failures) => {
        process.exitCode = failures ? 1 : 0;
    });
}
