package config

import (
	"fmt"
	neturl "net/url"
	"os"
	"strings"
)

// LoadLogsConfig reads LOGS_SERVICE_URL from the environment (or provides a sensible default)
// and validates the URL. Returns the resolved URL or an error.
func LoadLogsConfig() (string, error) {
	u := strings.TrimSpace(os.Getenv("LOGS_SERVICE_URL"))
	env := strings.TrimSpace(os.Getenv("ENVIRONMENT"))

	if u == "" {
		if strings.ToLower(env) == "docker" {
			u = "http://logs:8082/api/logs"
		} else {
			u = "http://localhost:8082/api/logs"
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

// LoadLogsConfigWithFallback loads the logs service URL and honors the
// LOGS_STRICT environment variable. If LOGS_STRICT is set to "false"
// and the configured URL is invalid, this function returns ("", false, nil)
// meaning logging is disabled but the service can continue startup.
// If LOGS_STRICT is true (default), it behaves the same as LoadLogsConfig
// and returns an error on invalid configuration.
// LoadLogsConfigFor returns the logs URL for a specific service. It first
// checks for per-service override like REVIEW_LOGS_URL (uppercased service
// name + _LOGS_URL). If not present it falls back to LOGS_SERVICE_URL and
// then to sensible defaults (docker vs local). Validation is performed.
func LoadLogsConfigFor(service string) (string, error) {
	// Per-service override, e.g., REVIEW_LOGS_URL
	svcKey := ""
	if strings.TrimSpace(service) != "" {
		svcKey = strings.ToUpper(service) + "_LOGS_URL"
	}

	var u string
	if svcKey != "" {
		u = strings.TrimSpace(os.Getenv(svcKey))
	}

	if u == "" {
		u = strings.TrimSpace(os.Getenv("LOGS_SERVICE_URL"))
	}

	env := strings.TrimSpace(os.Getenv("ENVIRONMENT"))
	if u == "" {
		if strings.ToLower(env) == "docker" {
			u = "http://logs:8082/api/logs"
		} else {
			u = "http://localhost:8082/api/logs"
		}
	}

	if err := validateLogsURL(u); err != nil {
		return "", fmt.Errorf("invalid logs url %q: %w", u, err)
	}
	return u, nil
}

// LoadLogsConfigWithFallbackFor behaves like LoadLogsConfigFor but respects
// LOGS_STRICT. If LOGS_STRICT=false and validation fails this returns
// ("", false, nil) indicating logging is disabled but startup can continue.
func LoadLogsConfigWithFallbackFor(service string) (string, bool, error) {
	strict := strings.TrimSpace(os.Getenv("LOGS_STRICT"))
	if strings.EqualFold(strict, "false") {
		// Try to load but don't fail startup when invalid
		u, err := LoadLogsConfigFor(service)
		if err != nil {
			return "", false, nil
		}
		return u, true, nil
	}

	u, err := LoadLogsConfigFor(service)
	if err != nil {
		return "", false, err
	}
	return u, true, nil
}
