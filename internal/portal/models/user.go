package models

import "time"

// User represents a user in the portal system
type User struct {
	GitHubID  int64     `json:"github_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
