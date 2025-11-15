# DevSmith Gin Sample App

Minimal Gin application demonstrating DevSmith cross-repo logging integration.

## Features

- ✅ DevSmith logger with batch ingestion
- ✅ Gin middleware for automatic request/response logging
- ✅ Panic recovery with error logging
- ✅ Multiple endpoints with different log levels
- ✅ Error handling and exception logging
- ✅ Health/metrics endpoints (skipped from logging)
- ✅ Graceful shutdown with log flushing

## Prerequisites

- Go 1.21+ installed
- DevSmith account with project created
- API key generated (see [Manual Test Guide](../../MANUAL_TEST_GUIDE.md))

## Setup

### 1. Download Dependencies

```bash
go mod download
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
DEVSMITH_SERVICE_NAME=gin-sample
```

### 3. Start Application

```bash
go run main.go
```

Server runs on http://localhost:8080

## Testing Endpoints

### Root Endpoint
```bash
curl http://localhost:8080/
```
**Expected Logs**: INFO - Root endpoint accessed

### Health Check (Not Logged)
```bash
curl http://localhost:8080/health
```
**Expected Logs**: None (skipped by middleware)

### Metrics (Not Logged)
```bash
curl http://localhost:8080/metrics
```
**Expected Logs**: None (skipped by middleware)

### Get Users
```bash
curl http://localhost:8080/api/users
```
**Expected Logs**: 
- INFO - Incoming request
- DEBUG - Fetching users list
- INFO - Request completed

### Create User (Success)
```bash
curl -X POST http://localhost:8080/api/users \
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
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email":"bob@example.com"}'
```
**Expected Logs**:
- INFO - Incoming request
- WARN - User creation failed - missing username
- INFO - Request completed (400 status)

### Error Endpoint
```bash
curl http://localhost:8080/api/error
```
**Expected Logs**:
- INFO - Incoming request
- WARN - Error endpoint called
- ERROR - Application error occurred
- INFO - Request completed (500 status)

### Panic Endpoint (Recovery Test)
```bash
curl http://localhost:8080/api/panic
```
**Expected Logs**:
- INFO - Incoming request
- WARN - Panic endpoint called
- ERROR - Panic recovered (with error details)
- INFO - Request completed (500 status)

### 404 Not Found
```bash
curl http://localhost:8080/nonexistent
```
**Expected Logs**:
- INFO - Incoming request
- WARN - 404 Not Found
- INFO - Request completed (404 status)

## Verify in DevSmith Dashboard

1. Navigate to http://localhost:3000/logs (or production URL)
2. Filter by project: `your-project-slug`
3. Filter by service: `gin-sample`
4. You should see all logs with:
   - Correct timestamps
   - Appropriate log levels
   - Context fields (method, path, status_code, etc.)
   - Tags (request, response, error, panic, etc.)

## Log Levels Used

- **DEBUG**: Detailed operations, database queries
- **INFO**: Request/response flow, business logic
- **WARN**: Validation failures, 404 errors, potential issues
- **ERROR**: Exceptions, crashes, panic recovery, critical failures

## Graceful Shutdown

The app handles SIGINT (Ctrl+C) and SIGTERM gracefully:

```bash
# Start app
go run main.go

# Stop with Ctrl+C
# Expected: Logs flushed before exit
```

## Building for Production

```bash
# Build binary
go build -o gin-sample main.go

# Run binary
./gin-sample
```

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
      "timestamp": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'",
      "level": "INFO",
      "message": "Test log",
      "service": "test"
    }]
  }'
```

**Check 3**: Application logs
```bash
# Check stdout for errors
go run main.go 2>&1 | grep -i error
```

### High Memory Usage

Reduce buffer size in `main.go`:
```go
logger := NewLogger(
    // ...
    50,          // Was 100 (bufferSize)
    3*time.Second,  // Was 5*time.Second (flush more frequently)
)
```

### Panic Not Recovered

Ensure middleware is registered:
```go
router.Use(gin.Recovery())  // Built-in recovery
router.Use(DevSmithMiddleware(...))  // Logs panics
```

## Next Steps

- Customize log contexts with your business logic
- Add custom tags for filtering
- Integrate with your existing error tracking
- Deploy to production with environment-specific configuration

## Reference

- [Manual Test Guide](../../MANUAL_TEST_GUIDE.md)
- [DevSmith Architecture](../../../../../CROSS_REPO_LOGGING_ARCHITECTURE.md)
- [Go Logger Documentation](../../../go/README.md)
