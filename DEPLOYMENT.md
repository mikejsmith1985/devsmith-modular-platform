# DevSmith Platform - Deployment Guide

**Version:** 1.0 Beta  
**Last Updated:** November 12, 2025  
**Target:** Beta Users - Self-Hosted Deployment

---

## 📋 Prerequisites

### Required Software

| Software | Minimum Version | Recommended | Purpose |
|----------|----------------|-------------|---------|
| **Docker** | 24.0+ | 27.0+ | Container runtime |
| **Docker Compose** | 2.20+ | 2.30+ | Multi-container orchestration |
| **Git** | 2.30+ | Latest | Source code management |
| **Ollama** | 0.1.0+ | Latest | AI model hosting (local) |

### System Requirements

Choose your configuration based on which AI model size you plan to use:

#### Option 1: Small Models (7B parameters)
- **CPU:** 4 cores minimum (8 cores recommended)
- **RAM:** 12GB minimum (16GB recommended)
  - 4GB for DevSmith services
  - 8GB for Ollama + 7B model
- **Storage:** 20GB free space
  - 10GB for Docker images
  - 5GB for 7B AI model
  - 5GB for database/logs
- **Best For:** Development, small teams, testing

#### Option 2: Medium Models (16B parameters)
- **CPU:** 8 cores minimum (16 cores recommended)
- **RAM:** 20GB minimum (32GB recommended)
  - 4GB for DevSmith services
  - 16GB for Ollama + 16B model
- **Storage:** 30GB free space
  - 10GB for Docker images
  - 15GB for 16B AI model
  - 5GB for database/logs
- **Best For:** Production use, larger codebases

#### Option 3: Large Models (32B parameters)
- **CPU:** 16 cores minimum (32 cores recommended)
- **RAM:** 40GB minimum (64GB recommended)
  - 4GB for DevSmith services
  - 32GB for Ollama + 32B model
- **Storage:** 50GB free space
  - 10GB for Docker images
  - 35GB for 32B AI model
  - 5GB for database/logs
- **Best For:** Enterprise use, maximum accuracy

### GitHub OAuth Application

You must register a GitHub OAuth application to enable user authentication.

**Steps:**
1. Go to [GitHub Settings → Developer Settings → OAuth Apps](https://github.com/settings/developers)
2. Click **"New OAuth App"**
3. Fill in the details:
   - **Application name:** DevSmith Platform (or your preferred name)
   - **Homepage URL:** `http://localhost:3000` (or your domain)
   - **Authorization callback URL:** `http://localhost:3000/auth/github/callback`
4. Click **"Register application"**
5. Copy the **Client ID** (you'll need this for `.env`)
6. Generate a **Client Secret** (you'll need this for `.env`)

**⚠️ Important:** Keep your Client Secret secure. Never commit it to version control.

---

## 🚀 Quick Start (15 Minutes)

### Step 1: Install Ollama (5 minutes)

Ollama provides local AI model hosting. It's required for the Review app and optional for other features.

#### Linux
```bash
curl -fsSL https://ollama.com/install.sh | sh
```

#### macOS
```bash
brew install ollama
```

#### Windows
Download from [ollama.com/download](https://ollama.com/download)

**Verify installation:**
```bash
ollama --version
# Should output: ollama version 0.x.x
```

### Step 2: Pull an AI Model (3-5 minutes)

Choose ONE model based on your system resources:

#### For 12-16GB RAM (Recommended for Beta)
```bash
ollama pull qwen2.5-coder:7b
```
**Size:** ~4.7GB  
**Speed:** Fast  
**Quality:** Good for most code review tasks

#### For 20-32GB RAM
```bash
ollama pull deepseek-coder-v2:16b
```
**Size:** ~9.8GB  
**Speed:** Moderate  
**Quality:** Excellent for complex code analysis

#### For 40GB+ RAM
```bash
ollama pull llama3.1:70b
```
**Size:** ~40GB  
**Speed:** Slower  
**Quality:** State-of-the-art code understanding

**Verify model downloaded:**
```bash
ollama list
# Should show your downloaded model
```

### Step 3: Start Ollama Service

#### Linux/macOS
```bash
ollama serve
```
Leave this terminal open. Ollama runs on `http://localhost:11434`

**OR** run as background service:
```bash
# macOS
brew services start ollama

# Linux (systemd)
sudo systemctl enable ollama
sudo systemctl start ollama
```

#### Windows
Ollama runs automatically as a service after installation.

**Test Ollama:**
```bash
curl http://localhost:11434/api/tags
# Should return JSON with your models
```

### Step 4: Clone Repository (1 minute)

```bash
git clone https://github.com/mikejsmith1985/devsmith-modular-platform.git
cd devsmith-modular-platform
```

### Step 5: Configure Environment (2 minutes)

Create your environment configuration:

```bash
cp .env.example .env
nano .env  # or use your preferred editor
```

**Required environment variables:**

```bash
# Database Configuration
DATABASE_URL=postgresql://devsmith:devsmith@postgres:5432/devsmith
POSTGRES_USER=devsmith
POSTGRES_PASSWORD=devsmith
POSTGRES_DB=devsmith

# Redis Configuration
REDIS_URL=redis:6379

# GitHub OAuth (from Step "Prerequisites")
GITHUB_CLIENT_ID=your_github_client_id_here
GITHUB_CLIENT_SECRET=your_github_client_secret_here
REDIRECT_URI=http://localhost:3000/auth/github/callback

# JWT Secret (generate a random string)
JWT_SECRET=$(openssl rand -base64 32)

# AI Configuration
OLLAMA_ENDPOINT=http://host.docker.internal:11434

# ⚠️ IMPORTANT: AI models are configured through the AI Factory UI (/llm-config)
# The OLLAMA_ENDPOINT is only needed so AI Factory can connect to your Ollama instance.
# DO NOT configure models via environment variables - they will be ignored.

# Service Ports (defaults are fine for most users)
PORT=3000
PORTAL_PORT=3001
REVIEW_PORT=8081
LOGS_PORT=8082
ANALYTICS_PORT=8083
```

**Generate JWT Secret (Linux/macOS):**
```bash
echo "JWT_SECRET=$(openssl rand -base64 32)" >> .env
```

**Generate JWT Secret (Windows PowerShell):**
```powershell
$secret = [Convert]::ToBase64String([byte[]](1..32 | ForEach-Object { Get-Random -Maximum 256 }))
Add-Content .env "JWT_SECRET=$secret"
```

### Step 6: Deploy DevSmith Platform (2 minutes)

**✨ NEW: Atomic Deployment Process**

```bash
# Deploy portal with atomic frontend+backend build
./scripts/deploy-portal.sh

# Start remaining services
docker-compose up -d

# Database migrations run automatically on service startup
# No manual commands needed - services are idempotent
```

**Note on Database Migrations:**
All services automatically run database migrations on startup. The Portal service will create the following tables if they don't exist:
- `portal.users` - User accounts
- `portal.llm_configs` - AI model configurations
- `portal.app_llm_preferences` - App-specific model preferences
- `portal.llm_usage_logs` - Token usage tracking

**You do NOT need to run manual migration commands.** Services are safe to restart multiple times - migrations are idempotent and will only create missing tables.

This atomic deployment:
- ✅ **Builds frontend inside Docker** (no local node_modules needed)
- ✅ **Deploys frontend + backend together** (no version drift)
- ✅ **Includes health checks** (automatic verification)
- ✅ **Single command operation** (eliminates manual errors)
- ✅ **Automatic database migrations** (creates tables on first startup)

**Alternative: Traditional method (if atomic deployment fails):**
```bash
docker-compose up -d
```

**Watch startup progress:**
```bash
docker-compose logs -f
# Press Ctrl+C to stop following logs
```

### Step 7: Verify Installation (2 minutes)

**Check all services are healthy:**
```bash
docker-compose ps
```

Expected output:
```
SERVICE     STATE     STATUS
frontend    running   Up 2 minutes (healthy)
portal      running   Up 2 minutes (healthy)
review      running   Up 2 minutes (healthy)
logs        running   Up 2 minutes (healthy)
analytics   running   Up 2 minutes (healthy)
postgres    running   Up 2 minutes (healthy)
redis       running   Up 2 minutes (healthy)
traefik     running   Up 2 minutes
jaeger      running   Up 2 minutes
```

**Test health endpoints:**
```bash
# Gateway health
curl http://localhost:3000/health
# Should return: {"status":"ok"}

# Portal health
curl http://localhost:3000/api/portal/health
# Should return: {"status":"healthy",...}

# Logs health
curl http://localhost:3000/api/logs/health
# Should return: {"status":"healthy",...}
```

### Step 8: First Login (1 minute)

1. Open browser: **http://localhost:3000**
2. Click **"Login with GitHub"**
3. Authorize the DevSmith application
4. You'll be redirected to the Dashboard

**🎉 Success!** You should see the Dashboard with 4 app cards:
- Health Monitoring
- Code Review
- AI Factory
- Projects

**⚠️ NEXT STEP REQUIRED:** Before using Review, Logs, or Analytics apps, you MUST configure at least one AI model in the AI Factory. See the next section for instructions.

---

## 🤖 Configure AI Models (AI Factory) ⚠️ REQUIRED

**IMPORTANT:** AI model configuration is now ONLY done through the AI Factory web interface. Environment variables for AI models are no longer supported.

After first login, you MUST configure at least one AI model before using Review, Logs, or Analytics features:

### Why AI Factory?

The AI Factory provides:
- ✅ **Per-user model selection** - Different team members can use different models
- ✅ **Per-app preferences** - Use fast local models for logs, powerful API models for code review
- ✅ **Secure API key storage** - Encrypted in database, never in environment files
- ✅ **Real-time switching** - Change models without restarting services
- ✅ **Easy testing** - Test connection before saving

### Step 1: Navigate to AI Factory

From Dashboard, click the **"AI Factory"** card (or navigate to `/llm-config`).

### Step 2: Add Your First Model (Required)

You must add at least ONE model before using AI-powered features.

#### Option A: Add Local Ollama Model (Recommended for Beta)

1. Click **"Add Model"** button
2. Fill in the form:
   - **Name:** "Local Qwen Coder" (or any descriptive name)
   - **Provider:** Select **"Ollama (Local)"**
   - **Model Name:** `qwen2.5-coder:7b` (must match the model you pulled earlier)
   - **Endpoint:** `http://host.docker.internal:11434` (pre-filled, connects to your Ollama instance)
   - **API Key:** Leave blank (not needed for Ollama)
   - **Set as Default:** ✅ Check this box
3. Click **"Test Connection"** 
   - Should show: ✅ **"Connection successful! Model is available."**
   - If it fails, verify:
     - Ollama is running (`ollama list` should work)
     - Model is pulled (`ollama list` should show `qwen2.5-coder:7b`)
     - Endpoint URL is correct
4. Click **"Save Model"**

**✅ Success!** You now have a working AI model configured.

#### Option B: Add Cloud API Model (OpenAI)

If you have an OpenAI API key:

1. Click **"Add Model"**
2. Fill in the form:
   - **Name:** "GPT-4 Turbo"
   - **Provider:** Select **"OpenAI (GPT)"**
   - **Model Name:** `gpt-4-turbo-preview` (or `gpt-3.5-turbo` for lower cost)
   - **API Key:** Paste your OpenAI API key (starts with `sk-`)
   - **Endpoint:** Leave as default (OpenAI official endpoint)
   - **Set as Default:** ✅ Check if you want this as primary model
3. Click **"Test Connection"** - should succeed if key is valid
4. Click **"Save Model"**

**💰 Cost Note:** OpenAI charges per request. `gpt-3.5-turbo` is ~10x cheaper than `gpt-4-turbo`.

#### Option C: Add Cloud API Model (Anthropic Claude)

If you have an Anthropic API key:

1. Click **"Add Model"**
2. Fill in the form:
   - **Name:** "Claude 3.5 Sonnet"
   - **Provider:** Select **"Anthropic (Claude)"**
   - **Model Name:** `claude-3-5-sonnet-20241022` (recommended) or `claude-3-opus-20240229`
   - **API Key:** Paste your Anthropic API key (starts with `sk-ant-`)
   - **Endpoint:** Leave as default (Anthropic official endpoint)
   - **Set as Default:** ✅ Check if you want this as primary model
3. Click **"Test Connection"** - should succeed if key is valid
4. Click **"Save Model"**

**💡 Quality Tip:** Claude is generally considered best for code analysis and detailed explanations.

### Step 3: Configure App Preferences (Optional)

By default, all apps use your "default" model. But you can customize:

1. In the **"App Preferences"** section, you'll see dropdowns:
   - **Review App:** Which model to use for code analysis
   - **Logs App:** Which model to use for log pattern analysis
   - **Analytics App:** Which model to use for insights
2. Select different models for each app (or leave as default)
3. Changes save automatically

**Example Configuration:**
- Review App: `claude-3-5-sonnet` (best quality for code review)
- Logs App: `qwen2.5-coder:7b` (fast local model, sufficient for logs)
- Analytics App: `gpt-4-turbo` (powerful for insights)

### Step 4: Verify Configuration

1. Navigate to **Review** app
2. Paste some test code:
   ```go
   func add(a, b int) int {
       return a + b
   }
   ```
3. Click **"Analyze"**
4. Should see AI analysis results (no 500 error)

**✅ If analysis works:** Configuration successful!  
**❌ If you get errors:** See Troubleshooting section below.

---

## 🔧 Troubleshooting

### "No AI model configured" Error

**Symptom:** Error message: "No AI model configured. Please configure an AI model in AI Factory (/llm-config)"

**Solution:**
1. Go to http://localhost:3000/llm-config
2. Add at least ONE model (see Step 2 above)
3. Set it as default
4. Try analysis again

### "AI Service Unavailable" or 500 Error

**Symptom:** Analysis fails with generic error

**Possible Causes & Solutions:**

#### 1. Ollama Not Running
```bash
# Check if Ollama is accessible
curl http://localhost:11434/api/tags

# If fails, start Ollama
ollama serve
```

#### 2. Model Not Pulled
```bash
# Check which models are available
ollama list

# If your configured model isn't listed, pull it
ollama pull qwen2.5-coder:7b
```

#### 3. Wrong Endpoint URL
- Ollama endpoint should be: `http://host.docker.internal:11434`
- This special hostname allows Docker containers to reach your host machine
- On Linux, you may need: `http://172.17.0.1:11434` instead

#### 4. Invalid API Key (for cloud providers)
- Go to AI Factory
- Click "Edit" on the model
- Click "Test Connection"
- If fails, verify your API key is correct and has credits/quota

#### 5. Model Name Mismatch
- In AI Factory, model name must EXACTLY match what Ollama has
- Example: `qwen2.5-coder:7b` not `qwen2.5-coder` or `qwen2.5-coder:latest`

### Check Review Service Logs

If issues persist, check what the Review service sees:

```bash
# View Review service logs
docker-compose logs review --tail=50

# Look for lines containing:
# - "Initializing AI client" (should show portal_url)
# - "GET /api/portal/app-llm-preferences" (Portal API calls)
# - Any error messages about models or providers
```

### Verify Portal API Responds

```bash
# Test Portal API (will fail auth, but should not 404)
curl http://localhost:3000/api/portal/app-llm-preferences

# Should return: {"error":"Authentication required"}
# If returns 404, Portal service may not be running
```

---

## 📖 Using the Platform

### Health Monitoring

1. Click **"Health"** from Dashboard
2. View system metrics:
   - Service status (all 9 services)
   - Database connectivity
   - Redis cache status
   - API response times

**Coming Soon:**
- Log aggregation from your projects
- Error trend analysis
- Alert configuration

### Code Review

1. Click **"Review"** from Dashboard
2. Choose input method:
   - **Paste Code:** Paste code snippet directly
   - **GitHub Repo:** Connect your GitHub repository
3. Select reading mode:
   - **Preview:** Quick overview (30 seconds)
   - **Skim:** Fast analysis (1-2 minutes)
   - **Scan:** Moderate depth (3-5 minutes)
   - **Detailed:** Thorough review (5-10 minutes)
   - **Strategic:** Architectural analysis (10-15 minutes)
4. Click **"Analyze"**
5. Review AI-generated insights

**⚠️ Note:** First analysis may take longer as AI model warms up.

### Projects (Cross-Repo Logging)

See **[API_INTEGRATION.md](./API_INTEGRATION.md)** for complete guide on:
- Creating projects
- Generating API keys
- Integrating external applications
- Sending logs via batch API

### Analytics

**Status:** Coming soon in future release

---

## 🔧 Troubleshooting

### Services Won't Start

**Symptom:** `docker-compose up -d` fails or services show "unhealthy"

**Solutions:**

1. **Check port conflicts:**
```bash
# Check if ports are already in use
lsof -i :3000  # Traefik gateway
lsof -i :5432  # PostgreSQL
lsof -i :6379  # Redis
```
If ports are in use, either stop the conflicting service or change ports in `docker-compose.yml`.

2. **Check Docker resources:**
```bash
docker system df
# Ensure you have enough disk space
```

3. **View service logs:**
```bash
docker-compose logs portal
docker-compose logs review
docker-compose logs logs
```

4. **Full restart:**
```bash
docker-compose down -v  # ⚠️ This deletes database data
docker-compose up -d --build
```

### Ollama Connection Failed

**Symptom:** Review app shows "AI service unavailable" error

**Solutions:**

1. **Verify Ollama is running:**
```bash
curl http://localhost:11434/api/tags
```
If this fails, start Ollama: `ollama serve`

2. **Check model is pulled:**
```bash
ollama list
# Should show your model (e.g., qwen2.5-coder:7b)
```

3. **Test from inside Docker:**
```bash
docker-compose exec review curl http://host.docker.internal:11434/api/tags
```

4. **Check AI Factory configuration:**
   - Go to http://localhost:3000/llm-config
   - Verify Ollama model is added
   - Click "Test Connection" - should be green ✅

### GitHub OAuth Login Fails

**Symptom:** "Invalid redirect_uri" or "Application suspended" error

**Solutions:**

1. **Verify `.env` configuration:**
```bash
cat .env | grep GITHUB
# GITHUB_CLIENT_ID should match GitHub app
# REDIRECT_URI should be: http://localhost:3000/auth/github/callback
```

2. **Check GitHub OAuth app settings:**
   - Go to [GitHub OAuth Apps](https://github.com/settings/developers)
   - Verify "Authorization callback URL" matches exactly: `http://localhost:3000/auth/github/callback`
   - Ensure app is not suspended

3. **Regenerate Client Secret:**
   - Generate new secret in GitHub OAuth app settings
   - Update `GITHUB_CLIENT_SECRET` in `.env`
   - Restart: `docker-compose restart portal`

### Database Migration Errors

**Symptom:** Services crash with "relation does not exist" errors

**Important:** As of 2025-11-16, all services run database migrations automatically on startup. If you see migration errors, it indicates a code issue, not a deployment issue.

**Solutions:**

1. **Check migration status:**
```bash
docker-compose exec postgres psql -U devsmith -d devsmith -c "\dt portal.*"
# Should show: portal.users, portal.llm_configs, portal.app_llm_preferences, portal.llm_usage_logs

docker-compose exec postgres psql -U devsmith -d devsmith -c "\dt logs.*"
# Should show: logs.entries, logs.projects, logs.ai_analysis, etc.
```

2. **Check service logs for migration errors:**
```bash
docker-compose logs portal | grep migration
# Should show: "Running database migrations..." and "Database migrations completed successfully"

docker-compose logs logs | grep migration
# Should show successful migration execution
```

3. **If migrations fail, restart the service:**
```bash
docker-compose restart portal
docker-compose logs portal --tail=100
```

4. **Reset database (⚠️ destroys all data - last resort only):**
```bash
docker-compose down -v
docker-compose up -d
# All services will automatically create tables on first startup
```

### Performance Issues

**Symptom:** Slow AI analysis, high CPU usage, system freezes

**Solutions:**

1. **Use smaller AI model:**
```bash
# Stop current model
ollama rm deepseek-coder-v2:16b

# Pull smaller model
ollama pull qwen2.5-coder:7b

# Update AI Factory to use new model
```

2. **Limit Docker resources:**
Edit `docker-compose.yml` and add:
```yaml
services:
  review:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
```

3. **Use cloud AI provider:**
   - Add OpenAI or Anthropic in AI Factory
   - Switch app preferences to cloud model
   - Much faster (but costs money)

---

## 🛠️ Advanced Configuration

### Custom Domain Setup

To use a custom domain instead of `localhost:3000`:

1. **Update `.env`:**
```bash
REDIRECT_URI=https://devsmith.yourdomain.com/auth/github/callback
```

2. **Update GitHub OAuth app:**
   - Change "Authorization callback URL" to match

3. **Configure reverse proxy (Nginx/Apache):**
```nginx
server {
    listen 80;
    server_name devsmith.yourdomain.com;
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

4. **Setup SSL with Let's Encrypt:**
```bash
certbot --nginx -d devsmith.yourdomain.com
```

### Multi-User Production Setup

For production use with multiple users:

1. **Use external database:**
```bash
# .env
DATABASE_URL=postgresql://user:pass@your-db-server:5432/devsmith
```

2. **Use external Redis:**
```bash
# .env
REDIS_URL=your-redis-server:6379
```

3. **Add rate limiting:** (See ARCHITECTURE.md)

4. **Setup monitoring:** (Jaeger UI: http://localhost:16686)

5. **Configure backups:**
```bash
# Daily PostgreSQL backup
docker-compose exec postgres pg_dump -U devsmith devsmith > backup-$(date +%Y%m%d).sql
```

---

## 📊 Monitoring & Observability

### Jaeger Tracing UI

Access distributed tracing at **http://localhost:16686**

- View request flows across services
- Identify performance bottlenecks
- Debug cross-service issues

### Service Logs

```bash
# View all logs
docker-compose logs -f

# View specific service
docker-compose logs -f review

# View last 100 lines
docker-compose logs --tail=100 portal
```

### Health Checks

All services expose health endpoints:

```bash
curl http://localhost:3000/api/portal/health
curl http://localhost:3000/api/review/health
curl http://localhost:3000/api/logs/health
curl http://localhost:3000/api/analytics/health
```

---

## 🔄 Updating DevSmith / Deploying Code Changes

### ✨ NEW: Atomic Deployment (Recommended)

**Portal Service (Frontend + Backend):**
```bash
./scripts/deploy-portal.sh
```

This script:
- ✅ Builds frontend inside Docker (eliminates node_modules sync issues)
- ✅ Compiles backend with embedded frontend (atomic deployment)
- ✅ Includes automatic health checks and verification
- ✅ Displays deployment status and bundle version
- ✅ Fails fast with clear error messages

**Other Services:**
```bash
docker-compose up -d --build <service>
```

**Note:** Services automatically run database migrations on startup. No manual migration commands needed.

### ⚠️ Legacy: Manual Frontend Deployment (Deprecated)

**❌ OLD METHOD (Error-Prone):**
```bash
cd frontend && npm run build      # Can fail or be skipped
cp -r dist/* ../apps/portal/static/  # Manual copy step
docker-compose up -d --build portal   # May use old files
```

**Issues with old method:**
- Manual steps can be forgotten or fail
- Frontend/backend version drift
- No automatic verification
- Hard to rollback on failure

**✅ NEW METHOD (Atomic):**
```bash
./scripts/deploy-portal.sh  # One command, atomic operation
```
```

### Standard Update Procedure

When you update code (via `git pull` or local development):

```bash
# Stop services (if running)
docker-compose down

# Pull latest code (if updating from git)
git pull origin main

# Rebuild each service with mandatory script
./scripts/rebuild-service.sh portal
./scripts/rebuild-service.sh review
./scripts/rebuild-service.sh logs
./scripts/rebuild-service.sh analytics

# Verify all services healthy
docker-compose ps
```

### Single Service Deployment

If you only modified one service:

```bash
# Example: Updated logs service
./scripts/rebuild-service.sh logs

# Script automatically:
# 1. Stops container
# 2. Removes container
# 3. Removes image (prevents cache)
# 4. Builds fresh with --no-cache
# 5. Starts container
# 6. Verifies deployment (5 checks)
```

### What the Verification Checks

The rebuild script automatically verifies:
- ✅ Container is running
- ✅ Container age < 120 seconds (catches stale deployments)
- ✅ Health endpoint responds
- ✅ No recent errors in logs
- ✅ Build timestamp present

**If any check fails**, the script will report the error and exit.

### Example Output

```
🔨 MANDATORY REBUILD: logs
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Step 1/6: Stopping container...
✅ Container stopped

Step 2/6: Removing container...
✅ Container removed

Step 3/6: Removing image...
✅ Image removed

Step 4/6: Building with fresh cache...
[+] Building 45.2s (30/30) FINISHED
✅ Build completed (45s)

Step 5/6: Starting container...
✅ Container started

Step 6/6: Verifying deployment...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🔍 DEPLOYMENT VERIFICATION: logs
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1️⃣ Checking if container is running...
✅ Container is running

2️⃣ Checking container age...
✅ Container is 8s old (fresh deployment)

3️⃣ Checking health endpoint...
✅ Health endpoint responding

4️⃣ Checking for recent errors...
✅ No recent errors found

5️⃣ Checking build timestamp...
✅ Build timestamp present in logs

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ ALL CHECKS PASSED - Deployment verified!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Troubleshooting Failed Deployments

#### Container Age Check Failed
```
❌ FAIL: Container is 2100s old (expected < 120s)
```
**Cause:** Container wasn't rebuilt, using cached code  
**Fix:** Ensure script completes all 6 steps

#### Health Endpoint Failed
```
❌ FAIL: Health endpoint not responding
```
**Cause:** Service failed to start  
**Fix:** Check logs: `docker logs devsmith-modular-platform-logs-1 --tail=50`

#### Recent Errors Found
```
❌ FAIL: Found recent errors in container logs
```
**Cause:** Runtime error after startup  
**Fix:** Review logs and fix code issue

---

## ⚙️ Environment Variables Reference

### Core Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `REDIS_URL` | Yes | - | Redis connection string |
| `JWT_SECRET` | Yes | - | Secret for JWT token signing |
| `GITHUB_CLIENT_ID` | Yes | - | GitHub OAuth app client ID |
| `GITHUB_CLIENT_SECRET` | Yes | - | GitHub OAuth app secret |
| `REDIRECT_URI` | Yes | - | OAuth callback URL |

### AI Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OLLAMA_ENDPOINT` | No | `http://host.docker.internal:11434` | Ollama API endpoint (only needed for AI Factory to connect) |

**⚠️ IMPORTANT:** The `OLLAMA_MODEL`, `ANTHROPIC_API_KEY`, and `OPENAI_API_KEY` environment variables are **NO LONGER SUPPORTED**. All AI model configuration must be done through the **AI Factory UI** (`/llm-config`). This ensures:
- Per-user model preferences
- Secure API key storage (encrypted in database)
- Real-time model switching without service restarts
- Per-app model selection

### Service Ports

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `3000` | Traefik gateway port |
| `PORTAL_PORT` | No | `3001` | Portal service internal port |
| `REVIEW_PORT` | No | `8081` | Review service internal port |
| `LOGS_PORT` | No | `8082` | Logs service internal port |
| `ANALYTICS_PORT` | No | `8083` | Analytics service internal port |

---

## 🆘 Getting Help

### Documentation

- **[USER_GUIDE.md](./USER_GUIDE.md)** - Feature walkthroughs
- **[API_INTEGRATION.md](./API_INTEGRATION.md)** - External app integration
- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - System architecture

### Community Support

- **GitHub Issues:** [Report bugs or request features](https://github.com/mikejsmith1985/devsmith-modular-platform/issues)
- **Discussions:** [Ask questions](https://github.com/mikejsmith1985/devsmith-modular-platform/discussions)

### Beta Feedback

As a beta user, your feedback is invaluable! Please report:
- Installation issues
- UI/UX problems
- Performance concerns
- Feature requests
- Documentation gaps

**Open an issue:** https://github.com/mikejsmith1985/devsmith-modular-platform/issues/new

---

**🎉 You're all set!** Enjoy using DevSmith Platform.
