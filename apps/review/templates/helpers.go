package templates

import reviewmodels "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"

// filterBySeverity filters issues by severity level
func filterBySeverity(issues []reviewmodels.CodeIssue, severity string) []reviewmodels.CodeIssue {
	var filtered []reviewmodels.CodeIssue
	for _, issue := range issues {
		if issue.Severity == severity {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}
