package config

import (
	"os"
	"testing"
)

func TestLoadLogsConfig_EnvVar(t *testing.T) {
	orig := os.Getenv("LOGS_SERVICE_URL")
	defer os.Setenv("LOGS_SERVICE_URL", orig)

	os.Setenv("LOGS_SERVICE_URL", "http://example:8082/api/logs")
	os.Unsetenv("ENVIRONMENT")

	u, err := LoadLogsConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u != "http://example:8082/api/logs" {
		t.Fatalf("expected http://example:8082/api/logs, got %s", u)
	}
}

func TestLoadLogsConfig_InvalidURL(t *testing.T) {
	orig := os.Getenv("LOGS_SERVICE_URL")
	defer os.Setenv("LOGS_SERVICE_URL", orig)

	os.Setenv("LOGS_SERVICE_URL", "ftp://invalid/path")
	_, err := LoadLogsConfig()
	if err == nil {
		t.Fatalf("expected error for invalid URL, got nil")
	}
}

func TestLoadLogsConfig_DefaultDocker(t *testing.T) {
	orig := os.Getenv("LOGS_SERVICE_URL")
	defer os.Setenv("LOGS_SERVICE_URL", orig)

	os.Unsetenv("LOGS_SERVICE_URL")
	os.Setenv("ENVIRONMENT", "docker")
	defer os.Unsetenv("ENVIRONMENT")

	u, err := LoadLogsConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u != "http://logs:8082/api/logs" {
		t.Fatalf("expected docker default http://logs:8082/api/logs, got %s", u)
	}
}

func TestLoadLogsConfig_DefaultLocal(t *testing.T) {
	origURL := os.Getenv("LOGS_SERVICE_URL")
	origEnv := os.Getenv("ENVIRONMENT")
	defer os.Setenv("LOGS_SERVICE_URL", origURL)
	defer os.Setenv("ENVIRONMENT", origEnv)

	os.Unsetenv("LOGS_SERVICE_URL")
	os.Unsetenv("ENVIRONMENT")

	u, err := LoadLogsConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u != "http://localhost:8082/api/logs" {
		t.Fatalf("expected local default http://localhost:8082/api/logs, got %s", u)
	}
}
