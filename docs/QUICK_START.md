# DevSmith Platform - Beta User Quick Start

**Get logging in 5 minutes or less!** 🚀

---

## One-Command Installation

```bash
curl -sSL https://raw.githubusercontent.com/mikejsmith1985/devsmith-modular-platform/main/scripts/quick-deploy.sh | bash
```

That's it! The script will:
1. ✅ Check prerequisites (Docker, Git, etc.)
2. ✅ Clone the repository
3. ✅ Generate your API token
4. ✅ Start all services
5. ✅ Create your first project
6. ✅ Run health checks

**Total time: ~5 minutes** (depending on internet speed)

---

## What You Get

After installation, you'll have:

- **Logs Service** running at `http://localhost:8082`
- **Your API Token** saved to `.deploy/api-credentials.txt`
- **Your First Project** ready to receive logs
- **Working Example** command to test immediately

---

## Quick Test

Your credentials file (`.deploy/api-credentials.txt`) will contain a ready-to-use curl command:

```bash
curl -X POST http://localhost:8082/api/logs/batch \
  -H "X-API-Key: YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "project_slug": "my-first-project",
    "logs": [
      {
        "timestamp": "2025-11-12T10:00:00Z",
        "level": "info",
        "message": "Hello from DevSmith!",
        "service_name": "my-app",
        "context": {"version": "1.0.0"}
      }
    ]
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "logs_received": 1,
  "project_id": 1
}
```

---

## Integration Examples

### Node.js / JavaScript

```javascript
// Install axios: npm install axios
const axios = require('axios');

const API_KEY = 'YOUR_API_TOKEN_HERE';
const LOGS_URL = 'http://localhost:8082/api/logs/batch';

async function sendLogs(logs) {
  try {
    const response = await axios.post(LOGS_URL, {
      project_slug: 'my-first-project',
      logs: logs
    }, {
      headers: {
        'X-API-Key': API_KEY,
        'Content-Type': 'application/json'
      }
    });
    
    console.log('Logs sent:', response.data);
  } catch (error) {
    console.error('Failed to send logs:', error.message);
  }
}

// Example usage
sendLogs([
  {
    timestamp: new Date().toISOString(),
    level: 'info',
    message: 'Application started',
    service_name: 'my-app',
    context: {
      version: '1.0.0',
      environment: 'production'
    }
  }
]);
```

### Python

```python
# Install requests: pip install requests
import requests
from datetime import datetime

API_KEY = 'YOUR_API_TOKEN_HERE'
LOGS_URL = 'http://localhost:8082/api/logs/batch'

def send_logs(logs):
    headers = {
        'X-API-Key': API_KEY,
        'Content-Type': 'application/json'
    }
    
    payload = {
        'project_slug': 'my-first-project',
        'logs': logs
    }
    
    try:
        response = requests.post(LOGS_URL, json=payload, headers=headers)
        response.raise_for_status()
        print(f"Logs sent: {response.json()}")
    except requests.exceptions.RequestException as e:
        print(f"Failed to send logs: {e}")

# Example usage
send_logs([
    {
        'timestamp': datetime.utcnow().isoformat() + 'Z',
        'level': 'info',
        'message': 'Application started',
        'service_name': 'my-app',
        'context': {
            'version': '1.0.0',
            'environment': 'production'
        }
    }
])
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

const (
    APIKey  = "YOUR_API_TOKEN_HERE"
    LogsURL = "http://localhost:8082/api/logs/batch"
)

type LogEntry struct {
    Timestamp   string                 `json:"timestamp"`
    Level       string                 `json:"level"`
    Message     string                 `json:"message"`
    ServiceName string                 `json:"service_name"`
    Context     map[string]interface{} `json:"context,omitempty"`
}

type BatchRequest struct {
    ProjectSlug string     `json:"project_slug"`
    Logs        []LogEntry `json:"logs"`
}

func sendLogs(logs []LogEntry) error {
    payload := BatchRequest{
        ProjectSlug: "my-first-project",
        Logs:        logs,
    }
    
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal logs: %w", err)
    }
    
    req, err := http.NewRequest("POST", LogsURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    
    req.Header.Set("X-API-Key", APIKey)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send logs: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    
    fmt.Println("Logs sent successfully!")
    return nil
}

func main() {
    logs := []LogEntry{
        {
            Timestamp:   time.Now().UTC().Format(time.RFC3339),
            Level:       "info",
            Message:     "Application started",
            ServiceName: "my-app",
            Context: map[string]interface{}{
                "version":     "1.0.0",
                "environment": "production",
            },
        },
    }
    
    if err := sendLogs(logs); err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Bash / Shell Scripts

```bash
#!/bin/bash

API_KEY="YOUR_API_TOKEN_HERE"
LOGS_URL="http://localhost:8082/api/logs/batch"

# Function to send logs
send_log() {
    local level=$1
    local message=$2
    local service=${3:-"bash-script"}
    
    curl -X POST "$LOGS_URL" \
      -H "X-API-Key: $API_KEY" \
      -H "Content-Type: application/json" \
      -d "{
        \"project_slug\": \"my-first-project\",
        \"logs\": [
          {
            \"timestamp\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\",
            \"level\": \"$level\",
            \"message\": \"$message\",
            \"service_name\": \"$service\",
            \"context\": {\"script\": \"$0\", \"user\": \"$USER\"}
          }
        ]
      }"
}

# Example usage
send_log "info" "Script started" "my-deployment-script"
send_log "error" "Deployment failed" "my-deployment-script"
```

---

## GitHub Actions Integration

```yaml
name: Send Deployment Logs to DevSmith

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Deploy Application
        run: |
          # Your deployment commands here
          echo "Deploying..."
      
      - name: Send Success Log to DevSmith
        if: success()
        run: |
          curl -X POST http://your-devsmith-server:8082/api/logs/batch \
            -H "X-API-Key: ${{ secrets.DEVSMITH_API_KEY }}" \
            -H "Content-Type: application/json" \
            -d "{
              \"project_slug\": \"my-project\",
              \"logs\": [{
                \"timestamp\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\",
                \"level\": \"info\",
                \"message\": \"Deployment succeeded for commit ${{ github.sha }}\",
                \"service_name\": \"github-actions\",
                \"context\": {
                  \"commit\": \"${{ github.sha }}\",
                  \"branch\": \"${{ github.ref }}\",
                  \"actor\": \"${{ github.actor }}\"
                }
              }]
            }"
      
      - name: Send Failure Log to DevSmith
        if: failure()
        run: |
          curl -X POST http://your-devsmith-server:8082/api/logs/batch \
            -H "X-API-Key: ${{ secrets.DEVSMITH_API_KEY }}" \
            -H "Content-Type: application/json" \
            -d "{
              \"project_slug\": \"my-project\",
              \"logs\": [{
                \"timestamp\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\",
                \"level\": \"error\",
                \"message\": \"Deployment failed for commit ${{ github.sha }}\",
                \"service_name\": \"github-actions\",
                \"context\": {
                  \"commit\": \"${{ github.sha }}\",
                  \"branch\": \"${{ github.ref }}\",
                  \"actor\": \"${{ github.actor }}\"
                }
              }]
            }"
```

---

## Querying Your Logs

### Get Recent Logs

```bash
curl "http://localhost:8082/api/logs?project_slug=my-first-project&limit=10"
```

### Filter by Level

```bash
curl "http://localhost:8082/api/logs?project_slug=my-first-project&level=error"
```

### Search by Service

```bash
curl "http://localhost:8082/api/logs?project_slug=my-first-project&service_name=my-app"
```

### Time Range Query

```bash
curl "http://localhost:8082/api/logs?project_slug=my-first-project&start_time=2025-11-12T00:00:00Z&end_time=2025-11-12T23:59:59Z"
```

---

## Managing Your Platform

### Check Service Status

```bash
docker-compose ps
```

### View Service Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f logs
```

### Restart Services

```bash
docker-compose restart
```

### Stop Platform

```bash
docker-compose down
```

### Start Platform

```bash
docker-compose up -d
```

### Update Platform

```bash
cd ~/devsmith-platform
git pull
docker-compose up -d --build
```

---

## Performance Expectations

Based on load testing with Simple Token Authentication:

- **Average Response Time:** ~14ms
- **Throughput:** 250+ requests/second
- **Failure Rate:** 0%
- **Logs Ingestion:** 25,000 logs/second
- **Concurrent Connections:** 10+

Your mileage may vary based on hardware, but this gives you an idea of what to expect.

---

## Troubleshooting

### Service Won't Start

```bash
# Check Docker daemon
docker ps

# Check service-specific logs
docker-compose logs logs

# Rebuild from scratch
docker-compose down -v
docker-compose up -d --build
```

### Authentication Fails (HTTP 401)

1. Check your API token: `cat .deploy/api-credentials.txt`
2. Verify token in database:
   ```bash
   docker-compose exec postgres psql -U devsmith -d devsmith -c "SELECT name, slug, api_token FROM logs.projects;"
   ```
3. Ensure `X-API-Key` header is set correctly

### Logs Not Appearing

1. Check service health: `curl http://localhost:8082/health`
2. Verify project slug matches: lowercase, hyphens instead of spaces
3. Check database:
   ```bash
   docker-compose exec postgres psql -U devsmith -d devsmith -c "SELECT COUNT(*) FROM logs.entries;"
   ```

### Performance Issues

1. Check system resources: `docker stats`
2. Verify database connection pool: Check logs for connection errors
3. Monitor response times: Add timing to your requests
4. Consider tuning Docker resource limits in `docker-compose.yml`

---

## Creating Additional Projects

```bash
# Connect to database
docker-compose exec postgres psql -U devsmith -d devsmith

# Insert new project
INSERT INTO logs.projects (name, slug, api_token, is_active)
VALUES (
    'My Second Project',
    'my-second-project',
    'your-new-token-here',  -- Generate with: openssl rand -hex 32
    true
);
```

---

## Next Steps

1. ✅ **Integrate with your app** - Use examples above
2. ✅ **Set up monitoring** - Query logs regularly
3. ✅ **Create dashboards** - Visualize your log data
4. ✅ **Secure your deployment** - Use HTTPS in production
5. ✅ **Join community** - Share feedback and contribute!

---

## Support

- **Issues:** [GitHub Issues](https://github.com/mikejsmith1985/devsmith-modular-platform/issues)
- **Documentation:** [Wiki](https://github.com/mikejsmith1985/devsmith-modular-platform/wiki)
- **Email:** support@devsmith.io (beta users only)

---

**Welcome to DevSmith Platform! Happy logging! 📊**
