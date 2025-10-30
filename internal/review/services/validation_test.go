package review_services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateCodeContent validates code content input
func TestValidateCodeContent_EmptyCode(t *testing.T) {
	err := ValidateCodeContent("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestValidateCodeContent_TooLarge(t *testing.T) {
	// Create code exceeding 10MB limit
	largeCode := strings.Repeat("x", 11*1024*1024)
	err := ValidateCodeContent(largeCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum")
}

func TestValidateCodeContent_ValidCode(t *testing.T) {
	validCode := "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}"
	err := ValidateCodeContent(validCode)
	assert.NoError(t, err)
}

func TestValidateCodeContent_LargeButValid(t *testing.T) {
	// Create code just under 10MB limit
	largeCode := strings.Repeat("x", 9*1024*1024)
	err := ValidateCodeContent(largeCode)
	assert.NoError(t, err)
}

// TestValidateScanQuery validates search query input
func TestValidateScanQuery_EmptyQuery(t *testing.T) {
	err := ValidateScanQuery("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestValidateScanQuery_TooShort(t *testing.T) {
	err := ValidateScanQuery("a")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least")
}

func TestValidateScanQuery_TooLong(t *testing.T) {
	longQuery := strings.Repeat("word ", 500)
	err := ValidateScanQuery(longQuery)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "maximum length")
}

func TestValidateScanQuery_Valid(t *testing.T) {
	validQuery := "Where is authentication handled?"
	err := ValidateScanQuery(validQuery)
	assert.NoError(t, err)
}

func TestValidateScanQuery_WithSpecialChars(t *testing.T) {
	validQuery := "Find all database queries with 'SELECT'"
	err := ValidateScanQuery(validQuery)
	assert.NoError(t, err)
}

// TestValidateReadingMode validates reading mode enum
func TestValidateReadingMode_Empty(t *testing.T) {
	err := ValidateReadingMode("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid reading mode")
}

func TestValidateReadingMode_Invalid(t *testing.T) {
	err := ValidateReadingMode("invalid_mode")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid reading mode")
}

func TestValidateReadingMode_Preview(t *testing.T) {
	err := ValidateReadingMode("preview")
	assert.NoError(t, err)
}

func TestValidateReadingMode_Skim(t *testing.T) {
	err := ValidateReadingMode("skim")
	assert.NoError(t, err)
}

func TestValidateReadingMode_Scan(t *testing.T) {
	err := ValidateReadingMode("scan")
	assert.NoError(t, err)
}

func TestValidateReadingMode_Detailed(t *testing.T) {
	err := ValidateReadingMode("detailed")
	assert.NoError(t, err)
}

func TestValidateReadingMode_Critical(t *testing.T) {
	err := ValidateReadingMode("critical")
	assert.NoError(t, err)
}

// TestValidateGitHubURL validates GitHub repository URLs
func TestValidateGitHubURL_Empty(t *testing.T) {
	err := ValidateGitHubURL("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestValidateGitHubURL_Invalid(t *testing.T) {
	testCases := []string{
		"not-a-url",
		"http://example.com",
		"https://gitlab.com/user/repo",
		"https://github.com/user", // Missing repo
		"ftp://github.com/user/repo",
	}
	for _, url := range testCases {
		err := ValidateGitHubURL(url)
		require.Error(t, err, "should reject invalid URL: %s", url)
	}
}

func TestValidateGitHubURL_Valid(t *testing.T) {
	validURLs := []string{
		"https://github.com/mikejsmith1985/devsmith-modular-platform",
		"https://github.com/user/repo",
		"github.com/user/repo",
	}
	for _, url := range validURLs {
		err := ValidateGitHubURL(url)
		assert.NoError(t, err, "should accept valid URL: %s", url)
	}
}

// TestValidateFilePath validates file paths for traversal attacks
func TestValidateFilePath_Empty(t *testing.T) {
	err := ValidateFilePath("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestValidateFilePath_PathTraversal(t *testing.T) {
	traversalPaths := []string{
		"../../../etc/passwd",
		"..\\..\\windows\\system32",
		"/etc/passwd",
		"C:\\Windows\\System32",
		"../../config.yml",
	}
	for _, path := range traversalPaths {
		err := ValidateFilePath(path)
		require.Error(t, err, "should reject path traversal: %s", path)
	}
}

func TestValidateFilePath_Valid(t *testing.T) {
	validPaths := []string{
		"main.go",
		"src/handler.go",
		"internal/services/auth.go",
		"cmd/review/handlers/review_handler.go",
	}
	for _, path := range validPaths {
		err := ValidateFilePath(path)
		assert.NoError(t, err, "should accept valid path: %s", path)
	}
}

// TestSanitizeCodeForDisplay validates XSS sanitization
func TestSanitizeCodeForDisplay_ValidCode(t *testing.T) {
	input := `func main() {
    fmt.Println("hello")
}`
	output := SanitizeCodeForDisplay(input)
	// Quotes will be escaped to &#34; for security
	assert.Contains(t, output, "fmt.Println")
	assert.Contains(t, output, "hello")
	// Verify it's escaped (quotes become &#34;)
	assert.Contains(t, output, "&#34;")
}

func TestSanitizeCodeForDisplay_HTMLTags(t *testing.T) {
	input := `<script>alert('xss')</script>`
	output := SanitizeCodeForDisplay(input)
	// Should be escaped
	assert.Contains(t, output, "&lt;script&gt;")
	assert.Contains(t, output, "&lt;/script&gt;")
	// Dangerous content is now escaped and safe to display
	assert.NotContains(t, output, "<script>")
}

func TestSanitizeCodeForDisplay_OnEventHandlers(t *testing.T) {
	input := `<img src="x" onerror="alert('xss')">`
	output := SanitizeCodeForDisplay(input)
	// Should be escaped
	assert.Contains(t, output, "&lt;img")
	assert.Contains(t, output, "&gt;")
	assert.NotContains(t, output, "<img")
	assert.NotContains(t, output, ">")
}

func TestSanitizeCodeForDisplay_HTMLEntities(t *testing.T) {
	input := `<>&"`
	output := SanitizeCodeForDisplay(input)
	// Should escape the HTML
	assert.Contains(t, output, "&lt;")
	assert.Contains(t, output, "&gt;")
	assert.Contains(t, output, "&amp;")
	assert.Contains(t, output, "&#34;")
}

func TestSanitizeCodeForDisplay_ValidHTMLInCode(t *testing.T) {
	// Code that happens to contain HTML-like content (e.g., template strings)
	input := `const html = "<div class='test'>content</div>";`
	output := SanitizeCodeForDisplay(input)
	// Should escape the HTML
	assert.Contains(t, output, "&lt;div")
	assert.Contains(t, output, "&lt;/div&gt;")
	assert.NotContains(t, output, "<div")
}

// TestValidateTitle validates review session title
func TestValidateTitle_Empty(t *testing.T) {
	err := ValidateTitle("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestValidateTitle_TooLong(t *testing.T) {
	longTitle := strings.Repeat("x", 300)
	err := ValidateTitle(longTitle)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "maximum length")
}

func TestValidateTitle_Valid(t *testing.T) {
	err := ValidateTitle("Review of user authentication service")
	assert.NoError(t, err)
}

func TestValidateTitle_WithSpecialChars(t *testing.T) {
	err := ValidateTitle("Review: User Auth & API Security")
	assert.NoError(t, err)
}

// TestValidateCodeSource validates code source type
func TestValidateCodeSource_Empty(t *testing.T) {
	err := ValidateCodeSource("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid code source")
}

func TestValidateCodeSource_Invalid(t *testing.T) {
	err := ValidateCodeSource("invalid_source")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid code source")
}

func TestValidateCodeSource_Paste(t *testing.T) {
	err := ValidateCodeSource("paste")
	assert.NoError(t, err)
}

func TestValidateCodeSource_GitHub(t *testing.T) {
	err := ValidateCodeSource("github")
	assert.NoError(t, err)
}

func TestValidateCodeSource_Upload(t *testing.T) {
	err := ValidateCodeSource("upload")
	assert.NoError(t, err)
}

// TestValidateGitHubBranch validates GitHub branch names
func TestValidateGitHubBranch_Empty(t *testing.T) {
	err := ValidateGitHubBranch("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestValidateGitHubBranch_Invalid(t *testing.T) {
	invalidBranches := []string{
		"feature/../../config", // Path traversal
		"../main",
		"..\\main",
	}
	for _, branch := range invalidBranches {
		err := ValidateGitHubBranch(branch)
		require.Error(t, err, "should reject invalid branch: %s", branch)
	}
}

func TestValidateGitHubBranch_Valid(t *testing.T) {
	validBranches := []string{
		"main",
		"develop",
		"feature/user-auth",
		"release/v1.0.0",
		"bugfix/issue-123",
	}
	for _, branch := range validBranches {
		err := ValidateGitHubBranch(branch)
		assert.NoError(t, err, "should accept valid branch: %s", branch)
	}
}
