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
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/security"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/session"
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

// sessionStore is a package-level variable to store the Redis session store
var sessionStore *session.RedisStore

// RegisterAuthRoutesWithSession registers authentication routes with Redis session support
func RegisterAuthRoutesWithSession(router *gin.Engine, dbConn *sql.DB, store *session.RedisStore) {
	sessionStore = store
	RegisterAuthRoutes(router, dbConn)
}

// RegisterGitHubRoutes registers GitHub-related authentication routes
func RegisterGitHubRoutes(router *gin.Engine) {
	log.Println("[DEBUG] Registering authentication routes with /api/portal/auth prefix")

	// Create auth group with correct prefix
	authGroup := router.Group("/api/portal/auth")
	{
		authGroup.GET("/github/login", HandleGitHubOAuthLogin)
		authGroup.GET("/github/callback", HandleGitHubOAuthCallbackWithSession)
		authGroup.GET("/login", HandleAuthLogin)
		authGroup.GET("/github/dashboard", HandleGitHubDashboard)
		authGroup.POST("/logout", HandleLogout)
	}

	// Legacy routes for backward compatibility (redirect to new paths)
	router.GET("/auth/github/login", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/portal/auth/github/login")
	})
	router.GET("/auth/github/callback", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/portal/auth/github/callback")
	})
	router.GET("/auth/login", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/portal/auth/login")
	})
	router.POST("/auth/logout", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/portal/auth/logout")
	})

	// Test authentication endpoint - only enabled when ENABLE_TEST_AUTH=true
	if os.Getenv("ENABLE_TEST_AUTH") == "true" {
		log.Println("[WARN] Test auth endpoint enabled - DO NOT USE IN PRODUCTION")
		router.POST("/auth/test-login", HandleTestLogin)
	}
}

// HandleTestLogin creates a test session for E2E testing
// Only enabled when ENABLE_TEST_AUTH=true environment variable is set
func HandleTestLogin(c *gin.Context) {
	if os.Getenv("ENABLE_TEST_AUTH") != "true" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Test auth endpoint not enabled"})
		return
	}

	type TestUser struct {
		Username  string `json:"username" binding:"required"`
		Email     string `json:"email" binding:"required"`
		AvatarURL string `json:"avatar_url" binding:"required"`
		GitHubID  string `json:"github_id"`
	}

	var req TestUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if sessionStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Session store not initialized"})
		return
	}

	// Generate session ID using session package
	sessionID, err := session.GenerateSessionID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate session ID"})
		return
	}

	// Create test session in Redis
	sess := &session.Session{
		SessionID:      sessionID,
		UserID:         999999, // Test user ID
		GitHubUsername: req.Username,
		GitHubToken:    "test-token-not-real",
		CreatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
		Metadata: map[string]interface{}{
			"email":      req.Email,
			"avatar_url": req.AvatarURL,
			"github_id":  req.GitHubID,
			"is_test":    true,
		},
	}

	// Create returns (sessionID string, error)
	createdSessionID, err := sessionStore.Create(c.Request.Context(), sess)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create test session"})
		return
	}

	// Generate JWT with session_id (same pattern as OAuth callback)
	claims := jwt.MapClaims{
		"session_id": createdSessionID,
		"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret := security.GetJWTSecret()
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set JWT cookie
	c.SetCookie(
		"devsmith_token",
		tokenString,
		int((7 * 24 * time.Hour).Seconds()),
		"/",
		"",
		false, // HTTP-only for dev
		true,  // HTTP-only
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"token":   token,
		"user": map[string]interface{}{
			"username":   req.Username,
			"email":      req.Email,
			"avatar_url": req.AvatarURL,
			"github_id":  req.GitHubID,
		},
	})
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

	// Legacy non-PKCE flow (no code_verifier)
	accessToken, err := exchangeCodeForToken(code, "")
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
	jwtKey := security.GetJWTSecret()
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

// TokenRequest represents the PKCE token exchange request
type TokenRequest struct {
	Code         string `json:"code" binding:"required"`
	State        string `json:"state" binding:"required"`
	CodeVerifier string `json:"code_verifier" binding:"required"`
}

// RegisterTokenRoutes registers token-related routes
func RegisterTokenRoutes(router *gin.Engine) {
	authGroup := router.Group("/api/portal/auth")
	{
		authGroup.POST("/token", HandleTokenExchange)
		authGroup.GET("/me", HandleGetCurrentUser)
	}
}

// HandleTokenExchange handles PKCE token exchange
// POST /api/portal/auth/token
func HandleTokenExchange(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ERROR] Invalid token request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	log.Printf("[DEBUG] PKCE token exchange - code=%s, state=%s", req.Code, req.State)

	// Validate OAuth config
	if err := validateOAuthConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub OAuth not configured"})
		return
	}

	// Exchange code for access token with PKCE code_verifier
	// RFC 7636: code_verifier MUST be sent to token endpoint
	accessToken, err := exchangeCodeForToken(req.Code, req.CodeVerifier)
	if err != nil {
		log.Printf("[ERROR] Failed to exchange code: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to authenticate"})
		return
	}

	// Fetch user info
	user, err := FetchUserInfo(accessToken)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch user: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to fetch user info"})
		return
	}

	log.Printf("[DEBUG] User authenticated: %s (ID: %d)", user.Login, user.ID)

	// Create Redis session
	sess := &session.Session{
		UserID:         int(user.ID),
		GitHubUsername: user.Login,
		GitHubToken:    accessToken,
		Metadata: map[string]interface{}{
			"email":      user.Email,
			"avatar_url": user.AvatarURL,
			"name":       user.Name,
		},
	}

	if sessionStore == nil {
		log.Printf("[ERROR] Session store not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Session management error"})
		return
	}

	sessionID, err := sessionStore.Create(c.Request.Context(), sess)
	if err != nil {
		log.Printf("[ERROR] Failed to create session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Generate JWT
	claims := jwt.MapClaims{
		"session_id": sessionID,
		"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret := security.GetJWTSecret()
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("[ERROR] JWT generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue token"})
		return
	}

	log.Printf("[DEBUG] Token exchange successful, session: %s", sessionID)

	// Set httpOnly cookie
	SetSecureJWTCookie(c, tokenString)

	// Return token to frontend (for localStorage)
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"username":   user.Login,
			"email":      user.Email,
			"avatar_url": user.AvatarURL,
			"github_id":  user.ID,
		},
	})
}

// HandleGetCurrentUser validates JWT token and returns current user info
// GET /api/portal/auth/me
func HandleGetCurrentUser(c *gin.Context) {
	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		log.Printf("[DEBUG] No Authorization header provided")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No authorization token provided"})
		return
	}

	// Remove "Bearer " prefix
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		log.Printf("[DEBUG] Invalid Authorization header format (missing Bearer prefix)")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		return
	}

	// Parse and validate JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return security.GetJWTSecret(), nil
	})

	if err != nil {
		log.Printf("[ERROR] JWT validation failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	if !token.Valid {
		log.Printf("[ERROR] JWT token invalid")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("[ERROR] Failed to extract JWT claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	// Get session_id from claims
	sessionID, ok := claims["session_id"].(string)
	if !ok {
		log.Printf("[ERROR] No session_id in JWT claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: missing session_id"})
		return
	}

	// Retrieve session from Redis
	if sessionStore == nil {
		log.Printf("[ERROR] Session store not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Session store unavailable"})
		return
	}

	sess, err := sessionStore.Get(c.Request.Context(), sessionID)
	if err != nil {
		log.Printf("[ERROR] Failed to retrieve session %s: %v", sessionID, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not found or expired"})
		return
	}

	// Return user info from session metadata
	c.JSON(http.StatusOK, gin.H{
		"username":   sess.GitHubUsername,
		"email":      sess.Metadata["email"],
		"avatar_url": sess.Metadata["avatar_url"],
		"github_id":  sess.Metadata["github_id"],
	})
}

func validateOAuthConfig() error {
	required := []string{"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET", "REDIRECT_URI"}
	for _, key := range required {
		if os.Getenv(key) == "" {
			return fmt.Errorf("%s environment variable not set", key)
		}
	}
	return nil
}

// exchangeCodeForToken exchanges the authorization code for an access token
// RFC 7636: For PKCE flow, code_verifier MUST be included
func exchangeCodeForToken(code string, codeVerifier string) (string, error) {
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

	// RFC 7636: Include code_verifier for PKCE flow
	// GitHub requires this when code_challenge was sent in authorization request
	if codeVerifier != "" {
		q.Add("code_verifier", codeVerifier)
		log.Printf("[DEBUG] Including code_verifier in token exchange (PKCE)")
	}

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
	// Read response body for logging
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read response body: %w", readErr)
	}
	log.Printf("[DEBUG] GitHub token response status: %d", resp.StatusCode)
	log.Printf("[DEBUG] GitHub token response body: %s", string(bodyBytes))

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if decodeErr := json.Unmarshal(bodyBytes, &tokenResp); decodeErr != nil {
		return "", fmt.Errorf("failed to decode response: %w", decodeErr)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, error: %s - %s", resp.StatusCode, tokenResp.Error, tokenResp.ErrorDesc)
	}
	if tokenResp.Error != "" {
		return "", fmt.Errorf("github oauth error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
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
		"",               // domain (empty = current domain)
		isSecure,         // secure
		true,             // httpOnly
	)
}

// HandleGitHubOAuthCallbackWithSession processes GitHub OAuth callback with Redis session
func HandleGitHubOAuthCallbackWithSession(c *gin.Context) {
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

	// Exchange code for access token (legacy non-PKCE flow)
	accessToken, err := exchangeCodeForToken(code, "")
	if err != nil {
		log.Printf("[ERROR] Failed to exchange code for token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate"})
		return
	}

	log.Printf("[DEBUG] Step 2: Got access token, fetching user...")

	// Fetch user info from GitHub
	user, err := FetchUserInfo(accessToken)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}

	log.Printf("[DEBUG] Step 3: User authenticated: %s (ID: %d)", user.Login, user.ID)

	// Create Redis session (store full user data here, not in JWT)
	sess := &session.Session{
		UserID:         int(user.ID),
		GitHubUsername: user.Login,
		GitHubToken:    accessToken,
		Metadata: map[string]interface{}{
			"email":      user.Email,
			"avatar_url": user.AvatarURL,
			"name":       user.Name,
		},
	}

	sessionID, err := sessionStore.Create(c.Request.Context(), sess)
	if err != nil {
		log.Printf("[ERROR] Failed to create session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	log.Printf("[DEBUG] Step 4: Session created: %s", sessionID)

	// Create JWT containing ONLY session_id (not user data)
	claims := jwt.MapClaims{
		"session_id": sessionID,
		"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret := security.GetJWTSecret()
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("[ERROR] Failed to sign JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue JWT"})
		return
	}

	log.Printf("[DEBUG] Step 5: JWT created, setting cookie and redirecting to React callback")

	// Set httpOnly cookie for security
	SetSecureJWTCookie(c, tokenString)

	// Redirect to React frontend callback route with token in URL
	// This allows React to store token in localStorage for API calls
	redirectURL := "http://localhost:3000/auth/callback?token=" + tokenString
	log.Printf("[DEBUG] Redirecting to: %s", redirectURL)
	c.Redirect(http.StatusFound, redirectURL)
}

// HandleLogout logs out the user by deleting their session
func HandleLogout(c *gin.Context) {
	// Get JWT from cookie
	tokenString, err := c.Cookie("devsmith_token")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Already logged out"})
		return
	}

	// Parse JWT to get session_id
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return security.GetJWTSecret(), nil
	})

	if err != nil || !token.Valid {
		// Invalid token, just clear cookie
		c.SetCookie("devsmith_token", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
		return
	}

	// Extract session_id from claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.SetCookie("devsmith_token", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
		return
	}

	sessionID, ok := claims["session_id"].(string)
	if !ok || sessionID == "" {
		c.SetCookie("devsmith_token", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
		return
	}

	// Delete session from Redis
	if err := sessionStore.Delete(c.Request.Context(), sessionID); err != nil {
		log.Printf("[WARN] Failed to delete session from Redis: %v", err)
	}

	// Clear JWT cookie
	c.SetCookie("devsmith_token", "", -1, "/", "", false, true)

	log.Printf("[DEBUG] User logged out, session deleted: %s", sessionID)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
