package testutils

import (
	"context"
	"log"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/stretchr/testify/mock"
)

type MockLogReader struct {
	mock.Mock
}

// CountByServiceAndLevel provides a mocked implementation with logging
func (m *MockLogReader) CountByServiceAndLevel(ctx context.Context, service, level string, startTime, endTime time.Time) (int, error) {
	log.Printf("MockLogReader.CountByServiceAndLevel called with service=%s, level=%s, startTime=%v, endTime=%v", service, level, startTime, endTime)
	args := m.Called(ctx, service, level, startTime, endTime)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	return args.Int(0), args.Error(1)
}

// FindLogs provides a mocked implementation with logging
func (m *MockLogReader) FindLogs(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.LogEntry, error) {
	log.Printf("MockLogReader.FindLogs called with filters=%v, limit=%d, offset=%d", filters, limit, offset)
	args := m.Called(ctx, filters, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.LogEntry), args.Error(1)
}

// FindAllServices provides a mocked implementation with logging
func (m *MockLogReader) FindAllServices(ctx context.Context) ([]string, error) {
	log.Printf("MockLogReader.FindAllServices called")
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// FindTopMessages provides a mocked implementation with logging
func (m *MockLogReader) FindTopMessages(ctx context.Context, service string, level string, startTime time.Time, endTime time.Time, limit int) ([]models.IssueItem, error) {
	log.Printf("MockLogReader.FindTopMessages called with service=%s, level=%s, startTime=%v, endTime=%v, limit=%d", service, level, startTime, endTime, limit)
	args := m.Called(ctx, service, level, startTime, endTime, limit)
	// Add log to verify the type of Count in mock data
	if args.Get(0) != nil {
		mockData := args.Get(0).([]models.IssueItem)
		for _, item := range mockData {
			log.Printf("Mock Data Item: %+v, Count Type: %T", item, item.Count)
		}
	}
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.IssueItem), args.Error(1)
}

// FindTopByMetric provides a mocked implementation with logging
func (m *MockLogReader) FindTopByMetric(ctx context.Context, metric string, limit int) ([]models.Aggregation, error) {
	log.Printf("MockLogReader.FindTopByMetric called with metric=%s, limit=%d", metric, limit)
	args := m.Called(ctx, metric, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Aggregation), args.Error(1)
}
