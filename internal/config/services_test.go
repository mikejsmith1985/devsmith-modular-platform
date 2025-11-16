package config

import (
	"os"
	"testing"
)

func TestGetServiceURL(t *testing.T) {
	tests := []struct {
		name        string
		service     string
		envVars     map[string]string
		expectedURL string
	}{
		{
			name:        "portal service in docker",
			service:     "portal",
			envVars:     map[string]string{"ENVIRONMENT": "docker"},
			expectedURL: "http://portal:8080",
		},
		{
			name:        "review service in docker",
			service:     "review",
			envVars:     map[string]string{"ENVIRONMENT": "docker"},
			expectedURL: "http://review:8081",
		},
		{
			name:        "logs service local",
			service:     "logs",
			envVars:     map[string]string{"ENVIRONMENT": "local"},
			expectedURL: "http://localhost:8082",
		},
		{
			name:        "analytics service local",
			service:     "analytics",
			envVars:     map[string]string{"ENVIRONMENT": "local"},
			expectedURL: "http://localhost:8083",
		},
		{
			name:        "gateway service docker",
			service:     "gateway",
			envVars:     map[string]string{"ENVIRONMENT": "docker"},
			expectedURL: "http://gateway:3000",
		},
		{
			name:        "gateway service local",
			service:     "gateway",
			envVars:     map[string]string{"ENVIRONMENT": "local"},
			expectedURL: "http://localhost:3000",
		},
		{
			name:        "invalid service returns empty",
			service:     "invalid",
			envVars:     map[string]string{"ENVIRONMENT": "docker"},
			expectedURL: "",
		},
		{
			name:        "explicit URL override",
			service:     "portal",
			envVars:     map[string]string{"PORTAL_URL": "http://custom:9999"},
			expectedURL: "http://custom:9999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			os.Unsetenv("ENVIRONMENT")
			os.Unsetenv("DOCKER")
			os.Unsetenv("PORTAL_URL")
			os.Unsetenv("REVIEW_URL")
			os.Unsetenv("LOGS_URL")
			os.Unsetenv("ANALYTICS_URL")
			os.Unsetenv("GATEWAY_URL")

			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			url := GetServiceURL(tt.service)

			if url != tt.expectedURL {
				t.Errorf("expected URL %s, got %s", tt.expectedURL, url)
			}
		})
	}
}

func TestGetServiceHealthURL(t *testing.T) {
	tests := []struct {
		name        string
		service     string
		envVars     map[string]string
		expectedURL string
	}{
		{
			name:        "portal health endpoint",
			service:     "portal",
			envVars:     map[string]string{"ENVIRONMENT": "docker"},
			expectedURL: "http://portal:8080/health",
		},
		{
			name:        "review health endpoint",
			service:     "review",
			envVars:     map[string]string{"ENVIRONMENT": "local"},
			expectedURL: "http://localhost:8081/health",
		},
		{
			name:        "gateway uses root not /health",
			service:     "gateway",
			envVars:     map[string]string{"ENVIRONMENT": "docker"},
			expectedURL: "http://gateway:3000/",
		},
		{
			name:        "invalid service returns empty",
			service:     "invalid",
			envVars:     map[string]string{"ENVIRONMENT": "docker"},
			expectedURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			os.Unsetenv("ENVIRONMENT")
			os.Unsetenv("DOCKER")

			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			url := GetServiceHealthURL(tt.service)

			if url != tt.expectedURL {
				t.Errorf("expected URL %s, got %s", tt.expectedURL, url)
			}
		})
	}
}

func TestGetGatewayURL(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectedURL string
	}{
		{
			name:        "default gateway URL",
			envVars:     map[string]string{},
			expectedURL: "http://localhost:3000",
		},
		{
			name:        "custom gateway URL from env",
			envVars:     map[string]string{"GATEWAY_URL": "http://custom:9000"},
			expectedURL: "http://custom:9000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear any existing GATEWAY_URL
			os.Unsetenv("GATEWAY_URL")

			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			url := GetGatewayURL()

			if url != tt.expectedURL {
				t.Errorf("expected URL %s, got %s", tt.expectedURL, url)
			}
		})
	}
}

func TestGetDatabaseURL(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectedURL string
	}{
		{
			name: "explicit DATABASE_URL from env",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://user:pass@customhost:5432/dbname?sslmode=require",
			},
			expectedURL: "postgresql://user:pass@customhost:5432/dbname?sslmode=require",
		},
		{
			name: "docker environment with defaults",
			envVars: map[string]string{
				"ENVIRONMENT": "docker",
			},
			expectedURL: "postgres://devsmith:devsmith@postgres:5432/devsmith?sslmode=disable",
		},
		{
			name: "local environment with defaults",
			envVars: map[string]string{
				"ENVIRONMENT": "local",
			},
			expectedURL: "postgres://devsmith:devsmith@localhost:5432/devsmith?sslmode=disable",
		},
		{
			name: "custom postgres parameters",
			envVars: map[string]string{
				"ENVIRONMENT":       "docker",
				"POSTGRES_USER":     "customuser",
				"POSTGRES_PASSWORD": "custompass",
				"POSTGRES_DB":       "customdb",
				"POSTGRES_PORT":     "5433",
				"POSTGRES_SSLMODE":  "require",
			},
			expectedURL: "postgres://customuser:custompass@postgres:5433/customdb?sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all database-related env vars
			os.Unsetenv("DATABASE_URL")
			os.Unsetenv("ENVIRONMENT")
			os.Unsetenv("DOCKER")
			os.Unsetenv("POSTGRES_USER")
			os.Unsetenv("POSTGRES_PASSWORD")
			os.Unsetenv("POSTGRES_DB")
			os.Unsetenv("POSTGRES_PORT")
			os.Unsetenv("POSTGRES_SSLMODE")

			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			url := GetDatabaseURL()

			if url != tt.expectedURL {
				t.Errorf("expected URL %s, got %s", tt.expectedURL, url)
			}
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name          string
		envKey        string
		envValue      string
		defaultValue  string
		expectedValue string
	}{
		{
			name:          "environment variable set",
			envKey:        "TEST_VAR",
			envValue:      "custom_value",
			defaultValue:  "default_value",
			expectedValue: "custom_value",
		},
		{
			name:          "environment variable not set",
			envKey:        "NONEXISTENT_VAR",
			envValue:      "",
			defaultValue:  "default_value",
			expectedValue: "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.envKey)
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnvOrDefault(tt.envKey, tt.defaultValue)

			if result != tt.expectedValue {
				t.Errorf("expected %s, got %s", tt.expectedValue, result)
			}
		})
	}
}
