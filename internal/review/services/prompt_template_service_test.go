package review_services

import (
	"context"
	"errors"
	"testing"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPromptTemplateRepository is a mock implementation for testing
type MockPromptTemplateRepository struct {
	mock.Mock
}

// Ensure mock implements the interface
var _ repositories.PromptTemplateRepositoryInterface = (*MockPromptTemplateRepository)(nil)

func (m *MockPromptTemplateRepository) FindByUserAndMode(ctx context.Context, userID int, mode, userLevel, outputMode string) (*review_models.PromptTemplate, error) {
	args := m.Called(ctx, userID, mode, userLevel, outputMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*review_models.PromptTemplate), args.Error(1)
}

func (m *MockPromptTemplateRepository) FindDefaultByMode(ctx context.Context, mode, userLevel, outputMode string) (*review_models.PromptTemplate, error) {
	args := m.Called(ctx, mode, userLevel, outputMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*review_models.PromptTemplate), args.Error(1)
}

func (m *MockPromptTemplateRepository) Upsert(ctx context.Context, template *review_models.PromptTemplate) (*review_models.PromptTemplate, error) {
	args := m.Called(ctx, template)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*review_models.PromptTemplate), args.Error(1)
}

func (m *MockPromptTemplateRepository) DeleteUserCustom(ctx context.Context, userID int, mode, userLevel, outputMode string) error {
	args := m.Called(ctx, userID, mode, userLevel, outputMode)
	return args.Error(0)
}

func (m *MockPromptTemplateRepository) SaveExecution(ctx context.Context, execution *review_models.PromptExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockPromptTemplateRepository) GetExecutionHistory(ctx context.Context, userID int, limit int) ([]*review_models.PromptExecution, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*review_models.PromptExecution), args.Error(1)
}

func (m *MockPromptTemplateRepository) UpdateExecutionRating(ctx context.Context, executionID int64, userID int, rating int) error {
	args := m.Called(ctx, executionID, userID, rating)
	return args.Error(0)
}

// Test: GetEffectivePrompt returns user custom over system default
func TestPromptTemplateService_GetEffectivePrompt_UserCustom(t *testing.T) {
	repo := new(MockPromptTemplateRepository)
	service := NewPromptTemplateService(repo)

	userID := 1
	mode := "preview"
	userLevel := "beginner"
	outputMode := "quick"

	customPrompt := &review_models.PromptTemplate{
		ID:         "custom-1",
		UserID:     &userID,
		Mode:       mode,
		UserLevel:  userLevel,
		OutputMode: outputMode,
		PromptText: "Custom preview prompt for {{code}}",
		Variables:  []string{"{{code}}"},
		IsDefault:  false,
	}

	repo.On("FindByUserAndMode", mock.Anything, userID, mode, userLevel, outputMode).
		Return(customPrompt, nil)

	result, err := service.GetEffectivePrompt(context.Background(), userID, mode, userLevel, outputMode)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "custom-1", result.ID)
	assert.Equal(t, "Custom preview prompt for {{code}}", result.PromptText)
	assert.True(t, result.IsCustom())
	repo.AssertExpectations(t)
}

// Test: GetEffectivePrompt falls back to system default if no custom
func TestPromptTemplateService_GetEffectivePrompt_FallbackToDefault(t *testing.T) {
	repo := new(MockPromptTemplateRepository)
	service := NewPromptTemplateService(repo)

	userID := 1
	mode := "preview"
	userLevel := "beginner"
	outputMode := "quick"

	defaultPrompt := &review_models.PromptTemplate{
		ID:         "default-preview-beginner-quick",
		UserID:     nil, // System default
		Mode:       mode,
		UserLevel:  userLevel,
		OutputMode: outputMode,
		PromptText: "Default preview prompt for {{code}}",
		Variables:  []string{"{{code}}"},
		IsDefault:  true,
	}

	// User has no custom prompt
	repo.On("FindByUserAndMode", mock.Anything, userID, mode, userLevel, outputMode).
		Return(nil, nil)

	// Fall back to system default
	repo.On("FindDefaultByMode", mock.Anything, mode, userLevel, outputMode).
		Return(defaultPrompt, nil)

	result, err := service.GetEffectivePrompt(context.Background(), userID, mode, userLevel, outputMode)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "default-preview-beginner-quick", result.ID)
	assert.False(t, result.IsCustom())
	assert.True(t, result.IsDefault)
	repo.AssertExpectations(t)
}

// Test: GetEffectivePrompt returns error if no default exists
func TestPromptTemplateService_GetEffectivePrompt_NoDefaultError(t *testing.T) {
	repo := new(MockPromptTemplateRepository)
	service := NewPromptTemplateService(repo)

	userID := 1
	mode := "preview"
	userLevel := "beginner"
	outputMode := "quick"

	repo.On("FindByUserAndMode", mock.Anything, userID, mode, userLevel, outputMode).
		Return(nil, nil)
	repo.On("FindDefaultByMode", mock.Anything, mode, userLevel, outputMode).
		Return(nil, nil)

	result, err := service.GetEffectivePrompt(context.Background(), userID, mode, userLevel, outputMode)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no default prompt found")
	repo.AssertExpectations(t)
}

// Test: SaveCustomPrompt validates required variables present
func TestPromptTemplateService_SaveCustomPrompt_ValidatesVariables(t *testing.T) {
	repo := new(MockPromptTemplateRepository)
	service := NewPromptTemplateService(repo)

	userID := 1
	mode := "preview"
	userLevel := "beginner"
	outputMode := "quick"
	invalidPrompt := "This prompt is missing the code variable" // No {{code}}

	result, err := service.SaveCustomPrompt(context.Background(), userID, mode, userLevel, outputMode, invalidPrompt)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "missing required variable")
	assert.Contains(t, err.Error(), "{{code}}")
	// Should never call Upsert since validation fails
	repo.AssertNotCalled(t, "Upsert")
}

// Test: SaveCustomPrompt creates valid custom prompt
func TestPromptTemplateService_SaveCustomPrompt_Success(t *testing.T) {
	repo := new(MockPromptTemplateRepository)
	service := NewPromptTemplateService(repo)

	userID := 1
	mode := "preview"
	userLevel := "beginner"
	outputMode := "quick"
	promptText := "Custom prompt with {{code}} variable"

	expectedTemplate := &review_models.PromptTemplate{
		ID:         "custom-1-preview-beginner-quick",
		UserID:     &userID,
		Mode:       mode,
		UserLevel:  userLevel,
		OutputMode: outputMode,
		PromptText: promptText,
		Variables:  []string{"{{code}}"},
		IsDefault:  false,
		Version:    1,
	}

	repo.On("Upsert", mock.Anything, mock.MatchedBy(func(t *review_models.PromptTemplate) bool {
		return t.UserID != nil && *t.UserID == userID &&
			t.Mode == mode &&
			t.UserLevel == userLevel &&
			t.OutputMode == outputMode &&
			t.PromptText == promptText
	})).Return(expectedTemplate, nil)

	result, err := service.SaveCustomPrompt(context.Background(), userID, mode, userLevel, outputMode, promptText)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, promptText, result.PromptText)
	assert.True(t, result.IsCustom())
	repo.AssertExpectations(t)
}

// Test: SaveCustomPrompt validates scan mode has query variable
func TestPromptTemplateService_SaveCustomPrompt_ScanModeRequiresQuery(t *testing.T) {
	repo := new(MockPromptTemplateRepository)
	service := NewPromptTemplateService(repo)

	userID := 1
	mode := "scan"
	userLevel := "beginner"
	outputMode := "quick"
	invalidPrompt := "Scan prompt with {{code}} but missing the query variable" // Has {{code}} but no {{query}}

	result, err := service.SaveCustomPrompt(context.Background(), userID, mode, userLevel, outputMode, invalidPrompt)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "missing required variable")
	assert.Contains(t, err.Error(), "{{query}}")
	// Should never call Upsert since validation fails
	repo.AssertNotCalled(t, "Upsert")
}

// Test: FactoryReset deletes user custom and returns to default
func TestPromptTemplateService_FactoryReset_Success(t *testing.T) {
	repo := new(MockPromptTemplateRepository)
	service := NewPromptTemplateService(repo)

	userID := 1
	mode := "preview"
	userLevel := "beginner"
	outputMode := "quick"

	repo.On("DeleteUserCustom", mock.Anything, userID, mode, userLevel, outputMode).
		Return(nil)

	err := service.FactoryReset(context.Background(), userID, mode, userLevel, outputMode)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// Test: FactoryReset returns error if delete fails
func TestPromptTemplateService_FactoryReset_DeleteError(t *testing.T) {
	repo := new(MockPromptTemplateRepository)
	service := NewPromptTemplateService(repo)

	userID := 1
	mode := "preview"
	userLevel := "beginner"
	outputMode := "quick"

	repo.On("DeleteUserCustom", mock.Anything, userID, mode, userLevel, outputMode).
		Return(errors.New("database error"))

	err := service.FactoryReset(context.Background(), userID, mode, userLevel, outputMode)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	repo.AssertExpectations(t)
}

// Test: RenderPrompt substitutes all variables correctly
func TestPromptTemplateService_RenderPrompt_Success(t *testing.T) {
	service := NewPromptTemplateService(nil)

	template := &review_models.PromptTemplate{
		PromptText: "Analyze this code: {{code}} with query: {{query}}",
		Variables:  []string{"{{code}}", "{{query}}"},
	}

	variables := map[string]string{
		"{{code}}":  "func main() {}",
		"{{query}}": "find bugs",
	}

	rendered, err := service.RenderPrompt(template, variables)

	assert.NoError(t, err)
	assert.Equal(t, "Analyze this code: func main() {} with query: find bugs", rendered)
	assert.NotContains(t, rendered, "{{")
}

// Test: RenderPrompt errors if variable missing
func TestPromptTemplateService_RenderPrompt_MissingVariable(t *testing.T) {
	service := NewPromptTemplateService(nil)

	template := &review_models.PromptTemplate{
		PromptText: "Analyze this code: {{code}} with query: {{query}}",
		Variables:  []string{"{{code}}", "{{query}}"},
	}

	variables := map[string]string{
		"{{code}}": "func main() {}",
		// Missing {{query}}
	}

	rendered, err := service.RenderPrompt(template, variables)

	assert.Error(t, err)
	assert.Empty(t, rendered)
	assert.Contains(t, err.Error(), "missing variable value")
	assert.Contains(t, err.Error(), "{{query}}")
}

// Test: RenderPrompt handles empty variables map
func TestPromptTemplateService_RenderPrompt_EmptyVariables(t *testing.T) {
	service := NewPromptTemplateService(nil)

	template := &review_models.PromptTemplate{
		PromptText: "Simple prompt with no variables",
		Variables:  []string{},
	}

	variables := map[string]string{}

	rendered, err := service.RenderPrompt(template, variables)

	assert.NoError(t, err)
	assert.Equal(t, "Simple prompt with no variables", rendered)
}

// Test: LogExecution records prompt usage
func TestPromptTemplateService_LogExecution_Success(t *testing.T) {
	repo := new(MockPromptTemplateRepository)
	service := NewPromptTemplateService(repo)

	execution := &review_models.PromptExecution{
		TemplateID:     "custom-1",
		UserID:         1,
		RenderedPrompt: "Rendered prompt text",
		Response:       "AI response",
		ModelUsed:      "claude-3-5-sonnet-20241022",
		LatencyMs:      1500,
		TokensUsed:     500,
	}

	repo.On("SaveExecution", mock.Anything, execution).Return(nil)

	err := service.LogExecution(context.Background(), execution)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// Test: LogExecution validates required fields
func TestPromptTemplateService_LogExecution_MissingFields(t *testing.T) {
	repo := new(MockPromptTemplateRepository)
	service := NewPromptTemplateService(repo)

	// Missing TemplateID
	execution := &review_models.PromptExecution{
		UserID:    1,
		ModelUsed: "claude-3-5-sonnet-20241022",
	}

	err := service.LogExecution(context.Background(), execution)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template_id is required")
}

// Test: ExtractVariables finds all template variables
func TestPromptTemplateService_ExtractVariables(t *testing.T) {
	service := NewPromptTemplateService(nil)

	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "Single variable",
			text:     "Analyze {{code}}",
			expected: []string{"{{code}}"},
		},
		{
			name:     "Multiple variables",
			text:     "Analyze {{code}} with {{query}} for {{file}}",
			expected: []string{"{{code}}", "{{query}}", "{{file}}"},
		},
		{
			name:     "Duplicate variables",
			text:     "Check {{code}} and verify {{code}} again",
			expected: []string{"{{code}}"},
		},
		{
			name:     "No variables",
			text:     "Static prompt text",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ExtractVariables(tt.text)
			// Use ElementsMatch to compare without caring about order
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}
