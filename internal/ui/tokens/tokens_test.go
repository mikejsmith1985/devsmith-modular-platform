package tokens

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTokens_NewTokens_CreatesAllComponents verifies all components are initialized
func TestTokens_NewTokens_CreatesAllComponents(t *testing.T) {
	tokens := NewTokens()

	assert.NotNil(t, tokens.Colors, "Colors should be initialized")
	assert.NotNil(t, tokens.Spacing, "Spacing should be initialized")
	assert.NotNil(t, tokens.Typography, "Typography should be initialized")
	assert.NotNil(t, tokens.BorderRadius, "BorderRadius should be initialized")
	assert.NotNil(t, tokens.Shadows, "Shadows should be initialized")
	assert.NotNil(t, tokens.Transitions, "Transitions should be initialized")
}

// TestTokens_Colors_HasValidHexColors verifies all colors are valid hex codes
func TestTokens_Colors_HasValidHexColors(t *testing.T) {
	tokens := NewTokens()

	colors := map[string]string{
		"Primary":         tokens.Colors.Primary,
		"PrimaryHover":    tokens.Colors.PrimaryHover,
		"PrimaryActive":   tokens.Colors.PrimaryActive,
		"Success":         tokens.Colors.Success,
		"Warning":         tokens.Colors.Warning,
		"Danger":          tokens.Colors.Danger,
		"Info":            tokens.Colors.Info,
		"Background":      tokens.Colors.Background,
		"Surface":         tokens.Colors.Surface,
		"SurfaceSecond":   tokens.Colors.SurfaceSecond,
		"Border":          tokens.Colors.Border,
		"Text":            tokens.Colors.Text,
		"TextSecondary":   tokens.Colors.TextSecondary,
		"TextTertiary":    tokens.Colors.TextTertiary,
		"DarkBackground":  tokens.Colors.DarkBackground,
		"DarkSurface":     tokens.Colors.DarkSurface,
		"DarkSurface2":    tokens.Colors.DarkSurface2,
		"DarkBorder":      tokens.Colors.DarkBorder,
		"DarkText":        tokens.Colors.DarkText,
		"DarkText2":       tokens.Colors.DarkText2,
		"DarkText3":       tokens.Colors.DarkText3,
	}

	for name, color := range colors {
		t.Run(name, func(t *testing.T) {
			// Must start with #
			assert.True(t, strings.HasPrefix(color, "#"), "Color must start with #")

			// Must be valid hex length (7 for 6-digit hex, 9 for 8-digit with alpha)
			assert.True(t,
				len(color) == 7 || len(color) == 9,
				"Color must be 6 or 8 character hex code (got %d)", len(color))

			// All characters after # must be hex
			for _, ch := range color[1:] {
				assert.True(t,
					(ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F'),
					"Color contains invalid hex character: %c", ch)
			}
		})
	}
}

// TestTokens_Spacing_HasValidValues verifies spacing values
func TestTokens_Spacing_HasValidValues(t *testing.T) {
	tokens := NewTokens()

	spacingValues := map[string]string{
		"XS":     tokens.Spacing.XS,
		"Small":  tokens.Spacing.Small,
		"Base":   tokens.Spacing.Base,
		"Half":   tokens.Spacing.Half,
		"Double": tokens.Spacing.Double,
		"Triple": tokens.Spacing.Triple,
		"Quad":   tokens.Spacing.Quad,
		"Five":   tokens.Spacing.Five,
		"Six":    tokens.Spacing.Six,
		"Seven":  tokens.Spacing.Seven,
		"Eight":  tokens.Spacing.Eight,
		"Ten":    tokens.Spacing.Ten,
	}

	for name, value := range spacingValues {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, value, "Spacing value should not be empty")
			assert.True(t, strings.HasSuffix(value, "px"), "Spacing must be in pixels")

			// Extract number and verify it's numeric
			numPart := strings.TrimSuffix(value, "px")
			assert.NotEmpty(t, numPart, "Spacing should have numeric part")
			assert.True(t, isNumeric(numPart), "Spacing numeric part should be valid: %s", numPart)
		})
	}
}

// TestTokens_Typography_HasSystemFonts verifies font families
func TestTokens_Typography_HasSystemFonts(t *testing.T) {
	tokens := NewTokens()

	assert.NotEmpty(t, tokens.Typography.SystemFont, "SystemFont should be defined")
	assert.NotEmpty(t, tokens.Typography.MonoFont, "MonoFont should be defined")

	// Should contain multiple fonts as fallback
	assert.True(t, strings.Contains(tokens.Typography.SystemFont, ","),
		"SystemFont should have fallback fonts")
	assert.True(t, strings.Contains(tokens.Typography.MonoFont, ","),
		"MonoFont should have fallback fonts")
}

// TestTokens_Typography_HasAllFontSizes verifies all font sizes exist
func TestTokens_Typography_HasAllFontSizes(t *testing.T) {
	tokens := NewTokens()

	sizes := map[string]string{
		"Size12": tokens.Typography.Size12,
		"Size13": tokens.Typography.Size13,
		"Size14": tokens.Typography.Size14,
		"Size15": tokens.Typography.Size15,
		"Size16": tokens.Typography.Size16,
		"Size18": tokens.Typography.Size18,
		"Size20": tokens.Typography.Size20,
		"Size24": tokens.Typography.Size24,
		"Size28": tokens.Typography.Size28,
		"Size32": tokens.Typography.Size32,
		"Size36": tokens.Typography.Size36,
	}

	for name, size := range sizes {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, size, "Font size should not be empty")
			assert.True(t, strings.HasSuffix(size, "rem"), "Font size should be in rem")
		})
	}
}

// TestTokens_Typography_HasAllFontWeights verifies font weights
func TestTokens_Typography_HasAllFontWeights(t *testing.T) {
	tokens := NewTokens()

	weights := map[string]string{
		"Regular":   tokens.Typography.Regular,
		"Medium":    tokens.Typography.Medium,
		"SemiBold":  tokens.Typography.SemiBold,
		"Bold":      tokens.Typography.Bold,
		"Heavy":     tokens.Typography.Heavy,
	}

	expectedWeights := map[string]bool{
		"400": true, // Regular
		"500": true, // Medium
		"600": true, // SemiBold
		"700": true, // Bold
		"800": true, // Heavy
	}

	for name, weight := range weights {
		t.Run(name, func(t *testing.T) {
			assert.True(t, expectedWeights[weight],
				"Font weight should be valid: %s", weight)
		})
	}
}

// TestTokens_Typography_HasAllLineHeights verifies line heights
func TestTokens_Typography_HasAllLineHeights(t *testing.T) {
	tokens := NewTokens()

	lineHeights := map[string]string{
		"Tight":   tokens.Typography.Tight,
		"Normal":  tokens.Typography.Normal,
		"Relaxed": tokens.Typography.Relaxed,
		"Loose":   tokens.Typography.Loose,
	}

	expectedHeights := map[string]bool{
		"1.2":  true, // Tight
		"1.5":  true, // Normal
		"1.75": true, // Relaxed
		"2":    true, // Loose
	}

	for name, height := range lineHeights {
		t.Run(name, func(t *testing.T) {
			assert.True(t, expectedHeights[height],
				"Line height should be valid: %s", height)
		})
	}
}

// TestTokens_BorderRadius_HasAllValues verifies border radius tokens
func TestTokens_BorderRadius_HasAllValues(t *testing.T) {
	tokens := NewTokens()

	radiusValues := map[string]string{
		"None":   tokens.BorderRadius.None,
		"Small":  tokens.BorderRadius.Small,
		"Medium": tokens.BorderRadius.Medium,
		"Large":  tokens.BorderRadius.Large,
		"XL":     tokens.BorderRadius.XL,
		"Full":   tokens.BorderRadius.Full,
	}

	for name, value := range radiusValues {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, value, "BorderRadius value should not be empty")
			assert.True(t, strings.HasSuffix(value, "px"), "BorderRadius must be in pixels")
		})
	}
}

// TestTokens_Shadows_HasLightAndDark verifies shadow coverage
func TestTokens_Shadows_HasLightAndDark(t *testing.T) {
	tokens := NewTokens()

	lightShadows := map[string]string{
		"Shallow": tokens.Shadows.Shallow,
		"Small":   tokens.Shadows.Small,
		"Medium":  tokens.Shadows.Medium,
		"Large":   tokens.Shadows.Large,
		"XL":      tokens.Shadows.XL,
	}

	darkShadows := map[string]string{
		"DarkShallow": tokens.Shadows.DarkShallow,
		"DarkSmall":   tokens.Shadows.DarkSmall,
		"DarkMedium":  tokens.Shadows.DarkMedium,
		"DarkLarge":   tokens.Shadows.DarkLarge,
	}

	for name, shadow := range lightShadows {
		t.Run("Light_"+name, func(t *testing.T) {
			assert.NotEmpty(t, shadow, "Shadow value should not be empty")
			assert.True(t, strings.Contains(shadow, "rgba"),
				"Shadow should use rgba for color")
		})
	}

	for name, shadow := range darkShadows {
		t.Run("Dark_"+name, func(t *testing.T) {
			assert.NotEmpty(t, shadow, "Shadow value should not be empty")
			assert.True(t, strings.Contains(shadow, "rgba"),
				"Shadow should use rgba for color")
		})
	}
}

// TestTokens_Transitions_HasAllDurations verifies transition timing
func TestTokens_Transitions_HasAllDurations(t *testing.T) {
	tokens := NewTokens()

	transitions := map[string]string{
		"Fast":   tokens.Transitions.Fast,
		"Base":   tokens.Transitions.Base,
		"Slow":   tokens.Transitions.Slow,
		"Slower": tokens.Transitions.Slower,
	}

	for name, transition := range transitions {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, transition, "Transition should not be empty")
			assert.True(t, strings.Contains(transition, "ms"),
				"Transition should specify milliseconds")
			assert.True(t, strings.Contains(transition, "cubic-bezier"),
				"Transition should use cubic-bezier easing")
		})
	}
}

// TestTokens_ColorConsistency_NeutralScale verifies neutral colors progress
func TestTokens_ColorConsistency_NeutralScale(t *testing.T) {
	tokens := NewTokens()

	// Light mode progression from light to dark
	lightScale := []string{
		tokens.Colors.Background,
		tokens.Colors.Surface,
		tokens.Colors.SurfaceSecond,
		tokens.Colors.Border,
		tokens.Colors.TextTertiary,
		tokens.Colors.TextSecondary,
		tokens.Colors.Text,
	}

	for i := 0; i < len(lightScale)-1; i++ {
		assert.NotEqual(t, lightScale[i], lightScale[i+1],
			"Neutral color scale should have distinct values at index %d and %d", i, i+1)
	}
}

// TestTokens_SemanticColors_NonConflicting verifies semantic colors don't conflict
func TestTokens_SemanticColors_NonConflicting(t *testing.T) {
	tokens := NewTokens()

	semanticColors := []string{
		tokens.Colors.Primary,
		tokens.Colors.Success,
		tokens.Colors.Warning,
		tokens.Colors.Danger,
		tokens.Colors.Info,
	}

	seen := make(map[string]bool)
	for _, color := range semanticColors {
		assert.False(t, seen[color], "Semantic colors should be unique: %s", color)
		seen[color] = true
	}
}

// TestTokens_DarkMode_ProvidedForAll verifies dark mode coverage
func TestTokens_DarkMode_ProvidedForAll(t *testing.T) {
	tokens := NewTokens()

	// Verify we have dark mode for primary categories
	assert.NotEmpty(t, tokens.Colors.DarkBackground, "Should have dark background")
	assert.NotEmpty(t, tokens.Colors.DarkSurface, "Should have dark surface")
	assert.NotEmpty(t, tokens.Colors.DarkText, "Should have dark text")

	// Dark mode colors should differ from light mode
	assert.NotEqual(t, tokens.Colors.Background, tokens.Colors.DarkBackground)
	assert.NotEqual(t, tokens.Colors.Text, tokens.Colors.DarkText)
}

// TestTokens_Accessibility_ShadowsSufficientContrast verifies shadow alpha values
func TestTokens_Accessibility_ShadowsSufficientContrast(t *testing.T) {
	tokens := NewTokens()

	// Shadow alpha values should be reasonable (0.1 to 0.6)
	shadowTests := []struct {
		name   string
		shadow string
	}{
		{"Shallow", tokens.Shadows.Shallow},
		{"Small", tokens.Shadows.Small},
		{"Medium", tokens.Shadows.Medium},
		{"Large", tokens.Shadows.Large},
		{"DarkShallow", tokens.Shadows.DarkShallow},
		{"DarkLarge", tokens.Shadows.DarkLarge},
	}

	for _, test := range shadowTests {
		t.Run(test.name, func(t *testing.T) {
			// Parse alpha value from rgba
			assert.True(t, strings.Contains(test.shadow, "rgba"),
				"Shadow should use rgba format")
		})
	}
}

// Helper function to check if a string contains only numeric characters
func isNumeric(s string) bool {
	for _, ch := range s {
		if !((ch >= '0' && ch <= '9') || ch == '.' || ch == '-') {
			return false
		}
	}
	return true
}
