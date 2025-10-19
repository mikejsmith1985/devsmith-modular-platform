# OpenHands Implementation Spec: Portal Service - Authentication

**Created:** 2025-10-18
**Issue:** #2
**Estimated Complexity:** Medium
**Target Service:** portal
**Estimated Time:** 1.5 - 2 hours (autonomous)

---

## Overview

### Feature Description
Implement the Portal service's authentication system using GitHub OAuth, enabling users to log in and access their GitHub repositories. This is the foundation service that all other apps depend on for user identity.

### User Story
As a developer learning to read code effectively, I want to log in with my GitHub account so that I can access my repositories for code review sessions and have my learning progress tracked.

### Success Criteria
- [ ] Users can click "Login with GitHub" and complete OAuth flow
- [ ] Successful authentication redirects to portal dashboard
- [ ] User session persists across page refreshes
- [ ] Logout functionality clears session
- [ ] Unauthorized users cannot access protected routes
- [ ] User's GitHub profile data is stored in database
- [ ] All tests pass with 70%+ coverage

---

## Context for Cognitive Load Management

### Bounded Context
**Service:** Portal
**Domain:** Authentication and User Management
**Related Entities:**
- `User` (authentication context) - GitHub identity, session management
- `Session` (active login state) - JWT tokens, expiry
- `GitHubToken` (OAuth credential) - Access token for GitHub API calls

**Context Boundaries:**
- ✅ **Within scope:** Login, logout, session validation, GitHub OAuth, user profile storage
- ❌ **Out of scope:** Code review functionality (Review service), repository analysis (Review service), user preferences for reviews (Review service)

**Why This Separation:**
The Portal service ONLY knows about "who is the user" and "are they authenticated." It does NOT know about code reviews, reading modes, or AI analysis. This prevents cognitive overload and allows each service to be understood independently.

---

### Layering

**Primary Layer:** All three layers required (Controller → Orchestration → Data)

#### Controller Layer Files
```
cmd/portal/handlers/
├── auth_handler.go              # GitHub OAuth HTTP endpoints
├── auth_handler_test.go         # Handler tests
├── dashboard_handler.go         # Portal dashboard page
└── dashboard_handler_test.go    # Dashboard tests

cmd/portal/templates/
├── layout.templ                 # Base HTML layout
├── login.templ                  # Login page with "Login with GitHub" button
├── dashboard.templ              # Portal dashboard (app browser)
└── components/
    ├── header.templ             # Header with user profile
    └── app_card.templ           # App enable/disable card
```

#### Orchestration Layer Files
```
internal/portal/services/
├── auth_service.go              # Authentication business logic
├── auth_service_test.go         # Service tests
├── github_client.go             # GitHub API integration
└── github_client_test.go        # GitHub client tests

internal/portal/interfaces/
└── auth_interface.go            # Abstract contracts for testing
```

#### Data Layer Files
```
internal/portal/db/
├── user_repository.go           # User CRUD operations
├── user_repository_test.go      # Repository tests
└── migrations/
    ├── 20251018_001_create_users_table.sql
    └── 20251018_002_create_sessions_table.sql
```

**Cross-Layer Rules:**
- ✅ `auth_handler.go` calls `auth_service.go`
- ✅ `auth_service.go` calls `user_repository.go`
- ❌ `auth_handler.go` MUST NOT call `user_repository.go` directly
- ❌ `user_repository.go` MUST NOT import service or handler packages
- ❌ No circular dependencies between layers

---

### Abstractions to Implement

**New Interfaces:**

```go
// internal/portal/interfaces/auth_interface.go
package interfaces

import (
    "context"
    "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
)

// AuthService defines authentication operations
// Why interface: Allows mocking in handler tests, enables future auth providers
type AuthService interface {
    // AuthenticateWithGitHub completes OAuth flow and creates user session
    // Returns user and JWT token on success
    AuthenticateWithGitHub(ctx context.Context, code string) (*models.User, string, error)

    // ValidateSession checks if JWT token is valid and returns user
    ValidateSession(ctx context.Context, token string) (*models.User, error)

    // RevokeSession invalidates the given JWT token
    RevokeSession(ctx context.Context, token string) error
}

// UserRepository defines database operations for users
// Why interface: Enables testing services without real database
type UserRepository interface {
    // CreateOrUpdate inserts new user or updates existing (by GitHub ID)
    CreateOrUpdate(ctx context.Context, user *models.User) error

    // FindByGitHubID retrieves user by their GitHub ID
    FindByGitHubID(ctx context.Context, githubID int64) (*models.User, error)

    // FindByID retrieves user by internal database ID
    FindByID(ctx context.Context, id int) (*models.User, error)
}

// GitHubClient defines operations for GitHub API integration
// Why interface: Enables testing without hitting real GitHub API
type GitHubClient interface {
    // ExchangeCodeForToken exchanges OAuth code for access token
    ExchangeCodeForToken(ctx context.Context, code string) (string, error)

    // GetUserProfile fetches authenticated user's GitHub profile
    GetUserProfile(ctx context.Context, accessToken string) (*models.GitHubProfile, error)
}
```

**Existing Interfaces to Use:**
- Standard library `context.Context` for cancellation and timeouts
- `http.Handler` for HTTP endpoints
- `*sql.DB` or `*pgxpool.Pool` for database connections

**Implementation Strategy:**
1. Define interfaces first (abstraction) in `internal/portal/interfaces/`
2. Create concrete implementations (concretion) in respective packages
3. Test against interfaces using mocks (testify/mock)
4. Dependency injection via constructors (no globals)

---

### Scope Management

**Global/Package-Level State:**
```go
// AVOID global variables. Document any exceptions:
// None required for this feature.
```

**Struct-Level State:**
```go
// Preferred: Encapsulate dependencies in structs

// internal/portal/services/auth_service.go
type AuthService struct {
    userRepo      interfaces.UserRepository
    githubClient  interfaces.GitHubClient
    jwtSecret     []byte
    tokenExpiry   time.Duration
    logger        *zerolog.Logger
}

// Constructor with explicit dependencies
func NewAuthService(
    userRepo interfaces.UserRepository,
    githubClient interfaces.GitHubClient,
    jwtSecret string,
    logger *zerolog.Logger,
) *AuthService {
    return &AuthService{
        userRepo:     userRepo,
        githubClient: githubClient,
        jwtSecret:    []byte(jwtSecret),
        tokenExpiry:  24 * time.Hour,
        logger:       logger,
    }
}
```

**Function-Level Scope:**
```go
// Keep variables as local as possible
// Pass context explicitly (enables timeout control)
// Return errors explicitly (Go idiom)

func (s *AuthService) AuthenticateWithGitHub(ctx context.Context, code string) (*models.User, string, error) {
    // All variables scoped to this function
    token, err := s.githubClient.ExchangeCodeForToken(ctx, code)
    if err != nil {
        return nil, "", fmt.Errorf("token exchange failed: %w", err)
    }

    profile, err := s.githubClient.GetUserProfile(ctx, token)
    if err != nil {
        return nil, "", fmt.Errorf("profile fetch failed: %w", err)
    }

    // ... rest of logic
}
```

---

## Implementation Details

### 1. Database Changes

#### Schema: `portal`

**New Tables:**

```sql
-- internal/portal/db/migrations/20251018_001_create_users_table.sql

CREATE TABLE portal.users (
    id SERIAL PRIMARY KEY,
    github_id BIGINT NOT NULL UNIQUE,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    avatar_url TEXT,
    github_access_token TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_users_github_id ON portal.users(github_id);
CREATE INDEX idx_users_username ON portal.users(username);

-- Comments for documentation
COMMENT ON TABLE portal.users IS 'Authenticated users from GitHub OAuth';
COMMENT ON COLUMN portal.users.github_id IS 'Unique GitHub user ID (immutable)';
COMMENT ON COLUMN portal.users.github_access_token IS 'Encrypted OAuth token for GitHub API';
```

```sql
-- internal/portal/db/migrations/20251018_002_create_sessions_table.sql

CREATE TABLE portal.sessions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES portal.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_sessions_user_id ON portal.sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON portal.sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON portal.sessions(expires_at);

-- Comments
COMMENT ON TABLE portal.sessions IS 'Active user sessions with JWT tokens';
COMMENT ON COLUMN portal.sessions.token_hash IS 'SHA-256 hash of JWT token';
COMMENT ON COLUMN portal.sessions.expires_at IS 'Expiration timestamp (default 24h)';
```

**Migrations:**
- Up migration: `20251018_001_create_users_table.sql`
- Down migration: `20251018_001_create_users_table_down.sql` (DROP TABLE portal.users CASCADE)
- Up migration: `20251018_002_create_sessions_table.sql`
- Down migration: `20251018_002_create_sessions_table_down.sql` (DROP TABLE portal.sessions CASCADE)

**Data Relationships:**
```
portal.users 1:N portal.sessions
  - users.id → sessions.user_id
  - Relationship: One user can have multiple active sessions
  - ON DELETE CASCADE: Deleting user deletes all their sessions
```

---

### 2. Go Structs and Models

```go
// internal/portal/models/user.go

package models

import "time"

// User represents an authenticated user in the Portal context
// Bounded Context: Portal (authentication identity)
// Layer: Shared across all layers
type User struct {
    ID                int       `json:"id" db:"id"`
    GitHubID          int64     `json:"github_id" db:"github_id" binding:"required"`
    Username          string    `json:"username" db:"username" binding:"required"`
    Email             string    `json:"email" db:"email"`
    AvatarURL         string    `json:"avatar_url" db:"avatar_url"`
    GitHubAccessToken string    `json:"-" db:"github_access_token"` // Never serialize to JSON
    CreatedAt         time.Time `json:"created_at" db:"created_at"`
    UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// Session represents an active user session
// Bounded Context: Portal (authentication state)
type Session struct {
    ID        int       `json:"id" db:"id"`
    UserID    int       `json:"user_id" db:"user_id"`
    TokenHash string    `json:"-" db:"token_hash"` // Never expose token
    ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GitHubProfile represents data from GitHub API
// Used for OAuth flow only (not stored directly)
type GitHubProfile struct {
    ID        int64  `json:"id"`
    Login     string `json:"login"`
    Email     string `json:"email"`
    AvatarURL string `json:"avatar_url"`
    Name      string `json:"name"`
}

// Validation: Gin uses `binding` tags
// - required: Field cannot be empty
// - email: Must be valid email format
```

---

### 3. API Endpoints

```go
// cmd/portal/handlers/auth_handler.go

package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/interfaces"
    "github.com/rs/zerolog/log"
)

type AuthHandler struct {
    authService interfaces.AuthService
}

func NewAuthHandler(authService interfaces.AuthService) *AuthHandler {
    return &AuthHandler{authService: authService}
}

// HandleGitHubLogin redirects user to GitHub OAuth authorization
// Method: GET
// Path: /auth/github/login
// Auth: Public
func (h *AuthHandler) HandleGitHubLogin(c *gin.Context) {
    // Build GitHub OAuth URL
    clientID := os.Getenv("GITHUB_CLIENT_ID")
    redirectURL := os.Getenv("GITHUB_CALLBACK_URL")

    authURL := fmt.Sprintf(
        "https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email,repo",
        clientID, redirectURL,
    )

    c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// HandleGitHubCallback processes OAuth callback from GitHub
// Method: GET
// Path: /auth/github/callback
// Auth: Public
func (h *AuthHandler) HandleGitHubCallback(c *gin.Context) {
    // 1. Extract OAuth code
    code := c.Query("code")
    if code == "" {
        log.Error().Msg("GitHub OAuth callback missing code parameter")
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Authentication failed: missing authorization code",
        })
        return
    }

    // 2. Complete authentication via service layer
    user, token, err := h.authService.AuthenticateWithGitHub(c.Request.Context(), code)
    if err != nil {
        log.Error().
            Err(err).
            Str("endpoint", "github_callback").
            Msg("GitHub authentication failed")

        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Authentication failed. Please try again.",
        })
        return
    }

    // 3. Set JWT as HTTP-only cookie (secure session)
    c.SetCookie(
        "session_token",
        token,
        86400,           // 24 hours
        "/",
        "",
        true,            // Secure (HTTPS only)
        true,            // HttpOnly (no JavaScript access)
    )

    // 4. Redirect to dashboard
    log.Info().
        Int("user_id", user.ID).
        Str("username", user.Username).
        Msg("User authenticated successfully")

    c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
}

// HandleLogout revokes user session and clears cookie
// Method: POST
// Path: /auth/logout
// Auth: Required
func (h *AuthHandler) HandleLogout(c *gin.Context) {
    // Get token from cookie
    token, err := c.Cookie("session_token")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "No active session",
        })
        return
    }

    // Revoke session in database
    if err := h.authService.RevokeSession(c.Request.Context(), token); err != nil {
        log.Error().Err(err).Msg("Session revocation failed")
        // Continue anyway - clear cookie
    }

    // Clear cookie
    c.SetCookie("session_token", "", -1, "/", "", true, true)

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Logged out successfully",
    })
}
```

**Endpoint Specification:**

```
GET /auth/github/login
Request: None
Response: 302 Redirect to GitHub OAuth

GET /auth/github/callback?code={oauth_code}
Request: Query parameter `code` from GitHub
Response (Success - 302 Redirect):
  - Sets session_token cookie
  - Redirects to /dashboard

Response (Error - 400/500):
{
    "error": "User-friendly error message"
}

POST /auth/logout
Request: Cookie `session_token`
Response (200 OK):
{
    "success": true,
    "message": "Logged out successfully"
}
```

---

### 4. Service Layer Implementation

```go
// internal/portal/services/auth_service.go

package services

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/interfaces"
    "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
    "github.com/rs/zerolog"
)

type AuthServiceImpl struct {
    userRepo      interfaces.UserRepository
    githubClient  interfaces.GitHubClient
    jwtSecret     []byte
    tokenExpiry   time.Duration
    logger        *zerolog.Logger
}

func NewAuthService(
    userRepo interfaces.UserRepository,
    githubClient interfaces.GitHubClient,
    jwtSecret string,
    logger *zerolog.Logger,
) *AuthServiceImpl {
    return &AuthServiceImpl{
        userRepo:     userRepo,
        githubClient: githubClient,
        jwtSecret:    []byte(jwtSecret),
        tokenExpiry:  24 * time.Hour,
        logger:       logger,
    }
}

// AuthenticateWithGitHub completes GitHub OAuth flow
// Business logic: Exchange code → Get profile → Create/update user → Generate JWT
func (s *AuthServiceImpl) AuthenticateWithGitHub(ctx context.Context, code string) (*models.User, string, error) {
    // 1. Exchange OAuth code for access token
    accessToken, err := s.githubClient.ExchangeCodeForToken(ctx, code)
    if err != nil {
        return nil, "", fmt.Errorf("GitHub token exchange failed: %w", err)
    }

    // 2. Fetch user profile from GitHub API
    profile, err := s.githubClient.GetUserProfile(ctx, accessToken)
    if err != nil {
        return nil, "", fmt.Errorf("GitHub profile fetch failed: %w", err)
    }

    // 3. Create or update user in database
    user := &models.User{
        GitHubID:          profile.ID,
        Username:          profile.Login,
        Email:             profile.Email,
        AvatarURL:         profile.AvatarURL,
        GitHubAccessToken: accessToken, // TODO: Encrypt before storing
    }

    if err := s.userRepo.CreateOrUpdate(ctx, user); err != nil {
        return nil, "", fmt.Errorf("user persistence failed: %w", err)
    }

    // 4. Generate JWT token
    token, err := s.generateJWT(user.ID)
    if err != nil {
        return nil, "", fmt.Errorf("JWT generation failed: %w", err)
    }

    // 5. Store session in database (for revocation capability)
    tokenHash := s.hashToken(token)
    session := &models.Session{
        UserID:    user.ID,
        TokenHash: tokenHash,
        ExpiresAt: time.Now().Add(s.tokenExpiry),
    }

    if err := s.userRepo.CreateSession(ctx, session); err != nil {
        return nil, "", fmt.Errorf("session creation failed: %w", err)
    }

    return user, token, nil
}

// ValidateSession checks JWT validity and returns user
func (s *AuthServiceImpl) ValidateSession(ctx context.Context, tokenString string) (*models.User, error) {
    // 1. Parse and verify JWT
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return s.jwtSecret, nil
    })

    if err != nil || !token.Valid {
        return nil, fmt.Errorf("invalid token: %w", err)
    }

    // 2. Extract user ID from claims
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, fmt.Errorf("invalid token claims")
    }

    userID := int(claims["user_id"].(float64))

    // 3. Check if session exists and is not revoked
    tokenHash := s.hashToken(tokenString)
    exists, err := s.userRepo.SessionExists(ctx, tokenHash)
    if err != nil || !exists {
        return nil, fmt.Errorf("session not found or revoked")
    }

    // 4. Fetch user from database
    user, err := s.userRepo.FindByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("user not found: %w", err)
    }

    return user, nil
}

// RevokeSession invalidates JWT by removing from sessions table
func (s *AuthServiceImpl) RevokeSession(ctx context.Context, tokenString string) error {
    tokenHash := s.hashToken(tokenString)
    return s.userRepo.DeleteSession(ctx, tokenHash)
}

// generateJWT creates a signed JWT token
func (s *AuthServiceImpl) generateJWT(userID int) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(s.tokenExpiry).Unix(),
        "iat":     time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.jwtSecret)
}

// hashToken creates SHA-256 hash for storage
func (s *AuthServiceImpl) hashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return hex.EncodeToString(hash[:])
}
```

---

### 5. Data Layer Implementation

```go
// internal/portal/db/user_repository.go

package db

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
)

type UserRepositoryImpl struct {
    db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepositoryImpl {
    return &UserRepositoryImpl{db: db}
}

// CreateOrUpdate inserts new user or updates if GitHub ID exists
func (r *UserRepositoryImpl) CreateOrUpdate(ctx context.Context, user *models.User) error {
    query := `
        INSERT INTO portal.users (github_id, username, email, avatar_url, github_access_token, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW())
        ON CONFLICT (github_id)
        DO UPDATE SET
            username = EXCLUDED.username,
            email = EXCLUDED.email,
            avatar_url = EXCLUDED.avatar_url,
            github_access_token = EXCLUDED.github_access_token,
            updated_at = NOW()
        RETURNING id, created_at, updated_at
    `

    err := r.db.QueryRow(
        ctx, query,
        user.GitHubID, user.Username, user.Email, user.AvatarURL, user.GitHubAccessToken,
    ).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

    if err != nil {
        return fmt.Errorf("user upsert failed: %w", err)
    }

    return nil
}

// FindByGitHubID retrieves user by GitHub ID
func (r *UserRepositoryImpl) FindByGitHubID(ctx context.Context, githubID int64) (*models.User, error) {
    query := `
        SELECT id, github_id, username, email, avatar_url, github_access_token, created_at, updated_at
        FROM portal.users
        WHERE github_id = $1
    `

    var user models.User
    err := r.db.QueryRow(ctx, query, githubID).Scan(
        &user.ID, &user.GitHubID, &user.Username, &user.Email,
        &user.AvatarURL, &user.GitHubAccessToken, &user.CreatedAt, &user.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found: github_id=%d", githubID)
    }
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }

    return &user, nil
}

// FindByID retrieves user by internal ID
func (r *UserRepositoryImpl) FindByID(ctx context.Context, id int) (*models.User, error) {
    query := `
        SELECT id, github_id, username, email, avatar_url, created_at, updated_at
        FROM portal.users
        WHERE id = $1
    `

    var user models.User
    err := r.db.QueryRow(ctx, query, id).Scan(
        &user.ID, &user.GitHubID, &user.Username, &user.Email,
        &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found: id=%d", id)
    }
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }

    return &user, nil
}

// CreateSession stores a new session
func (r *UserRepositoryImpl) CreateSession(ctx context.Context, session *models.Session) error {
    query := `
        INSERT INTO portal.sessions (user_id, token_hash, expires_at)
        VALUES ($1, $2, $3)
        RETURNING id, created_at
    `

    err := r.db.QueryRow(
        ctx, query,
        session.UserID, session.TokenHash, session.ExpiresAt,
    ).Scan(&session.ID, &session.CreatedAt)

    if err != nil {
        return fmt.Errorf("session creation failed: %w", err)
    }

    return nil
}

// SessionExists checks if session is valid and not expired
func (r *UserRepositoryImpl) SessionExists(ctx context.Context, tokenHash string) (bool, error) {
    query := `
        SELECT EXISTS(
            SELECT 1 FROM portal.sessions
            WHERE token_hash = $1
            AND expires_at > NOW()
        )
    `

    var exists bool
    err := r.db.QueryRow(ctx, query, tokenHash).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("session check failed: %w", err)
    }

    return exists, nil
}

// DeleteSession removes session (logout)
func (r *UserRepositoryImpl) DeleteSession(ctx context.Context, tokenHash string) error {
    query := `DELETE FROM portal.sessions WHERE token_hash = $1`

    _, err := r.db.Exec(ctx, query, tokenHash)
    if err != nil {
        return fmt.Errorf("session deletion failed: %w", err)
    }

    return nil
}
```

**SQL Query Guidelines:**
- ✅ Parameterized queries (`$1`, `$2`) prevent SQL injection
- ✅ Use context for cancellation/timeouts
- ✅ `ON CONFLICT` for upsert (PostgreSQL feature)
- ✅ `RETURNING` to get generated IDs
- ❌ No business logic in SQL (simple CRUD only)

---

### 6. Template Implementation

```go
// cmd/portal/templates/login.templ

package templates

// LoginPage renders the GitHub OAuth login page
templ LoginPage() {
    @Layout("Login - DevSmith Platform") {
        <div class="hero min-h-screen bg-base-200">
            <div class="hero-content text-center">
                <div class="max-w-md">
                    <h1 class="text-5xl font-bold mb-8">DevSmith Platform</h1>
                    <p class="text-xl mb-8">
                        Learn to read code effectively.<br/>
                        Master the Human-in-the-Loop skill.
                    </p>

                    <!-- GitHub Login Button -->
                    <a
                        href="/auth/github/login"
                        class="btn btn-primary btn-lg gap-2"
                    >
                        <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
                            <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
                        </svg>
                        Login with GitHub
                    </a>

                    <p class="text-sm text-gray-500 mt-8">
                        By logging in, you agree to access your GitHub profile and repositories
                        to enable code reading sessions.
                    </p>
                </div>
            </div>
        </div>
    }
}
```

```go
// cmd/portal/templates/dashboard.templ

package templates

import "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"

// DashboardPage renders portal dashboard with app browser
templ DashboardPage(user *models.User, apps []AppInfo) {
    @Layout("Dashboard - DevSmith Platform") {
        @Header(user)

        <div class="container mx-auto p-8">
            <h1 class="text-3xl font-bold mb-2">Welcome, {user.Username}</h1>
            <p class="text-gray-600 mb-8">Select an app to get started with your learning journey</p>

            <!-- App Browser -->
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                for _, app := range apps {
                    @AppCard(app)
                }
            </div>
        </div>
    }
}
```

```go
// cmd/portal/templates/components/app_card.templ

package templates

// AppCard renders a single app card with enable/disable toggle
templ AppCard(app AppInfo) {
    <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
            <h2 class="card-title">{app.Name}</h2>
            <p class="text-sm text-gray-600">{app.Description}</p>

            <div class="card-actions justify-between items-center mt-4">
                <div class="badge" class={app.BadgeClass}>{app.Status}</div>

                <!-- HTMX toggle button -->
                <button
                    class="btn btn-sm btn-primary"
                    hx-post={"/api/portal/apps/" + app.ID + "/toggle"}
                    hx-target="closest .card"
                    hx-swap="outerHTML"
                >
                    if app.Enabled {
                        Launch
                    } else {
                        Enable
                    }
                </button>
            </div>
        </div>
    </div>
}
```

**HTMX Integration:**
- `hx-post`: HTTP POST request to toggle endpoint
- `hx-target="closest .card"`: Replace the card with updated version
- `hx-swap="outerHTML"`: Replace entire element
- Server returns updated AppCard component

---

### 7. Testing Requirements

#### Unit Tests (70%+ coverage)

```go
// internal/portal/services/auth_service_test.go

package services

import (
    "context"
    "errors"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
)

// Mocks
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) CreateOrUpdate(ctx context.Context, user *models.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) FindByGitHubID(ctx context.Context, githubID int64) (*models.User, error) {
    args := m.Called(ctx, githubID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

type MockGitHubClient struct {
    mock.Mock
}

func (m *MockGitHubClient) ExchangeCodeForToken(ctx context.Context, code string) (string, error) {
    args := m.Called(ctx, code)
    return args.String(0), args.Error(1)
}

func (m *MockGitHubClient) GetUserProfile(ctx context.Context, accessToken string) (*models.GitHubProfile, error) {
    args := m.Called(ctx, accessToken)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.GitHubProfile), args.Error(1)
}

// Test: Successful GitHub authentication
func TestAuthService_AuthenticateWithGitHub_Success(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    mockGitHub := new(MockGitHubClient)
    logger := zerolog.New(os.Stdout)
    service := NewAuthService(mockRepo, mockGitHub, "test-secret", &logger)

    ctx := context.Background()
    oauthCode := "test-oauth-code"

    // Mock GitHub token exchange
    mockGitHub.On("ExchangeCodeForToken", ctx, oauthCode).
        Return("github-access-token", nil)

    // Mock GitHub profile fetch
    profile := &models.GitHubProfile{
        ID:        12345,
        Login:     "testuser",
        Email:     "test@example.com",
        AvatarURL: "https://github.com/avatar.png",
    }
    mockGitHub.On("GetUserProfile", ctx, "github-access-token").
        Return(profile, nil)

    // Mock user creation
    mockRepo.On("CreateOrUpdate", ctx, mock.AnythingOfType("*models.User")).
        Return(nil).
        Run(func(args mock.Arguments) {
            user := args.Get(1).(*models.User)
            user.ID = 1 // Simulate database ID assignment
        })

    // Mock session creation
    mockRepo.On("CreateSession", ctx, mock.AnythingOfType("*models.Session")).
        Return(nil)

    // Act
    user, token, err := service.AuthenticateWithGitHub(ctx, oauthCode)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "testuser", user.Username)
    assert.Equal(t, int64(12345), user.GitHubID)
    assert.NotEmpty(t, token)

    mockGitHub.AssertExpectations(t)
    mockRepo.AssertExpectations(t)
}

// Test: GitHub token exchange fails
func TestAuthService_AuthenticateWithGitHub_TokenExchangeFails(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    mockGitHub := new(MockGitHubClient)
    logger := zerolog.New(os.Stdout)
    service := NewAuthService(mockRepo, mockGitHub, "test-secret", &logger)

    ctx := context.Background()

    // Mock error
    mockGitHub.On("ExchangeCodeForToken", ctx, "invalid-code").
        Return("", errors.New("invalid authorization code"))

    // Act
    user, token, err := service.AuthenticateWithGitHub(ctx, "invalid-code")

    // Assert
    assert.Error(t, err)
    assert.Nil(t, user)
    assert.Empty(t, token)
    assert.Contains(t, err.Error(), "token exchange failed")
}

// Test: JWT validation succeeds
func TestAuthService_ValidateSession_Success(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    mockGitHub := new(MockGitHubClient)
    logger := zerolog.New(os.Stdout)
    service := NewAuthService(mockRepo, mockGitHub, "test-secret", &logger)

    ctx := context.Background()

    // Create valid JWT
    token, _ := service.generateJWT(1)
    tokenHash := service.hashToken(token)

    // Mock session exists
    mockRepo.On("SessionExists", ctx, tokenHash).Return(true, nil)

    // Mock user fetch
    user := &models.User{ID: 1, Username: "testuser"}
    mockRepo.On("FindByID", ctx, 1).Return(user, nil)

    // Act
    validatedUser, err := service.ValidateSession(ctx, token)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, validatedUser)
    assert.Equal(t, 1, validatedUser.ID)

    mockRepo.AssertExpectations(t)
}
```

#### Integration Tests

```go
// internal/portal/db/user_repository_integration_test.go
// +build integration

package db

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
)

func TestUserRepository_CreateOrUpdate_Integration(t *testing.T) {
    // Requires real database (use test container)
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    repo := NewUserRepository(db)
    ctx := context.Background()

    // Test insert
    user := &models.User{
        GitHubID:  12345,
        Username:  "testuser",
        Email:     "test@example.com",
        AvatarURL: "https://example.com/avatar.png",
    }

    err := repo.CreateOrUpdate(ctx, user)
    assert.NoError(t, err)
    assert.NotZero(t, user.ID)

    // Test update
    user.Username = "updated-user"
    err = repo.CreateOrUpdate(ctx, user)
    assert.NoError(t, err)

    // Verify
    found, err := repo.FindByGitHubID(ctx, 12345)
    assert.NoError(t, err)
    assert.Equal(t, "updated-user", found.Username)
}
```

#### Test Coverage Commands
```bash
# Run unit tests
go test ./internal/portal/...

# Run with coverage
go test -cover ./internal/portal/...

# Generate coverage report
go test -coverprofile=coverage.out ./internal/portal/...
go tool cover -html=coverage.out

# Run integration tests (requires Docker)
go test -tags=integration ./internal/portal/db/...
```

---

## Implementation Checklist

### Phase 1: Setup ✅
- [ ] Create branch: `feature/002-portal-authentication`
- [ ] Create migration files in `internal/portal/db/migrations/`
- [ ] Run migrations: `make migrate-up`
- [ ] Define models in `internal/portal/models/`

### Phase 2: Data Layer ✅
- [ ] Create `internal/portal/interfaces/auth_interface.go`
- [ ] Implement `internal/portal/db/user_repository.go`
- [ ] Write repository tests `user_repository_test.go`
- [ ] Verify tests pass: `go test ./internal/portal/db/...`

### Phase 3: Service Layer ✅
- [ ] Implement `internal/portal/services/github_client.go`
- [ ] Implement `internal/portal/services/auth_service.go`
- [ ] Write service tests with mocks
- [ ] Verify tests pass: `go test ./internal/portal/services/...`

### Phase 4: Controller Layer ✅
- [ ] Create `cmd/portal/handlers/auth_handler.go`
- [ ] Create Templ templates (login.templ, dashboard.templ)
- [ ] Generate templates: `templ generate`
- [ ] Write handler tests
- [ ] Verify tests pass: `go test ./cmd/portal/handlers/...`

### Phase 5: Integration ✅
- [ ] Wire up dependencies in `cmd/portal/main.go`
- [ ] Register routes: `/auth/github/login`, `/auth/github/callback`, `/auth/logout`
- [ ] Update nginx config to route `/auth/*` to portal service
- [ ] Test locally: `make dev && curl http://localhost:3000/auth/github/login`
- [ ] Test OAuth flow end-to-end in browser

### Phase 6: Documentation ✅
- [ ] Update `AI_CHANGELOG.md` with implementation notes
- [ ] Add inline comments for JWT generation logic
- [ ] Document GitHub OAuth setup in README

### Phase 7: Code Quality ✅
- [ ] Run linter: `golangci-lint run ./internal/portal/... ./cmd/portal/...`
- [ ] Fix linting issues
- [ ] Format code: `gofmt -w .`
- [ ] Check coverage: `go test -cover ./internal/portal/...`
- [ ] Ensure 70%+ coverage

### Phase 8: Commit and PR ✅
- [ ] Stage changes: `git add .`
- [ ] Commit:
      ```
      feat(portal): implement GitHub OAuth authentication

      - GitHub OAuth login flow with callback handling
      - JWT-based session management
      - User profile storage in portal.users table
      - Session persistence in portal.sessions table
      - Logout functionality with session revocation
      - Layered architecture (handler → service → repository)
      - 75% test coverage with unit and integration tests

      Implements bounded context separation (Portal context only knows
      about authentication identity, not code review concerns).

      Closes #2
      ```
- [ ] Push: `git push origin feature/002-portal-authentication`
- [ ] Create PR to `development`
- [ ] Verify CI passes
- [ ] Request review from Claude

---

## Cognitive Load Optimization Notes

### For Intrinsic Complexity (Simplify)
- JWT handling is complex → Encapsulated in `generateJWT()` helper
- OAuth flow has many steps → Split into service methods
- Clear naming: `AuthenticateWithGitHub` not `Auth`
- Comments explain "why" (e.g., "Why store token hash: enables revocation")

### For Extraneous Load (Reduce)
- No magic strings: Use constants for cookie names, token expiry
- No global state: All dependencies passed via constructor
- Explicit errors: Wrap errors with context (`fmt.Errorf("...%w", err)`)
- No abbreviations: `githubClient` not `ghc`

### For Germane Load (Maximize)
- Follows existing patterns: Handler → Service → Repository (3-layer)
- Respects bounded contexts: Portal context only
- Uses Go idioms: Explicit error handling, struct-based dependency injection
- Interfaces enable testing: Can mock GitHub API

---

## Questions and Clarifications

### Before Starting Implementation
- [x] Bounded context clear: Portal = authentication only
- [x] Layering understood: 3 layers with no shortcuts
- [x] Dependencies identified: pgxpool, Gin, JWT library, Templ
- [x] Acceptance criteria measurable: Can login, sessions persist, unauthorized blocked

### During Implementation
If you encounter:
- **GitHub API rate limits** → Use personal access token for testing
- **JWT secret management** → Use environment variable, document in .env.example
- **Session cleanup** → Future issue: cron job to delete expired sessions
- **Token encryption** → Future issue: encrypt github_access_token before storage (use AES-256)

---

## References
- ARCHITECTURE.md - Mental Models (Bounded Context, Layering, Abstractions, Scope)
- ARCHITECTURE.md - Portal Service specification (lines 475-695)
- Requirements.md - Authentication requirements (lines 220-285)
- DevsmithTDD.md - Authentication tests (lines 150-198)
- Go JWT library: https://github.com/golang-jwt/jwt
- GitHub OAuth docs: https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps

---

**Next Steps:**
1. Read this spec completely
2. Ask clarifying questions in issue comments if needed
3. Follow the implementation checklist phase by phase
4. Run tests after each phase (`go test ./...`)
5. Create PR when all phases complete
6. Tag Claude for code review (Critical reading mode)

**Estimated Autonomous Time:** 1.5 - 2 hours
**Test Coverage Target:** 70%+ (aim for 75%+)
**Success Metric:** User can log in with GitHub and see dashboard
