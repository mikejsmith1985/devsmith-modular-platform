package review_services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// MultiFileAnalyzer provides cross-file code analysis using AI
type MultiFileAnalyzer struct {
	aiProvider ai.Provider
	model      string
}

// NewMultiFileAnalyzer creates a new multi-file analyzer
func NewMultiFileAnalyzer(aiProvider ai.Provider, model string) *MultiFileAnalyzer {
	return &MultiFileAnalyzer{
		aiProvider: aiProvider,
		model:      model,
	}
}

// FileContent represents a single file to be analyzed
type FileContent struct {
	Path    string
	Content string
	Size    int64
}

// AnalyzeRequest contains parameters for multi-file analysis
type AnalyzeRequest struct {
	Files       []FileContent
	ReadingMode string
	Temperature float64
}

// AnalyzeResult contains the analysis results
type AnalyzeResult struct {
	Summary              string
	Dependencies         []review_models.CrossFileDependency
	SharedAbstractions   []review_models.SharedAbstraction
	ArchitecturePatterns []review_models.ArchitecturePattern
	Recommendations      []string
	DurationMs           int64
	InputTokens          int
	OutputTokens         int
}

// Analyze performs cross-file analysis using AI
func (m *MultiFileAnalyzer) Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResult, error) {
	startTime := time.Now()

	// Build combined prompt with file separators
	combinedContent := m.buildCombinedPrompt(req.Files, req.ReadingMode)

	// Call AI provider
	aiReq := &ai.Request{
		Prompt:      combinedContent,
		Model:       m.model,
		Temperature: req.Temperature,
		MaxTokens:   4000, // Allow for detailed cross-file analysis
	}

	aiResp, err := m.aiProvider.Generate(ctx, aiReq)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// Parse AI response (expect JSON structure)
	result, err := m.parseAIResponse(aiResp.Content)
	if err != nil {
		// If parsing fails, return basic result
		result = &AnalyzeResult{
			Summary:         aiResp.Content,
			Recommendations: []string{"Analysis completed, but response format was unexpected"},
		}
	}

	// Add timing and token metrics
	result.DurationMs = time.Since(startTime).Milliseconds()
	result.InputTokens = aiResp.InputTokens
	result.OutputTokens = aiResp.OutputTokens

	return result, nil
}

// buildCombinedPrompt creates a prompt optimized for multi-file analysis
func (m *MultiFileAnalyzer) buildCombinedPrompt(files []FileContent, readingMode string) string {
	var sb strings.Builder

	// Write analysis instructions
	sb.WriteString("You are analyzing multiple source code files together. ")
	sb.WriteString("Focus on cross-file dependencies, shared abstractions, and architectural patterns.\n\n")

	switch readingMode {
	case "preview":
		sb.WriteString("**Reading Mode: Preview** - Provide high-level overview of file structure and relationships.\n")
	case "skim":
		sb.WriteString("**Reading Mode: Skim** - Focus on interfaces, abstractions, and major workflows across files.\n")
	case "scan":
		sb.WriteString("**Reading Mode: Scan** - Identify specific patterns, imports, and connections between files.\n")
	case "detailed":
		sb.WriteString("**Reading Mode: Detailed** - Perform deep analysis of cross-file dependencies and data flow.\n")
	case "critical":
		sb.WriteString("**Reading Mode: Critical** - Evaluate architectural quality, identify issues, suggest improvements.\n")
	default:
		sb.WriteString("**Reading Mode: General** - Provide balanced analysis of file relationships.\n")
	}

	sb.WriteString("\n**Files to analyze:**\n\n")

	// Add each file with clear separators
	for i, file := range files {
		sb.WriteString(fmt.Sprintf("=== FILE %d/%d: %s (%d bytes) ===\n", i+1, len(files), file.Path, file.Size))
		sb.WriteString(file.Content)
		sb.WriteString("\n\n")
	}

	// Request structured JSON response
	sb.WriteString("\n**Instructions:**\n")
	sb.WriteString("Provide your analysis as JSON with the following structure:\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"summary\": \"Brief overview of the file relationships\",\n")
	sb.WriteString("  \"dependencies\": [{\"from_file\": \"path/a.go\", \"to_file\": \"path/b.go\", \"import_type\": \"import\", \"symbols\": [\"TypeName\"]}],\n")
	sb.WriteString("  \"shared_abstractions\": [{\"name\": \"Interface\", \"type\": \"interface\", \"files\": [\"a.go\", \"b.go\"], \"description\": \"Purpose\"}],\n")
	sb.WriteString("  \"architecture_patterns\": [{\"pattern\": \"MVC\", \"confidence\": 0.85, \"files\": [\"all\"], \"description\": \"Pattern details\"}],\n")
	sb.WriteString("  \"recommendations\": [\"Suggestion 1\", \"Suggestion 2\"]\n")
	sb.WriteString("}\n")

	return sb.String()
}

// parseAIResponse attempts to parse the AI response as JSON
func (m *MultiFileAnalyzer) parseAIResponse(content string) (*AnalyzeResult, error) {
	// Try to find JSON in the response (AI might wrap it in markdown)
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonContent := content[jsonStart : jsonEnd+1]

	// Parse JSON structure
	var parsed struct {
		Summary              string                                   `json:"summary"`
		Dependencies         []review_models.CrossFileDependency      `json:"dependencies"`
		SharedAbstractions   []review_models.SharedAbstraction        `json:"shared_abstractions"`
		ArchitecturePatterns []review_models.ArchitecturePattern      `json:"architecture_patterns"`
		Recommendations      []string                                 `json:"recommendations"`
	}

	if err := json.Unmarshal([]byte(jsonContent), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	result := &AnalyzeResult{
		Summary:              parsed.Summary,
		Dependencies:         parsed.Dependencies,
		SharedAbstractions:   parsed.SharedAbstractions,
		ArchitecturePatterns: parsed.ArchitecturePatterns,
		Recommendations:      parsed.Recommendations,
	}

	// Ensure slices are not nil
	if result.Dependencies == nil {
		result.Dependencies = []review_models.CrossFileDependency{}
	}
	if result.SharedAbstractions == nil {
		result.SharedAbstractions = []review_models.SharedAbstraction{}
	}
	if result.ArchitecturePatterns == nil {
		result.ArchitecturePatterns = []review_models.ArchitecturePattern{}
	}
	if result.Recommendations == nil {
		result.Recommendations = []string{}
	}

	return result, nil
}
