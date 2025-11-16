package logs_services

import (
	"context"
	"encoding/json"

	logs_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// LogRepositoryAdapter adapts logs_db.LogRepository to logs_services.LogRepository
type LogRepositoryAdapter struct {
	repo *logs_db.LogRepository
}

// NewLogRepositoryAdapter creates a new adapter
func NewLogRepositoryAdapter(repo *logs_db.LogRepository) LogRepository {
	return &LogRepositoryAdapter{repo: repo}
}

// GetByID adapts the repository GetByID method
func (a *LogRepositoryAdapter) GetByID(ctx context.Context, id int64) (*logs_models.LogEntry, error) {
	// Call the original repository
	dbEntry, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert metadata map to JSON bytes
	metadataBytes, err := json.Marshal(dbEntry.Metadata)
	if err != nil {
		metadataBytes = []byte("{}")
	}

	// Convert logs_db.LogEntry to logs_models.LogEntry
	return &logs_models.LogEntry{
		ID:        dbEntry.ID,
		Timestamp: dbEntry.CreatedAt, // Use CreatedAt as Timestamp
		Level:     dbEntry.Level,
		Service:   dbEntry.Service,
		Message:   dbEntry.Message,
		Metadata:  metadataBytes,
		CreatedAt: dbEntry.CreatedAt,
	}, nil
}
