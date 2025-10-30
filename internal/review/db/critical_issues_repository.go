package review_db

import (
	"context"
	"database/sql"
	"fmt"
)

// CriticalIssue represents a critical issue.
type CriticalIssue struct {
	SuggestedFix     string
	Description      string
	FilePath         string
	Status           string
	IssueType        string
	Severity         string
	CreatedAt        string
	ID               int64
	ReadingSessionID int64
	LineNumber       int
}

// CriticalIssuesRepository handles CRUD operations for CriticalIssue.
type CriticalIssuesRepository struct {
	DB *sql.DB
}

// NewCriticalIssuesRepository creates a new CriticalIssuesRepository.
func NewCriticalIssuesRepository(db *sql.DB) *CriticalIssuesRepository {
	return &CriticalIssuesRepository{DB: db}
}

// Create inserts a new critical issue.
func (r *CriticalIssuesRepository) Create(ctx context.Context, issue *CriticalIssue) (*CriticalIssue, error) {
	if issue.Status == "" {
		issue.Status = "open"
	}
	query := `INSERT INTO reviews.critical_issues (reading_session_id, issue_type, severity, file_path, line_number, description, suggested_fix, status) 
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
             RETURNING id, created_at`
	err := r.DB.QueryRowContext(ctx, query,
		issue.ReadingSessionID, issue.IssueType, issue.Severity, issue.FilePath, issue.LineNumber, issue.Description, issue.SuggestedFix, issue.Status,
	).Scan(&issue.ID, &issue.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("db: failed to create critical issue: %w", err)
	}
	return issue, nil
}

// GetByID retrieves a critical issue by ID.
func (r *CriticalIssuesRepository) GetByID(ctx context.Context, id int64) (*CriticalIssue, error) {
	row := r.DB.QueryRowContext(ctx,
		`SELECT id, reading_session_id, issue_type, severity, file_path, line_number, description, suggested_fix, status, created_at 
         FROM reviews.critical_issues WHERE id = $1`, id)
	var issue CriticalIssue
	err := row.Scan(&issue.ID, &issue.ReadingSessionID, &issue.IssueType, &issue.Severity, &issue.FilePath, &issue.LineNumber, &issue.Description, &issue.SuggestedFix, &issue.Status, &issue.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("db: failed to get critical issue: %w", err)
	}
	return &issue, nil
}

// GetByReadingSessionID retrieves issues for a reading session.
func (r *CriticalIssuesRepository) GetByReadingSessionID(ctx context.Context, sessionID int64) ([]*CriticalIssue, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, reading_session_id, issue_type, severity, file_path, line_number, description, suggested_fix, status, created_at 
         FROM reviews.critical_issues WHERE reading_session_id = $1 ORDER BY severity DESC, created_at DESC`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("db: failed to query critical issues: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("warning: failed to close rows: %v\n", closeErr)
		}
	}()

	var issues []*CriticalIssue
	for rows.Next() {
		var issue CriticalIssue
		if err := rows.Scan(&issue.ID, &issue.ReadingSessionID, &issue.IssueType, &issue.Severity, &issue.FilePath, &issue.LineNumber, &issue.Description, &issue.SuggestedFix, &issue.Status, &issue.CreatedAt); err != nil {
			return nil, fmt.Errorf("db: failed to scan critical issue: %w", err)
		}
		issues = append(issues, &issue)
	}
	return issues, nil
}

// Update updates an existing critical issue.
func (r *CriticalIssuesRepository) Update(ctx context.Context, issue *CriticalIssue) error {
	query := `UPDATE reviews.critical_issues SET issue_type = $1, severity = $2, file_path = $3, line_number = $4, description = $5, suggested_fix = $6, status = $7 WHERE id = $8`
	result, err := r.DB.ExecContext(ctx, query,
		issue.IssueType, issue.Severity, issue.FilePath, issue.LineNumber, issue.Description, issue.SuggestedFix, issue.Status, issue.ID,
	)
	if err != nil {
		return fmt.Errorf("db: failed to update critical issue: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return fmt.Errorf("db: critical issue not found")
	}
	return nil
}

// Delete deletes a critical issue by ID.
func (r *CriticalIssuesRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.DB.ExecContext(ctx, `DELETE FROM reviews.critical_issues WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("db: failed to delete critical issue: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return fmt.Errorf("db: critical issue not found")
	}
	return nil
}
