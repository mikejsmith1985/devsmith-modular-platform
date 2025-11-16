package portal_handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	portal_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/services"
)

// LLMConfigHandler handles HTTP requests for LLM configuration management
type LLMConfigHandler struct {
	service *portal_services.LLMConfigService
}

// NewLLMConfigHandler creates a new LLM configuration handler
func NewLLMConfigHandler(service *portal_services.LLMConfigService) *LLMConfigHandler {
	return &LLMConfigHandler{
		service: service,
	}
}

// getUserIDFromContext extracts the authenticated user ID from the Gin context
// Returns user ID and true if found, 0 and false otherwise
func getUserIDFromContext(c *gin.Context) (int, bool) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	// Try to convert to int (may be int or string depending on middleware)
	switch v := userIDInterface.(type) {
	case int:
		return v, true
	case string:
		userID, err := strconv.Atoi(v)
		if err != nil {
			return 0, false
		}
		return userID, true
	default:
		return 0, false
	}
}

// ListLLMConfigs handles GET /api/portal/llm-configs
// Returns all LLM configurations for the authenticated user
func (h *LLMConfigHandler) ListLLMConfigs(c *gin.Context) {
	// Extract user ID from authentication context
	userID, exists := getUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Get configs from service
	configs, err := h.service.ListUserConfigs(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve configurations"})
		return
	}

	// Convert to response format (exclude encrypted API keys)
	response := make([]gin.H, 0, len(configs))
	for _, config := range configs {
		configData := gin.H{
			"id":          config.ID,
			"name":        config.Provider + " - " + config.ModelName, // Computed name
			"provider":    config.Provider,
			"model":       config.ModelName,
			"is_default":  config.IsDefault,
			"created_at":  config.CreatedAt,
			"updated_at":  config.UpdatedAt,
			"has_api_key": config.APIKeyEncrypted.Valid && config.APIKeyEncrypted.String != "", // Boolean flag
		}
		if config.APIEndpoint.Valid {
			configData["endpoint"] = config.APIEndpoint.String
		}
		response = append(response, configData)
	}

	c.JSON(http.StatusOK, response)
}

// CreateLLMConfig handles POST /api/portal/llm-configs
// Creates a new LLM configuration for the authenticated user
func (h *LLMConfigHandler) CreateLLMConfig(c *gin.Context) {
	// Extract user ID
	userID, exists := getUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Parse request body
	var req struct {
		Provider  string `json:"provider" binding:"required"`
		Model     string `json:"model" binding:"required"`
		APIKey    string `json:"api_key"` // Optional (not required for Ollama)
		Endpoint  string `json:"endpoint"`
		IsDefault bool   `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate provider is one of the supported types
	validProviders := map[string]bool{
		"anthropic": true,
		"openai":    true,
		"ollama":    true,
		"deepseek":  true,
		"mistral":   true,
	}
	if !validProviders[strings.ToLower(req.Provider)] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider. Must be one of: anthropic, openai, ollama, deepseek, mistral"})
		return
	}

	// Validate API key required for non-Ollama providers
	if strings.ToLower(req.Provider) != "ollama" && req.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key required for " + req.Provider})
		return
	}

	// Create config via service
	config, err := h.service.CreateConfig(
		c.Request.Context(),
		userID,
		req.Provider,
		req.Model,
		req.APIKey,
		req.IsDefault,
		req.Endpoint,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create configuration: " + err.Error()})
		return
	}

	// Return created config (exclude encrypted API key)
	response := gin.H{
		"id":          config.ID,
		"name":        config.Provider + " - " + config.ModelName,
		"provider":    config.Provider,
		"model":       config.ModelName,
		"is_default":  config.IsDefault,
		"created_at":  config.CreatedAt,
		"updated_at":  config.UpdatedAt,
		"has_api_key": config.APIKeyEncrypted.Valid && config.APIKeyEncrypted.String != "",
	}
	if config.APIEndpoint.Valid {
		response["endpoint"] = config.APIEndpoint.String
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateLLMConfig handles PUT /api/portal/llm-configs/:id
// Updates an existing LLM configuration
func (h *LLMConfigHandler) UpdateLLMConfig(c *gin.Context) {
	// Extract user ID
	userID, exists := getUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Extract config ID from URL
	configID := c.Param("id")
	if configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Config ID required"})
		return
	}

	// Parse request body
	var req struct {
		Provider  *string `json:"provider"`
		Model     *string `json:"model_name"` // Use model_name to match service
		APIKey    *string `json:"api_key"`
		Endpoint  *string `json:"endpoint"`
		IsDefault *bool   `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Provider != nil {
		updates["provider"] = *req.Provider
	}
	if req.Model != nil {
		updates["model_name"] = *req.Model
	}
	if req.APIKey != nil {
		updates["api_key"] = *req.APIKey
	}
	if req.Endpoint != nil {
		updates["endpoint"] = *req.Endpoint
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}

	// Update via service
	if err := h.service.UpdateConfig(c.Request.Context(), userID, configID, updates); err != nil {
		// Check for specific errors
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
			return
		}
		if strings.Contains(err.Error(), "permission denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this configuration"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated successfully"})
}

// DeleteLLMConfig handles DELETE /api/portal/llm-configs/:id
// Deletes an LLM configuration
func (h *LLMConfigHandler) DeleteLLMConfig(c *gin.Context) {
	// Extract user ID
	userID, exists := getUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Extract config ID from URL
	configID := c.Param("id")
	if configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Config ID required"})
		return
	}

	// Delete via service
	if err := h.service.DeleteConfig(c.Request.Context(), userID, configID); err != nil {
		// Check for specific errors
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
			return
		}
		if strings.Contains(err.Error(), "permission denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this configuration"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete configuration: " + err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// TestLLMConnection handles POST /api/portal/llm-configs/test
// Tests connection to an LLM provider without saving the configuration
func (h *LLMConfigHandler) TestLLMConnection(c *gin.Context) {
	// Extract user ID (still require auth even though not saving)
	_, exists := getUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Parse request body
	var req struct {
		Provider string `json:"provider" binding:"required"`
		Model    string `json:"model" binding:"required"`
		APIKey   string `json:"api_key"` // Optional for Ollama
		Endpoint string `json:"endpoint"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate API key for non-Ollama providers
	if strings.ToLower(req.Provider) != "ollama" && req.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key required for " + req.Provider})
		return
	}

	// TODO: Implement actual connection test using AI factory
	// For now, return a mock success response
	// This will be implemented in a follow-up task

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Connection test successful (mock implementation - to be completed)",
	})
}

// GetAppPreferences handles GET /api/portal/app-llm-preferences
// Returns the LLM configuration preferences for each app
func (h *LLMConfigHandler) GetAppPreferences(c *gin.Context) {
	// Extract user ID
	userID, exists := getUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Define the apps we support
	apps := []string{"review", "logs", "analytics"}
	preferences := make(map[string]interface{})

	// Get effective config for each app
	for _, app := range apps {
		config, err := h.service.GetEffectiveConfig(c.Request.Context(), userID, app)
		if err != nil {
			// If error getting config, set to null
			preferences[app] = nil
		} else {
			preferences[app] = gin.H{
				"config_id": config.ID,
				"provider":  config.Provider,
				"model":     config.ModelName,
			}
		}
	}

	c.JSON(http.StatusOK, preferences)
}

// SetAppPreference handles PUT /api/portal/app-llm-preferences/:app
// Sets the preferred LLM configuration for a specific app
func (h *LLMConfigHandler) SetAppPreference(c *gin.Context) {
	// Extract user ID
	userID, exists := getUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Extract app name from URL
	appName := c.Param("app")
	if appName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "App name required"})
		return
	}

	// Validate app name
	validApps := map[string]bool{
		"review":    true,
		"logs":      true,
		"analytics": true,
	}
	if !validApps[appName] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid app name. Must be one of: review, logs, analytics"})
		return
	}

	// Parse request body
	var req struct {
		ConfigID string `json:"config_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Set preference via service
	if err := h.service.SetAppPreference(c.Request.Context(), userID, appName, req.ConfigID); err != nil {
		// Check for specific errors
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
			return
		}
		if strings.Contains(err.Error(), "permission denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to use this configuration"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set app preference: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "App preference set successfully"})
}

// GetUsageSummary handles GET /api/portal/llm-usage/summary
// Returns usage statistics for the authenticated user
func (h *LLMConfigHandler) GetUsageSummary(c *gin.Context) {
	// Extract user ID
	userID, exists := getUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Get period from query parameter (default to 30 days)
	period := c.DefaultQuery("period", "30d")

	// TODO: Implement actual usage tracking
	// For now, return mock data
	// This will be implemented when we add usage tracking to the LLM service

	_ = userID // Suppress unused variable warning
	_ = period

	c.JSON(http.StatusOK, gin.H{
		"total_tokens":   0,
		"total_requests": 0,
		"total_cost":     0.0,
		"period":         period,
		"note":           "Usage tracking to be implemented in future phase",
	})
}

// SetDefaultConfig handles PUT /api/portal/llm-configs/:id/set-default
// Sets the default LLM configuration for the user
func (h *LLMConfigHandler) SetDefaultConfig(c *gin.Context) {
	// Extract user ID
	userID, exists := getUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Extract config ID from URL
	configID := c.Param("id")
	if configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Config ID required"})
		return
	}

	// Parse request body
	var req struct {
		IsDefault bool `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Update via service
	updates := map[string]interface{}{
		"is_default": req.IsDefault,
	}

	if err := h.service.UpdateConfig(c.Request.Context(), userID, configID, updates); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
			return
		}
		if strings.Contains(err.Error(), "permission denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this configuration"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update default configuration: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Default configuration updated successfully",
	})
}

// RegisterLLMConfigRoutes registers all LLM configuration routes with the router group
// The router group should already have authentication middleware applied
func RegisterLLMConfigRoutes(routerGroup *gin.RouterGroup, service *portal_services.LLMConfigService) {
	handler := NewLLMConfigHandler(service)

	// All routes are within the provided group (which already has /api/portal prefix)
	routerGroup.GET("/llm-configs", handler.ListLLMConfigs)
	routerGroup.POST("/llm-configs", handler.CreateLLMConfig)
	routerGroup.PUT("/llm-configs/:id", handler.UpdateLLMConfig)
	routerGroup.PUT("/llm-configs/:id/set-default", handler.SetDefaultConfig)
	routerGroup.DELETE("/llm-configs/:id", handler.DeleteLLMConfig)
	routerGroup.POST("/llm-configs/test", handler.TestLLMConnection)
	routerGroup.GET("/app-llm-preferences", handler.GetAppPreferences)
	routerGroup.PUT("/app-llm-preferences/:app", handler.SetAppPreference)
	routerGroup.GET("/llm-usage/summary", handler.GetUsageSummary)
}
