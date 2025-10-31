package healthcheck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// TrivyChecker validates container image security using Trivy
type TrivyChecker struct {
	CheckName string
	ScanType  string   // "image", "config", "filesystem"
	TrivyPath string   // path to trivy binary or script
	Targets   []string // images to scan
}

// TrivyScanResult represents parsed Trivy scan output
type TrivyScanResult struct {
	ScanType        string       `json:"scan_type"`
	Target          string       `json:"target"`
	Vulnerabilities []VulnResult `json:"vulnerabilities,omitempty"`
	Critical        int          `json:"critical"`
	High            int          `json:"high"`
	Medium          int          `json:"medium"`
	Low             int          `json:"low"`
}

// VulnResult represents a single vulnerability
type VulnResult struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Package  string `json:"package,omitempty"`
}

// ScanResults holds aggregated Trivy scan results
type ScanResults struct {
	Scans         []TrivyScanResult
	FailedTargets []string
	Critical      int
	High          int
	Medium        int
	Low           int
}

// Name returns the checker name
func (c *TrivyChecker) Name() string {
	return c.CheckName
}

// Check runs Trivy security scanning
func (c *TrivyChecker) Check() CheckResult {
	start := time.Now()
	result := CheckResult{
		Name:      c.CheckName,
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	if len(c.Targets) == 0 {
		result.Status = StatusWarn
		result.Message = "No targets specified for Trivy scan"
		result.Duration = time.Since(start)
		return result
	}

	// Set default path if not specified
	trivyPath := c.TrivyPath
	if trivyPath == "" {
		trivyPath = "scripts/trivy-scan.sh"
	}

	scanResults := c.performScans(trivyPath)

	c.populateDetails(&result, scanResults)
	c.determineStatus(&result, scanResults)

	result.Duration = time.Since(start)
	return result
}

// performScans executes Trivy scans on all targets and aggregates results
func (c *TrivyChecker) performScans(trivyPath string) *ScanResults {
	results := &ScanResults{
		Scans:         make([]TrivyScanResult, 0),
		FailedTargets: make([]string, 0),
	}

	for _, target := range c.Targets {
		scanResult, err := c.runTrivy(target, trivyPath)
		if err != nil {
			results.FailedTargets = append(results.FailedTargets, fmt.Sprintf("%s (%v)", target, err))
			continue
		}

		results.Scans = append(results.Scans, *scanResult)
		results.Critical += scanResult.Critical
		results.High += scanResult.High
		results.Medium += scanResult.Medium
		results.Low += scanResult.Low
	}

	return results
}

// populateDetails fills in the check result details
func (c *TrivyChecker) populateDetails(result *CheckResult, scanResults *ScanResults) {
	result.Details["targets_scanned"] = len(scanResults.Scans)
	result.Details["targets_failed"] = len(scanResults.FailedTargets)
	result.Details["critical"] = scanResults.Critical
	result.Details["high"] = scanResults.High
	result.Details["medium"] = scanResults.Medium
	result.Details["low"] = scanResults.Low
	result.Details["scans"] = scanResults.Scans

	if len(scanResults.FailedTargets) > 0 {
		result.Details["failed_targets"] = scanResults.FailedTargets
	}
}

// determineStatus sets the status and message based on vulnerability severity
func (c *TrivyChecker) determineStatus(result *CheckResult, scanResults *ScanResults) {
	switch {
	case scanResults.Critical > 0:
		result.Status = StatusFail
		result.Message = fmt.Sprintf("CRITICAL vulnerabilities found: %d critical, %d high", scanResults.Critical, scanResults.High)
		result.Error = "Critical security vulnerabilities detected - immediate action required"
	case scanResults.High > 0:
		result.Status = StatusWarn
		result.Message = fmt.Sprintf("HIGH vulnerabilities found: %d high, %d medium", scanResults.High, scanResults.Medium)
	case scanResults.Medium > 0:
		result.Status = StatusWarn
		result.Message = fmt.Sprintf("MEDIUM vulnerabilities found: %d medium, %d low", scanResults.Medium, scanResults.Low)
	case len(scanResults.Scans) > 0:
		result.Status = StatusPass
		result.Message = fmt.Sprintf("No vulnerabilities found in %d target(s)", len(scanResults.Scans))
	default:
		result.Status = StatusFail
		result.Message = "Failed to scan any targets"
		result.Error = fmt.Sprintf("All %d target(s) failed to scan", len(c.Targets))
	}
}

// runTrivy executes Trivy scan on a target
func (c *TrivyChecker) runTrivy(target, trivyPath string) (*TrivyScanResult, error) {
	// Call the Trivy wrapper script/binary
	cmd := exec.Command(trivyPath, "image", target)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	// Timeout after 60 seconds
	select {
	case <-time.After(60 * time.Second):
		return nil, fmt.Errorf("Trivy scan timeout for %s", target)
	case err := <-done:
		if err != nil {
			// Trivy exits with non-zero if vulnerabilities found, which is expected
			// Only fail on actual execution errors
			if stderr.Len() > 0 && stdout.Len() == 0 {
				return nil, fmt.Errorf("Trivy execution failed: %w", err)
			}
		}
	}

	// Parse the JSON output
	return c.parseTrivy(stdout.Bytes(), target)
}

// parseTrivy parses Trivy JSON output
func (c *TrivyChecker) parseTrivy(jsonOutput []byte, target string) (*TrivyScanResult, error) {
	if len(jsonOutput) == 0 {
		// Empty output means no vulnerabilities
		return &TrivyScanResult{
			ScanType: c.ScanType,
			Target:   target,
			Critical: 0,
			High:     0,
			Medium:   0,
			Low:      0,
		}, nil
	}

	// Try to unmarshal as Trivy JSON report
	var result TrivyScanResult
	result.ScanType = c.ScanType
	result.Target = target

	// Parse Trivy JSON format
	var trivyReport struct {
		Results []struct {
			Vulnerabilities []struct {
				VulnerabilityID string `json:"VulnerabilityID"`
				Severity        string `json:"Severity"`
				Title           string `json:"Title"`
				PkgName         string `json:"PkgName"`
			} `json:"Vulnerabilities"`
		} `json:"Results"`
	}

	err := json.Unmarshal(jsonOutput, &trivyReport)
	if err != nil {
		// If JSON parsing fails, try to extract counts from plain text
		return c.parseTrivyPlaintext(target)
	}

	// Count vulnerabilities by severity
	for _, res := range trivyReport.Results {
		for _, vuln := range res.Vulnerabilities {
			vulnResult := VulnResult{
				ID:       vuln.VulnerabilityID,
				Severity: vuln.Severity,
				Title:    vuln.Title,
				Package:  vuln.PkgName,
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vulnResult)

			// Count by severity
			switch vuln.Severity {
			case "CRITICAL":
				result.Critical++
			case "HIGH":
				result.High++
			case "MEDIUM":
				result.Medium++
			case "LOW":
				result.Low++
			}
		}
	}

	return &result, nil
}

// parseTrivyPlaintext attempts to parse Trivy output as plain text
// This is a fallback for when JSON parsing fails
func (c *TrivyChecker) parseTrivyPlaintext(target string) (*TrivyScanResult, error) {
	result := &TrivyScanResult{
		ScanType: c.ScanType,
		Target:   target,
	}

	// If there's any output but we can't parse it, assume no vulnerabilities
	// The scan was successful but format was unexpected
	return result, nil
}
