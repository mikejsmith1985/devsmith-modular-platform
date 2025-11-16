// Package review_models contains data structures for review service analysis results and code abstractions.
package review_models

// ====================================================================================
// MODE OUTPUT STRUCTURES - Five Reading Modes
// ====================================================================================

// PreviewModeOutput contains results for Preview Mode analysis.
// Preview mode provides rapid structural assessment of code.
type PreviewModeOutput struct {
	Summary           string     `json:"summary"`
	FileTree          []FileNode `json:"file_tree"`
	BoundedContexts   []string   `json:"bounded_contexts"`
	TechStack         []string   `json:"tech_stack"`
	ArchitectureStyle string     `json:"architecture_style"`
	EntryPoints       []string   `json:"entry_points"`
	ExternalDeps      []string   `json:"external_dependencies"`
	Stats             CodeStats  `json:"stats"`
}

// FileNode represents a file or directory in the code structure
type FileNode struct {
	Name        string     `json:"name"`
	Type        string     `json:"type"` // "file" or "directory"
	Path        string     `json:"path"`
	Description string     `json:"description"`
	Children    []FileNode `json:"children,omitempty"`
}

// CodeStats provides statistical information about the code
type CodeStats struct {
	TotalFiles      int `json:"total_files"`
	TotalLines      int `json:"total_lines"`
	TotalFunctions  int `json:"total_functions"`
	TotalInterfaces int `json:"total_interfaces"`
	TotalTests      int `json:"total_tests"`
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

// ScanModeOutput contains results for Scan Mode analysis.
// It includes a summary and a list of code matches.
type ScanModeOutput struct {
	Summary string      `json:"summary"`
	Matches []CodeMatch `json:"matches"`
}

// DetailedModeOutput contains results for Detailed Mode analysis.
// Provides line-by-line explanation and algorithm analysis.
type DetailedModeOutput struct {
	Summary          string            `json:"summary"`
	LineExplanations []LineExplanation `json:"line_explanations"`
	AlgorithmSummary string            `json:"algorithm_summary"`
	Complexity       string            `json:"complexity"`
	EdgeCases        []string          `json:"edge_cases"`
	VariableTracking []VariableState   `json:"variable_tracking"`
	ControlFlow      []ControlFlowNode `json:"control_flow"`
}

// LineExplanation provides explanation for a specific line of code
type LineExplanation struct {
	LineNumber  int    `json:"line_number"`
	Code        string `json:"code"`
	Explanation string `json:"explanation"`
	Variables   string `json:"variables"` // Variable states at this line
}

// VariableState tracks variable values at specific points
type VariableState struct {
	LineNumber int               `json:"line_number"`
	Variables  map[string]string `json:"variables"`
}

// ControlFlowNode represents a node in the control flow graph
type ControlFlowNode struct {
	Type        string   `json:"type"` // "if", "loop", "function_call", etc.
	LineNumber  int      `json:"line_number"`
	Description string   `json:"description"`
	Children    []string `json:"children,omitempty"`
}

// CriticalModeOutput contains results for Critical Mode analysis.
// It includes the overall grade, summary, and a list of issues.
type CriticalModeOutput struct {
	OverallGrade string      `json:"overall_grade"`
	Summary      string      `json:"summary"`
	Issues       []CodeIssue `json:"issues"`
}

// ====================================================================================
// SUPPORTING STRUCTURES
// ====================================================================================

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
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Purpose     string   `json:"purpose"`
	Methods     []string `json:"methods"`
}

// DataModelInfo describes a data model and its fields.
// It includes the model name, fields, and purpose.
type DataModelInfo struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Purpose     string   `json:"purpose"`
	Fields      []string `json:"fields"`
}

// WorkflowInfo describes a workflow and its steps.
// It includes the workflow name and a list of steps.
type WorkflowInfo struct {
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
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

// ModelInfo represents information about an AI model
type ModelInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Size        string `json:"size"` // e.g., "7B", "16B"
	Available   bool   `json:"available"`
}

// ====================================================================================
// CONSTANTS - Mode Identifiers
// ====================================================================================

// PreviewMode is the string identifier for Preview Mode analysis
const PreviewMode = "preview"

// SkimMode is the string identifier for Skim Mode analysis
// SkimMode is the string identifier for Skim Mode analysis.
const SkimMode = "skim"

// ScanMode is the string identifier for Scan Mode analysis.
// ScanMode is the string identifier for Scan Mode analysis.
const ScanMode = "scan"

// DetailedMode is the string identifier for Detailed Mode analysis
const DetailedMode = "detailed"

// CriticalMode is the string identifier for Critical Mode analysis.
// CriticalMode is the string identifier for Critical Mode analysis.
const CriticalMode = "critical"

// ====================================================================================
// DATABASE MODELS
// ====================================================================================

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

// Review represents a code review session.
// It includes the session ID, title, code source, and timestamps.
type Review struct {
	Title        string `json:"title"`
	CodeSource   string `json:"code_source"`
	CreatedAt    string `json:"created_at"`
	LastAccessed string `json:"last_accessed"`
	ID           int64  `json:"id"`
}
