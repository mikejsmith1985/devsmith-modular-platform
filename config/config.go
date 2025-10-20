// Package config provides configuration loading for environment variables and service ports.
package config

import (
	"log"
	"os"
)

// Config holds environment variables and service port settings.
type Config struct {
	PostgresUser       string
	PostgresPassword   string
	PostgresDB         string
	PortalPort         string
	ReviewPort         string
	LogsPort           string
	AnalyticsPort      string
	NginxPort          string
	GitHubClientID     string
	GitHubClientSecret string
	JWTSecret          string
}

// LoadConfig loads configuration from environment variables and validates it.
func LoadConfig() *Config {
	cfg := &Config{
		PostgresUser:       getEnv("POSTGRES_USER"),
		PostgresPassword:   getEnv("POSTGRES_PASSWORD"),
		PostgresDB:         getEnv("POSTGRES_DB"),
		PortalPort:         getEnv("PORTAL_PORT"),
		ReviewPort:         getEnv("REVIEW_PORT"),
		LogsPort:           getEnv("LOGS_PORT"),
		AnalyticsPort:      getEnv("ANALYTICS_PORT"),
		NginxPort:          getEnv("NGINX_PORT"),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET"),
		JWTSecret:          getEnv("JWT_SECRET"),
	}
	cfg.Validate()
	return cfg
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Missing required environment variable: %s", key)
	}
	return val
}

// Validate checks if the configuration is valid and returns an error if not.
func (c *Config) Validate() error {
	// Add additional validation logic if needed
	return nil
}
