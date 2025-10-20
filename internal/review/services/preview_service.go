package services

import (
	"context"
)

// PreviewResult holds the analysis for Preview Mode
type PreviewResult struct {
	FileTree                []FileNode
	BoundedContexts         []string
	TechStack               []string
	ArchitecturePattern     string
	EntryPoints             []string
	ExternalDependencies    []string
	Summary                 string
	FunctionImplementations []string // Should be empty in Preview Mode
}

// FileNode represents a node in the file/folder tree
type FileNode struct {
	Name        string
	Path        string
	Description string
	Children    []FileNode
}

// PreviewService provides Preview Mode analysis
type PreviewService struct {
	// Add dependencies here (e.g., AI client, cache)
}

func NewPreviewService() *PreviewService {
	return &PreviewService{}
}

// AnalyzePreview analyzes the codebase in Preview Mode
func (s *PreviewService) AnalyzePreview(ctx context.Context, codebasePath string) (*PreviewResult, error) {
	// TODO: Integrate AI analysis logic here
	// For now, return mock data
	return &PreviewResult{
		FileTree:                []FileNode{{Name: "main.go", Path: "/main.go", Description: "Main entry point", Children: nil}},
		BoundedContexts:         []string{"Auth domain", "Review domain"},
		TechStack:               []string{"Go", "Gin"},
		ArchitecturePattern:     "layered",
		EntryPoints:             []string{"main.go"},
		ExternalDependencies:    []string{"PostgreSQL", "Redis"},
		Summary:                 "This is a Go microservice using Gin and PostgreSQL.",
		FunctionImplementations: []string{},
	}, nil
}
