// Package review_models contains data structures for review sessions and analysis.
package review_models

import "time"

// CodeReviewSession represents a complete code review session.
// It tracks the current state across all reading modes and maintains history.
//
//nolint:fieldalignment // Data models prioritize readability over memory optimization
type CodeReviewSession struct {
	ID              int64
	UserID          int64
	SessionDuration int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastAccessedAt  time.Time
	CompletedAt     *time.Time
	ModeStates      map[string]ModeState
	Title           string
	Description     string
	CodeSource      string
	CodeContent     string
	GithubRepo      string
	GithubBranch    string
	GithubPath      string
	Language        string
	Status          string
	CurrentMode     string
}

// ModeState tracks the state of a single reading mode within a session.
type ModeState struct {
	AnalysisDuration    int64
	ResultID            int64
	IssuesFound         int
	QualityScore        int
	AnalysisStartedAt   *time.Time
	AnalysisCompletedAt *time.Time
	Mode                string
	Status              string
	UserNotes           string
	LastError           string
	IsCompleted         bool
}

// SessionHistory tracks changes to a session over time.
type SessionHistory struct {
	ID        int64
	SessionID int64
	ActedBy   int64
	CreatedAt time.Time
	Action    string
	Mode      string
	OldValue  string
	NewValue  string
	Changes   string
}

// SessionStatistics provides overview metrics for a session.
type SessionStatistics struct {
	SessionID           int64
	TotalDuration       int64
	ModesCovered        int
	TotalIssuesFound    int
	CriticalIssuesCount int
	HighIssuesCount     int
	MediumIssuesCount   int
	AverageQualityScore int
	CreatedAt           time.Time
	FastestModeAnalysis string
	SlowestModeAnalysis string
}

// SessionFilter for querying sessions.
type SessionFilter struct {
	UserID    int64
	Limit     int
	Offset    int
	DateFrom  time.Time
	DateTo    time.Time
	Status    string
	Language  string
	SortBy    string
	SortOrder string
	HasErrors bool
}

// SessionSummary provides a brief overview for list views.
type SessionSummary struct {
	ID              int64
	DurationSeconds int64
	ModeProgress    int
	CreatedAt       time.Time
	LastAccessedAt  time.Time
	Title           string
	CodeSource      string
	Language        string
	Status          string
	CurrentMode     string
}
