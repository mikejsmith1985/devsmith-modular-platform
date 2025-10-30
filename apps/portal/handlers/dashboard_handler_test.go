package portal_handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	portal_handlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/handlers"
)

func TestDashboardHandler(t *testing.T) {
	// Set up Gin context and recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Initialize a valid HTTP request
	req := httptest.NewRequest(http.MethodGet, "/dashboard", http.NoBody)
	c.Request = req

	// Mock user claims using UserClaims structure
	c.Set("user", &portal_handlers.UserClaims{
		Username:  "testuser",
		Email:     "testuser@example.com",
		AvatarURL: "https://example.com/avatar.png",
	})

	// Call handler
	portal_handlers.DashboardHandler(c)

	// Assertions
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "testuser")
	assert.Contains(t, w.Body.String(), "testuser@example.com")
	assert.Contains(t, w.Body.String(), "https://example.com/avatar.png")
}

func TestGetUserInfoHandler(t *testing.T) {
	// Set up Gin context and recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Mock user claims using UserClaims structure
	c.Set("user", &portal_handlers.UserClaims{
		Username:  "testuser",
		Email:     "testuser@example.com",
		AvatarURL: "https://example.com/avatar.png",
		GithubID:  "",
		CreatedAt: time.Time{},
	})

	// Call handler
	portal_handlers.GetUserInfoHandler(c)

	// Assertions
	require.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"username":   "testuser",
		"email":      "testuser@example.com",
		"avatar_url": "https://example.com/avatar.png",
		"github_id":  "",
		"created_at": "0001-01-01T00:00:00Z",
	}, response)
}
