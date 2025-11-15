# Manual Testing Guide for Beta Users

This guide helps beta users integrate DevSmith cross-repo logging into their external applications and validate functionality.

## Prerequisites

### 1. DevSmith Account Setup
- Create account at DevSmith platform (http://localhost:3000 or production URL)
- Navigate to Projects page
- Click "Create Project"
- Fill in project details:
  - **Name**: Your project name (e.g., "My Node.js App")
  - **Slug**: URL-friendly identifier (e.g., "my-nodejs-app")
  - **Description**: Brief description of your project

### 2. Generate API Key
- After creating project, click "Show API Key"
- **IMPORTANT**: Copy the API key immediately - it's only shown once
- Store securely (environment variable, secrets manager)
- Format: `devsmith_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX`

### 3. Note Your Configuration
Write down these values for integration:
```
API Endpoint: http://localhost:8082/api/logs/batch (or production URL)
API Key: devsmith_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
Project Slug: my-nodejs-app
Service Name: your-service-name (e.g., "api-server", "worker")
```

---

## Integration Testing

### JavaScript/Node.js Integration

#### Step 1: Copy Sample Logger
```bash
# Download sample logger
curl -o logger.js https://raw.githubusercontent.com/YOUR_REPO/docs/integrations/javascript/logger.js

# Or copy from docs/integrations/javascript/logger.js
```

#### Step 2: Configure in Your App
```javascript
// config.js or .env
const logger = require('./logger');

const devsmithLogger = logger.createLogger({
  apiUrl: 'http://localhost:8082/api/logs/batch',
  apiKey: process.env.DEVSMITH_API_KEY,
  projectSlug: 'my-nodejs-app',
  serviceName: 'api-server',
  bufferSize: 100,
  flushInterval: 5000, // 5 seconds
  retryAttempts: 3
});

// Use in your app
devsmithLogger.info('Application started');
devsmithLogger.error('Database connection failed', { 
  error: err.message,
  stack: err.stack 
});
```

#### Step 3: Verify Logs Appear
1. Start your application
2. Generate some log events (requests, errors, etc.)
3. Navigate to DevSmith Logs page (http://localhost:3000/logs)
4. Filter by your project slug
5. **Expected**: Logs appear within 5 seconds
6. **Check**:
   - Timestamp is correct
   - Log level is correct (DEBUG/INFO/WARN/ERROR)
   - Message is complete
   - Context fields are present
   - Tags are included

---

### Python/Flask Integration

#### Step 1: Copy Sample Logger and Extension
```bash
# Download files
curl -o logger.py https://raw.githubusercontent.com/YOUR_REPO/docs/integrations/python/logger.py
curl -o flask_extension.py https://raw.githubusercontent.com/YOUR_REPO/docs/integrations/python/flask_extension.py
```

#### Step 2: Configure Flask App
```python
# app.py
from flask import Flask
from logger import DevSmithLogger
from flask_extension import DevSmithLogging

app = Flask(__name__)

# Initialize logger
logger = DevSmithLogger(
    api_url='http://localhost:8082/api/logs/batch',
    api_key=os.environ['DEVSMITH_API_KEY'],
    project_slug='my-flask-app',
    service_name='web-server',
    buffer_size=100,
    flush_interval=5.0
)

# Initialize Flask extension
logging_ext = DevSmithLogging(app, logger, skip_paths=['/health', '/metrics'])

@app.route('/test')
@logging_ext.log_route(context={'endpoint': 'test'})
def test():
    logger.info('Test endpoint called')
    return {'status': 'ok'}

if __name__ == '__main__':
    app.run()
```

#### Step 3: Verify Middleware Logging
1. Start Flask app
2. Make requests: `curl http://localhost:5000/test`
3. Check DevSmith Logs page
4. **Expected**: See automatic request/response logs:
   - Incoming request (method, path, headers)
   - Response (status code, duration)
   - Custom context from decorator

---

### Go/Gin Integration

#### Step 1: Copy Sample Files
```bash
# Download files
curl -o logger.go https://raw.githubusercontent.com/YOUR_REPO/docs/integrations/go/logger.go
curl -o gin_middleware.go https://raw.githubusercontent.com/YOUR_REPO/docs/integrations/go/gin_middleware.go
```

#### Step 2: Configure Gin App
```go
// main.go
package main

import (
    "github.com/gin-gonic/gin"
    "os"
    "your-module/logger"
    "your-module/gin_middleware"
)

func main() {
    // Initialize logger
    devsmithLogger := logger.New(logger.Config{
        APIURL:       "http://localhost:8082/api/logs/batch",
        APIKey:       os.Getenv("DEVSMITH_API_KEY"),
        ProjectSlug:  "my-go-app",
        ServiceName:  "api-server",
        BufferSize:   100,
        FlushInterval: 5 * time.Second,
    })
    defer devsmithLogger.Close()

    // Setup Gin with middleware
    router := gin.Default()
    router.Use(gin_middleware.DevSmithMiddleware(devsmithLogger, gin_middleware.MiddlewareOptions{
        SkipPaths: []string{"/health", "/metrics"},
        Tags:      []string{"production", "api"},
    }))

    router.GET("/test", func(c *gin.Context) {
        devsmithLogger.Info("Test endpoint", map[string]interface{}{
            "user_id": c.Query("user_id"),
        }, []string{"endpoint"})
        c.JSON(200, gin.H{"status": "ok"})
    })

    router.Run(":8080")
}
```

#### Step 3: Verify Panic Recovery
1. Add panic endpoint: `router.GET("/panic", func(c *gin.Context) { panic("test panic") })`
2. Call endpoint: `curl http://localhost:8080/panic`
3. Check DevSmith Logs
4. **Expected**: See ERROR log with panic message and stack trace

---

## Validation Checklist

### ✅ Basic Functionality
- [ ] Logs appear in DevSmith dashboard within 5 seconds
- [ ] All log levels work (DEBUG, INFO, WARN, ERROR)
- [ ] Timestamps are accurate
- [ ] Messages are complete (no truncation)
- [ ] Service name appears correctly

### ✅ Filtering
- [ ] Can filter by project slug
- [ ] Can filter by service name
- [ ] Can filter by log level
- [ ] Can search by message content
- [ ] Date range filtering works

### ✅ Context and Tags
- [ ] Context fields appear in log details
- [ ] Nested context objects preserved
- [ ] Tags are searchable
- [ ] Custom tags appear correctly

### ✅ Performance
- [ ] Application performance not noticeably impacted
- [ ] No blocking on log calls (async/background)
- [ ] Batch sending reduces network overhead
- [ ] Buffer flushes regularly (every 5 seconds)

### ✅ Error Handling
- [ ] Invalid API key shows clear error message
- [ ] Network failures don't crash application
- [ ] Retry logic works (logs eventually sent)
- [ ] Shutdown flushes remaining logs

### ✅ Long-Running Stability
- [ ] Application runs for 24+ hours without issues
- [ ] Memory usage stable (no leak)
- [ ] No accumulation of unsent logs
- [ ] Periodic flush works reliably

---

## Performance Testing

### Load Test Script
```bash
#!/bin/bash
# Generate 10,000 logs over 5 minutes

for i in {1..10000}; do
  curl -X POST http://localhost:8082/api/logs/batch \
    -H "Authorization: Bearer $DEVSMITH_API_KEY" \
    -H "Content-Type: application/json" \
    -d "{
      \"project_slug\": \"load-test\",
      \"logs\": [{
        \"timestamp\": \"$(date -Iseconds)\",
        \"level\": \"INFO\",
        \"message\": \"Load test log $i\",
        \"service\": \"load-test\"
      }]
    }" &
  
  if (( i % 100 == 0 )); then
    wait  # Wait for batch to complete
    echo "Sent $i logs"
  fi
done

wait
echo "Load test complete: 10,000 logs sent"
```

### Expected Results
- **Throughput**: 14,000-33,000 logs/second (see ARCHITECTURE.md)
- **Latency**: <100ms per batch request
- **Success Rate**: >99.9% (some failures acceptable with retry)
- **Dashboard**: Logs appear in real-time during load test

---

## Troubleshooting

### Issue: Logs Not Appearing

**Symptoms**: Application logs but nothing in DevSmith dashboard

**Checks**:
1. Verify API key is correct: `echo $DEVSMITH_API_KEY`
2. Test API endpoint:
   ```bash
   curl -X POST http://localhost:8082/api/logs/batch \
     -H "Authorization: Bearer $DEVSMITH_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{
       "project_slug": "test",
       "logs": [{
         "timestamp": "'$(date -Iseconds)'",
         "level": "INFO",
         "message": "Test log",
         "service": "test"
       }]
     }'
   ```
3. Check application logs for errors:
   - JavaScript: `console.error` output
   - Python: Check stderr
   - Go: Check logger output
4. Verify project slug matches database: Check Projects page

**Solutions**:
- Regenerate API key if invalid
- Check firewall/network connectivity
- Verify DevSmith platform is running
- Enable debug logging in sample logger

---

### Issue: High Memory Usage

**Symptoms**: Application memory grows over time

**Checks**:
1. Check buffer size: Should be 100-1000, not 10,000+
2. Verify flush interval: Should be 5-10 seconds, not 60+
3. Check for stuck logs: Are batches being sent?
4. Monitor DevSmith API response times

**Solutions**:
- Reduce buffer size: `bufferSize: 100`
- Decrease flush interval: `flushInterval: 5000` (5 seconds)
- Add retry backoff to prevent accumulation
- Check DevSmith API health: `curl http://localhost:8082/health`

---

### Issue: Duplicate Logs

**Symptoms**: Same log appears multiple times in dashboard

**Causes**:
- Retry logic with no deduplication
- Multiple logger instances created
- Application restarted mid-flush

**Solutions**:
- Ensure single logger instance (singleton pattern)
- Add idempotency keys to logs (optional)
- Check application startup logs for multiple initializations

---

### Issue: Missing Context Fields

**Symptoms**: Context appears empty or incomplete

**Checks**:
1. Verify context serialization:
   ```javascript
   // JavaScript
   JSON.stringify(context) // Should work
   
   # Python
   json.dumps(context)  # Should work
   
   // Go
   json.Marshal(context)  // Should work
   ```
2. Check for circular references (will break JSON)
3. Verify nested objects supported

**Solutions**:
- Use simple data types (string, number, boolean)
- Avoid circular references
- Flatten deep nesting if necessary

---

## Beta User Feedback Form

### Integration Experience
- **Difficulty Level** (1-5): _____
- **Time to Integrate** (minutes): _____
- **Documentation Clarity** (1-5): _____
- **Issues Encountered**: _____________________

### Performance
- **Application Impact** (1-5, 5 = no impact): _____
- **Dashboard Load Time** (seconds): _____
- **Log Search Speed** (1-5, 5 = instant): _____
- **Batch Performance** (observed logs/second): _____

### Feature Requests
- Most wanted features: _____________________
- Missing functionality: _____________________
- UI improvements: _____________________

### Bugs Found
- Description: _____________________
- Steps to reproduce: _____________________
- Expected behavior: _____________________
- Actual behavior: _____________________

### Overall Rating (1-5): _____

**Submit feedback to**: [GitHub Issues](https://github.com/YOUR_REPO/issues) or email support@devsmith.dev

---

## Next Steps After Beta

1. **Review Feedback**: Development team reviews all beta user feedback
2. **Bug Fixes**: Critical bugs fixed in next release
3. **Performance Optimization**: Based on real-world usage patterns
4. **Documentation Updates**: Based on common questions/issues
5. **Production Release**: Stable version for general availability

Thank you for participating in the DevSmith beta program! 🎉
