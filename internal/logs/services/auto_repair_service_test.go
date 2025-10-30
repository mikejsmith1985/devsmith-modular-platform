package services

import (
	"testing"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/healthcheck"
)

func TestDetermineRepairStrategy(t *testing.T) {
	tests := []struct {
		name       string
		issueType  string
		policy     *HealthPolicy
		expected   string
		reason     string
	}{
		{
			name:      "Timeout issues are restarted",
			issueType: "timeout",
			policy: &HealthPolicy{
				RepairStrategy: "rebuild", // Policy says rebuild, but timeout suggests restart
			},
			expected: "restart",
			reason:   "Timeout usually means hung service, restart first before rebuilding",
		},
		{
			name:      "Crash issues are rebuilt",
			issueType: "crash",
			policy: &HealthPolicy{
				RepairStrategy: "restart",
			},
			expected: "rebuild",
			reason:   "Container crashes may need fresh image",
		},
		{
			name:      "Dependency issues aren't repaired",
			issueType: "dependency",
			policy: &HealthPolicy{
				RepairStrategy: "restart",
			},
			expected: "none",
			reason:   "Can't fix dependency by restarting this service",
		},
		{
			name:      "Security issues are rebuilt",
			issueType: "security",
			policy: &HealthPolicy{
				RepairStrategy: "restart",
			},
			expected: "rebuild",
			reason:   "Vulnerabilities need fresh images with patches",
		},
		{
			name:      "Unknown issues use policy default",
			issueType: "unknown",
			policy: &HealthPolicy{
				RepairStrategy: "none",
			},
			expected: "none",
			reason:   "No specific strategy, use policy default",
		},
	}

	service := &AutoRepairService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.determineRepairStrategy(tt.issueType, tt.policy)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s. Reason: %s", tt.expected, result, tt.reason)
			}
		})
	}
}

func TestClassifyIssue(t *testing.T) {
	tests := []struct {
		name       string
		checkResult healthcheck.CheckResult
		expected   string
	}{
		{
			name: "Timeout classification",
			checkResult: healthcheck.CheckResult{
				Message: "Endpoint timeout after 5 seconds",
			},
			expected: "timeout",
		},
		{
			name: "Refused connection classification",
			checkResult: healthcheck.CheckResult{
				Message: "Connection refused - service not responding",
			},
			expected: "timeout",
		},
		{
			name: "Crash classification",
			checkResult: healthcheck.CheckResult{
				Message: "Container stopped/crash detected",
			},
			expected: "crash",
		},
		{
			name: "Dependency classification",
			checkResult: healthcheck.CheckResult{
				Message: "Dependent service 'logs' is not responding",
			},
			expected: "dependency",
		},
		{
			name: "Security classification",
			checkResult: healthcheck.CheckResult{
				Message: "CRITICAL vulnerabilities found: 5 critical, 2 high",
			},
			expected: "security",
		},
		{
			name: "Unknown classification",
			checkResult: healthcheck.CheckResult{
				Message: "Something went wrong",
			},
			expected: "unknown",
		},
	}

	service := &AutoRepairService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.classifyIssue(tt.checkResult)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestExtractServiceName(t *testing.T) {
	tests := []struct {
		checkName string
		expected  string
	}{
		{"http_portal", "portal"},
		{"http_review", "review"},
		{"gateway_routing", "routing"},
		{"database", ""},
		{"", ""},
		{"single", ""},
	}

	service := &AutoRepairService{}

	for _, tt := range tests {
		t.Run(tt.checkName, func(t *testing.T) {
			result := service.extractServiceName(tt.checkName)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestAnalyzeAndRepair(t *testing.T) {
	// This test requires a database, so we skip it in unit test mode
	// It would be tested in integration tests
	t.Skip("Integration test - requires database connection")
}
