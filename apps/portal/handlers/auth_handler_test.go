package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLoginFlow_RedirectsToGitHub(t *testing.T) {
	// Arrange
	router := gin.Default()
	router.GET("/auth/github/login", func(c *gin.Context) {
		clientID := os.Getenv("GITHUB_CLIENT_ID")
		c.Redirect(http.StatusFound, "https://github.com/login/oauth/authorize?client_id="+clientID)
	})

	// Mock environment variable for GitHub client ID
	os.Setenv("GITHUB_CLIENT_ID", "test-client-id")
	defer os.Unsetenv("GITHUB_CLIENT_ID")

	// Act
	req, _ := http.NewRequest("GET", "/auth/github/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusFound, w.Code, "Should redirect to GitHub OAuth")
	location := w.Header().Get("Location")
	assert.Contains(t, location, "https://github.com/login/oauth/authorize", "Should redirect to GitHub OAuth URL")
	assert.Contains(t, location, "client_id=test-client-id", "Should include client ID in redirect URL")
}
