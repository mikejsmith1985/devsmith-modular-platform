# DevSmith Platform - API Integration Guide

**Version:** 1.0 Beta  
**Last Updated:** November 12, 2025  
**Audience:** Developers integrating external applications

---

## 📖 Overview

The DevSmith Platform **Projects API** allows you to send logs from any application (Node.js, Go, Python, Java, etc.) to DevSmith for centralized monitoring, AI-powered diagnostics, and analytics.

**Key Features:**
- **Batch Ingestion:** Send up to 1000 logs per request (100x faster than individual requests)
- **Language Agnostic:** Works with any language that can make HTTP requests
- **Simple Authentication:** API key-based (no OAuth required for external apps)
- **AI Diagnostics:** Automatic error pattern detection and root cause analysis
- **Project Isolation:** Each project has its own logs, API key, and settings

---

## 🚀 Quick Start (5 Minutes)

### Step 1: Create a Project

1. Log in to DevSmith: **http://localhost:3000**
2. Click **"Projects"** from Dashboard
3. Click **"Create Project"** button
4. Fill in details:
   - **Name:** "My Node.js App" (descriptive name)
   - **Slug:** "my-nodejs-app" (unique identifier, lowercase, hyphens only)
5. Click **"Create"**
6. **Copy your API Key** - you'll need this! (Format: `dsk_xxxxxxxxxxxxx`)

⚠️ **Keep your API key secure!** Treat it like a password. Don't commit it to version control.

### Step 2: Install DevSmith Client (Optional)

We provide official client libraries for popular languages:

#### Node.js / TypeScript
```bash
npm install @devsmith/logger
# or
yarn add @devsmith/logger
```

#### Go
```bash
go get github.com/mikejsmith1985/devsmith-go-client
```

#### Python
```bash
pip install devsmith-logger
```

**Don't see your language?** No problem! See [Manual HTTP Integration](#manual-http-integration) below.

### Step 3: Send Your First Log

#### Node.js
```javascript
const { DevSmithLogger } = require('@devsmith/logger');

const logger = new DevSmithLogger({
  apiKey: 'dsk_your_api_key_here',
  projectSlug: 'my-nodejs-app',
  endpoint: 'http://localhost:3000/api/logs/batch' // or your domain
});

// Log messages
logger.info('Application started successfully');
logger.error('Database connection failed', { 
  error: 'ECONNREFUSED',
  host: 'localhost:5432'
});

// Flush logs (sends batch to DevSmith)
await logger.flush();
```

#### Go
```go
import "github.com/mikejsmith1985/devsmith-go-client"

logger := devsmith.NewLogger(devsmith.Config{
    APIKey:      "dsk_your_api_key_here",
    ProjectSlug: "my-golang-app",
    Endpoint:    "http://localhost:3000/api/logs/batch",
})

logger.Info("Application started successfully")
logger.Error("Database connection failed", map[string]interface{}{
    "error": "connection refused",
    "host":  "localhost:5432",
})

// Flush logs
logger.Flush()
```

#### Python
```python
from devsmith_logger import DevSmithLogger

logger = DevSmithLogger(
    api_key='dsk_your_api_key_here',
    project_slug='my-python-app',
    endpoint='http://localhost:3000/api/logs/batch'
)

logger.info('Application started successfully')
logger.error('Database connection failed', {
    'error': 'connection refused',
    'host': 'localhost:5432'
})

# Flush logs
logger.flush()
```

### Step 4: View Logs in DevSmith

1. Go to **http://localhost:3000/health**
2. Select your project from dropdown
3. View logs in real-time with AI-powered insights!

---

## 🔌 Manual HTTP Integration

If we don't have a client library for your language, you can integrate directly via HTTP.

### API Endpoint

```
POST http://localhost:3000/api/logs/batch
```

**For production:** Replace `localhost:3000` with your DevSmith domain

### Authentication

Include your API key in the `X-API-Key` header:

```http
X-API-Key: dsk_your_api_key_here
```

### Request Format

**Content-Type:** `application/json`

**Body:**
```json
{
  "project_slug": "my-app",
  "logs": [
    {
      "timestamp": "2025-11-12T14:30:00Z",
      "level": "info",
      "message": "User logged in successfully",
      "service_name": "auth-service",
      "context": {
        "user_id": 12345,
        "ip_address": "192.168.1.100"
      }
    },
    {
      "timestamp": "2025-11-12T14:30:15Z",
      "level": "error",
      "message": "Failed to send email notification",
      "service_name": "notification-service",
      "context": {
        "error": "SMTP timeout",
        "recipient": "user@example.com"
      }
    }
  ]
}
```

### Request Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `project_slug` | string | ✅ Yes | Project identifier (from Step 1) |
| `logs` | array | ✅ Yes | Array of log entries (max 1000 per request) |
| `logs[].timestamp` | string | ✅ Yes | ISO 8601 timestamp (UTC recommended) |
| `logs[].level` | string | ✅ Yes | Log level: `debug`, `info`, `warn`, `error`, `fatal` |
| `logs[].message` | string | ✅ Yes | Log message (max 10,000 characters) |
| `logs[].service_name` | string | ❌ No | Service/component name (e.g., "api-server", "worker") |
| `logs[].context` | object | ❌ No | Additional metadata (JSON object, max 50 fields) |

### Response Format

**Success (200 OK):**
```json
{
  "accepted": 2,
  "message": "Successfully ingested 2 log entries"
}
```

**Error (4xx/5xx):**
```json
{
  "error": "Invalid API key",
  "message": "API key is required. Get your key from DevSmith Portal."
}
```

### Example: cURL

```bash
curl -X POST http://localhost:3000/api/logs/batch \
  -H "X-API-Key: dsk_your_api_key_here" \
  -H "Content-Type: application/json" \
  -d '{
    "project_slug": "my-app",
    "logs": [
      {
        "timestamp": "2025-11-12T14:30:00Z",
        "level": "info",
        "message": "Application started",
        "service_name": "main",
        "context": {"version": "1.0.0"}
      }
    ]
  }'
```

---

## 🌍 Complete Integration Examples

### Node.js + Express

```javascript
// logger.js - Singleton DevSmith logger
const { DevSmithLogger } = require('@devsmith/logger');

const logger = new DevSmithLogger({
  apiKey: process.env.DEVSMITH_API_KEY,
  projectSlug: process.env.DEVSMITH_PROJECT_SLUG,
  endpoint: process.env.DEVSMITH_ENDPOINT || 'http://localhost:3000/api/logs/batch',
  flushInterval: 5000, // Auto-flush every 5 seconds
  batchSize: 100 // Or when 100 logs queued
});

// Graceful shutdown - flush remaining logs
process.on('SIGTERM', async () => {
  await logger.flush();
  process.exit(0);
});

module.exports = logger;

// app.js - Use in Express app
const express = require('express');
const logger = require('./logger');

const app = express();

// Log all requests
app.use((req, res, next) => {
  logger.info(`${req.method} ${req.path}`, {
    method: req.method,
    path: req.path,
    ip: req.ip,
    user_agent: req.get('user-agent')
  });
  next();
});

// Error handling middleware
app.use((err, req, res, next) => {
  logger.error('Request failed', {
    error: err.message,
    stack: err.stack,
    path: req.path,
    method: req.method
  });
  res.status(500).json({ error: 'Internal server error' });
});

app.listen(3000, () => {
  logger.info('Server started', { port: 3000 });
});
```

### Go + Gin

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/mikejsmith1985/devsmith-go-client"
)

var logger *devsmith.Logger

func main() {
    // Initialize DevSmith logger
    logger = devsmith.NewLogger(devsmith.Config{
        APIKey:      os.Getenv("DEVSMITH_API_KEY"),
        ProjectSlug: os.Getenv("DEVSMITH_PROJECT_SLUG"),
        Endpoint:    getEnvOrDefault("DEVSMITH_ENDPOINT", "http://localhost:3000/api/logs/batch"),
        FlushInterval: 5 * time.Second,
        BatchSize:   100,
    })

    // Graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        logger.Flush()
        os.Exit(0)
    }()

    // Setup Gin
    r := gin.Default()

    // Request logging middleware
    r.Use(func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        logger.Info("Request completed", map[string]interface{}{
            "method":      c.Request.Method,
            "path":        c.Request.URL.Path,
            "status":      c.Writer.Status(),
            "duration_ms": time.Since(start).Milliseconds(),
            "ip":          c.ClientIP(),
        })
    })

    // Error handling
    r.Use(gin.Recovery())
    r.Use(func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                logger.Error("Panic recovered", map[string]interface{}{
                    "error": err,
                    "path":  c.Request.URL.Path,
                })
                c.JSON(500, gin.H{"error": "Internal server error"})
            }
        }()
        c.Next()
    })

    // Routes
    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello World"})
    })

    logger.Info("Server starting", map[string]interface{}{"port": 8080})
    r.Run(":8080")
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### Python + Flask

```python
# logger.py - Singleton DevSmith logger
import os
import atexit
from devsmith_logger import DevSmithLogger

logger = DevSmithLogger(
    api_key=os.getenv('DEVSMITH_API_KEY'),
    project_slug=os.getenv('DEVSMITH_PROJECT_SLUG'),
    endpoint=os.getenv('DEVSMITH_ENDPOINT', 'http://localhost:3000/api/logs/batch'),
    flush_interval=5,  # Auto-flush every 5 seconds
    batch_size=100
)

# Flush on exit
atexit.register(logger.flush)

# app.py - Use in Flask app
from flask import Flask, request, jsonify
from logger import logger
import traceback

app = Flask(__name__)

# Request logging middleware
@app.before_request
def log_request():
    logger.info(f'{request.method} {request.path}', {
        'method': request.method,
        'path': request.path,
        'ip': request.remote_addr,
        'user_agent': request.headers.get('User-Agent')
    })

# Error handling
@app.errorhandler(Exception)
def handle_exception(e):
    logger.error('Request failed', {
        'error': str(e),
        'stack': traceback.format_exc(),
        'path': request.path,
        'method': request.method
    })
    return jsonify({'error': 'Internal server error'}), 500

@app.route('/')
def index():
    return jsonify({'message': 'Hello World'})

if __name__ == '__main__':
    logger.info('Server started', {'port': 5000})
    app.run(port=5000)
```

### Java + Spring Boot

```java
// DevSmithLogger.java - Logger implementation
package com.example.devsmith;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.scheduling.annotation.Scheduled;
import javax.annotation.PreDestroy;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.net.URI;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import com.fasterxml.jackson.databind.ObjectMapper;

@Component
public class DevSmithLogger {
    
    @Value("${devsmith.api-key}")
    private String apiKey;
    
    @Value("${devsmith.project-slug}")
    private String projectSlug;
    
    @Value("${devsmith.endpoint}")
    private String endpoint;
    
    private final List<LogEntry> buffer = new ArrayList<>();
    private final HttpClient httpClient = HttpClient.newHttpClient();
    private final ObjectMapper objectMapper = new ObjectMapper();
    private final int BATCH_SIZE = 100;
    
    public void info(String message, Map<String, Object> context) {
        addLog("info", message, context);
    }
    
    public void error(String message, Map<String, Object> context) {
        addLog("error", message, context);
    }
    
    private synchronized void addLog(String level, String message, Map<String, Object> context) {
        buffer.add(new LogEntry(
            Instant.now().toString(),
            level,
            message,
            "spring-boot-app",
            context
        ));
        
        if (buffer.size() >= BATCH_SIZE) {
            flush();
        }
    }
    
    @Scheduled(fixedDelay = 5000) // Auto-flush every 5 seconds
    public synchronized void flush() {
        if (buffer.isEmpty()) return;
        
        try {
            Map<String, Object> payload = Map.of(
                "project_slug", projectSlug,
                "logs", new ArrayList<>(buffer)
            );
            
            String json = objectMapper.writeValueAsString(payload);
            
            HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(endpoint))
                .header("X-API-Key", apiKey)
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(json))
                .build();
            
            httpClient.sendAsync(request, HttpResponse.BodyHandlers.ofString())
                .thenAccept(response -> {
                    if (response.statusCode() != 200) {
                        System.err.println("DevSmith log flush failed: " + response.body());
                    }
                });
            
            buffer.clear();
        } catch (Exception e) {
            System.err.println("Failed to flush logs to DevSmith: " + e.getMessage());
        }
    }
    
    @PreDestroy
    public void shutdown() {
        flush();
    }
    
    private record LogEntry(
        String timestamp,
        String level,
        String message,
        String service_name,
        Map<String, Object> context
    ) {}
}

// application.properties
# devsmith.api-key=dsk_your_api_key_here
# devsmith.project-slug=my-spring-app
# devsmith.endpoint=http://localhost:3000/api/logs/batch

// Usage in Controller
@RestController
public class MyController {
    
    @Autowired
    private DevSmithLogger logger;
    
    @GetMapping("/")
    public ResponseEntity<Map<String, String>> index() {
        logger.info("Home endpoint accessed", Map.of("method", "GET"));
        return ResponseEntity.ok(Map.of("message", "Hello World"));
    }
    
    @ExceptionHandler(Exception.class)
    public ResponseEntity<Map<String, String>> handleException(Exception e) {
        logger.error("Request failed", Map.of(
            "error", e.getMessage(),
            "stack", Arrays.toString(e.getStackTrace())
        ));
        return ResponseEntity.status(500)
            .body(Map.of("error", "Internal server error"));
    }
}
```

---

## ⚙️ Best Practices

### 1. Asynchronous Logging

**DO:** Log asynchronously to avoid blocking your application
```javascript
// Good - Non-blocking
logger.info('User created', { user_id: 123 });
// App continues immediately
```

**DON'T:** Wait for log flush in critical path
```javascript
// Bad - Blocks for up to 5 seconds!
await logger.flush();
res.json({ success: true });
```

### 2. Structured Logging

**DO:** Use the `context` field for structured data
```javascript
logger.error('Database query failed', {
  query: 'SELECT * FROM users',
  error: 'Connection timeout',
  duration_ms: 5000,
  retry_count: 3
});
```

**DON'T:** Stringify context into message
```javascript
// Bad - Harder to search/filter
logger.error('Database query failed: SELECT * FROM users, error: Connection timeout');
```

### 3. Sensitive Data

**DO:** Sanitize sensitive data before logging
```javascript
logger.info('User logged in', {
  user_id: 123,
  email: 'user@example.com',
  // DON'T log: password, credit card, SSN, etc.
});
```

**DON'T:** Log passwords, API keys, tokens, etc.

### 4. Batch Size & Frequency

**Recommended Settings:**
- **Batch Size:** 100-500 logs
- **Flush Interval:** 5-10 seconds
- **Max Batch Size:** 1000 logs (API limit)

**High-Volume Apps (>1000 logs/sec):**
```javascript
const logger = new DevSmithLogger({
  batchSize: 1000,      // Max batch size
  flushInterval: 1000,  // Flush every 1 second
  maxRetries: 3         // Retry failed requests
});
```

**Low-Volume Apps (<10 logs/sec):**
```javascript
const logger = new DevSmithLogger({
  batchSize: 50,        // Smaller batches
  flushInterval: 30000, // Flush every 30 seconds
});
```

### 5. Error Handling

Always handle flush errors gracefully:

```javascript
logger.on('error', (err) => {
  console.error('DevSmith logging failed:', err.message);
  // Fallback to local file logging
  fs.appendFileSync('fallback.log', JSON.stringify(err.logs));
});
```

### 6. Connection Pooling

For high-throughput applications, use connection pooling:

```javascript
const logger = new DevSmithLogger({
  http: {
    keepAlive: true,
    maxSockets: 10,
    timeout: 5000
  }
});
```

### 7. Graceful Shutdown

Always flush logs on shutdown:

```javascript
process.on('SIGTERM', async () => {
  console.log('Shutting down...');
  await logger.flush();
  process.exit(0);
});

process.on('SIGINT', async () => {
  console.log('Shutting down...');
  await logger.flush();
  process.exit(0);
});
```

---

## 🔒 Security Best Practices

### API Key Management

**DO:**
- Store API keys in environment variables
- Use different API keys for dev/staging/prod
- Rotate API keys periodically (every 90 days)
- Use secret management tools (AWS Secrets Manager, HashiCorp Vault)

**DON'T:**
- Commit API keys to version control
- Share API keys in chat/email
- Hardcode API keys in source code
- Reuse API keys across projects

### Network Security

**Production Checklist:**
- ✅ Use HTTPS endpoint (not HTTP)
- ✅ Validate SSL certificates
- ✅ Use firewall rules to restrict access
- ✅ Enable rate limiting (100 req/min default)

### Access Control

**Project-Level Isolation:**
- Each API key is scoped to ONE project
- Logs from one project cannot be accessed by another
- Deleting a project revokes its API key immediately

---

## 📊 Rate Limits

### Current Limits (Beta)

| Limit Type | Value | Scope |
|------------|-------|-------|
| **Requests per minute** | 100 | Per API key |
| **Logs per request** | 1,000 | Per batch |
| **Request size** | 1 MB | Per batch |
| **Message size** | 10 KB | Per log entry |
| **Context fields** | 50 | Per log entry |

### Rate Limit Headers

Responses include rate limit information:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1699876543
```

### Handling Rate Limits

**429 Too Many Requests Response:**
```json
{
  "error": "Rate limit exceeded",
  "retry_after": 60
}
```

**Retry Strategy:**
```javascript
async function sendLogsWithRetry(logs, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'X-API-Key': apiKey,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ project_slug, logs })
      });
      
      if (response.status === 429) {
        const retryAfter = response.headers.get('Retry-After') || 60;
        console.log(`Rate limited, retrying in ${retryAfter}s`);
        await sleep(retryAfter * 1000);
        continue;
      }
      
      if (response.ok) {
        return await response.json();
      }
      
      throw new Error(`HTTP ${response.status}: ${await response.text()}`);
    } catch (err) {
      if (i === maxRetries - 1) throw err;
      await sleep(Math.pow(2, i) * 1000); // Exponential backoff
    }
  }
}
```

---

## 🧪 Testing Your Integration

### 1. Test API Key

```bash
curl -X POST http://localhost:3000/api/logs/batch \
  -H "X-API-Key: dsk_your_api_key_here" \
  -H "Content-Type: application/json" \
  -d '{
    "project_slug": "test-project",
    "logs": [
      {
        "timestamp": "2025-11-12T14:00:00Z",
        "level": "info",
        "message": "Test log from API integration",
        "service_name": "test"
      }
    ]
  }'
```

**Expected:** `200 OK` with `{"accepted": 1, ...}`

### 2. Test Invalid API Key

```bash
curl -X POST http://localhost:3000/api/logs/batch \
  -H "X-API-Key: invalid_key" \
  -H "Content-Type: application/json" \
  -d '{"project_slug": "test", "logs": []}'
```

**Expected:** `401 Unauthorized` with error message

### 3. Test Batch Ingestion

```bash
# Generate 100 test logs
node -e "
const logs = Array.from({length: 100}, (_, i) => ({
  timestamp: new Date().toISOString(),
  level: 'info',
  message: \`Test log \${i+1}\`,
  service_name: 'test'
}));
console.log(JSON.stringify({project_slug: 'test', logs}));
" | curl -X POST http://localhost:3000/api/logs/batch \
  -H "X-API-Key: dsk_your_api_key_here" \
  -H "Content-Type: application/json" \
  -d @-
```

**Expected:** `200 OK` with `{"accepted": 100, ...}`

### 4. Verify Logs in UI

1. Go to http://localhost:3000/health
2. Select your project
3. You should see your test logs!

---

## 🆘 Troubleshooting

### "Invalid API key" Error

**Cause:** API key is incorrect or project was deleted

**Solution:**
1. Go to http://localhost:3000/projects
2. Verify project exists
3. Copy API key again (click "Show Key")
4. Update your application configuration

### "Project not found" Error

**Cause:** `project_slug` doesn't match any project

**Solution:**
1. Check project slug in DevSmith UI
2. Slugs are case-sensitive and use hyphens (e.g., "my-app", not "My App")

### Logs Not Appearing

**Possible Causes:**

1. **Logs not flushed yet**
   - Solution: Wait for auto-flush (5-10 seconds) or call `logger.flush()`

2. **Wrong project selected in UI**
   - Solution: Select correct project from dropdown

3. **Invalid timestamp format**
   - Solution: Use ISO 8601 format: `2025-11-12T14:30:00Z`

4. **Network connectivity**
   - Solution: Verify DevSmith is reachable: `curl http://localhost:3000/health`

### High Latency

**If logging is slow:**

1. **Increase batch size:**
   ```javascript
   batchSize: 500 // Instead of 100
   ```

2. **Reduce flush interval:**
   ```javascript
   flushInterval: 10000 // 10 seconds instead of 5
   ```

3. **Use connection pooling:**
   ```javascript
   http: { keepAlive: true, maxSockets: 10 }
   ```

4. **Check network latency:**
   ```bash
   time curl http://localhost:3000/health
   ```

---

## 📚 API Reference

### POST /api/logs/batch

Ingest logs in batch.

**Authentication:** X-API-Key header

**Request:**
- **Method:** POST
- **Content-Type:** application/json
- **Body:** See [Request Format](#request-format)

**Response:**
- **200 OK:** Logs accepted
- **400 Bad Request:** Invalid request format
- **401 Unauthorized:** Missing or invalid API key
- **403 Forbidden:** Project disabled
- **429 Too Many Requests:** Rate limit exceeded
- **500 Internal Server Error:** Server error

**Rate Limit:** 100 requests/minute per API key

---

## 🔗 Additional Resources

- **[DEPLOYMENT.md](./DEPLOYMENT.md)** - Setup DevSmith Platform
- **[USER_GUIDE.md](./USER_GUIDE.md)** - Using DevSmith web interface
- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Technical architecture

### Client Libraries

- **Node.js:** https://github.com/mikejsmith1985/devsmith-node-client
- **Go:** https://github.com/mikejsmith1985/devsmith-go-client
- **Python:** https://github.com/mikejsmith1985/devsmith-python-client
- **Java:** https://github.com/mikejsmith1985/devsmith-java-client

### Support

- **GitHub Issues:** https://github.com/mikejsmith1985/devsmith-modular-platform/issues
- **Discussions:** https://github.com/mikejsmith1985/devsmith-modular-platform/discussions

---

**🎉 Happy logging!** Your insights are just an API call away.
