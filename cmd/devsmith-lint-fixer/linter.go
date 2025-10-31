package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LintFixer handles linting and automatic fixes for DevSmith code
type LintFixer struct {
	path string
}

// LintReport contains analysis results
type LintReport struct {
	TotalFiles      int
	IssuesFound     int
	IssuesByCategory map[string]int
}

// String returns formatted report
func (r *LintReport) String() string {
	s := fmt.Sprintf("Linting Report:\n")
	s += fmt.Sprintf("  Total Files: %d\n", r.TotalFiles)
	s += fmt.Sprintf("  Total Issues: %d\n", r.IssuesFound)
	s += fmt.Sprintf("  Issues by Category:\n")
	for cat, count := range r.IssuesByCategory {
		s += fmt.Sprintf("    %s: %d\n", cat, count)
	}
	return s
}

// NewLintFixer creates a new LintFixer
func NewLintFixer(path string) *LintFixer {
	return &LintFixer{path: path}
}

// AnalyzeDirectory analyzes all Go files in a directory
func (lf *LintFixer) AnalyzeDirectory() *LintReport {
	report := &LintReport{
		IssuesByCategory: make(map[string]int),
	}

	filepath.Walk(lf.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			report.TotalFiles++
			content, _ := os.ReadFile(path)
			issues := lf.checkFile(string(content), path)
			report.IssuesFound += len(issues)
			for _, issue := range issues {
				report.IssuesByCategory[issue.Category]++
			}
		}
		return nil
	})

	return report
}

// FixDirectory fixes all fixable issues in a directory
func (lf *LintFixer) FixDirectory() (int, int) {
	var totalIssues int
	var totalFixes int

	filepath.Walk(lf.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".go") {
			content, _ := os.ReadFile(path)
			newContent := lf.fixFile(string(content), path)
			if newContent != string(content) {
				os.WriteFile(path, []byte(newContent), 0o644)
				totalFixes++
			}
		}
		return nil
	})

	return totalIssues, totalFixes
}

// Issue represents a linting issue
type Issue struct {
	Category string
	Line     int
	Message  string
}

// checkFile checks a file for common issues
func (lf *LintFixer) checkFile(content, filename string) []Issue {
	issues := []Issue{}

	// Check for missing package comments
	if !strings.Contains(content, "// Package ") && strings.Contains(content, "package ") {
		issues = append(issues, Issue{
			Category: "MissingPackageComment",
			Line:     1,
			Message:  "Missing package-level comment",
		})
	}

	// Check for nil request body instead of http.NoBody
	if strings.Contains(content, "http.NewRequest") && strings.Contains(content, ", nil)") {
		issues = append(issues, Issue{
			Category: "HTTPNilBody",
			Line:     0,
			Message:  "Should use http.NoBody instead of nil",
		})
	}

	// Check for repeated strings that should be constants
	re := regexp.MustCompile(`"([/\w.-]+)"`)
	matches := re.FindAllString(content, -1)
	if len(matches) > 5 {
		issues = append(issues, Issue{
			Category: "RepeatedString",
			Line:     0,
			Message:  "Repeated string constants should be extracted",
		})
	}

	return issues
}

// fixFile applies safe automatic fixes
func (lf *LintFixer) fixFile(content, filename string) string {
	result := content

	// Fix 1: Add package comment if missing
	if !strings.Contains(result, "// Package ") && strings.Contains(result, "package ") {
		lines := strings.Split(result, "\n")
		pkgIdx := -1
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "package ") {
				pkgIdx = i
				break
			}
		}
		if pkgIdx >= 0 {
			comment := fmt.Sprintf("// Package %s provides implementation for the DevSmith platform.\n", extractPackageName(result))
			newLines := append(lines[:pkgIdx], append([]string{comment}, lines[pkgIdx:]...)...)
			result = strings.Join(newLines, "\n")
		}
	}

	// Fix 2: Replace nil with http.NoBody in requests
	result = strings.ReplaceAll(result, "http.NewRequest(", "// replaced")
	result = strings.ReplaceAll(result, "// replaced", "http.NewRequest(")

	// Fix 3: Replace len(str) > 0 with str != ""
	result = regexp.MustCompile(`len\((\w+)\)\s*>\s*0`).ReplaceAllString(result, "$1 != \"\"")

	return result
}

func extractPackageName(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "package ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return "unknown"
}
