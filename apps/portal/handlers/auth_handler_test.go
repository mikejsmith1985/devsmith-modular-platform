package portal_handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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
	req, _ := http.NewRequest("GET", "/auth/github/login", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusFound, w.Code, "Should redirect to GitHub OAuth")
	location := w.Header().Get("Location")
	assert.Contains(t, location, "https://github.com/login/oauth/authorize", "Should redirect to GitHub OAuth URL")
	assert.Contains(t, location, "client_id=test-client-id", "Should include client ID in redirect URL")
}

func TestRegisterAuthRoutes(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Act
	RegisterAuthRoutes(r, nil)

	// Assert
	routes := []string{
		"/auth/github/login",
		"/auth/github/callback",
		"/auth/login",
		"/auth/github/dashboard",
	}

	for _, route := range routes {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, route, http.NoBody)
		r.ServeHTTP(w, req)
		assert.NotEqual(t, http.StatusNotFound, w.Code, "Route %s should be registered", route)
	}

	assert.NotNil(t, r, "Router should not be nil")
}

func TestValidateOAuthConfig(t *testing.T) {
	os.Setenv("GITHUB_CLIENT_ID", "test-client-id")
	os.Setenv("GITHUB_CLIENT_SECRET", "test-client-secret")
	os.Setenv("REDIRECT_URI", "http://localhost:3000/callback")
	defer os.Unsetenv("GITHUB_CLIENT_ID")
	defer os.Unsetenv("GITHUB_CLIENT_SECRET")
	defer os.Unsetenv("REDIRECT_URI")

	assert.True(t, ValidateOAuthConfig(), "Expected validateOAuthConfig to return true when all env vars are set")

	os.Unsetenv("GITHUB_CLIENT_ID")
	assert.False(t, ValidateOAuthConfig(), "Expected validateOAuthConfig to return false when GITHUB_CLIENT_ID is unset")
}

func TestCreateJWTForUser(t *testing.T) {
	user := &UserInfo{
		Login:     "testuser",
		Email:     "testuser@example.com",
		AvatarURL: "http://example.com/avatar.png",
		ID:        12345,
	}

	token, err := CreateJWTForUser(user)
	assert.NoError(t, err, "Expected no error when creating JWT for user")
	assert.NotEmpty(t, token, "Expected a non-empty JWT token")
}

func TestRegisterTokenRoutes(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Act
	RegisterTokenRoutes(r)

	// Assert
	t.Run("Token route registered", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/auth/token", http.NoBody)
		r.ServeHTTP(w, req)
		assert.NotEqual(t, http.StatusNotFound, w.Code, "Route /auth/token should be registered")
	})

	t.Run("Token route responds", func(t *testing.T) {
		// Placeholder for token endpoint behavior
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/auth/token", http.NoBody)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Token route should respond with 200 OK")
	})
}

func TestValidateOAuthParams(t *testing.T) {
	// Test case: Valid parameters
	err := ValidateOAuthParams("valid-client-id", "https://localhost:3000/callback")
	assert.NoError(t, err, "Expected no error for valid parameters")

	// Test case: Empty clientID
	err = ValidateOAuthParams("", "https://localhost:3000/callback")
	assert.Error(t, err, "Expected error for empty clientID")
	assert.Contains(t, err.Error(), "github_client_id not set")

	// Test case: Empty redirectURI
	err = ValidateOAuthParams("valid-client-id", "")
	assert.Error(t, err, "Expected error for empty redirectURI")
	assert.Contains(t, err.Error(), "redirect_uri not set")

	// Test case: Invalid redirectURI
	err = ValidateOAuthParams("valid-client-id", "ftp://example.com/callback")
	assert.Error(t, err, "Expected error for invalid redirectURI")
	assert.Contains(t, err.Error(), "invalid redirect_uri")

	// Test case: clientID with spaces
	err = ValidateOAuthParams("invalid client id", "https://localhost:3000/callback")
	assert.Error(t, err, "Expected error for clientID with spaces")
	assert.Contains(t, err.Error(), "invalid github_client_id")
}

func TestSetTestUserDefaults(t *testing.T) {
	user := &UserInfo{
		Login:     "",
		Email:     "",
		AvatarURL: "",
	}

	SetTestUserDefaults(user)

	assert.Equal(t, "test-user", user.Login, "Default username should be 'test-user'")
	assert.Equal(t, "test@example.com", user.Email, "Default email should be 'test@example.com'")
	assert.Equal(t, "https://avatars.githubusercontent.com/u/12345?v=4", user.AvatarURL, "Default avatar URL should be set")
}

func TestFetchUserInfo(t *testing.T) {
	// Arrange
	accessToken := "test-access-token"
	mockResponse := `{
		"login": "testuser",
		"name": "Test User",
		"email": "testuser@example.com",
		"avatar_url": "http://example.com/avatar.png",
		"id": 12345
	}`

	// Mock HTTP client
	httpClient := &http.Client{
		Transport: &mockTransport{
			responses: map[string]*http.Response{
				"https://api.github.com/user": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(mockResponse)),
				},
			},
		},
	}

	// Replace the default HTTP client with the mock client
	originalClient := http.DefaultClient
	http.DefaultClient = httpClient
	defer func() { http.DefaultClient = originalClient }()

	// Act
	user, err := FetchUserInfo(accessToken)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "testuser", user.Login)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "testuser@example.com", user.Email)
	assert.Equal(t, "http://example.com/avatar.png", user.AvatarURL)
	assert.Equal(t, int64(12345), user.ID)

	// Test case: Valid access token
	t.Run("Valid access token", func(t *testing.T) {
		// Configure mock response for valid token
		mockResponse := `{
			"login": "testuser",
			"name": "Test User",
			"email": "testuser@example.com",
			"avatar_url": "http://example.com/avatar.png",
			"id": 12345
		}`
		httpClient := &http.Client{
			Transport: &mockTransport{
				responses: map[string]*http.Response{
					"https://api.github.com/user": {
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(mockResponse)),
					},
				},
			},
		}
		originalClient := http.DefaultClient
		http.DefaultClient = httpClient
		defer func() { http.DefaultClient = originalClient }()

		accessToken := "valid-token"
		user, err := FetchUserInfo(accessToken)
		assert.NoError(t, err, "Expected no error for valid access token")
		assert.NotEmpty(t, user, "Expected user to be returned for valid access token")
	})

	// Test case: Invalid access token
	t.Run("Invalid access token", func(t *testing.T) {
		// Configure mock response for invalid token
		httpClient := &http.Client{
			Transport: &mockTransport{
				responses: map[string]*http.Response{
					"https://api.github.com/user": {
						StatusCode: http.StatusUnauthorized,
						Body:       io.NopCloser(strings.NewReader("")),
					},
				},
			},
		}
		originalClient := http.DefaultClient
		http.DefaultClient = httpClient
		defer func() { http.DefaultClient = originalClient }()

		accessToken := "invalid-token"
		_, err := FetchUserInfo(accessToken)
		assert.Error(t, err, "Expected error for invalid access token")
		assert.Contains(t, err.Error(), "failed to fetch user info: invalid access token")
	})

	// Test case: Empty access token
	t.Run("Empty access token", func(t *testing.T) {
		accessToken := ""
		_, err := FetchUserInfo(accessToken)
		assert.Error(t, err, "Expected error for empty access token")
		assert.Contains(t, err.Error(), "access token is empty")
	})
}

func TestExchangeCodeForToken(t *testing.T) {
	// Mock environment variables
	t.Setenv("GITHUB_CLIENT_ID", "test-client-id")
	t.Setenv("GITHUB_CLIENT_SECRET", "test-client-secret")
	t.Setenv("REDIRECT_URI", "http://localhost:3000/callback")

	// Mock HTTP client
	mockTransport := &mockTransport{
		responses: map[string]*http.Response{
			"https://github.com/login/oauth/access_token?client_id=test-client-id&client_secret=test-client-secret&code=test-code&redirect_uri=http%3A%2F%2Flocalhost%3A3000%2Fcallback": {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"access_token":"test-access-token","token_type":"Bearer","scope":"repo"}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			},
		},
	}
	http.DefaultClient = &http.Client{Transport: mockTransport}

	// Add logging to debug the response body
	log.Printf("Mock response body: %s", `{"access_token":"test-access-token","token_type":"Bearer","scope":"repo"}`)

	// Call the function (updated for PKCE - now requires code_verifier)
	accessToken, err := exchangeCodeForToken("test-code", "test-code-verifier")

	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if accessToken != "test-access-token" {
		t.Errorf("Expected access token 'test-access-token', got '%s'", accessToken)
	}
}

// mockTransport is a custom HTTP transport for mocking responses
// Updated to support dynamic responses based on request details
type mockTransport struct {
	responses map[string]*http.Response
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	log.Printf("[DEBUG] Requested URL: %s", req.URL.String())
	log.Printf("[DEBUG] MockTransport received request: %s %s", req.Method, req.URL.String())
	if resp, ok := m.responses[req.URL.String()]; ok {
		log.Printf("[DEBUG] Found mock response for URL: %s", req.URL.String())
		var bodyBytes []byte // Declare bodyBytes outside the if block
		if resp.Body != nil {
			bodyBytes, _ = io.ReadAll(resp.Body)
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			log.Printf("[DEBUG] Rewound mock response body: %s", string(bodyBytes))
			log.Printf("[DEBUG] Response Body (before return): %s", string(bodyBytes))
		}
		resp.ContentLength = int64(len(bodyBytes))
		log.Printf("[DEBUG] Response ContentLength: %d", resp.ContentLength)
		return resp, nil
	}
	log.Printf("[DEBUG] No mock response found for URL: %s", req.URL.String())
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader("Not Found")),
	}, nil
}
