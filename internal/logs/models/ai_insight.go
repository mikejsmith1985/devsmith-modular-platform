package logs_models

import "time"

// AIInsight represents AI-generated analysis for a log entry
type AIInsight struct {
	ID          int64     `json:"id" db:"id"`
	LogID       int64     `json:"log_id" db:"log_id"`
	Analysis    string    `json:"analysis" db:"analysis"`
	RootCause   string    `json:"root_cause" db:"root_cause"`
	Suggestions []string  `json:"suggestions" db:"suggestions"`
	ModelUsed   string    `json:"model_used" db:"model_used"`
	GeneratedAt time.Time `json:"generated_at" db:"generated_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
