// Package services provides GitHub API integration for the portal service.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
)

// GitHubClientImpl implements the GitHubClient interface for interacting with GitHub's API.
// It provides methods to exchange OAuth codes and fetch user profiles from GitHub.
type GitHubClientImpl struct {
	clientID     string
	clientSecret string
}

// NewGitHubClient creates a new GitHubClientImpl with the given client ID and secret.
func NewGitHubClient(clientID, clientSecret string) *GitHubClientImpl {
	return &GitHubClientImpl{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// ExchangeCodeForToken exchanges an OAuth code for a GitHub access token.
func (g *GitHubClientImpl) ExchangeCodeForToken(ctx context.Context, code string) (string, error) {
	url := "https://github.com/login/oauth/access_token"
	payload := fmt.Sprintf("client_id=%s&client_secret=%s&code=%s", g.clientID, g.clientSecret, code)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("GitHub error: %s", result.Error)
	}
	return result.AccessToken, nil
}

// GetUserProfile fetches the authenticated user's GitHub profile using the access token.
func (g *GitHubClientImpl) GetUserProfile(ctx context.Context, accessToken string) (*models.GitHubProfile, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var profile struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}
	return &models.GitHubProfile{
		ID:        profile.ID,
		Username:  profile.Login,
		Email:     profile.Email,
		AvatarURL: profile.AvatarURL,
	}, nil
}
