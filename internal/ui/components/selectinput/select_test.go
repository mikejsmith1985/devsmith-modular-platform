package selectinput

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelect_Render_Basic(t *testing.T) {
	// GIVEN: A basic select component with options
	props := SelectProps{
		Name:    "model",
		Label:   "AI Model",
		Options: []Option{{Value: "gpt4", Label: "GPT-4"}, {Value: "claude", Label: "Claude"}},
	}

	// WHEN: We render the select
	var buf bytes.Buffer
	ctx := context.Background()
	err := Select(props).Render(ctx, &buf)

	// THEN: The select should render with options
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "model")
	assert.Contains(t, content, "AI Model")
	assert.Contains(t, content, "gpt4")
	assert.Contains(t, content, "claude")
	assert.Contains(t, content, "GPT-4")
	assert.Contains(t, content, "Claude")
}

func TestSelect_Render_WithSelected(t *testing.T) {
	// GIVEN: A select with a pre-selected value
	props := SelectProps{
		Name:          "model",
		Label:         "AI Model",
		SelectedValue: "gpt4",
		Options:       []Option{{Value: "gpt4", Label: "GPT-4"}, {Value: "claude", Label: "Claude"}},
	}

	// WHEN: We render the select
	var buf bytes.Buffer
	ctx := context.Background()
	err := Select(props).Render(ctx, &buf)

	// THEN: The selected option should be marked as selected
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "selected")
	assert.Contains(t, content, "gpt4")
}

func TestSelect_Render_Disabled(t *testing.T) {
	// GIVEN: A disabled select
	props := SelectProps{
		Name:     "model",
		Label:    "AI Model",
		Disabled: true,
		Options:  []Option{{Value: "gpt4", Label: "GPT-4"}},
	}

	// WHEN: We render the select
	var buf bytes.Buffer
	ctx := context.Background()
	err := Select(props).Render(ctx, &buf)

	// THEN: The select should be disabled
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "disabled")
}

func TestSelect_Render_Required(t *testing.T) {
	// GIVEN: A required select
	props := SelectProps{
		Name:     "model",
		Label:    "AI Model",
		Required: true,
		Options:  []Option{{Value: "gpt4", Label: "GPT-4"}},
	}

	// WHEN: We render the select
	var buf bytes.Buffer
	ctx := context.Background()
	err := Select(props).Render(ctx, &buf)

	// THEN: The select should have required attribute
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "required")
	assert.Contains(t, content, "*")
}

func TestSelect_Render_WithError(t *testing.T) {
	// GIVEN: A select with error
	props := SelectProps{
		Name:     "model",
		Label:    "AI Model",
		Error:    "Please select a model",
		HasError: true,
		Options:  []Option{{Value: "gpt4", Label: "GPT-4"}},
	}

	// WHEN: We render the select
	var buf bytes.Buffer
	ctx := context.Background()
	err := Select(props).Render(ctx, &buf)

	// THEN: The error should be displayed
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Please select a model")
}

func TestSelect_Render_WithHelpText(t *testing.T) {
	// GIVEN: A select with help text
	props := SelectProps{
		Name:     "model",
		Label:    "AI Model",
		HelpText: "Select your preferred AI model for code analysis",
		Options:  []Option{{Value: "gpt4", Label: "GPT-4"}},
	}

	// WHEN: We render the select
	var buf bytes.Buffer
	ctx := context.Background()
	err := Select(props).Render(ctx, &buf)

	// THEN: The help text should be displayed
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Select your preferred AI model for code analysis")
}

func TestSelect_Render_WithPlaceholder(t *testing.T) {
	// GIVEN: A select with placeholder
	props := SelectProps{
		Name:        "model",
		Label:       "AI Model",
		Placeholder: "Choose a model...",
		Options:     []Option{{Value: "gpt4", Label: "GPT-4"}},
	}

	// WHEN: We render the select
	var buf bytes.Buffer
	ctx := context.Background()
	err := Select(props).Render(ctx, &buf)

	// THEN: The placeholder should be displayed
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Choose a model...")
}

func TestSelect_Render_Accessibility(t *testing.T) {
	// GIVEN: A select component
	props := SelectProps{
		Name:    "model",
		Label:   "AI Model",
		Options: []Option{{Value: "gpt4", Label: "GPT-4"}},
	}

	// WHEN: We render the select
	var buf bytes.Buffer
	ctx := context.Background()
	err := Select(props).Render(ctx, &buf)

	// THEN: The select should have accessibility attributes
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "aria-")
	assert.Contains(t, content, "role=")
}

func TestSelect_Render_MultipleOptions(t *testing.T) {
	// GIVEN: A select with many options
	props := SelectProps{
		Name:  "provider",
		Label: "Provider",
		Options: []Option{
			{Value: "ollama", Label: "Ollama (Local)"},
			{Value: "openai", Label: "OpenAI"},
			{Value: "anthropic", Label: "Anthropic"},
			{Value: "google", Label: "Google"},
		},
	}

	// WHEN: We render the select
	var buf bytes.Buffer
	ctx := context.Background()
	err := Select(props).Render(ctx, &buf)

	// THEN: All options should be present
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "ollama")
	assert.Contains(t, content, "openai")
	assert.Contains(t, content, "anthropic")
	assert.Contains(t, content, "google")
	assert.Contains(t, content, "Ollama (Local)")
	assert.Contains(t, content, "OpenAI")
	assert.Contains(t, content, "Anthropic")
	assert.Contains(t, content, "Google")
}
