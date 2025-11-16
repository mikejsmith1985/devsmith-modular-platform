// Package security provides shared authentication and authorization utilities
// for the DevSmith platform (Portal, Review, Logs, Analytics services).
package security

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// UserClaims represents the JWT claims for authenticated users.
// This structure is shared across all DevSmith services for consistent authentication.
type UserClaims struct {
	jwt.RegisteredClaims
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	GithubID  string    `json:"github_id"`
}

// GetJWTSecret returns the JWT signing secret from environment.
// Panics if JWT_SECRET is not set - this is intentional to prevent insecure defaults.
func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is not set - this is required for secure authentication")
	}
	return []byte(secret)
}

// ValidateJWT validates a JWT token string and returns the user claims
func ValidateJWT(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return GetJWTSecret(), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// CreateJWT creates a new JWT token for a user
func CreateJWT(username, email, avatarURL, githubID string) (string, error) {
	claims := UserClaims{
		Username:  username,
		Email:     email,
		AvatarURL: avatarURL,
		GithubID:  githubID,
		CreatedAt: time.Now(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetJWTSecret())
}
