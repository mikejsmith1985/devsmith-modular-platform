package review_services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
)

// mockAIProvider implements ai.Provider for testing
type mockAIProvider struct {
	responseContent string
	err             error
	inputTokens     int
	outputTokens    int
}

func (m *mockAIProvider) Generate(ctx context.Context, req *ai.Request) (*ai.Response, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &ai.Response{
		Content:      m.responseContent,
		Model:        req.Model,
		FinishReason: "complete",
		InputTokens:  m.inputTokens,
		OutputTokens: m.outputTokens,
		ResponseTime: 100 * time.Millisecond,
		CostUSD:      0,
	}, nil
}

func (m *mockAIProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *mockAIProvider) GetModelInfo() *ai.ModelInfo {
	return &ai.ModelInfo{
		Provider:    "test",
		Model:       "test-model",
		DisplayName: "Test Model",
	}
}

func TestMultiFileAnalyzer_Analyze_Success(t *testing.T) {
	// Setup
	mockResponse := `{
		"summary": "Three Go files with MVC pattern",
		"dependencies": [
			{
				"from_file": "handler.go",
				"to_file": "service.go",
				"import_type": "import",
				"symbols": ["UserService"]
			}
		],
		"shared_abstractions": [
			{
				"name": "Repository",
				"type": "interface",
				"files": ["service.go", "db.go"],
				"description": "Data access layer"
			}
		],
		"architecture_patterns": [
			{
				"pattern": "MVC",
				"confidence": 0.9,
				"files": ["handler.go", "service.go", "db.go"],
				"description": "Classic MVC pattern"
			}
		],
		"recommendations": ["Add more tests", "Extract constants"]
	}`

	mockProvider := &mockAIProvider{
		responseContent: mockResponse,
		inputTokens:     500,
		outputTokens:    200,
	}

	analyzer := NewMultiFileAnalyzer(mockProvider, "test-model")

	// Execute
	req := &AnalyzeRequest{
		Files: []FileContent{
			{Path: "handler.go", Content: "package main\n// handler", Size: 100},
			{Path: "service.go", Content: "package main\n// service", Size: 150},
			{Path: "db.go", Content: "package main\n// db", Size: 200},
		},
		ReadingMode: "critical",
		Temperature: 0.7,
	}

	result, err := analyzer.Analyze(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Three Go files with MVC pattern", result.Summary)
	assert.Len(t, result.Dependencies, 1)
	assert.Equal(t, "handler.go", result.Dependencies[0].FromFile)
	assert.Equal(t, "service.go", result.Dependencies[0].ToFile)
	assert.Len(t, result.SharedAbstractions, 1)
	assert.Equal(t, "Repository", result.SharedAbstractions[0].Name)
	assert.Len(t, result.ArchitecturePatterns, 1)
	assert.Equal(t, "MVC", result.ArchitecturePatterns[0].Pattern)
	assert.Len(t, result.Recommendations, 2)
	assert.GreaterOrEqual(t, result.DurationMs, int64(0)) // Duration may be 0 in tests
	assert.Equal(t, 500, result.InputTokens)
	assert.Equal(t, 200, result.OutputTokens)
}

func TestMultiFileAnalyzer_Analyze_NonJSONResponse(t *testing.T) {
	// Setup - AI returns plain text instead of JSON
	mockProvider := &mockAIProvider{
		responseContent: "This is a plain text analysis without JSON structure.",
		inputTokens:     300,
		outputTokens:    50,
	}

	analyzer := NewMultiFileAnalyzer(mockProvider, "test-model")

	// Execute
	req := &AnalyzeRequest{
		Files: []FileContent{
			{Path: "test.go", Content: "package main", Size: 50},
		},
		ReadingMode: "preview",
		Temperature: 0.5,
	}

	result, err := analyzer.Analyze(context.Background(), req)

	// Assert - should succeed but with fallback result
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Summary, "plain text analysis")
	assert.Len(t, result.Recommendations, 1) // Fallback recommendation
	assert.Len(t, result.Dependencies, 0)     // Empty slices
	assert.Len(t, result.SharedAbstractions, 0)
	assert.Len(t, result.ArchitecturePatterns, 0)
}

func TestMultiFileAnalyzer_Analyze_AIError(t *testing.T) {
	// Setup - AI provider returns error
	mockProvider := &mockAIProvider{
		err: assert.AnError,
	}

	analyzer := NewMultiFileAnalyzer(mockProvider, "test-model")

	// Execute
	req := &AnalyzeRequest{
		Files: []FileContent{
			{Path: "test.go", Content: "package main", Size: 50},
		},
		ReadingMode: "preview",
		Temperature: 0.5,
	}

	result, err := analyzer.Analyze(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "AI generation failed")
}

func TestMultiFileAnalyzer_BuildCombinedPrompt_PreviewMode(t *testing.T) {
	analyzer := NewMultiFileAnalyzer(nil, "test-model")

	files := []FileContent{
		{Path: "main.go", Content: "package main\nfunc main() {}", Size: 30},
		{Path: "utils.go", Content: "package main\nfunc helper() {}", Size: 32},
	}

	prompt := analyzer.buildCombinedPrompt(files, "preview")

	// Assert
	assert.Contains(t, prompt, "Preview")
	assert.Contains(t, prompt, "high-level overview")
	assert.Contains(t, prompt, "FILE 1/2: main.go")
	assert.Contains(t, prompt, "FILE 2/2: utils.go")
	assert.Contains(t, prompt, "package main")
	assert.Contains(t, prompt, "JSON")
}

func TestMultiFileAnalyzer_BuildCombinedPrompt_CriticalMode(t *testing.T) {
	analyzer := NewMultiFileAnalyzer(nil, "test-model")

	files := []FileContent{
		{Path: "handler.go", Content: "package main", Size: 15},
	}

	prompt := analyzer.buildCombinedPrompt(files, "critical")

	// Assert
	assert.Contains(t, prompt, "Critical")
	assert.Contains(t, prompt, "architectural quality")
	assert.Contains(t, prompt, "identify issues")
	assert.Contains(t, prompt, "suggest improvements")
}

func TestMultiFileAnalyzer_ParseAIResponse_ValidJSON(t *testing.T) {
	analyzer := NewMultiFileAnalyzer(nil, "test-model")

	// Valid JSON response
	response := `{
		"summary": "Test summary",
		"dependencies": [],
		"shared_abstractions": [],
		"architecture_patterns": [],
		"recommendations": ["Test rec"]
	}`

	result, err := analyzer.parseAIResponse(response)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Test summary", result.Summary)
	assert.NotNil(t, result.Dependencies)
	assert.NotNil(t, result.SharedAbstractions)
	assert.NotNil(t, result.ArchitecturePatterns)
	assert.Len(t, result.Recommendations, 1)
}

func TestMultiFileAnalyzer_ParseAIResponse_JSONWithMarkdown(t *testing.T) {
	analyzer := NewMultiFileAnalyzer(nil, "test-model")

	// JSON wrapped in markdown
	response := "```json\n" + `{
		"summary": "Wrapped in markdown",
		"dependencies": [],
		"shared_abstractions": [],
		"architecture_patterns": [],
		"recommendations": []
	}` + "\n```"

	result, err := analyzer.parseAIResponse(response)

	// Assert - should still parse successfully
	require.NoError(t, err)
	assert.Equal(t, "Wrapped in markdown", result.Summary)
}

func TestMultiFileAnalyzer_ParseAIResponse_NoJSON(t *testing.T) {
	analyzer := NewMultiFileAnalyzer(nil, "test-model")

	// No JSON in response
	response := "This is just plain text with no JSON structure."

	result, err := analyzer.parseAIResponse(response)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no JSON found")
}

func TestMultiFileAnalyzer_ParseAIResponse_InvalidJSON(t *testing.T) {
	analyzer := NewMultiFileAnalyzer(nil, "test-model")

	// Malformed JSON - has braces but invalid structure
	response := `{
		"summary": "Test",
		"dependencies": [
			{"invalid": "no closing bracket"
		]
	}`

	result, err := analyzer.parseAIResponse(response)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse JSON")
}

func TestMultiFileAnalyzer_ParseAIResponse_NilSlices(t *testing.T) {
	analyzer := NewMultiFileAnalyzer(nil, "test-model")

	// JSON with null arrays (should be converted to empty slices)
	response := `{
		"summary": "Test",
		"dependencies": null,
		"shared_abstractions": null,
		"architecture_patterns": null,
		"recommendations": null
	}`

	result, err := analyzer.parseAIResponse(response)

	// Assert - should convert nulls to empty slices
	require.NoError(t, err)
	assert.NotNil(t, result.Dependencies)
	assert.Len(t, result.Dependencies, 0)
	assert.NotNil(t, result.SharedAbstractions)
	assert.Len(t, result.SharedAbstractions, 0)
	assert.NotNil(t, result.ArchitecturePatterns)
	assert.Len(t, result.ArchitecturePatterns, 0)
	assert.NotNil(t, result.Recommendations)
	assert.Len(t, result.Recommendations, 0)
}

func TestMultiFileAnalyzer_RealWorldStructure(t *testing.T) {
	// Test with realistic multi-file structure
	mockResponse := `{
		"summary": "Go web application following layered architecture",
		"dependencies": [
			{
				"from_file": "handlers/user_handler.go",
				"to_file": "services/user_service.go",
				"import_type": "import",
				"symbols": ["UserService", "CreateUserRequest"]
			},
			{
				"from_file": "services/user_service.go",
				"to_file": "db/user_repository.go",
				"import_type": "import",
				"symbols": ["UserRepository"]
			}
		],
		"shared_abstractions": [
			{
				"name": "Repository",
				"type": "interface",
				"files": ["db/user_repository.go", "db/session_repository.go"],
				"description": "Common data access pattern",
				"complexity": "simple"
			},
			{
				"name": "Handler",
				"type": "pattern",
				"files": ["handlers/user_handler.go", "handlers/session_handler.go"],
				"description": "HTTP request handlers",
				"complexity": "moderate"
			}
		],
		"architecture_patterns": [
			{
				"pattern": "Layered Architecture",
				"confidence": 0.95,
				"files": ["handlers/", "services/", "db/"],
				"description": "Clear separation: Handlers → Services → Data"
			},
			{
				"pattern": "Repository Pattern",
				"confidence": 0.88,
				"files": ["db/user_repository.go", "db/session_repository.go"],
				"description": "Abstracted data access layer"
			}
		],
		"recommendations": [
			"Consider adding middleware for authentication",
			"Extract validation logic to separate package",
			"Add comprehensive error handling in handlers",
			"Implement request/response logging"
		]
	}`

	mockProvider := &mockAIProvider{
		responseContent: mockResponse,
		inputTokens:     1200,
		outputTokens:    500,
	}

	analyzer := NewMultiFileAnalyzer(mockProvider, "test-model")

	req := &AnalyzeRequest{
		Files: []FileContent{
			{Path: "handlers/user_handler.go", Content: "...", Size: 1500},
			{Path: "services/user_service.go", Content: "...", Size: 2000},
			{Path: "db/user_repository.go", Content: "...", Size: 1200},
		},
		ReadingMode: "detailed",
		Temperature: 0.6,
	}

	result, err := analyzer.Analyze(context.Background(), req)

	// Assert comprehensive result
	require.NoError(t, err)
	assert.Contains(t, result.Summary, "layered architecture")
	assert.Len(t, result.Dependencies, 2)
	assert.Len(t, result.SharedAbstractions, 2)
	assert.Len(t, result.ArchitecturePatterns, 2)
	assert.Len(t, result.Recommendations, 4)

	// Verify specific structures
	assert.Equal(t, "handlers/user_handler.go", result.Dependencies[0].FromFile)
	assert.Equal(t, "Repository", result.SharedAbstractions[0].Name)
	assert.Equal(t, "Layered Architecture", result.ArchitecturePatterns[0].Pattern)
	assert.Greater(t, result.ArchitecturePatterns[0].Confidence, 0.9)
}
