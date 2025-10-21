// Package models contains data structures for review service analysis results and code abstractions.
package models

// ScanModeOutput contains results for Scan Mode analysis.
// It includes a summary and a list of code matches.
type ScanModeOutput struct {
	Summary string      `json:"summary"`
	Matches []CodeMatch `json:"matches"`
}

// CodeMatch represents a code match found during Scan Mode analysis.
// It includes the relevance, file path, code snippet, context, and line number.
type CodeMatch struct {
	FilePath    string `json:"file"`
	CodeSnippet string `json:"code_snippet"`
	Context     string `json:"context"`
	Snippet     string
	Relevance   float64 `json:"relevance"`
	Line        int
}

// SkimMode is the string identifier for Skim Mode analysis
// SkimMode is the string identifier for Skim Mode analysis.
const SkimMode = "skim"

// ScanMode is the string identifier for Scan Mode analysis.
// ScanMode is the string identifier for Scan Mode analysis.
const ScanMode = "scan"

// AnalysisResult represents a cached or captured analysis result.
// It includes the analysis mode, prompt, summary, metadata, and other details.
type AnalysisResult struct {
	Mode      string
	Prompt    string
	Summary   string
	Metadata  string
	ModelUsed string
	RawOutput string
	ReviewID  int64
}

// SkimModeOutput represents the output of Skim Mode analysis.
// It includes functions, interfaces, data models, workflows, and a summary.
type SkimModeOutput struct {
	// Reordered fields for optimal memory alignment
	Summary    string
	Functions  []FunctionSignature `json:"functions"`
	Interfaces []InterfaceInfo     `json:"interfaces"`
	DataModels []DataModelInfo     `json:"data_models"`
	Workflows  []WorkflowInfo      `json:"workflows"`
}

// FunctionSignature describes a function's signature and purpose.
// It includes the function name, signature, and a brief description.
type FunctionSignature struct {
	Name        string `json:"name"`
	Signature   string `json:"signature"`
	Description string `json:"description"`
}

// InterfaceInfo describes an interface and its methods.
// It includes the interface name, methods, and purpose.
type InterfaceInfo struct {
	Name        string
	Description string
	Purpose     string   `json:"purpose"`
	Methods     []string `json:"methods"`
}

// DataModelInfo describes a data model and its fields.
// It includes the model name, fields, and purpose.
type DataModelInfo struct {
	Name        string
	Description string
	Purpose     string   `json:"purpose"`
	Fields      []string `json:"fields"`
}

// WorkflowInfo describes a workflow and its steps.
// It includes the workflow name and a list of steps.
type WorkflowInfo struct {
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
}

// CriticalModeOutput contains results for Critical Mode analysis.
// It includes the overall grade, summary, and a list of issues.
type CriticalModeOutput struct {
	OverallGrade string      `json:"overall_grade"`
	Summary      string      `json:"summary"`
	Issues       []CodeIssue `json:"issues"`
}

// CodeIssue represents an issue found during Critical Mode analysis.
// It includes the impact, fix suggestion, description, code snippet, category, severity, file, and line number.
type CodeIssue struct {
	Impact        string `json:"impact"`
	FixSuggestion string `json:"fix_suggestion"`
	Description   string `json:"description"`
	CodeSnippet   string `json:"code_snippet"`
	Category      string `json:"category"` // security, bug, performance, maintainability
	Severity      string `json:"severity"` // critical, high, medium, low
	File          string `json:"file"`
	Line          int    `json:"line"`
}

// CriticalMode is the string identifier for Critical Mode analysis.
// CriticalMode is the string identifier for Critical Mode analysis.
const CriticalMode = "critical"

// Review represents a code review session.
// It includes the session ID, title, code source, and timestamps.
type Review struct {
	Title        string `json:"title"`
	CodeSource   string `json:"code_source"`
	CreatedAt    string `json:"created_at"`
	LastAccessed string `json:"last_accessed"`
	ID           int64  `json:"id"`
}
