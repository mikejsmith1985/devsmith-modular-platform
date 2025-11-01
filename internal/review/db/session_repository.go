// Package review_db provides database access for review sessions.
package review_db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// SessionRepository handles CRUD and query operations for CodeReviewSession.
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new SessionRepository.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create inserts a new session and returns it with generated ID.
func (r *SessionRepository) Create(ctx context.Context, session *review_models.CodeReviewSession) (*review_models.CodeReviewSession, error) {
	query := `
		INSERT INTO review.sessions (
			user_id, title, description, code_source, code_content,
			github_repo, github_branch, github_path, language,
			status, current_mode, created_at, updated_at, last_accessed
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at, last_accessed
	`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		session.UserID, session.Title, session.Description,
		session.CodeSource, session.CodeContent,
		session.GithubRepo, session.GithubBranch, session.GithubPath,
		session.Language, session.Status, session.CurrentMode,
		now, now, now,
	).Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt, &session.LastAccessedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetByID retrieves a session by ID with all related mode states.
func (r *SessionRepository) GetByID(ctx context.Context, sessionID int64) (*review_models.CodeReviewSession, error) {
	query := `
		SELECT id, user_id, title, description, code_source, code_content,
		       github_repo, github_branch, github_path, language,
		       status, current_mode, created_at, updated_at, last_accessed,
		       completed_at, session_duration_seconds
		FROM review.sessions
		WHERE id = $1
	`

	session := &review_models.CodeReviewSession{}
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID, &session.UserID, &session.Title, &session.Description,
		&session.CodeSource, &session.CodeContent,
		&session.GithubRepo, &session.GithubBranch, &session.GithubPath,
		&session.Language, &session.Status, &session.CurrentMode,
		&session.CreatedAt, &session.UpdatedAt, &session.LastAccessedAt,
		&session.CompletedAt, &session.SessionDuration,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Load mode states
	modeStates, err := r.getModeStates(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load mode states: %w", err)
	}
	session.ModeStates = modeStates

	return session, nil
}

// Update updates an existing session.
func (r *SessionRepository) Update(ctx context.Context, session *review_models.CodeReviewSession) error {
	query := `
		UPDATE review.sessions
		SET title = $1, description = $2, code_source = $3, code_content = $4,
		    github_repo = $5, github_branch = $6, github_path = $7, language = $8,
		    status = $9, current_mode = $10, completed_at = $11,
		    session_duration_seconds = $12, updated_at = NOW()
		WHERE id = $13
	`

	result, err := r.db.ExecContext(ctx, query,
		session.Title, session.Description, session.CodeSource,
		session.CodeContent, session.GithubRepo, session.GithubBranch,
		session.GithubPath, session.Language, session.Status,
		session.CurrentMode, session.CompletedAt, session.SessionDuration,
		session.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check update result: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// UpdateLastAccessed updates the last_accessed timestamp.
func (r *SessionRepository) UpdateLastAccessed(ctx context.Context, sessionID int64) error {
	query := `UPDATE review.sessions SET last_accessed = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update last accessed: %w", err)
	}
	return nil
}

// List retrieves sessions with filtering and pagination.
func (r *SessionRepository) List(ctx context.Context, filter *review_models.SessionFilter) ([]*review_models.SessionSummary, error) {
	query := `
		SELECT s.id, s.title, s.code_source, s.language, s.status, s.current_mode,
		       COALESCE(ss.mode_progress, 0),
		       s.created_at, s.last_accessed,
		       EXTRACT(EPOCH FROM (COALESCE(s.completed_at, NOW()) - s.created_at))::BIGINT
		FROM review.sessions s
		LEFT JOIN review.session_summaries ss ON s.id = ss.id
		WHERE s.user_id = $1
	`

	args := []interface{}{filter.UserID}
	argCount := 2

	// Add status filter
	if filter.Status != "" {
		query += fmt.Sprintf(" AND s.status = $%d", argCount)
		args = append(args, filter.Status)
		argCount++
	}

	// Add language filter
	if filter.Language != "" {
		query += fmt.Sprintf(" AND s.language = $%d", argCount)
		args = append(args, filter.Language)
		argCount++
	}

	// Add date range filters
	if !filter.DateFrom.IsZero() {
		query += fmt.Sprintf(" AND s.created_at >= $%d", argCount)
		args = append(args, filter.DateFrom)
		argCount++
	}

	if !filter.DateTo.IsZero() {
		query += fmt.Sprintf(" AND s.created_at <= $%d", argCount)
		args = append(args, filter.DateTo)
		argCount++
	}

	// Add sort
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder != "" {
		sortOrder = filter.SortOrder
	}
	query += fmt.Sprintf(" ORDER BY s.%s %s", sortBy, sortOrder)

	// Add pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
		args = append(args, filter.Limit, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var summaries []*review_models.SessionSummary
	for rows.Next() {
		summary := &review_models.SessionSummary{}
		err := rows.Scan(
			&summary.ID, &summary.Title, &summary.CodeSource, &summary.Language,
			&summary.Status, &summary.CurrentMode, &summary.ModeProgress,
			&summary.CreatedAt, &summary.LastAccessedAt, &summary.DurationSeconds,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session summary: %w", err)
		}
		summaries = append(summaries, summary)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return summaries, nil
}

// Delete removes a session and all related data (cascade).
func (r *SessionRepository) Delete(ctx context.Context, sessionID int64) error {
	query := `DELETE FROM review.sessions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check delete result: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// getModeStates loads all mode states for a session.
func (r *SessionRepository) getModeStates(ctx context.Context, sessionID int64) (map[string]review_models.ModeState, error) {
	query := `
		SELECT mode, status, is_completed, analysis_started_at,
		       analysis_completed_at, analysis_duration_ms, result_id,
		       user_notes, issues_found, quality_score, last_error
		FROM review.mode_states
		WHERE session_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query mode states: %w", err)
	}
	defer rows.Close()

	modeStates := make(map[string]review_models.ModeState)
	for rows.Next() {
		var ms review_models.ModeState
		err := rows.Scan(
			&ms.Mode, &ms.Status, &ms.IsCompleted,
			&ms.AnalysisStartedAt, &ms.AnalysisCompletedAt, &ms.AnalysisDuration,
			&ms.ResultID, &ms.UserNotes, &ms.IssuesFound, &ms.QualityScore,
			&ms.LastError,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mode state: %w", err)
		}
		modeStates[ms.Mode] = ms
	}

	return modeStates, rows.Err()
}
