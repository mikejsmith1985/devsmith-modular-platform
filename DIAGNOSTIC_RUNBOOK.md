# DevSmith Platform UI Diagnostic Runbook
**Date Created:** 2025-11-06  
**Issue:** Port 80 returns 404/stale UI, port 3000 shows updated UI locally

---

## CRITICAL DISCOVERY
Your architecture uses **Traefik on port 3000**, NOT port 80. If you're accessing port 80, something else is serving that port (likely leftover nginx or system web server).

---

## Phase 1: INSTANT CHECKS (Copy/Paste Block)

```bash
# === NETWORK BINDING ANALYSIS ===
echo "=== WHO OWNS PORT 80? ==="
sudo ss -ltnp | grep ':80 ' || echo "Nothing on port 80"
sudo lsof -i :80 2>/dev/null || echo "lsof: Nothing on port 80"

echo -e "\n=== WHO OWNS PORT 3000 (Expected: Traefik)? ==="
sudo ss -ltnp | grep ':3000 ' || echo "Nothing on port 3000"

echo -e "\n=== ALL DOCKER CONTAINERS AND PORT MAPPINGS ==="
docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'

echo -e "\n=== TRAEFIK/NGINX CONTAINERS (Active or Stopped) ==="
docker ps -a | grep -E 'traefik|nginx' || echo "No traefik/nginx containers found"

echo -e "\n=== SYSTEM NGINX SERVICE ==="
systemctl status nginx 2>/dev/null || echo "nginx service not found"
ps aux | grep -E '[n]ginx' || echo "No nginx processes"

# === CONTENT COMPARISON ===
echo -e "\n=== FETCHING CONTENT FROM BOTH PORTS ==="
echo "Port 80 response:"
curl -sS -m 5 -D - http://127.0.0.1:80/ -o /tmp/port80.html 2>&1 | head -20

echo -e "\nPort 3000 response (Traefik -> Portal):"
curl -sS -m 5 -D - http://127.0.0.1:3000/ -o /tmp/port3000.html 2>&1 | head -20

echo -e "\n=== CONTENT HASH COMPARISON ==="
sha256sum /tmp/port80.html /tmp/port3000.html 2>/dev/null || echo "One or both files missing"

echo -e "\n=== FIRST 30 LINES OF PORT 80 HTML ==="
head -n 30 /tmp/port80.html 2>/dev/null || echo "No content from port 80"

echo -e "\n=== FIRST 30 LINES OF PORT 3000 HTML ==="
head -n 30 /tmp/port3000.html 2>/dev/null || echo "No content from port 3000"
```

---

## Phase 2: TRAEFIK ROUTING DIAGNOSTICS

```bash
# === TRAEFIK CONFIGURATION ===
echo "=== TRAEFIK CONTAINER STATUS ==="
docker ps | grep traefik

echo -e "\n=== TRAEFIK LOGS (Last 100 lines) ==="
docker logs devsmith-traefik --tail 100 2>&1 | grep -E 'error|ERROR|warn|WARN|router|service|Configuration' || docker logs devsmith-traefik --tail 100

echo -e "\n=== TRAEFIK LABELS ON ALL SERVICES ==="
for container in $(docker ps --format '{{.Names}}'); do
  echo -e "\n--- $container ---"
  docker inspect $container --format '{{range $k,$v := .Config.Labels}}{{if eq (printf "%.7s" $k) "traefik"}}{{$k}}={{$v}}{{"\n"}}{{end}}{{end}}' 2>/dev/null | head -10
done

echo -e "\n=== TRAEFIK DASHBOARD API (Routers) ==="
curl -sS http://127.0.0.1:8090/api/http/routers 2>/dev/null | jq -r '.[] | {name: .name, rule: .rule, service: .service, status: .status}' 2>/dev/null || echo "Traefik API not accessible"

echo -e "\n=== TRAEFIK DASHBOARD API (Services) ==="
curl -sS http://127.0.0.1:8090/api/http/services 2>/dev/null | jq -r '.[] | {name: .name, serverStatus: .serverStatus}' 2>/dev/null || echo "Traefik API not accessible"
```

---

## Phase 3: CSS/STATIC ASSET VERIFICATION

```bash
# === CSS DELIVERY CHECK ===
echo "=== REQUESTING CSS FROM PORT 3000 (Via Traefik) ==="
curl -sS -I http://127.0.0.1:3000/static/css/devsmith-theme.css 2>&1 | head -15

echo -e "\n=== CSS HASH FROM PORT 3000 ==="
curl -sS http://127.0.0.1:3000/static/css/devsmith-theme.css -o /tmp/css3000.css 2>/dev/null
sha256sum /tmp/css3000.css 2>/dev/null

echo -e "\n=== ACTUAL CSS FILES IN CONTAINERS ==="
for service in portal review logs analytics; do
  echo -e "\n--- $service container CSS files ---"
  docker exec devsmith-modular-platform-${service}-1 ls -lh /app/static/css/*.css 2>/dev/null || echo "No CSS files or container not found"
  docker exec devsmith-modular-platform-${service}-1 sha256sum /app/static/css/devsmith-theme.css 2>/dev/null | cut -d' ' -f1 || echo "CSS not found"
done

echo -e "\n=== HOST CSS FILES (Source) ==="
sha256sum apps/portal/static/css/devsmith-theme.css 2>/dev/null
sha256sum apps/review/static/css/devsmith-theme.css 2>/dev/null
sha256sum apps/logs/static/css/devsmith-theme.css 2>/dev/null
```

---

## Phase 4: DIRECT CONTAINER ACCESS (Bypass Traefik)

```bash
# === DIRECT SERVICE HEALTH CHECKS ===
echo "=== PORTAL (Direct Port 8080) ==="
curl -sS -I http://127.0.0.1:8080/ 2>&1 | head -10

echo -e "\n=== REVIEW (Direct Port 8081) ==="
curl -sS -I http://127.0.0.1:8081/ 2>&1 | head -10

echo -e "\n=== LOGS (Direct Port 8082) ==="
curl -sS -I http://127.0.0.1:8082/ 2>&1 | head -10

echo -e "\n=== ANALYTICS (Direct Port 8083) ==="
curl -sS -I http://127.0.0.1:8083/ 2>&1 | head -10

# === CONTAINER NETWORK IPS ===
echo -e "\n=== CONTAINER NETWORK IPS ==="
for service in portal review logs analytics; do
  ip=$(docker inspect devsmith-modular-platform-${service}-1 --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' 2>/dev/null)
  echo "$service: $ip"
done
```

---

## Phase 5: CACHE DETECTION

```bash
# === BROWSER CACHE HEADERS ===
echo "=== CACHE-CONTROL HEADERS FROM PORT 3000 ==="
curl -sS -I http://127.0.0.1:3000/ 2>&1 | grep -iE 'cache|etag|last-modified|expires'

echo -e "\n=== SERVICE WORKER CHECK ==="
curl -sS -I http://127.0.0.1:3000/service-worker.js 2>&1 | head -5

echo -e "\n=== STATIC FILE CACHE HEADERS ==="
curl -sS -I http://127.0.0.1:3000/static/css/devsmith-theme.css 2>&1 | grep -iE 'cache|etag|last-modified'
```

---

## Phase 6: VOLUME/MOUNT VERIFICATION

```bash
# === DOCKER VOLUME MOUNTS ===
echo "=== PORTAL VOLUME MOUNTS ==="
docker inspect devsmith-modular-platform-portal-1 --format '{{json .Mounts}}' 2>/dev/null | jq -r '.[] | {Source: .Source, Destination: .Destination, Type: .Type}' || echo "Container not found"

echo -e "\n=== FILE TIMESTAMPS IN CONTAINERS ==="
for service in portal review logs analytics; do
  echo -e "\n--- $service ---"
  docker exec devsmith-modular-platform-${service}-1 stat /app/static/css/devsmith-theme.css 2>/dev/null | grep Modify || echo "File not found"
done
```

---

## INTERPRETATION GUIDE

### 1. Port Binding Analysis
- **Expected:** Port 3000 shows `docker-proxy` or Traefik container
- **Port 80 active:** Something else is serving (likely nginx or Apache)
- **Port 80 empty:** Browser might be cached or wrong URL

### 2. Content Hash Comparison
- **Same hash:** Same content, issue is browser cache
- **Different hash:** Port 80 serves different/old content
- **Port 80 404/error:** Traefik routing issue or wrong service

### 3. Traefik Logs
Look for:
- `Configuration loaded` - Good, Traefik found services
- `Router <name> added` - Routes registered
- `404` or `Service unavailable` - Routing broken
- `error` - Configuration problem

### 4. Container CSS Hashes
- **All same hash:** CSS built correctly, likely cache issue
- **Different hashes:** Containers not rebuilt after CSS update
- **Missing files:** Build failed or wrong path

---

## ROOT CAUSE HYPOTHESES (Ranked)

### 1. **ACCESSING WRONG PORT** (90% Likelihood)
- You're accessing `http://localhost:80` instead of `http://localhost:3000`
- Traefik is on port 3000, port 80 may be leftover nginx
- **Fix:** Use `http://localhost:3000` everywhere

### 2. **Leftover nginx on Port 80** (70% Likelihood)
- System nginx or old container still running
- Serving stale static files
- **Fix:** `sudo systemctl stop nginx && sudo systemctl disable nginx`

### 3. **Browser Cache** (60% Likelihood)
- Hard refresh (Ctrl+Shift+R) didn't clear cache
- Service worker caching old version
- **Fix:** Clear site data, use incognito mode, check DevTools Application tab

### 4. **Traefik Misconfiguration** (40% Likelihood)
- StripPrefix middleware eating too much path
- Priority rules causing wrong routing
- **Fix:** Adjust labels, check Traefik dashboard

### 5. **Containers Not Rebuilt** (30% Likelihood)
- CSS updated on host but not in container
- `docker restart` doesn't rebuild images
- **Fix:** `docker-compose up -d --build portal review logs analytics`

### 6. **Volume Mount Serving Old Files** (20% Likelihood)
- Static files mounted from host overriding container build
- **Fix:** Check docker-compose.yml for volume mounts

---

## SHORT-TERM MITIGATIONS (Execute in Order)

```bash
# 1. Stop any system nginx
sudo systemctl stop nginx 2>/dev/null || echo "No system nginx"
docker stop $(docker ps -a | grep nginx | awk '{print $1}') 2>/dev/null || echo "No nginx containers"

# 2. Restart Traefik to reload config
docker restart devsmith-traefik

# 3. Rebuild and restart app containers with fresh CSS
docker-compose up -d --build portal review logs analytics

# 4. Clear browser cache completely
# (Do this manually: DevTools > Application > Clear storage > Clear site data)

# 5. Test with curl (bypasses browser cache)
curl -sS http://127.0.0.1:3000/ | head -30
```

---

## FINAL FIXES (Based on Root Cause)

### If Port 80 is Active (Most Likely)
```bash
# Option A: Stop system nginx permanently
sudo systemctl stop nginx
sudo systemctl disable nginx
sudo systemctl mask nginx

# Option B: Configure nginx to reverse proxy to Traefik
# (Edit /etc/nginx/sites-available/default)
location / {
    proxy_pass http://127.0.0.1:3000;
    proxy_set_header Host $host;
}
```

### If Traefik Misconfigured
```yaml
# Check docker-compose.yml labels for portal:
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.portal.rule=Host(`localhost`)"
  - "traefik.http.routers.portal.entrypoints=web"
  - "traefik.http.services.portal.loadbalancer.server.port=8080"
```

### If CSS Not Updated
```bash
# Full rebuild with no cache
docker-compose down
docker-compose build --no-cache portal review logs analytics
docker-compose up -d

# Verify CSS inside containers
docker exec devsmith-modular-platform-portal-1 sha256sum /app/static/css/devsmith-theme.css
```

### If Browser Cache Persistent
```bash
# Add cache-busting headers in Go handlers
w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
w.Header().Set("Pragma", "no-cache")
w.Header().Set("Expires", "0")
```

---

## VALIDATION CHECKLIST

Run these commands to confirm fix:

```bash
# 1. Verify Traefik owns port 3000
sudo ss -ltnp | grep ':3000' | grep -q docker && echo "✓ Traefik on 3000" || echo "✗ Port 3000 wrong"

# 2. Verify no nginx on port 80
sudo ss -ltnp | grep ':80' && echo "✗ Something on port 80" || echo "✓ Port 80 clear"

# 3. Verify HTML matches
curl -sS http://127.0.0.1:3000/ > /tmp/test1.html
curl -sS http://127.0.0.1:3000/ > /tmp/test2.html
sha256sum /tmp/test1.html /tmp/test2.html | awk '{print $1}' | uniq | wc -l | grep -q 1 && echo "✓ Consistent content" || echo "✗ Content varies"

# 4. Verify CSS loads
curl -sS http://127.0.0.1:3000/static/css/devsmith-theme.css | grep -q '.btn-primary' && echo "✓ CSS has custom classes" || echo "✗ CSS wrong"

# 5. Verify Traefik routes
curl -sS http://127.0.0.1:8090/api/http/routers | jq -r '.[].status' | grep -q "enabled" && echo "✓ Traefik routes enabled" || echo "✗ No routes"

# 6. Browser test
echo "Manual check: Open http://localhost:3000 in browser"
echo "- Check Network tab: CSS should be 200 OK, not 304"
echo "- Check Elements tab: <link> should point to /static/css/devsmith-theme.css"
echo "- Check Console: No 404 errors"
```

---

## AUTOMATED DIAGNOSTIC SCRIPT

Save this as `diagnose-ui.sh` and run with `bash diagnose-ui.sh`:

```bash
#!/bin/bash
set -e
OUTPUT_DIR="/tmp/devsmith-diagnostics-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$OUTPUT_DIR"

echo "DevSmith UI Diagnostics - $(date)" | tee "$OUTPUT_DIR/summary.txt"
echo "======================================" | tee -a "$OUTPUT_DIR/summary.txt"

# Port binding
echo -e "\n[1/8] Port binding analysis..." | tee -a "$OUTPUT_DIR/summary.txt"
sudo ss -ltnp | grep -E ':(80|3000|8080|8081|8082|8083) ' > "$OUTPUT_DIR/ports.txt" 2>&1 || true

# Docker status
echo "[2/8] Docker container status..." | tee -a "$OUTPUT_DIR/summary.txt"
docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' > "$OUTPUT_DIR/docker-ps.txt"

# Content fetch
echo "[3/8] Fetching content..." | tee -a "$OUTPUT_DIR/summary.txt"
curl -sS -D "$OUTPUT_DIR/port80-headers.txt" http://127.0.0.1:80/ -o "$OUTPUT_DIR/port80.html" 2>&1 || echo "Port 80 failed" > "$OUTPUT_DIR/port80.html"
curl -sS -D "$OUTPUT_DIR/port3000-headers.txt" http://127.0.0.1:3000/ -o "$OUTPUT_DIR/port3000.html" 2>&1 || echo "Port 3000 failed" > "$OUTPUT_DIR/port3000.html"

# Hashes
echo "[4/8] Computing hashes..." | tee -a "$OUTPUT_DIR/summary.txt"
sha256sum "$OUTPUT_DIR/port80.html" "$OUTPUT_DIR/port3000.html" > "$OUTPUT_DIR/hashes.txt" 2>&1 || true

# Traefik logs
echo "[5/8] Collecting Traefik logs..." | tee -a "$OUTPUT_DIR/summary.txt"
docker logs devsmith-traefik --tail 200 > "$OUTPUT_DIR/traefik-logs.txt" 2>&1 || echo "Traefik logs failed" > "$OUTPUT_DIR/traefik-logs.txt"

# Traefik API
echo "[6/8] Querying Traefik API..." | tee -a "$OUTPUT_DIR/summary.txt"
curl -sS http://127.0.0.1:8090/api/http/routers | jq '.' > "$OUTPUT_DIR/traefik-routers.json" 2>&1 || echo "[]" > "$OUTPUT_DIR/traefik-routers.json"
curl -sS http://127.0.0.1:8090/api/http/services | jq '.' > "$OUTPUT_DIR/traefik-services.json" 2>&1 || echo "[]" > "$OUTPUT_DIR/traefik-services.json"

# CSS verification
echo "[7/8] Checking CSS files..." | tee -a "$OUTPUT_DIR/summary.txt"
for service in portal review logs analytics; do
  docker exec devsmith-modular-platform-${service}-1 sha256sum /app/static/css/devsmith-theme.css 2>&1 >> "$OUTPUT_DIR/css-hashes.txt" || echo "$service: not found" >> "$OUTPUT_DIR/css-hashes.txt"
done

# System nginx
echo "[8/8] Checking for nginx..." | tee -a "$OUTPUT_DIR/summary.txt"
systemctl status nginx > "$OUTPUT_DIR/nginx-status.txt" 2>&1 || echo "nginx not running" > "$OUTPUT_DIR/nginx-status.txt"
ps aux | grep -E '[n]ginx' > "$OUTPUT_DIR/nginx-processes.txt" 2>&1 || echo "No nginx processes" > "$OUTPUT_DIR/nginx-processes.txt"

# Summary
echo -e "\n======================================" | tee -a "$OUTPUT_DIR/summary.txt"
echo "Diagnostics complete!" | tee -a "$OUTPUT_DIR/summary.txt"
echo "Results saved to: $OUTPUT_DIR" | tee -a "$OUTPUT_DIR/summary.txt"
echo -e "\nKey files:" | tee -a "$OUTPUT_DIR/summary.txt"
ls -lh "$OUTPUT_DIR" | tee -a "$OUTPUT_DIR/summary.txt"

# Create tarball
tar -czf "$OUTPUT_DIR.tar.gz" -C /tmp "$(basename $OUTPUT_DIR)"
echo -e "\nCompressed archive: $OUTPUT_DIR.tar.gz"
echo "Share this file for analysis if needed."
```

---

## EXPECTED OUTPUTS FOR HEALTHY SYSTEM

### Port Binding (ss -ltnp | grep ':3000')
```
tcp  LISTEN  0  4096  0.0.0.0:3000  0.0.0.0:*  users:(("docker-proxy",pid=123456))
```

### Traefik Logs (docker logs devsmith-traefik --tail 50)
```
time="..." level=info msg="Configuration loaded from flags."
time="..." level=info msg="Traefik version 2.10..."
time="..." level=info msg="Router portal@docker added"
time="..." level=info msg="Router review@docker added"
time="..." level=info msg="Router logs@docker added"
```

### Content Hash (sha256sum)
```
a1b2c3d4... /tmp/port3000.html
```

### CSS Response (curl -I http://127.0.0.1:3000/static/css/devsmith-theme.css)
```
HTTP/1.1 200 OK
Content-Type: text/css
Content-Length: 21504
```

---

## NEXT STEPS BASED ON DIAGNOSTIC OUTPUT

**Copy the output of Phase 1 commands and paste back. I will:**
1. Identify exact root cause
2. Provide specific fix commands
3. Explain why it happened
4. Suggest prevention measures

**Most likely fix will be:** Use `http://localhost:3000` instead of `http://localhost:80`, or stop system nginx if running.
