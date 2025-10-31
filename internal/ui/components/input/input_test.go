package input

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInput_Render_Text(t *testing.T) {
	// GIVEN: A text input component
	props := InputProps{
		Type:        "text",
		Name:        "username",
		Placeholder: "Enter username",
		Label:       "Username",
	}

	// WHEN: We render the input
	var buf bytes.Buffer
	ctx := context.Background()
	err := Input(props).Render(ctx, &buf)

	// THEN: The input should render without error
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "username", "Input should contain name attribute")
	assert.Contains(t, content, "Enter username", "Input should contain placeholder")
	assert.Contains(t, content, "Username", "Input should contain label")
	assert.Contains(t, content, "type=\"text\"", "Input should have type text")
}

func TestInput_Render_Password(t *testing.T) {
	// GIVEN: A password input component
	props := InputProps{
		Type:        "password",
		Name:        "password",
		Placeholder: "Enter password",
		Label:       "Password",
	}

	// WHEN: We render the input
	var buf bytes.Buffer
	ctx := context.Background()
	err := Input(props).Render(ctx, &buf)

	// THEN: The input should render as password type
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "type=\"password\"", "Input should have type password")
}

func TestInput_Render_Email(t *testing.T) {
	// GIVEN: An email input component
	props := InputProps{
		Type:        "email",
		Name:        "email",
		Placeholder: "your@email.com",
		Label:       "Email",
	}

	// WHEN: We render the input
	var buf bytes.Buffer
	ctx := context.Background()
	err := Input(props).Render(ctx, &buf)

	// THEN: The input should render as email type
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "type=\"email\"", "Input should have type email")
}

func TestInput_Render_WithValue(t *testing.T) {
	// GIVEN: An input with initial value
	props := InputProps{
		Type:  "text",
		Name:  "title",
		Label: "Title",
		Value: "My Title",
	}

	// WHEN: We render the input
	var buf bytes.Buffer
	ctx := context.Background()
	err := Input(props).Render(ctx, &buf)

	// THEN: The value should be set
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "My Title", "Input should contain value")
}

func TestInput_Render_Disabled(t *testing.T) {
	// GIVEN: A disabled input
	props := InputProps{
		Type:     "text",
		Name:     "disabled_field",
		Label:    "Disabled Field",
		Disabled: true,
	}

	// WHEN: We render the input
	var buf bytes.Buffer
	ctx := context.Background()
	err := Input(props).Render(ctx, &buf)

	// THEN: The input should be disabled
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "disabled", "Input should be disabled")
}

func TestInput_Render_Required(t *testing.T) {
	// GIVEN: A required input
	props := InputProps{
		Type:     "text",
		Name:     "required_field",
		Label:    "Required Field",
		Required: true,
	}

	// WHEN: We render the input
	var buf bytes.Buffer
	ctx := context.Background()
	err := Input(props).Render(ctx, &buf)

	// THEN: The input should have required attribute
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "required", "Input should be required")
}

func TestInput_Render_WithError(t *testing.T) {
	// GIVEN: An input with error message
	props := InputProps{
		Type:     "text",
		Name:     "error_field",
		Label:    "Error Field",
		Error:    "This field is required",
		HasError: true,
	}

	// WHEN: We render the input
	var buf bytes.Buffer
	ctx := context.Background()
	err := Input(props).Render(ctx, &buf)

	// THEN: The error message should be displayed
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "This field is required", "Input should contain error message")
}

func TestInput_Render_WithHelpText(t *testing.T) {
	// GIVEN: An input with help text
	props := InputProps{
		Type:     "text",
		Name:     "help_field",
		Label:    "Help Field",
		HelpText: "This is helpful information",
	}

	// WHEN: We render the input
	var buf bytes.Buffer
	ctx := context.Background()
	err := Input(props).Render(ctx, &buf)

	// THEN: The help text should be displayed
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "This is helpful information", "Input should contain help text")
}

func TestInput_Render_Accessibility(t *testing.T) {
	// GIVEN: An input component
	props := InputProps{
		Type:  "text",
		Name:  "accessible",
		Label: "Accessible Input",
	}

	// WHEN: We render the input
	var buf bytes.Buffer
	ctx := context.Background()
	err := Input(props).Render(ctx, &buf)

	// THEN: The input should have accessibility attributes
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "aria-", "Input should have aria attributes")
}

func TestInput_Render_AllPropsPopulated(t *testing.T) {
	// GIVEN: An input with all props (no error state so help text shows)
	props := InputProps{
		Type:        "email",
		Name:        "full_input",
		Label:       "Full Input",
		Placeholder: "placeholder@example.com",
		Value:       "current@example.com",
		Required:    true,
		Disabled:    false,
		Error:       "",
		HasError:    false,
		HelpText:    "Please enter a valid email",
	}

	// WHEN: We render the input
	var buf bytes.Buffer
	ctx := context.Background()
	err := Input(props).Render(ctx, &buf)

	// THEN: All elements should be present
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "type=\"email\"")
	assert.Contains(t, content, "Full Input")
	assert.Contains(t, content, "current@example.com")
	assert.Contains(t, content, "required")
	assert.Contains(t, content, "Please enter a valid email")
}
