// Package testutils provides mock implementations for testing analytics services.
package testutils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/stretchr/testify/mock"
)

// MockLogReader provides a mock implementation of the LogReader interface.
// It is used for testing purposes.
type MockLogReader struct {
	mock.Mock
}

// CountByServiceAndLevel provides a mocked implementation with logging.
// It simulates the behavior of the actual LogReader method.
func (m *MockLogReader) CountByServiceAndLevel(ctx context.Context, service, level string, startTime, endTime time.Time) (int, error) {
	log.Printf("MockLogReader.CountByServiceAndLevel called with service=%s, level=%s, startTime=%v, endTime=%v", service, level, startTime, endTime)
	args := m.Called(ctx, service, level, startTime, endTime)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	return args.Int(0), args.Error(1)
}

// FindLogs provides a mocked implementation with logging.
// It simulates the behavior of the actual LogReader method.
func (m *MockLogReader) FindLogs(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.LogEntry, error) {
	log.Printf("MockLogReader.FindLogs called with filters=%v, limit=%d, offset=%d", filters, limit, offset)
	args := m.Called(ctx, filters, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	entries, ok := args.Get(0).([]*models.LogEntry)
	if !ok {
		return nil, fmt.Errorf("unexpected type for log entries: %T", args.Get(0))
	}
	return entries, nil
}

// FindAllServices provides a mocked implementation with logging.
// It simulates the behavior of the actual LogReader method.
func (m *MockLogReader) FindAllServices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	result, ok := args.Get(0).([]string)
	if !ok {
		return nil, fmt.Errorf("unexpected type for result: %T", args.Get(0))
	}
	return result, nil
}

// FindTopMessages provides a mocked implementation with logging.
// It simulates the behavior of the actual LogReader method.
// Ensure error return values are checked and handled appropriately.
func (m *MockLogReader) FindTopMessages(ctx context.Context, service, level string, start, end time.Time, limit int) ([]models.IssueItem, error) {
	log.Printf("MockLogReader.FindTopMessages called with service=%s, level=%s, startTime=%v, endTime=%v, limit=%d", service, level, start, end, limit)
	args := m.Called(ctx, service, level, start, end, limit)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	if result, ok := args.Get(0).([]models.IssueItem); ok {
		return result, nil
	}
	return nil, args.Error(1)
}

// FindAggregations provides a mocked implementation with logging.
// It simulates the behavior of the actual LogReader method.
func (m *MockLogReader) FindAggregations(ctx context.Context, service, level string, startTime, endTime time.Time) ([]*models.Aggregation, error) {
	log.Printf("MockLogReader.FindAggregations called with service=%s, level=%s, startTime=%v, endTime=%v", service, level, startTime, endTime)
	args := m.Called(ctx, service, level, startTime, endTime)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	if result, ok := args.Get(0).([]*models.Aggregation); ok {
		return result, nil
	}
	return nil, args.Error(1)
}

// FindByRange provides a mocked implementation with logging.
// It simulates the behavior of the actual LogReader method.
// Ensure error return values are checked and handled appropriately.
func (m *MockLogReader) FindByRange(ctx context.Context, metricType, service string, start, end time.Time) ([]models.IssueItem, error) {
	log.Printf("MockLogReader.FindByRange called with metricType=%s, service=%s, start=%v, end=%v", metricType, service, start, end)
	args := m.Called(ctx, metricType, service, start, end)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	if result, ok := args.Get(0).([]models.IssueItem); ok {
		return result, nil
	}
	return nil, args.Error(1)
}
