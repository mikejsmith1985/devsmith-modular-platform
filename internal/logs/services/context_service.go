package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// ContextService manages correlation context (GREEN phase - full implementation)
type ContextService struct {
	repo *db.ContextRepository
}

// NewContextService creates a new context service (GREEN phase - full implementation)
func NewContextService(repo *db.ContextRepository) *ContextService {
	return &ContextService{repo: repo}
}

// GenerateCorrelationID creates a new unique correlation ID (GREEN phase)
func (s *ContextService) GenerateCorrelationID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if random fails
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(b)
}

// EnrichContext adds automatic metadata to context (GREEN phase)
func (s *ContextService) EnrichContext(
	ctx *models.CorrelationContext,
) *models.CorrelationContext {
	if ctx == nil {
		ctx = &models.CorrelationContext{}
	}

	// Generate correlation ID if missing
	if ctx.CorrelationID == "" {
		ctx.CorrelationID = s.GenerateCorrelationID()
	}

	// Add automatic enrichment
	if ctx.Hostname == "" {
		if host, err := os.Hostname(); err == nil {
			ctx.Hostname = host
		}
	}

	if ctx.Environment == "" {
		ctx.Environment = os.Getenv("ENVIRONMENT")
		if ctx.Environment == "" {
			ctx.Environment = "development"
		}
	}

	if ctx.Version == "" {
		ctx.Version = os.Getenv("SERVICE_VERSION")
		if ctx.Version == "" {
			ctx.Version = "dev"
		}
	}

	// Set timestamps
	now := time.Now()
	if ctx.CreatedAt.IsZero() {
		ctx.CreatedAt = now
	}
	ctx.UpdatedAt = now

	return ctx
}

// GetCorrelatedLogs retrieves all logs for a correlation ID (GREEN phase)
func (s *ContextService) GetCorrelatedLogs(
	ctx context.Context,
	correlationID string,
	limit, offset int,
) ([]models.LogEntry, error) {
	// Validate limit (max 1000, default 50)
	if limit > 1000 {
		limit = 1000
	}
	if limit <= 0 {
		limit = 50
	}

	if s.repo == nil {
		return nil, nil
	}

	return s.repo.GetCorrelatedLogs(ctx, correlationID, limit, offset)
}

// GetCorrelationMetadata returns summary of a correlation (GREEN phase)
func (s *ContextService) GetCorrelationMetadata(
	ctx context.Context,
	correlationID string,
) (map[string]interface{}, error) {
	if s.repo == nil {
		return make(map[string]interface{}), nil
	}

	count, err := s.repo.GetCorrelationCount(ctx, correlationID)
	if err != nil {
		return nil, err
	}

	metadata, err := s.repo.GetContextMetadata(ctx, correlationID)
	if err != nil {
		return nil, err
	}

	metadata["total_logs"] = count
	metadata["correlation_id"] = correlationID

	return metadata, nil
}

// GetTraceTimeline returns timeline of events for a correlation (GREEN phase)
func (s *ContextService) GetTraceTimeline(
	ctx context.Context,
	correlationID string,
) ([]map[string]interface{}, error) {
	if s.repo == nil {
		return nil, nil
	}

	logs, err := s.repo.GetCorrelatedLogs(ctx, correlationID, 1000, 0)
	if err != nil {
		return nil, err
	}

	timeline := make([]map[string]interface{}, 0, len(logs))
	for i := range logs {
		log := &logs[i]
		entry := map[string]interface{}{
			"timestamp": log.Timestamp,
			"level":     log.Level,
			"service":   log.Service,
			"message":   log.Message,
		}
		if log.Context != nil {
			entry["trace_id"] = log.Context.TraceID
			entry["span_id"] = log.Context.SpanID
		}
		timeline = append(timeline, entry)
	}

	return timeline, nil
}
