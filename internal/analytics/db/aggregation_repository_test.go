package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/stretchr/testify/assert"
)

func TestNewAggregationRepository(t *testing.T) {
	// Create a nil pool for testing constructor
	var pool *pgxpool.Pool

	repo := NewAggregationRepository(pool)

	assert.NotNil(t, repo)
	assert.Equal(t, pool, repo.db)
}

func TestAggregationRepository_Methods(t *testing.T) {
	// Test that repository can be created and has expected fields
	repo := NewAggregationRepository(nil)

	assert.NotNil(t, repo)
	assert.Nil(t, repo.db)

	// Verify methods exist by referencing them (don't call as they would panic with nil DB)
	_ = repo.Upsert
	_ = repo.FindByRange
	_ = repo.FindAllServices

	// Suppress unused imports
	_ = context.Background()
	_ = time.Now()
	_ = models.ErrorFrequency
}

func TestAggregationRepositoryInterface(t *testing.T) {
	// Verify that AggregationRepository implements the interface
	var _ AggregationRepositoryInterface = (*AggregationRepository)(nil)
}
