/**
 * Express.js Middleware for DevSmith Logging
 * 
 * Automatically logs HTTP requests/responses to DevSmith platform.
 * 
 * Installation:
 * 1. Copy logger.js into your project
 * 2. Copy this file (express-middleware.js) into your project
 * 3. Add to your Express app
 * 
 * Usage:
 *   const express = require('express');
 *   const { createDevSmithMiddleware } = require('./express-middleware');
 *   
 *   const app = express();
 *   
 *   app.use(createDevSmithMiddleware({
 *     apiKey: process.env.DEVSMITH_API_KEY,
 *     apiUrl: process.env.DEVSMITH_API_URL,
 *     projectSlug: 'my-app',
 *     serviceName: 'express-api'
 *   }));
 *   
 *   // Your routes...
 *   app.get('/', (req, res) => res.send('Hello World'));
 *   
 *   app.listen(3000);
 */

const DevSmithLogger = require('./logger');

/**
 * Create Express middleware for automatic request/response logging
 * 
 * @param {Object} config - Logger configuration
 * @param {string} config.apiKey - DevSmith API key (dsk_...)
 * @param {string} config.apiUrl - DevSmith API URL
 * @param {string} config.projectSlug - Project slug in DevSmith
 * @param {string} config.serviceName - Service name for this app
 * @param {boolean} config.logBody - Log request/response bodies (default: false)
 * @param {Array<string>} config.skipPaths - Paths to skip logging (e.g., ['/health'])
 * @param {Array<string>} config.redactHeaders - Headers to redact (e.g., ['authorization'])
 * @returns {Function} Express middleware function
 */
function createDevSmithMiddleware(config = {}) {
  const logger = new DevSmithLogger({
    apiKey: config.apiKey,
    apiUrl: config.apiUrl || 'http://localhost:3000',
    projectSlug: config.projectSlug,
    serviceName: config.serviceName
  });

  const logBody = config.logBody !== undefined ? config.logBody : false;
  const skipPaths = config.skipPaths || [];
  const redactHeaders = config.redactHeaders || ['authorization', 'cookie', 'x-api-key'];

  return function devsmithMiddleware(req, res, next) {
    // Skip logging for specified paths (e.g., health checks)
    if (skipPaths.includes(req.path)) {
      return next();
    }

    const startTime = Date.now();

    // Capture response data
    const originalSend = res.send;
    let responseBody = null;

    res.send = function (body) {
      if (logBody && body) {
        responseBody = body;
      }
      return originalSend.call(this, body);
    };

    // Log when response finishes
    res.on('finish', () => {
      const duration = Date.now() - startTime;
      const level = res.statusCode >= 500 ? 'ERROR' :
                    res.statusCode >= 400 ? 'WARN' : 'INFO';

      const context = {
        // Request info
        method: req.method,
        path: req.path,
        query: req.query,
        ip: req.ip || req.connection.remoteAddress,
        userAgent: req.get('user-agent'),
        
        // Response info
        statusCode: res.statusCode,
        duration: `${duration}ms`,
        
        // Headers (redacted)
        requestHeaders: redactHeadersObj(req.headers, redactHeaders),
        responseHeaders: redactHeadersObj(res.getHeaders(), redactHeaders)
      };

      // Optionally include bodies
      if (logBody) {
        if (req.body) {
          context.requestBody = JSON.stringify(req.body).substring(0, 1000); // Limit to 1KB
        }
        if (responseBody) {
          context.responseBody = typeof responseBody === 'string' 
            ? responseBody.substring(0, 1000)
            : JSON.stringify(responseBody).substring(0, 1000);
        }
      }

      const message = `${req.method} ${req.path} ${res.statusCode} ${duration}ms`;
      
      if (level === 'ERROR') {
        logger.error(message, context);
      } else if (level === 'WARN') {
        logger.warn(message, context);
      } else {
        logger.info(message, context);
      }
    });

    // Log unhandled errors
    res.on('error', (error) => {
      logger.error('Response error', {
        method: req.method,
        path: req.path,
        error: error.message,
        stack: error.stack
      });
    });

    next();
  };
}

/**
 * Redact sensitive headers
 * @param {Object} headers - Headers object
 * @param {Array<string>} redactList - Headers to redact
 * @returns {Object} Headers with redacted values
 */
function redactHeadersObj(headers, redactList) {
  const redacted = { ...headers };
  redactList.forEach(key => {
    if (redacted[key]) {
      redacted[key] = '[REDACTED]';
    }
  });
  return redacted;
}

/**
 * Error handler middleware - logs uncaught errors
 * 
 * Usage:
 *   app.use(createDevSmithErrorHandler(logger));
 */
function createDevSmithErrorHandler(logger) {
  return function devsmithErrorHandler(err, req, res, next) {
    logger.error('Unhandled error', {
      method: req.method,
      path: req.path,
      error: err.message,
      stack: err.stack,
      statusCode: err.statusCode || 500
    });

    // Pass to next error handler
    next(err);
  };
}

module.exports = {
  createDevSmithMiddleware,
  createDevSmithErrorHandler
};
