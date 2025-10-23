package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/handlers"
)

func TestDashboardHandler(t *testing.T) {
	// Set up Gin context and recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Initialize a valid HTTP request
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	c.Request = req

	// Mock user claims
	c.Set("user", jwt.MapClaims{
		"username":   "testuser",
		"email":      "testuser@example.com",
		"avatar_url": "https://example.com/avatar.png",
	})

	// Call handler
	handlers.DashboardHandler(c)

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

	// Mock user claims as jwt.MapClaims
	c.Set("user", jwt.MapClaims{
		"username":   "testuser",
		"email":      "testuser@example.com",
		"avatar_url": "https://example.com/avatar.png",
	})

	// Call handler
	handlers.GetUserInfoHandler(c)

	// Assertions
	require.Equal(t, http.StatusOK, w.Code)

	// Adjust expected response to include additional fields
	expectedResponse := map[string]interface{}{
		"username":   "testuser",
		"email":      "testuser@example.com",
		"avatar_url": "https://example.com/avatar.png",
		"created_at": nil,
		"github_id":  nil,
	}
	var actualResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse, actualResponse)
}
