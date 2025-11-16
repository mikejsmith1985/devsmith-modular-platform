# Quick Start: Cross-Repository Logging

Monitor your external applications with DevSmith's centralized logging.

---

## Overview

DevSmith can collect logs from **any application** regardless of language or framework. Simply:
1. Copy a sample logger file into your project
2. Configure with your API key
3. View logs in DevSmith Health dashboard

**Performance**: Batch processing = **100x faster** than individual log requests

---

## Step 1: Create a Project in DevSmith

1. Log into DevSmith: `http://localhost:3000`
2. Navigate to **Projects** (coming in Week 3 UI)
3. Click **Create New Project**
4. Fill in details:
   - **Name**: My Application
   - **Slug**: my-app (used in API calls)
   - **Description**: Production Node.js API
   - **Repository URL**: (optional) https://github.com/org/repo
5. Click **Create**
6. **IMPORTANT**: Copy your API key immediately (shown only once)
   - Format: `dsk_abc123xyz789...`
   - Store securely (password manager or environment variable)

---

## Step 2: Choose Your Language

### JavaScript/Node.js

**Copy the logger:**
```bash
curl -O https://raw.githubusercontent.com/mikejsmith1985/devsmith-modular-platform/main/docs/integrations/javascript/logger.js
```

Or manually copy: `docs/integrations/javascript/logger.js`

**Configure environment variables:**
```bash
# .env
DEVSMITH_API_KEY=dsk_abc123xyz789...
DEVSMITH_API_URL=http://localhost:3000
DEVSMITH_PROJECT_SLUG=my-app
DEVSMITH_SERVICE_NAME=api-server
```

**Use in your app:**
```javascript
require('dotenv').config();
const DevSmithLogger = require('./logger');

const logger = new DevSmithLogger({
  apiKey: process.env.DEVSMITH_API_KEY,
  apiUrl: process.env.DEVSMITH_API_URL,
  projectSlug: process.env.DEVSMITH_PROJECT_SLUG,
  serviceName: process.env.DEVSMITH_SERVICE_NAME
});

// Log as normal
logger.info('Server started', { port: 3000 });
logger.error('Database error', { code: 'ECONNREFUSED', host: 'localhost' });

// Logs are automatically batched and sent every 5 seconds or when 100 logs accumulated
```

**Express.js middleware example:**
```javascript
const express = require('express');
const app = express();

// Log all requests
app.use((req, res, next) => {
  logger.info('HTTP request', {
    method: req.method,
    path: req.path,
    ip: req.ip,
    userAgent: req.get('user-agent')
  });
  next();
});

// Log errors
app.use((err, req, res, next) => {
  logger.error('Unhandled error', {
    message: err.message,
    stack: err.stack,
    path: req.path
  });
  res.status(500).send('Internal Server Error');
});
```

---

### Python

**Copy the logger:**
```bash
curl -O https://raw.githubusercontent.com/mikejsmith1985/devsmith-modular-platform/main/docs/integrations/python/logger.py
```

Or manually copy: `docs/integrations/python/logger.py`

**Configure environment variables:**
```bash
# .env
DEVSMITH_API_KEY=dsk_abc123xyz789...
DEVSMITH_API_URL=http://localhost:3000
DEVSMITH_PROJECT_SLUG=my-app
DEVSMITH_SERVICE_NAME=api-server
```

**Use in your app:**
```python
import os
from logger import DevSmithLogger

logger = DevSmithLogger(
    api_key=os.getenv('DEVSMITH_API_KEY'),
    api_url=os.getenv('DEVSMITH_API_URL', 'http://localhost:3000'),
    project_slug=os.getenv('DEVSMITH_PROJECT_SLUG'),
    service_name=os.getenv('DEVSMITH_SERVICE_NAME')
)

# Log as normal
logger.info('Server started', port=3000)
logger.error('Database error', code='ECONNREFUSED', host='localhost')

# Logs are automatically batched and sent every 5 seconds or when 100 logs accumulated
```

**Flask decorator example:**
```python
from flask import Flask, request
import time

app = Flask(__name__)

@app.before_request
def log_request():
    logger.info('HTTP request', 
        method=request.method,
        path=request.path,
        ip=request.remote_addr
    )

@app.errorhandler(Exception)
def log_error(error):
    logger.error('Unhandled error',
        message=str(error),
        path=request.path
    )
    return 'Internal Server Error', 500
```

---

### Go

**Copy the logger:**
```bash
curl -O https://raw.githubusercontent.com/mikejsmith1985/devsmith-modular-platform/main/docs/integrations/go/logger.go
```

Or manually copy: `docs/integrations/go/logger.go` into your project

**Configure environment variables:**
```bash
# .env
DEVSMITH_API_KEY=dsk_abc123xyz789...
DEVSMITH_API_URL=http://localhost:3000
DEVSMITH_PROJECT_SLUG=my-app
DEVSMITH_SERVICE_NAME=api-server
```

**Use in your app:**
```go
package main

import (
    "os"
    "github.com/yourorg/yourproject/devsmithlogger"
)

func main() {
    logger := devsmithlogger.NewLogger(
        os.Getenv("DEVSMITH_API_KEY"),
        os.Getenv("DEVSMITH_API_URL"),
        os.Getenv("DEVSMITH_PROJECT_SLUG"),
        os.Getenv("DEVSMITH_SERVICE_NAME"),
    )
    defer logger.Close()

    // Log as normal
    logger.Info("Server started", map[string]interface{}{"port": 3000})
    logger.Error("Database error", map[string]interface{}{
        "code": "ECONNREFUSED",
        "host": "localhost",
    })

    // Logs are automatically batched and sent every 5 seconds or when 100 logs accumulated
}
```

**Gin middleware example:**
```go
package main

import (
    "github.com/gin-gonic/gin"
    "time"
)

func LoggingMiddleware(logger *devsmithlogger.DevSmithLogger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        logger.Info("HTTP request", map[string]interface{}{
            "method": c.Request.Method,
            "path":   c.Request.URL.Path,
            "ip":     c.ClientIP(),
        })
        
        c.Next()
        
        logger.Info("HTTP response", map[string]interface{}{
            "method":   c.Request.Method,
            "path":     c.Request.URL.Path,
            "status":   c.Writer.Status(),
            "duration": time.Since(start).Milliseconds(),
        })
    }
}

func main() {
    logger := devsmithlogger.NewLogger(
        os.Getenv("DEVSMITH_API_KEY"),
        os.Getenv("DEVSMITH_API_URL"),
        os.Getenv("DEVSMITH_PROJECT_SLUG"),
        os.Getenv("DEVSMITH_SERVICE_NAME"),
    )
    defer logger.Close()

    router := gin.Default()
    router.Use(LoggingMiddleware(logger))
    
    router.Run(":8080")
}
```

---

## Step 3: View Logs in DevSmith

1. Navigate to **Health** dashboard: `http://localhost:3000/health`
2. Use **Project Filter** to select your project: "My Application"
3. Use **Service Filter** to narrow down: "api-server"
4. View real-time logs with:
   - Level filtering (DEBUG, INFO, WARN, ERROR)
   - Search by message or context
   - AI-powered insights
   - Performance metrics

---

## Configuration Options

### Batch Size
Control how many logs are buffered before sending:

**JavaScript:**
```javascript
const logger = new DevSmithLogger({
  // ... other config
  batchSize: 50  // Send after 50 logs (default: 100)
});
```

**Python:**
```python
logger = DevSmithLogger(
    # ... other config
    batch_size=50  # Send after 50 logs (default: 100)
)
```

**Go:**
```go
logger := devsmithlogger.NewLoggerWithOptions(
    apiKey, apiURL, projectSlug, serviceName,
    50,             // batch size (default: 100)
    5*time.Second,  // flush interval (default: 5s)
)
```

### Flush Interval
Control how often logs are sent:

**JavaScript:**
```javascript
const logger = new DevSmithLogger({
  // ... other config
  flushInterval: 10000  // Send every 10 seconds (default: 5000ms)
});
```

**Python:**
```python
logger = DevSmithLogger(
    # ... other config
    flush_interval=10.0  # Send every 10 seconds (default: 5.0s)
)
```

**Go:**
```go
logger := devsmithlogger.NewLoggerWithOptions(
    apiKey, apiURL, projectSlug, serviceName,
    100,            // batch size
    10*time.Second, // Send every 10 seconds (default: 5s)
)
```

---

## Troubleshooting

### Logs not appearing in DevSmith?

**Check 1: API Key Valid**
```bash
curl -X POST http://localhost:3000/api/logs/batch \
  -H "Authorization: Bearer dsk_abc123xyz789..." \
  -H "Content-Type: application/json" \
  -d '{
    "project_slug": "my-app",
    "logs": [{"timestamp":"2025-11-11T12:00:00Z","level":"INFO","message":"Test","service":"test","context":{}}]
  }'
```

Expected: `200 OK` with `{"count": 1, "message": "Logs ingested successfully"}`

**Check 2: Network Connectivity**
```bash
curl http://localhost:3000/health
```

Expected: Health dashboard HTML

**Check 3: Check Logger Output**
Most loggers print errors to console/stderr:
- JavaScript: Check `console.error` output
- Python: Check stderr output
- Go: Check `fmt.Printf` output

**Check 4: Verify Project Slug**
Ensure `project_slug` in logger config matches project slug in DevSmith.

---

## Performance Benchmarks

**Individual Requests (WITHOUT batching):**
- 100 logs = 100 HTTP requests
- Time: ~5-10 seconds
- Database: 100 INSERT queries

**Batch Requests (WITH batching):**
- 100 logs = 1 HTTP request
- Time: ~50ms
- Database: 1 INSERT query

**Result: 100x faster with batching!**

---

## Security Best Practices

1. **Never commit API keys to git**
   - Use environment variables
   - Add `.env` to `.gitignore`

2. **Use HTTPS in production**
   - Change `apiUrl` to `https://your-devsmith.com`

3. **Rotate API keys regularly**
   - Regenerate keys every 90 days
   - DevSmith provides key regeneration in Projects page

4. **Limit API key permissions** (future feature)
   - Read-only vs write-only keys
   - Per-service keys

---

## Next Steps

- **Week 3**: Project management UI (create/edit/delete projects)
- **Week 4**: Advanced filtering and search
- **Phase 2**: Rate limiting tiers (Free/Pro/Enterprise)
- **Phase 3**: Log retention policies, S3 export

---

## Support

- **Documentation**: `docs/integrations/`
- **GitHub Issues**: https://github.com/mikejsmith1985/devsmith-modular-platform/issues
- **Architecture**: `CROSS_REPO_LOGGING_ARCHITECTURE.md`
