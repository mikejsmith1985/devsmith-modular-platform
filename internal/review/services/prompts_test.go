package review_services

import (
	"strings"
	"testing"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestBuildPreviewPrompt_ModeVariations tests that BuildPreviewPrompt correctly
// adjusts tone based on userMode and includes reasoning_trace for full outputMode
func TestBuildPreviewPrompt_ModeVariations(t *testing.T) {
	testCode := `package main

func main() {
    fmt.Println("Hello, World!")
}`

	tests := []struct {
		name            string
		userMode        string
		outputMode      string
		expectAnalogies bool // Beginner mode should have analogies
		expectTechnical bool // Expert mode should be technical
		expectReasoning bool // Full mode should have reasoning_trace
	}{
		{
			name:            "Beginner + Quick: simple language, no reasoning",
			userMode:        "beginner",
			outputMode:      "quick",
			expectAnalogies: true,
			expectTechnical: false,
			expectReasoning: false,
		},
		{
			name:            "Beginner + Full: simple language WITH reasoning",
			userMode:        "beginner",
			outputMode:      "full",
			expectAnalogies: true,
			expectTechnical: false,
			expectReasoning: true,
		},
		{
			name:            "Expert + Quick: technical, concise, no reasoning",
			userMode:        "expert",
			outputMode:      "quick",
			expectAnalogies: false,
			expectTechnical: true,
			expectReasoning: false,
		},
		{
			name:            "Expert + Full: technical WITH reasoning",
			userMode:        "expert",
			outputMode:      "full",
			expectAnalogies: false,
			expectTechnical: true,
			expectReasoning: true,
		},
		{
			name:            "Intermediate + Quick (defaults): balanced, no reasoning",
			userMode:        "intermediate",
			outputMode:      "quick",
			expectAnalogies: false,
			expectTechnical: false,
			expectReasoning: false,
		},
		{
			name:            "Novice + Full: clear terms WITH reasoning",
			userMode:        "novice",
			outputMode:      "full",
			expectAnalogies: false,
			expectTechnical: false,
			expectReasoning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := BuildPreviewPrompt(testCode, tt.userMode, tt.outputMode)

			// Verify prompt is not empty
			if prompt == "" {
				t.Fatal("BuildPreviewPrompt returned empty string")
			}

			// Check for analogies (beginner mode indicator)
			hasAnalogies := strings.Contains(strings.ToLower(prompt), "analog") ||
				strings.Contains(prompt, "as if teaching") ||
				strings.Contains(prompt, "simple, non-technical")
			if tt.expectAnalogies && !hasAnalogies {
				t.Errorf("Expected analogies for %s mode but found none in prompt. Prompt snippet: %s", tt.userMode, prompt[:min(200, len(prompt))])
			}

			// Check for technical terminology (expert mode indicator)
			hasTechnical := strings.Contains(prompt, "technical terminology") ||
				strings.Contains(prompt, "precise") ||
				strings.Contains(prompt, "architectural patterns")
			if tt.expectTechnical && !hasTechnical {
				t.Errorf("Expected technical terminology for %s mode but found none", tt.userMode)
			}

			// Check for reasoning_trace section (full mode indicator)
			hasReasoning := strings.Contains(prompt, "reasoning_trace")
			if tt.expectReasoning != hasReasoning {
				if tt.expectReasoning {
					t.Errorf("Expected reasoning_trace for %s output mode but not found", tt.outputMode)
				} else {
					t.Errorf("Did NOT expect reasoning_trace for %s output mode but found it", tt.outputMode)
				}
			}

			// Verify code is included in prompt
			if !strings.Contains(prompt, testCode) {
				t.Error("Test code not found in generated prompt")
			}
		})
	}
}

// TestBuildSkimPrompt_ModeVariations tests Skim mode prompt adjustments
func TestBuildSkimPrompt_ModeVariations(t *testing.T) {
	testCode := `func ProcessOrder(order Order) error {
    return validateAndSave(order)
}`

	tests := []struct {
		name            string
		userMode        string
		outputMode      string
		expectSimple    bool
		expectReasoning bool
	}{
		{"Beginner + Quick", "beginner", "quick", true, false},
		{"Beginner + Full", "beginner", "full", true, true},
		{"Expert + Quick", "expert", "quick", false, false},
		{"Expert + Full", "expert", "full", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := BuildSkimPrompt(testCode, tt.userMode, tt.outputMode)

			if prompt == "" {
				t.Fatal("BuildSkimPrompt returned empty string")
			}

			hasSimpleLanguage := strings.Contains(strings.ToLower(prompt), "simple") ||
				strings.Contains(strings.ToLower(prompt), "non-technical")
			if tt.expectSimple && !hasSimpleLanguage {
				t.Errorf("Expected simple language guidance for %s mode", tt.userMode)
			}

			hasReasoning := strings.Contains(prompt, "reasoning_trace")
			if tt.expectReasoning != hasReasoning {
				t.Errorf("Reasoning trace expectation mismatch for %s mode", tt.outputMode)
			}
		})
	}
}

// TestBuildScanPrompt_ModeVariations tests Scan mode with query patterns
func TestBuildScanPrompt_ModeVariations(t *testing.T) {
	testCode := `SELECT * FROM users WHERE email = ?`
	query := "SQL queries"

	tests := []struct {
		name            string
		userMode        string
		outputMode      string
		expectReasoning bool
	}{
		{"Intermediate + Quick (defaults)", "intermediate", "quick", false},
		{"Beginner + Full", "beginner", "full", true},
		{"Expert + Quick", "expert", "quick", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := BuildScanPrompt(testCode, query, tt.userMode, tt.outputMode)

			if prompt == "" {
				t.Fatal("BuildScanPrompt returned empty string")
			}

			if !strings.Contains(prompt, query) {
				t.Errorf("Query '%s' not found in prompt", query)
			}

			hasReasoning := strings.Contains(prompt, "reasoning_trace")
			if tt.expectReasoning != hasReasoning {
				t.Errorf("Reasoning trace expectation mismatch for %s mode", tt.outputMode)
			}
		})
	}
}

// TestBuildDetailedPrompt_ModeVariations tests line-by-line analysis mode
func TestBuildDetailedPrompt_ModeVariations(t *testing.T) {
	testCode := `for i := 0; i < len(items); i++ {
    process(items[i])
}`
	filename := "processor.go"

	tests := []struct {
		name            string
		userMode        string
		outputMode      string
		expectAnalogies bool
		expectReasoning bool
	}{
		{
			name:            "Beginner + Full: analogies AND reasoning",
			userMode:        "beginner",
			outputMode:      "full",
			expectAnalogies: true,
			expectReasoning: true,
		},
		{
			name:            "Expert + Quick: technical, concise",
			userMode:        "expert",
			outputMode:      "quick",
			expectAnalogies: false,
			expectReasoning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := BuildDetailedPrompt(testCode, filename, tt.userMode, tt.outputMode)

			if prompt == "" {
				t.Fatal("BuildDetailedPrompt returned empty string")
			}

			if !strings.Contains(prompt, filename) {
				t.Errorf("Filename '%s' not found in prompt", filename)
			}

			hasAnalogies := strings.Contains(strings.ToLower(prompt), "analog") ||
				strings.Contains(prompt, "as if teaching") ||
				strings.Contains(prompt, "simple, non-technical")
			if tt.expectAnalogies && !hasAnalogies {
				t.Error("Expected analogies in beginner mode prompt")
			}

			hasReasoning := strings.Contains(prompt, "reasoning_trace")
			if tt.expectReasoning != hasReasoning {
				t.Errorf("Reasoning trace expectation mismatch")
			}
		})
	}
}

// TestPromptBuilder_DefaultValues tests that empty/invalid modes get reasonable defaults
func TestPromptBuilder_DefaultValues(t *testing.T) {
	testCode := "package main"

	tests := []struct {
		name       string
		userMode   string
		outputMode string
	}{
		{"Empty modes", "", ""},
		{"Invalid user mode", "invalid", "quick"},
		{"Invalid output mode", "intermediate", "invalid"},
		{"Both invalid", "invalid", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic and should return valid prompt
			prompt := BuildPreviewPrompt(testCode, tt.userMode, tt.outputMode)
			if prompt == "" {
				t.Error("Expected valid prompt even with invalid modes")
			}

			// Invalid modes should fall back to intermediate/quick
			// So should NOT have reasoning_trace (which only appears in full mode)
			hasReasoning := strings.Contains(prompt, "reasoning_trace")
			if hasReasoning {
				t.Error("Invalid modes should default to quick (no reasoning_trace)")
			}
		})
	}
}

// TestPromptBuilder_CodeIncluded ensures all builders include the actual code
func TestPromptBuilder_CodeIncluded(t *testing.T) {
	testCode := "// UNIQUE_TEST_MARKER_12345"

	t.Run("Preview includes code", func(t *testing.T) {
		prompt := BuildPreviewPrompt(testCode, "intermediate", "quick")
		if !strings.Contains(prompt, testCode) {
			t.Error("Code not included in preview prompt")
		}
	})

	t.Run("Skim includes code", func(t *testing.T) {
		prompt := BuildSkimPrompt(testCode, "intermediate", "quick")
		if !strings.Contains(prompt, testCode) {
			t.Error("Code not included in skim prompt")
		}
	})

	t.Run("Scan includes code", func(t *testing.T) {
		prompt := BuildScanPrompt(testCode, "test pattern", "intermediate", "quick")
		if !strings.Contains(prompt, testCode) {
			t.Error("Code not included in scan prompt")
		}
	})

	t.Run("Detailed includes code", func(t *testing.T) {
		prompt := BuildDetailedPrompt(testCode, "test.go", "intermediate", "quick")
		if !strings.Contains(prompt, testCode) {
			t.Error("Code not included in detailed prompt")
		}
	})
}
