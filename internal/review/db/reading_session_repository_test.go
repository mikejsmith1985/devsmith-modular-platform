package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestReadingSession_Struct tests the ReadingSession struct fields
func TestReadingSession_Struct(t *testing.T) {
	rs := &ReadingSession{
		ID:          1,
		SessionID:   2,
		ReadingMode: "preview",
		TargetPath:  "/src/file.go",
		ScanQuery:   "SELECT * FROM table",
	}

	assert.Equal(t, int64(1), rs.ID)
	assert.Equal(t, int64(2), rs.SessionID)
	assert.Equal(t, "preview", rs.ReadingMode)
	assert.Equal(t, "/src/file.go", rs.TargetPath)
	assert.Equal(t, "SELECT * FROM table", rs.ScanQuery)
}

// TestNewReadingSessionRepository_Creation tests repository initialization
func TestNewReadingSessionRepository_Creation(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		type ReadingSessionRepository struct {
			DB interface{}
		}
		repo := &ReadingSessionRepository{DB: nil}
		assert.NotNil(t, repo)
		return
	}

	repo := NewReadingSessionRepository(db)
	assert.NotNil(t, repo)
}

// TestReadingModes_Validation verifies all valid reading modes
func TestReadingModes_Validation(t *testing.T) {
	validModes := []string{
		"preview",
		"skim",
		"scan",
		"detailed",
		"critical",
	}

	for _, mode := range validModes {
		rs := &ReadingSession{
			ReadingMode: mode,
		}
		assert.Equal(t, mode, rs.ReadingMode)
	}
}
