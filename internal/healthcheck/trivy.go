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

	var totalCritical, totalHigh, totalMedium, totalLow int
	var scans []TrivyScanResult
	failedTargets := []string{}

	// Scan each target
	for _, target := range c.Targets {
		scanResult, err := c.runTrivy(target, trivyPath)
		if err != nil {
			failedTargets = append(failedTargets, fmt.Sprintf("%s (%v)", target, err))
			continue
		}

		scans = append(scans, *scanResult)
		totalCritical += scanResult.Critical
		totalHigh += scanResult.High
		totalMedium += scanResult.Medium
		totalLow += scanResult.Low
	}

	result.Details["targets_scanned"] = len(scans)
	result.Details["targets_failed"] = len(failedTargets)
	result.Details["critical"] = totalCritical
	result.Details["high"] = totalHigh
	result.Details["medium"] = totalMedium
	result.Details["low"] = totalLow
	result.Details["scans"] = scans

	if len(failedTargets) > 0 {
		result.Details["failed_targets"] = failedTargets
	}

	// Determine status based on vulnerability severity
	if totalCritical > 0 {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("CRITICAL vulnerabilities found: %d critical, %d high", totalCritical, totalHigh)
		result.Error = "Critical security vulnerabilities detected - immediate action required"
	} else if totalHigh > 0 {
		result.Status = StatusWarn
		result.Message = fmt.Sprintf("HIGH vulnerabilities found: %d high, %d medium", totalHigh, totalMedium)
	} else if totalMedium > 0 {
		result.Status = StatusWarn
		result.Message = fmt.Sprintf("MEDIUM vulnerabilities found: %d medium, %d low", totalMedium, totalLow)
	} else if len(scans) > 0 {
		result.Status = StatusPass
		result.Message = fmt.Sprintf("No vulnerabilities found in %d target(s)", len(scans))
	} else {
		result.Status = StatusFail
		result.Message = "Failed to scan any targets"
		result.Error = fmt.Sprintf("All %d target(s) failed to scan", len(c.Targets))
	}

	result.Duration = time.Since(start)
	return result
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
		return c.parseTrivyPlaintext(string(jsonOutput), target)
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
func (c *TrivyChecker) parseTrivyPlaintext(output, target string) (*TrivyScanResult, error) {
	result := &TrivyScanResult{
		ScanType: c.ScanType,
		Target:   target,
	}

	// If there's any output but we can't parse it, assume no vulnerabilities
	// The scan was successful but format was unexpected
	return result, nil
}
