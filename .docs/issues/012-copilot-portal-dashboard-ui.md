# Issue #012: [COPILOT] Portal Service - Dashboard UI

**Type:** Feature (Copilot Implementation)
**Service:** Portal
**Depends On:** Issue #003 (Portal Authentication)
**Estimated Duration:** 45-60 minutes

---

## Summary

Create the main Portal dashboard UI that serves as the home page after authentication. This dashboard provides navigation to all three services (Review, Logs, Analytics) and displays user information.

**User Story:**
> As a logged-in developer, I want to see a clean dashboard with links to all platform services, so I can easily navigate to the tool I need.

---

## Bounded Context

**Portal Service Context:**
- **Responsibility:** User authentication, navigation hub, session management
- **Does NOT:** Implement service features (those live in Review, Logs, Analytics)
- **Boundaries:** Portal provides authentication and navigation only

**Why This Matters:**
- Dashboard UI lives in Portal (navigation hub)
- Review/Logs/Analytics UIs live in their own services
- Portal doesn't know about code review logic, log parsing, etc.

---

## Success Criteria

### Must Have (MVP)
- [ ] Dashboard page loads at `/dashboard` after successful login
- [ ] Shows GitHub username and avatar (from JWT)
- [ ] Displays 3 service cards: Review, Logs, Analytics
- [ ] Each card shows service description and status
- [ ] Each card links to service (e.g., Review → `http://localhost:8081`)
- [ ] Logout button clears JWT and redirects to `/login`
- [ ] Responsive design (works on desktop and tablet)
- [ ] Protected route (redirects to `/login` if not authenticated)

### Nice to Have (Post-MVP)
- Recent activity feed
- Service health indicators (up/down)
- Quick actions (e.g., "Start Review")

---

## Database Schema

**No new tables needed.**

Uses existing `portal.users` table from Issue #003.

---

## API Endpoints

### GET `/api/v1/dashboard/user`
**Purpose:** Get current user info for dashboard display

**Request:**
```http
GET /api/v1/dashboard/user
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "username": "mikejsmith1985",
  "avatar_url": "https://avatars.githubusercontent.com/u/12345",
  "github_id": 12345,
  "email": "mike@example.com",
  "created_at": "2025-10-18T10:30:00Z"
}
```

**Errors:**
- `401 Unauthorized` - Invalid or missing JWT
- `500 Internal Server Error` - Database error

---

## File Structure

```
apps/portal/
├── templates/
│   ├── layout.templ              # Existing (from Issue #003)
│   ├── login.templ               # Existing (from Issue #003)
│   └── dashboard.templ           # NEW - Dashboard page
├── handlers/
│   ├── auth_handler.go           # Existing
│   └── dashboard_handler.go      # NEW - Dashboard handler
└── static/
    ├── css/
    │   └── dashboard.css         # NEW - Dashboard styles
    └── js/
        └── dashboard.js          # NEW - Dashboard interactions

cmd/portal/
└── main.go                       # UPDATE - Add dashboard route
```

---

## Implementation Details

### 1. Dashboard Templ Template

**File:** `apps/portal/templates/dashboard.templ`

```go
package templates

templ Dashboard(user DashboardUser) {
	@Layout("Dashboard") {
		<div class="dashboard-container">
			<header class="dashboard-header">
				<div class="user-info">
					<img src={user.AvatarURL} alt={user.Username} class="avatar"/>
					<div class="user-details">
						<h2>{user.Username}</h2>
						<p class="user-email">{user.Email}</p>
					</div>
				</div>
				<button id="logout-btn" class="btn-logout">Logout</button>
			</header>

			<main class="dashboard-main">
				<h1>DevSmith Platform</h1>
				<p class="subtitle">AI-assisted code review and development analytics</p>

				<div class="services-grid">
					@ServiceCard(ServiceInfo{
						Name: "Review",
						Description: "Code review with 5 reading modes",
						URL: "http://localhost:8081",
						Icon: "📖",
						Status: "ready",
					})

					@ServiceCard(ServiceInfo{
						Name: "Logs",
						Description: "Real-time development logs",
						URL: "http://localhost:8082",
						Icon: "📝",
						Status: "ready",
					})

					@ServiceCard(ServiceInfo{
						Name: "Analytics",
						Description: "Trends and insights",
						URL: "http://localhost:8083",
						Icon: "📊",
						Status: "ready",
					})
				</div>
			</main>
		</div>
	}
}

templ ServiceCard(service ServiceInfo) {
	<div class="service-card">
		<div class="service-icon">{service.Icon}</div>
		<h3>{service.Name}</h3>
		<p class="service-description">{service.Description}</p>
		<span class={"service-status " + service.Status}>{service.Status}</span>
		<a href={templ.SafeURL(service.URL)} class="btn-primary">Open {service.Name}</a>
	</div>
}

type DashboardUser struct {
	Username  string
	Email     string
	AvatarURL string
}

type ServiceInfo struct {
	Name        string
	Description string
	URL         string
	Icon        string
	Status      string
}
```

---

### 2. Dashboard Handler

**File:** `apps/portal/handlers/dashboard_handler.go`

```go
package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"devsmith/apps/portal/templates"
)

// DashboardHandler serves the main dashboard page
func DashboardHandler(c *gin.Context) {
	// Extract user from JWT (middleware already validated)
	claims, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userClaims := claims.(jwt.MapClaims)

	user := templates.DashboardUser{
		Username:  userClaims["username"].(string),
		Email:     userClaims["email"].(string),
		AvatarURL: userClaims["avatar_url"].(string),
	}

	// Render dashboard template
	component := templates.Dashboard(user)
	component.Render(c.Request.Context(), c.Writer)
}

// GetUserInfoHandler returns current user info as JSON
func GetUserInfoHandler(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userClaims := claims.(jwt.MapClaims)

	c.JSON(http.StatusOK, gin.H{
		"username":   userClaims["username"],
		"email":      userClaims["email"],
		"avatar_url": userClaims["avatar_url"],
		"github_id":  userClaims["github_id"],
		"created_at": userClaims["created_at"],
	})
}
```

---

### 3. Update Main Router

**File:** `cmd/portal/main.go`

```go
// Add to existing routes (after auth routes from Issue #003)

// Dashboard routes (protected by JWT middleware)
authenticated := r.Group("/")
authenticated.Use(middleware.JWTAuthMiddleware())
{
	authenticated.GET("/dashboard", handlers.DashboardHandler)
	authenticated.GET("/api/v1/dashboard/user", handlers.GetUserInfoHandler)
}

// API routes for user info
api := r.Group("/api/v1")
api.Use(middleware.JWTAuthMiddleware())
{
	api.GET("/dashboard/user", handlers.GetUserInfoHandler)
}
```

---

### 4. Dashboard CSS

**File:** `apps/portal/static/css/dashboard.css`

```css
.dashboard-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  background: #f8f9fa;
  border-radius: 8px;
  margin-bottom: 2rem;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  border: 2px solid #0366d6;
}

.user-details h2 {
  margin: 0;
  font-size: 1.25rem;
  color: #24292e;
}

.user-email {
  margin: 0;
  font-size: 0.875rem;
  color: #586069;
}

.btn-logout {
  padding: 0.5rem 1rem;
  background: #dc3545;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background 0.2s;
}

.btn-logout:hover {
  background: #c82333;
}

.dashboard-main h1 {
  font-size: 2rem;
  margin-bottom: 0.5rem;
  color: #24292e;
}

.subtitle {
  color: #586069;
  margin-bottom: 2rem;
}

.services-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
  margin-top: 2rem;
}

.service-card {
  background: white;
  border: 1px solid #e1e4e8;
  border-radius: 8px;
  padding: 1.5rem;
  text-align: center;
  transition: transform 0.2s, box-shadow 0.2s;
}

.service-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.service-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.service-card h3 {
  margin: 0 0 0.5rem 0;
  color: #24292e;
}

.service-description {
  color: #586069;
  margin-bottom: 1rem;
  font-size: 0.875rem;
}

.service-status {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  margin-bottom: 1rem;
}

.service-status.ready {
  background: #d4edda;
  color: #155724;
}

.service-status.development {
  background: #fff3cd;
  color: #856404;
}

.btn-primary {
  display: inline-block;
  padding: 0.5rem 1.5rem;
  background: #0366d6;
  color: white;
  text-decoration: none;
  border-radius: 6px;
  font-weight: 500;
  transition: background 0.2s;
}

.btn-primary:hover {
  background: #0256c7;
}

@media (max-width: 768px) {
  .services-grid {
    grid-template-columns: 1fr;
  }

  .dashboard-header {
    flex-direction: column;
    gap: 1rem;
  }
}
```

---

### 5. Dashboard JavaScript

**File:** `apps/portal/static/js/dashboard.js`

```javascript
// Logout functionality
document.getElementById('logout-btn').addEventListener('click', async () => {
  // Clear JWT from localStorage
  localStorage.removeItem('devsmith_token');

  // Redirect to login
  window.location.href = '/login';
});

// Optional: Check service health on load
async function checkServiceHealth() {
  const services = [
    { name: 'Review', url: 'http://localhost:8081/health' },
    { name: 'Logs', url: 'http://localhost:8082/health' },
    { name: 'Analytics', url: 'http://localhost:8083/health' },
  ];

  for (const service of services) {
    try {
      const response = await fetch(service.url);
      if (response.ok) {
        console.log(`${service.name} service is healthy`);
      }
    } catch (err) {
      console.warn(`${service.name} service is not responding`);
    }
  }
}

// Run health check on page load
checkServiceHealth();
```

---

## Testing Requirements

### Unit Tests

**File:** `apps/portal/handlers/dashboard_handler_test.go`

```go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestDashboardHandler_Authenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock JWT middleware
	router.Use(func(c *gin.Context) {
		c.Set("user", jwt.MapClaims{
			"username":   "testuser",
			"email":      "test@example.com",
			"avatar_url": "https://example.com/avatar.jpg",
			"github_id":  float64(12345),
		})
		c.Next()
	})

	router.GET("/dashboard", DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "testuser")
	assert.Contains(t, w.Body.String(), "DevSmith Platform")
}

func TestDashboardHandler_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/dashboard", DashboardHandler)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))
}

func TestGetUserInfoHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("user", jwt.MapClaims{
			"username":   "testuser",
			"email":      "test@example.com",
			"avatar_url": "https://example.com/avatar.jpg",
			"github_id":  float64(12345),
			"created_at": "2025-10-18T10:30:00Z",
		})
		c.Next()
	})

	router.GET("/api/v1/dashboard/user", GetUserInfoHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/user", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "testuser")
	assert.Contains(t, w.Body.String(), "test@example.com")
}
```

### Manual Testing Checklist

- [ ] Navigate to `http://localhost:8080/dashboard` after login
- [ ] Verify GitHub username and avatar display correctly
- [ ] Click each service card link (Review, Logs, Analytics)
- [ ] Verify service URLs open in new tab (or same tab with back button)
- [ ] Click Logout button
- [ ] Verify redirect to `/login` and JWT cleared
- [ ] Try accessing `/dashboard` without authentication (should redirect)
- [ ] Test responsive design on mobile viewport

---

## Configuration

**File:** `.env` (Portal service)

```bash
# Existing from Issue #003
GITHUB_CLIENT_ID=your_github_oauth_app_id
GITHUB_CLIENT_SECRET=your_github_oauth_app_secret
JWT_SECRET=your_jwt_secret_key_min_32_chars
DATABASE_URL=postgresql://portal_user:portal_pass@localhost:5432/devsmith_portal

# Service URLs (for dashboard links)
REVIEW_SERVICE_URL=http://localhost:8081
LOGS_SERVICE_URL=http://localhost:8082
ANALYTICS_SERVICE_URL=http://localhost:8083
```

---

## Acceptance Criteria

Before marking this issue complete, verify:

- [x] Dashboard page loads at `/dashboard`
- [x] User info displays (username, email, avatar)
- [x] All 3 service cards render with correct links
- [x] Logout button works (clears JWT, redirects to login)
- [x] Protected route works (unauthenticated users redirected)
- [x] Responsive design works on desktop and tablet
- [x] `/api/v1/dashboard/user` endpoint returns user info
- [x] Unit tests pass (70%+ coverage)
- [x] Manual testing checklist complete
- [x] No console errors
- [x] No hardcoded URLs (all from .env)

---

## Branch Naming

```bash
feature/012-portal-dashboard-ui
```

---

## Notes

- This dashboard is intentionally simple (navigation hub only)
- Service-specific UIs will be built in Issues #013, #014, #015
- Service health checking is optional for MVP (nice-to-have)
- Dashboard uses Templ for server-side rendering (consistent with stack)
- JWT already includes all user data needed (no extra DB queries)

---

**Created:** 2025-10-20
**For:** Copilot Implementation
**Estimated Time:** 45-60 minutes
