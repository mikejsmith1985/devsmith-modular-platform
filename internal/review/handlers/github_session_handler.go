package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	review_db "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/github"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
)

// GitHubSessionHandler handles GitHub session HTTP endpoints
type GitHubSessionHandler struct {
	repo         review_db.GitHubRepositoryInterface
	githubClient github.ClientInterface
	aiAnalyzer   *review_services.MultiFileAnalyzer
}

// NewGitHubSessionHandler creates a new GitHub session handler
func NewGitHubSessionHandler(
	repo review_db.GitHubRepositoryInterface,
	client github.ClientInterface,
	aiAnalyzer *review_services.MultiFileAnalyzer,
) *GitHubSessionHandler {
	return &GitHubSessionHandler{
		repo:         repo,
		githubClient: client,
		aiAnalyzer:   aiAnalyzer,
	}
}

// CreateSessionRequest represents the request to create a GitHub session
type CreateSessionRequest struct {
	SessionID int64  `json:"session_id" binding:"required"`
	GitHubURL string `json:"github_url" binding:"required"`
	Branch    string `json:"branch"`
	Token     string `json:"token"` // GitHub API token (optional)
}

// CreateSessionResponse represents the response after creating a session
type CreateSessionResponse struct {
	ID               int64                       `json:"id"`
	SessionID        int64                       `json:"session_id"`
	Owner            string                      `json:"owner"`
	Repo             string                      `json:"repo"`
	Branch           string                      `json:"branch"`
	TotalFiles       int                         `json:"total_files"`
	TotalDirectories int                         `json:"total_directories"`
	FileTree         *review_models.FileTreeJSON `json:"file_tree,omitempty"`
	CreatedAt        string                      `json:"created_at"`
}

// CreateSession creates a new GitHub session with repository tree
func (h *GitHubSessionHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Validate and parse GitHub URL
	owner, repo, err := h.githubClient.ValidateURL(req.GitHubURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid GitHub URL", "details": err.Error()})
		return
	}

	// Default branch if not specified
	branch := req.Branch
	if branch == "" {
		branch = "main"
	}

	// Fetch repository tree
	repoTree, err := h.githubClient.GetRepoTree(c.Request.Context(), owner, repo, branch, req.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch repository tree", "details": err.Error()})
		return
	}

	// Convert github.TreeNode to review_models.TreeNode
	convertedNodes := convertTreeNodes(repoTree.RootNodes)

	// Marshal tree to JSONB
	treeJSON := &review_models.FileTreeJSON{RootNodes: convertedNodes}
	treeData, err := review_db.MarshalFileTree(treeJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal tree", "details": err.Error()})
		return
	}

	// Count files and directories
	totalFiles, totalDirs := countTreeNodes(convertedNodes)

	// Create GitHub session
	session := &review_models.GitHubSession{
		SessionID:        req.SessionID,
		GitHubURL:        req.GitHubURL,
		Owner:            owner,
		Repo:             repo,
		Branch:           branch,
		FileTree:         treeData,
		TotalFiles:       totalFiles,
		TotalDirectories: totalDirs,
	}

	err = h.repo.CreateGitHubSession(c.Request.Context(), session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session", "details": err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusCreated, CreateSessionResponse{
		ID:               session.ID,
		SessionID:        session.SessionID,
		Owner:            owner,
		Repo:             repo,
		Branch:           branch,
		TotalFiles:       totalFiles,
		TotalDirectories: totalDirs,
		FileTree:         treeJSON,
		CreatedAt:        session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetSession retrieves a GitHub session by ID
func (h *GitHubSessionHandler) GetSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	session, err := h.repo.GetGitHubSession(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "github session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session", "details": err.Error()})
		return
	}

	// Parse file tree
	fileTree, err := review_db.ParseFileTree(session.FileTree)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse file tree", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                session.ID,
		"session_id":        session.SessionID,
		"github_url":        session.GitHubURL,
		"owner":             session.Owner,
		"repo":              session.Repo,
		"branch":            session.Branch,
		"total_files":       session.TotalFiles,
		"total_directories": session.TotalDirectories,
		"file_tree":         fileTree,
		"created_at":        session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetTree retrieves the file tree for a session
func (h *GitHubSessionHandler) GetTree(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	session, err := h.repo.GetGitHubSession(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "github session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session", "details": err.Error()})
		return
	}

	// Parse file tree
	fileTree, err := review_db.ParseFileTree(session.FileTree)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse file tree", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"owner":       session.Owner,
		"repo":        session.Repo,
		"branch":      session.Branch,
		"tree":        fileTree,
		"last_synced": session.TreeLastSynced,
	})
}

// OpenFileRequest represents the request to open a file
type OpenFileRequest struct {
	FilePath string `json:"file_path" binding:"required"`
	Token    string `json:"token"` // GitHub API token
}

// OpenFile opens a file in a new tab
func (h *GitHubSessionHandler) OpenFile(c *gin.Context) {
	idStr := c.Param("id")
	githubSessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var req OpenFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Get GitHub session
	session, err := h.repo.GetGitHubSession(c.Request.Context(), githubSessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Fetch file content from GitHub
	fileContent, err := h.githubClient.GetFileContent(c.Request.Context(), session.Owner, session.Repo, req.FilePath, session.Branch, req.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch file content", "details": err.Error()})
		return
	}

	// Get existing open files to determine tab order
	openFiles, err := h.repo.GetOpenFiles(c.Request.Context(), githubSessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get open files", "details": err.Error()})
		return
	}

	// Create new open file entry
	tabID := uuid.New()
	openFile := &review_models.OpenFile{
		GitHubSessionID: githubSessionID,
		TabID:           tabID,
		FilePath:        req.FilePath,
		FileSHA:         fileContent.SHA,
		FileContent:     fileContent.Content,
		FileSize:        fileContent.Size,
		Language:        detectLanguage(req.FilePath),
		IsActive:        true,           // New tab is active
		TabOrder:        len(openFiles), // Append to end
	}

	err = h.repo.CreateOpenFile(c.Request.Context(), openFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file", "details": err.Error()})
		return
	}

	// Set this tab as active
	err = h.repo.SetActiveTab(c.Request.Context(), githubSessionID, tabID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set active tab", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"tab_id":       tabID,
		"file_path":    openFile.FilePath,
		"file_content": openFile.FileContent,
		"file_size":    openFile.FileSize,
		"language":     openFile.Language,
		"tab_order":    openFile.TabOrder,
		"opened_at":    openFile.OpenedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetOpenFiles retrieves all open files for a session
func (h *GitHubSessionHandler) GetOpenFiles(c *gin.Context) {
	idStr := c.Param("id")
	githubSessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	files, err := h.repo.GetOpenFiles(c.Request.Context(), githubSessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get open files", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
		"count": len(files),
	})
}

// CloseFile closes a file tab
func (h *GitHubSessionHandler) CloseFile(c *gin.Context) {
	tabIDStr := c.Param("tab_id")
	tabID, err := uuid.Parse(tabIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tab ID"})
		return
	}

	err = h.repo.CloseFile(c.Request.Context(), tabID)
	if err != nil {
		if err.Error() == "file not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close file", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File closed successfully"})
}

// SetActiveTabRequest represents request to set active tab
type SetActiveTabRequest struct {
	TabID string `json:"tab_id" binding:"required"`
}

// SetActiveTab sets a specific tab as active
func (h *GitHubSessionHandler) SetActiveTab(c *gin.Context) {
	idStr := c.Param("id")
	githubSessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var req SetActiveTabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	tabID, err := uuid.Parse(req.TabID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tab ID"})
		return
	}

	err = h.repo.SetActiveTab(c.Request.Context(), githubSessionID, tabID)
	if err != nil {
		if err.Error() == "tab not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tab not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set active tab", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tab activated successfully"})
}

// MultiFileAnalysisRequest represents request to analyze multiple files
type MultiFileAnalysisRequest struct {
	FilePaths   []string `json:"file_paths" binding:"required,min=2"`
	ReadingMode string   `json:"reading_mode" binding:"required"`
	UserMode    string   `json:"user_mode"`   // beginner, novice, intermediate, expert
	OutputMode  string   `json:"output_mode"` // quick, full
}

// AnalyzeMultipleFiles analyzes multiple files together
func (h *GitHubSessionHandler) AnalyzeMultipleFiles(c *gin.Context) {
	idStr := c.Param("id")
	githubSessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var req MultiFileAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Get GitHub session
	session, err := h.repo.GetGitHubSession(c.Request.Context(), githubSessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Build file contents for analysis
	var fileContents []review_services.FileContent
	for _, path := range req.FilePaths {
		// In production, fetch actual file content from GitHub
		content := fmt.Sprintf("// Content for %s in %s/%s\n", path, session.Owner, session.Repo)
		fileContents = append(fileContents, review_services.FileContent{
			Path:    path,
			Content: content,
		})
	}

	// Call AI analyzer service
	analyzeReq := &review_services.AnalyzeRequest{
		Files:       fileContents,
		ReadingMode: req.ReadingMode,
		Temperature: 0.3, // Lower temperature for more consistent analysis
	}

	result, err := h.aiAnalyzer.Analyze(c.Request.Context(), analyzeReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI analysis failed", "details": err.Error()})
		return
	}

	// Convert result to AIAnalysisResponse format for storage
	aiResponse := &review_models.AIAnalysisResponse{
		Summary:              result.Summary,
		Dependencies:         result.Dependencies,
		SharedAbstractions:   result.SharedAbstractions,
		ArchitecturePatterns: result.ArchitecturePatterns,
		Recommendations:      result.Recommendations,
	}

	// Create multi-file analysis record
	analysis := &review_models.MultiFileAnalysis{
		GitHubSessionID:    githubSessionID,
		FilePaths:          req.FilePaths,
		ReadingMode:        req.ReadingMode,
		CombinedContent:    "", // Optional: store combined content if needed
		AnalysisDurationMs: result.DurationMs,
	}

	aiResponseData, _ := json.Marshal(aiResponse)
	analysis.AIResponse = aiResponseData

	err = h.repo.CreateMultiFileAnalysis(c.Request.Context(), analysis)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create analysis", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis_id":   analysis.ID,
		"file_paths":    req.FilePaths,
		"reading_mode":  req.ReadingMode,
		"ai_response":   aiResponse,
		"duration_ms":   result.DurationMs,
		"input_tokens":  result.InputTokens,
		"output_tokens": result.OutputTokens,
		"created_at":    analysis.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// Helper functions

func countTreeNodes(nodes []review_models.TreeNode) (files, dirs int) {
	for _, node := range nodes {
		if node.Type == "file" {
			files++
		} else if node.Type == "dir" {
			dirs++
			childFiles, childDirs := countTreeNodes(node.Children)
			files += childFiles
			dirs += childDirs
		}
	}
	return
}

func detectLanguage(filePath string) string {
	// Simple language detection based on file extension
	ext := ""
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '.' {
			ext = filePath[i+1:]
			break
		}
	}

	switch ext {
	case "go":
		return "go"
	case "js", "jsx":
		return "javascript"
	case "ts", "tsx":
		return "typescript"
	case "py":
		return "python"
	case "java":
		return "java"
	case "rb":
		return "ruby"
	case "php":
		return "php"
	case "c", "h":
		return "c"
	case "cpp", "hpp", "cc":
		return "cpp"
	case "cs":
		return "csharp"
	case "rs":
		return "rust"
	case "md":
		return "markdown"
	case "json":
		return "json"
	case "yaml", "yml":
		return "yaml"
	case "xml":
		return "xml"
	case "html":
		return "html"
	case "css":
		return "css"
	case "sql":
		return "sql"
	default:
		return "text"
	}
}

// convertTreeNodes converts github.TreeNode to review_models.TreeNode
func convertTreeNodes(nodes []github.TreeNode) []review_models.TreeNode {
	result := make([]review_models.TreeNode, len(nodes))
	for i, node := range nodes {
		result[i] = review_models.TreeNode{
			Path:     node.Path,
			Type:     node.Type,
			SHA:      node.SHA,
			Size:     node.Size,
			Children: convertTreeNodes(node.Children), // Recursive conversion
		}
	}
	return result
}
