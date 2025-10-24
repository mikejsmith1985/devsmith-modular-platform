package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/templates"
)

// DashboardHandler serves the main dashboard page
func DashboardHandler(c *gin.Context) {
	userClaims, err := getUserClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] User claims: %+v\n", userClaims)

	dashboardUser := templates.DashboardUser{
		Username:  userClaims.Username,
		Email:     userClaims.Email,
		AvatarURL: userClaims.AvatarURL,
	}

	component := templates.Dashboard(dashboardUser)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("[ERROR] Failed to render dashboard component: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render dashboard"})
		return
	}
}

// getUserClaims extracts and validates user claims from context or cookie
func getUserClaims(c *gin.Context) (*UserClaims, error) {
	claims, exists := c.Get("user")
	if !exists {
		log.Println("[DEBUG] User context not found, trying cookie")
		return getUserClaimsFromCookie(c)
	}

	userClaims, ok := claims.(*UserClaims)
	if !ok {
		log.Printf("[DEBUG] Invalid user context type: %T\n", claims)
		return nil, fmt.Errorf("Invalid claims")
	}

	return userClaims, nil
}

// getUserClaimsFromCookie parses JWT from cookie and extracts claims
func getUserClaimsFromCookie(c *gin.Context) (*UserClaims, error) {
	cookie, err := c.Cookie("devsmith_token")
	if err != nil {
		return nil, fmt.Errorf("Authorization header or cookie missing")
	}

	jwtKey := getJWTKey()
	token, err := jwt.ParseWithClaims(cookie, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("Invalid token")
	}

	log.Printf("Parsed token: %+v\n", token)

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		log.Printf("Invalid token or claims. Token: %+v\n", token)
		return nil, fmt.Errorf("Invalid token claims")
	}

	log.Printf("Token is valid. Claims: %+v\n", claims)
	c.Set("user", claims)

	log.Printf("[DEBUG] Raw token: %s", cookie)
	parts := strings.Split(cookie, ".")
	if len(parts) == 3 {
		log.Printf("[DEBUG] Token header: %s", parts[0])
		log.Printf("[DEBUG] Token payload: %s", parts[1])
		log.Printf("[DEBUG] Token signature: %s", parts[2])
	} else {
		log.Printf("[DEBUG] Token format invalid: %s", cookie)
	}

	return claims, nil
}

// GetUserInfoHandler returns current user info as JSON
func GetUserInfoHandler(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userClaims, ok := claims.(*UserClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user claims"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username":   userClaims.Username,
		"email":      userClaims.Email,
		"avatar_url": userClaims.AvatarURL,
		"github_id":  userClaims.GithubID,
		"created_at": userClaims.CreatedAt,
	})
	log.Printf("Decoded JWT payload: %+v\n", userClaims)
}

// getJWTKey returns the shared JWT signing key.
func getJWTKey() []byte {
	return []byte("your-secret-key")
}
