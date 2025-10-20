package models

import "time"

// User represents a user in the portal system
type User struct {
	ID                 int       `json:"id"`
	GitHubID           int64     `json:"github_id"`
	Username           string    `json:"username"`
	Email              string    `json:"email"`
	AvatarURL          string    `json:"avatar_url"`
	GitHubAccessToken  string    `json:"github_access_token"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// GitHubProfile represents the user's GitHub profile data
type GitHubProfile struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}
