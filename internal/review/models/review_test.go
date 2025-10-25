package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanModeOutput_Basic(t *testing.T) {
	output := ScanModeOutput{
		Summary: "Found 5 matches",
		Matches: []CodeMatch{
			{FilePath: "main.go", Line: 10},
		},
	}

	assert.Equal(t, "Found 5 matches", output.Summary)
	assert.Len(t, output.Matches, 1)
}

func TestCodeMatch_Structure(t *testing.T) {
	match := CodeMatch{
		FilePath:    "handler.go",
		CodeSnippet: "func Handler() {}",
		Context:     "HTTP handler",
		Snippet:     "func Handler",
		Relevance:   0.95,
		Line:        42,
	}

	assert.Equal(t, "handler.go", match.FilePath)
	assert.Equal(t, 42, match.Line)
	assert.Equal(t, 0.95, match.Relevance)
}

func TestAnalysisResult_Fields(t *testing.T) {
	result := AnalysisResult{
		Mode:    "skim",
		Prompt:  "Analyze this code",
		Summary: "Code summary",
	}

	assert.Equal(t, "skim", result.Mode)
}

func TestSkimModeOutput_Structure(t *testing.T) {
	output := SkimModeOutput{
		Summary: "Skim analysis",
		Functions: []FunctionSignature{
			{Name: "Main", Signature: "func Main()", Description: "Entry point"},
		},
	}

	assert.Equal(t, "Skim analysis", output.Summary)
	assert.Len(t, output.Functions, 1)
}

func TestFunctionSignature_Fields(t *testing.T) {
	sig := FunctionSignature{
		Name:      "Calculate",
		Signature: "func Calculate(a, b int) int",
	}

	assert.Equal(t, "Calculate", sig.Name)
	assert.Contains(t, sig.Signature, "int")
}

func TestInterfaceInfo_Structure(t *testing.T) {
	iface := InterfaceInfo{
		Name:    "Reader",
		Purpose: "Define reading behavior",
		Methods: []string{"Read", "Close"},
	}

	assert.Equal(t, "Reader", iface.Name)
	assert.Len(t, iface.Methods, 2)
}

func TestDataModelInfo_Structure(t *testing.T) {
	model := DataModelInfo{
		Name:    "User",
		Purpose: "Store user data",
		Fields:  []string{"ID", "Name", "Email"},
	}

	assert.Equal(t, "User", model.Name)
	assert.Len(t, model.Fields, 3)
}

func TestWorkflowInfo_Structure(t *testing.T) {
	workflow := WorkflowInfo{
		Name:  "UserCreation",
		Steps: []string{"Validate", "Save", "SendEmail"},
	}

	assert.Equal(t, "UserCreation", workflow.Name)
	assert.Len(t, workflow.Steps, 3)
}

func TestCriticalModeOutput_Structure(t *testing.T) {
	output := CriticalModeOutput{
		OverallGrade: "B+",
		Summary:      "Good code quality",
	}

	assert.Equal(t, "B+", output.OverallGrade)
}

func TestCodeIssue_Structure(t *testing.T) {
	issue := CodeIssue{
		Impact:        "High",
		FixSuggestion: "Use mutex",
		Description:   "Race condition",
		CodeSnippet:   "globalVar++",
		Category:      "security",
		Severity:      "critical",
		File:          "main.go",
		Line:          55,
	}

	assert.Equal(t, "critical", issue.Severity)
	assert.Equal(t, 55, issue.Line)
}

func TestReview_Structure(t *testing.T) {
	review := Review{
		Title:      "My Review",
		CodeSource: "github",
		ID:         1,
	}

	assert.Equal(t, "My Review", review.Title)
	assert.Equal(t, int64(1), review.ID)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "skim", SkimMode)
	assert.Equal(t, "scan", ScanMode)
	assert.Equal(t, "critical", CriticalMode)
}

func TestZeroValueStructs(t *testing.T) {
	var output ScanModeOutput
	assert.Equal(t, "", output.Summary)
	assert.Len(t, output.Matches, 0)

	var result AnalysisResult
	assert.Equal(t, "", result.Mode)

	var issue CodeIssue
	assert.Equal(t, 0, issue.Line)
}

func TestMultipleMatches(t *testing.T) {
	output := ScanModeOutput{
		Matches: []CodeMatch{
			{FilePath: "a.go", Line: 1},
			{FilePath: "b.go", Line: 2},
			{FilePath: "c.go", Line: 3},
		},
	}

	assert.Len(t, output.Matches, 3)
	assert.Equal(t, "a.go", output.Matches[0].FilePath)
	assert.Equal(t, "c.go", output.Matches[2].FilePath)
}

func TestMultipleIssues(t *testing.T) {
	output := CriticalModeOutput{
		Issues: []CodeIssue{
			{Severity: "critical"},
			{Severity: "high"},
			{Severity: "medium"},
		},
	}

	assert.Len(t, output.Issues, 3)
}
