package button

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestButton_Render_Primary verifies primary button renders correctly
func TestButton_Render_Primary(t *testing.T) {
	props := ButtonProps{
		Label:   "Submit",
		Variant: VariantPrimary,
		Size:    SizeMedium,
		Type:    "submit",
	}

	var buf bytes.Buffer
	err := Button(props).Render(context.Background(), &buf)
	assert.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "devsmith-btn", "Should have base button class")
	assert.Contains(t, html, "btn-primary", "Should have primary variant class")
	assert.Contains(t, html, "btn-medium", "Should have medium size class")
	assert.Contains(t, html, "Submit", "Should have button label")
	assert.Contains(t, html, `type="submit"`, "Should have submit type")
}

// TestButton_Render_AllVariants verifies all button variants
func TestButton_Render_AllVariants(t *testing.T) {
	variants := map[string]ButtonVariant{
		"Primary":   VariantPrimary,
		"Secondary": VariantSecondary,
		"Danger":    VariantDanger,
		"Success":   VariantSuccess,
		"Outline":   VariantOutline,
	}

	for name, variant := range variants {
		t.Run(name, func(t *testing.T) {
			props := ButtonProps{
				Label:   "Test Button",
				Variant: variant,
				Size:    SizeMedium,
				Type:    "button",
			}

			var buf bytes.Buffer
			err := Button(props).Render(context.Background(), &buf)
			assert.NoError(t, err)

			html := buf.String()
			assert.Contains(t, html, "btn-"+string(variant), "Should contain variant class")
		})
	}
}

// TestButton_Render_AllSizes verifies all button sizes
func TestButton_Render_AllSizes(t *testing.T) {
	sizes := map[string]ButtonSize{
		"Small":  SizeSmall,
		"Medium": SizeMedium,
		"Large":  SizeLarge,
	}

	for name, size := range sizes {
		t.Run(name, func(t *testing.T) {
			props := ButtonProps{
				Label:   "Test Button",
				Variant: VariantPrimary,
				Size:    size,
				Type:    "button",
			}

			var buf bytes.Buffer
			err := Button(props).Render(context.Background(), &buf)
			assert.NoError(t, err)

			html := buf.String()
			assert.Contains(t, html, "btn-"+string(size), "Should contain size class")
		})
	}
}

// TestButton_Render_Loading verifies loading state
func TestButton_Render_Loading(t *testing.T) {
	props := ButtonProps{
		Label:   "Processing",
		Variant: VariantPrimary,
		Size:    SizeMedium,
		Type:    "button",
		Loading: true,
	}

	var buf bytes.Buffer
	err := Button(props).Render(context.Background(), &buf)
	assert.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "btn-loading", "Should have loading class")
	assert.Contains(t, html, "btn-spinner", "Should have spinner")
	assert.Contains(t, html, "disabled", "Should be disabled during loading")
}

// TestButton_Render_Disabled verifies disabled state
func TestButton_Render_Disabled(t *testing.T) {
	props := ButtonProps{
		Label:    "Disabled",
		Variant:  VariantPrimary,
		Size:     SizeMedium,
		Type:     "button",
		Disabled: true,
	}

	var buf bytes.Buffer
	err := Button(props).Render(context.Background(), &buf)
	assert.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "btn-disabled", "Should have disabled class")
	assert.Contains(t, html, "disabled", "Should have disabled attribute")
}

// TestButton_Render_WithIcon verifies button with icon
func TestButton_Render_WithIcon(t *testing.T) {
	props := ButtonProps{
		Label:   "Search",
		Variant: VariantPrimary,
		Size:    SizeMedium,
		Type:    "button",
		Icon:    "icon-search",
	}

	var buf bytes.Buffer
	err := Button(props).Render(context.Background(), &buf)
	assert.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "btn-icon", "Should have icon container")
	assert.Contains(t, html, "icon-search", "Should have icon class")
}

// TestButton_Render_FullWidth verifies full width button
func TestButton_Render_FullWidth(t *testing.T) {
	props := ButtonProps{
		Label:     "Full Width",
		Variant:   VariantPrimary,
		Size:      SizeMedium,
		Type:      "button",
		FullWidth: true,
	}

	var buf bytes.Buffer
	err := Button(props).Render(context.Background(), &buf)
	assert.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "btn-full-width", "Should have full-width class")
}

// TestButton_Render_WithID verifies button with ID attribute
func TestButton_Render_WithID(t *testing.T) {
	props := ButtonProps{
		Label:   "Button with ID",
		Variant: VariantPrimary,
		Size:    SizeMedium,
		Type:    "button",
		ID:      "submit-btn",
	}

	var buf bytes.Buffer
	err := Button(props).Render(context.Background(), &buf)
	assert.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, `id="submit-btn"`, "Should have id attribute")
}

// TestButton_Render_WithAriaLabel verifies accessibility label
func TestButton_Render_WithAriaLabel(t *testing.T) {
	props := ButtonProps{
		Label:     "X",
		Variant:   VariantPrimary,
		Size:      SizeSmall,
		Type:      "button",
		AriaLabel: "Close dialog",
	}

	var buf bytes.Buffer
	err := Button(props).Render(context.Background(), &buf)
	assert.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, `aria-label="Close dialog"`, "Should have aria-label")
}

// TestButton_Render_WithCustomClass verifies custom CSS classes
func TestButton_Render_WithCustomClass(t *testing.T) {
	props := ButtonProps{
		Label:   "Custom Button",
		Variant: VariantPrimary,
		Size:    SizeMedium,
		Type:    "button",
		Class:   "custom-class",
	}

	var buf bytes.Buffer
	err := Button(props).Render(context.Background(), &buf)
	assert.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "custom-class", "Should include custom class")
}

// TestButton_Render_AllStates verifies all button states
func TestButton_Render_AllStates(t *testing.T) {
	states := []struct {
		name      string
		disabled  bool
		loading   bool
		expected  string
	}{
		{"Normal", false, false, "btn-normal"},
		{"Disabled", true, false, "btn-disabled"},
		{"Loading", false, true, "btn-loading"},
		{"Loading overrides disabled", true, true, "btn-loading"},
	}

	for _, s := range states {
		t.Run(s.name, func(t *testing.T) {
			props := ButtonProps{
				Label:    "Test",
				Variant:  VariantPrimary,
				Size:     SizeMedium,
				Type:     "button",
				Disabled: s.disabled,
				Loading:  s.loading,
			}

			var buf bytes.Buffer
			err := Button(props).Render(context.Background(), &buf)
			assert.NoError(t, err)

			html := buf.String()
			assert.Contains(t, html, s.expected, "Should have correct state class")
		})
	}
}

// TestButton_Render_ButtonType verifies button HTML type
func TestButton_Render_ButtonType(t *testing.T) {
	types := []string{"button", "submit", "reset"}

	for _, btnType := range types {
		t.Run(btnType, func(t *testing.T) {
			props := ButtonProps{
				Label:   "Button",
				Variant: VariantPrimary,
				Size:    SizeMedium,
				Type:    btnType,
			}

			var buf bytes.Buffer
			err := Button(props).Render(context.Background(), &buf)
			assert.NoError(t, err)

			html := buf.String()
			assert.Contains(t, html, `type="`+btnType+`"`, "Should have correct button type")
		})
	}
}

// TestButton_Render_Structure verifies HTML structure
func TestButton_Render_Structure(t *testing.T) {
	props := ButtonProps{
		Label:   "Test",
		Variant: VariantPrimary,
		Size:    SizeMedium,
		Type:    "button",
	}

	var buf bytes.Buffer
	err := Button(props).Render(context.Background(), &buf)
	assert.NoError(t, err)

	html := buf.String()
	// Should be a button element
	assert.True(t, strings.HasPrefix(strings.TrimSpace(html), "<button"),
		"Should render as button element")
	assert.Contains(t, html, "</button>", "Should have closing button tag")
	// Should have label span
	assert.Contains(t, html, `<span class="btn-label">`, "Should have label span")
}

// TestButton_Render_DangerVariant verifies danger button styling
func TestButton_Render_DangerVariant(t *testing.T) {
	props := ButtonProps{
		Label:   "Delete",
		Variant: VariantDanger,
		Size:    SizeMedium,
		Type:    "button",
	}

	var buf bytes.Buffer
	err := Button(props).Render(context.Background(), &buf)
	assert.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "btn-danger", "Should have danger variant")
}

// TestButton_Render_SuccessVariant verifies success button styling
func TestButton_Render_SuccessVariant(t *testing.T) {
	props := ButtonProps{
		Label:   "Confirm",
		Variant: VariantSuccess,
		Size:    SizeMedium,
		Type:    "button",
	}

	var buf bytes.Buffer
	err := Button(props).Render(context.Background(), &buf)
	assert.NoError(t, err)

	html := buf.String()
	assert.Contains(t, html, "btn-success", "Should have success variant")
}

// TestButton_Render_AccessibilityClasses verifies accessibility features
func TestButton_Render_AccessibilityClasses(t *testing.T) {
	props := ButtonProps{
		Label:     "Accessible Button",
		Variant:   VariantPrimary,
		Size:      SizeMedium,
		Type:      "button",
		AriaLabel: "Important action",
	}

	var buf bytes.Buffer
	err := Button(props).Render(context.Background(), &buf)
	assert.NoError(t, err)

	html := buf.String()
	// Should be focusable
	assert.True(t, strings.Contains(html, "<button"), "Should be button element (naturally focusable)")
	// Should have aria label
	assert.Contains(t, html, "aria-label", "Should have ARIA label for accessibility")
}

// TestButtonClass_GeneratesCorrectClass verifies class generation function
func TestButtonClass_GeneratesCorrectClass(t *testing.T) {
	tests := []struct {
		name     string
		props    ButtonProps
		state    ButtonState
		expected string
	}{
		{
			"Basic primary button",
			ButtonProps{Variant: VariantPrimary, Size: SizeMedium},
			StateNormal,
			"devsmith-btn btn-primary btn-medium btn-normal",
		},
		{
			"Full width button",
			ButtonProps{Variant: VariantSecondary, Size: SizeSmall, FullWidth: true},
			StateNormal,
			"devsmith-btn btn-secondary btn-small btn-normal btn-full-width",
		},
		{
			"Button with custom class",
			ButtonProps{Variant: VariantDanger, Size: SizeLarge, Class: "my-custom"},
			StateLoading,
			"devsmith-btn btn-danger btn-large btn-loading btn-full-width my-custom",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Adjust FullWidth expectation for simple tests
			if !test.props.FullWidth {
				test.expected = strings.ReplaceAll(test.expected, " btn-full-width", "")
			}

			result := buttonClass(test.props, test.state)
			// Should contain all expected parts
			for _, part := range []string{
				"devsmith-btn",
				"btn-" + string(test.props.Variant),
				"btn-" + string(test.props.Size),
				"btn-" + string(test.state),
			} {
				assert.Contains(t, result, part, "Should contain %s", part)
			}
		})
	}
}
