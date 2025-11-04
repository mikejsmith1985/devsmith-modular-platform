package security

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateJWT(t *testing.T) {
	// Act
	token, err := CreateJWT("testuser", "test@example.com", "https://example.com/avatar.png", "12345")

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token can be parsed
	claims, err := ValidateJWT(token)
	require.NoError(t, err)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "12345", claims.GithubID)
}

func TestValidateJWT_ValidToken(t *testing.T) {
	// Arrange: Create a valid token
	token, err := CreateJWT("testuser", "test@example.com", "https://example.com/avatar.png", "12345")
	require.NoError(t, err)

	// Act
	claims, err := ValidateJWT(token)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "https://example.com/avatar.png", claims.AvatarURL)
	assert.Equal(t, "12345", claims.GithubID)
	assert.NotZero(t, claims.CreatedAt)
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	// Act
	claims, err := ValidateJWT("invalid.jwt.token")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	// Arrange: Create an expired token
	claims := UserClaims{
		Username:  "testuser",
		Email:     "test@example.com",
		GithubID:  "12345",
		CreatedAt: time.Now().Add(-48 * time.Hour),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-25 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(GetJWTSecret())
	require.NoError(t, err)

	// Act
	parsedClaims, err := ValidateJWT(tokenString)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestValidateJWT_WrongSigningMethod(t *testing.T) {
	// Arrange: Create token with wrong signing method
	claims := UserClaims{
		Username:  "testuser",
		GithubID:  "12345",
		CreatedAt: time.Now(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Use RS256 instead of HS256 (intentionally wrong)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// Act
	parsedClaims, err := ValidateJWT(tokenString)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
}

func TestGetJWTSecret(t *testing.T) {
	// Act
	secret := GetJWTSecret()

	// Assert
	assert.NotEmpty(t, secret)
	assert.Equal(t, []byte("your-secret-key"), secret) // Default secret
}
