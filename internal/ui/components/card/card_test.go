package card

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCard_Render_Basic(t *testing.T) {
	// GIVEN: A Card component with basic props
	props := CardProps{
		Title:       "Test Card",
		Description: "This is a test card",
	}

	// WHEN: We render the card
	var buf bytes.Buffer
	ctx := context.Background()
	err := Card(props).Render(ctx, &buf)

	// THEN: The card should render without error
	require.NoError(t, err, "Card should render without error")
	content := buf.String()
	assert.Contains(t, content, "Test Card", "Card should contain title")
	assert.Contains(t, content, "This is a test card", "Card should contain description")
}

func TestCard_Render_WithIcon(t *testing.T) {
	// GIVEN: A Card component with an icon
	props := CardProps{
		Title:       "Service Card",
		Description: "AI-Powered Code Review",
		Icon:        "🤖",
	}

	// WHEN: We render the card
	var buf bytes.Buffer
	ctx := context.Background()
	err := Card(props).Render(ctx, &buf)

	// THEN: The icon should be displayed
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "🤖", "Card should contain icon")
}

func TestCard_Render_WithBadge(t *testing.T) {
	// GIVEN: A Card component with a status badge
	props := CardProps{
		Title:       "Service Status",
		Description: "Current system status",
		BadgeText:   "Online",
		BadgeColor:  "green",
	}

	// WHEN: We render the card
	var buf bytes.Buffer
	ctx := context.Background()
	err := Card(props).Render(ctx, &buf)

	// THEN: The badge should be displayed
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Online", "Card should contain badge text")
	assert.Contains(t, content, "green", "Card should contain badge color")
}

func TestCard_Render_WithAction(t *testing.T) {
	// GIVEN: A Card component with an action button
	props := CardProps{
		Title:       "Review Service",
		Description: "Start code review",
		ActionText:  "Launch",
		ActionURL:   "http://localhost:8081",
	}

	// WHEN: We render the card
	var buf bytes.Buffer
	ctx := context.Background()
	err := Card(props).Render(ctx, &buf)

	// THEN: The action button should be displayed with correct link
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Launch", "Card should contain action button text")
	assert.Contains(t, content, "http://localhost:8081", "Card should contain action URL")
}

func TestCard_Render_WithStats(t *testing.T) {
	// GIVEN: A Card component with stats
	props := CardProps{
		Title:       "Analytics",
		Description: "Usage analytics",
		StatLabel:   "Last check",
		StatValue:   "2 hours ago",
	}

	// WHEN: We render the card
	var buf bytes.Buffer
	ctx := context.Background()
	err := Card(props).Render(ctx, &buf)

	// THEN: The stats should be displayed
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "2 hours ago", "Card should contain stat value")
}

func TestCard_Render_Accessibility(t *testing.T) {
	// GIVEN: A Card component
	props := CardProps{
		Title:       "Accessible Card",
		Description: "Test accessibility",
	}

	// WHEN: We render the card
	var buf bytes.Buffer
	ctx := context.Background()
	err := Card(props).Render(ctx, &buf)

	// THEN: The card should have proper accessibility attributes
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "role=", "Card should have role attribute")
}

func TestCard_Render_EmptyDescription(t *testing.T) {
	// GIVEN: A Card component with no description
	props := CardProps{
		Title: "Simple Card",
	}

	// WHEN: We render the card
	var buf bytes.Buffer
	ctx := context.Background()
	err := Card(props).Render(ctx, &buf)

	// THEN: The card should render without description
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Simple Card", "Card should contain title")
}

func TestCard_Render_AllPropsPopulated(t *testing.T) {
	// GIVEN: A Card component with all props
	props := CardProps{
		Title:       "Full Card",
		Description: "All props populated",
		Icon:        "✓",
		BadgeText:   "Active",
		BadgeColor:  "blue",
		ActionText:  "View",
		ActionURL:   "/path",
		StatLabel:   "Items",
		StatValue:   "42",
	}

	// WHEN: We render the card
	var buf bytes.Buffer
	ctx := context.Background()
	err := Card(props).Render(ctx, &buf)

	// THEN: All elements should be present
	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Full Card")
	assert.Contains(t, content, "All props populated")
	assert.Contains(t, content, "✓")
	assert.Contains(t, content, "Active")
	assert.Contains(t, content, "View")
	assert.Contains(t, content, "42")
}
