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
