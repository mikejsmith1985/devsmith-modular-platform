package review_models

import "time"

// PromptTemplate represents an AI prompt template for code review modes
type PromptTemplate struct {
	ID         string    `json:"id" db:"id"`
	UserID     *int      `json:"user_id,omitempty" db:"user_id"` // NULL = system default
	Mode       string    `json:"mode" db:"mode"`                 // "preview", "skim", "scan", "detailed", "critical"
	UserLevel  string    `json:"user_level" db:"user_level"`     // "beginner", "intermediate", "expert"
	OutputMode string    `json:"output_mode" db:"output_mode"`   // "quick", "detailed", "comprehensive"
	PromptText string    `json:"prompt_text" db:"prompt_text"`
	Variables  []string  `json:"variables" db:"-"` // Parsed from JSON
	IsDefault  bool      `json:"is_default" db:"is_default"`
	Version    int       `json:"version" db:"version"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// PromptExecution represents a logged execution of a prompt
type PromptExecution struct {
	ID             int64     `json:"id" db:"id"`
	TemplateID     string    `json:"template_id" db:"template_id"`
	UserID         int       `json:"user_id" db:"user_id"`
	RenderedPrompt string    `json:"rendered_prompt" db:"rendered_prompt"`
	Response       string    `json:"response" db:"response"`
	ModelUsed      string    `json:"model_used" db:"model_used"`
	LatencyMs      int       `json:"latency_ms" db:"latency_ms"`
	TokensUsed     int       `json:"tokens_used" db:"tokens_used"`
	UserRating     *int      `json:"user_rating,omitempty" db:"user_rating"` // 1-5 stars
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// IsCustom returns true if this is a user custom prompt (not system default)
func (pt *PromptTemplate) IsCustom() bool {
	return pt.UserID != nil
}

// CanBeDeleted returns true if this prompt can be deleted (user custom only)
func (pt *PromptTemplate) CanBeDeleted() bool {
	return pt.IsCustom() && !pt.IsDefault
}
