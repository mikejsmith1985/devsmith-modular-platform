package review_db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// GitHubRepository handles database operations for GitHub sessions
type GitHubRepository struct {
	db *sql.DB
}

// NewGitHubRepository creates a new GitHub repository instance
func NewGitHubRepository(db *sql.DB) *GitHubRepository {
	return &GitHubRepository{db: db}
}

// CreateGitHubSession creates a new GitHub session
func (r *GitHubRepository) CreateGitHubSession(ctx context.Context, session *review_models.GitHubSession) error {
	query := `
		INSERT INTO reviews.github_sessions 
		(session_id, github_url, owner, repo, branch, commit_sha, file_tree, total_files, total_directories, is_private, stars_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		session.SessionID, session.GitHubURL, session.Owner, session.Repo,
		session.Branch, session.CommitSHA, session.FileTree, session.TotalFiles,
		session.TotalDirectories, session.IsPrivate, session.StarsCount,
	).Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create github session: %w", err)
	}

	return nil
}

// GetGitHubSession retrieves a GitHub session by ID
//
//nolint:dupl // Scanning code is similar but serves different purposes
func (r *GitHubRepository) GetGitHubSession(ctx context.Context, id int64) (*review_models.GitHubSession, error) {
	query := `
		SELECT id, session_id, github_url, owner, repo, branch, commit_sha, 
		       file_tree, total_files, total_directories, tree_last_synced,
		       is_private, stars_count, created_at, updated_at
		FROM reviews.github_sessions
		WHERE id = $1
	`

	session := &review_models.GitHubSession{}
	var treeLastSynced sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID, &session.SessionID, &session.GitHubURL, &session.Owner,
		&session.Repo, &session.Branch, &session.CommitSHA, &session.FileTree,
		&session.TotalFiles, &session.TotalDirectories, &treeLastSynced,
		&session.IsPrivate, &session.StarsCount, &session.CreatedAt, &session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("github session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get github session: %w", err)
	}

	if treeLastSynced.Valid {
		session.TreeLastSynced = treeLastSynced.Time
	}

	return session, nil
}

// GetGitHubSessionBySessionID retrieves a GitHub session by review session ID
func (r *GitHubRepository) GetGitHubSessionBySessionID(ctx context.Context, sessionID int64) (*review_models.GitHubSession, error) {
	query := `
		SELECT id, session_id, github_url, owner, repo, branch, commit_sha, 
		       file_tree, total_files, total_directories, tree_last_synced,
		       is_private, stars_count, created_at, updated_at
		FROM reviews.github_sessions
		WHERE session_id = $1
	`

	session := &review_models.GitHubSession{}
	var treeLastSynced sql.NullTime

	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID, &session.SessionID, &session.GitHubURL, &session.Owner,
		&session.Repo, &session.Branch, &session.CommitSHA, &session.FileTree,
		&session.TotalFiles, &session.TotalDirectories, &treeLastSynced,
		&session.IsPrivate, &session.StarsCount, &session.CreatedAt, &session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("github session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get github session: %w", err)
	}

	if treeLastSynced.Valid {
		session.TreeLastSynced = treeLastSynced.Time
	}

	return session, nil
}

// UpdateFileTree updates the cached file tree for a GitHub session
func (r *GitHubRepository) UpdateFileTree(ctx context.Context, id int64, tree []byte, totalFiles, totalDirs int) error {
	query := `
		UPDATE reviews.github_sessions
		SET file_tree = $1, total_files = $2, total_directories = $3, tree_last_synced = $4
		WHERE id = $5
	`

	result, err := r.db.ExecContext(ctx, query, tree, totalFiles, totalDirs, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update file tree: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("github session not found")
	}

	return nil
}

// CreateOpenFile creates a new open file entry
func (r *GitHubRepository) CreateOpenFile(ctx context.Context, file *review_models.OpenFile) error {
	query := `
		INSERT INTO reviews.open_files
		(github_session_id, tab_id, file_path, file_sha, file_content, file_size, language, is_active, tab_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, opened_at, last_accessed, analysis_count
	`

	err := r.db.QueryRowContext(
		ctx, query,
		file.GitHubSessionID, file.TabID, file.FilePath, file.FileSHA,
		file.FileContent, file.FileSize, file.Language, file.IsActive, file.TabOrder,
	).Scan(&file.ID, &file.OpenedAt, &file.LastAccessed, &file.AnalysisCount)

	if err != nil {
		return fmt.Errorf("failed to create open file: %w", err)
	}

	return nil
}

// GetOpenFiles retrieves all open files for a GitHub session
func (r *GitHubRepository) GetOpenFiles(ctx context.Context, githubSessionID int64) ([]*review_models.OpenFile, error) {
	query := `
		SELECT id, github_session_id, tab_id, file_path, file_sha, file_content,
		       file_size, language, is_active, tab_order, opened_at, last_accessed, analysis_count
		FROM reviews.open_files
		WHERE github_session_id = $1
		ORDER BY tab_order ASC
	`

	rows, err := r.db.QueryContext(ctx, query, githubSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query open files: %w", err)
	}
	defer rows.Close()

	var files []*review_models.OpenFile
	for rows.Next() {
		file := &review_models.OpenFile{}
		err := rows.Scan(
			&file.ID, &file.GitHubSessionID, &file.TabID, &file.FilePath,
			&file.FileSHA, &file.FileContent, &file.FileSize, &file.Language,
			&file.IsActive, &file.TabOrder, &file.OpenedAt, &file.LastAccessed,
			&file.AnalysisCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan open file: %w", err)
		}
		files = append(files, file)
	}

	return files, nil
}

// GetOpenFileByTabID retrieves an open file by tab ID
func (r *GitHubRepository) GetOpenFileByTabID(ctx context.Context, tabID uuid.UUID) (*review_models.OpenFile, error) {
	query := `
		SELECT id, github_session_id, tab_id, file_path, file_sha, file_content,
		       file_size, language, is_active, tab_order, opened_at, last_accessed, analysis_count
		FROM reviews.open_files
		WHERE tab_id = $1
	`

	file := &review_models.OpenFile{}
	err := r.db.QueryRowContext(ctx, query, tabID).Scan(
		&file.ID, &file.GitHubSessionID, &file.TabID, &file.FilePath,
		&file.FileSHA, &file.FileContent, &file.FileSize, &file.Language,
		&file.IsActive, &file.TabOrder, &file.OpenedAt, &file.LastAccessed,
		&file.AnalysisCount,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("open file not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get open file: %w", err)
	}

	return file, nil
}

// SetActiveTab sets a tab as active and deactivates others
func (r *GitHubRepository) SetActiveTab(ctx context.Context, githubSessionID int64, tabID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Deactivate all tabs
	_, err = tx.ExecContext(ctx, `
		UPDATE reviews.open_files
		SET is_active = false
		WHERE github_session_id = $1
	`, githubSessionID)
	if err != nil {
		return fmt.Errorf("failed to deactivate tabs: %w", err)
	}

	// Activate the specified tab
	result, err := tx.ExecContext(ctx, `
		UPDATE reviews.open_files
		SET is_active = true
		WHERE tab_id = $1
	`, tabID)
	if err != nil {
		return fmt.Errorf("failed to activate tab: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("tab not found")
	}

	return tx.Commit()
}

// CloseFile closes an open file (deletes it)
func (r *GitHubRepository) CloseFile(ctx context.Context, tabID uuid.UUID) error {
	query := `DELETE FROM reviews.open_files WHERE tab_id = $1`

	result, err := r.db.ExecContext(ctx, query, tabID)
	if err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

// IncrementAnalysisCount increments the analysis count for a file
func (r *GitHubRepository) IncrementAnalysisCount(ctx context.Context, tabID uuid.UUID) error {
	query := `
		UPDATE reviews.open_files
		SET analysis_count = analysis_count + 1
		WHERE tab_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, tabID)
	if err != nil {
		return fmt.Errorf("failed to increment analysis count: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

// CreateMultiFileAnalysis creates a new multi-file analysis record
func (r *GitHubRepository) CreateMultiFileAnalysis(ctx context.Context, analysis *review_models.MultiFileAnalysis) error {
	query := `
		INSERT INTO reviews.multi_file_analysis
		(github_session_id, file_paths, reading_mode, combined_content, ai_response,
		 cross_file_dependencies, shared_abstractions, architecture_patterns, analysis_duration_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		analysis.GitHubSessionID, analysis.FilePaths, analysis.ReadingMode,
		analysis.CombinedContent, analysis.AIResponse, analysis.CrossFileDependencies,
		analysis.SharedAbstractions, analysis.ArchitecturePatterns, analysis.AnalysisDurationMs,
	).Scan(&analysis.ID, &analysis.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create multi-file analysis: %w", err)
	}

	return nil
}

// GetMultiFileAnalyses retrieves all multi-file analyses for a GitHub session
func (r *GitHubRepository) GetMultiFileAnalyses(ctx context.Context, githubSessionID int64) ([]*review_models.MultiFileAnalysis, error) {
	query := `
		SELECT id, github_session_id, file_paths, reading_mode, combined_content,
		       ai_response, cross_file_dependencies, shared_abstractions,
		       architecture_patterns, analysis_duration_ms, created_at
		FROM reviews.multi_file_analysis
		WHERE github_session_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, githubSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query multi-file analyses: %w", err)
	}
	defer rows.Close()

	var analyses []*review_models.MultiFileAnalysis
	for rows.Next() {
		analysis := &review_models.MultiFileAnalysis{}
		err := rows.Scan(
			&analysis.ID, &analysis.GitHubSessionID, &analysis.FilePaths,
			&analysis.ReadingMode, &analysis.CombinedContent, &analysis.AIResponse,
			&analysis.CrossFileDependencies, &analysis.SharedAbstractions,
			&analysis.ArchitecturePatterns, &analysis.AnalysisDurationMs, &analysis.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan multi-file analysis: %w", err)
		}
		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

// GetLatestMultiFileAnalysis retrieves the most recent multi-file analysis
func (r *GitHubRepository) GetLatestMultiFileAnalysis(ctx context.Context, githubSessionID int64) (*review_models.MultiFileAnalysis, error) {
	query := `
		SELECT id, github_session_id, file_paths, reading_mode, combined_content,
		       ai_response, cross_file_dependencies, shared_abstractions,
		       architecture_patterns, analysis_duration_ms, created_at
		FROM reviews.multi_file_analysis
		WHERE github_session_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	analysis := &review_models.MultiFileAnalysis{}
	err := r.db.QueryRowContext(ctx, query, githubSessionID).Scan(
		&analysis.ID, &analysis.GitHubSessionID, &analysis.FilePaths,
		&analysis.ReadingMode, &analysis.CombinedContent, &analysis.AIResponse,
		&analysis.CrossFileDependencies, &analysis.SharedAbstractions,
		&analysis.ArchitecturePatterns, &analysis.AnalysisDurationMs, &analysis.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no multi-file analysis found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest multi-file analysis: %w", err)
	}

	return analysis, nil
}

// ParseFileTree unmarshals JSONB file tree to Go struct
func ParseFileTree(data []byte) (*review_models.FileTreeJSON, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var tree review_models.FileTreeJSON
	err := json.Unmarshal(data, &tree)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file tree: %w", err)
	}

	return &tree, nil
}

// MarshalFileTree marshals Go struct to JSONB
func MarshalFileTree(tree *review_models.FileTreeJSON) ([]byte, error) {
	if tree == nil {
		return nil, nil
	}

	data, err := json.Marshal(tree)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal file tree: %w", err)
	}

	return data, nil
}
