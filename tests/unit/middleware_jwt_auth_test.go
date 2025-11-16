package integration

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	portal_handlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/handlers"
	portal_middleware "github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/middleware"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Set JWT_SECRET for unit tests
	os.Setenv("JWT_SECRET", "test-secret-key-for-unit-tests")
}

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	// Set up Gin context
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(portal_middleware.JWTAuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Generate a valid JWT token
	// Define claims using portal_handlers.UserClaims
	claims := portal_handlers.UserClaims{
		Username:  "testuser",
		Email:     "testuser@example.com",
		AvatarURL: "https://avatars.githubusercontent.com/u/123456?v=4",
		GithubID:  "123456",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Log the token's header and payload for debugging
	log.Printf("[DEBUG] Token header: %+v", token.Header)
	log.Printf("[DEBUG] Token payload: %+v", claims)

	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Log the generated token for debugging
	log.Printf("[DEBUG] Generated token: %s", signedToken)

	// Create a request with the token in the Authorization header
	req := httptest.NewRequest(http.MethodGet, "/protected", http.NoBody)
	// Remove the Bearer prefix from the Authorization header
	req.Header.Set("Authorization", signedToken)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Log the response for debugging
	log.Printf("[DEBUG] Response code: %d", w.Code)
	log.Printf("[DEBUG] Response body: %s", w.Body.String())

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}
