package review_services

import (
	"encoding/json"
	"strings"
)

// ExtractJSON finds and extracts the first valid JSON object or array from text.
// This handles cases where the AI model adds extra text before/after the JSON.
func ExtractJSON(text string) (string, error) {
	// Trim whitespace
	text = strings.TrimSpace(text)

	// Try parsing as-is first (fast path)
	if json.Valid([]byte(text)) {
		return text, nil
	}

	// Find JSON object boundaries { ... }
	startObj := strings.Index(text, "{")
	if startObj != -1 {
		// Find matching closing brace
		depth := 0
		for i := startObj; i < len(text); i++ {
			switch text[i] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					candidate := text[startObj : i+1]
					if json.Valid([]byte(candidate)) {
						return candidate, nil
					}
				}
			}
		}
	}

	// Find JSON array boundaries [ ... ]
	startArray := strings.Index(text, "[")
	if startArray != -1 {
		depth := 0
		for i := startArray; i < len(text); i++ {
			switch text[i] {
			case '[':
				depth++
			case ']':
				depth--
				if depth == 0 {
					candidate := text[startArray : i+1]
					if json.Valid([]byte(candidate)) {
						return candidate, nil
					}
				}
			}
		}
	}

	// If no valid JSON found, return original (will fail validation downstream)
	return text, nil
}
