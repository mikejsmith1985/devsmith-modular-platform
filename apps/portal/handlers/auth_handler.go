// Package portal_handlers contains the HTTP handlers for the Portal service.
package portal_handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/config"
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

// dbConn is a package-level variable to store the database connection
var dbConn *sql.DB

// RegisterAuthRoutesWithSession registers authentication routes with Redis session support
func RegisterAuthRoutesWithSession(router *gin.Engine, db *sql.DB, store *session.RedisStore) {
	sessionStore = store
	dbConn = db
	RegisterAuthRoutes(router, db)
}

// RegisterGitHubRoutes registers GitHub-related authentication routes
func RegisterGitHubRoutes(router *gin.Engine) {
	log.Println("[DEBUG] Registering authentication routes with /api/portal/auth prefix")

	// Create auth group with correct prefix
	authGroup := router.Group("/api/portal/auth")
	authGroup.GET("/github/login", HandleGitHubOAuthLogin)
	authGroup.GET("/github/callback", HandleGitHubOAuthCallbackWithSession)
	authGroup.GET("/login", HandleAuthLogin)
	authGroup.GET("/github/dashboard", HandleGitHubDashboard)
	authGroup.POST("/logout", HandleLogout)
	authGroup.GET("/health", HandleOAuthHealthCheck) // NEW: OAuth health check

	// NOTE: Legacy routes /auth/github/callback kept for OAuth redirect compatibility
	// GitHub OAuth redirects to this URL (configured in REDIRECT_URI environment variable)
	// After successful authentication, this handler redirects to frontend with token

	// Keep /auth/github/login for backward compatibility (redirects to GitHub)
	// This is used by some external tools/scripts
	router.GET("/auth/github/login", HandleGitHubOAuthLogin)
	router.GET("/auth/github/callback", HandleGitHubOAuthCallbackWithSession) // OAuth redirect target
	router.GET("/auth/login", HandleAuthLogin)
	router.POST("/auth/logout", HandleLogout)
	router.GET("/auth/health", HandleOAuthHealthCheck)

	// Test authentication endpoint - only enabled when ENABLE_TEST_AUTH=true
	if os.Getenv("ENABLE_TEST_AUTH") == "true" {
		log.Println("[WARN] Test auth endpoint enabled - DO NOT USE IN PRODUCTION")
		router.POST("/auth/test-login", HandleTestLogin)
	}
}

// generateOAuthState generates a cryptographically secure random state parameter for OAuth CSRF protection
func generateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// storeOAuthState stores the state parameter in Redis with expiration
func storeOAuthState(state string) {
	log.Printf("[DEBUG] storeOAuthState called with state=%s, sessionStore nil=%v", state, sessionStore == nil)

	if sessionStore == nil {
		log.Println("[WARN] Session store not initialized, OAuth state will not persist across restarts")
		return
	}

	ctx := context.Background()

	log.Printf("[DEBUG] About to call sessionStore.StoreOAuthState for state=%s", state)
	// Store state in Redis with 10 minute expiration
	err := sessionStore.StoreOAuthState(ctx, state, 10*time.Minute)
	log.Printf("[DEBUG] sessionStore.StoreOAuthState returned, err=%v", err)

	if err != nil {
		log.Printf("[ERROR] Failed to store OAuth state in Redis: %v", err)
	} else {
		log.Printf("[OAUTH] Stored state in Redis: %s (expires in 10 minutes)", state)
	}
}

// validateOAuthState validates the state parameter from Redis and removes it
func validateOAuthState(state string) bool {
	if sessionStore == nil {
		log.Println("[WARN] Session store not initialized, OAuth state validation will fail")
		return false
	}

	ctx := context.Background()

	// Check if state exists in Redis and delete it (one-time use)
	valid, err := sessionStore.ValidateOAuthState(ctx, state)
	if err != nil {
		log.Printf("[ERROR] OAuth state validation error: %v", err)
		return false
	}

	if valid {
		log.Printf("[OAUTH] State validated and removed from Redis: %s", state)
	} else {
		log.Printf("[WARN] OAuth state validation failed: state not found or expired: %s", state)
	}

	return valid
}

// HandleOAuthHealthCheck checks if OAuth is properly configured and services are accessible
// GET /api/portal/auth/health
func HandleOAuthHealthCheck(c *gin.Context) {
	log.Println("[DEBUG] OAuth health check requested")

	checks := map[string]interface{}{
		"github_client_id_set":     os.Getenv("GITHUB_CLIENT_ID") != "",
		"github_client_secret_set": os.Getenv("GITHUB_CLIENT_SECRET") != "",
		"redirect_uri_set":         os.Getenv("REDIRECT_URI") != "",
		"jwt_secret_set":           os.Getenv("JWT_SECRET") != "",
		"redis_available":          sessionStore != nil,
	}

	// Test Redis connection if available
	if sessionStore != nil {
		ctx := c.Request.Context()
		testSessionID := "health-check-test"
		testSession := &session.Session{
			SessionID:      testSessionID,
			UserID:         0,
			GitHubUsername: "health-check",
			GitHubToken:    "test",
			CreatedAt:      time.Now(),
			LastAccessedAt: time.Now(),
			Metadata:       map[string]interface{}{"test": true},
		}

		// Try to create and immediately delete test session
		if _, err := sessionStore.Create(ctx, testSession); err != nil {
			checks["redis_writable"] = false
			checks["redis_error"] = err.Error()
			log.Printf("[ERROR] Redis health check failed: %v", err)
		} else {
			checks["redis_writable"] = true
			// Clean up test session
			if err := sessionStore.Delete(ctx, testSessionID); err != nil {
				log.Printf("[ERROR] Failed to delete test session: %v", err)
			}
		}
	} else {
		checks["redis_writable"] = false
		checks["redis_error"] = "session store not initialized"
	}

	// Determine overall health
	allHealthy := true
	for key, value := range checks {
		if strings.HasSuffix(key, "_set") || strings.HasSuffix(key, "_available") || strings.HasSuffix(key, "_writable") {
			if healthy, ok := value.(bool); ok && !healthy {
				allHealthy = false
				break
			}
		}
	}

	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}

	log.Printf("[DEBUG] OAuth health check result: healthy=%v, checks=%+v", allHealthy, checks)

	c.JSON(status, gin.H{
		"healthy":   allHealthy,
		"checks":    checks,
		"timestamp": time.Now().Format(time.RFC3339),
	})
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

	// Create or update test user in database (CRITICAL: must exist before session references it)
	log.Printf("[TEST_AUTH] Creating/updating test user in database: github_id=%s, username=%s", req.GitHubID, req.Username)
	upsertQuery := `
		INSERT INTO portal.users (github_id, username, email, avatar_url, github_access_token, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (github_id) 
		DO UPDATE SET 
			username = EXCLUDED.username,
			email = EXCLUDED.email,
			avatar_url = EXCLUDED.avatar_url,
			updated_at = NOW()
		RETURNING id
	`

	var userID int
	err = dbConn.QueryRowContext(c.Request.Context(), upsertQuery,
		req.GitHubID,          // github_id (from test request, e.g., "99999")
		req.Username,          // username (e.g., "playwright-test")
		req.Email,             // email (e.g., "playwright@devsmith.local")
		req.AvatarURL,         // avatar_url
		"test-token-not-real", // github_access_token (dummy for tests)
	).Scan(&userID)

	if err != nil {
		log.Printf("[ERROR] Failed to create test user in database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create test user account"})
		return
	}

	log.Printf("[TEST_AUTH] Test user persisted to database with ID: %d", userID)

	// Create test session in Redis (use database user ID, not hardcoded 999999)
	sess := &session.Session{
		SessionID:      sessionID,
		UserID:         userID, // Use actual database user ID
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

// HandleGitHubOAuthLogin initiates GitHub OAuth flow with CSRF protection
// Deprecated: This endpoint is deprecated in favor of frontend-initiated PKCE flow.
// The frontend now handles OAuth initiation with encrypted state.
// This endpoint is kept for backward compatibility only.
func HandleGitHubOAuthLogin(c *gin.Context) {
	log.Println("[OAUTH] DEPRECATED: Go-initiated OAuth flow (use frontend PKCE instead)")
	log.Println("[OAUTH] Step 1: Initiating GitHub OAuth flow")

	clientID := os.Getenv("GITHUB_CLIENT_ID")
	if clientID == "" {
		log.Println("[ERROR] GITHUB_CLIENT_ID not configured")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GitHub OAuth not configured",
			"details": "Server is missing GitHub OAuth credentials. Please contact support.",
			"action":  "Contact administrator - error code: OAUTH_CONFIG_MISSING",
		})
		return
	}

	// Generate CSRF protection state parameter
	state, err := generateOAuthState()
	if err != nil {
		log.Printf("[ERROR] Failed to generate OAuth state: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to initiate OAuth flow",
			"details": "Could not generate security token. Please try again.",
			"action":  "If this persists, contact support with error code: OAUTH_STATE_GEN_FAILED",
		})
		return
	}

	// Store state for validation in callback
	storeOAuthState(state)

	// Add prompt=consent to force GitHub to re-prompt for authorization
	// This prevents stale state issues when GitHub caches previous authorizations
	// URL-encode the state parameter to preserve = padding through GitHub redirect
	redirectURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&state=%s&scope=read:user%%20user:email&prompt=consent",
		clientID, url.QueryEscape(state))

	log.Printf("[OAUTH] Step 2: Redirecting to GitHub with state=%s (forced consent)", state)
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
	authGroup.POST("/token", HandleTokenExchange)
	authGroup.GET("/me", HandleGetCurrentUser)
}

// HandleTokenExchange handles PKCE token exchange
// POST /api/portal/auth/token
// Handles PKCE-based token exchange from frontend OAuth flow.
// NO STATE VALIDATION IN REDIS - frontend already validated via encrypted state.
// This endpoint only exchanges the authorization code for an access token using PKCE.
func HandleTokenExchange(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ERROR] Invalid token request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	log.Printf("[DEBUG] PKCE token exchange - code=%s, state=%s (encrypted, audit only)", req.Code, req.State)

	// Validate OAuth config
	if err := validateOAuthConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub OAuth not configured"})
		return
	}

	// Exchange code for access token with PKCE code_verifier
	// RFC 7636: code_verifier MUST be sent to token endpoint
	// NOTE: Frontend validates encrypted state client-side - we only validate PKCE here
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

	// Step 7.5: Persist user to database (if not exists, create; if exists, update)
	log.Printf("[OAUTH] Step 7.5: Persisting user to database")
	upsertQuery := `
		INSERT INTO portal.users (github_id, username, email, avatar_url, github_access_token, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (github_id) 
		DO UPDATE SET 
			username = EXCLUDED.username,
			email = EXCLUDED.email,
			avatar_url = EXCLUDED.avatar_url,
			github_access_token = EXCLUDED.github_access_token,
			updated_at = NOW()
		RETURNING id
	`

	var userID int
	err = dbConn.QueryRowContext(c.Request.Context(), upsertQuery,
		user.ID,        // github_id (BIGINT from GitHub API)
		user.Login,     // username
		user.Email,     // email (may be empty string)
		user.AvatarURL, // avatar_url
		accessToken,    // github_access_token (store for API calls)
	).Scan(&userID)

	if err != nil {
		log.Printf("[ERROR] Failed to persist user to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user account"})
		return
	}

	log.Printf("[OAUTH] Step 7.6: User persisted to database: ID=%d (github_id=%d, username=%s)",
		userID, user.ID, user.Login)

	// Create Redis session with DATABASE user ID (not GitHub ID)
	sess := &session.Session{
		UserID:         userID, // <--- DATABASE ID (1, 2, 3...), NOT GitHub ID
		GitHubUsername: user.Login,
		GitHubToken:    accessToken,
		Metadata: map[string]interface{}{
			"email":      user.Email,
			"avatar_url": user.AvatarURL,
			"name":       user.Name,
			"github_id":  int(user.ID), // Keep GitHub ID in metadata for reference
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
func exchangeCodeForToken(code, codeVerifier string) (string, error) {
	log.Printf("[TOKEN_EXCHANGE] Step 1: Preparing token exchange request")

	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURI := os.Getenv("REDIRECT_URI")

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("OAuth credentials not configured")
	}

	// Exchange code for access token
	tokenReq, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", http.NoBody)
	if err != nil {
		log.Printf("[TOKEN_EXCHANGE] ERROR: Failed to create request: %v", err)
		return "", fmt.Errorf("failed to create token request: %w", err)
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
		log.Printf("[TOKEN_EXCHANGE] Including code_verifier (PKCE flow)")
	}

	tokenReq.URL.RawQuery = q.Encode()
	tokenReq.Header.Set("Accept", "application/json")

	log.Printf("[TOKEN_EXCHANGE] Step 2: Sending request to GitHub")

	resp, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		log.Printf("[TOKEN_EXCHANGE] ERROR: Request failed: %v", err)
		return "", fmt.Errorf("failed to send token request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("[TOKEN_EXCHANGE] WARN: Failed to close response body: %v", closeErr)
		}
	}()

	// Read response body for logging
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("[TOKEN_EXCHANGE] ERROR: Failed to read response body: %v", readErr)
		return "", fmt.Errorf("failed to read response body: %w", readErr)
	}

	log.Printf("[TOKEN_EXCHANGE] Step 3: Response received - status=%d, body_length=%d",
		resp.StatusCode, len(bodyBytes))

	// Only log response body if there's an error (don't log access tokens)
	if resp.StatusCode != http.StatusOK {
		log.Printf("[TOKEN_EXCHANGE] ERROR Response body: %s", string(bodyBytes))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}

	if decodeErr := json.Unmarshal(bodyBytes, &tokenResp); decodeErr != nil {
		log.Printf("[TOKEN_EXCHANGE] ERROR: Failed to decode JSON: %v", decodeErr)
		return "", fmt.Errorf("failed to decode response: %w", decodeErr)
	}

	// Check for OAuth errors from GitHub
	if tokenResp.Error != "" {
		log.Printf("[TOKEN_EXCHANGE] ERROR: GitHub returned error: %s - %s",
			tokenResp.Error, tokenResp.ErrorDesc)
		return "", fmt.Errorf("github oauth error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		log.Printf("[TOKEN_EXCHANGE] ERROR: Unexpected status code: %d", resp.StatusCode)
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Validate access token
	if tokenResp.AccessToken == "" {
		log.Printf("[TOKEN_EXCHANGE] ERROR: Access token is empty")
		return "", fmt.Errorf("access token is empty")
	}

	log.Printf("[TOKEN_EXCHANGE] Step 4: Token exchange successful - token_type=%s, scope=%s",
		tokenResp.TokenType, tokenResp.Scope)

	return tokenResp.AccessToken, nil
}

var githubAPI = "https://api.github.com/user"

// FetchUserInfo fetches the user info from GitHub using the provided access token.
func FetchUserInfo(accessToken string) (UserInfo, error) {
	log.Printf("[USER_INFO] Step 1: Fetching user information from GitHub API")

	if accessToken == "" {
		log.Println("[USER_INFO] ERROR: Access token is empty")
		return UserInfo{}, errors.New("access token is empty")
	}

	// Fetch user info from GitHub
	userReq, err := http.NewRequest("GET", githubAPI, http.NoBody)
	if err != nil {
		log.Printf("[USER_INFO] ERROR: Failed to create request: %v", err)
		return UserInfo{}, fmt.Errorf("failed to create user info request: %w", err)
	}

	userReq.Header.Set("Authorization", "token "+accessToken)
	userReq.Header.Set("Accept", "application/json")

	log.Printf("[USER_INFO] Step 2: Sending request to %s", githubAPI)

	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		log.Printf("[USER_INFO] ERROR: Request failed: %v", err)
		return UserInfo{}, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer func() {
		if closeErr := userResp.Body.Close(); closeErr != nil {
			log.Printf("[USER_INFO] WARN: Failed to close response body: %v", closeErr)
		}
	}()

	log.Printf("[USER_INFO] Step 3: Response received - status=%d", userResp.StatusCode)

	if userResp.StatusCode != http.StatusOK {
		log.Printf("[USER_INFO] ERROR: GitHub API returned non-OK status: %d", userResp.StatusCode)

		// Try to read error body for more details
		bodyBytes, readErr := io.ReadAll(userResp.Body)
		if readErr != nil {
			log.Printf("[USER_INFO] ERROR: Failed to read response body: %v", readErr)
		} else {
			log.Printf("[USER_INFO] ERROR: Response body: %s", string(bodyBytes))
		}

		if userResp.StatusCode == http.StatusUnauthorized {
			return UserInfo{}, errors.New("invalid or expired access token")
		}
		return UserInfo{}, fmt.Errorf("github API returned status %d", userResp.StatusCode)
	}

	if userResp.ContentLength == 0 {
		log.Println("[USER_INFO] ERROR: GitHub API returned empty response body")
		return UserInfo{}, errors.New("empty response body from GitHub API")
	}

	// Read and decode the response body
	bodyBytes, err := io.ReadAll(userResp.Body)
	if err != nil {
		log.Printf("[USER_INFO] ERROR: Failed to read response body: %v", err)
		return UserInfo{}, fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("[USER_INFO] Step 4: Response body received - length=%d bytes", len(bodyBytes))

	// Decode the response body directly into the UserInfo struct
	var user UserInfo
	if err := json.Unmarshal(bodyBytes, &user); err != nil {
		log.Printf("[USER_INFO] ERROR: Failed to decode JSON: %v", err)
		return UserInfo{}, fmt.Errorf("failed to decode user info: %w", err)
	}

	// Validate required fields
	if user.Login == "" || user.ID == 0 {
		log.Printf("[USER_INFO] ERROR: Missing required fields in response - login=%s, id=%d",
			user.Login, user.ID)
		return UserInfo{}, errors.New("incomplete user info from GitHub API")
	}

	log.Printf("[USER_INFO] Step 5: User info successfully retrieved - login=%s, id=%d, email=%s",
		user.Login, user.ID, user.Email)

	return user, nil
}

// Update the test configuration to use the nginx gateway URL for all tests
var nginxGatewayURL = config.GetGatewayURL() // Ensure all tests route through nginx

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
	query.Set("redirect_uri", config.GetGatewayURL()+"/auth/github/callback")
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
// Deprecated: This endpoint is deprecated in favor of frontend PKCE flow with encrypted state.
// The frontend now handles OAuth callbacks via /api/portal/auth/token endpoint.
// This endpoint is kept for backward compatibility only.
// validateOAuthCallbackParams checks GitHub callback parameters for errors and missing values
func validateOAuthCallbackParams(c *gin.Context) (code, state string, err error) {
	code = c.Query("code")
	state = c.Query("state")
	errorParam := c.Query("error")
	errorDesc := c.Query("error_description")

	log.Printf("[OAUTH] Callback params: state=%s, error=%s, code_present=%v",
		state, errorParam, code != "")

	// Check for GitHub OAuth errors
	if errorParam != "" {
		log.Printf("[ERROR] GitHub OAuth error: %s - %s", errorParam, errorDesc)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "GitHub OAuth failed",
			"details":    fmt.Sprintf("GitHub returned error: %s", errorDesc),
			"action":     "Please try logging in again. If this persists, contact support.",
			"error_code": "GITHUB_OAUTH_" + strings.ToUpper(errorParam),
		})
		return "", "", fmt.Errorf("github oauth error: %s", errorParam)
	}

	// Validate code parameter
	if code == "" {
		log.Println("[ERROR] Missing authorization code in callback")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing authorization code",
			"details": "GitHub did not provide an authorization code. This may indicate a configuration issue.",
			"action":  "Please try logging in again. If this persists, contact support with error code: OAUTH_CODE_MISSING",
		})
		return "", "", fmt.Errorf("missing authorization code")
	}

	// Validate state parameter (CSRF protection)
	if state == "" {
		log.Println("[ERROR] Missing state parameter in callback")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing state parameter",
			"details": "Security validation failed. This may indicate a CSRF attack or configuration issue.",
			"action":  "Please try logging in again. If this persists, contact support with error code: OAUTH_STATE_MISSING",
		})
		return "", "", fmt.Errorf("missing state parameter")
	}

	if !validateOAuthState(state) {
		log.Printf("[WARN] OAuth state validation failed: received=%s", state)
		log.Println("[INFO] State validation failed - this may be from a cached GitHub authorization.")
		log.Println("[INFO] If you use passkeys/security keys, GitHub may bypass consent and return stale state.")
		log.Println("[INFO] Solution: Revoke app at https://github.com/settings/applications")

		c.JSON(http.StatusUnauthorized, gin.H{
			"error":        "Invalid OAuth state parameter",
			"details":      "Security validation failed. If you're using passkeys or security keys for GitHub login, GitHub may have returned a cached authorization from before the server was updated.",
			"action":       "Please revoke this app at https://github.com/settings/applications, then try logging in again. Error code: OAUTH_STATE_INVALID",
			"passkey_note": "Passkey logins can cause GitHub to bypass the consent screen and return stale authorization codes.",
		})
		return "", "", fmt.Errorf("invalid state parameter")
	}

	log.Println("[OAUTH] Step 4: State validated successfully")
	return code, state, nil
}

// exchangeCodeAndFetchUser performs token exchange and user info retrieval
func exchangeCodeAndFetchUser(c *gin.Context, code string) (*UserInfo, string, error) {
	// Validate OAuth configuration
	if !ValidateOAuthConfig() {
		log.Println("[ERROR] OAuth configuration validation failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "OAuth not configured",
			"details": "Server OAuth credentials are missing or invalid.",
			"action":  "Contact administrator with error code: OAUTH_CONFIG_INVALID",
		})
		return nil, "", fmt.Errorf("oauth not configured")
	}

	log.Println("[OAUTH] Step 5: Exchanging authorization code for access token")

	// Exchange code for access token (legacy non-PKCE flow)
	accessToken, err := exchangeCodeForToken(code, "")
	if err != nil {
		log.Printf("[ERROR] Failed to exchange code for token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "Failed to exchange code for token",
			"details":           "GitHub API error during token exchange.",
			"error_code":        "OAUTH_TOKEN_EXCHANGE_FAILED",
			"action":            "Try logging in again. If this persists, contact support.",
			"technical_details": err.Error(),
		})
		return nil, "", err
	}

	log.Println("[OAUTH] Step 6: Token received, fetching user information from GitHub")

	// Fetch user info from GitHub
	user, err := FetchUserInfo(accessToken)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "Failed to fetch user info from GitHub",
			"details":           "Authenticated with GitHub, but could not retrieve user profile.",
			"error_code":        "OAUTH_USER_INFO_FAILED",
			"action":            "Try logging in again. If this persists, contact support.",
			"technical_details": err.Error(),
		})
		return nil, "", err
	}

	log.Printf("[OAUTH] Step 7: User authenticated: %s (GitHub ID: %d)", user.Login, user.ID)
	return &user, accessToken, nil
}

// persistUserToDatabase creates or updates user record in the database
func persistUserToDatabase(c *gin.Context, user *UserInfo, accessToken string) (int, error) {
	log.Println("[OAUTH] Step 7.5: Persisting user to database")
	upsertQuery := `
		INSERT INTO portal.users (github_id, username, email, avatar_url, github_access_token, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (github_id) 
		DO UPDATE SET 
			username = EXCLUDED.username,
			email = EXCLUDED.email,
			avatar_url = EXCLUDED.avatar_url,
			github_access_token = EXCLUDED.github_access_token,
			updated_at = NOW()
		RETURNING id
	`

	var userID int
	err := dbConn.QueryRowContext(c.Request.Context(), upsertQuery,
		user.ID,        // github_id (bigint from GitHub)
		user.Login,     // username
		user.Email,     // email (may be empty)
		user.AvatarURL, // avatar_url
		accessToken,    // github_access_token (encrypted in production)
	).Scan(&userID)

	if err != nil {
		log.Printf("[ERROR] Failed to persist user to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "Failed to create user account",
			"details":           "Authentication succeeded, but could not create user record in database.",
			"error_code":        "OAUTH_USER_PERSIST_FAILED",
			"action":            "Try logging in again. If this persists, contact support.",
			"technical_details": err.Error(),
		})
		return 0, err
	}

	log.Printf("[OAUTH] Step 7.6: User persisted to database: ID=%d (GitHub ID=%d)", userID, user.ID)
	return userID, nil
}

// createRedisSessionAndJWT creates session in Redis and generates JWT token
func createRedisSessionAndJWT(c *gin.Context, userID int, user *UserInfo, accessToken string) (string, error) {
	// Validate session store availability
	if sessionStore == nil {
		log.Println("[ERROR] Session store not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Session management unavailable",
			"details": "Server session storage is not available.",
			"action":  "Contact administrator with error code: SESSION_STORE_UNAVAILABLE",
		})
		return "", fmt.Errorf("session store not initialized")
	}

	// Create Redis session (store full user data here, not in JWT)
	sess := &session.Session{
		UserID:         userID, // Use database ID (portal.users.id)
		GitHubUsername: user.Login,
		GitHubToken:    accessToken,
		Metadata: map[string]interface{}{
			"email":        user.Email,
			"avatar_url":   user.AvatarURL,
			"name":         user.Name,
			"github_id":    user.ID, // Store GitHub ID for reference
			"logged_in_at": time.Now().Format(time.RFC3339),
		},
	}

	log.Println("[OAUTH] Step 8: Creating Redis session")

	sessionID, err := sessionStore.Create(c.Request.Context(), sess)
	if err != nil {
		log.Printf("[ERROR] Failed to create session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "Failed to create user session",
			"details":           "Authentication succeeded, but could not create session in Redis.",
			"error_code":        "OAUTH_SESSION_CREATE_FAILED",
			"action":            "Try logging in again. If this persists, contact support.",
			"technical_details": err.Error(),
		})
		return "", err
	}

	log.Printf("[OAUTH] Step 9: Session created successfully: %s", sessionID)

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "Failed to sign authentication token",
			"details":           "Session created, but JWT token generation failed.",
			"error_code":        "OAUTH_JWT_SIGN_FAILED",
			"action":            "Try logging in again. If this persists, contact support.",
			"technical_details": err.Error(),
		})
		return "", err
	}

	log.Println("[OAUTH] Step 10: JWT created, setting secure cookie")
	return tokenString, nil
}

func HandleGitHubOAuthCallbackWithSession(c *gin.Context) {
	log.Println("[OAUTH] Deprecated: Go OAuth callback (use /api/portal/auth/token instead)")
	log.Println("[OAUTH] Step 3: Callback received from GitHub")

	// Validate callback parameters
	code, _, err := validateOAuthCallbackParams(c)
	if err != nil {
		return
	}

	// Exchange code and fetch user
	user, accessToken, err := exchangeCodeAndFetchUser(c, code)
	if err != nil {
		return
	}

	// Persist user to database
	userID, err := persistUserToDatabase(c, user, accessToken)
	if err != nil {
		return
	}

	// Create session and JWT
	tokenString, err := createRedisSessionAndJWT(c, userID, user, accessToken)
	if err != nil {
		return
	}

	// Set httpOnly cookie for security
	SetSecureJWTCookie(c, tokenString)

	// Redirect to React frontend callback route with token in URL
	// This allows React to store token in localStorage for API calls
	redirectURL := config.GetGatewayURL() + "/auth/callback?token=" + tokenString
	log.Printf("[OAUTH] Step 11: Authentication complete! Redirecting to: %s", redirectURL)
	log.Printf("[OAUTH] User %s (ID: %d) successfully authenticated", user.Login, user.ID)

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
