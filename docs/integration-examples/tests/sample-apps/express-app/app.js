const express = require('express');
require('dotenv').config();

// DevSmith Logger (copy from docs/integrations/javascript/logger.js)
const logger = {
  createLogger: function(config) {
    const buffer = [];
    let flushTimer = null;

    const flush = () => {
      if (buffer.length === 0) return;
      
      const batch = buffer.splice(0, buffer.length);
      
      fetch(config.apiUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${config.apiKey}`
        },
        body: JSON.stringify({
          project_slug: config.projectSlug,
          logs: batch
        })
      }).catch(err => console.error('DevSmith flush error:', err));
    };

    const log = (level, message, context = {}, tags = []) => {
      buffer.push({
        timestamp: new Date().toISOString(),
        level,
        message,
        service: config.serviceName,
        context,
        tags
      });

      if (buffer.length >= config.bufferSize) {
        clearTimeout(flushTimer);
        flush();
      } else if (!flushTimer) {
        flushTimer = setTimeout(() => {
          flush();
          flushTimer = null;
        }, config.flushInterval);
      }
    };

    return {
      debug: (msg, ctx, tags) => log('DEBUG', msg, ctx, tags),
      info: (msg, ctx, tags) => log('INFO', msg, ctx, tags),
      warn: (msg, ctx, tags) => log('WARN', msg, ctx, tags),
      error: (msg, ctx, tags) => log('ERROR', msg, ctx, tags),
      flush: () => flush()
    };
  }
};

// Express Middleware (copy from docs/integrations/javascript/express_middleware.js)
function devsmithMiddleware(logger, options = {}) {
  const skipPaths = options.skipPaths || [];
  const tags = options.tags || [];

  return (req, res, next) => {
    if (skipPaths.includes(req.path)) {
      return next();
    }

    const start = Date.now();
    
    logger.info('Incoming request', {
      method: req.method,
      path: req.path,
      headers: req.headers,
      query: req.query
    }, ['request', ...tags]);

    const originalSend = res.send;
    res.send = function(body) {
      const duration = Date.now() - start;
      
      logger.info('Request completed', {
        method: req.method,
        path: req.path,
        status_code: res.statusCode,
        duration_ms: duration
      }, ['response', ...tags]);

      return originalSend.call(this, body);
    };

    next();
  };
}

// Initialize logger
const devsmithLogger = logger.createLogger({
  apiUrl: process.env.DEVSMITH_API_URL,
  apiKey: process.env.DEVSMITH_API_KEY,
  projectSlug: process.env.DEVSMITH_PROJECT_SLUG,
  serviceName: process.env.DEVSMITH_SERVICE_NAME,
  bufferSize: 100,
  flushInterval: 5000
});

// Create Express app
const app = express();
app.use(express.json());

// Add DevSmith middleware (skip health checks)
app.use(devsmithMiddleware(devsmithLogger, {
  skipPaths: ['/health'],
  tags: ['production', 'api']
}));

// Routes
app.get('/', (req, res) => {
  devsmithLogger.info('Root endpoint accessed', { 
    ip: req.ip 
  }, ['endpoint', 'public']);
  
  res.json({ 
    status: 'ok', 
    message: 'DevSmith Express Sample App',
    endpoints: [
      'GET / - This page',
      'GET /health - Health check (not logged)',
      'GET /api/users - Get users list',
      'POST /api/users - Create user',
      'GET /api/error - Trigger error for testing'
    ]
  });
});

app.get('/health', (req, res) => {
  // Health check endpoint - skipped by middleware
  res.json({ status: 'healthy' });
});

app.get('/api/users', (req, res) => {
  devsmithLogger.debug('Fetching users list', {
    page: req.query.page || 1,
    limit: req.query.limit || 10
  }, ['users', 'api']);

  // Simulate database query
  const users = [
    { id: 1, name: 'Alice' },
    { id: 2, name: 'Bob' }
  ];

  res.json({ users, count: users.length });
});

app.post('/api/users', (req, res) => {
  const userData = req.body;
  
  devsmithLogger.info('Creating new user', {
    username: userData.username,
    email: userData.email
  }, ['users', 'create']);

  // Simulate validation
  if (!userData.username) {
    devsmithLogger.warn('User creation failed - missing username', {
      provided_fields: Object.keys(userData)
    }, ['validation', 'error']);

    return res.status(400).json({ error: 'Username required' });
  }

  // Simulate user creation
  const newUser = {
    id: Math.floor(Math.random() * 10000),
    ...userData,
    created_at: new Date().toISOString()
  };

  devsmithLogger.info('User created successfully', {
    user_id: newUser.id,
    username: newUser.username
  }, ['users', 'success']);

  res.status(201).json({ user: newUser });
});

app.get('/api/error', (req, res) => {
  devsmithLogger.warn('Error endpoint called - simulating error', {
    ip: req.ip
  }, ['error', 'test']);

  try {
    // Simulate error
    throw new Error('Simulated database connection error');
  } catch (err) {
    devsmithLogger.error('Application error occurred', {
      error: err.message,
      stack: err.stack,
      endpoint: '/api/error'
    }, ['error', 'exception']);

    res.status(500).json({ 
      error: 'Internal server error',
      message: err.message 
    });
  }
});

// 404 handler
app.use((req, res) => {
  devsmithLogger.warn('404 Not Found', {
    method: req.method,
    path: req.path,
    ip: req.ip
  }, ['404', 'routing']);

  res.status(404).json({ error: 'Not found' });
});

// Error handler
app.use((err, req, res, next) => {
  devsmithLogger.error('Unhandled error', {
    error: err.message,
    stack: err.stack,
    path: req.path,
    method: req.method
  }, ['error', 'unhandled']);

  res.status(500).json({ error: 'Internal server error' });
});

// Graceful shutdown
process.on('SIGTERM', () => {
  devsmithLogger.info('Received SIGTERM - flushing logs', {}, ['shutdown']);
  devsmithLogger.flush();
  setTimeout(() => process.exit(0), 1000);
});

process.on('SIGINT', () => {
  devsmithLogger.info('Received SIGINT - flushing logs', {}, ['shutdown']);
  devsmithLogger.flush();
  setTimeout(() => process.exit(0), 1000);
});

// Start server
const PORT = process.env.PORT || 3001;
app.listen(PORT, () => {
  devsmithLogger.info('Express server started', {
    port: PORT,
    env: process.env.NODE_ENV
  }, ['startup', 'server']);

  console.log(`Server running on http://localhost:${PORT}`);
  console.log('DevSmith logging enabled');
});
