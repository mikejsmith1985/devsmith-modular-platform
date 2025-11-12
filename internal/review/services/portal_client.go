package review_services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PortalClient handles communication with the Portal service's AI Factory API
type PortalClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPortalClient creates a new Portal API client
func NewPortalClient(portalURL string) *PortalClient {
	return &PortalClient{
		baseURL: portalURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// LLMConfig represents an AI model configuration from Portal's AI Factory
type LLMConfig struct {
	ID           string  `json:"id"`
	UserID       int     `json:"user_id"`
	Provider     string  `json:"provider"`
	ModelName    string  `json:"model_name"`
	APIEndpoint  string  `json:"api_endpoint,omitempty"`
	APIKey       string  `json:"api_key,omitempty"` // Decrypted by Portal
	IsDefault    bool    `json:"is_default"`
	MaxTokens    int     `json:"max_tokens"`
	Temperature  float64 `json:"temperature"`
}

// AppPreferencesResponse is the response from Portal's app preferences endpoint
type AppPreferencesResponse struct {
	Review    *LLMConfig `json:"review"`
	Logs      *LLMConfig `json:"logs"`
	Analytics *LLMConfig `json:"analytics"`
}

// GetEffectiveConfigForApp fetches the user's effective LLM configuration for a specific app
// This respects the user's AI Factory settings: app-specific preference > default > system default
func (c *PortalClient) GetEffectiveConfigForApp(ctx context.Context, sessionToken, appName string) (*LLMConfig, error) {
	// Build request URL
	url := fmt.Sprintf("%s/api/portal/app-llm-preferences", c.baseURL)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie for authentication
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: sessionToken,
	})

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Portal API: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Portal API returned %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var prefsResp AppPreferencesResponse
	if err := json.NewDecoder(resp.Body).Decode(&prefsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract config for requested app
	var config *LLMConfig
	switch appName {
	case "review":
		config = prefsResp.Review
	case "logs":
		config = prefsResp.Logs
	case "analytics":
		config = prefsResp.Analytics
	default:
		return nil, fmt.Errorf("unknown app name: %s", appName)
	}

	if config == nil {
		return nil, fmt.Errorf("no LLM configuration found for app: %s. Please configure a model in AI Factory", appName)
	}

	return config, nil
}
