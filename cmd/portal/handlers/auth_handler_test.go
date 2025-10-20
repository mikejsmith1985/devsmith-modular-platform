package handlers

import (
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
