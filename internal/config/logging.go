// Package config provides configuration management for the platform services.
package config

import (
	"fmt"
	"net/url"
	"os"
)

// LoadLogsConfig loads and validates the logging service configuration.
// It reads from the LOGS_SERVICE_URL environment variable and provides
// sensible defaults based on the deployment environment (local, Docker, production).
//
// Environment Variables:
//   - LOGS_SERVICE_URL: Full URL to logs service endpoint (e.g., http://logs:8082/api/logs)
//   - ENVIRONMENT: Deployment environment (docker, local, production) - used for defaults
//
// Defaults:
//   - Docker: http://logs:8082/api/logs (uses Docker internal DNS)
//   - Local: http://localhost:8082/api/logs (localhost development)
//   - Other: Must be explicitly configured
//
// Returns:
//   - string: Validated logs service URL
//   - error: If URL is invalid or required env var is missing in non-default scenario
//
// Example:
//   logsURL, err := LoadLogsConfig()
//   if err != nil {
//       log.Fatalf("Failed to load logging config: %v", err)
//   }
func LoadLogsConfig() (string, error) {
	// Try to read from environment
	logsURL := os.Getenv("LOGS_SERVICE_URL")

	// If not set, use environment-specific default
	if logsURL == "" {
		env := os.Getenv("ENVIRONMENT")
		switch env {
		case "docker":
			// Docker Compose: Use service name for internal DNS
			logsURL = "http://logs:8082/api/logs"
		case "local", "":
			// Local development: Use localhost
			logsURL = "http://localhost:8082/api/logs"
		default:
			// Production-like environments require explicit configuration
			return "", fmt.Errorf(
				"LOGS_SERVICE_URL must be set for environment '%s' (use http://..., https://...)",
				env,
			)
		}
	}

	// Validate the URL
	if err := validateLogsURL(logsURL); err != nil {
		return "", err
	}

	return logsURL, nil
}

// validateLogsURL ensures the logs service URL is properly formatted and valid.
// Checks:
//   - URL parses correctly
//   - Scheme is http or https
//   - Path is /api/logs
//   - Host is not empty
func validateLogsURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("logs service URL cannot be empty")
	}

	// Parse the URL
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid logs URL format: %w", err)
	}

	// Validate scheme
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf(
			"invalid URL scheme '%s': must be http or https (got: %s)",
			parsed.Scheme, urlStr,
		)
	}

	// Validate host
	if parsed.Host == "" {
		return fmt.Errorf(
			"invalid URL: missing host (got: %s)",
			urlStr,
		)
	}

	// Validate path
	if parsed.Path != "/api/logs" {
		return fmt.Errorf(
			"invalid URL path '%s': must be /api/logs (got: %s)",
			parsed.Path, urlStr,
		)
	}

	return nil
}
