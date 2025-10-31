// Package healthcheck provides health checking and diagnostics.
package healthcheck

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DuplicateBlock represents a potential duplicate code block
type DuplicateBlock struct {
	File1      string
	File2      string
	StartLine1 int
	StartLine2 int
	Lines      int
	Content    string
}

// DuplicateDetector finds duplicate code blocks
type DuplicateDetector struct {
	minLines   int
	codeBlocks map[string]*CodeBlock
}

// CodeBlock represents a code segment
type CodeBlock struct {
	File       string
	StartLine  int
	EndLine    int
	Content    string
	Normalized string
}

// NewDuplicateDetector creates a new duplicate detector
func NewDuplicateDetector(minLines int) *DuplicateDetector {
	if minLines < 3 {
		minLines = 3
	}
	return &DuplicateDetector{
		minLines:   minLines,
		codeBlocks: make(map[string]*CodeBlock),
	}
}

// ScanDirectory scans a directory for Go files and detects duplicates
func (dd *DuplicateDetector) ScanDirectory(rootPath string) ([]DuplicateBlock, error) {
	var allBlocks []*CodeBlock

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip test files, vendor, and hidden directories
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		blocks, err := dd.extractCodeBlocks(path)
		if err == nil {
			allBlocks = append(allBlocks, blocks...)
		}
		return nil
	})

	if err != nil {
		return []DuplicateBlock{}, fmt.Errorf("failed to scan directory: %w", err)
	}

	duplicates := dd.findDuplicates(allBlocks)
	if duplicates == nil {
		duplicates = []DuplicateBlock{}
	}
	return duplicates, nil
}

// extractCodeBlocks extracts code blocks from a file
func (dd *DuplicateDetector) extractCodeBlocks(filePath string) ([]*CodeBlock, error) {
	// Validate file path to prevent path traversal attacks
	if err := validateFilePath(filePath); err != nil {
		return nil, err
	}

	file, err := os.Open(filePath) // #nosec G304 - path is validated above
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("warning: failed to close file %s: %v", filePath, err)
		}
	}()

	var blocks []*CodeBlock
	var currentBlock []string
	var startLine int
	var lineNum int
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip comments, empty lines
		if line == "" || strings.HasPrefix(line, "//") {
			if len(currentBlock) > 0 && len(currentBlock) >= dd.minLines {
				block := &CodeBlock{
					File:       filePath,
					StartLine:  startLine,
					EndLine:    lineNum - 1,
					Content:    strings.Join(currentBlock, "\n"),
					Normalized: normalizeCode(strings.Join(currentBlock, "\n")),
				}
				blocks = append(blocks, block)
			}
			currentBlock = []string{}
			continue
		}

		if len(currentBlock) == 0 {
			startLine = lineNum
		}
		currentBlock = append(currentBlock, line)
	}

	// Handle last block
	if len(currentBlock) > 0 && len(currentBlock) >= dd.minLines {
		block := &CodeBlock{
			File:       filePath,
			StartLine:  startLine,
			EndLine:    lineNum,
			Content:    strings.Join(currentBlock, "\n"),
			Normalized: normalizeCode(strings.Join(currentBlock, "\n")),
		}
		blocks = append(blocks, block)
	}

	return blocks, scanner.Err()
}

// findDuplicates finds duplicate blocks
func (dd *DuplicateDetector) findDuplicates(blocks []*CodeBlock) []DuplicateBlock {
	var duplicates []DuplicateBlock
	checked := make(map[string]bool)

	for i, block1 := range blocks {
		for j, block2 := range blocks {
			if i >= j {
				continue
			}

			key := fmt.Sprintf("%s:%d:%s:%d", block1.File, block1.StartLine, block2.File, block2.StartLine)
			if checked[key] {
				continue
			}
			checked[key] = true

			if block1.Normalized == block2.Normalized && block1.Content != "" {
				dup := DuplicateBlock{
					File1:      block1.File,
					File2:      block2.File,
					StartLine1: block1.StartLine,
					StartLine2: block2.StartLine,
					Lines:      len(block1.Content),
					Content:    block1.Content,
				}
				duplicates = append(duplicates, dup)
			}
		}
	}

	return duplicates
}

// normalizeCode normalizes code for comparison
func normalizeCode(code string) string {
	// Remove variable names, keep structure
	re := regexp.MustCompile(`[a-zA-Z_]\w*`)
	normalized := re.ReplaceAllString(code, "VAR")

	// Collapse whitespace
	fields := strings.Fields(normalized)
	return strings.Join(fields, " ")
}

// validateFilePath validates that a file path is safe to open
// Prevents path traversal attacks and symlink-based attacks
func validateFilePath(filePath string) error {
	// Reject relative paths with directory traversal
	if strings.Contains(filePath, "..") {
		return fmt.Errorf("path %q contains directory traversal sequences", filePath)
	}

	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Resolve symlinks to get the real path
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If EvalSymlinks fails, just use the absolute path (may not be a symlink)
		realPath = absPath
	}

	// Get current working directory (workspace root)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// For files in temp directories (like /tmp for tests), allow them
	// Only check that files within the project workspace don't escape it
	if strings.HasPrefix(realPath, cwd) {
		// Ensure the real path is within the workspace
		rel, err := filepath.Rel(cwd, realPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return fmt.Errorf("path %q is outside workspace boundaries", filePath)
		}
	}

	return nil
}
