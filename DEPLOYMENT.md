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
OLLAMA_MODEL=qwen2.5-coder:7b  # Match the model you pulled

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

### Step 6: Start DevSmith Platform (2 minutes)

```bash
docker-compose up -d
```

This will:
- Pull Docker images (~2GB download on first run)
- Start 9 services (PostgreSQL, Redis, Traefik, Frontend, Portal, Review, Logs, Analytics, Jaeger)
- Run database migrations automatically
- Wait for health checks to pass

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

---

## 🤖 Configure AI Models (AI Factory)

After first login, configure your AI models through the web interface:

### Step 1: Navigate to AI Factory

From Dashboard, click the **"AI Factory"** card.

### Step 2: Add Ollama Model

1. Click **"Add Model"** button
2. Fill in the form:
   - **Name:** "Local Qwen Coder" (or any name you prefer)
   - **Provider:** Select "Ollama (Local)"
   - **Model:** `qwen2.5-coder:7b` (or the model you pulled)
   - **API Key:** Leave blank (not needed for Ollama)
   - **Endpoint:** `http://host.docker.internal:11434` (pre-filled)
   - **Set as Default:** ✅ Check this box
3. Click **"Test Connection"** - should show ✅ Success
4. Click **"Save"**

### Step 3: Configure App Preferences

Set which model each app should use:

1. In the **"App Preferences"** section, you'll see dropdowns for each app:
   - **Review App:** Select your Ollama model
   - **Logs App:** Select your Ollama model (or leave as default)
2. Changes save automatically

### Optional: Add Cloud AI Providers

If you have API keys for cloud providers:

#### Add OpenAI
1. Click **"Add Model"**
2. Select **"OpenAI (GPT)"**
3. Enter your OpenAI API key
4. Select model (e.g., `gpt-4-turbo-preview`)
5. Save

#### Add Anthropic (Claude)
1. Click **"Add Model"**
2. Select **"Anthropic (Claude)"**
3. Enter your Anthropic API key
4. Select model (e.g., `claude-3-5-sonnet-20241022`)
5. Save

**💡 Tip:** You can switch between models anytime. Local Ollama is free but slower. Cloud APIs are faster but cost money.

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

**Solutions:**

1. **Check migration status:**
```bash
docker-compose exec postgres psql -U devsmith -d devsmith -c "\dt logs.*"
# Should show: logs.entries, logs.projects, logs.ai_analysis, etc.
```

2. **Manually run migrations:**
```bash
docker-compose exec logs /logs migrate
```

3. **Reset database (⚠️ destroys data):**
```bash
docker-compose down -v
docker-compose up -d
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

## 🔄 Updating DevSmith

To update to a newer version:

```bash
# Stop services
docker-compose down

# Pull latest code
git pull origin main

# Rebuild images
docker-compose build

# Start services (migrations run automatically)
docker-compose up -d

# Verify health
docker-compose ps
```

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
| `OLLAMA_ENDPOINT` | No | `http://host.docker.internal:11434` | Ollama API endpoint |
| `OLLAMA_MODEL` | No | `qwen2.5-coder:7b` | Default AI model |

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
