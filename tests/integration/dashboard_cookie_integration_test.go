package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/handlers"
	"github.com/stretchr/testify/assert"
)

type UserClaims struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	GithubID  string `json:"github_id"` // Added GithubID field
	jwt.RegisteredClaims
}

func TestDashboardWithOAuthCookies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/dashboard", handlers.DashboardHandler)

	// Generate a valid JWT for the devsmith_token cookie
	jwtKey := []byte("your-secret-key")
	claims := UserClaims{
		Username:  "testuser",
		Email:     "testuser@example.com",
		AvatarURL: "https://avatars.githubusercontent.com/u/123456?v=4",
		GithubID:  "123456", // Added GithubID field
		// Removed RegisteredClaims field for debugging
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Simulate request with cookies set as by OAuth callback
	req := httptest.NewRequest("GET", "/dashboard", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "devsmith_token", Value: signedToken})
	req.AddCookie(&http.Cookie{Name: "devsmith_user", Value: "testuser"})
	req.AddCookie(&http.Cookie{Name: "devsmith_avatar", Value: "https://avatars.githubusercontent.com/u/123456?v=4"})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Dashboard should return 200 OK")
	body := w.Body.String()
	assert.Contains(t, body, "testuser", "Dashboard should display username")
	assert.Contains(t, body, "https://avatars.githubusercontent.com/u/123456?v=4", "Dashboard should display avatar URL")
}
