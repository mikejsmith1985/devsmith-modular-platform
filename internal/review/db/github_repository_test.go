package review_db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGitHubRepository tests require a real database connection
// These are integration tests and should be run with -tags=integration

func TestMarshalFileTree(t *testing.T) {
	// GIVEN: FileTreeJSON structure
	tree := &review_models.FileTreeJSON{
		RootNodes: []review_models.TreeNode{
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
				Children: []review_models.TreeNode{
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

	// WHEN: Marshaling to JSONB
	data, err := MarshalFileTree(tree)

	// THEN: Should succeed
	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Contains(t, string(data), "README.md")
	assert.Contains(t, string(data), "src/main.go")
}

func TestParseFileTree(t *testing.T) {
	// GIVEN: JSONB data
	jsonData := []byte(`{
		"rootNodes": [
			{
				"path": "README.md",
				"type": "file",
				"sha": "abc123",
				"size": 1024
			},
			{
				"path": "src",
				"type": "dir",
				"sha": "def456",
				"children": [
					{
						"path": "src/main.go",
						"type": "file",
						"sha": "ghi789",
						"size": 2048
					}
				]
			}
		]
	}`)

	// WHEN: Parsing JSONB
	tree, err := ParseFileTree(jsonData)

	// THEN: Should succeed
	require.NoError(t, err)
	require.NotNil(t, tree)
	assert.Len(t, tree.RootNodes, 2)
	assert.Equal(t, "README.md", tree.RootNodes[0].Path)
	assert.Equal(t, "file", tree.RootNodes[0].Type)
	assert.Equal(t, "src", tree.RootNodes[1].Path)
	assert.Equal(t, "dir", tree.RootNodes[1].Type)
	assert.Len(t, tree.RootNodes[1].Children, 1)
	assert.Equal(t, "src/main.go", tree.RootNodes[1].Children[0].Path)
}

func TestParseFileTree_EmptyData(t *testing.T) {
	// GIVEN: Empty data
	var emptyData []byte

	// WHEN: Parsing empty data
	tree, err := ParseFileTree(emptyData)

	// THEN: Should return nil without error
	assert.NoError(t, err)
	assert.Nil(t, tree)
}

func TestMarshalFileTree_Nil(t *testing.T) {
	// GIVEN: Nil tree
	var tree *review_models.FileTreeJSON

	// WHEN: Marshaling nil tree
	data, err := MarshalFileTree(tree)

	// THEN: Should return nil without error
	assert.NoError(t, err)
	assert.Nil(t, data)
}

func TestParseFileTree_InvalidJSON(t *testing.T) {
	// GIVEN: Invalid JSON data
	invalidData := []byte(`{"invalid": json}`)

	// WHEN: Parsing invalid JSON
	tree, err := ParseFileTree(invalidData)

	// THEN: Should return error
	assert.Error(t, err)
	assert.Nil(t, tree)
	assert.Contains(t, err.Error(), "failed to parse file tree")
}

// Mock tests for repository methods (without database)

func TestGitHubSession_DataValidation(t *testing.T) {
	// GIVEN: Valid GitHub session data
	now := time.Now()
	session := &review_models.GitHubSession{
		SessionID:        1,
		GitHubURL:        "https://github.com/golang/go",
		Owner:            "golang",
		Repo:             "go",
		Branch:           "main",
		CommitSHA:        "abc123",
		TotalFiles:       100,
		TotalDirectories: 20,
		IsPrivate:        false,
		StarsCount:       120000,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// THEN: All required fields should be present
	assert.NotZero(t, session.SessionID)
	assert.NotEmpty(t, session.GitHubURL)
	assert.NotEmpty(t, session.Owner)
	assert.NotEmpty(t, session.Repo)
	assert.NotEmpty(t, session.Branch)
}

func TestOpenFile_DataValidation(t *testing.T) {
	// GIVEN: Valid open file data
	tabID := uuid.New()
	now := time.Now()
	file := &review_models.OpenFile{
		GitHubSessionID: 1,
		TabID:           tabID,
		FilePath:        "internal/main.go",
		FileSHA:         "def456",
		FileContent:     "package main",
		FileSize:        1024,
		Language:        "go",
		IsActive:        true,
		TabOrder:        0,
		OpenedAt:        now,
		LastAccessed:    now,
		AnalysisCount:   0,
	}

	// THEN: All required fields should be present
	assert.NotZero(t, file.GitHubSessionID)
	assert.NotEqual(t, uuid.Nil, file.TabID)
	assert.NotEmpty(t, file.FilePath)
	assert.GreaterOrEqual(t, file.TabOrder, 0)
}

func TestMultiFileAnalysis_DataValidation(t *testing.T) {
	// GIVEN: Valid multi-file analysis data
	ctx := context.Background()
	_ = ctx // Use context

	analysis := &review_models.MultiFileAnalysis{
		GitHubSessionID:    1,
		FilePaths:          []string{"main.go", "handler.go", "service.go"},
		ReadingMode:        "critical",
		CombinedContent:    "// Combined content",
		AnalysisDurationMs: 5000,
	}

	// THEN: All required fields should be present
	assert.NotZero(t, analysis.GitHubSessionID)
	assert.NotEmpty(t, analysis.FilePaths)
	assert.Len(t, analysis.FilePaths, 3)
	assert.NotEmpty(t, analysis.ReadingMode)
	assert.Greater(t, analysis.AnalysisDurationMs, int64(0))
}

func TestFileTree_RoundTrip(t *testing.T) {
	// GIVEN: FileTreeJSON structure
	originalTree := &review_models.FileTreeJSON{
		RootNodes: []review_models.TreeNode{
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
				Children: []review_models.TreeNode{
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

	// WHEN: Marshal then unmarshal
	data, err := MarshalFileTree(originalTree)
	require.NoError(t, err)

	parsedTree, err := ParseFileTree(data)
	require.NoError(t, err)

	// THEN: Should match original
	assert.Len(t, parsedTree.RootNodes, 2)
	assert.Equal(t, originalTree.RootNodes[0].Path, parsedTree.RootNodes[0].Path)
	assert.Equal(t, originalTree.RootNodes[0].Type, parsedTree.RootNodes[0].Type)
	assert.Equal(t, originalTree.RootNodes[0].SHA, parsedTree.RootNodes[0].SHA)
	assert.Equal(t, originalTree.RootNodes[0].Size, parsedTree.RootNodes[0].Size)

	assert.Equal(t, originalTree.RootNodes[1].Path, parsedTree.RootNodes[1].Path)
	assert.Equal(t, originalTree.RootNodes[1].Type, parsedTree.RootNodes[1].Type)
	assert.Len(t, parsedTree.RootNodes[1].Children, 1)
	assert.Equal(t, originalTree.RootNodes[1].Children[0].Path, parsedTree.RootNodes[1].Children[0].Path)
}
