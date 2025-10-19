# Aider Tasks for Issue #001

## Current Status
Copilot completed ~70% of Issue #001. The following items need to be fixed/completed.

## Priority 1: Fix docker-compose.yml (CRITICAL)

**Problem:** Current configuration won't work. Services will fail to build and communicate.

**Fix Required:**
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: devsmith-postgres
    environment:
      POSTGRES_DB: devsmith
      POSTGRES_USER: devsmith
      POSTGRES_PASSWORD: ${DB_PASSWORD:-devsmith_local}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./docker/postgres/init-schemas.sql:/docker-entrypoint-initdb.d/01-schemas.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U devsmith -d devsmith"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - devsmith-network

  portal:
    build:
      context: .                          # ← FIX: Root context
      dockerfile: cmd/portal/Dockerfile   # ← FIX: Specify dockerfile
    container_name: devsmith-portal
    environment:
      - PORT=8080                         # ← ADD: Service port
      - DATABASE_URL=postgres://devsmith:${DB_PASSWORD:-devsmith_local}@postgres:5432/devsmith
      - GITHUB_CLIENT_ID=${GITHUB_CLIENT_ID}
      - GITHUB_CLIENT_SECRET=${GITHUB_CLIENT_SECRET}
    depends_on:
      postgres:
        condition: service_healthy        # ← ADD: Wait for postgres
    healthcheck:                          # ← ADD: Health check
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    networks:
      - devsmith-network

  # Repeat for review (port 8081), logs (port 8082), analytics (port 8083)

  nginx:
    image: nginx:alpine
    container_name: devsmith-nginx
    ports:
      - "3000:80"                         # ← FIX: Port mapping
    volumes:
      - ./docker/nginx/nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - portal
      - review
      - logs
      - analytics
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:80/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    networks:
      - devsmith-network

volumes:
  postgres_data:

networks:
  devsmith-network:
    driver: bridge
```

**Reference:** See `.docs/issues/001-copilot-project-scaffolding.md` lines 640-773

---

## Priority 2: Fix docker/nginx/nginx.conf

**Problems:**
1. Listening on port 3000 (should be 80)
2. Missing events block
3. Wrong upstream server definitions (missing ports)
4. Wrong routing paths

**Fix Required:**
```nginx
events {
    worker_connections 1024;
}

http {
    # Logging
    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    # Upstream services (specify ports!)
    upstream portal {
        server portal:8080;
    }

    upstream review {
        server review:8081;
    }

    upstream logs {
        server logs:8082;
    }

    upstream analytics {
        server analytics:8083;
    }

    # Main server block
    server {
        listen 80;              # ← FIX: Listen on 80, not 3000
        server_name localhost;

        # Portal (default)
        location / {            # ← FIX: Root for portal
            proxy_pass http://portal;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Review app
        location /review {
            proxy_pass http://review;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Logs app
        location /logs {
            proxy_pass http://logs;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            # WebSocket support for log streaming
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
        }

        # Analytics app
        location /analytics {
            proxy_pass http://analytics;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Health check endpoint
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}
```

**Reference:** See `.docs/issues/001-copilot-project-scaffolding.md` lines 483-568

---

## Priority 3: Create scripts/ directory

**Missing:**
- `scripts/setup.sh`
- `scripts/dev.sh`
- `scripts/test.sh`

**Reference:** See `.docs/issues/001-copilot-project-scaffolding.md` lines 572-636

**After creating, run:**
```bash
chmod +x scripts/*.sh
```

---

## Priority 4: Verification

After fixes, run these tests from Issue #001 (lines 842-879):

```bash
# Clean slate
make clean

# Run setup
make setup

# Start development environment
make dev
# In another terminal:

# Test health checks
curl http://localhost:3000/health          # Nginx
curl http://localhost:3000/                # Portal (via nginx)
curl http://localhost:3000/review          # Review (via nginx)
curl http://localhost:3000/logs            # Logs (via nginx)
curl http://localhost:3000/analytics       # Analytics (via nginx)

# Check all services healthy
docker-compose ps
# All services should show "healthy"

# Verify database schemas
docker-compose exec postgres psql -U devsmith -d devsmith -c "\dn"
# Should show: portal, reviews, logs, analytics schemas

# Run tests
make test
```

---

## Summary

**Aider needs to:**
1. ✅ Fix docker-compose.yml (critical - won't work without this)
2. ✅ Fix nginx.conf (critical - wrong ports and routing)
3. ✅ Create scripts/ directory
4. ✅ Verify everything works with testing commands
5. ✅ Update devlog with completion notes

**Then Issue #001 will be complete and ready for PR!**
