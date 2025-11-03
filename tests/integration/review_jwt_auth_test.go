package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	review_handlers "github.com/mikejsmith1985/devsmith-modular-platform/apps/review/handlers"
)

// RED PHASE: These tests should fail initially because Review service doesn't validate JWT yet

// TestReviewModeWithoutJWT tests that mode endpoints require authentication
func TestReviewModeWithoutJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Setup minimal handler (will be replaced with real setup)
	handler := &review_handlers.UIHandler{}
	router.POST("/api/review/modes/preview", handler.HandlePreviewMode)

	// Act: POST without JWT
	code := `package main\nfunc main() {}`
	req := httptest.NewRequest("POST", "/api/review/modes/preview",
		bytes.NewBufferString(fmt.Sprintf("pasted_code=%s&model=mistral:7b-instruct", code)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should require authentication")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "authentication", "Error should mention authentication")
}

// TestReviewModeWithValidJWT tests that mode endpoints accept valid JWT
func TestReviewModeWithValidJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Setup minimal handler (will be replaced with real setup)
	handler := &review_handlers.UIHandler{}
	router.POST("/api/review/modes/preview", handler.HandlePreviewMode)

	// Create valid JWT token matching Portal's format
	token := createTestJWT(t, "testuser", "test@example.com", "12345")

	// Act: POST with valid JWT in cookie
	code := `package main\nfunc main() {}`
	req := httptest.NewRequest("POST", "/api/review/modes/preview",
		bytes.NewBufferString(fmt.Sprintf("pasted_code=%s&model=mistral:7b-instruct", code)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "devsmith_token",
		Value: token,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Should return 200 OK (analysis succeeds)
	assert.Equal(t, http.StatusOK, w.Code, "Should accept valid JWT and process request")
}

// TestReviewModeWithInvalidJWT tests that mode endpoints reject invalid JWT
func TestReviewModeWithInvalidJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	handler := &review_handlers.UIHandler{}
	router.POST("/api/review/modes/preview", handler.HandlePreviewMode)

	// Act: POST with invalid JWT
	code := `package main\nfunc main() {}`
	req := httptest.NewRequest("POST", "/api/review/modes/preview",
		bytes.NewBufferString(fmt.Sprintf("pasted_code=%s&model=mistral:7b-instruct", code)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "devsmith_token",
		Value: "invalid.jwt.token",
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should reject invalid JWT")
}

// TestReviewSessionCreationWithUserID tests that sessions are tied to authenticated user
func TestReviewSessionCreationWithUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	handler := &review_handlers.UIHandler{}
	router.POST("/api/review/sessions", handler.CreateSessionHandler)

	// Create JWT for user with ID "12345"
	token := createTestJWT(t, "testuser", "test@example.com", "12345")

	// Act: Create session with JWT
	req := httptest.NewRequest("POST", "/api/review/sessions",
		bytes.NewBufferString(`{"title":"Test Session","pasted_code":"package main"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "devsmith_token",
		Value: token,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Session created successfully
	assert.Equal(t, http.StatusOK, w.Code, "Should create session with user context")

	// Parse response to verify user_id is populated
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Session should have user_id field matching JWT github_id
	sessionData := response["session"].(map[string]interface{})
	assert.Equal(t, "12345", sessionData["user_id"], "Session should be tied to authenticated user")
}

// Helper: createTestJWT creates a valid JWT token matching Portal's format
func createTestJWT(t *testing.T, username, email, githubID string) string {
	// UserClaims matching Portal's structure
	type UserClaims struct {
		jwt.RegisteredClaims
		CreatedAt time.Time `json:"created_at"`
		Username  string    `json:"username"`
		Email     string    `json:"email"`
		AvatarURL string    `json:"avatar_url"`
		GithubID  string    `json:"github_id"`
	}

	claims := UserClaims{
		Username:  username,
		Email:     email,
		GithubID:  githubID,
		AvatarURL: "https://example.com/avatar.png",
		CreatedAt: time.Now(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("your-secret-key")) // Must match Portal's key
	require.NoError(t, err)

	return tokenString
}
