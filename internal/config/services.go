// Package config provides configuration loading for DevSmith services.
package config

import (
	"fmt"
	"os"
	"strings"
)

// ServicePorts defines default ports for each service
var ServicePorts = map[string]string{
	"gateway":   "3000",
	"portal":    "8080",
	"review":    "8081",
	"logs":      "8082",
	"analytics": "8083",
}

// GetServiceURL returns the URL for a service based on environment configuration.
// It checks in this order:
// 1. Per-service environment variable (e.g., PORTAL_URL)
// 2. Constructs URL based on ENVIRONMENT variable (docker vs local)
func GetServiceURL(service string) string {
	// Check for explicit environment variable first
	envKey := strings.ToUpper(service) + "_URL"
	if url := os.Getenv(envKey); url != "" {
		return strings.TrimRight(url, "/")
	}

	// Determine environment
	env := os.Getenv("ENVIRONMENT")
	isDocker := strings.EqualFold(env, "docker") || os.Getenv("DOCKER") == "true"

	port, exists := ServicePorts[service]
	if !exists {
		// Unknown service, return empty
		return ""
	}

	if isDocker {
		// Docker internal DNS: http://service-name:port
		return fmt.Sprintf("http://%s:%s", service, port)
	}

	// Local development: http://localhost:port
	return fmt.Sprintf("http://localhost:%s", port)
}

// GetServiceHealthURL returns the health check URL for a service
func GetServiceHealthURL(service string) string {
	baseURL := GetServiceURL(service)
	if baseURL == "" {
		return ""
	}

	// Gateway doesn't have /health, just root
	if service == "gateway" {
		return baseURL + "/"
	}

	return baseURL + "/health"
}

// GetGatewayURL returns the gateway URL (used for redirects)
func GetGatewayURL() string {
	return GetServiceURL("gateway")
}

// GetDatabaseURL returns the PostgreSQL connection string
func GetDatabaseURL() string {
	// Check explicit DATABASE_URL first
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return dbURL
	}

	// Determine environment
	env := os.Getenv("ENVIRONMENT")
	isDocker := strings.EqualFold(env, "docker") || os.Getenv("DOCKER") == "true"

	host := "localhost"
	if isDocker {
		host = "postgres"
	}

	// Default connection parameters
	user := getEnvOrDefault("POSTGRES_USER", "devsmith")
	password := getEnvOrDefault("POSTGRES_PASSWORD", "devsmith")
	database := getEnvOrDefault("POSTGRES_DB", "devsmith")
	port := getEnvOrDefault("POSTGRES_PORT", "5432")
	sslMode := getEnvOrDefault("POSTGRES_SSLMODE", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, database, sslMode)
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
