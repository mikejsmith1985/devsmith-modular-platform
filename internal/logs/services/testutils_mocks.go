// Package services contains business logic for the logs service.
package services

import (
	"context"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/mock"
)

// MockContextRepository is a mock for ContextRepository used in testing
type MockContextRepository struct {
	mock.Mock
}

// GetCorrelatedLogs returns correlated logs with testify mock framework
func (m *MockContextRepository) GetCorrelatedLogs(
	ctx context.Context,
	correlationID string,
	limit, offset int,
) ([]models.LogEntry, error) {
	args := m.Called(ctx, correlationID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	//nolint:errcheck // type assertion error not needed in mock
	return args.Get(0).([]models.LogEntry), args.Error(1)
}

// GetCorrelationCount returns count with testify mock framework
func (m *MockContextRepository) GetCorrelationCount(
	ctx context.Context,
	correlationID string,
) (int, error) {
	args := m.Called(ctx, correlationID)
	return args.Int(0), args.Error(1)
}

// GetRecentCorrelations returns correlation IDs with testify mock framework
func (m *MockContextRepository) GetRecentCorrelations(
	ctx context.Context,
	minutes, limit int,
) ([]string, error) {
	args := m.Called(ctx, minutes, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	//nolint:errcheck // type assertion error not needed in mock
	return args.Get(0).([]string), args.Error(1)
}

// GetContextMetadata returns metadata with testify mock framework
func (m *MockContextRepository) GetContextMetadata(
	ctx context.Context,
	correlationID string,
) (map[string]interface{}, error) {
	args := m.Called(ctx, correlationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	//nolint:errcheck // type assertion error not needed in mock
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// NewMockContextRepository creates a new mock context repository
func NewMockContextRepository() *MockContextRepository {
	return &MockContextRepository{}
}
