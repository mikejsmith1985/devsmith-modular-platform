# Integration Guide: Cross-Repository Logging

This guide walks you through integrating your application with the DevSmith Logs platform to centralize logging across multiple repositories and services.

## 📋 Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start (5 Minutes)](#quick-start-5-minutes)
3. [Detailed Setup](#detailed-setup)
4. [Language-Specific Integration](#language-specific-integration)
5. [Verification](#verification)
6. [Best Practices](#best-practices)
7. [Troubleshooting](#troubleshooting)

---

## Prerequisites

Before you begin, ensure you have:

- ✅ Access to DevSmith Logs platform (http://localhost:3000)
- ✅ GitHub account for authentication
- ✅ Application you want to monitor (JavaScript/Python/Go)
- ✅ Network connectivity to the logs API

---

## Quick Start (5 Minutes)

### Step 1: Create a Project in DevSmith

1. Navigate to http://localhost:3000 and log in with GitHub
2. Go to **Projects** page (sidebar or http://localhost:3000/projects)
3. Click **"Create New Project"**
4. Fill in the form:
   - **Name**: Your project name (e.g., "My E-Commerce App")
   - **Slug**: URL-friendly identifier (e.g., "my-ecommerce-app")
   - **Description**: Brief description of your project
   - **Repository URL**: GitHub repository URL (optional)
5. Click **"Create Project"**

### Step 2: Copy Your API Key

⚠️ **IMPORTANT**: Your API key is shown **only once**!

After creating the project, a modal will display your API key:

```
proj_abc123def456ghi789jkl012mno345
```

**Save this key immediately** - you cannot retrieve it later!

Options:
- Click **"Copy to Clipboard"** button
- Write it down in a secure location
- Add it to your password manager

### Step 3: Download Sample Integration Code

1. Go to **Integration Docs** page (http://localhost:3000/integration-docs)
2. Select your programming language:
   - **JavaScript** (Node.js, Express, React, etc.)
   - **Python** (Flask, Django, FastAPI, etc.)
   - **Go** (Gin, Echo, standard library)
3. Copy the sample code for **Basic Setup** or **Framework Middleware**
4. Save it to your project (e.g., `utils/logs-client.js`)

### Step 4: Configure Environment Variables

Add these environment variables to your application:

```bash
# .env file
LOGS_API_URL=http://localhost:8082
LOGS_API_KEY=proj_abc123def456ghi789jkl012mno345  # Your actual key
SERVICE_NAME=my-service-name
```

**Never commit your API key to version control!**

### Step 5: Deploy and Verify

1. Start your application with the new environment variables
2. Trigger some log events (requests, errors, etc.)
3. Go to **Health Dashboard** (http://localhost:3000/health)
4. Filter by your project name in the dropdown
5. You should see your logs appearing in real-time! 🎉

---

## Detailed Setup

### Understanding the Architecture

DevSmith Logs uses a **REST API** approach for universal compatibility:

```
Your Application → HTTP POST → DevSmith Logs API → Database → Health Dashboard
```

**Key Concepts**:
- **Project**: A logical grouping of logs from one application/repository
- **API Key**: Authentication token tied to a specific project
- **Batch Ingestion**: Send multiple log entries in one request (100x faster)
- **Service Name**: Identifier for the specific service within a project

### Project Structure

When you create a project, the system generates:

```json
{
  "id": 1,
  "name": "My E-Commerce App",
  "slug": "my-ecommerce-app",
  "description": "Production e-commerce platform",
  "repository_url": "https://github.com/myorg/ecommerce",
  "api_key": "proj_abc123...",  // Shown once, then hashed
  "is_active": true,
  "created_at": "2025-11-10T12:00:00Z"
}
```

### API Key Security

API keys are:
- ✅ Hashed with bcrypt before storage (never stored in plain text)
- ✅ Used as Bearer tokens in Authorization header
- ✅ Rate limited to 1,000 requests/minute per key
- ✅ Can be regenerated if compromised (old key immediately invalidated)
- ✅ Can be deactivated to stop all logging from that project

**Best Practices**:
- Store API keys in environment variables (not code)
- Use separate projects for dev/staging/production
- Rotate keys regularly (monthly or quarterly)
- Deactivate unused projects immediately

---

## Language-Specific Integration

### JavaScript (Node.js)

#### Basic Setup (Copy-Paste Ready)

Create `utils/logs-client.js`:

```javascript
const LOGS_API_URL = process.env.LOGS_API_URL || 'http://localhost:8082';
const LOGS_API_KEY = process.env.LOGS_API_KEY;
const SERVICE_NAME = process.env.SERVICE_NAME || 'unknown-service';
const BATCH_SIZE = 100;
const FLUSH_INTERVAL = 5000; // 5 seconds

let logBuffer = [];
let flushTimer = null;

async function flushLogs() {
  if (logBuffer.length === 0) return;
  
  const batch = { entries: [...logBuffer] };
  logBuffer = [];
  
  try {
    const response = await fetch(`${LOGS_API_URL}/api/logs/batch`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${LOGS_API_KEY}`,
      },
      body: JSON.stringify(batch),
    });
    
    if (!response.ok) {
      console.error('Failed to send logs:', response.status);
    }
  } catch (error) {
    console.error('Error sending logs:', error);
  }
}

function scheduleFlush() {
  if (flushTimer) clearTimeout(flushTimer);
  flushTimer = setTimeout(flushLogs, FLUSH_INTERVAL);
}

function log(level, message, metadata = {}) {
  logBuffer.push({
    level,
    message,
    service: SERVICE_NAME,
    metadata,
    timestamp: new Date().toISOString(),
  });
  
  if (logBuffer.length >= BATCH_SIZE) {
    flushLogs();
  } else {
    scheduleFlush();
  }
}

// Public API
module.exports = {
  debug: (msg, meta) => log('DEBUG', msg, meta),
  info: (msg, meta) => log('INFO', msg, meta),
  warn: (msg, meta) => log('WARNING', msg, meta),
  error: (msg, meta) => log('ERROR', msg, meta),
  flush: flushLogs,
};

// Flush on process exit
process.on('beforeExit', flushLogs);
process.on('SIGINT', async () => {
  await flushLogs();
  process.exit(0);
});
```

#### Usage in Your Code

```javascript
const logger = require('./utils/logs-client');

// Simple logging
logger.info('User logged in', { user_id: 123 });
logger.error('Database connection failed', { error: 'ECONNREFUSED' });

// With structured metadata
logger.warn('High memory usage', {
  memory_used_mb: 850,
  threshold_mb: 1024,
  process: 'api-server',
});
```

#### Express.js Middleware

```javascript
const logger = require('./utils/logs-client');

app.use((req, res, next) => {
  const start = Date.now();
  
  res.on('finish', () => {
    const duration = Date.now() - start;
    logger.info('HTTP Request', {
      method: req.method,
      path: req.path,
      status: res.statusCode,
      duration_ms: duration,
      ip: req.ip,
    });
  });
  
  next();
});
```

---

### Python

#### Basic Setup (Copy-Paste Ready)

Create `utils/logs_client.py`:

```python
import os
import time
import json
import threading
import requests
from datetime import datetime
from typing import Dict, Any

LOGS_API_URL = os.getenv('LOGS_API_URL', 'http://localhost:8082')
LOGS_API_KEY = os.getenv('LOGS_API_KEY')
SERVICE_NAME = os.getenv('SERVICE_NAME', 'unknown-service')
BATCH_SIZE = 100
FLUSH_INTERVAL = 5  # seconds

class LogsClient:
    def __init__(self):
        self.buffer = []
        self.lock = threading.Lock()
        self.flush_timer = None
        self.start_flush_timer()
    
    def start_flush_timer(self):
        if self.flush_timer:
            self.flush_timer.cancel()
        self.flush_timer = threading.Timer(FLUSH_INTERVAL, self.flush)
        self.flush_timer.daemon = True
        self.flush_timer.start()
    
    def flush(self):
        with self.lock:
            if not self.buffer:
                self.start_flush_timer()
                return
            
            batch = {'entries': list(self.buffer)}
            self.buffer.clear()
        
        try:
            response = requests.post(
                f'{LOGS_API_URL}/api/logs/batch',
                json=batch,
                headers={
                    'Content-Type': 'application/json',
                    'Authorization': f'Bearer {LOGS_API_KEY}',
                },
                timeout=5,
            )
            response.raise_for_status()
        except Exception as e:
            print(f'Error sending logs: {e}')
        
        self.start_flush_timer()
    
    def log(self, level: str, message: str, metadata: Dict[str, Any] = None):
        entry = {
            'level': level,
            'message': message,
            'service': SERVICE_NAME,
            'metadata': metadata or {},
            'timestamp': datetime.utcnow().isoformat() + 'Z',
        }
        
        with self.lock:
            self.buffer.append(entry)
            should_flush = len(self.buffer) >= BATCH_SIZE
        
        if should_flush:
            self.flush()
    
    def debug(self, message: str, metadata: Dict[str, Any] = None):
        self.log('DEBUG', message, metadata)
    
    def info(self, message: str, metadata: Dict[str, Any] = None):
        self.log('INFO', message, metadata)
    
    def warn(self, message: str, metadata: Dict[str, Any] = None):
        self.log('WARNING', message, metadata)
    
    def error(self, message: str, metadata: Dict[str, Any] = None):
        self.log('ERROR', message, metadata)

# Global instance
logger = LogsClient()
```

#### Usage in Your Code

```python
from utils.logs_client import logger

# Simple logging
logger.info('User logged in', {'user_id': 123})
logger.error('Database connection failed', {'error': 'Connection refused'})

# With structured metadata
logger.warn('High memory usage', {
    'memory_used_mb': 850,
    'threshold_mb': 1024,
    'process': 'api-server',
})
```

#### Flask Middleware

```python
from flask import Flask, request, g
from utils.logs_client import logger
import time

app = Flask(__name__)

@app.before_request
def before_request():
    g.start_time = time.time()

@app.after_request
def after_request(response):
    if hasattr(g, 'start_time'):
        duration = (time.time() - g.start_time) * 1000
        logger.info('HTTP Request', {
            'method': request.method,
            'path': request.path,
            'status': response.status_code,
            'duration_ms': duration,
            'ip': request.remote_addr,
        })
    return response
```

---

### Go

#### Basic Setup (Copy-Paste Ready)

Create `pkg/logs/client.go`:

```go
package logs

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "sync"
    "time"
)

const (
    defaultBatchSize     = 100
    defaultFlushInterval = 5 * time.Second
)

type LogEntry struct {
    Level     string                 `json:"level"`
    Message   string                 `json:"message"`
    Service   string                 `json:"service"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Timestamp string                 `json:"timestamp"`
}

type LogsClient struct {
    apiURL        string
    apiKey        string
    serviceName   string
    buffer        []LogEntry
    mu            sync.Mutex
    flushTimer    *time.Timer
    httpClient    *http.Client
}

func NewLogsClient() *LogsClient {
    client := &LogsClient{
        apiURL:      os.Getenv("LOGS_API_URL"),
        apiKey:      os.Getenv("LOGS_API_KEY"),
        serviceName: os.Getenv("SERVICE_NAME"),
        buffer:      make([]LogEntry, 0, defaultBatchSize),
        httpClient:  &http.Client{Timeout: 5 * time.Second},
    }
    
    if client.apiURL == "" {
        client.apiURL = "http://localhost:8082"
    }
    if client.serviceName == "" {
        client.serviceName = "unknown-service"
    }
    
    client.startFlushTimer()
    return client
}

func (c *LogsClient) startFlushTimer() {
    c.flushTimer = time.AfterFunc(defaultFlushInterval, func() {
        c.Flush()
        c.startFlushTimer()
    })
}

func (c *LogsClient) Flush() {
    c.mu.Lock()
    if len(c.buffer) == 0 {
        c.mu.Unlock()
        return
    }
    
    batch := map[string]interface{}{
        "entries": c.buffer,
    }
    c.buffer = make([]LogEntry, 0, defaultBatchSize)
    c.mu.Unlock()
    
    payload, err := json.Marshal(batch)
    if err != nil {
        fmt.Printf("Error marshaling logs: %v\n", err)
        return
    }
    
    req, err := http.NewRequest("POST", c.apiURL+"/api/logs/batch", bytes.NewBuffer(payload))
    if err != nil {
        fmt.Printf("Error creating request: %v\n", err)
        return
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.apiKey)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        fmt.Printf("Error sending logs: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        fmt.Printf("Failed to send logs: %d\n", resp.StatusCode)
    }
}

func (c *LogsClient) log(level, message string, metadata map[string]interface{}) {
    entry := LogEntry{
        Level:     level,
        Message:   message,
        Service:   c.serviceName,
        Metadata:  metadata,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }
    
    c.mu.Lock()
    c.buffer = append(c.buffer, entry)
    shouldFlush := len(c.buffer) >= defaultBatchSize
    c.mu.Unlock()
    
    if shouldFlush {
        c.Flush()
    }
}

func (c *LogsClient) Debug(message string, metadata map[string]interface{}) {
    c.log("DEBUG", message, metadata)
}

func (c *LogsClient) Info(message string, metadata map[string]interface{}) {
    c.log("INFO", message, metadata)
}

func (c *LogsClient) Warn(message string, metadata map[string]interface{}) {
    c.log("WARNING", message, metadata)
}

func (c *LogsClient) Error(message string, metadata map[string]interface{}) {
    c.log("ERROR", message, metadata)
}
```

#### Usage in Your Code

```go
package main

import (
    "yourapp/pkg/logs"
)

var logger *logs.LogsClient

func main() {
    logger = logs.NewLogsClient()
    defer logger.Flush()
    
    // Simple logging
    logger.Info("User logged in", map[string]interface{}{
        "user_id": 123,
    })
    
    logger.Error("Database connection failed", map[string]interface{}{
        "error": "connection refused",
    })
}
```

#### Gin Middleware

```go
func LoggingMiddleware(logger *logs.LogsClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Milliseconds()
        logger.Info("HTTP Request", map[string]interface{}{
            "method":      c.Request.Method,
            "path":        c.Request.URL.Path,
            "status":      c.Writer.Status(),
            "duration_ms": duration,
            "ip":          c.ClientIP(),
        })
    }
}

// In main.go
router := gin.Default()
router.Use(LoggingMiddleware(logger))
```

---

## Verification

### Check Logs in Dashboard

1. Go to http://localhost:3000/health
2. Click **Project** filter dropdown
3. Select your project name
4. You should see logs appearing with:
   - ✅ Correct log level (DEBUG/INFO/WARNING/ERROR)
   - ✅ Your service name
   - ✅ Timestamps
   - ✅ Metadata fields

### Verify Batch Ingestion

Check that logs are being sent in batches (not individually):

```bash
# Monitor network traffic (optional)
tcpdump -i any -A 'host localhost and port 8082' | grep -A 10 "POST /api/logs/batch"

# You should see JSON payloads with multiple entries:
# {"entries":[...100 logs...]}
```

### Test API Key

```bash
curl -X POST http://localhost:8082/api/logs/batch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "entries": [
      {
        "level": "INFO",
        "message": "Test log from curl",
        "service": "test-service",
        "metadata": {"test": true}
      }
    ]
  }'

# Expected: {"inserted": 1}
```

---

## Best Practices

### 1. Batch Size and Flush Interval

**Recommended Settings**:
- **Batch Size**: 100-1000 logs (default: 100)
- **Flush Interval**: 5-10 seconds (default: 5s)
- **Max Batch Size**: 1000 logs (API limit)

**Tuning for Your Workload**:

| Scenario | Batch Size | Flush Interval |
|----------|------------|----------------|
| Low traffic (<10 req/sec) | 50 | 10s |
| Medium traffic (10-100 req/sec) | 100 | 5s |
| High traffic (100-1000 req/sec) | 500 | 2s |
| Very high traffic (>1000 req/sec) | 1000 | 1s |

### 2. Metadata Structure

**Good Metadata**:
```json
{
  "user_id": 123,
  "action": "purchase",
  "amount": 99.99,
  "currency": "USD",
  "cart_items": 3
}
```

**Bad Metadata** (avoid):
```json
{
  "data": "user=123,action=purchase,amount=99.99",  // Hard to query
  "huge_json": {...},  // >10KB of nested data
  "password": "secret123"  // Never log sensitive data!
}
```

### 3. Error Handling

Always handle logging failures gracefully:

```javascript
// ❌ Bad: Crashes if logging fails
logger.error('Something failed');

// ✅ Good: Catches logging errors
try {
  logger.error('Something failed', metadata);
} catch (err) {
  console.error('Failed to send log:', err);
  // Application continues running
}
```

### 4. Rate Limiting

**API Limits**:
- 1,000 requests/minute per API key
- 1,000 logs per batch maximum

**Handling Rate Limits**:
```javascript
// Implement exponential backoff
async function sendLogsWithRetry(batch, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    const response = await fetch(url, options);
    
    if (response.status === 429) {
      // Rate limited - wait and retry
      const delay = Math.pow(2, i) * 1000; // 1s, 2s, 4s
      await new Promise(resolve => setTimeout(resolve, delay));
      continue;
    }
    
    return response;
  }
}
```

### 5. Security

**DO**:
- ✅ Store API keys in environment variables
- ✅ Use HTTPS in production
- ✅ Rotate API keys regularly
- ✅ Deactivate compromised keys immediately
- ✅ Use separate projects for dev/staging/prod

**DON'T**:
- ❌ Commit API keys to version control
- ❌ Log sensitive data (passwords, credit cards, tokens)
- ❌ Share API keys across teams
- ❌ Use same key for dev and production

---

## Troubleshooting

See **[TROUBLESHOOTING_GUIDE.md](./TROUBLESHOOTING_GUIDE.md)** for detailed solutions to common issues.

**Quick Checks**:

1. **Logs not appearing?**
   - Check `LOGS_API_KEY` is set correctly
   - Verify API URL is reachable: `curl http://localhost:8082/api/health`
   - Check project is active in Projects page

2. **401 Unauthorized?**
   - Verify Bearer token format: `Authorization: Bearer YOUR_KEY`
   - Check API key hasn't been regenerated
   - Ensure project isn't deactivated

3. **429 Too Many Requests?**
   - You've hit rate limit (1000 req/min)
   - Increase batch size to reduce request count
   - Implement exponential backoff retry logic

4. **400 Bad Request?**
   - Check JSON structure matches API format
   - Verify `entries` array exists
   - Ensure batch size ≤ 1000 logs

---

## Performance Benchmarks

**Expected Throughput**:
- Individual requests: ~140 logs/second
- Batch ingestion (100 logs): ~14,000 logs/second (100x improvement)
- Batch ingestion (1000 logs): ~33,000 logs/second (235x improvement)

**Latency**:
- p50: <100ms
- p95: <500ms
- p99: <1000ms

---

## Next Steps

- 📖 Read [TROUBLESHOOTING_GUIDE.md](./TROUBLESHOOTING_GUIDE.md) for solutions to common issues
- 🔍 Explore the **Integration Docs** page for more code examples
- 📊 Use the **Health Dashboard** to analyze your logs
- 🔐 Manage your projects and API keys in the **Projects** page

---

## Support

If you encounter issues not covered in this guide:

1. Check [TROUBLESHOOTING_GUIDE.md](./TROUBLESHOOTING_GUIDE.md)
2. Review logs in your application
3. Check DevSmith Logs service logs: `docker logs devsmith-modular-platform-logs-1`
4. Create an issue in the GitHub repository

---

**Last Updated**: 2025-11-11  
**Version**: 1.0
