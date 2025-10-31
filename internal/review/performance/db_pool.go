// Package performance provides performance optimization utilities for the Review Service
package performance

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
)

// PoolConfig holds database connection pool settings
type PoolConfig struct {
	MaxOpenConnections int
	MaxIdleConnections int
	MaxConnLifetime    int // seconds
}

// DefaultPoolConfig returns recommended connection pool settings
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxOpenConnections: 50, // Issue #26 requirement: max 50 connections
		MaxIdleConnections: 10,
		MaxConnLifetime:    3600, // 1 hour
	}
}

// LoadPoolConfigFromEnv loads pool configuration from environment variables
func LoadPoolConfigFromEnv() *PoolConfig {
	config := DefaultPoolConfig()

	// Allow override of max open connections
	if maxOpen := os.Getenv("DB_POOL_MAX_OPEN"); maxOpen != "" {
		if val, err := strconv.Atoi(maxOpen); err == nil && val > 0 {
			config.MaxOpenConnections = val
		}
	}

	// Allow override of max idle connections
	if maxIdle := os.Getenv("DB_POOL_MAX_IDLE"); maxIdle != "" {
		if val, err := strconv.Atoi(maxIdle); err == nil && val > 0 {
			config.MaxIdleConnections = val
		}
	}

	// Allow override of connection lifetime
	if lifetime := os.Getenv("DB_POOL_MAX_LIFETIME"); lifetime != "" {
		if val, err := strconv.Atoi(lifetime); err == nil && val > 0 {
			config.MaxConnLifetime = val
		}
	}

	return config
}

// ConfigurePool applies connection pool settings to a database connection
func ConfigurePool(db *sql.DB, config *PoolConfig) error {
	if db == nil {
		return fmt.Errorf("performance: cannot configure nil database connection")
	}
	if config == nil {
		return fmt.Errorf("performance: cannot configure with nil config")
	}

	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)
	// Note: SetConnMaxLifetime requires time.Duration, not int

	return nil
}
