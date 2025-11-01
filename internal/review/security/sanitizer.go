package security

import "github.com/microcosm-cc/bluemonday"

var codePolicy = bluemonday.StrictPolicy()

// SanitizeCodeInput removes potentially malicious HTML and script content from code input
// while preserving the actual code content. Uses bluemonday's strict policy for XSS prevention.
func SanitizeCodeInput(input string) string {
	return codePolicy.Sanitize(input)
}
