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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	repoTree       *github.RepoTree
	fileContent    *github.FileContent
	repoMetadata   *github.RepoMetadata
	codeFetch      *github.CodeFetch
	getTreeError   error
	getContentErr  error
	fetchCodeErr   error
}

func (m *mockGitHubClient) FetchCode(ctx context.Context, owner, repo, branch string, token string) (*github.CodeFetch, error) {
	if m.fetchCodeErr != nil {
		return nil, m.fetchCodeErr
	}
	if m.codeFetch != nil {
		return m.codeFetch, nil
	}
	// Default mock response
	return &github.CodeFetch{
		Code:      "// Mock code content",
		CommitSHA: "abc123",
		Branch:    branch,
		Metadata: &github.RepoMetadata{
			Owner: owner,
			Name:  repo,
		},
	}, nil
}

func (m *mockGitHubClient) GetRepoMetadata(ctx context.Context, owner, repo string, token string) (*github.RepoMetadata, error) {
	if m.repoMetadata != nil {
		return m.repoMetadata, nil
	}
	return &github.RepoMetadata{
		Owner:      owner,
		Name:       repo,
		StarsCount: 100,
		IsPrivate:  false,
	}, nil
}

func (m *mockGitHubClient) ValidateURL(url string) (owner string, repo string, err error) {
	return "test", "repo", nil
}

func (m *mockGitHubClient) GetRateLimit(ctx context.Context, token string) (remaining int, resetTime time.Time, err error) {
	return 5000, time.Now().Add(1 * time.Hour), nil
}

func (m *mockGitHubClient) GetRepoTree(ctx context.Context, owner, repo, branch, token string) (*github.RepoTree, error) {
	if m.getTreeError != nil {
		return nil, m.getTreeError
	}
	if m.repoTree != nil {
		return m.repoTree, nil
	}
	return &github.RepoTree{
		Owner:  owner,
		Repo:   repo,
		Branch: branch,
		RootNodes: []github.TreeNode{
			{Path: "main.go", Type: "file", Size: 1024},
			{Path: "handler.go", Type: "file", Size: 2048},
			{Path: "service.go", Type: "file", Size: 3072},
		},
	}, nil
}

func (m *mockGitHubClient) GetFileContent(ctx context.Context, owner, repo, path, branch, token string) (*github.FileContent, error) {
	if m.getContentErr != nil {
		return nil, m.getContentErr
	}
	if m.fileContent != nil {
		return m.fileContent, nil
	}
	return &github.FileContent{
		Path:    path,
		Content: "// Mock file content for " + path,
		SHA:     "file123",
		Size:    100,
	}, nil
}

func (m *mockGitHubClient) GetPullRequest(ctx context.Context, owner, repo string, prNumber int, token string) (*github.PullRequest, error) {
	return nil, nil // Not needed for these tests
}

func (m *mockGitHubClient) GetPRFiles(ctx context.Context, owner, repo string, prNumber int, token string) ([]github.PRFile, error) {
	return nil, nil // Not needed for these tests
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

func (m *mockAIProvider) GetModelInfo() *ai.ModelInfo {
	return &ai.ModelInfo{
		Provider:    "mock",
		Model:       "test-model",
		DisplayName: "Mock Test Model",
		MaxTokens:   4000,
	}
}

// setupTestHandler creates a test handler with in-memory repository and mocks
func setupTestHandler() (*handlers.GitHubSessionHandler, review_db.GitHubRepositoryInterface) {
	repo := review_db.NewInMemoryGitHubRepository()

	mockGitHub := &mockGitHubClient{
		repoTree: &github.RepoTree{
			Owner:  "test",
			Repo:   "repo",
			Branch: "main",
			RootNodes: []github.TreeNode{
				{Path: "main.go", Type: "file", Size: 1024},
				{Path: "handler.go", Type: "file", Size: 2048},
				{Path: "service.go", Type: "file", Size: 3072},
			},
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
	sessionID := int64(response["id"].(float64))
	storedSession, err := repo.GetGitHubSession(context.Background(), sessionID)
	require.NoError(t, err)
	assert.Equal(t, sessionID, storedSession.ID)
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
		TabID:           uuid.Must(uuid.Parse("00000000-0000-0000-0000-000000000001")),
		FilePath:        "main.go",
		IsActive:        true,
		TabOrder:        0,
	}
	file2 := &review_models.OpenFile{
		GitHubSessionID: session.ID,
		TabID:           uuid.Must(uuid.Parse("00000000-0000-0000-0000-000000000002")),
		FilePath:        "handler.go",
		IsActive:        false,
		TabOrder:        1,
	}
	err = repo.CreateOpenFile(context.Background(), file1)
	require.NoError(t, err)
	err = repo.CreateOpenFile(context.Background(), file2)
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
