package models

// SkimMode is the string identifier for Skim Mode analysis
const SkimMode = "skim"

// AnalysisResult represents a cached/captured analysis result
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
type SkimModeOutput struct {
	Functions  []FunctionSignature `json:"functions"`
	Interfaces []InterfaceInfo     `json:"interfaces"`
	DataModels []DataModelInfo     `json:"data_models"`
	Workflows  []WorkflowInfo      `json:"workflows"`
	Summary    string              `json:"summary"`
}

type FunctionSignature struct {
	Name        string `json:"name"`
	Signature   string `json:"signature"`
	Description string `json:"description"`
}

type InterfaceInfo struct {
	Name    string   `json:"name"`
	Methods []string `json:"methods"`
	Purpose string   `json:"purpose"`
}

type DataModelInfo struct {
	Name    string   `json:"name"`
	Fields  []string `json:"fields"`
	Purpose string   `json:"purpose"`
}

type WorkflowInfo struct {
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
}
