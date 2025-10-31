package badge

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBadge_Render_Default(t *testing.T) {
	// GIVEN: A basic badge
	props := BadgeProps{Text: "Active"}

	// WHEN: We render it
	var buf bytes.Buffer
	err := Badge(props).Render(context.Background(), &buf)

	// THEN: Should render without error
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Active")
	assert.Contains(t, content, "role=\"status\"")
}

func TestBadge_Render_Variants(t *testing.T) {
	variants := []string{"success", "error", "warning", "info", "neutral"}

	for _, variant := range variants {
		t.Run(variant, func(t *testing.T) {
			props := BadgeProps{Text: "Test", Variant: variant}
			var buf bytes.Buffer
			err := Badge(props).Render(context.Background(), &buf)
			require.NoError(t, err)
			assert.Contains(t, buf.String(), "Test")
		})
	}
}

func TestBadge_Render_WithIcon(t *testing.T) {
	props := BadgeProps{Text: "Status", Icon: "✓"}
	var buf bytes.Buffer
	err := Badge(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "✓")
	assert.Contains(t, content, "Status")
}

func TestBadge_Render_Pill(t *testing.T) {
	props := BadgeProps{Text: "Tag", Pill: true}
	var buf bytes.Buffer
	err := Badge(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Tag")
}

func TestBadge_Render_Dismissible(t *testing.T) {
	props := BadgeProps{Text: "Closeable", Dismissible: true}
	var buf bytes.Buffer
	err := Badge(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Closeable")
	assert.Contains(t, content, "×") // Close button
}

func TestBadge_Render_Size(t *testing.T) {
	sizes := []string{"sm", "md", "lg"}

	for _, size := range sizes {
		t.Run(size, func(t *testing.T) {
			props := BadgeProps{Text: "Size", Size: size}
			var buf bytes.Buffer
			err := Badge(props).Render(context.Background(), &buf)
			require.NoError(t, err)
			assert.Contains(t, buf.String(), "Size")
		})
	}
}
