// Package logs_services provides service implementations for logs operations.
package logs_services

import (
	"regexp"
	"strings"
)

// PatternMatcher classifies log messages based on error patterns
type PatternMatcher struct {
	patterns map[string]*regexp.Regexp
}

// NewPatternMatcher creates a new pattern matcher with predefined patterns
func NewPatternMatcher() *PatternMatcher {
	return &PatternMatcher{
		patterns: map[string]*regexp.Regexp{
			"db_connection":   regexp.MustCompile(`(?i)(connection refused|database.*timeout|pg.*connect|pq.*connection)`),
			"auth_failure":    regexp.MustCompile(`(?i)(unauthorized|authentication.*failed|invalid.*token|JWT.*validation)`),
			"null_pointer":    regexp.MustCompile(`(?i)(nil pointer|null reference|undefined|panic.*nil pointer)`),
			"rate_limit":      regexp.MustCompile(`(?i)(rate limit|too many requests|429|API.*rate)`),
			"network_timeout": regexp.MustCompile(`(?i)(timeout|i/o timeout|context deadline|request timeout)`),
		},
	}
}

// Classify determines the issue type based on the log message
func (p *PatternMatcher) Classify(logMsg string) string {
	// Normalize the message for matching
	normalizedMsg := strings.TrimSpace(logMsg)

	// Check patterns in order (first match wins)
	// Order matters - more specific patterns should come first
	checkOrder := []string{
		"db_connection",
		"auth_failure",
		"null_pointer",
		"rate_limit",
		"network_timeout",
	}

	for _, issueType := range checkOrder {
		if pattern, exists := p.patterns[issueType]; exists {
			if pattern.MatchString(normalizedMsg) {
				return issueType
			}
		}
	}

	return "unknown"
}

// AddPattern adds a new pattern to the matcher
func (p *PatternMatcher) AddPattern(issueType string, pattern *regexp.Regexp) {
	p.patterns[issueType] = pattern
}

// GetPatterns returns all configured patterns
func (p *PatternMatcher) GetPatterns() map[string]*regexp.Regexp {
	return p.patterns
}
