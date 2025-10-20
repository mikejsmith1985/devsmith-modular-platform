// Package models contains data structures for review service analysis results and code abstractions.
package models

// ScanModeOutput contains results for Scan Mode analysis.
type ScanModeOutput struct {
	Matches []CodeMatch `json:"matches"`
	Summary string      `json:"summary"`
}

// CodeMatch represents a code match found during scan mode analysis.
type CodeMatch struct {
	File        string  `json:"file"`
	Line        int     `json:"line"`
	CodeSnippet string  `json:"code_snippet"`
	Relevance   float64 `json:"relevance"`
	Context     string  `json:"context"`
}

// SkimMode is the string identifier for Skim Mode analysis
// SkimMode is the string identifier for Skim Mode analysis.
const SkimMode = "skim"

// ScanMode is the string identifier for Scan Mode analysis.
const ScanMode = "scan"

// AnalysisResult represents a cached/captured analysis result
// AnalysisResult represents a cached/captured analysis result.
type AnalysisResult struct {
	ReviewID  int64
	Mode      string
	Prompt    string
	RawOutput string
	Summary   string
	Metadata  string
	ModelUsed string
}

// SkimModeOutput represents Skim Mode analysis
// SkimModeOutput represents Skim Mode analysis output.
type SkimModeOutput struct {
	Functions  []FunctionSignature `json:"functions"`
	Interfaces []InterfaceInfo     `json:"interfaces"`
	DataModels []DataModelInfo     `json:"data_models"`
	Workflows  []WorkflowInfo      `json:"workflows"`
	Summary    string              `json:"summary"`
}

// FunctionSignature describes a function's signature and purpose.
type FunctionSignature struct {
	Name        string `json:"name"`
	Signature   string `json:"signature"`
	Description string `json:"description"`
}

// InterfaceInfo describes an interface and its methods.
type InterfaceInfo struct {
	Name    string   `json:"name"`
	Methods []string `json:"methods"`
	Purpose string   `json:"purpose"`
}

// DataModelInfo describes a data model and its fields.
type DataModelInfo struct {
	Name    string   `json:"name"`
	Fields  []string `json:"fields"`
	Purpose string   `json:"purpose"`
}

// WorkflowInfo describes a workflow and its steps.
type WorkflowInfo struct {
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
}

// CriticalModeOutput contains results for Critical Mode analysis.
type CriticalModeOutput struct {
	Issues       []CodeIssue `json:"issues"`
	Summary      string      `json:"summary"`
	OverallGrade string      `json:"overall_grade"`
}

// CodeIssue represents an issue found during Critical Mode analysis.
type CodeIssue struct {
	Severity      string `json:"severity"` // critical, high, medium, low
	Category      string `json:"category"` // security, bug, performance, maintainability
	File          string `json:"file"`
	Line          int    `json:"line"`
	CodeSnippet   string `json:"code_snippet"`
	Description   string `json:"description"`
	Impact        string `json:"impact"`
	FixSuggestion string `json:"fix_suggestion"`
}

// CriticalMode is the string identifier for Critical Mode analysis.
const CriticalMode = "critical"
