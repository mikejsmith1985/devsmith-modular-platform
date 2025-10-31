package toast

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToast_Render_Success(t *testing.T) {
	props := ToastProps{
		Type:    "success",
		Title:   "Success",
		Message: "Operation completed successfully",
	}

	var buf bytes.Buffer
	err := Toast(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Success")
	assert.Contains(t, content, "Operation completed successfully")
	assert.Contains(t, content, "role=\"alert\"")
}

func TestToast_Render_Error(t *testing.T) {
	props := ToastProps{
		Type:    "error",
		Title:   "Error",
		Message: "Something went wrong",
	}

	var buf bytes.Buffer
	err := Toast(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Error")
}

func TestToast_Render_WithDismiss(t *testing.T) {
	props := ToastProps{
		Type:        "info",
		Title:       "Info",
		Message:     "Information",
		Dismissible: true,
	}

	var buf bytes.Buffer
	err := Toast(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Info")
	assert.Contains(t, content, "dismiss notification")
}

func TestToast_Render_AllTypes(t *testing.T) {
	types := []string{"success", "error", "warning", "info"}

	for _, toastType := range types {
		t.Run(toastType, func(t *testing.T) {
			props := ToastProps{
				Type:    toastType,
				Title:   "Test",
				Message: "Test message",
			}

			var buf bytes.Buffer
			err := Toast(props).Render(context.Background(), &buf)
			require.NoError(t, err)
			assert.Contains(t, buf.String(), "Test")
		})
	}
}
