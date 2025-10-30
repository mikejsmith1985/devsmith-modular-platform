package healthcheck

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatJSON formats a health report as JSON
func FormatJSON(report HealthReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatHuman formats a health report for human readability
func FormatHuman(report HealthReport) string {
	var sb strings.Builder

	// Header
	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	sb.WriteString("  DevSmith Platform Health Check\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	// System Info
	sb.WriteString(fmt.Sprintf("Environment: %s\n", report.SystemInfo.Environment))
	sb.WriteString(fmt.Sprintf("Hostname:    %s\n", report.SystemInfo.Hostname))
	sb.WriteString(fmt.Sprintf("Go Version:  %s\n", report.SystemInfo.GoVersion))
	sb.WriteString(fmt.Sprintf("Timestamp:   %s\n\n", report.Timestamp.Format("2006-01-02 15:04:05")))

	// Overall Status
	statusSymbol := getStatusSymbol(report.Status)
	sb.WriteString(fmt.Sprintf("Overall Status: %s %s\n\n", statusSymbol, report.Status))

	// Summary
	sb.WriteString("Summary:\n")
	sb.WriteString(fmt.Sprintf("  Total Checks:  %d\n", report.Summary.Total))
	sb.WriteString(fmt.Sprintf("  ✓ Passed:      %d\n", report.Summary.Passed))
	if report.Summary.Warned > 0 {
		sb.WriteString(fmt.Sprintf("  ⚠ Warnings:    %d\n", report.Summary.Warned))
	}
	if report.Summary.Failed > 0 {
		sb.WriteString(fmt.Sprintf("  ✗ Failed:      %d\n", report.Summary.Failed))
	}
	sb.WriteString(fmt.Sprintf("  Duration:      %v\n\n", report.Duration))

	// Individual Checks
	sb.WriteString("Detailed Results:\n")
	sb.WriteString("───────────────────────────────────────────────────────────────\n")

	for _, check := range report.Checks {
		symbol := getStatusSymbol(check.Status)
		sb.WriteString(fmt.Sprintf("\n%s %s\n", symbol, check.Name))
		sb.WriteString(fmt.Sprintf("  Status:   %s\n", check.Status))
		sb.WriteString(fmt.Sprintf("  Message:  %s\n", check.Message))
		sb.WriteString(fmt.Sprintf("  Duration: %v\n", check.Duration))

		if check.Error != "" {
			sb.WriteString(fmt.Sprintf("  Error:    %s\n", check.Error))
		}

		if len(check.Details) > 0 {
			sb.WriteString("  Details:\n")
			for key, value := range check.Details {
				sb.WriteString(fmt.Sprintf("    %s: %v\n", key, value))
			}
		}
	}

	sb.WriteString("\n═══════════════════════════════════════════════════════════════\n")

	return sb.String()
}

// getStatusSymbol returns an emoji/symbol for the status
func getStatusSymbol(status CheckStatus) string {
	switch status {
	case StatusPass:
		return "✓"
	case StatusWarn:
		return "⚠"
	case StatusFail:
		return "✗"
	default:
		return "?"
	}
}

