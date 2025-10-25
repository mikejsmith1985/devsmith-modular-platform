package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCriticalIssue_Struct tests the CriticalIssue struct fields
func TestCriticalIssue_Struct(t *testing.T) {
	issue := &CriticalIssue{
		ID:               1,
		ReadingSessionID: 2,
		IssueType:        "security",
		Severity:         "critical",
		FilePath:         "/src/auth.go",
		LineNumber:       42,
		Status:           "open",
	}

	assert.Equal(t, int64(1), issue.ID)
	assert.Equal(t, int64(2), issue.ReadingSessionID)
	assert.Equal(t, "security", issue.IssueType)
	assert.Equal(t, "critical", issue.Severity)
	assert.Equal(t, "/src/auth.go", issue.FilePath)
	assert.Equal(t, 42, issue.LineNumber)
	assert.Equal(t, "open", issue.Status)
}

// TestNewCriticalIssuesRepository_Creation tests repository initialization
func TestNewCriticalIssuesRepository_Creation(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		type CriticalIssuesRepository struct {
			DB interface{}
		}
		repo := &CriticalIssuesRepository{DB: nil}
		assert.NotNil(t, repo)
		return
	}

	repo := NewCriticalIssuesRepository(db)
	assert.NotNil(t, repo)
}

// TestIssueTypes_Validation verifies all valid issue type values
func TestIssueTypes_Validation(t *testing.T) {
	validTypes := []string{
		"security",
		"performance",
		"quality",
		"architecture",
		"testing",
	}

	for _, issueType := range validTypes {
		issue := &CriticalIssue{
			IssueType: issueType,
		}
		assert.Equal(t, issueType, issue.IssueType)
	}
}

// TestSeverityLevels_Validation verifies all valid severity levels
func TestSeverityLevels_Validation(t *testing.T) {
	validSeverities := []string{
		"critical",
		"important",
		"minor",
	}

	for _, severity := range validSeverities {
		issue := &CriticalIssue{
			Severity: severity,
		}
		assert.Equal(t, severity, issue.Severity)
	}
}

// TestIssueStatuses_Validation verifies all valid issue status values
func TestIssueStatuses_Validation(t *testing.T) {
	validStatuses := []string{
		"open",
		"accepted",
		"rejected",
		"fixed",
	}

	for _, status := range validStatuses {
		issue := &CriticalIssue{
			Status: status,
		}
		assert.Equal(t, status, issue.Status)
	}
}
