// Package review_models contains data structures for review sessions and analysis.
package review_models

import "time"

// CodeReviewSession represents a complete code review session.
// It tracks the current state across all reading modes and maintains history.
type CodeReviewSession struct {
	ID               int64                    `json:"id"`
	UserID           int64                    `json:"user_id"`
	Title            string                   `json:"title"`
	Description      string                   `json:"description,omitempty"`
	CodeSource       string                   `json:"code_source"` // paste, github, upload
	CodeContent      string                   `json:"code_content"`
	GithubRepo       string                   `json:"github_repo,omitempty"`
	GithubBranch     string                   `json:"github_branch,omitempty"`
	GithubPath       string                   `json:"github_path,omitempty"`
	Language         string                   `json:"language,omitempty"` // go, python, etc
	Status           string                   `json:"status"`             // active, completed, archived
	CurrentMode      string                   `json:"current_mode,omitempty"`
	ModeStates       map[string]ModeState     `json:"mode_states"` // Track state per mode
	CreatedAt        time.Time                `json:"created_at"`
	UpdatedAt        time.Time                `json:"updated_at"`
	LastAccessedAt   time.Time                `json:"last_accessed_at"`
	CompletedAt      *time.Time               `json:"completed_at,omitempty"`
	SessionDuration  int64                    `json:"session_duration_seconds"` // Total time spent
}

// ModeState tracks the state of a single reading mode within a session.
type ModeState struct {
	Mode              string            `json:"mode"` // critical, preview, skim, scan, detailed
	Status            string            `json:"status"`
	IsCompleted       bool              `json:"is_completed"`
	AnalysisStartedAt *time.Time        `json:"analysis_started_at,omitempty"`
	AnalysisCompletedAt *time.Time      `json:"analysis_completed_at,omitempty"`
	AnalysisDuration  int64             `json:"analysis_duration_ms"`
	ResultID          int64             `json:"result_id,omitempty"` // FK to analysis result
	UserNotes         string            `json:"user_notes,omitempty"`
	IssuesFound       int               `json:"issues_found"`
	QualityScore      int               `json:"quality_score"` // 0-100
	LastError         string            `json:"last_error,omitempty"`
}

// SessionHistory tracks changes to a session over time.
type SessionHistory struct {
	ID              int64     `json:"id"`
	SessionID       int64     `json:"session_id"`
	Action          string    `json:"action"` // created, mode_started, mode_completed, notes_updated, status_changed
	Mode            string    `json:"mode,omitempty"`
	OldValue        string    `json:"old_value,omitempty"`
	NewValue        string    `json:"new_value,omitempty"`
	Changes         string    `json:"changes"` // JSON diff
	ActedBy         int64     `json:"acted_by"` // User ID
	CreatedAt       time.Time `json:"created_at"`
}

// SessionStatistics provides overview metrics for a session.
type SessionStatistics struct {
	SessionID               int64     `json:"session_id"`
	TotalDuration           int64     `json:"total_duration_seconds"`
	ModesCovered            int       `json:"modes_covered"` // How many modes completed
	TotalIssuesFound        int       `json:"total_issues_found"`
	CriticalIssuesCount     int       `json:"critical_issues_count"`
	HighIssuesCount         int       `json:"high_issues_count"`
	MediumIssuesCount       int       `json:"medium_issues_count"`
	AverageQualityScore     int       `json:"average_quality_score"` // Across all modes
	FastestModeAnalysis     string    `json:"fastest_mode_analysis"` // Mode name
	SlowestModeAnalysis     string    `json:"slowest_mode_analysis"` // Mode name
	CreatedAt               time.Time `json:"created_at"`
}

// SessionFilter for querying sessions.
type SessionFilter struct {
	UserID       int64
	Status       string // active, completed, archived
	Language     string
	DateFrom     time.Time
	DateTo       time.Time
	HasErrors    bool
	SortBy       string // created, updated, accessed
	SortOrder    string // asc, desc
	Limit        int
	Offset       int
}

// SessionSummary provides a brief overview for list views.
type SessionSummary struct {
	ID             int64     `json:"id"`
	Title          string    `json:"title"`
	CodeSource     string    `json:"code_source"`
	Language       string    `json:"language,omitempty"`
	Status         string    `json:"status"`
	CurrentMode    string    `json:"current_mode,omitempty"`
	ModeProgress   int       `json:"mode_progress"` // 0-100% based on modes completed
	CreatedAt      time.Time `json:"created_at"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
	DurationSeconds int64    `json:"duration_seconds"`
}
