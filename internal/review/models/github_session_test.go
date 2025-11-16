package review_models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubSession_Structure(t *testing.T) {
	// GIVEN: GitHub session data
	now := time.Now()
	session := &GitHubSession{
		ID:               1,
		SessionID:        42,
		GitHubURL:        "https://github.com/golang/go",
		Owner:            "golang",
		Repo:             "go",
		Branch:           "main",
		CommitSHA:        "abc123",
		TotalFiles:       100,
		TotalDirectories: 20,
		TreeLastSynced:   now,
		IsPrivate:        false,
		StarsCount:       120000,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// THEN: All fields should be accessible
	require.NotNil(t, session)
	assert.Equal(t, int64(1), session.ID)
	assert.Equal(t, int64(42), session.SessionID)
	assert.Equal(t, "https://github.com/golang/go", session.GitHubURL)
	assert.Equal(t, "golang", session.Owner)
	assert.Equal(t, "go", session.Repo)
	assert.Equal(t, "main", session.Branch)
	assert.Equal(t, "abc123", session.CommitSHA)
	assert.Equal(t, 100, session.TotalFiles)
	assert.Equal(t, 20, session.TotalDirectories)
	assert.False(t, session.IsPrivate)
	assert.Equal(t, 120000, session.StarsCount)
}

func TestOpenFile_Structure(t *testing.T) {
	// GIVEN: Open file data
	tabID := uuid.New()
	now := time.Now()
	openFile := &OpenFile{
		ID:              1,
		GitHubSessionID: 42,
		TabID:           tabID,
		FilePath:        "internal/main.go",
		FileSHA:         "def456",
		FileContent:     "package main\n\nfunc main() {}",
		FileSize:        1024,
		Language:        "go",
		IsActive:        true,
		TabOrder:        0,
		OpenedAt:        now,
		LastAccessed:    now,
		AnalysisCount:   3,
	}

	// THEN: All fields should be accessible
	require.NotNil(t, openFile)
	assert.Equal(t, int64(1), openFile.ID)
	assert.Equal(t, int64(42), openFile.GitHubSessionID)
	assert.Equal(t, tabID, openFile.TabID)
	assert.Equal(t, "internal/main.go", openFile.FilePath)
	assert.Equal(t, "def456", openFile.FileSHA)
	assert.Contains(t, openFile.FileContent, "package main")
	assert.Equal(t, int64(1024), openFile.FileSize)
	assert.Equal(t, "go", openFile.Language)
	assert.True(t, openFile.IsActive)
	assert.Equal(t, 0, openFile.TabOrder)
	assert.Equal(t, 3, openFile.AnalysisCount)
}

func TestMultiFileAnalysis_Structure(t *testing.T) {
	// GIVEN: Multi-file analysis data
	now := time.Now()
	analysis := &MultiFileAnalysis{
		ID:                 1,
		GitHubSessionID:    42,
		FilePaths:          []string{"main.go", "handler.go", "service.go"},
		ReadingMode:        "critical",
		CombinedContent:    "// Combined file contents",
		AnalysisDurationMs: 5000,
		CreatedAt:          now,
	}

	// THEN: All fields should be accessible
	require.NotNil(t, analysis)
	assert.Equal(t, int64(1), analysis.ID)
	assert.Equal(t, int64(42), analysis.GitHubSessionID)
	assert.Len(t, analysis.FilePaths, 3)
	assert.Equal(t, "main.go", analysis.FilePaths[0])
	assert.Equal(t, "handler.go", analysis.FilePaths[1])
	assert.Equal(t, "service.go", analysis.FilePaths[2])
	assert.Equal(t, "critical", analysis.ReadingMode)
	assert.Equal(t, int64(5000), analysis.AnalysisDurationMs)
}

func TestTreeNode_Structure(t *testing.T) {
	// GIVEN: Tree node with children
	treeNode := &TreeNode{
		Path: "src",
		Type: "dir",
		SHA:  "abc123",
		Size: 0,
		Children: []TreeNode{
			{
				Path: "src/main.go",
				Type: "file",
				SHA:  "def456",
				Size: 2048,
			},
			{
				Path: "src/handler.go",
				Type: "file",
				SHA:  "ghi789",
				Size: 1024,
			},
		},
	}

	// THEN: Tree structure should be correct
	require.NotNil(t, treeNode)
	assert.Equal(t, "src", treeNode.Path)
	assert.Equal(t, "dir", treeNode.Type)
	assert.Len(t, treeNode.Children, 2)
	assert.Equal(t, "src/main.go", treeNode.Children[0].Path)
	assert.Equal(t, "file", treeNode.Children[0].Type)
	assert.Equal(t, int64(2048), treeNode.Children[0].Size)
}

func TestFileTreeJSON_Structure(t *testing.T) {
	// GIVEN: File tree JSON structure
	fileTree := &FileTreeJSON{
		RootNodes: []TreeNode{
			{
				Path: "README.md",
				Type: "file",
				SHA:  "abc123",
				Size: 1024,
			},
			{
				Path: "src",
				Type: "dir",
				SHA:  "def456",
				Children: []TreeNode{
					{
						Path: "src/main.go",
						Type: "file",
						SHA:  "ghi789",
						Size: 2048,
					},
				},
			},
		},
	}

	// THEN: Structure should be valid
	require.NotNil(t, fileTree)
	assert.Len(t, fileTree.RootNodes, 2)
	assert.Equal(t, "README.md", fileTree.RootNodes[0].Path)
	assert.Equal(t, "file", fileTree.RootNodes[0].Type)
	assert.Equal(t, "src", fileTree.RootNodes[1].Path)
	assert.Equal(t, "dir", fileTree.RootNodes[1].Type)
	assert.Len(t, fileTree.RootNodes[1].Children, 1)
}

func TestCrossFileDependency_Structure(t *testing.T) {
	// GIVEN: Cross-file dependency
	dep := &CrossFileDependency{
		FromFile:   "main.go",
		ToFile:     "handler.go",
		ImportType: "import",
		Symbols:    []string{"HandleRequest", "NewHandler"},
	}

	// THEN: Dependency structure should be valid
	require.NotNil(t, dep)
	assert.Equal(t, "main.go", dep.FromFile)
	assert.Equal(t, "handler.go", dep.ToFile)
	assert.Equal(t, "import", dep.ImportType)
	assert.Len(t, dep.Symbols, 2)
	assert.Contains(t, dep.Symbols, "HandleRequest")
	assert.Contains(t, dep.Symbols, "NewHandler")
}

func TestSharedAbstraction_Structure(t *testing.T) {
	// GIVEN: Shared abstraction
	abstraction := &SharedAbstraction{
		Name:        "Handler",
		Type:        "interface",
		Files:       []string{"handler.go", "middleware.go"},
		Description: "Common HTTP handler interface",
		Complexity:  "simple",
	}

	// THEN: Abstraction structure should be valid
	require.NotNil(t, abstraction)
	assert.Equal(t, "Handler", abstraction.Name)
	assert.Equal(t, "interface", abstraction.Type)
	assert.Len(t, abstraction.Files, 2)
	assert.Equal(t, "Common HTTP handler interface", abstraction.Description)
	assert.Equal(t, "simple", abstraction.Complexity)
}

func TestArchitecturePattern_Structure(t *testing.T) {
	// GIVEN: Architecture pattern
	pattern := &ArchitecturePattern{
		Pattern:     "Repository",
		Confidence:  0.95,
		Files:       []string{"repository.go", "user_repository.go"},
		Description: "Repository pattern for data access",
		Evidence:    []string{"Repository interface defined", "Concrete implementations exist"},
	}

	// THEN: Pattern structure should be valid
	require.NotNil(t, pattern)
	assert.Equal(t, "Repository", pattern.Pattern)
	assert.Equal(t, 0.95, pattern.Confidence)
	assert.Len(t, pattern.Files, 2)
	assert.Equal(t, "Repository pattern for data access", pattern.Description)
	assert.Len(t, pattern.Evidence, 2)
}

func TestAIAnalysisResponse_Structure(t *testing.T) {
	// GIVEN: AI analysis response
	response := &AIAnalysisResponse{
		Summary: "Architecture analysis of 3 files",
		Dependencies: []CrossFileDependency{
			{
				FromFile:   "main.go",
				ToFile:     "handler.go",
				ImportType: "import",
			},
		},
		SharedAbstractions: []SharedAbstraction{
			{
				Name:  "Handler",
				Type:  "interface",
				Files: []string{"handler.go", "middleware.go"},
			},
		},
		ArchitecturePatterns: []ArchitecturePattern{
			{
				Pattern:    "MVC",
				Confidence: 0.85,
				Files:      []string{"main.go", "handler.go", "model.go"},
			},
		},
		Recommendations: []string{"Consider adding dependency injection", "Extract common interfaces"},
		Issues: []AnalysisIssue{
			{
				File:        "main.go",
				Line:        42,
				Severity:    "high",
				Category:    "architecture",
				Description: "Direct database access in handler",
				Suggestion:  "Use repository pattern",
			},
		},
	}

	// THEN: Response structure should be valid
	require.NotNil(t, response)
	assert.Equal(t, "Architecture analysis of 3 files", response.Summary)
	assert.Len(t, response.Dependencies, 1)
	assert.Len(t, response.SharedAbstractions, 1)
	assert.Len(t, response.ArchitecturePatterns, 1)
	assert.Len(t, response.Recommendations, 2)
	assert.Len(t, response.Issues, 1)
	assert.Equal(t, "high", response.Issues[0].Severity)
	assert.Equal(t, "architecture", response.Issues[0].Category)
}

func TestAnalysisIssue_Structure(t *testing.T) {
	// GIVEN: Analysis issue
	issue := &AnalysisIssue{
		File:        "handler.go",
		Line:        123,
		Severity:    "critical",
		Category:    "security",
		Description: "SQL injection vulnerability",
		Suggestion:  "Use parameterized queries",
	}

	// THEN: Issue structure should be valid
	require.NotNil(t, issue)
	assert.Equal(t, "handler.go", issue.File)
	assert.Equal(t, 123, issue.Line)
	assert.Equal(t, "critical", issue.Severity)
	assert.Equal(t, "security", issue.Category)
	assert.Equal(t, "SQL injection vulnerability", issue.Description)
	assert.Equal(t, "Use parameterized queries", issue.Suggestion)
}
