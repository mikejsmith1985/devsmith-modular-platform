package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v57/github"
	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
	"golang.org/x/oauth2"
)

// GitHubHandler handles GitHub repository integration endpoints
type GitHubHandler struct {
	logger         *logger.Logger
	previewService review_services.PreviewAnalyzer
}

// NewGitHubHandler creates a new GitHub handler
func NewGitHubHandler(logger *logger.Logger, previewService review_services.PreviewAnalyzer) *GitHubHandler {
	return &GitHubHandler{
		logger:         logger,
		previewService: previewService,
	}
}

// TreeNode represents a node in the file tree
type TreeNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	Type     string      `json:"type"` // "file" or "dir"
	Size     int         `json:"size,omitempty"`
	Children []*TreeNode `json:"children,omitempty"`
}

// TreeResponse represents the repository tree response
type TreeResponse struct {
	Owner       string      `json:"owner"`
	Repo        string      `json:"repo"`
	Branch      string      `json:"branch"`
	Tree        []*TreeNode `json:"tree"`
	EntryPoints []string    `json:"entry_points"`
	FileCount   int         `json:"file_count"`
}

// FileResponse represents a single file content response
type FileResponse struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Language string `json:"language"`
	Size     int    `json:"size"`
	SHA      string `json:"sha"`
}

// QuickScanResponse represents the quick repo scan response
type QuickScanResponse struct {
	Owner        string                 `json:"owner"`
	Repo         string                 `json:"repo"`
	Branch       string                 `json:"branch"`
	Files        []*FileResponse        `json:"files"`
	Analysis     map[string]interface{} `json:"analysis"`
	FetchedAt    time.Time              `json:"fetched_at"`
	FilesFetched int                    `json:"files_fetched"`
}

// GetRepoTree fetches the repository tree structure without file contents
func (h *GitHubHandler) GetRepoTree(c *gin.Context) {
	repoURL := c.Query("url")
	branch := c.Query("branch")

	if repoURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Repository URL is required"})
		return
	}

	// Parse owner and repo from URL (github.com/owner/repo)
	owner, repo, err := parseGitHubURL(repoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid GitHub URL: %v", err)})
		return
	}

	// Get GitHub token from session
	token, exists := c.Get("github_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitHub authentication required"})
		return
	}

	// Create GitHub client
	client := createGitHubClient(c.Request.Context(), token.(string))

	// If no branch specified, get default branch
	if branch == "" {
		repository, _, err := client.Repositories.Get(c.Request.Context(), owner, repo)
		if err != nil {
			h.logger.Error("Failed to get repository", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch repository"})
			return
		}
		branch = repository.GetDefaultBranch()
	}

	// Get repository tree
	tree, _, err := client.Git.GetTree(c.Request.Context(), owner, repo, branch, true)
	if err != nil {
		h.logger.Error("Failed to get repository tree", "error", err)
		handleGitHubError(c, err)
		return
	}

	// Build tree structure
	treeNodes := buildTreeStructure(tree.Entries)
	entryPoints := identifyEntryPoints(tree.Entries)

	response := &TreeResponse{
		Owner:       owner,
		Repo:        repo,
		Branch:      branch,
		Tree:        treeNodes,
		EntryPoints: entryPoints,
		FileCount:   len(tree.Entries),
	}

	c.JSON(http.StatusOK, response)
}

// GetRepoFile fetches a single file's content from the repository
func (h *GitHubHandler) GetRepoFile(c *gin.Context) {
	repoURL := c.Query("url")
	path := c.Query("path")
	branch := c.Query("branch")

	if repoURL == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Repository URL and file path are required"})
		return
	}

	// Normalize path: remove leading slash if present
	path = strings.TrimPrefix(path, "/")

	h.logger.Info("Fetching file from GitHub",
		"url", repoURL,
		"path", path,
		"branch", branch,
	)

	// Parse owner and repo from URL
	owner, repo, err := parseGitHubURL(repoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid GitHub URL: %v", err)})
		return
	}

	// Get GitHub token from session
	token, exists := c.Get("github_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitHub authentication required"})
		return
	}

	// Create GitHub client
	client := createGitHubClient(c.Request.Context(), token.(string))

	// Get file content
	opts := &github.RepositoryContentGetOptions{}
	if branch != "" {
		opts.Ref = branch
	}

	fileContent, _, _, err := client.Repositories.GetContents(c.Request.Context(), owner, repo, path, opts)
	if err != nil {
		h.logger.Error("Failed to get file content", "error", err, "path", path)
		handleGitHubError(c, err)
		return
	}

	if fileContent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Decode content
	content, err := fileContent.GetContent()
	if err != nil {
		h.logger.Error("Failed to decode file content", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode file content"})
		return
	}

	response := &FileResponse{
		Path:     path,
		Content:  content,
		Language: detectLanguageFromPath(path),
		Size:     fileContent.GetSize(),
		SHA:      fileContent.GetSHA(),
	}

	c.JSON(http.StatusOK, response)
}

// QuickRepoScan performs a quick repository scan by fetching core files
func (h *GitHubHandler) QuickRepoScan(c *gin.Context) {
	repoURL := c.Query("url")
	branch := c.Query("branch")

	if repoURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Repository URL is required"})
		return
	}

	// Parse owner and repo from URL
	owner, repo, err := parseGitHubURL(repoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid GitHub URL: %v", err)})
		return
	}

	// Get GitHub token from session
	token, exists := c.Get("github_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitHub authentication required"})
		return
	}

	// Create GitHub client
	client := createGitHubClient(c.Request.Context(), token.(string))

	// If no branch specified, get default branch
	if branch == "" {
		repository, _, err := client.Repositories.Get(c.Request.Context(), owner, repo)
		if err != nil {
			h.logger.Error("Failed to get repository", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch repository"})
			return
		}
		branch = repository.GetDefaultBranch()
	}

	// Get repository tree to find actual core files
	tree, _, err := client.Git.GetTree(c.Request.Context(), owner, repo, branch, true)
	if err != nil {
		h.logger.Error("Failed to get repository tree", "error", err)
		handleGitHubError(c, err)
		return
	}

	// Core file patterns to search for
	corePatterns := []string{
		"README.md", "README.rst", "README.txt", "README",
		"package.json", "package-lock.json",
		"go.mod", "go.sum",
		"requirements.txt", "Pipfile", "pyproject.toml",
		"Cargo.toml", "Cargo.lock",
		"pom.xml", "build.gradle", "build.gradle.kts",
		"LICENSE", "LICENSE.md", "LICENSE.txt",
		"CONTRIBUTING.md",
		".gitignore",
		"docker-compose.yml", "docker-compose.yaml",
		"Dockerfile",
		"Makefile",
	}

	// Entry point patterns
	entryPatterns := []string{
		"main.go", "app.go",
		"index.js", "app.js", "server.js",
		"main.py", "app.py", "__main__.py",
		"main.rs", "lib.rs",
		"Main.java", "Program.cs",
		"index.html",
	}

	// Find matching core files in tree
	var coreFilePaths []string
	for _, entry := range tree.Entries {
		if entry.GetType() != "blob" {
			continue
		}

		path := entry.GetPath()
		name := getFileName(path)

		// Check if matches core pattern
		for _, pattern := range corePatterns {
			if name == pattern || path == pattern {
				coreFilePaths = append(coreFilePaths, path)
				break
			}
		}

		// Check if matches entry point pattern
		for _, pattern := range entryPatterns {
			if name == pattern || strings.HasSuffix(path, "/"+pattern) {
				coreFilePaths = append(coreFilePaths, path)
				break
			}
		}
	}

	// Limit to first 10 files to avoid huge responses
	if len(coreFilePaths) > 10 {
		coreFilePaths = coreFilePaths[:10]
	}

	// Fetch file contents
	var files []*FileResponse
	opts := &github.RepositoryContentGetOptions{Ref: branch}

	for _, path := range coreFilePaths {
		fileContent, _, _, err := client.Repositories.GetContents(c.Request.Context(), owner, repo, path, opts)
		if err != nil {
			h.logger.Warn("Failed to fetch core file", "error", err, "path", path)
			continue
		}

		if fileContent == nil {
			continue
		}

		content, err := fileContent.GetContent()
		if err != nil {
			h.logger.Warn("Failed to decode file", "error", err, "path", path)
			continue
		}

		files = append(files, &FileResponse{
			Path:     path,
			Content:  content,
			Language: detectLanguageFromPath(path),
			Size:     fileContent.GetSize(),
			SHA:      fileContent.GetSHA(),
		})
	}

	if len(files) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No core files found in repository"})
		return
	}

	// Combine file contents for Preview Mode analysis
	var combinedContent strings.Builder
	combinedContent.WriteString(fmt.Sprintf("# Repository: %s/%s (branch: %s)\n\n", owner, repo, branch))

	for _, file := range files {
		combinedContent.WriteString(fmt.Sprintf("## File: %s\n", file.Path))
		combinedContent.WriteString(fmt.Sprintf("```%s\n", file.Language))
		combinedContent.WriteString(file.Content)
		combinedContent.WriteString("\n```\n\n")
	}

	// Call Preview Mode AI analysis with default modes (intermediate/quick)
	// Extract user_mode and output_mode from query params if provided
	userMode := c.DefaultQuery("user_mode", "intermediate")
	outputMode := c.DefaultQuery("output_mode", "quick")

	result, err := h.previewService.AnalyzePreview(c.Request.Context(), combinedContent.String(), userMode, outputMode)
	if err != nil {
		h.logger.Error("Preview analysis failed for quick scan", "error", err)
		// Fall back to basic response if AI analysis fails
		analysis := map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("AI analysis failed: %v", err),
		}
		response := &QuickScanResponse{
			Owner:        owner,
			Repo:         repo,
			Branch:       branch,
			Files:        files,
			Analysis:     analysis,
			FetchedAt:    time.Now(),
			FilesFetched: len(files),
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Convert PreviewModeOutput to map for response
	analysis := map[string]interface{}{
		"summary":            result.Summary,
		"file_tree":          result.FileTree,
		"bounded_contexts":   result.BoundedContexts,
		"tech_stack":         result.TechStack,
		"architecture_style": result.ArchitectureStyle,
		"entry_points":       result.EntryPoints,
		"external_deps":      result.ExternalDeps,
		"stats":              result.Stats,
	}

	response := &QuickScanResponse{
		Owner:        owner,
		Repo:         repo,
		Branch:       branch,
		Files:        files,
		Analysis:     analysis,
		FetchedAt:    time.Now(),
		FilesFetched: len(files),
	}

	h.logger.Info("Quick repo scan completed",
		"owner", owner,
		"repo", repo,
		"branch", branch,
		"files_fetched", len(files),
	)

	c.JSON(http.StatusOK, response)
}

// Helper functions

func parseGitHubURL(url string) (string, string, error) {
	// Remove protocol if present
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Remove github.com prefix
	url = strings.TrimPrefix(url, "github.com/")

	// Split into owner/repo
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format")
	}

	owner := parts[0]
	repo := strings.TrimSuffix(parts[1], ".git")

	return owner, repo, nil
}

func createGitHubClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func buildTreeStructure(entries []*github.TreeEntry) []*TreeNode {
	// Build a map of all entries
	nodeMap := make(map[string]*TreeNode)
	var rootNodes []*TreeNode

	// First pass: create all nodes
	for _, entry := range entries {
		// Convert GitHub type to frontend-friendly type
		nodeType := entry.GetType()
		if nodeType == "blob" {
			nodeType = "file"
		} else if nodeType == "tree" {
			nodeType = "directory"
		}

		node := &TreeNode{
			Name: getFileName(entry.GetPath()),
			Path: entry.GetPath(),
			Type: nodeType,
			Size: entry.GetSize(),
		}

		if entry.GetType() == "tree" {
			node.Children = []*TreeNode{}
		}

		nodeMap[entry.GetPath()] = node
	}

	// Second pass: build tree hierarchy
	for _, entry := range entries {
		path := entry.GetPath()
		node := nodeMap[path]

		// Find parent directory
		parentPath := getParentPath(path)
		if parentPath == "" {
			// Root level node
			rootNodes = append(rootNodes, node)
		} else {
			// Add to parent's children
			if parent, exists := nodeMap[parentPath]; exists {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return rootNodes
}

func getFileName(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func getParentPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "/")
}

func identifyEntryPoints(entries []*github.TreeEntry) []string {
	entryPoints := []string{}
	entryPointFiles := map[string]bool{
		"main.go":      true,
		"cmd/main.go":  true,
		"index.js":     true,
		"src/index.js": true,
		"app.js":       true,
		"src/app.js":   true,
		"main.py":      true,
		"app.py":       true,
		"__init__.py":  true,
		"main.rs":      true,
		"lib.rs":       true,
		"Main.java":    true,
		"Program.cs":   true,
		"index.html":   true,
	}

	for _, entry := range entries {
		if entry.GetType() == "blob" {
			path := entry.GetPath()
			if entryPointFiles[path] || entryPointFiles[getFileName(path)] {
				entryPoints = append(entryPoints, path)
			}
		}
	}

	return entryPoints
}

func detectLanguageFromPath(path string) string {
	ext := strings.ToLower(strings.TrimPrefix(strings.ToLower(getFileExtension(path)), "."))

	languageMap := map[string]string{
		"js":   "javascript",
		"jsx":  "javascript",
		"ts":   "typescript",
		"tsx":  "typescript",
		"py":   "python",
		"go":   "go",
		"rs":   "rust",
		"java": "java",
		"c":    "c",
		"cpp":  "cpp",
		"cs":   "csharp",
		"rb":   "ruby",
		"php":  "php",
		"html": "html",
		"css":  "css",
		"json": "json",
		"xml":  "xml",
		"yaml": "yaml",
		"yml":  "yaml",
		"md":   "markdown",
		"sh":   "bash",
		"sql":  "sql",
	}

	if lang, exists := languageMap[ext]; exists {
		return lang
	}

	return "plaintext"
}

func getFileExtension(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) > 1 {
		return "." + parts[len(parts)-1]
	}
	return ""
}

func handleGitHubError(c *gin.Context, err error) {
	if githubErr, ok := err.(*github.ErrorResponse); ok {
		switch githubErr.Response.StatusCode {
		case http.StatusNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository or resource not found"})
		case http.StatusForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": "Access forbidden. Repository may be private or rate limit exceeded."})
		case http.StatusUnauthorized:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "GitHub authentication failed"})
		case 429: // Rate limit
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "GitHub API rate limit exceeded. Please try again later."})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub API error"})
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to communicate with GitHub"})
	}
}
