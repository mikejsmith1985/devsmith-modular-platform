package healthcheck

import (
	"strings"
	"testing"
	"time"
)

func TestFormatJSON(t *testing.T) {
	report := HealthReport{
		Status:    StatusPass,
		Timestamp: time.Now(),
		Duration:  100 * time.Millisecond,
		Checks: []CheckResult{
			{
				Name:      "test_check",
				Status:    StatusPass,
				Message:   "OK",
				Duration:  50 * time.Millisecond,
				Timestamp: time.Now(),
			},
		},
		Summary: Summary{
			Total:  1,
			Passed: 1,
		},
		SystemInfo: SystemInfo{
			Environment: "test",
			Hostname:    "localhost",
			GoVersion:   "go1.21",
			Timestamp:   time.Now(),
		},
	}

	output, err := FormatJSON(&report)
	if err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	if !strings.Contains(output, "test_check") {
		t.Error("Expected output to contain check name")
	}

	if !strings.Contains(output, "pass") {
		t.Error("Expected output to contain status")
	}
}

func TestFormatHuman(t *testing.T) {
	report := HealthReport{
		Status:    StatusPass,
		Timestamp: time.Now(),
		Duration:  100 * time.Millisecond,
		Checks: []CheckResult{
			{
				Name:      "test_check",
				Status:    StatusPass,
				Message:   "All good",
				Duration:  50 * time.Millisecond,
				Timestamp: time.Now(),
			},
		},
		Summary: Summary{
			Total:  1,
			Passed: 1,
		},
		SystemInfo: SystemInfo{
			Environment: "test",
			Hostname:    "localhost",
			GoVersion:   "go1.21",
			Timestamp:   time.Now(),
		},
	}

	output := FormatHuman(&report)

	expectedStrings := []string{
		"DevSmith Platform Health Check",
		"test_check",
		"All good",
		"Overall Status",
		"Summary:",
		"Total Checks:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

func TestGetStatusSymbol(t *testing.T) {
	tests := []struct {
		status   CheckStatus
		expected string
	}{
		{StatusPass, "✓"},
		{StatusWarn, "⚠"},
		{StatusFail, "✗"},
		{StatusUnknown, "?"},
	}

	for _, tt := range tests {
		result := getStatusSymbol(tt.status)
		if result != tt.expected {
			t.Errorf("Expected symbol %s for status %s, got %s", tt.expected, tt.status, result)
		}
	}
}
