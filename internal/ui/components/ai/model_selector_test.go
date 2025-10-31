package ai

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	aicore "github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
)

// TestModelSelector_Render_DisplaysAllModels verifies model list display
func TestModelSelector_Render_DisplaysAllModels(t *testing.T) {
	models := []*aicore.ModelInfo{
		{Provider: "ollama", Model: "local", DisplayName: "Ollama Local", CostPer1kInputTokens: 0.0},
		{Provider: "anthropic", Model: "haiku", DisplayName: "Claude Haiku", CostPer1kInputTokens: 0.00080},
		{Provider: "openai", Model: "gpt-4o", DisplayName: "GPT-4o", CostPer1kInputTokens: 0.005},
	}

	var buf bytes.Buffer
	err := ModelSelector(models, "anthropic:haiku", "review").Render(context.Background(), &buf)

	assert.NoError(t, err)
	html := buf.String()

	// Verify all models present
	assert.Contains(t, html, "Ollama Local")
	assert.Contains(t, html, "Claude Haiku")
	assert.Contains(t, html, "GPT-4o")
}

// TestModelSelector_Render_HighlightsCurrentSelection verifies selected model
func TestModelSelector_Render_HighlightsCurrentSelection(t *testing.T) {
	models := []*aicore.ModelInfo{
		{Provider: "ollama", Model: "local", DisplayName: "Ollama", CostPer1kInputTokens: 0.0},
		{Provider: "anthropic", Model: "haiku", DisplayName: "Claude Haiku", CostPer1kInputTokens: 0.00080},
	}

	var buf bytes.Buffer
	err := ModelSelector(models, "anthropic:haiku", "review").Render(context.Background(), &buf)

	assert.NoError(t, err)
	html := buf.String()

	// Should show which is selected
	assert.Contains(t, html, "selected")
	assert.Contains(t, html, "anthropic:haiku")
}

// TestModelSelector_Render_DisplaysCosts verifies pricing display
func TestModelSelector_Render_DisplaysCosts(t *testing.T) {
	models := []*aicore.ModelInfo{
		{Provider: "ollama", Model: "local", DisplayName: "Ollama (Free)", CostPer1kInputTokens: 0.0},
		{Provider: "anthropic", Model: "haiku", DisplayName: "Claude Haiku", CostPer1kInputTokens: 0.00080},
	}

	var buf bytes.Buffer
	err := ModelSelector(models, "ollama:local", "review").Render(context.Background(), &buf)

	assert.NoError(t, err)
	html := buf.String()

	// Should display cost information
	assert.Contains(t, html, "Free") // Ollama is free
	assert.Contains(t, html, "$0") // or similar cost indication
}

// TestModelSelector_Render_IncludesAppName verifies app context
func TestModelSelector_Render_IncludesAppName(t *testing.T) {
	models := []*aicore.ModelInfo{
		{Provider: "ollama", Model: "local", DisplayName: "Ollama", CostPer1kInputTokens: 0.0},
	}

	var buf bytes.Buffer
	err := ModelSelector(models, "ollama:local", "review").Render(context.Background(), &buf)

	assert.NoError(t, err)
	html := buf.String()

	// Should reference the app
	assert.Contains(t, html, "review")
}

// TestCostAlertBanner_Render_DisplaysWhenThresholdExceeded verifies alert
func TestCostAlertBanner_Render_DisplaysWhenThresholdExceeded(t *testing.T) {
	var buf bytes.Buffer
	err := CostAlertBanner(true, 0.75).Render(context.Background(), &buf)

	assert.NoError(t, err)
	html := buf.String()

	// Should display alert when triggered
	assert.Contains(t, html, "warning") // CSS class or similar
	assert.Contains(t, html, "75.0%")   // Percentage used (with decimal)
}

// TestCostAlertBanner_Render_HiddenWhenUnderThreshold verifies hidden state
func TestCostAlertBanner_Render_HiddenWhenUnderThreshold(t *testing.T) {
	var buf bytes.Buffer
	err := CostAlertBanner(false, 0.30).Render(context.Background(), &buf)

	assert.NoError(t, err)
	html := buf.String()

	// Should not show alert when under threshold
	// Should either be empty or have display:none
	if html != "" {
		assert.Contains(t, html, "display:none") // Or hidden class
	}
}

// TestBudgetWarning_Render_DisplaysWhenExceeded verifies budget warning
func TestBudgetWarning_Render_DisplaysWhenExceeded(t *testing.T) {
	var buf bytes.Buffer
	err := BudgetWarning(true, 10.0, 12.5).Render(context.Background(), &buf)

	assert.NoError(t, err)
	html := buf.String()

	// Should display budget exceeded warning
	assert.Contains(t, html, "12.5") // Current cost
	assert.Contains(t, html, "10.0") // Budget limit
}

// TestBudgetWarning_Render_HiddenWhenWithinBudget verifies hidden when ok
func TestBudgetWarning_Render_HiddenWhenWithinBudget(t *testing.T) {
	var buf bytes.Buffer
	err := BudgetWarning(false, 10.0, 5.0).Render(context.Background(), &buf)

	assert.NoError(t, err)
	html := buf.String()

	// Should not show warning when within budget
	if html != "" {
		assert.Contains(t, html, "display:none") // Or hidden class
	}
}

// TestRememberPreference_Render_IncludesCheckbox verifies checkbox
func TestRememberPreference_Render_IncludesCheckbox(t *testing.T) {
	var buf bytes.Buffer
	err := RememberPreference(true).Render(context.Background(), &buf)

	assert.NoError(t, err)
	html := buf.String()

	// Should have checkbox for remembering preference
	assert.Contains(t, html, "checkbox")
	assert.Contains(t, html, "Remember") // Label text
}
