// Package models contains user and profile data structures for the portal service.
package models

import "time"

// User represents a user in the portal system
type User struct {
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	AvatarURL         string    `json:"avatar_url"`
	GitHubAccessToken string    `json:"github_access_token"`
	ID                int       `json:"id"`
	GitHubID          int64     `json:"github_id"`
}

// GitHubProfile represents the user's GitHub profile data
type GitHubProfile struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	ID        int64  `json:"id"`
}
