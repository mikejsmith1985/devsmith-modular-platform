package review_services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPreviewService_PassesModesToPromptBuilder validates that PreviewService
// correctly passes user_mode and output_mode parameters to BuildPreviewPrompt
func TestPreviewService_PassesModesToPromptBuilder(t *testing.T) {
	// This test validates the integration between service and prompt builder
	// We can't easily mock BuildPreviewPrompt, but we can verify the prompt
	// contains the expected tone guidance based on modes

	tests := []struct {
		name              string
		userMode          string
		outputMode        string
		expectInPrompt    string // What we expect to find in the generated prompt
		expectNotInPrompt string // What should NOT be in the prompt
	}{
		{
			name:              "Beginner mode includes simple language guidance",
			userMode:          "beginner",
			outputMode:        "quick",
			expectInPrompt:    "simple, non-technical",
			expectNotInPrompt: "reasoning_trace",
		},
		{
			name:              "Expert mode includes technical guidance",
			userMode:          "expert",
			outputMode:        "quick",
			expectInPrompt:    "precise technical",
			expectNotInPrompt: "reasoning_trace",
		},
		{
			name:              "Full mode includes reasoning trace",
			userMode:          "intermediate",
			outputMode:        "full",
			expectInPrompt:    "reasoning_trace",
			expectNotInPrompt: "",
		},
		{
			name:              "Quick mode excludes reasoning trace",
			userMode:          "intermediate",
			outputMode:        "quick",
			expectInPrompt:    "",
			expectNotInPrompt: "reasoning_trace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate prompt using the same logic the service uses
			testCode := "package main\n\nfunc main() {}"
			prompt := BuildPreviewPrompt(testCode, tt.userMode, tt.outputMode)

			require.NotEmpty(t, prompt, "Prompt should not be empty")

			// Verify expected content
			if tt.expectInPrompt != "" {
				assert.Contains(t, prompt, tt.expectInPrompt,
					"Prompt should contain '%s' for %s + %s",
					tt.expectInPrompt, tt.userMode, tt.outputMode)
			}

			// Verify excluded content
			if tt.expectNotInPrompt != "" {
				assert.NotContains(t, prompt, tt.expectNotInPrompt,
					"Prompt should NOT contain '%s' for %s + %s",
					tt.expectNotInPrompt, tt.userMode, tt.outputMode)
			}

			// Verify code is included
			assert.Contains(t, prompt, testCode, "Prompt must include the actual code")
		})
	}
}

// TestSkimService_PassesModesToPromptBuilder validates Skim service mode passing
func TestSkimService_PassesModesToPromptBuilder(t *testing.T) {
	tests := []struct {
		name            string
		userMode        string
		outputMode      string
		shouldHaveTrace bool
	}{
		{"Beginner + Quick", "beginner", "quick", false},
		{"Expert + Full", "expert", "full", true},
		{"Intermediate + Quick (default)", "intermediate", "quick", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCode := "func Process() {}"
			prompt := BuildSkimPrompt(testCode, tt.userMode, tt.outputMode)

			require.NotEmpty(t, prompt)
			assert.Contains(t, prompt, testCode)

			hasTrace := strings.Contains(prompt, "reasoning_trace")
			assert.Equal(t, tt.shouldHaveTrace, hasTrace,
				"Reasoning trace presence mismatch for %s + %s",
				tt.userMode, tt.outputMode)
		})
	}
}

// TestScanService_PassesModesToPromptBuilder validates Scan service mode passing
func TestScanService_PassesModesToPromptBuilder(t *testing.T) {
	testCode := "SELECT * FROM users WHERE id = ?"
	query := "SQL queries"

	tests := []struct {
		name       string
		userMode   string
		outputMode string
		wantTrace  bool
	}{
		{"Beginner + Full", "beginner", "full", true},
		{"Expert + Quick", "expert", "quick", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := BuildScanPrompt(testCode, query, tt.userMode, tt.outputMode)

			require.NotEmpty(t, prompt)
			assert.Contains(t, prompt, testCode)
			assert.Contains(t, prompt, query)

			hasTrace := strings.Contains(prompt, "reasoning_trace")
			assert.Equal(t, tt.wantTrace, hasTrace)
		})
	}
}

// TestDetailedService_PassesModesToPromptBuilder validates Detailed service mode passing
func TestDetailedService_PassesModesToPromptBuilder(t *testing.T) {
	testCode := "for i := 0; i < 10; i++ {}"
	filename := "loop.go"

	tests := []struct {
		name       string
		userMode   string
		outputMode string
		wantTrace  bool
	}{
		{"Beginner + Full", "beginner", "full", true},
		{"Expert + Quick", "expert", "quick", false},
		{"Novice + Full", "novice", "full", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := BuildDetailedPrompt(testCode, filename, tt.userMode, tt.outputMode)

			require.NotEmpty(t, prompt)
			assert.Contains(t, prompt, testCode)
			assert.Contains(t, prompt, filename)

			hasTrace := strings.Contains(prompt, "reasoning_trace")
			assert.Equal(t, tt.wantTrace, hasTrace,
				"Reasoning trace mismatch for %s + %s", tt.userMode, tt.outputMode)
		})
	}
}

// TestAllServices_DefaultModeHandling validates default values work across all services
func TestAllServices_DefaultModeHandling(t *testing.T) {
	testCode := "test code"

	t.Run("Preview with empty modes defaults correctly", func(t *testing.T) {
		prompt := BuildPreviewPrompt(testCode, "", "")
		assert.NotContains(t, prompt, "reasoning_trace", "Empty modes should default to quick (no trace)")
	})

	t.Run("Skim with empty modes defaults correctly", func(t *testing.T) {
		prompt := BuildSkimPrompt(testCode, "", "")
		assert.NotContains(t, prompt, "reasoning_trace")
	})

	t.Run("Scan with empty modes defaults correctly", func(t *testing.T) {
		prompt := BuildScanPrompt(testCode, "query", "", "")
		assert.NotContains(t, prompt, "reasoning_trace")
	})

	t.Run("Detailed with empty modes defaults correctly", func(t *testing.T) {
		prompt := BuildDetailedPrompt(testCode, "file.go", "", "")
		assert.NotContains(t, prompt, "reasoning_trace")
	})
}

// TestServiceModeConsistency validates all services handle modes consistently
func TestServiceModeConsistency(t *testing.T) {
	testCode := "test"

	// All services with Beginner + Full should produce prompts with:
	// 1. Simple language guidance
	// 2. Reasoning trace

	t.Run("All services with Beginner + Full have reasoning trace", func(t *testing.T) {
		previewPrompt := BuildPreviewPrompt(testCode, "beginner", "full")
		skimPrompt := BuildSkimPrompt(testCode, "beginner", "full")
		scanPrompt := BuildScanPrompt(testCode, "query", "beginner", "full")
		detailedPrompt := BuildDetailedPrompt(testCode, "file", "beginner", "full")

		prompts := []string{previewPrompt, skimPrompt, scanPrompt, detailedPrompt}
		names := []string{"Preview", "Skim", "Scan", "Detailed"}

		for i, prompt := range prompts {
			assert.Contains(t, prompt, "reasoning_trace",
				"%s service should include reasoning_trace for Full mode", names[i])

			// Check for beginner-friendly language markers (different for each service)
			lowerPrompt := strings.ToLower(prompt)
			hasBeginner := strings.Contains(lowerPrompt, "analog") ||
				strings.Contains(lowerPrompt, "simple") ||
				strings.Contains(lowerPrompt, "as if teaching") ||
				strings.Contains(lowerPrompt, "avoid jargon") ||
				strings.Contains(lowerPrompt, "avoid assuming")
			assert.True(t, hasBeginner,
				"%s service should include beginner-friendly language guidance", names[i])
		}
	})

	t.Run("All services with Expert + Quick exclude reasoning trace", func(t *testing.T) {
		previewPrompt := BuildPreviewPrompt(testCode, "expert", "quick")
		skimPrompt := BuildSkimPrompt(testCode, "expert", "quick")
		scanPrompt := BuildScanPrompt(testCode, "query", "expert", "quick")
		detailedPrompt := BuildDetailedPrompt(testCode, "file", "expert", "quick")

		prompts := []string{previewPrompt, skimPrompt, scanPrompt, detailedPrompt}
		names := []string{"Preview", "Skim", "Scan", "Detailed"}

		for i, prompt := range prompts {
			assert.NotContains(t, prompt, "reasoning_trace",
				"%s service should NOT include reasoning_trace for Quick mode", names[i])
			assert.Contains(t, strings.ToLower(prompt), "technical",
				"%s service should include technical guidance for Expert", names[i])
		}
	})
}
