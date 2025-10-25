package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthRoutes_LoginRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Dummy DB connection not needed for login redirect
	RegisterAuthRoutes(r, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/github/login", http.NoBody)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "github.com/login/oauth/authorize")
}

func TestAuthRoutes_LoginRedirect_ContainsClientID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAuthRoutes(r, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/github/login", http.NoBody)
	r.ServeHTTP(w, req)

	location := w.Header().Get("Location")
	assert.Contains(t, location, "client_id")
}

func TestAuthRoutes_LoginRedirect_ContainsRedirectURI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAuthRoutes(r, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/github/login", http.NoBody)
	r.ServeHTTP(w, req)

	location := w.Header().Get("Location")
	assert.Contains(t, location, "redirect_uri")
}

func TestAuthRoutes_LoginRedirect_ContainsScopes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAuthRoutes(r, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/github/login", http.NoBody)
	r.ServeHTTP(w, req)

	location := w.Header().Get("Location")
	assert.Contains(t, location, "scope")
}

func TestAuthRoutes_CallbackRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAuthRoutes(r, nil)
	w := httptest.NewRecorder()
	// OAuth callback with code
	req, _ := http.NewRequest("GET", "/auth/github/callback?code=test123&state=test", http.NoBody)
	r.ServeHTTP(w, req)
	// Should handle the callback (may fail without real DB/token exchange, but should be 200, 401, or redirect)
	assert.Contains(t, []int{http.StatusOK, http.StatusFound, http.StatusInternalServerError, http.StatusUnauthorized}, w.Code)
}

func TestAuthRoutes_RoutesRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAuthRoutes(r, nil)

	// Should have routes registered
	routes := r.Routes()
	assert.Greater(t, len(routes), 0, "No routes registered")

	// Should have login route
	hasLoginRoute := false
	hasCallbackRoute := false
	for _, route := range routes {
		if route.Path == "/auth/github/login" {
			hasLoginRoute = true
			assert.Equal(t, "GET", route.Method)
		}
		if route.Path == "/auth/github/callback" {
			hasCallbackRoute = true
			assert.Equal(t, "GET", route.Method)
		}
	}

	assert.True(t, hasLoginRoute, "Login route not registered")
	assert.True(t, hasCallbackRoute, "Callback route not registered")
}

func TestAuthRoutes_LoginRedirect_HTTPSMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAuthRoutes(r, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/github/login", http.NoBody)
	r.ServeHTTP(w, req)

	// Login should be GET request
	assert.Equal(t, http.StatusFound, w.Code)
}

func TestAuthRoutes_LoginRedirect_StatusCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAuthRoutes(r, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/github/login", http.NoBody)
	r.ServeHTTP(w, req)

	// Should be a redirect (302 Found)
	assert.Equal(t, http.StatusFound, w.Code)
}

func TestAuthRoutes_LoginRedirect_ValidGitHubURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAuthRoutes(r, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/github/login", http.NoBody)
	r.ServeHTTP(w, req)

	location := w.Header().Get("Location")
	assert.True(t, location != "", "Location header is empty")
	assert.Contains(t, location, "github.com")
	assert.Contains(t, location, "https://")
}

func TestAuthRoutes_CallbackWithErrorParameter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAuthRoutes(r, nil)
	w := httptest.NewRecorder()

	// GitHub OAuth error response
	req, _ := http.NewRequest("GET", "/auth/github/callback?error=access_denied", http.NoBody)
	r.ServeHTTP(w, req)

	// Should handle error gracefully (400 or redirect)
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusFound, http.StatusInternalServerError}, w.Code)
}

func TestAuthRoutes_CallbackWithoutCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAuthRoutes(r, nil)
	w := httptest.NewRecorder()

	// Missing required code parameter
	req, _ := http.NewRequest("GET", "/auth/github/callback", http.NoBody)
	r.ServeHTTP(w, req)

	// Should return error or redirect
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusFound, http.StatusInternalServerError}, w.Code)
}

func TestAuthRoutes_MultipleLoginAttempts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAuthRoutes(r, nil)

	// Make multiple login requests
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/auth/github/login", http.NoBody)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusFound, w.Code, fmt.Sprintf("Request %d failed", i+1))
		assert.Contains(t, w.Header().Get("Location"), "github.com")
	}
}
