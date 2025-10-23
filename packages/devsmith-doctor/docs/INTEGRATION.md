# Docker Compose Integration for DevSmith Doctor

## Add to Your Existing docker-compose.yml

Add this service definition to your `docker-compose.yml`:

```yaml
services:
  # ... your existing services (postgres, portal, logs, etc.) ...

  # DevSmith Doctor service
  doctor:
    build: ./packages/devsmith-doctor/backend
    container_name: devsmith-doctor
    ports:
      - "8084:8000"
    volumes:
      # Read-only access to validation results
      - ./.validation:/app/.validation:ro
      # Read-only access to configs for analysis
      - ./docker-compose.yml:/app/docker-compose.yml:ro
      - ./nginx.conf:/app/nginx.conf:ro
      # Docker socket for executing fixes
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - PROJECT_ROOT=/app
      - LOG_SERVICE_URL=http://logs:8000
      - POSTGRES_URL=postgresql://user:password@postgres:5432/devsmith
    depends_on:
      logs:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 10s
    networks:
      - devsmith-network
    restart: unless-stopped

networks:
  devsmith-network:
    driver: bridge
```

## Update nginx.conf

Add this location block to your `nginx.conf` to route Doctor requests:

```nginx
# DevSmith Doctor API
location /api/doctor/ {
    proxy_pass http://doctor:8000/api/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}

# DevSmith Doctor Dashboard (if using integrated frontend)
location /doctor {
    proxy_pass http://doctor:8000;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

## Environment Variables

Create a `.env` file with these variables (or add to existing):

```bash
# DevSmith Doctor Configuration
DOCTOR_AUTO_APPLY_THRESHOLD=0.8
DOCTOR_MAX_ATTEMPTS=3
DOCTOR_LOG_FIXES=true
DOCTOR_ESCALATE_TO_AGENT=false

# Integration Settings
DOCTOR_ENABLE_NGINXFMT=true
DOCTOR_ENABLE_HADOLINT=true
DOCTOR_ENABLE_COMPOSE_LINT=true
```

## Start the Service

```bash
# Build and start Doctor
docker-compose up -d doctor

# View logs
docker-compose logs -f doctor

# Check health
curl http://localhost:8084/health
```

## Integrate with docker-validate.sh

Update your validation script to optionally trigger Doctor:

```bash
# At the end of docker-validate.sh
if [ "$AUTO_FIX" = "true" ]; then
    echo "🏥 Running DevSmith Doctor auto-fix..."
    curl -X POST http://localhost:8084/api/diagnose
    # Could also call specific fix endpoints
fi
```

Or create a combined script `scripts/validate-and-fix.sh`:

```bash
#!/bin/bash
set -e

echo "🔍 Step 1: Validating Docker setup..."
./scripts/docker-validate.sh

if [ $? -ne 0 ]; then
    echo "❌ Validation failed. Running Doctor..."
    
    echo "🏥 Step 2: Diagnosing issues..."
    curl -X POST http://localhost:8084/api/diagnose | jq '.'
    
    echo "🔧 Step 3: Auto-fixing safe issues..."
    # This would trigger auto-fix via API
    # Implementation depends on your specific needs
    
    echo "🔍 Step 4: Re-validating..."
    ./scripts/docker-validate.sh
fi
```

## Access the Dashboard

Once integrated:

- **API**: http://localhost:8084/api/
- **Dashboard**: http://localhost:3000/doctor (via nginx proxy)
- **Direct**: http://localhost:8084 (if not using nginx proxy)

## CLI Integration

Install the CLI tool:

```bash
# Make CLI executable
chmod +x ./packages/devsmith-doctor/cli/devsmith-doctor

# Optionally link to PATH
sudo ln -s $(pwd)/packages/devsmith-doctor/cli/devsmith-doctor /usr/local/bin/

# Now you can run from anywhere
devsmith-doctor --help
```

## Complete Workflow Example

```bash
# 1. Start all services including Doctor
./scripts/dev.sh

# 2. Make changes to your docker setup
vim docker-compose.yml

# 3. Restart services
docker-compose up -d --build

# 4. Validate (this creates .validation/status.json)
./scripts/docker-validate.sh

# 5. If issues found, run Doctor
devsmith-doctor --mode auto

# 6. Verify fixes worked
./scripts/docker-validate.sh

# 7. View fix history
curl http://localhost:8084/api/history | jq '.'
```

## Monitoring & Alerts

Add to your monitoring setup:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'devsmith-doctor'
    static_configs:
      - targets: ['doctor:8000']
    metrics_path: '/metrics'
```

Or integrate with your existing devsmith-logs:

```python
# Doctor automatically logs to devsmith-logs
# View in logs dashboard with tag: service=doctor
```

## Troubleshooting

### Doctor can't access Docker socket
```bash
# Ensure Docker socket is mounted
docker-compose exec doctor ls -la /var/run/docker.sock

# If permission denied, add doctor user to docker group
docker-compose exec doctor usermod -aG docker app
```

### Can't read validation file
```bash
# Ensure .validation directory exists and has correct permissions
mkdir -p .validation
chmod 755 .validation

# Run docker-validate.sh first
./scripts/docker-validate.sh
```

### API not accessible
```bash
# Check if service is running
docker-compose ps doctor

# Check logs
docker-compose logs doctor

# Test direct access
curl http://localhost:8084/health
```

## Next Steps

1. **Test the integration**: `docker-compose up -d doctor`
2. **Run a diagnosis**: `curl -X POST http://localhost:8084/api/diagnose`
3. **Try the CLI**: `devsmith-doctor --mode interactive`
4. **Add to your workflow**: Update your dev scripts to use Doctor
5. **Monitor fixes**: Check the fix history in the dashboard

---

**See also:**
- [Main README](../README.md) - Full Doctor documentation
- [Pattern Library](../docs/PATTERNS.md) - All fix patterns
- [API Reference](../docs/API.md) - Complete API docs
