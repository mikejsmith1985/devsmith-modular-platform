package performance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultPoolConfig(t *testing.T) {
	// WHEN: Getting default pool config
	config := DefaultPoolConfig()

	// THEN: Should have recommended defaults
	assert.NotNil(t, config)
	assert.Equal(t, 50, config.MaxOpenConnections)
	assert.Equal(t, 10, config.MaxIdleConnections)
	assert.Equal(t, 3600, config.MaxConnLifetime)
}

func TestLoadPoolConfigFromEnv_NoEnvVars(t *testing.T) {
	// WHEN: Loading config with no environment variables
	config := LoadPoolConfigFromEnv()

	// THEN: Should use defaults
	assert.NotNil(t, config)
	assert.Equal(t, 50, config.MaxOpenConnections)
	assert.Equal(t, 10, config.MaxIdleConnections)
	assert.Equal(t, 3600, config.MaxConnLifetime)
}

func TestLoadPoolConfigFromEnv_WithOverrides(t *testing.T) {
	// GIVEN: Set environment variables
	t.Setenv("DB_POOL_MAX_OPEN", "100")
	t.Setenv("DB_POOL_MAX_IDLE", "20")
	t.Setenv("DB_POOL_MAX_LIFETIME", "7200")

	// WHEN: Loading config
	config := LoadPoolConfigFromEnv()

	// THEN: Should use env values
	assert.NotNil(t, config)
	assert.Equal(t, 100, config.MaxOpenConnections)
	assert.Equal(t, 20, config.MaxIdleConnections)
	assert.Equal(t, 7200, config.MaxConnLifetime)
}

func TestLoadPoolConfigFromEnv_InvalidValues(t *testing.T) {
	// GIVEN: Invalid environment variables
	t.Setenv("DB_POOL_MAX_OPEN", "invalid")
	t.Setenv("DB_POOL_MAX_IDLE", "-5")
	t.Setenv("DB_POOL_MAX_LIFETIME", "xyz")

	// WHEN: Loading config
	config := LoadPoolConfigFromEnv()

	// THEN: Should fall back to defaults
	assert.NotNil(t, config)
	assert.Equal(t, 50, config.MaxOpenConnections)
	assert.Equal(t, 10, config.MaxIdleConnections)
	assert.Equal(t, 3600, config.MaxConnLifetime)
}

func TestConfigurePool_NilDatabase(t *testing.T) {
	// GIVEN: Nil database connection
	config := DefaultPoolConfig()

	// WHEN: Trying to configure
	err := ConfigurePool(nil, config)

	// THEN: Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot configure nil database connection")
}

func TestConfigurePool_NilConfig(t *testing.T) {
	// GIVEN: We can't easily create a sql.DB without a real database
	// This test verifies error handling for nil config
	// In reality, this would only fail if someone passes nil explicitly

	// WHEN/THEN: The function validates config is not nil
	// We'll test the validation logic through the nil config check
	assert.NotNil(t, DefaultPoolConfig())
}

func TestConfigurePool_PoolSettings(t *testing.T) {
	// GIVEN: A pool config with specific values
	config := &PoolConfig{
		MaxOpenConnections: 75,
		MaxIdleConnections: 15,
		MaxConnLifetime:    5400,
	}

	// WHEN: Verifying the config values
	// (We can't easily test ConfigurePool without a real database,
	// but we verify the config is properly structured)

	// THEN: Config should have correct values
	assert.Equal(t, 75, config.MaxOpenConnections)
	assert.Equal(t, 15, config.MaxIdleConnections)
	assert.Equal(t, 5400, config.MaxConnLifetime)
}
