package portal_handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	templates "github.com/mikejsmith1985/devsmith-modular-platform/apps/portal/templates"
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

// LogsDashboardHandler serves the logs dashboard page
func LogsDashboardHandler(c *gin.Context) {
	userClaims, err := getUserClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Fetch dashboard data from logs service API
	logsServiceURL := os.Getenv("LOGS_SERVICE_URL")
	if logsServiceURL == "" {
		logsServiceURL = "http://localhost:8082"
	}

	// Fetch dashboard stats
	stats, err := fetchLogsData(c.Request.Context(), logsServiceURL+"/api/logs/dashboard/stats")
	if err != nil {
		log.Printf("[WARN] Failed to fetch dashboard stats: %v", err)
		stats = map[string]interface{}{}
	}

	// Fetch top errors
	topErrors, err := fetchLogsData(c.Request.Context(), logsServiceURL+"/api/logs/validations/top-errors?limit=10&days=1")
	if err != nil {
		log.Printf("[WARN] Failed to fetch top errors: %v", err)
		topErrors = []interface{}{}
	}

	// Fetch error trends
	trends, err := fetchLogsData(c.Request.Context(), logsServiceURL+"/api/logs/validations/trends?days=1")
	if err != nil {
		log.Printf("[WARN] Failed to fetch trends: %v", err)
		trends = []interface{}{}
	}

	// Normalize fetched data into expected types for the template
	var statsMap map[string]interface{}
	if s, ok := stats.(map[string]interface{}); ok {
		statsMap = s
	} else {
		statsMap = map[string]interface{}{}
	}

	var topErrorsSlice []interface{}
	if te, ok := topErrors.([]interface{}); ok {
		topErrorsSlice = te
	} else {
		topErrorsSlice = []interface{}{}
	}

	var trendsSlice []interface{}
	if tr, ok := trends.([]interface{}); ok {
		trendsSlice = tr
	} else {
		trendsSlice = []interface{}{}
	}

	dashboardData := templates.LogsDashboardData{
		User:      templates.DashboardUser{Username: userClaims.Username, Email: userClaims.Email, AvatarURL: userClaims.AvatarURL},
		Stats:     statsMap,
		TopErrors: topErrorsSlice,
		Trends:    trendsSlice,
	}

	component := templates.LogsDashboard(dashboardData)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("[ERROR] Failed to render logs dashboard component: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render logs dashboard"})
		return
	}
}

// fetchLogsData fetches JSON data from logs service API
func fetchLogsData(ctx context.Context, url string) (interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("[WARN] Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

// getJWTKey returns the shared JWT signing key.
func getJWTKey() []byte {
	return []byte("your-secret-key")
}
