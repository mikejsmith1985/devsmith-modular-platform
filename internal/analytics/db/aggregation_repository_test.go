package db

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestNewAggregationRepository(t *testing.T) {
	var pool *pgxpool.Pool

	repo := NewAggregationRepository(pool)

	assert.NotNil(t, repo)
	assert.Equal(t, pool, repo.db)
}

func TestAggregationRepositoryInterface(t *testing.T) {
	var _ AggregationRepositoryInterface = (*AggregationRepository)(nil)
}
