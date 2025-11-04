//go:build integration

package integration

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	portal_handlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/handlers"
	portal_middleware "github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Load environment variables from .env.test using absolute path
	if err := godotenv.Load("/home/mikej/projects/DevSmith-Modular-Platform/.env.test"); err != nil {
		log.Fatalf("Error loading .env.test: %v", err)
	}

	// Run tests
	os.Exit(m.Run())
}

// Integration test for the portal login flow (no mocks, real HTTP)
func TestPortalLoginFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Enable test authentication endpoint
	os.Setenv("ENABLE_TEST_AUTH", "true")
	// Set test client ID for GitHub OAuth
	os.Setenv("GITHUB_CLIENT_ID", "test-client-id")
	os.Setenv("REDIRECT_URI", "https://localhost:3000/auth/callback")

	// Register auth routes (includes /auth/test-login)
	portal_handlers.RegisterAuthRoutes(router, nil)

	// Register dashboard route with JWT middleware (like in main.go)
	authenticated := router.Group("/")
	authenticated.Use(portal_middleware.JWTAuthMiddleware())
	authenticated.GET("/dashboard", portal_handlers.DashboardHandler)

	// Simulate user visiting /auth/login
	req := httptest.NewRequest("GET", "/auth/login", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should redirect to GitHub OAuth (real URL, not stub)
	assert.Equal(t, http.StatusFound, w.Code)
	location := w.Header().Get("Location")
	assert.Contains(t, location, "github.com/login/oauth/authorize", "Should redirect to GitHub OAuth")

	// Simulate dashboard access without authentication
	dashReq := httptest.NewRequest("GET", "/dashboard", http.NoBody)
	dashW := httptest.NewRecorder()
	router.ServeHTTP(dashW, dashReq)

	// Should redirect to login if not authenticated
	assert.Equal(t, http.StatusFound, dashW.Code)
	assert.Contains(t, dashW.Header().Get("Location"), "/login", "Should redirect to login if not authenticated")

	// Use /auth/test-login to generate a valid JWT token
	testLoginReq := httptest.NewRequest("POST", "/auth/test-login", strings.NewReader(`{
		"username": "testuser",
		"email": "testuser@example.com",
		"avatar_url": "https://avatars.githubusercontent.com/u/123456?v=4"
	}`))
	testLoginReq.Header.Set("Content-Type", "application/json")
	testLoginW := httptest.NewRecorder()
	router.ServeHTTP(testLoginW, testLoginReq)

	assert.Equal(t, http.StatusOK, testLoginW.Code, "Test login should succeed")
	var testLoginResp struct {
		Token string `json:"token"`
	}
	json.Unmarshal(testLoginW.Body.Bytes(), &testLoginResp)
	token := testLoginResp.Token

	// Create a NEW request for authenticated dashboard access
	authDashReq := httptest.NewRequest("GET", "/dashboard", http.NoBody)
	authDashReq.AddCookie(&http.Cookie{
		Name:  "devsmith_token",
		Value: token,
	})
	authDashW := httptest.NewRecorder()
	router.ServeHTTP(authDashW, authDashReq)

	// Should return 200 OK if authenticated
	assert.Equal(t, http.StatusOK, authDashW.Code, "Should return 200 OK when authenticated with valid token")
	assert.Contains(t, authDashW.Body.String(), "testuser", "Dashboard should display username")
}
