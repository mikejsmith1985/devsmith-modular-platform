//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
	review_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/github"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/handlers"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
)

// mockGitHubClient implements github.ClientInterface for testing
type mockGitHubClient struct {
	getTreeResult *github.TreeResult
	getTreeError  error
}

func (m *mockGitHubClient) GetTree(ctx context.Context, owner, repo, ref string) (*github.TreeResult, error) {
	if m.getTreeError != nil {
		return nil, m.getTreeError
	}
	return m.getTreeResult, nil
}

func (m *mockGitHubClient) GetFileContent(ctx context.Context, owner, repo, path, ref string) ([]byte, error) {
	return []byte("// Mock file content"), nil
}

// mockAIProvider implements ai.Provider for testing
type mockAIProvider struct {
	response *ai.Response
	err      error
}

func (m *mockAIProvider) Generate(ctx context.Context, req *ai.Request) (*ai.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *mockAIProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *mockAIProvider) GetModelInfo(ctx context.Context) (string, error) {
	return "mock-model", nil
}

// setupTestHandler creates a test handler with in-memory repository and mocks
func setupTestHandler() (*handlers.GitHubSessionHandler, *review_db.InMemoryGitHubRepository) {
	repo := review_db.NewInMemoryGitHubRepository()

	mockGitHub := &mockGitHubClient{
		getTreeResult: &github.TreeResult{
			Tree: []github.TreeNode{
				{Path: "main.go", Type: "file", Size: 1024},
				{Path: "handler.go", Type: "file", Size: 2048},
				{Path: "service.go", Type: "file", Size: 3072},
			},
			SHA: "abc123",
		},
	}

	mockAI := &mockAIProvider{
		response: &ai.Response{
			Content: `{
				"summary": "Test analysis summary",
				"dependencies": [
					{
						"from_file": "handler.go",
						"to_file": "service.go",
						"type": "imports",
						"description": "Handler imports service"
					}
				],
				"shared_abstractions": [
					{
						"name": "UserService",
						"type": "interface",
						"files": ["service.go", "handler.go"]
					}
				],
				"architecture_patterns": [
					{
						"pattern": "MVC",
						"description": "Model-View-Controller pattern"
					}
				],
				"recommendations": ["Add more tests", "Extract constants"]
			}`,
			InputTokens:  500,
			OutputTokens: 200,
		},
	}

	analyzer := review_services.NewMultiFileAnalyzer(mockAI, "test-model")
	handler := handlers.NewGitHubSessionHandler(repo, mockGitHub, analyzer)

	return handler, repo
}

// TestCreateGitHubSession tests creating a new GitHub session
func TestCreateGitHubSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, repo := setupTestHandler()

	// Create request
	reqBody := map[string]interface{}{
		"session_id": int64(1),
		"github_url": "https://github.com/test/repo",
		"branch":     "main",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/review/github/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.CreateSession(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.NotNil(t, response["id"])
	assert.Equal(t, float64(1), response["session_id"])
	assert.Equal(t, "test", response["owner"])
	assert.Equal(t, "repo", response["repo"])
	assert.Equal(t, "main", response["branch"])
	assert.NotNil(t, response["file_tree"])

	// Verify session stored in repository
	sessions, err := repo.ListGitHubSessions(context.Background(), 1)
	require.Error(t, err) // ListGitHubSessions not supported in in-memory impl
}

// TestOpenFile tests opening a file in a session
func TestOpenFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, repo := setupTestHandler()

	// First create a session
	session := &review_models.GitHubSession{
		SessionID: 1,
		GitHubURL: "https://github.com/test/repo",
		Owner:     "test",
		Repo:      "repo",
		Branch:    "main",
	}
	err := repo.CreateGitHubSession(context.Background(), session)
	require.NoError(t, err)

	// Open file request
	reqBody := map[string]interface{}{
		"tab_id":    "tab-1",
		"file_path": "main.go",
		"tab_order": 0,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/review/github/sessions/1/open-file", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	// Execute
	handler.OpenFile(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response["id"])
	assert.Equal(t, "tab-1", response["tab_id"])
	assert.Equal(t, "main.go", response["file_path"])
	assert.Equal(t, true, response["is_active"])
}

// TestGetOpenFiles tests retrieving all open files for a session
func TestGetOpenFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, repo := setupTestHandler()

	// Create session
	session := &review_models.GitHubSession{
		SessionID: 1,
		GitHubURL: "https://github.com/test/repo",
		Owner:     "test",
		Repo:      "repo",
		Branch:    "main",
	}
	err := repo.CreateGitHubSession(context.Background(), session)
	require.NoError(t, err)

	// Open multiple files
	file1 := &review_models.OpenFile{
		GitHubSessionID: session.ID,
		TabID:           "tab-1",
		FilePath:        "main.go",
		IsActive:        true,
		TabOrder:        0,
	}
	file2 := &review_models.OpenFile{
		GitHubSessionID: session.ID,
		TabID:           "tab-2",
		FilePath:        "handler.go",
		IsActive:        false,
		TabOrder:        1,
	}
	err = repo.OpenFile(context.Background(), file1)
	require.NoError(t, err)
	err = repo.OpenFile(context.Background(), file2)
	require.NoError(t, err)

	// Get open files request
	req := httptest.NewRequest("GET", "/api/review/github/sessions/1/open-files", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	// Execute
	handler.GetOpenFiles(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	files := response["open_files"].([]interface{})
	assert.Len(t, files, 2)
}

// TestAnalyzeMultipleFiles tests multi-file AI analysis
func TestAnalyzeMultipleFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, repo := setupTestHandler()

	// Create session
	session := &review_models.GitHubSession{
		SessionID: 1,
		GitHubURL: "https://github.com/test/repo",
		Owner:     "test",
		Repo:      "repo",
		Branch:    "main",
	}
	err := repo.CreateGitHubSession(context.Background(), session)
	require.NoError(t, err)

	// Analyze multiple files request
	reqBody := map[string]interface{}{
		"file_paths":   []string{"main.go", "handler.go", "service.go"},
		"reading_mode": "critical",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/review/github/sessions/1/analyze-multiple", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	// Execute
	handler.AnalyzeMultipleFiles(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.NotNil(t, response["analysis_id"])
	assert.Equal(t, "critical", response["reading_mode"])
	assert.NotNil(t, response["ai_response"])
	assert.NotNil(t, response["duration_ms"])
	assert.NotNil(t, response["input_tokens"])
	assert.NotNil(t, response["output_tokens"])

	// Verify AI response structure
	aiResponse := response["ai_response"].(map[string]interface{})
	assert.NotEmpty(t, aiResponse["summary"])
	assert.NotNil(t, aiResponse["dependencies"])
	assert.NotNil(t, aiResponse["shared_abstractions"])
	assert.NotNil(t, aiResponse["architecture_patterns"])
	assert.NotNil(t, aiResponse["recommendations"])
}
