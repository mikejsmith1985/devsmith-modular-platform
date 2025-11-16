package tests
//go:build integration
// +build integration

package devsmith

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Test configuration loaded from .test-config.json
type TestConfig struct {
	APIKey        string `json:"apiKey"`
	ProjectSlug   string `json:"projectSlug"`
	ProjectID     int    `json:"projectId"`
	APIUrl        string `json:"apiUrl"`
	BatchEndpoint string `json:"batchEndpoint"`
}

var testConfig TestConfig

// LoadTestConfig loads the shared test configuration
func loadTestConfig() error {
	configPath := filepath.Join("../../tests", ".test-config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read test config: %w", err)
	}

	if err := json.Unmarshal(data, &testConfig); err != nil {
		return fmt.Errorf("failed to parse test config: %w", err)
	}

	return nil
}

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	if err := loadTestConfig(); err != nil {
		fmt.Printf("Failed to load test config: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
