package version

import (
	"fmt"
	"os"
	"time"
)

// Version information - set during build
var (
	// Version is the semantic version (e.g., "v0.1.0")
	Version = "dev"

	// CommitHash is the git commit hash
	CommitHash = "unknown"

	// BuildTime is when the binary was built
	BuildTime = "unknown"

	// BuildNumber is the CI build number (optional)
	BuildNumber = "0"
)

// CacheBuster returns a string suitable for cache busting in URLs
// Format: v0.1.0-abc1234 or dev-abc1234-20251112
func CacheBuster() string {
	if Version != "dev" {
		// Production: use version-commithash (e.g., v0.1.0-abc1234)
		return fmt.Sprintf("%s-%s", Version, shortHash())
	}

	// Development: use dev-commithash-buildtime (changes every build)
	return fmt.Sprintf("dev-%s-%s", shortHash(), buildTimestamp())
}

// QueryParam returns the cache buster as a URL query parameter
// Example: ?v=v0.1.0-abc1234
func QueryParam() string {
	return fmt.Sprintf("v=%s", CacheBuster())
}

// ShortVersion returns a short version string for display
func ShortVersion() string {
	if Version != "dev" {
		return Version
	}
	return fmt.Sprintf("dev-%s", shortHash())
}

// FullVersion returns complete version information
func FullVersion() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, CommitHash, BuildTime)
}

// shortHash returns first 7 chars of commit hash
func shortHash() string {
	if len(CommitHash) >= 7 {
		return CommitHash[:7]
	}
	return CommitHash
}

// buildTimestamp returns formatted build time for cache busting
func buildTimestamp() string {
	if BuildTime != "unknown" {
		// Parse and format as YYYYMMDD
		if t, err := time.Parse(time.RFC3339, BuildTime); err == nil {
			return t.Format("20060102")
		}
	}
	// Fallback to current date in dev
	return time.Now().Format("20060102")
}

// IsDevelopment returns true if running development build
func IsDevelopment() bool {
	return Version == "dev"
}

// IsProduction returns true if running production build
func IsProduction() bool {
	return Version != "dev"
}

// GetFromEnv allows overriding version from environment (for testing)
func GetFromEnv() {
	if v := os.Getenv("APP_VERSION"); v != "" {
		Version = v
	}
	if h := os.Getenv("GIT_COMMIT_HASH"); h != "" {
		CommitHash = h
	}
	if t := os.Getenv("BUILD_TIME"); t != "" {
		BuildTime = t
	}
}
