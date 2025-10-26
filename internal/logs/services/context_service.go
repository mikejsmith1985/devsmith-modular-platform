// Package services provides business logic for the logs service.
// This package handles correlation context management for distributed tracing.
package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// Constants for context service configuration
const (
	// CorrelationIDLength is the length of generated correlation IDs in bytes (32 hex chars = 16 bytes)
	CorrelationIDLength = 16

	// DefaultLogLimit is the default pagination limit for log queries
	DefaultLogLimit = 50

	// MaxLogLimit is the maximum allowed pagination limit
	MaxLogLimit = 1000

	// DefaultContextTimeoutMinutes is the default window for recent correlations
	DefaultContextTimeoutMinutes = 5

	// DefaultEnvironment is used when ENVIRONMENT variable is not set
	DefaultEnvironment = "development"

	// DefaultVersion is used when SERVICE_VERSION variable is not set
	DefaultVersion = "dev"
)

// ContextService manages correlation context for distributed tracing.
// It handles ID generation, context enrichment, and log retrieval for correlated requests.
type ContextService struct {
	repo *db.ContextRepository
}

// NewContextService creates a new context service with the given repository.
// The repository handles all database operations for correlation context.
//
// Example:
//
//	repo := db.NewContextRepository(sqlDB)
//	svc := NewContextService(repo)
//	correlationID := svc.GenerateCorrelationID()
func NewContextService(repo *db.ContextRepository) *ContextService {
	return &ContextService{repo: repo}
}

// GenerateCorrelationID creates a new unique correlation ID for request tracing.
// Returns a 32-character hexadecimal string (16 random bytes encoded).
// Each call generates a unique ID suitable for distributed system tracing.
//
// Example:
//
//	id := svc.GenerateCorrelationID() // "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
func (s *ContextService) GenerateCorrelationID() string {
	bytes := make([]byte, CorrelationIDLength)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("%016x", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// EnrichContext adds automatic metadata to a correlation context.
// Populates fields like hostname, environment, version, and timestamps.
// If context is nil, creates a new one. If fields are already set, preserves them.
//
// Fields populated:
// - CorrelationID: Generated if not present
// - Hostname: System hostname
// - Environment: From ENVIRONMENT env var or default
// - Version: From SERVICE_VERSION env var or default
// - CreatedAt: Current time if not set
// - UpdatedAt: Always updated to current time
//
// Example:
//
//	ctx := &models.CorrelationContext{Method: "POST", Path: "/api/logs"}
//	enriched := svc.EnrichContext(ctx)
//	// enriched now has hostname, environment, version, timestamps
func (s *ContextService) EnrichContext(ctx *models.CorrelationContext) *models.CorrelationContext {
	if ctx == nil {
		ctx = &models.CorrelationContext{}
	}

	// Generate ID if not provided
	if ctx.CorrelationID == "" {
		ctx.CorrelationID = s.GenerateCorrelationID()
	}

	// Add hostname if not present
	if ctx.Hostname == "" {
		if hostname, err := os.Hostname(); err == nil {
			ctx.Hostname = hostname
		}
	}

	// Add environment from config
	if ctx.Environment == "" {
		if env := os.Getenv("ENVIRONMENT"); env != "" {
			ctx.Environment = env
		} else {
			ctx.Environment = DefaultEnvironment
		}
	}

	// Add version from config
	if ctx.Version == "" {
		if version := os.Getenv("SERVICE_VERSION"); version != "" {
			ctx.Version = version
		} else {
			ctx.Version = DefaultVersion
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

// GetCorrelatedLogs retrieves all logs associated with a correlation ID.
// Supports pagination with automatic limit validation (capped at MaxLogLimit).
// Returns logs ordered by timestamp descending for chronological viewing.
//
// Parameters:
// - correlationID: Required - the correlation ID to query for
// - limit: Pagination limit (0 uses default, values > MaxLogLimit are capped)
// - offset: Number of results to skip (cannot be negative)
//
// Returns logs ordered by timestamp DESC, then by ID DESC.
//
// Example:
//
//	logs, err := svc.GetCorrelatedLogs(ctx, "abc123", 50, 0)
//	if err != nil {
//	    log.Printf("Query failed: %v", err)
//	}
func (s *ContextService) GetCorrelatedLogs(
	ctx context.Context,
	correlationID string,
	limit, offset int,
) ([]models.LogEntry, error) {
	if s.repo == nil {
		return []models.LogEntry{}, nil
	}

	// Validate and normalize pagination parameters
	if limit <= 0 {
		limit = DefaultLogLimit
	}
	if limit > MaxLogLimit {
		limit = MaxLogLimit
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetCorrelatedLogs(ctx, correlationID, limit, offset)
}

// GetCorrelationMetadata retrieves aggregated metadata for a correlation.
// Returns information about all logs in the correlation including:
// - total_logs: Count of logs in this correlation
// - correlation_id: The correlation ID
// - services: Unique list of services involved
// - trace_ids: Unique list of OpenTelemetry trace IDs
//
// This provides a summary view for distributed tracing analysis.
//
// Example:
//
//	metadata, err := svc.GetCorrelationMetadata(ctx, "abc123")
//	// metadata = {
//	//   "total_logs": 15,
//	//   "correlation_id": "abc123",
//	//   "services": ["portal", "analytics"],
//	//   "trace_ids": ["trace-xyz"]
//	// }
func (s *ContextService) GetCorrelationMetadata(
	ctx context.Context,
	correlationID string,
) (map[string]interface{}, error) {
	if s.repo == nil {
		return make(map[string]interface{}), nil
	}
	return s.repo.GetContextMetadata(ctx, correlationID)
}

// GetTraceTimeline retrieves a chronological timeline of events for a correlation.
// Returns an ordered list of log entries with trace/span information,
// suitable for visualizing request flow through multiple services.
//
// Each entry contains:
// - timestamp: When the event occurred
// - level: Log level (debug, info, warn, error)
// - service: Service that generated the log
// - message: Log message
// - trace_id: OpenTelemetry trace ID
// - span_id: OpenTelemetry span ID
//
// Example:
//
//	timeline, err := svc.GetTraceTimeline(ctx, "abc123")
//	// timeline = [
//	//   {
//	//     "timestamp": "2025-01-01T12:00:00Z",
//	//     "level": "info",
//	//     "service": "portal",
//	//     "message": "Request started",
//	//     "trace_id": "trace-xyz",
//	//     "span_id": "span-1"
//	//   },
//	//   ...
//	// ]
func (s *ContextService) GetTraceTimeline(
	ctx context.Context,
	correlationID string,
) ([]map[string]interface{}, error) {
	if s.repo == nil {
		return nil, nil
	}

	logs, err := s.repo.GetCorrelatedLogs(ctx, correlationID, MaxLogLimit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs for timeline: %w", err)
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
