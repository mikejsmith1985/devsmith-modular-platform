// Package config provides configuration loading for DevSmith services.
package config

import (
	"fmt"
	neturl "net/url"
	"os"
	"strings"
)

const (
	// logsServiceURLDocker is the logs service URL in Docker environment
	logsServiceURLDocker = "http://logs:8082/api/logs"
	// logsServiceURLLocal is the logs service URL in local development
	logsServiceURLLocal = "http://localhost:8082/api/logs"
)

// LoadLogsConfig reads LOGS_SERVICE_URL from the environment (or provides a sensible default)
// and validates the URL. Returns the resolved URL or an error.
func LoadLogsConfig() (string, error) {
	u := strings.TrimSpace(os.Getenv("LOGS_SERVICE_URL"))
	env := strings.TrimSpace(os.Getenv("ENVIRONMENT"))

	if u == "" {
		if strings.EqualFold(env, "docker") {
			u = logsServiceURLDocker
		} else {
			u = logsServiceURLLocal
		}
	}

	if err := validateLogsURL(u); err != nil {
		return "", fmt.Errorf("invalid LOGS_SERVICE_URL %q: %w", u, err)
	}

	return u, nil
}

// validateLogsURL checks the scheme, host, and path of the logs URL.
func validateLogsURL(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return fmt.Errorf("logs service URL cannot be empty")
	}
	parsed, err := neturl.Parse(raw)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("invalid scheme: %s", parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("invalid URL: missing host")
	}
	// Accept /api/logs or /api/logs/ (allow trailing slash)
	if !strings.HasPrefix(parsed.Path, "/api/logs") {
		return fmt.Errorf("invalid path: %s (must start with /api/logs)", parsed.Path)
	}
	return nil
}

// LoadLogsConfigFor loads the logs service URL for a specific service.
func LoadLogsConfigFor(service string) (url string, isOverride bool, err error) {
	env := os.Getenv("ENVIRONMENT")

	var u string
	if strings.EqualFold(env, "docker") {
		u = logsServiceURLDocker
	} else {
		u = logsServiceURLLocal
	}

	// Per-service override, e.g., REVIEW_LOGS_URL
	svcKey := ""
	if strings.TrimSpace(service) != "" {
		svcKey = strings.ToUpper(service) + "_LOGS_URL"
	}

	if svcKey != "" {
		override := strings.TrimSpace(os.Getenv(svcKey))
		if override != "" {
			u = override
		}
	}

	if u == "" {
		u = strings.TrimSpace(os.Getenv("LOGS_SERVICE_URL"))
	}

	if u == "" {
		if strings.EqualFold(env, "docker") {
			u = logsServiceURLDocker
		} else {
			u = logsServiceURLLocal
		}
	}

	if err := validateLogsURL(u); err != nil {
		return "", false, fmt.Errorf("invalid logs url %q: %w", u, err)
	}
	return u, true, nil
}

// LoadLogsConfigWithFallbackFor loads the logs service URL with fallback logic.
func LoadLogsConfigWithFallbackFor(service string) (url string, isOverride bool, err error) {
	env := os.Getenv("ENVIRONMENT")

	// Check for per-service override first
	override := os.Getenv(strings.ToUpper(service) + "_LOGS_URL")
	if override != "" {
		return override, true, nil
	}

	// Then check global LOGS_SERVICE_URL
	global := os.Getenv("LOGS_SERVICE_URL")
	if global != "" {
		return global, true, nil
	}

	// Fall back to default
	if strings.EqualFold(env, "docker") {
		return logsServiceURLDocker, true, nil
	}
	return logsServiceURLLocal, true, nil
}
