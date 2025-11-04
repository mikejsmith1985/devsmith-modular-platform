package handlers

import (
	"testing"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/github"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/stretchr/testify/assert"
)

// Test helper functions

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"main.go", "go"},
		{"app.js", "javascript"},
		{"component.tsx", "typescript"},
		{"script.py", "python"},
		{"README.md", "markdown"},
		{"config.json", "json"},
		{"docker-compose.yml", "yaml"},
		{"unknown.xyz", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detectLanguage(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountTreeNodes(t *testing.T) {
	nodes := []review_models.TreeNode{
		{Path: "main.go", Type: "file", SHA: "abc123"},
		{Path: "src", Type: "dir", SHA: "def456", Children: []review_models.TreeNode{
			{Path: "src/app.go", Type: "file", SHA: "ghi789"},
			{Path: "src/utils", Type: "dir", SHA: "jkl012", Children: []review_models.TreeNode{
				{Path: "src/utils/helper.go", Type: "file", SHA: "mno345"},
			}},
		}},
	}

	files, dirs := countTreeNodes(nodes)
	assert.Equal(t, 3, files)
	assert.Equal(t, 2, dirs)
}

func TestConvertTreeNodes(t *testing.T) {
	githubNodes := []github.TreeNode{
		{Path: "main.go", Type: "file", SHA: "abc123", Size: 100},
		{Path: "src", Type: "dir", SHA: "def456", Children: []github.TreeNode{
			{Path: "src/app.go", Type: "file", SHA: "ghi789", Size: 200},
		}},
	}

	converted := convertTreeNodes(githubNodes)

	assert.Equal(t, 2, len(converted))
	assert.Equal(t, "main.go", converted[0].Path)
	assert.Equal(t, "file", converted[0].Type)
	assert.Equal(t, "abc123", converted[0].SHA)
	assert.Equal(t, int64(100), converted[0].Size)

	assert.Equal(t, "src", converted[1].Path)
	assert.Equal(t, 1, len(converted[1].Children))
	assert.Equal(t, "src/app.go", converted[1].Children[0].Path)
	assert.Equal(t, int64(200), converted[1].Children[0].Size)
}

func TestConvertTreeNodes_EmptySlice(t *testing.T) {
	converted := convertTreeNodes([]github.TreeNode{})
	assert.Equal(t, 0, len(converted))
}

func TestConvertTreeNodes_DeeplyNested(t *testing.T) {
	githubNodes := []github.TreeNode{
		{
			Path: "root", 
			Type: "dir", 
			SHA: "a", 
			Children: []github.TreeNode{
				{
					Path: "root/level1", 
					Type: "dir", 
					SHA: "b", 
					Children: []github.TreeNode{
						{
							Path: "root/level1/level2", 
							Type: "dir", 
							SHA: "c", 
							Children: []github.TreeNode{
								{Path: "root/level1/level2/deep.go", Type: "file", SHA: "d", Size: 42},
							},
						},
					},
				},
			},
		},
	}

	converted := convertTreeNodes(githubNodes)

	assert.Equal(t, 1, len(converted))
	assert.Equal(t, "root", converted[0].Path)
	assert.Equal(t, 1, len(converted[0].Children))
	assert.Equal(t, "root/level1", converted[0].Children[0].Path)
	assert.Equal(t, 1, len(converted[0].Children[0].Children))
	assert.Equal(t, "root/level1/level2", converted[0].Children[0].Children[0].Path)
	assert.Equal(t, 1, len(converted[0].Children[0].Children[0].Children))
	assert.Equal(t, "root/level1/level2/deep.go", converted[0].Children[0].Children[0].Children[0].Path)
	assert.Equal(t, int64(42), converted[0].Children[0].Children[0].Children[0].Size)
}
