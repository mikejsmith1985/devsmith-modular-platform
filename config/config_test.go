package config

import (
	"os"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	cfg := &Config{
		PostgresUser:       "testuser",
		PostgresPassword:   "testpass",
		PostgresDB:         "testdb",
		PortalPort:         "8080",
		ReviewPort:         "8081",
		LogsPort:           "8082",
		AnalyticsPort:      "8083",
		NginxPort:          "80",
		GitHubClientID:     "test_client_id",
		GitHubClientSecret: "test_client_secret",
		JWTSecret:          "test_jwt_secret",
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() returned unexpected error: %v", err)
	}
}

func TestLoadConfig_Success(t *testing.T) {
	// Set up test environment variables
	envVars := map[string]string{
		"POSTGRES_USER":        "testuser",
		"POSTGRES_PASSWORD":    "testpass",
		"POSTGRES_DB":          "testdb",
		"PORTAL_PORT":          "8080",
		"REVIEW_PORT":          "8081",
		"LOGS_PORT":            "8082",
		"ANALYTICS_PORT":       "8083",
		"NGINX_PORT":           "80",
		"GITHUB_CLIENT_ID":     "test_client_id",
		"GITHUB_CLIENT_SECRET": "test_client_secret",
		"JWT_SECRET":           "test_jwt_secret",
	}

	// Set environment variables
	for key, val := range envVars {
		os.Setenv(key, val)
	}

	// Clean up after test
	defer func() {
		for key := range envVars {
			os.Unsetenv(key)
		}
	}()

	cfg := LoadConfig()

	// Verify all fields are populated correctly
	if cfg.PostgresUser != "testuser" {
		t.Errorf("Expected PostgresUser to be 'testuser', got '%s'", cfg.PostgresUser)
	}
	if cfg.PostgresPassword != "testpass" {
		t.Errorf("Expected PostgresPassword to be 'testpass', got '%s'", cfg.PostgresPassword)
	}
	if cfg.PostgresDB != "testdb" {
		t.Errorf("Expected PostgresDB to be 'testdb', got '%s'", cfg.PostgresDB)
	}
	if cfg.PortalPort != "8080" {
		t.Errorf("Expected PortalPort to be '8080', got '%s'", cfg.PortalPort)
	}
	if cfg.ReviewPort != "8081" {
		t.Errorf("Expected ReviewPort to be '8081', got '%s'", cfg.ReviewPort)
	}
	if cfg.LogsPort != "8082" {
		t.Errorf("Expected LogsPort to be '8082', got '%s'", cfg.LogsPort)
	}
	if cfg.AnalyticsPort != "8083" {
		t.Errorf("Expected AnalyticsPort to be '8083', got '%s'", cfg.AnalyticsPort)
	}
	if cfg.NginxPort != "80" {
		t.Errorf("Expected NginxPort to be '80', got '%s'", cfg.NginxPort)
	}
	if cfg.GitHubClientID != "test_client_id" {
		t.Errorf("Expected GitHubClientID to be 'test_client_id', got '%s'", cfg.GitHubClientID)
	}
	if cfg.GitHubClientSecret != "test_client_secret" {
		t.Errorf("Expected GitHubClientSecret to be 'test_client_secret', got '%s'", cfg.GitHubClientSecret)
	}
	if cfg.JWTSecret != "test_jwt_secret" {
		t.Errorf("Expected JWTSecret to be 'test_jwt_secret', got '%s'", cfg.JWTSecret)
	}
}

func TestGetEnv_Success(t *testing.T) {
	key := "TEST_ENV_VAR"
	expectedValue := "test_value"

	os.Setenv(key, expectedValue)
	defer os.Unsetenv(key)

	result := getEnv(key)
	if result != expectedValue {
		t.Errorf("Expected getEnv('%s') to return '%s', got '%s'", key, expectedValue, result)
	}
}
