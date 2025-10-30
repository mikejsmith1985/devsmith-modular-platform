package analytics_db

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestNewLogReader(t *testing.T) {
	var pool *pgxpool.Pool

	reader := NewLogReader(pool)

	assert.NotNil(t, reader)
	assert.Equal(t, pool, reader.db)
}

func TestLogReaderInterface(t *testing.T) {
	var _ LogReaderInterface = (*LogReader)(nil)
}
