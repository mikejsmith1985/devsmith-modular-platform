/**
 * DevSmith Logger - JavaScript/Node.js Integration
 * 
 * Copy this file into your project and customize the configuration.
 * 
 * Usage:
 *   const DevSmithLogger = require('./logger');
 *   const logger = new DevSmithLogger({
 *     apiKey: process.env.DEVSMITH_API_KEY,
 *     apiUrl: process.env.DEVSMITH_API_URL || 'http://localhost:3000',
 *     projectSlug: 'my-project',
 *     serviceName: 'api-server'
 *   });
 *   
 *   logger.info('User logged in', { userId: 123 });
 *   logger.error('Database error', { code: 'ECONNREFUSED' });
 */

const https = require('https');
const http = require('http');

class DevSmithLogger {
  constructor(config) {
    this.apiKey = config.apiKey;
    this.apiUrl = config.apiUrl || 'http://localhost:3000';
    this.projectSlug = config.projectSlug;
    this.serviceName = config.serviceName;
    this.batchSize = config.batchSize || 100;
    this.flushInterval = config.flushInterval || 5000; // 5 seconds
    
    this.buffer = [];
    this.timer = null;
    
    // Validate required config
    if (!this.apiKey) {
      throw new Error('DevSmithLogger: apiKey is required');
    }
    if (!this.projectSlug) {
      throw new Error('DevSmithLogger: projectSlug is required');
    }
    if (!this.serviceName) {
      throw new Error('DevSmithLogger: serviceName is required');
    }
    
    // Setup flush timer
    this.startTimer();
    
    // Flush on process exit
    process.on('exit', () => this.flush());
    process.on('SIGINT', () => {
      this.flush();
      process.exit();
    });
  }
  
  startTimer() {
    this.timer = setInterval(() => {
      if (this.buffer.length > 0) {
        this.flush();
      }
    }, this.flushInterval);
  }
  
  log(level, message, context = {}) {
    const entry = {
      timestamp: new Date().toISOString(),
      level: level.toUpperCase(),
      message: message,
      service: this.serviceName,
      context: context
    };
    
    this.buffer.push(entry);
    
    // Flush if batch size reached
    if (this.buffer.length >= this.batchSize) {
      this.flush();
    }
  }
  
  flush() {
    if (this.buffer.length === 0) {
      return;
    }
    
    const logs = [...this.buffer];
    this.buffer = [];
    
    const payload = JSON.stringify({
      project_slug: this.projectSlug,
      logs: logs
    });
    
    const url = new URL('/api/logs/batch', this.apiUrl);
    const isHttps = url.protocol === 'https:';
    const client = isHttps ? https : http;
    
    const options = {
      hostname: url.hostname,
      port: url.port || (isHttps ? 443 : 80),
      path: url.pathname,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(payload),
        'Authorization': `Bearer ${this.apiKey}`
      }
    };
    
    const req = client.request(options, (res) => {
      let data = '';
      res.on('data', (chunk) => {
        data += chunk;
      });
      res.on('end', () => {
        if (res.statusCode !== 200 && res.statusCode !== 201) {
          console.error(`DevSmith Logger: Failed to send logs (${res.statusCode}):`, data);
        }
      });
    });
    
    req.on('error', (error) => {
      console.error('DevSmith Logger: Network error:', error.message);
      // Re-add logs to buffer for retry
      this.buffer.push(...logs);
    });
    
    req.write(payload);
    req.end();
  }
  
  // Convenience methods
  debug(message, context) {
    this.log('DEBUG', message, context);
  }
  
  info(message, context) {
    this.log('INFO', message, context);
  }
  
  warn(message, context) {
    this.log('WARN', message, context);
  }
  
  error(message, context) {
    this.log('ERROR', message, context);
  }
}

module.exports = DevSmithLogger;
