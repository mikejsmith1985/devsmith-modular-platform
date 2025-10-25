// Package service provides business logic for log operations.
package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockRepository struct {
	InsertFn       func(ctx context.Context, entry interface{}) (int64, error)
	QueryFn        func(ctx context.Context, filters, page interface{}) ([]interface{}, error)
	GetByIDFn      func(ctx context.Context, id int64) (interface{}, error)
	DeleteByIDFn   func(ctx context.Context, id int64) error
	DeleteBeforeFn func(ctx context.Context, ts interface{}) (int64, error)
}

func (m *MockRepository) Insert(ctx context.Context, entry interface{}) (int64, error) {
	if m.InsertFn != nil {
		return m.InsertFn(ctx, entry)
	}
	return 1, nil
}

func (m *MockRepository) Query(ctx context.Context, filters, page interface{}) ([]interface{}, error) {
	if m.QueryFn != nil {
		return m.QueryFn(ctx, filters, page)
	}
	return []interface{}{}, nil
}

func (m *MockRepository) GetByID(ctx context.Context, id int64) (interface{}, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) DeleteByID(ctx context.Context, id int64) error {
	if m.DeleteByIDFn != nil {
		return m.DeleteByIDFn(ctx, id)
	}
	return nil
}

func (m *MockRepository) DeleteBefore(ctx context.Context, ts interface{}) (int64, error) {
	if m.DeleteBeforeFn != nil {
		return m.DeleteBeforeFn(ctx, ts)
	}
	return 0, nil
}

func TestNewLogService(t *testing.T) {
	repo := &MockRepository{}
	svc := NewLogService(repo)
	assert.NotNil(t, svc)
}

func TestInsert_Valid(t *testing.T) {
	repo := &MockRepository{
		InsertFn: func(ctx context.Context, entry interface{}) (int64, error) {
			return 42, nil
		},
	}
	svc := NewLogService(repo)

	entry := map[string]interface{}{
		"service": "portal",
		"level":   "info",
		"message": "test",
	}

	id, err := svc.Insert(context.Background(), entry)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), id)
}

func TestQuery_Valid(t *testing.T) {
	repo := &MockRepository{
		QueryFn: func(ctx context.Context, filters interface{}, page interface{}) ([]interface{}, error) {
			return []interface{}{
				map[string]interface{}{"id": 1, "service": "portal"},
			}, nil
		},
	}
	svc := NewLogService(repo)

	filters := map[string]interface{}{"service": "portal"}
	page := map[string]int{"limit": 10, "offset": 0}

	results, err := svc.Query(context.Background(), filters, page)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestGetByID_Valid(t *testing.T) {
	repo := &MockRepository{
		GetByIDFn: func(ctx context.Context, id int64) (interface{}, error) {
			return map[string]interface{}{"id": id, "service": "portal"}, nil
		},
	}
	svc := NewLogService(repo)

	entry, err := svc.GetByID(context.Background(), 42)
	assert.NoError(t, err)
	assert.NotNil(t, entry)
}

func TestStats_Valid(t *testing.T) {
	svc := NewLogService(&MockRepository{})
	stats, err := svc.Stats(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestDeleteByID_Valid(t *testing.T) {
	repo := &MockRepository{
		DeleteByIDFn: func(ctx context.Context, id int64) error {
			return nil
		},
	}
	svc := NewLogService(repo)

	err := svc.DeleteByID(context.Background(), 42)
	assert.NoError(t, err)
}

func TestDelete_Valid(t *testing.T) {
	repo := &MockRepository{
		DeleteBeforeFn: func(ctx context.Context, ts interface{}) (int64, error) {
			return 25, nil
		},
	}
	svc := NewLogService(repo)

	count, err := svc.Delete(context.Background(), map[string]interface{}{"before": "2025-01-01"})
	assert.NoError(t, err)
	assert.Equal(t, int64(25), count)
}
