package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	portal_handlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/handlers"
	"github.com/stretchr/testify/assert"
)

func TestHandleTestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/auth/test-login", portal_handlers.HandleTestLogin)

	t.Run("Invalid JSON request", func(t *testing.T) {
		body := `{"username": "testuser", "email": "test@example.com"` // Missing closing brace
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/auth/test-login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		portal_handlers.HandleTestLogin(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid JSON format")
	})

	t.Run("Missing fields in JSON", func(t *testing.T) {
		body := `{"username": "testuser"}` // Missing email and avatar_url
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/auth/test-login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		portal_handlers.HandleTestLogin(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Missing required fields in request body")
	})

	t.Run("Successful test login", func(t *testing.T) {
		requestBody := `{
			"username": "testuser",
			"email": "testuser@example.com",
			"avatar_url": "https://example.com/avatar.png"
		}`
		req, _ := http.NewRequest(http.MethodPost, "/auth/test-login", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "testuser")
		assert.Contains(t, w.Body.String(), "testuser@example.com")
		assert.Contains(t, w.Body.String(), "https://example.com/avatar.png")
	})
}
