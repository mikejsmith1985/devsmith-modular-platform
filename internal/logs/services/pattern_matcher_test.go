package logs_services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewPatternMatcher tests pattern matcher creation
func TestNewPatternMatcher(t *testing.T) {
	matcher := NewPatternMatcher()

	assert.NotNil(t, matcher)
	assert.NotNil(t, matcher.patterns)
}

// TestClassify_DatabaseConnection tests database connection error pattern
func TestClassify_DatabaseConnection(t *testing.T) {
	matcher := NewPatternMatcher()

	testCases := []struct {
		message  string
		expected string
	}{
		{"connection refused to database", "db_connection"},
		{"database timeout after 30s", "db_connection"},
		{"pg: connection failed", "db_connection"},
		{"pq: connection to server failed", "db_connection"},
		{"some other error", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			result := matcher.Classify(tc.message)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestClassify_AuthFailure tests authentication failure pattern
func TestClassify_AuthFailure(t *testing.T) {
	matcher := NewPatternMatcher()

	testCases := []struct {
		message  string
		expected string
	}{
		{"unauthorized access", "auth_failure"},
		{"authentication failed", "auth_failure"},
		{"invalid token provided", "auth_failure"},
		{"JWT validation failed", "auth_failure"},
		{"some other error", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			result := matcher.Classify(tc.message)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestClassify_NullPointer tests null pointer error pattern
func TestClassify_NullPointer(t *testing.T) {
	matcher := NewPatternMatcher()

	testCases := []struct {
		message  string
		expected string
	}{
		{"nil pointer dereference", "null_pointer"},
		{"null reference exception", "null_pointer"},
		{"undefined is not an object (evaluating 'x.y')", "null_pointer"},
		{"panic: runtime error: invalid memory address or nil pointer dereference", "null_pointer"},
		{"some other error", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			result := matcher.Classify(tc.message)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestClassify_RateLimit tests rate limiting error pattern
func TestClassify_RateLimit(t *testing.T) {
	matcher := NewPatternMatcher()

	testCases := []struct {
		message  string
		expected string
	}{
		{"rate limit exceeded", "rate_limit"},
		{"too many requests", "rate_limit"},
		{"HTTP 429 Too Many Requests", "rate_limit"},
		{"API rate limit hit", "rate_limit"},
		{"some other error", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			result := matcher.Classify(tc.message)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestClassify_NetworkTimeout tests network timeout error pattern
func TestClassify_NetworkTimeout(t *testing.T) {
	matcher := NewPatternMatcher()

	testCases := []struct {
		message  string
		expected string
	}{
		{"connection timeout after 30s", "network_timeout"},
		{"i/o timeout while reading", "network_timeout"},
		{"context deadline exceeded", "network_timeout"},
		{"request timeout", "network_timeout"},
		{"some other error", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			result := matcher.Classify(tc.message)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestClassify_CaseInsensitive tests that pattern matching is case-insensitive
func TestClassify_CaseInsensitive(t *testing.T) {
	matcher := NewPatternMatcher()

	testCases := []struct {
		message  string
		expected string
	}{
		{"CONNECTION REFUSED TO DATABASE", "db_connection"},
		{"Authentication Failed", "auth_failure"},
		{"Nil Pointer Dereference", "null_pointer"},
		{"RATE LIMIT EXCEEDED", "rate_limit"},
		{"Context Deadline Exceeded", "network_timeout"},
	}

	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			result := matcher.Classify(tc.message)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestClassify_FirstMatchWins tests priority of pattern matching
func TestClassify_FirstMatchWins(t *testing.T) {
	matcher := NewPatternMatcher()

	// A message that could match multiple patterns should match the first one
	message := "database connection timeout" // Could be db_connection or network_timeout

	result := matcher.Classify(message)

	// Should match db_connection since it's checked first
	assert.Equal(t, "db_connection", result)
}
