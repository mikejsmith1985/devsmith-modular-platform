# DevSmith Platform - 5 Minute Quick Start

Get the DevSmith logging platform running in 5 minutes or less!

## Prerequisites

- Docker and Docker Compose installed
- That's it! No Go, no PostgreSQL installation needed.

## Quick Start (Production)

### Step 1: Download Configuration (30 seconds)

```bash
# Download the production docker-compose file
curl -O https://raw.githubusercontent.com/mikejsmith1985/devsmith-modular-platform/main/docker-compose.prod.yml

# Download the example environment file
curl -O https://raw.githubusercontent.com/mikejsmith1985/devsmith-modular-platform/main/.env.example
mv .env.example .env
```

### Step 2: Configure Environment (1 minute)

Edit `.env` and set at minimum:

```bash
# Change the default password!
DB_PASSWORD=your-secure-password-here

# For GitHub OAuth (optional for logs-only deployment)
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
JWT_SECRET=your-random-secret-key
```

### Step 3: Start Everything (2 minutes)

```bash
# Pull images and start all services
docker-compose -f docker-compose.prod.yml up -d

# Wait for services to be healthy (auto-runs migrations)
docker-compose -f docker-compose.prod.yml ps
```

### Step 4: Create Your First Project (30 seconds)

```bash
# Create a project and get your API token
curl -X POST http://localhost:8082/api/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Application",
    "slug": "my-app"
  }'

# Response includes your API token:
# {
#   "id": 1,
#   "name": "My Application",
#   "slug": "my-app",
#   "api_token": "generated-token-here",
#   "is_active": true
# }
```

### Step 5: Start Logging! (30 seconds)

```bash
# Test your API token
curl -X POST http://localhost:8082/api/logs/batch \
  -H "X-API-Key: YOUR-API-TOKEN-HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "project_slug": "my-app",
    "logs": [
      {
        "timestamp": "2025-11-12T10:00:00Z",
        "level": "info",
        "message": "Application started successfully!",
        "service_name": "my-service",
        "context": {"version": "1.0.0"}
      }
    ]
  }'
```

**Done!** Your DevSmith platform is running. 🎉

## Services Running

- **Logs API**: http://localhost:8082
- **Portal**: http://localhost:8080 (requires GitHub OAuth setup)
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

## Health Checks

```bash
# Check logs service
curl http://localhost:8082/health

# Check portal service
curl http://localhost:8080/health

# View all containers
docker-compose -f docker-compose.prod.yml ps
```

## View Logs

```bash
# Logs service logs
docker-compose -f docker-compose.prod.yml logs -f logs

# All services
docker-compose -f docker-compose.prod.yml logs -f

# Database logs
docker-compose -f docker-compose.prod.yml logs -f postgres
```

## Stop Everything

```bash
# Stop services (keeps data)
docker-compose -f docker-compose.prod.yml stop

# Remove everything (DELETES DATA!)
docker-compose -f docker-compose.prod.yml down -v
```

## Integration Examples

### GitHub Actions

```yaml
# .github/workflows/build.yml
steps:
  - name: Send logs to DevSmith
    run: |
      curl -X POST https://your-devsmith-instance.com/api/logs/batch \
        -H "X-API-Key: ${{ secrets.DEVSMITH_API_TOKEN }}" \
        -H "Content-Type: application/json" \
        -d '{
          "project_slug": "my-app",
          "logs": [{
            "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
            "level": "info",
            "message": "Build completed",
            "service_name": "github-actions",
            "context": {
              "workflow": "${{ github.workflow }}",
              "run_id": "${{ github.run_id }}"
            }
          }]
        }'
```

### Node.js

```javascript
const axios = require('axios');

async function sendLog(level, message, context = {}) {
  await axios.post('http://localhost:8082/api/logs/batch', {
    project_slug: 'my-app',
    logs: [{
      timestamp: new Date().toISOString(),
      level: level,
      message: message,
      service_name: 'my-nodejs-app',
      context: context
    }]
  }, {
    headers: {
      'X-API-Key': process.env.DEVSMITH_API_TOKEN
    }
  });
}

// Usage
sendLog('info', 'User logged in', { userId: 123 });
```

### Python

```python
import requests
from datetime import datetime

def send_log(level, message, context=None):
    requests.post(
        'http://localhost:8082/api/logs/batch',
        headers={'X-API-Key': os.getenv('DEVSMITH_API_TOKEN')},
        json={
            'project_slug': 'my-app',
            'logs': [{
                'timestamp': datetime.utcnow().isoformat() + 'Z',
                'level': level,
                'message': message,
                'service_name': 'my-python-app',
                'context': context or {}
            }]
        }
    )

# Usage
send_log('info', 'User logged in', {'user_id': 123})
```

## Troubleshooting

### Service won't start
```bash
# Check logs
docker-compose -f docker-compose.prod.yml logs logs

# Restart specific service
docker-compose -f docker-compose.prod.yml restart logs
```

### Database connection issues
```bash
# Check postgres is healthy
docker-compose -f docker-compose.prod.yml ps postgres

# Check database exists
docker-compose -f docker-compose.prod.yml exec postgres \
  psql -U devsmith -d devsmith -c '\l'
```

### Can't create projects
```bash
# Check migrations ran
docker-compose -f docker-compose.prod.yml exec postgres \
  psql -U devsmith -d devsmith -c '\dt logs.*'

# Should show: logs.projects, logs.entries
```

## Production Deployment

For production, you should:

1. **Use proper secrets management** (not `.env` files)
2. **Set up SSL/TLS** (use nginx or Traefik reverse proxy)
3. **Configure backups** for PostgreSQL
4. **Set resource limits** in docker-compose.yml
5. **Monitor services** with Prometheus/Grafana
6. **Use external databases** for better reliability

See [PRODUCTION_DEPLOYMENT.md](./PRODUCTION_DEPLOYMENT.md) for details.

## Support

- **Issues**: https://github.com/mikejsmith1985/devsmith-modular-platform/issues
- **Docs**: https://github.com/mikejsmith1985/devsmith-modular-platform/wiki

## Next Steps

- Set up GitHub OAuth for Portal access
- Configure custom domains
- Add more services (Review, Analytics)
- Set up monitoring and alerts
