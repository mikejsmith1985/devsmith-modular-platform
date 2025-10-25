package db

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// SAVE METHOD TESTS - CYCLE 1
// ============================================================================

func TestLogRepository_Save_Success(t *testing.T) {
	tests := []struct {
		name      string
		entry     *LogEntry
		wantID    int64
		wantError bool
	}{
		{
			name: "valid entry returns ID",
			entry: &LogEntry{
				CreatedAt: time.Now(),
				Level:     "error",
				Message:   "Database connection failed",
				Service:   "db_service",
				Metadata:  map[string]interface{}{"user_id": 123},
			},
			wantID:    1,
			wantError: false,
		},
		{
			name: "minimal valid entry",
			entry: &LogEntry{
				CreatedAt: time.Now(),
				Level:     "info",
				Message:   "Test",
				Service:   "test",
			},
			wantID:    2,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &LogRepository{}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			got, err := repo.Save(ctx, tt.entry)

			if (err != nil) != tt.wantError {
				t.Errorf("Save() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError && got <= 0 {
				t.Errorf("Save() returned invalid ID %d", got)
			}
		})
	}
}

func TestLogRepository_Save_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		entry   *LogEntry
		wantErr bool
	}{
		{"nil entry", nil, true},
		{"empty message", &LogEntry{CreatedAt: time.Now(), Level: "info", Service: "test", Message: ""}, true},
		{"empty level", &LogEntry{CreatedAt: time.Now(), Level: "", Service: "test", Message: "msg"}, true},
		{"empty service", &LogEntry{CreatedAt: time.Now(), Level: "info", Service: "", Message: "msg"}, true},
		{"zero timestamp", &LogEntry{CreatedAt: time.Time{}, Level: "info", Service: "test", Message: "msg"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &LogRepository{}
			_, err := repo.Save(context.Background(), tt.entry)

			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLogRepository_Save_ContextCancellation(t *testing.T) {
	repo := &LogRepository{}
	entry := &LogEntry{
		CreatedAt: time.Now(),
		Level:     "info",
		Service:   "test",
		Message:   "test",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.Save(ctx, entry)
	if err == nil {
		t.Error("Save() should error on cancelled context")
	}
}
