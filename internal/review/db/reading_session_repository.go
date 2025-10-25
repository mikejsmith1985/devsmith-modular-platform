package db

import (
	"context"
	"database/sql"
	"fmt"
)

// ReadingSession represents a reading session in the database.
type ReadingSession struct {
	AIResponse      string
	ScanQuery       string
	UserAnnotations string
	TargetPath      string
	ReadingMode     string
	CreatedAt       string
	ID              int64
	SessionID       int64
}

// ReadingSessionRepository handles CRUD operations for ReadingSession.
type ReadingSessionRepository struct {
	DB *sql.DB
}

// NewReadingSessionRepository creates a new ReadingSessionRepository.
func NewReadingSessionRepository(db *sql.DB) *ReadingSessionRepository {
	return &ReadingSessionRepository{DB: db}
}

// Create inserts a new reading session and returns it with ID.
func (r *ReadingSessionRepository) Create(ctx context.Context, rs *ReadingSession) (*ReadingSession, error) {
	query := `INSERT INTO reviews.reading_sessions (session_id, reading_mode, target_path, scan_query, ai_response, user_annotations) 
             VALUES ($1, $2, $3, $4, $5, $6) 
             RETURNING id, created_at`
	err := r.DB.QueryRowContext(ctx, query,
		rs.SessionID, rs.ReadingMode, rs.TargetPath, rs.ScanQuery, rs.AIResponse, rs.UserAnnotations,
	).Scan(&rs.ID, &rs.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("db: failed to create reading session: %w", err)
	}
	return rs, nil
}

// GetByID retrieves a reading session by ID.
func (r *ReadingSessionRepository) GetByID(ctx context.Context, id int64) (*ReadingSession, error) {
	row := r.DB.QueryRowContext(ctx,
		`SELECT id, session_id, reading_mode, target_path, scan_query, ai_response, user_annotations, created_at 
         FROM reviews.reading_sessions WHERE id = $1`, id)
	var rs ReadingSession
	err := row.Scan(&rs.ID, &rs.SessionID, &rs.ReadingMode, &rs.TargetPath, &rs.ScanQuery, &rs.AIResponse, &rs.UserAnnotations, &rs.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("db: failed to get reading session: %w", err)
	}
	return &rs, nil
}

// GetBySessionID retrieves all reading sessions for a review session.
func (r *ReadingSessionRepository) GetBySessionID(ctx context.Context, sessionID int64) ([]*ReadingSession, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, session_id, reading_mode, target_path, scan_query, ai_response, user_annotations, created_at 
         FROM reviews.reading_sessions WHERE session_id = $1 ORDER BY created_at DESC`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("db: failed to query reading sessions: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("warning: failed to close rows: %v\n", closeErr)
		}
	}()

	var sessions []*ReadingSession
	for rows.Next() {
		var rs ReadingSession
		if err := rows.Scan(&rs.ID, &rs.SessionID, &rs.ReadingMode, &rs.TargetPath, &rs.ScanQuery, &rs.AIResponse, &rs.UserAnnotations, &rs.CreatedAt); err != nil {
			return nil, fmt.Errorf("db: failed to scan reading session: %w", err)
		}
		sessions = append(sessions, &rs)
	}
	return sessions, nil
}

// Update updates an existing reading session.
func (r *ReadingSessionRepository) Update(ctx context.Context, rs *ReadingSession) error {
	query := `UPDATE reviews.reading_sessions SET reading_mode = $1, target_path = $2, scan_query = $3, ai_response = $4, user_annotations = $5 WHERE id = $6`
	result, err := r.DB.ExecContext(ctx, query,
		rs.ReadingMode, rs.TargetPath, rs.ScanQuery, rs.AIResponse, rs.UserAnnotations, rs.ID,
	)
	if err != nil {
		return fmt.Errorf("db: failed to update reading session: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return fmt.Errorf("db: reading session not found")
	}
	return nil
}

// Delete deletes a reading session by ID.
func (r *ReadingSessionRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.DB.ExecContext(ctx, `DELETE FROM reviews.reading_sessions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("db: failed to delete reading session: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return fmt.Errorf("db: reading session not found")
	}
	return nil
}
