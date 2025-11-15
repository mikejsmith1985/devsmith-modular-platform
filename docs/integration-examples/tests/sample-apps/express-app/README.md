# DevSmith Express Sample App

Minimal Express.js application demonstrating DevSmith cross-repo logging integration.

## Features

- ✅ DevSmith logger with batch ingestion
- ✅ Express middleware for automatic request/response logging
- ✅ Multiple endpoints with different log levels
- ✅ Error handling and exception logging
- ✅ Health check endpoint (skipped from logging)
- ✅ Graceful shutdown with log flushing

## Prerequisites

- Node.js 16+ installed
- DevSmith account with project created
- API key generated (see [Manual Test Guide](../../MANUAL_TEST_GUIDE.md))

## Setup

### 1. Install Dependencies

```bash
npm install
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` and fill in your DevSmith credentials:

```env
DEVSMITH_API_URL=http://localhost:8082/api/logs/batch
DEVSMITH_API_KEY=devsmith_YOUR_API_KEY_HERE
DEVSMITH_PROJECT_SLUG=your-project-slug
DEVSMITH_SERVICE_NAME=express-sample
```

### 3. Start Application

```bash
# Development (with auto-reload)
npm run dev

# Production
npm start
```

Server runs on http://localhost:3001

## Testing Endpoints

### Root Endpoint
```bash
curl http://localhost:3001/
```
**Expected Logs**: INFO - Root endpoint accessed

### Health Check (Not Logged)
```bash
curl http://localhost:3001/health
```
**Expected Logs**: None (skipped by middleware)

### Get Users
```bash
curl http://localhost:3001/api/users
```
**Expected Logs**: 
- INFO - Incoming request
- DEBUG - Fetching users list
- INFO - Request completed

### Create User (Success)
```bash
curl -X POST http://localhost:3001/api/users \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com"}'
```
**Expected Logs**:
- INFO - Incoming request
- INFO - Creating new user
- INFO - User created successfully
- INFO - Request completed

### Create User (Validation Error)
```bash
curl -X POST http://localhost:3001/api/users \
  -H "Content-Type: application/json" \
  -d '{"email":"bob@example.com"}'
```
**Expected Logs**:
- INFO - Incoming request
- WARN - User creation failed - missing username
- INFO - Request completed (400 status)

### Error Endpoint
```bash
curl http://localhost:3001/api/error
```
**Expected Logs**:
- INFO - Incoming request
- WARN - Error endpoint called
- ERROR - Application error occurred (with stack trace)
- INFO - Request completed (500 status)

### 404 Not Found
```bash
curl http://localhost:3001/nonexistent
```
**Expected Logs**:
- INFO - Incoming request
- WARN - 404 Not Found
- INFO - Request completed (404 status)

## Verify in DevSmith Dashboard

1. Navigate to http://localhost:3000/logs (or production URL)
2. Filter by project: `your-project-slug`
3. Filter by service: `express-sample`
4. You should see all logs with:
   - Correct timestamps
   - Appropriate log levels
   - Context fields (method, path, status_code, etc.)
   - Tags (request, response, error, etc.)

## Log Levels Used

- **DEBUG**: Database queries, detailed operations
- **INFO**: Request/response flow, business logic
- **WARN**: Validation failures, 404 errors, potential issues
- **ERROR**: Exceptions, crashes, critical failures

## Troubleshooting

### Logs Not Appearing

**Check 1**: Verify API key
```bash
echo $DEVSMITH_API_KEY
```

**Check 2**: Test API endpoint directly
```bash
curl -X POST http://localhost:8082/api/logs/batch \
  -H "Authorization: Bearer $DEVSMITH_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "project_slug": "your-project-slug",
    "logs": [{
      "timestamp": "'$(date -Iseconds)'",
      "level": "INFO",
      "message": "Test log",
      "service": "test"
    }]
  }'
```

**Check 3**: Application logs
```bash
# Check for errors in console output
npm start
```

### High Memory Usage

Reduce buffer size in `app.js`:
```javascript
bufferSize: 50,  // Was 100
flushInterval: 3000  // Was 5000 (flush more frequently)
```

## Next Steps

- Customize log contexts with your business logic
- Add custom tags for filtering
- Integrate with your existing error tracking
- Deploy to production with environment-specific configuration

## Reference

- [Manual Test Guide](../../MANUAL_TEST_GUIDE.md)
- [DevSmith Architecture](../../../../../CROSS_REPO_LOGGING_ARCHITECTURE.md)
- [JavaScript Logger Documentation](../../../javascript/README.md)
