package modal

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModal_Render_Basic(t *testing.T) {
	props := ModalProps{
		Title:   "Confirm Action",
		Content: "Are you sure?",
	}

	var buf bytes.Buffer
	err := Modal(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Confirm Action")
	assert.Contains(t, content, "Are you sure?")
	assert.Contains(t, content, "role=\"dialog\"")
}

func TestModal_Render_WithButtons(t *testing.T) {
	props := ModalProps{
		Title:   "Delete Item",
		Content: "This cannot be undone",
		PrimaryAction: &Action{
			Label: "Delete",
			ID:    "delete-btn",
		},
		SecondaryAction: &Action{
			Label: "Cancel",
			ID:    "cancel-btn",
		},
	}

	var buf bytes.Buffer
	err := Modal(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Delete")
	assert.Contains(t, content, "Cancel")
	assert.Contains(t, content, "delete-btn")
	assert.Contains(t, content, "cancel-btn")
}

func TestModal_Render_Closeable(t *testing.T) {
	props := ModalProps{
		Title:      "Info",
		Content:    "Information",
		Closeable:  true,
		CloseLabel: "Close",
	}

	var buf bytes.Buffer
	err := Modal(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Close")
}

func TestModal_Render_Sizes(t *testing.T) {
	sizes := []string{"sm", "md", "lg", "xl"}

	for _, size := range sizes {
		t.Run(size, func(t *testing.T) {
			props := ModalProps{
				Title:   "Test",
				Content: "Content",
				Size:    size,
			}

			var buf bytes.Buffer
			err := Modal(props).Render(context.Background(), &buf)
			require.NoError(t, err)
		})
	}
}

func TestModal_Render_Scrollable(t *testing.T) {
	props := ModalProps{
		Title:      "Long Content",
		Content:    "This is a very long content that requires scrolling",
		Scrollable: true,
	}

	var buf bytes.Buffer
	err := Modal(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Long Content")
}

func TestModal_Render_WithImage(t *testing.T) {
	props := ModalProps{
		Title:       "Visual Content",
		Content:     "Image modal",
		ImageURL:    "/images/preview.png",
		ImageAlt:    "Preview image",
		HasImage:    true,
	}

	var buf bytes.Buffer
	err := Modal(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "/images/preview.png")
	assert.Contains(t, content, "Preview image")
}

func TestModal_Render_Danger(t *testing.T) {
	props := ModalProps{
		Title:    "Delete Everything",
		Content:  "This will permanently delete all data",
		IsDanger: true,
	}

	var buf bytes.Buffer
	err := Modal(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Delete Everything")
}
