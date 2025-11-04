// Package portal_handlers contains the HTTP handlers for the Portal service.
package portal_handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// UserClaims represents the JWT claims for authenticated users.
// It includes standard claims and additional user-specific fields.
type UserClaims struct {
	jwt.RegisteredClaims
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	GithubID  string    `json:"github_id"`
}

// RegisterAuthRoutes registers authentication-related routes
func RegisterAuthRoutes(router *gin.Engine, dbConn *sql.DB) {
	RegisterGitHubRoutes(router)
	RegisterTokenRoutes(router)
}

// RegisterGitHubRoutes registers GitHub-related authentication routes
func RegisterGitHubRoutes(router *gin.Engine) {
	log.Println("[DEBUG] Registering authentication routes")

	router.GET("/auth/github/login", HandleGitHubOAuthLogin)
	router.GET("/auth/github/callback", HandleGitHubOAuthCallback)
	router.GET("/auth/login", HandleAuthLogin)
	router.GET("/auth/github/dashboard", HandleGitHubDashboard)
}

// HandleGitHubOAuthLogin initiates GitHub OAuth flow
func HandleGitHubOAuthLogin(c *gin.Context) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	if clientID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub OAuth client ID not configured"})
		return
	}
	redirectURL := "https://github.com/login/oauth/authorize?client_id=" + clientID
	c.Redirect(http.StatusFound, redirectURL)
}

// HandleGitHubOAuthCallback processes GitHub OAuth callback
func HandleGitHubOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code in callback"})
		return
	}

	if !ValidateOAuthConfig() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub OAuth credentials not configured"})
		return
	}

	log.Printf("[DEBUG] Step 1: Received callback code=%s", code)
	log.Printf("[DEBUG] Step 2: Exchanging code for token...")

	accessToken, err := exchangeCodeForToken(code)
	if err != nil {
		log.Printf("[ERROR] Failed to exchange code for token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate"})
		return
	}

	log.Printf("[DEBUG] Step 3: Got access token, fetching user...")
	user, err := FetchUserInfo(accessToken)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}

	log.Printf("[DEBUG] Step 4: Creating JWT claims: %+v", user)
	tokenString, err := CreateJWTForUser(&UserInfo{
		Login:     user.Login,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		ID:        user.ID,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to sign JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue JWT"})
		return
	}

	log.Printf("[DEBUG] Step 5: Setting cookies and redirecting")
	SetSecureJWTCookie(c, tokenString)
	c.Redirect(http.StatusFound, "/dashboard")
}

// HandleAuthLogin handles the main login route
func HandleAuthLogin(c *gin.Context) {
	log.Println("[DEBUG] Login route registered")

	clientID := os.Getenv("GITHUB_CLIENT_ID")
	redirectURI := os.Getenv("REDIRECT_URI")

	if err := ValidateOAuthParams(clientID, redirectURI); err != nil {
		log.Printf("[ERROR] %v", err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	redirectURL := "https://github.com/login/oauth/authorize?client_id=" + clientID +
		"&redirect_uri=" + redirectURI + "&scope=read:user%20user:email"
	log.Printf("[DEBUG] Redirecting to GitHub OAuth: %s", redirectURL)
	c.Redirect(http.StatusFound, redirectURL)
}

// HandleGitHubDashboard is a placeholder dashboard route
func HandleGitHubDashboard(c *gin.Context) {
	log.Println("[DEBUG] Dashboard route registered")
	userID := c.Query("user_id")
	if userID == "" {
		log.Println("[ERROR] Missing user_id in dashboard request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_id"})
		return
	}
	log.Printf("[DEBUG] Dashboard accessed by user_id: %s", userID)
	c.JSON(http.StatusOK, gin.H{"message": "Dashboard route", "user_id": userID})
}

// ValidateOAuthConfig validates the OAuth configuration by checking environment variables.
func ValidateOAuthConfig() bool {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURI := os.Getenv("REDIRECT_URI")
	return clientID != "" && clientSecret != "" && redirectURI != ""
}

// UserInfo represents user information retrieved from GitHub OAuth.
type UserInfo struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	ID        int64  `json:"id"`
}

// CreateJWTForUser creates a JWT token for the given user information.
func CreateJWTForUser(user *UserInfo) (string, error) {
	jwtKey := []byte("your-secret-key")
	claims := UserClaims{
		Username:  user.Login,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		GithubID:  fmt.Sprintf("%d", user.ID),
		CreatedAt: time.Now(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ValidateOAuthParams validates OAuth parameters including client ID and redirect URI.
func ValidateOAuthParams(clientID, redirectURI string) error {
	if clientID == "" {
		return fmt.Errorf("github_client_id not set")
	}
	if redirectURI == "" {
		return fmt.Errorf("redirect_uri not set")
	}
	if !strings.HasPrefix(redirectURI, "https://") && !strings.HasPrefix(redirectURI, "http://localhost") {
		return fmt.Errorf("invalid redirect_uri: must start with https:// or http://localhost")
	}
	if strings.Contains(clientID, " ") {
		return fmt.Errorf("invalid github_client_id: contains spaces")
	}
	return nil
}

// SetTestUserDefaults sets default values for test user information.
func SetTestUserDefaults(user *UserInfo) {
	if user.Login == "" {
		user.Login = "test-user"
	}
	if user.Email == "" {
		user.Email = "test@example.com"
	}
	if user.AvatarURL == "" {
		user.AvatarURL = "https://avatars.githubusercontent.com/u/12345?v=4"
	}
}

// RegisterTokenRoutes registers token-related routes
func RegisterTokenRoutes(router *gin.Engine) {
	router.POST("/auth/token", func(c *gin.Context) {
		// Token generation logic
	})
}

// exchangeCodeForToken exchanges the authorization code for an access token
func exchangeCodeForToken(code string) (string, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURI := os.Getenv("REDIRECT_URI")

	// Exchange code for access token
	tokenReq, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", http.NoBody)
	if err != nil {
		return "", err
	}
	q := tokenReq.URL.Query()
	q.Add("client_id", clientID)
	q.Add("client_secret", clientSecret)
	q.Add("code", code)
	q.Add("redirect_uri", redirectURI)
	tokenReq.URL.RawQuery = q.Encode()
	tokenReq.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("[ERROR] Failed to close response body: %v", closeErr)
		}
	}()
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&tokenResp); decodeErr != nil {
		return "", decodeErr
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("access token is empty")
	}
	return tokenResp.AccessToken, nil
}

var githubAPI = "https://api.github.com/user"

// FetchUserInfo fetches the user info from GitHub using the provided access token.
func FetchUserInfo(accessToken string) (UserInfo, error) {
	if accessToken == "" {
		return UserInfo{}, errors.New("access token is empty")
	}

	// Fetch user info from GitHub
	userReq, err := http.NewRequest("GET", githubAPI, http.NoBody)
	if err != nil {
		log.Printf("[ERROR] Failed to create new request: %v", err)
		return UserInfo{}, err
	}
	userReq.Header.Set("Authorization", "token "+accessToken)
	log.Printf("[DEBUG] Using HTTP client: %v", http.DefaultClient)
	log.Printf("[DEBUG] Request URL: %s", userReq.URL)
	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		return UserInfo{}, err
	}
	defer func() {
		if closeErr := userResp.Body.Close(); closeErr != nil {
			log.Printf("[ERROR] Failed to close user response body: %v", closeErr)
		}
	}()

	if userResp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] GitHub API returned status: %d", userResp.StatusCode)
		return UserInfo{}, errors.New("failed to fetch user info: invalid access token")
	}

	if userResp.ContentLength == 0 {
		log.Printf("[ERROR] GitHub API returned an empty response body")
		return UserInfo{}, errors.New("failed to fetch user info: empty response body")
	}

	// Read and log the response body for debugging
	bodyBytes, err := io.ReadAll(userResp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read response body: %v", err)
		return UserInfo{}, errors.New("failed to read response body")
	}
	log.Printf("[DEBUG] Response Body: %s", string(bodyBytes))

	// Decode the response body directly into the UserInfo struct
	var user UserInfo
	if err := json.Unmarshal(bodyBytes, &user); err != nil {
		log.Printf("[ERROR] Failed to decode user response: %v", err)
		return UserInfo{}, errors.New("failed to decode user info")
	}

	log.Printf("[DEBUG] Decoded UserInfo: %+v", user)
	return user, nil
}

// Update the test configuration to use the nginx gateway URL for all tests
const nginxGatewayURL = "http://localhost:3000" // Ensure all tests route through nginx

// TestPortalLoginFlow tests the login flow through the nginx gateway.
func TestPortalLoginFlow(t *testing.T) {
	// Update the test to use nginxGatewayURL
	testServer := nginxGatewayURL + "/auth/github/login"
	t.Log("Testing login flow through nginx at:", testServer)
	// Ensure the Gin context (c) is properly passed to the function and used for the redirectURI query.
	c := &gin.Context{}
	req, err := http.NewRequest(http.MethodGet, testServer, http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create new HTTP request: %v", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	query := req.URL.Query()
	query.Set("redirect_uri", "http://localhost:3000/auth/github/callback")
	req.URL.RawQuery = query.Encode()
	redirectURI := c.Query("redirect_uri")
	log.Printf("DEBUG: Handling login request. Redirect URI: %s", redirectURI)
	// Add logic to use testServer in the test
	// ...existing code...
}

// FetchUserInfoHandler is a gin.HandlerFunc that wraps fetchUserInfo
func FetchUserInfoHandler(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing access token"})
		return
	}

	user, err := FetchUserInfo(accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// SetSecureJWTCookie sets JWT token in HTTP cookie with secure flags
// Parameters:
// - c: Gin context
// - tokenString: JWT token value
// Security flags:
// - HttpOnly: Prevents JavaScript XSS from stealing token
// - Secure: HTTPS-only transmission in production
// - SameSite=Strict: CSRF protection
// - 24-hour expiry
func SetSecureJWTCookie(c *gin.Context, tokenString string) {
	// In production (HTTPS), use Secure flag. In development/test, allow HTTP.
	isSecure := strings.HasPrefix(os.Getenv("REDIRECT_URI"), "https://")

	c.SetCookie(
		"devsmith_token", // name
		tokenString,      // value
		86400,            // maxAge (24 hours)
		"/",              // path
		"",               // domain
		isSecure,         // secure (HTTPS only in production)
		true,             // httpOnly (XSS protection)
	)

	// Set SameSite=Lax for CSRF protection while allowing top-level navigation
	// SameSite=Strict blocks cookies on link navigation (e.g., dashboard -> /review)
	// SameSite=Lax allows cookies on top-level GET requests (safe for navigation)
	c.SetSameSite(http.SameSiteLaxMode)
}
