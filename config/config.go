package config

import (
	"log"
	"os"
)

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

func (c *Config) Validate() {
	// Add additional validation logic if needed
}
