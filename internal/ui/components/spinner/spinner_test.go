package spinner

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpinner_Render_Default(t *testing.T) {
	props := SpinnerProps{}

	var buf bytes.Buffer
	err := Spinner(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "role=\"status\"")
	assert.Contains(t, content, "Loading")
}

func TestSpinner_Render_Sizes(t *testing.T) {
	sizes := []string{"sm", "md", "lg"}

	for _, size := range sizes {
		t.Run(size, func(t *testing.T) {
			props := SpinnerProps{Size: size}
			var buf bytes.Buffer
			err := Spinner(props).Render(context.Background(), &buf)
			require.NoError(t, err)
			assert.NoError(t, err)
		})
	}
}

func TestSpinner_Render_WithText(t *testing.T) {
	props := SpinnerProps{
		Text: "Processing...",
	}

	var buf bytes.Buffer
	err := Spinner(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Processing...")
}

func TestSpinner_Render_Variants(t *testing.T) {
	variants := []string{"primary", "secondary", "success", "danger"}

	for _, variant := range variants {
		t.Run(variant, func(t *testing.T) {
			props := SpinnerProps{Variant: variant}
			var buf bytes.Buffer
			err := Spinner(props).Render(context.Background(), &buf)
			require.NoError(t, err)
		})
	}
}

func TestSpinner_Render_FullPage(t *testing.T) {
	props := SpinnerProps{
		FullPage: true,
		Text:     "Loading application...",
	}

	var buf bytes.Buffer
	err := Spinner(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Loading application...")
	assert.Contains(t, content, "fixed")
}
