package healthcheck

import "time"

// CheckResult represents the result of a single health check
type CheckResult struct {
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Name      string                 `json:"name"`
	Status    CheckStatus            `json:"status"`
	Message   string                 `json:"message"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
}

// CheckStatus represents the status of a health check
type CheckStatus string

// CheckStatus constants represent the possible health check outcomes.
const (
	StatusPass    CheckStatus = "pass"
	StatusWarn    CheckStatus = "warn"
	StatusFail    CheckStatus = "fail"
	StatusUnknown CheckStatus = "unknown"
)

// HealthReport represents the overall system health report
type HealthReport struct {
	SystemInfo SystemInfo    `json:"system_info"`
	Timestamp  time.Time     `json:"timestamp"`
	Status     CheckStatus   `json:"status"`
	Checks     []CheckResult `json:"checks"`
	Summary    Summary       `json:"summary"`
	Duration   time.Duration `json:"duration"`
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
	Timestamp   time.Time `json:"timestamp"`
	Environment string    `json:"environment"`
	Hostname    string    `json:"hostname"`
	GoVersion   string    `json:"go_version"`
}

// Checker is the interface for all health check implementations
type Checker interface {
	Name() string
	Check() CheckResult
}
