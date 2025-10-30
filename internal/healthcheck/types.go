package healthcheck

import "time"

// CheckResult represents the result of a single health check
type CheckResult struct {
	Name      string                 `json:"name"`
	Status    CheckStatus            `json:"status"`
	Message   string                 `json:"message"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// CheckStatus represents the status of a health check
type CheckStatus string

const (
	StatusPass    CheckStatus = "pass"
	StatusWarn    CheckStatus = "warn"
	StatusFail    CheckStatus = "fail"
	StatusUnknown CheckStatus = "unknown"
)

// HealthReport represents the overall system health report
type HealthReport struct {
	Status     CheckStatus   `json:"status"`
	Timestamp  time.Time     `json:"timestamp"`
	Duration   time.Duration `json:"duration"`
	Checks     []CheckResult `json:"checks"`
	Summary    Summary       `json:"summary"`
	SystemInfo SystemInfo    `json:"system_info"`
}

// Summary provides aggregate statistics
type Summary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Warned  int `json:"warned"`
	Failed  int `json:"failed"`
	Unknown int `json:"unknown"`
}

// SystemInfo provides environment context
type SystemInfo struct {
	Environment string    `json:"environment"`
	Hostname    string    `json:"hostname"`
	GoVersion   string    `json:"go_version"`
	Timestamp   time.Time `json:"timestamp"`
}

// Checker is the interface for all health check implementations
type Checker interface {
	Name() string
	Check() CheckResult
}

