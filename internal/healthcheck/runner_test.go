package healthcheck

import (
	"testing"
	"time"
)

// MockChecker is a test checker
type MockChecker struct {
	name   string
	status CheckStatus
	err    string
}

func (m *MockChecker) Name() string {
	return m.name
}

func (m *MockChecker) Check() CheckResult {
	return CheckResult{
		Name:      m.name,
		Status:    m.status,
		Message:   "mock check",
		Error:     m.err,
		Duration:  10 * time.Millisecond,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}
}

func TestRunnerWithNoCheckers(t *testing.T) {
	runner := NewRunner()
	report := runner.Run()

	if report.Summary.Total != 0 {
		t.Errorf("Expected 0 checks, got %d", report.Summary.Total)
	}

	if report.Status != StatusPass {
		t.Errorf("Expected status %s, got %s", StatusPass, report.Status)
	}
}

func TestRunnerWithPassingChecks(t *testing.T) {
	runner := NewRunner()
	runner.AddChecker(&MockChecker{name: "test1", status: StatusPass})
	runner.AddChecker(&MockChecker{name: "test2", status: StatusPass})

	report := runner.Run()

	if report.Summary.Total != 2 {
		t.Errorf("Expected 2 checks, got %d", report.Summary.Total)
	}

	if report.Summary.Passed != 2 {
		t.Errorf("Expected 2 passed, got %d", report.Summary.Passed)
	}

	if report.Status != StatusPass {
		t.Errorf("Expected status %s, got %s", StatusPass, report.Status)
	}
}

func TestRunnerWithFailedCheck(t *testing.T) {
	runner := NewRunner()
	runner.AddChecker(&MockChecker{name: "test1", status: StatusPass})
	runner.AddChecker(&MockChecker{name: "test2", status: StatusFail, err: "mock error"})

	report := runner.Run()

	if report.Summary.Total != 2 {
		t.Errorf("Expected 2 checks, got %d", report.Summary.Total)
	}

	if report.Summary.Passed != 1 {
		t.Errorf("Expected 1 passed, got %d", report.Summary.Passed)
	}

	if report.Summary.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", report.Summary.Failed)
	}

	if report.Status != StatusFail {
		t.Errorf("Expected status %s, got %s", StatusFail, report.Status)
	}
}

func TestRunnerWithWarning(t *testing.T) {
	runner := NewRunner()
	runner.AddChecker(&MockChecker{name: "test1", status: StatusPass})
	runner.AddChecker(&MockChecker{name: "test2", status: StatusWarn})

	report := runner.Run()

	if report.Summary.Total != 2 {
		t.Errorf("Expected 2 checks, got %d", report.Summary.Total)
	}

	if report.Summary.Passed != 1 {
		t.Errorf("Expected 1 passed, got %d", report.Summary.Passed)
	}

	if report.Summary.Warned != 1 {
		t.Errorf("Expected 1 warned, got %d", report.Summary.Warned)
	}

	if report.Status != StatusWarn {
		t.Errorf("Expected status %s, got %s", StatusWarn, report.Status)
	}
}

func TestDetermineOverallStatus(t *testing.T) {
	tests := []struct {
		name     string
		checks   []CheckResult
		expected CheckStatus
	}{
		{
			name:     "all passing",
			checks:   []CheckResult{{Status: StatusPass}, {Status: StatusPass}},
			expected: StatusPass,
		},
		{
			name:     "one warning",
			checks:   []CheckResult{{Status: StatusPass}, {Status: StatusWarn}},
			expected: StatusWarn,
		},
		{
			name:     "one failure",
			checks:   []CheckResult{{Status: StatusPass}, {Status: StatusFail}},
			expected: StatusFail,
		},
		{
			name:     "warning and failure",
			checks:   []CheckResult{{Status: StatusWarn}, {Status: StatusFail}},
			expected: StatusFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineOverallStatus(tt.checks)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCalculateSummary(t *testing.T) {
	checks := []CheckResult{
		{Status: StatusPass},
		{Status: StatusPass},
		{Status: StatusWarn},
		{Status: StatusFail},
		{Status: StatusUnknown},
	}

	summary := calculateSummary(checks)

	if summary.Total != 5 {
		t.Errorf("Expected total 5, got %d", summary.Total)
	}
	if summary.Passed != 2 {
		t.Errorf("Expected 2 passed, got %d", summary.Passed)
	}
	if summary.Warned != 1 {
		t.Errorf("Expected 1 warned, got %d", summary.Warned)
	}
	if summary.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", summary.Failed)
	}
	if summary.Unknown != 1 {
		t.Errorf("Expected 1 unknown, got %d", summary.Unknown)
	}
}
