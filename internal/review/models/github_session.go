// Package review_models contains data structures for review sessions and analysis.
package review_models

import (
	"time"

	"github.com/google/uuid"
)

// GitHubSession represents a GitHub repository session with cached tree structure.
// Associated with a review session for GitHub-sourced code analysis.
type GitHubSession struct {
	ID               int64     `json:"id" db:"id"`
	SessionID        int64     `json:"session_id" db:"session_id"`
	GitHubURL        string    `json:"github_url" db:"github_url"`
	Owner            string    `json:"owner" db:"owner"`
	Repo             string    `json:"repo" db:"repo"`
	Branch           string    `json:"branch" db:"branch"`
	CommitSHA        string    `json:"commit_sha,omitempty" db:"commit_sha"`
	FileTree         []byte    `json:"file_tree,omitempty" db:"file_tree"` // JSONB stored as []byte
	TotalFiles       int       `json:"total_files" db:"total_files"`
	TotalDirectories int       `json:"total_directories" db:"total_directories"`
	TreeLastSynced   time.Time `json:"tree_last_synced,omitempty" db:"tree_last_synced"`
	IsPrivate        bool      `json:"is_private" db:"is_private"`
	StarsCount       int       `json:"stars_count" db:"stars_count"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// OpenFile represents a file opened in a tab within the multi-tab UI.
// Each tab has a unique UUID and tracks its position and activity state.
type OpenFile struct {
	ID              int64     `json:"id" db:"id"`
	GitHubSessionID int64     `json:"github_session_id" db:"github_session_id"`
	TabID           uuid.UUID `json:"tab_id" db:"tab_id"`
	FilePath        string    `json:"file_path" db:"file_path"`
	FileSHA         string    `json:"file_sha,omitempty" db:"file_sha"`
	FileContent     string    `json:"file_content,omitempty" db:"file_content"`
	FileSize        int64     `json:"file_size" db:"file_size"`
	Language        string    `json:"language,omitempty" db:"language"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	TabOrder        int       `json:"tab_order" db:"tab_order"`
	OpenedAt        time.Time `json:"opened_at" db:"opened_at"`
	LastAccessed    time.Time `json:"last_accessed" db:"last_accessed"`
	AnalysisCount   int       `json:"analysis_count" db:"analysis_count"`
}

// MultiFileAnalysis represents an analysis performed across multiple files.
// Tracks cross-file dependencies, shared abstractions, and architecture patterns.
type MultiFileAnalysis struct {
	ID                    int64     `json:"id" db:"id"`
	GitHubSessionID       int64     `json:"github_session_id" db:"github_session_id"`
	FilePaths             []string  `json:"file_paths" db:"file_paths"` // Array type
	ReadingMode           string    `json:"reading_mode" db:"reading_mode"`
	CombinedContent       string    `json:"combined_content,omitempty" db:"combined_content"`
	AIResponse            []byte    `json:"ai_response,omitempty" db:"ai_response"`                         // JSONB
	CrossFileDependencies []byte    `json:"cross_file_dependencies,omitempty" db:"cross_file_dependencies"` // JSONB
	SharedAbstractions    []byte    `json:"shared_abstractions,omitempty" db:"shared_abstractions"`         // JSONB
	ArchitecturePatterns  []byte    `json:"architecture_patterns,omitempty" db:"architecture_patterns"`     // JSONB
	AnalysisDurationMs    int64     `json:"analysis_duration_ms" db:"analysis_duration_ms"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
}

// TreeNode represents a file or directory in the repository tree (for JSON marshaling).
type TreeNode struct {
	Path     string     `json:"path"`
	Type     string     `json:"type"` // "file" or "dir"
	SHA      string     `json:"sha"`
	Size     int64      `json:"size,omitempty"`
	Children []TreeNode `json:"children,omitempty"`
}

// FileTreeJSON represents the structure stored in github_sessions.file_tree JSONB column.
type FileTreeJSON struct {
	RootNodes []TreeNode `json:"rootNodes"`
}

// CrossFileDependency represents a detected dependency between files.
type CrossFileDependency struct {
	FromFile   string   `json:"from_file"`
	ToFile     string   `json:"to_file"`
	ImportType string   `json:"import_type"`       // "import", "require", "include", etc.
	Symbols    []string `json:"symbols,omitempty"` // Specific symbols imported
}

// SharedAbstraction represents a common pattern or interface found across files.
type SharedAbstraction struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"` // "interface", "base_class", "pattern", etc.
	Files       []string `json:"files"`
	Description string   `json:"description"`
	Complexity  string   `json:"complexity,omitempty"` // "simple", "moderate", "complex"
}

// ArchitecturePattern represents a detected architecture pattern.
type ArchitecturePattern struct {
	Pattern     string   `json:"pattern"`    // "MVC", "Repository", "Factory", etc.
	Confidence  float64  `json:"confidence"` // 0.0 to 1.0
	Files       []string `json:"files"`
	Description string   `json:"description"`
	Evidence    []string `json:"evidence"` // Specific code evidence
}

// AIAnalysisResponse represents the parsed AI response for multi-file analysis.
type AIAnalysisResponse struct {
	Summary              string                `json:"summary"`
	Dependencies         []CrossFileDependency `json:"dependencies"`
	SharedAbstractions   []SharedAbstraction   `json:"shared_abstractions"`
	ArchitecturePatterns []ArchitecturePattern `json:"architecture_patterns"`
	Recommendations      []string              `json:"recommendations"`
	Issues               []AnalysisIssue       `json:"issues,omitempty"`
}

// AnalysisIssue represents a specific issue found during analysis.
type AnalysisIssue struct {
	File        string `json:"file"`
	Line        int    `json:"line,omitempty"`
	Severity    string `json:"severity"` // "critical", "high", "medium", "low"
	Category    string `json:"category"` // "architecture", "security", "performance", etc.
	Description string `json:"description"`
	Suggestion  string `json:"suggestion,omitempty"`
}
