package healthcheck

import (
	"os"
	"runtime"
	"time"
)

// Runner executes health checks and generates reports
type Runner struct {
	Checkers []Checker
}

// NewRunner creates a new health check runner
func NewRunner() *Runner {
	return &Runner{
		Checkers: []Checker{},
	}
}

// AddChecker adds a health checker to the runner
func (r *Runner) AddChecker(checker Checker) {
	r.Checkers = append(r.Checkers, checker)
}

// Run executes all health checks and returns a report
func (r *Runner) Run() HealthReport {
	start := time.Now()

	report := HealthReport{
		Timestamp:  start,
		Checks:     make([]CheckResult, 0, len(r.Checkers)),
		SystemInfo: getSystemInfo(),
	}

	// Run all checks
	for _, checker := range r.Checkers {
		result := checker.Check()
		report.Checks = append(report.Checks, result)
	}

	// Calculate summary
	report.Summary = calculateSummary(report.Checks)

	// Determine overall status
	report.Status = determineOverallStatus(report.Checks)

	report.Duration = time.Since(start)

	return report
}

// calculateSummary generates aggregate statistics
func calculateSummary(checks []CheckResult) Summary {
	summary := Summary{
		Total: len(checks),
	}

	for _, check := range checks {
		switch check.Status {
		case StatusPass:
			summary.Passed++
		case StatusWarn:
			summary.Warned++
		case StatusFail:
			summary.Failed++
		case StatusUnknown:
			summary.Unknown++
		}
	}

	return summary
}

// determineOverallStatus calculates the overall system status
func determineOverallStatus(checks []CheckResult) CheckStatus {
	hasFailed := false
	hasWarned := false

	for _, check := range checks {
		if check.Status == StatusFail {
			hasFailed = true
		}
		if check.Status == StatusWarn {
			hasWarned = true
		}
	}

	if hasFailed {
		return StatusFail
	}
	if hasWarned {
		return StatusWarn
	}
	return StatusPass
}

// getSystemInfo gathers system information
func getSystemInfo() SystemInfo {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "unknown"
	}

	return SystemInfo{
		Environment: environment,
		Hostname:    hostname,
		GoVersion:   runtime.Version(),
		Timestamp:   time.Now(),
	}
}
