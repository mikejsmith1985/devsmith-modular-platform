package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadLogsConfig_EnvVarSet tests loading config with explicit LOGS_SERVICE_URL
func TestLoadLogsConfig_EnvVarSet(t *testing.T) {
	// Setup
	os.Setenv("LOGS_SERVICE_URL", "http://logs.example.com:8082/api/logs")
	defer os.Unsetenv("LOGS_SERVICE_URL")

	// Execute
	url, err := LoadLogsConfig()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "http://logs.example.com:8082/api/logs", url)
}

// TestLoadLogsConfig_DockerDefault tests default config for Docker environment
func TestLoadLogsConfig_DockerDefault(t *testing.T) {
	// Setup: Unset explicit URL, set Docker environment
	os.Unsetenv("LOGS_SERVICE_URL")
	os.Setenv("ENVIRONMENT", "docker")
	defer os.Unsetenv("ENVIRONMENT")

	// Execute
	url, err := LoadLogsConfig()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "http://logs:8082/api/logs", url)
}

// TestLoadLogsConfig_LocalDefault tests default config for local development
func TestLoadLogsConfig_LocalDefault(t *testing.T) {
	// Setup: Unset both env vars (should default to local)
	os.Unsetenv("LOGS_SERVICE_URL")
	os.Unsetenv("ENVIRONMENT")

	// Execute
	url, err := LoadLogsConfig()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8082/api/logs", url)
}

// TestLoadLogsConfig_LocalExplicit tests explicit local environment
func TestLoadLogsConfig_LocalExplicit(t *testing.T) {
	// Setup
	os.Unsetenv("LOGS_SERVICE_URL")
	os.Setenv("ENVIRONMENT", "local")
	defer os.Unsetenv("ENVIRONMENT")

	// Execute
	url, err := LoadLogsConfig()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8082/api/logs", url)
}

// TestLoadLogsConfig_ProductionMissingConfig tests that production requires explicit config
func TestLoadLogsConfig_ProductionMissingConfig(t *testing.T) {
	// Setup
	os.Unsetenv("LOGS_SERVICE_URL")
	os.Setenv("ENVIRONMENT", "production")
	defer os.Unsetenv("ENVIRONMENT")

	// Execute
	url, err := LoadLogsConfig()

	// Assert
	assert.Error(t, err)
	assert.Empty(t, url)
	assert.Contains(t, err.Error(), "LOGS_SERVICE_URL must be set")
}

// TestLoadLogsConfig_ProductionWithConfig tests production with explicit config
func TestLoadLogsConfig_ProductionWithConfig(t *testing.T) {
	// Setup
	os.Setenv("LOGS_SERVICE_URL", "https://logs.prod.example.com/api/logs")
	os.Setenv("ENVIRONMENT", "production")
	defer os.Unsetenv("LOGS_SERVICE_URL")
	defer os.Unsetenv("ENVIRONMENT")

	// Execute
	url, err := LoadLogsConfig()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "https://logs.prod.example.com/api/logs", url)
}

// TestValidateLogsURL_ValidHTTP tests validation of valid HTTP URL
func TestValidateLogsURL_ValidHTTP(t *testing.T) {
	err := validateLogsURL("http://localhost:8082/api/logs")
	assert.NoError(t, err)
}

// TestValidateLogsURL_ValidHTTPS tests validation of valid HTTPS URL
func TestValidateLogsURL_ValidHTTPS(t *testing.T) {
	err := validateLogsURL("https://logs.example.com/api/logs")
	assert.NoError(t, err)
}

// TestValidateLogsURL_ValidDockerDNS tests validation of Docker internal DNS
func TestValidateLogsURL_ValidDockerDNS(t *testing.T) {
	err := validateLogsURL("http://logs:8082/api/logs")
	assert.NoError(t, err)
}

// TestValidateLogsURL_InvalidScheme tests rejection of invalid scheme
func TestValidateLogsURL_InvalidScheme(t *testing.T) {
	err := validateLogsURL("ftp://logs:8082/api/logs")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid URL scheme")
}

// TestValidateLogsURL_InvalidPath tests rejection of invalid path
func TestValidateLogsURL_InvalidPath(t *testing.T) {
	err := validateLogsURL("http://logs:8082/logs")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid URL path")
}

// TestValidateLogsURL_MissingHost tests rejection of missing host
func TestValidateLogsURL_MissingHost(t *testing.T) {
	err := validateLogsURL("http:///api/logs")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing host")
}

// TestValidateLogsURL_Empty tests rejection of empty URL
func TestValidateLogsURL_Empty(t *testing.T) {
	err := validateLogsURL("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

// TestValidateLogsURL_InvalidFormat tests rejection of malformed URL
func TestValidateLogsURL_InvalidFormat(t *testing.T) {
	err := validateLogsURL("not a url at all")
	assert.Error(t, err)
}

// TestValidateLogsURL_WithPort tests URL with explicit port
func TestValidateLogsURL_WithPort(t *testing.T) {
	err := validateLogsURL("http://logs.example.com:9000/api/logs")
	assert.NoError(t, err)
}

// TestValidateLogsURL_WithAuthentication tests URL with authentication
func TestValidateLogsURL_WithAuthentication(t *testing.T) {
	err := validateLogsURL("https://user:pass@logs.example.com/api/logs")
	assert.NoError(t, err)
}
