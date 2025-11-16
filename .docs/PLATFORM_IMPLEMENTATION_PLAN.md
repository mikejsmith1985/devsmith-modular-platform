# DevSmith Platform: Complete Implementation Plan

**Version:** 2.0  
**Created:** 2025-11-05  
**Status:** Ready for Implementation  
**Priority:** Infrastructure → Testing → Features

---

## Executive Summary

This document consolidates all platform improvements into a single prioritized implementation plan. Priority is based on:

1. **Infrastructure First** - Redis (session management) + Traefik (gateway resilience)
2. **Quality Assurance** - Comprehensive E2E testing with visual validation
3. **User Experience** - Consistent styling across all apps
4. **Compliance** - Automated enforcement of development standards

**Estimated Total Duration:** 2-3 weeks

---

## Priority 1: Infrastructure (Week 1) - HIGHEST ROI

### 1.1 Redis Session Store (2-3 days)

**Business Value:**
- ✅ **Single Sign-On** - Login once, access all apps (solves #1 user complaint)
- ✅ **Central logout** - Logout from Portal invalidates all sessions
- ✅ **Session management** - Admin can view/revoke active sessions
- ✅ **Audit trail** - Track session creation/destruction in logs
- ✅ **Scalability** - Redis handles millions of sessions (industry standard)

**Technical Implementation:**

#### Phase 1.1.1: Add Redis to docker-compose.yml (30 min)

```yaml
# docker-compose.yml
services:
  redis:
    image: redis:7-alpine
    container_name: devsmith-redis
    command: redis-server --maxmemory 256mb --maxmemory-policy allkeys-lru
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    networks:
      - devsmith-network
    restart: unless-stopped

volumes:
  redis-data:
    driver: local
```

#### Phase 1.1.2: Create Session Package (2 hours)

```go
// internal/session/redis_store.go
package session

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    "github.com/redis/go-redis/v9"
)

type RedisStore struct {
    client *redis.Client
    ttl    time.Duration
}

type Session struct {
    SessionID        string                 `json:"session_id"`
    UserID           int                    `json:"user_id"`
    GitHubUsername   string                 `json:"github_username"`
    GitHubToken      string                 `json:"github_token"`
    CreatedAt        time.Time              `json:"created_at"`
    LastAccessedAt   time.Time              `json:"last_accessed_at"`
    Metadata         map[string]interface{} `json:"metadata"`
}

func NewRedisStore(addr string, ttl time.Duration) (*RedisStore, error) {
    client := redis.NewClient(&redis.Options{
        Addr:         addr,
        Password:     "",
        DB:           0,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        PoolSize:     10,
    })
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("redis ping failed: %w", err)
    }
    
    return &RedisStore{client: client, ttl: ttl}, nil
}

func (s *RedisStore) Create(ctx context.Context, session *Session) (string, error) {
    session.CreatedAt = time.Now()
    session.LastAccessedAt = time.Now()
    
    data, err := json.Marshal(session)
    if err != nil {
        return "", fmt.Errorf("marshal session: %w", err)
    }
    
    key := fmt.Sprintf("session:%s", session.SessionID)
    if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
        return "", fmt.Errorf("redis set: %w", err)
    }
    
    return session.SessionID, nil
}

func (s *RedisStore) Get(ctx context.Context, sessionID string) (*Session, error) {
    key := fmt.Sprintf("session:%s", sessionID)
    data, err := s.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, nil // Session not found
    }
    if err != nil {
        return nil, fmt.Errorf("redis get: %w", err)
    }
    
    var session Session
    if err := json.Unmarshal(data, &session); err != nil {
        return nil, fmt.Errorf("unmarshal session: %w", err)
    }
    
    // Update last accessed time
    session.LastAccessedAt = time.Now()
    if err := s.Update(ctx, &session); err != nil {
        // Log but don't fail - session still valid
    }
    
    return &session, nil
}

func (s *RedisStore) Update(ctx context.Context, session *Session) error {
    data, err := json.Marshal(session)
    if err != nil {
        return fmt.Errorf("marshal session: %w", err)
    }
    
    key := fmt.Sprintf("session:%s", session.SessionID)
    if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
        return fmt.Errorf("redis set: %w", err)
    }
    
    return nil
}

func (s *RedisStore) Delete(ctx context.Context, sessionID string) error {
    key := fmt.Sprintf("session:%s", sessionID)
    if err := s.client.Del(ctx, key).Err(); err != nil {
        return fmt.Errorf("redis del: %w", err)
    }
    
    return nil
}

func (s *RedisStore) Close() error {
    return s.client.Close()
}
```

#### Phase 1.1.3: Update Portal OAuth Handler (3 hours)

```go
// apps/portal/handlers/auth_handler.go
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
    code := c.Query("code")
    if code == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code parameter"})
        return
    }
    
    // Exchange code for access token
    accessToken, err := h.githubClient.ExchangeCode(c.Request.Context(), code)
    if err != nil {
        h.logger.Error("Failed to exchange code", "error", err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to authenticate"})
        return
    }
    
    // Get user info from GitHub
    githubUser, err := h.githubClient.GetUser(c.Request.Context(), accessToken)
    if err != nil {
        h.logger.Error("Failed to get user info", "error", err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get user info"})
        return
    }
    
    // Find or create user in database
    user, err := h.userRepo.FindOrCreateByGitHubID(c.Request.Context(), githubUser.ID, githubUser.Login)
    if err != nil {
        h.logger.Error("Failed to create user", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }
    
    // Create Redis session (NEW)
    sessionID := uuid.New().String()
    session := &session.Session{
        SessionID:      sessionID,
        UserID:         user.ID,
        GitHubUsername: githubUser.Login,
        GitHubToken:    accessToken,
        Metadata:       map[string]interface{}{
            "login_ip": c.ClientIP(),
            "user_agent": c.Request.UserAgent(),
        },
    }
    
    if _, err := h.sessionStore.Create(c.Request.Context(), session); err != nil {
        h.logger.Error("Failed to create session", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
        return
    }
    
    // Create JWT containing session_id (NOT user data)
    claims := jwt.MapClaims{
        "session_id": sessionID,
        "exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
        "iat":        time.Now().Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(h.jwtSecret))
    if err != nil {
        h.logger.Error("Failed to generate JWT", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }
    
    // Set cookie
    c.SetCookie(
        "devsmith_token",
        tokenString,
        int((7 * 24 * time.Hour).Seconds()),
        "/",
        "",
        false, // Set true in production with HTTPS
        true,  // HttpOnly
    )
    
    c.Redirect(http.StatusFound, "/dashboard")
}
```

#### Phase 1.1.4: Update Service Middlewares (2 hours per service)

```go
// internal/review/middleware/redis_session_auth.go
package middleware

import (
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "devsmith/internal/session"
)

func RedisSessionAuthMiddleware(sessionStore *session.RedisStore, jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Get JWT from cookie
        tokenString, err := c.Cookie("devsmith_token")
        if err != nil {
            // No portal session - check for service-specific auth (standalone mode)
            c.Redirect(http.StatusFound, "/auth/github/login")
            c.Abort()
            return
        }
        
        // 2. Parse JWT to get session_id
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(jwtSecret), nil
        })
        if err != nil || !token.Valid {
            c.Redirect(http.StatusFound, "/auth/github/login")
            c.Abort()
            return
        }
        
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.Redirect(http.StatusFound, "/auth/github/login")
            c.Abort()
            return
        }
        
        sessionID, ok := claims["session_id"].(string)
        if !ok {
            c.Redirect(http.StatusFound, "/auth/github/login")
            c.Abort()
            return
        }
        
        // 3. Validate session in Redis
        session, err := sessionStore.Get(c.Request.Context(), sessionID)
        if err != nil || session == nil {
            // Session expired or invalid
            c.Redirect(http.StatusFound, "/auth/github/login")
            c.Abort()
            return
        }
        
        // 4. Session valid - set user context
        c.Set("user_id", session.UserID)
        c.Set("github_username", session.GitHubUsername)
        c.Set("github_token", session.GitHubToken)
        c.Set("auth_mode", "portal")
        
        c.Next()
    }
}
```

#### Testing Checklist

- [ ] Redis container starts healthy
- [ ] Portal login creates session in Redis: `redis-cli KEYS "session:*"`
- [ ] JWT contains only session_id (inspect with jwt.io)
- [ ] Review app recognizes Portal session (no re-auth)
- [ ] Logs app recognizes Portal session (no re-auth)
- [ ] Logout deletes session from Redis
- [ ] Session expires after 7 days (test with TTL override)
- [ ] Direct service access (http://localhost:8081/review) initiates OAuth

**Acceptance Criteria:**
✅ User logs into Portal once → can access Review, Logs, Analytics without re-auth  
✅ Session stored in Redis with 7-day TTL  
✅ Logout from Portal clears session across all services  

---

### 1.2 Traefik Gateway Migration (1 day)

**Business Value:**
- ✅ **Service auto-discovery** - Add new service with Docker labels, no config edits
- ✅ **Auto-reload** - Traefik detects docker-compose changes automatically
- ✅ **Built-in dashboard** - http://localhost:8080 shows all routes and health
- ✅ **Health checks** - Automatic failover if service is down
- ✅ **Simpler config** - No nginx .conf files to maintain

**Technical Implementation:**

#### Phase 1.2.1: Add Traefik to docker-compose.yml (30 min)

```yaml
# docker-compose.yml
services:
  traefik:
    image: traefik:v2.10
    container_name: devsmith-traefik
    command:
      - "--api.dashboard=true"
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:3000"
      - "--log.level=INFO"
      - "--accesslog=true"
    ports:
      - "3000:3000"   # Main gateway
      - "8080:8080"   # Traefik dashboard
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    networks:
      - devsmith-network
    restart: unless-stopped

  portal:
    # ... existing portal config ...
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.portal.rule=Host(`localhost`) && PathPrefix(`/`)"
      - "traefik.http.routers.portal.entrypoints=web"
      - "traefik.http.services.portal.loadbalancer.server.port=8080"
      - "traefik.http.services.portal.loadbalancer.healthcheck.path=/health"
      - "traefik.http.services.portal.loadbalancer.healthcheck.interval=10s"

  review:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.review.rule=Host(`localhost`) && PathPrefix(`/review`)"
      - "traefik.http.routers.review.entrypoints=web"
      - "traefik.http.services.review.loadbalancer.server.port=8081"
      - "traefik.http.services.review.loadbalancer.healthcheck.path=/health"

  logs:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.logs.rule=Host(`localhost`) && PathPrefix(`/logs`)"
      - "traefik.http.routers.logs.entrypoints=web"
      - "traefik.http.services.logs.loadbalancer.server.port=8082"
      - "traefik.http.services.logs.loadbalancer.healthcheck.path=/health"

  analytics:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.analytics.rule=Host(`localhost`) && PathPrefix(`/analytics`)"
      - "traefik.http.routers.analytics.entrypoints=web"
      - "traefik.http.services.analytics.loadbalancer.server.port=8083"
      - "traefik.http.services.analytics.loadbalancer.healthcheck.path=/health"

# Remove nginx service entirely
```

#### Phase 1.2.2: Remove Nginx Config (15 min)

```bash
# Delete nginx configuration
rm -rf docker/nginx/

# Update .gitignore (if nginx was tracked)
echo "docker/nginx/" >> .gitignore
```

#### Testing Checklist

- [ ] Traefik dashboard accessible at http://localhost:8080
- [ ] Dashboard shows all 4 services (Portal, Review, Logs, Analytics)
- [ ] Portal accessible at http://localhost:3000
- [ ] Review accessible at http://localhost:3000/review
- [ ] Health checks show green in Traefik dashboard
- [ ] Stop Review service → Traefik shows red, traffic not routed
- [ ] Add new service → Appears in dashboard automatically

**Acceptance Criteria:**
✅ All services accessible through Traefik on port 3000  
✅ Dashboard shows service health in real-time  
✅ No manual config files required for routing  

---

## Priority 2: Testing & Validation (Week 2) - HIGH ROI

### 2.1 Comprehensive E2E Test Suite (3-4 days)

**Business Value:**
- ✅ **Prevent regressions** - Catch UI breaks before users see them
- ✅ **Visual validation** - Percy detects styling changes automatically
- ✅ **Documentation** - Tests serve as living documentation of workflows
- ✅ **Confidence** - Deploy knowing every workflow is tested

**Technical Implementation:**

#### Phase 2.1.1: Test Infrastructure Setup (1 day)

```bash
# Install dependencies
npm install --save-dev @playwright/test @percy/playwright

# Create test structure
mkdir -p tests/e2e/{fixtures,portal,review,logs,analytics,cross-service}

# Initialize Playwright config
npx playwright install
```

```typescript
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
  ],
  webServer: {
    command: 'docker-compose up',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
    timeout: 120000,
  },
});
```

#### Phase 2.1.2: Create Auth Fixture (2 hours)

```typescript
// tests/e2e/fixtures/auth.fixture.ts
import { test as base, expect } from '@playwright/test';

export const test = base.extend<{ authenticatedPage: Page }>({
  authenticatedPage: async ({ page }, use) => {
    // Navigate to portal
    await page.goto('/');
    
    // Click GitHub login
    await page.click('text=Login with GitHub');
    
    // Wait for OAuth redirect (mock in test environment)
    await page.waitForURL('**/auth/github/callback**');
    
    // Should redirect to dashboard
    await page.waitForURL('**/dashboard');
    
    // Verify authentication cookie exists
    const cookies = await page.context().cookies();
    const authCookie = cookies.find(c => c.name === 'devsmith_token');
    expect(authCookie).toBeDefined();
    
    await use(page);
  },
});
```

#### Phase 2.1.3: Complete Element Interaction Test (2 hours)

```typescript
// tests/e2e/logs/dashboard.visual.spec.ts
import { test, expect } from '../fixtures/auth.fixture';
import { percySnapshot } from '@percy/playwright';

test.describe('Logs Dashboard - Complete Interaction', () => {
  test('should interact with every element and validate visually', async ({ authenticatedPage: page }) => {
    // Navigate to Logs
    await page.click('text=Logs');
    await page.waitForURL('**/logs');
    await percySnapshot(page, 'Logs Dashboard - Initial Load');
    
    // Test 1: Stat cards (all 5 levels)
    const statCards = ['debug', 'info', 'warning', 'error', 'critical'];
    for (const level of statCards) {
      await page.click(`[data-stat-card="${level}"]`);
      await page.waitForSelector(`[data-filtered="${level}"]`);
      expect(await page.locator('.log-row').count()).toBeGreaterThan(0);
      await percySnapshot(page, `Logs Dashboard - ${level} Filter Active`);
    }
    
    // Test 2: Search functionality
    await page.fill('input[name="search"]', 'connection refused');
    await page.press('input[name="search"]', 'Enter');
    await page.waitForSelector('[data-search-active="true"]');
    await percySnapshot(page, 'Logs Dashboard - Search Active');
    
    // Test 3: Time range filters
    await page.selectOption('select[name="time-range"]', '1h');
    await page.waitForSelector('[data-time-range="1h"]');
    await percySnapshot(page, 'Logs Dashboard - 1 Hour Filter');
    
    // Test 4: Log row click (opens modal)
    await page.click('.log-row:first-child');
    await page.waitForSelector('.log-detail-modal');
    await percySnapshot(page, 'Logs Dashboard - Detail Modal Open');
    
    // Test 5: Modal tabs
    const tabs = ['details', 'context', 'ai-analysis'];
    for (const tab of tabs) {
      await page.click(`[data-tab="${tab}"]`);
      await page.waitForSelector(`[data-tab-content="${tab}"]`);
      await percySnapshot(page, `Logs Dashboard - Modal Tab ${tab}`);
    }
    
    // Test 6: Close modal
    await page.click('.modal-close');
    await page.waitForSelector('.log-detail-modal', { state: 'hidden' });
    
    // Test 7: Refresh button
    await page.click('[data-action="refresh"]');
    await page.waitForSelector('[data-refreshing="true"]');
    await page.waitForSelector('[data-refreshing="false"]');
    
    // Test 8: Dark mode toggle
    await page.click('[data-action="toggle-theme"]');
    await page.waitForSelector('body.dark-mode');
    await percySnapshot(page, 'Logs Dashboard - Dark Mode');
    
    // Test 9: Keyboard shortcuts modal
    await page.keyboard.press('?');
    await page.waitForSelector('.shortcuts-modal');
    await percySnapshot(page, 'Logs Dashboard - Shortcuts Modal');
    await page.keyboard.press('Escape');
    
    // Test 10: Export logs
    const downloadPromise = page.waitForEvent('download');
    await page.click('[data-action="export"]');
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toContain('.csv');
  });
});
```

#### Phase 2.1.4: Cross-Service Navigation Test (1 hour)

```typescript
// tests/e2e/cross-service/session-persistence.spec.ts
import { test, expect } from '../fixtures/auth.fixture';
import { percySnapshot } from '@percy/playwright';

test('should maintain session across Portal → Review → Logs', async ({ authenticatedPage: page }) => {
  // Start at Portal dashboard
  await percySnapshot(page, 'Session Test - Portal Dashboard');
  
  // Navigate to Review (should NOT re-authenticate)
  await page.click('text=Review');
  await page.waitForURL('**/review');
  
  // Verify no OAuth redirect happened
  expect(page.url()).not.toContain('github.com/login');
  await percySnapshot(page, 'Session Test - Review App');
  
  // Navigate to Logs (should NOT re-authenticate)
  await page.click('text=Logs');
  await page.waitForURL('**/logs');
  
  // Verify no OAuth redirect happened
  expect(page.url()).not.toContain('github.com/login');
  await percySnapshot(page, 'Session Test - Logs App');
  
  // Return to Portal
  await page.click('text=Dashboard');
  await page.waitForURL('**/dashboard');
  await percySnapshot(page, 'Session Test - Back to Portal');
  
  // Verify user still authenticated (no login button visible)
  await expect(page.locator('text=Login with GitHub')).not.toBeVisible();
  await expect(page.locator('[data-user-menu]')).toBeVisible();
});
```

#### Testing Checklist

- [ ] All Portal elements tested (login, logout, service cards)
- [ ] All Review elements tested (modes, file tree, analysis)
- [ ] All Logs elements tested (stat cards, search, filters, modals)
- [ ] All Analytics elements tested (charts, filters, export)
- [ ] Cross-service navigation tested (no re-auth required)
- [ ] Dark mode tested in all apps
- [ ] Percy screenshots captured for all views
- [ ] Visual regressions detected on CSS changes

**Acceptance Criteria:**
✅ 50+ test cases covering all apps and interactions  
✅ Percy visual regression for all major views  
✅ Tests run in CI/CD on every PR  

---

## Priority 3: User Experience (Week 2-3) - MEDIUM ROI

### 3.1 Styling Migration to devsmith-logs Design (1-2 days)

**Business Value:**
- ✅ **Brand consistency** - All apps look cohesive
- ✅ **Professional appearance** - Frosted glass, smooth animations
- ✅ **Dark mode** - Better for long coding sessions
- ✅ **Accessibility** - High contrast, proper focus states

**Technical Implementation:**

#### Phase 3.1.1: Create Shared Theme (4 hours)

**Key Clarification: Bootstrap Icons ONLY (No Bootstrap CSS/JS)**

```css
/* internal/ui/static/css/devsmith-theme.css */

/* Bootstrap Icons ONLY - 2.5KB, no JavaScript, no framework conflicts */
@import url('https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.css');

/* Tailwind directives */
@tailwind base;
@tailwind components;
@tailwind utilities;

/* Design System Variables */
:root {
  /* Dark Mode Colors (Default) */
  --bg-primary: #1a1d2e;
  --bg-secondary: #2d3148;
  --bg-card: rgba(45, 49, 72, 0.95);
  --text-primary: #e4e4e7;
  --text-secondary: #a1a1aa;
  
  /* Accent Colors */
  --accent-debug: #20c997;
  --accent-info: #0dcaf0;
  --accent-warning: #ffc107;
  --accent-error: #dc3545;
  --accent-critical: #d63384;
  
  /* Shadows */
  --shadow-sm: 0 2px 8px rgba(0, 0, 0, 0.08);
  --shadow-md: 0 4px 16px rgba(0, 0, 0, 0.12);
  --shadow-lg: 0 8px 24px rgba(0, 0, 0, 0.16);
}

/* Tailwind Component Classes */
@layer components {
  .ds-card {
    @apply bg-[var(--bg-card)] rounded-xl p-6 shadow-lg;
    @apply border border-gray-700/30;
    backdrop-filter: blur(10px);
    transition: all 200ms ease;
  }
  
  .ds-card:hover {
    @apply shadow-xl -translate-y-0.5;
  }
  
  .ds-stat-card {
    @apply bg-[var(--bg-secondary)] rounded-xl p-6 flex items-center gap-4;
    @apply cursor-pointer transition-all duration-200;
    @apply border-l-4 hover:shadow-xl hover:-translate-y-1;
  }
  
  .ds-stat-card.debug { @apply border-[var(--accent-debug)]; }
  .ds-stat-card.info { @apply border-[var(--accent-info)]; }
  .ds-stat-card.error { @apply border-[var(--accent-error)]; }
  
  .ds-stat-icon {
    @apply w-16 h-16 rounded-lg flex items-center justify-center text-3xl;
  }
  
  .ds-btn-primary {
    @apply px-4 py-2 bg-indigo-600 text-white rounded-lg;
    @apply hover:bg-indigo-700 hover:shadow-md transition-all;
  }
}

/* Light Mode (Optional) */
body.light-mode {
  --bg-primary: #ffffff;
  --bg-secondary: #f3f4f6;
  --bg-card: rgba(255, 255, 255, 0.95);
  --text-primary: #111827;
  --text-secondary: #4b5563;
}
```

**Icon Usage (Bootstrap Icons CSS classes only):**

```html
<!-- Stat card with icon (Tailwind + Bootstrap Icon) -->
<div class="ds-stat-card debug">
  <div class="ds-stat-icon bg-green-500/15 text-green-500">
    <i class="bi bi-bug-fill"></i> <!-- Bootstrap Icon -->
  </div>
  <div>
    <div class="text-3xl font-bold">1,234</div>
    <div class="text-sm text-gray-400">Debug Logs</div>
  </div>
</div>

<!-- Button with icon -->
<button class="ds-btn-primary flex items-center gap-2">
  <i class="bi bi-plus-circle"></i>
  <span>Add Entry</span>
</button>
```

**What We're NOT importing:**
- ❌ Bootstrap CSS (`bootstrap.min.css`) - Conflicts with Tailwind
- ❌ Bootstrap JavaScript - No JS needed
- ❌ jQuery - Not needed

**Stack:**
- Icons: Bootstrap Icons (CSS only, 2.5KB)
- Styling: Tailwind CSS
- JavaScript: HTMX + Alpine.js

#### Phase 3.1.2: Update All Service Templates (1 day)

```go
// apps/portal/templates/dashboard.templ
package templates

templ Dashboard(services []Service) {
  @Layout("Dashboard") {
    <div class="container mx-auto p-6">
      <h1 class="text-3xl font-bold mb-8">DevSmith Platform</h1>
      
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        for _, service := range services {
          <div class="ds-card" hx-get={"/"+service.Path} hx-push-url="true">
            <div class="flex items-center gap-4 mb-4">
              <i class={"bi " + service.Icon + " text-4xl text-indigo-500"}></i>
              <h2 class="text-2xl font-bold">{service.Name}</h2>
            </div>
            <p class="text-gray-400">{service.Description}</p>
          </div>
        }
      </div>
    </div>
  }
}
```

#### Testing Checklist

- [ ] Shared theme file created
- [ ] Portal templates updated with ds-* classes
- [ ] Review templates updated
- [ ] Logs templates updated
- [ ] Analytics templates updated
- [ ] Dark mode toggle works in all apps
- [ ] Frosted glass effect visible on cards
- [ ] Hover animations smooth (200ms transition)
- [ ] Visual consistency across all apps

**Acceptance Criteria:**
✅ All apps use shared devsmith-theme.css  
✅ Consistent look and feel across Portal, Review, Logs, Analytics  
✅ Dark mode toggle functional in all apps  

---

## Priority 4: Compliance (Week 3) - LOW EFFORT, HIGH ACCOUNTABILITY

### 4.1 Automated Compliance System (4-6 hours)

**Business Value:**
- ✅ **Quality assurance** - Blocks non-compliant commits automatically
- ✅ **Accountability** - Pre-commit hook enforces standards
- ✅ **Documentation** - VERIFICATION.md required for "complete" work
- ✅ **Testing** - Screenshots + regression tests mandatory

**Technical Implementation:**

#### Phase 4.1.1: Pre-Commit Hook (3 hours)

```bash
#!/bin/bash
# .git/hooks/pre-commit-copilot-compliance

set -e

echo "🔍 DevSmith Compliance Check..."

# Check 1: Commit message contains "complete" or "ready for review"
COMMIT_MSG=$(cat .git/COMMIT_EDITMSG 2>/dev/null || echo "")
if echo "$COMMIT_MSG" | grep -iq "complete\|ready for review"; then
    echo "⚠️  Detected 'complete' or 'ready for review' in commit message"
    echo "✅ Verifying compliance requirements..."
    
    # Check 2: VERIFICATION.md exists
    if [ ! -f "test-results/manual-verification-$(date +%Y%m%d)/VERIFICATION.md" ]; then
        echo "❌ FAILED: VERIFICATION.md not found"
        echo "📄 Create: test-results/manual-verification-$(date +%Y%m%d)/VERIFICATION.md"
        exit 1
    fi
    
    # Check 3: Screenshots exist
    SCREENSHOT_COUNT=$(find test-results/manual-verification-$(date +%Y%m%d)/ -name "*.png" | wc -l)
    if [ "$SCREENSHOT_COUNT" -lt 3 ]; then
        echo "❌ FAILED: Less than 3 screenshots found ($SCREENSHOT_COUNT)"
        echo "📸 Capture screenshots of: initial state, interaction, result"
        exit 1
    fi
    
    # Check 4: Regression tests passed
    if [ ! -f "test-results/regression-$(date +%Y%m%d)/summary.txt" ]; then
        echo "⚠️  Regression test results not found"
        echo "🧪 Run: bash scripts/regression-test.sh"
        exit 1
    fi
    
    if ! grep -q "Passed: .* ✓, Failed: 0" "test-results/regression-$(date +%Y%m%d)/summary.txt"; then
        echo "❌ FAILED: Regression tests have failures"
        cat test-results/regression-$(date +%Y%m%d)/summary.txt
        exit 1
    fi
    
    # Check 5: Test files exist for code changes
    MODIFIED_FILES=$(git diff --cached --name-only --diff-filter=AM | grep -E '\.(go|ts|tsx)$' || true)
    if [ -n "$MODIFIED_FILES" ]; then
        MISSING_TESTS=0
        while IFS= read -r file; do
            if [[ "$file" == *_test.* ]]; then
                continue  # Skip test files themselves
            fi
            
            TEST_FILE="${file%.*}_test.${file##*.}"
            if [ ! -f "$TEST_FILE" ]; then
                echo "⚠️  No test file for: $file"
                MISSING_TESTS=$((MISSING_TESTS + 1))
            fi
        done <<< "$MODIFIED_FILES"
        
        if [ $MISSING_TESTS -gt 0 ]; then
            echo "❌ FAILED: $MISSING_TESTS file(s) missing tests"
            exit 1
        fi
    fi
    
    echo "✅ All compliance checks passed!"
fi

# Standard pre-commit checks (always run)
echo "🔧 Running standard checks..."

# Format check
gofmt -l . | grep -v vendor
if [ $? -eq 0 ]; then
    echo "❌ FAILED: Go files not formatted"
    echo "🔧 Run: gofmt -w ."
    exit 1
fi

# Build check
go build ./... >/dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "❌ FAILED: Build errors"
    go build ./...
    exit 1
fi

echo "✅ Pre-commit checks complete!"
```

#### Phase 4.1.2: Install Hook (30 min)

```bash
#!/bin/bash
# scripts/install-compliance-hook.sh

cp .git/hooks/pre-commit-copilot-compliance .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

echo "✅ Compliance hook installed"
echo "📝 To bypass (emergency only): git commit --no-verify"
```

#### Phase 4.1.3: CI/CD Integration (1 hour)

```yaml
# .github/workflows/compliance-check.yml
name: Compliance Check

on:
  pull_request:
    branches: [development, main]

jobs:
  compliance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Check for VERIFICATION.md
        if: contains(github.event.pull_request.title, 'complete') || contains(github.event.pull_request.title, 'ready for review')
        run: |
          if [ ! -f "test-results/manual-verification-*/VERIFICATION.md" ]; then
            echo "❌ VERIFICATION.md required for completion"
            exit 1
          fi
      
      - name: Check for screenshots
        if: contains(github.event.pull_request.title, 'complete')
        run: |
          SCREENSHOT_COUNT=$(find test-results/manual-verification-*/ -name "*.png" 2>/dev/null | wc -l)
          if [ "$SCREENSHOT_COUNT" -lt 3 ]; then
            echo "❌ At least 3 screenshots required"
            exit 1
          fi
      
      - name: Run regression tests
        run: |
          docker-compose up -d
          bash scripts/regression-test.sh
```

#### Testing Checklist

- [ ] Hook blocks commits with "complete" but no VERIFICATION.md
- [ ] Hook blocks commits with less than 3 screenshots
- [ ] Hook blocks commits if regression tests failed
- [ ] Hook allows regular commits without compliance checks
- [ ] CI/CD rejects PRs missing compliance requirements

**Acceptance Criteria:**
✅ Pre-commit hook installed and functional  
✅ Cannot commit "complete" work without evidence  
✅ CI/CD enforces compliance on PRs  

---

## Handoff Section - Critical Information for Next Session

### Current State Summary (2025-11-05)

**What's Working:**
- ✅ Portal OAuth login functional (JWT_SECRET fix applied)
- ✅ All services running healthy
- ✅ Basic regression tests passing
- ✅ Logs app operational (Phase 1 complete)

**What's Broken:**
- ⚠️ Logs page formatting issues (CSS not loading properly)
- ⚠️ User must re-authenticate for each service (no shared sessions)
- ⚠️ Nginx gateway manual configuration required for new services

**What's Planned:**
- 🚧 Redis session store (Priority 1.1)
- 🚧 Traefik migration (Priority 1.2)
- 🚧 E2E test suite (Priority 2.1)
- 🚧 Styling migration (Priority 3.1)
- 🚧 Compliance system (Priority 4.1)

### Active Todo List

```markdown
# Priority Todo List (Start Here)

## Infrastructure (Week 1)
- [ ] Add Redis to docker-compose.yml
- [ ] Create internal/session/redis_store.go
- [ ] Update Portal OAuth handler to create Redis sessions
- [ ] Update Review middleware to check Redis sessions
- [ ] Update Logs middleware to check Redis sessions
- [ ] Update Analytics middleware to check Redis sessions
- [ ] Test: Login to Portal → Access Review without re-auth
- [ ] Add Traefik to docker-compose.yml
- [ ] Convert all services to use Traefik labels
- [ ] Remove nginx service and configs
- [ ] Test: Traefik dashboard shows all services
- [ ] Test: Health checks work in Traefik

## Testing (Week 2)
- [ ] Install Playwright + Percy
- [ ] Create auth fixture for authenticated tests
- [ ] Write complete element interaction test for Logs
- [ ] Write complete element interaction test for Review
- [ ] Write complete element interaction test for Portal
- [ ] Write cross-service navigation test
- [ ] Set up Percy visual regression
- [ ] Run tests in CI/CD

## User Experience (Week 2-3)
- [ ] Create internal/ui/static/css/devsmith-theme.css
- [ ] Update Portal templates with Tailwind + ds-* classes
- [ ] Update Review templates
- [ ] Update Logs templates (fix current formatting issues)
- [ ] Update Analytics templates
- [ ] Add dark mode toggle to all apps
- [ ] Test: Visual consistency across all apps

## Compliance (Week 3)
- [ ] Create pre-commit hook script
- [ ] Install hook: scripts/install-compliance-hook.sh
- [ ] Test: Hook blocks incomplete commits
- [ ] Add CI/CD compliance check workflow
- [ ] Test: CI/CD blocks non-compliant PRs

## Investigate & Fix
- [ ] Logs formatting issue (CSS not loading)
  - Check: apps/logs/static/css/*.css files exist
  - Check: Browser console for 404 errors on CSS files
  - Check: Nginx/Traefik serving /static/ correctly
```

### Key Files to Review

**Redis Session Implementation:**
- `internal/session/redis_store.go` (create this - code provided above)
- `apps/portal/handlers/auth_handler.go` (update OAuth callback)
- `internal/review/middleware/redis_session_auth.go` (create this)
- `docker-compose.yml` (add Redis service)

**Traefik Migration:**
- `docker-compose.yml` (add Traefik, convert service labels)
- `docker/nginx/` (delete this directory)

**Testing Infrastructure:**
- `tests/e2e/fixtures/auth.fixture.ts` (create this)
- `tests/e2e/logs/dashboard.visual.spec.ts` (create this)
- `playwright.config.ts` (create this)
- `package.json` (add Playwright + Percy dependencies)

**Styling System:**
- `internal/ui/static/css/devsmith-theme.css` (create this)
- `apps/*/templates/**/*.templ` (update all templates)

**Compliance:**
- `.git/hooks/pre-commit-copilot-compliance` (create this)
- `.github/workflows/compliance-check.yml` (create this)

### Known Issues & Debugging Steps

**Issue 1: Logs Formatting Broken**
```bash
# Check if CSS files exist
ls -la apps/logs/static/css/

# Check browser console
# Navigate to http://localhost:3000/logs
# Open DevTools → Console
# Look for 404 errors on CSS files

# Check Traefik/Nginx routing
curl -I http://localhost:3000/logs/static/css/dashboard.css

# Expected: 200 OK
# If 404: Update Traefik labels or nginx config
```

**Issue 2: Redis Connection Failed**
```bash
# Check Redis running
docker-compose ps redis

# Test Redis connectivity
docker-compose exec redis redis-cli ping
# Expected: PONG

# Check Redis logs
docker-compose logs redis --tail=50
```

**Issue 3: Traefik Not Showing Service**
```bash
# Check Traefik dashboard
open http://localhost:8080/dashboard/

# Check service labels in docker-compose.yml
docker-compose config | grep -A 10 "traefik.enable"

# Restart Traefik to reload config
docker-compose restart traefik
```

### Environment Variables Required

```bash
# .env (ensure these are set)
JWT_SECRET=dev-secret-key-change-in-production
REDIS_URL=redis:6379
OLLAMA_ENDPOINT=http://host.docker.internal:11434
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret

# Percy (for visual testing)
PERCY_TOKEN=your_percy_token_from_percy_io
```

### Testing Checklist Before Declaring Complete

Before saying work is "complete" or "ready for review":

- [ ] Regression tests run: `bash scripts/regression-test.sh`
- [ ] All tests PASS (100% pass rate, not 13/14)
- [ ] Manual testing completed with screenshots
- [ ] Screenshots saved to `test-results/manual-verification-$(date +%Y%m%d)/`
- [ ] VERIFICATION.md created with embedded screenshots
- [ ] Visual inspection: No loading spinners, no errors, UI matches expectations
- [ ] Cross-service navigation tested (Portal → Review → Logs)
- [ ] Docker services healthy: `docker-compose ps`
- [ ] ERROR_LOG.md updated if any errors encountered

**RULE ZERO: Do not say work is complete unless ALL boxes checked above.**

### Quick Start Commands for Next Session

```bash
# Start all services
docker-compose up -d

# Check service health
docker-compose ps

# View logs for debugging
docker-compose logs -f portal
docker-compose logs -f review
docker-compose logs -f logs

# Run regression tests
bash scripts/regression-test.sh

# Install compliance hook
bash scripts/install-compliance-hook.sh

# Run Playwright tests
npx playwright test

# Run Percy visual tests
npx percy exec -- npx playwright test
```

### Priority Sequence for Next Session

1. **Start with Infrastructure** (highest ROI, enables everything else)
   - Redis session store (2-3 days)
   - Traefik migration (1 day)
   
2. **Then Testing** (prevents regressions)
   - E2E test suite (3-4 days)
   
3. **Then UX** (improves user experience)
   - Styling migration (1-2 days)
   
4. **Finally Compliance** (enforcement)
   - Pre-commit hooks (4-6 hours)

**Total Estimate: 2-3 weeks for complete implementation**

### Success Metrics

**Week 1 Success:**
- ✅ User logs in once, accesses all apps (no re-auth)
- ✅ Traefik dashboard shows all services
- ✅ Services auto-discovered when added to docker-compose

**Week 2 Success:**
- ✅ 50+ Playwright tests passing
- ✅ Percy captures visual regressions
- ✅ CI/CD runs tests on every PR

**Week 3 Success:**
- ✅ All apps have consistent styling
- ✅ Pre-commit hook blocks incomplete work
- ✅ Dark mode works in all apps

### Final Notes

This document consolidates:
- ✅ IMPLEMENTATION_ROADMAP.md (5-phase plan)
- ✅ COMPREHENSIVE_PLATFORM_IMPROVEMENTS.md (OAuth, testing, styling, gateway, compliance)
- ✅ BOOTSTRAP_ICONS_CLARIFICATION.md (styling clarification)

All information needed to continue implementation is now in this single document. The old documents will be deleted after this file is created.

**Next steps:**
1. Read this document completely
2. Start with Priority 1.1 (Redis session store)
3. Follow the code examples provided
4. Test each phase thoroughly before moving to next
5. Update ERROR_LOG.md if any issues encountered
6. Do NOT say "complete" until all acceptance criteria met

Good luck! 🚀
