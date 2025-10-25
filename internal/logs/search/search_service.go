package search

import (
	"context"
	"time"
)

// Service handles search operations on logs.
type Service struct {
	parser *QueryParser
	store  Store
}

// Store defines the interface for log storage.
type Store interface {
	SearchLogs(ctx context.Context, query *ParsedQuery) ([]*LogEntry, error)
	SaveLog(ctx context.Context, entry *LogEntry) error
	DeleteLog(ctx context.Context, id int64) error
}

// LogEntry represents a single log entry.
// nolint:govet // fieldalignment: timestamp field ordering necessary for domain model clarity
type LogEntry struct {
	Timestamp time.Time
	ID        int64
	Service   string
	Level     string
	Message   string
}

// Error represents a search operation error.
type Error struct {
	Code    string
	Message string
}

// NewError creates a new search error.
func NewError(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

// NewService creates a new search service.
func NewService(parser *QueryParser, store Store) *Service {
	return &Service{
		parser: parser,
		store:  store,
	}
}

// Search performs a query string search without accessing the database.
func (s *Service) Search(ctx context.Context, queryString string) ([]*LogEntry, error) {
	parsed := s.parser.Parse(queryString)
	if !parsed.IsValid {
		return nil, NewError("invalid query", parsed.ErrorMsg)
	}

	// Return results from store
	return s.store.SearchLogs(ctx, parsed)
}
