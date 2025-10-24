package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestNewLogReader(t *testing.T) {
	var pool *pgxpool.Pool

	reader := NewLogReader(pool)

	assert.NotNil(t, reader)
	assert.Equal(t, pool, reader.db)
}

func TestLogReader_Methods(t *testing.T) {
	// Test that log reader can be created
	reader := NewLogReader(nil)

	assert.NotNil(t, reader)
	assert.Nil(t, reader.db)

	// Verify methods exist (don't call as they would panic with nil DB)
	_ = reader.FindTopMessages
	_ = reader.FindAllServices
	_ = reader.CountByServiceAndLevel

	// Suppress unused imports
	_ = context.Background()
	_ = time.Now()
}

func TestLogReaderInterface(t *testing.T) {
	// Verify that LogReader implements the interface
	var _ LogReaderInterface = (*LogReader)(nil)
}
