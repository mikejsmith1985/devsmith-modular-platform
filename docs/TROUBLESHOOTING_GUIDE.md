# Troubleshooting Guide: Cross-Repository Logging

This guide provides solutions to common issues when integrating with DevSmith Logs platform.

## 📋 Table of Contents

1. [Authentication Issues](#authentication-issues)
2. [Logs Not Appearing](#logs-not-appearing)
3. [Rate Limiting Issues](#rate-limiting-issues)
4. [Network and Connectivity](#network-and-connectivity)
5. [Batch Size and Performance](#batch-size-and-performance)
6. [JSON and Data Format](#json-and-data-format)
7. [Dashboard and UI Issues](#dashboard-and-ui-issues)
8. [Debugging Techniques](#debugging-techniques)

---

## Authentication Issues

### 🔴 Problem: "401 Unauthorized - Invalid API key"

**Symptoms**:
```json
{
  "error": "Invalid API key"
}
```

**Common Causes**:

1. **Using the wrong API key**
   - ✅ **Solution**: Copy the API key from the project creation modal (shown only once)
   - If you lost the key, regenerate it in **Projects** page → **Regenerate Key** button

2. **Missing "Bearer" prefix**
   ```javascript
   // ❌ Wrong
   headers: { 'Authorization': 'proj_abc123...' }
   
   // ✅ Correct
   headers: { 'Authorization': 'Bearer proj_abc123...' }
   ```

3. **Using the hashed key from database**
   - The API key shown in the database is bcrypt-hashed (`$2a$10$...`)
   - You need the **original key** shown once during creation
   - ✅ **Solution**: Regenerate the API key if you lost it

4. **Trailing spaces or newlines**
   ```javascript
   // ❌ Wrong (extra space)
   const API_KEY = 'proj_abc123... ';
   
   // ✅ Correct (trim whitespace)
   const API_KEY = process.env.LOGS_API_KEY.trim();
   ```

**Quick Test**:
```bash
# Test your API key
curl -X POST http://localhost:8082/api/logs/batch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{"entries":[{"level":"INFO","message":"Test","service":"test"}]}'

# Expected: {"inserted": 1}
# If 401: API key is invalid
```

---

### 🔴 Problem: "401 Unauthorized - Authorization header required"

**Symptoms**:
```json
{
  "error": "Authorization header required"
}
```

**Cause**: Missing `Authorization` header in HTTP request

**Solutions**:

1. **Check header is set**:
   ```javascript
   // ❌ Wrong (no Authorization header)
   fetch(url, {
     method: 'POST',
     headers: { 'Content-Type': 'application/json' }
   });
   
   // ✅ Correct
   fetch(url, {
     method: 'POST',
     headers: {
       'Content-Type': 'application/json',
       'Authorization': `Bearer ${LOGS_API_KEY}`
     }
   });
   ```

2. **Environment variable not loaded**:
   ```javascript
   // Debug: Check if env var exists
   console.log('API Key:', process.env.LOGS_API_KEY ? 'SET' : 'MISSING');
   
   // Load environment variables (Node.js)
   require('dotenv').config();
   ```

---

### 🔴 Problem: "403 Forbidden - Project is deactivated"

**Symptoms**:
```json
{
  "error": "Project is deactivated"
}
```

**Cause**: Project was deactivated in the Projects page

**Solution**:
1. Go to http://localhost:3000/projects
2. Check if project status shows "Deactivated"
3. If deactivated by mistake, you need to:
   - Create a new project (deactivated projects cannot be reactivated)
   - Update your application with the new API key

**Prevention**: Don't deactivate projects that are still in use!

---

## Logs Not Appearing

### 🟡 Problem: Logs not showing up in Health dashboard

**Diagnostic Checklist**:

1. **Verify API is reachable**:
   ```bash
   curl http://localhost:8082/api/health
   # Expected: {"status":"ok"}
   ```

2. **Check project_id is correct**:
   - API key is tied to a specific project
   - Logs automatically tagged with `project_id` based on API key
   - In dashboard, select the correct project in filter dropdown

3. **Check service name**:
   ```bash
   # Check if logs exist with your service name
   curl -X GET http://localhost:8082/api/logs?service=YOUR_SERVICE_NAME
   ```

4. **Verify logs are being sent**:
   ```javascript
   // Add debug logging
   console.log('Sending logs to:', LOGS_API_URL);
   console.log('Batch size:', batch.entries.length);
   
   const response = await fetch(url, options);
   console.log('Response status:', response.status);
   const body = await response.json();
   console.log('Response body:', body);
   ```

5. **Check buffer is flushing**:
   - Logs are buffered and sent in batches
   - If batch size not reached, logs wait for flush interval (default 5s)
   - Force flush on application shutdown:
     ```javascript
     // Node.js
     process.on('SIGINT', async () => {
       await logger.flush();
       process.exit(0);
     });
     
     // Python
     import atexit
     atexit.register(logger.flush)
     ```

6. **Check database**:
   ```bash
   # Connect to PostgreSQL
   docker exec -it devsmith-modular-platform-postgres-1 psql -U devsmith -d devsmith
   
   # Query logs
   SELECT COUNT(*) FROM logs.entries WHERE service = 'YOUR_SERVICE_NAME';
   SELECT * FROM logs.entries ORDER BY created_at DESC LIMIT 10;
   ```

---

## Rate Limiting Issues

### 🟠 Problem: "429 Too Many Requests"

**Symptoms**:
```json
{
  "error": "Rate limit exceeded"
}
```

**Cause**: Exceeded 1,000 requests/minute per API key

**Solutions**:

1. **Increase batch size**:
   ```javascript
   // ❌ Bad: Sending 1 log per request = 60 logs = 60 requests/min
   logs.forEach(log => sendLog(log));
   
   // ✅ Good: Sending 100 logs per request = 6000 logs = 60 requests/min
   sendBatch(logs); // Max 1000 logs per batch
   ```

2. **Implement exponential backoff**:
   ```javascript
   async function sendLogsWithRetry(batch, maxRetries = 3) {
     for (let attempt = 0; attempt < maxRetries; attempt++) {
       const response = await fetch(url, options);
       
       if (response.status === 429) {
         // Rate limited - wait before retry
         const delay = Math.pow(2, attempt) * 1000; // 1s, 2s, 4s
         console.log(`Rate limited, retrying in ${delay}ms...`);
         await new Promise(resolve => setTimeout(resolve, delay));
         continue;
       }
       
       if (response.ok) return response;
       throw new Error(`HTTP ${response.status}`);
     }
     
     throw new Error('Max retries exceeded');
   }
   ```

3. **Reduce flush frequency**:
   ```javascript
   // If you're hitting rate limits, increase flush interval
   const FLUSH_INTERVAL = 10000; // 10 seconds instead of 5
   const BATCH_SIZE = 500; // Increase batch size
   ```

4. **Use multiple API keys**:
   - Create separate projects for different services
   - Each API key gets its own 1,000 req/min limit
   - Distribute load across multiple keys

**Calculate your rate**:
```
Requests per minute = (Logs per minute) / (Logs per batch)

Example:
- 10,000 logs/minute
- 100 logs per batch
= 100 requests/minute (under limit ✅)

Bad example:
- 10,000 logs/minute
- 10 logs per batch
= 1,000 requests/minute (at limit ⚠️)
```

---

## Network and Connectivity

### 🔵 Problem: Network timeouts or connection refused

**Symptoms**:
- `ECONNREFUSED`
- `ETIMEDOUT`
- `Network request failed`

**Solutions**:

1. **Verify LOGS_API_URL is correct**:
   ```javascript
   console.log('API URL:', process.env.LOGS_API_URL);
   // Expected: http://localhost:8082 (development)
   //       or: https://logs.yourdomain.com (production)
   ```

2. **Check service is running**:
   ```bash
   # Check if logs service is running
   docker ps | grep logs
   
   # Check logs service health
   curl http://localhost:8082/api/health
   ```

3. **Check firewall rules**:
   ```bash
   # Test connectivity from your app server
   telnet localhost 8082
   
   # If blocked, check firewall
   sudo ufw status
   sudo iptables -L
   ```

4. **Use correct URL in Docker**:
   ```javascript
   // ❌ Wrong: localhost from inside container = container itself
   LOGS_API_URL=http://localhost:8082
   
   // ✅ Correct: use service name or host.docker.internal
   LOGS_API_URL=http://logs:8082  // If in same Docker network
   LOGS_API_URL=http://host.docker.internal:8082  // If accessing host
   ```

5. **Increase timeout**:
   ```javascript
   // If network is slow, increase timeout
   fetch(url, {
     ...options,
     signal: AbortSignal.timeout(10000) // 10 seconds
   });
   ```

---

## Batch Size and Performance

### 🟢 Problem: "400 Bad Request - Batch exceeds maximum size"

**Symptoms**:
```json
{
  "error": "Batch exceeds maximum size of 1000 logs"
}
```

**Cause**: Sent more than 1,000 logs in a single batch

**Solution**:
```javascript
// Split large batches into chunks
function splitIntoBatches(logs, maxSize = 1000) {
  const batches = [];
  for (let i = 0; i < logs.length; i += maxSize) {
    batches.push(logs.slice(i, i + maxSize));
  }
  return batches;
}

// Send multiple batches
const batches = splitIntoBatches(allLogs, 1000);
for (const batch of batches) {
  await sendBatch({ entries: batch });
}
```

---

### 🟢 Problem: Slow log ingestion or high latency

**Symptoms**:
- Logs take 5+ seconds to appear in dashboard
- High memory usage in application

**Solutions**:

1. **Optimize batch size**:
   ```javascript
   // Too small = too many requests = slow
   const BATCH_SIZE = 10; // ❌ Bad
   
   // Optimal = balance between latency and throughput
   const BATCH_SIZE = 100; // ✅ Good
   
   // For high-traffic apps
   const BATCH_SIZE = 500; // ✅ Even better
   ```

2. **Reduce flush interval for real-time logs**:
   ```javascript
   // If you need near real-time logging
   const FLUSH_INTERVAL = 1000; // 1 second
   const BATCH_SIZE = 50; // Smaller batches, more frequent
   ```

3. **Use async/non-blocking logging**:
   ```javascript
   // ❌ Bad: Blocks main thread
   function log(message) {
     const response = syncHttpRequest(url, data);
     return response;
   }
   
   // ✅ Good: Non-blocking
   function log(message) {
     // Add to buffer, send async
     buffer.push(entry);
     if (buffer.length >= BATCH_SIZE) {
       sendAsync(buffer); // Fire and forget
     }
   }
   ```

---

## JSON and Data Format

### 🟣 Problem: "400 Bad Request - Invalid JSON"

**Symptoms**:
```json
{
  "error": "Invalid JSON format"
}
```

**Common Causes**:

1. **Missing required fields**:
   ```javascript
   // ❌ Wrong: Missing 'entries'
   { "logs": [...] }
   
   // ✅ Correct
   { "entries": [...] }
   ```

2. **Empty entries array**:
   ```javascript
   // ❌ Wrong: Empty array
   { "entries": [] }
   
   // ✅ Correct: At least one entry
   { "entries": [{ "level": "INFO", "message": "Test", "service": "test" }] }
   ```

3. **Missing required log fields**:
   ```javascript
   // ❌ Wrong: Missing 'service'
   {
     "entries": [
       { "level": "INFO", "message": "Test" }
     ]
   }
   
   // ✅ Correct: All required fields
   {
     "entries": [
       {
         "level": "INFO",
         "message": "Test log message",
         "service": "my-service"
       }
     ]
   }
   ```

4. **Invalid timestamp format**:
   ```javascript
   // ❌ Wrong: Unix timestamp
   "timestamp": 1699704000
   
   // ✅ Correct: ISO 8601 / RFC3339
   "timestamp": "2025-11-10T12:00:00Z"
   "timestamp": new Date().toISOString()
   ```

5. **Malformed JSON**:
   ```javascript
   // ❌ Wrong: Missing comma, quotes, etc.
   { entries: [{ level: INFO }] }
   
   // ✅ Correct: Valid JSON
   { "entries": [{ "level": "INFO" }] }
   
   // Validate before sending
   try {
     JSON.parse(JSON.stringify(batch));
   } catch (e) {
     console.error('Invalid JSON:', e);
   }
   ```

---

### 🟣 Problem: "400 Bad Request - Missing Content-Type header"

**Symptoms**:
```json
{
  "error": "Content-Type must be application/json"
}
```

**Solution**:
```javascript
// Always set Content-Type header
fetch(url, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json', // Required!
    'Authorization': `Bearer ${API_KEY}`
  },
  body: JSON.stringify(batch)
});
```

---

## Dashboard and UI Issues

### 🟤 Problem: Dashboard shows "No logs found"

**Diagnostic Steps**:

1. **Check project filter**:
   - Click **Project** dropdown at top of page
   - Select your project (not "All Projects")
   - If project not in list, verify it was created successfully

2. **Check service filter**:
   - If service filter is set, only logs from that service show
   - Set to "All Services" to see everything

3. **Check date range**:
   - Dashboard shows recent logs (last 24 hours by default)
   - If testing old data, adjust date range

4. **Check log level filter**:
   - If set to "ERROR" only, you won't see INFO/DEBUG logs
   - Set to "All Levels" to see everything

5. **Force refresh**:
   - Click **Refresh** button in dashboard
   - Or reload page (Ctrl+R / Cmd+R)

---

### 🟤 Problem: Logs appear delayed or not in real-time

**Cause**: WebSocket connection not established

**Solutions**:

1. **Check WebSocket connection**:
   ```javascript
   // Open browser console (F12)
   // Look for WebSocket connection
   // Should see: WS ws://localhost:8082/ws/logs
   ```

2. **Check browser console for errors**:
   ```
   WebSocket connection failed
   Error: WebSocket is already in CLOSING or CLOSED state
   ```

3. **Disable browser extensions**:
   - Ad blockers can block WebSocket connections
   - Try in incognito mode

4. **Check firewall/proxy**:
   - Some corporate firewalls block WebSocket
   - Test from different network

---

## Debugging Techniques

### Enable Verbose Logging

**JavaScript**:
```javascript
const DEBUG = process.env.DEBUG === 'true';

function log(level, message, metadata) {
  if (DEBUG) {
    console.log('[LOGS CLIENT]', {
      level,
      message,
      metadata,
      buffer_size: logBuffer.length,
      api_url: LOGS_API_URL,
    });
  }
  // ... rest of logging logic
}
```

**Python**:
```python
import logging

logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)

def log(self, level, message, metadata):
    logger.debug(f'[LOGS CLIENT] {level}: {message}, buffer_size={len(self.buffer)}')
    # ... rest of logging logic
```

**Go**:
```go
const DEBUG = true

func (c *LogsClient) log(level, message string, metadata map[string]interface{}) {
    if DEBUG {
        fmt.Printf("[LOGS CLIENT] level=%s, message=%s, buffer_size=%d\n", 
            level, message, len(c.buffer))
    }
    // ... rest of logging logic
}
```

---

### Test API Directly with curl

**Test Authentication**:
```bash
# Should return 401
curl -i -X POST http://localhost:8082/api/logs/batch \
  -H "Content-Type: application/json" \
  -d '{"entries":[{"level":"INFO","message":"Test","service":"test"}]}'

# Should return 200
curl -i -X POST http://localhost:8082/api/logs/batch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{"entries":[{"level":"INFO","message":"Test","service":"test"}]}'
```

**Test Batch Ingestion**:
```bash
# Send 10 logs
curl -X POST http://localhost:8082/api/logs/batch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "entries": [
      {"level":"DEBUG","message":"Log 1","service":"test"},
      {"level":"INFO","message":"Log 2","service":"test"},
      {"level":"WARNING","message":"Log 3","service":"test"},
      {"level":"ERROR","message":"Log 4","service":"test"},
      {"level":"INFO","message":"Log 5","service":"test"},
      {"level":"INFO","message":"Log 6","service":"test"},
      {"level":"INFO","message":"Log 7","service":"test"},
      {"level":"INFO","message":"Log 8","service":"test"},
      {"level":"INFO","message":"Log 9","service":"test"},
      {"level":"INFO","message":"Log 10","service":"test"}
    ]
  }'

# Expected response: {"inserted": 10}
```

---

### Check Service Logs

**Logs Service**:
```bash
# View logs service logs
docker logs devsmith-modular-platform-logs-1 --tail 100 -f

# Look for errors like:
# - "Invalid API key"
# - "Rate limit exceeded"
# - "Batch exceeds maximum size"
```

**PostgreSQL Logs**:
```bash
# View database logs
docker logs devsmith-modular-platform-postgres-1 --tail 50

# Look for:
# - Connection errors
# - Slow queries
# - Constraint violations
```

---

### Use Network Monitoring

**Browser DevTools**:
1. Open browser console (F12)
2. Go to **Network** tab
3. Filter by "logs" or "batch"
4. Check request/response:
   - Status code (should be 200)
   - Request headers (Authorization present?)
   - Request payload (valid JSON?)
   - Response body (error message?)

**Command Line (tcpdump)**:
```bash
# Monitor traffic to logs API
sudo tcpdump -i any -A 'port 8082'

# Look for POST /api/logs/batch requests
# Check Authorization headers
# Verify JSON payloads
```

---

## Quick Reference: HTTP Status Codes

| Code | Meaning | Common Cause | Solution |
|------|---------|--------------|----------|
| 200 | OK | Success | - |
| 400 | Bad Request | Invalid JSON, missing fields, batch too large | Check JSON format and batch size |
| 401 | Unauthorized | Invalid/missing API key | Check Authorization header |
| 403 | Forbidden | Project deactivated | Reactivate project or create new one |
| 405 | Method Not Allowed | Used GET instead of POST | Use POST method |
| 429 | Too Many Requests | Rate limit exceeded | Increase batch size, add backoff |
| 500 | Internal Server Error | Server issue | Check service logs |

---

## Still Having Issues?

If this guide didn't solve your problem:

1. ✅ Check [INTEGRATION_GUIDE.md](./INTEGRATION_GUIDE.md) for setup instructions
2. ✅ Review **Integration Docs** page in the dashboard
3. ✅ Check DevSmith Logs service logs:
   ```bash
   docker logs devsmith-modular-platform-logs-1 --tail 200
   ```
4. ✅ Check database for logs:
   ```sql
   SELECT * FROM logs.entries WHERE service = 'YOUR_SERVICE' LIMIT 10;
   ```
5. ✅ Create an issue on GitHub with:
   - Error message
   - Steps to reproduce
   - Code snippet (sanitize API keys!)
   - Service logs

---

**Last Updated**: 2025-11-11  
**Version**: 1.0
