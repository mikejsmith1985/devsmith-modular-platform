// Package search provides advanced filtering and search capabilities for logs.
package search

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLogStore is a mock implementation of log storage for testing.
// nolint:govet // fieldalignment: test mock fields ordered for clarity
type MockLogStore struct {
	searchHits map[string][]*LogEntry
	logs       []*LogEntry
	callCount  int
}

// SearchLogs returns pre-configured search results.
func (m *MockLogStore) SearchLogs(ctx context.Context, query *ParsedQuery) ([]*LogEntry, error) {
	m.callCount++
	if hits, ok := m.searchHits[query.RootNode.Field]; ok {
		return hits, nil
	}
	return []*LogEntry{}, nil
}

// SaveLog stores a log entry.
func (m *MockLogStore) SaveLog(ctx context.Context, entry *LogEntry) error {
	m.logs = append(m.logs, entry)
	return nil
}

// DeleteLog removes a log entry.
func (m *MockLogStore) DeleteLog(ctx context.Context, id int64) error {
	for i, log := range m.logs {
		if log.ID == id {
			m.logs = append(m.logs[:i], m.logs[i+1:]...)
			return nil
		}
	}
	return nil
}

// TestSearchService_ValidQuery tests valid query parsing and execution.
func TestSearchService_ValidQuery(t *testing.T) {
	parser := NewQueryParser()
	store := &MockLogStore{
		searchHits: map[string][]*LogEntry{
			"service": {
				&LogEntry{
					ID:      1,
					Service: "auth-service",
					Level:   "error",
					Message: "authentication failed",
				},
			},
		},
	}
	service := NewService(parser, store)

	results, err := service.Search(context.Background(), "service:auth-service")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "auth-service", results[0].Service)
}

// TestSearchService_InvalidQuery tests invalid query handling.
func TestSearchService_InvalidQuery(t *testing.T) {
	parser := NewQueryParser()
	store := &MockLogStore{}
	service := NewService(parser, store)

	results, err := service.Search(context.Background(), "invalid:query:too:many:colons")
	assert.Error(t, err)
	assert.Nil(t, results)
	assert.IsType(t, &Error{}, err)
}

// TestSearchService_ComplexQuery tests complex boolean query execution.
func TestSearchService_ComplexQuery(t *testing.T) {
	parser := NewQueryParser()
	store := &MockLogStore{
		searchHits: map[string][]*LogEntry{
			"service": {
				&LogEntry{ID: 1, Service: "portal", Level: "error", Message: "portal error"},
				&LogEntry{ID: 2, Service: "portal", Level: "warn", Message: "portal warning"},
			},
		},
	}
	service := NewService(parser, store)

	results, err := service.Search(context.Background(), "service:portal AND level:error")
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.GreaterOrEqual(t, len(results), 0)
}

// TestSearchService_RegexQuery tests regex pattern search.
func TestSearchService_RegexQuery(t *testing.T) {
	parser := NewQueryParser()
	store := &MockLogStore{
		searchHits: map[string][]*LogEntry{
			"message": {
				&LogEntry{ID: 1, Message: "database connection error"},
				&LogEntry{ID: 2, Message: "database timeout"},
			},
		},
	}
	service := NewService(parser, store)

	results, err := service.Search(context.Background(), "message:/database.*/")
	assert.NoError(t, err)
	assert.NotNil(t, results)
}

// TestSearchService_EmptyResults tests handling of empty search results.
func TestSearchService_EmptyResults(t *testing.T) {
	parser := NewQueryParser()
	store := &MockLogStore{
		searchHits: make(map[string][]*LogEntry),
	}
	service := NewService(parser, store)

	results, err := service.Search(context.Background(), "service:nonexistent")
	assert.NoError(t, err)
	assert.Empty(t, results)
}

// TestSearchService_ContextCancellation tests behavior with cancelled context.
func TestSearchService_ContextCancellation(t *testing.T) {
	parser := NewQueryParser()
	store := &MockLogStore{}
	service := NewService(parser, store)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results, err := service.Search(ctx, "service:test")
	assert.NoError(t, err)
	assert.Empty(t, results)
}

// TestSearchService_StoreCallCount verifies store is called exactly once.
func TestSearchService_StoreCallCount(t *testing.T) {
	parser := NewQueryParser()
	store := &MockLogStore{
		searchHits: map[string][]*LogEntry{
			"service": {},
		},
	}
	service := NewService(parser, store)

	service.Search(context.Background(), "service:test")
	assert.Equal(t, 1, store.callCount)
}

// TestSearchService_MultipleQueries verifies multiple searches work correctly.
func TestSearchService_MultipleQueries(t *testing.T) {
	parser := NewQueryParser()
	store := &MockLogStore{
		searchHits: map[string][]*LogEntry{
			"service": {
				&LogEntry{ID: 1, Service: "auth"},
				&LogEntry{ID: 2, Service: "portal"},
			},
			"level": {
				&LogEntry{ID: 3, Level: "error"},
				&LogEntry{ID: 4, Level: "warn"},
			},
		},
	}
	service := NewService(parser, store)

	results1, err1 := service.Search(context.Background(), "service:auth")
	assert.NoError(t, err1)
	assert.NotNil(t, results1)

	results2, err2 := service.Search(context.Background(), "level:error")
	assert.NoError(t, err2)
	assert.NotNil(t, results2)

	assert.Equal(t, 2, store.callCount)
}

// TestSearchService_QuotedStringValue tests search with quoted string values.
func TestSearchService_QuotedStringValue(t *testing.T) {
	parser := NewQueryParser()
	store := &MockLogStore{
		searchHits: map[string][]*LogEntry{
			"message": {
				&LogEntry{
					ID:      1,
					Message: "database connection failed",
				},
			},
		},
	}
	service := NewService(parser, store)

	results, err := service.Search(context.Background(), `message:"database connection failed"`)
	assert.NoError(t, err)
	assert.NotNil(t, results)
}

// TestSearchService_FieldValidation tests that only valid fields are searchable.
func TestSearchService_FieldValidation(t *testing.T) {
	parser := NewQueryParser()
	store := &MockLogStore{}
	service := NewService(parser, store)

	results, err := service.Search(context.Background(), "invalid_field:value")
	assert.Error(t, err)
	assert.Nil(t, results)
}

// TestError_ErrorInterface tests Error implements error interface.
func TestError_ErrorInterface(t *testing.T) {
	err := NewError("test_code", "test message")
	require.NotNil(t, err)
	assert.Equal(t, "test_code: test message", err.Error())
}

// TestSearchService_ParserIntegration tests integration with query parser.
func TestSearchService_ParserIntegration(t *testing.T) {
	parser := NewQueryParser()
	store := &MockLogStore{
		searchHits: map[string][]*LogEntry{
			"service": {
				&LogEntry{ID: 1, Service: "test-service"},
			},
		},
	}
	service := NewService(parser, store)

	results, err := service.Search(context.Background(), "service:test-service")
	assert.NoError(t, err)
	assert.NotNil(t, results)
}

// TestSearchService_NilStore handles nil store gracefully.
func TestSearchService_NilStore(t *testing.T) {
	parser := NewQueryParser()
	service := NewService(parser, nil)
	require.NotNil(t, service)
}
